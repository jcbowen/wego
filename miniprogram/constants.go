package miniprogram

// 微信小程序相关 API 地址
const (
	// Code2SessionURL 小程序登录凭证校验地址
	Code2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"

	// GetPhoneNumberURL 通过组件 code 换取手机号地址
	GetPhoneNumberURL = "https://api.weixin.qq.com/wxa/business/getuserphonenumber"

	// AccessTokenURL 获取小程序 access_token 地址
	AccessTokenURL = "https://api.weixin.qq.com/cgi-bin/token"

	// GrantTypeAuthorizationCode 授权类型：authorization_code
	GrantTypeAuthorizationCode = "authorization_code"

	// GrantTypeClientCredential 授权类型：client_credential
	GrantTypeClientCredential = "client_credential"
)
