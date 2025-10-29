package openplatform

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
	APIClearQuotaURL           = "https://api.weixin.qq.com/cgi-bin/clear_quota"
	APIGetApiQuotaURL          = "https://api.weixin.qq.com/cgi-bin/openapi/quota/get"
	APIGetRidInfoURL           = "https://api.weixin.qq.com/cgi-bin/openapi/rid/get"
	APIClearComponentQuotaURL  = "https://api.weixin.qq.com/cgi-bin/component/clear_quota"
	APIModifyServerDomainURL   = "https://api.weixin.qq.com/cgi-bin/component/modify_wxa_server_domain"
	APIGetJumpDomainFileURL    = "https://api.weixin.qq.com/cgi-bin/component/get_domain_confirmfile"
	APIModifyJumpDomainURL     = "https://api.weixin.qq.com/cgi-bin/component/modify_wxa_jump_domain"
	APIGetTemplateDraftListURL = "https://api.weixin.qq.com/wxa/gettemplatedraftlist"
	APIAddToTemplateURL        = "https://api.weixin.qq.com/wxa/addtotemplate"
	APIGetTemplateListURL      = "https://api.weixin.qq.com/wxa/gettemplatelist"
	APIDeleteTemplateURL       = "https://api.weixin.qq.com/wxa/deletetemplate"
)

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
)

// 事件处理相关常量
const (
	// EventResponseSuccess 事件处理成功响应
	// 根据微信官方文档要求，接收POST请求后只需直接返回字符串"success"
	EventResponseSuccess = "success"

	// EventTimestampTolerance 事件时间戳容忍范围（秒）
	// 用于防止重放攻击，建议设置为5分钟（300秒）
	EventTimestampTolerance = 300
)

// 授权类型常量
const (
	// AuthTypeOfficialAccount 公众号授权
	AuthTypeOfficialAccount = 1

	// AuthTypeMiniProgram 小程序授权
	AuthTypeMiniProgram = 2

	// AuthTypeBoth 公众号和小程序都展示
	AuthTypeBoth = 3

	// AuthTypeMiniProgramPromoter 小程序推客账号
	AuthTypeMiniProgramPromoter = 4

	// AuthTypeChannels 视频号账号
	AuthTypeChannels = 5

	// AuthTypeAll 全部
	AuthTypeAll = 6

	// AuthTypeLiveStreamingAssistant 带货助手账号
	AuthTypeLiveStreamingAssistant = 8
)
