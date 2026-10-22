package official_account

import (
	"context"
)

// SubscribeClient 订阅通知客户端（兼容层）
//
// Deprecated: 请改用 SubscribeNotifyClient 及其语义化方法（GetNotifyCategory/GetPubTemplateTitles/
// GetPubTemplateKeywords/GetPersonalTemplateList/AddPersonalTemplate/DeletePersonalTemplate/
// SendSubscribeNotify），旧方法名中的 Wxa/New 字样易与小程序订阅消息混淆。
// 本类型仅作过渡兼容：结构保留，方法全部转发到 SubscribeNotifyClient，接口路径与序列化行为完全一致。
type SubscribeClient struct {
	Client *Client
	notify *SubscribeNotifyClient
}

// NewSubscribeClient 创建新的订阅通知客户端
//
// Deprecated: 请改用 NewSubscribeNotifyClient。
//
// 参数:
//   - client *Client 公众号客户端实例
//
// 返回:
//   - *SubscribeClient 兼容层客户端（方法转发到 SubscribeNotifyClient）
func NewSubscribeClient(client *Client) *SubscribeClient {
	return &SubscribeClient{
		Client: client,
		notify: NewSubscribeNotifyClient(client),
	}
}

// GetCategory 获取订阅通知分类列表
//
// Deprecated: 请改用 SubscribeNotifyClient.GetNotifyCategory。
func (c *SubscribeClient) GetCategory(ctx context.Context) (*GetCategoryResponse, error) {
	return c.notify.GetNotifyCategory(ctx)
}

// GetPubNewTemplateTitles 获取公共模板标题列表
//
// Deprecated: 请改用 SubscribeNotifyClient.GetPubTemplateTitles。
func (c *SubscribeClient) GetPubNewTemplateTitles(ctx context.Context, categoryID int, start, limit int) (*GetPubNewTemplateTitlesResponse, error) {
	return c.notify.GetPubTemplateTitles(ctx, categoryID, start, limit)
}

// GetPubNewTemplateKeywords 获取公共模板关键词列表
//
// Deprecated: 请改用 SubscribeNotifyClient.GetPubTemplateKeywords。
func (c *SubscribeClient) GetPubNewTemplateKeywords(ctx context.Context, tid int) (*GetPubNewTemplateKeywordsResponse, error) {
	return c.notify.GetPubTemplateKeywords(ctx, tid)
}

// GetWxaPubNewTemplate 获取当前帐号下的个人模板列表
//
// Deprecated: 请改用 SubscribeNotifyClient.GetPersonalTemplateList。
func (c *SubscribeClient) GetWxaPubNewTemplate(ctx context.Context) (*GetWxaPubNewTemplateResponse, error) {
	return c.notify.GetPersonalTemplateList(ctx)
}

// AddWxaNewTemplate 添加个人模板
//
// Deprecated: 请改用 SubscribeNotifyClient.AddPersonalTemplate。
func (c *SubscribeClient) AddWxaNewTemplate(ctx context.Context, tid int, kidList []int, sceneDesc string) (*AddWxaNewTemplateResponse, error) {
	return c.notify.AddPersonalTemplate(ctx, tid, kidList, sceneDesc)
}

// DelWxaNewTemplate 删除个人模板
//
// Deprecated: 请改用 SubscribeNotifyClient.DeletePersonalTemplate。
func (c *SubscribeClient) DelWxaNewTemplate(ctx context.Context, priTmplID string) (*DelWxaNewTemplateResponse, error) {
	return c.notify.DeletePersonalTemplate(ctx, priTmplID)
}

// SendNewSubscribeMsg 发送订阅通知
//
// Deprecated: 请改用 SubscribeNotifyClient.SendSubscribeNotify。
func (c *SubscribeClient) SendNewSubscribeMsg(ctx context.Context, request *SendNewSubscribeMsgRequest) (*SendNewSubscribeMsgResponse, error) {
	return c.notify.SendSubscribeNotify(ctx, request)
}

// TemplateSubscribe 发送一次性订阅消息
//
// Deprecated: 请改用 SubscribeNotifyClient.TemplateSubscribe。
func (c *SubscribeClient) TemplateSubscribe(ctx context.Context, request *TemplateSubscribeRequest) (*TemplateSubscribeResponse, error) {
	return c.notify.TemplateSubscribe(ctx, request)
}
