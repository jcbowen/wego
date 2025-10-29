package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileStorage 文件存储实现
// 将令牌数据持久化到本地文件系统
type FileStorage struct {
	mu                        sync.RWMutex
	baseDir                   string
	componentTokenFile        string
	preAuthCodeFile           string
	componentVerifyTicketFile string
	authorizerTokensDir       string
	prevEncodingAESKeysDir    string // 上一次EncodingAESKey存储目录
}

// NewFileStorage 创建文件存储实例
func NewFileStorage(baseDir string) (*FileStorage, error) {
	// 确保基础目录存在
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}

	storage := &FileStorage{
		baseDir:                   baseDir,
		componentTokenFile:        filepath.Join(baseDir, "component_token.json"),
		preAuthCodeFile:           filepath.Join(baseDir, "pre_auth_code.json"),
		componentVerifyTicketFile: filepath.Join(baseDir, "component_verify_ticket.txt"),
		authorizerTokensDir:       filepath.Join(baseDir, "authorizer_tokens"),
		prevEncodingAESKeysDir:    filepath.Join(baseDir, "prev_encoding_aes_keys"),
	}

	// 确保授权方令牌目录存在
	if err := os.MkdirAll(storage.authorizerTokensDir, 0755); err != nil {
		return nil, err
	}

	// 确保上一次EncodingAESKey存储目录存在
	if err := os.MkdirAll(storage.prevEncodingAESKeysDir, 0755); err != nil {
		return nil, err
	}

	return storage, nil
}

// SaveComponentToken 保存组件令牌到文件
func (s *FileStorage) SaveComponentToken(ctx context.Context, token *ComponentAccessToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.saveToFile(s.componentTokenFile, token)
}

// GetComponentToken 从文件读取组件令牌
func (s *FileStorage) GetComponentToken(ctx context.Context) (*ComponentAccessToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var token ComponentAccessToken
	if err := s.loadFromFile(s.componentTokenFile, &token); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// 检查是否过期
	if time.Now().After(token.ExpiresAt) {
		return nil, nil
	}

	return &token, nil
}

// DeleteComponentToken 删除组件令牌文件
func (s *FileStorage) DeleteComponentToken(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return os.Remove(s.componentTokenFile)
}

// SavePreAuthCode 保存预授权码到文件
func (s *FileStorage) SavePreAuthCode(ctx context.Context, code *PreAuthCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.saveToFile(s.preAuthCodeFile, code)
}

// GetPreAuthCode 从文件读取预授权码
func (s *FileStorage) GetPreAuthCode(ctx context.Context) (*PreAuthCode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var code PreAuthCode
	if err := s.loadFromFile(s.preAuthCodeFile, &code); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// 检查是否过期
	if time.Now().After(code.ExpiresAt) {
		return nil, nil
	}

	return &code, nil
}

// DeletePreAuthCode 删除预授权码文件
func (s *FileStorage) DeletePreAuthCode(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return os.Remove(s.preAuthCodeFile)
}

// SaveVerifyTicket 保存验证票据到文件
func (s *FileStorage) SaveComponentVerifyTicket(ctx context.Context, ticket string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建票据结构，记录创建时间和过期时间
	verifyTicket := &ComponentVerifyTicket{
		Ticket:    ticket,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(12 * time.Hour), // 12小时有效期
	}

	return s.saveToFile(s.componentVerifyTicketFile, verifyTicket)
}

// GetVerifyTicket 从文件读取验证票据
func (s *FileStorage) GetComponentVerifyTicket(ctx context.Context) (*ComponentVerifyTicket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var verifyTicket ComponentVerifyTicket
	if err := s.loadFromFile(s.componentVerifyTicketFile, &verifyTicket); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// 检查票据是否过期
	if time.Now().After(verifyTicket.ExpiresAt) {
		return nil, nil
	}

	return &verifyTicket, nil
}

// DeleteVerifyTicket 删除验证票据文件
func (s *FileStorage) DeleteComponentVerifyTicket(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return os.Remove(s.componentVerifyTicketFile)
}

// SaveAuthorizerToken 保存授权方令牌到文件
func (s *FileStorage) SaveAuthorizerToken(ctx context.Context, authorizerAppID string, token *AuthorizerAccessToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.authorizerTokensDir, authorizerAppID+".json")
	return s.saveToFile(filename, token)
}

// GetAuthorizerToken 从文件读取授权方令牌
func (s *FileStorage) GetAuthorizerToken(ctx context.Context, authorizerAppID string) (*AuthorizerAccessToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := filepath.Join(s.authorizerTokensDir, authorizerAppID+".json")
	var token AuthorizerAccessToken
	if err := s.loadFromFile(filename, &token); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// 检查是否过期
	if time.Now().After(token.ExpiresAt) {
		return nil, nil
	}

	return &token, nil
}

// DeleteAuthorizerToken 删除授权方令牌文件
func (s *FileStorage) DeleteAuthorizerToken(ctx context.Context, authorizerAppID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.authorizerTokensDir, authorizerAppID+".json")
	return os.Remove(filename)
}

// ClearAuthorizerTokens 清除所有授权方令牌
func (s *FileStorage) ClearAuthorizerTokens(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	files, err := os.ReadDir(s.authorizerTokensDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			filename := filepath.Join(s.authorizerTokensDir, file.Name())
			os.Remove(filename)
		}
	}

	return nil
}

// ListAuthorizerTokens 列出所有已存储的授权方appid
func (s *FileStorage) ListAuthorizerTokens(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	files, err := os.ReadDir(s.authorizerTokensDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	appids := make([]string, 0, len(files))
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			appid := file.Name()[:len(file.Name())-5] // 移除.json后缀
			appids = append(appids, appid)
		}
	}

	return appids, nil
}

// Ping 存储健康检查
func (s *FileStorage) Ping(ctx context.Context) error {
	// 检查基础目录是否可写
	testFile := filepath.Join(s.baseDir, ".ping_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return err
	}
	os.Remove(testFile)
	return nil
}

// saveToFile 将数据保存到文件
func (s *FileStorage) saveToFile(filename string, data interface{}) error {
	// 创建临时文件
	tempFile := filename + ".tmp"

	file, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		os.Remove(tempFile)
		return err
	}

	// 原子性替换文件
	if err := file.Close(); err != nil {
		os.Remove(tempFile)
		return err
	}

	return os.Rename(tempFile, filename)
}

// loadFromFile 从文件加载数据
func (s *FileStorage) loadFromFile(filename string, data interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(data)
}

// SavePrevEncodingAESKey 保存上一次的EncodingAESKey到文件
func (s *FileStorage) SavePrevEncodingAESKey(ctx context.Context, appID string, prevKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prevKeyData := &PrevEncodingAESKey{
		AppID:              appID,
		PrevEncodingAESKey: prevKey,
		UpdatedAt:          time.Now(),
	}

	filename := filepath.Join(s.prevEncodingAESKeysDir, appID+".json")
	return s.saveToFile(filename, prevKeyData)
}

// GetPrevEncodingAESKey 从文件读取上一次的EncodingAESKey
func (s *FileStorage) GetPrevEncodingAESKey(ctx context.Context, appID string) (*PrevEncodingAESKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := filepath.Join(s.prevEncodingAESKeysDir, appID+".json")
	var prevKey PrevEncodingAESKey
	if err := s.loadFromFile(filename, &prevKey); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return &prevKey, nil
}

// DeletePrevEncodingAESKey 删除上一次的EncodingAESKey文件
func (s *FileStorage) DeletePrevEncodingAESKey(ctx context.Context, appID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.prevEncodingAESKeysDir, appID+".json")
	return os.Remove(filename)
}
