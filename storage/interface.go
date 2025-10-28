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

	// 授权方令牌相关方法
	SaveAuthorizerToken(ctx context.Context, authorizerAppID string, token *AuthorizerAccessToken) error
	GetAuthorizerToken(ctx context.Context, authorizerAppID string) (*AuthorizerAccessToken, error)
	DeleteAuthorizerToken(ctx context.Context, authorizerAppID string) error
	ClearAuthorizerTokens(ctx context.Context) error            // 清除所有授权方令牌
	ListAuthorizerTokens(ctx context.Context) ([]string, error) // 返回所有已存储的授权方appid

	// 存储健康检查
	Ping(ctx context.Context) error
}