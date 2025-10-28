package core

import (
	"fmt"
)

// WeGoConfig 微信开放平台配置结构体
type WeGoConfig struct {
	ComponentAppID     string `json:"component_app_id"`     // 第三方平台appid
	ComponentAppSecret string `json:"component_app_secret"` // 第三方平台appsecret
	ComponentToken     string `json:"component_token"`      // 消息校验Token
	EncodingAESKey     string `json:"encoding_aes_key"`     // 消息加解密Key
	RedirectURI        string `json:"redirect_uri"`         // 授权回调URI
}

// Validate 验证配置的有效性
func (c *WeGoConfig) Validate() error {
	if c.ComponentAppID == "" {
		return fmt.Errorf("ComponentAppID不能为空")
	}
	if c.ComponentAppSecret == "" {
		return fmt.Errorf("ComponentAppSecret不能为空")
	}
	if c.EncodingAESKey != "" && len(c.EncodingAESKey) != 43 {
		return fmt.Errorf("EncodingAESKey必须是43位长度")
	}
	return nil
}