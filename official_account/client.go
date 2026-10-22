package official_account

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/logger"
	"github.com/jcbowen/wego/storage"
)

// accessTokenExpireMargin access_token 过期安全余量，避免网络延迟导致临界过期
const accessTokenExpireMargin = 5 * time.Minute

// Client 微信公众号客户端
type Client struct {
	config     *Config
	httpClient core.HTTPClient
	storage    storage.TokenStorage
	logger     logger.LoggerInterface
	req        *core.Request
	tokenMu    sync.RWMutex // 保护 access_token 的获取与刷新，避免并发重复刷新

	stableTokenClient *StableTokenClient // 稳定版access_token客户端
}

// NewClientE 创建新的微信公众号客户端（使用默认文件存储），出错时返回 error 而不是 panic
// @param config *Config 公众号配置信息
// @param opts ...any 可选参数，支持以下类型：
//   - debugger.LoggerInterface: 自定义日志器
//   - core.HTTPClient: 自定义HTTP客户端
//
// @return *Client 公众号客户端实例
// @return error 文件存储创建失败时返回错误
func NewClientE(config *Config, opts ...any) (*Client, error) {
	// 使用当前工作目录下的 wego_storage 文件夹作为默认存储路径
	fileStorage, err := storage.NewFileStorage("./runtime/wego_storage")
	if err != nil {
		return nil, fmt.Errorf("文件存储创建失败: %w", err)
	}
	return NewMPClientWithStorageE(config, fileStorage, opts...)
}

// NewClient 创建新的微信公众号客户端（使用默认文件存储）
//
// Deprecated: 配置或存储非法时会直接 log.Panic 终止进程，依赖库不应越权决定宿主进程生死。
// 请改用 NewClientE。
func NewClient(config *Config, opts ...any) *Client {
	client, err := NewClientE(config, opts...)
	if err != nil {
		log.Panic("创建微信公众号客户端失败: " + err.Error())
	}
	return client
}

// NewMPClientWithStorageE 创建新的微信公众号客户端（使用自定义存储），出错时返回 error 而不是 panic
// @param config *Config 公众号配置信息
// @param tokenStorage storage.TokenStorage 自定义存储实例
// @param opts ...any 可选参数，支持以下类型：
//   - debugger.LoggerInterface: 自定义日志器
//   - core.HTTPClient: 自定义HTTP客户端
//
// @return *Client 公众号客户端实例
// @return error 配置校验失败时返回错误
func NewMPClientWithStorageE(config *Config, tokenStorage storage.TokenStorage, opts ...any) (*Client, error) {
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

	// 遍历所有可选参数，根据类型进行相应设置
	core.ApplyClientOptions(opts, func(msg string) { client.logger.Warn(msg) }, client.SetLogger, client.SetHTTPClient)

	client.req = core.NewRequest(client.httpClient, client.logger)
	client.stableTokenClient = NewStableTokenClient(client)

	return client, nil
}

// NewMPClientWithStorage 创建新的微信公众号客户端（使用自定义存储）
//
// Deprecated: 配置非法时会直接 log.Panic 终止进程，请改用 NewMPClientWithStorageE。
func NewMPClientWithStorage(config *Config, tokenStorage storage.TokenStorage, opts ...any) *Client {
	client, err := NewMPClientWithStorageE(config, tokenStorage, opts...)
	if err != nil {
		log.Panic("创建微信公众号客户端失败: " + err.Error())
	}
	return client
}

// GetStableTokenClient 获取稳定版access_token客户端
func (c *Client) GetStableTokenClient() *StableTokenClient {
	return c.stableTokenClient
}

// SetLogger 设置自定义日志器
func (c *Client) SetLogger(log debugger.LoggerInterface) {
	// 如果传入的logger为nil，则使用默认logger
	if log == nil {
		c.logger = logger.NewDefaultLoggerInterface()
	} else {
		// 使用适配器将debugger logger转换为wego logger
		c.logger = logger.NewDebuggerLoggerAdapter(log)
	}
	// 同时更新请求对象中的日志器
	if c.req != nil {
		c.req = core.NewRequest(c.httpClient, c.logger)
	}
}

// SetHTTPClient 设置自定义HTTP客户端
func (c *Client) SetHTTPClient(client core.HTTPClient) {
	c.httpClient = client
	// 同时更新请求对象中的HTTP客户端
	if c.req != nil {
		c.req = core.NewRequest(client, c.logger)
	}
}

// GetAccessToken 获取公众号access_token
func (c *Client) GetAccessToken(ctx context.Context) (string, error) {
	// 读锁查缓存，命中直接返回
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

	// 未命中则加锁刷新
	return c.refreshAccessToken(ctx)
}

// GetStableAccessToken 获取稳定版access_token
func (c *Client) GetStableAccessToken(ctx context.Context) (string, error) {
	return c.stableTokenClient.GetStableAccessTokenWithAutoRefresh(ctx)
}

// GetStableAccessTokenInfo 获取稳定版access_token详细信息
func (c *Client) GetStableAccessTokenInfo(ctx context.Context, mode StableAccessTokenMode) (*StableAccessTokenInfo, error) {
	return c.stableTokenClient.GetStableAccessToken(ctx, mode)
}

// MakeRequestWithStableToken 使用稳定版access_token发送请求
func (c *Client) MakeRequestWithStableToken(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	return c.stableTokenClient.MakeRequestWithStableToken(ctx, method, url, body, result)
}

// refreshAccessToken 刷新access_token
func (c *Client) refreshAccessToken(ctx context.Context) (string, error) {
	// 互斥锁 + 双重检查，避免并发场景下重复刷新（微信侧并发刷新会互相覆盖并返回 40001）
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

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

	// 参数走 Query String，与官方 /cgi-bin/token 的 GET 方式一致
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    URLAccessToken,
		Query: map[string]string{
			"grant_type": core.GrantTypeClientCredential,
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

// isTokenValid 判断 access_token 是否仍在有效期内（考虑安全余量）
func (c *Client) isTokenValid(expiresAt time.Time) bool {
	return time.Now().Before(expiresAt.Add(-accessTokenExpireMargin))
}

// GetConfig 获取配置信息
func (c *Client) GetConfig() *Config {
	return c.config
}

// GetLogger 获取日志器
func (c *Client) GetLogger() logger.LoggerInterface {
	return c.logger
}

// ClearQuota 清空API调用次数
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/openApi/clear_quota.html
// 接口说明：清空公众号API的调用次数限制，每月共10次清零操作机会
// 注意事项：每个帐号每月共10次清零操作机会，清零生效一次即用掉一次机会
// 请求方式：POST https://api.weixin.qq.com/cgi-bin/clear_quota?access_token=ACCESS_TOKEN
// 请求体：{"appid":"APPID"}
func (c *Client) ClearQuota(ctx context.Context) error {
	// 获取access_token
	accessToken, err := c.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("获取access_token失败: %v", err)
	}

	// 构建请求URL
	url := fmt.Sprintf("%s?access_token=%s", URLClearQuota, accessToken)

	// 构建请求体
	request := ClearQuotaRequest{
		AppID: c.config.AppID,
	}

	var response ClearQuotaResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    url,
		Body:   request,
		Result: &response,
	})
	if err != nil {
		return fmt.Errorf("清空API调用次数失败: %v", err)
	}

	if !response.IsSuccess() {
		return &response.APIResponse
	}

	return nil
}
