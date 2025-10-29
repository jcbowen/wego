# 04-Token生成介绍

## 概述

微信开放平台第三方平台涉及多种Token的管理，包括ComponentAccessToken、AuthorizerAccessToken等。WeGo库提供了完整的Token生成、刷新和存储机制。

## Token类型

### 1. ComponentAccessToken（第三方平台Token）
- 用途：调用第三方平台API
- 有效期：2小时
- 获取方式：使用component_verify_ticket换取

### 4. ComponentVerifyTicket（验证票据）
- 用途：获取ComponentAccessToken的凭证
- 有效期：12小时
- 推送频率：微信服务器每10分钟推送一次

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
    SaveComponentToken(ctx context.Context, token *ComponentAccessToken) error
    GetComponentToken(ctx context.Context) (*ComponentAccessToken, error)
    
    // AuthorizerAccessToken相关
    SaveAuthorizerToken(ctx context.Context, authorizerAppID string, token *AuthorizerAccessToken) error
    GetAuthorizerToken(ctx context.Context, authorizerAppID string) (*AuthorizerAccessToken, error)
    
    // PreAuthCode相关
    SavePreAuthCode(ctx context.Context, code *PreAuthCode) error
    GetPreAuthCode(ctx context.Context) (*PreAuthCode, error)
    
    // VerifyTicket相关
    SaveVerifyTicket(ctx context.Context, ticket string) error
    GetVerifyTicket(ctx context.Context) (string, error)
}

// ComponentAccessToken 第三方平台Token结构体
type ComponentAccessToken struct {
    AccessToken string    `json:"component_access_token"`
    ExpiresIn   int       `json:"expires_in"`
    ExpiresAt   time.Time `json:"expires_at"`
}

// AuthorizerAccessToken 授权方Token结构体
type AuthorizerAccessToken struct {
    AuthorizerAppID        string    `json:"authorizer_appid"`
    AuthorizerAccessToken  string    `json:"authorizer_access_token"`
    ExpiresIn              int       `json:"expires_in"`
    ExpiresAt              time.Time `json:"expires_at"`
    AuthorizerRefreshToken string    `json:"authorizer_refresh_token"`
}

// PreAuthCode 预授权码结构体
type PreAuthCode struct {
    PreAuthCode string    `json:"pre_auth_code"`
    ExpiresIn   int       `json:"expires_in"`
    ExpiresAt   time.Time `json:"expires_at"`
}
```

### Token自动管理

WeGo库实现了Token的自动获取和刷新，通过OpenPlatformClient结构体管理：

```go
// OpenPlatformClient 微信开放平台客户端结构体
type OpenPlatformClient struct {
    config      *OpenPlatformConfig
    httpClient  *http.Client
    logger      Logger
    storage     TokenStorage
}

// GetComponentToken 获取第三方平台Token（自动管理）
func (c *OpenPlatformClient) GetComponentToken(ctx context.Context) (*ComponentAccessToken, error) {
    // 从存储中获取Token
    token, err := c.storage.GetComponentToken(ctx)
    if err == nil && token != nil && !c.isTokenExpired(token) {
        return token, nil
    }
    
    // 获取VerifyTicket
    ticket, err := c.storage.GetVerifyTicket(ctx)
    if err != nil || ticket == "" {
        return nil, fmt.Errorf("获取VerifyTicket失败: %v", err)
    }
    
    // 调用API获取新Token
    apiClient := api.NewAPIClient(c)
    newToken, err := apiClient.GetComponentAccessToken(ctx, ticket)
    if err != nil {
        return nil, fmt.Errorf("获取ComponentAccessToken失败: %v", err)
    }
    
    // 存储新Token
    err = c.storage.SaveComponentToken(ctx, newToken)
    if err != nil {
        return nil, fmt.Errorf("存储ComponentAccessToken失败: %v", err)
    }
    
    return newToken, nil
}

// GetAuthorizerAccessToken 获取授权方Token（自动管理）
func (c *OpenPlatformClient) GetAuthorizerAccessToken(ctx context.Context, authorizerAppID string) (string, error) {
    // 从存储中获取Token
    token, err := c.storage.GetAuthorizerToken(ctx, authorizerAppID)
    if err == nil && token != nil && !c.isTokenExpired(token) {
        return token.AuthorizerAccessToken, nil
    }
    
    // Token过期，使用refresh_token刷新
    if token != nil && token.RefreshToken != "" {
        apiClient := api.NewAPIClient(c)
        authInfo, err := apiClient.RefreshAuthorizerToken(ctx, authorizerAppID, token.RefreshToken)
        if err != nil {
            return "", fmt.Errorf("刷新AuthorizerAccessToken失败: %v", err)
        }
        
        // 存储新Token
        newToken := &AuthorizerAccessToken{
            AuthorizerAppID:        authorizerAppID,
            AuthorizerAccessToken:  authInfo.AuthorizerAccessToken,
            ExpiresIn:              authInfo.ExpiresIn,
            ExpiresAt:              time.Now().Add(time.Duration(authInfo.ExpiresIn) * time.Second),
            RefreshToken:           authInfo.RefreshToken,
        }
        err = c.storage.SaveAuthorizerToken(ctx, authorizerAppID, newToken)
        if err != nil {
            return "", fmt.Errorf("存储AuthorizerAccessToken失败: %v", err)
        }
        
        return newToken.AuthorizerAccessToken, nil
    }
    
    return "", fmt.Errorf("无法获取有效的AuthorizerAccessToken")
}

// isTokenExpired 检查Token是否过期
func (c *OpenPlatformClient) isTokenExpired(token interface{}) bool {
    switch t := token.(type) {
    case *ComponentAccessToken:
        return time.Now().After(t.ExpiresAt)
    case *AuthorizerAccessToken:
        return time.Now().After(t.ExpiresAt)
    case *PreAuthCode:
        return time.Now().After(t.ExpiresAt)
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
    
    "github.com/jcbowen/wego"
)

func main() {
    config := &wego.OpenPlatformConfig{
        ComponentAppID:     "your_component_appid",
        ComponentAppSecret: "your_component_appsecret",
        ComponentToken:     "your_component_token",
        EncodingAESKey:     "your_encoding_aes_key",
    }
    
    // 创建WeGo客户端
    client := wego.NewOpenPlatform(config)
    
    // 创建带超时的上下文
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // 获取ComponentAccessToken（自动管理）
    apiClient := client.API()
    componentToken, err := apiClient.GetComponentAccessToken(ctx, "verify_ticket_here")
    if err != nil {
        log.Printf("获取ComponentAccessToken失败: %v\n", err)
        return
    }
    fmt.Printf("ComponentAccessToken: %s, 有效期: %d秒\n", 
        componentToken.AccessToken, componentToken.ExpiresIn)
    
    // 使用授权方AppID获取AuthorizerAccessToken
    authorizerAppID := "授权方AppID"
    authClient := client.Auth()
    authorizerToken, err := authClient.GetAuthorizerAccessToken(ctx, authorizerAppID)
    if err != nil {
        log.Printf("获取AuthorizerAccessToken失败: %v\n", err)
        return
    }
    fmt.Printf("AuthorizerAccessToken: %s\n", authorizerToken)
    
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
    
    "github.com/jcbowen/wego"
)

// DatabaseStorage 数据库存储实现
type DatabaseStorage struct {
    db *sql.DB
}

func (s *DatabaseStorage) SaveComponentToken(ctx context.Context, token *wego.ComponentAccessToken) error {
    _, err := s.db.ExecContext(ctx, `
        INSERT INTO component_tokens (access_token, expires_in, expires_at) 
        VALUES (?, ?, ?) 
        ON DUPLICATE KEY UPDATE access_token = ?, expires_in = ?, expires_at = ?
    `, token.AccessToken, token.ExpiresIn, token.ExpiresAt, 
       token.AccessToken, token.ExpiresIn, token.ExpiresAt)
    
    return err
}

func (s *DatabaseStorage) GetComponentToken(ctx context.Context) (*wego.ComponentAccessToken, error) {
    var accessToken string
    var expiresIn int
    var expiresAt time.Time
    
    err := s.db.QueryRowContext(ctx, `
        SELECT access_token, expires_in, expires_at FROM component_tokens 
        WHERE expires_at > ? 
        ORDER BY expires_at DESC LIMIT 1
    `, time.Now()).Scan(&accessToken, &expiresIn, &expiresAt)
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    
    if err != nil {
        return nil, err
    }
    
    return &wego.ComponentAccessToken{
        AccessToken: accessToken,
        ExpiresIn:   expiresIn,
        ExpiresAt:   expiresAt,
    }, nil
}

func (s *DatabaseStorage) SaveAuthorizerToken(ctx context.Context, authorizerAppID string, token *wego.AuthorizerAccessToken) error {
    _, err := s.db.ExecContext(ctx, `
        INSERT INTO authorizer_tokens (app_id, authorizer_access_token, refresh_token, expires_in, expires_at) 
        VALUES (?, ?, ?, ?, ?) 
        ON DUPLICATE KEY UPDATE authorizer_access_token = ?, refresh_token = ?, expires_in = ?, expires_at = ?
    `, authorizerAppID, token.AuthorizerAccessToken, token.RefreshToken, token.ExpiresIn, token.ExpiresAt,
       token.AuthorizerAccessToken, token.RefreshToken, token.ExpiresIn, token.ExpiresAt)
    
    return err
}

func (s *DatabaseStorage) GetAuthorizerToken(ctx context.Context, authorizerAppID string) (*wego.AuthorizerAccessToken, error) {
    var authorizerAccessToken, refreshToken string
    var expiresIn int
    var expiresAt time.Time
    
    err := s.db.QueryRowContext(ctx, `
        SELECT authorizer_access_token, refresh_token, expires_in, expires_at 
        FROM authorizer_tokens 
        WHERE app_id = ? AND expires_at > ?
    `, authorizerAppID, time.Now()).Scan(&authorizerAccessToken, &refreshToken, &expiresIn, &expiresAt)
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    
    if err != nil {
        return nil, err
    }
    
    return &wego.AuthorizerAccessToken{
        AuthorizerAppID:       authorizerAppID,
        AuthorizerAccessToken: authorizerAccessToken,
        RefreshToken:          refreshToken,
        ExpiresIn:            expiresIn,
        ExpiresAt:            expiresAt,
    }, nil
}

// 实现其他接口方法...
func (s *DatabaseStorage) SavePreAuthCode(ctx context.Context, code *wego.PreAuthCode) error {
    _, err := s.db.ExecContext(ctx, `
        INSERT INTO pre_auth_codes (pre_auth_code, expires_in, expires_at) 
        VALUES (?, ?, ?) 
        ON DUPLICATE KEY UPDATE pre_auth_code = ?, expires_in = ?, expires_at = ?
    `, code.PreAuthCode, code.ExpiresIn, code.ExpiresAt,
       code.PreAuthCode, code.ExpiresIn, code.ExpiresAt)
    return err
}

func (s *DatabaseStorage) GetPreAuthCode(ctx context.Context) (*wego.PreAuthCode, error) {
    var preAuthCode string
    var expiresIn int
    var expiresAt time.Time
    
    err := s.db.QueryRowContext(ctx, `
        SELECT pre_auth_code, expires_in, expires_at FROM pre_auth_codes 
        WHERE expires_at > ? 
        ORDER BY expires_at DESC LIMIT 1
    `, time.Now()).Scan(&preAuthCode, &expiresIn, &expiresAt)
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    
    if err != nil {
        return nil, err
    }
    
    return &wego.PreAuthCode{
        PreAuthCode: preAuthCode,
        ExpiresIn:   expiresIn,
        ExpiresAt:   expiresAt,
    }, nil
}

func (s *DatabaseStorage) SaveVerifyTicket(ctx context.Context, ticket string) error {
    _, err := s.db.ExecContext(ctx, `
        INSERT INTO verify_tickets (ticket, created_at) 
        VALUES (?, ?) 
        ON DUPLICATE KEY UPDATE ticket = ?, created_at = ?
    `, ticket, time.Now(), ticket, time.Now())
    return err
}

func (s *DatabaseStorage) GetVerifyTicket(ctx context.Context) (string, error) {
    var ticket string
    var createdAt time.Time
    
    err := s.db.QueryRowContext(ctx, `
        SELECT ticket, created_at FROM verify_tickets 
        ORDER BY created_at DESC LIMIT 1
    `).Scan(&ticket, &createdAt)
    
    if err == sql.ErrNoRows {
        return "", nil
    }
    
    if err != nil {
        return "", err
    }
    
    return ticket, nil
}

// 使用自定义存储
func main() {
    config := &wego.OpenPlatformConfig{
        ComponentAppID:     "your_component_appid",
        ComponentAppSecret: "your_component_appsecret",
        ComponentToken:     "your_component_token",
        EncodingAESKey:     "your_encoding_aes_key",
    }
    
    db, _ := sql.Open("mysql", "user:pass@/dbname")
    storage := &DatabaseStorage{db: db}
    
    // 使用自定义存储创建客户端
    client := wego.NewWeGoWithStorage(storage, config)
    
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

### 1. 获取第三方平台Token
- **方法**: `GetComponentAccessToken(ctx context.Context, verifyTicket string)`
- **请求参数**: 
  - `component_appid` - 第三方平台AppID
  - `component_appsecret` - 第三方平台AppSecret  
  - `component_verify_ticket` - 微信服务器推送的VerifyTicket
- **返回值**: `component_access_token` (有效期7200秒)
- **API地址**: `https://api.weixin.qq.com/cgi-bin/component/api_component_token`

### 2. 获取预授权码
- **方法**: `GetPreAuthCode(ctx context.Context)`
- **请求参数**: `component_appid` - 第三方平台AppID
- **返回值**: `pre_auth_code` (有效期600秒)
- **API地址**: `https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode`

### 3. 使用授权码换取授权信息
- **方法**: `QueryAuth(ctx context.Context, authorizationCode string)`
- **请求参数**: 
  - `component_appid` - 第三方平台AppID
  - `authorization_code` - 授权码
- **返回值**: 授权方信息，包括`authorizer_access_token`、`authorizer_refresh_token`等
- **API地址**: `https://api.weixin.qq.com/cgi-bin/component/api_query_auth`

### 4. 刷新授权方Token
- **方法**: `RefreshAuthorizerToken(ctx context.Context, authorizerAppID, refreshToken string)`
- **请求参数**: 
  - `component_appid` - 第三方平台AppID
  - `authorizer_appid` - 授权方AppID
  - `authorizer_refresh_token` - 授权方刷新Token
- **返回值**: 新的`authorizer_access_token`和`authorizer_refresh_token`
- **API地址**: `https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token`

### 5. 客户端创建
- `NewWeGoWithStorage(storage TokenStorage, config *OpenPlatformConfig)` - 使用自定义存储创建客户端
- `SaveVerifyTicket(ctx context.Context, ticket string)` - 保存VerifyTicket（用于获取ComponentAccessToken）

通过WeGo库，您可以轻松管理各种Token，确保第三方平台服务的稳定运行。