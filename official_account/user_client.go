package official_account

import (
	"context"
	"fmt"

	"github.com/jcbowen/wego/core"
)

// UserClient 用户管理客户端
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/User_Management/Get_users_basic_information_UnionID.html
type UserClient struct {
	Client *Client
}

// NewUserClient 创建新的用户管理客户端
func NewUserClient(client *Client) *UserClient {
	return &UserClient{
		Client: client,
	}
}

// 用户语言版本
const (
	LangZhCN = "zh_CN" // 简体中文
	LangZhTW = "zh_TW" // 繁体中文
	LangEN   = "en"    // 英文
)

// UserInfo 用户基本信息
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/User_Management/Get_users_basic_information_UnionID.html
type UserInfo struct {
	Subscribe      int    `json:"subscribe"`       // 是否订阅该公众号（0否，1是）
	OpenID         string `json:"openid"`          // 用户的 OpenID
	Nickname       string `json:"nickname"`        // 昵称
	Sex            int    `json:"sex"`             // 性别（1男，2女，0未知）
	Language       string `json:"language"`        // 用户的语言
	City           string `json:"city"`            // 城市
	Province       string `json:"province"`        // 省份
	Country        string `json:"country"`         // 国家
	HeadImageURL   string `json:"headimgurl"`      // 头像 URL
	SubscribeTime  int64  `json:"subscribe_time"`  // 关注时间戳
	UnionID        string `json:"unionid"`         // 开放平台 UnionID（已绑定开放平台账号时返回）
	Remark         string `json:"remark"`          // 备注
	GroupID        int    `json:"groupid"`         // 用户所在分组 ID
	TagIDList      []int  `json:"tagid_list"`      // 用户被打上的标签 ID 列表
	SubscribeScene string `json:"subscribe_scene"` // 关注渠道来源
	QRScene        int    `json:"qr_scene"`        // 二维码扫码场景值
	QRSceneStr     string `json:"qr_scene_str"`    // 二维码扫码场景字符串
}

// GetUserInfoResponse 获取用户基本信息响应（错误时携带 errcode/errmsg）
type GetUserInfoResponse struct {
	core.APIResponse
	UserInfo
}

// BatchGetUserInfoItem 批量获取用户信息的单项请求
type BatchGetUserInfoItem struct {
	OpenID string `json:"openid"` // 用户 OpenID
	Lang   string `json:"lang"`   // 国家地区语言版本（zh_CN/zh_TW/en）
}

// BatchGetUserInfoRequest 批量获取用户信息请求
// 注意：单次请求最多支持 100 个 openid
type BatchGetUserInfoRequest struct {
	UserList []BatchGetUserInfoItem `json:"user_list"`
}

// BatchGetUserInfoResponse 批量获取用户信息响应
type BatchGetUserInfoResponse struct {
	core.APIResponse
	UserInfoList []struct {
		UserInfo
		ErrCode int    `json:"errcode"` // 单个用户的获取结果（该用户不存在等）
		ErrMsg  string `json:"errmsg"`
	} `json:"user_info_list"`
}

// GetUserListResponse 获取关注者列表响应
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/User_Management/Get_users_list.html
type GetUserListResponse struct {
	core.APIResponse
	Total int      `json:"total"` // 关注该公众账号的总用户数
	Count int      `json:"count"` // 拉取的 openid 个数，最大值 10000
	Data  struct { // openid 列表
		OpenID []string `json:"openid"`
	} `json:"data"`
	NextOpenID string `json:"next_openid"` // 拉取列表的最后一个用户的 openid（为空表示已取完）
}

// UpdateUserRemarkRequest 设置用户备注名请求
type UpdateUserRemarkRequest struct {
	OpenID string `json:"openid"` // 用户 OpenID
	Remark string `json:"remark"` // 备注名（长度小于 30 个字符）
}

// UpdateUserRemarkResponse 设置用户备注名响应
type UpdateUserRemarkResponse struct {
	core.APIResponse
}

// UserTag 用户标签
// 参考文档：https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
type UserTag struct {
	ID    int    `json:"id"`              // 标签 ID
	Name  string `json:"name"`            // 标签名
	Count int    `json:"count,omitempty"` // 此标签下粉丝数（仅获取标签列表时返回）
}

// CreateUserTagRequest 创建标签请求（官方接口仅接受 name，不接受 id）
type CreateUserTagRequest struct {
	Tag struct {
		Name string `json:"name"` // 标签名（30 个字符以内）
	} `json:"tag"`
}

// CreateUserTagResponse 创建标签响应
type CreateUserTagResponse struct {
	core.APIResponse
	Tag UserTag `json:"tag"`
}

// GetTagsResponse 获取公众号已创建的标签响应
type GetTagsResponse struct {
	core.APIResponse
	Tags []UserTag `json:"tags"`
}

// UpdateUserTagRequest 更新标签请求
type UpdateUserTagRequest struct {
	Tag UserTag `json:"tag"` // id 为必填，name 为新标签名
}

// UpdateUserTagResponse 更新标签响应
type UpdateUserTagResponse struct {
	core.APIResponse
}

// DeleteUserTagRequest 删除标签请求
type DeleteUserTagRequest struct {
	Tag UserTag `json:"tag"` // 仅 id 生效
}

// DeleteUserTagResponse 删除标签响应
type DeleteUserTagResponse struct {
	core.APIResponse
}

// BatchTagUsersRequest 批量为用户打标签请求
type BatchTagUsersRequest struct {
	OpenIDList []string `json:"openid_list"` // 用户 OpenID 列表（最多 50 个）
	TagID      int      `json:"tagid"`       // 标签 ID
}

// BatchTagUsersResponse 批量为用户打标签响应
type BatchTagUsersResponse struct {
	core.APIResponse
}

// BatchUntagUsersRequest 批量取消用户标签请求
type BatchUntagUsersRequest struct {
	OpenIDList []string `json:"openid_list"` // 用户 OpenID 列表（最多 50 个）
	TagID      int      `json:"tagid"`       // 标签 ID
}

// BatchUntagUsersResponse 批量取消用户标签响应
type BatchUntagUsersResponse struct {
	core.APIResponse
}

// GetUsersByTagRequest 获取标签下粉丝列表请求
type GetUsersByTagRequest struct {
	TagID      int    `json:"tagid"`       // 标签 ID
	NextOpenID string `json:"next_openid"` // 第一个拉取的 openid，不填默认从头开始拉取
}

// GetUsersByTagResponse 获取标签下粉丝列表响应
type GetUsersByTagResponse struct {
	core.APIResponse
	Count int      `json:"count"` // 拉取的 openid 个数，最大值 10000
	Data  struct { // openid 列表
		OpenID []string `json:"openid"`
	} `json:"data"`
	NextOpenID string `json:"next_openid"` // 拉取列表的最后一个用户的 openid
}

// GetTagIDListRequest 获取用户身上的标签列表请求
type GetTagIDListRequest struct {
	OpenID string `json:"openid"` // 用户 OpenID
}

// GetTagIDListResponse 获取用户身上的标签列表响应
type GetTagIDListResponse struct {
	core.APIResponse
	TagIDList []int `json:"tagid_list"` // 用户被打上的标签 ID 列表
}

// GetUserInfo 获取用户基本信息（包括订阅状态）
//
// 参数:
//   - ctx: 上下文
//   - openID: 用户的 OpenID
//   - lang: 返回国家地区语言版本（zh_CN/zh_TW/en），传空默认 zh_CN
//
// 返回:
//   - *UserInfo: 用户基本信息
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/User_Management/Get_users_basic_information_UnionID.html
func (c *UserClient) GetUserInfo(ctx context.Context, openID, lang string) (*UserInfo, error) {
	if lang == "" {
		lang = LangZhCN
	}

	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetUserInfoResponse
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    URLUserInfo,
		Query: map[string]string{
			"access_token": accessToken,
			"openid":       openID,
			"lang":         lang,
		},
		Result: &result,
	})
	if err != nil {
		return nil, err
	}

	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result.UserInfo, nil
}

// BatchGetUsers 批量获取用户基本信息
//
// 参数:
//   - ctx: 上下文
//   - openIDs: 用户 OpenID 列表，单次最多 100 个
//   - lang: 返回国家地区语言版本（zh_CN/zh_TW/en），传空默认 zh_CN
//
// 返回:
//   - *BatchGetUserInfoResponse: 用户信息列表（含单项错误码）
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/User_Management/Get_users_basic_information_UnionID.html
func (c *UserClient) BatchGetUsers(ctx context.Context, openIDs []string, lang string) (*BatchGetUserInfoResponse, error) {
	if len(openIDs) == 0 {
		return nil, fmt.Errorf("openIDs 不能为空")
	}
	if len(openIDs) > 100 {
		return nil, fmt.Errorf("openIDs 单次最多支持 100 个，当前 %d 个", len(openIDs))
	}
	if lang == "" {
		lang = LangZhCN
	}

	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	userList := make([]BatchGetUserInfoItem, 0, len(openIDs))
	for _, openID := range openIDs {
		userList = append(userList, BatchGetUserInfoItem{OpenID: openID, Lang: lang})
	}

	var result BatchGetUserInfoResponse
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLGetUserInfoBatch,
		Query: map[string]string{
			"access_token": accessToken,
		},
		Body:   BatchGetUserInfoRequest{UserList: userList},
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

// GetUserList 获取关注者列表（一次最多拉取 10000 个 openid，可通过 nextOpenID 翻页）
//
// 参数:
//   - ctx: 上下文
//   - nextOpenID: 第一个拉取的 OpenID，传空表示从头开始拉取
//
// 返回:
//   - *GetUserListResponse: 关注者 openid 列表（NextOpenID 为空表示已取完）
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/User_Management/Get_users_list.html
func (c *UserClient) GetUserList(ctx context.Context, nextOpenID string) (*GetUserListResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	query := map[string]string{
		"access_token": accessToken,
	}
	if nextOpenID != "" {
		query["next_openid"] = nextOpenID
	}

	var result GetUserListResponse
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    URLGetUserList,
		Query:  query,
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

// UpdateUserRemark 设置用户备注名
//
// 参数:
//   - ctx: 上下文
//   - openID: 用户 OpenID
//   - remark: 备注名，长度需小于 30 个字符
//
// 返回:
//   - *UpdateUserRemarkResponse: 设置结果
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/User_Management/Update_remark.html
func (c *UserClient) UpdateUserRemark(ctx context.Context, openID, remark string) (*UpdateUserRemarkResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result UpdateUserRemarkResponse
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLUpdateRemark,
		Query: map[string]string{
			"access_token": accessToken,
		},
		Body:   UpdateUserRemarkRequest{OpenID: openID, Remark: remark},
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

// CreateTag 创建用户标签（公众号最多可创建 100 个标签）
//
// 参数:
//   - ctx: 上下文
//   - name: 标签名（30 个字符以内）
//
// 返回:
//   - *CreateUserTagResponse: 创建结果（含标签 ID）
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (c *UserClient) CreateTag(ctx context.Context, name string) (*CreateUserTagResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var req CreateUserTagRequest
	req.Tag.Name = name

	var result CreateUserTagResponse
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLCreateTag,
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

// GetTags 获取公众号已创建的所有标签
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - *GetTagsResponse: 标签列表（含各标签下粉丝数）
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (c *UserClient) GetTags(ctx context.Context) (*GetTagsResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetTagsResponse
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "GET",
		URL:    URLGetTags,
		Query: map[string]string{
			"access_token": accessToken,
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

// UpdateTag 更新标签名称
//
// 参数:
//   - ctx: 上下文
//   - tagID: 标签 ID
//   - name: 新标签名（30 个字符以内）
//
// 返回:
//   - *UpdateUserTagResponse: 更新结果
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (c *UserClient) UpdateTag(ctx context.Context, tagID int, name string) (*UpdateUserTagResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result UpdateUserTagResponse
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLUpdateTag,
		Query: map[string]string{
			"access_token": accessToken,
		},
		Body:   UpdateUserTagRequest{Tag: UserTag{ID: tagID, Name: name}},
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

// DeleteTag 删除标签
//
// 参数:
//   - ctx: 上下文
//   - tagID: 标签 ID
//
// 返回:
//   - *DeleteUserTagResponse: 删除结果
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (c *UserClient) DeleteTag(ctx context.Context, tagID int) (*DeleteUserTagResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result DeleteUserTagResponse
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLDeleteTag,
		Query: map[string]string{
			"access_token": accessToken,
		},
		Body:   DeleteUserTagRequest{Tag: UserTag{ID: tagID}},
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

// BatchTagUsers 批量为用户打标签
//
// 参数:
//   - ctx: 上下文
//   - tagID: 标签 ID
//   - openIDs: 用户 OpenID 列表，单次最多 50 个
//
// 返回:
//   - *BatchTagUsersResponse: 打标签结果
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (c *UserClient) BatchTagUsers(ctx context.Context, tagID int, openIDs []string) (*BatchTagUsersResponse, error) {
	if len(openIDs) == 0 {
		return nil, fmt.Errorf("openIDs 不能为空")
	}
	if len(openIDs) > 50 {
		return nil, fmt.Errorf("openIDs 单次最多支持 50 个，当前 %d 个", len(openIDs))
	}

	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result BatchTagUsersResponse
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLSetUserTag,
		Query: map[string]string{
			"access_token": accessToken,
		},
		Body:   BatchTagUsersRequest{OpenIDList: openIDs, TagID: tagID},
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

// BatchUntagUsers 批量取消用户标签
//
// 参数:
//   - ctx: 上下文
//   - tagID: 标签 ID
//   - openIDs: 用户 OpenID 列表，单次最多 50 个
//
// 返回:
//   - *BatchUntagUsersResponse: 取消标签结果
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (c *UserClient) BatchUntagUsers(ctx context.Context, tagID int, openIDs []string) (*BatchUntagUsersResponse, error) {
	if len(openIDs) == 0 {
		return nil, fmt.Errorf("openIDs 不能为空")
	}
	if len(openIDs) > 50 {
		return nil, fmt.Errorf("openIDs 单次最多支持 50 个，当前 %d 个", len(openIDs))
	}

	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result BatchUntagUsersResponse
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLCancelUserTag,
		Query: map[string]string{
			"access_token": accessToken,
		},
		Body:   BatchUntagUsersRequest{OpenIDList: openIDs, TagID: tagID},
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

// GetUsersByTag 获取标签下粉丝列表（一次最多 10000 个，可通过 nextOpenID 翻页）
//
// 参数:
//   - ctx: 上下文
//   - tagID: 标签 ID
//   - nextOpenID: 第一个拉取的 OpenID，传空表示从头开始拉取
//
// 返回:
//   - *GetUsersByTagResponse: 粉丝 openid 列表（NextOpenID 为空表示已取完）
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (c *UserClient) GetUsersByTag(ctx context.Context, tagID int, nextOpenID string) (*GetUsersByTagResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetUsersByTagResponse
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLGetTagUsers,
		Query: map[string]string{
			"access_token": accessToken,
		},
		Body:   GetUsersByTagRequest{TagID: tagID, NextOpenID: nextOpenID},
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

// GetTagIDList 获取用户身上的标签列表
//
// 参数:
//   - ctx: 上下文
//   - openID: 用户 OpenID
//
// 返回:
//   - *GetTagIDListResponse: 用户的标签 ID 列表
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://developers.weixin.qq.com/doc/offiaccount/User_Management/User_Tag_Management.html
func (c *UserClient) GetTagIDList(ctx context.Context, openID string) (*GetTagIDListResponse, error) {
	accessToken, err := c.Client.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result GetTagIDListResponse
	err = c.Client.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLGetUserTags,
		Query: map[string]string{
			"access_token": accessToken,
		},
		Body:   GetTagIDListRequest{OpenID: openID},
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
