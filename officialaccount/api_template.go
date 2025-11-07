package officialaccount

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jcbowen/wego/core"
)

type Template struct {
	req *core.Request
}

func NewTemplate(req *core.Request) *Template {
	return &Template{
		req: req,
	}
}

// TemplateMessageRequest 发送模板消息请求
type TemplateMessageRequest struct {
	ToUser      string                         `json:"touser"`
	TemplateID  string                         `json:"template_id"`
	URL         string                         `json:"url,omitempty"`
	MiniProgram TemplateMessageMiniProgram     `json:"miniprogram,omitempty"`
	Data        map[string]TemplateMessageData `json:"data"`
}

// TemplateMessageMiniProgram 小程序信息
type TemplateMessageMiniProgram struct {
	AppID    string `json:"appid"`
	PagePath string `json:"pagepath"`
}

// TemplateMessageData 模板消息数据
type TemplateMessageData struct {
	Value string `json:"value"`
	Color string `json:"color,omitempty"`
}

// SendTemplateMessageResponse 发送模板消息响应
type SendTemplateMessageResponse struct {
	core.APIResponse
	MsgID int64 `json:"msgid"`
}

// SendTemplateMessage 发送公众号模板消息
func (c *Template) SendTemplateMessage(ctx context.Context, template *TemplateMessageRequest, accessToken string) (*SendTemplateMessageResponse, error) {
	// 验证参数
	if template == nil {
		return nil, fmt.Errorf("模板消息不能为空")
	}
	if template.ToUser == "" {
		return nil, fmt.Errorf("接收用户不能为空")
	}
	if template.TemplateID == "" {
		return nil, fmt.Errorf("模板ID不能为空")
	}
	if template.Data == nil {
		return nil, fmt.Errorf("模板数据不能为空")
	}

	var result SendTemplateMessageResponse

	apiURL := fmt.Sprintf("%s?access_token=%s", APIMessageTemplateSendURL, url.QueryEscape(accessToken))
	err := c.req.Make(ctx, "POST", apiURL, template, &result)

	if err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}
