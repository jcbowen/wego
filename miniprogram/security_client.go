package miniprogram

import (
	"context"

	"github.com/jcbowen/wego/core"
)

// MsgSecCheckRequest 文本内容安全检测请求（version 2）
//
// 注意：v2 接口必须传入内容发布用户的 openid 与 scene，否则会报错。
// 参考文档：https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/sec-center/sec-check/msgSecCheck.html
type MsgSecCheckRequest struct {
	Content    string `json:"content"`            // 要检测的文本内容（最长 2500 字）
	Version    string `json:"version"`            // 接口版本号，固定填 "2"
	OpenID     string `json:"openid"`             // 用户的 openid（用户需在近两小时访问过小程序）
	Scene      int    `json:"scene"`              // 场景值：1资料 2评论 3论坛 4社交日志
	Title      string `json:"title,omitempty"`    // 文本标题（可选，需填写内容标题）
	Nickname   string `json:"nickname,omitempty"` // 用户昵称（可选）
	Signature  string `json:"signature,omitempty"`
	HeadAvatar string `json:"head_avatar,omitempty"` // 用户头像图片链接（可选）
}

// MsgSecCheckResult 检测结果
type MsgSecCheckResult struct {
	Suggest string `json:"suggest"` // 建议：pass(通过)/review(需人工复审)/risky(不通过)
	Label   int    `json:"label"`   // 命中标签：100正常 10001广告 20001时政 20002色情 20003辱骂 20006违法犯罪 20008欺诈 20012版权 21000其他
}

// MsgSecCheckResponse 文本内容安全检测响应
type MsgSecCheckResponse struct {
	core.APIResponse
	Result  *MsgSecCheckResult `json:"result"`   // 检测结果（errcode 为 0 时存在）
	TraceID string             `json:"trace_id"` // 追踪 ID，用于异步结果关联与工单排查
	Detail  []struct {         // 详细检测结果
		Strategy string `json:"strategy"`
		ErrCode  int    `json:"errcode"`
		Suggest  string `json:"suggest"`
		Label    int    `json:"label"`
		Prob     int    `json:"prob"`
	} `json:"detail"`
}

// MsgSecCheck 文本内容安全检测（同步返回结果）
//
// 参数:
//   - ctx: 上下文
//   - content: 要检测的文本内容（最长 2500 字）
//   - openID: 内容发布用户的 openid
//   - scene: 场景值（1资料 2评论 3论坛 4社交日志）
//
// 返回:
//   - *MsgSecCheckResponse: 检测结果（Result.Suggest 为 pass/review/risky）
//   - error: 请求失败或微信返回错误时返回错误
//
// 使用示例:
//
//	resp, err := client.MsgSecCheck(ctx, "待检测文本", openid, miniprogram.SecCheckSceneComment)
//	if err == nil && resp.Result.Suggest == "risky" { /* 拦截内容 */ }
//
// 文档:
//   - https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/sec-center/sec-check/msgSecCheck.html
func (c *Client) MsgSecCheck(ctx context.Context, content, openID string, scene int) (*MsgSecCheckResponse, error) {
	accessToken, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result MsgSecCheckResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    MsgSecCheckURL,
		Query: map[string]string{
			"access_token": accessToken,
		},
		Body: MsgSecCheckRequest{
			Content: content,
			Version: "2",
			OpenID:  openID,
			Scene:   scene,
		},
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

// MediaCheckAsyncRequest 音视频内容安全检测请求（version 2，异步）
type MediaCheckAsyncRequest struct {
	MediaURL  string `json:"media_url"`  // 要检测的图片/音频/视频的 url（支持图片和音频，音频小于 5MB）
	MediaType int    `json:"media_type"` // 媒体类型：2图片 3音频 4视频
	Version   string `json:"version"`    // 接口版本号，固定填 "2"
	OpenID    string `json:"openid"`     // 用户的 openid（用户需在近两小时访问过小程序）
	Scene     int    `json:"scene"`      // 场景值：1资料 2评论 3论坛 4社交日志
}

// MediaCheckAsyncResponse 音视频内容安全检测响应（结果通过消息推送异步告知）
type MediaCheckAsyncResponse struct {
	core.APIResponse
	TraceID string `json:"trace_id"` // 追踪 ID，检测结果推送时会携带该值
}

// MediaCheckAsync 音视频内容安全检测（异步，检测结果通过消息推送服务器异步告知）
//
// 参数:
//   - ctx: 上下文
//   - mediaURL: 要检测的媒体资源 url
//   - mediaType: 媒体类型（2图片 3音频 4视频）
//   - openID: 内容发布用户的 openid
//   - scene: 场景值（1资料 2评论 3论坛 4社交日志）
//
// 返回:
//   - *MediaCheckAsyncResponse: 提交结果（携带 TraceID，实际检测结果走消息推送）
//   - error: 请求失败或微信返回错误时返回错误
//
// 注意: 异步检测结果通过小程序消息推送（wxa_media_check 推送事件）送达，需在消息推送服务中处理。
//
// 文档:
//   - https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/sec-center/sec-check/mediaCheckAsync.html
func (c *Client) MediaCheckAsync(ctx context.Context, mediaURL string, mediaType int, openID string, scene int) (*MediaCheckAsyncResponse, error) {
	accessToken, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result MediaCheckAsyncResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    MediaCheckAsyncURL,
		Query: map[string]string{
			"access_token": accessToken,
		},
		Body: MediaCheckAsyncRequest{
			MediaURL:  mediaURL,
			MediaType: mediaType,
			Version:   "2",
			OpenID:    openID,
			Scene:     scene,
		},
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
