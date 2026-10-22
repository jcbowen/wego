package official_account

import (
	"context"
	"fmt"
	"time"

	"github.com/jcbowen/wego/core"
)

// GetJSAPITicketResponse 获取 JSAPI ticket 响应
type GetJSAPITicketResponse struct {
	core.APIResponse
	Ticket    string `json:"ticket"`     // JSAPI ticket（用于 JSSDK 签名）
	ExpiresIn int    `json:"expires_in"` // 有效期（秒，一般为 7200）
}

// GetJSAPITicket 获取公众号 JSAPI ticket（带内存缓存，提前 5 分钟过期，供 JSSDK 签名使用）
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - string: JSAPI ticket
//   - error: 请求失败或微信返回错误时返回错误
//
// 使用示例:
//
//	ticket, err := wegoClient.OfficialAccountClient.GetJSAPITicket(ctx)
//	// 之后用 ticket + noncestr + timestamp + url 计算 JSSDK 签名
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/JS-SDK.html
func (c *Client) GetJSAPITicket(ctx context.Context) (string, error) {
	c.jsapiTicketMu.Lock()
	defer c.jsapiTicketMu.Unlock()

	// 缓存命中（考虑安全余量）
	if c.jsapiTicket != "" && time.Now().Before(c.jsapiTicketExpiresAt) {
		return c.jsapiTicket, nil
	}

	accessToken, err := c.GetAccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("获取 access_token 失败: %w", err)
	}

	var result GetJSAPITicketResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    URLGetTicket,
		Query: map[string]string{
			"access_token": accessToken,
			"type":         "jsapi",
		},
		Result: &result,
	})
	if err != nil {
		return "", err
	}

	if !result.IsSuccess() {
		return "", &result.APIResponse
	}

	// 有效期减去 5 分钟安全余量，避免临界过期
	c.jsapiTicket = result.Ticket
	c.jsapiTicketExpiresAt = time.Now().Add(time.Duration(result.ExpiresIn)*time.Second - accessTokenExpireMargin)

	return c.jsapiTicket, nil
}

// InvalidateJSAPITicket 使缓存的 JSAPI ticket 失效（遇到 40001/42001 等失效错误码后可调用强制刷新）
func (c *Client) InvalidateJSAPITicket() {
	c.jsapiTicketMu.Lock()
	defer c.jsapiTicketMu.Unlock()
	c.jsapiTicket = ""
	c.jsapiTicketExpiresAt = time.Time{}
}
