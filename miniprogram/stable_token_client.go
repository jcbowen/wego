package miniprogram

import (
	"context"

	"github.com/jcbowen/wego/core"
)

// StableAccessTokenRequest 获取稳定版 access_token 请求参数
type StableAccessTokenRequest struct {
	GrantType    string `json:"grant_type"`              // 获取 access_token 固定填 client_credential
	AppID        string `json:"appid"`                   // 小程序 appid
	Secret       string `json:"secret"`                  // 小程序 appsecret
	ForceRefresh bool   `json:"force_refresh,omitempty"` // 是否强制刷新，true-是；false-否
}

// StableAccessTokenResponse 获取稳定版 access_token 响应
type StableAccessTokenResponse struct {
	core.APIResponse
	AccessToken string `json:"access_token"` // 获取到的凭证
	ExpiresIn   int    `json:"expires_in"`   // 凭证有效时间，单位：秒（正常为 7200）
}

// GetStableAccessToken 获取稳定版 access_token（POST /cgi-bin/stable_token）
//
// 稳定版 access_token：结果不下发/不依赖缓存，单 appid 调用频率限制见官方文档。
// 普通模式（forceRefresh=false）有效期内重复调用返回相同 token；
// 强制刷新（forceRefresh=true）受官方限频约束（每天次数与最小间隔限制，见官方文档）。
//
// 注意：本方法仅直接调用接口并返回结果，不写入现有 token 缓存，
// 也不会改变 GetAccessToken 的默认 token 获取路径。
//
// 参数:
//   - ctx: 上下文
//   - forceRefresh: 是否强制刷新（对应官方请求字段 force_refresh）
//
// 返回:
//   - *StableAccessTokenResponse: 包含 AccessToken 与 ExpiresIn
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-access-token/getStableAccessToken.html
func (c *Client) GetStableAccessToken(ctx context.Context, forceRefresh bool) (*StableAccessTokenResponse, error) {
	var result StableAccessTokenResponse
	err := c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    StableTokenURL,
		Body: StableAccessTokenRequest{
			GrantType:    GrantTypeClientCredential,
			AppID:        c.config.AppID,
			Secret:       c.config.AppSecret,
			ForceRefresh: forceRefresh,
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
