package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jcbowen/jcbaseGo/component/redis"
)

// RedisStorage Redis存储实现
// 基于jcbaseGo的redis组件实现TokenStorage接口
// 使用Redis作为持久化存储，支持自动过期和分布式部署
//
// 键命名规则：
// - component_token: 组件令牌
// - pre_auth_code: 预授权码
// - verify_ticket: 验证票据
// - authorizer_token:{appid}: 授权方令牌
// - prev_aes_key:{appid}: 上一次的EncodingAESKey
// - authorizer_appids: 授权方appid集合

type RedisStorage struct {
	client    *redis.Instance // jcbaseGo Redis实例
	keyPrefix string          // 键前缀，用于区分不同应用实例
}

// RedisConfig Redis存储配置选项
type RedisConfig struct {
	RedisInstance *redis.Instance // jcbaseGo Redis实例
	KeyPrefix     string          // 键前缀，用于区分不同应用实例，默认"wego:"
}

// NewRedisStorage 创建Redis存储实例
//
// 参数:
//   config: Redis配置选项
//
// 返回:
//   *RedisStorage: Redis存储实例
//   error: 配置验证失败时返回错误
//
// 示例:
//   redisInstance := redis.New(redis.Config{Addr: "localhost:6379"})
//   storage, err := NewRedisStorage(&RedisConfig{
//       RedisInstance: redisInstance,
//       KeyPrefix: "myapp:",
//   })
func NewRedisStorage(config *RedisConfig) (*RedisStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	
	if config.RedisInstance == nil {
		return nil, fmt.Errorf("redis instance cannot be nil")
	}
	
	if config.KeyPrefix == "" {
		config.KeyPrefix = "wego:"
	}
	
	return &RedisStorage{
		client:    config.RedisInstance,
		keyPrefix: config.KeyPrefix,
	}, nil
}

// buildKey 构建完整的Redis键名
//
// 参数:
//   parts: 键名组成部分
//
// 返回:
//   string: 完整的Redis键名
func (s *RedisStorage) buildKey(parts ...string) string {
	return s.keyPrefix + strings.Join(parts, ":")
}

// Ping 检查Redis连接状态
//
// 参数:
//   ctx: 上下文
//
// 返回:
//   error: 连接正常返回nil，否则返回错误
func (s *RedisStorage) Ping(ctx context.Context) error {
	_, err := s.client.Ping()
	return err
}

// SaveComponentToken 保存组件令牌到Redis
//
// 参数:
//   ctx: 上下文
//   token: 组件令牌信息
//
// 返回:
//   error: 保存失败时返回错误
func (s *RedisStorage) SaveComponentToken(ctx context.Context, token *ComponentAccessToken) error {
	if token == nil {
		return fmt.Errorf("token cannot be nil")
	}
	
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}
	
	key := s.buildKey("component_token")
	expireAt := time.Until(token.ExpiresAt)
	
	if err := s.client.Set(key, string(data), expireAt); err != nil {
		return fmt.Errorf("failed to save component token: %w", err)
	}
	
	return nil
}

// GetComponentToken 从Redis获取组件令牌
//
// 参数:
//   ctx: 上下文
//
// 返回:
//   *ComponentAccessToken: 组件令牌信息，不存在或过期返回nil
//   error: 获取失败时返回错误
func (s *RedisStorage) GetComponentToken(ctx context.Context) (*ComponentAccessToken, error) {
	key := s.buildKey("component_token")
	
	data, err := s.client.GetString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get component token: %w", err)
	}
	if data == "" {
		return nil, nil
	}
	
	var token ComponentAccessToken
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal component token: %w", err)
	}
	
	// 检查是否过期
	if time.Now().After(token.ExpiresAt) {
		// 令牌已过期，删除它
		if err := s.DeleteComponentToken(ctx); err != nil {
			return nil, fmt.Errorf("failed to delete expired token: %w", err)
		}
		return nil, nil
	}
	
	return &token, nil
}

// DeleteComponentToken 从Redis删除组件令牌
//
// 参数:
//   ctx: 上下文
//
// 返回:
//   error: 删除失败时返回错误
func (s *RedisStorage) DeleteComponentToken(ctx context.Context) error {
	key := s.buildKey("component_token")
	
	if err := s.client.Del(key); err != nil {
		return fmt.Errorf("failed to delete component token: %w", err)
	}
	
	return nil
}

// SavePreAuthCode 保存预授权码到Redis
//
// 参数:
//   ctx: 上下文
//   code: 预授权码信息
//
// 返回:
//   error: 保存失败时返回错误
func (s *RedisStorage) SavePreAuthCode(ctx context.Context, code *PreAuthCode) error {
	if code == nil {
		return fmt.Errorf("pre auth code cannot be nil")
	}
	
	data, err := json.Marshal(code)
	if err != nil {
		return fmt.Errorf("failed to marshal pre auth code: %w", err)
	}
	
	key := s.buildKey("pre_auth_code")
	expireAt := time.Until(code.ExpiresAt)
	
	if err := s.client.Set(key, string(data), expireAt); err != nil {
		return fmt.Errorf("failed to save pre auth code: %w", err)
	}
	
	return nil
}

// GetPreAuthCode 从Redis获取预授权码
//
// 参数:
//   ctx: 上下文
//
// 返回:
//   *PreAuthCode: 预授权码信息，不存在或过期返回nil
//   error: 获取失败时返回错误
func (s *RedisStorage) GetPreAuthCode(ctx context.Context) (*PreAuthCode, error) {
	key := s.buildKey("pre_auth_code")
	
	data, err := s.client.GetString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get pre auth code: %w", err)
	}
	if data == "" {
		return nil, nil
	}
	
	var code PreAuthCode
	if err := json.Unmarshal([]byte(data), &code); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pre auth code: %w", err)
	}
	
	// 检查是否过期
	if time.Now().After(code.ExpiresAt) {
		// 预授权码已过期，删除它
		if err := s.DeletePreAuthCode(ctx); err != nil {
			return nil, fmt.Errorf("failed to delete expired pre auth code: %w", err)
		}
		return nil, nil
	}
	
	return &code, nil
}

// DeletePreAuthCode 从Redis删除预授权码
//
// 参数:
//   ctx: 上下文
//
// 返回:
//   error: 删除失败时返回错误
func (s *RedisStorage) DeletePreAuthCode(ctx context.Context) error {
	key := s.buildKey("pre_auth_code")
	
	if err := s.client.Del(key); err != nil {
		return fmt.Errorf("failed to delete pre auth code: %w", err)
	}
	
	return nil
}

// SaveComponentVerifyTicket 保存验证票据到Redis
//
// 参数:
//   ctx: 上下文
//   ticket: 票据内容
//
// 返回:
//   error: 保存失败时返回错误
func (s *RedisStorage) SaveComponentVerifyTicket(ctx context.Context, ticket string) error {
	if ticket == "" {
		return fmt.Errorf("ticket cannot be empty")
	}
	
	ticketData := &ComponentVerifyTicket{
		Ticket:    ticket,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(12 * time.Hour), // 12小时有效期
	}
	
	data, err := json.Marshal(ticketData)
	if err != nil {
		return fmt.Errorf("failed to marshal verify ticket: %w", err)
	}
	
	key := s.buildKey("verify_ticket")
	expireAt := time.Until(ticketData.ExpiresAt)
	
	if err := s.client.Set(key, string(data), expireAt); err != nil {
		return fmt.Errorf("failed to save verify ticket: %w", err)
	}
	
	return nil
}

// GetComponentVerifyTicket 从Redis获取验证票据
//
// 参数:
//   ctx: 上下文
//
// 返回:
//   *ComponentVerifyTicket: 验证票据信息，包含创建时间
//   error: 获取失败时返回错误
func (s *RedisStorage) GetComponentVerifyTicket(ctx context.Context) (*ComponentVerifyTicket, error) {
	key := s.buildKey("verify_ticket")
	
	data, err := s.client.GetString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get verify ticket: %w", err)
	}
	if data == "" {
		return nil, nil
	}
	
	var ticket ComponentVerifyTicket
	if err := json.Unmarshal([]byte(data), &ticket); err != nil {
		return nil, fmt.Errorf("failed to unmarshal verify ticket: %w", err)
	}
	
	// 检查是否过期
	if time.Now().After(ticket.ExpiresAt) {
		// 票据已过期，删除它
		if err := s.DeleteComponentVerifyTicket(ctx); err != nil {
			return nil, fmt.Errorf("failed to delete expired verify ticket: %w", err)
		}
		return nil, nil
	}
	
	return &ticket, nil
}

// DeleteComponentVerifyTicket 从Redis删除验证票据
//
// 参数:
//   ctx: 上下文
//
// 返回:
//   error: 删除失败时返回错误
func (s *RedisStorage) DeleteComponentVerifyTicket(ctx context.Context) error {
	key := s.buildKey("verify_ticket")
	
	if err := s.client.Del(key); err != nil {
		return fmt.Errorf("failed to delete verify ticket: %w", err)
	}
	
	return nil
}

// SaveAuthorizerToken 保存授权方令牌到Redis
//
// 参数:
//   ctx: 上下文
//   authorizerAppID: 授权方应用ID
//   token: 授权方令牌信息
//
// 返回:
//   error: 保存失败时返回错误
func (s *RedisStorage) SaveAuthorizerToken(ctx context.Context, authorizerAppID string, token *AuthorizerAccessToken) error {
	if authorizerAppID == "" {
		return fmt.Errorf("authorizer app id cannot be empty")
	}
	if token == nil {
		return fmt.Errorf("token cannot be nil")
	}
	
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal authorizer token: %w", err)
	}
	
	// 保存令牌
	tokenKey := s.buildKey("authorizer_token", authorizerAppID)
	expireAt := time.Until(token.ExpiresAt)
	
	if err := s.client.Set(tokenKey, string(data), expireAt); err != nil {
		return fmt.Errorf("failed to save authorizer token: %w", err)
	}
	
	// 将appid添加到集合中
	setKey := s.buildKey("authorizer_appids")
	if _, err := s.client.SAdd(setKey, authorizerAppID); err != nil {
		return fmt.Errorf("failed to add authorizer appid to set: %w", err)
	}
	
	return nil
}

// GetAuthorizerToken 从Redis获取授权方令牌
//
// 参数:
//   ctx: 上下文
//   authorizerAppID: 授权方应用ID
//
// 返回:
//   *AuthorizerAccessToken: 授权方令牌信息，不存在或过期返回nil
//   error: 获取失败时返回错误
func (s *RedisStorage) GetAuthorizerToken(ctx context.Context, authorizerAppID string) (*AuthorizerAccessToken, error) {
	if authorizerAppID == "" {
		return nil, fmt.Errorf("authorizer app id cannot be empty")
	}
	
	tokenKey := s.buildKey("authorizer_token", authorizerAppID)
	
	data, err := s.client.GetString(tokenKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorizer token: %w", err)
	}
	if data == "" {
		return nil, nil
	}
	
	var token AuthorizerAccessToken
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal authorizer token: %w", err)
	}
	
	// 检查是否过期
	if time.Now().After(token.ExpiresAt) {
		// 令牌已过期，删除它
		if err := s.DeleteAuthorizerToken(ctx, authorizerAppID); err != nil {
			return nil, fmt.Errorf("failed to delete expired authorizer token: %w", err)
		}
		return nil, nil
	}
	
	return &token, nil
}

// DeleteAuthorizerToken 从Redis删除授权方令牌
//
// 参数:
//   ctx: 上下文
//   authorizerAppID: 授权方应用ID
//
// 返回:
//   error: 删除失败时返回错误
func (s *RedisStorage) DeleteAuthorizerToken(ctx context.Context, authorizerAppID string) error {
	if authorizerAppID == "" {
		return fmt.Errorf("authorizer app id cannot be empty")
	}
	
	tokenKey := s.buildKey("authorizer_token", authorizerAppID)
	setKey := s.buildKey("authorizer_appids")
	
	// 删除令牌
	if err := s.client.Del(tokenKey); err != nil {
		return fmt.Errorf("failed to delete authorizer token: %w", err)
	}
	
	// 从集合中移除appid
	if _, err := s.client.SRem(setKey, authorizerAppID); err != nil {
		return fmt.Errorf("failed to remove authorizer appid from set: %w", err)
	}
	
	return nil
}

// ClearAuthorizerTokens 清除所有授权方令牌
//
// 参数:
//   ctx: 上下文
//
// 返回:
//   error: 清除失败时返回错误
func (s *RedisStorage) ClearAuthorizerTokens(ctx context.Context) error {
	setKey := s.buildKey("authorizer_appids")
	
	// 获取所有appid
	appids, err := s.client.SMembers(setKey)
	if err != nil {
		return fmt.Errorf("failed to get authorizer appids: %w", err)
	}
	
	// 删除所有令牌
	for _, appid := range appids {
		tokenKey := s.buildKey("authorizer_token", appid)
		if err := s.client.Del(tokenKey); err != nil {
			return fmt.Errorf("failed to delete authorizer token for %s: %w", appid, err)
		}
	}
	
	// 删除集合
	if err := s.client.Del(setKey); err != nil {
		return fmt.Errorf("failed to delete authorizer appids set: %w", err)
	}
	
	return nil
}

// ListAuthorizerTokens 返回所有已存储的授权方appid
//
// 参数:
//   ctx: 上下文
//
// 返回:
//   []string: 授权方appid列表
//   error: 获取失败时返回错误
func (s *RedisStorage) ListAuthorizerTokens(ctx context.Context) ([]string, error) {
	setKey := s.buildKey("authorizer_appids")
	
	appids, err := s.client.SMembers(setKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorizer appids: %w", err)
	}
	
	// 过滤掉已过期的令牌
	validAppids := make([]string, 0, len(appids))
	for _, appid := range appids {
		token, err := s.GetAuthorizerToken(ctx, appid)
		if err != nil {
			return nil, fmt.Errorf("failed to check token for %s: %w", appid, err)
		}
		if token != nil {
			validAppids = append(validAppids, appid)
		}
	}
	
	return validAppids, nil
}

// SavePrevEncodingAESKey 保存上一次的EncodingAESKey到Redis
//
// 参数:
//   ctx: 上下文
//   appID: 应用ID
//   prevKey: 上一次的EncodingAESKey
//
// 返回:
//   error: 保存失败时返回错误
func (s *RedisStorage) SavePrevEncodingAESKey(ctx context.Context, appID string, prevKey string) error {
	if appID == "" {
		return fmt.Errorf("app id cannot be empty")
	}
	if prevKey == "" {
		return fmt.Errorf("previous encoding aes key cannot be empty")
	}
	
	prevKeyData := &PrevEncodingAESKey{
		AppID:              appID,
		PrevEncodingAESKey: prevKey,
		UpdatedAt:          time.Now(),
	}
	
	data, err := json.Marshal(prevKeyData)
	if err != nil {
		return fmt.Errorf("failed to marshal previous encoding aes key: %w", err)
	}
	
	key := s.buildKey("prev_aes_key", appID)
	
	// 保存上一次的EncodingAESKey，不设置过期时间
	if err := s.client.Set(key, string(data), 0); err != nil {
		return fmt.Errorf("failed to save previous encoding aes key: %w", err)
	}
	
	return nil
}

// GetPrevEncodingAESKey 从Redis获取上一次的EncodingAESKey
//
// 参数:
//   ctx: 上下文
//   appID: 应用ID
//
// 返回:
//   *PrevEncodingAESKey: 上一次的EncodingAESKey信息
//   error: 获取失败时返回错误
func (s *RedisStorage) GetPrevEncodingAESKey(ctx context.Context, appID string) (*PrevEncodingAESKey, error) {
	if appID == "" {
		return nil, fmt.Errorf("app id cannot be empty")
	}
	
	key := s.buildKey("prev_aes_key", appID)
	
	data, err := s.client.GetString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous encoding aes key: %w", err)
	}
	if data == "" {
		return nil, nil
	}
	
	var prevKey PrevEncodingAESKey
	if err := json.Unmarshal([]byte(data), &prevKey); err != nil {
		return nil, fmt.Errorf("failed to unmarshal previous encoding aes key: %w", err)
	}
	
	return &prevKey, nil
}

// DeletePrevEncodingAESKey 从Redis删除上一次的EncodingAESKey
//
// 参数:
//   ctx: 上下文
//   appID: 应用ID
//
// 返回:
//   error: 删除失败时返回错误
func (s *RedisStorage) DeletePrevEncodingAESKey(ctx context.Context, appID string) error {
	if appID == "" {
		return fmt.Errorf("app id cannot be empty")
	}
	
	key := s.buildKey("prev_aes_key", appID)
	
	if err := s.client.Del(key); err != nil {
		return fmt.Errorf("failed to delete previous encoding aes key: %w", err)
	}
	
	return nil
}