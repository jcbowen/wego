package storage

import (
	"context"
	"sync"
	"time"
)

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	mu                  sync.RWMutex
	componentToken      *ComponentAccessToken
	preAuthCode         *PreAuthCode
	verifyTicket        *ComponentVerifyTicket
	authorizerTokens    map[string]*AuthorizerAccessToken
	prevEncodingAESKeys map[string]*PrevEncodingAESKey // 上一次的EncodingAESKey存储
}

// NewMemoryStorage 创建内存存储实例
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		authorizerTokens:    make(map[string]*AuthorizerAccessToken),
		prevEncodingAESKeys: make(map[string]*PrevEncodingAESKey),
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

// DeletePreAuthCode 删除预授权码
func (s *MemoryStorage) DeletePreAuthCode(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.preAuthCode = nil
	return nil
}

// SaveVerifyTicket 保存验证票据
func (s *MemoryStorage) SaveComponentVerifyTicket(ctx context.Context, ticket string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建票据结构，记录创建时间和过期时间
	s.verifyTicket = &ComponentVerifyTicket{
		Ticket:    ticket,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(12 * time.Hour), // 12小时有效期
	}
	return nil
}

// GetVerifyTicket 获取验证票据
func (s *MemoryStorage) GetComponentVerifyTicket(ctx context.Context) (*ComponentVerifyTicket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.verifyTicket == nil {
		return nil, nil
	}

	// 检查票据是否过期
	if time.Now().After(s.verifyTicket.ExpiresAt) {
		s.mu.RUnlock()
		s.mu.Lock()
		s.verifyTicket = nil
		s.mu.Unlock()
		return nil, nil
	}

	return s.verifyTicket, nil
}

// DeleteVerifyTicket 删除验证票据
func (s *MemoryStorage) DeleteComponentVerifyTicket(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.verifyTicket = nil
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

// SavePrevEncodingAESKey 保存上一次的EncodingAESKey
func (s *MemoryStorage) SavePrevEncodingAESKey(ctx context.Context, appID string, prevKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.prevEncodingAESKeys[appID] = &PrevEncodingAESKey{
		AppID:              appID,
		PrevEncodingAESKey: prevKey,
		UpdatedAt:          time.Now(),
	}
	return nil
}

// GetPrevEncodingAESKey 获取上一次的EncodingAESKey
func (s *MemoryStorage) GetPrevEncodingAESKey(ctx context.Context, appID string) (*PrevEncodingAESKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if prevKey, exists := s.prevEncodingAESKeys[appID]; exists && prevKey != nil {
		return prevKey, nil
	}
	return nil, nil
}

// DeletePrevEncodingAESKey 删除上一次的EncodingAESKey
func (s *MemoryStorage) DeletePrevEncodingAESKey(ctx context.Context, appID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.prevEncodingAESKeys, appID)
	return nil
}
