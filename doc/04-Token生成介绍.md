# 04-Token生成介绍

## 概述

微信开放平台第三方平台涉及多种Token的管理，包括ComponentAccessToken、AuthorizerAccessToken等。wxopen组件提供了完整的Token生成、刷新和存储机制。

## Token类型

### 1. ComponentAccessToken（第三方平台Token）
- 用途：调用第三方平台API
- 有效期：2小时
- 获取方式：使用component_verify_ticket换取

### 2. AuthorizerAccessToken（授权方Token）
- 用途：代授权方调用API
- 有效期：2小时
- 获取方式：使用授权码或refresh_token换取

### 3. RefreshToken（刷新Token）
- 用途：刷新AuthorizerAccessToken
- 有效期：30天
- 获取方式：授权时获取

## wxopen组件实现

### Token管理接口

```go
// TokenStorage Token存储接口
type TokenStorage interface {
    // ComponentAccessToken相关
    SetComponentAccessToken(token string, expiresIn int) error
    GetComponentAccessToken() (string, error)
    
    // AuthorizerAccessToken相关
    SetAuthorizerAccessToken(appID, token string, expiresIn int) error
    GetAuthorizerAccessToken(appID string) (string, error)
    
    // RefreshToken相关
    SetAuthorizerRefreshToken(appID, refreshToken string) error
    GetAuthorizerRefreshToken(appID string) (string, error)
    
    // VerifyTicket相关
    SetVerifyTicket(ticket string) error
    GetVerifyTicket() (string, error)
}
```

### Token自动管理

wxopen组件实现了Token的自动获取和刷新：

```go
// 获取ComponentAccessToken（自动管理）
func (client *WxOpenClient) GetComponentAccessToken() (string, int, error) {
    // 先从存储获取
    token, err := client.storage.GetComponentAccessToken()
    if err == nil && token != "" {
        return token, 7200, nil // 假设还有2小时有效期
    }
    
    // 获取失败或过期，重新获取
    ticket, err := client.GetVerifyTicket()
    if err != nil {
        return "", 0, fmt.Errorf("获取VerifyTicket失败: %v", err)
    }
    
    // 调用API获取新Token
    newToken, expiresIn, err := client.api.GetComponentAccessToken(ticket)
    if err != nil {
        return "", 0, fmt.Errorf("获取ComponentAccessToken失败: %v", err)
    }
    
    // 存储新Token
    err = client.storage.SetComponentAccessToken(newToken, expiresIn)
    if err != nil {
        return "", 0, fmt.Errorf("存储ComponentAccessToken失败: %v", err)
    }
    
    return newToken, expiresIn, nil
}

// 获取AuthorizerAccessToken（自动管理）
func (client *WxOpenClient) GetAuthorizerAccessToken(appID string) (string, int, error) {
    // 先从存储获取
    token, err := client.storage.GetAuthorizerAccessToken(appID)
    if err == nil && token != "" {
        return token, 7200, nil
    }
    
    // 获取失败或过期，使用refresh_token刷新
    refreshToken, err := client.storage.GetAuthorizerRefreshToken(appID)
    if err != nil || refreshToken == "" {
        return "", 0, fmt.Errorf("获取RefreshToken失败: %v", err)
    }
    
    // 调用API刷新Token
    newTokenInfo, err := client.api.RefreshAuthorizerToken(appID, refreshToken)
    if err != nil {
        return "", 0, fmt.Errorf("刷新AuthorizerAccessToken失败: %v", err)
    }
    
    // 存储新Token
    err = client.storage.SetAuthorizerAccessToken(appID, newTokenInfo.AccessToken, newTokenInfo.ExpiresIn)
    if err != nil {
        return "", 0, fmt.Errorf("存储AuthorizerAccessToken失败: %v", err)
    }
    
    // 更新refresh_token（如果返回了新的）
    if newTokenInfo.RefreshToken != "" {
        err = client.storage.SetAuthorizerRefreshToken(appID, newTokenInfo.RefreshToken)
        if err != nil {
            return "", 0, fmt.Errorf("存储RefreshToken失败: %v", err)
        }
    }
    
    return newTokenInfo.AccessToken, newTokenInfo.ExpiresIn, nil
}
```

## 完整示例

### 1. Token管理示例

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/wego"
)

func main() {
    config := &wego.WxOpenConfig{
        ComponentAppID:     "your_component_appid",
        ComponentAppSecret: "your_component_appsecret",
        ComponentToken:     "your_component_token",
        EncodingAESKey:     "your_encoding_aes_key",
    }
    
    client := wego.NewWxOpenClient(config)
    
    // 获取ComponentAccessToken（自动管理）
    componentToken, expiresIn, err := client.GetComponentAccessToken()
    if err != nil {
        fmt.Printf("获取ComponentAccessToken失败: %v\n", err)
        return
    }
    fmt.Printf("ComponentAccessToken: %s, 有效期: %d秒\n", componentToken, expiresIn)
    
    // 使用授权方AppID获取AuthorizerAccessToken
    authorizerAppID := "授权方AppID"
    authorizerToken, expiresIn, err := client.GetAuthorizerAccessToken(authorizerAppID)
    if err != nil {
        fmt.Printf("获取AuthorizerAccessToken失败: %v\n", err)
        return
    }
    fmt.Printf("AuthorizerAccessToken: %s, 有效期: %d秒\n", authorizerToken, expiresIn)
    
    // 使用Token调用API
    err = callAPIWithToken(client, authorizerAppID)
    if err != nil {
        fmt.Printf("调用API失败: %v\n", err)
        return
    }
}

// 使用Token调用API的示例
func callAPIWithToken(client *wego.WxOpenClient, appID string) error {
    // 获取授权方客户端
    authorizerClient := client.GetAuthorizerClient(appID)
    
    // 调用获取用户列表API
    users, total, count, nextOpenID, err := authorizerClient.GetUserList("")
    if err != nil {
        return fmt.Errorf("获取用户列表失败: %v", err)
    }
    
    fmt.Printf("用户总数: %d, 本次获取: %d, 下一个OpenID: %s\n", total, count, nextOpenID)
    for _, user := range users {
        fmt.Printf("用户: %s\n", user)
    }
    
    return nil
}
```

### 2. 自定义存储实现

```go
// 数据库存储实现
type DatabaseStorage struct {
    db *sql.DB
}

func (s *DatabaseStorage) SetComponentAccessToken(token string, expiresIn int) error {
    expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
    
    _, err := s.db.Exec(`
        INSERT INTO component_tokens (token, expires_at, created_at) 
        VALUES (?, ?, ?) 
        ON DUPLICATE KEY UPDATE token = ?, expires_at = ?, updated_at = ?
    `, token, expiresAt, time.Now(), token, expiresAt, time.Now())
    
    return err
}

func (s *DatabaseStorage) GetComponentAccessToken() (string, error) {
    var token string
    var expiresAt time.Time
    
    err := s.db.QueryRow(`
        SELECT token FROM component_tokens 
        WHERE expires_at > ? 
        ORDER BY created_at DESC LIMIT 1
    `, time.Now()).Scan(&token)
    
    if err == sql.ErrNoRows {
        return "", nil
    }
    
    return token, err
}

// 类似的实现其他Token存储方法...

// 使用自定义存储
func main() {
    config := &wego.WxOpenConfig{
        // ... 配置
    }
    
    db, _ := sql.Open("mysql", "user:pass@/dbname")
    storage := &DatabaseStorage{db: db}
    
    client := wego.NewWxOpenClientWithStorage(config, storage)
    // ... 使用客户端
}
```

## Token管理最佳实践

### 1. 缓存策略
- 实现Token缓存，减少API调用
- 设置合理的缓存过期时间
- 处理缓存失效情况

### 2. 刷新机制
- 提前刷新Token（如提前5分钟）
- 处理刷新失败的重试机制
- 监控Token使用情况

### 3. 存储安全
- 加密存储敏感Token
- 定期清理过期Token
- 备份重要Token数据

### 4. 监控告警
- 监控Token获取失败率
- 设置Token过期告警
- 记录Token使用日志

## 注意事项

### 1. Token有效期
- ComponentAccessToken：2小时
- AuthorizerAccessToken：2小时  
- RefreshToken：30天
- 预授权码：20分钟

### 2. 调用频率限制
- 每个Token有调用频率限制
- 避免频繁获取Token
- 实现Token复用机制

### 3. 错误处理
- 处理Token过期错误
- 实现Token自动刷新
- 记录错误日志

## 常见问题

### Q: Token获取失败
A: 检查网络连接、AppID/AppSecret是否正确、VerifyTicket是否存在

### Q: Token刷新失败
A: 检查RefreshToken是否有效、授权是否已被取消

### Q: Token存储异常
A: 检查存储实现、数据库连接、权限设置

## 相关API

- `GetComponentAccessToken()` - 获取第三方平台Token
- `GetAuthorizerAccessToken()` - 获取授权方Token
- `RefreshAuthorizerToken()` - 刷新授权方Token
- `SetTokenStorage()` - 设置自定义Token存储

通过wxopen组件，您可以轻松管理各种Token，确保第三方平台服务的稳定运行。