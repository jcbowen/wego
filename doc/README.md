# 微信开放平台第三方平台开发教程

本目录包含微信开放平台第三方平台开发和微信公众号开发的完整教程文档，基于WeGo库的实现。

## 文档列表

### 微信公众号开发
- [微信公众号开发库使用指南](./officialaccount/README.md) - WeGo微信公众号开发库完整使用说明

### 微信开放平台开发
#### 基础概念
- [01-授权流程技术说明](./openplatform/01-授权流程技术说明.md) - 第三方平台授权流程详解
- [02-验证票据(component_verify_ticket)](./openplatform/02-验证票据.md) - 验证票据的接收和处理
- [03-授权变更通知推送](./openplatform/03-授权变更通知推送.md) - 授权状态变更的事件处理
- [04-Token生成介绍](./openplatform/04-Token生成介绍.md) - 各种Token的生成和管理

#### 消息处理
- [05-消息推送介绍](./openplatform/05-消息推送介绍.md) - 消息推送服务的配置和使用
- [06-消息加解密技术介绍](./openplatform/06-消息加解密技术介绍.md) - 加解密技术实现细节

#### 接口调用
- [07-代调用接口介绍](./openplatform/07-代调用接口介绍.md) - 代公众号/小程序调用接口
- [08-代公众号网页授权](./openplatform/08-代公众号网页授权.md) - OAuth2.0网页授权实现
- [09-消息与事件处理](./openplatform/09-消息与事件处理.md) - 消息和事件处理
- [10-JS-SDK使用说明](./openplatform/10-JS-SDK使用说明.md) - 前端JS SDK集成
- [11-视频号店铺授权与开放平台账号绑定](./openplatform/11-视频号店铺授权与开放平台账号绑定.md) - 视频号小店管理和开放平台账号绑定

#### 稳定版Token功能
- [稳定版access_token使用指南](./officialaccount/stable_token.md) - 稳定版access_token的获取和管理
- [稳定版Token API接口说明](./officialaccount/stable_token_api.md) - 稳定版Token相关API接口详解

#### 订阅消息功能
- [订阅消息使用指南](./officialaccount/subscribe_message.md) - 订阅通知消息功能完整使用说明
- [订阅消息API接口说明](./officialaccount/subscribe_message_api.md) - 订阅消息相关API接口详解

## 快速开始

### 1. 安装

```bash
go get github.com/jcbowen/wego
```

### 2. 初始化客户端

```go
package main

import (
    "context"
    "fmt"
    "github.com/jcbowen/wego"
)

func main() {
    // 配置微信开放平台参数
    config := &wego.OpenPlatformConfig{
        ComponentAppID:     "your_component_app_id",
        ComponentAppSecret: "your_component_app_secret",
        ComponentToken:     "your_component_token",
        EncodingAESKey:     "your_encoding_aes_key",
        RedirectURI:        "your_redirect_uri",
    }

    // 创建WeGo实例
    wegoClient := wego.NewWeGo(config)

    // 使用各个功能模块
    openPlatformAuth := wegoClient.OpenPlatformAuth()
    openPlatformMessage := wegoClient.OpenPlatformMessage()
    cryptoClient := wegoClient.Crypto()
    storageClient := wegoClient.Storage()

    fmt.Println("WeGo客户端初始化成功！")
}
```

### 3. 配置存储

WeGo库支持多种存储方式：
- **内存存储**（默认）- 通过`storage.NewStorageClient().NewMemoryStorage()`创建
- **文件存储** - 通过`storage.NewStorageClient().NewFileStorage("path/to/storage.json")`创建
- **数据库存储** - 通过`storage.NewStorageClient().NewDBStorage(db)`创建，需要传入GORM数据库实例
- **自定义存储** - 需要实现`storage.TokenStorage`接口，并通过`wego.NewWeGoWithStorage()`使用

### 4. 处理消息和事件

WeGo提供了完整的消息处理器，支持：
- 验证票据接收
- 授权变更事件处理  
- 用户消息代收
- 消息加解密

## 功能特性

✅ **模块化设计** - 按功能模块组织代码，便于扩展和维护
✅ **完整的API封装** - 支持微信开放平台和微信公众号核心API
✅ **消息处理** - 支持微信消息的接收、解析和处理
✅ **授权管理** - 提供完整的授权流程管理
✅ **安全加密** - 支持微信消息的加密和解密
✅ **存储抽象** - 支持多种存储后端（内存、文件、数据库等）
✅ **类型安全** - 完整的类型定义和错误处理
✅ **稳定版Token** - 支持稳定版access_token获取和管理（通过`StableTokenClient`）
✅ **订阅消息** - 支持订阅通知消息功能（通过`SubscribeClient`）
✅ **数据库存储** - 支持GORM数据库存储后端
✅ **自动刷新** - 支持token自动刷新机制
✅ **多客户端支持** - 支持同时配置开放平台和公众号客户端
✅ **自定义存储** - 支持通过`NewWeGoWithStorage`使用自定义存储实现

## 项目结构

```
wego/
├── core/           # 核心配置和客户端
├── crypto/         # 加密解密功能
├── message/        # 消息处理功能
├── officialaccount/ # 公众号开发功能
├── openplatform/   # 开放平台功能
├── storage/        # 存储抽象层
├── types/          # 类型定义
└── doc/           # 技术文档
```

## 模块说明

### Core 模块
核心配置和客户端实现，包含：
- `OpenPlatformConfig` - 开放平台配置结构体
- `MPConfig` - 公众号配置结构体
- `WeGo` - 主客户端，支持多种客户端配置
- 令牌管理和HTTP客户端

### OpenPlatform 模块
微信开放平台功能，包含：
- `APIClient` - 开放平台API客户端（通过`OpenPlatformClient`字段访问）
- `AuthClient` - 授权管理客户端（通过`OpenPlatformAuth()`方法获取）
- API地址常量定义
- API响应结构体
- 授权信息数据结构
- 事件处理器接口

### OfficialAccount 模块
微信公众号开发功能，包含：
- `MPClient` - 公众号主客户端（通过`OfficialAccountClient`字段访问）
- `MPAPIClient` - 公众号API客户端（通过`OfficialAccountAPI()`方法获取）
- `MenuClient` - 菜单管理客户端（通过`OfficialAccountMenu()`方法获取）
- `MessageClient` - 消息处理客户端（通过`OfficialAccountMessage()`方法获取）
- `TemplateClient` - 模板消息客户端（通过`OfficialAccountTemplate()`方法获取）
- `CustomClient` - 客服消息客户端（通过`OfficialAccountCustom()`方法获取）
- `MaterialClient` - 素材管理客户端（通过`OfficialAccountMaterial()`方法获取）
- `StableTokenClient` - 稳定版access_token客户端（通过`MPAPIClient`的`GetStableTokenClient()`方法获取）
- `SubscribeClient` - 订阅消息客户端（通过`officialaccount.NewSubscribeClient()`创建）

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
- `TokenStorage` 接口
- `MemoryStorage` 内存存储实现
- `FileStorage` 文件存储实现

## 示例代码

每个文档都包含详细的示例代码，展示如何使用WeGo库实现相应的功能。

## 注意事项

1. 请确保在微信开放平台正确配置服务器地址和Token
2. 消息加解密需要使用43位的EncodingAESKey
3. 所有API调用都需要验证IP白名单
4. Token管理需要持久化存储以保证服务稳定性
5. 建议在生产环境中使用数据库存储Token

## 技术支持

如果在使用过程中遇到问题，请参考具体文档中的常见问题部分，或查看库源代码。