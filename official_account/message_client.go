package official_account

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

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
	Speed     int `json:"speed"`
	RealSpeed int `json:"real_speed"`
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

// MassSendRequest 根据OpenID群发消息请求
type MassSendRequest struct {
	Touser  []string    `json:"touser"`
	MsgType string      `json:"msgtype"`
	Content interface{} `json:"content"`
}

// MassSendResponse 根据OpenID群发消息响应
type MassSendResponse struct {
	core.APIResponse
	MsgID int64 `json:"msg_id"`
}

// PreviewRequest 预览消息请求
type PreviewRequest struct {
	Touser  string      `json:"touser"`
	MsgType string      `json:"msgtype"`
	Content interface{} `json:"content"`
}

// PreviewResponse 预览消息响应
type PreviewResponse struct {
	core.APIResponse
	MsgID int64 `json:"msg_id"`
}

// SendAllRequest 根据标签群发消息请求
type SendAllRequest struct {
	Filter  Filter      `json:"filter"`
	MsgType string      `json:"msgtype"`
	Content interface{} `json:"content"`
}

// Filter 筛选条件
type Filter struct {
	IsToAll bool `json:"is_to_all"`
	TagID   int  `json:"tag_id"`
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

	apiURL := fmt.Sprintf("%s?access_token=%s", APIUploadImageURL, url.QueryEscape(accessToken))

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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIDeleteMassMsgURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIGetSpeedURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "GET", apiURL, nil, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIMassMsgGetURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// MassSend 根据OpenID群发消息
func (c *MessageClient) MassSend(ctx context.Context, touser []string, msgType string, content interface{}) (*MassSendResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := MassSendRequest{
		Touser:  touser,
		MsgType: msgType,
		Content: content,
	}

	var result MassSendResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APIMassSendURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// Preview 预览消息
func (c *MessageClient) Preview(ctx context.Context, touser, msgType string, content interface{}) (*PreviewResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := PreviewRequest{
		Touser:  touser,
		MsgType: msgType,
		Content: content,
	}

	var result PreviewResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APIPreviewURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// SendAll 根据标签群发消息
func (c *MessageClient) SendAll(ctx context.Context, filter Filter, msgType string, content interface{}) (*SendAllResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := SendAllRequest{
		Filter:  filter,
		MsgType: msgType,
		Content: content,
	}

	var result SendAllResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APISendAllURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APISetSpeedURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIUploadNewsMsgURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
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
	} `json:"add_friend_reply_info"` // 关注后自动回复的信息
	MessageDefaultReplyInfo struct {
		Type    string `json:"type"`    // 自动回复类型
		Content string `json:"content"` // 对于文本类型，content是文本内容，对于图文、图片、语音、视频，是mediaID
	} `json:"message_default_reply_info"` // 消息自动回复的信息
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIGetCurrentAutoreplyInfoURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "GET", apiURL, nil, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}
