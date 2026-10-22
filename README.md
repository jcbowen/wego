# WeGo - 微信开发封装库

WeGo是一个模块化的微信开发封装库，专门为微信开放平台第三方平台开发和微信公众号开发设计。该库提供了完整的微信开放平台API封装、微信公众号API封装、消息处理、授权管理等功能。

## 特性

- 🏗️ **模块化设计** - 按功能模块组织代码，便于扩展和维护
- 🔐 **完整的API封装** - 支持微信开放平台和微信公众号所有核心API
- 📨 **消息处理** - 支持微信消息的接收、解析和处理
- 🔑 **授权管理** - 提供完整的授权流程管理
- 🔒 **安全加密** - 支持微信消息的加密和解密
- 💾 **存储抽象** - 支持多种存储后端（内存、文件、数据库等）
- 📚 **类型安全** - 完整的类型定义和错误处理
- 🔄 **稳定版Token** - 支持稳定版access_token获取和管理
- 📢 **订阅消息** - 支持订阅通知消息功能

## 项目结构

```
wego/
├── core/           # 核心配置和客户端
├── crypto/         # 加密解密功能
├── message/        # 消息处理功能
├── miniprogram/    # 小程序开发功能
├── official_account/ # 公众号开发功能
├── openplatform/   # 开放平台功能
├── payment/        # 微信支付 V3 功能
├── storage/        # 存储抽象层
├── types/          # 类型定义
└── doc/           # 技术文档
```

## 导出类型

```go
type (
    // API通用响应
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

    // 微信公众号相关类型
    OfficialAccountConfig         = official_account.Config
    OfficialAccountClient         = official_account.Client
    OfficialAccountAPIClient      = official_account.APIClient
    OfficialAccountMenuClient     = official_account.MenuClient
    OfficialAccountMessageClient  = official_account.MessageClient
    OfficialAccountTemplateClient = official_account.TemplateClient
    OfficialAccountCustomClient   = official_account.CustomClient
    OfficialAccountMaterialClient = official_account.MaterialClient

    // 微信公众号消息与数据结构
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

    // 微信支付相关类型
    PaymentConfig              = payment.Config
    PaymentClient              = payment.Client
    PaymentAmount              = payment.Amount
    PaymentRefundAmount        = payment.RefundAmount
    PaymentPayer               = payment.Payer
    PaymentSceneInfo           = payment.SceneInfo
    PaymentH5Info              = payment.H5Info
    PaymentJSAPIRequest        = payment.JSAPIRequest
    PaymentH5Request           = payment.H5Request
    PaymentPrepayResponse      = payment.PrepayResponse
    PaymentJSAPIPayParams      = payment.JSAPIPayParams
    PaymentH5Response          = payment.H5Response
    PaymentOrder               = payment.Order
    PaymentRefundRequest       = payment.RefundRequest
    PaymentRefundResponse      = payment.RefundResponse
    PaymentCloseOrderRequest   = payment.CloseOrderRequest
    PaymentNativeRequest       = payment.NativeRequest
    PaymentNativeResponse      = payment.NativeResponse
    PaymentAppRequest          = payment.AppRequest
    PaymentAppPayParams        = payment.AppPayParams
    PaymentBillRequest         = payment.BillRequest
    PaymentBillResponse        = payment.BillResponse
    PaymentProfitSharingRequest         = payment.ProfitSharingRequest
    PaymentProfitSharingReceiver        = payment.ProfitSharingReceiver
    PaymentProfitSharingResponse        = payment.ProfitSharingResponse
    PaymentProfitSharingReturnRequest   = payment.ProfitSharingReturnRequest
    PaymentProfitSharingReturnResponse  = payment.ProfitSharingReturnResponse
    PaymentProfitSharingUnfreezeRequest = payment.ProfitSharingUnfreezeRequest
    PaymentProfitSharingAmountsResponse = payment.ProfitSharingAmountsResponse
    PaymentNotification        = payment.Notification
    PaymentResource            = payment.Resource
    PaymentNotificationHeaders = payment.NotificationHeaders
    PaymentAPIError            = payment.APIError
    PaymentAPIErrorDetail      = payment.APIErrorDetail
)
```

## 快速开始

### 安装

```bash
go get github.com/jcbowen/wego
```

### 存储说明

WeGo库提供多种存储后端支持：

- **文件存储（默认）**：使用`./runtime/wego_storage`目录保存Token数据
- **内存存储**：适合开发和测试环境，重启后数据丢失
- **数据库存储**：支持GORM数据库存储
- **自定义存储**：实现`storage.TokenStorage`接口即可

**默认存储策略**：
- 默认使用文件存储，数据持久化到本地文件
- 文件存储创建失败时自动回退到系统临时目录（`os.TempDir()/wego_storage`）；临时目录也不可用时，E 版构造返回error，旧版`wego.New`/`wego.NewWithStorage`为兼容保留，失败时会panic
- 可通过`wego.NewWithStorage(storage, configs...)`指定自定义存储

### 稳定版Token说明

WeGo库支持稳定版access_token功能（`official_account.StableTokenClient`）：

- **普通模式**：优先使用缓存的token（内存缓存，5分钟过期余量），避免频繁刷新
- **强制刷新模式**：强制刷新获取新token；官方限制`force_refresh`每天20次且间隔30秒，库内已内置限频器（超限返回明确错误）
- **存储支持**：默认内存存储，可通过`Client.GetStableTokenClient().SetStableTokenStorage(s)`注入自定义持久化存储（实现`StableAccessTokenStorage`接口）

### E 版构造函数（推荐）

依赖库不应决定宿主进程生死。全库统一提供E版构造/获取方法，配置或依赖非法时返回`error`而非panic：

- 模块级：`official_account.NewClientE`、`miniprogram.NewClientE`、`openplatform.NewClientE`、`payment.NewClientE`（及各自`NewXxxWithStorageE`）
- 入口级：`wego.NewE(configs...)`、`wego.NewWithStorageE(storage, configs...)`
- `WeGo`的全部getter均有`XxxE`变体（如`OpenPlatformAuthE`、`OfficialAccountAPIE`、`PaymentE`），未初始化对应客户端时返回error
- 旧版`wego.New`/各getter仍可用（兼容层转发E版，失败时panic），已标`Deprecated`

### 基本使用

#### 只使用开放平台

```go
package main

import (
	"fmt"
	"github.com/jcbowen/wego"
)

func main() {
	// 配置微信开放平台参数
	openPlatformConfig := &wego.OpenPlatformConfig{
		ComponentAppID:     "your_component_app_id",
		ComponentAppSecret: "your_component_app_secret",
		ComponentToken:     "your_component_token",
		EncodingAESKey:     "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG", // 43位
		RedirectURI:        "your_redirect_uri",
	}

    // 创建WeGo实例（只初始化开放平台），推荐使用NewE获取校验错误
    wegoClient, err := wego.NewE(openPlatformConfig)
    if err != nil {
        panic(err)
    }

    // 使用开放平台功能（E版getter，未初始化时返回error）
    authClient, err := wegoClient.OpenPlatformAuthE()
    if err != nil {
        panic(err)
    }
    messageClient, err := wegoClient.OpenPlatformMessageE()
    if err != nil {
        panic(err)
    }
    // 需要直接调用开放平台API时可使用底层客户端：
    // wegoClient.OpenPlatformClient.GetComponentAccessToken(ctx, "...")

	fmt.Printf("开放平台客户端初始化成功！authClient就绪: %v, messageClient就绪: %v\n",
		authClient != nil, messageClient != nil)
}
```

#### 只使用公众号

```go
package main

import (
	"fmt"
	"github.com/jcbowen/wego"
)

func main() {
	// 配置公众号参数
	officialAccountConfig := &wego.OfficialAccountConfig{
		AppID:     "your_mp_app_id",
		AppSecret: "your_mp_app_secret",
		Token:     "your_mp_token",
		AESKey:    "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG", // 43位
	}

    // 创建WeGo实例（只初始化公众号），推荐使用NewE获取校验错误
    wegoClient, err := wego.NewE(officialAccountConfig)
    if err != nil {
        panic(err)
    }

	// 使用公众号功能（E版getter，未初始化时返回error）
	apiClient, err := wegoClient.OfficialAccountAPIE()
	if err != nil {
		panic(err)
	}
	menuClient, err := wegoClient.OfficialAccountMenuE()
	if err != nil {
		panic(err)
	}
	messageClient, err := wegoClient.OfficialAccountMessageE()
	if err != nil {
		panic(err)
	}

	fmt.Printf("公众号客户端初始化成功！apiClient: %v, menuClient: %v, messageClient: %v\n",
		apiClient != nil, menuClient != nil, messageClient != nil)
}
```

#### 同时使用开放平台和公众号

```go
package main

import (
	"fmt"
	"github.com/jcbowen/wego"
)

func main() {
	// 配置开放平台参数
	openPlatformConfig := &wego.OpenPlatformConfig{
		ComponentAppID:     "your_component_app_id",
		ComponentAppSecret: "your_component_app_secret",
		ComponentToken:     "your_component_token",
		EncodingAESKey:     "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG", // 43位
		RedirectURI:        "your_redirect_uri",
	}

	// 配置公众号参数
	officialAccountConfig := &wego.OfficialAccountConfig{
		AppID:     "your_mp_app_id",
		AppSecret: "your_mp_app_secret",
		Token:     "your_mp_token",
		AESKey:    "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG", // 43位
	}

    // 创建WeGo实例（同时初始化开放平台和公众号），推荐使用NewE获取校验错误
    wegoClient, err := wego.NewE(openPlatformConfig, officialAccountConfig)
    if err != nil {
        panic(err)
    }

	// 检查客户端是否初始化
	fmt.Printf("开放平台已初始化: %v\n", wegoClient.OpenPlatformClient != nil)
	fmt.Printf("公众号已初始化: %v\n", wegoClient.OfficialAccountClient != nil)

    // 使用开放平台功能（E版getter，未初始化时返回error）
    openPlatformAuth, err := wegoClient.OpenPlatformAuthE()
    if err != nil {
        panic(err)
    }
    openPlatformMessage, err := wegoClient.OpenPlatformMessageE()
    if err != nil {
        panic(err)
    }

	// 使用公众号功能
	officialAccountAPI, err := wegoClient.OfficialAccountAPIE()
	if err != nil {
		panic(err)
	}
	
	// 使用通用功能
	cryptoClient := wegoClient.Crypto()
	storageClient := wegoClient.Storage()

	fmt.Printf("所有客户端初始化成功！openPlatformAuth: %v, openPlatformMessage: %v, officialAccountAPI: %v\n",
		openPlatformAuth != nil, openPlatformMessage != nil, officialAccountAPI != nil)
	fmt.Printf("通用功能就绪！cryptoClient: %v, storageClient: %v\n",
		cryptoClient != nil, storageClient != nil)
}
```

#### 使用稳定版Token功能

```go
package main

import (
	"context"
	"fmt"
	"github.com/jcbowen/wego"
)

func main() {
	// 配置公众号参数
	config := &wego.OfficialAccountConfig{
		AppID:     "your_mp_app_id",
		AppSecret: "your_mp_app_secret",
		Token:     "your_mp_token",
		AESKey:    "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG", // 43位
	}

    // 创建WeGo实例
    wegoClient, err := wego.NewE(config)
    if err != nil {
        panic(err)
    }

    // 获取公众号API客户端（E版getter，未初始化时返回error）
    apiClient, err := wegoClient.OfficialAccountAPIE()
    if err != nil {
        panic(err)
    }

    // 使用稳定版token发送请求
	ctx := context.Background()
	
    // 普通模式获取稳定版token（通过底层Client访问稳定版客户端）
    tokenInfo, err := apiClient.Client.GetStableTokenClient().GetStableAccessTokenNormal(ctx)
	if err != nil {
		fmt.Printf("获取稳定版token失败: %v\n", err)
		return
	}
	fmt.Printf("稳定版token: %s, 过期时间: %v\n", tokenInfo.AccessToken, tokenInfo.ExpiresAt)

    // 使用稳定版token发送API请求
	var result map[string]interface{}
    err = apiClient.Client.GetStableTokenClient().MakeRequestWithStableToken(ctx, "GET", 
        "https://api.weixin.qq.com/cgi-bin/user/get", nil, &result)
	if err != nil {
		fmt.Printf("API请求失败: %v\n", err)
		return
	}
	fmt.Printf("API响应: %+v\n", result)
}
```

#### 使用订阅消息功能

```go
package main

import (
	"context"
	"fmt"
	"github.com/jcbowen/wego"
)

func main() {
	// 配置公众号参数
	config := &wego.OfficialAccountConfig{
		AppID:     "your_mp_app_id",
		AppSecret: "your_mp_app_secret",
		Token:     "your_mp_token",
		AESKey:    "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG", // 43位
	}

	// 创建WeGo实例
	wegoClient, err := wego.NewE(config)
	if err != nil {
		panic(err)
	}

	// 获取订阅消息客户端（E版getter，未初始化时返回error）
	subscribeClient, err := wegoClient.OfficialAccountSubscribeE()
	if err != nil {
		panic(err)
	}

	// 获取订阅通知分类列表
	ctx := context.Background()
	categoryResp, err := subscribeClient.GetCategory(ctx)
	if err != nil {
		fmt.Printf("获取分类列表失败: %v\n", err)
		return
	}
	fmt.Printf("分类列表: %+v\n", categoryResp.Data)

	// 获取公共模板标题列表
	titlesResp, err := subscribeClient.GetPubNewTemplateTitles(ctx, 1, 0, 10)
	if err != nil {
		fmt.Printf("获取模板标题失败: %v\n", err)
		return
	}
	fmt.Printf("模板标题列表: %+v\n", titlesResp.Data)
}
```

## 模块说明

### Core 模块

核心配置和客户端实现，包含：

- `Request` - 统一HTTP请求封装（含敏感参数日志脱敏）
- 令牌管理和HTTP客户端

### OpenPlatform 模块

微信开放平台功能，包含：

- `APIClient` - 开放平台API客户端
- `AuthClient` - 授权管理客户端
- API地址常量定义
- API响应结构体
- 授权信息数据结构
- 事件处理器接口

### OfficialAccount 模块

微信公众号开发功能，包含：

- `Client` - 公众号主客户端（token互斥锁 + 5分钟过期余量）
- `APIClient` - 公众号API客户端
- `MenuClient` - 菜单管理客户端
- `MessageClient` - 消息处理客户端（群发/预览）
- `TemplateClient` - 模板消息客户端
- `CustomClient` - 客服消息客户端
- `MaterialClient` - 素材管理客户端
- `StableTokenClient` - 稳定版access_token客户端（内置缓存与force_refresh限频）
- `SubscribeClient` - 订阅消息客户端
- `OAuthClient` - 网页授权客户端

**客户端获取方法**：
- `OfficialAccountAPI()` - 获取公众号API客户端
- `OfficialAccountMenu()` - 获取菜单管理客户端
- `OfficialAccountMessage()` - 获取消息处理客户端
- `OfficialAccountTemplate()` - 获取模板消息客户端
- `OfficialAccountCustom()` - 获取客服消息客户端
- `OfficialAccountMaterial()` - 获取素材管理客户端
- `OfficialAccountSubscribe()` - 获取订阅消息客户端
- 以上均有`XxxE`变体（未初始化时返回error而非panic，推荐）
- `Client.GetStableTokenClient()` - 获取稳定版Token客户端（公众号客户端方法）

### Payment 模块

微信支付 API V3 功能，包含：

- `Client` - 支付主客户端，内置请求签名与应答/回调验签
- 下单：`JSAPIPay`、`H5Pay`、`NativePay`、`AppPay`
- 查单：`QueryOrder`（按商户订单号）、`QueryOrderByTransactionID`（按微信支付订单号）
- 关单：`CloseOrder`
- 退款：`Refund`、`RefundQuery`
- 调起支付：`BuildJSAPIPayParams`（小程序/公众号）、`BuildAppPayParams`（App）
- 敏感信息：`EncryptSensitiveData`、`EncryptSerialNo`（RSAES-OAEP 加密）
- 账单：`ApplyTradeBill`、`ApplyFundFlowBill`、`DownloadBill`
- 分账：`ProfitSharing`、`ProfitSharingQuery`、`ProfitSharingReturn`、`ProfitSharingReturnQuery`、`ProfitSharingUnfreeze`、`ProfitSharingAmounts`
- 回调：`VerifyNotification`、`VerifyNotificationByHeaders`、`DecryptNotificationResource`
- 错误处理：`APIError` + `AsAPIError` / `IsAPIError`，可提取微信错误码做差异化处理
- 构造：`NewClientE`、`NewClientWithStorageE`（返回 error，推荐）；`NewClient`、`NewClientWithStorage` 已 Deprecated

**验签模式**：同时支持微信支付公钥（官方推荐）与平台证书，详见 [微信支付模块说明](doc/微信支付模块说明.md)。

### Message 模块

消息处理功能，包含：

- 消息类型常量
- 消息结构体定义
- 消息处理器接口

### Crypto 模块

加密解密功能，包含：

- AES密钥解码
- 消息加密和解密
- PKCS7填充处理

### Storage 模块

存储抽象层，包含：

- `TokenStorage` 接口 - 定义组件令牌、预授权码、验证票据、授权方令牌等存储方法
- `MemoryStorage` 内存存储实现 - 基于内存的临时存储
- `FileStorage` 文件存储实现（默认存储） - 基于本地文件的持久化存储
- `DBStorage` 数据库存储实现 - 基于数据库的持久化存储
- 支持自定义存储后端

**默认存储策略**：
- 默认存储为文件存储，使用`./runtime/wego_storage`目录保存Token数据
- 文件存储创建失败时自动回退到系统临时目录；临时目录也不可用时，E版构造返回error，旧版构造panic
- 可通过`wego.NewWithStorageE(storage, configs...)`指定自定义存储
- 稳定版token默认使用内存存储，可通过`SetStableTokenStorage`注入持久化实现

## 示例

查看 `doc/` 目录获取完整的使用示例和技术文档：

- [微信公众号开发库使用指南](doc/officialaccount/README.md)
- [微信支付模块说明](doc/微信支付模块说明.md)
- [授权流程技术说明](doc/openplatform/01-授权流程技术说明.md)
- [消息加解密技术介绍](doc/openplatform/06-消息加解密技术介绍.md)

## 文档

详细的技术文档请查看 `doc/` 目录：

- 授权流程技术说明
- 消息加解密说明
- Token生成介绍
- 消息与事件处理

## 依赖

- Go 1.26.0+
- gorm.io/gorm v1.31.2
- github.com/jcbowen/jcbaseGo v0.16.3

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！
