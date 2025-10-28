package wego

import (
	"github.com/jcbowen/wego/api"
	"github.com/jcbowen/wego/auth"
	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/crypto"
	"github.com/jcbowen/wego/message"
	"github.com/jcbowen/wego/officialaccount"
	"github.com/jcbowen/wego/storage"
)

// WeGo 微信开发封装库主结构体
type WeGo struct {
	Client *core.WegoClient
}

// NewWeGo 创建新的WeGo实例
func NewWeGo(config *core.WeGoConfig) *WeGo {
	client := core.NewWegoClient(config)
	return &WeGo{
		Client: client,
	}
}

// NewWeGoWithStorage 创建新的WeGo实例（使用自定义存储）
func NewWeGoWithStorage(config *core.WeGoConfig, storage storage.TokenStorage) *WeGo {
	client := core.NewWegoClientWithStorage(config, storage)
	return &WeGo{
		Client: client,
	}
}

// API 返回API相关功能
func (w *WeGo) API() *api.APIClient {
	return api.NewAPIClient(w.Client)
}

// Auth 返回授权相关功能
func (w *WeGo) Auth() *auth.AuthClient {
	return auth.NewAuthClient(w.Client)
}

// Message 返回消息处理相关功能
func (w *WeGo) Message() *message.MessageClient {
	return message.NewMessageClient(w.Client)
}

// Crypto 返回加密解密相关功能
func (w *WeGo) Crypto() *crypto.CryptoClient {
	return crypto.NewCryptoClient()
}

// Storage 返回存储相关功能
func (w *WeGo) Storage() *storage.StorageClient {
	return storage.NewStorageClient()
}

// MP 返回微信公众号开发相关功能
func (w *WeGo) MP(config *officialaccount.MPConfig) *officialaccount.MPClient {
	return officialaccount.NewMPClient(config)
}

// GetConfig 获取配置信息
func (w *WeGo) GetConfig() *core.WeGoConfig {
	return w.Client.GetConfig()
}

// 导出常用类型和常量
var (
// API相关 - 这些常量需要在实际使用时定义
// APIComponentTokenURL     = api.APIComponentTokenURL
// APICreatePreAuthCodeURL = api.APICreatePreAuthCodeURL
// APIQueryAuthURL         = api.APIQueryAuthURL
// APIComponentAPIURL      = api.APIComponentAPIURL
// APIComponentAPIURLV2    = api.APIComponentAPIURLV2

// 消息类型 - 这些常量需要在实际使用时定义
// MsgTypeText       = message.MsgTypeText
// MsgTypeImage      = message.MsgTypeImage
// MsgTypeVoice      = message.MsgTypeVoice
// MsgTypeVideo      = message.MsgTypeVideo
// MsgTypeShortVideo = message.MsgTypeShortVideo
// MsgTypeLocation   = message.MsgTypeLocation
// MsgTypeLink       = message.MsgTypeLink
// MsgTypeEvent      = message.MsgTypeEvent

// 事件类型 - 这些常量需要在实际使用时定义
// EventSubscribe   = message.EventSubscribe
// EventUnsubscribe = message.EventUnsubscribe
// EventScan        = message.EventScan
// EventLocation    = message.EventLocation
// EventClick       = message.EventClick
// EventView        = message.EventView
)

// 导出常用结构体类型
type (
	WeGoConfig            = core.WeGoConfig
	APIResponse           = core.APIResponse
	AuthorizationInfo     = core.AuthorizationInfo
	AuthorizerInfo        = api.AuthorizerInfo
	TokenStorage          = storage.TokenStorage
	MemoryStorage         = storage.MemoryStorage
	DBStorage             = storage.DBStorage
	FileStorage           = storage.FileStorage
	ComponentAccessToken  = storage.ComponentAccessToken
	PreAuthCode           = storage.PreAuthCode
	AuthorizerAccessToken = storage.AuthorizerAccessToken

	// 微信公众号开发相关类型
	MPConfig       = officialaccount.MPConfig
	MPClient       = officialaccount.MPClient
	MPAPIClient    = officialaccount.MPAPIClient
	MenuClient     = officialaccount.MenuClient
	MessageClient  = officialaccount.MessageClient
	TemplateClient = officialaccount.TemplateClient
	CustomClient   = officialaccount.CustomClient
	MaterialClient = officialaccount.MaterialClient

	// 微信公众号数据结构体
	Menu                   = officialaccount.Menu
	Button                 = officialaccount.Button
	SendTemplateMsgRequest = officialaccount.SendTemplateMsgRequest
	TemplateData           = officialaccount.TemplateData
	MPTextMessage          = officialaccount.TextMessage
	MPImageMessage         = officialaccount.ImageMessage
	VoiceMessage           = officialaccount.VoiceMessage
	VideoMessage           = officialaccount.VideoMessage
	MusicMessage           = officialaccount.MusicMessage
	NewsMessage            = officialaccount.NewsMessage
	MPNewsMessage          = officialaccount.MPNewsMessage
	WXCardMessage          = officialaccount.WXCardMessage
	MiniProgramPageMessage = officialaccount.MiniProgramPageMessage
	NewsArticle            = officialaccount.NewsArticle
)
