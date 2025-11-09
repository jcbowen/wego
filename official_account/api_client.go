package official_account

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jcbowen/wego/core"
)

// APIClient 微信公众号API客户端
type APIClient struct {
	Client *Client
}

// NewMPAPIClient 创建新的微信公众号API客户端
func NewMPAPIClient(client *Client) *APIClient {
	return &APIClient{
		Client: client,
	}
}

// CallbackCheckRequest 网络通信检测请求
type CallbackCheckRequest struct {
	Action        string `json:"action"`
	CheckOperator string `json:"check_operator"`
}

// CallbackCheckResponse 网络通信检测响应
type CallbackCheckResponse struct {
	core.APIResponse
	DNS  []string `json:"dns"`
	Ping []string `json:"ping"`
}

// GetApiDomainIpResponse 获取微信API服务器IP响应
type GetApiDomainIpResponse struct {
	core.APIResponse
	IPList []string `json:"ip_list"`
}

// GetCallbackIpResponse 获取微信推送服务器IP响应
type GetCallbackIpResponse struct {
	core.APIResponse
	IPList []string `json:"ip_list"`
}

// CallbackCheck 网络通信检测
func (c *APIClient) CallbackCheck(ctx context.Context, action, checkOperator string) (*CallbackCheckResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := CallbackCheckRequest{
		Action:        action,
		CheckOperator: checkOperator,
	}

	var result CallbackCheckResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLCallbackCheck, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetApiDomainIp 获取微信API服务器IP
func (c *APIClient) GetApiDomainIp(ctx context.Context) (*GetApiDomainIpResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetApiDomainIpResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetApiDomainIp, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetCallbackIp 获取微信推送服务器IP
func (c *APIClient) GetCallbackIp(ctx context.Context) (*GetCallbackIpResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetCallbackIpResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetCallbackIp, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}
