package openplatform

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/wego/crypto"
	"github.com/jcbowen/wego/storage"
)

// HTTPClient HTTP客户端接口
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// APIClient API客户端
type APIClient struct {
	config       *OpenPlatformConfig
	httpClient   HTTPClient
	storage      storage.TokenStorage
	logger       debugger.LoggerInterface
	eventHandler EventHandler          // 事件处理器
	crypt        *crypto.WXBizMsgCrypt // 消息加解密实例
}

// NewAPIClient 创建新的API客户端（使用默认文件存储）
func NewAPIClient(config *OpenPlatformConfig) *APIClient {
	// 使用当前工作目录下的 wego_storage 文件夹作为默认存储路径
	fileStorage, err := storage.NewFileStorage("./runtime/wego_storage")
	if err != nil {
		// 如果文件存储创建失败，回退到内存存储并输出日志
		logger := &debugger.DefaultLogger{}
		logger.Warn(fmt.Sprintf("文件存储创建失败，回退到内存存储: %v", err))
		return NewAPIClientWithStorage(config, storage.NewMemoryStorage())
	}
	return NewAPIClientWithStorage(config, fileStorage)
}

// NewAPIClientWithStorage 创建新的API客户端（使用自定义存储）
func NewAPIClientWithStorage(config *OpenPlatformConfig, storage storage.TokenStorage) *APIClient {
	client := &APIClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		storage:    storage,
		logger:     &debugger.DefaultLogger{},
		crypt:      crypto.NewWXBizMsgCrypt(config.ComponentToken, config.EncodingAESKey, config.ComponentAppID),
	}

	return client
}

// SetLogger 设置自定义日志器
func (c *APIClient) SetLogger(logger debugger.LoggerInterface) {
	c.logger = logger
}

// SetHTTPClient 设置自定义HTTP客户端
func (c *APIClient) SetHTTPClient(client HTTPClient) {
	c.httpClient = client
}

// SetEventHandler 设置事件处理器
func (c *APIClient) SetEventHandler(handler EventHandler) {
	c.eventHandler = handler
}

// GetEventHandler 获取事件处理器
func (c *APIClient) GetEventHandler() EventHandler {
	if c.eventHandler == nil {
		return &DefaultEventHandler{}
	}
	return c.eventHandler
}

// GetConfig 获取配置信息
func (c *APIClient) GetConfig() *OpenPlatformConfig {
	return c.config
}

// GetLogger 获取日志器
func (c *APIClient) GetLogger() debugger.LoggerInterface {
	return c.logger
}

// SetComponentToken 设置开放平台令牌
func (c *APIClient) SetComponentToken(token *storage.ComponentAccessToken) error {
	return c.storage.SaveComponentToken(context.Background(), token)
}

// GetComponentToken 获取开放平台令牌
func (c *APIClient) GetComponentToken(ctx context.Context) (*storage.ComponentAccessToken, error) {
	return c.storage.GetComponentToken(ctx)
}

// SetPreAuthCode 设置预授权码
func (c *APIClient) SetPreAuthCode(ctx context.Context, preAuthCode *storage.PreAuthCode) error {
	return c.storage.SavePreAuthCode(ctx, preAuthCode)
}

// GetPreAuthCode 获取预授权码
func (c *APIClient) GetPreAuthCode(ctx context.Context) (*storage.PreAuthCode, error) {
	return c.storage.GetPreAuthCode(ctx)
}

// SetAuthorizerToken 设置授权方token信息
func (c *APIClient) SetAuthorizerToken(authorizerAppID, accessToken, refreshToken string, expiresIn int) error {
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
func (c *APIClient) GetAuthorizerAccessToken(ctx context.Context, authorizerAppID string) (string, error) {
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

// GetComponentVerifyTicket 获取验证票据
// @param ctx context.Context 上下文
// @return *storage.ComponentVerifyTicket 验证票据结构，包含票据内容和有效期信息
// @return error 错误信息
func (c *APIClient) GetComponentVerifyTicket(ctx context.Context) (*storage.ComponentVerifyTicket, error) {
	return c.storage.GetComponentVerifyTicket(ctx)
}

// SaveComponentVerifyTicket 保存验证票据
// @param ctx context.Context 上下文
// @param ticket string 票据内容
// @return error 错误信息
func (c *APIClient) SaveComponentVerifyTicket(ctx context.Context, ticket string) error {
	return c.storage.SaveComponentVerifyTicket(ctx, ticket)
}

// refreshAuthorizerAccessToken 刷新授权方access_token
func (c *APIClient) refreshAuthorizerAccessToken(ctx context.Context, authorizerAppID string) (string, error) {
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

// MakeRequest 发送HTTP请求的通用方法
func (c *APIClient) MakeRequest(ctx context.Context, method, url string, body interface{}, result interface{}) error {
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
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			c.logger.Error(fmt.Sprintf("关闭响应体失败: %v", err))
		}
	}(resp.Body)

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
func (c *APIClient) MakeRequestRaw(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	return resp, nil
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
	APIResponse
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
	APIResponse
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
	APIResponse
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
	APIResponse
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
	APIResponse
	TotalCount int `json:"total_count"`
	List       []struct {
		AuthorizerAppID string `json:"authorizer_appid"`
		RefreshToken    string `json:"refresh_token"`
		AuthTime        int64  `json:"auth_time"`
	} `json:"list"`
}

// GetComponentAccessToken 获取第三方平台access_token
func (c *APIClient) GetComponentAccessToken(ctx context.Context, verifyTicket string) (*storage.ComponentAccessToken, error) {
	// 先从存储中获取
	if token, err := c.GetComponentToken(ctx); err == nil && token != nil && token.ExpiresAt.After(time.Now()) {
		return token, nil
	}

	// 如果verifyTicket为空，从存储中获取验证票据
	if verifyTicket == "" {
		verifyTicketObj, err := c.storage.GetComponentVerifyTicket(ctx)
		if err != nil {
			return nil, fmt.Errorf("获取验证票据失败: %v", err)
		}
		if verifyTicketObj == nil {
			return nil, fmt.Errorf("验证票据不存在或已过期")
		}
		verifyTicket = verifyTicketObj.Ticket
	}

	request := ComponentTokenRequest{
		ComponentAppID:        c.config.ComponentAppID,
		ComponentAppSecret:    c.config.ComponentAppSecret,
		ComponentVerifyTicket: verifyTicket,
	}

	var result struct {
		APIResponse
		ComponentAccessToken string `json:"component_access_token"`
		ExpiresIn            int    `json:"expires_in"`
	}

	err := c.MakeRequest(ctx, "POST", APIComponentTokenURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	token := &storage.ComponentAccessToken{
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

// GetPreAuthCodeFromAPI 从微信API获取预授权码
func (c *APIClient) GetPreAuthCodeFromAPI(ctx context.Context) (*PreAuthCodeResponse, error) {
	// 先从存储中获取
	if preAuthCode, err := c.storage.GetPreAuthCode(ctx); err == nil && preAuthCode != nil && preAuthCode.ExpiresAt.After(time.Now()) {
		return &PreAuthCodeResponse{
			APIResponse: APIResponse{ErrCode: 0, ErrMsg: ""},
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
	apiURL := fmt.Sprintf("%s?component_access_token=%s", APIPreAuthCodeURL, url.QueryEscape(componentToken.AccessToken))
	err = c.MakeRequest(ctx, "POST", apiURL, request, &result)
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
func (c *APIClient) HandleAuthorizationEvent(ctx context.Context, xmlData []byte, msgSignature, timestamp, nonce, encryptType string) (string, error) {
	// 首先检测是否为加密消息
	var encryptedMsg struct {
		XMLName    xml.Name `xml:"xml"`
		AppId      string   `xml:"AppId"`
		Encrypt    string   `xml:"Encrypt"`
		CreateTime int64    `xml:"CreateTime"`
	}

	// 记录接收到的参数用于调试
	c.logger.Debug(fmt.Sprintf("处理授权事件，参数 - timestamp: %s, nonce: %s, encrypt_type: %s, msg_signature: %s",
		timestamp, nonce, encryptType, msgSignature))

	// 判断消息类型：根据encrypt_type参数或XML内容检测
	isEncrypted := encryptType == "aes" && msgSignature != ""

	// 如果URL参数表明是加密消息，或者XML内容包含Encrypt字段，则进行解密
	if isEncrypted {
		c.logger.Debug("URL参数表明是加密消息，开始解密处理")

		// 尝试解析XML获取加密内容
		if err := xml.Unmarshal(xmlData, &encryptedMsg); err != nil {
			c.logger.Error(fmt.Sprintf("解析加密消息XML失败: %v", err))
			return "success", nil // 即使解析失败也返回success
		}

		if encryptedMsg.Encrypt == "" {
			c.logger.Warn("URL参数表明是加密消息，但XML中未找到Encrypt字段")
			return "success", nil
		}

		c.logger.Debug(fmt.Sprintf("检测到加密消息，开始解密处理，AppId: %s", encryptedMsg.AppId))

		// 解密消息
		decryptedData, err := c.DecryptMessage(encryptedMsg.Encrypt, msgSignature, timestamp, nonce)
		if err != nil {
			c.logger.Error(fmt.Sprintf("解密授权事件消息失败: %v", err))
			return "success", nil // 即使解密失败也返回success
		}

		c.logger.Debug(fmt.Sprintf("解密成功，解密后内容: %s", string(decryptedData)))

		// 使用解密后的数据继续处理
		xmlData = decryptedData
	} else {
		// 明文消息：检查XML是否包含Encrypt字段（可能是误传参数）
		if err := xml.Unmarshal(xmlData, &encryptedMsg); err == nil && encryptedMsg.Encrypt != "" {
			c.logger.Warn("XML包含Encrypt字段但URL参数未表明是加密消息，可能参数传递有误")
			// 继续按明文处理，但记录警告
		}
	}

	// 解析XML获取基础事件信息
	var baseEvent AuthorizationEvent
	err := xml.Unmarshal(xmlData, &baseEvent)
	if err != nil {
		c.logger.Error(fmt.Sprintf("解析授权事件XML失败: %v", err))
		return "success", nil // 即使解析失败也返回success
	}

	// 验证事件签名和时间戳
	if err = c.validateAuthorizationEvent(&baseEvent); err != nil {
		c.logger.Error(fmt.Sprintf("授权事件验证失败: %v", err))
		return "success", nil // 即使验证失败也返回success
	}

	// 根据事件类型进行处理
	switch baseEvent.InfoType {
	case "authorized":
		var event AuthorizedEvent
		err = xml.Unmarshal(xmlData, &event)
		if err != nil {
			c.logger.Error(fmt.Sprintf("解析授权成功事件失败: %v", err))
			break
		}
		c.logger.Debug(fmt.Sprintf("解析授权成功事件成功，事件内容: %+v", event))
		if err := c.GetEventHandler().HandleAuthorized(ctx, &event); err != nil {
			c.logger.Error(fmt.Sprintf("处理授权成功事件失败: %v", err))
		}

	case "unauthorized":
		var event UnauthorizedEvent
		err = xml.Unmarshal(xmlData, &event)
		if err != nil {
			c.logger.Error(fmt.Sprintf("解析取消授权事件失败: %v", err))
			break
		}
		c.logger.Debug(fmt.Sprintf("解析取消授权事件成功，事件内容: %+v", event))
		if err := c.GetEventHandler().HandleUnauthorized(ctx, &event); err != nil {
			c.logger.Error(fmt.Sprintf("处理取消授权事件失败: %v", err))
		}

	case "updateauthorized":
		var event UpdateAuthorizedEvent
		err := xml.Unmarshal(xmlData, &event)
		if err != nil {
			c.logger.Error(fmt.Sprintf("解析授权更新事件失败: %v", err))
			break
		}
		c.logger.Debug(fmt.Sprintf("解析授权更新事件成功，事件内容: %+v", event))
		if err = c.GetEventHandler().HandleUpdateAuthorized(ctx, &event); err != nil {
			c.logger.Error(fmt.Sprintf("处理授权更新事件失败: %v", err))
		}

	case "component_verify_ticket":
		var event ComponentVerifyTicketEvent
		err := xml.Unmarshal(xmlData, &event)
		if err != nil {
			c.logger.Error(fmt.Sprintf("解析验证票据事件失败: %v", err))
			// 根据微信官方文档要求，即使解析失败也必须返回success
			break
		}
		c.logger.Debug(fmt.Sprintf("解析验证票据事件成功，事件内容: %+v", event))
		// 存储验证票据
		if err := c.storage.SaveComponentVerifyTicket(ctx, event.ComponentVerifyTicket); err != nil {
			c.logger.Error(fmt.Sprintf("存储验证票据失败: %v", err))
			// 根据微信官方文档要求，即使存储失败也必须返回success
		}

		if err := c.GetEventHandler().HandleComponentVerifyTicket(ctx, &event); err != nil {
			c.logger.Error(fmt.Sprintf("处理验证票据事件失败: %v", err))
			// 根据微信官方文档要求，即使处理失败也必须返回success
		}

	case "encoding_aes_key_changed":
		var event EncodingAESKeyChangedEvent
		err := xml.Unmarshal(xmlData, &event)
		if err != nil {
			c.logger.Error(fmt.Sprintf("解析EncodingAESKey变更事件失败: %v", err))
			break
		}
		c.logger.Debug(fmt.Sprintf("解析EncodingAESKey变更事件成功，事件内容: %+v", event))
		// 保存上一次的EncodingAESKey
		if c.crypt != nil {
			err2 := c.crypt.SetPrevEncodingAESKey(c.config.EncodingAESKey)
			if err2 != nil {
				c.logger.Error(fmt.Sprintf("设置上一次EncodingAESKey失败: %v", err2))
				break
			}
		}

		// 更新配置中的EncodingAESKey
		c.config.EncodingAESKey = event.NewEncodingAESKey

		if err := c.GetEventHandler().HandleEncodingAESKeyChanged(ctx, &event); err != nil {
			c.logger.Error(fmt.Sprintf("处理EncodingAESKey变更事件失败: %v", err))
		}

	default:
		c.logger.Warn(fmt.Sprintf("收到未知的授权事件类型: %s", baseEvent.InfoType))
	}

	// 根据微信官方文档要求，必须返回"success"字符串
	return "success", nil
}

// validateAuthorizationEvent 验证授权事件
func (c *APIClient) validateAuthorizationEvent(event *AuthorizationEvent) error {
	// 验证AppID是否匹配
	if event.AppId != c.config.ComponentAppID {
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
func (c *APIClient) DecryptMessage(encryptedMsg, msgSignature, timestamp, nonce string) ([]byte, error) {
	// 验证消息签名
	if err := c.verifySignature(msgSignature, timestamp, nonce, encryptedMsg); err != nil {
		return nil, fmt.Errorf("消息签名验证失败: %v", err)
	}

	// 使用crypto包中的解密实现
	wxCrypt := crypto.NewWXBizMsgCrypt(c.config.ComponentToken, c.config.EncodingAESKey, c.config.ComponentAppID)

	// 解密消息
	decryptedMsg, err := wxCrypt.DecryptMsg(msgSignature, timestamp, nonce, encryptedMsg)
	if err != nil {
		return nil, fmt.Errorf("消息解密失败: %v", err)
	}

	return []byte(decryptedMsg), nil
}

// verifySignature 验证消息签名
func (c *APIClient) verifySignature(signature, timestamp, nonce, encryptedMsg string) error {
	// 根据微信开放平台签名算法验证签名
	// 签名算法：sha1(sort(token, timestamp, nonce, encryptedMsg))

	// 获取配置中的Token
	token := c.config.ComponentToken

	// 使用crypto包中的签名验证实现
	wxCrypt := crypto.NewWXBizMsgCrypt(token, c.config.EncodingAESKey, c.config.ComponentAppID)

	// 验证签名
	if !wxCrypt.VerifySignature(signature, timestamp, nonce, encryptedMsg) {
		return fmt.Errorf("签名验证失败")
	}

	return nil
}

// QueryAuth 使用授权码换取授权信息
func (c *APIClient) QueryAuth(ctx context.Context, authorizationCode string) (*QueryAuthResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := QueryAuthRequest{
		ComponentAppID:    c.config.ComponentAppID,
		AuthorizationCode: authorizationCode,
	}

	var result QueryAuthResponse
	queryAuthUrl := fmt.Sprintf("%s?component_access_token=%s", APIQueryAuthURL, url.QueryEscape(componentToken.AccessToken))
	err = c.MakeRequest(ctx, "POST", queryAuthUrl, request, &result)
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
func (c *APIClient) RefreshAuthorizerToken(ctx context.Context, authorizerAppID, refreshToken string) (*AuthorizationInfo, error) {
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
		APIResponse
		AuthorizationInfo
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token?component_access_token=%s",
		url.QueryEscape(componentToken.AccessToken))

	err = c.MakeRequest(ctx, "POST", apiURL, request, &result)
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
func (c *APIClient) GetAuthorizerInfo(ctx context.Context, authorizerAppID string) (*GetAuthorizerInfoResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := GetAuthorizerInfoRequest{
		ComponentAppID:  c.config.ComponentAppID,
		AuthorizerAppID: authorizerAppID,
	}

	var result GetAuthorizerInfoResponse
	getAuthorizerInfoUrl := fmt.Sprintf("%s?component_access_token=%s", APIGetAuthorizerInfoURL, url.QueryEscape(componentToken.AccessToken))
	err = c.MakeRequest(ctx, "POST", getAuthorizerInfoUrl, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetAuthorizerList 获取授权方列表
func (c *APIClient) GetAuthorizerList(ctx context.Context, offset, count int) (*GetAuthorizerListResponse, error) {
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
	getAuthorizerListUrl := fmt.Sprintf("%s?component_access_token=%s", APIGetAuthorizerListURL, url.QueryEscape(componentToken.AccessToken))
	err = c.MakeRequest(ctx, "POST", getAuthorizerListUrl, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GenerateAuthURL 生成授权链接
// authType: 授权类型 (1:手机端仅展示公众号, 2:仅展示小程序, 3:公众号和小程序都展示, 4:小程序推客账号, 5:视频号账号, 6:全部, 8:带货助手账号)
// platform: 平台类型 ("pc": PC端, "mobile": 移动端)
func (c *APIClient) GenerateAuthURL(preAuthCode string, authType int, bizAppID string, platform string) string {
	var baseURL string

	if platform == "mobile" {
		// 移动端授权页面URL
		baseURL = "https://open.weixin.qq.com/wxaopen/safe/bindcomponent"
	} else {
		// PC端授权页面URL
		baseURL = "https://mp.weixin.qq.com/cgi-bin/componentloginpage"
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
func (c *APIClient) GeneratePcAuthURL(preAuthCode string, authType int, bizAppID string) string {
	return c.GenerateAuthURL(preAuthCode, authType, bizAppID, "pc")
}

// GenerateMobileAuthURL 生成移动端授权链接
func (c *APIClient) GenerateMobileAuthURL(preAuthCode string, authType int, bizAppID string) string {
	return c.GenerateAuthURL(preAuthCode, authType, bizAppID, "mobile")
}

// ClearQuota 重置API调用次数
func (c *APIClient) ClearQuota(ctx context.Context) (*APIResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := ClearQuotaRequest{
		ComponentAppID: c.config.ComponentAppID,
	}

	var result APIResponse
	clearQuotaUrl := fmt.Sprintf("%s?access_token=%s", APIClearQuotaURL, url.QueryEscape(componentToken.AccessToken))
	err = c.MakeRequest(ctx, "POST", clearQuotaUrl, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result
	}

	return &result, nil
}

// GetApiQuota 查询API调用额度
func (c *APIClient) GetApiQuota(ctx context.Context, authorizerAppID string) (*GetApiQuotaResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := GetApiQuotaRequest{
		ComponentAppID:  c.config.ComponentAppID,
		AuthorizerAppID: authorizerAppID,
	}

	var result GetApiQuotaResponse
	getApiQuotaUrl := fmt.Sprintf("%s?access_token=%s", APIGetApiQuotaURL, url.QueryEscape(componentToken.AccessToken))
	err = c.MakeRequest(ctx, "POST", getApiQuotaUrl, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetRidInfo 查询rid信息
func (c *APIClient) GetRidInfo(ctx context.Context, rid string) (*GetRidInfoResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := GetRidInfoRequest{
		RID: rid,
	}

	var result GetRidInfoResponse
	getRidInfoUrl := fmt.Sprintf("%s?access_token=%s", APIGetRidInfoURL, url.QueryEscape(componentToken.AccessToken))
	err = c.MakeRequest(ctx, "POST", getRidInfoUrl, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// ClearComponentQuota 使用AppSecret重置第三方平台API调用次数
func (c *APIClient) ClearComponentQuota(ctx context.Context) (*APIResponse, error) {
	request := ClearComponentQuotaRequest{
		ComponentAppID:     c.config.ComponentAppID,
		ComponentAppSecret: c.config.ComponentAppSecret,
	}

	var result APIResponse
	err := c.MakeRequest(ctx, "POST", APIClearComponentQuotaURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result
	}

	return &result, nil
}

// SetAuthorizerOption 设置授权方选项信息
func (c *APIClient) SetAuthorizerOption(ctx context.Context, authorizerAppID, optionName, optionValue string) (*APIResponse, error) {
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

	var result APIResponse
	setAuthorizerOptionUrl := fmt.Sprintf("%s?component_access_token=%s", APISetAuthorizerOptionURL, url.QueryEscape(componentToken.AccessToken))
	err = c.MakeRequest(ctx, "POST", setAuthorizerOptionUrl, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result
	}

	return &result, nil
}

// GetAuthorizerOption 获取授权方选项信息
func (c *APIClient) GetAuthorizerOption(ctx context.Context, authorizerAppID, optionName string) (*GetAuthorizerOptionResponse, error) {
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
	getAuthorizerOptionUrl := fmt.Sprintf("%s?component_access_token=%s", APIGetAuthorizerOptionURL, url.QueryEscape(componentToken.AccessToken))
	err = c.MakeRequest(ctx, "POST", getAuthorizerOptionUrl, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetTemplateDraftList 获取草稿箱列表
func (c *APIClient) GetTemplateDraftList(ctx context.Context) (*GetTemplateDraftListResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	var result GetTemplateDraftListResponse
	getTemplateDraftListUrl := fmt.Sprintf("%s?access_token=%s", APIGetTemplateDraftListURL, url.QueryEscape(componentToken.AccessToken))
	err = c.MakeRequest(ctx, "GET", getTemplateDraftListUrl, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// AddToTemplate 将草稿添加到模板库
func (c *APIClient) AddToTemplate(ctx context.Context, draftID int64, templateType int) (*APIResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := AddToTemplateRequest{
		DraftID:      draftID,
		TemplateType: templateType,
	}

	var result APIResponse
	addToTemplateUrl := fmt.Sprintf("%s?access_token=%s", APIAddToTemplateURL, url.QueryEscape(componentToken.AccessToken))
	err = c.MakeRequest(ctx, "POST", addToTemplateUrl, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result
	}

	return &result, nil
}

// GetTemplateList 获取模板列表
func (c *APIClient) GetTemplateList(ctx context.Context) (*GetTemplateListResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	var result GetTemplateListResponse
	getTemplateListUrl := fmt.Sprintf("%s?access_token=%s", APIGetTemplateListURL, url.QueryEscape(componentToken.AccessToken))
	err = c.MakeRequest(ctx, "GET", getTemplateListUrl, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// DeleteTemplate 删除代码模板
func (c *APIClient) DeleteTemplate(ctx context.Context, templateID int64) (*APIResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := DeleteTemplateRequest{
		TemplateID: templateID,
	}

	var result APIResponse
	deleteTemplateUrl := fmt.Sprintf("%s?access_token=%s", APIDeleteTemplateURL, url.QueryEscape(componentToken.AccessToken))
	err = c.MakeRequest(ctx, "POST", deleteTemplateUrl, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result
	}

	return &result, nil
}
