package officialaccount

import (
	"context"
	"fmt"
	"time"
)

// StableTokenClient 稳定版access_token客户端
type StableTokenClient struct {
	client *MPClient
}

// NewStableTokenClient 创建稳定版access_token客户端
func NewStableTokenClient(client *MPClient) *StableTokenClient {
	return &StableTokenClient{
		client: client,
	}
}

// GetStableAccessToken 获取稳定版access_token
func (c *StableTokenClient) GetStableAccessToken(ctx context.Context, mode StableAccessTokenMode) (*StableAccessTokenInfo, error) {
	return c.getStableAccessTokenWithMode(ctx, mode)
}

// GetStableAccessTokenNormal 以普通模式获取稳定版access_token
func (c *StableTokenClient) GetStableAccessTokenNormal(ctx context.Context) (*StableAccessTokenInfo, error) {
	return c.getStableAccessTokenWithMode(ctx, StableAccessTokenModeNormal)
}

// GetStableAccessTokenForceRefresh 以强制刷新模式获取稳定版access_token
func (c *StableTokenClient) GetStableAccessTokenForceRefresh(ctx context.Context) (*StableAccessTokenInfo, error) {
	return c.getStableAccessTokenWithMode(ctx, StableAccessTokenModeForceRefresh)
}

// getStableAccessTokenWithMode 根据模式获取稳定版access_token
func (c *StableTokenClient) getStableAccessTokenWithMode(ctx context.Context, mode StableAccessTokenMode) (*StableAccessTokenInfo, error) {
	// 检查是否已有有效的token
	if mode == StableAccessTokenModeNormal {
		if token, err := c.getValidStableAccessToken(ctx); err == nil && token != nil {
			return token, nil
		}
	}

	// 构建请求参数
	request := StableAccessTokenRequest{
		GrantType:    "client_credential",
		AppID:        c.client.config.AppID,
		Secret:       c.client.config.AppSecret,
		ForceRefresh: mode == StableAccessTokenModeForceRefresh,
	}

	// 调用API获取稳定版access_token
	var result StableAccessTokenResponse
	err := c.client.MakeRequest(ctx, "POST", APIStableAccessTokenURL, request, &result)
	if err != nil {
		return nil, fmt.Errorf("获取稳定版access_token失败: %v", err)
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	// 构建token信息
	tokenInfo := &StableAccessTokenInfo{
		AccessToken: result.AccessToken,
		ExpiresIn:   result.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
		Mode:        mode,
	}

	// 保存到存储（如果有存储接口）
	if c.client.storage != nil {
		// 这里需要将StableAccessTokenInfo转换为存储格式
		// 由于存储接口目前只支持AuthorizerAccessToken，我们需要扩展存储接口
		// 暂时先不保存，后续可以扩展存储接口
	}

	return tokenInfo, nil
}

// getValidStableAccessToken 获取有效的稳定版access_token
func (c *StableTokenClient) getValidStableAccessToken(ctx context.Context) (*StableAccessTokenInfo, error) {
	// 这里需要从存储中获取token
	// 由于存储接口目前不支持稳定版token，暂时返回nil
	// 后续可以扩展存储接口来支持
	return nil, nil
}

// IsStableAccessTokenValid 检查稳定版access_token是否有效
func (c *StableTokenClient) IsStableAccessTokenValid(ctx context.Context, token string) (bool, error) {
	// 简单的检查：如果token不为空且长度合理，则认为有效
	// 实际应该调用微信API验证token有效性
	if token == "" {
		return false, nil
	}
	
	// 这里可以添加更复杂的验证逻辑
	// 例如检查token格式、调用微信API验证等
	return true, nil
}

// RefreshStableAccessTokenIfNeeded 如果需要则刷新稳定版access_token
func (c *StableTokenClient) RefreshStableAccessTokenIfNeeded(ctx context.Context) (string, error) {
	// 检查当前token是否即将过期（提前5分钟）
	if token, err := c.getValidStableAccessToken(ctx); err == nil && token != nil {
		if time.Until(token.ExpiresAt) > 5*time.Minute {
			return token.AccessToken, nil
		}
	}

	// 获取新的稳定版access_token
	newToken, err := c.GetStableAccessTokenNormal(ctx)
	if err != nil {
		return "", err
	}

	return newToken.AccessToken, nil
}

// GetStableAccessTokenWithAutoRefresh 自动刷新获取稳定版access_token
func (c *StableTokenClient) GetStableAccessTokenWithAutoRefresh(ctx context.Context) (string, error) {
	return c.RefreshStableAccessTokenIfNeeded(ctx)
}

// MakeRequestWithStableToken 使用稳定版access_token发送请求
func (c *StableTokenClient) MakeRequestWithStableToken(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	// 获取稳定版access_token
	token, err := c.RefreshStableAccessTokenIfNeeded(ctx)
	if err != nil {
		return fmt.Errorf("获取稳定版access_token失败: %v", err)
	}

	// 构建带token的URL
	fullURL := fmt.Sprintf("%s?access_token=%s", url, token)

	// 发送请求
	return c.client.MakeRequest(ctx, method, fullURL, body, result)
}