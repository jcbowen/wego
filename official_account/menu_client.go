package official_account

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jcbowen/wego/core"
)

// MenuClient 自定义菜单客户端
type MenuClient struct {
	Client *Client
}

// NewMenuClient 创建新的自定义菜单客户端
func NewMenuClient(client *Client) *MenuClient {
	return &MenuClient{
		Client: client,
	}
}

// Button 菜单按钮
type Button struct {
	Type      string   `json:"type,omitempty"`
	Name      string   `json:"name"`
	Key       string   `json:"key,omitempty"`
	URL       string   `json:"url,omitempty"`
	MediaID   string   `json:"media_id,omitempty"`
	AppID     string   `json:"appid,omitempty"`
	PagePath  string   `json:"pagepath,omitempty"`
	SubButton []Button `json:"sub_button,omitempty"`
}

// Menu 菜单结构
type Menu struct {
	Button []Button `json:"button"`
}

// ConditionalMenu 个性化菜单
type ConditionalMenu struct {
	Button    []Button  `json:"button"`
	MatchRule MatchRule `json:"matchrule"`
	MenuID    string    `json:"menuid,omitempty"`
}

// MatchRule 匹配规则
type MatchRule struct {
	TagID              string `json:"tag_id,omitempty"`
	Sex                string `json:"sex,omitempty"`
	Country            string `json:"country,omitempty"`
	Province           string `json:"province,omitempty"`
	City               string `json:"city,omitempty"`
	ClientPlatformType string `json:"client_platform_type,omitempty"`
	Language           string `json:"language,omitempty"`
}

// CreateMenuResponse 创建菜单响应
type CreateMenuResponse struct {
	core.APIResponse
}

// GetCurrentMenuResponse 获取当前菜单信息响应
type GetCurrentMenuResponse struct {
	core.APIResponse
	IsMenuOpen   int `json:"is_menu_open"`
	SelfMenuInfo struct {
		Button []Button `json:"button"`
	} `json:"selfmenu_info"`
}

// GetMenuResponse 获取菜单响应
type GetMenuResponse struct {
	core.APIResponse
	Menu            Menu              `json:"menu"`
	ConditionalMenu []ConditionalMenu `json:"conditionalmenu"`
}

// DeleteMenuResponse 删除菜单响应
type DeleteMenuResponse struct {
	core.APIResponse
}

// AddConditionalMenuResponse 添加个性化菜单响应
type AddConditionalMenuResponse struct {
	core.APIResponse
	MenuID string `json:"menuid"`
}

// DeleteConditionalMenuResponse 删除个性化菜单响应
type DeleteConditionalMenuResponse struct {
	core.APIResponse
}

// TryMatchMenuRequest 测试个性化菜单匹配请求
type TryMatchMenuRequest struct {
	UserID string `json:"user_id"`
}

// TryMatchMenuResponse 测试个性化菜单匹配响应
type TryMatchMenuResponse struct {
	core.APIResponse
	Button []Button `json:"button"`
}

// CreateMenu 创建自定义菜单
func (c *MenuClient) CreateMenu(ctx context.Context, menu *Menu) (*CreateMenuResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result CreateMenuResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APICreateMenuURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, menu, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetCurrentSelfmenuInfo 获取当前菜单信息
func (c *MenuClient) GetCurrentSelfmenuInfo(ctx context.Context) (*GetCurrentMenuResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetCurrentMenuResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APIGetCurrentMenuURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetMenu 获取菜单
func (c *MenuClient) GetMenu(ctx context.Context) (*GetMenuResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetMenuResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APIGetMenuURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// DeleteMenu 删除菜单
func (c *MenuClient) DeleteMenu(ctx context.Context) (*DeleteMenuResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result DeleteMenuResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APIDeleteMenuURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// AddConditionalMenu 添加个性化菜单
func (c *MenuClient) AddConditionalMenu(ctx context.Context, menu *ConditionalMenu) (*AddConditionalMenuResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result AddConditionalMenuResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APIAddConditionalMenuURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, menu, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// DeleteConditionalMenu 删除个性化菜单
func (c *MenuClient) DeleteConditionalMenu(ctx context.Context, menuID string) (*DeleteConditionalMenuResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result DeleteConditionalMenuResponse
	apiURL := fmt.Sprintf("%s?access_token=%s&menuid=%s", APIDeleteConditionalMenuURL, url.QueryEscape(accessToken), url.QueryEscape(menuID))
	err = c.Client.req.Make(ctx, "POST", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// TryMatchMenu 测试个性化菜单匹配
func (c *MenuClient) TryMatchMenu(ctx context.Context, userID string) (*TryMatchMenuResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := TryMatchMenuRequest{
		UserID: userID,
	}

	var result TryMatchMenuResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APITryMatchMenuURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}
