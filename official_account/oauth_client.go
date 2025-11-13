package official_account

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/wego/core"
)

// OAuthClient 网页授权客户端
type OAuthClient struct {
	client *Client
	logger debugger.LoggerInterface
}

// NewOAuthClient 创建网页授权客户端
func NewOAuthClient(client *Client) *OAuthClient {
	return &OAuthClient{
		client: client,
		logger: client.GetLogger(),
	}
}

// GenerateAuthorizeURL 生成网页授权URL
// @param ctx 上下文
// @param req 授权请求参数
// @return 授权URL
// @return 错误信息
func (o *OAuthClient) GenerateAuthorizeURL(ctx context.Context, req *OAuthAuthorizeRequest) (string, error) {
	if req.AppID == "" {
		return "", fmt.Errorf("AppID不能为空")
	}
	if req.RedirectURI == "" {
		return "", fmt.Errorf("RedirectURI不能为空")
	}
	if req.Scope == "" {
		return "", fmt.Errorf("Scope不能为空")
	}

	// 验证scope值
	if req.Scope != core.OAuthScopeBase && req.Scope != core.OAuthScopeUserInfo {
		return "", fmt.Errorf("无效的scope值: %s", req.Scope)
	}

	params := url.Values{}
	params.Set("appid", req.AppID)
	params.Set("redirect_uri", req.RedirectURI)
	params.Set("response_type", core.ResponseTypeCode)
	params.Set("scope", req.Scope)
	if req.State != "" {
		params.Set("state", req.State)
	}
	params.Set("connect_redirect", "1")

	authorizeURL := URLConnectOAuth2Authorize + "?" + params.Encode() + "#wechat_redirect"

    o.logger.Info("生成网页授权URL", map[string]interface{}{"url": authorizeURL})
	return authorizeURL, nil
}

// GetAccessToken 通过code获取网页授权access_token
// @param ctx 上下文
// @param code 授权码
// @return 授权响应
// @return 错误信息
func (o *OAuthClient) GetAccessToken(ctx context.Context, code string) (*OAuthAccessTokenResponse, error) {
	if code == "" {
		return nil, fmt.Errorf("code不能为空")
	}

	config := o.client.GetConfig()
	req := &OAuthAccessTokenRequest{
		AppID:     config.AppID,
		Secret:    config.AppSecret,
		Code:      code,
		GrantType: core.GrantTypeAuthorizationCode,
	}

	var resp OAuthAccessTokenResponse
	err := o.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    URLSnsOAuth2AccessToken,
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, fmt.Errorf("获取网页授权access_token失败: %v", err)
	}

	if !resp.IsSuccess() {
		return nil, &resp.APIResponse
	}

    o.logger.Info("获取网页授权access_token成功", map[string]interface{}{"openid": resp.OpenID, "scope": resp.Scope})
	return &resp, nil
}

// RefreshAccessToken 刷新网页授权access_token
// @param ctx 上下文
// @param refreshToken 刷新token
// @return 授权响应
// @return 错误信息
func (o *OAuthClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*OAuthAccessTokenResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refreshToken不能为空")
	}

	config := o.client.GetConfig()
	req := &OAuthRefreshTokenRequest{
		AppID:        config.AppID,
		RefreshToken: refreshToken,
		GrantType:    core.GrantTypeRefreshToken,
	}

	var resp OAuthAccessTokenResponse
	err := o.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    URLSnsOAuth2RefreshToken,
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, fmt.Errorf("刷新网页授权access_token失败: %v", err)
	}

	if !resp.IsSuccess() {
		return nil, &resp.APIResponse
	}

    o.logger.Info("刷新网页授权access_token成功", map[string]interface{}{"openid": resp.OpenID})
	return &resp, nil
}

// GetUserInfo 获取用户信息
// @param ctx 上下文
// @param accessToken 网页授权access_token
// @param openID 用户openid
// @param lang 语言，默认为zh_CN
// @return 用户信息
// @return 错误信息
func (o *OAuthClient) GetUserInfo(ctx context.Context, accessToken, openID, lang string) (*OAuthUserInfoResponse, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("accessToken不能为空")
	}
	if openID == "" {
		return nil, fmt.Errorf("openID不能为空")
	}

	if lang == "" {
		lang = "zh_CN"
	}

	req := &OAuthUserInfoRequest{
		AccessToken: accessToken,
		OpenID:      openID,
		Lang:        lang,
	}

	var resp OAuthUserInfoResponse
	err := o.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    URLSnsUserInfo,
		Query:  req,
		Result: &resp,
	})
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %v", err)
	}

	if !resp.IsSuccess() {
		return nil, &resp.APIResponse
	}

    o.logger.Info("获取用户信息成功", map[string]interface{}{"openid": resp.OpenID, "nickname": resp.Nickname})
	return &resp, nil
}

// ValidateAccessToken 检验授权凭证（access_token）是否有效
// @param ctx 上下文
// @param accessToken 网页授权access_token
// @param openID 用户openid
// @return 是否有效
// @return 错误信息
func (o *OAuthClient) ValidateAccessToken(ctx context.Context, accessToken, openID string) (bool, error) {
	if accessToken == "" {
		return false, fmt.Errorf("accessToken不能为空")
	}
	if openID == "" {
		return false, fmt.Errorf("openID不能为空")
	}

	req := &OAuthAuthRequest{
		AccessToken: accessToken,
		OpenID:      openID,
	}

	var resp core.APIResponse
	err := o.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    URLSnsAuth,
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return false, fmt.Errorf("检验access_token失败: %v", err)
	}

	// 如果返回的错误码为0，则表示有效
	if resp.ErrCode == 0 {
        o.logger.Info("检验access_token有效", map[string]interface{}{"openid": openID})
		return true, nil
	}

    o.logger.Warn("检验access_token无效", map[string]interface{}{"openid": openID, "errcode": resp.ErrCode, "errmsg": resp.ErrMsg})
	return false, nil
}

// CompleteOAuthFlow 完整的网页授权流程
// @param ctx 上下文
// @param code 授权码
// @param scope 授权作用域
// @return 用户信息（如果scope为snsapi_userinfo）
// @return 授权响应（包含openid等基础信息）
// @return 错误信息
func (o *OAuthClient) CompleteOAuthFlow(ctx context.Context, code, scope string) (*OAuthUserInfoResponse, *OAuthAccessTokenResponse, error) {
	// 1. 获取access_token
	accessTokenResp, err := o.GetAccessToken(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("获取access_token失败: %v", err)
	}

	// 2. 如果scope是snsapi_userinfo，获取用户信息
	var userInfo *OAuthUserInfoResponse
	if scope == "snsapi_userinfo" {
		userInfo, err = o.GetUserInfo(ctx, accessTokenResp.AccessToken, accessTokenResp.OpenID, "zh_CN")
		if err != nil {
			o.logger.Warn("获取用户信息失败", map[string]interface{}{"error": err.Error()})
			// 不返回错误，继续返回基础信息
		}
	}

	return userInfo, accessTokenResp, nil
}

// BuildOAuthURL 快速构建授权URL的便捷方法
// @param ctx 上下文
// @param redirectURI 重定向URI
// @param scope 授权作用域
// @param state 状态参数
// @return 授权URL
// @return 错误信息
func (o *OAuthClient) BuildOAuthURL(ctx context.Context, redirectURI, scope, state string) (string, error) {
	config := o.client.GetConfig()
	req := &OAuthAuthorizeRequest{
		AppID:       config.AppID,
		RedirectURI: redirectURI,
		Scope:       scope,
		State:       state,
	}
	return o.GenerateAuthorizeURL(ctx, req)
}

// GetUserInfoByCode 通过code直接获取用户信息（适用于snsapi_userinfo模式）
// @param ctx 上下文
// @param code 授权码
// @return 用户信息
// @return 授权响应（包含openid等基础信息）
// @return 错误信息
func (o *OAuthClient) GetUserInfoByCode(ctx context.Context, code string) (*OAuthUserInfoResponse, *OAuthAccessTokenResponse, error) {
	return o.CompleteOAuthFlow(ctx, code, "snsapi_userinfo")
}

// GetOpenIDByCode 通过code直接获取openid（适用于snsapi_base模式）
// @param ctx 上下文
// @param code 授权码
// @return openid
// @return 授权响应（包含其他基础信息）
// @return 错误信息
func (o *OAuthClient) GetOpenIDByCode(ctx context.Context, code string) (string, *OAuthAccessTokenResponse, error) {
	accessTokenResp, err := o.GetAccessToken(ctx, code)
	if err != nil {
		return "", nil, fmt.Errorf("获取access_token失败: %v", err)
	}
	return accessTokenResp.OpenID, accessTokenResp, nil
}
