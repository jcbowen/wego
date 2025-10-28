package api

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/storage"
)

// APIClient API客户端
type APIClient struct {
	Client *core.WegoClient
}

// NewAPIClient 创建新的API客户端
func NewAPIClient(client *core.WegoClient) *APIClient {
	return &APIClient{
		Client: client,
	}
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
	if token, err := c.Client.GetComponentToken(ctx); err == nil && token != nil && token.ExpiresAt.After(time.Now()) {
		return token, nil
	}

	request := ComponentTokenRequest{
		ComponentAppID:        c.Client.GetConfig().ComponentAppID,
		ComponentAppSecret:    c.Client.GetConfig().ComponentAppSecret,
		ComponentVerifyTicket: verifyTicket,
	}

	var result struct {
		APIResponse
		ComponentAccessToken string `json:"component_access_token"`
		ExpiresIn            int    `json:"expires_in"`
	}

	err := c.Client.MakeRequest(ctx, "POST", APIComponentTokenURL, request, &result)
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
	if err := c.Client.SetComponentToken(token); err != nil {
		c.Client.GetLogger().Warnf("保存组件令牌失败: %v", err)
	}

	return token, nil
}

// GetPreAuthCode 获取预授权码
func (c *APIClient) GetPreAuthCode(ctx context.Context) (*PreAuthCodeResponse, error) {
	// 先从存储中获取
	if preAuthCode, err := c.Client.GetPreAuthCode(ctx); err == nil && preAuthCode != nil && preAuthCode.ExpiresAt.After(time.Now()) {
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
		ComponentAppID: c.Client.GetConfig().ComponentAppID,
	}

	var result PreAuthCodeResponse
	apiURL := fmt.Sprintf("%s?component_access_token=%s", APIPreAuthCodeURL, url.QueryEscape(componentToken.AccessToken))
	err = c.Client.MakeRequest(ctx, "POST", apiURL, request, &result)
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
	if err := c.Client.SetPreAuthCode(ctx, preAuthCode); err != nil {
		c.Client.GetLogger().Warnf("保存预授权码失败: %v", err)
	}

	return &result, nil
}

// QueryAuth 使用授权码换取授权信息
func (c *APIClient) QueryAuth(ctx context.Context, authorizationCode string) (*QueryAuthResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := QueryAuthRequest{
		ComponentAppID:    c.Client.GetConfig().ComponentAppID,
		AuthorizationCode: authorizationCode,
	}

	var result QueryAuthResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIQueryAuthURL, url.QueryEscape(componentToken.AccessToken))
	err = c.Client.MakeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	// 缓存授权方token
	if err := c.Client.SetAuthorizerToken(
		result.AuthorizationInfo.AuthorizerAppID,
		result.AuthorizationInfo.AuthorizerAccessToken,
		result.AuthorizationInfo.AuthorizerRefreshToken,
		result.AuthorizationInfo.ExpiresIn,
	); err != nil {
		c.Client.GetLogger().Warnf("缓存授权方token失败: %v", err)
	}

	return &result, nil
}

// RefreshAuthorizerToken 刷新授权方access_token
func (c *APIClient) RefreshAuthorizerToken(ctx context.Context, authorizerAppID, refreshToken string) (*AuthorizerTokenResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := AuthorizerTokenRequest{
		ComponentAppID:         c.Client.GetConfig().ComponentAppID,
		AuthorizerAppID:        authorizerAppID,
		AuthorizerRefreshToken: refreshToken,
	}

	var result AuthorizerTokenResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIAuthorizerTokenURL, url.QueryEscape(componentToken.AccessToken))
	err = c.Client.MakeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	// 更新缓存
	if err := c.Client.SetAuthorizerToken(
		authorizerAppID,
		result.AuthorizerAccessToken,
		result.AuthorizerRefreshToken,
		result.ExpiresIn,
	); err != nil {
		c.Client.GetLogger().Warnf("更新授权方token失败: %v", err)
	}

	return &result, nil
}

// GetAuthorizerInfo 获取授权方信息
func (c *APIClient) GetAuthorizerInfo(ctx context.Context, authorizerAppID string) (*GetAuthorizerInfoResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := GetAuthorizerInfoRequest{
		ComponentAppID:  c.Client.GetConfig().ComponentAppID,
		AuthorizerAppID: authorizerAppID,
	}

	var result GetAuthorizerInfoResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIGetAuthorizerInfoURL, url.QueryEscape(componentToken.AccessToken))
	err = c.Client.MakeRequest(ctx, "POST", url, request, &result)
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
		ComponentAppID: c.Client.GetConfig().ComponentAppID,
		Offset:         offset,
		Count:          count,
	}

	var result GetAuthorizerListResponse
	url := fmt.Sprintf("%s?component_access_token=%s", "https://api.weixin.qq.com/cgi-bin/component/api_get_authorizer_list", url.QueryEscape(componentToken.AccessToken))
	err = c.Client.MakeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GenerateAuthURL 生成授权链接
func (c *APIClient) GenerateAuthURL(preAuthCode string, authType int, bizAppID string) string {
	baseURL := "https://mp.weixin.qq.com/cgi-bin/componentloginpage"
	params := url.Values{
		"component_appid": {c.Client.GetConfig().ComponentAppID},
		"pre_auth_code":   {preAuthCode},
		"redirect_uri":    {c.Client.GetConfig().RedirectURI},
	}

	if authType > 0 {
		params.Set("auth_type", fmt.Sprintf("%d", authType))
	}

	if bizAppID != "" {
		params.Set("biz_appid", bizAppID)
	}

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// ClearQuota 重置API调用次数
func (c *APIClient) ClearQuota(ctx context.Context) (*APIResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := ClearQuotaRequest{
		ComponentAppID: c.Client.GetConfig().ComponentAppID,
	}

	var result APIResponse
	url := fmt.Sprintf("%s?access_token=%s", APIClearQuotaURL, url.QueryEscape(componentToken.AccessToken))
	err = c.Client.MakeRequest(ctx, "POST", url, request, &result)
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
		ComponentAppID:  c.Client.GetConfig().ComponentAppID,
		AuthorizerAppID: authorizerAppID,
	}

	var result GetApiQuotaResponse
	url := fmt.Sprintf("%s?access_token=%s", APIGetApiQuotaURL, url.QueryEscape(componentToken.AccessToken))
	err = c.Client.MakeRequest(ctx, "POST", url, request, &result)
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
	url := fmt.Sprintf("%s?access_token=%s", APIGetRidInfoURL, url.QueryEscape(componentToken.AccessToken))
	err = c.Client.MakeRequest(ctx, "POST", url, request, &result)
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
		ComponentAppID:     c.Client.GetConfig().ComponentAppID,
		ComponentAppSecret: c.Client.GetConfig().ComponentAppSecret,
	}

	var result APIResponse
	err := c.Client.MakeRequest(ctx, "POST", APIClearComponentQuotaURL, request, &result)
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
		ComponentAppID:  c.Client.GetConfig().ComponentAppID,
		AuthorizerAppID: authorizerAppID,
		OptionName:      optionName,
		OptionValue:     optionValue,
	}

	var result APIResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APISetAuthorizerOptionURL, url.QueryEscape(componentToken.AccessToken))
	err = c.Client.MakeRequest(ctx, "POST", url, request, &result)
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
		ComponentAppID:  c.Client.GetConfig().ComponentAppID,
		AuthorizerAppID: authorizerAppID,
		OptionName:      optionName,
	}

	var result GetAuthorizerOptionResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIGetAuthorizerOptionURL, url.QueryEscape(componentToken.AccessToken))
	err = c.Client.MakeRequest(ctx, "POST", url, request, &result)
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
	url := fmt.Sprintf("%s?access_token=%s", APIGetTemplateDraftListURL, url.QueryEscape(componentToken.AccessToken))
	err = c.Client.MakeRequest(ctx, "GET", url, nil, &result)
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
		DraftID:     draftID,
		TemplateType: templateType,
	}

	var result APIResponse
	url := fmt.Sprintf("%s?access_token=%s", APIAddToTemplateURL, url.QueryEscape(componentToken.AccessToken))
	err = c.Client.MakeRequest(ctx, "POST", url, request, &result)
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
	url := fmt.Sprintf("%s?access_token=%s", APIGetTemplateListURL, url.QueryEscape(componentToken.AccessToken))
	err = c.Client.MakeRequest(ctx, "GET", url, nil, &result)
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
	url := fmt.Sprintf("%s?access_token=%s", APIDeleteTemplateURL, url.QueryEscape(componentToken.AccessToken))
	err = c.Client.MakeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result
	}

	return &result, nil
}