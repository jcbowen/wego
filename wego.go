package wego

import (
	"github.com/jcbowen/wego/api"
	"github.com/jcbowen/wego/auth"
	"github.com/jcbowen/wego/crypto"
	"github.com/jcbowen/wego/message"
	"github.com/jcbowen/wego/officialaccount"
	"github.com/jcbowen/wego/openplatform"
	"github.com/jcbowen/wego/storage"
)

// WeGo 微信开发封装库主结构体
type WeGo struct {
	// 开放平台客户端
	OpenPlatformClient *openplatform.APIClient

	// 公众号客户端
	OfficialAccountClient *officialaccount.MPClient
}

// NewWeGo 创建新的WeGo实例，支持多种客户端配置
// 同类型的客户端只能初始化一个，比如第一个参数是微信公众号，后面的不管还有几个公众号配置都忽略掉
func NewWeGo(configs ...any) *WeGo {
	wego := &WeGo{}

	for _, config := range configs {
		switch cfg := config.(type) {
		case *openplatform.OpenPlatformConfig:
			// 如果还没有初始化过开放平台客户端，则初始化
			if wego.OpenPlatformClient == nil {
				wego.OpenPlatformClient = openplatform.NewAPIClient(cfg)
			}
		case *officialaccount.MPConfig:
			// 如果还没有初始化过公众号客户端，则初始化
			if wego.OfficialAccountClient == nil {
				wego.OfficialAccountClient = officialaccount.NewMPClient(cfg)
			}
		default:
			// 忽略不支持的配置类型
			continue
		}
	}

	return wego
}

// NewWeGoWithStorage 创建新的WeGo实例（使用自定义存储）
// 同类型的客户端只能初始化一个，比如第一个参数是微信公众号，后面的不管还有几个公众号配置都忽略掉
func NewWeGoWithStorage(storage storage.TokenStorage, configs ...any) *WeGo {
	wego := &WeGo{}

	for _, config := range configs {
		switch cfg := config.(type) {
		case *openplatform.OpenPlatformConfig:
			// 如果还没有初始化过开放平台客户端，则初始化
			if wego.OpenPlatformClient == nil {
				wego.OpenPlatformClient = openplatform.NewAPIClientWithStorage(cfg, storage)
			}
		case *officialaccount.MPConfig:
			// 如果还没有初始化过公众号客户端，则初始化
			if wego.OfficialAccountClient == nil {
				wego.OfficialAccountClient = officialaccount.NewMPClientWithStorage(cfg, storage)
			}
		default:
			// 忽略不支持的配置类型
			continue
		}
	}

	return wego
}

// OpenPlatformAPI 返回开放平台API相关功能
func (w *WeGo) OpenPlatformAPI() *api.APIClient {
	if w.OpenPlatformClient == nil {
		panic("未初始化开放平台客户端")
	}
	return api.NewAPIClient(w.OpenPlatformClient)
}

// OpenPlatformAuth 返回开放平台授权相关功能
func (w *WeGo) OpenPlatformAuth() *auth.AuthClient {
	if w.OpenPlatformClient == nil {
		panic("未初始化开放平台客户端")
	}
	return auth.NewAuthClient(w.OpenPlatformClient)
}

// OpenPlatformMessage 返回开放平台消息处理相关功能
func (w *WeGo) OpenPlatformMessage() *message.MessageClient {
	if w.OpenPlatformClient == nil {
		panic("未初始化开放平台客户端")
	}
	return message.NewMessageClient(w.OpenPlatformClient)
}

// OfficialAccountAPI 返回公众号API相关功能
func (w *WeGo) OfficialAccountAPI() *officialaccount.MPAPIClient {
	if w.OfficialAccountClient == nil {
		panic("未初始化公众号客户端")
	}
	return officialaccount.NewMPAPIClient(w.OfficialAccountClient)
}

// OfficialAccountMenu 返回公众号菜单相关功能
func (w *WeGo) OfficialAccountMenu() *officialaccount.MenuClient {
	if w.OfficialAccountClient == nil {
		panic("未初始化公众号客户端")
	}
	return officialaccount.NewMenuClient(w.OfficialAccountClient)
}

// OfficialAccountMessage 返回公众号消息相关功能
func (w *WeGo) OfficialAccountMessage() *officialaccount.MessageClient {
	if w.OfficialAccountClient == nil {
		panic("未初始化公众号客户端")
	}
	return officialaccount.NewMessageClient(w.OfficialAccountClient)
}

// OfficialAccountTemplate 返回公众号模板消息相关功能
func (w *WeGo) OfficialAccountTemplate() *officialaccount.TemplateClient {
	if w.OfficialAccountClient == nil {
		panic("未初始化公众号客户端")
	}
	return officialaccount.NewTemplateClient(w.OfficialAccountClient)
}

// OfficialAccountCustom 返回公众号客服消息相关功能
func (w *WeGo) OfficialAccountCustom() *officialaccount.CustomClient {
	if w.OfficialAccountClient == nil {
		panic("未初始化公众号客户端")
	}
	return officialaccount.NewCustomClient(w.OfficialAccountClient)
}

// OfficialAccountMaterial 返回公众号素材管理相关功能
func (w *WeGo) OfficialAccountMaterial() *officialaccount.MaterialClient {
	if w.OfficialAccountClient == nil {
		panic("未初始化公众号客户端")
	}
	return officialaccount.NewMaterialClient(w.OfficialAccountClient)
}

// Crypto 返回加密解密相关功能
func (w *WeGo) Crypto() *crypto.CryptoClient {
	return crypto.NewCryptoClient()
}

// Storage 返回存储相关功能
func (w *WeGo) Storage() *storage.StorageClient {
	return storage.NewStorageClient()
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
	// 开放平台相关类型
	OpenPlatformConfig    = openplatform.OpenPlatformConfig
	APIResponse           = openplatform.APIResponse
	AuthorizationInfo     = openplatform.AuthorizationInfo
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
	UserInfo               = auth.UserInfo
)
