package official_account

import (
	"fmt"
	"time"

	"github.com/jcbowen/wego/core"
)

// OAuthAuthorizeRequest 网页授权请求参数
type OAuthAuthorizeRequest struct {
	AppID       string // 公众号的唯一标识
	RedirectURI string // 授权后重定向的回调链接地址
	Scope       string // 应用授权作用域，snsapi_base（不弹出授权页面，直接跳转，只能获取用户openid），snsapi_userinfo（弹出授权页面，可通过openid拿到昵称、性别、所在地）
	State       string // 重定向后会带上state参数，开发者可以填写a-zA-Z0-9的参数值，最多128字节
}

// OAuthAccessTokenRequest 获取网页授权access_token请求参数
type OAuthAccessTokenRequest struct {
	AppID     string // 公众号的唯一标识
	Secret    string // 公众号的appsecret
	Code      string // 填写第一步获取的code参数
	GrantType string // 填写为authorization_code
}

// OAuthAccessTokenResponse 获取网页授权access_token响应
type OAuthAccessTokenResponse struct {
	core.APIResponse
	AccessToken  string `json:"access_token"`  // 网页授权接口调用凭证,注意：此access_token与基础支持的access_token不同
	ExpiresIn    int    `json:"expires_in"`    // access_token接口调用凭证超时时间，单位（秒）
	RefreshToken string `json:"refresh_token"` // 用户刷新access_token
	OpenID       string `json:"openid"`        // 用户唯一标识，请注意，在未关注公众号时，用户访问公众号的网页，也会产生一个用户和公众号唯一的OpenID
	Scope        string `json:"scope"`         // 用户授权的作用域，使用逗号（,）分隔
	UnionID      string `json:"unionid"`       // 只有在用户将公众号绑定到微信开放平台帐号后，才会出现该字段
}

// OAuthRefreshTokenRequest 刷新网页授权access_token请求参数
type OAuthRefreshTokenRequest struct {
	AppID        string // 公众号的唯一标识
	RefreshToken string // 填写通过access_token获取到的refresh_token参数
	GrantType    string // 填写为refresh_token
}

// OAuthUserInfoRequest 获取用户信息请求参数
type OAuthUserInfoRequest struct {
	AccessToken string // 网页授权接口调用凭证
	OpenID      string // 用户的唯一标识
	Lang        string // 返回国家地区语言版本，zh_CN 简体，zh_TW 繁体，en 英语
}

// OAuthUserInfoResponse 获取用户信息响应
type OAuthUserInfoResponse struct {
	core.APIResponse
	OpenID     string   `json:"openid"`            // 用户的唯一标识
	Nickname   string   `json:"nickname"`          // 用户昵称
	Sex        int      `json:"sex"`               // 用户的性别，值为1时是男性，值为2时是女性，值为0时是未知
	Province   string   `json:"province"`          // 用户个人资料填写的省份
	City       string   `json:"city"`              // 普通用户个人资料填写的城市
	Country    string   `json:"country"`           // 国家，如中国为CN
	HeadImgURL string   `json:"headimgurl"`        // 用户头像，最后一个数值代表正方形头像大小（有0、46、64、96、132数值可选，0代表640*640正方形头像），用户没有头像时该项为空。若用户更换头像，原有头像URL将失效。
	Privilege  []string `json:"privilege"`         // 用户特权信息，json 数组，如微信沃卡用户为（chinaunicom）
	UnionID    string   `json:"unionid,omitempty"` // 只有在用户将公众号绑定到微信开放平台帐号后，才会出现该字段
}

// OAuthAuthRequest 检验授权凭证（access_token）是否有效请求参数
type OAuthAuthRequest struct {
	AccessToken string // 网页授权接口调用凭证
	OpenID      string // 用户的唯一标识
}

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
