package storage

import (
	"context"
	"time"
)

// ComponentAccessToken 第三方平台access_token
type ComponentAccessToken struct {
	AccessToken string    `json:"component_access_token"`
	ExpiresIn   int       `json:"expires_in"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// PreAuthCode 预授权码
type PreAuthCode struct {
	PreAuthCode string    `json:"pre_auth_code"`
	ExpiresIn   int       `json:"expires_in"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// AuthorizerAccessToken 授权方access_token
type AuthorizerAccessToken struct {
	AuthorizerAppID        string    `json:"authorizer_appid"`
	AuthorizerAccessToken  string    `json:"authorizer_access_token"`
	ExpiresIn              int       `json:"expires_in"`
	ExpiresAt              time.Time `json:"expires_at"`
	AuthorizerRefreshToken string    `json:"authorizer_refresh_token"`
}

// ComponentVerifyTicket 验证票据结构
type ComponentVerifyTicket struct {
	Ticket    string    `json:"ticket"`     // 票据内容
	CreatedAt time.Time `json:"created_at"` // 创建时间
	ExpiresAt time.Time `json:"expires_at"` // 过期时间（创建时间+12小时）
}

// PrevEncodingAESKey 上一次的EncodingAESKey结构体
type PrevEncodingAESKey struct {
	AppID              string    `json:"appid"`                 // 应用ID
	PrevEncodingAESKey string    `json:"prev_encoding_aes_key"` // 上一次的EncodingAESKey
	UpdatedAt          time.Time `json:"updated_at"`            // 更新时间
}

// TokenStorage 令牌存储接口
type TokenStorage interface {
	// 组件令牌相关方法
	SaveComponentToken(ctx context.Context, token *ComponentAccessToken) error
	GetComponentToken(ctx context.Context) (*ComponentAccessToken, error)
	DeleteComponentToken(ctx context.Context) error

	// 预授权码相关方法
	SavePreAuthCode(ctx context.Context, code *PreAuthCode) error
	GetPreAuthCode(ctx context.Context) (*PreAuthCode, error)
	DeletePreAuthCode(ctx context.Context) error

	// 验证票据相关方法
	SaveComponentVerifyTicket(ctx context.Context, ticket string) error
	GetComponentVerifyTicket(ctx context.Context) (*ComponentVerifyTicket, error) // 返回票据结构，包含创建时间
	DeleteComponentVerifyTicket(ctx context.Context) error

	// 授权方令牌相关方法
	SaveAuthorizerToken(ctx context.Context, authorizerAppID string, token *AuthorizerAccessToken) error
	GetAuthorizerToken(ctx context.Context, authorizerAppID string) (*AuthorizerAccessToken, error)
	DeleteAuthorizerToken(ctx context.Context, authorizerAppID string) error
	ClearAuthorizerTokens(ctx context.Context) error            // 清除所有授权方令牌
	ListAuthorizerTokens(ctx context.Context) ([]string, error) // 返回所有已存储的授权方appid

	// 上一次EncodingAESKey相关方法
	SavePrevEncodingAESKey(ctx context.Context, appID string, prevKey string) error
	GetPrevEncodingAESKey(ctx context.Context, appID string) (*PrevEncodingAESKey, error)
	DeletePrevEncodingAESKey(ctx context.Context, appID string) error

	// 存储健康检查
	Ping(ctx context.Context) error
}
