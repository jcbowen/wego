// Package official_account 微信公众号API - 模板消息
// 本接口支持「服务号（仅认证）」账号类型调用。其他账号类型如无特殊说明，均不可调用。
package official_account

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
	ToUser      string                         `json:"touser"`                  // 接收者（用户）的 openid
	TemplateID  string                         `json:"template_id"`             // 所需下发的订阅模板id
	URL         string                         `json:"url,omitempty"`           // 模板跳转链接（海外账号没有跳转能力,url 和 miniprogram 同时不填，无跳转，url 和 miniprogram 同时填写，优先跳转小程序）
	MiniProgram TemplateMessageMiniProgram     `json:"miniprogram,omitempty"`   // 跳转小程序时填写（url 和 miniprogram 同时不填，无跳转，page 和 miniprogram 同时填写，优先跳转小程序）
	Data        map[string]TemplateMessageData `json:"data"`                    // 模板内容，需根据模板给定的格式给出（参考注意事项），格式形如 { "key1": { "value": any }, "key2": { "value": any }
	ClientMsgID string                         `json:"client_msg_id,omitempty"` // 防重入id。对于同一个openid + client_msg_id, 只发送一条消息,10分钟有效,超过10分钟不保证效果。若无防重入需求，可不填
}

// TemplateMessageMiniProgram 小程序信息
// 跳转小程序时填写（url 和 miniprogram 同时不填，无跳转，page 和 miniprogram 同时填写，优先跳转小程序）
type TemplateMessageMiniProgram struct {
	AppID    string `json:"appid"`    // 小程序appid
	PagePath string `json:"pagepath"` // 小程序跳转路径
}

// TemplateMessageData 模板消息数据
// 模板内容，需根据模板给定的格式给出（参考注意事项），格式形如 { "key1": { "value": any }, "key2": { "value": any } }
type TemplateMessageData struct {
	Value string `json:"value"`
	// Color string `json:"color,omitempty"` // 已经废弃
}

// SendTemplateMessageResponse 发送模板消息响应
type SendTemplateMessageResponse struct {
	core.APIResponse
	MsgID int64 `json:"msgid"` // 消息id
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

	apiURL := fmt.Sprintf("%s?access_token=%s", URLMessageTemplateSend, url.QueryEscape(accessToken))
	err := c.req.Make(ctx, "POST", apiURL, template, &result)

	if err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}
