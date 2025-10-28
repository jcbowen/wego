# WeGo 微信公众号开发库

WeGo 是一个 Go 语言编写的微信公众号开发库，提供了完整的微信公众号 API 封装。

## 功能特性

- ✅ 基础接口（获取access_token、服务器IP等）
- ✅ 自定义菜单管理
- ✅ 消息管理（群发、模板消息等）
- ✅ 客服消息
- ✅ 素材管理
- ✅ 模板消息
- ✅ 完整的错误处理
- ✅ 自动access_token管理
- ✅ 支持多种存储后端

## 快速开始

### 安装

```bash
go get github.com/jcbowen/wego/officialaccount
```

### 基本使用

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jcbowen/wego/officialaccount"
)

func main() {
	// 1. 创建配置
	config := &officialaccount.MPConfig{
		AppID:     "your_app_id",
		AppSecret: "your_app_secret",
		Token:     "your_token",
		AESKey:    "your_aes_key",
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	// 2. 创建客户端
	client := officialaccount.NewMPClient(config)

	// 3. 使用各种功能客户端
	ctx := context.Background()

	// 基础接口
	apiClient := officialaccount.NewMPAPIClient(client)
	ips, err := apiClient.GetApiDomainIp(ctx)
	if err != nil {
		log.Printf("获取API服务器IP失败: %v", err)
	} else {
		fmt.Printf("API服务器IP: %v\n", ips.IPList)
	}

	// 自定义菜单
	menuClient := officialaccount.NewMenuClient(client)
	menu := &officialaccount.Menu{
		Button: []officialaccount.Button{
			{
				Type: "click",
				Name: "今日歌曲",
				Key:  "V1001_TODAY_MUSIC",
			},
		},
	}
	resp, err := menuClient.CreateMenu(ctx, menu)
	if err != nil {
		log.Printf("创建菜单失败: %v", err)
	} else {
		fmt.Printf("创建菜单成功: %v\n", resp.IsSuccess())
	}
}
```

## 详细使用说明

### 配置

```go
config := &officialaccount.MPConfig{
	AppID:     "wx1234567890abcdef",  // 公众号AppID
	AppSecret: "your_app_secret",     // 公众号AppSecret
	Token:     "your_token",          // 消息校验Token
	AESKey:    "your_aes_key",       // 消息加解密Key（可选）
}

// 验证配置
if err := config.Validate(); err != nil {
	// 处理错误
}
```

### 客户端初始化

```go
// 创建基础客户端
client := officialaccount.NewMPClient(config)

// 如果需要使用存储（推荐）
storage := core.NewMemoryStorage() // 内存存储
// storage := core.NewRedisStorage("redis://localhost:6379") // Redis存储
client.SetStorage(storage)
```

### 各功能模块使用

#### 1. 基础接口

```go
apiClient := officialaccount.NewMPAPIClient(client)

// 获取API服务器IP
ips, err := apiClient.GetApiDomainIp(ctx)

// 获取推送服务器IP
callbackIPs, err := apiClient.GetCallbackIp(ctx)

// 网络检测
checkResult, err := apiClient.CallbackCheck(ctx, "action", "check_operator")
```

#### 2. 自定义菜单

```go
menuClient := officialaccount.NewMenuClient(client)

// 创建菜单
menu := &officialaccount.Menu{
	Button: []officialaccount.Button{
		{
			Type: "click",
			Name: "菜单1",
			Key:  "menu1",
		},
	},
}
resp, err := menuClient.CreateMenu(ctx, menu)

// 获取菜单
currentMenu, err := menuClient.GetCurrentMenu(ctx)

// 删除菜单
deleteResp, err := menuClient.DeleteMenu(ctx)
```

#### 3. 模板消息

```go
templateClient := officialaccount.NewTemplateClient(client)

// 发送模板消息
msg := &officialaccount.SendTemplateMsgRequest{
	Touser:     "user_openid",
	TemplateID: "template_id",
	Data: map[string]officialaccount.TemplateData{
		"first": {
			Value: "您好，您有新的订单",
			Color: "#173177",
		},
	},
}
resp, err := templateClient.SendTemplateMessage(ctx, msg)

// 获取行业信息
industry, err := templateClient.GetIndustry(ctx)

// 设置行业
setResp, err := templateClient.SetIndustry(ctx, "1", "2")
```

#### 4. 客服消息

```go
customClient := officialaccount.NewCustomClient(client)

// 发送文本消息
textMsg := &officialaccount.TextMessage{
	MsgType: "text",
	Text: struct {
		Content string `json:"content"`
	}{
		Content: "Hello World",
	},
}
resp, err := customClient.SendCustomMessage(ctx, "user_openid", textMsg)

// 添加客服账号
addResp, err := customClient.AddCustomAccount(ctx, "test@test", "客服昵称")
```

#### 5. 素材管理

```go
materialClient := officialaccount.NewMaterialClient(client)

// 获取素材总数
countResp, err := materialClient.GetMaterialCount(ctx)

// 获取素材列表
listResp, err := materialClient.GetMaterialList(ctx, "image", 0, 20)

// 上传临时素材
uploadResp, err := materialClient.UploadTempMedia(ctx, "image", "image.jpg", fileData)
```

#### 6. 消息管理

```go
messageClient := officialaccount.NewMessageClient(client)

// 群发消息
massMsg := &officialaccount.MassMessage{
	Filter: &officialaccount.Filter{
		IsToAll: true,
	},
	MsgType: "text",
	Text: &officialaccount.TextContent{
		Content: "群发消息内容",
	},
}
resp, err := messageClient.SendMassMessage(ctx, massMsg)
```

## 错误处理

所有API调用都会返回错误信息，建议对每个API调用进行错误处理：

```go
resp, err := apiClient.GetApiDomainIp(ctx)
if err != nil {
	// 处理网络错误或API错误
	log.Printf("API调用失败: %v", err)
	return
}

if !resp.IsSuccess() {
	// 处理微信API返回的错误
	log.Printf("微信API错误: %d - %s", resp.ErrCode, resp.ErrMsg)
	return
}

// 处理成功响应
fmt.Printf("API服务器IP: %v", resp.IPList)
```

## 存储支持

WeGo支持多种存储后端用于access_token缓存：

### 内存存储（默认）

```go
storage := core.NewMemoryStorage()
client.SetStorage(storage)
```

### 文件存储

```go
storage := core.NewFileStorage("/path/to/token.json")
client.SetStorage(storage)
```

### Redis存储

```go
storage := core.NewRedisStorage("redis://localhost:6379")
client.SetStorage(storage)
```

### 数据库存储

```go
storage := core.NewDBStorage(db, "access_tokens")
client.SetStorage(storage)
```

## 示例代码

完整的示例代码请参考 `examples/mp/` 目录。

## 许可证

本项目采用 MIT 许可证。