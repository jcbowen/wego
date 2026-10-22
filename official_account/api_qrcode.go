// Package official_account 微信公众号API - 参数二维码
// 本接口支持「服务号（仅认证）」账号类型调用。其他账号类型如无特殊说明，均不可调用
package official_account

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/jcbowen/wego/core"
)

const (
	ActionNameQRScene         = "QR_SCENE"           // 临时整型场景值
	ActionNameQRStrScene      = "QR_STR_SCENE"       // 临时字符串场景值
	ActionNameQRLimitScene    = "QR_LIMIT_SCENE"     // 永久整型场景值
	ActionNameQRLimitStrScene = "QR_LIMIT_STR_SCENE" // 永久字符串场景值
)

// QRCodeRequest 二维码请求
type QRCodeRequest struct {
	ExpireSeconds int64  `json:"expire_seconds,omitempty"` // 二维码有效时间（秒），最大2592000，仅临时二维码需要
	ActionName    string `json:"action_name"`              // 二维码类型：QR_SCENE(临时整型)/QR_STR_SCENE(临时字符串)/QR_LIMIT_SCENE(永久整型)/QR_LIMIT_STR_SCENE(永久字符串)
	ActionInfo    struct {
		Scene struct {
			SceneID  int    `json:"scene_id,omitempty"`  // 场景值ID，临时二维码时为32位非0整型，永久二维码时最大值为100000（目前参数只支持1--100000）
			SceneStr string `json:"scene_str,omitempty"` // 场景值ID（字符串形式的ID），字符串类型，长度限制为1到64
		} `json:"scene"` // 场景信息
	} `json:"action_info"` // 二维码详细信息
}

// QRCodeResponse 二维码响应
type QRCodeResponse struct {
	core.APIResponse
	Ticket        string `json:"ticket"`         // 获取的二维码ticket，凭借此ticket可以在有效时间内换取二维码
	ExpireSeconds int64  `json:"expire_seconds"` // 该二维码有效时间，以秒为单位。 最大不超过2592000（即30天）。
	URL           string `json:"url"`            // 二维码图片解析后的地址，开发者可根据该地址自行生成需要的二维码图片
}

type Qrcode struct {
	req *core.Request
}

func NewQrcode(req *core.Request) *Qrcode {
	return &Qrcode{req: req}
}

// Create 生成带参数的二维码
func (q *Qrcode) Create(ctx context.Context, qrCode *QRCodeRequest, accessToken string) (*QRCodeResponse, error) {
	// 验证参数
	if qrCode == nil {
		return nil, fmt.Errorf("二维码请求不能为空")
	}
	if qrCode.ActionName == "" {
		return nil, fmt.Errorf("二维码动作名称不能为空")
	}

	apiURL := fmt.Sprintf("%s?access_token=%s", URLQRCodeCreate, accessToken)

	var result QRCodeResponse
	err := q.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    apiURL,
		Body:   qrCode,
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

// CreateQRCode 生成带参数的二维码（Client 级入口，自动获取 access_token）
//
// 参数:
//   - ctx: 上下文
//   - req: 二维码请求参数（ActionName 取值见 ActionNameXxx 常量）
//
// 返回:
//   - *QRCodeResponse: 含 ticket 与二维码图片解析地址
//   - error: 请求失败或微信返回错误时返回错误
//
// 使用示例:
//
//	resp, err := client.CreateQRCode(ctx, &official_account.QRCodeRequest{
//		ExpireSeconds: 2592000,
//		ActionName:    official_account.ActionNameQRStrScene,
//		// ActionInfo.Scene.SceneStr = "invite_1001"
//	})
//	// 之后用 client.GetQRCodeImage(ctx, resp.Ticket) 换取图片，或直接用 resp.URL
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/Account_Management/Generating_a_Parametric_QR_Code.html
func (c *Client) CreateQRCode(ctx context.Context, req *QRCodeRequest) (*QRCodeResponse, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	return NewQrcode(c.req).Create(ctx, req, token)
}

// GetQRCodeImage 通过 ticket 换取二维码图片二进制（可直接保存或展示）
//
// 参数:
//   - ctx: 上下文
//   - ticket: 换取二维码的 ticket（来自 CreateQRCode 响应，需 urlencode）
//
// 返回:
//   - []byte: 图片二进制数据
//   - error: 请求失败时返回错误
//
// 说明:
//   - 仅支持服务号且已认证；ticket 有效期与二维码类型一致（临时码最长 30 天）
//
// 使用示例:
//
//	img, err := client.GetQRCodeImage(ctx, ticket)
//	if err == nil { os.WriteFile("qrcode.jpg", img, 0644) }
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/Account_Management/Generating_a_Parametric_QR_Code.html
func (c *Client) GetQRCodeImage(ctx context.Context, ticket string) ([]byte, error) {
	if ticket == "" {
		return nil, fmt.Errorf("二维码ticket不能为空")
	}

	showURL := fmt.Sprintf("%s?ticket=%s", URLQRCodeShow, url.QueryEscape(ticket))

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, showURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	resp, err := c.req.MakeRaw(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 先读响应体（图片仅几十 KB），非 2xx 时错误信息可携带响应体便于排障
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 官方对非法/过期 ticket 返回 404 等非 2xx 状态码，需按错误处理而非当图片返回
	if err := core.CheckRawResponseStatus(resp.StatusCode, data); err != nil {
		return nil, err
	}

	// 微信错误时返回 JSON 而非图片，检测并抛出错误码（以 Content-Type 为准，避免误判图片字节）
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		var apiResp core.APIResponse
		if err := json.Unmarshal(data, &apiResp); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}
		if !apiResp.IsSuccess() {
			return nil, &apiResp
		}
		return nil, fmt.Errorf("接口返回异常: 收到 JSON 成功响应但无图片数据")
	}

	return data, nil
}

// ShowQrcode 通过ticket换取二维码
// ticket正确情况下，http 返回码是200，是一张图片，可以直接展示或者下载
func (q *Qrcode) ShowQrcode(ticket string) (string, error) {
	// 验证参数
	if ticket == "" {
		return "", fmt.Errorf("二维码ticket不能为空")
	}

	return fmt.Sprintf("%s?ticket=%s", URLQRCodeShow, url.QueryEscape(ticket)), nil
}
