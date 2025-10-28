# 微信开放平台组件 (wego)

这是一个用于对接微信开放平台的Go语言组件，支持第三方平台开发，包括公众号/小程序授权管理、API代调用等功能。

## 功能特性

- ✅ 组件令牌管理 (Component Access Token)
- ✅ 预授权码获取和授权链接生成
- ✅ 授权信息换取和令牌刷新
- ✅ 授权方API代调用 (公众号/小程序)
- ✅ 消息处理 (事件推送、消息加解密)
- ✅ 客服消息发送
- ✅ 菜单管理
- ✅ 用户信息获取
- ✅ 模板消息发送
- ✅ 批量用户信息获取
- ✅ 小程序码生成
- ✅ 订阅消息发送

## 安装

```go
go get github.com/jcbowen/jcbaseGo/component/wego
```

## 快速开始

### 1. 初始化客户端

```go
import "github.com/jcbowen/jcbaseGo/component/wego"

config := &wego.WxOpenConfig{
    ComponentAppID:     "your_component_appid",
    ComponentAppSecret: "your_component_appsecret",
    ComponentToken:     "your_component_token",
    EncodingAESKey:     "your_encoding_aes_key",
    RedirectURI:        "your_redirect_uri",
}

client := wego.NewWxOpenClient(config)
```

### 2. 处理授权流程

```go
// 获取预授权码
preAuthCode, err := client.GetPreAuthCode(context.Background())
if err != nil {
    log.Fatal(err)
}

// 生成授权链接
authURL := client.GenerateAuthURL(preAuthCode.PreAuthCode, 0)
fmt.Println("授权链接:", authURL)

// 使用授权码换取授权信息
authInfo, err := client.QueryAuth(context.Background(), "authorization_code")
if err != nil {
    log.Fatal(err)
}

// 缓存授权方token
client.SetAuthorizerToken(
    authInfo.AuthorizationInfo.AuthorizerAppID,
    authInfo.AuthorizationInfo.AuthorizerAccessToken,
    authInfo.AuthorizationInfo.AuthorizerRefreshToken,
    authInfo.AuthorizationInfo.ExpiresIn,
)
```

### 3. 代公众号调用API

```go
// 创建授权方客户端
authorizerClient := wego.NewAuthorizerClient(client, "authorizer_appid")

// 发送客服消息
textMsg := &wego.TextMessage{
    Content: "Hello, World!",
}
err := authorizerClient.SendCustomMessage(context.Background(), "user_openid", textMsg)
if err != nil {
    log.Fatal(err)
}

// 获取用户信息
userInfo, err := authorizerClient.GetUserInfo(context.Background(), "user_openid")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("用户昵称: %s\n", userInfo.Nickname)
```

### 4. 处理微信服务器消息

```go
// 创建消息处理器
processor := wego.NewMessageProcessor(config)

// 注册事件处理器
processor.RegisterEventHandler(wego.EventTypeComponentVerifyTicket, &wego.DefaultEventHandler{})
processor.RegisterEventHandler(wego.EventTypeAuthorized, &wego.DefaultEventHandler{})

// 验证签名
if processor.VerifySignature(signature, timestamp, nonce) {
    // 处理消息
    result, err := processor.ProcessMessage(xmlData)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("处理结果:", result)
}
```

## API 参考

### 核心类

#### WxOpenClient

微信开放平台客户端主类，负责管理access_token和提供基础API。

**主要方法:**
- `GetComponentAccessToken()` - 获取第三方平台access_token
- `GetPreAuthCode()` - 获取预授权码
- `QueryAuth()` - 使用授权码换取授权信息
- `RefreshAuthorizerToken()` - 刷新授权方access_token
- `GetAuthorizerInfo()` - 获取授权方信息
- `GenerateAuthURL()` - 生成授权链接

#### AuthorizerClient

授权方API客户端，用于代公众号/小程序调用API。

**主要方法:**
- `SendCustomMessage()` - 发送客服消息
- `CreateMenu()` / `GetMenu()` / `DeleteMenu()` - 菜单管理
- `GetUserInfo()` / `GetUserList()` - 用户管理
- `SendTemplateMessage()` - 发送模板消息
- `UploadMedia()` / `GetMedia()` - 素材管理
- `CreateQRCode()` - 创建二维码

#### MessageProcessor

消息处理器，负责处理微信服务器推送的消息和事件。

**主要方法:**
- `ProcessMessage()` - 处理消息
- `VerifySignature()` - 验证签名
- `EncryptMessage()` / `DecryptMessage()` - 消息加密解密
- `GenerateTextResponse()` - 生成文本回复

### 配置说明

#### WxOpenConfig

```go
type WxOpenConfig struct {
    ComponentAppID     string // 第三方平台appid
    ComponentAppSecret string // 第三方平台appsecret
    ComponentToken     string // 消息校验Token
    EncodingAESKey     string // 消息加解密Key
    RedirectURI        string // 授权回调URI
}
```

## 事件处理

### 组件验证票据事件

```go
type ComponentVerifyTicketHandler struct{}

func (h *ComponentVerifyTicketHandler) HandleEvent(event *wego.EventMessage) (interface{}, error) {
    // 解析验证票据事件
    ticketEvent, err := wego.ParseComponentVerifyTicket(eventXML)
    if err != nil {
        return nil, err
    }
    
    // 保存component_verify_ticket
    saveTicket(ticketEvent.ComponentVerifyTicket)
    
    return "success", nil
}
```

### 授权事件

```go
type AuthorizedEventHandler struct{}

func (h *AuthorizedEventHandler) HandleEvent(event *wego.EventMessage) (interface{}, error) {
    // 解析授权成功事件
    authEvent, err := wego.ParseAuthorizedEvent(eventXML)
    if err != nil {
        return nil, err
    }
    
    // 处理授权成功逻辑
    handleAuthorization(authEvent.AuthorizerAppID, authEvent.AuthorizationCode)
    
    return "success", nil
}
```

## 错误处理

所有API方法都返回error，可以通过类型断言判断错误类型：

```go
result, err := client.GetPreAuthCode(ctx)
if err != nil {
    if apiErr, ok := err.(*wego.APIResponse); ok {
        fmt.Printf("微信API错误: %d - %s\n", apiErr.ErrCode, apiErr.ErrMsg)
    } else {
        fmt.Printf("其他错误: %v\n", err)
    }
}
```

## 缓存管理

组件内置了access_token的缓存机制，会自动处理token的刷新。如果需要手动清除缓存：

```go
client.ClearCache()
```

## 日志配置

可以设置自定义日志器：

```go
type CustomLogger struct{}

func (l *CustomLogger) Debugf(format string, args ...interface{}) {
    // 实现debug日志
}

func (l *CustomLogger) Infof(format string, args ...interface{}) {
    // 实现info日志
}

// 设置自定义日志器
client.SetLogger(&CustomLogger{})
```

## 示例代码

更多完整示例请参考 `example/` 目录。

## 注意事项

1. **ComponentVerifyTicket** 需要妥善保存，它是获取component_access_token的必要参数
2. **授权方refresh_token** 需要安全存储，用于刷新授权方access_token
3. **消息加解密** 需要配置正确的EncodingAESKey
4. **API调用频率** 需要遵守微信开放平台的频率限制

## 版本历史

- v1.0.0: 初始版本，支持基础API和消息处理

## 许可证

MIT License