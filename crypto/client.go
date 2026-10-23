package crypto

import (
	"context"
	"crypto/sha1"
	"crypto/subtle"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jcbowen/wego/core"
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

// DecryptMiniProgramData 解密微信小程序加密数据（用户信息、手机号等）
//
// 参数:
//   - sessionKey:  登录接口获取的 session_key，Base64 编码
//   - encryptedData: 前端返回的 encryptedData，Base64 编码
//   - iv: 前端返回的 iv，Base64 编码
//   - result:  解密后 JSON 反序列化的目标结构体指针
//
// 返回:
//   - error: 解密或反序列化失败时返回错误
func (c *CryptoClient) DecryptMiniProgramData(sessionKey, encryptedData, iv string, result interface{}) error {
	return DecryptMiniProgramData(sessionKey, encryptedData, iv, result)
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
	Token              string
	EncodingAESKey     string
	PrevEncodingAESKey string // 上一次的EncodingAESKey（官方要求支持）
	AppID              string
	storage            storage.TokenStorage // 使用现有的存储系统
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
func NewWXBizMsgCryptWithStorage(token, encodingAESKey, appID string, storage storage.TokenStorage) (*WXBizMsgCrypt, error) {
	crypto := &WXBizMsgCrypt{
		Token:          token,
		EncodingAESKey: encodingAESKey,
		AppID:          appID,
		storage:        storage,
	}

	// 从存储中加载上一次的EncodingAESKey
	if err := crypto.loadPrevEncodingAESKey(); err != nil {
		return nil, fmt.Errorf("加载上一次EncodingAESKey失败: %w", err)
	}

	return crypto, nil
}

// SetPrevEncodingAESKey 设置上一次的EncodingAESKey（符合微信官方规范）
func (c *WXBizMsgCrypt) SetPrevEncodingAESKey(prevKey string) error {
	c.PrevEncodingAESKey = prevKey

	// 如果配置了存储系统，则保存到存储中
	if c.storage != nil {
		ctx := context.Background()
		err := c.storage.SavePrevEncodingAESKey(ctx, c.AppID, prevKey)
		if err != nil {
			return fmt.Errorf("保存上一次EncodingAESKey到存储失败: %w", err)
		}
	}

	return nil
}

// loadPrevEncodingAESKey 从存储中加载上一次的EncodingAESKey
func (c *WXBizMsgCrypt) loadPrevEncodingAESKey() error {
	if c.storage == nil {
		return nil
	}

	ctx := context.Background()
	prevKey, err := c.storage.GetPrevEncodingAESKey(ctx, c.AppID)
	if err != nil {
		return err
	}
	if prevKey != nil {
		c.PrevEncodingAESKey = prevKey.PrevEncodingAESKey
	}
	return nil
}

// EncryptMsg 加密消息（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Message_encryption_and_decryption.html" index="0">0</mcreference>
func (c *WXBizMsgCrypt) EncryptMsg(replyMsg, timestamp, nonce string) (string, string, error) {
	aesKey, err := DecodeAESKey(c.EncodingAESKey)
	if err != nil {
		return "", "", fmt.Errorf("EncodingAESKey解码失败: %w", err)
	}

	encrypted, err := EncryptMsg(replyMsg, c.AppID, aesKey)
	if err != nil {
		return "", "", fmt.Errorf("加密失败: %w", err)
	}

	// 生成签名（符合微信官方规范：token、timestamp、nonce、msg_encrypt）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	signature := genSignature(c.Token, timestamp, nonce, encrypted)
	return encrypted, signature, nil
}

// DecryptMsg 解密消息（符合微信官方规范，支持使用上一次的EncodingAESKey）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Message_encryption_and_decryption.html" index="0">0</mcreference>
//
// 兼容方法：不关心「实际解密成功所用密钥」时使用。
// 密钥轮换过渡期需要按官方要求「用哪个 AESKey 解密成功，就用此 AESKey 加密回包」时，请改用 DecryptMsgWithKey。
func (c *WXBizMsgCrypt) DecryptMsg(msgSignature, timestamp, nonce, encryptedMsg string) (string, error) {
	plainText, _, err := c.DecryptMsgWithKey(msgSignature, timestamp, nonce, encryptedMsg)
	return plainText, err
}

// DecryptMsgWithKey 解密消息并返回实际解密成功所用的 EncodingAESKey（符合微信官方规范）
//
// 首先验证消息签名；使用当前密钥解密失败时，降级尝试上一次的密钥（密钥轮换过渡期）。
//
// 参数:
//   - msgSignature: 消息签名 msg_signature
//   - timestamp: 时间戳
//   - nonce: 随机数
//   - encryptedMsg: 密文消息
//
// 返回:
//   - plainText: 解密后的明文
//   - usedAesKey: 实际解密成功所用的 EncodingAESKey（当前密钥成功时返回当前密钥，降级成功时返回上一次的密钥）
//   - error: 验签或解密失败时返回错误
func (c *WXBizMsgCrypt) DecryptMsgWithKey(msgSignature, timestamp, nonce, encryptedMsg string) (plainText string, usedAesKey string, err error) {
	// 校验时间戳是否超出容忍窗口（防重放），为空或非数字时保守跳过校验
	if err := checkTimestamp(timestamp); err != nil {
		return "", "", err
	}

	// 首先验证消息签名（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	if !c.VerifySignature(msgSignature, timestamp, nonce, encryptedMsg) {
		return "", "", fmt.Errorf("消息签名验证失败")
	}

	// 首先使用当前的EncodingAESKey尝试解密
	aesKey, err := DecodeAESKey(c.EncodingAESKey)
	if err != nil {
		return "", "", fmt.Errorf("当前EncodingAESKey解码失败: %w", err)
	}

	result, err := DecryptMsg(encryptedMsg, aesKey)
	if err == nil {
		return result, c.EncodingAESKey, nil
	}

	// 如果当前密钥解密失败，尝试使用上一次的EncodingAESKey（官方要求）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Message_encryption_and_decryption.html" index="0">0</mcreference>
	if c.PrevEncodingAESKey != "" {
		prevAesKey, err := DecodeAESKey(c.PrevEncodingAESKey)
		if err != nil {
			return "", "", fmt.Errorf("上一次EncodingAESKey解码失败: %w", err)
		}

		result, err = DecryptMsg(encryptedMsg, prevAesKey)
		if err == nil {
			// 使用上一次密钥解密成功，返回明文与该密钥。
			// 注意：不修改实例的当前/上一次密钥状态，由调用方按官方要求
			// 「用哪个 AESKey 解密成功，就用此 AESKey 加密回包」自行选择回包密钥。
			return result, c.PrevEncodingAESKey, nil
		}
	}

	return "", "", fmt.Errorf("解密失败，当前和上一次EncodingAESKey均无法解密: %w", err)
}

// EncryptMsgWithKey 用指定的 EncodingAESKey 加密回包（密钥轮换过渡期使用）
//
// 官方要求：用哪个 AESKey 解密成功，就用此 AESKey 加密回包。
// 过渡期收到用旧密钥加密的消息时，应将 DecryptMsgWithKey 返回的 usedAesKey 传入本方法回包。
//
// 参数:
//   - replyMsg: 回复消息明文
//   - timestamp: 时间戳
//   - nonce: 随机数
//   - encodingAESKey: 用于加密的 EncodingAESKey（通常为 DecryptMsgWithKey 返回的 usedAesKey）
//
// 返回:
//   - encrypted: 加密后的密文
//   - signature: 消息签名（签名使用当前配置的 Token，Token 不随密钥轮换变化）
//   - error: 密钥解码或加密失败时返回错误
func (c *WXBizMsgCrypt) EncryptMsgWithKey(replyMsg, timestamp, nonce, encodingAESKey string) (string, string, error) {
	if encodingAESKey == "" {
		return "", "", fmt.Errorf("EncodingAESKey不能为空")
	}

	aesKey, err := DecodeAESKey(encodingAESKey)
	if err != nil {
		return "", "", fmt.Errorf("EncodingAESKey解码失败: %w", err)
	}

	encrypted, err := EncryptMsg(replyMsg, c.AppID, aesKey)
	if err != nil {
		return "", "", fmt.Errorf("加密失败: %w", err)
	}

	// 生成签名（符合微信官方规范：token、timestamp、nonce、msg_encrypt）
	signature := genSignature(c.Token, timestamp, nonce, encrypted)
	return encrypted, signature, nil
}

// VerifyURL 验证URL（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Message_encryption_and_decryption.html" index="0">0</mcreference>
func (c *WXBizMsgCrypt) VerifyURL(msgSignature, timestamp, nonce, echostr string) (string, error) {
	// 校验时间戳是否超出容忍窗口（防重放），为空或非数字时保守跳过校验
	if err := checkTimestamp(timestamp); err != nil {
		return "", err
	}

	// 验证签名（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	if !c.VerifySignature(msgSignature, timestamp, nonce, echostr) {
		return "", fmt.Errorf("签名验证失败")
	}

	// 解密echostr（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Message_encryption_and_decryption.html" index="0">0</mcreference>
	aesKey, err := DecodeAESKey(c.EncodingAESKey)
	if err != nil {
		return "", fmt.Errorf("EncodingAESKey解码失败: %w", err)
	}

	result, err := DecryptMsg(echostr, aesKey)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
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
	// 常量时间比较防时序侧信道；长度不同时返回 0（视为不匹配）
	return subtle.ConstantTimeCompare([]byte(expectedSignature), []byte(msgSignature)) == 1
}

// checkTimestamp 校验微信消息时间戳是否超出容忍窗口（防重放）
//
// 容忍窗口与 core.EventTimestampTolerance（300 秒）对齐。
//
// 参数:
//   - timestamp: 微信推送的 URL 查询参数 timestamp（秒级 Unix 时间戳字符串）
//
// 返回:
//   - error: 时间戳超出容忍窗口时返回错误；timestamp 为空或非数字时保守跳过校验返回 nil
func checkTimestamp(timestamp string) error {
	if timestamp == "" {
		return nil
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil
	}
	offset := time.Now().Unix() - ts
	if offset > core.EventTimestampTolerance || offset < -core.EventTimestampTolerance {
		return fmt.Errorf("时间戳超出容忍窗口（±%d 秒），疑似重放请求", core.EventTimestampTolerance)
	}
	return nil
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
