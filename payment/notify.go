package payment

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// NotificationHeaders 微信支付回调通知请求头
//
// 字段说明:
//   - WechatpaySerial: 平台证书序列号
//   - WechatpayNonce: 随机串
//   - WechatpayTimestamp: 时间戳
//   - WechatpaySignature: 签名
//   - WechatpaySignatureType: 签名类型
//   - WechatpayRequestID: 请求 ID
type NotificationHeaders struct {
	WechatpaySerial        string
	WechatpayNonce         string
	WechatpayTimestamp     string
	WechatpaySignature     string
	WechatpaySignatureType string
	WechatpayRequestID     string
}

// ParseNotificationHeaders 从 HTTP 请求头中解析微信支付回调通知头
//
// 参数:
//   - header: HTTP 请求头
//
// 返回:
//   - *NotificationHeaders: 解析后的通知头
func ParseNotificationHeaders(header http.Header) *NotificationHeaders {
	return &NotificationHeaders{
		WechatpaySerial:        header.Get("Wechatpay-Serial"),
		WechatpayNonce:         header.Get("Wechatpay-Nonce"),
		WechatpayTimestamp:     header.Get("Wechatpay-Timestamp"),
		WechatpaySignature:     header.Get("Wechatpay-Signature"),
		WechatpaySignatureType: header.Get("Wechatpay-Signature-Type"),
		WechatpayRequestID:     header.Get("Wechatpay-Request-ID"),
	}
}

// VerifyNotification 验证微信支付回调通知签名并解密资源
//
// 参数:
//   - ctx: 上下文
//   - header: HTTP 请求头
//   - body: 请求体字节数组
//
// 返回:
//   - *Notification: 解密后的通知对象
//   - []byte: 解密后的资源 JSON 数据
//   - error: 验签或解密失败时返回错误
//
// 文档:
//   - https://pay.weixin.qq.com/wiki/doc/apiv3/wechatpay/wechatpay4_1.shtml
func (c *Client) VerifyNotification(ctx context.Context, header http.Header, body []byte) (*Notification, []byte, error) {
	headers := ParseNotificationHeaders(header)
	return c.VerifyNotificationByHeaders(ctx, headers, body)
}

// VerifyNotificationByHeaders 通过解析后的通知头验证回调通知
//
// 参数:
//   - ctx: 上下文
//   - headers: 通知头信息
//   - body: 请求体字节数组
//
// 返回:
//   - *Notification: 解密后的通知对象
//   - []byte: 解密后的资源 JSON 数据
//   - error: 验签或解密失败时返回错误
//
// 说明:
//   - 验签密钥由通知头 Wechatpay-Serial 决定：公钥 ID 走公钥，否则走平台证书
//   - 官方切换灰度期内回调由微信随机使用公钥或平台证书签名，两者均需支持
func (c *Client) VerifyNotificationByHeaders(ctx context.Context, headers *NotificationHeaders, body []byte) (*Notification, []byte, error) {
	if headers == nil {
		return nil, nil, fmt.Errorf("通知头不能为空")
	}
	if headers.WechatpaySerial == "" || headers.WechatpayNonce == "" ||
		headers.WechatpayTimestamp == "" || headers.WechatpaySignature == "" {
		return nil, nil, fmt.Errorf("通知头缺少必要字段")
	}

	// 校验时间戳，允许 5 分钟偏差
	timestamp, err := strconv.ParseInt(headers.WechatpayTimestamp, 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("时间戳格式错误: %w", err)
	}
	if diff := absInt64(time.Now().Unix() - timestamp); diff > 300 {
		return nil, nil, fmt.Errorf("通知时间戳偏差过大: %d 秒", diff)
	}

	// 构造验签名串
	message := headers.WechatpayTimestamp + "\n" + headers.WechatpayNonce + "\n" + string(body) + "\n"
	hash := sha256.Sum256([]byte(message))

	signature, err := base64.StdEncoding.DecodeString(headers.WechatpaySignature)
	if err != nil {
		return nil, nil, fmt.Errorf("签名Base64解码失败: %w", err)
	}

	pubKey, err := c.resolveVerifyPublicKey(ctx, headers.WechatpaySerial)
	if err != nil {
		return nil, nil, err
	}

	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], signature); err != nil {
		return nil, nil, fmt.Errorf("回调通知签名验证失败: %w", err)
	}

	// 解析通知 JSON
	var notification Notification
	if err := json.Unmarshal(body, &notification); err != nil {
		return nil, nil, fmt.Errorf("通知JSON解析失败: %w", err)
	}

	// 解密资源
	plainBytes, err := c.DecryptNotificationResource(&notification.Resource)
	if err != nil {
		return nil, nil, fmt.Errorf("解密通知资源失败: %w", err)
	}

	return &notification, plainBytes, nil
}

// DecryptNotificationResource 解密微信支付回调通知资源
//
// 参数:
//   - resource: 通知资源对象
//
// 返回:
//   - []byte: 解密后的 JSON 数据
//   - error: 解密失败时返回错误
func (c *Client) DecryptNotificationResource(resource *Resource) ([]byte, error) {
	if resource == nil {
		return nil, fmt.Errorf("资源对象不能为空")
	}

	plainText, err := decryptCertificate(c.config.APIV3Key, &EncryptCertificate{
		Algorithm:      resource.Algorithm,
		Nonce:          resource.Nonce,
		AssociatedData: resource.AssociatedData,
		Ciphertext:     resource.Ciphertext,
	})
	if err != nil {
		return nil, err
	}

	return []byte(plainText), nil
}

// absInt64 返回 int64 绝对值
func absInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
