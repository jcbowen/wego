// Package official_account 微信公众号API - 数据统计分析（datacube）
// 参考：https://developers.weixin.qq.com/doc/offiaccount/Analytics/Menu_Analysis.html
package official_account

import (
	"context"
	"fmt"
	"time"

	"github.com/jcbowen/wego/core"
)

// 请求体（各数据统计接口统一为 begin_date/end_date）
type dataCubeDateRangeRequest struct {
	BeginDate string `json:"begin_date"` // 获取数据的起始日期（yyyy-MM-dd）
	EndDate   string `json:"end_date"`   // 获取数据的结束日期（yyyy-MM-dd）
}

// UserSummaryItem 用户增长统计项（按分布渠道拆分）
//
// 注意：官方新版文档的 user_source 渠道枚举与旧版差异较大，以本结构注释为准。
type UserSummaryItem struct {
	RefDate    string `json:"ref_date"`    // 数据日期（yyyy-MM-dd）
	UserSource int    `json:"user_source"` // 用户来源渠道（0其他合计/1公众号搜索/17名片分享/30扫描二维码/57文章内账号名称/100微信广告/161他人转载/149小程序关注/200视频号/201直播）
	NewUser    int    `json:"new_user"`    // 新增的用户数量
	CancelUser int    `json:"cancel_user"` // 取消关注的用户数量
}

// GetUserSummaryResponse 用户增长统计响应（getusersummary，最大时间跨度 7 天）
type GetUserSummaryResponse struct {
	core.APIResponse
	List []UserSummaryItem `json:"list"`
}

// UserCumulateItem 用户累计统计项
type UserCumulateItem struct {
	RefDate      string `json:"ref_date"`      // 数据日期（yyyy-MM-dd）
	CumulateUser int    `json:"cumulate_user"` // 总用户量
}

// GetUserCumulateResponse 用户累计统计响应（getusercumulate，最大时间跨度 7 天）
type GetUserCumulateResponse struct {
	core.APIResponse
	List []UserCumulateItem `json:"list"`
}

// ArticleSummaryItem 图文分析-图文群发每日数据项
type ArticleSummaryItem struct {
	RefDate          string `json:"ref_date"`            // 数据日期
	MsgID            string `json:"msgid"`               // 图文消息的 id（如 "12003_3"）
	Title            string `json:"title"`               // 图文消息的标题
	IntPageReadUser  int    `json:"int_page_read_user"`  // 图文页的到达人数
	IntPageReadCount int    `json:"int_page_read_count"` // 图文页的到达次数
	OriPageReadUser  int    `json:"ori_page_read_user"`  // 原文页的到达人数
	OriPageReadCount int    `json:"ori_page_read_count"` // 原文页的到达次数
	ShareUser        int    `json:"share_user"`          // 分享的人数
	ShareCount       int    `json:"share_count"`         // 分享的次数
	AddToFavUser     int    `json:"add_to_fav_user"`     // 收藏的人数
	AddToFavCount    int    `json:"add_to_fav_count"`    // 收藏的次数
}

// GetArticleSummaryResponse 图文群发每日数据响应（getarticlesummary，最大时间跨度 1 天，只查询近一天）
//
// 注意：官方已标注本接口停止维护，新接口为获取发表内容每日阅读/分享数据（api_getarticleread/api_getarticleshare）。
type GetArticleSummaryResponse struct {
	core.APIResponse
	List []ArticleSummaryItem `json:"list"`
}

// ArticleTotalDetailItem 图文统计-单篇图文按渠道拆分数据项
type ArticleTotalDetailItem struct {
	StatDate         string `json:"stat_date"`           // 统计的日期
	TargetUser       int    `json:"target_user"`         // 送达人数（图文群发时触达的用户数）
	IntPageReadUser  int    `json:"int_page_read_user"`  // 图文页的到达人数
	IntPageReadCount int    `json:"int_page_read_count"` // 图文页的到达次数
	OriPageReadUser  int    `json:"ori_page_read_user"`  // 原文页的到达人数
	OriPageReadCount int    `json:"ori_page_read_count"` // 原文页的到达次数
	ShareUser        int    `json:"share_user"`          // 分享的人数
	ShareCount       int    `json:"share_count"`         // 分享的次数
	AddToFavUser     int    `json:"add_to_fav_user"`     // 收藏的人数
	AddToFavCount    int    `json:"add_to_fav_count"`    // 收藏的次数
}

// ArticleTotalItem 图文分析-图文群发总数据项
type ArticleTotalItem struct {
	RefDate string                   `json:"ref_date"` // 数据日期
	MsgID   string                   `json:"msgid"`    // 图文消息的 id（如 "12003_3"）
	Title   string                   `json:"title"`    // 图文消息的标题
	Details []ArticleTotalDetailItem `json:"details"`  // 按渠道拆分的明细
}

// GetArticleTotalResponse 图文群发总数据响应（getarticletotal，最大时间跨度 1 天，只查询近一天）
//
// 注意：官方已标注本接口停止维护，新接口为获取发表内容概况总数据（api_getbizsummary）。
type GetArticleTotalResponse struct {
	core.APIResponse
	List []ArticleTotalItem `json:"list"`
}

// UserReadItem 图文分析-阅读数据项（按日）
type UserReadItem struct {
	RefDate          string `json:"ref_date"`            // 数据日期
	UserSource       int    `json:"user_source"`         // 用户来源渠道（同 UserSummaryItem.UserSource 枚举）
	IntPageReadUser  int    `json:"int_page_read_user"`  // 图文页的到达人数
	IntPageReadCount int    `json:"int_page_read_count"` // 图文页的到达次数
	OriPageReadUser  int    `json:"ori_page_read_user"`  // 原文页的到达人数
	OriPageReadCount int    `json:"ori_page_read_count"` // 原文页的到达次数
	ShareUser        int    `json:"share_user"`          // 分享的人数
	ShareCount       int    `json:"share_count"`         // 分享的次数
	AddToFavUser     int    `json:"add_to_fav_user"`     // 收藏的人数
	AddToFavCount    int    `json:"add_to_fav_count"`    // 收藏的次数
}

// GetUserReadResponse 图文阅读数据响应（getuserread，最大时间跨度 3 天）
type GetUserReadResponse struct {
	core.APIResponse
	List []UserReadItem `json:"list"`
}

// UserReadHourItem 图文分析-阅读数据项（按小时）
type UserReadHourItem struct {
	UserReadItem
	RefHour int `json:"ref_hour"` // 数据的小时（0-23）
}

// GetUserReadHourResponse 图文阅读数据响应（getuserreadhour，最大时间跨度 1 天）
type GetUserReadHourResponse struct {
	core.APIResponse
	List []UserReadHourItem `json:"list"`
}

// UserShareItem 图文分析-分享数据项（按日）
type UserShareItem struct {
	RefDate    string `json:"ref_date"`    // 数据日期
	ShareScene int    `json:"share_scene"` // 分享的场景（1好友转发/2朋友圈/255其他）
	ShareCount int    `json:"share_count"` // 分享的次数
	ShareUser  int    `json:"share_user"`  // 分享的人数
}

// GetUserShareResponse 图文分享数据响应（getusershare，最大时间跨度 7 天）
//
// 注意：官方已标注本接口停止维护，新接口为获取发表内容每日分享数据（api_getarticleshare）。
type GetUserShareResponse struct {
	core.APIResponse
	List []UserShareItem `json:"list"`
}

// UserShareHourItem 图文分析-分享数据项（按小时）
type UserShareHourItem struct {
	UserShareItem
	RefHour int `json:"ref_hour"` // 数据的小时（0-23）
}

// GetUserShareHourResponse 图文分享数据响应（getusersharehour，最大时间跨度 1 天）
type GetUserShareHourResponse struct {
	core.APIResponse
	List []UserShareHourItem `json:"list"`
}

// InterfaceSummaryItem 接口分析项（按日）
type InterfaceSummaryItem struct {
	RefDate       string `json:"ref_date"`        // 数据日期
	CallbackCount int    `json:"callback_count"`  // 发出 callback 请求的总次数
	FailCount     int    `json:"fail_count"`      // 失败次数
	TotalTimeCost int    `json:"total_time_cost"` // 总耗时（毫秒）
	MaxTimeCost   int    `json:"max_time_cost"`   // 最大耗时（毫秒）
}

// GetInterfaceSummaryResponse 接口分析数据响应（getinterfacesummary，最大时间跨度 30 天）
type GetInterfaceSummaryResponse struct {
	core.APIResponse
	List []InterfaceSummaryItem `json:"list"`
}

// InterfaceSummaryHourItem 接口分析项（按小时）
type InterfaceSummaryHourItem struct {
	InterfaceSummaryItem
	RefHour int `json:"ref_hour"` // 数据的小时（0-23）
}

// GetInterfaceSummaryHourResponse 接口分析数据响应（getinterfacesummaryhour，最大时间跨度 1 天）
type GetInterfaceSummaryHourResponse struct {
	core.APIResponse
	List []InterfaceSummaryHourItem `json:"list"`
}

// validateDateRange 校验日期范围参数：格式 yyyy-MM-dd 且跨度不超过 maxDays 天
func validateDateRange(beginDate, endDate string, maxDays int) error {
	begin, err := time.ParseInLocation("2006-01-02", beginDate, time.Local)
	if err != nil {
		return fmt.Errorf("begin_date 格式错误，应为 yyyy-MM-dd: %w", err)
	}
	end, err := time.ParseInLocation("2006-01-02", endDate, time.Local)
	if err != nil {
		return fmt.Errorf("end_date 格式错误，应为 yyyy-MM-dd: %w", err)
	}
	if end.Before(begin) {
		return fmt.Errorf("end_date 不能早于 begin_date")
	}
	if days := int(end.Sub(begin).Hours() / 24); days >= maxDays {
		return fmt.Errorf("查询跨度不能超过 %d 天（当前 %d 天）", maxDays, days+1)
	}
	return nil
}

// postDateRange 发送 begin_date/end_date 型数据统计请求的公共实现
func (c *Client) postDateRange(ctx context.Context, apiURL, beginDate, endDate string, result interface{}) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    apiURL,
		Query:  map[string]string{"access_token": token},
		Body: dataCubeDateRangeRequest{
			BeginDate: beginDate,
			EndDate:   endDate,
		},
		Result: result,
	})
	if err != nil {
		return err
	}
	// 错误码由调用方携带具体 APIResponse 返回
	return nil
}

// GetUserSummary 获取用户增长统计（按渠道拆分，最大时间跨度 7 天）
//
// 参数:
//   - ctx: 上下文
//   - beginDate: 起始日期（yyyy-MM-dd）
//   - endDate: 结束日期（yyyy-MM-dd）
//
// 返回:
//   - *GetUserSummaryResponse: 按渠道拆分的用户增长数据
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/Analytics/User_Analysis_Data_Interface.html
func (c *Client) GetUserSummary(ctx context.Context, beginDate, endDate string) (*GetUserSummaryResponse, error) {
	if err := validateDateRange(beginDate, endDate, 7); err != nil {
		return nil, err
	}

	var result GetUserSummaryResponse
	if err := c.postDateRange(ctx, URLDataCubeGetUserSummary, beginDate, endDate, &result); err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}
	return &result, nil
}

// GetUserCumulate 获取累计用户统计（最大时间跨度 7 天）
//
// 参数:
//   - ctx: 上下文
//   - beginDate: 起始日期（yyyy-MM-dd）
//   - endDate: 结束日期（yyyy-MM-dd）
//
// 返回:
//   - *GetUserCumulateResponse: 每日累计用户数据
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/Analytics/User_Analysis_Data_Interface.html
func (c *Client) GetUserCumulate(ctx context.Context, beginDate, endDate string) (*GetUserCumulateResponse, error) {
	if err := validateDateRange(beginDate, endDate, 7); err != nil {
		return nil, err
	}

	var result GetUserCumulateResponse
	if err := c.postDateRange(ctx, URLDataCubeGetUserCumulate, beginDate, endDate, &result); err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}
	return &result, nil
}

// GetArticleSummary 获取图文群发每日数据（最大时间跨度 1 天，只查询近一天）
//
// 参数:
//   - ctx: 上下文
//   - beginDate: 起始日期（yyyy-MM-dd）
//   - endDate: 结束日期（yyyy-MM-dd）
//
// 返回:
//   - *GetArticleSummaryResponse: 图文群发每日数据
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/Analytics/Article_Analysis_Data_Interface.html
func (c *Client) GetArticleSummary(ctx context.Context, beginDate, endDate string) (*GetArticleSummaryResponse, error) {
	if err := validateDateRange(beginDate, endDate, 1); err != nil {
		return nil, err
	}

	var result GetArticleSummaryResponse
	if err := c.postDateRange(ctx, URLDataCubeGetArticleSummary, beginDate, endDate, &result); err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}
	return &result, nil
}

// GetArticleTotal 获取图文群发总数据（最大时间跨度 1 天，只查询近一天）
//
// 参数:
//   - ctx: 上下文
//   - beginDate: 起始日期（yyyy-MM-dd）
//   - endDate: 结束日期（yyyy-MM-dd）
//
// 返回:
//   - *GetArticleTotalResponse: 按渠道拆分的图文总数据
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/Analytics/Article_Analysis_Data_Interface.html
func (c *Client) GetArticleTotal(ctx context.Context, beginDate, endDate string) (*GetArticleTotalResponse, error) {
	if err := validateDateRange(beginDate, endDate, 1); err != nil {
		return nil, err
	}

	var result GetArticleTotalResponse
	if err := c.postDateRange(ctx, URLDataCubeGetArticleTotal, beginDate, endDate, &result); err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}
	return &result, nil
}

// GetUserRead 获取图文阅读数据（按日，最大时间跨度 3 天）
//
// 参数:
//   - ctx: 上下文
//   - beginDate: 起始日期（yyyy-MM-dd）
//   - endDate: 结束日期（yyyy-MM-dd）
//
// 返回:
//   - *GetUserReadResponse: 每日阅读数据
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/Analytics/Article_Analysis_Data_Interface.html
func (c *Client) GetUserRead(ctx context.Context, beginDate, endDate string) (*GetUserReadResponse, error) {
	if err := validateDateRange(beginDate, endDate, 3); err != nil {
		return nil, err
	}

	var result GetUserReadResponse
	if err := c.postDateRange(ctx, URLDataCubeGetUserRead, beginDate, endDate, &result); err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}
	return &result, nil
}

// GetUserReadHour 获取图文阅读数据（按小时，最大时间跨度 1 天）
//
// 参数:
//   - ctx: 上下文
//   - beginDate: 起始日期（yyyy-MM-dd）
//   - endDate: 结束日期（yyyy-MM-dd）
//
// 返回:
//   - *GetUserReadHourResponse: 按小时的阅读数据
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/Analytics/Article_Analysis_Data_Interface.html
func (c *Client) GetUserReadHour(ctx context.Context, beginDate, endDate string) (*GetUserReadHourResponse, error) {
	if err := validateDateRange(beginDate, endDate, 1); err != nil {
		return nil, err
	}

	var result GetUserReadHourResponse
	if err := c.postDateRange(ctx, URLDataCubeGetUserReadHour, beginDate, endDate, &result); err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}
	return &result, nil
}

// GetUserShare 获取图文分享数据（按日，最大时间跨度 7 天）
//
// 参数:
//   - ctx: 上下文
//   - beginDate: 起始日期（yyyy-MM-dd）
//   - endDate: 结束日期（yyyy-MM-dd）
//
// 返回:
//   - *GetUserShareResponse: 每日分享数据
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/Analytics/Article_Analysis_Data_Interface.html
func (c *Client) GetUserShare(ctx context.Context, beginDate, endDate string) (*GetUserShareResponse, error) {
	if err := validateDateRange(beginDate, endDate, 7); err != nil {
		return nil, err
	}

	var result GetUserShareResponse
	if err := c.postDateRange(ctx, URLDataCubeGetUserShare, beginDate, endDate, &result); err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}
	return &result, nil
}

// GetUserShareHour 获取图文分享数据（按小时，最大时间跨度 1 天）
//
// 参数:
//   - ctx: 上下文
//   - beginDate: 起始日期（yyyy-MM-dd）
//   - endDate: 结束日期（yyyy-MM-dd）
//
// 返回:
//   - *GetUserShareHourResponse: 按小时的分享数据
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/Analytics/Article_Analysis_Data_Interface.html
func (c *Client) GetUserShareHour(ctx context.Context, beginDate, endDate string) (*GetUserShareHourResponse, error) {
	if err := validateDateRange(beginDate, endDate, 1); err != nil {
		return nil, err
	}

	var result GetUserShareHourResponse
	if err := c.postDateRange(ctx, URLDataCubeGetUserShareHour, beginDate, endDate, &result); err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}
	return &result, nil
}

// GetInterfaceSummary 获取接口分析数据（按日，最大时间跨度 30 天）
//
// 参数:
//   - ctx: 上下文
//   - beginDate: 起始日期（yyyy-MM-dd）
//   - endDate: 结束日期（yyyy-MM-dd）
//
// 返回:
//   - *GetInterfaceSummaryResponse: 每日接口调用数据
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/Analytics/Interface_Analysis_Data_Interface.html
func (c *Client) GetInterfaceSummary(ctx context.Context, beginDate, endDate string) (*GetInterfaceSummaryResponse, error) {
	if err := validateDateRange(beginDate, endDate, 30); err != nil {
		return nil, err
	}

	var result GetInterfaceSummaryResponse
	if err := c.postDateRange(ctx, URLDataCubeGetInterfaceSummary, beginDate, endDate, &result); err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}
	return &result, nil
}

// GetInterfaceSummaryHour 获取接口分析数据（按小时，最大时间跨度 1 天）
//
// 参数:
//   - ctx: 上下文
//   - beginDate: 起始日期（yyyy-MM-dd）
//   - endDate: 结束日期（yyyy-MM-dd）
//
// 返回:
//   - *GetInterfaceSummaryHourResponse: 按小时的接口调用数据
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/Analytics/Interface_Analysis_Data_Interface.html
func (c *Client) GetInterfaceSummaryHour(ctx context.Context, beginDate, endDate string) (*GetInterfaceSummaryHourResponse, error) {
	if err := validateDateRange(beginDate, endDate, 1); err != nil {
		return nil, err
	}

	var result GetInterfaceSummaryHourResponse
	if err := c.postDateRange(ctx, URLDataCubeGetInterfaceSummaryHour, beginDate, endDate, &result); err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}
	return &result, nil
}
