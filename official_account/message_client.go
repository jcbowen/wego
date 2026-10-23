package official_account

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"

	"github.com/jcbowen/wego/core"
)

// MessageClient 消息管理客户端
type MessageClient struct {
	Client *Client
}

// NewMessageClient 创建新的消息管理客户端
func NewMessageClient(client *Client) *MessageClient {
	return &MessageClient{
		Client: client,
	}
}

// UploadImageResponse 上传图文消息图片响应
type UploadImageResponse struct {
	core.APIResponse
	URL string `json:"url"`
}

// DeleteMassMsgRequest 删除群发消息请求
type DeleteMassMsgRequest struct {
	MsgID      int64 `json:"msg_id"`
	ArticleIdx int   `json:"article_idx,omitempty"`
}

// DeleteMassMsgResponse 删除群发消息响应
type DeleteMassMsgResponse struct {
	core.APIResponse
}

// GetSpeedResponse 获取群发速度响应
type GetSpeedResponse struct {
	core.APIResponse
	Speed     int `json:"speed"`     // 群发速度级别
	RealSpeed int `json:"realspeed"` // 实时群发速度级别（官方字段名无下划线）
}

// MassMsgGetRequest 获取群发消息发送状态请求
type MassMsgGetRequest struct {
	MsgID int64 `json:"msg_id"`
}

// MassMsgGetResponse 获取群发消息发送状态响应
type MassMsgGetResponse struct {
	core.APIResponse
	MsgID     int64  `json:"msg_id"`
	MsgStatus string `json:"msg_status"`
}

// MassContent 群发/预览消息的内容体接口
//
// 官方文档要求：群发与预览接口的内容体**以消息类型命名**，与 msgtype 平级，例如
//
//	{"touser":["OPENID"],"msgtype":"text","text":{"content":"hello"}}
//	{"filter":{"is_to_all":true},"msgtype":"mpnews","mpnews":{"media_id":"xxx"}}
//
// 早期版本用统一的 "content" 字段承载，微信侧无法识别，会直接报错。
type MassContent interface {
	// ContentKey 返回内容体在请求体中的键名，同时作为 msgtype 的取值
	// 取值范围：text / image / voice / mpnews / mpvideo / wxcard
	ContentKey() string
}

// MassTextContent 文本消息内容体
type MassTextContent struct {
	Content string `json:"content"`
}

func (c MassTextContent) ContentKey() string { return "text" }

// MassImageContent 图片消息内容体（旧格式，键 image + media_id 单图）
//
// Deprecated: 官方现行示例为 images 键 + media_ids 数组（见 MassImagesContent），
// 旧格式是否仍被微信接受待真机验证，建议优先使用 MassImagesContent。
type MassImageContent struct {
	MediaID string `json:"media_id"`
}

func (c MassImageContent) ContentKey() string { return "image" }

// MassImagesContent 图片消息内容体（官方现行示例格式，键 images）
//
// 官方示例（msgtype 仍为 image，内容键为 images）:
//
//	{"filter":{"is_to_all":true},"msgtype":"image","images":{"media_ids":["MEDIA_ID1"],"recommend":"...","title":"..."}}
//
// 文档:
//   - https://developers.weixin.qq.com/doc/service/api/notify/message/api_sendall.html
type MassImagesContent struct {
	MediaIDs           []string `json:"media_ids"`                       // 用于群发的图片 media_id 列表
	Recommend          string   `json:"recommend,omitempty"`             // 推荐语
	Title              string   `json:"title,omitempty"`                 // 标题
	NeedOpenComment    int      `json:"need_open_comment,omitempty"`     // 开启评论（1-开启，0-关闭）
	OnlyFansCanComment int      `json:"only_fans_can_comment,omitempty"` // 只有粉丝能评论（1-开启，0-关闭）
}

func (c MassImagesContent) ContentKey() string { return "images" }

// MassVoiceContent 语音消息内容体
type MassVoiceContent struct {
	MediaID string `json:"media_id"`
}

func (c MassVoiceContent) ContentKey() string { return "voice" }

// MassMPNewsContent 图文消息内容体
type MassMPNewsContent struct {
	MediaID string `json:"media_id"`
}

func (c MassMPNewsContent) ContentKey() string { return "mpnews" }

// MassMPVideoContent 视频消息内容体
type MassMPVideoContent struct {
	MediaID     string `json:"media_id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

func (c MassMPVideoContent) ContentKey() string { return "mpvideo" }

// MassWXCardContent 卡券消息内容体
type MassWXCardContent struct {
	CardID  string `json:"card_id"`
	CardExt string `json:"card_ext,omitempty"`
}

func (c MassWXCardContent) ContentKey() string { return "wxcard" }

// derefMassContent 将指针形态的内容体归一化为值形态，便于统一 switch 处理；
// 类型化 nil 指针（如 (*MassTextContent)(nil)）同样归一为 nil
func derefMassContent(content MassContent) MassContent {
	if content == nil {
		return nil
	}
	v := reflect.ValueOf(content)
	if v.Kind() != reflect.Ptr {
		return content
	}
	if v.IsNil() {
		return nil
	}
	if deref, ok := v.Elem().Interface().(MassContent); ok {
		return deref
	}
	return content
}

// MassContentFields 群发/预览请求体中「按类型命名」的内容字段
//
// 该结构以匿名方式嵌入各请求体，序列化后与 msgtype 平级，符合官方格式。
type MassContentFields struct {
	Text    *MassTextContent    `json:"text,omitempty"`
	Image   *MassImageContent   `json:"image,omitempty"`
	Images  *MassImagesContent  `json:"images,omitempty"` // 官方现行图片格式
	Voice   *MassVoiceContent   `json:"voice,omitempty"`
	MPNews  *MassMPNewsContent  `json:"mpnews,omitempty"`
	MPVideo *MassMPVideoContent `json:"mpvideo,omitempty"`
	WXCard  *MassWXCardContent  `json:"wxcard,omitempty"`
}

// setContent 根据内容体类型填充对应字段
// content 为 nil 或类型化 nil 指针（如 (*MassTextContent)(nil)）时返回错误
func (f *MassContentFields) setContent(content MassContent) error {
	c := derefMassContent(content)
	if c == nil {
		return fmt.Errorf("群发消息内容不能为空")
	}
	switch c := c.(type) {
	case MassTextContent:
		f.Text = &c
	case MassImageContent:
		f.Image = &c
	case MassImagesContent:
		f.Images = &c
	case MassVoiceContent:
		f.Voice = &c
	case MassMPNewsContent:
		f.MPNews = &c
	case MassMPVideoContent:
		f.MPVideo = &c
	case MassWXCardContent:
		f.WXCard = &c
	default:
		return fmt.Errorf("不支持的群发消息类型: %T", content)
	}
	return nil
}

// massMsgType 返回内容体对应的 msgtype 取值
//
// 说明:
//   - MassImagesContent 的内容键为 images，但 msgtype 官方仍要求为 image，
//     是唯一键名与 msgtype 不一致的类型
func massMsgType(content MassContent) string {
	if _, ok := derefMassContent(content).(MassImagesContent); ok {
		return "image"
	}
	return content.ContentKey()
}

// MassSendRequest 根据OpenID群发消息请求
type MassSendRequest struct {
	MassContentFields
	Touser      []string `json:"touser"`
	MsgType     string   `json:"msgtype"`
	ClientMsgID string   `json:"clientmsgid,omitempty"` // 开发者侧群发消息 id（长度限制 32 字节，用于避免重复推送；不填则后台以群发范围和内容摘要值作为 clientmsgid）
}

// NewMassSendRequest 构造按 OpenID 群发的请求体
func NewMassSendRequest(touser []string, content MassContent) (*MassSendRequest, error) {
	if len(touser) == 0 {
		return nil, fmt.Errorf("touser不能为空")
	}
	req := &MassSendRequest{Touser: touser}
	if err := req.setContent(content); err != nil {
		return nil, err
	}
	req.MsgType = massMsgType(content)
	return req, nil
}

// MassSendResponse 根据OpenID群发消息响应
type MassSendResponse struct {
	core.APIResponse
	MsgID int64 `json:"msg_id"`
}

// PreviewRequest 预览消息请求
type PreviewRequest struct {
	MassContentFields
	Touser      string `json:"touser"` // 接收消息用户的 openid（touser、towxname 二选一）
	MsgType     string `json:"msgtype"`
	WxName      string `json:"towxname,omitempty"`    // 接收消息用户的微信号，实现对微信号的预览（touser、towxname 二选一，每日限 100 次）
	MsgDataID   int64  `json:"msg_data_id,omitempty"` // 按消息 id 预览（图文分析数据接口中 msgid 的前半部分）
	ClientMsgID string `json:"clientmsgid,omitempty"` // 开发者侧消息 id（长度限制 32 字节，用于避免重复推送）
}

// NewPreviewRequest 构造预览请求体
func NewPreviewRequest(touser string, content MassContent) (*PreviewRequest, error) {
	if touser == "" {
		return nil, fmt.Errorf("touser不能为空")
	}
	req := &PreviewRequest{Touser: touser}
	if err := req.setContent(content); err != nil {
		return nil, err
	}
	req.MsgType = massMsgType(content)
	return req, nil
}

// PreviewResponse 预览消息响应
type PreviewResponse struct {
	core.APIResponse
	MsgID int64 `json:"msg_id"`
}

// SendAllRequest 根据标签群发消息请求
type SendAllRequest struct {
	MassContentFields
	Filter      Filter `json:"filter"`
	MsgType     string `json:"msgtype"`
	ClientMsgID string `json:"clientmsgid,omitempty"` // 开发者侧群发消息 id（长度限制 32 字节，用于避免重复推送；不填则后台以群发范围和内容摘要值作为 clientmsgid）
}

// NewSendAllRequest 构造按标签群发的请求体
func NewSendAllRequest(filter Filter, content MassContent) (*SendAllRequest, error) {
	req := &SendAllRequest{Filter: filter}
	if err := req.setContent(content); err != nil {
		return nil, err
	}
	req.MsgType = massMsgType(content)
	return req, nil
}

// Filter 筛选条件
//
// 说明:
//   - is_to_all 官方参数表标注 string 但官方示例全部为布尔（文档自身矛盾），
//     采用 FlexBool 双形态兼容：请求侧序列化为布尔（与示例一致），响应侧双形态均可解析
//   - tag_id 官方参数表标注 string 但示例多为数字，保持数字形态（与官方示例多数一致）
type Filter struct {
	IsToAll core.FlexBool `json:"is_to_all"`
	TagID   int           `json:"tag_id,omitempty"`
}

// SendAllResponse 根据标签群发消息响应
type SendAllResponse struct {
	core.APIResponse
	MsgID int64 `json:"msg_id"`
}

// SetSpeedRequest 设置群发速度请求
type SetSpeedRequest struct {
	Speed int `json:"speed"`
}

// SetSpeedResponse 设置群发速度响应
type SetSpeedResponse struct {
	core.APIResponse
}

// UploadNewsMsgRequest 上传图文消息素材请求
type UploadNewsMsgRequest struct {
	Articles []Article `json:"articles"`
}

// Article 图文消息文章
type Article struct {
	Title              string `json:"title"`
	ThumbMediaID       string `json:"thumb_media_id"`
	Author             string `json:"author,omitempty"`
	Digest             string `json:"digest,omitempty"`
	ShowCoverPic       int    `json:"show_cover_pic"`
	Content            string `json:"content"`
	ContentSourceURL   string `json:"content_source_url,omitempty"`
	NeedOpenComment    int    `json:"need_open_comment,omitempty"`
	OnlyFansCanComment int    `json:"only_fans_can_comment,omitempty"`
}

// UploadNewsMsgResponse 上传图文消息素材响应
type UploadNewsMsgResponse struct {
	core.APIResponse
	MediaID string `json:"media_id"`
}

// UploadImage 上传图文消息图片
// 接口文档：https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Batch_Sends_and_Originality_Checks.html
// 功能：上传图文消息图片，获取图片URL，该URL可用于图文消息的图片展示
// 限制：图片大小不超过2MB，支持JPG、PNG格式
// 请求方式：POST multipart/form-data
func (c *MessageClient) UploadImage(ctx context.Context, filename string, imageData []byte) (*UploadImageResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s?access_token=%s", URLUploadImage, accessToken)

	// 创建multipart/form-data请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建文件字段
	part, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return nil, fmt.Errorf("创建文件字段失败: %v", err)
	}

	// 写入文件数据
	if _, err := part.Write(imageData); err != nil {
		return nil, fmt.Errorf("写入文件数据失败: %v", err)
	}

	// 关闭writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("关闭multipart writer失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	resp, err := c.Client.req.MakeRaw(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 非 2xx 状态码提前判错，避免把网关/代理层错误当响应体解析
	if err := core.CheckRawResponseStatus(resp.StatusCode, respBody); err != nil {
		return nil, err
	}

	var result UploadImageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// DeleteMassMsg 删除群发消息
func (c *MessageClient) DeleteMassMsg(ctx context.Context, msgID int64, articleIdx int) (*DeleteMassMsgResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := DeleteMassMsgRequest{
		MsgID:      msgID,
		ArticleIdx: articleIdx,
	}

	var result DeleteMassMsgResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLDeleteMassMsg, accessToken)
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

// GetSpeed 获取群发速度
func (c *MessageClient) GetSpeed(ctx context.Context) (*GetSpeedResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetSpeedResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetSpeed, accessToken)
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
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

// MassMsgGet 获取群发消息发送状态
func (c *MessageClient) MassMsgGet(ctx context.Context, msgID int64) (*MassMsgGetResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := MassMsgGetRequest{
		MsgID: msgID,
	}

	var result MassMsgGetResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLMassMsgGet, accessToken)
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

// MassSend 根据OpenID群发消息
// content 需传入类型化内容体（如 &MassTextContent{Content:"hello"}），
// 内部会按官方格式展开为 {"touser":[...],"msgtype":"text","text":{"content":"hello"}}
func (c *MessageClient) MassSend(ctx context.Context, touser []string, content MassContent) (*MassSendResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request, err := NewMassSendRequest(touser, content)
	if err != nil {
		return nil, err
	}

	var result MassSendResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLMassSend, accessToken)
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

// Preview 预览消息
func (c *MessageClient) Preview(ctx context.Context, touser string, content MassContent) (*PreviewResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request, err := NewPreviewRequest(touser, content)
	if err != nil {
		return nil, err
	}

	var result PreviewResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLPreview, accessToken)
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

// SendAll 根据标签群发消息
func (c *MessageClient) SendAll(ctx context.Context, filter Filter, content MassContent) (*SendAllResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request, err := NewSendAllRequest(filter, content)
	if err != nil {
		return nil, err
	}

	var result SendAllResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLSendAll, accessToken)
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

// SetSpeed 设置群发速度
func (c *MessageClient) SetSpeed(ctx context.Context, speed int) (*SetSpeedResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := SetSpeedRequest{
		Speed: speed,
	}

	var result SetSpeedResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLSetSpeed, accessToken)
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

// UploadNewsMsg 上传图文消息素材
func (c *MessageClient) UploadNewsMsg(ctx context.Context, articles []Article) (*UploadNewsMsgResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := UploadNewsMsgRequest{
		Articles: articles,
	}

	var result UploadNewsMsgResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLUploadNewsMsg, accessToken)
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

// GetCurrentAutoreplyInfoResponse 获取当前自动回复信息响应
type GetCurrentAutoreplyInfoResponse struct {
	core.APIResponse
	IsAddFriendReplyOpen int `json:"is_add_friend_reply_open"` // 关注后自动回复是否开启，0代表未开启，1代表开启
	IsAutoreplyOpen      int `json:"is_autoreply_open"`        // 消息自动回复是否开启，0代表未开启，1代表开启
	AddFriendReplyInfo   struct {
		Type    string `json:"type"`    // 自动回复类型
		Content string `json:"content"` // 对于文本类型，content是文本内容，对于图文、图片、语音、视频，是mediaID
	} `json:"add_friend_autoreply_info"` // 关注后自动回复的信息（官方字段名，原文为 autoreply 而非 reply）
	MessageDefaultReplyInfo struct {
		Type    string `json:"type"`    // 自动回复类型
		Content string `json:"content"` // 对于文本类型，content是文本内容，对于图文、图片、语音、视频，是mediaID
	} `json:"message_default_autoreply_info"` // 消息自动回复的信息（官方字段名，原文为 autoreply 而非 reply）
	KeywordAutoreplyInfo struct {
		List []struct {
			RuleName        string `json:"rule_name"`   // 规则名称
			CreateTime      int64  `json:"create_time"` // 创建时间
			ReplyMode       string `json:"reply_mode"`  // 回复模式，reply_all代表全部回复，random_one代表随机回复一条
			KeywordListInfo []struct {
				Type      string `json:"type"`       // 匹配模式，contain代表消息中含有该关键词即可，equal表示消息内容必须和关键词相同
				MatchMode string `json:"match_mode"` // 匹配模式，contain代表消息中含有该关键词即可，equal表示消息内容必须和关键词相同
				Content   string `json:"content"`    // 关键词内容
			} `json:"keyword_list_info"` // 关键词列表
			ReplyListInfo []struct {
				Type     string `json:"type"`    // 回复类型
				Content  string `json:"content"` // 回复内容，对于文本类型，content是文本内容，对于图文、图片、语音、视频，是mediaID
				NewsInfo struct {
					List []struct {
						Title      string `json:"title"`       // 图文消息标题
						Author     string `json:"author"`      // 作者
						Digest     string `json:"digest"`      // 摘要
						ShowCover  int    `json:"show_cover"`  // 是否显示封面，0为不显示，1为显示
						CoverURL   string `json:"cover_url"`   // 封面图片的URL
						ContentURL string `json:"content_url"` // 正文的URL
						SourceURL  string `json:"source_url"`  // 原文链接，若获取的图文消息无原文链接，则可能无此字段
					} `json:"list"` // 图文消息的文章列表
				} `json:"news_info"` // 图文消息的信息
			} `json:"reply_list_info"` // 回复列表
		} `json:"list"` // 关键词自动回复规则列表
	} `json:"keyword_autoreply_info"` // 关键词自动回复信息
}

// GetCurrentAutoreplyInfo 获取当前自动回复设置
// 接口文档：https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Getting_Rules_for_Auto_Replies.html
// 功能：获取公众号当前使用的自动回复规则，包括关注后自动回复、消息自动回复和关键词自动回复
// 请求方式：GET
func (c *MessageClient) GetCurrentAutoreplyInfo(ctx context.Context) (*GetCurrentAutoreplyInfoResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetCurrentAutoreplyInfoResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetCurrentAutoreplyInfo, accessToken)
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
