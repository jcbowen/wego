package officialaccount

import (
	"context"
	"fmt"
	"net/url"
	
	"github.com/jcbowen/wego/core"
)

// TemplateClient 模板消息客户端
type TemplateClient struct {
	Client *MPClient
}

// NewTemplateClient 创建新的模板消息客户端
func NewTemplateClient(client *MPClient) *TemplateClient {
	return &TemplateClient{
		Client: client,
	}
}

// SendTemplateMsgRequest 发送模板消息请求
type SendTemplateMsgRequest struct {
	Touser     string            `json:"touser"`
	TemplateID string            `json:"template_id"`
	URL        string            `json:"url,omitempty"`
	MiniProgram MiniProgramInfo  `json:"miniprogram,omitempty"`
	Data       map[string]TemplateData `json:"data"`
}

// MiniProgramInfo 小程序信息
type MiniProgramInfo struct {
	AppID    string `json:"appid"`
	PagePath string `json:"pagepath"`
}

// TemplateData 模板消息数据
type TemplateData struct {
	Value string `json:"value"`
	Color string `json:"color,omitempty"`
}

// SendTemplateMsgResponse 发送模板消息响应
type SendTemplateMsgResponse struct {
	core.APIResponse
	MsgID int64 `json:"msgid"`
}

// AddTemplateRequest 选用模板请求
type AddTemplateRequest struct {
	TemplateIDShort string `json:"template_id_short"`
}

// AddTemplateResponse 选用模板响应
type AddTemplateResponse struct {
	core.APIResponse
	TemplateID string `json:"template_id"`
}

// QueryBlockTmplMsgRequest 查询拦截的模板消息请求
type QueryBlockTmplMsgRequest struct {
	BeginDate string `json:"begin_date"`
	EndDate   string `json:"end_date"`
	Offset    int    `json:"offset"`
	Count     int    `json:"count"`
}

// BlockTmplMsg 拦截的模板消息
type BlockTmplMsg struct {
	TemplateID string `json:"template_id"`
	BlockTime  string `json:"block_time"`
	BlockType  int    `json:"block_type"`
	BlockReason string `json:"block_reason"`
}

// QueryBlockTmplMsgResponse 查询拦截的模板消息响应
type QueryBlockTmplMsgResponse struct {
	core.APIResponse
	TotalCount int            `json:"total_count"`
	List       []BlockTmplMsg `json:"list"`
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
	Title          string `json:"title"`
	PrimaryIndustry string `json:"primary_industry"`
	DeputyIndustry  string `json:"deputy_industry"`
	Content        string `json:"content"`
	Example        string `json:"example"`
}

// GetIndustryResponse 获取行业信息响应
type GetIndustryResponse struct {
	core.APIResponse
	PrimaryIndustry IndustryInfo `json:"primary_industry"`
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
func (c *TemplateClient) SendTemplateMessage(ctx context.Context, request *SendTemplateMsgRequest) (*SendTemplateMsgResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result SendTemplateMsgResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APISendTemplateMsgURL, url.QueryEscape(accessToken))
	err = c.Client.MakeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// AddTemplate 选用模板
func (c *TemplateClient) AddTemplate(ctx context.Context, templateIDShort string) (*AddTemplateResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := AddTemplateRequest{
		TemplateIDShort: templateIDShort,
	}

	var result AddTemplateResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APIAddTemplateURL, url.QueryEscape(accessToken))
	err = c.Client.MakeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// QueryBlockTmplMsg 查询拦截的模板消息
func (c *TemplateClient) QueryBlockTmplMsg(ctx context.Context, beginDate, endDate string, offset, count int) (*QueryBlockTmplMsgResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := QueryBlockTmplMsgRequest{
		BeginDate: beginDate,
		EndDate:   endDate,
		Offset:    offset,
		Count:     count,
	}

	var result QueryBlockTmplMsgResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APIQueryBlockTmplMsgURL, url.QueryEscape(accessToken))
	err = c.Client.MakeRequest(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIDeleteTemplateURL, url.QueryEscape(accessToken))
	err = c.Client.MakeRequest(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIGetAllTemplatesURL, url.QueryEscape(accessToken))
	err = c.Client.MakeRequest(ctx, "GET", apiURL, nil, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIGetIndustryURL, url.QueryEscape(accessToken))
	err = c.Client.MakeRequest(ctx, "GET", apiURL, nil, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APISetIndustryURL, url.QueryEscape(accessToken))
	err = c.Client.MakeRequest(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}