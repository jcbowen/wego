package openplatform

// 微信开放平台API地址常量
const (
	APIComponentAccessTokenURL = "https://api.weixin.qq.com/sns/oauth2/component/access_token"
	APIComponentTokenURL       = "https://api.weixin.qq.com/cgi-bin/component/api_component_token"
	APIPreAuthCodeURL          = "https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode"
	APIQueryAuthURL            = "https://api.weixin.qq.com/cgi-bin/component/api_query_auth"
	APIAuthorizerTokenURL      = "https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token"
	APIGetAuthorizerInfoURL    = "https://api.weixin.qq.com/cgi-bin/component/api_get_authorizer_info"
	APIGetAuthorizerListURL    = "https://api.weixin.qq.com/cgi-bin/component/api_get_authorizer_list"
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

	// 小程序

	APIWxaGetTemplateDraftListURL = "https://api.weixin.qq.com/wxa/gettemplatedraftlist"
	APIWxaAddToTemplateURL        = "https://api.weixin.qq.com/wxa/addtotemplate"
	APIWxaGetTemplateListURL      = "https://api.weixin.qq.com/wxa/gettemplatelist"
	APIWxaDeleteTemplateURL       = "https://api.weixin.qq.com/wxa/deletetemplate"
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
