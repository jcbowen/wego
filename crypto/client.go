package crypto

import (
	"context"
	"crypto/sha1"
	"fmt"
	"sort"
	"strings"
	"sync"
	
	"github.com/jcbowen/wego/storage"
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

// WXBizMsgCrypt 微信消息加解密实例（符合微信官方规范）
type WXBizMsgCrypt struct {
	Token           string
	EncodingAESKey  string
	PrevEncodingAESKey string // 上一次的EncodingAESKey（官方要求支持）
	AppID           string
	storage         storage.TokenStorage // 使用现有的存储系统
}

// NewWXBizMsgCrypt 创建新的微信消息加解密实例
func NewWXBizMsgCrypt(token, encodingAESKey, appID string) *WXBizMsgCrypt {
	return &WXBizMsgCrypt{
		Token:          token,
		EncodingAESKey: encodingAESKey,
		AppID:          appID,
	}
}

// NewWXBizMsgCryptWithStorage 创建新的微信消息加解密实例（使用存储系统）
func NewWXBizMsgCryptWithStorage(token, encodingAESKey, appID string, storage storage.TokenStorage) *WXBizMsgCrypt {
	crypto := &WXBizMsgCrypt{
		Token:          token,
		EncodingAESKey: encodingAESKey,
		AppID:          appID,
		storage:        storage,
	}
	
	// 从存储中加载上一次的EncodingAESKey
	crypto.loadPrevEncodingAESKey()
	
	return crypto
}

// SetPrevEncodingAESKey 设置上一次的EncodingAESKey（符合微信官方规范）
func (c *WXBizMsgCrypt) SetPrevEncodingAESKey(prevKey string) error {
	c.PrevEncodingAESKey = prevKey
	
	// 如果配置了存储系统，则保存到存储中
	if c.storage != nil {
		ctx := context.Background()
		err := c.storage.SavePrevEncodingAESKey(ctx, c.AppID, prevKey)
		if err != nil {
			return fmt.Errorf("保存上一次EncodingAESKey到存储失败: %v", err)
		}
	}
	
	return nil
}

// loadPrevEncodingAESKey 从存储中加载上一次的EncodingAESKey
func (c *WXBizMsgCrypt) loadPrevEncodingAESKey() {
	if c.storage == nil {
		return
	}
	
	ctx := context.Background()
	prevKey, err := c.storage.GetPrevEncodingAESKey(ctx, c.AppID)
	if err == nil && prevKey != nil {
		c.PrevEncodingAESKey = prevKey.PrevEncodingAESKey
	}
}

// EncryptMsg 加密消息（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Message_encryption_and_decryption.html" index="0">0</mcreference>
func (c *WXBizMsgCrypt) EncryptMsg(replyMsg, timestamp, nonce string) (string, string, error) {
	aesKey, err := DecodeAESKey(c.EncodingAESKey)
	if err != nil {
		return "", "", fmt.Errorf("EncodingAESKey解码失败: %v", err)
	}

	encrypted, err := EncryptMsg(replyMsg, c.AppID, aesKey)
	if err != nil {
		return "", "", fmt.Errorf("加密失败: %v", err)
	}

	// 生成签名（符合微信官方规范：token、timestamp、nonce、msg_encrypt）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	signature := genSignature(c.Token, timestamp, nonce, encrypted)
	return encrypted, signature, nil
}

// DecryptMsg 解密消息（符合微信官方规范，支持使用上一次的EncodingAESKey）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Message_encryption_and_decryption.html" index="0">0</mcreference>
func (c *WXBizMsgCrypt) DecryptMsg(msgSignature, timestamp, nonce, encryptedMsg string) (string, error) {
	// 首先验证消息签名（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	if !c.VerifySignature(msgSignature, timestamp, nonce, encryptedMsg) {
		return "", fmt.Errorf("消息签名验证失败")
	}

	// 首先使用当前的EncodingAESKey尝试解密
	aesKey, err := DecodeAESKey(c.EncodingAESKey)
	if err != nil {
		return "", fmt.Errorf("当前EncodingAESKey解码失败: %v", err)
	}

	result, err := DecryptMsg(encryptedMsg, aesKey)
	if err == nil {
		return result, nil
	}

	// 如果当前密钥解密失败，尝试使用上一次的EncodingAESKey（官方要求）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Message_encryption_and_decryption.html" index="0">0</mcreference>
	if c.PrevEncodingAESKey != "" {
		prevAesKey, err := DecodeAESKey(c.PrevEncodingAESKey)
		if err != nil {
			return "", fmt.Errorf("上一次EncodingAESKey解码失败: %v", err)
		}

		result, err = DecryptMsg(encryptedMsg, prevAesKey)
		if err == nil {
			// 使用上一次密钥解密成功，说明EncodingAESKey已经更新
			// 需要将当前密钥保存为上一次密钥，并更新当前密钥为新的密钥
			if c.storage != nil {
				// 保存当前密钥为上一次密钥
				ctx := context.Background()
				err := c.storage.SavePrevEncodingAESKey(ctx, c.AppID, c.EncodingAESKey)
				if err != nil {
					// 保存失败不影响解密结果，但记录错误
					fmt.Printf("警告：保存上一次EncodingAESKey失败: %v\n", err)
				}
			}

			// 更新实例中的密钥状态
			// 将当前密钥保存为上一次密钥
			prevKey := c.EncodingAESKey
			// 更新当前密钥为新的EncodingAESKey（即上一次成功解密的密钥）
			c.EncodingAESKey = c.PrevEncodingAESKey
			// 更新上一次密钥为原来的当前密钥
			c.PrevEncodingAESKey = prevKey

			return result, nil
		}
	}

	return "", fmt.Errorf("解密失败，当前和上一次EncodingAESKey均无法解密: %v", err)
}

// VerifyURL 验证URL（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Message_encryption_and_decryption.html" index="0">0</mcreference>
func (c *WXBizMsgCrypt) VerifyURL(msgSignature, timestamp, nonce, echostr string) (string, error) {
	// 验证签名（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	if !c.VerifySignature(msgSignature, timestamp, nonce, echostr) {
		return "", fmt.Errorf("签名验证失败")
	}

	// 解密echostr（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Message_encryption_and_decryption.html" index="0">0</mcreference>
	aesKey, err := DecodeAESKey(c.EncodingAESKey)
	if err != nil {
		return "", fmt.Errorf("EncodingAESKey解码失败: %v", err)
	}

	result, err := DecryptMsg(echostr, aesKey)
	if err != nil {
		return "", fmt.Errorf("解密失败: %v", err)
	}

	return result, nil
}

// VerifySignature 验证消息签名（符合微信官方规范）
func (c *WXBizMsgCrypt) VerifySignature(msgSignature, timestamp, nonce, encryptedMsg string) bool {
	return validateSignature(c.Token, timestamp, nonce, encryptedMsg, msgSignature)
}

// GenerateSignature 生成消息签名（符合微信官方规范）
func (c *WXBizMsgCrypt) GenerateSignature(timestamp, nonce, encryptedMsg string) string {
	return genSignature(c.Token, timestamp, nonce, encryptedMsg)
}

// validateSignature 验证签名
func validateSignature(token, timestamp, nonce, encrypted, msgSignature string) bool {
	expectedSignature := genSignature(token, timestamp, nonce, encrypted)
	return expectedSignature == msgSignature
}

// genSignature 生成签名（符合微信官方规范：token、timestamp、nonce、msg_encrypt）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
func genSignature(token, timestamp, nonce, encrypted string) string {
	// 微信官方要求参数顺序：token、timestamp、nonce、msg_encrypt<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	params := []string{token, timestamp, nonce, encrypted}
	sort.Strings(params)
	str := strings.Join(params, "")
	hash := sha1.Sum([]byte(str))
	return fmt.Sprintf("%x", hash)
}