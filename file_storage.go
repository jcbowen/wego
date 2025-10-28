package wego

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
	mu                  sync.RWMutex
	baseDir             string
	componentTokenFile  string
	preAuthCodeFile     string
	authorizerTokensDir string
}

// NewFileStorage 创建文件存储实例
func NewFileStorage(baseDir string) (*FileStorage, error) {
	// 确保基础目录存在
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}

	storage := &FileStorage{
		baseDir:             baseDir,
		componentTokenFile:  filepath.Join(baseDir, "component_token.json"),
		preAuthCodeFile:     filepath.Join(baseDir, "pre_auth_code.json"),
		authorizerTokensDir: filepath.Join(baseDir, "authorizer_tokens"),
	}

	// 确保授权方令牌目录存在
	if err := os.MkdirAll(storage.authorizerTokensDir, 0755); err != nil {
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
