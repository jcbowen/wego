package openplatform

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/crypto"
	"github.com/jcbowen/wego/logger"
	"github.com/jcbowen/wego/storage"
)

// tokenExpireMargin token 过期安全余量，避免网络延迟导致临界过期（对齐 miniprogram 包写法）
const tokenExpireMargin = 5 * time.Minute

// Client API客户端
type Client struct {
	config       *Config
	httpClient   core.HTTPClient
	storage      storage.TokenStorage
	logger       logger.LoggerInterface
	eventHandler EventHandler // 事件处理器
	req          *core.Request

	// cryptMu 保护 config.EncodingAESKey 的读写（encoding_aes_key_changed 事件回调
	// 会更新密钥，而消息解密并发进行，直接读写存在数据竞争）
	cryptMu sync.RWMutex

	// compTokenMu 串行化 component_access_token 的刷新，避免并发重复请求微信API
	compTokenMu sync.Mutex

	// authTokMu 串行化授权方 access_token 的刷新，避免并发重复请求微信API
	authTokMu sync.Mutex

	eventErrorCallback EventErrorCallback // 事件处理异常回调（可选，用于上层落盘排障）

	jsapiTicketMu  sync.Mutex
	jsapiTicketMap map[string]*jsapiTicketInfo // JSAPI Ticket 缓存（按授权方appid区分）
}

// NewClientE 创建新的API客户端（使用默认文件存储），出错时返回 error 而不是 panic
// @param config *Config 开放平台配置信息
// @param opt ...any 可选参数，支持以下类型：
//   - debugger.LoggerInterface: 自定义日志器
//   - HTTPClient: 自定义HTTP客户端
//   - EventHandler: 自定义事件处理器
//   - EventErrorCallback: 事件处理异常回调（可选，用于上层落盘排障）
//
// @return *Client API客户端实例
// @return error 文件存储创建失败或配置校验失败时返回错误
func NewClientE(config *Config, opt ...any) (*Client, error) {
	// 使用当前工作目录下的 ./runtime/wego_storage 文件夹作为默认存储路径
	fileStorage, err := storage.NewFileStorage("./runtime/wego_storage")
	if err != nil {
		return nil, fmt.Errorf("文件存储创建失败: %w", err)
	}
	return NewClientWithStorageE(config, fileStorage, opt...)
}

// NewClient 创建新的API客户端（使用默认文件存储）
//
// Deprecated: 配置或存储非法时会直接 log.Panic 终止进程，依赖库不应越权决定宿主进程生死。
// 请改用 NewClientE。
func NewClient(config *Config, opt ...any) *Client {
	client, err := NewClientE(config, opt...)
	if err != nil {
		log.Panic("创建开放平台客户端失败: " + err.Error())
	}
	return client
}

// NewClientWithStorageE 创建新的API客户端（使用自定义存储），出错时返回 error 而不是 panic
func NewClientWithStorageE(config *Config, tokenStorage storage.TokenStorage, opt ...any) (*Client, error) {
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
	core.ApplyClientOptions(opt, func(msg string) { client.logger.Warn(msg) }, func(l debugger.LoggerInterface) {
		client.SetLogger(l)
	}, client.SetHTTPClient,
		func(option any) (handled bool) {
			if handler, ok := option.(EventHandler); ok {
				client.SetEventHandler(handler)
				return true
			}
			if cb, ok := option.(EventErrorCallback); ok {
				client.SetEventErrorCallback(cb)
				return true
			}
			return false
		})

	client.req = core.NewRequest(client.httpClient, client.logger)

	return client, nil
}

// NewClientWithStorage 创建新的API客户端（使用自定义存储）
//
// Deprecated: 配置非法时会直接 log.Panic 终止进程，请改用 NewClientWithStorageE。
func NewClientWithStorage(config *Config, tokenStorage storage.TokenStorage, opt ...any) *Client {
	client, err := NewClientWithStorageE(config, tokenStorage, opt...)
	if err != nil {
		log.Panic("创建开放平台客户端失败: " + err.Error())
	}
	return client
}

// SetLogger 设置日志器
func (c *Client) SetLogger(log any) {
	if log == nil {
		c.logger = logger.NewDefaultLoggerInterface()
	} else {
		switch v := log.(type) {
		case debugger.LoggerInterface:
			// 如果是debugger logger，创建适配器
			c.logger = logger.NewDebuggerLoggerAdapter(v)
		default:
			// 未知类型，使用默认logger
			c.logger = logger.NewDefaultLoggerInterface()
		}
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

// SetEventHandler 设置事件处理器
func (c *Client) SetEventHandler(handler EventHandler) {
	c.eventHandler = handler
}

// GetEventHandler 获取事件处理器
func (c *Client) GetEventHandler() EventHandler {
	if c.eventHandler == nil {
		return &DefaultEventHandler{}
	}
	return c.eventHandler
}

// SetEventErrorCallback 设置事件处理异常回调（可选）
//
// 参数:
//   - cb EventErrorCallback 事件处理异常回调；事件解析/解密/验签失败或处理器返回错误时被同步调用，
//     不影响微信官方要求的「返回 success」行为
func (c *Client) SetEventErrorCallback(cb EventErrorCallback) {
	c.eventErrorCallback = cb
}

// notifyEventError 通过回调通知事件处理异常（未设置回调时为空操作；回调自身 panic 时恢复并记日志，防影响主流程）
//
// 参数:
//   - source string 失败环节标识（取值见 EventErrorCallback 注释）
//   - err error 原始错误
func (c *Client) notifyEventError(source string, err error) {
	if c.eventErrorCallback == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(fmt.Sprintf("事件错误回调自身发生panic: %v", r))
		}
	}()
	c.eventErrorCallback(source, err)
}

// GetConfig 获取配置信息
func (c *Client) GetConfig() *Config {
	return c.config
}

// GetLogger 获取日志器
func (c *Client) GetLogger() logger.LoggerInterface {
	return c.logger
}

// SetComponentToken 设置开放平台令牌
func (c *Client) SetComponentToken(token *storage.ComponentAccessToken) error {
	return c.storage.SaveComponentToken(context.Background(), token)
}

// GetComponentToken 获取开放平台令牌
// @param ctx context.Context 上下文
// @return *storage.ComponentAccessToken 组件访问令牌，包含令牌内容和有效期信息
// @return error 错误信息，如果令牌获取失败或自动刷新失败则返回错误
func (c *Client) GetComponentToken(ctx context.Context) (*storage.ComponentAccessToken, error) {
	// 从存储获取令牌
	token, err := c.storage.GetComponentToken(ctx)
	if err != nil {
		return nil, err
	}

	// 如果令牌存在且仍在有效期内（含安全余量），直接返回
	if token != nil && c.isTokenValid(token.ExpiresAt) {
		return token, nil
	}

	// 令牌不存在或临近过期，加锁刷新（内部双重检查，避免并发重复请求微信API）
	c.compTokenMu.Lock()
	defer c.compTokenMu.Unlock()

	token, err = c.storage.GetComponentToken(ctx)
	if err != nil {
		return nil, err
	}
	if token != nil && c.isTokenValid(token.ExpiresAt) {
		return token, nil
	}

	c.logger.Info("ComponentAccessToken不存在或已过期，自动获取新令牌...")
	newToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("自动获取ComponentAccessToken失败: %w", err)
	}

	return newToken, nil
}

// SetPreAuthCode 设置预授权码
func (c *Client) SetPreAuthCode(ctx context.Context, preAuthCode *storage.PreAuthCode) error {
	return c.storage.SavePreAuthCode(ctx, preAuthCode)
}

// GetPreAuthCode 获取预授权码（优先复用缓存，临近过期自动重新获取）
//
// 复用策略：先读存储，取到且未过期（留 60 秒安全余量）则直接返回，
// 避免每次都调微信API；否则调用微信API获取新预授权码并写回存储
func (c *Client) GetPreAuthCode(ctx context.Context) (*storage.PreAuthCode, error) {
	// 先读缓存：取到且未过期（留 60 秒余量）则直接复用
	code, err := c.storage.GetPreAuthCode(ctx)
	if err == nil && code != nil && code.PreAuthCode != "" &&
		time.Now().Before(code.ExpiresAt.Add(-preAuthCodeExpireMargin)) {
		return code, nil
	}

	// 缓存缺失或临近过期，从微信API获取（内部获取成功后会写回存储）
	resp, err := c.GetPreAuthCodeFromAPI(ctx)
	if err != nil {
		return nil, err
	}

	return &storage.PreAuthCode{
		PreAuthCode: resp.PreAuthCode,
		ExpiresIn:   resp.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	}, nil
}

// SetAuthorizerToken 设置授权方token信息
func (c *Client) SetAuthorizerToken(authorizerAppID, accessToken, refreshToken string, expiresIn int) error {
	token := &storage.AuthorizerAccessToken{
		AuthorizerAppID:        authorizerAppID,
		AuthorizerAccessToken:  accessToken,
		ExpiresIn:              expiresIn,
		ExpiresAt:              time.Now().Add(time.Duration(expiresIn) * time.Second),
		AuthorizerRefreshToken: refreshToken,
	}

	return c.storage.SaveAuthorizerToken(context.Background(), authorizerAppID, token)
}

// GetAuthorizerAccessToken 获取授权方access_token
func (c *Client) GetAuthorizerAccessToken(ctx context.Context, authorizerAppID string) (string, error) {
	// 从存储中获取授权方token
	token, err := c.storage.GetAuthorizerToken(ctx, authorizerAppID)
	if err != nil {
		return "", err
	}

	// 仍在有效期内（含安全余量）则直接返回
	if token != nil && c.isTokenValid(token.ExpiresAt) {
		return token.AuthorizerAccessToken, nil
	}

	// 重新获取授权方access_token（内部加锁，避免并发重复刷新）
	return c.refreshAuthorizerAccessToken(ctx, authorizerAppID)
}

// isTokenValid 判断 token 是否仍在有效期内（考虑安全余量，对齐 miniprogram 包写法）
func (c *Client) isTokenValid(expiresAt time.Time) bool {
	return time.Now().Before(expiresAt.Add(-tokenExpireMargin))
}

// GetComponentVerifyTicket 获取验证票据
// @param ctx context.Context 上下文
// @return *storage.ComponentVerifyTicket 验证票据结构，包含票据内容和有效期信息
// @return error 错误信息
func (c *Client) GetComponentVerifyTicket(ctx context.Context) (*storage.ComponentVerifyTicket, error) {
	return c.storage.GetComponentVerifyTicket(ctx)
}

// SaveComponentVerifyTicket 保存验证票据
// @param ctx context.Context 上下文
// @param ticket string 票据内容
// @return error 错误信息
func (c *Client) SaveComponentVerifyTicket(ctx context.Context, ticket string) error {
	return c.storage.SaveComponentVerifyTicket(ctx, ticket)
}

// refreshAuthorizerAccessToken 刷新授权方access_token
func (c *Client) refreshAuthorizerAccessToken(ctx context.Context, authorizerAppID string) (string, error) {
	// 加锁串行化刷新，避免并发重复请求微信API
	c.authTokMu.Lock()
	defer c.authTokMu.Unlock()

	// 双重检查：加锁后再次从存储中获取
	token, err := c.storage.GetAuthorizerToken(ctx, authorizerAppID)
	if err != nil {
		return "", err
	}

	if token != nil && c.isTokenValid(token.ExpiresAt) {
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

// ComponentTokenRequest 获取component_access_token请求参数
type ComponentTokenRequest struct {
	ComponentAppID        string `json:"component_appid"`
	ComponentAppSecret    string `json:"component_appsecret"`
	ComponentVerifyTicket string `json:"component_verify_ticket"`
}

// PreAuthCodeRequest 获取预授权码请求参数
type PreAuthCodeRequest struct {
	ComponentAppID string `json:"component_appid"`
}

// PreAuthCodeResponse 预授权码响应
type PreAuthCodeResponse struct {
	core.APIResponse
	PreAuthCode string `json:"pre_auth_code"`
	ExpiresIn   int    `json:"expires_in"`
}

// QueryAuthRequest 使用授权码换取授权信息请求参数
type QueryAuthRequest struct {
	ComponentAppID    string `json:"component_appid"`
	AuthorizationCode string `json:"authorization_code"`
}

// QueryAuthResponse 授权信息响应
type QueryAuthResponse struct {
	core.APIResponse
	AuthorizationInfo AuthorizationInfo `json:"authorization_info"`
}

// AuthorizerTokenRequest 刷新授权方token请求参数
type AuthorizerTokenRequest struct {
	ComponentAppID         string `json:"component_appid"`
	AuthorizerAppID        string `json:"authorizer_appid"`
	AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
}

// AuthorizerTokenResponse 授权方token响应
type AuthorizerTokenResponse struct {
	core.APIResponse
	AuthorizerAccessToken  string `json:"authorizer_access_token"`
	ExpiresIn              int    `json:"expires_in"`
	AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
}

// GetAuthorizerInfoRequest 获取授权方信息请求参数
type GetAuthorizerInfoRequest struct {
	ComponentAppID  string `json:"component_appid"`
	AuthorizerAppID string `json:"authorizer_appid"`
}

// GetAuthorizerInfoResponse 授权方信息响应
type GetAuthorizerInfoResponse struct {
	core.APIResponse
	AuthorizerInfo AuthorizerInfo `json:"authorizer_info"`
}

// GetAuthorizerListRequest 获取授权方列表请求参数
type GetAuthorizerListRequest struct {
	ComponentAppID string `json:"component_appid"`
	Offset         int    `json:"offset"`
	Count          int    `json:"count"`
}

// GetAuthorizerListResponse 获取授权方列表响应
type GetAuthorizerListResponse struct {
	core.APIResponse
	TotalCount int `json:"total_count"`
	List       []struct {
		AuthorizerAppID string `json:"authorizer_appid"`
		RefreshToken    string `json:"refresh_token,omitempty"` // 虽然文档写了，但是实际上获取不到
		AuthTime        int64  `json:"auth_time"`
	} `json:"list"`
}

// GetComponentAccessToken 获取第三方平台access_token
func (c *Client) GetComponentAccessToken(ctx context.Context, verifyTicket string) (*storage.ComponentAccessToken, error) {
	// 直接从存储中获取令牌，不调用GetComponentToken避免递归
	token, err := c.storage.GetComponentToken(ctx)
	if err != nil {
		return nil, err
	}

	// 如果令牌存在且未过期，直接返回
	if token != nil && token.ExpiresAt.After(time.Now()) {
		return token, nil
	}

	// 如果verifyTicket为空，从存储中获取验证票据
	if verifyTicket == "" {
		verifyTicketObj, err := c.storage.GetComponentVerifyTicket(ctx)
		if err != nil {
			wd, _ := os.Getwd()
			return nil, fmt.Errorf("获取验证票据失败: %v [storage=%T wd=%s]", err, c.storage, wd)
		}
		if verifyTicketObj == nil || time.Now().After(verifyTicketObj.ExpiresAt) {
			wd, _ := os.Getwd()
			return nil, fmt.Errorf("验证票据不存在或已过期 [storage=%T wd=%s]", c.storage, wd)
		}
		verifyTicket = verifyTicketObj.Ticket
	}

	request := ComponentTokenRequest{
		ComponentAppID:        c.config.ComponentAppID,
		ComponentAppSecret:    c.config.ComponentAppSecret,
		ComponentVerifyTicket: verifyTicket,
	}

	var result struct {
		core.APIResponse
		ComponentAccessToken string `json:"component_access_token"`
		ExpiresIn            int    `json:"expires_in"`
	}

	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLComponentToken,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	token = &storage.ComponentAccessToken{
		AccessToken: result.ComponentAccessToken,
		ExpiresIn:   result.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}

	// 保存到存储
	if err := c.SetComponentToken(token); err != nil {
		c.logger.Warn(fmt.Sprintf("保存开放平台令牌失败: %v", err))
	}

	return token, nil
}

// preAuthCodeExpireMargin 预授权码过期安全余量，避免授权流程进行中预授权码失效
const preAuthCodeExpireMargin = 60 * time.Second

// GetPreAuthCodeFromAPI 从微信API获取预授权码
func (c *Client) GetPreAuthCodeFromAPI(ctx context.Context) (*PreAuthCodeResponse, error) {
	// 先从存储中获取（留 60 秒安全余量，与 GetPreAuthCode 复用策略保持一致）
	if preAuthCode, err := c.storage.GetPreAuthCode(ctx); err == nil && preAuthCode != nil &&
		time.Now().Before(preAuthCode.ExpiresAt.Add(-preAuthCodeExpireMargin)) {
		return &PreAuthCodeResponse{
			APIResponse: core.APIResponse{ErrCode: 0, ErrMsg: ""},
			PreAuthCode: preAuthCode.PreAuthCode,
			ExpiresIn:   preAuthCode.ExpiresIn,
		}, nil
	}

	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := PreAuthCodeRequest{
		ComponentAppID: c.config.ComponentAppID,
	}

	var result PreAuthCodeResponse
	apiURL := fmt.Sprintf("%s?component_access_token=%s", URLPreAuthCode, componentToken.AccessToken)
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    apiURL,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	// 保存到存储
	preAuthCode := &storage.PreAuthCode{
		PreAuthCode: result.PreAuthCode,
		ExpiresIn:   result.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}

	// 保存到存储
	if err := c.SetPreAuthCode(ctx, preAuthCode); err != nil {
		c.logger.Warn(fmt.Sprintf("保存预授权码失败: %v", err))
	}

	return &result, nil
}

// HandleAuthorizationEvent 处理微信开放平台授权事件
// 支持明文和加密两种消息格式
// 根据微信官方文档<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/authorize_event.html" index="0">0</mcreference>，
// 接收POST请求后只需直接返回字符串"success"
func (c *Client) HandleAuthorizationEvent(ctx context.Context, xmlData []byte, msgSignature, timestamp, nonce, encryptType string) (string, error) {
	// 首先检测是否为加密消息
	var encryptedMsg struct {
		XMLName    xml.Name `xml:"xml"`
		AppId      string   `xml:"AppId"`
		Encrypt    string   `xml:"Encrypt"`
		CreateTime int64    `xml:"CreateTime"`
	}

	// 记录接收到的参数用于调试
	c.logger.Info(fmt.Sprintf("处理授权事件，参数 - timestamp: %s, nonce: %s, encrypt_type: %s, msg_signature: %s",
		timestamp, nonce, encryptType, msgSignature))

	// 解析XML判断消息类型：以 XML 是否包含 Encrypt 字段为准，
	// 不依赖 encrypt_type query 参数（宿主网关可能未透传该参数，
	// 若仅按 encrypt_type 判定会导致加密事件被静默丢弃，如 component_verify_ticket）
	if err := xml.Unmarshal(xmlData, &encryptedMsg); err != nil {
		c.logger.Error(fmt.Sprintf("解析消息XML失败: %v", err))
		c.notifyEventError("parse_encrypted_xml", err)
		return "success", nil // 即使解析失败也返回success
	}

	if encryptedMsg.Encrypt != "" {
		if encryptType != "" && encryptType != "aes" {
			c.logger.Warn(fmt.Sprintf("XML包含Encrypt字段但encrypt_type=%s，仍按加密消息处理", encryptType))
		}
		c.logger.Info(fmt.Sprintf("检测到加密消息，开始解密处理，AppId: %s", encryptedMsg.AppId))

		// 解密消息
		decryptedData, err := c.DecryptMessage(encryptedMsg.Encrypt, msgSignature, timestamp, nonce)
		if err != nil {
			c.logger.Error(fmt.Sprintf("解密授权事件消息失败: %v", err))
			c.notifyEventError("decrypt", err)
			return "success", nil // 即使解密失败也返回success
		}

		c.logger.Info(fmt.Sprintf("解密成功，解密后内容: %s", string(decryptedData)))

		// 使用解密后的数据继续处理
		xmlData = decryptedData
	}

	// 解析XML获取基础事件信息
	var baseEvent AuthorizationEvent
	err := xml.Unmarshal(xmlData, &baseEvent)
	if err != nil {
		c.logger.Error(fmt.Sprintf("解析授权事件XML失败: %v", err))
		c.notifyEventError("parse_event_xml", err)
		return "success", nil // 即使解析失败也返回success
	}

	// 验证事件签名和时间戳
	if err = c.validateAuthorizationEvent(&baseEvent); err != nil {
		c.logger.Error(fmt.Sprintf("授权事件验证失败: %v", err))
		c.notifyEventError("validate_event", err)
		return "success", nil // 即使验证失败也返回success
	}

	// 根据事件类型进行处理
	switch baseEvent.InfoType {
	case "authorized":
		var event AuthorizedEvent
		err = xml.Unmarshal(xmlData, &event)
		if err != nil {
			c.logger.Error(fmt.Sprintf("解析授权成功事件失败: %v", err))
			c.notifyEventError("parse_authorized_event", err)
			break
		}
		c.logger.Info(fmt.Sprintf("解析授权成功事件成功，事件内容: %+v", event))
		if err := c.GetEventHandler().HandleAuthorized(ctx, &event); err != nil {
			c.logger.Error(fmt.Sprintf("处理授权成功事件失败: %v", err))
			c.notifyEventError("handle_authorized", err)
		}

	case "unauthorized":
		var event UnauthorizedEvent
		err = xml.Unmarshal(xmlData, &event)
		if err != nil {
			c.logger.Error(fmt.Sprintf("解析取消授权事件失败: %v", err))
			c.notifyEventError("parse_unauthorized_event", err)
			break
		}
		c.logger.Info(fmt.Sprintf("解析取消授权事件成功，事件内容: %+v", event))
		if err := c.GetEventHandler().HandleUnauthorized(ctx, &event); err != nil {
			c.logger.Error(fmt.Sprintf("处理取消授权事件失败: %v", err))
			c.notifyEventError("handle_unauthorized", err)
		}

	case "updateauthorized":
		var event UpdateAuthorizedEvent
		err := xml.Unmarshal(xmlData, &event)
		if err != nil {
			c.logger.Error(fmt.Sprintf("解析授权更新事件失败: %v", err))
			c.notifyEventError("parse_updateauthorized_event", err)
			break
		}
		c.logger.Info(fmt.Sprintf("解析授权更新事件成功，事件内容: %+v", event))
		if err = c.GetEventHandler().HandleUpdateAuthorized(ctx, &event); err != nil {
			c.logger.Error(fmt.Sprintf("处理授权更新事件失败: %v", err))
			c.notifyEventError("handle_updateauthorized", err)
		}

	case "component_verify_ticket":
		var event ComponentVerifyTicketEvent
		err := xml.Unmarshal(xmlData, &event)
		if err != nil {
			c.logger.Error(fmt.Sprintf("解析验证票据事件失败: %v", err))
			c.notifyEventError("parse_component_verify_ticket_event", err)
			// 根据微信官方文档要求，即使解析失败也必须返回success
			break
		}
		c.logger.Info(fmt.Sprintf("解析验证票据事件成功，事件内容: %+v", event))
		// 存储验证票据
		if err := c.storage.SaveComponentVerifyTicket(ctx, event.ComponentVerifyTicket); err != nil {
			c.logger.Error(fmt.Sprintf("存储验证票据失败: %v", err))
			c.notifyEventError("save_component_verify_ticket", err)
			// 根据微信官方文档要求，即使存储失败也必须返回success
		}

		if err := c.GetEventHandler().HandleComponentVerifyTicket(ctx, &event); err != nil {
			c.logger.Error(fmt.Sprintf("处理验证票据事件失败: %v", err))
			c.notifyEventError("handle_component_verify_ticket", err)
			// 根据微信官方文档要求，即使处理失败也必须返回success
		}

	case "encoding_aes_key_changed":
		var event EncodingAESKeyChangedEvent
		err := xml.Unmarshal(xmlData, &event)
		if err != nil {
			c.logger.Error(fmt.Sprintf("解析EncodingAESKey变更事件失败: %v", err))
			c.notifyEventError("parse_encoding_aes_key_changed_event", err)
			break
		}
		c.logger.Info(fmt.Sprintf("解析EncodingAESKey变更事件成功，事件内容: %+v", event))

		// 持久化旧密钥（官方要求支持密钥轮换过渡期使用旧密钥解密），
		// 之后解密实例通过 NewWXBizMsgCryptWithStorage 从存储加载该密钥作为 PrevKey
		c.cryptMu.RLock()
		oldKey := c.config.EncodingAESKey
		c.cryptMu.RUnlock()
		if oldKey != "" && oldKey != event.NewEncodingAESKey {
			if err := c.storage.SavePrevEncodingAESKey(ctx, c.config.ComponentAppID, oldKey); err != nil {
				c.logger.Error(fmt.Sprintf("保存上一次EncodingAESKey失败: %v", err))
				c.notifyEventError("save_prev_encoding_aes_key", err)
			}
		}

		// 更新配置中的EncodingAESKey（加锁保护，避免与并发解密产生数据竞争）
		c.cryptMu.Lock()
		c.config.EncodingAESKey = event.NewEncodingAESKey
		c.cryptMu.Unlock()

		if err := c.GetEventHandler().HandleEncodingAESKeyChanged(ctx, &event); err != nil {
			c.logger.Error(fmt.Sprintf("处理EncodingAESKey变更事件失败: %v", err))
			c.notifyEventError("handle_encoding_aes_key_changed", err)
		}

	default:
		c.logger.Warn(fmt.Sprintf("收到未知的授权事件类型: %s", baseEvent.InfoType))
	}

	// 根据微信官方文档要求，必须返回"success"字符串
	return "success", nil
}

// validateAuthorizationEvent 验证授权事件
func (c *Client) validateAuthorizationEvent(event *AuthorizationEvent) error {
	// 验证AppID是否匹配（官方对事件AppID校验标注为「可选」：
	// 部分事件回调可能不携带AppId，为空时跳过校验仅记告警；非空且不匹配时才判为错误）
	if event.AppId == "" {
		c.logger.Warn("授权事件未携带AppId，跳过AppID校验")
	} else if event.AppId != c.config.ComponentAppID {
		return fmt.Errorf("AppID不匹配: expected=%s, actual=%s", c.config.ComponentAppID, event.AppId)
	}

	// 验证时间戳，防止重放攻击
	currentTime := time.Now().Unix()

	// 处理CreateTime为0的情况（某些微信回调可能不包含CreateTime字段）
	if event.CreateTime == 0 {
		c.logger.Warn("授权事件CreateTime为0，跳过时间戳验证")
		return nil
	}

	// 检查时间戳是否在有效范围内（当前时间±5分钟）
	if math.Abs(float64(currentTime-event.CreateTime)) > 300 { // 5分钟容忍
		return fmt.Errorf("时间戳过期: current=%d, event=%d", currentTime, event.CreateTime)
	}

	return nil
}

// DecryptMessage 解密消息（用于处理加密的授权事件）
//
// 使用带存储的加解密实例：密钥轮换过渡期会自动降级尝试存储中的上一次密钥（PrevKey）
func (c *Client) DecryptMessage(encryptedMsg, msgSignature, timestamp, nonce string) ([]byte, error) {
	// 验证消息签名
	if err := c.verifySignature(msgSignature, timestamp, nonce, encryptedMsg); err != nil {
		return nil, fmt.Errorf("消息签名验证失败: %v", err)
	}

	c.cryptMu.RLock()
	aesKey := c.config.EncodingAESKey
	c.cryptMu.RUnlock()

	// 使用crypto包中的解密实现（带存储，支持密钥轮换过渡期的旧密钥解密）
	wxCrypt, err := crypto.NewWXBizMsgCryptWithStorage(c.config.ComponentToken, aesKey, c.config.ComponentAppID, c.storage)
	if err != nil {
		return nil, fmt.Errorf("创建解密实例失败: %v", err)
	}

	// 解密消息
	decryptedMsg, err := wxCrypt.DecryptMsg(msgSignature, timestamp, nonce, encryptedMsg)
	if err != nil {
		return nil, fmt.Errorf("消息解密失败: %v", err)
	}

	return []byte(decryptedMsg), nil
}

// verifySignature 验证消息签名
func (c *Client) verifySignature(signature, timestamp, nonce, encryptedMsg string) error {
	// 根据微信开放平台签名算法验证签名
	// 签名算法：sha1(sort(token, timestamp, nonce, encryptedMsg))

	// 获取配置中的Token
	token := c.config.ComponentToken

	// 使用crypto包中的签名验证实现
	c.cryptMu.RLock()
	aesKey := c.config.EncodingAESKey
	c.cryptMu.RUnlock()
	wxCrypt := crypto.NewWXBizMsgCrypt(token, aesKey, c.config.ComponentAppID)

	// 验证签名
	if !wxCrypt.VerifySignature(signature, timestamp, nonce, encryptedMsg) {
		return fmt.Errorf("签名验证失败")
	}

	return nil
}

// QueryAuth 使用授权码换取授权信息
func (c *Client) QueryAuth(ctx context.Context, authorizationCode string) (*QueryAuthResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := QueryAuthRequest{
		ComponentAppID:    c.config.ComponentAppID,
		AuthorizationCode: authorizationCode,
	}

	var result QueryAuthResponse
	queryAuthUrl := fmt.Sprintf("%s?component_access_token=%s", URLQueryAuth, componentToken.AccessToken)
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    queryAuthUrl,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	// 缓存授权方token
	if err := c.SetAuthorizerToken(
		result.AuthorizationInfo.AuthorizerAppID,
		result.AuthorizationInfo.AuthorizerAccessToken,
		result.AuthorizationInfo.AuthorizerRefreshToken,
		result.AuthorizationInfo.ExpiresIn,
	); err != nil {
		c.logger.Warn(fmt.Sprintf("缓存授权方token失败: %v", err))
	}

	return &result, nil
}

// RefreshAuthorizerToken 刷新授权方access_token
func (c *Client) RefreshAuthorizerToken(ctx context.Context, authorizerAppID, refreshToken string) (*AuthorizationInfo, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := AuthorizerTokenRequest{
		ComponentAppID:         c.config.ComponentAppID,
		AuthorizerAppID:        authorizerAppID,
		AuthorizerRefreshToken: refreshToken,
	}

	var result struct {
		core.APIResponse
		AuthorizationInfo
	}

	apiURL := fmt.Sprintf("%s?component_access_token=%s", URLAuthorizerToken,
		componentToken.AccessToken)

	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    apiURL,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	// 更新缓存
	if err := c.SetAuthorizerToken(
		authorizerAppID,
		result.AuthorizerAccessToken,
		result.AuthorizerRefreshToken,
		result.ExpiresIn,
	); err != nil {
		c.logger.Warn(fmt.Sprintf("更新授权方token失败: %v", err))
	}

	return &result.AuthorizationInfo, nil
}

// GetAuthorizerInfo 获取授权方信息
func (c *Client) GetAuthorizerInfo(ctx context.Context, authorizerAppID string) (*GetAuthorizerInfoResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := GetAuthorizerInfoRequest{
		ComponentAppID:  c.config.ComponentAppID,
		AuthorizerAppID: authorizerAppID,
	}

	var result GetAuthorizerInfoResponse
	getAuthorizerInfoUrl := fmt.Sprintf("%s?component_access_token=%s", URLGetAuthorizerInfo, componentToken.AccessToken)
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    getAuthorizerInfoUrl,
		Body:   request,
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

// GetAuthorizerList 获取授权方列表
func (c *Client) GetAuthorizerList(ctx context.Context, offset, count int) (*GetAuthorizerListResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := GetAuthorizerListRequest{
		ComponentAppID: c.config.ComponentAppID,
		Offset:         offset,
		Count:          count,
	}

	var result GetAuthorizerListResponse
	getAuthorizerListUrl := fmt.Sprintf("%s?component_access_token=%s", URLGetAuthorizerList, componentToken.AccessToken)
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    getAuthorizerListUrl,
		Body:   request,
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

// GetAllAuthorizers 获取所有授权方列表，自动处理分页
// 该方法会循环调用GetAuthorizerList，直到获取所有授权方数据
// 返回所有授权方信息的切片和可能的错误
func (c *Client) GetAllAuthorizers(ctx context.Context) (allAuthorizers []struct {
	AuthorizerAppID string `json:"authorizer_appid"`
	RefreshToken    string `json:"refresh_token,omitempty"` // 虽然文档写了，但是实际上获取不到
	AuthTime        int64  `json:"auth_time"`
}, err error) {
	const (
		pageSize = 500 // 每页获取的数量，微信API最大支持500
		maxRetry = 3   // 最大重试次数
	)

	offset := 0

	for {
		var response *GetAuthorizerListResponse

		// 重试机制
		for retry := 0; retry < maxRetry; retry++ {
			response, err = c.GetAuthorizerList(ctx, offset, pageSize)
			if err == nil {
				break
			}

			// 如果是最后一次重试仍然失败，则返回错误
			if retry == maxRetry-1 {
				return nil, fmt.Errorf("获取授权方列表失败，重试%d次后仍然失败: %v", maxRetry, err)
			}

			// 等待一段时间后重试
			time.Sleep(time.Duration(retry+1) * time.Second)
		}

		// 检查响应是否有效
		if response == nil {
			return nil, fmt.Errorf("获取授权方列表响应为空")
		}

		// 添加当前页的数据到总列表
		if len(response.List) > 0 {
			allAuthorizers = append(allAuthorizers, response.List...)
		}

		// 判断是否还有更多数据
		// 如果当前页返回的数量小于请求的数量，或者已经获取了所有数据，则结束循环
		if len(response.List) < pageSize || len(allAuthorizers) >= response.TotalCount {
			break
		}

		// 更新offset，准备获取下一页
		offset += pageSize

		// 防止无限循环，如果获取的数据量超过总数量，则退出
		if offset >= response.TotalCount {
			break
		}

		// 添加小延迟，避免对API造成过大压力
		time.Sleep(100 * time.Millisecond)
	}

	return allAuthorizers, nil
}

// GenerateAuthURL 生成授权链接
// authType: 授权类型 (1:手机端仅展示公众号, 2:仅展示小程序, 3:公众号和小程序都展示, 4:小程序推客账号, 5:视频号账号, 6:全部, 8:带货助手账号)
// platform: 平台类型 ("pc": PC端, "mobile": 移动端)
func (c *Client) GenerateAuthURL(preAuthCode string, authType int, bizAppID string, platform string) string {
	var baseURL string

	if platform == "mobile" {
		// 移动端授权页面URL
		baseURL = core.MobileAuthPageURL
	} else {
		// PC端授权页面URL
		baseURL = core.PCAuthPageURL
	}

	// 如果authType不在可选范围内，则取默认值6
	if authType < 1 || authType > 8 || authType == 7 {
		authType = 6
	}

	params := url.Values{
		"component_appid": {c.config.ComponentAppID},
		"pre_auth_code":   {preAuthCode},
		"redirect_uri":    {c.config.RedirectURI},
		"auth_type":       {fmt.Sprintf("%d", authType)},
	}

	if bizAppID != "" {
		params.Set("biz_appid", bizAppID)
	}

	// 移动端授权链接需要添加action和no_scan参数
	if platform == "mobile" {
		params.Set("action", "bindcomponent")
		params.Set("no_scan", "1")
	}

	urlStr := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 移动端授权链接需要添加#wechat_redirect后缀
	if platform == "mobile" {
		urlStr += "#wechat_redirect"
	}

	return urlStr
}

// GeneratePcAuthURL 生成PC端授权链接
func (c *Client) GeneratePcAuthURL(preAuthCode string, authType int, bizAppID string) string {
	return c.GenerateAuthURL(preAuthCode, authType, bizAppID, "pc")
}

// GenerateMobileAuthURL 生成移动端授权链接
func (c *Client) GenerateMobileAuthURL(preAuthCode string, authType int, bizAppID string) string {
	return c.GenerateAuthURL(preAuthCode, authType, bizAppID, "mobile")
}

// ClearQuota 重置第三方平台自身的API调用次数
// 参考文档：https://developers.weixin.qq.com/doc/service/api/apimanage/api_clearquota.html
// 注意事项：清第三方平台自身quota需用 component_access_token，每个账号每月共10次机会
func (c *Client) ClearQuota(ctx context.Context) (*core.APIResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	return c.clearQuota(ctx, componentToken.AccessToken, c.config.ComponentAppID)
}

// ClearAuthorizerQuota 代授权方（公众号/小程序）重置API调用次数
// 注意事项：代商家清除 quota 必须使用 authorizer_access_token，且消耗的是授权方自身的清零机会
func (c *Client) ClearAuthorizerQuota(ctx context.Context, authorizerAppID string) (*core.APIResponse, error) {
	if authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	authorizerToken, err := c.GetAuthorizerAccessToken(ctx, authorizerAppID)
	if err != nil {
		return nil, err
	}

	return c.clearQuota(ctx, authorizerToken, authorizerAppID)
}

// clearQuota 重置API调用次数的实际实现
// 官方请求体仅包含 {"appid":"要被清空的账号的appid"}
func (c *Client) clearQuota(ctx context.Context, accessToken, appID string) (*core.APIResponse, error) {
	request := ClearQuotaRequest{
		AppID: appID,
	}

	var result core.APIResponse
	clearQuotaURL := fmt.Sprintf("%s?access_token=%s", URLClearQuota, accessToken)
	err := c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    clearQuotaURL,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result
	}

	return &result, nil
}

// GetApiQuota 查询API调用额度
// authorizerAppID 为空时查询第三方平台自身额度（component_access_token），
// 非空时查询授权方额度（authorizer_access_token）
// 参考文档：https://developers.weixin.qq.com/doc/service/api/apimanage/api_quota_get.html
func (c *Client) GetApiQuota(ctx context.Context, authorizerAppID string) (*GetApiQuotaResponse, error) {
	accessToken := ""
	targetAppID := c.config.ComponentAppID

	if authorizerAppID != "" {
		authorizerToken, err := c.GetAuthorizerAccessToken(ctx, authorizerAppID)
		if err != nil {
			return nil, err
		}
		accessToken = authorizerToken
		targetAppID = authorizerAppID
	} else {
		componentToken, err := c.GetComponentAccessToken(ctx, "")
		if err != nil {
			return nil, err
		}
		accessToken = componentToken.AccessToken
	}

	request := GetApiQuotaRequest{
		AppID: targetAppID,
	}

	var result GetApiQuotaResponse
	getApiQuotaURL := fmt.Sprintf("%s?access_token=%s", URLGetApiQuota, accessToken)
	err := c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    getApiQuotaURL,
		Body:   request,
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

// GetRidInfo 查询rid信息
func (c *Client) GetRidInfo(ctx context.Context, rid string) (*GetRidInfoResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := GetRidInfoRequest{
		RID: rid,
	}

	var result GetRidInfoResponse
	getRidInfoUrl := fmt.Sprintf("%s?access_token=%s", URLGetRidInfo, componentToken.AccessToken)
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    getRidInfoUrl,
		Body:   request,
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

// ClearComponentQuota 使用AppSecret重置第三方平台自身API调用次数
// 参考文档：https://developers.weixin.qq.com/doc/oplatform/openApi/openapi/api_componentclearquota_v2.html
// 注意事项：
//  1. 该接口无 query 参数，仅支持 POST，路径为 /cgi-bin/component/clear_quota/v2
//  2. 请求体字段为 appsecret（不是 component_appsecret）
//  3. 与 clear_quota 共用每月10次清零机会
func (c *Client) ClearComponentQuota(ctx context.Context) (*core.APIResponse, error) {
	return c.clearComponentQuota(ctx, "")
}

// ClearAuthorizerQuotaBySecret 使用AppSecret代授权方（公众号/小程序）重置API调用次数
// 该接口通过 appsecret 调用，可在授权方 access_token 已耗尽时使用，appid 必须是已授权的账号
func (c *Client) ClearAuthorizerQuotaBySecret(ctx context.Context, authorizerAppID string) (*core.APIResponse, error) {
	if authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}
	return c.clearComponentQuota(ctx, authorizerAppID)
}

// clearComponentQuota 使用AppSecret重置API调用次数的实际实现
func (c *Client) clearComponentQuota(ctx context.Context, authorizerAppID string) (*core.APIResponse, error) {
	request := ClearComponentQuotaRequest{
		ComponentAppID: c.config.ComponentAppID,
		AppSecret:      c.config.ComponentAppSecret,
		AppID:          authorizerAppID,
	}

	var result core.APIResponse
	err := c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLClearComponentQuota,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result
	}

	return &result, nil
}

// SetAuthorizerOption 设置授权方选项信息
func (c *Client) SetAuthorizerOption(ctx context.Context, authorizerAppID, optionName, optionValue string) (*core.APIResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := SetAuthorizerOptionRequest{
		ComponentAppID:  c.config.ComponentAppID,
		AuthorizerAppID: authorizerAppID,
		OptionName:      optionName,
		OptionValue:     optionValue,
	}

	var result core.APIResponse
	setAuthorizerOptionUrl := fmt.Sprintf("%s?component_access_token=%s", URLSetAuthorizerOption, componentToken.AccessToken)
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    setAuthorizerOptionUrl,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result
	}

	return &result, nil
}

// GetAuthorizerOption 获取授权方选项信息
func (c *Client) GetAuthorizerOption(ctx context.Context, authorizerAppID, optionName string) (*GetAuthorizerOptionResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := GetAuthorizerOptionRequest{
		ComponentAppID:  c.config.ComponentAppID,
		AuthorizerAppID: authorizerAppID,
		OptionName:      optionName,
	}

	var result GetAuthorizerOptionResponse
	getAuthorizerOptionUrl := fmt.Sprintf("%s?component_access_token=%s", URLGetAuthorizerOption, componentToken.AccessToken)
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    getAuthorizerOptionUrl,
		Body:   request,
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

// GetTemplateDraftList 获取草稿箱列表
func (c *Client) GetTemplateDraftList(ctx context.Context) (*GetTemplateDraftListResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	var result GetTemplateDraftListResponse
	getTemplateDraftListUrl := fmt.Sprintf("%s?access_token=%s", URLWxaGetTemplateDraftList, componentToken.AccessToken)
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    getTemplateDraftListUrl,
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

// AddToTemplate 将草稿添加到模板库
func (c *Client) AddToTemplate(ctx context.Context, draftID int64, templateType int) (*core.APIResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := AddToTemplateRequest{
		DraftID:      draftID,
		TemplateType: templateType,
	}

	var result core.APIResponse
	addToTemplateUrl := fmt.Sprintf("%s?access_token=%s", URLWxaAddToTemplate, componentToken.AccessToken)
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    addToTemplateUrl,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result
	}

	return &result, nil
}

// GetTemplateList 获取模板列表
func (c *Client) GetTemplateList(ctx context.Context) (*GetTemplateListResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	var result GetTemplateListResponse
	getTemplateListUrl := fmt.Sprintf("%s?access_token=%s", URLWxaGetTemplateList, componentToken.AccessToken)
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    getTemplateListUrl,
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

// DeleteTemplate 删除代码模板
func (c *Client) DeleteTemplate(ctx context.Context, templateID int64) (*core.APIResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := DeleteTemplateRequest{
		TemplateID: templateID,
	}

	var result core.APIResponse
	deleteTemplateUrl := fmt.Sprintf("%s?access_token=%s", URLWxaDeleteTemplate, componentToken.AccessToken)
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    deleteTemplateUrl,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result
	}

	return &result, nil
}
