# WeGo 技术文档索引

本目录包含微信开放平台第三方平台开发、微信公众号开发、小程序开发与微信支付的完整技术文档，基于 WeGo 库的实现。

## 文档列表

### 微信公众号开发
- [微信公众号开发库使用指南](./officialaccount/README.md) - WeGo微信公众号开发库完整使用说明
- [网页授权功能详解](./officialaccount/网页授权功能详解.md) - 微信公众号网页授权完整实现指南

### 微信小程序开发
- [小程序模块说明](./miniprogram/小程序模块说明.md) - 登录、手机号、订阅消息、内容安全、小程序码、稳定版 access_token、用户数据解密

### 微信支付
- [微信支付模块说明](./微信支付模块说明.md) - 支付 V3 接口调用、公钥与平台证书双验签模式、回调处理

### 微信开放平台开发
#### 基础概念
- [01-授权流程技术说明](./openplatform/01-授权流程技术说明.md) - 第三方平台授权流程详解
- [02-验证票据(component_verify_ticket)](./openplatform/02-验证票据.md) - 验证票据的接收和处理
- [03-授权变更通知推送](./openplatform/03-授权变更通知推送.md) - 授权状态变更的事件处理
- [04-Token生成介绍](./openplatform/04-Token生成介绍.md) - 各种Token的生成和管理
- [网页授权功能详解](./openplatform/网页授权功能详解.md) - 微信开放平台网页授权完整实现指南

### 网页授权专题文档
- [网页授权文档索引](./网页授权文档索引.md) - 微信网页授权完整文档索引和快速开始指南
- [网页授权错误码处理和安全注意事项](./网页授权错误码处理和安全注意事项.md) - 网页授权错误码详解和安全最佳实践
- [网页授权常见问题排查指南和最佳实践](./网页授权常见问题排查指南和最佳实践.md) - 网页授权常见问题解决方案和开发最佳实践

#### 消息处理
- [05-消息推送介绍](./openplatform/05-消息推送介绍.md) - 消息推送服务的配置和使用
- [06-消息加解密技术介绍](./openplatform/06-消息加解密技术介绍.md) - 加解密技术实现细节、密钥轮换与 SecureMessageProcessor

#### 接口调用
- [07-代调用接口介绍](./openplatform/07-代调用接口介绍.md) - 代公众号/小程序调用接口
- [08-代公众号网页授权](./openplatform/08-代公众号网页授权.md) - OAuth2.0网页授权实现
- [09-代公众号处理消息和事件](./openplatform/09-代公众号处理消息和事件.md) - 消息和事件处理
- [10-代公众号使用js sdk](./openplatform/10-代公众号使用js sdk.md) - 前端JS SDK集成

## 快速开始

### 1. 安装

```bash
go get github.com/jcbowen/wego
```

### 2. 初始化客户端

```go
package main

import (
    "fmt"
    "github.com/jcbowen/wego"
)

func main() {
    // 配置微信开放平台参数
    config := &wego.OpenPlatformConfig{
        ComponentAppID:     "your_component_app_id",
        ComponentAppSecret: "your_component_app_secret",
        ComponentToken:     "your_component_token",
        EncodingAESKey:     "your_encoding_aes_key", // 43位
        RedirectURI:        "your_redirect_uri",
    }

    // 创建WeGo实例（推荐 E 版构造：配置非法时返回 error 而非 panic）
    wegoClient, err := wego.NewE(config)
    if err != nil {
        panic(err)
    }

    // 使用各个功能模块（E 版 getter：客户端未初始化时返回 error 而非 panic）
    openPlatformAuth, err := wegoClient.OpenPlatformAuthE()
    if err != nil {
        panic(err)
    }
    openPlatformMessage, err := wegoClient.OpenPlatformMessageE()
    if err != nil {
        panic(err)
    }
    cryptoClient := wegoClient.Crypto()
    storageClient := wegoClient.Storage()

    fmt.Println("WeGo客户端初始化成功！")
    _ = openPlatformAuth
    _ = openPlatformMessage
    _ = cryptoClient
    _ = storageClient
}
```

> 依赖库不应决定宿主进程生死。全库统一提供 E 版构造/获取方法（返回 error），
> 旧版 `wego.New` / 各非 E 版 getter 已标注 `Deprecated`，仅为兼容保留。
> 详见根 [README](../README.md#e-版构造函数推荐)。

### 3. 配置存储

WeGo库支持多种存储方式：
- 文件存储（默认）：使用`./runtime/wego_storage`目录持久化Token，写入采用临时文件 + 重命名的原子写
- 内存存储：用于开发/测试，重启后数据丢失
- 数据库存储：基于GORM的持久化存储（多实例部署推荐）
- 自定义存储：实现`storage.TokenStorage`接口，并通过`wego.NewWithStorageE(storage, configs...)`使用

### 4. 处理消息和事件

WeGo提供了完整的消息处理器，支持：
- 验证票据接收
- 授权变更事件处理
- 用户消息代收
- 消息加解密（含密钥轮换）

## 功能特性

✅ **模块化设计** - 按功能模块组织代码，便于扩展和维护
✅ **完整的API封装** - 支持微信开放平台和微信公众号核心API
✅ **消息处理** - 支持微信消息的接收、解析和处理
✅ **授权管理** - 提供完整的授权流程管理
✅ **安全加密** - 支持微信消息的加密和解密，支持 EncodingAESKey 密钥轮换
✅ **存储抽象** - 支持多种存储后端（内存、文件、数据库等）
✅ **类型安全** - 完整的类型定义和错误处理
✅ **稳定版Token** - 支持稳定版access_token获取和管理（`StableTokenClient`）
✅ **订阅消息** - 支持订阅通知消息功能（`SubscribeNotifyClient`）
✅ **E版构造** - 全库统一提供返回 error 的构造函数与 getter，失败不再 panic
✅ **微信支付** - 支持支付 V3 下单、查单、关单、退款、账单、分账、商家转账与敏感信息加密，兼容微信支付公钥与平台证书两种验签模式
✅ **日志脱敏** - 统一 HTTP 请求封装内置敏感参数、请求体及 XML 敏感字段的日志脱敏

## 项目结构

```
wego/
├── core/            # 核心配置、HTTP请求封装（含日志脱敏）与通用类型
├── crypto/          # 加密解密功能
├── logger/          # 日志抽象与适配器
├── message/         # 消息处理功能
├── miniprogram/     # 小程序开发功能
├── official_account/ # 公众号开发功能
├── openplatform/    # 开放平台功能
├── payment/         # 微信支付 V3 功能
├── storage/         # 存储抽象层
├── types/           # 类型定义
└── doc/             # 技术文档
```

## 模块说明

### Core 模块
核心配置和客户端实现，包含：
- `Config` - 各模块配置结构体的公共基础
- `Request` - 统一 HTTP 请求封装（含敏感参数、请求体及 XML 敏感字段的日志脱敏）
- `FlexString` / `FlexBool` - 双形态 JSON 字段解析（兼容官方接口字段类型变化）
- 令牌管理和HTTP客户端

### OpenPlatform 模块
微信开放平台功能，包含：
- `APIClient` - 开放平台API客户端（通过 `OpenPlatformClient` 字段访问）
- `AuthClient` - 授权管理客户端（推荐通过 `OpenPlatformAuthE()` 方法获取）
- 代小程序代码管理、域名设置、事件回调处理（`EventHandler`，可选 `SetEventErrorCallback`）
- 密钥轮换：`encoding_aes_key_changed` 事件自动持久化旧密钥（`SavePrevEncodingAESKey`）
- API地址常量、API响应结构体、授权信息数据结构、事件处理器接口

### OfficialAccount 模块
微信公众号开发功能，包含：
- `Client` - 公众号主客户端（token互斥锁 + 5分钟过期余量 + JSAPI ticket缓存）
- `APIClient` - 公众号API客户端（推荐通过 `OfficialAccountAPIE()` 方法获取）
- `MenuClient` / `MessageClient` / `TemplateClient` / `CustomClient` / `MaterialClient` / `UserClient`
- `StableTokenClient` - 稳定版access_token客户端（内置缓存与 force_refresh 限频）
- `SubscribeNotifyClient` - 公众号订阅通知客户端（旧的 `SubscribeClient` 已标 `Deprecated`，转发新实现）
- `OAuthClient` - 网页授权客户端
- 带参数二维码、数据分析、客服账号头像上传

### MiniProgram 模块
微信小程序功能，包含：
- `Client` - 小程序客户端（access_token 自动缓存与刷新，互斥锁并发保护）
- 登录（`Code2Session`）、手机号（`GetPhoneNumber`）、订阅消息（`SendSubscribeMessage`）
- 内容安全（`MsgSecCheck`/`MsgSecCheckFull`/`MediaCheckAsync`）
- 小程序码（`GetWxaCode`/`GetWxaCodeUnlimit`/`CreateWxaQrcode`）
- 稳定版 access_token（`GetStableAccessToken`）、用户数据解密（`DecryptUserInfo`）

详见 [小程序模块说明](./miniprogram/小程序模块说明.md)。

### Payment 模块
微信支付 API V3 功能，包含：
- `Client` - 支付主客户端，内置请求签名与应答/回调验签
- 下单、查单、关单、退款、账单、分账、商家转账
- 回调验签与解密，兼容微信支付公钥与平台证书两种验签模式
- `APIError` 错误处理与 E 版构造

详见 [微信支付模块说明](./微信支付模块说明.md)。

### Message 模块
消息处理功能，包含：
- 消息类型常量、消息结构体定义、消息处理器接口
- `SecureMessageProcessor` - 公众号/第三方平台加密消息处理器（密钥轮换：解密成功用哪个 AESKey，回包即用该 key 加密；可选 `WithTokenStorage` 注入存储）

### Crypto 模块
加密解密功能，包含：
- AES密钥解码、消息加密和解密（`DecryptMsgWithKey`/`EncryptMsgWithKey` 支持密钥轮换）
- 回调签名校验（含时间戳容差校验，超时判定为无效请求）
- PKCS7填充处理、小程序客户端数据解密

### Logger 模块
日志抽象层，包含：
- `LoggerInterface` - 日志接口（与 jcbaseGo `debugger.LoggerInterface` 完全兼容）
- `DefaultLogger` - 不依赖外部组件的默认控制台实现（`NewDefaultLoggerInterface`/`NewLoggerWithLevel`/`NewLoggerWithFields`）
- `DebuggerLoggerAdapter` - debugger logger 适配器，及双向转换工具（`ConvertToDebuggerLogger`）
- 各模块客户端均可通过构造 opts 或 `SetLogger` 注入自定义日志器

### Storage 模块
存储抽象层，包含：
- `TokenStorage` 接口 - 定义组件令牌、预授权码、验证票据、授权方令牌等存储方法
- `MemoryStorage` 内存存储实现 - 基于内存的临时存储
- `FileStorage` 文件存储实现（默认存储） - 基于本地文件的持久化存储（写入采用临时文件 + 重命名的原子写）
- `DBStorage` 数据库存储实现 - 基于 GORM 的持久化存储（多实例部署推荐）
- 支持自定义存储后端

## 示例代码

每个文档都包含详细的示例代码，展示如何使用WeGo库实现相应的功能。根级快速开始示例见 [README](../README.md)。

## 注意事项

1. 请确保在微信开放平台正确配置服务器地址和Token
2. 消息加解密需要使用43位的EncodingAESKey
3. 所有API调用都需要验证IP白名单
4. Token管理需要持久化存储以保证服务稳定性
5. 建议在生产环境使用数据库存储Token（多实例部署必须共享存储，避免token互相顶掉）

## 技术支持

如果在使用过程中遇到问题，请参考具体文档中的常见问题部分，或查看库源代码。
