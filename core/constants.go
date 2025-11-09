package core

// API相关公共常量
const (
	// 基础错误码
	ErrCodeSuccess           = 0
	ErrCodeInvalidCredential = 40001
	ErrCodeInvalidGrantType  = 40002
	ErrCodeInvalidOpenID     = 40003
	ErrCodeInvalidMediaType  = 40004
	ErrCodeInvalidFileSize   = 40005
	ErrCodeInvalidFileFormat = 40006
	ErrCodeInvalidParams     = 40013
	ErrCodeUnauthorized      = 48001
	ErrCodeAccessDenied      = 48004
	ErrCodeAPIQuotaExceeded  = 45009

	// 基础API域名
	BaseAPIURL       = "https://api.weixin.qq.com"
	OpenBaseURL      = "https://open.weixin.qq.com"
	MPBaseURL        = "https://mp.weixin.qq.com"
	PayAPIURL        = "https://api.mch.weixin.qq.com"
	ResURL           = "https://res.wx.qq.com"
	
	// 授权页面URL
	MobileAuthPageURL   = "https://open.weixin.qq.com/wxaopen/safe/bindcomponent"
	PCAuthPageURL       = "https://mp.weixin.qq.com/cgi-bin/componentloginpage"
	SubscribeAuthPageURL = "https://mp.weixin.qq.com/mp/subscribemsg"

	// 通用授权参数
	GrantTypeAuthorizationCode = "authorization_code"
	GrantTypeRefreshToken      = "refresh_token"
	GrantTypeClientCredential  = "client_credential"
	ResponseTypeCode           = "code"

	// OAuth作用域
	OAuthScopeBase     = "snsapi_base"
	OAuthScopeUserInfo = "snsapi_userinfo"

	// 通用操作结果
	SuccessResponse = "success"
	// 事件响应成功
	EventResponseSuccess = "success"

	// 消息类型常量
	MessageTypeText       = "text"
	MessageTypeImage      = "image"
	MessageTypeVoice      = "voice"
	MessageTypeVideo      = "video"
	MessageTypeShortVideo = "shortvideo"
	MessageTypeLocation   = "location"
	MessageTypeLink       = "link"
	MessageTypeEvent      = "event"

	// 事件类型常量
	EventTypeComponentVerifyTicket = "component_verify_ticket"
	EventTypeUnauthorized          = "unauthorized"
	EventTypeAuthorized            = "authorized"
	EventTypeUpdateAuthorized      = "updateauthorized"

	// 时间相关常量
	EventTimestampTolerance = 300 // 事件时间戳容忍范围（秒）
)