package openplatform

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jcbowen/wego/officialaccount"
)

// AuthClient 授权相关客户端
type AuthClient struct {
	client *APIClient
}

// NewAuthClient 创建授权客户端
func NewAuthClient(client *APIClient) *AuthClient {
	return &AuthClient{
		client: client,
	}
}

// AuthorizerClient 授权方API客户端
type AuthorizerClient struct {
	authClient      *AuthClient
	authorizerAppID string
}

// NewAuthorizerClient 创建授权方API客户端
func (c *AuthClient) NewAuthorizerClient(authorizerAppID string) *AuthorizerClient {
	return &AuthorizerClient{
		authClient:      c,
		authorizerAppID: authorizerAppID,
	}
}

// TextMessage 文本消息
type TextMessage struct {
	Content string `json:"content"`
}

// ImageMessage 图片消息
type ImageMessage struct {
	MediaID string `json:"media_id"`
}

// VoiceMessage 语音消息
type VoiceMessage struct {
	MediaID string `json:"media_id"`
}

// VideoMessage 视频消息
type VideoMessage struct {
	MediaID     string `json:"media_id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// MusicMessage 音乐消息
type MusicMessage struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	MusicURL     string `json:"musicurl"`
	HQMusicURL   string `json:"hqmusicurl"`
	ThumbMediaID string `json:"thumb_media_id"`
}

// NewsMessage 图文消息
type NewsMessage struct {
	Articles []Article `json:"articles"`
}

// Article 图文消息文章
type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	PicURL      string `json:"picurl"`
}

// SendCustomMessage 发送客服消息
func (c *AuthorizerClient) SendCustomMessage(ctx context.Context, toUser string, message interface{}) error {
	accessToken, err := c.authClient.client.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"touser":  toUser,
		"msgtype": c.getMessageType(message),
	}

	// 根据消息类型设置对应的字段
	switch msg := message.(type) {
	case *TextMessage:
		request["text"] = map[string]string{"content": msg.Content}
	case *ImageMessage:
		request["image"] = map[string]string{"media_id": msg.MediaID}
	default:
		return fmt.Errorf("不支持的客服消息类型")
	}

	var result APIResponse
	err = c.authClient.client.req.Make(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// getMessageType 获取消息类型
func (c *AuthorizerClient) getMessageType(message interface{}) string {
	switch message.(type) {
	case *TextMessage:
		return "text"
	case *ImageMessage:
		return "image"
	case *VoiceMessage:
		return "voice"
	case *VideoMessage:
		return "video"
	case *MusicMessage:
		return "music"
	case *NewsMessage:
		return "news"
	default:
		return ""
	}
}

// Menu 菜单结构
type Menu struct {
	Button []Button `json:"button"`
}

// Button 菜单按钮
type Button struct {
	Type      string   `json:"type,omitempty"`
	Name      string   `json:"name"`
	Key       string   `json:"key,omitempty"`
	URL       string   `json:"url,omitempty"`
	SubButton []Button `json:"sub_button,omitempty"`
}

// CreateMenu 创建自定义菜单
func (c *AuthorizerClient) CreateMenu(ctx context.Context, menu *Menu) error {
	accessToken, err := c.authClient.client.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("%s?access_token=%s", officialaccount.APICreateMenuURL, url.QueryEscape(accessToken))

	var result APIResponse
	err = c.authClient.client.req.Make(ctx, "POST", apiURL, menu, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// GetMenu 获取菜单
func (c *AuthorizerClient) GetMenu(ctx context.Context) (*Menu, error) {
	accessToken, err := c.authClient.client.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s?access_token=%s", officialaccount.APIGetMenuURL, url.QueryEscape(accessToken))

	var result struct {
		APIResponse
		Menu Menu `json:"menu"`
	}

	err = c.authClient.client.req.Make(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result.Menu, nil
}

// DeleteMenu 删除自定义菜单
func (c *AuthorizerClient) DeleteMenu(ctx context.Context) error {
	accessToken, err := c.authClient.client.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("%s?access_token=%s", officialaccount.APIDeleteMenuURL, url.QueryEscape(accessToken))

	var result APIResponse
	err = c.authClient.client.req.Make(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// CallAPI 代调用API（支持context）
func (c *AuthorizerClient) CallAPI(ctx context.Context, apiURL string, params interface{}) ([]byte, error) {
	// 1. 获取授权方AccessToken
	token, err := c.authClient.client.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, fmt.Errorf("获取AccessToken失败: %v", err)
	}

	// 2. 构造完整URL
	fullURL := fmt.Sprintf("%s?access_token=%s", apiURL, token)

	// 3. 创建HTTP请求
	var req *http.Request
	if params != nil {
		// POST请求
		jsonData, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("参数序列化失败: %v", err)
		}

		req, err = http.NewRequestWithContext(ctx, "POST", fullURL, strings.NewReader(string(jsonData)))
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		// GET请求
		req, err = http.NewRequestWithContext(ctx, "GET", fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %v", err)
		}
	}

	// 4. 发送HTTP请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API调用失败: %v", err)
	}
	defer resp.Body.Close()

	// 5. 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 6. 解析响应
	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("响应解析失败: %v", err)
	}

	// 7. 检查错误码
	if apiResp.ErrCode != 0 {
		return nil, fmt.Errorf("API返回错误: %d - %s", apiResp.ErrCode, apiResp.ErrMsg)
	}

	return respBody, nil
}

// CallAPIWithQuery 支持查询参数的API调用
func (c *AuthorizerClient) CallAPIWithQuery(ctx context.Context, baseURL string, queryParams map[string]string, postData interface{}) ([]byte, error) {
	// 获取授权方AccessToken
	token, err := c.authClient.client.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, fmt.Errorf("获取AccessToken失败: %v", err)
	}

	// 构造查询参数
	params := url.Values{}
	params.Set("access_token", token)
	for key, value := range queryParams {
		params.Set(key, value)
	}

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 根据是否有postData决定请求方法
	if postData != nil {
		return c.CallAPI(ctx, fullURL, postData)
	}
	return c.CallAPI(ctx, fullURL, nil)
}

// SendTextMessage 发送文本消息
func (c *AuthorizerClient) SendTextMessage(ctx context.Context, toUser, content string) error {
	return c.SendCustomMessage(ctx, toUser, &TextMessage{Content: content})
}

// SendImageMessage 发送图片消息
func (c *AuthorizerClient) SendImageMessage(ctx context.Context, toUser, mediaID string) error {
	return c.SendCustomMessage(ctx, toUser, &ImageMessage{MediaID: mediaID})
}

// SendNewsMessage 发送图文消息
func (c *AuthorizerClient) SendNewsMessage(ctx context.Context, toUser string, articles []Article) error {
	message := map[string]interface{}{
		"touser":  toUser,
		"msgtype": "news",
		"news": map[string]interface{}{
			"articles": articles,
		},
	}

	apiURL := officialaccount.APIMessageCustomSendURL

	_, err := c.CallAPI(ctx, apiURL, message)
	return err
}

// ConditionalMenu 个性化菜单
type ConditionalMenu struct {
	Button    []Button `json:"button"`
	MatchRule struct {
		TagID              string `json:"tag_id,omitempty"`
		Sex                string `json:"sex,omitempty"`
		Country            string `json:"country,omitempty"`
		Province           string `json:"province,omitempty"`
		City               string `json:"city,omitempty"`
		ClientPlatformType string `json:"client_platform_type,omitempty"`
		Language           string `json:"language,omitempty"`
	} `json:"matchrule"`
}

// CreateConditionalMenu 创建个性化菜单
func (c *AuthorizerClient) CreateConditionalMenu(ctx context.Context, menu *ConditionalMenu) error {
	apiURL := "https://api.weixin.qq.com/cgi-bin/menu/addconditional"

	_, err := c.CallAPI(ctx, apiURL, menu)
	return err
}

// DeleteConditionalMenu 删除个性化菜单
func (c *AuthorizerClient) DeleteConditionalMenu(ctx context.Context, menuID string) error {
	apiURL := "https://api.weixin.qq.com/cgi-bin/menu/delconditional"

	params := map[string]interface{}{
		"menuid": menuID,
	}

	_, err := c.CallAPI(ctx, apiURL, params)
	return err
}

// CallAPIWithRetry 带重试的API调用
func (c *AuthorizerClient) CallAPIWithRetry(ctx context.Context, apiURL string, params interface{}, maxRetries int) ([]byte, error) {
	for i := 0; i < maxRetries; i++ {
		resp, err := c.CallAPI(ctx, apiURL, params)
		if err == nil {
			return resp, nil
		}

		// 检查是否是Token过期错误
		if strings.Contains(err.Error(), "40001") || strings.Contains(err.Error(), "42001") {
			// Token过期，清除缓存并重试
			// 这里需要实现缓存清除逻辑
			continue
		}

		// 其他错误，直接返回
		return nil, err
	}

	return nil, fmt.Errorf("API调用重试%d次后失败", maxRetries)
}

// JSSDKCacher JS-SDK配置缓存器
type JSSDKCacher struct {
	cache sync.Map
}

// GetCachedConfig 获取缓存的JS-SDK配置
func (c *JSSDKCacher) GetCachedConfig(url string, jsAPIList []string) (*JSSDKConfig, error) {
	cacheKey := c.generateCacheKey(url, jsAPIList)

	if cached, exists := c.cache.Load(cacheKey); exists {
		if config, ok := cached.(*JSSDKConfig); ok {
			return config, nil
		}
	}

	return nil, fmt.Errorf("缓存未命中")
}

// CacheConfig 缓存JS-SDK配置
func (c *JSSDKCacher) CacheConfig(url string, jsAPIList []string, config *JSSDKConfig) {
	cacheKey := c.generateCacheKey(url, jsAPIList)
	c.cache.Store(cacheKey, config)
}

// generateCacheKey 生成缓存键
func (c *JSSDKCacher) generateCacheKey(url string, jsAPIList []string) string {
	sortedAPIs := make([]string, len(jsAPIList))
	copy(sortedAPIs, jsAPIList)
	sort.Strings(sortedAPIs)

	keyData := url + "|" + strings.Join(sortedAPIs, ",")

	hash := sha1.Sum([]byte(keyData))
	return hex.EncodeToString(hash[:])
}

// GetConfigOptimized 优化后的JS-SDK配置获取
func (jm *JSSDKManager) GetConfigOptimized(url string, jsAPIList []string) (*JSSDKConfig, error) {
	// 验证参数
	if url == "" {
		return nil, fmt.Errorf("URL不能为空")
	}
	if len(jsAPIList) == 0 {
		return nil, fmt.Errorf("JSAPI列表不能为空")
	}
	if jm.authorizerClient.authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	// 先尝试从缓存获取
	if cachedConfig, err := jm.cacher.GetCachedConfig(url, jsAPIList); err == nil && cachedConfig != nil {
		return cachedConfig, nil
	}

	// 缓存未命中，重新生成配置
	ctx := context.Background()
	accessToken, err := jm.authorizerClient.authClient.client.GetAuthorizerAccessToken(ctx, jm.authorizerClient.authorizerAppID)
	if err != nil {
		return nil, fmt.Errorf("获取AccessToken失败: %v", err)
	}

	ticket, err := jm.getJSAPITicket(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("获取JSAPI Ticket失败: %v", err)
	}

	config := jm.generateSignature(url, ticket, jsAPIList)

	// 缓存新生成的配置
	jm.cacher.CacheConfig(url, jsAPIList, config)

	return config, nil
}

// GetConfigWithRetry 带重试的JS-SDK配置获取
func (jm *JSSDKManager) GetConfigWithRetry(url string, jsAPIList []string, maxRetries int) (*JSSDKConfig, error) {
	// 验证参数
	if url == "" {
		return nil, fmt.Errorf("URL不能为空")
	}
	if len(jsAPIList) == 0 {
		return nil, fmt.Errorf("JSAPI列表不能为空")
	}
	if jm.authorizerClient.authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}
	if maxRetries <= 0 {
		return nil, fmt.Errorf("重试次数必须大于0")
	}

	var lastErr error

	for i := 0; i < maxRetries; i++ {
		config, err := jm.GetConfigOptimized(url, jsAPIList)
		if err == nil {
			return config, nil
		}

		lastErr = err
		time.Sleep(time.Duration(i+1) * time.Second) // 指数退避
	}

	return nil, fmt.Errorf("获取JS-SDK配置失败，重试%d次后仍然失败: %v", maxRetries, lastErr)
}

// validateURL 验证URL格式
func (jm *JSSDKManager) validateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL不能为空")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL必须以http://或https://开头")
	}

	return nil
}

// UserInfo 用户信息
type UserInfo struct {
	APIResponse
	Subscribe      int    `json:"subscribe"`
	OpenID         string `json:"openid"`
	Nickname       string `json:"nickname"`
	Sex            int    `json:"sex"`
	Language       string `json:"language"`
	City           string `json:"city"`
	Province       string `json:"province"`
	Country        string `json:"country"`
	HeadImgURL     string `json:"headimgurl"`
	SubscribeTime  int64  `json:"subscribe_time"`
	UnionID        string `json:"unionid,omitempty"`
	Remark         string `json:"remark"`
	GroupID        int    `json:"groupid"`
	TagIDList      []int  `json:"tagid_list"`
	SubscribeScene string `json:"subscribe_scene"`
	QRScene        int    `json:"qr_scene"`
	QRSceneStr     string `json:"qr_scene_str"`
}

// GetUserInfo 获取用户信息
func (c *AuthorizerClient) GetUserInfo(ctx context.Context, openID string) (*UserInfo, error) {
	// 验证参数
	if openID == "" {
		return nil, fmt.Errorf("用户OpenID不能为空")
	}
	if c.authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	accessToken, err := c.authClient.client.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN",
		url.QueryEscape(accessToken), url.QueryEscape(openID))

	var result UserInfo
	err = c.authClient.client.req.Make(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// UserList 用户列表
type UserList struct {
	APIResponse
	Total int `json:"total"`
	Count int `json:"count"`
	Data  struct {
		OpenID []string `json:"openid"`
	} `json:"data"`
	NextOpenID string `json:"next_openid"`
}

// GetUserList 获取用户列表
func (c *AuthorizerClient) GetUserList(ctx context.Context, nextOpenID string) (*UserList, error) {
	accessToken, err := c.authClient.client.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/user/get?access_token=%s", url.QueryEscape(accessToken))
	if nextOpenID != "" {
		apiURL += "&next_openid=" + url.QueryEscape(nextOpenID)
	}

	var result UserList
	err = c.authClient.client.req.Make(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// SendTemplateMessage 发送模板消息
func (c *AuthorizerClient) SendTemplateMessage(ctx context.Context, template *officialaccount.TemplateMessageRequest) (*officialaccount.SendTemplateMessageResponse, error) {
	if c.authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	accessToken, err := c.authClient.client.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	return officialaccount.NewTemplate(c.authClient.client.req).SendTemplateMessage(ctx, template, accessToken)
}

// MediaResponse 媒体文件响应
type MediaResponse struct {
	APIResponse
	Type      string `json:"type"`
	MediaID   string `json:"media_id"`
	CreatedAt int64  `json:"created_at"`
}

// UploadMedia 上传临时素材
func (c *AuthorizerClient) UploadMedia(ctx context.Context, mediaType, filename string, data []byte) (*MediaResponse, error) {
	// 验证参数
	if mediaType == "" {
		return nil, fmt.Errorf("媒体类型不能为空")
	}
	if filename == "" {
		return nil, fmt.Errorf("文件名不能为空")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("媒体数据不能为空")
	}
	if c.authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	accessToken, err := c.authClient.client.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/media/upload?access_token=%s&type=%s",
		url.QueryEscape(accessToken), url.QueryEscape(mediaType))

	// 创建multipart/form-data请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建文件字段
	part, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return nil, fmt.Errorf("创建文件字段失败: %v", err)
	}

	// 写入文件数据
	if _, err := part.Write(data); err != nil {
		return nil, fmt.Errorf("写入文件数据失败: %v", err)
	}

	// 关闭writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("关闭multipart writer失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	resp, err := c.authClient.client.req.MakeRaw(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result MediaResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetMedia 获取临时素材
func (c *AuthorizerClient) GetMedia(ctx context.Context, mediaID string) ([]byte, error) {
	// 验证参数
	if mediaID == "" {
		return nil, fmt.Errorf("媒体ID不能为空")
	}
	if c.authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	accessToken, err := c.authClient.client.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/media/get?access_token=%s&media_id=%s",
		url.QueryEscape(accessToken), url.QueryEscape(mediaID))

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 发送请求
	resp, err := c.authClient.client.req.MakeRaw(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应类型
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// 如果是JSON响应，说明有错误
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %v", err)
		}

		var result APIResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("解析响应失败: %v", err)
		}

		if !result.IsSuccess() {
			return nil, &result
		}
	}

	// 读取媒体文件数据
	mediaData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取媒体文件失败: %v", err)
	}

	return mediaData, nil
}

// OAuthClient 网页授权客户端
type OAuthClient struct {
	authorizerClient *AuthorizerClient
	redirectURI      string
}

// GetOAuthClient 创建OAuth客户端
func (c *AuthorizerClient) GetOAuthClient(redirectURI string) *OAuthClient {
	return &OAuthClient{
		authorizerClient: c,
		redirectURI:      redirectURI,
	}
}

// GetAuthorizeURL 生成授权页面URL
func (oc *OAuthClient) GetAuthorizeURL(scope, state string) string {
	// 验证参数
	if oc.authorizerClient.authorizerAppID == "" {
		panic("authorizerAppID不能为空")
	}
	if oc.redirectURI == "" {
		panic("redirectURI不能为空")
	}
	if scope == "" {
		panic("scope不能为空")
	}

	// 严格按照微信官方文档要求的参数顺序和大小写
	params := url.Values{}
	params.Set("appid", oc.authorizerClient.authorizerAppID)
	params.Set("redirect_uri", oc.redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", scope)
	if state != "" {
		params.Set("state", state)
	}
	params.Set("component_appid", oc.authorizerClient.authClient.client.GetConfig().ComponentAppID)

	// 使用微信官方文档指定的授权URL
	return "https://open.weixin.qq.com/connect/oauth2/authorize?" + params.Encode() + "#wechat_redirect"
}

// GetBaseAuthorizeURL 生成静默授权URL
func (oc *OAuthClient) GetBaseAuthorizeURL(state string) string {
	return oc.GetAuthorizeURL("snsapi_base", state)
}

// GetUserInfoAuthorizeURL 生成用户信息授权URL
func (oc *OAuthClient) GetUserInfoAuthorizeURL(state string) string {
	return oc.GetAuthorizeURL("snsapi_userinfo", state)
}

// OAuthToken 网页授权Token
type OAuthToken struct {
	APIResponse
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
}

// GetAccessToken 使用授权码获取AccessToken
func (oc *OAuthClient) GetAccessToken(ctx context.Context, code string) (*OAuthToken, error) {
	// 验证参数
	if code == "" {
		return nil, fmt.Errorf("授权码不能为空")
	}
	if oc.authorizerClient.authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	// 需要先获取组件令牌
	componentToken, err := oc.authorizerClient.authClient.client.GetComponentToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取ComponentAccessToken失败: %v", err)
	}

	// 使用微信官方文档指定的URL
	apiURL := "https://api.weixin.qq.com/sns/oauth2/component/access_token"

	// 严格按照微信官方文档要求的参数格式
	params := map[string]interface{}{
		"appid":                  oc.authorizerClient.authorizerAppID,
		"code":                   code,
		"grant_type":             "authorization_code",
		"component_appid":        oc.authorizerClient.authClient.client.GetConfig().ComponentAppID,
		"component_access_token": componentToken.AccessToken,
	}

	var oauthToken OAuthToken
	err = oc.authorizerClient.authClient.client.req.Make(ctx, "POST", apiURL, params, &oauthToken)
	if err != nil {
		return nil, err
	}

	if !oauthToken.IsSuccess() {
		return nil, &oauthToken.APIResponse
	}

	return &oauthToken, nil
}

// JSSDKManager JS-SDK管理器
type JSSDKManager struct {
	authorizerClient *AuthorizerClient
	cacher           *JSSDKCacher
}

// GetJSSDKManager 创建JS-SDK管理器
func (c *AuthorizerClient) GetJSSDKManager() *JSSDKManager {
	return &JSSDKManager{
		authorizerClient: c,
		cacher:           &JSSDKCacher{},
	}
}

// JSSDKConfig JS-SDK配置结构
type JSSDKConfig struct {
	AppID     string   `json:"appId" ini:"app_id"`          // 公众号appid
	Timestamp int64    `json:"timestamp" ini:"timestamp"`   // 时间戳
	NonceStr  string   `json:"nonceStr" ini:"nonce_str"`    // 随机字符串
	Signature string   `json:"signature" ini:"signature"`   // 签名
	JSAPIList []string `json:"jsApiList" ini:"js_api_list"` // JS-SDK调用权限列表
}

// GetConfig 生成JS-SDK配置
func (jm *JSSDKManager) GetConfig(ctx context.Context, url string, jsAPIList []string) (*JSSDKConfig, error) {
	// 验证参数
	if url == "" {
		return nil, fmt.Errorf("URL不能为空")
	}
	if len(jsAPIList) == 0 {
		return nil, fmt.Errorf("JSAPI列表不能为空")
	}
	if jm.authorizerClient.authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	// 获取授权方AccessToken
	accessToken, err := jm.authorizerClient.authClient.client.GetAuthorizerAccessToken(ctx, jm.authorizerClient.authorizerAppID)
	if err != nil {
		return nil, fmt.Errorf("获取AccessToken失败: %v", err)
	}

	// 获取JSAPI Ticket
	ticket, err := jm.getJSAPITicket(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("获取JSAPI Ticket失败: %v", err)
	}

	// 生成签名
	config := jm.generateSignature(url, ticket, jsAPIList)

	return config, nil
}

// getJSAPITicket 获取JSAPI Ticket
func (jm *JSSDKManager) getJSAPITicket(ctx context.Context, accessToken string) (string, error) {
	apiURL := "https://api.weixin.qq.com/cgi-bin/ticket/getticket"

	params := map[string]interface{}{
		"access_token": accessToken,
		"type":         "jsapi",
	}

	var result struct {
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
		Ticket    string `json:"ticket"`
		ExpiresIn int    `json:"expires_in"`
	}

	err := jm.authorizerClient.authClient.client.req.Make(ctx, "GET", apiURL, params, &result)
	if err != nil {
		return "", err
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("获取JSAPI Ticket失败: %s", result.ErrMsg)
	}

	return result.Ticket, nil
}

// generateSignature 生成签名
func (jm *JSSDKManager) generateSignature(url, ticket string, jsAPIList []string) *JSSDKConfig {
	nonceStr := generateNonceStr()
	timestamp := time.Now().Unix()

	// 签名算法
	signStr := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%d&url=%s",
		ticket, nonceStr, timestamp, url)

	signature := sha1.Sum([]byte(signStr))
	signatureHex := hex.EncodeToString(signature[:])

	return &JSSDKConfig{
		AppID:     jm.authorizerClient.authorizerAppID,
		Timestamp: timestamp,
		NonceStr:  nonceStr,
		Signature: signatureHex,
		JSAPIList: jsAPIList,
	}
}

// generateNonceStr 生成随机字符串
func generateNonceStr() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bts := make([]byte, 16)
	for i := range bts {
		bts[i] = letters[rand.Intn(len(letters))]
	}
	return string(bts)
}

// QRCodeRequest 二维码请求
type QRCodeRequest struct {
	ExpireSeconds int64  `json:"expire_seconds,omitempty"`
	ActionName    string `json:"action_name"`
	ActionInfo    struct {
		Scene struct {
			SceneID  int    `json:"scene_id,omitempty"`
			SceneStr string `json:"scene_str,omitempty"`
		} `json:"scene"`
	} `json:"action_info"`
}

// QRCodeResponse 二维码响应
type QRCodeResponse struct {
	APIResponse
	Ticket        string `json:"ticket"`
	ExpireSeconds int64  `json:"expire_seconds"`
	URL           string `json:"url"`
}

// CreateQRCode 创建二维码
func (c *AuthorizerClient) CreateQRCode(ctx context.Context, qrCode *QRCodeRequest) (*QRCodeResponse, error) {
	// 验证参数
	if qrCode == nil {
		return nil, fmt.Errorf("二维码请求不能为空")
	}
	if qrCode.ActionName == "" {
		return nil, fmt.Errorf("二维码动作名称不能为空")
	}
	if c.authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	accessToken, err := c.authClient.client.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token=%s", url.QueryEscape(accessToken))

	var result QRCodeResponse
	err = c.authClient.client.req.Make(ctx, "POST", apiURL, qrCode, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetQRCodeURL 获取二维码图片URL
func (c *AuthorizerClient) GetQRCodeURL(ticket string) (string, error) {
	// 验证参数
	if ticket == "" {
		return "", fmt.Errorf("二维码ticket不能为空")
	}

	return fmt.Sprintf("https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=%s", url.QueryEscape(ticket)), nil
}

// MiniProgramClient 小程序客户端
type MiniProgramClient struct {
	authorizerClient *AuthorizerClient
}

// GetMiniProgramClient 创建小程序客户端
func (c *AuthorizerClient) GetMiniProgramClient() *MiniProgramClient {
	return &MiniProgramClient{
		authorizerClient: c,
	}
}

// WXACodeRequest 小程序码请求
type WXACodeRequest struct {
	Path  string `json:"path"`
	Width int    `json:"width,omitempty"`
}

// WXACodeResponse 小程序码响应
type WXACodeResponse struct {
	APIResponse
	ContentType string `json:"contentType"`
	Buffer      []byte `json:"buffer"`
}

// GetWXACode 获取小程序码
func (mpc *MiniProgramClient) GetWXACode(ctx context.Context, request *WXACodeRequest) (*WXACodeResponse, error) {
	// 验证参数
	if request == nil {
		return nil, fmt.Errorf("小程序码请求不能为空")
	}
	if request.Path == "" {
		return nil, fmt.Errorf("小程序码路径不能为空")
	}
	if mpc.authorizerClient.authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	accessToken, err := mpc.authorizerClient.authClient.client.GetAuthorizerAccessToken(ctx, mpc.authorizerClient.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/getwxacode?access_token=%s", url.QueryEscape(accessToken))

	var result WXACodeResponse
	err = mpc.authorizerClient.authClient.client.req.Make(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// OAuthUserInfo 网页授权用户信息
type OAuthUserInfo struct {
	APIResponse
	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string   `json:"unionid"`
}

// GetUserInfo 获取用户信息
func (oc *OAuthClient) GetUserInfo(ctx context.Context, accessToken, openID string) (*OAuthUserInfo, error) {
	// 验证参数
	if accessToken == "" {
		return nil, fmt.Errorf("accessToken不能为空")
	}
	if openID == "" {
		return nil, fmt.Errorf("openID不能为空")
	}

	// 使用微信官方文档指定的URL
	apiURL := "https://api.weixin.qq.com/sns/userinfo"

	// 严格按照微信官方文档要求的参数格式
	params := map[string]interface{}{
		"access_token": accessToken,
		"openid":       openID,
		"lang":         "zh_CN",
	}

	var userInfo OAuthUserInfo
	err := oc.authorizerClient.authClient.client.req.Make(ctx, "GET", apiURL, params, &userInfo)
	if err != nil {
		return nil, err
	}

	if !userInfo.IsSuccess() {
		return nil, &userInfo.APIResponse
	}

	return &userInfo, nil
}

// RefreshToken 刷新AccessToken
func (oc *OAuthClient) RefreshToken(ctx context.Context, refreshToken string) (*OAuthToken, error) {
	// 验证参数
	if refreshToken == "" {
		return nil, fmt.Errorf("refreshToken不能为空")
	}
	if oc.authorizerClient.authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	// 需要先获取组件令牌
	componentToken, err := oc.authorizerClient.authClient.client.GetComponentToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取ComponentAccessToken失败: %v", err)
	}

	// 使用微信官方文档指定的URL
	apiURL := "https://api.weixin.qq.com/sns/oauth2/component/refresh_token"

	// 严格按照微信官方文档要求的参数格式
	params := map[string]interface{}{
		"appid":                  oc.authorizerClient.authorizerAppID,
		"grant_type":             "refresh_token",
		"refresh_token":          refreshToken,
		"component_appid":        oc.authorizerClient.authClient.client.GetConfig().ComponentAppID,
		"component_access_token": componentToken.AccessToken,
	}

	var oauthToken OAuthToken
	err = oc.authorizerClient.authClient.client.req.Make(ctx, "POST", apiURL, params, &oauthToken)
	if err != nil {
		return nil, err
	}

	if !oauthToken.IsSuccess() {
		return nil, &oauthToken.APIResponse
	}

	return &oauthToken, nil
}
