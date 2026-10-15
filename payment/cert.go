package payment

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/jcbowen/wego/storage"
)

// resolveVerifyPublicKey 根据响应头 Wechatpay-Serial 选择验签公钥
//
// 参数:
//   - ctx: 上下文
//   - serialNo: 响应头 Wechatpay-Serial 的值
//
// 返回:
//   - *rsa.PublicKey: 用于验签的公钥
//   - error: 未配置对应密钥或获取失败时返回错误
//
// 说明:
//   - 以 PUB_KEY_ID_ 开头表示微信使用「微信支付公钥」签名，需使用配置的公钥验签
//   - 其余为平台证书序列号，使用平台证书验签（未缓存时自动下载）
//   - 官方灰度期内回调场景由微信随机决定使用哪种密钥，故必须以响应头为准
//
// 使用示例:
//   - pubKey, err := client.resolveVerifyPublicKey(ctx, resp.Header.Get("Wechatpay-Serial"))
func (c *Client) resolveVerifyPublicKey(ctx context.Context, serialNo string) (*rsa.PublicKey, error) {
	if serialNo == "" {
		return nil, fmt.Errorf("Wechatpay-Serial不能为空")
	}

	if strings.HasPrefix(serialNo, PublicKeyIDPrefix) {
		if c.publicKey == nil {
			return nil, fmt.Errorf("响应使用微信支付公钥(%s)签名，但当前未配置 PublicKeyID/PublicKey", serialNo)
		}
		return c.publicKey, nil
	}

	cert, err := c.getPlatformCert(ctx, serialNo)
	if err != nil {
		return nil, fmt.Errorf("获取平台证书失败: %w", err)
	}

	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("平台证书公钥类型错误")
	}

	return pubKey, nil
}

// decryptCertificate 使用 API V3 密钥解密平台证书
//
// 参数:
//   - apiV3Key: API V3 密钥
//   - cert: 加密的证书信息
//
// 返回:
//   - string: 解密后的 PEM 证书
//   - error: 解密失败时返回错误
func decryptCertificate(apiV3Key string, cert *EncryptCertificate) (string, error) {
	if cert.Algorithm != "AEAD_AES_256_GCM" {
		return "", fmt.Errorf("不支持的加密算法: %s", cert.Algorithm)
	}

	key := []byte(apiV3Key)
	if len(key) != 32 {
		return "", fmt.Errorf("APIV3Key长度必须为32字节")
	}

	cipherText, err := base64.StdEncoding.DecodeString(cert.Ciphertext)
	if err != nil {
		return "", fmt.Errorf("证书密文Base64解码失败: %w", err)
	}

	nonce := []byte(cert.Nonce)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建AES密码器失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM失败: %w", err)
	}

	plainText, err := gcm.Open(nil, nonce, cipherText, []byte(cert.AssociatedData))
	if err != nil {
		return "", fmt.Errorf("解密证书失败: %w", err)
	}

	return string(plainText), nil
}

// parseCertificate 解析 PEM 证书
//
// 参数:
//   - certPEM: PEM 格式证书
//
// 返回:
//   - *x509.Certificate: 解析后的证书
//   - error: 解析失败时返回错误
func parseCertificate(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("无法解析证书PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	return cert, nil
}

// certRefreshMinInterval 平台证书下载失败后的重试最小间隔
//
// 说明:
//   - 微信对 /v3/certificates 有频率限制，失败后短时间内不应反复重试
const certRefreshMinInterval = time.Minute

// downloadPlatformCerts 从微信服务器下载平台证书并缓存
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 下载或保存失败时返回错误
//
// 说明:
//   - 该接口不做应答验签：验签依赖平台证书本身，属于"先有鸡后有蛋"，微信官方 SDK 同样处理
//   - 并发调用会被串行化，失败后在 certRefreshMinInterval 内直接返回上次错误，避免被限流
//   - 仅使用微信支付公钥模式的商户不会调用该接口
func (c *Client) downloadPlatformCerts(ctx context.Context) error {
	c.certRefreshMu.Lock()
	defer c.certRefreshMu.Unlock()

	if !c.certRefreshFail.IsZero() && time.Since(c.certRefreshFail) < certRefreshMinInterval {
		return fmt.Errorf("平台证书下载最近已失败，%v 内不再重试: %w", certRefreshMinInterval, ErrCertDownloadFailed)
	}

	if err := c.fetchPlatformCerts(ctx); err != nil {
		c.certRefreshFail = time.Now()
		return err
	}

	c.certRefreshFail = time.Time{}
	return nil
}

// fetchPlatformCerts 实际执行平台证书下载与缓存
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 下载或保存失败时返回错误
func (c *Client) fetchPlatformCerts(ctx context.Context) error {
	var result CertificatesResponse
	if err := c.doRequest(ctx, "GET", PayCertificatesURL, nil, &result, false); err != nil {
		return fmt.Errorf("下载平台证书失败: %w", err)
	}

	if len(result.Data) == 0 {
		return fmt.Errorf("%w: 接口未返回任何平台证书（若商户已切换为微信支付公钥模式，请配置 PublicKeyID/PublicKey）", ErrCertDownloadFailed)
	}

	loaded := 0
	for _, item := range result.Data {
		certPEM, err := decryptCertificate(c.config.APIV3Key, &item.EncryptCertificate)
		if err != nil {
			c.logger.Error("解密平台证书失败", map[string]interface{}{
				"serial_no": item.SerialNo,
				"error":     err.Error(),
			})
			continue
		}

		cert, err := parseCertificate(certPEM)
		if err != nil {
			c.logger.Error("解析平台证书失败", map[string]interface{}{
				"serial_no": item.SerialNo,
				"error":     err.Error(),
			})
			continue
		}

		expiresAt := cert.NotAfter
		if !expiresAt.IsZero() {
			expiresAt = expiresAt.Add(-time.Minute)
		}

		storedCert := &storage.WechatPayCertificate{
			MchID:       c.config.MchID,
			SerialNo:    item.SerialNo,
			Certificate: certPEM,
			ExpiresAt:   expiresAt,
		}
		if err := c.storage.SaveWechatPayCert(ctx, c.config.MchID, storedCert); err != nil {
			c.logger.Error("保存平台证书到存储失败", map[string]interface{}{
				"serial_no": item.SerialNo,
				"error":     err.Error(),
			})
			continue
		}

		func() {
			c.certMu.Lock()
			defer c.certMu.Unlock()
			c.certs[item.SerialNo] = cert
		}()
		loaded++
	}

	if loaded == 0 {
		return fmt.Errorf("%w: 全部 %d 张证书解密或保存失败（请检查 APIV3Key 是否为 32 字节）", ErrCertDownloadFailed, len(result.Data))
	}

	return nil
}

// getPlatformCert 获取指定序列号的微信支付平台证书
//
// 参数:
//   - ctx: 上下文
//   - serialNo: 证书序列号
//
// 返回:
//   - *x509.Certificate: 平台证书
//   - error: 获取失败时返回错误
func (c *Client) getPlatformCert(ctx context.Context, serialNo string) (*x509.Certificate, error) {
	if serialNo == "" {
		return nil, fmt.Errorf("serialNo不能为空")
	}

	// 检查内存缓存
	c.certMu.RLock()
	if cert, ok := c.certs[serialNo]; ok && time.Now().Before(cert.NotAfter) {
		c.certMu.RUnlock()
		return cert, nil
	}
	c.certMu.RUnlock()

	// 从存储中读取
	//
	// 存储读取失败（含自定义存储把"不存在"实现为 error 的情况）不应阻断流程，
	// 直接回退到重新下载，否则首次使用永远拿不到证书
	storedCert, err := c.storage.GetWechatPayCert(ctx, c.config.MchID, serialNo)
	if err != nil {
		c.logger.Warn("从存储读取平台证书失败，将重新下载", map[string]interface{}{
			"serial_no": serialNo,
			"error":     err.Error(),
		})
	}

	if storedCert != nil {
		cert, err := parseCertificate(storedCert.Certificate)
		if err != nil {
			return nil, fmt.Errorf("解析存储的平台证书失败: %w", err)
		}
		if time.Now().Before(cert.NotAfter) {
			c.certMu.Lock()
			c.certs[serialNo] = cert
			c.certMu.Unlock()
			return cert, nil
		}
	}

	// 重新下载平台证书
	if err := c.downloadPlatformCerts(ctx); err != nil {
		return nil, err
	}

	c.certMu.RLock()
	defer c.certMu.RUnlock()
	if cert, ok := c.certs[serialNo]; ok {
		return cert, nil
	}

	return nil, fmt.Errorf("无法获取序列号为 %s 的平台证书（若商户已切换为微信支付公钥模式，请配置 PublicKeyID/PublicKey）", serialNo)
}
