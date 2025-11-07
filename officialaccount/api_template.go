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

// SendTemplateMsgRequest 发送模板消息请求
type SendTemplateMsgRequest struct {
	ToUser      string                  `json:"touser"`
	TemplateID  string                  `json:"template_id"`
	URL         string                  `json:"url,omitempty"`
	MiniProgram MiniProgramInfo         `json:"miniprogram,omitempty"`
	Data        map[string]TemplateData `json:"data"`
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

// SendTemplateMessage 发送公众号模板消息
func (c *Template) SendTemplateMessage(ctx context.Context, template *SendTemplateMsgRequest, accessToken string) (*SendTemplateMsgResponse, error) {
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

	var result SendTemplateMsgResponse

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
