package wego

import (
	"log"

	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/crypto"
	"github.com/jcbowen/wego/message"
	"github.com/jcbowen/wego/officialaccount"
	"github.com/jcbowen/wego/openplatform"
	"github.com/jcbowen/wego/storage"
	"github.com/jcbowen/wego/types"
)

// WeGo 微信开发封装库主结构体
type WeGo struct {
	// 开放平台客户端
	OpenPlatformClient *openplatform.APIClient

	// 公众号客户端
	OfficialAccountClient *officialaccount.Client
}

// NewWeGo 创建新的WeGo实例，支持多种客户端配置
// 同类型的客户端只能初始化一个，比如第一个参数是微信公众号，后面的不管还有几个公众号配置都忽略掉
func NewWeGo(configs ...any) *WeGo {
	wego := &WeGo{}

	for _, config := range configs {
		switch cfg := config.(type) {
		case *openplatform.Config:
			// 如果还没有初始化过开放平台客户端，则初始化
			if wego.OpenPlatformClient == nil {
				wego.OpenPlatformClient = openplatform.NewAPIClient(cfg)
			}
		case *officialaccount.Config:
			// 如果还没有初始化过公众号客户端，则初始化
			if wego.OfficialAccountClient == nil {
				wego.OfficialAccountClient = officialaccount.NewMPClient(cfg)
			}
		default:
			log.Printf("警告：不支持的配置类型 %T", cfg)
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
		case *openplatform.Config:
			// 如果还没有初始化过开放平台客户端，则初始化
			if wego.OpenPlatformClient == nil {
				wego.OpenPlatformClient = openplatform.NewAPIClientWithStorage(cfg, storage)
			}
		case *officialaccount.Config:
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

// SetLogger 设置日志记录器
func (w *WeGo) SetLogger(logger debugger.LoggerInterface) {
	if w.OpenPlatformClient != nil {
		w.OpenPlatformClient.SetLogger(logger)
	}
	if w.OfficialAccountClient != nil {
		w.OfficialAccountClient.SetLogger(logger)
	}
}

// OpenPlatformAuth 返回开放平台授权相关功能
func (w *WeGo) OpenPlatformAuth() *openplatform.AuthClient {
	if w.OpenPlatformClient == nil {
		panic("未初始化开放平台客户端")
	}
	return openplatform.NewAuthClient(w.OpenPlatformClient)
}

// OpenPlatformMessage 返回开放平台消息处理相关功能
func (w *WeGo) OpenPlatformMessage() *message.MessageClient {
	if w.OpenPlatformClient == nil {
		panic("未初始化开放平台客户端")
	}
	return message.NewMessageClient(w.OpenPlatformClient)
}

// OfficialAccountAPI 返回公众号API相关功能
func (w *WeGo) OfficialAccountAPI() *officialaccount.APIClient {
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
	APIResponse = core.APIResponse

	// 存储相关类型
	TokenStorage          = storage.TokenStorage
	MemoryStorage         = storage.MemoryStorage
	DBStorage             = storage.DBStorage
	FileStorage           = storage.FileStorage
	ComponentAccessToken  = storage.ComponentAccessToken
	PreAuthCode           = storage.PreAuthCode
	AuthorizerAccessToken = storage.AuthorizerAccessToken

	// 开放平台相关类型
	OpenPlatformConfig            = openplatform.Config
	OpenPlatformAuthorizationInfo = openplatform.AuthorizationInfo
	OpenPlatformAuthorizerInfo    = openplatform.AuthorizerInfo

	// 微信公众号开发相关类型
	OfficialAccountConfig         = officialaccount.Config
	OfficialAccountClient         = officialaccount.Client
	OfficialAccountAPIClient      = officialaccount.APIClient
	OfficialAccountMenuClient     = officialaccount.MenuClient
	OfficialAccountMessageClient  = officialaccount.MessageClient
	OfficialAccountTemplateClient = officialaccount.TemplateClient
	OfficialAccountCustomClient   = officialaccount.CustomClient
	OfficialAccountMaterialClient = officialaccount.MaterialClient

	// 微信公众号数据结构体
	OfficialAccountMenu                   = officialaccount.Menu
	OfficialAccountButton                 = officialaccount.Button
	OfficialAccountTemplateMessageRequest = officialaccount.TemplateMessageRequest
	OfficialAccountTemplateMessageData    = officialaccount.TemplateMessageData
	OfficialAccountMessageText            = officialaccount.MessageText
	OfficialAccountMessageImage           = officialaccount.MessageImage
	OfficialAccountMessageVoice           = officialaccount.MessageVoice
	OfficialAccountMessageVideo           = officialaccount.MessageVideo
	OfficialAccountMusicMessage           = officialaccount.MessageMusic
	OfficialAccountNewsMessage            = officialaccount.MessageNews
	OfficialAccountWXCardMessage          = officialaccount.MessageWXCard
	OfficialAccountMiniProgramPageMessage = officialaccount.MessageMiniProgramPage
	OfficialAccountNewsArticle            = officialaccount.NewsArticle
	UserInfo                              = types.OAuthUserInfoResponse
)
