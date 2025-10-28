package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/jcbowen/wego/storage"
)

// HTTPClient HTTP客户端接口
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Logger 日志接口
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// DefaultLogger 默认日志实现
type DefaultLogger struct{}

func (l *DefaultLogger) Debugf(format string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+format+"\n", args...)
}

func (l *DefaultLogger) Infof(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

func (l *DefaultLogger) Warnf(format string, args ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}

func (l *DefaultLogger) Errorf(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

// APIResponse 微信API通用响应结构
type APIResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// Error 实现error接口
func (r *APIResponse) Error() string {
	return fmt.Sprintf("微信API错误[%d]: %s", r.ErrCode, r.ErrMsg)
}

// IsSuccess 检查API响应是否成功
func (r *APIResponse) IsSuccess() bool {
	return r.ErrCode == 0
}

// AuthorizationInfo 授权信息
type AuthorizationInfo struct {
	AuthorizerAppID        string `json:"authorizer_appid"`
	AuthorizerAccessToken  string `json:"authorizer_access_token"`
	ExpiresIn              int    `json:"expires_in"`
	AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
}

// WegoClient 微信开放平台客户端
type WegoClient struct {
	config     *WeGoConfig
	httpClient HTTPClient
	storage    storage.TokenStorage
	logger     Logger
}

// NewWegoClient 创建新的微信开放平台客户端（使用默认内存存储）
func NewWegoClient(config *WeGoConfig) *WegoClient {
	return NewWegoClientWithStorage(config, storage.NewMemoryStorage())
}

// NewWegoClientWithStorage 创建新的微信开放平台客户端（使用自定义存储）
func NewWegoClientWithStorage(config *WeGoConfig, storage storage.TokenStorage) *WegoClient {
	client := &WegoClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		storage:    storage,
		logger:     &DefaultLogger{},
	}

	return client
}

// SetLogger 设置自定义日志器
func (c *WegoClient) SetLogger(logger Logger) {
	c.logger = logger
}

// SetHTTPClient 设置自定义HTTP客户端
func (c *WegoClient) SetHTTPClient(client HTTPClient) {
	c.httpClient = client
}

// SetComponentToken 设置组件令牌
func (c *WegoClient) SetComponentToken(token *storage.ComponentAccessToken) error {
	return c.storage.SaveComponentToken(context.Background(), token)
}

// GetComponentToken 获取组件令牌
func (c *WegoClient) GetComponentToken(ctx context.Context) (*storage.ComponentAccessToken, error) {
	return c.storage.GetComponentToken(ctx)
}

// GetAuthorizerAccessToken 获取授权方access_token
func (c *WegoClient) GetAuthorizerAccessToken(ctx context.Context, authorizerAppID string) (string, error) {
	// 从存储中获取授权方token
	token, err := c.storage.GetAuthorizerToken(ctx, authorizerAppID)
	if err != nil {
		return "", err
	}

	if token != nil && time.Now().Before(token.ExpiresAt) {
		return token.AuthorizerAccessToken, nil
	}

	// 重新获取授权方access_token
	return c.refreshAuthorizerAccessToken(ctx, authorizerAppID)
}

// refreshAuthorizerAccessToken 刷新授权方access_token
func (c *WegoClient) refreshAuthorizerAccessToken(ctx context.Context, authorizerAppID string) (string, error) {
	// 双重检查：再次从存储中获取
	token, err := c.storage.GetAuthorizerToken(ctx, authorizerAppID)
	if err != nil {
		return "", err
	}

	if token != nil && time.Now().Before(token.ExpiresAt) {
		return token.AuthorizerAccessToken, nil
	}

	// 调用微信API刷新授权方access_token
	if token != nil && token.AuthorizerRefreshToken != "" {
		// 使用refresh_token刷新access_token
		result, err := c.RefreshAuthorizerToken(ctx, authorizerAppID, token.AuthorizerRefreshToken)
		if err != nil {
			return "", err
		}

		// 更新存储
	newToken := &storage.AuthorizerAccessToken{
		AuthorizerAppID:        authorizerAppID,
		AuthorizerAccessToken:  result.AuthorizerAccessToken,
		ExpiresIn:              result.ExpiresIn,
		ExpiresAt:              time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
		AuthorizerRefreshToken: result.AuthorizerRefreshToken,
	}

	if err := c.storage.SaveAuthorizerToken(ctx, authorizerAppID, newToken); err != nil {
		return "", fmt.Errorf("保存授权方token失败: %v", err)
	}

	return result.AuthorizerAccessToken, nil
	}

	return "", fmt.Errorf("无法获取授权方access_token：缺少refresh_token")
}

// SetAuthorizerToken 设置授权方token信息
func (c *WegoClient) SetAuthorizerToken(authorizerAppID, accessToken, refreshToken string, expiresIn int) error {
	token := &storage.AuthorizerAccessToken{
		AuthorizerAppID:        authorizerAppID,
		AuthorizerAccessToken:  accessToken,
		ExpiresIn:              expiresIn,
		ExpiresAt:              time.Now().Add(time.Duration(expiresIn) * time.Second),
		AuthorizerRefreshToken: refreshToken,
	}

	return c.storage.SaveAuthorizerToken(context.Background(), authorizerAppID, token)
}

// RefreshAuthorizerToken 刷新授权方access_token
func (c *WegoClient) RefreshAuthorizerToken(ctx context.Context, authorizerAppID, refreshToken string) (*AuthorizationInfo, error) {
	// 获取组件access_token
	componentToken, err := c.GetComponentToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取组件token失败: %v", err)
	}

	// 构建请求参数
	request := map[string]string{
		"component_appid":           c.config.ComponentAppID,
		"authorizer_appid":          authorizerAppID,
		"authorizer_refresh_token":   refreshToken,
	}

	// 调用微信API刷新授权方token
	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token?component_access_token=%s", 
		url.QueryEscape(componentToken.AccessToken))

	var result struct {
		APIResponse
		AuthorizationInfo
	}

	err = c.MakeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result.AuthorizationInfo, nil
}

// MakeRequest 发送HTTP请求的通用方法
func (c *WegoClient) MakeRequest(ctx context.Context, method, url string, body interface{}, result interface{}) error {
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
func (c *WegoClient) MakeRequestRaw(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	return resp, nil
}

// GetConfig 获取配置信息
func (c *WegoClient) GetConfig() *WeGoConfig {
	return c.config
}

// GetLogger 获取日志器
func (c *WegoClient) GetLogger() Logger {
	return c.logger
}

// SetPreAuthCode 设置预授权码
func (c *WegoClient) SetPreAuthCode(ctx context.Context, preAuthCode *storage.PreAuthCode) error {
	return c.storage.SavePreAuthCode(ctx, preAuthCode)
}

// GetPreAuthCode 获取预授权码
func (c *WegoClient) GetPreAuthCode(ctx context.Context) (*storage.PreAuthCode, error) {
	return c.storage.GetPreAuthCode(ctx)
}