package payment

import "fmt"

// Config 微信支付 V3 配置结构体
//
// 验签模式说明（平台证书与微信支付公钥可任选其一，也可同时配置）:
//   - 仅使用平台证书：不填 PublicKeyID 与 PublicKey，维持传统模式
//   - 仅使用微信支付公钥：填写 PublicKeyID 与 PublicKey，不再依赖平台证书下载
//   - 两者都配置：优先使用微信支付公钥，但应答/回调仍以响应头 Wechatpay-Serial 为准自动回退到平台证书
type Config struct {
	AppID        string `json:"app_id" ini:"app_id"`                 // 应用 ID（公众号/小程序 appid）
	MchID        string `json:"mch_id" ini:"mch_id"`                 // 商户号
	APIV3Key     string `json:"api_v3_key" ini:"api_v3_key"`         // API V3 密钥（32 字节）
	CertSerialNo string `json:"cert_serial_no" ini:"cert_serial_no"` // 商户 API 证书序列号，用于请求签名
	PrivateKey   string `json:"private_key" ini:"private_key"`       // 商户私钥 PEM 字符串
	NotifyURL    string `json:"notify_url" ini:"notify_url"`         // 支付结果通知 URL
	PublicKeyID  string `json:"public_key_id" ini:"public_key_id"`   // 微信支付公钥 ID，形如 PUB_KEY_ID_xxx（官方推荐）
	PublicKey    string `json:"public_key" ini:"public_key"`         // 微信支付公钥 PEM 字符串（官方推荐）
}

// UsePublicKey 判断当前配置是否启用微信支付公钥验签
//
// 返回:
//   - bool: PublicKeyID 与 PublicKey 均已配置时返回 true
//
// 使用示例:
//   - if cfg.UsePublicKey() { fmt.Println("公钥模式") }
func (c *Config) UsePublicKey() bool {
	return c.PublicKeyID != "" && c.PublicKey != ""
}

// Validate 验证配置的有效性
//
// 返回:
//   - error: 配置无效时返回错误
func (c *Config) Validate() error {
	if c.AppID == "" {
		return fmt.Errorf("AppID不能为空")
	}
	if c.MchID == "" {
		return fmt.Errorf("MchID不能为空")
	}
	if c.APIV3Key == "" {
		return fmt.Errorf("APIV3Key不能为空")
	}
	if len(c.APIV3Key) != 32 {
		return fmt.Errorf("APIV3Key长度必须为32字节，当前长度: %d", len(c.APIV3Key))
	}
	if c.CertSerialNo == "" {
		return fmt.Errorf("CertSerialNo不能为空")
	}
	if c.PrivateKey == "" {
		return fmt.Errorf("PrivateKey不能为空")
	}
	// 公钥 ID 与公钥内容必须成对配置，避免只配一半导致验签失败
	if c.PublicKeyID != "" && c.PublicKey == "" {
		return fmt.Errorf("已配置PublicKeyID时必须同时配置PublicKey")
	}
	if c.PublicKey != "" && c.PublicKeyID == "" {
		return fmt.Errorf("已配置PublicKey时必须同时配置PublicKeyID")
	}
	return nil
}
