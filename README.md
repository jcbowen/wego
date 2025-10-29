# WeGo - 微信开发封装库

WeGo是一个模块化的微信开发封装库，专门为微信开放平台第三方平台开发设计。该库提供了完整的微信开放平台API封装、消息处理、授权管理等功能。

## 特性

- 🏗️ **模块化设计** - 按功能模块组织代码，便于扩展和维护
- 🔐 **完整的API封装** - 支持微信开放平台所有核心API
- 📨 **消息处理** - 支持微信消息的接收、解析和处理
- 🔑 **授权管理** - 提供完整的授权流程管理
- 🔒 **安全加密** - 支持微信消息的加密和解密
- 💾 **存储抽象** - 支持多种存储后端（内存、文件、数据库等）
- 📚 **类型安全** - 完整的类型定义和错误处理

## 项目结构

```
wego/
├── core/           # 核心配置和客户端
├── api/            # API相关功能
├── auth/           # 授权相关功能
├── message/        # 消息处理功能
├── crypto/         # 加密解密功能
├── storage/        # 存储抽象层
├── types/          # 类型定义
├── openplatform/   # 开放平台功能
├── officialaccount/ # 公众号开发功能
└── doc/           # 技术文档
```

## 快速开始

### 安装

```bash
go get github.com/jcbowen/wego
```

### 存储说明

WeGo库提供多种存储后端支持：

- **文件存储（默认）**：使用`wego_storage`目录保存Token数据
- **内存存储**：适合开发和测试环境，重启后数据丢失
- **自定义存储**：支持数据库等自定义存储实现

**默认存储策略**：
- 默认使用文件存储，数据持久化到本地文件
- 如果文件存储创建失败，会自动回退到内存存储并记录警告日志
- 可通过`NewWeGoWithStorage`方法指定自定义存储

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
		EncodingAESKey:     "your_encoding_aes_key",
		RedirectURI:        "your_redirect_uri",
	}

	// 创建WeGo实例（只初始化开放平台）
	wegoClient := wego.NewWeGo(openPlatformConfig)

	// 使用开放平台功能
	apiClient := wegoClient.OpenPlatformAPI()
	authClient := wegoClient.OpenPlatformAuth()
	messageClient := wegoClient.OpenPlatformMessage()

	fmt.Println("开放平台客户端初始化成功！")
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
	officialAccountConfig := &wego.MPConfig{
		AppID:     "your_mp_app_id",
		AppSecret: "your_mp_app_secret",
		Token:     "your_mp_token",
		AESKey:    "your_mp_aes_key",
	}

	// 创建WeGo实例（只初始化公众号）
	wegoClient := wego.NewWeGo(officialAccountConfig)

	// 使用公众号功能
	apiClient := wegoClient.OfficialAccountAPI()
	menuClient := wegoClient.OfficialAccountMenu()
	messageClient := wegoClient.OfficialAccountMessage()

	fmt.Println("公众号客户端初始化成功！")
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
		EncodingAESKey:     "your_encoding_aes_key",
		RedirectURI:        "your_redirect_uri",
	}

	// 配置公众号参数
	officialAccountConfig := &wego.MPConfig{
		AppID:     "your_mp_app_id",
		AppSecret: "your_mp_app_secret",
		Token:     "your_mp_token",
		AESKey:    "your_mp_aes_key",
	}

	// 创建WeGo实例（同时初始化开放平台和公众号）
	wegoClient := wego.NewWeGo(openPlatformConfig, officialAccountConfig)

	// 检查客户端是否初始化
	fmt.Printf("开放平台已初始化: %v\n", wegoClient.HasOpenPlatform())
	fmt.Printf("公众号已初始化: %v\n", wegoClient.HasOfficialAccount())

	// 使用开放平台功能
	openPlatformAPI := wegoClient.OpenPlatformAPI()
	
	// 使用公众号功能
	officialAccountAPI := wegoClient.OfficialAccountAPI()
	
	// 使用通用功能
	cryptoClient := wegoClient.Crypto()
	storageClient := wegoClient.Storage()

	fmt.Println("所有客户端初始化成功！")
}
```

## 模块说明

### Core 模块

核心配置和客户端实现，包含：

- `OpenPlatformConfig` - 开放平台配置结构体
- `WegoClient` - 主客户端
- 令牌管理和HTTP客户端

### API 模块

微信开放平台API封装，包含：

- API地址常量定义
- API响应结构体
- 授权信息数据结构

### Auth 模块

授权管理功能，包含：

- `AuthorizerClient` - 授权方客户端
- 代调用API接口
- 用户信息管理
- 媒体文件上传下载

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
- `FileStorage` 文件存储实现（默认存储）
- 支持自定义存储后端

**默认存储策略**：
- 默认存储为文件存储
- 文件存储使用`wego_storage`目录保存Token数据
- 如果文件存储创建失败，会自动回退到内存存储并记录警告日志

## 示例

查看 `doc/` 目录获取完整的使用示例和技术文档：

- [微信公众号开发库使用指南](doc/officialaccount/README.md)
- [授权流程技术说明](doc/01-授权流程技术说明.md)
- [消息加解密技术介绍](doc/06-消息加解密技术介绍.md)

## 文档

详细的技术文档请查看 `doc/` 目录：

- 授权流程技术说明
- 消息加解密说明
- Token生成介绍
- 消息与事件处理

## 依赖

- Go 1.23.0+
- gorm.io/gorm v1.31.0

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！
