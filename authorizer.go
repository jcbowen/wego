package wego

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
	"strings"
	"time"
)

// AuthorizerClient 授权方API客户端
type AuthorizerClient struct {
	wegoClient      *WegoClient
	authorizerAppID string
}

// NewAuthorizerClient 创建授权方API客户端
func NewAuthorizerClient(wegoClient *WegoClient, authorizerAppID string) *AuthorizerClient {
	return &AuthorizerClient{
		wegoClient:      wegoClient,
		authorizerAppID: authorizerAppID,
	}
}

// SendCustomMessage 发送客服消息
func (c *AuthorizerClient) SendCustomMessage(ctx context.Context, toUser string, message interface{}) error {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
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
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
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

// CreateMenu 创建自定义菜单
func (c *AuthorizerClient) CreateMenu(ctx context.Context, menu *Menu) error {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/menu/create?access_token=%s", url.QueryEscape(accessToken))

	var result APIResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, menu, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
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

// GetMenu 获取菜单
func (c *AuthorizerClient) GetMenu(ctx context.Context) (*Menu, error) {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/menu/get?access_token=%s", url.QueryEscape(accessToken))

	var result struct {
		APIResponse
		Menu Menu `json:"menu"`
	}

	err = c.wegoClient.makeRequest(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result.Menu, nil
}

// DeleteMenu 删除菜单
func (c *AuthorizerClient) DeleteMenu(ctx context.Context) error {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/menu/delete?access_token=%s", url.QueryEscape(accessToken))

	var result APIResponse
	err = c.wegoClient.makeRequest(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// GetUserInfo 获取用户信息
func (c *AuthorizerClient) GetUserInfo(ctx context.Context, openID string) (*UserInfo, error) {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN",
		url.QueryEscape(accessToken), url.QueryEscape(openID))

	var result UserInfo
	err = c.wegoClient.makeRequest(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &result.APIResponse
	}

	return &result, nil
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

// GetUserList 获取用户列表
func (c *AuthorizerClient) GetUserList(ctx context.Context, nextOpenID string) (*UserList, error) {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/user/get?access_token=%s", url.QueryEscape(accessToken))
	if nextOpenID != "" {
		apiURL += "&next_openid=" + url.QueryEscape(nextOpenID)
	}

	var result UserList
	err = c.wegoClient.makeRequest(ctx, "GET", apiURL, nil, &result)
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

// SendTemplateMessage 发送模板消息
func (c *AuthorizerClient) SendTemplateMessage(ctx context.Context, template *TemplateMessage) error {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s", url.QueryEscape(accessToken))

	var result struct {
		APIResponse
		MsgID int64 `json:"msgid"`
	}

	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, template, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result.APIResponse
	}

	return nil
}

// TemplateMessage 模板消息
type TemplateMessage struct {
	ToUser      string                   `json:"touser"`
	TemplateID  string                   `json:"template_id"`
	URL         string                   `json:"url,omitempty"`
	MiniProgram *TemplateMiniProgramInfo `json:"miniprogram,omitempty"`
	Data        map[string]TemplateData  `json:"data"`
}

// TemplateData 模板消息数据
type TemplateData struct {
	Value string `json:"value"`
	Color string `json:"color,omitempty"`
}

// TemplateMiniProgramInfo 模板消息小程序信息
type TemplateMiniProgramInfo struct {
	AppID    string `json:"appid"`
	PagePath string `json:"pagepath"`
}

// UploadMedia 上传临时素材
func (c *AuthorizerClient) UploadMedia(ctx context.Context, mediaType, filename string, data []byte) (*MediaResponse, error) {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
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
	resp, err := c.wegoClient.httpClient.Do(req)
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

// MediaResponse 媒体文件响应
type MediaResponse struct {
	APIResponse
	Type      string `json:"type"`
	MediaID   string `json:"media_id"`
	CreatedAt int64  `json:"created_at"`
}

// GetMedia 获取临时素材
func (c *AuthorizerClient) GetMedia(ctx context.Context, mediaID string) ([]byte, error) {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
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
	resp, err := c.wegoClient.httpClient.Do(req)
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
	params := url.Values{}
	params.Set("appid", oc.authorizerClient.authorizerAppID)
	params.Set("redirect_uri", oc.redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", scope)
	params.Set("state", state)
	params.Set("component_appid", oc.authorizerClient.wegoClient.config.ComponentAppID)

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
	// 需要先获取verifyTicket，这里传递空字符串让方法内部处理
	componentToken, err := oc.authorizerClient.wegoClient.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("获取ComponentAccessToken失败: %v", err)
	}

	apiURL := "https://api.weixin.qq.com/sns/oauth2/component/access_token"

	params := map[string]interface{}{
		"appid":                  oc.authorizerClient.authorizerAppID,
		"code":                   code,
		"grant_type":             "authorization_code",
		"component_appid":        oc.authorizerClient.wegoClient.config.ComponentAppID,
		"component_access_token": componentToken.AccessToken,
	}

	var oauthToken OAuthToken
	err = oc.authorizerClient.wegoClient.makeRequest(ctx, "POST", apiURL, params, &oauthToken)
	if err != nil {
		return nil, err
	}

	if !oauthToken.IsSuccess() {
		return nil, &oauthToken.APIResponse
	}

	return &oauthToken, nil
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
	apiURL := "https://api.weixin.qq.com/sns/userinfo"

	params := map[string]interface{}{
		"access_token": accessToken,
		"openid":       openID,
		"lang":         "zh_CN",
	}

	var userInfo OAuthUserInfo
	err := oc.authorizerClient.wegoClient.makeRequest(ctx, "GET", apiURL, params, &userInfo)
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
	// 需要先获取verifyTicket，这里传递空字符串让方法内部处理
	componentToken, err := oc.authorizerClient.wegoClient.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("获取ComponentAccessToken失败: %v", err)
	}

	apiURL := "https://api.weixin.qq.com/sns/oauth2/component/refresh_token"

	params := map[string]interface{}{
		"appid":                  oc.authorizerClient.authorizerAppID,
		"grant_type":             "refresh_token",
		"refresh_token":          refreshToken,
		"component_appid":        oc.authorizerClient.wegoClient.config.ComponentAppID,
		"component_access_token": componentToken.AccessToken,
	}

	var oauthToken OAuthToken
	err = oc.authorizerClient.wegoClient.makeRequest(ctx, "POST", apiURL, params, &oauthToken)
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
}

// GetJSSDKManager 创建JS-SDK管理器
func (c *AuthorizerClient) GetJSSDKManager() *JSSDKManager {
	return &JSSDKManager{
		authorizerClient: c,
	}
}

// JSSDKConfig JS-SDK配置结构
type JSSDKConfig struct {
	AppID     string   `json:"appId"`
	Timestamp int64    `json:"timestamp"`
	NonceStr  string   `json:"nonceStr"`
	Signature string   `json:"signature"`
	JSAPIList []string `json:"jsApiList"`
}

// GetConfig 生成JS-SDK配置
func (jm *JSSDKManager) GetConfig(ctx context.Context, url string, jsAPIList []string) (*JSSDKConfig, error) {
	// 获取授权方AccessToken
	accessToken, err := jm.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, jm.authorizerClient.authorizerAppID)
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

	err := jm.authorizerClient.wegoClient.makeRequest(ctx, "GET", apiURL, params, &result)
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
	bytes := make([]byte, 16)
	for i := range bytes {
		bytes[i] = letters[rand.Intn(len(letters))]
	}
	return string(bytes)
}

// CreateQRCode 创建二维码
func (c *AuthorizerClient) CreateQRCode(ctx context.Context, qrCode *QRCodeRequest) (*QRCodeResponse, error) {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token=%s", url.QueryEscape(accessToken))

	var result QRCodeResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, qrCode, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
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

// GetQRCodeURL 获取二维码图片URL
func (c *AuthorizerClient) GetQRCodeURL(ticket string) string {
	return fmt.Sprintf("https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=%s", url.QueryEscape(ticket))
}

// MiniProgramClient 小程序API客户端
type MiniProgramClient struct {
	authorizerClient *AuthorizerClient
}

// NewMiniProgramClient 创建小程序API客户端
func NewMiniProgramClient(authorizerClient *AuthorizerClient) *MiniProgramClient {
	return &MiniProgramClient{
		authorizerClient: authorizerClient,
	}
}

// GetWXACode 获取小程序码
func (c *MiniProgramClient) GetWXACode(ctx context.Context, path string, width int) ([]byte, error) {
	accessToken, err := c.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerClient.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/getwxacode?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"path":  path,
		"width": width,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.authorizerClient.wegoClient.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应类型
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// 返回的是错误信息
		body, _ := io.ReadAll(resp.Body)
		var result APIResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
		return nil, &result
	}

	// 返回图片数据
	return io.ReadAll(resp.Body)
}

// CreateWXAQrCode 创建小程序二维码
func (c *MiniProgramClient) CreateWXAQrCode(ctx context.Context, path string, width int) ([]byte, error) {
	accessToken, err := c.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerClient.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/wxaapp/createwxaqrcode?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"path":  path,
		"width": width,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.authorizerClient.wegoClient.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应类型
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// 返回的是错误信息
		body, _ := io.ReadAll(resp.Body)
		var result APIResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
		return nil, &result
	}

	// 返回图片数据
	return io.ReadAll(resp.Body)
}

// CommitCode 提交代码
func (c *MiniProgramClient) CommitCode(ctx context.Context, templateID int, extJSON, userVersion, userDesc string) error {
	accessToken, err := c.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerClient.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/commit?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"template_id":  templateID,
		"ext_json":     extJSON,
		"user_version": userVersion,
		"user_desc":    userDesc,
	}

	var result APIResponse
	err = c.authorizerClient.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// GetLatestAuditStatus 获取最新审核状态
func (c *MiniProgramClient) GetLatestAuditStatus(ctx context.Context) (*AuditStatusResponse, error) {
	accessToken, err := c.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerClient.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/getlatestauditstatus?access_token=%s", url.QueryEscape(accessToken))

	var result AuditStatusResponse
	err = c.authorizerClient.wegoClient.makeRequest(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// Release 发布已审核通过的小程序
func (c *MiniProgramClient) Release(ctx context.Context) error {
	accessToken, err := c.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerClient.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/release?access_token=%s", url.QueryEscape(accessToken))

	var result APIResponse
	err = c.authorizerClient.wegoClient.makeRequest(ctx, "POST", apiURL, nil, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// RevertCodeRelease 版本回退
func (c *MiniProgramClient) RevertCodeRelease(ctx context.Context) error {
	accessToken, err := c.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerClient.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/revertcoderelease?access_token=%s", url.QueryEscape(accessToken))

	var result APIResponse
	err = c.authorizerClient.wegoClient.makeRequest(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// ChangeVisitStatus 修改小程序线上代码的可见状态
func (c *MiniProgramClient) ChangeVisitStatus(ctx context.Context, action string) error {
	accessToken, err := c.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerClient.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/change_visitstatus?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"action": action, // open: 开启 close: 关闭
	}

	var result APIResponse
	err = c.authorizerClient.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// AuditStatusResponse 审核状态响应
type AuditStatusResponse struct {
	APIResponse
	AuditID    int    `json:"auditid"`
	Status     int    `json:"status"`
	Reason     string `json:"reason"`
	ScreenShot string `json:"ScreenShot"`
}

// GetTemplateList 获取代码模板列表
func (c *MiniProgramClient) GetTemplateList(ctx context.Context) (*TemplateListResponse, error) {
	accessToken, err := c.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerClient.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/gettemplatelist?access_token=%s", url.QueryEscape(accessToken))

	var result TemplateListResponse
	err = c.authorizerClient.wegoClient.makeRequest(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// DeleteTemplate 删除指定代码模板
func (c *MiniProgramClient) DeleteTemplate(ctx context.Context, templateID int) error {
	accessToken, err := c.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerClient.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/deletetemplate?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"template_id": templateID,
	}

	var result APIResponse
	err = c.authorizerClient.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// GetCategory 获取账号可以设置的所有类目
func (c *MiniProgramClient) GetCategory(ctx context.Context) (*CategoryResponse, error) {
	accessToken, err := c.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerClient.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/get_category?access_token=%s", url.QueryEscape(accessToken))

	var result CategoryResponse
	err = c.authorizerClient.wegoClient.makeRequest(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetPage 获取小程序的页面配置
func (c *MiniProgramClient) GetPage(ctx context.Context) (*PageResponse, error) {
	accessToken, err := c.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerClient.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/get_page?access_token=%s", url.QueryEscape(accessToken))

	var result PageResponse
	err = c.authorizerClient.wegoClient.makeRequest(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// SubmitAudit 提交审核
func (c *MiniProgramClient) SubmitAudit(ctx context.Context, auditData *AuditData) (*SubmitAuditResponse, error) {
	accessToken, err := c.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerClient.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/submit_audit?access_token=%s", url.QueryEscape(accessToken))

	var result SubmitAuditResponse
	err = c.authorizerClient.wegoClient.makeRequest(ctx, "POST", apiURL, auditData, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// UndoCodeAudit 撤销审核
func (c *MiniProgramClient) UndoCodeAudit(ctx context.Context) error {
	accessToken, err := c.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerClient.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/wxa/undocodeaudit?access_token=%s", url.QueryEscape(accessToken))

	var result APIResponse
	err = c.authorizerClient.wegoClient.makeRequest(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// TemplateListResponse 模板列表响应
type TemplateListResponse struct {
	APIResponse
	TemplateList []TemplateInfo `json:"template_list"`
}

// TemplateInfo 模板信息
type TemplateInfo struct {
	TemplateID   int    `json:"template_id"`
	TemplateType int    `json:"template_type"`
	UserVersion  string `json:"user_version"`
	UserDesc     string `json:"user_desc"`
	CreateTime   int64  `json:"create_time"`
}

// CategoryResponse 类目响应
type CategoryResponse struct {
	APIResponse
	CategoryList []CategoryInfo `json:"category_list"`
}

// CategoryInfo 类目信息
type CategoryInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// PageResponse 页面响应
type PageResponse struct {
	APIResponse
	PageList []string `json:"page_list"`
}

// SubmitAuditResponse 提交审核响应
type SubmitAuditResponse struct {
	APIResponse
	AuditID int `json:"auditid"`
}

// AuditData 审核数据
type AuditData struct {
	ItemList []AuditItem `json:"item_list"`
}

// AuditItem 审核项
type AuditItem struct {
	Address     string `json:"address"`
	Tag         string `json:"tag"`
	FirstClass  string `json:"first_class"`
	SecondClass string `json:"second_class"`
	ThirdClass  string `json:"third_class"`
	FirstID     int    `json:"first_id"`
	SecondID    int    `json:"second_id"`
	ThirdID     int    `json:"third_id"`
	Title       string `json:"title"`
}

// SendSubscribeMessage 发送订阅消息
func (c *MiniProgramClient) SendSubscribeMessage(ctx context.Context, message *SubscribeMessage) error {
	accessToken, err := c.authorizerClient.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerClient.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", url.QueryEscape(accessToken))

	var result APIResponse
	err = c.authorizerClient.wegoClient.makeRequest(ctx, "POST", apiURL, message, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// SubscribeMessage 订阅消息
type SubscribeMessage struct {
	ToUser           string                   `json:"touser"`
	TemplateID       string                   `json:"template_id"`
	Page             string                   `json:"page,omitempty"`
	Data             map[string]SubscribeData `json:"data"`
	MiniprogramState string                   `json:"miniprogram_state,omitempty"`
	Lang             string                   `json:"lang,omitempty"`
}

// SubscribeData 订阅消息数据
type SubscribeData struct {
	Value string `json:"value"`
}

// CommonAPI 通用API方法
func (c *AuthorizerClient) CommonAPI(ctx context.Context, apiPath string, params map[string]string, data interface{}, result interface{}) error {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com%s?access_token=%s", apiPath, url.QueryEscape(accessToken))

	// 添加查询参数
	if len(params) > 0 {
		queryParams := url.Values{}
		for key, value := range params {
			queryParams.Set(key, value)
		}
		apiURL += "&" + queryParams.Encode()
	}

	method := "GET"
	if data != nil {
		method = "POST"
	}

	err = c.wegoClient.makeRequest(ctx, method, apiURL, data, result)
	if err != nil {
		return err
	}

	return nil
}

// BatchGetUserInfo 批量获取用户信息
func (c *AuthorizerClient) BatchGetUserInfo(ctx context.Context, openIDs []string) ([]UserInfo, error) {
	if len(openIDs) > 100 {
		return nil, fmt.Errorf("一次最多获取100个用户信息")
	}

	userList := make([]map[string]string, len(openIDs))
	for i, openID := range openIDs {
		userList[i] = map[string]string{
			"openid": openID,
			"lang":   "zh_CN",
		}
	}

	request := map[string]interface{}{
		"user_list": userList,
	}

	var result struct {
		APIResponse
		UserInfoList []UserInfo `json:"user_info_list"`
	}

	err := c.CommonAPI(ctx, "/cgi-bin/user/info/batchget", nil, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return result.UserInfoList, nil
}

// AuthorizerOptionResponse 授权方选项信息响应
type AuthorizerOptionResponse struct {
	APIResponse
	AuthorizerAppID string `json:"authorizer_appid"`
	OptionName      string `json:"option_name"`
	OptionValue     string `json:"option_value"`
}

// DomainModifyResponse 域名修改响应
type DomainModifyResponse struct {
	APIResponse
	RequestDomain  []string `json:"requestdomain"`
	SocketDomain   []string `json:"socketdomain"`
	UploadDomain   []string `json:"uploaddomain"`
	DownloadDomain []string `json:"downloaddomain"`
}

// DomainConfirmFileResponse 域名校验文件响应
type DomainConfirmFileResponse struct {
	APIResponse
	FileName    string `json:"file_name"`
	FileContent string `json:"file_content"`
}

// TemplateDraftListResponse 模板草稿列表响应
type TemplateDraftListResponse struct {
	APIResponse
	DraftList []TemplateDraft `json:"draft_list"`
}

// TemplateDraft 模板草稿
type TemplateDraft struct {
	DraftID     int    `json:"draft_id"`
	UserVersion string `json:"user_version"`
	UserDesc    string `json:"user_desc"`
	CreateTime  int64  `json:"create_time"`
}

// OpenAccountResponse 开放平台账号响应
type OpenAccountResponse struct {
	APIResponse
	OpenAppID string `json:"open_appid"`
}

// CreateOpenAccount 创建开放平台账号并绑定公众号/小程序
func (c *AuthorizerClient) CreateOpenAccount(ctx context.Context) (*OpenAccountResponse, error) {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/open/create?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"appid": c.authorizerAppID,
	}

	var result OpenAccountResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// BindOpenAccount 绑定公众号/小程序到已有开放平台账号
func (c *AuthorizerClient) BindOpenAccount(ctx context.Context, openAppID string) error {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/open/bind?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"appid":      c.authorizerAppID,
		"open_appid": openAppID,
	}

	var result APIResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// UnbindOpenAccount 解绑公众号/小程序
func (c *AuthorizerClient) UnbindOpenAccount(ctx context.Context, openAppID string) error {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/open/unbind?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"appid":      c.authorizerAppID,
		"open_appid": openAppID,
	}

	var result APIResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// SetAuthorizerOptionInfo 设置授权方选项信息
func (c *AuthorizerClient) SetAuthorizerOptionInfo(ctx context.Context, optionName, optionValue string) error {
	componentToken, err := c.wegoClient.GetComponentAccessToken(ctx, "")
	if err != nil {
		return err
	}

	request := map[string]interface{}{
		"component_appid":  c.wegoClient.config.ComponentAppID,
		"authorizer_appid": c.authorizerAppID,
		"option_name":      optionName,
		"option_value":     optionValue,
	}

	var result APIResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APISetAuthorizerOptionURL, url.QueryEscape(componentToken.AccessToken))
	err = c.wegoClient.makeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// GetAuthorizerOptionInfo 获取授权方选项信息
func (c *AuthorizerClient) GetAuthorizerOptionInfo(ctx context.Context, optionName string) (*AuthorizerOptionResponse, error) {
	componentToken, err := c.wegoClient.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := map[string]interface{}{
		"component_appid":  c.wegoClient.config.ComponentAppID,
		"authorizer_appid": c.authorizerAppID,
		"option_name":      optionName,
	}

	var result AuthorizerOptionResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIGetAuthorizerOptionURL, url.QueryEscape(componentToken.AccessToken))
	err = c.wegoClient.makeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// ModifyThirdpartyServerDomain 设置第三方平台服务器域名
func (c *AuthorizerClient) ModifyThirdpartyServerDomain(ctx context.Context, action string, domains []string) (*DomainModifyResponse, error) {
	componentToken, err := c.wegoClient.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := map[string]interface{}{
		"action":            action,
		"wxa_server_domain": domains,
	}

	var result DomainModifyResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIModifyServerDomainURL, url.QueryEscape(componentToken.AccessToken))
	err = c.wegoClient.makeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetThirdpartyJumpDomainConfirmFile 获取第三方平台业务域名校验文件
func (c *AuthorizerClient) GetThirdpartyJumpDomainConfirmFile(ctx context.Context) (*DomainConfirmFileResponse, error) {
	componentToken, err := c.wegoClient.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	var result DomainConfirmFileResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIGetJumpDomainFileURL, url.QueryEscape(componentToken.AccessToken))
	err = c.wegoClient.makeRequest(ctx, "GET", url, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// ModifyThirdpartyJumpDomain 设置第三方平台业务域名
func (c *AuthorizerClient) ModifyThirdpartyJumpDomain(ctx context.Context, action string, domains []string) (*DomainModifyResponse, error) {
	componentToken, err := c.wegoClient.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := map[string]interface{}{
		"action":          action,
		"wxa_jump_domain": domains,
	}

	var result DomainModifyResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIModifyJumpDomainURL, url.QueryEscape(componentToken.AccessToken))
	err = c.wegoClient.makeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetTemplateDraftList 获取代码模板草稿列表
func (c *AuthorizerClient) GetTemplateDraftList(ctx context.Context) (*TemplateDraftListResponse, error) {
	componentToken, err := c.wegoClient.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	var result TemplateDraftListResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIGetTemplateDraftListURL, url.QueryEscape(componentToken.AccessToken))
	err = c.wegoClient.makeRequest(ctx, "GET", url, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// AddToTemplate 将草稿添加到代码模板库
func (c *AuthorizerClient) AddToTemplate(ctx context.Context, draftID int) (*APIResponse, error) {
	componentToken, err := c.wegoClient.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := map[string]interface{}{
		"draft_id": draftID,
	}

	var result APIResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIAddToTemplateURL, url.QueryEscape(componentToken.AccessToken))
	err = c.wegoClient.makeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result
	}

	return &result, nil
}

// GetOpenAccount 获取公众号/小程序所绑定的开放平台账号
func (c *AuthorizerClient) GetOpenAccount(ctx context.Context) (*OpenAccountResponse, error) {
	accessToken, err := c.wegoClient.GetAuthorizerAccessToken(ctx, c.authorizerAppID)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/open/get?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"appid": c.authorizerAppID,
	}

	var result OpenAccountResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// ShopAccountResponse 视频号小店商家信息响应
type ShopAccountResponse struct {
	APIResponse
	Data ShopAccountData `json:"data"`
}

// ShopAccountData 视频号小店商家信息数据
type ShopAccountData struct {
	ServiceAgentPath        string            `json:"service_agent_path"`
	ServiceAgentPhone       string            `json:"service_agent_phone"`
	ServiceAgentType        []int             `json:"service_agent_type"`
	DefaultReceivingAddress *ReceivingAddress `json:"default_receiving_address,omitempty"`
}

// ReceivingAddress 收货地址
type ReceivingAddress struct {
	ReceiverName    string `json:"receiver_name"`
	DetailedAddress string `json:"detailed_address"`
	TelNumber       string `json:"tel_number"`
	Country         string `json:"country,omitempty"`
	Province        string `json:"province,omitempty"`
	City            string `json:"city,omitempty"`
	Town            string `json:"town,omitempty"`
}

// ShopAccountUpdateRequest 更新商家信息请求
type ShopAccountUpdateRequest struct {
	ServiceAgentPath        string            `json:"service_agent_path,omitempty"`
	ServiceAgentPhone       string            `json:"service_agent_phone,omitempty"`
	ServiceAgentType        []int             `json:"service_agent_type,omitempty"`
	DefaultReceivingAddress *ReceivingAddress `json:"default_receiving_address,omitempty"`
}

// getShopAccessToken 获取视频号小店access_token
func (c *AuthorizerClient) getShopAccessToken(ctx context.Context) (string, error) {
	// 视频号小店使用标准的access_token获取方式
	// 需要先注册视频号小店获取access_token
	registerResult, err := c.RegisterShop(ctx)
	if err != nil {
		return "", fmt.Errorf("注册视频号小店失败: %v", err)
	}

	if !registerResult.IsSuccess() {
		return "", &registerResult.APIResponse
	}

	return registerResult.Data.AccessToken, nil
}

// GetShopAccountInfo 获取视频号小店商家信息
func (c *AuthorizerClient) GetShopAccountInfo(ctx context.Context) (*ShopAccountResponse, error) {
	// 视频号小店接口使用标准的access_token
	accessToken, err := c.getShopAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/shop/account/get_info?access_token=%s", url.QueryEscape(accessToken))

	var result ShopAccountResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// UpdateShopAccountInfo 更新视频号小店商家信息
func (c *AuthorizerClient) UpdateShopAccountInfo(ctx context.Context, request *ShopAccountUpdateRequest) error {
	// 视频号小店接口使用标准的access_token
	accessToken, err := c.getShopAccessToken(ctx)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/shop/account/update_info?access_token=%s", url.QueryEscape(accessToken))

	var result APIResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// ShopRegisterResponse 视频号小店注册响应
type ShopRegisterResponse struct {
	APIResponse
	Data ShopRegisterData `json:"data"`
}

// ShopRegisterData 视频号小店注册数据
type ShopRegisterData struct {
	AccessToken string `json:"access_token"`
	ExpireIn    int    `json:"expire_in"`
}

// RegisterShop 注册视频号小店
func (c *AuthorizerClient) RegisterShop(ctx context.Context) (*ShopRegisterResponse, error) {
	// 视频号小店注册接口使用第三方平台component_access_token
	componentToken, err := c.wegoClient.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("获取ComponentAccessToken失败: %v", err)
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/shop/register?access_token=%s", url.QueryEscape(componentToken.AccessToken))

	request := map[string]interface{}{
		"appid": c.authorizerAppID,
	}

	var result ShopRegisterResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// ShopCallbackConfig 视频号小店回调配置
type ShopCallbackConfig struct {
	URL   string `json:"url"`
	Token string `json:"token"`
	Key   string `json:"encoding_aes_key"`
}

// SetShopCallback 设置视频号小店回调配置
func (c *AuthorizerClient) SetShopCallback(ctx context.Context, config *ShopCallbackConfig) error {
	// 视频号小店接口使用标准的access_token
	accessToken, err := c.getShopAccessToken(ctx)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/shop/callback/set?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"url":              config.URL,
		"token":            config.Token,
		"encoding_aes_key": config.Key,
	}

	var result APIResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// GetShopCallback 获取视频号小店回调配置
func (c *AuthorizerClient) GetShopCallback(ctx context.Context) (*ShopCallbackConfig, error) {
	// 视频号小店接口使用标准的access_token
	accessToken, err := c.getShopAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/shop/callback/get?access_token=%s", url.QueryEscape(accessToken))

	var result struct {
		APIResponse
		URL   string `json:"url"`
		Token string `json:"token"`
		Key   string `json:"encoding_aes_key"`
	}

	err = c.wegoClient.makeRequest(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &ShopCallbackConfig{
		URL:   result.URL,
		Token: result.Token,
		Key:   result.Key,
	}, nil
}

// ShopOrderResponse 视频号小店订单响应
type ShopOrderResponse struct {
	APIResponse
	Data ShopOrderData `json:"data"`
}

// ShopOrderData 视频号小店订单数据
type ShopOrderData struct {
	OrderID     string `json:"order_id"`
	Status      int    `json:"status"`
	CreateTime  int64  `json:"create_time"`
	UpdateTime  int64  `json:"update_time"`
	TotalAmount int    `json:"total_amount"`
	Currency    string `json:"currency"`
}

// GetShopOrder 获取视频号小店订单详情
func (c *AuthorizerClient) GetShopOrder(ctx context.Context, orderID string) (*ShopOrderResponse, error) {
	// 视频号小店接口使用标准的access_token
	accessToken, err := c.getShopAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/shop/order/get?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"order_id": orderID,
	}

	var result ShopOrderResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// ShopOrderListResponse 视频号小店订单列表响应
type ShopOrderListResponse struct {
	APIResponse
	Data  []ShopOrderData `json:"data"`
	Total int             `json:"total"`
}

// GetShopOrderList 获取视频号小店订单列表
func (c *AuthorizerClient) GetShopOrderList(ctx context.Context, page, pageSize int) (*ShopOrderListResponse, error) {
	// 视频号小店接口使用标准的access_token
	accessToken, err := c.getShopAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/shop/order/list?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"page":      page,
		"page_size": pageSize,
	}

	var result ShopOrderListResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// ShopProductResponse 视频号小店商品响应
type ShopProductResponse struct {
	APIResponse
	Data ShopProductData `json:"data"`
}

// ShopProductData 视频号小店商品数据
type ShopProductData struct {
	ProductID  string `json:"product_id"`
	Title      string `json:"title"`
	Price      int    `json:"price"`
	Stock      int    `json:"stock"`
	Status     int    `json:"status"`
	CreateTime int64  `json:"create_time"`
}

// GetShopProduct 获取视频号小店商品详情
func (c *AuthorizerClient) GetShopProduct(ctx context.Context, productID string) (*ShopProductResponse, error) {
	// 视频号小店接口使用标准的access_token
	accessToken, err := c.getShopAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/shop/product/get?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"product_id": productID,
	}

	var result ShopProductResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// ShopProductListResponse 视频号小店商品列表响应
type ShopProductListResponse struct {
	APIResponse
	Data  []ShopProductData `json:"data"`
	Total int               `json:"total"`
}

// GetShopProductList 获取视频号小店商品列表
func (c *AuthorizerClient) GetShopProductList(ctx context.Context, page, pageSize int) (*ShopProductListResponse, error) {
	// 视频号小店接口使用标准的access_token
	accessToken, err := c.getShopAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/shop/product/list?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"page":      page,
		"page_size": pageSize,
	}

	var result ShopProductListResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// ShopProductAddRequest 添加商品请求
type ShopProductAddRequest struct {
	Title       string   `json:"title"`
	Price       int      `json:"price"`
	Stock       int      `json:"stock"`
	Description string   `json:"description,omitempty"`
	Images      []string `json:"images,omitempty"`
}

// AddShopProduct 添加视频号小店商品
func (c *AuthorizerClient) AddShopProduct(ctx context.Context, request *ShopProductAddRequest) (*ShopProductResponse, error) {
	// 视频号小店接口使用标准的access_token
	accessToken, err := c.getShopAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/shop/product/add?access_token=%s", url.QueryEscape(accessToken))

	var result ShopProductResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// ShopDeliveryResponse 视频号小店发货响应
type ShopDeliveryResponse struct {
	APIResponse
}

// DeliverShopOrder 视频号小店订单发货
func (c *AuthorizerClient) DeliverShopOrder(ctx context.Context, orderID, logisticsCompany, trackingNumber string) error {
	// 视频号小店接口使用标准的access_token
	accessToken, err := c.getShopAccessToken(ctx)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/shop/delivery/send?access_token=%s", url.QueryEscape(accessToken))

	request := map[string]interface{}{
		"order_id":          orderID,
		"logistics_company": logisticsCompany,
		"tracking_number":   trackingNumber,
	}

	var result ShopDeliveryResponse
	err = c.wegoClient.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result.APIResponse
	}

	return nil
}
