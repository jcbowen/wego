package officialaccount

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

// MessageText 文本消息
type MessageText struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

func (m *MessageText) GetMsgType() string {
	return "text"
}

// MessageImage 图片消息
type MessageImage struct {
	MsgType string `json:"msgtype"`
	Image   struct {
		MediaID string `json:"media_id"`
	} `json:"image"`
}

func (m *MessageImage) GetMsgType() string {
	return "image"
}

// MessageVoice 语音消息
type MessageVoice struct {
	MsgType string `json:"msgtype"`
	Voice   struct {
		MediaID string `json:"media_id"`
	} `json:"voice"`
}

func (m *MessageVoice) GetMsgType() string {
	return "voice"
}

// MessageVideo 视频消息
type MessageVideo struct {
	MsgType string `json:"msgtype"`
	Video   struct {
		MediaID      string `json:"media_id"`
		ThumbMediaID string `json:"thumb_media_id"`
		Title        string `json:"title"`
		Description  string `json:"description"`
	} `json:"video"`
}

func (m *MessageVideo) GetMsgType() string {
	return "video"
}

// MessageMusic 音乐消息
type MessageMusic struct {
	MsgType string `json:"msgtype"`
	Music   struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		MusicURL     string `json:"musicurl"`
		HQMusicURL   string `json:"hqmusicurl"`
		ThumbMediaID string `json:"thumb_media_id"`
	} `json:"music"`
}

func (m *MessageMusic) GetMsgType() string {
	return "music"
}

// MessageNews 图文消息
type MessageNews struct {
	MsgType string `json:"msgtype"`
	News    struct {
		Articles []CustomArticle `json:"articles"`
	} `json:"news"`
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
	MsgType string `json:"msgtype"`
	MPNews  struct {
		MediaID string `json:"media_id"`
	} `json:"mpnews"`
}

func (m *MPNewsMessage) GetMsgType() string {
	return "mpnews"
}

// MessageWXCard 卡券消息
type MessageWXCard struct {
	MsgType string `json:"msgtype"`
	WXCard  struct {
		CardID string `json:"card_id"`
	} `json:"wxcard"`
}

func (m *MessageWXCard) GetMsgType() string {
	return "wxcard"
}

// MessageMiniProgramPage 小程序卡片消息
type MessageMiniProgramPage struct {
	MsgType         string `json:"msgtype"`
	MiniProgramPage struct {
		Title        string `json:"title"`
		AppID        string `json:"appid"`
		PagePath     string `json:"pagepath"`
		ThumbMediaID string `json:"thumb_media_id"`
	} `json:"miniprogrampage"`
}

func (m *MessageMiniProgramPage) GetMsgType() string {
	return "miniprogrampage"
}

// SendCustomMessageRequest 发送客服消息请求
type SendCustomMessageRequest struct {
	Touser          string                  `json:"touser"`
	MsgType         string                  `json:"msgtype"`
	Text            *MessageText            `json:"text,omitempty"`
	Image           *MessageImage           `json:"image,omitempty"`
	Voice           *MessageVoice           `json:"voice,omitempty"`
	Video           *MessageVideo           `json:"video,omitempty"`
	Music           *MessageMusic           `json:"music,omitempty"`
	News            *MessageNews            `json:"news,omitempty"`
	MPNews          *MPNewsMessage          `json:"mpnews,omitempty"`
	WXCard          *MessageWXCard          `json:"wxcard,omitempty"`
	MiniProgramPage *MessageMiniProgramPage `json:"miniprogrampage,omitempty"`
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

	// 根据消息类型设置对应的消息字段
	switch msg := message.(type) {
	case *MessageText:
		request.Text = msg
	case *MessageImage:
		request.Image = msg
	case *MessageVoice:
		request.Voice = msg
	case *MessageVideo:
		request.Video = msg
	case *MessageMusic:
		request.Music = msg
	case *MessageNews:
		request.News = msg
	case *MPNewsMessage:
		request.MPNews = msg
	case *MessageWXCard:
		request.WXCard = msg
	case *MessageMiniProgramPage:
		request.MiniProgramPage = msg
	default:
		return nil, fmt.Errorf("unsupported message type: %s", message.GetMsgType())
	}

	var result SendCustomMessageResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", APIMessageCustomSendURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIAddCustomAccountURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIUpdateCustomAccountURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIDeleteCustomAccountURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APISetCustomAccountHeadImgURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIGetAllCustomAccountsURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "GET", apiURL, nil, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIGetOnlineCustomAccountsURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "GET", apiURL, nil, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APICreateCustomSessionURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APICloseCustomSessionURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIGetCustomSessionURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIGetCustomSessionListURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIGetWaitCaseURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "GET", apiURL, nil, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APITypingURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
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
	apiURL := fmt.Sprintf("%s?access_token=%s", APIGetMsgRecordURL, url.QueryEscape(accessToken))
	err = c.Client.req.Make(ctx, "POST", apiURL, request, &result)
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}
