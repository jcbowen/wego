package miniprogram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jcbowen/wego/core"
)

// 小程序码 env_version 取值
const (
	EnvVersionRelease = "release" // 正式版
	EnvVersionTrial   = "trial"   // 体验版
	EnvVersionDevelop = "develop" // 开发版
)

// GetWxaCodeRequest 获取小程序码请求（getwxacode）
//
// 注意：path 为官方必填参数（最长 1024 字符，可携带页面参数，空将返回 40159 错误），
// 与 createwxaqrcode 生成的码总数上限共 10 万（超限报 45029），如需大量生成请使用 getwxacodeunlimit。
type GetWxaCodeRequest struct {
	Path       string     `json:"path"`                  // 页面路径（官方必填，最长 1024 字符）
	Width      int        `json:"width,omitempty"`       // 小程序码宽度（280-1280，默认 430）
	AutoColor  bool       `json:"auto_color"`            // 自动配置线条颜色
	LineColor  *LineColor `json:"line_color,omitempty"`  // 线条颜色（auto_color 为 false 时生效）
	IsHyaline  bool       `json:"is_hyaline"`            // 是否需要透明底色
	EnvVersion string     `json:"env_version,omitempty"` // 小程序版本：release/trial/develop
}

// LineColor 小程序码线条颜色（RGB）
type LineColor struct {
	R int `json:"r"` // 红色分量（0-255）
	G int `json:"g"` // 绿色分量（0-255）
	B int `json:"b"` // 蓝色分量（0-255）
}

// GetWxaCodeUnlimitRequest 获取小程序码请求（getwxacodeunlimit）
//
// 注意：scene 参数即小程序码携带的业务参数（最长 32 个可见字符），页面通过 onLoad(query.scene) 获取。
type GetWxaCodeUnlimitRequest struct {
	Scene      string     `json:"scene"`                 // 业务参数（最长 32 个可见字符，必填）
	Page       string     `json:"page,omitempty"`        // 页面路径（必须为已发布页面），空则默认主页
	CheckPath  *bool      `json:"check_path,omitempty"`  // 校验 page 是否存在（不传时官方默认 true；传 false 时不校验，未发布页面可能无法正常打开）
	EnvVersion string     `json:"env_version,omitempty"` // 小程序版本：release/trial/develop
	Width      int        `json:"width,omitempty"`       // 小程序码宽度（280-1280，默认 430）
	AutoColor  bool       `json:"auto_color"`            // 自动配置线条颜色
	LineColor  *LineColor `json:"line_color,omitempty"`  // 线条颜色（auto_color 为 false 时生效）
	IsHyaline  bool       `json:"is_hyaline"`            // 是否需要透明底色
}

// CreateWxaQrcodeRequest 获取小程序二维码请求（createwxaqrcode，旧接口）
//
// 注意：path 长度小于 128 字符，总数量 10 万个，生成后数量无法增减。
type CreateWxaQrcodeRequest struct {
	Path  string `json:"path"`            // 页面路径（不能带参数，官方必填，长度小于 128 字符）
	Width int    `json:"width,omitempty"` // 二维码宽度（280-1280，默认 430）
}

// WXACodeResponse 小程序码/二维码接口响应
//
// 成功时 Data 为图片二进制数据，ContentType 为图片类型；失败时微信返回 JSON 错误。
type WXACodeResponse struct {
	core.APIResponse
	Data        []byte `json:"-"` // 图片二进制数据
	ContentType string `json:"-"` // 响应 Content-Type（如 image/jpeg）
	NumUsed     string `json:"-"` // 响应头 Num-Used（已生成码数量，与 10 万总量上限对照；官方在 getwxacode/createQRCode 文档记载该头，getwxacodeunlimit 无总量限制，官方未记载其返回，为空属正常）
}

// GetWxaCode 获取不限制数量之外的小程序码（带 path，总数上限 10 万）
//
// 参数:
//   - ctx: 上下文
//   - req: 请求参数
//
// 返回:
//   - *WXACodeResponse: Data 为图片二进制数据
//   - error: 请求失败或微信返回错误时返回错误
//
// 使用示例:
//
//	resp, err := client.GetWxaCode(ctx, &miniprogram.GetWxaCodeRequest{Path: "pages/index/index", Width: 430})
//	if err == nil { os.WriteFile("qrcode.jpg", resp.Data, 0644) }
//
// 文档:
//   - https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/qrcode-link/qr-code/getWXACode.html
func (c *Client) GetWxaCode(ctx context.Context, req *GetWxaCodeRequest) (*WXACodeResponse, error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path 不能为空（官方必填参数，空将返回 40159 错误）")
	}
	return c.postWXACodeBinary(ctx, GetWxaCodeURL, req)
}

// GetWxaCodeUnlimit 获取不限量的小程序码（带 scene 参数，数量暂无限制）
//
// 参数:
//   - ctx: 上下文
//   - req: 请求参数（Scene 必填，最长 32 个可见字符）
//
// 返回:
//   - *WXACodeResponse: Data 为图片二进制数据
//   - error: 请求失败或微信返回错误时返回错误
//
// 使用示例:
//
//	resp, err := client.GetWxaCodeUnlimit(ctx, &miniprogram.GetWxaCodeUnlimitRequest{
//		Scene: "id=123",
//		Page:  "pages/detail/index",
//	})
//	if err == nil { os.WriteFile("qrcode.jpg", resp.Data, 0644) }
//
// 文档:
//   - https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/qrcode-link/qr-code/getUnlimitedQRCode.html
func (c *Client) GetWxaCodeUnlimit(ctx context.Context, req *GetWxaCodeUnlimitRequest) (*WXACodeResponse, error) {
	if req.Scene == "" {
		return nil, fmt.Errorf("scene 不能为空（最长 32 个可见字符）")
	}
	return c.postWXACodeBinary(ctx, GetWxaCodeUnlimitURL, req)
}

// CreateWxaQrcode 获取小程序二维码（旧接口，带 path，总数上限 10 万）
//
// 参数:
//   - ctx: 上下文
//   - req: 请求参数
//
// 返回:
//   - *WXACodeResponse: Data 为图片二进制数据
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/qrcode-link/qr-code/createQRCode.html
func (c *Client) CreateWxaQrcode(ctx context.Context, req *CreateWxaQrcodeRequest) (*WXACodeResponse, error) {
	if req.Path == "" {
		return nil, fmt.Errorf("path 不能为空（官方必填参数，长度小于 128 字符）")
	}
	return c.postWXACodeBinary(ctx, CreateWxaQrcodeURL, req)
}

// postWXACodeBinary 发送小程序码类接口请求并处理「成功二进制 / 失败 JSON」的混合响应
//
// 参数:
//   - ctx: 上下文
//   - apiURL: 接口地址（不含 access_token）
//   - body: 请求参数结构体
//
// 返回:
//   - *WXACodeResponse: 成功时 Data 为图片二进制；失败时返回微信 JSON 错误
//   - error: 请求失败或微信返回错误时返回错误
func (c *Client) postWXACodeBinary(ctx context.Context, apiURL string, body interface{}) (*WXACodeResponse, error) {
	accessToken, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf("%s?access_token=%s", apiURL, accessToken)

	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.req.MakeRaw(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 非 2xx 状态码提前判错，避免把网关/代理层错误当小程序码数据返回
	if err := core.CheckRawResponseStatus(resp.StatusCode, respBody); err != nil {
		return nil, err
	}

	contentType := resp.Header.Get("Content-Type")
	// JSON 响应说明生成失败，解析错误码（正常成功时微信返回 image/jpeg 二进制）
	if strings.Contains(contentType, "application/json") {
		var apiResp core.APIResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			return nil, fmt.Errorf("解析响应失败: %v", err)
		}
		if !apiResp.IsSuccess() {
			return nil, &apiResp
		}
		// errcode==0 的 JSON 响应不携带图片数据，正常不应出现；显式报错避免把 JSON 当图片返回
		return nil, fmt.Errorf("接口返回异常: 收到 JSON 成功响应但无图片数据, errmsg: %s", apiResp.ErrMsg)
	}

	return &WXACodeResponse{
		APIResponse: core.APIResponse{ErrCode: 0, ErrMsg: "success"},
		Data:        respBody,
		ContentType: contentType,
		NumUsed:     resp.Header.Get("Num-Used"),
	}, nil
}
