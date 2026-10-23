// Package openplatform 微信开放平台API - 调用授权方接口
package openplatform

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/official_account"
)

// 标准错误定义
var (
	ErrInvalidURL        = errors.New("invalid URL format")
	ErrEmptyJSAPIList    = errors.New("JSAPI list cannot be empty")
	ErrEmptyAppID        = errors.New("authorizer AppID cannot be empty")
	ErrCacheMiss         = errors.New("cache miss or expired")
	ErrInvalidMaxRetries = errors.New("max retries must be greater than 0")
	ErrTicketFetchFailed = errors.New("failed to fetch JSAPI ticket")
	ErrAccessTokenFailed = errors.New("failed to get authorizer access token")
)

// AuthClient 授权相关客户端
type AuthClient struct {
	client *Client
}

// NewAuthClient 创建授权客户端
func NewAuthClient(client *Client) *AuthClient {
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
	MediaID      string `json:"media_id"`
	ThumbMediaID string `json:"thumb_media_id"` // 视频缩略图的媒体ID（官方要求必填）
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
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

	apiURL := fmt.Sprintf("%s?access_token=%s", official_account.URLMessageCustomSend, accessToken)

	request := map[string]interface{}{
		"touser":  toUser,
		"msgtype": c.getMessageType(message),
	}

	// 根据消息类型设置对应的字段（与官方一致：msgtype 在外层，内容体不含 msgtype）
	switch msg := message.(type) {
	case *TextMessage:
		request["text"] = map[string]string{"content": msg.Content}
	case *ImageMessage:
		request["image"] = map[string]string{"media_id": msg.MediaID}
	case *VoiceMessage:
		request["voice"] = map[string]string{"media_id": msg.MediaID}
	case *VideoMessage:
		request["video"] = map[string]string{
			"media_id":       msg.MediaID,
			"thumb_media_id": msg.ThumbMediaID,
			"title":          msg.Title,
			"description":    msg.Description,
		}
	case *MusicMessage:
		request["music"] = map[string]string{
			"title":          msg.Title,
			"description":    msg.Description,
			"musicurl":       msg.MusicURL,
			"hqmusicurl":     msg.HQMusicURL,
			"thumb_media_id": msg.ThumbMediaID,
		}
	case *NewsMessage:
		request["news"] = map[string]interface{}{"articles": msg.Articles}
	default:
		return fmt.Errorf("不支持的客服消息类型")
	}

	var result core.APIResponse
	err = c.authClient.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    apiURL,
		Body:   request,
		Result: &result,
	})
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

	apiURL := fmt.Sprintf("%s?access_token=%s", official_account.URLCreateMenu, accessToken)

	var result core.APIResponse
	err = c.authClient.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    apiURL,
		Body:   menu,
		Result: &result,
	})
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

	apiURL := fmt.Sprintf("%s?access_token=%s", official_account.URLGetMenu, accessToken)

	var result struct {
		core.APIResponse
		Menu Menu `json:"menu"`
	}

	err = c.authClient.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    apiURL,
		Result: &result,
	})
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

	apiURL := fmt.Sprintf("%s?access_token=%s", official_account.URLDeleteMenu, accessToken)

	var result core.APIResponse
	err = c.authClient.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    apiURL,
		Result: &result,
	})
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

	// 2. 构造完整URL（兼容调用方已携带 query 的场景，避免出现双 ? 破坏参数解析）
	fullURL := fmt.Sprintf("%s?access_token=%s", apiURL, token)
	if strings.Contains(apiURL, "?") {
		fullURL = fmt.Sprintf("%s&access_token=%s", apiURL, token)
	}

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

	// 4. 发送HTTP请求（使用注入的HTTP客户端，保证自定义超时/代理/埋点生效）
	resp, err := c.authClient.client.req.MakeRaw(req)
	if err != nil {
		return nil, fmt.Errorf("API调用失败: %v", err)
	}
	defer resp.Body.Close()

	// 5. 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 5.1 校验HTTP状态码，避免 500/302 被当作正常响应解析
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API调用失败 - 状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 6. 解析响应并统一返回错误类型
	var apiResp core.APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("响应解析失败: %v", err)
	}

	// 7. 检查错误码
	if !apiResp.IsSuccess() {
		return nil, fmt.Errorf("API返回错误: %d - %s", apiResp.ErrCode, apiResp.ErrMsg)
	}

	return respBody, nil
}

// CallAPIWithQuery 支持查询参数的API调用
//
// 说明: access_token 由 CallAPI 统一注入，此处仅拼接业务查询参数，
// 避免 access_token 被二次拼接导致参数污染（历史缺陷，已修复）。
func (c *AuthorizerClient) CallAPIWithQuery(ctx context.Context, baseURL string, queryParams map[string]string, postData interface{}) ([]byte, error) {
	// 构造业务查询参数（不含 access_token，由 CallAPI 统一注入）
	params := url.Values{}
	for key, value := range queryParams {
		params.Set(key, value)
	}

	fullURL := baseURL
	if encoded := params.Encode(); encoded != "" {
		fullURL = fmt.Sprintf("%s?%s", baseURL, encoded)
	}

	// 根据是否有postData决定请求方法
	return c.CallAPI(ctx, fullURL, postData)
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

	apiURL := official_account.URLMessageCustomSend

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
	apiURL := official_account.URLAddConditionalMenu

	_, err := c.CallAPI(ctx, apiURL, menu)
	return err
}

// DeleteConditionalMenu 删除个性化菜单
func (c *AuthorizerClient) DeleteConditionalMenu(ctx context.Context, menuID string) error {
	apiURL := official_account.URLDeleteConditionalMenu

	params := map[string]interface{}{
		"menuid": menuID,
	}

	_, err := c.CallAPI(ctx, apiURL, params)
	return err
}

// CallAPIWithRetry 带重试的API调用
// 遇到 40001（access_token 无效）/ 42001（access_token 过期）时会将本地缓存的
// 授权方 access_token 置为过期（保留 refresh_token）并重新获取后重试；其他错误直接返回
func (c *AuthorizerClient) CallAPIWithRetry(ctx context.Context, apiURL string, params interface{}, maxRetries int) ([]byte, error) {
	for i := 0; i < maxRetries; i++ {
		resp, err := c.CallAPI(ctx, apiURL, params)
		if err == nil {
			return resp, nil
		}

		// 检查是否是Token过期/无效错误
		if strings.Contains(err.Error(), "40001") || strings.Contains(err.Error(), "42001") {
			// Token过期/无效：仅失效 access_token，强制下次调用走 refresh 流程；
			// 必须保留 refresh_token（删除整条记录会导致 refresh_token 丢失，代调用将永久不可用）
			c.invalidateAuthorizerAccessToken(ctx)
			continue
		}

		// 其他错误，直接返回
		return nil, err
	}

	return nil, fmt.Errorf("API调用重试%d次后失败", maxRetries)
}

// invalidateAuthorizerAccessToken 将授权方 access_token 置为过期（保留 refresh_token）
//
// 用于 40001/42001 后强制下次调用走 refresh_token 刷新流程；
// 若缓存记录不存在则不做任何处理（下次调用会因缺少 token 自然报错）。
func (c *AuthorizerClient) invalidateAuthorizerAccessToken(ctx context.Context) {
	token, err := c.authClient.client.storage.GetAuthorizerToken(ctx, c.authorizerAppID)
	if err != nil {
		c.authClient.client.logger.Warn("读取授权方token缓存失败", map[string]interface{}{
			"authorizer_appid": c.authorizerAppID,
			"error":            err.Error(),
		})
		return
	}
	if token == nil {
		return
	}

	token.AuthorizerAccessToken = ""
	token.ExpiresAt = time.Time{}
	if err := c.authClient.client.storage.SaveAuthorizerToken(ctx, c.authorizerAppID, token); err != nil {
		c.authClient.client.logger.Warn("失效授权方token缓存失败", map[string]interface{}{
			"authorizer_appid": c.authorizerAppID,
			"error":            err.Error(),
		})
	}
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Hits       int64 // 缓存命中次数
	Misses     int64 // 缓存未命中次数
	Evictions  int64 // 缓存淘汰次数
	TotalItems int64 // 总缓存项数
}

// JSSDKCacher JS-SDK配置缓存器
type JSSDKCacher struct {
	cache sync.Map
	stats CacheStats
	mu    sync.RWMutex
}

// isJSSDKConfigValid 判断缓存的 JS-SDK 配置是否仍有效
// 优先按 ticket 实际过期时间判断；历史缓存缺失该值时回退固定 defaultTicketExpiresIn 秒
func isJSSDKConfigValid(config *JSSDKConfig) bool {
	if !config.expiresAt.IsZero() {
		return time.Now().Before(config.expiresAt)
	}
	return time.Now().Unix()-config.Timestamp < defaultTicketExpiresIn
}

// GetCachedConfig 获取缓存的JS-SDK配置 - 带统计功能
func (c *JSSDKCacher) GetCachedConfig(url string, jsAPIList []string) (*JSSDKConfig, error) {
	cacheKey := c.generateCacheKey(url, jsAPIList)

	if cached, exists := c.cache.Load(cacheKey); exists {
		if config, ok := cached.(*JSSDKConfig); ok {
			// 按 ticket 实际过期时间判断是否有效
			if isJSSDKConfigValid(config) {
				c.mu.Lock()
				c.stats.Hits++
				c.mu.Unlock()
				return config, nil
			}
		}
	}

	c.mu.Lock()
	c.stats.Misses++
	c.mu.Unlock()
	return nil, fmt.Errorf("%w: 缓存未命中或已过期", ErrCacheMiss)
}

// CacheConfig 缓存JS-SDK配置 - 带统计功能
func (c *JSSDKCacher) CacheConfig(url string, jsAPIList []string, config *JSSDKConfig) {
	cacheKey := c.generateCacheKey(url, jsAPIList)
	c.cache.Store(cacheKey, config)

	c.mu.Lock()
	c.stats.TotalItems++
	c.mu.Unlock()
}

// GetStats 获取缓存统计信息
func (c *JSSDKCacher) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// CleanupExpired 清理过期缓存
func (c *JSSDKCacher) CleanupExpired() {
	expiredCount := 0

	c.cache.Range(func(key, value interface{}) bool {
		if config, ok := value.(*JSSDKConfig); ok {
			// 按 ticket 实际过期时间判断是否过期（缺失该值时回退固定 7200 秒）
			if !isJSSDKConfigValid(config) {
				c.cache.Delete(key)
				expiredCount++
			}
		}
		return true
	})

	c.mu.Lock()
	c.stats.Evictions += int64(expiredCount)
	c.stats.TotalItems -= int64(expiredCount)
	if c.stats.TotalItems < 0 {
		c.stats.TotalItems = 0
	}
	c.mu.Unlock()
}

// generateCacheKey 生成缓存键 - 性能优化版本
func (c *JSSDKCacher) generateCacheKey(url string, jsAPIList []string) string {
	var builder strings.Builder
	builder.Grow(len(url) + len(jsAPIList)*20 + 1)

	builder.WriteString(url)
	builder.WriteString("|")

	sortedAPIs := make([]string, len(jsAPIList))
	copy(sortedAPIs, jsAPIList)
	sort.Strings(sortedAPIs)

	for i, api := range sortedAPIs {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(api)
	}

	hash := sha1.Sum([]byte(builder.String()))
	return hex.EncodeToString(hash[:])
}

// GetConfigOptimized 优化后的JS-SDK配置获取
func (jm *JSSDKManager) GetConfigOptimized(ctx context.Context, url string, jsAPIList []string) (*JSSDKConfig, error) {
	// 验证参数
	if url == "" {
		return nil, fmt.Errorf("%w: URL不能为空", ErrInvalidURL)
	}
	if len(jsAPIList) == 0 {
		return nil, fmt.Errorf("%w: JSAPI列表不能为空", ErrEmptyJSAPIList)
	}
	if jm.authorizerClient.authorizerAppID == "" {
		return nil, fmt.Errorf("%w: 授权方AppID不能为空", ErrEmptyAppID)
	}

	// 验证URL
	if err := jm.validateURL(url); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	// 先尝试从缓存获取
	if cachedConfig, err := jm.cacher.GetCachedConfig(url, jsAPIList); err == nil && cachedConfig != nil {
		return cachedConfig, nil
	}

	// 缓存未命中，重新生成配置
	accessToken, err := jm.authorizerClient.authClient.client.GetAuthorizerAccessToken(ctx, jm.authorizerClient.authorizerAppID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAccessTokenFailed, err)
	}

	ticket, ticketExpiresAt, err := jm.getJSAPITicket(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTicketFetchFailed, err)
	}

	config := jm.generateSignature(url, ticket, jsAPIList)
	// 记录 ticket 实际过期时间，供缓存按真实有效期判断
	config.expiresAt = ticketExpiresAt

	// 缓存新生成的配置
	jm.cacher.CacheConfig(url, jsAPIList, config)

	return config, nil
}

// GetConfigWithRetry 带重试的JS-SDK配置获取
func (jm *JSSDKManager) GetConfigWithRetry(ctx context.Context, url string, jsAPIList []string, maxRetries int) (*JSSDKConfig, error) {
	// 验证参数
	if url == "" {
		return nil, fmt.Errorf("%w: URL不能为空", ErrInvalidURL)
	}
	if len(jsAPIList) == 0 {
		return nil, fmt.Errorf("%w: JSAPI列表不能为空", ErrEmptyJSAPIList)
	}
	if jm.authorizerClient.authorizerAppID == "" {
		return nil, fmt.Errorf("%w: 授权方AppID不能为空", ErrEmptyAppID)
	}
	if maxRetries <= 0 {
		return nil, fmt.Errorf("%w: 重试次数必须大于0", ErrInvalidMaxRetries)
	}

	// 验证URL
	if err := jm.validateURL(url); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	var lastErr error

	for i := 0; i < maxRetries; i++ {
		config, err := jm.GetConfigOptimized(ctx, url, jsAPIList)
		if err == nil {
			return config, nil
		}

		lastErr = err
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ctx.Err(), lastErr)
		case <-time.After(time.Duration(i+1) * time.Second): // 指数退避
			continue
		}
	}

	return nil, fmt.Errorf("%w: 重试%d次后仍然失败: %v", ErrTicketFetchFailed, maxRetries, lastErr)
}

// validateURL 验证URL
func (jm *JSSDKManager) validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL不能为空")
	}

	// 验证URL格式
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("URL格式错误: %v", err)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL缺少主机名")
	}

	// 验证协议
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL必须使用http或https协议")
	}

	return nil
}

// UserInfo 用户信息
type UserInfo struct {
	core.APIResponse
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

	apiURL := fmt.Sprintf("%s?access_token=%s&openid=%s&lang=zh_CN",
		official_account.URLUserInfo, accessToken, url.QueryEscape(openID))

	var result UserInfo
	err = c.authClient.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    apiURL,
		Result: &result,
	})
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
	core.APIResponse
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

	apiURL := fmt.Sprintf("%s?access_token=%s", official_account.URLGetUserList, accessToken)
	if nextOpenID != "" {
		apiURL += "&next_openid=" + url.QueryEscape(nextOpenID)
	}

	var result UserList
	err = c.authClient.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    apiURL,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// SendTemplateMessage 发送模板消息
func (c *AuthorizerClient) SendTemplateMessage(ctx context.Context, template *official_account.TemplateMessageRequest) (*official_account.SendTemplateMessageResponse, error) {
	if c.authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	accessToken, err := c.authClient.client.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	return official_account.NewTemplate(c.authClient.client.req).SendTemplateMessage(ctx, template, accessToken)
}

// MediaResponse 媒体文件响应
type MediaResponse struct {
	core.APIResponse
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

	apiURL := fmt.Sprintf("%s?access_token=%s&type=%s",
		official_account.URLUploadMaterial, accessToken, url.QueryEscape(mediaType))

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

	// 非 2xx 状态码提前判错，避免把网关/代理层错误当响应体解析
	if err := core.CheckRawResponseStatus(resp.StatusCode, respBody); err != nil {
		return nil, err
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

	apiURL := fmt.Sprintf("%s?access_token=%s&media_id=%s",
		official_account.URLGetMaterial, accessToken, url.QueryEscape(mediaID))

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

	// 统一读取响应体，便于状态码与 JSON 错误判定
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 非 2xx 状态码提前判错，避免把网关/代理层错误当媒体数据返回
	if err := core.CheckRawResponseStatus(resp.StatusCode, respBody); err != nil {
		return nil, err
	}

	// JSON 响应说明有错误（媒体接口正常返回二进制）
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		var result core.APIResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("解析响应失败: %v", err)
		}

		if !result.IsSuccess() {
			return nil, &result
		}
	}

	return respBody, nil
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
//   - opts
//     -- scope：授权类型，snsapi_base 静默授权，snsapi_userinfo 授权并获取用户信息
//     -- state：用于保持请求和回调的状态，授权请求后原样带回给第三方。
//     -- forcePopup 强制此次授权需要用户弹窗确认；默认为false；需要注意的是，若用户命中了特殊场景下的静默授权逻辑，则此参数不生效
//
// Deprecated: 参数非法时会直接 panic 终止进程，请改用 GetAuthorizeURLE。
func (oc *OAuthClient) GetAuthorizeURL(opts ...string) string {
	authorizeURL, err := oc.GetAuthorizeURLE(opts...)
	if err != nil {
		panic(err.Error())
	}
	return authorizeURL
}

// GetAuthorizeURLE 生成授权页面URL（E 版：参数非法时返回 error 而非 panic）
func (oc *OAuthClient) GetAuthorizeURLE(opts ...string) (string, error) {
	// 解析参数
	var scope, state, forcePopup string
	if len(opts) > 0 {
		scope = opts[0]
	}
	if len(opts) > 1 {
		state = opts[1]
	}
	if len(opts) > 2 {
		if opts[2] == "true" {
			forcePopup = "true"
		}
	}

	// 验证参数
	if oc.authorizerClient.authorizerAppID == "" {
		return "", fmt.Errorf("authorizerAppID不能为空")
	}
	if oc.redirectURI == "" {
		return "", fmt.Errorf("redirectURI不能为空")
	}
	if scope == "" {
		return "", fmt.Errorf("scope不能为空")
	}

	// 严格按照微信官方文档要求的参数顺序和大小写
	params := url.Values{}
	params.Set("appid", oc.authorizerClient.authorizerAppID)
	params.Set("redirect_uri", oc.redirectURI)
	params.Set("response_type", core.ResponseTypeCode)
	params.Set("scope", scope)
	if state != "" {
		params.Set("state", state)
	}
	if forcePopup == "true" {
		params.Set("force_popup", forcePopup)
	}
	params.Set("component_appid", oc.authorizerClient.authClient.client.GetConfig().ComponentAppID)

	// 使用微信官方文档指定的授权URL
	return fmt.Sprintf("%s?%s#wechat_redirect", official_account.URLConnectOAuth2Authorize, params.Encode()), nil
}

// GetBaseAuthorizeURL 生成静默授权URL
// Deprecated: 请改用 GetBaseAuthorizeURLE。
func (oc *OAuthClient) GetBaseAuthorizeURL(state string) string {
	return oc.GetAuthorizeURL("snsapi_base", state)
}

// GetBaseAuthorizeURLE 生成静默授权URL（E 版）
func (oc *OAuthClient) GetBaseAuthorizeURLE(state string) (string, error) {
	return oc.GetAuthorizeURLE("snsapi_base", state)
}

// GetUserInfoAuthorizeURL 生成用户信息授权URL
// Deprecated: 请改用 GetUserInfoAuthorizeURLE。
func (oc *OAuthClient) GetUserInfoAuthorizeURL(state string) string {
	return oc.GetAuthorizeURL("snsapi_userinfo", state)
}

// GetUserInfoAuthorizeURLE 生成用户信息授权URL（E 版）
func (oc *OAuthClient) GetUserInfoAuthorizeURLE(state string) (string, error) {
	return oc.GetAuthorizeURLE("snsapi_userinfo", state)
}

// OAuthToken 网页授权Token
type OAuthToken struct {
	core.APIResponse
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
		oc.authorizerClient.authClient.client.logger.Error("授权方AppID为空，无法进行网页授权")
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	// 记录关键参数信息用于调试
	oc.authorizerClient.authClient.client.logger.Info(fmt.Sprintf("网页授权参数验证 - 授权方AppID: %s, Code长度: %d",
		oc.authorizerClient.authorizerAppID, len(code)))

	// 需要先获取组件令牌
	componentToken, err := oc.authorizerClient.authClient.client.GetComponentToken(ctx)
	if err != nil {
		oc.authorizerClient.authClient.client.logger.Error(fmt.Sprintf("获取ComponentAccessToken失败: %v", err))
		return nil, fmt.Errorf("获取ComponentAccessToken失败: %v", err)
	}

	// 检查组件令牌是否为空
	if componentToken == nil {
		oc.authorizerClient.authClient.client.logger.Error("ComponentAccessToken为空，请先获取有效的组件令牌")
		return nil, fmt.Errorf("ComponentAccessToken为空，请先获取有效的组件令牌")
	}

	oc.authorizerClient.authClient.client.logger.Info("已获取ComponentAccessToken", map[string]interface{}{
		"component_access_token": "[REDACTED]",
	})

	// 严格按照微信官方文档要求的参数格式
	params := map[string]string{
		"appid":                  oc.authorizerClient.authorizerAppID,
		"code":                   code,
		"grant_type":             core.GrantTypeAuthorizationCode,
		"component_appid":        oc.authorizerClient.authClient.client.GetConfig().ComponentAppID,
		"component_access_token": componentToken.AccessToken,
	}

	// 记录完整的请求参数用于调试
	oc.authorizerClient.authClient.client.logger.Info("网页授权请求参数详情:", map[string]interface{}{
		"appid":                  oc.authorizerClient.authorizerAppID,
		"code_length":            len(code),
		"component_appid":        oc.authorizerClient.authClient.client.GetConfig().ComponentAppID,
		"component_access_token": "[REDACTED]", // 出于安全考虑隐藏token内容
		"grant_type":             core.GrantTypeAuthorizationCode,
	})

	var oauthToken OAuthToken
	err = oc.authorizerClient.authClient.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    URLComponentOAuth2AccessToken,
		Query:  params,
		Result: &oauthToken,
	})
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

	// expiresAt 所依赖 JSAPI ticket 的实际过期时间（用于缓存过期判断；
	// 历史缓存缺失该值时回退按 defaultTicketExpiresIn 计算）
	expiresAt time.Time
}

// GetConfig 生成JS-SDK配置
func (jm *JSSDKManager) GetConfig(ctx context.Context, url string, jsAPIList []string) (*JSSDKConfig, error) {
	// 验证参数
	if url == "" {
		return nil, fmt.Errorf("%w: URL不能为空", ErrInvalidURL)
	}
	if len(jsAPIList) == 0 {
		return nil, fmt.Errorf("%w: JSAPI列表不能为空", ErrEmptyJSAPIList)
	}
	if jm.authorizerClient.authorizerAppID == "" {
		return nil, fmt.Errorf("%w: 授权方AppID不能为空", ErrEmptyAppID)
	}

	// 验证URL
	if err := jm.validateURL(url); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	// 获取授权方AccessToken
	accessToken, err := jm.authorizerClient.authClient.client.GetAuthorizerAccessToken(ctx, jm.authorizerClient.authorizerAppID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAccessTokenFailed, err)
	}

	// 获取JSAPI Ticket（GetConfig 不走缓存器，仅生成配置）
	ticket, _, err := jm.getJSAPITicket(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTicketFetchFailed, err)
	}

	// 生成签名
	config := jm.generateSignature(url, ticket, jsAPIList)

	return config, nil
}

// jsapiTicketExpireMargin JSAPI Ticket 过期安全余量
const jsapiTicketExpireMargin = 5 * time.Minute

// defaultTicketExpiresIn ticket 有效期默认值（秒），官方正常返回 7200；
// 响应缺失 expires_in 时以此回退
const defaultTicketExpiresIn = 7200

// jsapiTicketInfo JSAPI Ticket 缓存项
type jsapiTicketInfo struct {
	ticket    string
	expiresAt time.Time
}

// getCachedJSAPITicket 读取缓存的JSAPI Ticket（含安全余量判断）
//
// 返回票据内容、实际过期时间及是否命中缓存
func (c *Client) getCachedJSAPITicket(appID string) (string, time.Time, bool) {
	c.jsapiTicketMu.Lock()
	defer c.jsapiTicketMu.Unlock()

	info, ok := c.jsapiTicketMap[appID]
	if !ok {
		return "", time.Time{}, false
	}
	if time.Now().Before(info.expiresAt.Add(-jsapiTicketExpireMargin)) {
		return info.ticket, info.expiresAt, true
	}
	// 已过期（含余量），移除
	delete(c.jsapiTicketMap, appID)
	return "", time.Time{}, false
}

// cacheJSAPITicket 缓存JSAPI Ticket，返回写入的过期时间
// expiresIn 缺失（<=0）时回退 defaultTicketExpiresIn
func (c *Client) cacheJSAPITicket(appID, ticket string, expiresIn int) time.Time {
	c.jsapiTicketMu.Lock()
	defer c.jsapiTicketMu.Unlock()

	if expiresIn <= 0 {
		expiresIn = defaultTicketExpiresIn
	}
	if c.jsapiTicketMap == nil {
		c.jsapiTicketMap = make(map[string]*jsapiTicketInfo)
	}
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	c.jsapiTicketMap[appID] = &jsapiTicketInfo{
		ticket:    ticket,
		expiresAt: expiresAt,
	}
	return expiresAt
}

// getJSAPITicket 获取JSAPI Ticket（优先使用缓存）
// ticket 接口有独立调用配额，不能每次生成 JSSDK 配置都请求
//
// 返回票据内容及其对应后端的有效期（响应缺失 expires_in 时回退 7200 秒），
// 供 JSSDK 配置缓存按 ticket 实际过期时间判断
func (jm *JSSDKManager) getJSAPITicket(ctx context.Context, accessToken string) (string, time.Time, error) {
	appID := jm.authorizerClient.authorizerAppID

	// 优先读缓存
	if ticket, expiresAt, ok := jm.authorizerClient.authClient.client.getCachedJSAPITicket(appID); ok {
		return ticket, expiresAt, nil
	}

	apiURL := official_account.URLGetTicket

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

	err := jm.authorizerClient.authClient.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    apiURL,
		Query:  params,
		Result: &result,
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("调用微信API获取JSAPI Ticket失败: %v", err)
	}

	if result.ErrCode != 0 {
		return "", time.Time{}, fmt.Errorf("获取JSAPI Ticket失败，错误码: %d, 错误信息: %s", result.ErrCode, result.ErrMsg)
	}

	if result.Ticket == "" {
		return "", time.Time{}, fmt.Errorf("获取JSAPI Ticket失败，返回票据为空")
	}

	// 写入缓存并取回按实际 expiresIn 计算的过期时间
	expiresAt := jm.authorizerClient.authClient.client.cacheJSAPITicket(appID, result.Ticket, result.ExpiresIn)

	return result.Ticket, expiresAt, nil
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

// generateNonceStr 生成随机字符串（crypto/rand 密码学安全实现，线程安全）
//
// 返回:
//   - string: 16 位字母数字随机串；熵源不可用（极罕见）时回退时间戳保证可用性
func generateNonceStr() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bts := make([]byte, 16)

	if _, err := crand.Read(bts); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	for i := range bts {
		bts[i] = letters[bts[i]%byte(len(letters))]
	}

	return string(bts)
}

// Create 创建二维码
func (c *AuthorizerClient) Create(ctx context.Context, qrCode *official_account.QRCodeRequest) (*official_account.QRCodeResponse, error) {
	if c.authorizerAppID == "" {
		return nil, fmt.Errorf("授权方AppID不能为空")
	}

	accessToken, err := c.authClient.client.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	return official_account.NewQrcode(c.authClient.client.req).Create(ctx, qrCode, accessToken)
}

// GetQRCodeURL 获取二维码图片URL
func (c *AuthorizerClient) GetQRCodeURL(ticket string) (string, error) {
	return official_account.NewQrcode(c.authClient.client.req).ShowQrcode(ticket)
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

// WXACodeResponse 获取小程序码响应
//
// 成功时 Buffer 为图片二进制数据、ContentType 为图片类型；失败时返回 core.APIResponse 错误。
type WXACodeResponse struct {
	core.APIResponse
	ContentType string `json:"-"` // 图片 Content-Type（如 image/jpeg）
	Buffer      []byte `json:"-"` // 图片二进制数据
}

// GetWXACode 获取小程序码
//
// 参数:
//   - ctx: 上下文
//   - request: 小程序码请求（Path 必填）
//
// 返回:
//   - *WXACodeResponse: 成功时 Buffer 为图片二进制，可直接保存为图片
//   - error: 请求失败或微信返回错误时返回错误
//
// 说明:
//   - 官方接口成功时返回图片二进制、失败时才返回 JSON，因此必须走 MakeRaw
//     并按 Content-Type 分流，不能当作普通 JSON 接口解析
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

	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetWxaCode, accessToken)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := mpc.authorizerClient.authClient.client.req.MakeRaw(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 非 2xx 状态码提前判错，避免把网关/代理层错误当小程序码数据返回
	if err := core.CheckRawResponseStatus(resp.StatusCode, respBody); err != nil {
		return nil, err
	}

	contentType := resp.Header.Get("Content-Type")
	// JSON 响应说明获取失败，解析错误码（正常成功时返回 image/* 二进制）
	if strings.Contains(contentType, "application/json") {
		var apiResp core.APIResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}
		if !apiResp.IsSuccess() {
			return nil, &apiResp
		}
		return nil, fmt.Errorf("接口返回异常: 收到 JSON 成功响应但无小程序码数据, errmsg: %s", apiResp.ErrMsg)
	}

	return &WXACodeResponse{
		APIResponse: core.APIResponse{ErrCode: 0, ErrMsg: "success"},
		ContentType: contentType,
		Buffer:      respBody,
	}, nil
}

// GetUserInfo 获取用户信息
func (oc *OAuthClient) GetUserInfo(ctx context.Context, accessToken, openID string) (*official_account.OAuthUserInfoResponse, error) {
	// 验证参数
	if accessToken == "" {
		return nil, fmt.Errorf("accessToken不能为空")
	}
	if openID == "" {
		return nil, fmt.Errorf("openID不能为空")
	}

	// 使用开放平台OAuth2用户信息的正确URL - 注意这里需要使用sns/userinfo而不是cgi-bin/user/info
	apiURL := URLComponentOAuth2UserInfo

	// 严格按照微信官方文档要求的参数格式
	params := map[string]interface{}{
		"access_token": accessToken,
		"openid":       openID,
		"lang":         "zh_CN",
	}

	var userInfo official_account.OAuthUserInfoResponse
	err := oc.authorizerClient.authClient.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    apiURL,
		Query:  params,
		Result: &userInfo,
	})
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

	// 检查组件令牌是否为空
	if componentToken == nil {
		oc.authorizerClient.authClient.client.logger.Error("ComponentAccessToken为空，请先获取有效的组件令牌")
		return nil, fmt.Errorf("ComponentAccessToken为空，请先获取有效的组件令牌")
	}

	oc.authorizerClient.authClient.client.logger.Info("已获取ComponentAccessToken", map[string]interface{}{
		"component_access_token": "[REDACTED]",
	})

	// 刷新网页授权 token 走 sns/oauth2/component/refresh_token，
	// 而非 api_authorizer_token（后者用于刷新授权方 access_token）
	apiURL := URLSnsOAuth2ComponentRefreshToken

	// 严格按照微信官方文档要求的参数格式
	params := map[string]string{
		"appid":                  oc.authorizerClient.authorizerAppID,
		"grant_type":             core.GrantTypeRefreshToken,
		"refresh_token":          refreshToken,
		"component_appid":        oc.authorizerClient.authClient.client.GetConfig().ComponentAppID,
		"component_access_token": componentToken.AccessToken,
	}

	var oauthToken OAuthToken
	err = oc.authorizerClient.authClient.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    apiURL,
		Query:  params,
		Result: &oauthToken,
	})
	if err != nil {
		return nil, err
	}

	if !oauthToken.IsSuccess() {
		return nil, &oauthToken.APIResponse
	}

	return &oauthToken, nil
}
