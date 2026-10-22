// Package openplatform 微信开放平台API - 代小程序代码管理
// 参考：https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/code/commit.html
package openplatform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/jcbowen/wego/core"
)

// 代小程序审核状态（GetWxaAuditStatusResponse.Status）
const (
	WxaAuditStatusApproved  = 0 // 审核成功
	WxaAuditStatusRejected  = 1 // 审核被拒绝
	WxaAuditStatusAuditing  = 2 // 审核中
	WxaAuditStatusWithdrawn = 3 // 已撤回
	WxaAuditStatusDelayed   = 4 // 审核延后
)

// CommitWxaCodeRequest 上传并发布小程序代码请求（POST /wxa/commit）
//
// 注意：
//   - 提交成功后会同时生成体验版
//   - 标准模板库的模板 ext_json 仅支持 {"extAppid":"", "ext": {}, "window": {}}
type CommitWxaCodeRequest struct {
	TemplateID  string `json:"template_id"`  // 代码模板 ID（从代码模板列表获取）
	ExtJSON     string `json:"ext_json"`     // ext.json 配置内容（JSON 字符串，支持 extAppid/ext/extPages/pages/window 等）
	UserVersion string `json:"user_version"` // 代码版本号（长度不超过 64 字符）
	UserDesc    string `json:"user_desc"`    // 代码描述
}

// WxaCategory 小程序类目信息（get_category 响应项，也是提审时 item_list 的类目来源）
type WxaCategory struct {
	FirstClass  string `json:"first_class,omitempty"`  // 一级类目名称
	SecondClass string `json:"second_class,omitempty"` // 二级类目名称
	ThirdClass  string `json:"third_class,omitempty"`  // 三级类目名称（可选）
	FirstID     int64  `json:"first_id,omitempty"`     // 一级类目 ID
	SecondID    int64  `json:"second_id,omitempty"`    // 二级类目 ID
	ThirdID     int64  `json:"third_id,omitempty"`     // 三级类目 ID（可选）
}

// GetWxaCategoryResponse 获取审核时可填写的类目信息响应（GET /wxa/get_category）
type GetWxaCategoryResponse struct {
	core.APIResponse
	CategoryList []WxaCategory `json:"category_list"` // 已设置的二级类目及可选三级类目列表
}

// WxaAuditItem 提交审核的审核项
//
// 类目（first_class/second_class/third_class 及对应 ID）必须来自 GetWxaCategory 已配置的类目。
type WxaAuditItem struct {
	Address     string `json:"address,omitempty"`     // 小程序页面路径（来自 get_page 接口）
	Tag         string `json:"tag,omitempty"`         // 标签（空格分隔，至多 10 个，单个长度至多 20）
	FirstClass  string `json:"first_class"`           // 一级类目名称（必填）
	SecondClass string `json:"second_class"`          // 二级类目名称（必填）
	ThirdClass  string `json:"third_class,omitempty"` // 三级类目名称
	FirstID     int64  `json:"first_id"`              // 一级类目 ID（必填）
	SecondID    int64  `json:"second_id"`             // 二级类目 ID（必填）
	ThirdID     int64  `json:"third_id,omitempty"`    // 三级类目 ID
	Title       string `json:"title,omitempty"`       // 页面标题（长度至多 32）
}

// WxaAuditPreviewInfo 提审预览信息（小程序页面截图和操作录屏，mediaid 来自 UploadWxaAuditMedia，有效期 3 天）
type WxaAuditPreviewInfo struct {
	VideoIDList []string `json:"video_id_list,omitempty"` // 录屏 mediaid 列表
	PicIDList   []string `json:"pic_id_list,omitempty"`   // 截屏 mediaid 列表
}

// WxaUGCDeclare 用户生成内容（UGC）场景信息安全声明
//
// 当小程序第一次提交审核且类目包含社交-社区/论坛、社交-笔记、社交-问答其中之一时必填。
type WxaUGCDeclare struct {
	Scene          []int  `json:"scene,omitempty"`            // UGC 场景：0不涉及/1用户资料/2图片/3视频/4文本/5其他（可多选，填 0 时无需其他字段）
	OtherSceneDesc string `json:"other_scene_desc,omitempty"` // scene 选"其他"时的说明（不超 256 字）
	Method         []int  `json:"method,omitempty"`           // 内容安全机制：1平台API/2其他审核产品/3人工审核/4未做审核
	HasAuditTeam   int    `json:"has_audit_team,omitempty"`   // 是否有审核团队：0无/1有
	AuditDesc      string `json:"audit_desc,omitempty"`       // 审核机制说明（不超 256 字）
}

// SubmitWxaAuditRequest 提交小程序代码审核请求（POST /wxa/submit_audit）
//
// 注意：
//   - item_list 至多 5 项，类目必填且须为小程序已配置的类目
//   - 仅上个版本被驳回时 feedback_info/feedback_stuff 才生效
type SubmitWxaAuditRequest struct {
	ItemList      []WxaAuditItem       `json:"item_list"`                // 审核项列表（1-5 项）
	PreviewInfo   *WxaAuditPreviewInfo `json:"preview_info,omitempty"`   // 预览信息
	VersionDesc   string               `json:"version_desc,omitempty"`   // 版本说明和功能解释
	FeedbackInfo  string               `json:"feedback_info,omitempty"`  // 反馈内容（至多 200 字，仅上次被驳回时生效）
	FeedbackStuff string               `json:"feedback_stuff,omitempty"` // 用 | 分割的 media_id 列表（至多 5 张）
	UGCDeclare    *WxaUGCDeclare       `json:"ugc_declare,omitempty"`    // UGC 信息安全声明
}

// SubmitWxaAuditResponse 提交审核响应
type SubmitWxaAuditResponse struct {
	core.APIResponse
	AuditID int64 `json:"auditid"` // 审核编号（查询审核状态时使用）
}

// GetWxaAuditStatusRequest 查询审核状态请求（POST /wxa/get_auditstatus）
type GetWxaAuditStatusRequest struct {
	AuditID int64 `json:"auditid"` // 提交审核时获得的审核 id
}

// GetWxaAuditStatusResponse 查询审核状态响应
type GetWxaAuditStatusResponse struct {
	core.APIResponse
	Status     int    `json:"status"`               // 审核状态：0成功/1拒绝/2审核中/3已撤回/4延后
	Reason     string `json:"reason,omitempty"`     // status=1 拒绝原因；status=4 延后原因
	Screenshot string `json:"screenshot,omitempty"` // status=1 时审核失败截图，用 | 分隔的 media_id 列表
}

// UploadWxaAuditMediaResponse 提审素材上传响应（POST /wxa/uploadmedia，multipart）
type UploadWxaAuditMediaResponse struct {
	core.APIResponse
	Type    string `json:"type"`    // 素材类型（image/video）
	MediaID string `json:"mediaid"` // 素材 mediaid（有效期 3 天，用于 preview_info）
}

// WxaTrialQrcodeResponse 体验版二维码响应（GET /wxa/get_qrcode）
//
// 成功时 Data 为图片二进制（image/jpeg），可直接保存为图片；失败时微信返回 JSON 错误。
type WxaTrialQrcodeResponse struct {
	core.APIResponse
	Data        []byte `json:"-"` // 图片二进制数据
	ContentType string `json:"-"` // 响应 Content-Type（如 image/jpeg）
}

// CommitWxaCode 为已授权的小程序上传代码并生成体验版
//
// 参数:
//   - ctx: 上下文
//   - authorizerAppID: 授权方小程序 AppID
//   - req: 上传代码请求参数
//
// 返回:
//   - *core.APIResponse: 接口响应（errcode=0 表示成功）
//   - error: 请求失败或微信返回错误时返回错误
//
// 使用示例:
//
//	err := openPlatformClient.CommitWxaCode(ctx, authorizerAppID, &openplatform.CommitWxaCodeRequest{
//		TemplateID:  "12",
//		ExtJSON:     `{"extAppid":"wx123","ext":{"attr1":"v1"}}`,
//		UserVersion: "V1.0",
//		UserDesc:    "首个版本",
//	})
//
// 文档:
//   - https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/code/commit.html
func (c *Client) CommitWxaCode(ctx context.Context, authorizerAppID string, req *CommitWxaCodeRequest) (*core.APIResponse, error) {
	if authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}
	if req == nil || req.TemplateID == "" {
		return nil, fmt.Errorf("template_id 不能为空")
	}

	token, err := c.GetAuthorizerAccessToken(ctx, authorizerAppID)
	if err != nil {
		return nil, err
	}

	var result core.APIResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLWxaCommit,
		Query:  map[string]string{"access_token": token},
		Body:   req,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result
	}
	return &result, nil
}

// GetWxaCategory 获取审核时可填写的类目信息（已设置的二级类目及可选三级类目）
//
// 参数:
//   - ctx: 上下文
//   - authorizerAppID: 授权方小程序 AppID
//
// 返回:
//   - *GetWxaCategoryResponse: 类目列表
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/category/get_category.html
func (c *Client) GetWxaCategory(ctx context.Context, authorizerAppID string) (*GetWxaCategoryResponse, error) {
	if authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	token, err := c.GetAuthorizerAccessToken(ctx, authorizerAppID)
	if err != nil {
		return nil, err
	}

	var result GetWxaCategoryResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    URLWxaGetCategory,
		Query:  map[string]string{"access_token": token},
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}
	return &result, nil
}

// SubmitWxaAudit 将上传的代码提交审核
//
// 参数:
//   - ctx: 上下文
//   - authorizerAppID: 授权方小程序 AppID
//   - req: 提审请求参数（item_list 1-5 项，类目来自 GetWxaCategory）
//
// 返回:
//   - *SubmitWxaAuditResponse: 含审核编号 auditid
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/code/submit_audit.html
func (c *Client) SubmitWxaAudit(ctx context.Context, authorizerAppID string, req *SubmitWxaAuditRequest) (*SubmitWxaAuditResponse, error) {
	if authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}
	if req == nil || len(req.ItemList) == 0 || len(req.ItemList) > 5 {
		return nil, fmt.Errorf("item_list 必须为 1-5 项")
	}

	token, err := c.GetAuthorizerAccessToken(ctx, authorizerAppID)
	if err != nil {
		return nil, err
	}

	var result SubmitWxaAuditResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLWxaSubmitAudit,
		Query:  map[string]string{"access_token": token},
		Body:   req,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}
	return &result, nil
}

// UploadWxaAuditMedia 上传提审素材（小程序页面截图/操作录屏），返回的 mediaid 用于提审 preview_info
//
// 参数:
//   - ctx: 上下文
//   - authorizerAppID: 授权方小程序 AppID
//   - filename: 文件名（用于 multipart 表单，如 "shot.png"/"demo.mp4"）
//   - media: 素材二进制内容（图片 2M：PNG/JPEG/JPG/GIF；视频 10MB：MP4）
//
// 返回:
//   - *UploadWxaAuditMediaResponse: 含 mediaid（有效期 3 天）
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/code/submit_audit.html
func (c *Client) UploadWxaAuditMedia(ctx context.Context, authorizerAppID, filename string, media []byte) (*UploadWxaAuditMediaResponse, error) {
	if authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}
	if len(media) == 0 {
		return nil, fmt.Errorf("素材内容不能为空")
	}

	token, err := c.GetAuthorizerAccessToken(ctx, authorizerAppID)
	if err != nil {
		return nil, err
	}

	body, contentType, err := buildMultipartFileBody("media", filename, media)
	if err != nil {
		return nil, err
	}

	var result UploadWxaAuditMediaResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLWxaUploadAuditMedia,
		Query:  map[string]string{"access_token": token},
		Body:   body,
		Headers: map[string]string{
			"Content-Type": contentType,
		},
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}
	return &result, nil
}

// GetWxaAuditStatus 查询指定发布审核单的审核状态
//
// 参数:
//   - ctx: 上下文
//   - authorizerAppID: 授权方小程序 AppID
//   - auditID: 提交审核时获得的审核 id
//
// 返回:
//   - *GetWxaAuditStatusResponse: Status 取值见 WxaAuditStatusXxx 常量
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/code/get_auditstatus.html
func (c *Client) GetWxaAuditStatus(ctx context.Context, authorizerAppID string, auditID int64) (*GetWxaAuditStatusResponse, error) {
	if authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}
	if auditID == 0 {
		return nil, fmt.Errorf("auditid 不能为空")
	}

	token, err := c.GetAuthorizerAccessToken(ctx, authorizerAppID)
	if err != nil {
		return nil, err
	}

	var result GetWxaAuditStatusResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLWxaGetAuditStatus,
		Query:  map[string]string{"access_token": token},
		Body:   GetWxaAuditStatusRequest{AuditID: auditID},
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}
	return &result, nil
}

// ReleaseWxaCode 发布最后一个审核通过的小程序代码版本
//
// 参数:
//   - ctx: 上下文
//   - authorizerAppID: 授权方小程序 AppID
//
// 返回:
//   - *core.APIResponse: 接口响应
//   - error: 请求失败或微信返回错误时返回错误
//
// 说明:
//   - 官方要求 POST 必须携带空 JSON 对象 "{}"，否则报 44002（empty post data）
//
// 文档:
//   - https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/code/release.html
func (c *Client) ReleaseWxaCode(ctx context.Context, authorizerAppID string) (*core.APIResponse, error) {
	if authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	token, err := c.GetAuthorizerAccessToken(ctx, authorizerAppID)
	if err != nil {
		return nil, err
	}

	var result core.APIResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLWxaRelease,
		Query:  map[string]string{"access_token": token},
		Body:   struct{}{}, // 官方要求：post 的 data 为空对象 {}，不能不传
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result
	}
	return &result, nil
}

// RevertWxaCodeRelease 回退小程序线上版本
//
// 参数:
//   - ctx: 上下文
//   - authorizerAppID: 授权方小程序 AppID
//   - appVersion: 目标版本号，为空时回滚到上一个线上版本
//
// 返回:
//   - *core.APIResponse: 接口响应
//   - error: 请求失败或微信返回错误时返回错误
//
// 说明:
//   - 无上一个线上版本无法回退；最多保存最近发布/回退的 5 个版本；当前版本已回退过则不能再回退
//   - 该接口为 GET 请求，版本号通过 query 参数 app_version 传递
//
// 文档:
//   - https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/code/revertcoderelease.html
func (c *Client) RevertWxaCodeRelease(ctx context.Context, authorizerAppID, appVersion string) (*core.APIResponse, error) {
	if authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	token, err := c.GetAuthorizerAccessToken(ctx, authorizerAppID)
	if err != nil {
		return nil, err
	}

	query := map[string]string{"access_token": token}
	if appVersion != "" {
		query["app_version"] = appVersion
	}

	var result core.APIResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    URLWxaRevertRelease,
		Query:  query,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result
	}
	return &result, nil
}

// GetWxaTrialQrcode 获取授权小程序的体验版二维码
//
// 参数:
//   - ctx: 上下文
//   - authorizerAppID: 授权方小程序 AppID
//   - path: 扫码后进入的页面路径（可带参数，如 "page/index?action=1"），为空进入默认主页
//
// 返回:
//   - *WxaTrialQrcodeResponse: Data 为图片二进制，可直接保存为图片
//   - error: 请求失败或微信返回错误时返回错误
//
// 使用示例:
//
//	resp, err := openPlatformClient.GetWxaTrialQrcode(ctx, authorizerAppID, "pages/index/index")
//	if err == nil { os.WriteFile("trial.jpg", resp.Data, 0644) }
//
// 文档:
//   - https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/code/get_qrcode.html
func (c *Client) GetWxaTrialQrcode(ctx context.Context, authorizerAppID, path string) (*WxaTrialQrcodeResponse, error) {
	if authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	token, err := c.GetAuthorizerAccessToken(ctx, authorizerAppID)
	if err != nil {
		return nil, err
	}

	// path 需 urlencode：如 page/index?action=1 需填入 page%2Findex%3Faction%3D1
	fullURL := fmt.Sprintf("%s?access_token=%s", URLWxaGetQrcode, url.QueryEscape(token))
	if path != "" {
		fullURL += "&path=" + url.QueryEscape(path)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	resp, err := c.req.MakeRaw(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	// JSON 响应说明获取失败，解析错误码（正常成功时返回 image/jpeg 二进制）
	if strings.Contains(contentType, "application/json") {
		var apiResp core.APIResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}
		if !apiResp.IsSuccess() {
			return nil, &apiResp
		}
		return nil, fmt.Errorf("接口返回异常: 收到 JSON 成功响应但无二维码数据, errmsg: %s", apiResp.ErrMsg)
	}

	return &WxaTrialQrcodeResponse{
		APIResponse: core.APIResponse{ErrCode: 0, ErrMsg: "success"},
		Data:        respBody,
		ContentType: contentType,
	}, nil
}

// buildMultipartFileBody 构造单文件 multipart/form-data 请求体
//
// 参数:
//   - fieldName: 表单字段名（微信素材类接口固定为 media）
//   - filename: 文件名
//   - content: 文件二进制内容
//
// 返回:
//   - []byte: 请求体内容
//   - string: 需要设置的 Content-Type（含 boundary）
//   - error: 构造失败时返回错误
func buildMultipartFileBody(fieldName, filename string, content []byte) ([]byte, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return nil, "", fmt.Errorf("创建multipart表单失败: %w", err)
	}
	if _, err := part.Write(content); err != nil {
		return nil, "", fmt.Errorf("写入multipart表单失败: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("关闭multipart writer失败: %w", err)
	}

	return buf.Bytes(), writer.FormDataContentType(), nil
}
