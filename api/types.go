package api

import "fmt"

// APIResponse 微信API通用响应结构
type APIResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// Error 实现error接口
func (r *APIResponse) Error() string {
	return fmt.Sprintf("微信API错误[%d]: %s", r.ErrCode, r.ErrMsg)
}

// IsSuccess 检查API响应是否成功
func (r *APIResponse) IsSuccess() bool {
	return r.ErrCode == 0
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

// ClearQuotaRequest 重置API调用次数请求
type ClearQuotaRequest struct {
	ComponentAppID string `json:"component_appid"`
}

// GetApiQuotaRequest 查询API调用额度请求
type GetApiQuotaRequest struct {
	ComponentAppID  string `json:"component_appid"`
	AuthorizerAppID string `json:"authorizer_appid"`
}

// QuotaItem API调用额度项
type QuotaItem struct {
	API            string `json:"api"`
	DailyQuota     int    `json:"daily_quota"`
	DailyUsed      int    `json:"daily_used"`
	DailyRemaining int    `json:"daily_remaining"`
}

// GetApiQuotaResponse 查询API调用额度响应
type GetApiQuotaResponse struct {
	APIResponse
	Quota []QuotaItem `json:"quota"`
}

// GetRidInfoRequest 查询rid信息请求
type GetRidInfoRequest struct {
	RID string `json:"rid"`
}

// RidInfo rid信息
type RidInfo struct {
	Request struct {
		InvokeTime   string `json:"invoke_time"`
		CostInMs    int    `json:"cost_in_ms"`
		RequestURL  string `json:"request_url"`
		RequestBody string `json:"request_body"`
		Response    string `json:"response"`
		ClientIP    string `json:"client_ip"`
	} `json:"request"`
}

// GetRidInfoResponse 查询rid信息响应
type GetRidInfoResponse struct {
	APIResponse
	RidInfo RidInfo `json:"rid_info"`
}

// ClearComponentQuotaRequest 使用AppSecret重置第三方平台API调用次数请求
type ClearComponentQuotaRequest struct {
	ComponentAppID     string `json:"component_appid"`
	ComponentAppSecret string `json:"component_appsecret"`
}

// SetAuthorizerOptionRequest 设置授权方选项信息请求
type SetAuthorizerOptionRequest struct {
	ComponentAppID  string `json:"component_appid"`
	AuthorizerAppID string `json:"authorizer_appid"`
	OptionName      string `json:"option_name"`
	OptionValue     string `json:"option_value"`
}

// GetAuthorizerOptionRequest 获取授权方选项信息请求
type GetAuthorizerOptionRequest struct {
	ComponentAppID  string `json:"component_appid"`
	AuthorizerAppID string `json:"authorizer_appid"`
	OptionName      string `json:"option_name"`
}

// GetAuthorizerOptionResponse 获取授权方选项信息响应
type GetAuthorizerOptionResponse struct {
	APIResponse
	AuthorizerAppID string `json:"authorizer_appid"`
	OptionName      string `json:"option_name"`
	OptionValue     string `json:"option_value"`
}

// TemplateDraft 草稿箱模板
type TemplateDraft struct {
	CreateTime             int64  `json:"create_time"`
	UserVersion           string `json:"user_version"`
	UserDesc              string `json:"user_desc"`
	DraftID               int64  `json:"draft_id"`
}

// GetTemplateDraftListResponse 获取草稿箱列表响应
type GetTemplateDraftListResponse struct {
	APIResponse
	DraftList []TemplateDraft `json:"draft_list"`
}

// AddToTemplateRequest 将草稿添加到模板库请求
type AddToTemplateRequest struct {
	DraftID     int64  `json:"draft_id"`
	TemplateType int    `json:"template_type"`
}

// Template 模板库模板
type Template struct {
	CreateTime   int64  `json:"create_time"`
	UserVersion  string `json:"user_version"`
	UserDesc     string `json:"user_desc"`
	TemplateID   int64  `json:"template_id"`
}

// GetTemplateListResponse 获取模板列表响应
type GetTemplateListResponse struct {
	APIResponse
	TemplateList []Template `json:"template_list"`
}

// DeleteTemplateRequest 删除代码模板请求
type DeleteTemplateRequest struct {
	TemplateID int64 `json:"template_id"`
}