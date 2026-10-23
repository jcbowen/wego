package official_account

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/jcbowen/wego/core"
)

// MaterialClient 素材管理客户端
type MaterialClient struct {
	Client *Client
}

// NewMaterialClient 创建新的素材管理客户端
func NewMaterialClient(client *Client) *MaterialClient {
	return &MaterialClient{
		Client: client,
	}
}

// MaterialType 素材类型
type MaterialType string

const (
	MaterialTypeImage MaterialType = "image"
	MaterialTypeVoice MaterialType = "voice"
	MaterialTypeVideo MaterialType = "video"
	MaterialTypeThumb MaterialType = "thumb"
	MaterialTypeNews  MaterialType = "news"
)

// UploadMaterialResponse 上传素材响应
type UploadMaterialResponse struct {
	core.APIResponse
	MediaID string `json:"media_id"`
	URL     string `json:"url"`
}

// GetMaterialCountResponse 获取素材总数响应
type GetMaterialCountResponse struct {
	core.APIResponse
	VoiceCount int `json:"voice_count"`
	VideoCount int `json:"video_count"`
	ImageCount int `json:"image_count"`
	NewsCount  int `json:"news_count"`
}

// BatchGetMaterialRequest 批量获取素材请求
type BatchGetMaterialRequest struct {
	Type   MaterialType `json:"type"`
	Offset int          `json:"offset"`
	Count  int          `json:"count"`
}

// BatchGetMaterialResponse 批量获取素材响应
type BatchGetMaterialResponse struct {
	core.APIResponse
	TotalCount int            `json:"total_count"`
	ItemCount  int            `json:"item_count"`
	Item       []MaterialItem `json:"item"`
}

// MaterialItem 素材项
type MaterialItem struct {
	MediaID    string `json:"media_id"`
	Name       string `json:"name"`
	UpdateTime int64  `json:"update_time"`
	URL        string `json:"url"`
}

// NewsMaterialItem 图文素材项
type NewsMaterialItem struct {
	MediaID    string      `json:"media_id"`
	Content    NewsContent `json:"content"`
	UpdateTime int64       `json:"update_time"`
}

// NewsContent 图文内容
type NewsContent struct {
	NewsItem []NewsArticle `json:"news_item"`
}

// NewsArticle 图文文章
type NewsArticle struct {
	Title              string `json:"title"`
	Author             string `json:"author"`
	Digest             string `json:"digest"`
	Content            string `json:"content"`
	ContentSourceURL   string `json:"content_source_url"`
	ThumbMediaID       string `json:"thumb_media_id"`
	ShowCoverPic       int    `json:"show_cover_pic"`
	URL                string `json:"url"`
	ThumbURL           string `json:"thumb_url"`
	NeedOpenComment    int    `json:"need_open_comment"`
	OnlyFansCanComment int    `json:"only_fans_can_comment"`
}

// MaterialUploadImageResponse 素材管理上传图片响应
type MaterialUploadImageResponse struct {
	core.APIResponse
	URL string `json:"url"`
}

// UploadVideoRequest 上传视频素材请求
type UploadVideoRequest struct {
	MediaID     string `json:"media_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UploadVideoResponse 上传视频素材响应
type UploadVideoResponse struct {
	core.APIResponse
}

// DeleteMaterialRequest 删除素材请求
type DeleteMaterialRequest struct {
	MediaID string `json:"media_id"`
}

// DeleteMaterialResponse 删除素材响应
type DeleteMaterialResponse struct {
	core.APIResponse
}

// UploadMaterial 上传临时素材
// 接口文档：https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/New_temporary_materials.html
// 功能：上传临时素材，获取media_id，该media_id可用于消息发送
// 限制：图片大小不超过2MB，语音大小不超过2MB，视频大小不超过10MB，缩略图大小不超过64KB
// 请求方式：POST multipart/form-data
func (c *MaterialClient) UploadMaterial(ctx context.Context, materialType MaterialType, filename string, data []byte) (*UploadMaterialResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s?access_token=%s&type=%s", URLUploadMaterial, accessToken, materialType)

	// 创建multipart/form-data请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建文件字段
	part, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return nil, fmt.Errorf("创建文件字段失败: %v", err)
	}

	// 写入文件数据
	if _, err := part.Write(data); err != nil {
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

	var result UploadMaterialResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// GetMaterialResponse 获取永久素材响应
//
// 根据素材类型区分两种返回形态：
//   - 图片/语音/缩略图等二进制素材：Data 为原始字节流，ContentType 为响应 Content-Type
//   - 视频素材：微信返回 JSON（title/description/down_url），解析到对应字段
type GetMaterialResponse struct {
	core.APIResponse
	ContentType string `json:"-"`           // 响应 Content-Type
	Data        []byte `json:"-"`           // 二进制素材的原始数据（非视频素材时有效）
	Title       string `json:"title"`       // 视频素材标题（仅视频素材返回）
	Description string `json:"description"` // 视频素材描述（仅视频素材返回）
	DownURL     string `json:"down_url"`    // 视频素材下载地址（仅视频素材返回，官方文档字段）
}

// GetMaterial 获取永久素材
//
// 参数:
//   - ctx: 上下文
//   - mediaID: 素材 media_id
//
// 返回:
//   - *GetMaterialResponse: 二进制素材填充 Data/ContentType；视频素材填充 Title/Description/DownURL
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/Get_materials.html
func (c *MaterialClient) GetMaterial(ctx context.Context, mediaID string) (*GetMaterialResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetMaterialPermanent, accessToken)

	// 素材接口的响应可能是二进制流，必须走 MakeRaw 并按 Content-Type 区分处理
	requestBody, err := json.Marshal(map[string]string{"media_id": mediaID})
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.req.MakeRaw(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 非 2xx 状态码提前判错，避免把网关/代理层错误当素材数据返回
	if err := core.CheckRawResponseStatus(resp.StatusCode, respBody); err != nil {
		return nil, err
	}

	contentType := resp.Header.Get("Content-Type")
	// JSON 响应：要么是错误信息，要么是视频素材的元数据
	if strings.Contains(contentType, "application/json") {
		var result GetMaterialResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("解析响应失败: %v", err)
		}
		if !result.IsSuccess() {
			return nil, &result.APIResponse
		}
		result.ContentType = contentType
		return &result, nil
	}

	// 二进制素材：直接返回原始数据
	return &GetMaterialResponse{
		APIResponse: core.APIResponse{ErrCode: 0, ErrMsg: "success"},
		ContentType: contentType,
		Data:        respBody,
	}, nil
}

// DeleteMaterial 删除永久素材
func (c *MaterialClient) DeleteMaterial(ctx context.Context, mediaID string) (*DeleteMaterialResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := DeleteMaterialRequest{
		MediaID: mediaID,
	}

	var result DeleteMaterialResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLDeleteMaterial, accessToken)
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

// GetMaterialCount 获取素材总数
func (c *MaterialClient) GetMaterialCount(ctx context.Context) (*GetMaterialCountResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetMaterialCountResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetMaterialCount, accessToken)
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

// BatchGetMaterial 批量获取素材列表
func (c *MaterialClient) BatchGetMaterial(ctx context.Context, materialType MaterialType, offset, count int) (*BatchGetMaterialResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := BatchGetMaterialRequest{
		Type:   materialType,
		Offset: offset,
		Count:  count,
	}

	var result BatchGetMaterialResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLBatchGetMaterial, accessToken)
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

// UploadImage 上传图文消息内的图片获取URL
// 接口文档：https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/Adding_Permanent_Assets.html
// 功能：上传图文消息内的图片获取URL，该URL可用于图文消息的图片展示
// 限制：图片大小不超过2MB，支持JPG、PNG格式
// 请求方式：POST multipart/form-data
func (c *MaterialClient) UploadImage(ctx context.Context, filename string, data []byte) (*MaterialUploadImageResponse, error) {
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
	if _, err := part.Write(data); err != nil {
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

	var result MaterialUploadImageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// AddMaterialRequest 新增永久素材请求
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/Adding_Permanent_Assets.html
// 接口说明：新增永久素材，支持图片、语音、缩略图类型
// 请求方式：POST multipart/form-data
// 注意事项：图片大小不超过2MB，语音大小不超过2MB，缩略图大小不超过64KB
type AddMaterialRequest struct {
	MediaID     string `json:"media_id,omitempty"`    // 临时素材media_id（用于视频素材）
	Title       string `json:"title,omitempty"`       // 视频素材标题
	Description string `json:"description,omitempty"` // 视频素材描述
}

// AddMaterialResponse 新增永久素材响应
type AddMaterialResponse struct {
	core.APIResponse
	MediaID string `json:"media_id"` // 新增的永久素材media_id
	URL     string `json:"url"`      // 新增的图片素材URL（仅图片素材返回）
}

// AddMaterial 新增永久素材
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/Adding_Permanent_Assets.html
// 接口说明：新增永久素材，支持图片、语音、缩略图类型
// 注意事项：
//   - 图片大小不超过2MB，支持JPG、PNG格式
//   - 语音大小不超过2MB，支持AMR、MP3格式
//   - 缩略图大小不超过64KB，支持JPG格式
//   - 视频素材需要通过UploadMaterial先上传临时素材，再调用UploadVideo转换为永久素材
func (c *MaterialClient) AddMaterial(ctx context.Context, materialType MaterialType, filename string, data []byte) (*AddMaterialResponse, error) {
	if materialType == MaterialTypeVideo {
		return nil, fmt.Errorf("视频素材请使用 AddMaterialVideo 方法（官方要求额外携带 description 表单字段）")
	}

	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s?access_token=%s&type=%s", URLUploadVideo, accessToken, materialType)

	// 创建multipart/form-data请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建文件字段
	part, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return nil, fmt.Errorf("创建文件字段失败: %v", err)
	}

	// 写入文件数据
	if _, err := part.Write(data); err != nil {
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

	var result AddMaterialResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// AddMaterialVideo 新增永久视频素材
//
// 官方文档：https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/Adding_Permanent_Assets.html
// 接口说明：新增永久视频素材（type=video），视频最大不超过 20M
// 请求方式：POST multipart/form-data
// 注意事项：
//   - 视频素材除 media 文件字段外，必须额外携带 description 表单字段
//   - description 值为 JSON 字符串：{"title":TITLE,"introduction":INTRODUCTION}（注意是 introduction）
//
// 参数:
//   - ctx context.Context 请求上下文
//   - filename string 视频文件名（如 demo.mp4）
//   - data []byte 视频文件内容
//   - title string 视频标题
//   - introduction string 视频介绍
//
// 返回:
//   - *AddMaterialResponse 上传结果（MediaID）
//   - error 请求失败或微信返回业务错误时返回错误
func (c *MaterialClient) AddMaterialVideo(ctx context.Context, filename string, data []byte, title, introduction string) (*AddMaterialResponse, error) {
	if title == "" {
		return nil, fmt.Errorf("title不能为空")
	}
	if introduction == "" {
		return nil, fmt.Errorf("introduction不能为空")
	}

	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s?access_token=%s&type=%s", URLUploadVideo, accessToken, MaterialTypeVideo)

	// 创建multipart/form-data请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// description 表单字段（官方要求 JSON：{"title":..,"introduction":..}）
	description, err := json.Marshal(map[string]string{"title": title, "introduction": introduction})
	if err != nil {
		return nil, fmt.Errorf("序列化description失败: %v", err)
	}
	if err := writer.WriteField("description", string(description)); err != nil {
		return nil, fmt.Errorf("写入description字段失败: %v", err)
	}

	// 创建文件字段
	part, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return nil, fmt.Errorf("创建文件字段失败: %v", err)
	}

	// 写入文件数据
	if _, err := part.Write(data); err != nil {
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

	var result AddMaterialResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// UploadVideo 上传视频素材
func (c *MaterialClient) UploadVideo(ctx context.Context, mediaID, title, description string) (*UploadVideoResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := UploadVideoRequest{
		MediaID:     mediaID,
		Title:       title,
		Description: description,
	}

	var result UploadVideoResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLUploadVideo, accessToken)
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

// GetHDVoiceRequest 获取高清语音请求
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/Get_hd_voice.html
// 接口说明：获取高清语音素材，用于JS-SDK的语音播放功能
// 请求方式：POST application/json
type GetHDVoiceRequest struct {
	MediaID string `json:"media_id"` // 语音素材的media_id
}

// GetHDVoiceResponse 获取高清语音响应
type GetHDVoiceResponse struct {
	core.APIResponse
	Data []byte `json:"-"` // 语音文件数据
}

// GetHDVoice 获取高清语音素材
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Asset_Management/Get_hd_voice.html
// 接口说明：获取高清语音素材，用于JS-SDK的语音播放功能
// 注意事项：
//   - 该接口返回的是语音文件的二进制数据
//   - 需要先通过UploadMaterial上传语音素材获取media_id
//   - 主要用于JS-SDK的wx.downloadVoice和wx.playVoice功能
func (c *MaterialClient) GetHDVoice(ctx context.Context, mediaID string) (*GetHDVoiceResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := GetHDVoiceRequest{
		MediaID: mediaID,
	}

	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetHDVoice, accessToken)

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 序列化请求体
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %v", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(requestBody))

	// 发送请求
	resp, err := c.Client.req.MakeRaw(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应数据
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 非 2xx 状态码提前判错，避免把网关/代理层错误当语音数据返回
	if err := core.CheckRawResponseStatus(resp.StatusCode, respBody); err != nil {
		return nil, err
	}

	// 检查响应类型
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// 如果是JSON响应，说明有错误
		var result GetHDVoiceResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("解析响应失败: %v", err)
		}

		if !result.IsSuccess() {
			return nil, &result.APIResponse
		}

		return &result, nil
	}

	// 如果是二进制数据，直接返回
	result := &GetHDVoiceResponse{
		APIResponse: core.APIResponse{
			ErrCode: 0,
			ErrMsg:  "success",
		},
		Data: respBody,
	}

	return result, nil
}

// DraftArticle 草稿文章
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Add_draft.html
// 接口说明：草稿文章结构，用于新建草稿和获取草稿
// 注意事项：
//   - 图文消息的正文中支持插入图片，图片需先通过素材管理接口上传获取URL
//   - 图文消息的正文中支持插入视频，视频需先通过素材管理接口上传获取media_id
//   - 图文消息的正文中支持插入音频，音频需先通过素材管理接口上传获取media_id
type DraftArticle struct {
	Title              string `json:"title"`                 // 标题
	Author             string `json:"author"`                // 作者
	Digest             string `json:"digest"`                // 图文消息的摘要
	Content            string `json:"content"`               // 图文消息的具体内容，支持HTML标签
	ContentSourceURL   string `json:"content_source_url"`    // 图文消息的原文地址
	ThumbMediaID       string `json:"thumb_media_id"`        // 图文消息的封面图片素材ID
	ShowCoverPic       int    `json:"show_cover_pic"`        // 是否显示封面，0为不显示，1为显示
	NeedOpenComment    int    `json:"need_open_comment"`     // 是否打开评论，0不打开，1打开
	OnlyFansCanComment int    `json:"only_fans_can_comment"` // 是否粉丝才可评论，0所有人可评论，1粉丝才可评论
}

// AddDraftRequest 新建草稿请求
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Add_draft.html
// 接口说明：新建草稿，将图文消息保存到草稿箱
// 请求方式：POST application/json
// 注意事项：
//   - 草稿箱中的草稿不会发布到公众号
//   - 草稿可以后续通过发布接口发布到公众号
//   - 草稿箱最多保存100篇草稿
type AddDraftRequest struct {
	Articles []DraftArticle `json:"articles"` // 图文消息，一个图文消息支持1到8条图文
}

// AddDraftResponse 新建草稿响应
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Add_draft.html
// 接口说明：新建草稿响应，返回草稿的media_id
// 响应字段：
//   - media_id: 草稿的media_id，用于后续操作
type AddDraftResponse struct {
	core.APIResponse
	MediaID string `json:"media_id"` // 草稿的media_id
}

// GetDraftRequest 获取草稿请求
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Get_draft.html
// 接口说明：获取草稿内容
// 请求方式：POST application/json
// 注意事项：
//   - 只能获取草稿箱中的草稿
//   - 草稿的media_id通过新建草稿接口获取
type GetDraftRequest struct {
	MediaID string `json:"media_id"` // 草稿的media_id
}

// GetDraftResponse 获取草稿响应
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Get_draft.html
// 接口说明：获取草稿响应，返回草稿的详细信息
// 响应字段：
//   - news_item: 图文消息列表
type GetDraftResponse struct {
	core.APIResponse
	NewsItem []DraftArticle `json:"news_item"` // 图文消息列表
}

// DeleteDraftRequest 删除草稿请求
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Delete_draft.html
// 接口说明：删除草稿
// 请求方式：POST application/json
// 注意事项：
//   - 只能删除草稿箱中的草稿
//   - 删除后无法恢复
type DeleteDraftRequest struct {
	MediaID string `json:"media_id"` // 草稿的media_id
}

// DeleteDraftResponse 删除草稿响应
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Delete_draft.html
// 接口说明：删除草稿响应
type DeleteDraftResponse struct {
	core.APIResponse
}

// GetDraftCountResponse 获取草稿总数响应
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Get_draft_count.html
// 接口说明：获取草稿箱中的草稿总数
// 响应字段：
//   - total_count: 草稿总数
type GetDraftCountResponse struct {
	core.APIResponse
	TotalCount int `json:"total_count"` // 草稿总数
}

// BatchGetDraftRequest 批量获取草稿列表请求
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Batch_get_draft.html
// 接口说明：批量获取草稿列表
// 请求方式：POST application/json
// 注意事项：
//   - offset: 从全部草稿的该偏移位置开始返回，0表示从第一个草稿返回
//   - count: 返回草稿的数量，取值在1到20之间
//   - no_content: 是否不返回content字段，1表示不返回，0表示返回
type BatchGetDraftRequest struct {
	Offset    int `json:"offset"`     // 偏移位置
	Count     int `json:"count"`      // 返回数量
	NoContent int `json:"no_content"` // 是否不返回content字段
}

// DraftItem 草稿项
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Batch_get_draft.html
// 接口说明：草稿项，用于批量获取草稿列表响应
type DraftItem struct {
	MediaID    string       `json:"media_id"`    // 草稿的media_id
	Content    DraftContent `json:"content"`     // 草稿内容
	UpdateTime int64        `json:"update_time"` // 更新时间
}

// DraftContent 草稿内容
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Batch_get_draft.html
// 接口说明：草稿内容，包含图文消息列表
type DraftContent struct {
	NewsItem []DraftArticle `json:"news_item"` // 图文消息列表
}

// BatchGetDraftResponse 批量获取草稿列表响应
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Batch_get_draft.html
// 接口说明：批量获取草稿列表响应
// 响应字段：
//   - total_count: 草稿总数
//   - item_count: 本次获取的草稿数量
//   - item: 草稿列表
type BatchGetDraftResponse struct {
	core.APIResponse
	TotalCount int         `json:"total_count"` // 草稿总数
	ItemCount  int         `json:"item_count"`  // 本次获取的草稿数量
	Item       []DraftItem `json:"item"`        // 草稿列表
}

// UpdateDraftRequest 修改草稿请求
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Update_draft.html
// 接口说明：修改草稿
// 请求方式：POST application/json
// 注意事项：
//   - 只能修改草稿箱中的草稿
//   - 修改后草稿的media_id不变
type UpdateDraftRequest struct {
	MediaID  string       `json:"media_id"` // 草稿的media_id
	Index    int          `json:"index"`    // 要更新的文章在图文消息中的位置，第一篇为0
	Articles DraftArticle `json:"articles"` // 图文消息
}

// UpdateDraftResponse 修改草稿响应
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Update_draft.html
// 接口说明：修改草稿响应
type UpdateDraftResponse struct {
	core.APIResponse
}

// AddDraft 新建草稿
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Add_draft.html
// 接口说明：新建草稿，将图文消息保存到草稿箱
// 注意事项：
//   - 草稿箱最多保存100篇草稿
//   - 图文消息支持1到8条图文
//   - 草稿不会自动发布到公众号
func (c *MaterialClient) AddDraft(ctx context.Context, articles []DraftArticle) (*AddDraftResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := AddDraftRequest{
		Articles: articles,
	}

	var result AddDraftResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLAddDraft, accessToken)
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

// GetDraft 获取草稿
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Get_draft.html
// 接口说明：获取草稿内容
// 注意事项：
//   - 只能获取草稿箱中的草稿
//   - 返回草稿的详细信息
func (c *MaterialClient) GetDraft(ctx context.Context, mediaID string) (*GetDraftResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := GetDraftRequest{
		MediaID: mediaID,
	}

	var result GetDraftResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetDraft, accessToken)
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

// DeleteDraft 删除草稿
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Delete_draft.html
// 接口说明：删除草稿
// 注意事项：
//   - 只能删除草稿箱中的草稿
//   - 删除后无法恢复
func (c *MaterialClient) DeleteDraft(ctx context.Context, mediaID string) (*DeleteDraftResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := DeleteDraftRequest{
		MediaID: mediaID,
	}

	var result DeleteDraftResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLDeleteDraft, accessToken)
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

// GetDraftCount 获取草稿总数
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Get_draft_count.html
// 接口说明：获取草稿箱中的草稿总数
// 注意事项：
//   - 返回草稿箱中的草稿总数
func (c *MaterialClient) GetDraftCount(ctx context.Context) (*GetDraftCountResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetDraftCountResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLGetDraftCount, accessToken)
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

// BatchGetDraft 批量获取草稿列表
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Batch_get_draft.html
// 接口说明：批量获取草稿列表
// 注意事项：
//   - offset: 从全部草稿的该偏移位置开始返回，0表示从第一个草稿返回
//   - count: 返回草稿的数量，取值在1到20之间
//   - no_content: 是否不返回content字段，1表示不返回，0表示返回
func (c *MaterialClient) BatchGetDraft(ctx context.Context, offset, count, noContent int) (*BatchGetDraftResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := BatchGetDraftRequest{
		Offset:    offset,
		Count:     count,
		NoContent: noContent,
	}

	var result BatchGetDraftResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLBatchGetDraft, accessToken)
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

// UpdateDraft 修改草稿
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/Draft_Box/Update_draft.html
// 接口说明：修改草稿
// 注意事项：
//   - 只能修改草稿箱中的草稿
//   - 修改后草稿的media_id不变
func (c *MaterialClient) UpdateDraft(ctx context.Context, mediaID string, index int, article DraftArticle) (*UpdateDraftResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	request := UpdateDraftRequest{
		MediaID:  mediaID,
		Index:    index,
		Articles: article,
	}

	var result UpdateDraftResponse
	apiURL := fmt.Sprintf("%s?access_token=%s", URLUpdateDraft, accessToken)
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
