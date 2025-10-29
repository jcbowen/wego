package openplatform

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



// OpenPlatformClient 微信开放平台客户端
type OpenPlatformClient struct {
	config     *OpenPlatformConfig
	httpClient HTTPClient
	storage    storage.TokenStorage
	logger     Logger
}

// NewOpenPlatformClient 创建新的微信开放平台客户端（使用默认文件存储）
func NewOpenPlatformClient(config *OpenPlatformConfig) *OpenPlatformClient {
	// 使用当前工作目录下的 wego_storage 文件夹作为默认存储路径
	fileStorage, err := storage.NewFileStorage("wego_storage")
	if err != nil {
		// 如果文件存储创建失败，回退到内存存储并输出日志
		logger := &DefaultLogger{}
		logger.Warnf("文件存储创建失败，回退到内存存储: %v", err)
		return NewOpenPlatformClientWithStorage(config, storage.NewMemoryStorage())
	}
	return NewOpenPlatformClientWithStorage(config, fileStorage)
}

// NewOpenPlatformClientWithStorage 创建新的微信开放平台客户端（使用自定义存储）
func NewOpenPlatformClientWithStorage(config *OpenPlatformConfig, storage storage.TokenStorage) *OpenPlatformClient {
	client := &OpenPlatformClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		storage:    storage,
		logger:     &DefaultLogger{},
	}

	return client
}

// SetLogger 设置自定义日志器
func (c *OpenPlatformClient) SetLogger(logger Logger) {
	c.logger = logger
}

// SetHTTPClient 设置自定义HTTP客户端
func (c *OpenPlatformClient) SetHTTPClient(client HTTPClient) {
	c.httpClient = client
}

// SetComponentToken 设置组件令牌
func (c *OpenPlatformClient) SetComponentToken(token *storage.ComponentAccessToken) error {
	return c.storage.SaveComponentToken(context.Background(), token)
}

// GetComponentToken 获取组件令牌
func (c *OpenPlatformClient) GetComponentToken(ctx context.Context) (*storage.ComponentAccessToken, error) {
	return c.storage.GetComponentToken(ctx)
}

// GetAuthorizerAccessToken 获取授权方access_token
func (c *OpenPlatformClient) GetAuthorizerAccessToken(ctx context.Context, authorizerAppID string) (string, error) {
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
func (c *OpenPlatformClient) refreshAuthorizerAccessToken(ctx context.Context, authorizerAppID string) (string, error) {
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
func (c *OpenPlatformClient) SetAuthorizerToken(authorizerAppID, accessToken, refreshToken string, expiresIn int) error {
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
func (c *OpenPlatformClient) RefreshAuthorizerToken(ctx context.Context, authorizerAppID, refreshToken string) (*AuthorizationInfo, error) {
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
func (c *OpenPlatformClient) MakeRequest(ctx context.Context, method, url string, body interface{}, result interface{}) error {
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
func (c *OpenPlatformClient) MakeRequestRaw(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	return resp, nil
}

// GetConfig 获取配置信息
func (c *OpenPlatformClient) GetConfig() *OpenPlatformConfig {
	return c.config
}

// GetLogger 获取日志器
func (c *OpenPlatformClient) GetLogger() Logger {
	return c.logger
}

// SetPreAuthCode 设置预授权码
func (c *OpenPlatformClient) SetPreAuthCode(ctx context.Context, preAuthCode *storage.PreAuthCode) error {
	return c.storage.SavePreAuthCode(ctx, preAuthCode)
}

// GetPreAuthCode 获取预授权码
func (c *OpenPlatformClient) GetPreAuthCode(ctx context.Context) (*storage.PreAuthCode, error) {
	return c.storage.GetPreAuthCode(ctx)
}