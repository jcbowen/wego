package miniprogram

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/crypto"
	"github.com/jcbowen/wego/logger"
	"github.com/jcbowen/wego/storage"
)

// accessTokenExpireMargin access_token 过期安全余量，避免网络延迟导致临界过期
const accessTokenExpireMargin = 5 * time.Minute

// Client 微信小程序客户端
type Client struct {
	config     *Config
	httpClient core.HTTPClient
	storage    storage.TokenStorage
	logger     logger.LoggerInterface
	req        *core.Request
	tokenMu    sync.RWMutex
}

// NewClientE 创建新的小程序客户端（使用默认文件存储），出错时返回 error 而不是 panic
//
// 参数:
//   - config: 小程序配置信息
//   - opts: 可选参数，支持 debugger.LoggerInterface 与 core.HTTPClient
//
// 返回:
//   - *Client: 小程序客户端实例
//   - error: 文件存储创建失败或配置校验失败时返回错误
func NewClientE(config *Config, opts ...any) (*Client, error) {
	fileStorage, err := storage.NewFileStorage("./runtime/wego_storage")
	if err != nil {
		return nil, fmt.Errorf("文件存储创建失败: %w", err)
	}
	return NewClientWithStorageE(config, fileStorage, opts...)
}

// NewClient 创建新的小程序客户端（使用默认文件存储）
//
// Deprecated: 配置或存储非法时会直接 log.Panic 终止进程，依赖库不应越权决定宿主进程生死。
// 请改用 NewClientE。
func NewClient(config *Config, opts ...any) *Client {
	client, err := NewClientE(config, opts...)
	if err != nil {
		log.Panic("创建小程序客户端失败: " + err.Error())
	}
	return client
}

// NewClientWithStorageE 创建新的小程序客户端（使用自定义存储），出错时返回 error 而不是 panic
//
// 参数:
//   - config: 小程序配置信息
//   - tokenStorage: 自定义存储实例
//   - opts: 可选参数，支持 debugger.LoggerInterface 与 core.HTTPClient
//
// 返回:
//   - *Client: 小程序客户端实例
//   - error: 配置校验失败时返回错误
func NewClientWithStorageE(config *Config, tokenStorage storage.TokenStorage, opts ...any) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("配置不能为空")
	}
	if tokenStorage == nil {
		return nil, fmt.Errorf("存储不能为空")
	}
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置校验失败: %w", err)
	}

	client := &Client{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		storage:    tokenStorage,
		logger:     logger.NewDefaultLoggerInterface(),
	}

	core.ApplyClientOptions(opts, func(msg string) { client.logger.Warn(msg) }, client.SetLogger, client.SetHTTPClient)

	client.req = core.NewRequest(client.httpClient, client.logger)
	return client, nil
}

// NewClientWithStorage 创建新的小程序客户端（使用自定义存储）
//
// Deprecated: 配置非法时会直接 log.Panic 终止进程，请改用 NewClientWithStorageE。
func NewClientWithStorage(config *Config, tokenStorage storage.TokenStorage, opts ...any) *Client {
	client, err := NewClientWithStorageE(config, tokenStorage, opts...)
	if err != nil {
		log.Panic("创建小程序客户端失败: " + err.Error())
	}
	return client
}

// SetLogger 设置自定义日志器
//
// 参数:
//   - log: 自定义日志器，nil 则使用默认日志器
func (c *Client) SetLogger(log debugger.LoggerInterface) {
	if log == nil {
		c.logger = logger.NewDefaultLoggerInterface()
	} else {
		c.logger = logger.NewDebuggerLoggerAdapter(log)
	}
	if c.req != nil {
		c.req = core.NewRequest(c.httpClient, c.logger)
	}
}

// SetHTTPClient 设置自定义 HTTP 客户端
//
// 参数:
//   - client: 自定义 HTTP 客户端
func (c *Client) SetHTTPClient(client core.HTTPClient) {
	c.httpClient = client
	if c.req != nil {
		c.req = core.NewRequest(client, c.logger)
	}
}

// GetConfig 获取配置信息
//
// 返回:
//   - *Config: 小程序配置副本指针
func (c *Client) GetConfig() *Config {
	return c.config
}

// GetLogger 获取日志器
//
// 返回:
//   - logger.LoggerInterface: 日志器实例
func (c *Client) GetLogger() logger.LoggerInterface {
	return c.logger
}

// GetAccessToken 获取小程序 access_token
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - string: access_token
//   - error: 获取失败时返回错误
func (c *Client) GetAccessToken(ctx context.Context) (string, error) {
	c.tokenMu.RLock()
	token, err := c.storage.GetAuthorizerToken(ctx, c.config.AppID)
	if err != nil {
		c.tokenMu.RUnlock()
		return "", err
	}
	if token != nil && c.isTokenValid(token.ExpiresAt) {
		c.tokenMu.RUnlock()
		return token.AuthorizerAccessToken, nil
	}
	c.tokenMu.RUnlock()

	return c.refreshAccessToken(ctx)
}

// refreshAccessToken 刷新小程序 access_token
func (c *Client) refreshAccessToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// 双重检查，避免并发重复刷新
	token, err := c.storage.GetAuthorizerToken(ctx, c.config.AppID)
	if err != nil {
		return "", err
	}
	if token != nil && c.isTokenValid(token.ExpiresAt) {
		return token.AuthorizerAccessToken, nil
	}

	var result struct {
		core.APIResponse
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    AccessTokenURL,
		Query: map[string]string{
			"grant_type": GrantTypeClientCredential,
			"appid":      c.config.AppID,
			"secret":     c.config.AppSecret,
		},
		Result: &result,
	})
	if err != nil {
		return "", err
	}

	if !result.IsSuccess() {
		return "", &result.APIResponse
	}

	newToken := &storage.AuthorizerAccessToken{
		AuthorizerAppID:       c.config.AppID,
		AuthorizerAccessToken: result.AccessToken,
		ExpiresIn:             result.ExpiresIn,
		ExpiresAt:             time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}
	if err := c.storage.SaveAuthorizerToken(ctx, c.config.AppID, newToken); err != nil {
		return "", fmt.Errorf("保存小程序 access_token 失败: %w", err)
	}

	return result.AccessToken, nil
}

// isTokenValid 判断 access_token 是否仍在有效期内（考虑安全余量）
func (c *Client) isTokenValid(expiresAt time.Time) bool {
	return time.Now().Before(expiresAt.Add(-accessTokenExpireMargin))
}

// Code2Session 小程序登录凭证校验
//
// 参数:
//   - ctx: 上下文
//   - code: 前端 wx.login 获取的临时登录凭证
//
// 返回:
//   - *Code2SessionResponse: 登录凭证校验结果
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-login/code2Session.html
func (c *Client) Code2Session(ctx context.Context, code string) (*Code2SessionResponse, error) {
	var result Code2SessionResponse
	err := c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    Code2SessionURL,
		Query: map[string]string{
			"appid":      c.config.AppID,
			"secret":     c.config.AppSecret,
			"js_code":    code,
			"grant_type": GrantTypeAuthorizationCode,
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

// GetPhoneNumber 通过组件 code 换取用户手机号
//
// 参数:
//   - ctx: 上下文
//   - code: 前端 getPhoneNumber 组件返回的动态令牌
//   - openid: 可选参数，用户唯一标识，填入后微信会校验 openid 与 code 的绑定关系
//
// 返回:
//   - *GetPhoneNumberResponse: 手机号换取结果
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-info/phone-number/getPhoneNumber.html
func (c *Client) GetPhoneNumber(ctx context.Context, code string, openid ...string) (*GetPhoneNumberResponse, error) {
	accessToken, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取 access_token 失败: %w", err)
	}

	req := GetPhoneNumberRequest{Code: code}
	if len(openid) > 0 && openid[0] != "" {
		req.OpenID = openid[0]
	}

	var result GetPhoneNumberResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    GetPhoneNumberURL,
		Query: map[string]string{
			"access_token": accessToken,
		},
		Body:   req,
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

// DecryptUserInfo 解密微信小程序用户信息或手机号
//
// 参数:
//   - sessionKey:  登录接口返回的 session_key，Base64 编码
//   - encryptedData: 前端返回的 encryptedData，Base64 编码
//   - iv: 前端返回的 iv，Base64 编码
//   - result:  解密后 JSON 反序列化的目标结构体指针
//
// 返回:
//   - error: 解密失败或水印校验不通过时返回错误
//
// 示例:
//
//	var userInfo miniprogram.UserInfo
//	err := client.DecryptUserInfo(sessionKey, encryptedData, iv, &userInfo)
func (c *Client) DecryptUserInfo(sessionKey, encryptedData, iv string, result interface{}) error {
	// 先解密到临时结构，校验数据水印
	var wrapper struct {
		Watermark Watermark `json:"watermark"`
	}
	if err := crypto.DecryptMiniProgramData(sessionKey, encryptedData, iv, &wrapper); err != nil {
		return err
	}
	if wrapper.Watermark.AppID != "" && wrapper.Watermark.AppID != c.config.AppID {
		return fmt.Errorf("解密数据水印 appid 不匹配，期望 %s，实际 %s", c.config.AppID, wrapper.Watermark.AppID)
	}

	return crypto.DecryptMiniProgramData(sessionKey, encryptedData, iv, result)
}
