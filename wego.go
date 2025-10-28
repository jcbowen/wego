package wego

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jcbowen/jcbaseGo/component/helper"
)

// WeGoConfig 微信开放平台配置结构体
type WeGoConfig struct {
	ComponentAppID     string `json:"component_app_id"`     // 第三方平台appid
	ComponentAppSecret string `json:"component_app_secret"` // 第三方平台appsecret
	ComponentToken     string `json:"component_token"`      // 消息校验Token
	EncodingAESKey     string `json:"encoding_aes_key"`     // 消息加解密Key
	RedirectURI        string `json:"redirect_uri"`         // 授权回调URI
}

// Validate 验证配置的有效性
func (c *WeGoConfig) Validate() error {
	if c.ComponentAppID == "" {
		return fmt.Errorf("ComponentAppID不能为空")
	}
	if c.ComponentAppSecret == "" {
		return fmt.Errorf("ComponentAppSecret不能为空")
	}
	if c.EncodingAESKey != "" && len(c.EncodingAESKey) != 43 {
		return fmt.Errorf("EncodingAESKey必须是43位长度")
	}
	return nil
}

// ComponentAccessToken 第三方平台access_token
type ComponentAccessToken struct {
	AccessToken string    `json:"component_access_token"`
	ExpiresIn   int       `json:"expires_in"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// PreAuthCode 预授权码
type PreAuthCode struct {
	PreAuthCode string    `json:"pre_auth_code"`
	ExpiresIn   int       `json:"expires_in"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// AuthorizerAccessToken 授权方access_token
type AuthorizerAccessToken struct {
	AuthorizerAppID        string    `json:"authorizer_appid"`
	AuthorizerAccessToken  string    `json:"authorizer_access_token"`
	ExpiresIn              int       `json:"expires_in"`
	ExpiresAt              time.Time `json:"expires_at"`
	AuthorizerRefreshToken string    `json:"authorizer_refresh_token"`
}

// HTTPClient HTTP客户端接口
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// WegoClient 微信开放平台客户端
type WegoClient struct {
	config     *WeGoConfig
	httpClient HTTPClient
	storage    TokenStorage
	logger     Logger
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

// NewWegoClient 创建新的微信开放平台客户端（使用默认内存存储）
func NewWegoClient(config *WeGoConfig) *WegoClient {
	return NewWegoClientWithStorage(config, NewMemoryStorage())
}

// NewWegoClientWithStorage 创建新的微信开放平台客户端（使用自定义存储）
func NewWegoClientWithStorage(config *WeGoConfig, storage TokenStorage) *WegoClient {
	client := &WegoClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		storage:    storage,
		logger:     &DefaultLogger{},
	}

	// 设置默认值
	if err := helper.CheckAndSetDefault(client.config); err != nil {
		client.logger.Warnf("设置默认值失败: %v", err)
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
func (c *WegoClient) SetComponentToken(token *ComponentAccessToken) error {
	return c.storage.SaveComponentToken(context.Background(), token)
}

// GetComponentToken 获取组件令牌
func (c *WegoClient) GetComponentToken(ctx context.Context) (*ComponentAccessToken, error) {
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
		newToken := &AuthorizerAccessToken{
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
	token := &AuthorizerAccessToken{
		AuthorizerAppID:        authorizerAppID,
		AuthorizerAccessToken:  accessToken,
		ExpiresIn:              expiresIn,
		ExpiresAt:              time.Now().Add(time.Duration(expiresIn) * time.Second),
		AuthorizerRefreshToken: refreshToken,
	}

	return c.storage.SaveAuthorizerToken(context.Background(), authorizerAppID, token)
}

// GetAuthorizerToken 获取授权方token信息
func (c *WegoClient) GetAuthorizerToken(ctx context.Context, authorizerAppID string) (*AuthorizerAccessToken, error) {
	token, err := c.storage.GetAuthorizerToken(ctx, authorizerAppID)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// APIResponse 微信API通用响应结构
type APIResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// IsSuccess 检查API响应是否成功
func (r *APIResponse) IsSuccess() bool {
	return r.ErrCode == 0
}

// Error 实现error接口
func (r *APIResponse) Error() string {
	return fmt.Sprintf("微信API错误: %d - %s", r.ErrCode, r.ErrMsg)
}

// ParseAPIResponse 解析API响应
func ParseAPIResponse(data []byte) (*APIResponse, error) {
	var resp APIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("解析API响应失败: %v", err)
	}
	return &resp, nil
}
