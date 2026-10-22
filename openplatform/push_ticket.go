// Package openplatform 微信开放平台API - 启动票据推送服务
// 参考：https://developers.weixin.qq.com/doc/oplatform/openApi/ticket-token/api_startpushticket.html
package openplatform

import (
	"context"

	"github.com/jcbowen/wego/core"
)

// StartPushTicketRequest 启动票据推送服务请求
type StartPushTicketRequest struct {
	ComponentAppID  string `json:"component_appid"`  // 平台型第三方平台的 AppID
	ComponentSecret string `json:"component_secret"` // 平台型第三方平台的 APPSECRET
}

// StartPushTicket 启动 component_verify_ticket 推送服务
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - *core.APIResponse: 接口响应（errcode=0 表示已启动或本来就处于正常状态 "in a normal state"）
//   - error: 请求失败或微信返回错误时返回错误
//
// 说明:
//   - 用于第三方平台错过 ticket 推送或推送异常后的主动恢复
//   - 新版官方文档明确该接口无 query 参数，凭证信息全部通过请求体传递
//
// 使用示例:
//
//	resp, err := openPlatformClient.StartPushTicket(ctx)
//
// 文档:
//   - https://developers.weixin.qq.com/doc/oplatform/openApi/ticket-token/api_startpushticket.html
func (c *Client) StartPushTicket(ctx context.Context) (*core.APIResponse, error) {
	request := StartPushTicketRequest{
		ComponentAppID:  c.config.ComponentAppID,
		ComponentSecret: c.config.ComponentAppSecret,
	}

	var result core.APIResponse
	err := c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLStartPushTicket,
		Body:   request,
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result
	}
	return &result, nil
}
