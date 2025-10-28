package wego

import (
	"context"
	"sync"
	"time"
)

// TokenStorage 令牌存储接口
type TokenStorage interface {
	// 组件令牌相关方法
	SaveComponentToken(ctx context.Context, token *ComponentAccessToken) error
	GetComponentToken(ctx context.Context) (*ComponentAccessToken, error)
	DeleteComponentToken(ctx context.Context) error

	// 预授权码相关方法
	SavePreAuthCode(ctx context.Context, code *PreAuthCode) error
	GetPreAuthCode(ctx context.Context) (*PreAuthCode, error)
	DeletePreAuthCode(ctx context.Context) error

	// 授权方令牌相关方法
	SaveAuthorizerToken(ctx context.Context, authorizerAppID string, token *AuthorizerAccessToken) error
	GetAuthorizerToken(ctx context.Context, authorizerAppID string) (*AuthorizerAccessToken, error)
	DeleteAuthorizerToken(ctx context.Context, authorizerAppID string) error
	ClearAuthorizerTokens(ctx context.Context) error            // 清除所有授权方令牌
	ListAuthorizerTokens(ctx context.Context) ([]string, error) // 返回所有已存储的授权方appid

	// 存储健康检查
	Ping(ctx context.Context) error
}

// MemoryStorage 内存存储实现（默认）
type MemoryStorage struct {
	mu               sync.RWMutex
	componentToken   *ComponentAccessToken
	preAuthCode      *PreAuthCode
	authorizerTokens map[string]*AuthorizerAccessToken
}

// NewMemoryStorage 创建内存存储实例
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		authorizerTokens: make(map[string]*AuthorizerAccessToken),
	}
}

func (s *MemoryStorage) SaveComponentToken(ctx context.Context, token *ComponentAccessToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.componentToken = token
	return nil
}

func (s *MemoryStorage) GetComponentToken(ctx context.Context) (*ComponentAccessToken, error) {
	s.mu.RLock()
	if s.componentToken == nil {
		s.mu.RUnlock()
		return nil, nil
	}
	// 检查是否过期
	if time.Now().After(s.componentToken.ExpiresAt) {
		s.mu.RUnlock()
		s.mu.Lock()
		s.componentToken = nil
		s.mu.Unlock()
		return nil, nil
	}
	token := s.componentToken
	s.mu.RUnlock()
	return token, nil
}

func (s *MemoryStorage) DeleteComponentToken(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.componentToken = nil
	return nil
}

func (s *MemoryStorage) SavePreAuthCode(ctx context.Context, code *PreAuthCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.preAuthCode = code
	return nil
}

func (s *MemoryStorage) GetPreAuthCode(ctx context.Context) (*PreAuthCode, error) {
	s.mu.RLock()
	if s.preAuthCode == nil {
		s.mu.RUnlock()
		return nil, nil
	}
	// 检查是否过期
	if time.Now().After(s.preAuthCode.ExpiresAt) {
		s.mu.RUnlock()
		s.mu.Lock()
		s.preAuthCode = nil
		s.mu.Unlock()
		return nil, nil
	}
	code := s.preAuthCode
	s.mu.RUnlock()
	return code, nil
}

func (s *MemoryStorage) DeletePreAuthCode(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.preAuthCode = nil
	return nil
}

func (s *MemoryStorage) SaveAuthorizerToken(ctx context.Context, authorizerAppID string, token *AuthorizerAccessToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.authorizerTokens[authorizerAppID] = token
	return nil
}

func (s *MemoryStorage) GetAuthorizerToken(ctx context.Context, authorizerAppID string) (*AuthorizerAccessToken, error) {
	s.mu.RLock()
	token, exists := s.authorizerTokens[authorizerAppID]
	if !exists || token == nil {
		s.mu.RUnlock()
		return nil, nil
	}
	// 检查是否过期
	if time.Now().After(token.ExpiresAt) {
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.authorizerTokens, authorizerAppID)
		s.mu.Unlock()
		return nil, nil
	}
	tokenCopy := token
	s.mu.RUnlock()
	return tokenCopy, nil
}

func (s *MemoryStorage) DeleteAuthorizerToken(ctx context.Context, authorizerAppID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.authorizerTokens, authorizerAppID)
	return nil
}

func (s *MemoryStorage) ClearAuthorizerTokens(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.authorizerTokens = make(map[string]*AuthorizerAccessToken)
	return nil
}

func (s *MemoryStorage) ListAuthorizerTokens(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	appids := make([]string, 0, len(s.authorizerTokens))
	for appid := range s.authorizerTokens {
		appids = append(appids, appid)
	}
	return appids, nil
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}
