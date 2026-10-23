package miniprogram

// 微信小程序相关 API 地址
const (
	// Code2SessionURL 小程序登录凭证校验地址
	Code2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"

	// GetPhoneNumberURL 通过组件 code 换取手机号地址
	GetPhoneNumberURL = "https://api.weixin.qq.com/wxa/business/getuserphonenumber"

	// AccessTokenURL 获取小程序 access_token 地址
	AccessTokenURL = "https://api.weixin.qq.com/cgi-bin/token"

	// StableTokenURL 获取小程序稳定版 access_token 地址
	StableTokenURL = "https://api.weixin.qq.com/cgi-bin/stable_token"

	// GrantTypeAuthorizationCode 授权类型：authorization_code
	GrantTypeAuthorizationCode = "authorization_code"

	// GrantTypeClientCredential 授权类型：client_credential
	GrantTypeClientCredential = "client_credential"

	// SubscribeMessageSendURL 发送订阅消息地址
	SubscribeMessageSendURL = "https://api.weixin.qq.com/cgi-bin/message/subscribe/send"

	// MsgSecCheckURL 内容安全检测（文本）地址
	MsgSecCheckURL = "https://api.weixin.qq.com/wxa/msg_sec_check"

	// MediaCheckAsyncURL 内容安全检测（音视频异步）地址，检测结果通过消息推送异步告知
	MediaCheckAsyncURL = "https://api.weixin.qq.com/wxa/media_check_async"

	// GetWxaCodeURL 获取小程序码（带 path 参数，数量有限制）地址
	GetWxaCodeURL = "https://api.weixin.qq.com/wxa/getwxacode"

	// GetWxaCodeUnlimitURL 获取小程序码（带 scene 参数，数量不限）地址
	GetWxaCodeUnlimitURL = "https://api.weixin.qq.com/wxa/getwxacodeunlimit"

	// CreateWxaQrcodeURL 获取小程序二维码（旧接口，带 path 参数）地址
	CreateWxaQrcodeURL = "https://api.weixin.qq.com/cgi-bin/wxaapp/createwxaqrcode"
)
