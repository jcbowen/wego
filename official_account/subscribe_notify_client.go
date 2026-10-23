package official_account

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jcbowen/wego/core"
)

// SubscribeNotifyClient 公众号订阅通知客户端
//
// 职责边界（与 miniprogram 包的订阅消息互不重叠）：
//   - 本客户端封装「公众号订阅通知」全套接口（/wxaapi/newtmpl/* 公众号版 + /cgi-bin/message/subscribe/bizsend 发送），
//     使用公众号 access_token；
//   - 小程序订阅消息（/cgi-bin/message/subscribe/send）请使用 miniprogram 包的 SendSubscribeMessage。
type SubscribeNotifyClient struct {
	Client *Client
}

// NewSubscribeNotifyClient 创建新的公众号订阅通知客户端
//
// 参数:
//   - client *Client 公众号客户端实例，不能为空
//
// 返回:
//   - *SubscribeNotifyClient 订阅通知客户端
//
// 使用示例:
//
//	notifyClient, _ := wegoClient.OfficialAccountSubscribeNotifyE()
//	notifyClient.SendSubscribeNotify(ctx, &official_account.SendNewSubscribeMsgRequest{...})
func NewSubscribeNotifyClient(client *Client) *SubscribeNotifyClient {
	return &SubscribeNotifyClient{
		Client: client,
	}
}

// CategoryInfo 分类信息
type CategoryInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GetCategoryResponse 获取分类列表响应
type GetCategoryResponse struct {
	core.APIResponse
	Data []CategoryInfo `json:"data"`
}

// PubTemplateTitleInfo 公共模板标题信息
//
// 说明: 官方 tid 为 string 类型
type PubTemplateTitleInfo struct {
	TID        string `json:"tid"`
	Title      string `json:"title"`
	Type       int    `json:"type"`
	CategoryID int    `json:"categoryId"`
}

// GetPubNewTemplateTitlesResponse 获取公共模板标题列表响应
type GetPubNewTemplateTitlesResponse struct {
	core.APIResponse
	Count int                    `json:"count"`
	Data  []PubTemplateTitleInfo `json:"data"`
}

// PubTemplateKeywordInfo 公共模板关键词信息
type PubTemplateKeywordInfo struct {
	Kid     int    `json:"kid"`
	Name    string `json:"name"`
	Example string `json:"example"`
	Rule    string `json:"rule"`
}

// GetPubNewTemplateKeywordsResponse 获取公共模板关键词列表响应
type GetPubNewTemplateKeywordsResponse struct {
	core.APIResponse
	Count int                      `json:"count"` // 关键词总数（官方响应含 count）
	Data  []PubTemplateKeywordInfo `json:"data"`
}

// WxaPubNewTemplateInfo 个人模板信息
type WxaPubNewTemplateInfo struct {
	PriTmplID string `json:"priTmplId"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Example   string `json:"example"`
	Type      int    `json:"type"`
}

// GetWxaPubNewTemplateResponse 获取个人模板列表响应
type GetWxaPubNewTemplateResponse struct {
	core.APIResponse
	Data []WxaPubNewTemplateInfo `json:"data"`
}

// AddWxaNewTemplateRequest 添加个人模板请求
//
// 说明: 官方 tid 为 string 类型
type AddWxaNewTemplateRequest struct {
	TID       string `json:"tid"`
	KidList   []int  `json:"kidList"`
	SceneDesc string `json:"sceneDesc"`
}

// AddWxaNewTemplateResponse 添加个人模板响应
type AddWxaNewTemplateResponse struct {
	core.APIResponse
	PriTmplID string `json:"priTmplId"`
}

// DelWxaNewTemplateRequest 删除个人模板请求
type DelWxaNewTemplateRequest struct {
	PriTmplID string `json:"priTmplId"`
}

// DelWxaNewTemplateResponse 删除个人模板响应
type DelWxaNewTemplateResponse struct {
	core.APIResponse
}

// SubscribeMsgData 订阅通知数据
//
// 说明: 官方 bizsend 的 data 各项仅含 value 字段，不含 name
type SubscribeMsgData struct {
	Value string `json:"value"`
}

// SendNewSubscribeMsgRequest 发送订阅通知请求（/cgi-bin/message/subscribe/bizsend）
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Subscription_Messages/Message_Send_API.html
type SendNewSubscribeMsgRequest struct {
	Touser     string                      `json:"touser"`
	TemplateID string                      `json:"template_id"`
	Page       string                      `json:"page,omitempty"`
	Data       map[string]SubscribeMsgData `json:"data"`
	// MiniprogramState 小程序状态，官方为 string：
	// developer-开发版 / trial-体验版 / formal-正式版（默认）
	MiniprogramState string `json:"miniprogram_state,omitempty"`
	// Lang 进入小程序查看的语言：zh_CN / zh_TW / zh_HK / en
	Lang string `json:"lang,omitempty"`
	// Miniprogram 跳转小程序配置（可选）
	Miniprogram *TemplateMessageMiniProgram `json:"miniprogram,omitempty"`
}

// SendNewSubscribeMsgResponse 发送订阅通知响应
type SendNewSubscribeMsgResponse struct {
	core.APIResponse
}

// GetNotifyCategory 获取订阅通知类目列表
// 接口文档：https://developers.weixin.qq.com/doc/offiaccount/Subscription_Messages/Category_list_API.html
// 请求方式：GET /wxaapi/newtmpl/getcategory
//
// 参数:
//   - ctx context.Context 请求上下文
//
// 返回:
//   - *GetCategoryResponse 类目列表
//   - error 请求失败或微信返回业务错误时返回错误
func (c *SubscribeNotifyClient) GetNotifyCategory(ctx context.Context) (*GetCategoryResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetCategoryResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetCategory, accessToken)
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    apiURL,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetPubTemplateTitles 获取公共模板标题列表
// 接口文档：https://developers.weixin.qq.com/doc/offiaccount/Subscription_Messages/Template_API.html
// 请求方式：GET /wxaapi/newtmpl/getpubtemplatetitles
//
// 参数:
//   - ctx context.Context 请求上下文
//   - categoryID int 类目 ID（GetNotifyCategory 返回的 id）
//   - start int 起始位置（从 0 开始）
//   - limit int 单页数量（最大 30）
//
// 返回:
//   - *GetPubNewTemplateTitlesResponse 公共模板标题列表
//   - error 请求失败或微信返回业务错误时返回错误
func (c *SubscribeNotifyClient) GetPubTemplateTitles(ctx context.Context, categoryID int, start, limit int) (*GetPubNewTemplateTitlesResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetPubNewTemplateTitlesResponse
	apiURL := fmt.Sprintf("%s?access_token=%s&ids=%d&start=%d&limit=%d",
		URLGetPubTemplateTitles, accessToken, categoryID, start, limit)
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    apiURL,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetPubTemplateKeywords 获取公共模板关键词列表
// 接口文档：https://developers.weixin.qq.com/doc/offiaccount/Subscription_Messages/Template_API.html
// 请求方式：GET /wxaapi/newtmpl/getpubtemplatekeywords
//
// 参数:
//   - ctx context.Context 请求上下文
//   - tid string 公共模板标题 ID（GetPubTemplateTitles 返回的 tid，官方为 string）
//
// 返回:
//   - *GetPubNewTemplateKeywordsResponse 关键词列表
//   - error 请求失败或微信返回业务错误时返回错误
func (c *SubscribeNotifyClient) GetPubTemplateKeywords(ctx context.Context, tid string) (*GetPubNewTemplateKeywordsResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetPubNewTemplateKeywordsResponse
	apiURL := fmt.Sprintf("%s?access_token=%s&tid=%s",
		URLGetPubTemplateKeywords, accessToken, url.QueryEscape(tid))
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    apiURL,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetPersonalTemplateList 获取当前帐号下的个人模板列表
// 接口文档：https://developers.weixin.qq.com/doc/offiaccount/Subscription_Messages/Template_API.html
// 请求方式：GET /wxaapi/newtmpl/gettemplate
//
// 参数:
//   - ctx context.Context 请求上下文
//
// 返回:
//   - *GetWxaPubNewTemplateResponse 个人模板列表
//   - error 请求失败或微信返回业务错误时返回错误
func (c *SubscribeNotifyClient) GetPersonalTemplateList(ctx context.Context) (*GetWxaPubNewTemplateResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetWxaPubNewTemplateResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetWxaPubTemplate, accessToken)
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    apiURL,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// AddPersonalTemplate 组合模板并添加至帐号下的个人模板库
// 接口文档：https://developers.weixin.qq.com/doc/offiaccount/Subscription_Messages/Template_API.html
// 请求方式：POST /wxaapi/newtmpl/addtemplate
//
// 参数:
//   - ctx context.Context 请求上下文
//   - tid string 公共模板标题 ID（官方为 string）
//   - kidList []int 关键词 ID 列表（与模板关键词顺序对应）
//   - sceneDesc string 服务场景描述（15 字以内）
//
// 返回:
//   - *AddWxaNewTemplateResponse 含新增的个人模板 ID（PriTmplID）
//   - error 请求失败或微信返回业务错误时返回错误
func (c *SubscribeNotifyClient) AddPersonalTemplate(ctx context.Context, tid string, kidList []int, sceneDesc string) (*AddWxaNewTemplateResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := AddWxaNewTemplateRequest{
		TID:       tid,
		KidList:   kidList,
		SceneDesc: sceneDesc,
	}

	var result AddWxaNewTemplateResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLAddWxaNewTemplate, accessToken)
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    apiURL,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// DeletePersonalTemplate 删除帐号下的个人模板
// 接口文档：https://developers.weixin.qq.com/doc/offiaccount/Subscription_Messages/Template_API.html
// 请求方式：POST /wxaapi/newtmpl/deltemplate
//
// 参数:
//   - ctx context.Context 请求上下文
//   - priTmplID string 个人模板 ID
//
// 返回:
//   - *DelWxaNewTemplateResponse 删除结果
//   - error 请求失败或微信返回业务错误时返回错误
func (c *SubscribeNotifyClient) DeletePersonalTemplate(ctx context.Context, priTmplID string) (*DelWxaNewTemplateResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := DelWxaNewTemplateRequest{
		PriTmplID: priTmplID,
	}

	var result DelWxaNewTemplateResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLDelWxaNewTemplate, accessToken)
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    apiURL,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// SendSubscribeNotify 发送订阅通知（bizsend）
// 接口文档：https://developers.weixin.qq.com/doc/offiaccount/Subscription_Messages/Message_Send_API.html
// 请求方式：POST /cgi-bin/message/subscribe/bizsend
//
// 参数:
//   - ctx context.Context 请求上下文
//   - request *SendNewSubscribeMsgRequest 发送请求（Touser/TemplateID/Data 必填）
//
// 返回:
//   - *SendNewSubscribeMsgResponse 发送结果
//   - error 请求失败或微信返回业务错误时返回错误
func (c *SubscribeNotifyClient) SendSubscribeNotify(ctx context.Context, request *SendNewSubscribeMsgRequest) (*SendNewSubscribeMsgResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result SendNewSubscribeMsgResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLSendNewSubscribeMsg, accessToken)
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    apiURL,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// TemplateSubscribeContent 一次性订阅消息正文
type TemplateSubscribeContent struct {
	Value string `json:"value"`
	Color string `json:"color,omitempty"`
}

// TemplateSubscribeData 一次性订阅消息 data 体
// 官方固定结构：data.content.value/color（content 为固定键名，非动态字段）
type TemplateSubscribeData struct {
	Content TemplateSubscribeContent `json:"content"`
}

// TemplateSubscribeRequest 一次性订阅消息请求
// 接口文档：https://developers.weixin.qq.com/doc/offiaccount/Message_Management/One-time_subscription_info.html
// 功能：发送一次性订阅消息
// 请求方式：POST
// 注意：一次性订阅消息只能发送给已经授权订阅的用户
// 用户授权订阅的URL格式：%s?action=get_confirm&appid=APPID&scene=SCENE&template_id=TEMPLATE_ID&redirect_url=REDIRECT_URL&reserved=RESERVED#wechat_redirect
type TemplateSubscribeRequest struct {
	Touser     string                `json:"touser"`
	TemplateID string                `json:"template_id"`
	URL        string                `json:"url,omitempty"`
	Scene      string                `json:"scene"`
	Title      string                `json:"title"`
	Data       TemplateSubscribeData `json:"data"`
}

// TemplateSubscribeResponse 一次性订阅消息响应
type TemplateSubscribeResponse struct {
	core.APIResponse
}

// TemplateSubscribe 发送一次性订阅消息
// 接口文档：https://developers.weixin.qq.com/doc/offiaccount/Message_Management/One-time_subscription_info.html
// 功能：发送一次性订阅消息
// 请求方式：POST
// 注意：此接口用于发送一次性订阅消息，用户需要先通过授权URL进行订阅授权
//
// 参数:
//   - ctx context.Context 请求上下文
//   - request *TemplateSubscribeRequest 发送请求（Touser/TemplateID/Scene/Title/Data 必填）
//
// 返回:
//   - *TemplateSubscribeResponse 发送结果
//   - error 请求失败或微信返回业务错误时返回错误
func (c *SubscribeNotifyClient) TemplateSubscribe(ctx context.Context, request *TemplateSubscribeRequest) (*TemplateSubscribeResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result TemplateSubscribeResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLTemplateSubscribe, accessToken)
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    apiURL,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}
