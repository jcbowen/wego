package openplatform

import (
	"context"

	"github.com/jcbowen/wego/core"
)

// OAuthComponentAccessTokenRequest 第三方平台代公众号获取网页授权access_token请求参数
type OAuthComponentAccessTokenRequest struct {
	ComponentAppID       string `json:"component_appid"`        // 第三方平台appid
	ComponentAppSecret   string `json:"component_appsecret"`    // 第三方平台appsecret
	AppID                string `json:"appid"`                  // 授权公众号appid
	Code                 string `json:"code"`                   // 填写第一步获取的code参数
	GrantType            string `json:"grant_type"`             // 填写为authorization_code
	ComponentAccessToken string `json:"component_access_token"` // 第三方平台component_access_token
}

// OAuthComponentRefreshTokenRequest 第三方平台代公众号刷新网页授权access_token请求参数
type OAuthComponentRefreshTokenRequest struct {
	ComponentAppID       string `json:"component_appid"`        // 第三方平台appid
	AppID                string `json:"appid"`                  // 授权公众号appid
	RefreshToken         string `json:"refresh_token"`          // 填写通过access_token获取到的refresh_token参数
	GrantType            string `json:"grant_type"`             // 填写为refresh_token
	ComponentAccessToken string `json:"component_access_token"` // 第三方平台component_access_token
}

// AuthorizationInfo 授权信息
type AuthorizationInfo struct {
	AuthorizerAppID        string     `json:"authorizer_appid"`
	AuthorizerAccessToken  string     `json:"authorizer_access_token"`
	ExpiresIn              int        `json:"expires_in"` // authorizer_access_token的有效期
	AuthorizerRefreshToken string     `json:"authorizer_refresh_token"`
	FuncInfo               []FuncInfo `json:"func_info"`
}

// FuncScopeCategory 授权给开发者的权限集
type FuncScopeCategory struct {
	Id int `json:"id"`
}

type FuncInfo struct {
	FuncScopeCategory FuncScopeCategory `json:"funcscope_category"`
	ConfirmInfo       ConfirmInfo       `json:"confirm_info,omitempty"`
}

type ConfirmInfo struct {
	NeedConfirm    int `json:"need_confirm"`
	AlreadyConfirm int `json:"already_confirm"`
	CanConfirm     int `json:"can_confirm"`
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
	ChannelsInfo    *ChannelsInfo    `json:"channels_info,omitempty"`
}

// ServiceTypeInfo 账号类型
type ServiceTypeInfo struct {
	ID int `json:"id"`
}

// VerifyTypeInfo 认证类型
type VerifyTypeInfo struct {
	ID int `json:"id"`
}

// ChannelsInfo 视频号账号类型；如果该授权账号为视频号则返回该字段
type ChannelsInfo struct {
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
	core.APIResponse
	Quota []QuotaItem `json:"quota"`
}

// GetRidInfoRequest 查询rid信息请求
type GetRidInfoRequest struct {
	RID string `json:"rid"`
}

// RidInfo rid信息
type RidInfo struct {
	Request struct {
		InvokeTime  string `json:"invoke_time"`
		CostInMs    int    `json:"cost_in_ms"`
		RequestURL  string `json:"request_url"`
		RequestBody string `json:"request_body"`
		Response    string `json:"response"`
		ClientIP    string `json:"client_ip"`
	} `json:"request"`
}

// GetRidInfoResponse 查询rid信息响应
type GetRidInfoResponse struct {
	core.APIResponse
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
	core.APIResponse
	AuthorizerAppID string `json:"authorizer_appid"`
	OptionName      string `json:"option_name"`
	OptionValue     string `json:"option_value"`
}

// TemplateDraft 草稿箱模板
type TemplateDraft struct {
	CreateTime  int64  `json:"create_time"`
	UserVersion string `json:"user_version"`
	UserDesc    string `json:"user_desc"`
	DraftID     int64  `json:"draft_id"`
}

// GetTemplateDraftListResponse 获取草稿箱列表响应
type GetTemplateDraftListResponse struct {
	core.APIResponse
	DraftList []TemplateDraft `json:"draft_list"`
}

// AddToTemplateRequest 将草稿添加到模板库请求
type AddToTemplateRequest struct {
	DraftID      int64 `json:"draft_id"`
	TemplateType int   `json:"template_type"`
}

// Template 模板库模板
type Template struct {
	CreateTime  int64  `json:"create_time"`
	UserVersion string `json:"user_version"`
	UserDesc    string `json:"user_desc"`
	TemplateID  int64  `json:"template_id"`
}

// GetTemplateListResponse 获取模板列表响应
type GetTemplateListResponse struct {
	core.APIResponse
	TemplateList []Template `json:"template_list"`
}

// DeleteTemplateRequest 删除代码模板请求
type DeleteTemplateRequest struct {
	TemplateID int64 `json:"template_id"`
}

// AuthorizationEvent 授权变更事件基础结构
// 对应微信官方文档中的授权变更通知推送格式
// https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/authorize_event.html
// 所有授权变更事件共有的字段
// 接收POST请求后，只需直接返回字符串success
// 字段说明：
// - AppId: 第三方平台appid
// - CreateTime: 时间戳
// - InfoType: 通知类型
// - AuthorizerAppid: 公众号/服务号/小程序/微信小店/带货助手/视频号助手的appid
// - AuthorizationCode: 授权码，可用于获取授权信息（仅authorized和updateauthorized事件）
// - AuthorizationCodeExpiredTime: 授权码过期时间 单位秒（仅authorized和updateauthorized事件）
// - PreAuthCode: 预授权码（仅authorized和updateauthorized事件）
type AuthorizationEvent struct {
	AppId                        string `xml:"AppId" json:"appid"`
	CreateTime                   int64  `xml:"CreateTime" json:"create_time"`
	InfoType                     string `xml:"InfoType" json:"info_type"`
	AuthorizerAppid              string `xml:"AuthorizerAppid" json:"authorizer_appid"`
	AuthorizationCode            string `xml:"AuthorizationCode,omitempty" json:"authorization_code,omitempty"`
	AuthorizationCodeExpiredTime int64  `xml:"AuthorizationCodeExpiredTime,omitempty" json:"authorization_code_expired_time,omitempty"`
	PreAuthCode                  string `xml:"PreAuthCode,omitempty" json:"pre_auth_code,omitempty"`
}

// AuthorizedEvent 授权成功事件
// InfoType: authorized
// 当公众号/服务号/小程序/微信小店/带货助手/视频号助手对第三方平台进行授权时触发
type AuthorizedEvent struct {
	AuthorizationEvent
}

// UnauthorizedEvent 取消授权事件
// InfoType: unauthorized
// 当公众号/服务号/小程序/微信小店/带货助手/视频号助手取消对第三方平台的授权时触发
type UnauthorizedEvent struct {
	AuthorizationEvent
}

// UpdateAuthorizedEvent 授权更新事件
// InfoType: updateauthorized
// 当授权方更新授权时触发。如果更新授权时，授权的权限集没有发生变化，将不会触发授权更新通知
type UpdateAuthorizedEvent struct {
	AuthorizationEvent
}

// ComponentVerifyTicketEvent 验证票据推送事件
// InfoType: component_verify_ticket
// 微信服务器每隔10分钟推送component_verify_ticket到第三方平台的消息接收URL
// 根据微信官方文档<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/component_verify_ticket.html" index="0">0</mcreference>，
// 接收POST请求后只需直接返回字符串"success"
type ComponentVerifyTicketEvent struct {
	AuthorizationEvent
	ComponentVerifyTicket string `xml:"ComponentVerifyTicket" json:"component_verify_ticket"`
}

// EncodingAESKeyChangedEvent EncodingAESKey变更事件
// InfoType: encoding_aes_key_changed
// 当第三方平台在微信开放平台后台修改EncodingAESKey时，微信服务器会推送此事件
// 根据微信官方规范<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Message_encryption_and_decryption.html" index="1">1</mcreference>，
// 需要保存上一次的EncodingAESKey以确保平滑过渡
type EncodingAESKeyChangedEvent struct {
	AuthorizationEvent
	NewEncodingAESKey string `xml:"NewEncodingAESKey" json:"new_encoding_aes_key"`
}

// 授权变更事件类型常量
// 对应微信官方文档中的InfoType字段说明
// https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/authorize_event.html
const (
	// InfoTypeAuthorized 授权成功事件
	// 当公众号/服务号/小程序/微信小店/带货助手/视频号助手对第三方平台进行授权时触发
	InfoTypeAuthorized = "authorized"

	// InfoTypeUnauthorized 取消授权事件
	// 当公众号/服务号/小程序/微信小店/带货助手/视频号助手取消对第三方平台的授权时触发
	InfoTypeUnauthorized = "unauthorized"

	// InfoTypeUpdateAuthorized 授权更新事件
	// 当授权方更新授权时触发。如果更新授权时，授权的权限集没有发生变化，将不会触发授权更新通知
	InfoTypeUpdateAuthorized = "updateauthorized"

	// InfoTypeComponentVerifyTicket 验证票据推送事件
	// 微信服务器每隔10分钟推送component_verify_ticket到第三方平台的消息接收URL
	// https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/component_verify_ticket.html
	InfoTypeComponentVerifyTicket = "component_verify_ticket"
)

// EventHandler 事件处理器接口
type EventHandler interface {
	// HandleAuthorized 处理授权成功事件
	HandleAuthorized(ctx context.Context, event *AuthorizedEvent) error

	// HandleUnauthorized 处理取消授权事件
	HandleUnauthorized(ctx context.Context, event *UnauthorizedEvent) error

	// HandleUpdateAuthorized 处理授权更新事件
	HandleUpdateAuthorized(ctx context.Context, event *UpdateAuthorizedEvent) error

	// HandleComponentVerifyTicket 处理验证票据推送事件
	// 根据微信官方文档<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/component_verify_ticket.html" index="0">0</mcreference>，
	// 接收POST请求后只需直接返回字符串"success"
	HandleComponentVerifyTicket(ctx context.Context, event *ComponentVerifyTicketEvent) error

	// HandleEncodingAESKeyChanged 处理EncodingAESKey变更事件
	// 当第三方平台在微信开放平台后台修改EncodingAESKey时，微信服务器会推送此事件
	// 根据微信官方规范<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Message_encryption_and_decryption.html" index="1">1</mcreference>，
	// 需要保存上一次的EncodingAESKey以确保平滑过渡
	HandleEncodingAESKeyChanged(ctx context.Context, event *EncodingAESKeyChangedEvent) error
}

// DefaultEventHandler 默认事件处理器
type DefaultEventHandler struct{}

func (h *DefaultEventHandler) HandleAuthorized(ctx context.Context, event *AuthorizedEvent) error {
	return nil
}

func (h *DefaultEventHandler) HandleUnauthorized(ctx context.Context, event *UnauthorizedEvent) error {
	return nil
}

func (h *DefaultEventHandler) HandleUpdateAuthorized(ctx context.Context, event *UpdateAuthorizedEvent) error {
	return nil
}

func (h *DefaultEventHandler) HandleComponentVerifyTicket(ctx context.Context, event *ComponentVerifyTicketEvent) error {
	// 根据微信官方文档<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/component_verify_ticket.html" index="0">0</mcreference>，
	// 接收POST请求后只需直接返回字符串"success"
	// 这里可以添加票据存储逻辑，但即使处理失败也必须返回success
	return nil
}

func (h *DefaultEventHandler) HandleEncodingAESKeyChanged(ctx context.Context, event *EncodingAESKeyChangedEvent) error {
	return nil
}
