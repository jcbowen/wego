package wego

import (
	"log"

	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/crypto"
	"github.com/jcbowen/wego/message"
	"github.com/jcbowen/wego/official_account"
	"github.com/jcbowen/wego/openplatform"
	"github.com/jcbowen/wego/storage"
	"github.com/jcbowen/wego/types"
)

// WeGo 微信开发封装库主结构体
type WeGo struct {
	// 开放平台客户端
	OpenPlatformClient *openplatform.Client

	// 公众号客户端
	OfficialAccountClient *official_account.Client
}

// New 创建新的WeGo实例，支持多种客户端配置和可选参数
// 同类型的客户端只能初始化一个，比如第一个参数是微信公众号，后面的不管还有几个公众号配置都忽略掉
// @param configParams ...any 配置参数，支持以下类型：
//   - openplatform.Config 或 *openplatform.Config: 开放平台配置
//   - official_account.Config 或 *official_account.Config: 公众号配置
//
// @param optParams ...any 可选参数，支持以下类型：
//   - debugger.LoggerInterface: 自定义日志器
//   - core.HTTPClient: 自定义HTTP客户端
//   - openplatform.EventHandler: 开放平台事件处理器
//
// @return *WeGo WeGo实例
func New(params ...any) *WeGo {
	wego := &WeGo{}

	// 分离配置参数和可选参数
	var configParams []any
	var optParams []any

	for _, param := range params {
		// 检查是否为配置参数（支持指针和非指针类型）
		if isConfigParam(param) {
			configParams = append(configParams, param)
		} else {
			optParams = append(optParams, param)
		}
	}

	for _, config := range configParams {
		switch cfg := config.(type) {
		case openplatform.Config, *openplatform.Config:
			// 如果还没有初始化过开放平台客户端，则初始化
			if wego.OpenPlatformClient == nil {
				// 统一转换为指针类型
				var openPlatformConfig *openplatform.Config
				if ptr, ok := cfg.(*openplatform.Config); ok {
					openPlatformConfig = ptr
				} else {
					configCopy := cfg.(openplatform.Config)
					openPlatformConfig = &configCopy
				}
				wego.OpenPlatformClient = openplatform.NewClient(openPlatformConfig, optParams...)
			}
		case official_account.Config, *official_account.Config:
			// 如果还没有初始化过公众号客户端，则初始化
			if wego.OfficialAccountClient == nil {
				// 统一转换为指针类型
				var officialAccountConfig *official_account.Config
				if ptr, ok := cfg.(*official_account.Config); ok {
					officialAccountConfig = ptr
				} else {
					configCopy := cfg.(official_account.Config)
					officialAccountConfig = &configCopy
				}
				wego.OfficialAccountClient = official_account.NewClient(officialAccountConfig, optParams...)
			}
		default:
			log.Printf("警告：不支持的配置类型 %T", cfg)
			// 忽略不支持的配置类型
			continue
		}
	}

	return wego
}

// isConfigParam 检查参数是否为配置参数（支持指针和非指针类型）
// @param param any 待检查的参数
// @return bool 如果是配置参数返回true，否则返回false
func isConfigParam(param any) bool {
	switch param.(type) {
	case openplatform.Config, *openplatform.Config,
		official_account.Config, *official_account.Config:
		return true
	default:
		return false
	}
}

// NewWithStorage 创建新的WeGo实例（使用自定义存储），支持可选参数
// 同类型的客户端只能初始化一个，比如第一个参数是微信公众号，后面的不管还有几个公众号配置都忽略掉
// @param storage storage.TokenStorage 自定义存储实例
// @param configParams ...any 配置参数，支持以下类型：
//   - openplatform.Config 或 *openplatform.Config: 开放平台配置
//   - official_account.Config 或 *official_account.Config: 公众号配置
//
// @param optParams ...any 可选参数，支持以下类型：
//   - debugger.LoggerInterface: 自定义日志器
//   - core.HTTPClient: 自定义HTTP客户端
//   - openplatform.EventHandler: 开放平台事件处理器
//
// @return *WeGo WeGo实例
func NewWithStorage(storage storage.TokenStorage, params ...any) *WeGo {
	wego := &WeGo{}

	// 分离配置参数和可选参数
	var configParams []any
	var optParams []any

	for _, param := range params {
		// 检查是否为配置参数（支持指针和非指针类型）
		if isConfigParam(param) {
			configParams = append(configParams, param)
		} else {
			optParams = append(optParams, param)
		}
	}

	for _, config := range configParams {
		switch cfg := config.(type) {
		case openplatform.Config, *openplatform.Config:
			// 如果还没有初始化过开放平台客户端，则初始化
			if wego.OpenPlatformClient == nil {
				// 统一转换为指针类型
				var openPlatformConfig *openplatform.Config
				if ptr, ok := cfg.(*openplatform.Config); ok {
					openPlatformConfig = ptr
				} else {
					configCopy := cfg.(openplatform.Config)
					openPlatformConfig = &configCopy
				}
				wego.OpenPlatformClient = openplatform.NewClientWithStorage(openPlatformConfig, storage, optParams...)
			}
		case official_account.Config, *official_account.Config:
			// 如果还没有初始化过公众号客户端，则初始化
			if wego.OfficialAccountClient == nil {
				// 统一转换为指针类型
				var officialAccountConfig *official_account.Config
				if ptr, ok := cfg.(*official_account.Config); ok {
					officialAccountConfig = ptr
				} else {
					configCopy := cfg.(official_account.Config)
					officialAccountConfig = &configCopy
				}
				wego.OfficialAccountClient = official_account.NewMPClientWithStorage(officialAccountConfig, storage, optParams...)
			}
		default:
			// 忽略不支持的配置类型
			continue
		}
	}

	return wego
}

// SetLogger 设置日志记录器
func (w *WeGo) SetLogger(log debugger.LoggerInterface) {
	if w.OpenPlatformClient != nil {
		w.OpenPlatformClient.SetLogger(log)
	}
	if w.OfficialAccountClient != nil {
		w.OfficialAccountClient.SetLogger(log)
	}
}

// SetHTTPClient 设置自定义HTTP客户端
func (w *WeGo) SetHTTPClient(client core.HTTPClient) {
	if w.OpenPlatformClient != nil {
		w.OpenPlatformClient.SetHTTPClient(client)
	}
	if w.OfficialAccountClient != nil {
		w.OfficialAccountClient.SetHTTPClient(client)
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
func (w *WeGo) OfficialAccountAPI() *official_account.APIClient {
	if w.OfficialAccountClient == nil {
		panic("未初始化公众号客户端")
	}
	return official_account.NewMPAPIClient(w.OfficialAccountClient)
}

// OfficialAccountMenu 返回公众号菜单相关功能
func (w *WeGo) OfficialAccountMenu() *official_account.MenuClient {
	if w.OfficialAccountClient == nil {
		panic("未初始化公众号客户端")
	}
	return official_account.NewMenuClient(w.OfficialAccountClient)
}

// OfficialAccountMessage 返回公众号消息相关功能
func (w *WeGo) OfficialAccountMessage() *official_account.MessageClient {
	if w.OfficialAccountClient == nil {
		panic("未初始化公众号客户端")
	}
	return official_account.NewMessageClient(w.OfficialAccountClient)
}

// OfficialAccountTemplate 返回公众号模板消息相关功能
func (w *WeGo) OfficialAccountTemplate() *official_account.TemplateClient {
	if w.OfficialAccountClient == nil {
		panic("未初始化公众号客户端")
	}
	return official_account.NewTemplateClient(w.OfficialAccountClient)
}

// OfficialAccountCustom 返回公众号客服消息相关功能
func (w *WeGo) OfficialAccountCustom() *official_account.CustomClient {
	if w.OfficialAccountClient == nil {
		panic("未初始化公众号客户端")
	}
	return official_account.NewCustomClient(w.OfficialAccountClient)
}

// OfficialAccountOAuth 返回公众号网页授权相关功能
func (w *WeGo) OfficialAccountOAuth() *official_account.OAuthClient {
	if w.OfficialAccountClient == nil {
		panic("未初始化公众号客户端")
	}
	return official_account.NewOAuthClient(w.OfficialAccountClient)
}

// OpenPlatformOAuth 返回开放平台网页授权相关功能（代公众号授权）
func (w *WeGo) OpenPlatformOAuth(redirectURI string) *openplatform.OAuthClient {
	if w.OpenPlatformClient == nil {
		panic("未初始化开放平台客户端")
	}
	authClient := openplatform.NewAuthClient(w.OpenPlatformClient)
	authorizerClient := authClient.NewAuthorizerClient("") // 这里需要传入具体的authorizerAppID
	return authorizerClient.GetOAuthClient(redirectURI)
}

// OfficialAccountMaterial 返回公众号素材管理相关功能
func (w *WeGo) OfficialAccountMaterial() *official_account.MaterialClient {
	if w.OfficialAccountClient == nil {
		panic("未初始化公众号客户端")
	}
	return official_account.NewMaterialClient(w.OfficialAccountClient)
}

// OfficialAccountSubscribe 返回公众号订阅消息相关功能
// 功能：获取订阅消息客户端实例，用于管理订阅消息相关功能
// 返回值：*official_account.SubscribeClient 订阅消息客户端指针
func (w *WeGo) OfficialAccountSubscribe() *official_account.SubscribeClient {
	if w.OfficialAccountClient == nil {
		panic("未初始化公众号客户端")
	}
	apiClient := official_account.NewMPAPIClient(w.OfficialAccountClient)
	return apiClient.GetSubscribeClient()
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
	OfficialAccountConfig         = official_account.Config
	OfficialAccountClient         = official_account.Client
	OfficialAccountAPIClient      = official_account.APIClient
	OfficialAccountMenuClient     = official_account.MenuClient
	OfficialAccountMessageClient  = official_account.MessageClient
	OfficialAccountTemplateClient = official_account.TemplateClient
	OfficialAccountCustomClient   = official_account.CustomClient
	OfficialAccountMaterialClient = official_account.MaterialClient

	// 微信公众号数据结构体
	OfficialAccountMenu                   = official_account.Menu
	OfficialAccountButton                 = official_account.Button
	OfficialAccountTemplateMessageRequest = official_account.TemplateMessageRequest
	OfficialAccountTemplateMessageData    = official_account.TemplateMessageData
	OfficialAccountMessageText            = official_account.MessageText
	OfficialAccountMessageImage           = official_account.MessageImage
	OfficialAccountMessageVoice           = official_account.MessageVoice
	OfficialAccountMessageVideo           = official_account.MessageVideo
	OfficialAccountMusicMessage           = official_account.MessageMusic
	OfficialAccountNewsMessage            = official_account.MessageNews
	OfficialAccountWXCardMessage          = official_account.MessageWXCard
	OfficialAccountMiniProgramPageMessage = official_account.MessageMiniProgramPage
	OfficialAccountNewsArticle            = official_account.NewsArticle
	UserInfo                              = types.OAuthUserInfoResponse
)
