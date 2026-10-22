package official_account

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jcbowen/wego/core"
)

// CustomClient 客服消息客户端
type CustomClient struct {
	Client *Client
}

// NewCustomClient 创建新的客服消息客户端
func NewCustomClient(client *Client) *CustomClient {
	return &CustomClient{
		Client: client,
	}
}

// CustomMessage 客服消息接口
type CustomMessage interface {
	GetMsgType() string
}

// 以下为客服消息的「内容体」结构，用于 SendCustomMessageRequest 中与 msgtype 平级组装。
//
// 官方客服消息请求体格式（参考 https://developers.weixin.qq.com/doc/service/api/message/sendmsg/api_customsend.html）：
//
//	{"touser":"OPENID","msgtype":"text","text":{"content":"hello"}}
//
// 即内容体中**不再**包含 msgtype，msgtype 只在最外层出现一次。
type (
	// CustomTextContent 文本消息内容体
	CustomTextContent struct {
		Content string `json:"content"`
	}
	// CustomImageContent 图片消息内容体
	CustomImageContent struct {
		MediaID string `json:"media_id"`
	}
	// CustomVoiceContent 语音消息内容体
	CustomVoiceContent struct {
		MediaID string `json:"media_id"`
	}
	// CustomVideoContent 视频消息内容体
	CustomVideoContent struct {
		MediaID      string `json:"media_id"`
		ThumbMediaID string `json:"thumb_media_id"`
		Title        string `json:"title,omitempty"`
		Description  string `json:"description,omitempty"`
	}
	// CustomMusicContent 音乐消息内容体
	CustomMusicContent struct {
		Title        string `json:"title,omitempty"`
		Description  string `json:"description,omitempty"`
		MusicURL     string `json:"musicurl"`
		HQMusicURL   string `json:"hqmusicurl"`
		ThumbMediaID string `json:"thumb_media_id"`
	}
	// CustomNewsContent 图文消息内容体
	CustomNewsContent struct {
		Articles []CustomArticle `json:"articles"`
	}
	// CustomMPNewsContent 永久图文消息内容体
	CustomMPNewsContent struct {
		MediaID string `json:"media_id"`
	}
	// CustomWXCardContent 卡券消息内容体
	CustomWXCardContent struct {
		CardID string `json:"card_id"`
	}
	// CustomMiniProgramPageContent 小程序卡片消息内容体
	CustomMiniProgramPageContent struct {
		Title        string `json:"title"`
		AppID        string `json:"appid"`
		PagePath     string `json:"pagepath"`
		ThumbMediaID string `json:"thumb_media_id"`
	}
)

// MessageText 文本消息
type MessageText struct {
	MsgType string            `json:"msgtype"`
	Text    CustomTextContent `json:"text"`
}

func (m *MessageText) GetMsgType() string {
	return "text"
}

// MessageImage 图片消息
type MessageImage struct {
	MsgType string             `json:"msgtype"`
	Image   CustomImageContent `json:"image"`
}

func (m *MessageImage) GetMsgType() string {
	return "image"
}

// MessageVoice 语音消息
type MessageVoice struct {
	MsgType string             `json:"msgtype"`
	Voice   CustomVoiceContent `json:"voice"`
}

func (m *MessageVoice) GetMsgType() string {
	return "voice"
}

// MessageVideo 视频消息
type MessageVideo struct {
	MsgType string             `json:"msgtype"`
	Video   CustomVideoContent `json:"video"`
}

func (m *MessageVideo) GetMsgType() string {
	return "video"
}

// MessageMusic 音乐消息
type MessageMusic struct {
	MsgType string             `json:"msgtype"`
	Music   CustomMusicContent `json:"music"`
}

func (m *MessageMusic) GetMsgType() string {
	return "music"
}

// MessageNews 图文消息
type MessageNews struct {
	MsgType string            `json:"msgtype"`
	News    CustomNewsContent `json:"news"`
}

func (m *MessageNews) GetMsgType() string {
	return "news"
}

// CustomArticle 客服图文消息文章
type CustomArticle struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	PicURL      string `json:"picurl"`
}

// MPNewsMessage 图文消息（发送时使用media_id）
type MPNewsMessage struct {
	MsgType string              `json:"msgtype"`
	MPNews  CustomMPNewsContent `json:"mpnews"`
}

func (m *MPNewsMessage) GetMsgType() string {
	return "mpnews"
}

// MessageWXCard 卡券消息
type MessageWXCard struct {
	MsgType string              `json:"msgtype"`
	WXCard  CustomWXCardContent `json:"wxcard"`
}

func (m *MessageWXCard) GetMsgType() string {
	return "wxcard"
}

// MessageMiniProgramPage 小程序卡片消息
type MessageMiniProgramPage struct {
	MsgType         string                       `json:"msgtype"`
	MiniProgramPage CustomMiniProgramPageContent `json:"miniprogrampage"`
}

func (m *MessageMiniProgramPage) GetMsgType() string {
	return "miniprogrampage"
}

// SendCustomMessageRequest 发送客服消息请求
//
// 结构必须与官方一致：msgtype 只在最外层出现，各类型字段承载**纯内容体**。
type SendCustomMessageRequest struct {
	Touser          string                        `json:"touser"`
	MsgType         string                        `json:"msgtype"`
	Text            *CustomTextContent            `json:"text,omitempty"`
	Image           *CustomImageContent           `json:"image,omitempty"`
	Voice           *CustomVoiceContent           `json:"voice,omitempty"`
	Video           *CustomVideoContent           `json:"video,omitempty"`
	Music           *CustomMusicContent           `json:"music,omitempty"`
	News            *CustomNewsContent            `json:"news,omitempty"`
	MPNews          *CustomMPNewsContent          `json:"mpnews,omitempty"`
	WXCard          *CustomWXCardContent          `json:"wxcard,omitempty"`
	MiniProgramPage *CustomMiniProgramPageContent `json:"miniprogrampage,omitempty"`
}

// SendCustomMessageResponse 发送客服消息响应
type SendCustomMessageResponse struct {
	core.APIResponse
}

// CustomAccount 客服账号
type CustomAccount struct {
	KfAccount    string `json:"kf_account"`
	KfNick       string `json:"kf_nick"`
	KfID         string `json:"kf_id"`
	KfHeadImgURL string `json:"kf_headimgurl"`
}

// AddCustomAccountRequest 添加客服账号请求
type AddCustomAccountRequest struct {
	KfAccount string `json:"kf_account"`
	Nickname  string `json:"nickname"`
}

// AddCustomAccountResponse 添加客服账号响应
type AddCustomAccountResponse struct {
	core.APIResponse
}

// UpdateCustomAccountRequest 修改客服账号请求
type UpdateCustomAccountRequest struct {
	KfAccount string `json:"kf_account"`
	Nickname  string `json:"nickname"`
}

// UpdateCustomAccountResponse 修改客服账号响应
type UpdateCustomAccountResponse struct {
	core.APIResponse
}

// DeleteCustomAccountRequest 删除客服账号请求
type DeleteCustomAccountRequest struct {
	KfAccount string `json:"kf_account"`
}

// DeleteCustomAccountResponse 删除客服账号响应
type DeleteCustomAccountResponse struct {
	core.APIResponse
}

// SetCustomAccountHeadImgRequest 设置客服账号头像请求
type SetCustomAccountHeadImgRequest struct {
	KfAccount string `json:"kf_account"`
}

// SetCustomAccountHeadImgResponse 设置客服账号头像响应
type SetCustomAccountHeadImgResponse struct {
	core.APIResponse
}

// GetAllCustomAccountsResponse 获取所有客服账号响应
type GetAllCustomAccountsResponse struct {
	core.APIResponse
	KfList []CustomAccount `json:"kf_list"`
}

// GetOnlineCustomAccountsResponse 获取在线客服接待信息响应
type GetOnlineCustomAccountsResponse struct {
	core.APIResponse
	KfOnlineList []OnlineCustomAccount `json:"kf_online_list"`
}

// OnlineCustomAccount 在线客服账号
type OnlineCustomAccount struct {
	KfAccount    string `json:"kf_account"`
	Status       int    `json:"status"`
	KfID         string `json:"kf_id"`
	AcceptedCase int    `json:"accepted_case"`
}

// CreateCustomSessionRequest 创建客服会话请求
type CreateCustomSessionRequest struct {
	KfAccount string `json:"kf_account"`
	OpenID    string `json:"openid"`
}

// CreateCustomSessionResponse 创建客服会话响应
type CreateCustomSessionResponse struct {
	core.APIResponse
}

// CloseCustomSessionRequest 关闭客服会话请求
type CloseCustomSessionRequest struct {
	KfAccount string `json:"kf_account"`
	OpenID    string `json:"openid"`
}

// CloseCustomSessionResponse 关闭客服会话响应
type CloseCustomSessionResponse struct {
	core.APIResponse
}

// GetCustomSessionRequest 获取客服会话请求
type GetCustomSessionRequest struct {
	OpenID string `json:"openid"`
}

// GetCustomSessionResponse 获取客服会话响应
type GetCustomSessionResponse struct {
	core.APIResponse
	KfAccount  string `json:"kf_account"`
	Createtime int64  `json:"createtime"`
}

// GetCustomSessionListRequest 获取客服会话列表请求
type GetCustomSessionListRequest struct {
	KfAccount string `json:"kf_account"`
}

// GetCustomSessionListResponse 获取客服会话列表响应
type GetCustomSessionListResponse struct {
	core.APIResponse
	SessionList []CustomSession `json:"sessionlist"`
}

// CustomSession 客服会话
type CustomSession struct {
	OpenID     string `json:"openid"`
	Createtime int64  `json:"createtime"`
}

// GetWaitCaseResponse 获取未接入会话列表响应
type GetWaitCaseResponse struct {
	core.APIResponse
	Count        int            `json:"count"`
	WaitCaseList []WaitCaseInfo `json:"waitcaselist"`
}

// WaitCaseInfo 未接入会话信息
type WaitCaseInfo struct {
	LatestTime int64  `json:"latest_time"`
	OpenID     string `json:"openid"`
}

// GetMsgRecordRequest 获取聊天记录请求
type GetMsgRecordRequest struct {
	StartTime int64 `json:"starttime"`
	EndTime   int64 `json:"endtime"`
	MsgID     int64 `json:"msgid"`
	Number    int   `json:"number"`
}

// GetMsgRecordResponse 获取聊天记录响应
type GetMsgRecordResponse struct {
	core.APIResponse
	RecordList []MsgRecord `json:"recordlist"`
}

// MsgRecord 聊天记录
type MsgRecord struct {
	OpenID   string `json:"openid"`
	OperCode int    `json:"opercode"`
	Text     string `json:"text"`
	Time     int64  `json:"time"`
	Worker   string `json:"worker"`
}

// SendCustomMessage 发送客服消息
func (c *CustomClient) SendCustomMessage(ctx context.Context, touser string, message CustomMessage) (*SendCustomMessageResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := SendCustomMessageRequest{
		Touser:  touser,
		MsgType: message.GetMsgType(),
	}

	// 根据消息类型设置对应的消息内容体（只取内容，不再嵌套 msgtype）
	switch msg := message.(type) {
	case *MessageText:
		request.Text = &msg.Text
	case *MessageImage:
		request.Image = &msg.Image
	case *MessageVoice:
		request.Voice = &msg.Voice
	case *MessageVideo:
		request.Video = &msg.Video
	case *MessageMusic:
		request.Music = &msg.Music
	case *MessageNews:
		request.News = &msg.News
	case *MPNewsMessage:
		request.MPNews = &msg.MPNews
	case *MessageWXCard:
		request.WXCard = &msg.WXCard
	case *MessageMiniProgramPage:
		request.MiniProgramPage = &msg.MiniProgramPage
	default:
		return nil, fmt.Errorf("unsupported message type: %s", message.GetMsgType())
	}

	var result SendCustomMessageResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLMessageCustomSend, url.QueryEscape(accessToken))
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

// AddCustomAccount 添加客服账号
func (c *CustomClient) AddCustomAccount(ctx context.Context, kfAccount, nickname string) (*AddCustomAccountResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := AddCustomAccountRequest{
		KfAccount: kfAccount,
		Nickname:  nickname,
	}

	var result AddCustomAccountResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLAddCustomAccount, url.QueryEscape(accessToken))
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

// UpdateCustomAccount 修改客服账号
func (c *CustomClient) UpdateCustomAccount(ctx context.Context, kfAccount, nickname string) (*UpdateCustomAccountResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := UpdateCustomAccountRequest{
		KfAccount: kfAccount,
		Nickname:  nickname,
	}

	var result UpdateCustomAccountResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLUpdateCustomAccount, url.QueryEscape(accessToken))
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

// DeleteCustomAccount 删除客服账号
func (c *CustomClient) DeleteCustomAccount(ctx context.Context, kfAccount string) (*DeleteCustomAccountResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := DeleteCustomAccountRequest{
		KfAccount: kfAccount,
	}

	var result DeleteCustomAccountResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLDeleteCustomAccount, url.QueryEscape(accessToken))
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

// SetCustomAccountHeadImg 设置客服账号头像
func (c *CustomClient) SetCustomAccountHeadImg(ctx context.Context, kfAccount string) (*SetCustomAccountHeadImgResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := SetCustomAccountHeadImgRequest{
		KfAccount: kfAccount,
	}

	var result SetCustomAccountHeadImgResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLSetCustomAccountHeadImg, url.QueryEscape(accessToken))
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

// GetAllCustomAccounts 获取所有客服账号
func (c *CustomClient) GetAllCustomAccounts(ctx context.Context) (*GetAllCustomAccountsResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetAllCustomAccountsResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetAllCustomAccounts, url.QueryEscape(accessToken))
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

// GetOnlineCustomAccounts 获取在线客服接待信息
func (c *CustomClient) GetOnlineCustomAccounts(ctx context.Context) (*GetOnlineCustomAccountsResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetOnlineCustomAccountsResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetOnlineCustomAccounts, url.QueryEscape(accessToken))
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

// CreateCustomSession 创建客服会话
func (c *CustomClient) CreateCustomSession(ctx context.Context, kfAccount, openid string) (*CreateCustomSessionResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := CreateCustomSessionRequest{
		KfAccount: kfAccount,
		OpenID:    openid,
	}

	var result CreateCustomSessionResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLCreateCustomSession, url.QueryEscape(accessToken))
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

// CloseCustomSession 关闭客服会话
func (c *CustomClient) CloseCustomSession(ctx context.Context, kfAccount, openid string) (*CloseCustomSessionResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := CloseCustomSessionRequest{
		KfAccount: kfAccount,
		OpenID:    openid,
	}

	var result CloseCustomSessionResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLCloseCustomSession, url.QueryEscape(accessToken))
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

// GetCustomSession 获取客服会话
func (c *CustomClient) GetCustomSession(ctx context.Context, openid string) (*GetCustomSessionResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := GetCustomSessionRequest{
		OpenID: openid,
	}

	var result GetCustomSessionResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetCustomSession, url.QueryEscape(accessToken))
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

// GetCustomSessionList 获取客服会话列表
func (c *CustomClient) GetCustomSessionList(ctx context.Context, kfAccount string) (*GetCustomSessionListResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := GetCustomSessionListRequest{
		KfAccount: kfAccount,
	}

	var result GetCustomSessionListResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetCustomSessionList, url.QueryEscape(accessToken))
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

// GetWaitCase 获取未接入会话列表
func (c *CustomClient) GetWaitCase(ctx context.Context) (*GetWaitCaseResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetWaitCaseResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetWaitCase, url.QueryEscape(accessToken))
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

// TypingRequest 客服输入状态请求
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Service_Center_messages.html
// 接口说明：控制客服输入状态，让用户看到客服"正在输入"的状态
// 请求方式：POST https://api.weixin.qq.com/cgi-bin/message/custom/typing?access_token=ACCESS_TOKEN
// 请求体：{"touser":"OPENID","command":"Typing"}
type TypingRequest struct {
	ToUser  string `json:"touser"`  // 用户OpenID
	Command string `json:"command"` // 命令："Typing"表示正在输入，"CancelTyping"表示取消输入
}

// TypingResponse 客服输入状态响应
type TypingResponse struct {
	core.APIResponse
}

// Typing 控制客服输入状态
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Service_Center_messages.html
// 接口说明：控制客服输入状态，让用户看到客服"正在输入"的状态
// 注意事项：此接口需要客服账号已绑定且在线
func (c *CustomClient) Typing(ctx context.Context, toUser, command string) (*TypingResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := TypingRequest{
		ToUser:  toUser,
		Command: command,
	}

	var result TypingResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLTyping, url.QueryEscape(accessToken))
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

// TypingStart 开始客服输入状态
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Service_Center_messages.html
// 接口说明：让用户看到客服"正在输入"的状态
// 注意事项：此接口需要客服账号已绑定且在线
func (c *CustomClient) TypingStart(ctx context.Context, toUser string) (*TypingResponse, error) {
	return c.Typing(ctx, toUser, "Typing")
}

// TypingCancel 取消客服输入状态
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Service_Center_messages.html
// 接口说明：取消客服"正在输入"的状态
// 注意事项：此接口需要客服账号已绑定且在线
func (c *CustomClient) TypingCancel(ctx context.Context, toUser string) (*TypingResponse, error) {
	return c.Typing(ctx, toUser, "CancelTyping")
}

// GetMsgRecord 获取聊天记录
func (c *CustomClient) GetMsgRecord(ctx context.Context, startTime, endTime int64, msgID int64, number int) (*GetMsgRecordResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := GetMsgRecordRequest{
		StartTime: startTime,
		EndTime:   endTime,
		MsgID:     msgID,
		Number:    number,
	}

	var result GetMsgRecordResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetMsgRecord, url.QueryEscape(accessToken))
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
