package payment

import (
	"context"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/jcbowen/wego/storage"
)

// getLatestPlatformCert 获取最新且未过期的平台证书记录
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - *storage.WechatPayCertificate: 最新的平台证书记录
//   - error: 下载或读取失败时返回错误
func (c *Client) getLatestPlatformCert(ctx context.Context) (*storage.WechatPayCertificate, error) {
	stored, err := c.storage.GetLatestWechatPayCert(ctx, c.config.MchID)
	if err == nil && stored != nil && certUsable(stored.ExpiresAt) {
		return stored, nil
	}

	if err := c.downloadPlatformCerts(ctx); err != nil {
		return nil, err
	}

	stored, err = c.storage.GetLatestWechatPayCert(ctx, c.config.MchID)
	if err != nil {
		return nil, fmt.Errorf("读取最新平台证书失败: %w", err)
	}
	if stored == nil {
		return nil, ErrNoEncryptKey
	}

	return stored, nil
}

// certUsable 判断证书过期时间是否仍然可用
//
// 参数:
//   - expiresAt: 证书过期时间，零值表示无过期限制
//
// 返回:
//   - bool: 未过期或零值时返回 true
func certUsable(expiresAt time.Time) bool {
	return expiresAt.IsZero() || time.Now().Before(expiresAt)
}

// resolveEncryptPublicKey 获取用于敏感信息加密的微信支付公钥
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - *rsa.PublicKey: 加密用公钥
//   - string: 该公钥对应的 Wechatpay-Serial 值（公钥 ID 或平台证书序列号）
//   - error: 无可用密钥时返回错误
//
// 说明:
//   - 与验签不同，加密是商户主动行为，故按配置优先：配了微信支付公钥就优先使用
//   - 未配置公钥时回退到最新平台证书
func (c *Client) resolveEncryptPublicKey(ctx context.Context) (*rsa.PublicKey, string, error) {
	if c.publicKey != nil {
		return c.publicKey, c.config.PublicKeyID, nil
	}

	stored, err := c.getLatestPlatformCert(ctx)
	if err != nil {
		return nil, "", err
	}

	cert, err := parseCertificate(stored.Certificate)
	if err != nil {
		return nil, "", fmt.Errorf("解析平台证书失败: %w", err)
	}

	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, "", fmt.Errorf("平台证书公钥类型错误")
	}

	return pubKey, stored.SerialNo, nil
}

// EncryptSerialNo 返回当前用于敏感信息加密的 Wechatpay-Serial 值
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - string: 微信支付公钥 ID 或平台证书序列号
//   - error: 无可用密钥时返回错误
//
// 使用示例:
//   - serial, err := client.EncryptSerialNo(ctx)
func (c *Client) EncryptSerialNo(ctx context.Context) (string, error) {
	_, serial, err := c.resolveEncryptPublicKey(ctx)
	return serial, err
}

// EncryptSensitiveData 加密微信支付要求的敏感信息
//
// 参数:
//   - ctx: 上下文
//   - plaintext: 明文，如分账接收方姓名、身份证号等
//
// 返回:
//   - string: Base64 编码的密文，可直接作为接口字段值
//   - error: 无可用密钥或加密失败时返回错误
//
// 说明:
//   - 使用 RSAES-OAEP 算法（SHA-1 + MGF1-SHA1），与微信支付官方要求一致
//   - 优先使用配置的微信支付公钥，未配置时使用最新平台证书
//
// 使用示例:
//   - name, err := client.EncryptSensitiveData(ctx, "张三")
func (c *Client) EncryptSensitiveData(ctx context.Context, plaintext string) (string, error) {
	pubKey, _, err := c.resolveEncryptPublicKey(ctx)
	if err != nil {
		return "", err
	}
	return EncryptSensitiveWithPublicKey(pubKey, plaintext)
}

// EncryptSensitiveWithPublicKey 使用指定公钥加密敏感信息
//
// 参数:
//   - pubKey: 微信支付公钥或平台证书公钥
//   - plaintext: 明文
//
// 返回:
//   - string: Base64 编码的密文
//   - error: 明文为空或加密失败时返回错误
//
// 使用示例:
//   - cipher, err := payment.EncryptSensitiveWithPublicKey(pubKey, "张三")
func EncryptSensitiveWithPublicKey(pubKey *rsa.PublicKey, plaintext string) (string, error) {
	if pubKey == nil {
		return "", fmt.Errorf("公钥不能为空")
	}
	if plaintext == "" {
		return "", fmt.Errorf("待加密内容不能为空")
	}

	cipherText, err := rsa.EncryptOAEP(sha1.New(), crand.Reader, pubKey, []byte(plaintext), nil)
	if err != nil {
		return "", fmt.Errorf("敏感信息加密失败: %w", err)
	}

	return base64.StdEncoding.EncodeToString(cipherText), nil
}
