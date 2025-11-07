package officialaccount

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/storage"
)

// Client 微信公众号客户端
type Client struct {
	config     *Config
	httpClient core.HTTPClient
	storage    storage.TokenStorage
	logger     debugger.LoggerInterface
	req        *core.Request

	stableTokenClient *StableTokenClient // 稳定版access_token客户端
}

// NewClient 创建新的微信公众号客户端（使用默认文件存储）
// @param config *Config 公众号配置信息
// @param opts ...any 可选参数，支持以下类型：
//   - debugger.LoggerInterface: 自定义日志器
//   - core.HTTPClient: 自定义HTTP客户端
//
// @return *Client 公众号客户端实例
func NewClient(config *Config, opts ...any) *Client {
	// 使用当前工作目录下的 wego_storage 文件夹作为默认存储路径
	fileStorage, err := storage.NewFileStorage("./runtime/wego_storage")
	if err != nil {
		// 如果文件存储创建失败，回退到内存存储并输出日志
		logger := &debugger.DefaultLogger{}
		logger.Warn("文件存储创建失败，回退到内存存储: " + err.Error())
		return NewMPClientWithStorage(config, storage.NewMemoryStorage(), opts...)
	}
	return NewMPClientWithStorage(config, fileStorage, opts...)
}

// NewMPClientWithStorage 创建新的微信公众号客户端（使用自定义存储）
// @param config *Config 公众号配置信息
// @param storage storage.TokenStorage 自定义存储实例
// @param opts ...any 可选参数，支持以下类型：
//   - debugger.LoggerInterface: 自定义日志器
//   - core.HTTPClient: 自定义HTTP客户端
//
// @return *Client 公众号客户端实例
func NewMPClientWithStorage(config *Config, storage storage.TokenStorage, opts ...any) *Client {
	client := &Client{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		storage:    storage,
		logger:     &debugger.DefaultLogger{},
	}

	// 遍历所有可选参数，根据类型进行相应设置
	if len(opts) > 0 {
		for _, option := range opts {
			switch v := option.(type) {
			case debugger.LoggerInterface:
				// 设置自定义日志器
				client.SetLogger(v)
			case core.HTTPClient:
				// 设置自定义HTTP客户端
				client.SetHTTPClient(v)
			default:
				// 记录未知类型的可选参数
				client.logger.Warn(fmt.Sprintf("未知的可选参数类型: %T", v))
			}
		}
	}

	client.req = core.NewRequest(client.httpClient, client.logger)
	client.stableTokenClient = NewStableTokenClient(client)

	return client
}

// GetStableTokenClient 获取稳定版access_token客户端
func (c *Client) GetStableTokenClient() *StableTokenClient {
	return c.stableTokenClient
}

// SetLogger 设置自定义日志器
func (c *Client) SetLogger(logger debugger.LoggerInterface) {
	c.logger = logger
	// 同时更新请求对象中的日志器
	if c.req != nil {
		c.req = core.NewRequest(c.httpClient, logger)
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

	err = core.NewRequest(c.httpClient, c.logger).Make(ctx, "GET", apiURL, nil, &result)
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

// GetConfig 获取配置信息
func (c *Client) GetConfig() *Config {
	return c.config
}

// GetLogger 获取日志器
func (c *Client) GetLogger() debugger.LoggerInterface {
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
	url := fmt.Sprintf("%s?access_token=%s", APIClearQuotaURL, accessToken)

	// 构建请求体
	request := ClearQuotaRequest{
		AppID: c.config.AppID,
	}

	var response ClearQuotaResponse
	err = core.NewRequest(c.httpClient, c.logger).Make(ctx, "POST", url, request, &response)
	if err != nil {
		return fmt.Errorf("清空API调用次数失败: %v", err)
	}

	if !response.IsSuccess() {
		return &response.APIResponse
	}

	return nil
}
