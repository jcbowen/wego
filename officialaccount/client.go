package officialaccount

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/storage"
)

// MPClient 微信公众号客户端
type MPClient struct {
	config     *MPConfig
	httpClient core.HTTPClient
	storage    storage.TokenStorage
	logger     core.Logger

	stableTokenClient *StableTokenClient // 稳定版access_token客户端
}

// NewMPClient 创建新的微信公众号客户端（使用默认内存存储）
func NewMPClient(config *MPConfig) *MPClient {
	return NewMPClientWithStorage(config, storage.NewMemoryStorage())
}

// NewMPClientWithStorage 创建新的微信公众号客户端（使用自定义存储）
func NewMPClientWithStorage(config *MPConfig, storage storage.TokenStorage) *MPClient {
	client := &MPClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		storage:    storage,
		logger:     &core.DefaultLogger{},
	}

	client.stableTokenClient = NewStableTokenClient(client)

	return client
}

// GetStableTokenClient 获取稳定版access_token客户端
func (c *MPClient) GetStableTokenClient() *StableTokenClient {
	return c.stableTokenClient
}

// SetLogger 设置自定义日志器
func (c *MPClient) SetLogger(logger core.Logger) {
	c.logger = logger
}

// SetHTTPClient 设置自定义HTTP客户端
func (c *MPClient) SetHTTPClient(client core.HTTPClient) {
	c.httpClient = client
}

// GetAccessToken 获取公众号access_token
func (c *MPClient) GetAccessToken(ctx context.Context) (string, error) {
	// 从存储中获取token
	token, err := c.storage.GetAuthorizerToken(ctx, c.config.AppID)
	if err != nil {
		return "", err
	}

	if token != nil && time.Now().Before(token.ExpiresAt) {
		return token.AuthorizerAccessToken, nil
	}

	// 重新获取access_token
	return c.refreshAccessToken(ctx)
}

// GetStableAccessToken 获取稳定版access_token
func (c *MPClient) GetStableAccessToken(ctx context.Context) (string, error) {
	return c.stableTokenClient.GetStableAccessTokenWithAutoRefresh(ctx)
}

// GetStableAccessTokenInfo 获取稳定版access_token详细信息
func (c *MPClient) GetStableAccessTokenInfo(ctx context.Context, mode StableAccessTokenMode) (*StableAccessTokenInfo, error) {
	return c.stableTokenClient.GetStableAccessToken(ctx, mode)
}

// MakeRequestWithStableToken 使用稳定版access_token发送请求
func (c *MPClient) MakeRequestWithStableToken(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	return c.stableTokenClient.MakeRequestWithStableToken(ctx, method, url, body, result)
}

// refreshAccessToken 刷新access_token
func (c *MPClient) refreshAccessToken(ctx context.Context) (string, error) {
	// 双重检查：再次从存储中获取
	token, err := c.storage.GetAuthorizerToken(ctx, c.config.AppID)
	if err != nil {
		return "", err
	}

	if token != nil && time.Now().Before(token.ExpiresAt) {
		return token.AuthorizerAccessToken, nil
	}

	// 调用微信API获取access_token
	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		c.config.AppID, c.config.AppSecret)

	var result struct {
		core.APIResponse
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	err = c.MakeRequest(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return "", err
	}

	if !result.IsSuccess() {
		return "", &result.APIResponse
	}

	// 更新存储
	newToken := &storage.AuthorizerAccessToken{
		AuthorizerAppID:       c.config.AppID,
		AuthorizerAccessToken: result.AccessToken,
		ExpiresIn:             result.ExpiresIn,
		ExpiresAt:             time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}

	if err := c.storage.SaveAuthorizerToken(ctx, c.config.AppID, newToken); err != nil {
		return "", fmt.Errorf("保存公众号token失败: %v", err)
	}

	return result.AccessToken, nil
}

// MakeRequest 发送HTTP请求的通用方法
func (c *MPClient) MakeRequest(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("序列化请求体失败: %v", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应体失败: %v", err)
	}

	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	return nil
}

// MakeRequestRaw 发送原始HTTP请求，返回响应对象
func (c *MPClient) MakeRequestRaw(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	return resp, nil
}

// GetConfig 获取配置信息
func (c *MPClient) GetConfig() *MPConfig {
	return c.config
}

// GetLogger 获取日志器
func (c *MPClient) GetLogger() core.Logger {
	return c.logger
}

// ClearQuota 清空API调用次数
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/openApi/clear_quota.html
// 接口说明：清空公众号API的调用次数限制，每月共10次清零操作机会
// 注意事项：每个帐号每月共10次清零操作机会，清零生效一次即用掉一次机会
// 请求方式：POST https://api.weixin.qq.com/cgi-bin/clear_quota?access_token=ACCESS_TOKEN
// 请求体：{"appid":"APPID"}
func (c *MPClient) ClearQuota(ctx context.Context) error {
	// 获取access_token
	accessToken, err := c.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("获取access_token失败: %v", err)
	}

	// 构建请求URL
	url := fmt.Sprintf("%s?access_token=%s", APIClearQuotaURL, accessToken)

	// 构建请求体
	request := ClearQuotaRequest{
		AppID: c.config.AppID,
	}

	var response ClearQuotaResponse
	err = c.MakeRequest(ctx, "POST", url, request, &response)
	if err != nil {
		return fmt.Errorf("清空API调用次数失败: %v", err)
	}

	if !response.IsSuccess() {
		return &response.APIResponse
	}

	return nil
}
