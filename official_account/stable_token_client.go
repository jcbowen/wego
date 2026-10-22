package official_account

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jcbowen/wego/core"
)

// 官方对 getStableAccessToken 的限制（参考
// https://developers.weixin.qq.com/doc/service/api/base/api_getstableaccesstoken.html）：
//   - 普通模式：有效期内重复调用返回相同 token，不刷新（可放心本地缓存）
//   - 强制刷新：每天最多 20 次，且两次强制刷新需间隔 30 秒
const (
	forceRefreshDailyLimit  = 20
	forceRefreshMinInterval = 30 * time.Second
)

// StableTokenClient 稳定版access_token客户端
type StableTokenClient struct {
	client *Client

	mu            sync.RWMutex             // 保护内存缓存与限频计数
	cache         *StableAccessTokenInfo   // 进程内缓存
	stableStorage StableAccessTokenStorage // 可选的持久化存储，跨进程共享时注入

	forceDate  string    // 当前限频计数所属日期（YYYY-MM-DD，服务器时区）
	forceCount int       // 当日强制刷新次数
	lastForce  time.Time // 上次强制刷新时间
}

// NewStableTokenClient 创建稳定版access_token客户端
func NewStableTokenClient(client *Client) *StableTokenClient {
	return &StableTokenClient{
		client: client,
	}
}

// SetStableTokenStorage 注入稳定版access_token的持久化存储
// 注入后多实例/多进程部署时可共享缓存，避免重复调用接口
func (c *StableTokenClient) SetStableTokenStorage(s StableAccessTokenStorage) {
	c.stableStorage = s
}

// GetStableAccessToken 获取稳定版access_token
func (c *StableTokenClient) GetStableAccessToken(ctx context.Context, mode StableAccessTokenMode) (*StableAccessTokenInfo, error) {
	return c.getStableAccessTokenWithMode(ctx, mode)
}

// GetStableAccessTokenNormal 以普通模式获取稳定版access_token
// 普通模式优先使用本地缓存，缓存失效才请求微信
func (c *StableTokenClient) GetStableAccessTokenNormal(ctx context.Context) (*StableAccessTokenInfo, error) {
	return c.getStableAccessTokenWithMode(ctx, StableAccessTokenModeNormal)
}

// GetStableAccessTokenForceRefresh 以强制刷新模式获取稳定版access_token
// 注意：官方限制每天 20 次、间隔 30 秒，超出后本方法直接返回错误
func (c *StableTokenClient) GetStableAccessTokenForceRefresh(ctx context.Context) (*StableAccessTokenInfo, error) {
	return c.getStableAccessTokenWithMode(ctx, StableAccessTokenModeForceRefresh)
}

// getStableAccessTokenWithMode 根据模式获取稳定版access_token
func (c *StableTokenClient) getStableAccessTokenWithMode(ctx context.Context, mode StableAccessTokenMode) (*StableAccessTokenInfo, error) {
	// 普通模式优先使用缓存；强制刷新模式按定义必须请求微信
	if mode == StableAccessTokenModeNormal {
		if token, err := c.getValidStableAccessToken(ctx); err == nil && token != nil {
			return token, nil
		}
	}

	// 互斥锁保护刷新过程，避免并发重复请求
	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查：拿锁期间可能已有并发方完成刷新
	if mode == StableAccessTokenModeNormal {
		if token := validStableToken(c.cache); token != nil {
			return token, nil
		}
	}

	// 强制刷新限频检查（官方：20 次/天、间隔 30 秒）
	if mode == StableAccessTokenModeForceRefresh {
		if err := c.checkForceRefreshLimitLocked(); err != nil {
			return nil, err
		}
	}

	// 构建请求参数
	request := StableAccessTokenRequest{
		GrantType:    core.GrantTypeClientCredential,
		AppID:        c.client.config.AppID,
		Secret:       c.client.config.AppSecret,
		ForceRefresh: mode == StableAccessTokenModeForceRefresh,
	}

	// 调用API获取稳定版access_token
	var result StableAccessTokenResponse
	err := c.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLStableAccessToken,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, fmt.Errorf("获取稳定版access_token失败: %v", err)
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	// 构建token信息
	tokenInfo := &StableAccessTokenInfo{
		AccessToken: result.AccessToken,
		ExpiresIn:   result.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
		Mode:        mode,
	}

	// 写入缓存与存储
	c.cache = tokenInfo
	c.saveToStorage(tokenInfo)

	if mode == StableAccessTokenModeForceRefresh {
		c.recordForceRefreshLocked()
	}

	return tokenInfo, nil
}

// getValidStableAccessToken 获取有效的稳定版access_token（内存优先，其次持久化存储）
func (c *StableTokenClient) getValidStableAccessToken(ctx context.Context) (*StableAccessTokenInfo, error) {
	c.mu.RLock()
	token := validStableToken(c.cache)
	c.mu.RUnlock()
	if token != nil {
		return token, nil
	}

	if c.stableStorage != nil {
		stored, err := c.stableStorage.GetStableAccessToken(c.client.config.AppID)
		if err != nil {
			return nil, err
		}
		if token = validStableToken(stored); token != nil {
			// 回填进程内缓存
			c.mu.Lock()
			if c.cache == nil || validStableToken(c.cache) == nil {
				c.cache = token
			}
			c.mu.Unlock()
			return token, nil
		}
	}

	return nil, nil
}

// saveToStorage 将token写入持久化存储（失败仅告警，不影响主流程）
func (c *StableTokenClient) saveToStorage(token *StableAccessTokenInfo) {
	if c.stableStorage == nil {
		return
	}
	if err := c.stableStorage.SetStableAccessToken(c.client.config.AppID, token); err != nil {
		c.client.logger.Warn("保存稳定版access_token到存储失败", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// checkForceRefreshLimitLocked 检查强制刷新限频（调用方需已持有写锁）
func (c *StableTokenClient) checkForceRefreshLimitLocked() error {
	today := time.Now().Format("2006-01-02")
	if c.forceDate != today {
		// 跨天重置计数（时区以服务器为准）
		c.forceDate = today
		c.forceCount = 0
	}
	if !c.lastForce.IsZero() && time.Since(c.lastForce) < forceRefreshMinInterval {
		return fmt.Errorf("两次强制刷新间隔不足%v（官方限制30秒），请稍后再试", forceRefreshMinInterval)
	}
	if c.forceCount >= forceRefreshDailyLimit {
		return fmt.Errorf("今日强制刷新次数已达上限（%d次/天，官方限制）", forceRefreshDailyLimit)
	}
	return nil
}

// recordForceRefreshLocked 记录一次强制刷新（调用方需已持有写锁）
func (c *StableTokenClient) recordForceRefreshLocked() {
	c.forceCount++
	c.lastForce = time.Now()
}

// validStableToken 判断稳定版token是否仍在有效期内（考虑安全余量）
func validStableToken(token *StableAccessTokenInfo) *StableAccessTokenInfo {
	if token == nil || token.AccessToken == "" {
		return nil
	}
	if time.Now().Before(token.ExpiresAt.Add(-accessTokenExpireMargin)) {
		return token
	}
	return nil
}

// IsStableAccessTokenValid 检查稳定版access_token是否有效
// 当前实现为本地格式校验；如需强校验可调用任一仅返回 openid 的接口验证
func (c *StableTokenClient) IsStableAccessTokenValid(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, nil
	}
	return true, nil
}

// RefreshStableAccessTokenIfNeeded 如果需要则刷新稳定版access_token
func (c *StableTokenClient) RefreshStableAccessTokenIfNeeded(ctx context.Context) (string, error) {
	// 检查当前token是否即将过期（提前5分钟）
	if token, err := c.getValidStableAccessToken(ctx); err == nil && token != nil {
		return token.AccessToken, nil
	}

	// 获取新的稳定版access_token
	newToken, err := c.GetStableAccessTokenNormal(ctx)
	if err != nil {
		return "", err
	}

	return newToken.AccessToken, nil
}

// GetStableAccessTokenWithAutoRefresh 自动刷新获取稳定版access_token
func (c *StableTokenClient) GetStableAccessTokenWithAutoRefresh(ctx context.Context) (string, error) {
	return c.RefreshStableAccessTokenIfNeeded(ctx)
}

// MakeRequestWithStableToken 使用稳定版access_token发送请求
func (c *StableTokenClient) MakeRequestWithStableToken(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	// 获取稳定版access_token
	token, err := c.RefreshStableAccessTokenIfNeeded(ctx)
	if err != nil {
		return fmt.Errorf("获取稳定版access_token失败: %v", err)
	}

	// 构建带token的URL
	fullURL := fmt.Sprintf("%s?access_token=%s", url, token)

	// 发送请求
	return c.client.req.Make(ctx, &core.ReqMakeOpt{
		Method: method,
		URL:    fullURL,
		Body:   body,
		Result: result,
	})
}
