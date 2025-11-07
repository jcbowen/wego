package officialaccount

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jcbowen/wego/core"
)

// SubscribeClient 订阅通知客户端
type SubscribeClient struct {
	Client *Client
}

// NewSubscribeClient 创建新的订阅通知客户端
func NewSubscribeClient(client *Client) *SubscribeClient {
	return &SubscribeClient{
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
type PubTemplateTitleInfo struct {
	TID        int    `json:"tid"`
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
	Data []PubTemplateKeywordInfo `json:"data"`
}

// WxaPubNewTemplateInfo 小程序公共模板信息
type WxaPubNewTemplateInfo struct {
	PriTmplID string `json:"priTmplId"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Example   string `json:"example"`
	Type      int    `json:"type"`
}

// GetWxaPubNewTemplateResponse 获取小程序公共模板列表响应
type GetWxaPubNewTemplateResponse struct {
	core.APIResponse
	Data []WxaPubNewTemplateInfo `json:"data"`
}

// AddWxaNewTemplateRequest 添加小程序模板请求
type AddWxaNewTemplateRequest struct {
	TID       int    `json:"tid"`
	KidList   []int  `json:"kidList"`
	SceneDesc string `json:"sceneDesc"`
}

// AddWxaNewTemplateResponse 添加小程序模板响应
type AddWxaNewTemplateResponse struct {
	core.APIResponse
	PriTmplID string `json:"priTmplId"`
}

// DelWxaNewTemplateRequest 删除小程序模板请求
type DelWxaNewTemplateRequest struct {
	PriTmplID string `json:"priTmplId"`
}

// DelWxaNewTemplateResponse 删除小程序模板响应
type DelWxaNewTemplateResponse struct {
	core.APIResponse
}

// SubscribeMsgData 订阅消息数据
type SubscribeMsgData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SendNewSubscribeMsgRequest 发送订阅消息请求
type SendNewSubscribeMsgRequest struct {
	Touser      string                      `json:"touser"`
	TemplateID  string                      `json:"template_id"`
	Page        string                      `json:"page,omitempty"`
	Data        map[string]SubscribeMsgData `json:"data"`
	Miniprogram *TemplateMessageMiniProgram `json:"miniprogram_state,omitempty"`
	Lang        string                      `json:"lang,omitempty"`
}

// SendNewSubscribeMsgResponse 发送订阅消息响应
type SendNewSubscribeMsgResponse struct {
	core.APIResponse
}

// GetCategory 获取订阅通知分类列表
// 接口文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/subscribe-message/subscribeMessage.getCategory.html
// 功能：获取小程序账号的类目
// 请求方式：GET
func (c *SubscribeClient) GetCategory(ctx context.Context) (*GetCategoryResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetCategoryResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APIGetCategoryURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetPubNewTemplateTitles 获取公共模板标题列表
// 接口文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/subscribe-message/subscribeMessage.getPubTemplateTitleList.html
// 功能：获取公共模板标题列表
// 请求方式：GET
func (c *SubscribeClient) GetPubNewTemplateTitles(ctx context.Context, categoryID int, start, limit int) (*GetPubNewTemplateTitlesResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetPubNewTemplateTitlesResponse
	apiURL := fmt.Sprintf("%s?access_token=%s&ids=%d&start=%d&limit=%d",
		APIGetPubNewTemplateTitlesURL, url.QueryEscape(accessToken), categoryID, start, limit)
	err = c.Client.req.Make(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetPubNewTemplateKeywords 获取公共模板关键词列表
// 接口文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/subscribe-message/subscribeMessage.getPubTemplateKeyWordsById.html
// 功能：获取模板标题下的关键词列表
// 请求方式：GET
func (c *SubscribeClient) GetPubNewTemplateKeywords(ctx context.Context, tid int) (*GetPubNewTemplateKeywordsResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetPubNewTemplateKeywordsResponse
	apiURL := fmt.Sprintf("%s?access_token=%s&tid=%d",
		APIGetPubNewTemplateKeywordsURL, url.QueryEscape(accessToken), tid)
	err = c.Client.req.Make(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetWxaPubNewTemplate 获取小程序公共模板列表
// 接口文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/subscribe-message/subscribeMessage.getTemplateList.html
// 功能：获取当前帐号下的个人模板列表
// 请求方式：GET
func (c *SubscribeClient) GetWxaPubNewTemplate(ctx context.Context) (*GetWxaPubNewTemplateResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetWxaPubNewTemplateResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APIGetWxaPubNewTemplateURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// AddWxaNewTemplate 添加小程序模板
// 接口文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/subscribe-message/subscribeMessage.addTemplate.html
// 功能：组合模板并添加至帐号下的个人模板库
// 请求方式：POST
func (c *SubscribeClient) AddWxaNewTemplate(ctx context.Context, tid int, kidList []int, sceneDesc string) (*AddWxaNewTemplateResponse, error) {
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIAddWxaNewTemplateURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// DelWxaNewTemplate 删除小程序模板
// 接口文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/subscribe-message/subscribeMessage.deleteTemplate.html
// 功能：删除帐号下的个人模板
// 请求方式：POST
func (c *SubscribeClient) DelWxaNewTemplate(ctx context.Context, priTmplID string) (*DelWxaNewTemplateResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := DelWxaNewTemplateRequest{
		PriTmplID: priTmplID,
	}

	var result DelWxaNewTemplateResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APIDelWxaNewTemplateURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// SendNewSubscribeMsg 发送订阅消息
// 接口文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/subscribe-message/subscribeMessage.send.html
// 功能：发送订阅消息
// 请求方式：POST
func (c *SubscribeClient) SendNewSubscribeMsg(ctx context.Context, request *SendNewSubscribeMsgRequest) (*SendNewSubscribeMsgResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result SendNewSubscribeMsgResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APISendNewSubscribeMsgURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// TemplateSubscribeData 一次性订阅消息数据
type TemplateSubscribeData struct {
	Value string `json:"value"`
	Color string `json:"color,omitempty"`
}

// TemplateSubscribeRequest 一次性订阅消息请求
// 接口文档：https://developers.weixin.qq.com/doc/offiaccount/Message_Management/One-time_subscription_info.html
// 功能：发送一次性订阅消息
// 请求方式：POST
// 注意：一次性订阅消息只能发送给已经授权订阅的用户
// 用户授权订阅的URL格式：https://mp.weixin.qq.com/mp/subscribemsg?action=get_confirm&appid=APPID&scene=SCENE&template_id=TEMPLATE_ID&redirect_url=REDIRECT_URL&reserved=RESERVED#wechat_redirect
type TemplateSubscribeRequest struct {
	Touser     string                           `json:"touser"`
	TemplateID string                           `json:"template_id"`
	URL        string                           `json:"url,omitempty"`
	Scene      string                           `json:"scene"`
	Title      string                           `json:"title"`
	Data       map[string]TemplateSubscribeData `json:"data"`
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
func (c *SubscribeClient) TemplateSubscribe(ctx context.Context, request *TemplateSubscribeRequest) (*TemplateSubscribeResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result TemplateSubscribeResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APITemplateSubscribeURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}
