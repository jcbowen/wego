package officialaccount

import (
	"fmt"
)

// MPConfig 微信公众号配置结构体
type MPConfig struct {
	AppID     string `json:"app_id"`     // 公众号appid
	AppSecret string `json:"app_secret"` // 公众号appsecret
	Token     string `json:"token"`      // 消息校验Token
	AESKey    string `json:"aes_key"`    // 消息加解密Key
}

// StableAccessTokenMode 稳定版access_token模式
type StableAccessTokenMode string

const (
	// StableAccessTokenModeNormal 普通模式
	StableAccessTokenModeNormal StableAccessTokenMode = "normal"
	// StableAccessTokenModeForceRefresh 强制刷新模式
	StableAccessTokenModeForceRefresh StableAccessTokenMode = "force_refresh"
)

// Validate 验证配置的有效性
func (c *MPConfig) Validate() error {
	if c.AppID == "" {
		return fmt.Errorf("AppID不能为空")
	}
	if c.AppSecret == "" {
		return fmt.Errorf("AppSecret不能为空")
	}
	if c.AESKey != "" && len(c.AESKey) != 43 {
		return fmt.Errorf("AESKey必须是43位长度")
	}
	return nil
}