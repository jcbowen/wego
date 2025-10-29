package officialaccount

import (
	"fmt"
	"time"

	"github.com/jcbowen/wego/core"
)

// StableAccessTokenRequest 获取稳定版access_token请求参数
type StableAccessTokenRequest struct {
	GrantType    string `json:"grant_type"`              // 获取access_token填写client_credential
	AppID        string `json:"appid"`                   // 公众号appid
	Secret       string `json:"secret"`                  // 公众号appsecret
	ForceRefresh bool   `json:"force_refresh,omitempty"` // 是否强制刷新，true-是；false-否
}

// StableAccessTokenResponse 获取稳定版access_token响应
type StableAccessTokenResponse struct {
	core.APIResponse
	AccessToken string `json:"access_token"` // 获取到的凭证
	ExpiresIn   int    `json:"expires_in"`   // 凭证有效时间，单位：秒
}

// StableAccessTokenInfo 稳定版access_token信息
type StableAccessTokenInfo struct {
	AccessToken string                `json:"access_token"` // 稳定版access_token
	ExpiresIn   int                   `json:"expires_in"`   // 凭证有效时间，单位：秒
	ExpiresAt   time.Time             `json:"expires_at"`   // 过期时间
	Mode        StableAccessTokenMode `json:"mode"`         // 获取模式
}

// StableAccessTokenStorage 稳定版access_token存储接口
type StableAccessTokenStorage interface {
	// SetStableAccessToken 设置稳定版access_token
	SetStableAccessToken(appID string, token *StableAccessTokenInfo) error
	// GetStableAccessToken 获取稳定版access_token
	GetStableAccessToken(appID string) (*StableAccessTokenInfo, error)
}

// DefaultStableAccessTokenStorage 默认稳定版access_token存储实现（基于内存）
type DefaultStableAccessTokenStorage struct {
	tokens map[string]*StableAccessTokenInfo
}

// NewDefaultStableAccessTokenStorage 创建默认稳定版access_token存储
func NewDefaultStableAccessTokenStorage() *DefaultStableAccessTokenStorage {
	return &DefaultStableAccessTokenStorage{
		tokens: make(map[string]*StableAccessTokenInfo),
	}
}

// SetStableAccessToken 设置稳定版access_token
func (s *DefaultStableAccessTokenStorage) SetStableAccessToken(appID string, token *StableAccessTokenInfo) error {
	s.tokens[appID] = token
	return nil
}

// ClearQuotaRequest 清空API调用次数请求参数
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/openApi/clear_quota.html
// 接口说明：清空公众号API的调用次数限制，每月共10次清零操作机会
// 注意事项：每个帐号每月共10次清零操作机会，清零生效一次即用掉一次机会
// 请求方式：POST https://api.weixin.qq.com/cgi-bin/clear_quota?access_token=ACCESS_TOKEN
// 请求体：{"appid":"APPID"}
type ClearQuotaRequest struct {
	AppID string `json:"appid"` // 公众号appid
}

// ClearQuotaResponse 清空API调用次数响应
type ClearQuotaResponse struct {
	core.APIResponse
}

// GetStableAccessToken 获取稳定版access_token
func (s *DefaultStableAccessTokenStorage) GetStableAccessToken(appID string) (*StableAccessTokenInfo, error) {
	token, exists := s.tokens[appID]
	if !exists {
		return nil, nil
	}
	return token, nil
}

// StableAccessTokenConfig 稳定版access_token配置
type StableAccessTokenConfig struct {
	AppID       string                   `json:"app_id" ini:"app_id"`             // 公众号appid
	AppSecret   string                   `json:"app_secret" ini:"app_secret"`     // 公众号appsecret
	DefaultMode StableAccessTokenMode    `json:"default_mode" ini:"default_mode"` // 默认获取模式
	Storage     StableAccessTokenStorage `json:"-"`                               // 存储接口
}

// Validate 验证稳定版access_token配置的有效性
func (c *StableAccessTokenConfig) Validate() error {
	if c.AppID == "" {
		return fmt.Errorf("AppID不能为空")
	}
	if c.AppSecret == "" {
		return fmt.Errorf("AppSecret不能为空")
	}
	if c.Storage == nil {
		c.Storage = NewDefaultStableAccessTokenStorage()
	}
	return nil
}
