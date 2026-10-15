package miniprogram

import "github.com/jcbowen/wego/core"

// Code2SessionResponse 小程序登录凭证校验响应
type Code2SessionResponse struct {
	core.APIResponse
	OpenID     string `json:"openid"`            // 用户唯一标识
	SessionKey string `json:"session_key"`       // 会话密钥
	UnionID    string `json:"unionid,omitempty"` // 用户在开放平台的唯一标识，可能为空
}

// PhoneInfo 手机号信息
type PhoneInfo struct {
	PhoneNumber     string `json:"phoneNumber"`     // 用户绑定的手机号（国外手机号会有国家码）
	PurePhoneNumber string `json:"purePhoneNumber"` // 没有国家码的手机号
	CountryCode     string `json:"countryCode"`     // 国家码
}

// GetPhoneNumberRequest 通过组件 code 换取手机号请求
type GetPhoneNumberRequest struct {
	Code   string `json:"code"`             // 前端 getPhoneNumber 组件返回的动态令牌
	OpenID string `json:"openid,omitempty"` // 可选，用户唯一标识，填入后微信会校验 openid 与 code 的绑定关系
}

// GetPhoneNumberResponse 通过组件 code 换取手机号响应
type GetPhoneNumberResponse struct {
	core.APIResponse
	PhoneInfo PhoneInfo `json:"phone_info"` // 手机号信息
}

// Watermark 微信开放数据水印
//
// 字段说明:
//   - AppID: 数据归属的小程序 appid
//   - Timestamp: 数据生成时间戳
//
// 注意:
//   - 官方建议服务端解密后校验 watermark.appid 是否与自身 appid 一致
//   - 可结合 timestamp 做时效性校验
//
// 文档:
//   - https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/signature.html
type Watermark struct {
	AppID     string `json:"appid"`     // 小程序 appid
	Timestamp int64  `json:"timestamp"` // 数据水印时间戳
}

// UserInfo 解密后的微信小程序用户信息
type UserInfo struct {
	OpenID    string    `json:"openId"`            // 用户 openId
	NickName  string    `json:"nickName"`          // 用户昵称
	Gender    int       `json:"gender"`            // 用户性别，0 未知，1 男，2 女
	City      string    `json:"city"`              // 用户所在城市
	Province  string    `json:"province"`          // 用户所在省份
	Country   string    `json:"country"`           // 用户所在国家
	AvatarURL string    `json:"avatarUrl"`         // 用户头像 URL
	UnionID   string    `json:"unionId,omitempty"` // 用户在开放平台的唯一标识
	Watermark Watermark `json:"watermark"`         // 数据水印
}
