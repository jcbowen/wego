package types

import (
	"encoding/xml"
	"time"
)

// APIResponse 通用API响应结构
type APIResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// ComponentTokenResponse 获取第三方平台access_token响应
type ComponentTokenResponse struct {
	APIResponse
	ComponentAccessToken string `json:"component_access_token"` // 第三方平台access_token
	ExpiresIn            int    `json:"expires_in"`             // 过期时间，单位秒
}

// PreAuthCodeResponse 获取预授权码响应
type PreAuthCodeResponse struct {
	APIResponse
	PreAuthCode string `json:"pre_auth_code"` // 预授权码
	ExpiresIn   int    `json:"expires_in"`    // 有效期，单位秒
}

// QueryAuthResponse 使用授权码换取授权信息响应
type QueryAuthResponse struct {
	APIResponse
	AuthorizationInfo AuthorizationInfo `json:"authorization_info"` // 授权信息
}

// AuthorizationInfo 授权信息
type AuthorizationInfo struct {
	AuthorizerAppID        string `json:"authorizer_appid"`         // 授权方appid
	AuthorizerAccessToken  string `json:"authorizer_access_token"`  // 授权方access_token
	ExpiresIn              int    `json:"expires_in"`               // 过期时间，单位秒
	AuthorizerRefreshToken string `json:"authorizer_refresh_token"` // 刷新令牌
}

// AuthorizerInfoResponse 获取授权方信息响应
type AuthorizerInfoResponse struct {
	APIResponse
	AuthorizerInfo AuthorizerInfo `json:"authorizer_info"` // 授权方信息
}

// AuthorizerInfo 授权方信息
type AuthorizerInfo struct {
	NickName        string `json:"nick_name"` // 授权方昵称
	HeadImg         string `json:"head_img"`  // 授权方头像
	ServiceTypeInfo struct {
		ID int `json:"id"` // 服务类型ID
	} `json:"service_type_info"`
	VerifyTypeInfo struct {
		ID int `json:"id"` // 认证类型ID
	} `json:"verify_type_info"`
	UserName      string `json:"user_name"`
	PrincipalName string `json:"principal_name"`
	BusinessInfo  struct {
		OpenStore int `json:"open_store"` // 是否开通门店功能
		OpenScan  int `json:"open_scan"`  // 是否开通扫码功能
		OpenPay   int `json:"open_pay"`   // 是否开通支付功能
		OpenCard  int `json:"open_card"`  // 是否开通卡券功能
		OpenShake int `json:"open_shake"` // 是否开通摇一摇功能
	} `json:"business_info"`
	Alias     string `json:"alias"`      // 授权方别名
	QrcodeURL string `json:"qrcode_url"` // 二维码图片URL
}

// JSAPITicketResponse JS-SDK票据响应
type JSAPITicketResponse struct {
	APIResponse
	Ticket    string `json:"ticket"`     // JS-SDK票据
	ExpiresIn int    `json:"expires_in"` // 过期时间，单位秒
}

// OAuthAccessTokenResponse OAuth access_token响应
type OAuthAccessTokenResponse struct {
	APIResponse
	AccessToken  string `json:"access_token"`  // OAuth access_token
	ExpiresIn    int    `json:"expires_in"`    // 过期时间，单位秒
	RefreshToken string `json:"refresh_token"` // 刷新令牌
	OpenID       string `json:"openid"`        // 用户OpenID
	Scope        string `json:"scope"`         // 授权作用域
}

// OAuthUserInfoResponse OAuth用户信息响应
type OAuthUserInfoResponse struct {
	APIResponse
	OpenID     string   `json:"openid"`     // 用户OpenID
	Nickname   string   `json:"nickname"`   // 用户昵称
	Sex        int      `json:"sex"`        // 用户性别，1为男性，2为女性
	Province   string   `json:"province"`   // 用户所在省份
	City       string   `json:"city"`       // 用户所在城市
	Country    string   `json:"country"`    // 用户所在国家
	HeadImgURL string   `json:"headimgurl"` // 用户头像URL
	Privilege  []string `json:"privilege"`  // 用户特权信息
	UnionID    string   `json:"unionid"`    // 开放平台UnionID
}

// MessageEvent 消息事件基础结构
type MessageEvent struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`   // 接收方账号（收到消息时为开发者的AppID，发送消息时为接收方账号）
	FromUserName string   `xml:"FromUserName"` // 发送方账号（收到消息时为发送方账号，发送消息时为开发者的AppID）
	CreateTime   int64    `xml:"CreateTime"`   // 消息创建时间（时间戳）
	MsgType      string   `xml:"MsgType"`      // 消息类型
	Event        string   `xml:"Event"`        // 事件类型
}

// TextMessageEvent 文本消息事件
type TextMessageEvent struct {
	MessageEvent
	Content string `xml:"Content"` // 文本消息内容
	MsgID   int64  `xml:"MsgId"`   // 消息ID
}

// ImageMessageEvent 图片消息事件
type ImageMessageEvent struct {
	MessageEvent
	PicURL  string `xml:"PicUrl"`  // 图片消息的URL
	MediaID string `xml:"MediaId"` // 图片消息的媒体ID
	MsgID   int64  `xml:"MsgId"`   // 消息ID
}

// SubscribeEvent 关注事件
type SubscribeEvent struct {
	MessageEvent
	EventKey string `xml:"EventKey"` // 扫描带参数二维码事件时，此字段为二维码的参数值
	Ticket   string `xml:"Ticket"`   // 二维码的ticket，可用来换取二维码图片
}

// UnsubscribeEvent 取消关注事件
type UnsubscribeEvent struct {
	MessageEvent
}

// MenuClickEvent 菜单点击事件
type MenuClickEvent struct {
	MessageEvent
	EventKey string `xml:"EventKey"` // 菜单KEY值
}

// MenuViewEvent 菜单跳转事件
type MenuViewEvent struct {
	MessageEvent
	EventKey string `xml:"EventKey"` // 菜单跳转的URL
}

// LocationEvent 上报地理位置事件
type LocationEvent struct {
	MessageEvent
	Latitude  float64 `xml:"Latitude"`  // 地理位置纬度
	Longitude float64 `xml:"Longitude"` // 地理位置经度
	Precision float64 `xml:"Precision"` // 地理位置精度
}

// TemplateSendJobFinishEvent 模板消息发送任务完成事件
type TemplateSendJobFinishEvent struct {
	MessageEvent
	MsgID  int64  `xml:"MsgID"`  // 消息ID
	Status string `xml:"Status"` // 发送状态：success-成功，failed:user block-用户拒收，failed:system failed-其他原因失败
}

// JSSDKConfig JS-SDK配置
type JSSDKConfig struct {
	AppID     string `json:"appId" ini:"appId"`         // 应用ID
	Timestamp int64  `json:"timestamp" ini:"timestamp"` // 时间戳
	NonceStr  string `json:"nonceStr" ini:"nonceStr"`   // 随机字符串
	Signature string `json:"signature" ini:"signature"` // 签名
}

// JSSDKAPITicket JS-SDK API票据
type JSSDKAPITicket struct {
	Ticket    string    `json:"ticket"`     // JS-SDK票据
	ExpiresIn int       `json:"expires_in"` // 过期时间，单位秒
	ExpiresAt time.Time `json:"expires_at"` // 过期时间，时间戳
}

// MediaUploadResponse 媒体文件上传响应
type MediaUploadResponse struct {
	APIResponse
	Type      string `json:"type"`       // 媒体文件类型，分别有图片（image）、语音（voice）、视频（video）和文件（file）
	MediaID   string `json:"media_id"`   // 媒体文件ID
	CreatedAt int64  `json:"created_at"` // 创建时间，时间戳
}

// MaterialNewsItem 图文消息素材项
type MaterialNewsItem struct {
	Title              string `json:"title"`                 // 标题
	Author             string `json:"author"`                // 作者
	Digest             string `json:"digest"`                // 摘要
	Content            string `json:"content"`               // 内容
	ContentSourceURL   string `json:"content_source_url"`    // 内容来源URL
	ThumbMediaID       string `json:"thumb_media_id"`        // 封面媒体ID
	ShowCoverPic       int    `json:"show_cover_pic"`        // 是否显示封面，0为不显示，1为显示
	NeedOpenComment    int    `json:"need_open_comment"`     // 是否需要开启评论，0为否，1为是
	OnlyFansCanComment int    `json:"only_fans_can_comment"` // 是否仅粉丝可评论，0为否，1为是
}

// MaterialNewsResponse 图文消息素材响应
type MaterialNewsResponse struct {
	APIResponse
	MediaID string `json:"media_id"` // 图文消息素材ID
}

// CustomMessage 客服消息
type CustomMessage struct {
	ToUser          string                        `json:"touser"`                    // 接收方用户ID
	MsgType         string                        `json:"msgtype"`                   // 消息类型
	Text            *CustomMessageText            `json:"text,omitempty"`            // 文本消息内容
	Image           *CustomMessageImage           `json:"image,omitempty"`           // 图片消息内容
	Voice           *CustomMessageVoice           `json:"voice,omitempty"`           // 语音消息内容
	Video           *CustomMessageVideo           `json:"video,omitempty"`           // 视频消息内容
	Music           *CustomMessageMusic           `json:"music,omitempty"`           // 音乐消息内容
	News            *CustomMessageNews            `json:"news,omitempty"`            // 图文消息内容
	MpNews          *CustomMessageMpNews          `json:"mpnews,omitempty"`          // 图文消息（mpnews）内容
	WxCard          *CustomMessageWxCard          `json:"wxcard,omitempty"`          // 卡券消息内容
	MiniProgramPage *CustomMessageMiniProgramPage `json:"miniprogrampage,omitempty"` // 小程序卡片消息内容
}

// CustomMessageText 客服文本消息
type CustomMessageText struct {
	Content string `json:"content"` // 文本消息内容
}

// CustomMessageImage 客服图片消息
type CustomMessageImage struct {
	MediaID string `json:"media_id"` // 图片消息素材ID
}

// CustomMessageVoice 客服语音消息
type CustomMessageVoice struct {
	MediaID string `json:"media_id"`
}

// CustomMessageVideo 客服视频消息
type CustomMessageVideo struct {
	MediaID      string `json:"media_id"`       // 视频消息素材ID
	ThumbMediaID string `json:"thumb_media_id"` // 视频消息封面媒体ID
	Title        string `json:"title"`          // 视频消息标题
	Description  string `json:"description"`    // 视频消息描述
}

// CustomMessageMusic 客服音乐消息
type CustomMessageMusic struct {
	Title        string `json:"title"`          // 音乐消息标题
	Description  string `json:"description"`    // 音乐消息描述
	MusicURL     string `json:"musicurl"`       // 音乐消息URL
	HQMusicURL   string `json:"hqmusicurl"`     // 音乐消息高品质URL
	ThumbMediaID string `json:"thumb_media_id"` // 音乐消息封面媒体ID
}

// CustomMessageNews 客服图文消息
type CustomMessageNews struct {
	Articles []CustomMessageNewsArticle `json:"articles"` // 图文消息文章列表
}

// CustomMessageNewsArticle 客服图文消息文章
type CustomMessageNewsArticle struct {
	Title       string `json:"title"`       // 文章标题
	Description string `json:"description"` // 文章描述
	URL         string `json:"url"`         // 文章URL
	PicURL      string `json:"picurl"`      // 文章封面图片URL
}

// CustomMessageMpNews 客服图文消息（mpnews）
type CustomMessageMpNews struct {
	MediaID string `json:"media_id"` // 图文消息（mpnews）素材ID
}

// CustomMessageWxCard 客服卡券消息
type CustomMessageWxCard struct {
	CardID string `json:"card_id"` // 卡券消息卡券ID
}

// CustomMessageMiniProgramPage 客服小程序卡片消息
type CustomMessageMiniProgramPage struct {
	Title        string `json:"title"`          // 小程序卡片消息标题
	PagePath     string `json:"pagepath"`       // 小程序卡片消息页面路径
	ThumbMediaID string `json:"thumb_media_id"` // 小程序卡片消息封面媒体ID
}

// MenuButton 菜单按钮
type MenuButton struct {
	Type      string       `json:"type"`                 // 菜单按钮类型
	Name      string       `json:"name"`                 // 菜单按钮名称
	Key       string       `json:"key,omitempty"`        // 菜单按钮点击事件KEY值，用于消息接口推送，不超过64字节
	URL       string       `json:"url,omitempty"`        // 菜单按钮点击事件URL，用户点击菜单按钮后会跳转的URL，不超过1024字节
	SubButton []MenuButton `json:"sub_button,omitempty"` // 子菜单按钮列表
}

// MenuResponse 菜单响应
type MenuResponse struct {
	APIResponse
	Menu struct {
		Button []MenuButton `json:"button"` // 菜单按钮列表
	} `json:"menu"`
}

// TemplateMessage 模板消息
type TemplateMessage struct {
	ToUser      string                      `json:"touser"`                // 接收方用户ID
	TemplateID  string                      `json:"template_id"`           // 模板消息ID
	URL         string                      `json:"url"`                   // 模板消息点击事件URL，用户点击模板消息后会跳转的URL，不超过1024字节
	Data        map[string]interface{}      `json:"data"`                  // 模板消息数据，键值对形式，键为模板消息中定义的变量名，值为变量对应的值
	MiniProgram *TemplateMessageMiniProgram `json:"miniprogram,omitempty"` // 模板消息小程序，可选字段
}

// TemplateMessageMiniProgram 模板消息小程序
type TemplateMessageMiniProgram struct {
	AppID    string `json:"appid"`    // 模板消息小程序AppID
	PagePath string `json:"pagepath"` // 模板消息小程序页面路径
}

// TemplateMessageResponse 模板消息响应
type TemplateMessageResponse struct {
	APIResponse
	MsgID int64 `json:"msgid"` // 模板消息ID，用于标识模板消息的唯一性
}