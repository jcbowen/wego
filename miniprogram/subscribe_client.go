package miniprogram

import (
	"context"

	"github.com/jcbowen/wego/core"
)

// 订阅消息 miniprogram_state 取值
const (
	SubscribeMsgStateFormal    = "formal"    // 正式版
	SubscribeMsgStateTrial     = "trial"     // 体验版
	SubscribeMsgStateDeveloper = "developer" // 开发版
)

// 订阅消息 lang 取值
const (
	SubscribeMsgLangZhCN = "zh_CN" // 简体中文
	SubscribeMsgLangZhHK = "zh_HK" // 繁体中文（香港）
	SubscribeMsgLangZhTW = "zh_TW" // 繁体中文（台湾）
	SubscribeMsgLangEn   = "en_US" // 英文
)

// 内容安全检测 scene 取值
const (
	SecCheckSceneProfile = 1 // 资料
	SecCheckSceneComment = 2 // 评论
	SecCheckSceneForum   = 3 // 论坛
	SecCheckSceneSocial  = 4 // 社交日志
)

// 内容安全检测 media_type 取值
// 官方文档：https://developers.weixin.qq.com/miniprogram/dev/server/API/sec-center/sec-check/api_mediacheckasync.html
// 官方原文「1:音频;2:图片」，官方不存在视频类型（提交视频将返回 40004）
const (
	SecCheckMediaTypeAudio = 1 // 音频
	SecCheckMediaTypeImage = 2 // 图片
)

// SubscribeMessageDataValue 订阅消息数据项的值
type SubscribeMessageDataValue struct {
	Value string `json:"value"` // 模板字段值
}

// SubscribeMessageData 订阅消息数据，key 为模板字段编号（如 thing1、time2）
type SubscribeMessageData map[string]SubscribeMessageDataValue

// SendSubscribeMessageRequest 发送订阅消息请求
type SendSubscribeMessageRequest struct {
	ToUser           string               `json:"touser"`                      // 接收用户的 OpenID
	TemplateID       string               `json:"template_id"`                 // 订阅消息模板 ID
	Page             string               `json:"page,omitempty"`              // 点击模板卡片跳转页面（空则无跳转）
	MiniprogramState string               `json:"miniprogram_state,omitempty"` // 跳转小程序类型：formal/trial/developer（不传时由微信按默认值处理）
	Lang             string               `json:"lang,omitempty"`              // 语言类型：zh_CN/zh_HK/zh_TW/en_US（不传时官方默认 zh_CN）
	Data             SubscribeMessageData `json:"data"`                        // 模板内容
}

// SendSubscribeMessageResponse 发送订阅消息响应
type SendSubscribeMessageResponse struct {
	core.APIResponse
}

// SendSubscribeMessage 发送订阅消息
//
// 参数:
//   - ctx: 上下文
//   - req: 发送请求（ToUser/TemplateID/Data 必填；MiniprogramState/Lang 传空时由微信按默认值处理）
//
// 返回:
//   - *SendSubscribeMessageResponse: 发送结果
//   - error: 请求失败或微信返回错误时返回错误
//
// 使用示例:
//
//	resp, err := client.SendSubscribeMessage(ctx, &miniprogram.SendSubscribeMessageRequest{
//		ToUser:     "openid",
//		TemplateID: "template_id",
//		Page:       "pages/index/index",
//		Data: miniprogram.SubscribeMessageData{
//			"thing1": {Value: "审核通过"},
//			"time2":  {Value: "2026-10-22 20:00"},
//		},
//	})
//
// 文档:
//   - https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-message-management/subscribe-message/sendMessage.html
func (c *Client) SendSubscribeMessage(ctx context.Context, req *SendSubscribeMessageRequest) (*SendSubscribeMessageResponse, error) {
	accessToken, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result SendSubscribeMessageResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    SubscribeMessageSendURL,
		Query: map[string]string{
			"access_token": accessToken,
		},
		Body:   req,
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
