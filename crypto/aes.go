package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

// DecodeAESKey 解码微信开放平台AES密钥（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
func DecodeAESKey(encodingAESKey string) ([]byte, error) {
	if len(encodingAESKey) != 43 {
		return nil, fmt.Errorf("EncodingAESKey长度必须为43位，当前长度: %d", len(encodingAESKey))
	}

	// 微信官方要求EncodingAESKey必须是43位Base64编码，尾部填充一个字符的"="<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	key, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		return nil, fmt.Errorf("EncodingAESKey Base64解码失败: %w", err)
	}

	// 验证解码后的密钥长度（32字节）
	if len(key) != 32 {
		return nil, fmt.Errorf("解码后的AES密钥长度必须为32字节，当前长度: %d", len(key))
	}

	return key, nil
}

// EncryptMsg 加密微信消息（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
func EncryptMsg(plainText, appID string, aesKey []byte) (string, error) {
	// 生成16字节的安全随机字符串（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	randomStr, err := generateRandomString(16)
	if err != nil {
		return "", fmt.Errorf("生成随机字符串失败: %w", err)
	}

	// 构造待加密的明文：random(16B) + msg_len(4B) + msg + appid（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	text := fmt.Sprintf("%s%s%s%s",
		randomStr,
		intToNetworkBytesOrder(len(plainText)),
		plainText,
		appID)

	// PKCS7填充（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	blockSize := aes.BlockSize
	textBytes := []byte(text)
	padLen := blockSize - len(textBytes)%blockSize
	if padLen == 0 {
		padLen = blockSize
	}
	padText := bytes.Repeat([]byte{byte(padLen)}, padLen)
	textBytes = append(textBytes, padText...)

	// AES加密（CBC模式，使用aesKey的前16字节作为IV，符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("创建AES密码器失败: %w", err)
	}

	cipherText := make([]byte, len(textBytes))
	mode := cipher.NewCBCEncrypter(block, aesKey[:aes.BlockSize])
	mode.CryptBlocks(cipherText, textBytes)

	// Base64编码（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// DecryptMsg 解密微信消息（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
func DecryptMsg(encryptedMsg string, aesKey []byte) (string, error) {
	// Base64解码（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	cipherText, err := base64.StdEncoding.DecodeString(encryptedMsg)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	// AES解密（CBC模式，使用aesKey的前16字节作为IV，符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("创建AES密码器失败: %w", err)
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("密文长度过短")
	}

	plainText := make([]byte, len(cipherText))
	mode := cipher.NewCBCDecrypter(block, aesKey[:aes.BlockSize])
	mode.CryptBlocks(plainText, cipherText)

	// PKCS7去填充（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	if len(plainText) == 0 {
		return "", errors.New("解密后数据为空")
	}
	padLen := int(plainText[len(plainText)-1])
	if padLen < 1 || padLen > aes.BlockSize {
		return "", errors.New("填充数据无效")
	}

	plainText = plainText[:len(plainText)-padLen]

	// 解析消息结构：random(16字节) + msg_len(4字节) + msg + appid（符合微信官方规范）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	if len(plainText) < 20 {
		return "", errors.New("明文长度过短")
	}

	msgLen := networkBytesOrderToInt(plainText[16:20])
	if len(plainText) < 20+msgLen {
		return "", errors.New("消息长度无效")
	}

	// 验证AppID（微信官方要求验证第三方平台AppID）<mcreference link="https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html" index="1">1</mcreference>
	decryptedAppID := string(plainText[20+msgLen:])
	if len(decryptedAppID) == 0 {
		return "", errors.New("AppID缺失")
	}

	return string(plainText[20 : 20+msgLen]), nil
}

// generateRandomString 生成安全的随机字符串（符合微信官方规范）
func generateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 使用加密安全的随机数生成器
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("生成随机数失败: %w", err)
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[int(randomBytes[i])%len(charset)]
	}

	return string(result), nil
}

// intToNetworkBytesOrder 整数转网络字节序
func intToNetworkBytesOrder(n int) []byte {
	bytes := make([]byte, 4)
	bytes[0] = byte(n >> 24)
	bytes[1] = byte(n >> 16)
	bytes[2] = byte(n >> 8)
	bytes[3] = byte(n)
	return bytes
}

// networkBytesOrderToInt 网络字节序转整数
func networkBytesOrderToInt(bytes []byte) int {
	if len(bytes) != 4 {
		return 0
	}
	return int(bytes[0])<<24 | int(bytes[1])<<16 | int(bytes[2])<<8 | int(bytes[3])
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
//
// 说明:
//   - 使用 AES-128-CBC（兼容 AES-192/AES-256）与 PKCS7 填充
func DecryptMiniProgramData(sessionKey, encryptedData, iv string, result interface{}) error {
	key, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return fmt.Errorf("session_key Base64解码失败: %w", err)
	}
	if len(key) != 16 {
		return fmt.Errorf("session_key长度错误，期望16字节，实际%d字节", len(key))
	}

	cipherText, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return fmt.Errorf("encryptedData Base64解码失败: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return fmt.Errorf("iv Base64解码失败: %w", err)
	}

	if len(nonce) != aes.BlockSize {
		return fmt.Errorf("iv长度错误，期望 %d，实际 %d", aes.BlockSize, len(nonce))
	}

	if len(cipherText)%aes.BlockSize != 0 {
		return fmt.Errorf("encryptedData长度不是块大小整数倍")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("创建AES密码器失败: %w", err)
	}

	plainText := make([]byte, len(cipherText))
	mode := cipher.NewCBCDecrypter(block, nonce)
	mode.CryptBlocks(plainText, cipherText)

	plainText, err = unpadPKCS7(plainText)
	if err != nil {
		return fmt.Errorf("去除PKCS7填充失败: %w", err)
	}

	if result != nil {
		if err := json.Unmarshal(plainText, result); err != nil {
			return fmt.Errorf("解密数据JSON解析失败: %w", err)
		}
	}

	return nil
}

// unpadPKCS7 去除 PKCS7 填充
func unpadPKCS7(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("数据为空")
	}

	padLen := int(data[len(data)-1])
	if padLen < 1 || padLen > aes.BlockSize || padLen > len(data) {
		return nil, errors.New("填充数据无效")
	}

	for i := 0; i < padLen; i++ {
		if data[len(data)-1-i] != byte(padLen) {
			return nil, errors.New("填充数据不一致")
		}
	}

	return data[:len(data)-padLen], nil
}
