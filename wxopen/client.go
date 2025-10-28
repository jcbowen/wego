package wxopen

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/message"
	"github.com/jcbowen/wego/storage"
)

// WXOpenClient 微信开放平台客户端
// 提供微信开放平台第三方平台的核心功能封装
type WXOpenClient struct {
	Client  *core.WegoClient
	Storage storage.TokenStorage
}

// NewWXOpenClient 创建微信开放平台客户端
func NewWXOpenClient(config *core.WeGoConfig) *WXOpenClient {
	client := core.NewWegoClient(config)
	storage := storage.NewMemoryStorage()
	
	return &WXOpenClient{
		Client:  client,
		Storage: storage,
	}
}

// NewWXOpenClientWithStorage 创建微信开放平台客户端（使用自定义存储）
func NewWXOpenClientWithStorage(config *core.WeGoConfig, storage storage.TokenStorage) *WXOpenClient {
	client := core.NewWegoClientWithStorage(config, storage)
	
	return &WXOpenClient{
		Client:  client,
		Storage: storage,
	}
}

// HandleComponentVerifyTicket 处理component_verify_ticket事件
func (w *WXOpenClient) HandleComponentVerifyTicket(xmlData []byte) error {
	var event message.ComponentVerifyTicketEvent
	if err := xml.Unmarshal(xmlData, &event); err != nil {
		return fmt.Errorf("解析component_verify_ticket事件失败: %v", err)
	}

	// 保存component_verify_ticket到存储
	// 这里需要根据实际存储实现来保存票据
	// 通常component_verify_ticket需要持久化存储，用于后续获取component_access_token
	
	fmt.Printf("收到component_verify_ticket: %s\n", event.ComponentVerifyTicket)
	return nil
}

// HandleAuthorizedEvent 处理授权成功事件
func (w *WXOpenClient) HandleAuthorizedEvent(xmlData []byte) (*message.AuthorizedEvent, error) {
	var event message.AuthorizedEvent
	if err := xml.Unmarshal(xmlData, &event); err != nil {
		return nil, fmt.Errorf("解析授权成功事件失败: %v", err)
	}

	fmt.Printf("收到授权成功事件，授权方AppID: %s\n", event.AuthorizerAppID)
	return &event, nil
}

// HandleUnauthorizedEvent 处理取消授权事件
func (w *WXOpenClient) HandleUnauthorizedEvent(xmlData []byte) (*message.UnauthorizedEvent, error) {
	var event message.UnauthorizedEvent
	if err := xml.Unmarshal(xmlData, &event); err != nil {
		return nil, fmt.Errorf("解析取消授权事件失败: %v", err)
	}

	fmt.Printf("收到取消授权事件，授权方AppID: %s\n", event.AuthorizerAppID)
	
	// 清理该授权方的相关数据
	if err := w.Storage.DeleteAuthorizerToken(context.Background(), event.AuthorizerAppID); err != nil {
		fmt.Printf("清理授权方令牌失败: %v\n", err)
	}
	
	return &event, nil
}

// HandleUpdateAuthorizedEvent 处理授权更新事件
func (w *WXOpenClient) HandleUpdateAuthorizedEvent(xmlData []byte) (*message.UpdateAuthorizedEvent, error) {
	var event message.UpdateAuthorizedEvent
	if err := xml.Unmarshal(xmlData, &event); err != nil {
		return nil, fmt.Errorf("解析授权更新事件失败: %v", err)
	}

	fmt.Printf("收到授权更新事件，授权方AppID: %s\n", event.AuthorizerAppID)
	return &event, nil
}

// ProcessMessage 处理微信推送消息
func (w *WXOpenClient) ProcessMessage(xmlData []byte) (interface{}, error) {
	// 解析基础消息类型
	var baseMsg message.Message
	if err := xml.Unmarshal(xmlData, &baseMsg); err != nil {
		return nil, fmt.Errorf("解析XML消息失败: %v", err)
	}

	// 根据消息类型进行处理
	switch baseMsg.MsgType {
	case message.MessageTypeEvent:
		return w.processEventMessage(xmlData)
	default:
		return nil, fmt.Errorf("不支持的消息类型: %s", baseMsg.MsgType)
	}
}

// processEventMessage 处理事件消息
func (w *WXOpenClient) processEventMessage(xmlData []byte) (interface{}, error) {
	var event message.EventMessage
	if err := xml.Unmarshal(xmlData, &event); err != nil {
		return nil, fmt.Errorf("解析事件消息失败: %v", err)
	}

	switch event.Event {
	case message.EventTypeComponentVerifyTicket:
		return nil, w.HandleComponentVerifyTicket(xmlData)
	case message.EventTypeAuthorized:
		return w.HandleAuthorizedEvent(xmlData)
	case message.EventTypeUnauthorized:
		return w.HandleUnauthorizedEvent(xmlData)
	case message.EventTypeUpdateAuthorized:
		return w.HandleUpdateAuthorizedEvent(xmlData)
	default:
		return nil, fmt.Errorf("不支持的事件类型: %s", event.Event)
	}
}

// GetAuthorizedApps 获取已授权的应用列表
func (w *WXOpenClient) GetAuthorizedApps(ctx context.Context) ([]string, error) {
	return w.Storage.ListAuthorizerTokens(ctx)
}

// IsAppAuthorized 检查应用是否已授权
func (w *WXOpenClient) IsAppAuthorized(ctx context.Context, authorizerAppID string) (bool, error) {
	token, err := w.Storage.GetAuthorizerToken(ctx, authorizerAppID)
	if err != nil {
		return false, err
	}
	
	return token != nil && time.Now().Before(token.ExpiresAt), nil
}

// RefreshAuthorizerToken 刷新授权方令牌
func (w *WXOpenClient) RefreshAuthorizerToken(ctx context.Context, authorizerAppID string) error {
	token, err := w.Storage.GetAuthorizerToken(ctx, authorizerAppID)
	if err != nil {
		return err
	}
	
	if token == nil {
		return fmt.Errorf("授权方令牌不存在")
	}

	// 这里需要调用微信API刷新令牌
	// 实际实现需要调用微信的刷新令牌接口
	
	return nil
}

// ClearAllTokens 清理所有令牌数据
func (w *WXOpenClient) ClearAllTokens(ctx context.Context) error {
	if err := w.Storage.DeleteComponentToken(ctx); err != nil {
		return err
	}
	
	if err := w.Storage.DeletePreAuthCode(ctx); err != nil {
		return err
	}
	
	if err := w.Storage.ClearAuthorizerTokens(ctx); err != nil {
		return err
	}
	
	return nil
}

// HealthCheck 健康检查
func (w *WXOpenClient) HealthCheck(ctx context.Context) error {
	return w.Storage.Ping(ctx)
}

// ComponentInfo 组件信息
type ComponentInfo struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
	Token     string `json:"token"`
	AESKey    string `json:"aes_key"`
}

// GetComponentInfo 获取组件信息
func (w *WXOpenClient) GetComponentInfo() *ComponentInfo {
	config := w.Client.GetConfig()
	return &ComponentInfo{
		AppID:     config.ComponentAppID,
		AppSecret: config.ComponentAppSecret,
		Token:     config.ComponentToken,
		AESKey:    config.EncodingAESKey,
	}
}

// Statistics 统计信息
type Statistics struct {
	TotalAuthorizedApps int `json:"total_authorized_apps"`
	ActiveApps          int `json:"active_apps"`
	ExpiredApps         int `json:"expired_apps"`
}

// GetStatistics 获取统计信息
func (w *WXOpenClient) GetStatistics(ctx context.Context) (*Statistics, error) {
	appIDs, err := w.Storage.ListAuthorizerTokens(ctx)
	if err != nil {
		return nil, err
	}

	stats := &Statistics{
		TotalAuthorizedApps: len(appIDs),
	}

	for _, appID := range appIDs {
		token, err := w.Storage.GetAuthorizerToken(ctx, appID)
		if err != nil {
			continue
		}
		
		if token != nil {
			if time.Now().Before(token.ExpiresAt) {
				stats.ActiveApps++
			} else {
				stats.ExpiredApps++
			}
		}
	}

	return stats, nil
}