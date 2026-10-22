package wego

import (
	"fmt"
	"log"

	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/crypto"
	"github.com/jcbowen/wego/message"
	"github.com/jcbowen/wego/miniprogram"
	"github.com/jcbowen/wego/official_account"
	"github.com/jcbowen/wego/openplatform"
	"github.com/jcbowen/wego/payment"
	"github.com/jcbowen/wego/storage"
	"github.com/jcbowen/wego/types"
)

// WeGo 微信开发封装库主结构体
type WeGo struct {
	// 开放平台客户端
	OpenPlatformClient *openplatform.Client

	// 公众号客户端
	OfficialAccountClient *official_account.Client

	// 微信小程序客户端
	MiniProgramClient *miniprogram.Client

	// 微信支付客户端
	PaymentClient *payment.Client
}

// New 创建新的WeGo实例，支持多种客户端配置和可选参数
// 同类型的客户端只能初始化一个，比如第一个参数是微信公众号，后面的不管还有几个公众号配置都忽略掉
//
// Deprecated: 任一客户端初始化失败时会直接 log.Panic 终止进程，依赖库不应越权决定宿主进程生死。
// 请改用 NewE。
//
// @param configParams ...any 配置参数，支持以下类型：
//   - openplatform.Config 或 *openplatform.Config: 开放平台配置
//   - official_account.Config 或 *official_account.Config: 公众号配置
//   - miniprogram.Config 或 *miniprogram.Config: 微信小程序配置
//   - payment.Config 或 *payment.Config: 微信支付配置
//
// @param optParams ...any 可选参数，支持以下类型：
//   - debugger.LoggerInterface: 自定义日志器
//   - core.HTTPClient: 自定义HTTP客户端
//   - openplatform.EventHandler: 开放平台事件处理器
//
// @return *WeGo WeGo实例
func New(params ...any) *WeGo {
	wego, err := NewE(params...)
	if err != nil {
		log.Panic("创建WeGo实例失败: " + err.Error())
	}
	return wego
}

// NewE 创建新的WeGo实例（出错返回 error），支持多种客户端配置和可选参数
// 同类型的客户端只能初始化一个，比如第一个参数是微信公众号，后面的不管还有几个公众号配置都忽略掉
//
// @param params ...any 配置参数与可选参数，说明同 New
//
// @return *WeGo WeGo实例
// @return error 任一客户端初始化失败时返回错误
func NewE(params ...any) (*WeGo, error) {
	return newWeGo(nil, params...)
}

// NewWithStorage 创建新的WeGo实例（使用自定义存储），支持可选参数
//
// Deprecated: 任一客户端初始化失败时会直接 log.Panic 终止进程，请改用 NewWithStorageE。
func NewWithStorage(tokenStorage storage.TokenStorage, params ...any) *WeGo {
	wego, err := NewWithStorageE(tokenStorage, params...)
	if err != nil {
		log.Panic("创建WeGo实例失败: " + err.Error())
	}
	return wego
}

// NewWithStorageE 创建新的WeGo实例（使用自定义存储，出错返回 error）
//
// @param tokenStorage storage.TokenStorage 自定义存储实例，不能为空
// @param params ...any 配置参数与可选参数，说明同 New
//
// @return *WeGo WeGo实例
// @return error 任一客户端初始化失败时返回错误
func NewWithStorageE(tokenStorage storage.TokenStorage, params ...any) (*WeGo, error) {
	if tokenStorage == nil {
		return nil, fmt.Errorf("存储不能为空")
	}
	return newWeGo(tokenStorage, params...)
}

// newWeGo WeGo实例的公共构建逻辑
// tokenStorage 为 nil 时使用各客户端的默认文件存储
func newWeGo(tokenStorage storage.TokenStorage, params ...any) (*WeGo, error) {
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
				client, err := buildClient[openplatform.Client](tokenStorage, toConfigPtr[openplatform.Config](cfg), optParams...)
				if err != nil {
					return nil, fmt.Errorf("初始化开放平台客户端失败: %w", err)
				}
				wego.OpenPlatformClient = client
			}
		case official_account.Config, *official_account.Config:
			// 如果还没有初始化过公众号客户端，则初始化
			if wego.OfficialAccountClient == nil {
				client, err := buildClient[official_account.Client](tokenStorage, toConfigPtr[official_account.Config](cfg), optParams...)
				if err != nil {
					return nil, fmt.Errorf("初始化公众号客户端失败: %w", err)
				}
				wego.OfficialAccountClient = client
			}
		case miniprogram.Config, *miniprogram.Config:
			// 如果还没有初始化过小程序客户端，则初始化
			if wego.MiniProgramClient == nil {
				client, err := buildClient[miniprogram.Client](tokenStorage, toConfigPtr[miniprogram.Config](cfg), optParams...)
				if err != nil {
					return nil, fmt.Errorf("初始化小程序客户端失败: %w", err)
				}
				wego.MiniProgramClient = client
			}
		case payment.Config, *payment.Config:
			// 如果还没有初始化过支付客户端，则初始化
			if wego.PaymentClient == nil {
				client, err := buildClient[payment.Client](tokenStorage, toConfigPtr[payment.Config](cfg), optParams...)
				if err != nil {
					return nil, fmt.Errorf("初始化支付客户端失败: %w", err)
				}
				wego.PaymentClient = client
			}
		default:
			log.Printf("警告：不支持的配置类型 %T", cfg)
			// 忽略不支持的配置类型
			continue
		}
	}

	return wego, nil
}

// buildClient 根据配置类型创建对应的客户端实例
// C 为客户端类型，T 为配置类型；tokenStorage 非 nil 时使用自定义存储，否则使用默认文件存储
func buildClient[C any, T any](tokenStorage storage.TokenStorage, config *T, opts ...any) (*C, error) {
	switch c := any(config).(type) {
	case *openplatform.Config:
		if tokenStorage != nil {
			client, err := openplatform.NewClientWithStorageE(c, tokenStorage, opts...)
			return castClient[C](client, err)
		}
		client, err := openplatform.NewClientE(c, opts...)
		return castClient[C](client, err)
	case *official_account.Config:
		if tokenStorage != nil {
			client, err := official_account.NewMPClientWithStorageE(c, tokenStorage, opts...)
			return castClient[C](client, err)
		}
		client, err := official_account.NewClientE(c, opts...)
		return castClient[C](client, err)
	case *miniprogram.Config:
		if tokenStorage != nil {
			client, err := miniprogram.NewClientWithStorageE(c, tokenStorage, opts...)
			return castClient[C](client, err)
		}
		client, err := miniprogram.NewClientE(c, opts...)
		return castClient[C](client, err)
	case *payment.Config:
		if tokenStorage != nil {
			client, err := payment.NewClientWithStorageE(c, tokenStorage, opts...)
			return castClient[C](client, err)
		}
		client, err := payment.NewClientE(c, opts...)
		return castClient[C](client, err)
	default:
		return nil, fmt.Errorf("不支持的配置类型: %T", config)
	}
}

// castClient 将 any 类型的客户端断言为目标类型
func castClient[C any](client any, err error) (*C, error) {
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*C)
	if !ok {
		return nil, fmt.Errorf("客户端类型断言失败: %T", client)
	}
	return typed, nil
}

// toConfigPtr 将配置参数统一转换为指针类型
func toConfigPtr[T any](cfg any) *T {
	if ptr, ok := cfg.(*T); ok {
		return ptr
	}
	configCopy := cfg.(T)
	return &configCopy
}

// isConfigParam 检查参数是否为配置参数（支持指针和非指针类型）
// @param param any 待检查的参数
// @return bool 如果是配置参数返回true，否则返回false
func isConfigParam(param any) bool {
	switch param.(type) {
	case openplatform.Config, *openplatform.Config,
		official_account.Config, *official_account.Config,
		miniprogram.Config, *miniprogram.Config,
		payment.Config, *payment.Config:
		return true
	default:
		return false
	}
}

// SetLogger 设置日志记录器
func (w *WeGo) SetLogger(log debugger.LoggerInterface) {
	if w.OpenPlatformClient != nil {
		w.OpenPlatformClient.SetLogger(log)
	}
	if w.OfficialAccountClient != nil {
		w.OfficialAccountClient.SetLogger(log)
	}
	if w.MiniProgramClient != nil {
		w.MiniProgramClient.SetLogger(log)
	}
	if w.PaymentClient != nil {
		w.PaymentClient.SetLogger(log)
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
	if w.MiniProgramClient != nil {
		w.MiniProgramClient.SetHTTPClient(client)
	}
	if w.PaymentClient != nil {
		w.PaymentClient.SetHTTPClient(client)
	}
}

// OpenPlatform 返回开放平台客户端（含代小程序代码管理、域名设置、push ticket 等能力）
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 OpenPlatformE。
func (w *WeGo) OpenPlatform() *openplatform.Client {
	openClient, err := w.OpenPlatformE()
	if err != nil {
		panic(err.Error())
	}
	return openClient
}

// OpenPlatformE 返回开放平台客户端（E 版：未初始化时返回 error）
func (w *WeGo) OpenPlatformE() (*openplatform.Client, error) {
	if w.OpenPlatformClient == nil {
		return nil, fmt.Errorf("未初始化开放平台客户端")
	}
	return w.OpenPlatformClient, nil
}

// OpenPlatformAuth 返回开放平台授权相关功能
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 OpenPlatformAuthE。
func (w *WeGo) OpenPlatformAuth() *openplatform.AuthClient {
	authClient, err := w.OpenPlatformAuthE()
	if err != nil {
		panic(err.Error())
	}
	return authClient
}

// OpenPlatformAuthE 返回开放平台授权相关功能（E 版：未初始化时返回 error）
func (w *WeGo) OpenPlatformAuthE() (*openplatform.AuthClient, error) {
	if w.OpenPlatformClient == nil {
		return nil, fmt.Errorf("未初始化开放平台客户端")
	}
	return openplatform.NewAuthClient(w.OpenPlatformClient), nil
}

// OpenPlatformMessage 返回开放平台消息处理相关功能
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 OpenPlatformMessageE。
func (w *WeGo) OpenPlatformMessage() *message.MessageClient {
	messageClient, err := w.OpenPlatformMessageE()
	if err != nil {
		panic(err.Error())
	}
	return messageClient
}

// OpenPlatformMessageE 返回开放平台消息处理相关功能（E 版：未初始化时返回 error）
func (w *WeGo) OpenPlatformMessageE() (*message.MessageClient, error) {
	if w.OpenPlatformClient == nil {
		return nil, fmt.Errorf("未初始化开放平台客户端")
	}
	return message.NewMessageClient(w.OpenPlatformClient), nil
}

// OfficialAccountAPI 返回公众号API相关功能
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 OfficialAccountAPIE。
func (w *WeGo) OfficialAccountAPI() *official_account.APIClient {
	apiClient, err := w.OfficialAccountAPIE()
	if err != nil {
		panic(err.Error())
	}
	return apiClient
}

// OfficialAccountAPIE 返回公众号API相关功能（E 版：未初始化时返回 error）
func (w *WeGo) OfficialAccountAPIE() (*official_account.APIClient, error) {
	if w.OfficialAccountClient == nil {
		return nil, fmt.Errorf("未初始化公众号客户端")
	}
	return official_account.NewMPAPIClient(w.OfficialAccountClient), nil
}

// OfficialAccountMenu 返回公众号菜单相关功能
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 OfficialAccountMenuE。
func (w *WeGo) OfficialAccountMenu() *official_account.MenuClient {
	menuClient, err := w.OfficialAccountMenuE()
	if err != nil {
		panic(err.Error())
	}
	return menuClient
}

// OfficialAccountMenuE 返回公众号菜单相关功能（E 版：未初始化时返回 error）
func (w *WeGo) OfficialAccountMenuE() (*official_account.MenuClient, error) {
	if w.OfficialAccountClient == nil {
		return nil, fmt.Errorf("未初始化公众号客户端")
	}
	return official_account.NewMenuClient(w.OfficialAccountClient), nil
}

// OfficialAccountMessage 返回公众号消息相关功能
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 OfficialAccountMessageE。
func (w *WeGo) OfficialAccountMessage() *official_account.MessageClient {
	messageClient, err := w.OfficialAccountMessageE()
	if err != nil {
		panic(err.Error())
	}
	return messageClient
}

// OfficialAccountMessageE 返回公众号消息相关功能（E 版：未初始化时返回 error）
func (w *WeGo) OfficialAccountMessageE() (*official_account.MessageClient, error) {
	if w.OfficialAccountClient == nil {
		return nil, fmt.Errorf("未初始化公众号客户端")
	}
	return official_account.NewMessageClient(w.OfficialAccountClient), nil
}

// OfficialAccountTemplate 返回公众号模板消息相关功能
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 OfficialAccountTemplateE。
func (w *WeGo) OfficialAccountTemplate() *official_account.TemplateClient {
	templateClient, err := w.OfficialAccountTemplateE()
	if err != nil {
		panic(err.Error())
	}
	return templateClient
}

// OfficialAccountTemplateE 返回公众号模板消息相关功能（E 版：未初始化时返回 error）
func (w *WeGo) OfficialAccountTemplateE() (*official_account.TemplateClient, error) {
	if w.OfficialAccountClient == nil {
		return nil, fmt.Errorf("未初始化公众号客户端")
	}
	return official_account.NewTemplateClient(w.OfficialAccountClient), nil
}

// OfficialAccountCustom 返回公众号客服消息相关功能
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 OfficialAccountCustomE。
func (w *WeGo) OfficialAccountCustom() *official_account.CustomClient {
	customClient, err := w.OfficialAccountCustomE()
	if err != nil {
		panic(err.Error())
	}
	return customClient
}

// OfficialAccountCustomE 返回公众号客服消息相关功能（E 版：未初始化时返回 error）
func (w *WeGo) OfficialAccountCustomE() (*official_account.CustomClient, error) {
	if w.OfficialAccountClient == nil {
		return nil, fmt.Errorf("未初始化公众号客户端")
	}
	return official_account.NewCustomClient(w.OfficialAccountClient), nil
}

// OfficialAccountOAuth 返回公众号网页授权相关功能
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 OfficialAccountOAuthE。
func (w *WeGo) OfficialAccountOAuth() *official_account.OAuthClient {
	oauthClient, err := w.OfficialAccountOAuthE()
	if err != nil {
		panic(err.Error())
	}
	return oauthClient
}

// OfficialAccountOAuthE 返回公众号网页授权相关功能（E 版：未初始化时返回 error）
func (w *WeGo) OfficialAccountOAuthE() (*official_account.OAuthClient, error) {
	if w.OfficialAccountClient == nil {
		return nil, fmt.Errorf("未初始化公众号客户端")
	}
	return official_account.NewOAuthClient(w.OfficialAccountClient), nil
}

// OpenPlatformOAuth 返回开放平台网页授权相关功能（代公众号授权）
//
// 参数:
//   - redirectURI: 授权后重定向地址
//   - authorizerAppID: 授权方公众号/小程序的 AppID，不能为空
//
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 OpenPlatformOAuthE。
func (w *WeGo) OpenPlatformOAuth(redirectURI, authorizerAppID string) *openplatform.OAuthClient {
	oauthClient, err := w.OpenPlatformOAuthE(redirectURI, authorizerAppID)
	if err != nil {
		panic(err.Error())
	}
	return oauthClient
}

// OpenPlatformOAuthE 返回开放平台网页授权相关功能（E 版：未初始化时返回 error）
//
// 参数:
//   - redirectURI: 授权后重定向地址
//   - authorizerAppID: 授权方公众号/小程序的 AppID，不能为空
//
// 返回:
//   - *openplatform.OAuthClient: 网页授权客户端
//   - error: 客户端未初始化或 authorizerAppID 为空时返回错误
func (w *WeGo) OpenPlatformOAuthE(redirectURI, authorizerAppID string) (*openplatform.OAuthClient, error) {
	if w.OpenPlatformClient == nil {
		return nil, fmt.Errorf("未初始化开放平台客户端")
	}
	if authorizerAppID == "" {
		return nil, fmt.Errorf("authorizerAppID 不能为空，代公众号网页授权必须指定授权方 AppID")
	}
	authClient := openplatform.NewAuthClient(w.OpenPlatformClient)
	authorizerClient := authClient.NewAuthorizerClient(authorizerAppID)
	return authorizerClient.GetOAuthClient(redirectURI), nil
}

// OfficialAccountMaterial 返回公众号素材管理相关功能
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 OfficialAccountMaterialE。
func (w *WeGo) OfficialAccountMaterial() *official_account.MaterialClient {
	materialClient, err := w.OfficialAccountMaterialE()
	if err != nil {
		panic(err.Error())
	}
	return materialClient
}

// OfficialAccountMaterialE 返回公众号素材管理相关功能（E 版：未初始化时返回 error）
func (w *WeGo) OfficialAccountMaterialE() (*official_account.MaterialClient, error) {
	if w.OfficialAccountClient == nil {
		return nil, fmt.Errorf("未初始化公众号客户端")
	}
	return official_account.NewMaterialClient(w.OfficialAccountClient), nil
}

// OfficialAccountSubscribe 返回公众号订阅通知相关功能（兼容层）
// 功能：获取订阅通知客户端实例（方法转发到 SubscribeNotifyClient）
// 返回值：*official_account.SubscribeClient 兼容层客户端指针
// Deprecated: 客户端未初始化时会直接 panic 终止进程，且本入口指向兼容层，
// 请改用 OfficialAccountSubscribeNotifyE（返回 *official_account.SubscribeNotifyClient）。
func (w *WeGo) OfficialAccountSubscribe() *official_account.SubscribeClient {
	subscribeClient, err := w.OfficialAccountSubscribeE()
	if err != nil {
		panic(err.Error())
	}
	return subscribeClient
}

// OfficialAccountSubscribeE 返回公众号订阅通知相关功能（兼容层，E 版：未初始化时返回 error）
//
// Deprecated: 本入口指向兼容层，请改用 OfficialAccountSubscribeNotifyE
// （返回 *official_account.SubscribeNotifyClient，语义化方法名）。
func (w *WeGo) OfficialAccountSubscribeE() (*official_account.SubscribeClient, error) {
	if w.OfficialAccountClient == nil {
		return nil, fmt.Errorf("未初始化公众号客户端")
	}
	apiClient := official_account.NewMPAPIClient(w.OfficialAccountClient)
	return apiClient.GetSubscribeClient(), nil
}

// OfficialAccountSubscribeNotify 返回公众号订阅通知相关功能
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 OfficialAccountSubscribeNotifyE。
func (w *WeGo) OfficialAccountSubscribeNotify() *official_account.SubscribeNotifyClient {
	notifyClient, err := w.OfficialAccountSubscribeNotifyE()
	if err != nil {
		panic(err.Error())
	}
	return notifyClient
}

// OfficialAccountSubscribeNotifyE 返回公众号订阅通知相关功能（E 版：未初始化时返回 error）
// 功能：公众号「订阅通知」全套接口（类目/模板管理 + bizsend 发送）；
// 小程序订阅消息请使用 MiniProgramE 客户端的 SendSubscribeMessage。
//
// 返回:
//   - *official_account.SubscribeNotifyClient 订阅通知客户端
//   - error 未初始化公众号客户端时返回错误
func (w *WeGo) OfficialAccountSubscribeNotifyE() (*official_account.SubscribeNotifyClient, error) {
	if w.OfficialAccountClient == nil {
		return nil, fmt.Errorf("未初始化公众号客户端")
	}
	apiClient := official_account.NewMPAPIClient(w.OfficialAccountClient)
	return apiClient.GetSubscribeNotifyClient(), nil
}

// OfficialAccountUser 返回公众号用户管理相关功能
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 OfficialAccountUserE。
func (w *WeGo) OfficialAccountUser() *official_account.UserClient {
	userClient, err := w.OfficialAccountUserE()
	if err != nil {
		panic(err.Error())
	}
	return userClient
}

// OfficialAccountUserE 返回公众号用户管理相关功能（E 版：未初始化时返回 error）
//
// 返回:
//   - *official_account.UserClient: 用户管理客户端（用户信息、标签管理）
//   - error: 未初始化公众号客户端时返回错误
func (w *WeGo) OfficialAccountUserE() (*official_account.UserClient, error) {
	if w.OfficialAccountClient == nil {
		return nil, fmt.Errorf("未初始化公众号客户端")
	}
	return official_account.NewUserClient(w.OfficialAccountClient), nil
}

// OfficialAccount 返回公众号基础客户端
// 注意：JSSDK ticket（GetJSAPITicket/InvalidateJSAPITicket）挂在该客户端上。
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 OfficialAccountE。
func (w *WeGo) OfficialAccount() *official_account.Client {
	oaClient, err := w.OfficialAccountE()
	if err != nil {
		panic(err.Error())
	}
	return oaClient
}

// OfficialAccountE 返回公众号基础客户端（E 版：未初始化时返回 error）
//
// 返回:
//   - *official_account.Client: 公众号基础客户端
//   - error: 未初始化公众号客户端时返回错误
func (w *WeGo) OfficialAccountE() (*official_account.Client, error) {
	if w.OfficialAccountClient == nil {
		return nil, fmt.Errorf("未初始化公众号客户端")
	}
	return w.OfficialAccountClient, nil
}

// MiniProgram 返回微信小程序客户端
//
// 返回:
//   - *miniprogram.Client: 小程序客户端
//
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 MiniProgramE。
func (w *WeGo) MiniProgram() *miniprogram.Client {
	miniProgramClient, err := w.MiniProgramE()
	if err != nil {
		panic(err.Error())
	}
	return miniProgramClient
}

// MiniProgramE 返回微信小程序客户端（E 版：未初始化时返回 error）
func (w *WeGo) MiniProgramE() (*miniprogram.Client, error) {
	if w.MiniProgramClient == nil {
		return nil, fmt.Errorf("未初始化小程序客户端")
	}
	return w.MiniProgramClient, nil
}

// Payment 返回微信支付客户端
//
// 返回:
//   - *payment.Client: 支付客户端
//
// Deprecated: 客户端未初始化时会直接 panic 终止进程，请改用 PaymentE。
func (w *WeGo) Payment() *payment.Client {
	paymentClient, err := w.PaymentE()
	if err != nil {
		panic(err.Error())
	}
	return paymentClient
}

// PaymentE 返回微信支付客户端（E 版：未初始化时返回 error）
func (w *WeGo) PaymentE() (*payment.Client, error) {
	if w.PaymentClient == nil {
		return nil, fmt.Errorf("未初始化支付客户端")
	}
	return w.PaymentClient, nil
}

// Crypto 返回加密解密相关功能
func (w *WeGo) Crypto() *crypto.CryptoClient {
	return crypto.NewCryptoClient()
}

// Storage 返回存储相关功能
func (w *WeGo) Storage() *storage.StorageClient {
	return storage.NewStorageClient()
}

// 导出常用结构体类型
type (
	APIResponse = core.APIResponse

	// 存储相关类型
	TokenStorage          = storage.TokenStorage
	DBStorage             = storage.DBStorage
	SqliteStorage         = storage.SqliteStorage
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

	// 微信小程序相关类型
	MiniProgramConfig                 = miniprogram.Config
	MiniProgramClient                 = miniprogram.Client
	MiniProgramCode2SessionResponse   = miniprogram.Code2SessionResponse
	MiniProgramGetPhoneNumberResponse = miniprogram.GetPhoneNumberResponse
	MiniProgramPhoneInfo              = miniprogram.PhoneInfo
	MiniProgramUserInfo               = miniprogram.UserInfo

	// 微信支付相关类型
	PaymentConfig                       = payment.Config
	PaymentClient                       = payment.Client
	PaymentAmount                       = payment.Amount
	PaymentRefundAmount                 = payment.RefundAmount
	PaymentPayer                        = payment.Payer
	PaymentSceneInfo                    = payment.SceneInfo
	PaymentH5Info                       = payment.H5Info
	PaymentJSAPIRequest                 = payment.JSAPIRequest
	PaymentH5Request                    = payment.H5Request
	PaymentPrepayResponse               = payment.PrepayResponse
	PaymentH5Response                   = payment.H5Response
	PaymentOrder                        = payment.Order
	PaymentRefundRequest                = payment.RefundRequest
	PaymentRefundResponse               = payment.RefundResponse
	PaymentCloseOrderRequest            = payment.CloseOrderRequest
	PaymentJSAPIPayParams               = payment.JSAPIPayParams
	PaymentNativeRequest                = payment.NativeRequest
	PaymentNativeResponse               = payment.NativeResponse
	PaymentAppRequest                   = payment.AppRequest
	PaymentAppPayParams                 = payment.AppPayParams
	PaymentBillRequest                  = payment.BillRequest
	PaymentBillResponse                 = payment.BillResponse
	PaymentProfitSharingRequest         = payment.ProfitSharingRequest
	PaymentProfitSharingReceiver        = payment.ProfitSharingReceiver
	PaymentProfitSharingResponse        = payment.ProfitSharingResponse
	PaymentProfitSharingReceiverResult  = payment.ProfitSharingReceiverResult
	PaymentProfitSharingReturnRequest   = payment.ProfitSharingReturnRequest
	PaymentProfitSharingReturnResponse  = payment.ProfitSharingReturnResponse
	PaymentProfitSharingUnfreezeRequest = payment.ProfitSharingUnfreezeRequest
	PaymentProfitSharingAmountsResponse = payment.ProfitSharingAmountsResponse
	PaymentNotification                 = payment.Notification
	PaymentResource                     = payment.Resource
	PaymentNotificationHeaders          = payment.NotificationHeaders
	PaymentAPIError                     = payment.APIError
	PaymentAPIErrorDetail               = payment.APIErrorDetail
)
