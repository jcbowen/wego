package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
)

// DecodeAESKey 解码微信开放平台AES密钥
func DecodeAESKey(encodingAESKey string) ([]byte, error) {
	if len(encodingAESKey) != 43 {
		return nil, fmt.Errorf("invalid encodingAESKey length: %d", len(encodingAESKey))
	}

	key, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		return nil, fmt.Errorf("failed to decode encodingAESKey: %v", err)
	}

	return key, nil
}

// EncryptMsg 加密微信消息
func EncryptMsg(plainText, appID string, aesKey []byte) (string, error) {
	// 生成16字节的随机字符串
	randomStr := generateRandomString(16)

	// 构造待加密的明文：randomStr + networkBytesOrder(plainText length) + plainText + appID
	text := fmt.Sprintf("%s%s%s%s",
		randomStr,
		intToNetworkBytesOrder(len(plainText)),
		plainText,
		appID)

	// PKCS7填充
	blockSize := 32
	textBytes := []byte(text)
	padLen := blockSize - len(textBytes)%blockSize
	if padLen == 0 {
		padLen = blockSize
	}
	padText := bytes.Repeat([]byte{byte(padLen)}, padLen)
	textBytes = append(textBytes, padText...)

	// AES加密
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	cipherText := make([]byte, len(textBytes))
	mode := cipher.NewCBCEncrypter(block, aesKey[:16])
	mode.CryptBlocks(cipherText, textBytes)

	// Base64编码
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// DecryptMsg 解密微信消息
func DecryptMsg(encryptedMsg string, aesKey []byte) (string, error) {
	// Base64解码
	cipherText, err := base64.StdEncoding.DecodeString(encryptedMsg)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted message: %v", err)
	}

	// AES解密
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	plainText := make([]byte, len(cipherText))
	mode := cipher.NewCBCDecrypter(block, aesKey[:16])
	mode.CryptBlocks(plainText, cipherText)

	// PKCS7去填充
	padLen := int(plainText[len(plainText)-1])
	if padLen < 1 || padLen > 32 {
		return "", errors.New("invalid padding")
	}

	plainText = plainText[:len(plainText)-padLen]

	// 解析消息结构：randomStr(16字节) + msgLen(4字节) + msg + appID
	if len(plainText) < 20 {
		return "", errors.New("plaintext too short")
	}

	msgLen := networkBytesOrderToInt(plainText[16:20])
	if len(plainText) < 20+msgLen {
		return "", errors.New("invalid message length")
	}

	return string(plainText[20 : 20+msgLen]), nil
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
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