package miniprogram

import "fmt"

// Config 微信小程序配置结构体
type Config struct {
	AppID     string `json:"app_id" ini:"app_id"`         // 小程序 appid
	AppSecret string `json:"app_secret" ini:"app_secret"` // 小程序 appsecret
}

// Validate 验证配置的有效性
//
// 返回:
//   - error: 配置无效时返回错误
func (c *Config) Validate() error {
	if c.AppID == "" {
		return fmt.Errorf("AppID不能为空")
	}
	if c.AppSecret == "" {
		return fmt.Errorf("AppSecret不能为空")
	}
	return nil
}
