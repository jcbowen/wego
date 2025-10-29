# 04-Token生成介绍

## 概述

微信开放平台第三方平台涉及多种Token的管理，包括ComponentAccessToken、AuthorizerAccessToken等。WeGo库提供了完整的Token生成、刷新和存储机制。

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

## WeGo库实现

### 默认存储策略

- **文件存储（默认）**：使用`wego_storage`目录保存Token数据，确保数据持久化
- **自动回退机制**：如果文件存储创建失败，会自动回退到内存存储并记录警告日志
- **生产环境推荐**：文件存储适合生产环境使用，避免重启后Token丢失

### 存储实现策略

- **文件存储（默认）**: 适合单机部署和生产环境，数据持久化到本地文件
- **内存存储**: 适合开发和测试环境，重启后数据丢失
- **数据库存储**: 适合分布式部署，支持多实例共享数据

### Token存储接口

根据实际的代码实现，TokenStorage接口定义如下：

```go
// TokenStorage Token存储接口
type TokenStorage interface {
    // ComponentAccessToken相关
    SetComponentAccessToken(token *ComponentAccessToken) error
    GetComponentAccessToken() (*ComponentAccessToken, error)
    
    // AuthorizerAccessToken相关
    SetAuthorizerAccessToken(authorizerAppID string, token *AuthorizerAccessToken) error
    GetAuthorizerAccessToken(authorizerAppID string) (*AuthorizerAccessToken, error)
    
    // PreAuthCode相关
    SetPreAuthCode(code *PreAuthCode) error
    GetPreAuthCode() (*PreAuthCode, error)
    
    // VerifyTicket相关
    SetVerifyTicket(ticket string) error
    GetVerifyTicket() (string, error)
}

// ComponentAccessToken 第三方平台Token结构体
type ComponentAccessToken struct {
    Token     string    `json:"token"`
    ExpiresIn int       `json:"expires_in"`
    CreatedAt time.Time `json:"created_at"`
}

// AuthorizerAccessToken 授权方Token结构体
type AuthorizerAccessToken struct {
    AccessToken  string    `json:"access_token"`
    ExpiresIn    int       `json:"expires_in"`
    RefreshToken string    `json:"refresh_token"`
    CreatedAt    time.Time `json:"created_at"`
}

// PreAuthCode 预授权码结构体
type PreAuthCode struct {
    PreAuthCode string    `json:"pre_auth_code"`
    ExpiresIn   int       `json:"expires_in"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### Token自动管理

WeGo库实现了Token的自动获取和刷新，通过WegoClient结构体管理：

```go
// WegoClient 核心客户端结构体
type WegoClient struct {
    config      *WeGoConfig
    httpClient  *http.Client
    logger      Logger
    storage     TokenStorage
    
    // Token缓存
    componentToken *ComponentAccessToken
    preAuthCode    *PreAuthCode
}

// GetComponentAccessToken 获取第三方平台Token（自动管理）
func (c *WegoClient) GetComponentAccessToken(ctx context.Context) (*ComponentAccessToken, error) {
    // 检查缓存中的Token是否有效
    if c.componentToken != nil && !c.isTokenExpired(c.componentToken) {
        return c.componentToken, nil
    }
    
    // 从存储中获取Token
    token, err := c.storage.GetComponentAccessToken()
    if err == nil && token != nil && !c.isTokenExpired(token) {
        c.componentToken = token
        return token, nil
    }
    
    // 获取VerifyTicket
    ticket, err := c.storage.GetVerifyTicket()
    if err != nil || ticket == "" {
        return nil, fmt.Errorf("获取VerifyTicket失败: %v", err)
    }
    
    // 调用API获取新Token
    newToken, err := c.api.GetComponentAccessToken(ctx, ticket)
    if err != nil {
        return nil, fmt.Errorf("获取ComponentAccessToken失败: %v", err)
    }
    
    // 存储新Token
    newToken.CreatedAt = time.Now()
    err = c.storage.SetComponentAccessToken(newToken)
    if err != nil {
        return nil, fmt.Errorf("存储ComponentAccessToken失败: %v", err)
    }
    
    c.componentToken = newToken
    return newToken, nil
}

// GetAuthorizerAccessToken 获取授权方Token（自动管理）
func (c *WegoClient) GetAuthorizerAccessToken(ctx context.Context, authorizerAppID string) (*AuthorizerAccessToken, error) {
    // 从存储中获取Token
    token, err := c.storage.GetAuthorizerAccessToken(authorizerAppID)
    if err == nil && token != nil && !c.isTokenExpired(token) {
        return token, nil
    }
    
    // Token过期，使用refresh_token刷新
    if token != nil && token.RefreshToken != "" {
        newToken, err := c.api.RefreshAuthorizerToken(ctx, authorizerAppID, token.RefreshToken)
        if err != nil {
            return nil, fmt.Errorf("刷新AuthorizerAccessToken失败: %v", err)
        }
        
        // 存储新Token
        newToken.CreatedAt = time.Now()
        err = c.storage.SetAuthorizerAccessToken(authorizerAppID, newToken)
        if err != nil {
            return nil, fmt.Errorf("存储AuthorizerAccessToken失败: %v", err)
        }
        
        return newToken, nil
    }
    
    return nil, fmt.Errorf("无法获取有效的AuthorizerAccessToken")
}

// isTokenExpired 检查Token是否过期
func (c *WegoClient) isTokenExpired(token interface{}) bool {
    switch t := token.(type) {
    case *ComponentAccessToken:
        return time.Since(t.CreatedAt) > time.Duration(t.ExpiresIn-300)*time.Second // 提前5分钟过期
    case *AuthorizerAccessToken:
        return time.Since(t.CreatedAt) > time.Duration(t.ExpiresIn-300)*time.Second
    case *PreAuthCode:
        return time.Since(t.CreatedAt) > time.Duration(t.ExpiresIn-60)*time.Second // 提前1分钟过期
    }
    return true
}
```

## 完整示例

### 1. Token管理示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/bowen/wego"
)

func main() {
    config := &wego.WeGoConfig{
        ComponentAppID:     "your_component_appid",
        ComponentAppSecret: "your_component_appsecret",
        ComponentToken:     "your_component_token",
        EncodingAESKey:     "your_encoding_aes_key",
    }
    
    // 创建WeGo客户端
    client := wego.NewWeGo(config)
    
    // 创建带超时的上下文
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // 获取ComponentAccessToken（自动管理）
    componentToken, err := client.GetComponentAccessToken(ctx)
    if err != nil {
        log.Printf("获取ComponentAccessToken失败: %v\n", err)
        return
    }
    fmt.Printf("ComponentAccessToken: %s, 有效期: %d秒\n", 
        componentToken.Token, componentToken.ExpiresIn)
    
    // 使用授权方AppID获取AuthorizerAccessToken
    authorizerAppID := "授权方AppID"
    authorizerToken, err := client.GetAuthorizerAccessToken(ctx, authorizerAppID)
    if err != nil {
        log.Printf("获取AuthorizerAccessToken失败: %v\n", err)
        return
    }
    fmt.Printf("AuthorizerAccessToken: %s, 有效期: %d秒\n", 
        authorizerToken.AccessToken, authorizerToken.ExpiresIn)
    
    // 使用Token调用API
    err = callAPIWithToken(ctx, client, authorizerAppID)
    if err != nil {
        log.Printf("调用API失败: %v\n", err)
        return
    }
}

// 使用Token调用API的示例
func callAPIWithToken(ctx context.Context, client *wego.WeGo, authorizerAppID string) error {
    // 获取授权方客户端
    authorizerClient := client.Auth(authorizerAppID)
    
    // 调用获取用户列表API
    // 注意：实际实现中可能需要根据具体的API接口进行调整
    // 这里仅作为示例展示调用模式
    
    // 示例：发送客服消息
    message := &wego.TextMessage{
        ToUser:  "用户OpenID",
        Content: "这是一条测试消息",
    }
    
    err := authorizerClient.SendCustomMessage(ctx, message)
    if err != nil {
        return fmt.Errorf("发送客服消息失败: %v", err)
    }
    
    fmt.Println("客服消息发送成功")
    return nil
}
```

### 2. 自定义存储实现

```go
import (
    "database/sql"
    "time"
    
    "github.com/bowen/wego"
)

// DatabaseStorage 数据库存储实现
type DatabaseStorage struct {
    db *sql.DB
}

func (s *DatabaseStorage) SetComponentAccessToken(token *wego.ComponentAccessToken) error {
    _, err := s.db.Exec(`
        INSERT INTO component_tokens (token, expires_in, created_at) 
        VALUES (?, ?, ?) 
        ON DUPLICATE KEY UPDATE token = ?, expires_in = ?, created_at = ?
    `, token.Token, token.ExpiresIn, token.CreatedAt, 
       token.Token, token.ExpiresIn, token.CreatedAt)
    
    return err
}

func (s *DatabaseStorage) GetComponentAccessToken() (*wego.ComponentAccessToken, error) {
    var token string
    var expiresIn int
    var createdAt time.Time
    
    err := s.db.QueryRow(`
        SELECT token, expires_in, created_at FROM component_tokens 
        WHERE created_at > ? 
        ORDER BY created_at DESC LIMIT 1
    `, time.Now().Add(-2*time.Hour)).Scan(&token, &expiresIn, &createdAt)
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    
    if err != nil {
        return nil, err
    }
    
    return &wego.ComponentAccessToken{
        Token:     token,
        ExpiresIn: expiresIn,
        CreatedAt: createdAt,
    }, nil
}

func (s *DatabaseStorage) SetAuthorizerAccessToken(authorizerAppID string, token *wego.AuthorizerAccessToken) error {
    _, err := s.db.Exec(`
        INSERT INTO authorizer_tokens (app_id, access_token, refresh_token, expires_in, created_at) 
        VALUES (?, ?, ?, ?, ?) 
        ON DUPLICATE KEY UPDATE access_token = ?, refresh_token = ?, expires_in = ?, created_at = ?
    `, authorizerAppID, token.AccessToken, token.RefreshToken, token.ExpiresIn, token.CreatedAt,
       token.AccessToken, token.RefreshToken, token.ExpiresIn, token.CreatedAt)
    
    return err
}

func (s *DatabaseStorage) GetAuthorizerAccessToken(authorizerAppID string) (*wego.AuthorizerAccessToken, error) {
    var accessToken, refreshToken string
    var expiresIn int
    var createdAt time.Time
    
    err := s.db.QueryRow(`
        SELECT access_token, refresh_token, expires_in, created_at 
        FROM authorizer_tokens 
        WHERE app_id = ? AND created_at > ?
    `, authorizerAppID, time.Now().Add(-2*time.Hour)).Scan(&accessToken, &refreshToken, &expiresIn, &createdAt)
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    
    if err != nil {
        return nil, err
    }
    
    return &wego.AuthorizerAccessToken{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    expiresIn,
        CreatedAt:    createdAt,
    }, nil
}

// 实现其他接口方法...
func (s *DatabaseStorage) SetPreAuthCode(code *wego.PreAuthCode) error {
    // 实现预授权码存储
    return nil
}

func (s *DatabaseStorage) GetPreAuthCode() (*wego.PreAuthCode, error) {
    // 实现预授权码获取
    return nil, nil
}

func (s *DatabaseStorage) SetVerifyTicket(ticket string) error {
    // 实现VerifyTicket存储
    return nil
}

func (s *DatabaseStorage) GetVerifyTicket() (string, error) {
    // 实现VerifyTicket获取
    return "", nil
}

// 使用自定义存储
func main() {
    config := &wego.WeGoConfig{
        ComponentAppID:     "your_component_appid",
        ComponentAppSecret: "your_component_appsecret",
        ComponentToken:     "your_component_token",
        EncodingAESKey:     "your_encoding_aes_key",
    }
    
    db, _ := sql.Open("mysql", "user:pass@/dbname")
    storage := &DatabaseStorage{db: db}
    
    // 使用自定义存储创建客户端
    client := wego.NewWeGoWithStorage(config, storage)
    
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

### 4. 上下文管理
- 所有Token获取方法都需要传入context.Context参数
- 支持超时控制和取消操作
- 建议为每个API调用设置合理的超时时间

## 常见问题

### Q: Token获取失败
A: 检查网络连接、AppID/AppSecret是否正确、VerifyTicket是否存在、context是否超时

### Q: Token刷新失败
A: 检查RefreshToken是否有效、授权是否已被取消、context是否超时

### Q: Token存储异常
A: 检查存储实现、数据库连接、权限设置、Token结构体字段是否正确

### Q: 如何实现自定义存储
A: 实现TokenStorage接口的所有方法，然后使用`wego.NewWeGoWithStorage(config, storage)`创建客户端

## 相关API

- `GetComponentAccessToken(ctx context.Context)` - 获取第三方平台Token
- `GetAuthorizerAccessToken(ctx context.Context, authorizerAppID string)` - 获取授权方Token
- `RefreshAuthorizerToken(ctx context.Context, authorizerAppID, refreshToken string)` - 刷新授权方Token
- `NewWeGoWithStorage(config *WeGoConfig, storage TokenStorage)` - 使用自定义存储创建客户端
- `SetVerifyTicket(ticket string)` - 设置VerifyTicket（用于获取ComponentAccessToken）

通过WeGo库，您可以轻松管理各种Token，确保第三方平台服务的稳定运行。