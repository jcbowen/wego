package storage

import (
	"context"
	"encoding/json"
	"fmt"
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
	wechatPayCertsDir         string // 微信支付平台证书存储目录
}

// NewFileStorage 创建文件存储实例
func NewFileStorage(baseDir string) (*FileStorage, error) {
	// 确保基础目录存在；如果创建失败，回退到系统临时目录
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		fallback := filepath.Join(os.TempDir(), "wego_storage")
		if err2 := os.MkdirAll(fallback, 0755); err2 != nil {
			return nil, err
		}
		baseDir = fallback
	}

	storage := &FileStorage{
		baseDir:                   baseDir,
		componentTokenFile:        filepath.Join(baseDir, "component_token.json"),
		preAuthCodeFile:           filepath.Join(baseDir, "pre_auth_code.json"),
		componentVerifyTicketFile: filepath.Join(baseDir, "component_verify_ticket.json"),
		authorizerTokensDir:       filepath.Join(baseDir, "authorizer_tokens"),
		prevEncodingAESKeysDir:    filepath.Join(baseDir, "prev_encoding_aes_keys"),
		wechatPayCertsDir:         filepath.Join(baseDir, "wechat_pay_certs"),
	}

	// 确保授权方令牌目录存在
	if err := os.MkdirAll(storage.authorizerTokensDir, 0755); err != nil {
		return nil, err
	}

	// 确保上一次EncodingAESKey存储目录存在
	if err := os.MkdirAll(storage.prevEncodingAESKeysDir, 0755); err != nil {
		return nil, err
	}

	// 确保微信支付平台证书目录存在
	if err := os.MkdirAll(storage.wechatPayCertsDir, 0755); err != nil {
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

	return &token, nil
}

// DeleteComponentToken 删除组件令牌文件
func (s *FileStorage) DeleteComponentToken(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return removeFileIfExists(s.componentTokenFile)
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

	return &code, nil
}

// DeletePreAuthCode 删除预授权码文件
func (s *FileStorage) DeletePreAuthCode(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return removeFileIfExists(s.preAuthCodeFile)
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

	return &verifyTicket, nil
}

// DeleteVerifyTicket 删除验证票据文件
func (s *FileStorage) DeleteComponentVerifyTicket(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return removeFileIfExists(s.componentVerifyTicketFile)
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

	return &token, nil
}

// DeleteAuthorizerToken 删除授权方令牌文件
func (s *FileStorage) DeleteAuthorizerToken(ctx context.Context, authorizerAppID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.authorizerTokensDir, authorizerAppID+".json")
	return removeFileIfExists(filename)
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
			if err := removeFileIfExists(filename); err != nil {
				return err
			}
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

// saveToFile 将数据原子化保存到文件
//
// 原子写流程:
//  1. 通过 os.CreateTemp 在目标目录生成唯一临时文件，避免多进程写固定 .tmp 名互踩（Windows/Linux 通用）
//  2. 写入 JSON 数据后调用 Sync 刷盘，再手动 Close（Windows 上 rename 前文件必须已关闭）
//  3. 按目标文件原有权限模式修正临时文件权限（目标不存在时使用 0600，避免令牌等敏感数据宽权限落盘）
//  4. 通过 Rename 原子替换目标文件；任一步骤失败时兜底清理（关闭并删除）临时文件，不影响原目标文件
func (s *FileStorage) saveToFile(filename string, data interface{}) error {
	tmp, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".tmp")
	if err != nil {
		return err
	}
	tempFile := tmp.Name()

	// 失败路径兜底清理；rename 成功后置空 tempFile 以跳过清理
	defer func() {
		if tempFile != "" {
			_ = tmp.Close() // 已手动 Close 时重复调用返回错误，忽略即可
			_ = os.Remove(tempFile)
		}
	}()

	encoder := json.NewEncoder(tmp)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return err
	}

	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	// 权限处理：目标文件已存在时保留其原有权限模式，否则使用 0600
	mode := os.FileMode(0600)
	if info, err := os.Stat(filename); err == nil {
		mode = info.Mode().Perm()
	}
	if err := os.Chmod(tempFile, mode); err != nil {
		return err
	}

	if err := os.Rename(tempFile, filename); err != nil {
		return err
	}
	tempFile = ""
	return nil
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

// removeFileIfExists 删除文件，若文件不存在则返回 nil
func removeFileIfExists(filename string) error {
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
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
	return removeFileIfExists(filename)
}

// SaveWechatPayCert 保存微信支付平台证书到文件
func (s *FileStorage) SaveWechatPayCert(ctx context.Context, mchID string, cert *WechatPayCertificate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if mchID == "" {
		return fmt.Errorf("mchID cannot be empty")
	}
	if cert == nil {
		return fmt.Errorf("cert cannot be nil")
	}
	if cert.SerialNo == "" {
		return fmt.Errorf("serialNo cannot be empty")
	}

	dir := filepath.Join(s.wechatPayCertsDir, mchID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(dir, cert.SerialNo+".json")
	return s.saveToFile(filename, cert)
}

// GetWechatPayCert 从文件读取指定序列号的微信支付平台证书
func (s *FileStorage) GetWechatPayCert(ctx context.Context, mchID, serialNo string) (*WechatPayCertificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if mchID == "" || serialNo == "" {
		return nil, fmt.Errorf("mchID and serialNo cannot be empty")
	}

	filename := filepath.Join(s.wechatPayCertsDir, mchID, serialNo+".json")
	var cert WechatPayCertificate
	if err := s.loadFromFile(filename, &cert); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return &cert, nil
}

// GetLatestWechatPayCert 从文件读取最新的微信支付平台证书
func (s *FileStorage) GetLatestWechatPayCert(ctx context.Context, mchID string) (*WechatPayCertificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if mchID == "" {
		return nil, fmt.Errorf("mchID cannot be empty")
	}

	dir := filepath.Join(s.wechatPayCertsDir, mchID)
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var latest *WechatPayCertificate
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filename := filepath.Join(dir, file.Name())
		var cert WechatPayCertificate
		if err := s.loadFromFile(filename, &cert); err != nil {
			continue
		}

		if latest == nil || cert.ExpiresAt.After(latest.ExpiresAt) {
			latest = &cert
		}
	}

	return latest, nil
}

// DeleteWechatPayCert 从文件删除指定微信支付平台证书
func (s *FileStorage) DeleteWechatPayCert(ctx context.Context, mchID, serialNo string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if mchID == "" || serialNo == "" {
		return fmt.Errorf("mchID and serialNo cannot be empty")
	}

	filename := filepath.Join(s.wechatPayCertsDir, mchID, serialNo+".json")
	return removeFileIfExists(filename)
}

// ListWechatPayCertSerialNos 列出文件中指定商户的所有微信支付平台证书序列号
func (s *FileStorage) ListWechatPayCertSerialNos(ctx context.Context, mchID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if mchID == "" {
		return nil, fmt.Errorf("mchID cannot be empty")
	}

	dir := filepath.Join(s.wechatPayCertsDir, mchID)
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	serialNos := make([]string, 0, len(files))
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			serialNos = append(serialNos, file.Name()[:len(file.Name())-5])
		}
	}

	return serialNos, nil
}
