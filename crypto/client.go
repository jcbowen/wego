package crypto

import (
	"crypto/sha1"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// CryptoClient 加密解密客户端
type CryptoClient struct{}

// NewCryptoClient 创建新的加密解密客户端
func NewCryptoClient() *CryptoClient {
	return &CryptoClient{}
}

// DecodeAESKey 解码微信开放平台AES密钥
func (c *CryptoClient) DecodeAESKey(encodingAESKey string) ([]byte, error) {
	return DecodeAESKey(encodingAESKey)
}

// EncryptMsg 加密微信消息
func (c *CryptoClient) EncryptMsg(plainText, appID string, aesKey []byte) (string, error) {
	return EncryptMsg(plainText, appID, aesKey)
}

// DecryptMsg 解密微信消息
func (c *CryptoClient) DecryptMsg(encryptedMsg string, aesKey []byte) (string, error) {
	return DecryptMsg(encryptedMsg, aesKey)
}

// CryptoCache 加密解密缓存
type CryptoCache struct {
	cache map[string]*WXBizMsgCrypt
	mu    sync.RWMutex
}

// NewCryptoCache 创建新的加密解密缓存
func NewCryptoCache() *CryptoCache {
	return &CryptoCache{
		cache: make(map[string]*WXBizMsgCrypt),
	}
}

// Get 获取缓存的加密解密实例
func (c *CryptoCache) Get(appID string) *WXBizMsgCrypt {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if crypto, exists := c.cache[appID]; exists {
		return crypto
	}
	return nil
}

// Set 设置缓存的加密解密实例
func (c *CryptoCache) Set(appID string, crypto *WXBizMsgCrypt) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.cache[appID] = crypto
}

// WXBizMsgCrypt 微信消息加解密实例
type WXBizMsgCrypt struct {
	Token          string
	EncodingAESKey string
	AppID          string
}

// NewWXBizMsgCrypt 创建新的微信消息加解密实例
func NewWXBizMsgCrypt(token, encodingAESKey, appID string) *WXBizMsgCrypt {
	return &WXBizMsgCrypt{
		Token:          token,
		EncodingAESKey: encodingAESKey,
		AppID:          appID,
	}
}

// EncryptMsg 加密消息
func (c *WXBizMsgCrypt) EncryptMsg(plainText, timestamp, nonce string) (string, error) {
	aesKey, err := DecodeAESKey(c.EncodingAESKey)
	if err != nil {
		return "", err
	}
	
	return EncryptMsg(plainText, c.AppID, aesKey)
}

// DecryptMsg 解密消息
func (c *WXBizMsgCrypt) DecryptMsg(msgSignature, timestamp, nonce, encryptedMsg string) (string, error) {
	aesKey, err := DecodeAESKey(c.EncodingAESKey)
	if err != nil {
		return "", err
	}
	
	return DecryptMsg(encryptedMsg, aesKey)
}

// VerifyURL 验证URL
func (c *WXBizMsgCrypt) VerifyURL(msgSignature, timestamp, nonce, echostr string) (string, error) {
	// 验证签名
	if !validateSignature(c.Token, timestamp, nonce, echostr, msgSignature) {
		return "", fmt.Errorf("签名验证失败")
	}
	
	// 解密echostr
	aesKey, err := DecodeAESKey(c.EncodingAESKey)
	if err != nil {
		return "", err
	}
	
	return DecryptMsg(echostr, aesKey)
}

// validateSignature 验证签名
func validateSignature(token, timestamp, nonce, encrypted, msgSignature string) bool {
	expectedSignature := genSignature(token, timestamp, nonce, encrypted)
	return expectedSignature == msgSignature
}

// genSignature 生成签名
func genSignature(token, timestamp, nonce, encrypted string) string {
	params := []string{token, timestamp, nonce, encrypted}
	sort.Strings(params)
	str := strings.Join(params, "")
	hash := sha1.Sum([]byte(str))
	return fmt.Sprintf("%x", hash)
}