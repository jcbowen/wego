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
	ComponentAccessToken string `json:"component_access_token"`
	ExpiresIn            int    `json:"expires_in"`
}

// PreAuthCodeResponse 获取预授权码响应
type PreAuthCodeResponse struct {
	APIResponse
	PreAuthCode string `json:"pre_auth_code"`
	ExpiresIn   int    `json:"expires_in"`
}

// QueryAuthResponse 使用授权码换取授权信息响应
type QueryAuthResponse struct {
	APIResponse
	AuthorizationInfo AuthorizationInfo `json:"authorization_info"`
}

// AuthorizationInfo 授权信息
type AuthorizationInfo struct {
	AuthorizerAppID        string `json:"authorizer_appid"`
	AuthorizerAccessToken  string `json:"authorizer_access_token"`
	ExpiresIn              int    `json:"expires_in"`
	AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
}

// AuthorizerInfoResponse 获取授权方信息响应
type AuthorizerInfoResponse struct {
	APIResponse
	AuthorizerInfo AuthorizerInfo `json:"authorizer_info"`
}

// AuthorizerInfo 授权方信息
type AuthorizerInfo struct {
	NickName        string `json:"nick_name"`
	HeadImg         string `json:"head_img"`
	ServiceTypeInfo struct {
		ID int `json:"id"`
	} `json:"service_type_info"`
	VerifyTypeInfo struct {
		ID int `json:"id"`
	} `json:"verify_type_info"`
	UserName     string `json:"user_name"`
	PrincipalName string `json:"principal_name"`
	BusinessInfo struct {
		OpenStore int `json:"open_store"`
		OpenScan  int `json:"open_scan"`
		OpenPay   int `json:"open_pay"`
		OpenCard  int `json:"open_card"`
		OpenShake int `json:"open_shake"`
	} `json:"business_info"`
	Alias     string `json:"alias"`
	QrcodeURL string `json:"qrcode_url"`
}

// JSAPITicketResponse JS-SDK票据响应
type JSAPITicketResponse struct {
	APIResponse
	Ticket    string `json:"ticket"`
	ExpiresIn int    `json:"expires_in"`
}

// OAuthAccessTokenResponse OAuth access_token响应
type OAuthAccessTokenResponse struct {
	APIResponse
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
}

// OAuthUserInfoResponse OAuth用户信息响应
type OAuthUserInfoResponse struct {
	APIResponse
	OpenID     string `json:"openid"`
	Nickname   string `json:"nickname"`
	Sex        int    `json:"sex"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	HeadImgURL string `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string `json:"unionid"`
}

// MessageEvent 消息事件基础结构
type MessageEvent struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Event        string   `xml:"Event"`
}

// TextMessageEvent 文本消息事件
type TextMessageEvent struct {
	MessageEvent
	Content string `xml:"Content"`
	MsgID   int64  `xml:"MsgId"`
}

// ImageMessageEvent 图片消息事件
type ImageMessageEvent struct {
	MessageEvent
	PicURL  string `xml:"PicUrl"`
	MediaID string `xml:"MediaId"`
	MsgID   int64  `xml:"MsgId"`
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
	MsgID   int64  `xml:"MsgID"`   // 消息ID
	Status  string `xml:"Status"`  // 发送状态：success-成功，failed:user block-用户拒收，failed:system failed-其他原因失败
}

// JSSDKConfig JS-SDK配置
type JSSDKConfig struct {
	AppID     string `json:"appId"`
	Timestamp int64  `json:"timestamp"`
	NonceStr  string `json:"nonceStr"`
	Signature string `json:"signature"`
}

// JSSDKAPITicket JS-SDK API票据
type JSSDKAPITicket struct {
	Ticket    string    `json:"ticket"`
	ExpiresIn int       `json:"expires_in"`
	ExpiresAt time.Time `json:"expires_at"`
}

// MediaUploadResponse 媒体文件上传响应
type MediaUploadResponse struct {
	APIResponse
	Type      string `json:"type"`
	MediaID   string `json:"media_id"`
	CreatedAt int64  `json:"created_at"`
}

// MaterialNewsItem 图文消息素材项
type MaterialNewsItem struct {
	Title              string `json:"title"`
	Author             string `json:"author"`
	Digest             string `json:"digest"`
	Content            string `json:"content"`
	ContentSourceURL   string `json:"content_source_url"`
	ThumbMediaID       string `json:"thumb_media_id"`
	ShowCoverPic       int    `json:"show_cover_pic"`
	NeedOpenComment    int    `json:"need_open_comment"`
	OnlyFansCanComment int    `json:"only_fans_can_comment"`
}

// MaterialNewsResponse 图文消息素材响应
type MaterialNewsResponse struct {
	APIResponse
	MediaID string `json:"media_id"`
}

// CustomMessage 客服消息
type CustomMessage struct {
	ToUser  string                 `json:"touser"`
	MsgType string                 `json:"msgtype"`
	Text    *CustomMessageText     `json:"text,omitempty"`
	Image   *CustomMessageImage   `json:"image,omitempty"`
	Voice   *CustomMessageVoice   `json:"voice,omitempty"`
	Video   *CustomMessageVideo   `json:"video,omitempty"`
	Music   *CustomMessageMusic   `json:"music,omitempty"`
	News    *CustomMessageNews    `json:"news,omitempty"`
	MpNews  *CustomMessageMpNews  `json:"mpnews,omitempty"`
	WxCard  *CustomMessageWxCard  `json:"wxcard,omitempty"`
	MiniProgramPage *CustomMessageMiniProgramPage `json:"miniprogrampage,omitempty"`
}

// CustomMessageText 客服文本消息
type CustomMessageText struct {
	Content string `json:"content"`
}

// CustomMessageImage 客服图片消息
type CustomMessageImage struct {
	MediaID string `json:"media_id"`
}

// CustomMessageVoice 客服语音消息
type CustomMessageVoice struct {
	MediaID string `json:"media_id"`
}

// CustomMessageVideo 客服视频消息
type CustomMessageVideo struct {
	MediaID      string `json:"media_id"`
	ThumbMediaID string `json:"thumb_media_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
}

// CustomMessageMusic 客服音乐消息
type CustomMessageMusic struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	MusicURL     string `json:"musicurl"`
	HQMusicURL   string `json:"hqmusicurl"`
	ThumbMediaID string `json:"thumb_media_id"`
}

// CustomMessageNews 客服图文消息
type CustomMessageNews struct {
	Articles []CustomMessageNewsArticle `json:"articles"`
}

// CustomMessageNewsArticle 客服图文消息文章
type CustomMessageNewsArticle struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	PicURL      string `json:"picurl"`
}

// CustomMessageMpNews 客服图文消息（mpnews）
type CustomMessageMpNews struct {
	MediaID string `json:"media_id"`
}

// CustomMessageWxCard 客服卡券消息
type CustomMessageWxCard struct {
	CardID string `json:"card_id"`
}

// CustomMessageMiniProgramPage 客服小程序卡片消息
type CustomMessageMiniProgramPage struct {
	Title        string `json:"title"`
	PagePath     string `json:"pagepath"`
	ThumbMediaID string `json:"thumb_media_id"`
}

// MenuButton 菜单按钮
type MenuButton struct {
	Type      string       `json:"type"`
	Name      string       `json:"name"`
	Key       string       `json:"key,omitempty"`
	URL       string       `json:"url,omitempty"`
	SubButton []MenuButton `json:"sub_button,omitempty"`
}

// MenuResponse 菜单响应
type MenuResponse struct {
	APIResponse
	Menu struct {
		Button []MenuButton `json:"button"`
	} `json:"menu"`
}

// TemplateMessage 模板消息
type TemplateMessage struct {
	ToUser     string                 `json:"touser"`
	TemplateID string                 `json:"template_id"`
	URL        string                 `json:"url"`
	Data       map[string]interface{} `json:"data"`
	MiniProgram *TemplateMessageMiniProgram `json:"miniprogram,omitempty"`
}

// TemplateMessageMiniProgram 模板消息小程序
type TemplateMessageMiniProgram struct {
	AppID    string `json:"appid"`
	PagePath string `json:"pagepath"`
}

// TemplateMessageResponse 模板消息响应
type TemplateMessageResponse struct {
	APIResponse
	MsgID int64 `json:"msgid"`
}