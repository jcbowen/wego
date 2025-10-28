package wego

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// 微信开放平台API地址常量
const (
	APIComponentTokenURL       = "https://api.weixin.qq.com/cgi-bin/component/api_component_token"
	APIPreAuthCodeURL          = "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode"
	APIQueryAuthURL            = "https://api.weixin.qq.com/cgi-bin/component/api_query_auth"
	APIAuthorizerTokenURL      = "https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token"
	APIGetAuthorizerInfoURL    = "https://api.weixin.qq.com/cgi-bin/component/api_get_authorizer_info"
	APIGetAuthorizerOptionURL  = "https://api.weixin.qq.com/cgi-bin/component/api_get_authorizer_option"
	APISetAuthorizerOptionURL  = "https://api.weixin.qq.com/cgi-bin/component/api_set_authorizer_option"
	APIStartPushTicketURL      = "https://api.weixin.qq.com/cgi-bin/component/api_start_push_ticket"
	APIClearQuotaURL           = "https://api.weixin.qq.com/cgi-bin/component/clear_quota"
	APIGetApiQuotaURL          = "https://api.weixin.qq.com/cgi-bin/component/api_get_authorizer_quota"
	APIGetRidInfoURL           = "https://api.weixin.qq.com/cgi-bin/component/api_query_auth"
	APIClearComponentQuotaURL  = "https://api.weixin.qq.com/cgi-bin/component/clear_quota"
	APIModifyServerDomainURL   = "https://api.weixin.qq.com/cgi-bin/component/modify_wxa_server_domain"
	APIGetJumpDomainFileURL    = "https://api.weixin.qq.com/cgi-bin/component/get_domain_confirmfile"
	APIModifyJumpDomainURL     = "https://api.weixin.qq.com/cgi-bin/component/modify_wxa_jump_domain"
	APIGetTemplateDraftListURL = "https://api.weixin.qq.com/wxa/gettemplatedraftlist"
	APIAddToTemplateURL        = "https://api.weixin.qq.com/wxa/addtotemplate"
	APIGetTemplateListURL      = "https://api.weixin.qq.com/wxa/gettemplatelist"
	APIDeleteTemplateURL       = "https://api.weixin.qq.com/wxa/deletetemplate"
)

// StartPushTicketRequest 启动票据推送服务请求参数
type StartPushTicketRequest struct {
	ComponentAppID     string `json:"component_appid"`
	ComponentAppSecret string `json:"component_appsecret"`
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

// AuthorizationInfo 授权信息
type AuthorizationInfo struct {
	AuthorizerAppID        string              `json:"authorizer_appid"`
	AuthorizerAccessToken  string              `json:"authorizer_access_token"`
	ExpiresIn              int                 `json:"expires_in"`
	AuthorizerRefreshToken string              `json:"authorizer_refresh_token"`
	FuncInfo               []FuncScopeCategory `json:"func_info"`
}

// FuncScopeCategory 授权给开发者的权限集
type FuncScopeCategory struct {
	FuncScopeCategoryID int `json:"funcscope_category"`
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

// AuthorizerInfo 授权方信息
type AuthorizerInfo struct {
	NickName        string           `json:"nick_name"`
	HeadImg         string           `json:"head_img"`
	ServiceTypeInfo ServiceTypeInfo  `json:"service_type_info"`
	VerifyTypeInfo  VerifyTypeInfo   `json:"verify_type_info"`
	UserName        string           `json:"user_name"`
	PrincipalName   string           `json:"principal_name"`
	BusinessInfo    BusinessInfo     `json:"business_info"`
	Alias           string           `json:"alias"`
	QrcodeURL       string           `json:"qrcode_url"`
	Signature       string           `json:"signature"`
	MiniProgramInfo *MiniProgramInfo `json:"MiniProgramInfo,omitempty"`
	RegisterType    int              `json:"register_type"`
	AccountStatus   int              `json:"account_status"`
	BasicConfig     *BasicConfigInfo `json:"basic_config,omitempty"`
}

// ServiceTypeInfo 账号类型
type ServiceTypeInfo struct {
	ID int `json:"id"`
}

// VerifyTypeInfo 认证类型
type VerifyTypeInfo struct {
	ID int `json:"id"`
}

// BusinessInfo 商业功能开通情况
type BusinessInfo struct {
	OpenStore int `json:"open_store"`
	OpenScan  int `json:"open_scan"`
	OpenPay   int `json:"open_pay"`
	OpenCard  int `json:"open_card"`
	OpenShake int `json:"open_shake"`
}

// MiniProgramInfo 小程序信息
type MiniProgramInfo struct {
	Network     NetworkInfo `json:"network"`
	Categories  []Category  `json:"categories"`
	VisitStatus int         `json:"visit_status"`
}

// NetworkInfo 网络信息
type NetworkInfo struct {
	RequestDomain   []string `json:"RequestDomain"`
	WsRequestDomain []string `json:"WsRequestDomain"`
	UploadDomain    []string `json:"UploadDomain"`
	DownloadDomain  []string `json:"DownloadDomain"`
	BizDomain       []string `json:"BizDomain"`
	UDPDomain       []string `json:"UDPDomain"`
}

// Category 小程序类目
type Category struct {
	First  string `json:"first"`
	Second string `json:"second"`
}

// BasicConfigInfo 基础配置信息
type BasicConfigInfo struct {
	IsPhoneConfigured bool `json:"is_phone_configured"`
	IsEmailConfigured bool `json:"is_email_configured"`
}

// ApiQuotaResponse API调用额度响应
type ApiQuotaResponse struct {
	APIResponse
	Quota struct {
		DailyLimit int `json:"daily_limit"`
		Used       int `json:"used"`
		Remain     int `json:"remain"`
	} `json:"quota"`
}

// RidInfoResponse rid信息响应
type RidInfoResponse struct {
	APIResponse
	Request struct {
		InvokeTime   int64  `json:"invoke_time"`
		CostInMs     int    `json:"cost_in_ms"`
		RequestURL   string `json:"request_url"`
		RequestBody  string `json:"request_body"`
		ResponseBody string `json:"response_body"`
		ClientIP     string `json:"client_ip"`
	} `json:"request"`
}

// StartPushTicket 启动票据推送服务
func (c *WegoClient) StartPushTicket(ctx context.Context) error {
	request := StartPushTicketRequest{
		ComponentAppID:     c.config.ComponentAppID,
		ComponentAppSecret: c.config.ComponentAppSecret,
	}

	var result struct {
		APIResponse
		ComponentAppID string `json:"component_appid"`
	}
	err := c.makeRequest(ctx, "POST", APIStartPushTicketURL, request, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result.APIResponse
	}

	c.logger.Infof("票据推送服务启动成功, component_appid: %s", result.ComponentAppID)
	return nil
}

// GetComponentAccessToken 获取第三方平台access_token
func (c *WegoClient) GetComponentAccessToken(ctx context.Context, verifyTicket string) (*ComponentAccessToken, error) {
	// 先从存储中获取
	if token, err := c.storage.GetComponentToken(ctx); err == nil && token != nil && token.ExpiresAt.After(time.Now()) {
		return token, nil
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

	err := c.makeRequest(ctx, "POST", APIComponentTokenURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	token := &ComponentAccessToken{
		AccessToken: result.ComponentAccessToken,
		ExpiresIn:   result.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}

	// 保存到存储
	if err := c.storage.SaveComponentToken(ctx, token); err != nil {
		c.logger.Errorf("保存组件令牌失败: %v", err)
	}

	return token, nil
}

// GetPreAuthCode 获取预授权码
func (c *WegoClient) GetPreAuthCode(ctx context.Context) (*PreAuthCodeResponse, error) {
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
	err = c.makeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	// 保存到存储
	preAuthCode := &PreAuthCode{
		PreAuthCode: result.PreAuthCode,
		ExpiresIn:   result.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}
	if err := c.storage.SavePreAuthCode(ctx, preAuthCode); err != nil {
		c.logger.Errorf("保存预授权码失败: %v", err)
	}

	return &result, nil
}

// QueryAuth 使用授权码换取授权信息
func (c *WegoClient) QueryAuth(ctx context.Context, authorizationCode string) (*QueryAuthResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := QueryAuthRequest{
		ComponentAppID:    c.config.ComponentAppID,
		AuthorizationCode: authorizationCode,
	}

	var result QueryAuthResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIQueryAuthURL, url.QueryEscape(componentToken.AccessToken))
	err = c.makeRequest(ctx, "POST", url, request, &result)
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
		c.logger.Errorf("缓存授权方token失败: %v", err)
	}

	return &result, nil
}

// RefreshAuthorizerToken 刷新授权方access_token
func (c *WegoClient) RefreshAuthorizerToken(ctx context.Context, authorizerAppID, refreshToken string) (*AuthorizerTokenResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := AuthorizerTokenRequest{
		ComponentAppID:         c.config.ComponentAppID,
		AuthorizerAppID:        authorizerAppID,
		AuthorizerRefreshToken: refreshToken,
	}

	var result AuthorizerTokenResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIAuthorizerTokenURL, url.QueryEscape(componentToken.AccessToken))
	err = c.makeRequest(ctx, "POST", url, request, &result)
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
		c.logger.Errorf("更新授权方token失败: %v", err)
	}

	return &result, nil
}

// GetAuthorizerInfo 获取授权方信息
func (c *WegoClient) GetAuthorizerInfo(ctx context.Context, authorizerAppID string) (*GetAuthorizerInfoResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := GetAuthorizerInfoRequest{
		ComponentAppID:  c.config.ComponentAppID,
		AuthorizerAppID: authorizerAppID,
	}

	var result GetAuthorizerInfoResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIGetAuthorizerInfoURL, url.QueryEscape(componentToken.AccessToken))
	err = c.makeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// makeRequest 发起HTTP请求的通用方法实现
func (c *WegoClient) makeRequest(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("序列化请求体失败: %v", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.logger.Debugf("发起请求: %s %s", method, url)
	if reqBody != nil {
		c.logger.Debugf("请求体: %s", string(reqBody))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应体失败: %v", err)
	}

	c.logger.Debugf("响应状态: %d", resp.StatusCode)
	c.logger.Debugf("响应体: %s", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	// 先检查是否是API错误响应
	apiResp, err := ParseAPIResponse(respBody)
	if err != nil {
		return err
	}

	if !apiResp.IsSuccess() {
		return apiResp
	}

	// 解析成功响应
	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("解析响应体失败: %v", err)
	}

	return nil
}

// GenerateAuthURL 生成授权链接
func (c *WegoClient) GenerateAuthURL(preAuthCode string, authType int, bizAppID string) string {
	baseURL := "https://mp.weixin.qq.com/cgi-bin/componentloginpage"
	params := url.Values{
		"component_appid": {c.config.ComponentAppID},
		"pre_auth_code":   {preAuthCode},
		"redirect_uri":    {c.config.RedirectURI},
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
func (c *WegoClient) ClearQuota(ctx context.Context) error {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return err
	}

	request := map[string]interface{}{
		"component_appid": c.config.ComponentAppID,
	}

	var result APIResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIClearQuotaURL, url.QueryEscape(componentToken.AccessToken))
	err = c.makeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// GetApiQuota 查询API调用额度
func (c *WegoClient) GetApiQuota(ctx context.Context, authorizerAppID, cgiPath string) (*ApiQuotaResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := map[string]interface{}{
		"component_appid":  c.config.ComponentAppID,
		"authorizer_appid": authorizerAppID,
		"cgi_path":         cgiPath,
	}

	var result ApiQuotaResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIGetApiQuotaURL, url.QueryEscape(componentToken.AccessToken))
	err = c.makeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetRidInfo 查询rid信息
func (c *WegoClient) GetRidInfo(ctx context.Context, rid string) (*RidInfoResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := map[string]interface{}{
		"component_appid": c.config.ComponentAppID,
		"rid":             rid,
	}

	var result RidInfoResponse
	url := fmt.Sprintf("%s?component_access_token=%s", APIGetRidInfoURL, url.QueryEscape(componentToken.AccessToken))
	err = c.makeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// ClearComponentQuotaByAppSecret 使用AppSecret重置第三方平台API调用次数
func (c *WegoClient) ClearComponentQuotaByAppSecret(ctx context.Context) error {
	request := map[string]interface{}{
		"component_appid":     c.config.ComponentAppID,
		"component_appsecret": c.config.ComponentAppSecret,
	}

	var result APIResponse
	err := c.makeRequest(ctx, "POST", APIClearComponentQuotaURL, request, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
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

// GetAuthorizerList 获取授权方列表
func (c *WegoClient) GetAuthorizerList(ctx context.Context, offset, count int) (*GetAuthorizerListResponse, error) {
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
	url := fmt.Sprintf("%s?component_access_token=%s", "https://api.weixin.qq.com/cgi-bin/component/api_get_authorizer_list", url.QueryEscape(componentToken.AccessToken))
	err = c.makeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// SetAuthorizerOptionInfoRequest 设置授权方选项信息请求参数
type SetAuthorizerOptionInfoRequest struct {
	ComponentAppID  string `json:"component_appid"`
	AuthorizerAppID string `json:"authorizer_appid"`
	OptionName      string `json:"option_name"`
	OptionValue     string `json:"option_value"`
}

// SetAuthorizerOptionInfo 设置授权方选项信息
func (c *WegoClient) SetAuthorizerOptionInfo(ctx context.Context, authorizerAppID, optionName, optionValue string) error {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return err
	}

	request := SetAuthorizerOptionInfoRequest{
		ComponentAppID:  c.config.ComponentAppID,
		AuthorizerAppID: authorizerAppID,
		OptionName:      optionName,
		OptionValue:     optionValue,
	}

	var result APIResponse
	url := fmt.Sprintf("%s?component_access_token=%s", "https://api.weixin.qq.com/cgi-bin/component/api_set_authorizer_option", url.QueryEscape(componentToken.AccessToken))
	err = c.makeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return err
	}

	if !result.IsSuccess() {
		return &result
	}

	return nil
}

// GetAuthorizerOptionInfoRequest 获取授权方选项信息请求参数
type GetAuthorizerOptionInfoRequest struct {
	ComponentAppID  string `json:"component_appid"`
	AuthorizerAppID string `json:"authorizer_appid"`
	OptionName      string `json:"option_name"`
}

// GetAuthorizerOptionInfoResponse 获取授权方选项信息响应
type GetAuthorizerOptionInfoResponse struct {
	APIResponse
	AuthorizerAppID string `json:"authorizer_appid"`
	OptionName      string `json:"option_name"`
	OptionValue     string `json:"option_value"`
}

// GetAuthorizerOptionInfo 获取授权方选项信息
func (c *WegoClient) GetAuthorizerOptionInfo(ctx context.Context, authorizerAppID, optionName string) (*GetAuthorizerOptionInfoResponse, error) {
	componentToken, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	request := GetAuthorizerOptionInfoRequest{
		ComponentAppID:  c.config.ComponentAppID,
		AuthorizerAppID: authorizerAppID,
		OptionName:      optionName,
	}

	var result GetAuthorizerOptionInfoResponse
	url := fmt.Sprintf("%s?component_access_token=%s", "https://api.weixin.qq.com/cgi-bin/component/api_get_authorizer_option", url.QueryEscape(componentToken.AccessToken))
	err = c.makeRequest(ctx, "POST", url, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// ClearCache 清除所有缓存
func (c *WegoClient) ClearCache() {
	ctx := context.Background()

	// 清除组件令牌
	if err := c.storage.DeleteComponentToken(ctx); err != nil {
		c.logger.Errorf("清除组件令牌失败: %v", err)
	}

	// 清除预授权码
	if err := c.storage.DeletePreAuthCode(ctx); err != nil {
		c.logger.Errorf("清除预授权码失败: %v", err)
	}

	// 清除所有授权方令牌
	if err := c.storage.ClearAuthorizerTokens(ctx); err != nil {
		c.logger.Errorf("清除授权方令牌失败: %v", err)
	}
}
