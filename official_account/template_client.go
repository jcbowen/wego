package official_account

import (
	"context"
	"fmt"

	"github.com/jcbowen/wego/core"
)

// TemplateClient 模板消息客户端
type TemplateClient struct {
	Client *Client
}

// NewTemplateClient 创建新的模板消息客户端
func NewTemplateClient(client *Client) *TemplateClient {
	return &TemplateClient{
		Client: client,
	}
}

// AddTemplateRequest 选用模板请求
type AddTemplateRequest struct {
	TemplateIDShort string   `json:"template_id_short"`
	KeywordNameList []string `json:"keyword_name_list"` // 模板中关键词的名称（必填，须与模板库中模板对应的关键词匹配，为空返回 40246）
}

// AddTemplateResponse 选用模板响应
type AddTemplateResponse struct {
	core.APIResponse
	TemplateID string `json:"template_id"`
}

// QueryBlockTmplMsgRequest 查询拦截的模板消息请求
//
// 官方文档：https://developers.weixin.qq.com/doc/service/api/notify/template/api_queryblocktmplmsg.html
type QueryBlockTmplMsgRequest struct {
	TmplMsgID string `json:"tmpl_msg_id"` // 被拦截的模板消息id（必填）
	LargestID int64  `json:"largest_id"`  // 上一页查询结果最大的id，用于翻页，第一次传0（必填）
	Limit     int    `json:"limit"`       // 单页查询的大小，最大100（必填）
}

// BlockTmplMsgInfo 被拦截的模板消息记录
type BlockTmplMsgInfo struct {
	ID            core.FlexString `json:"id"`             // 记录唯一ID，用于多次查询的largest_id（官方参数表为 string、示例为数字，做双形态兼容）
	TmplMsgID     string          `json:"tmpl_msg_id"`    // 被拦截的模板消息id
	Title         string          `json:"title"`          // 模板消息的标题
	Content       string          `json:"content"`        // 模板消息的内容
	SendTimestamp int64           `json:"send_timestamp"` // 下发的时间戳
	OpenID        string          `json:"openid"`         // 下发目标用户的openid
}

// QueryBlockTmplMsgResponse 查询拦截的模板消息响应
type QueryBlockTmplMsgResponse struct {
	core.APIResponse
	MsgInfo *BlockTmplMsgInfo `json:"msginfo"` // 消息信息
}

// DeleteTemplateRequest 删除模板请求
type DeleteTemplateRequest struct {
	TemplateID string `json:"template_id"`
}

// DeleteTemplateResponse 删除模板响应
type DeleteTemplateResponse struct {
	core.APIResponse
}

// GetAllTemplatesResponse 获取已选用模板列表响应
type GetAllTemplatesResponse struct {
	core.APIResponse
	TemplateList []TemplateInfo `json:"template_list"`
}

// TemplateInfo 模板信息
type TemplateInfo struct {
	TemplateID      string `json:"template_id"`
	Title           string `json:"title"`
	PrimaryIndustry string `json:"primary_industry"`
	DeputyIndustry  string `json:"deputy_industry"`
	Content         string `json:"content"`
	Example         string `json:"example"`
}

// GetIndustryResponse 获取行业信息响应
type GetIndustryResponse struct {
	core.APIResponse
	PrimaryIndustry   IndustryInfo `json:"primary_industry"`
	SecondaryIndustry IndustryInfo `json:"secondary_industry"`
}

// IndustryInfo 行业信息
type IndustryInfo struct {
	FirstClass  string `json:"first_class"`
	SecondClass string `json:"second_class"`
}

// SetIndustryRequest 设置所属行业请求
type SetIndustryRequest struct {
	IndustryID1 string `json:"industry_id1"`
	IndustryID2 string `json:"industry_id2"`
}

// SetIndustryResponse 设置所属行业响应
type SetIndustryResponse struct {
	core.APIResponse
}

// SendTemplateMessage 发送模板消息
func (c *TemplateClient) SendTemplateMessage(ctx context.Context, request *TemplateMessageRequest) (*SendTemplateMessageResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result *SendTemplateMessageResponse
	result, err = NewTemplate(c.Client.req).SendTemplateMessage(ctx, request, accessToken)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// AddTemplate 选用模板
//
// 参数:
//   - ctx: 上下文
//   - templateIDShort: 模板库中模板的编号
//   - keywordNameList: 模板中关键词的名称（须与模板库中模板对应的关键词匹配，为空返回 40246）
//
// 返回:
//   - *AddTemplateResponse: 含选用后获得的模板ID
//   - error: 请求失败或微信返回错误时返回错误
func (c *TemplateClient) AddTemplate(ctx context.Context, templateIDShort string, keywordNameList []string) (*AddTemplateResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := AddTemplateRequest{
		TemplateIDShort: templateIDShort,
		KeywordNameList: keywordNameList,
	}

	var result AddTemplateResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLTemplateAddTemplate, accessToken)
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

// QueryBlockTmplMsg 查询拦截的模板消息
//
// 参数:
//   - ctx: 上下文
//   - tmplMsgID: 被拦截的模板消息id（必填）
//   - largestID: 上一页查询结果最大的id，用于翻页，第一次传0
//   - limit: 单页查询的大小，最大100
//
// 返回:
//   - *QueryBlockTmplMsgResponse: 被拦截的模板消息信息
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/service/api/notify/template/api_queryblocktmplmsg.html
func (c *TemplateClient) QueryBlockTmplMsg(ctx context.Context, tmplMsgID string, largestID int64, limit int) (*QueryBlockTmplMsgResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := QueryBlockTmplMsgRequest{
		TmplMsgID: tmplMsgID,
		LargestID: largestID,
		Limit:     limit,
	}

	var result QueryBlockTmplMsgResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLQueryBlockTmplMsg, accessToken)
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

// DeleteTemplate 删除模板
func (c *TemplateClient) DeleteTemplate(ctx context.Context, templateID string) (*DeleteTemplateResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := DeleteTemplateRequest{
		TemplateID: templateID,
	}

	var result DeleteTemplateResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLTemplateDelPrivateTemplate, accessToken)
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

// GetAllTemplates 获取已选用模板列表
func (c *TemplateClient) GetAllTemplates(ctx context.Context) (*GetAllTemplatesResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetAllTemplatesResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLTemplateGetAllPrivateTemplate, accessToken)
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

// GetIndustry 获取行业信息
func (c *TemplateClient) GetIndustry(ctx context.Context) (*GetIndustryResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetIndustryResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLTemplateGetIndustry, accessToken)
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

// SetIndustry 设置所属行业
func (c *TemplateClient) SetIndustry(ctx context.Context, industryID1, industryID2 string) (*SetIndustryResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := SetIndustryRequest{
		IndustryID1: industryID1,
		IndustryID2: industryID2,
	}

	var result SetIndustryResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLTemplateSetIndustry, accessToken)
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
