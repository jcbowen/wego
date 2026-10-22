// Package openplatform 微信开放平台API - 第三方平台域名设置
// 参考：
//   - https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/ThirdParty/domain/modify_server_domain.html
//   - https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/ThirdParty/domain/modify_jump_domain.html
package openplatform

import (
	"context"
	"fmt"

	"github.com/jcbowen/wego/core"
)

// 域名操作类型（action）
const (
	DomainActionAdd    = "add"    // 添加（删除不存在的域名视为成功）
	DomainActionDelete = "delete" // 删除
	DomainActionSet    = "set"    // 覆盖
	DomainActionGet    = "get"    // 获取（同时返回测试版和全网发布版）
)

// ModifyWxaServerDomainRequest 设置第三方平台小程序服务器域名请求
//
// 注意：
//   - 每月只可修改申请 50 次
//   - 域名最多 1000 个，以 ; 隔开；不带 http:// 协议头，不带 URI 路径，如 "www.qq.com;wx.qq.com"
//   - action=get 时无需传域名字段
type ModifyWxaServerDomainRequest struct {
	Action                    string `json:"action"`                                 // 操作类型：add/delete/set/get
	WxaServerDomain           string `json:"wxa_server_domain,omitempty"`            // 服务器域名列表（; 分隔）
	IsModifyPublishedTogether *bool  `json:"is_modify_published_together,omitempty"` // 是否同时修改"全网发布版"（默认 false 只改测试版）
}

// ModifyWxaServerDomainResponse 设置第三方平台小程序服务器域名响应
type ModifyWxaServerDomainResponse struct {
	core.APIResponse
	PublishedWxaServerDomain string `json:"published_wxa_server_domain,omitempty"` // 生效的"全网发布版"域名（修改失败或无已发布平台时不返回）
	TestingWxaServerDomain   string `json:"testing_wxa_server_domain,omitempty"`   // 生效的"测试版"域名
	InvalidWxaServerDomain   string `json:"invalid_wxa_server_domain,omitempty"`   // 未通过验证的域名
}

// ModifyWxaJumpDomainRequest 设置第三方平台小程序业务域名请求
//
// 注意：
//   - 配置前需将校验文件放置在域名根目录（校验文件通过 URLGetJumpDomainFile 获取）
//   - 域名最多 200 个，以 ; 隔开；不带协议头与 URI 路径
//   - action=get 时无需传域名字段
type ModifyWxaJumpDomainRequest struct {
	Action                    string `json:"action"`                                 // 操作类型：add/delete/set/get
	WxaJumpH5Domain           string `json:"wxa_jump_h5_domain,omitempty"`           // 业务域名列表（; 分隔）
	IsModifyPublishedTogether *bool  `json:"is_modify_published_together,omitempty"` // 是否同时修改"全网发布版"（默认 false 只改测试版）
}

// ModifyWxaJumpDomainResponse 设置第三方平台小程序业务域名响应
type ModifyWxaJumpDomainResponse struct {
	core.APIResponse
	PublishedWxaJumpH5Domain string `json:"published_wxa_jump_h5_domain,omitempty"` // 生效的"全网发布版"业务域名
	TestingWxaJumpH5Domain   string `json:"testing_wxa_jump_h5_domain,omitempty"`   // 生效的"测试版"业务域名
	InvalidWxaJumpH5Domain   string `json:"invalid_wxa_jump_h5_domain,omitempty"`   // 未通过验证的域名
}

// ModifyWxaServerDomain 设置第三方平台自身的小程序服务器域名
//
// 授权给第三方代开发的小程序，其服务器域名只能为第三方平台配置的域名；
// 代小程序发布代码之前需先通过本接口配置平台自身域名。
//
// 参数:
//   - ctx: 上下文
//   - req: 请求参数（action=add/delete/set/get）
//
// 返回:
//   - *ModifyWxaServerDomainResponse: 各版本生效域名与未通过验证的域名
//   - error: 请求失败或微信返回错误时返回错误
//
// 使用示例:
//
//	resp, err := openPlatformClient.ModifyWxaServerDomain(ctx, &openplatform.ModifyWxaServerDomainRequest{
//		Action:          openplatform.DomainActionSet,
//		WxaServerDomain: "www.qq.com;wx.qq.com",
//	})
//
// 文档:
//   - https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/ThirdParty/domain/modify_server_domain.html
func (c *Client) ModifyWxaServerDomain(ctx context.Context, req *ModifyWxaServerDomainRequest) (*ModifyWxaServerDomainResponse, error) {
	if req == nil || req.Action == "" {
		return nil, fmt.Errorf("action 不能为空（add/delete/set/get）")
	}

	token, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	var result ModifyWxaServerDomainResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLModifyServerDomain,
		Query:  map[string]string{"access_token": token.AccessToken},
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

// ModifyWxaJumpDomain 设置第三方平台自身的小程序业务域名
//
// 参数:
//   - ctx: 上下文
//   - req: 请求参数（action=add/delete/set/get）
//
// 返回:
//   - *ModifyWxaJumpDomainResponse: 各版本生效域名与未通过验证的域名
//   - error: 请求失败或微信返回错误时返回错误
//
// 使用示例:
//
//	resp, err := openPlatformClient.ModifyWxaJumpDomain(ctx, &openplatform.ModifyWxaJumpDomainRequest{
//		Action:          openplatform.DomainActionSet,
//		WxaJumpH5Domain: "www.qq.com;wx.qq.com",
//	})
//
// 文档:
//   - https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/ThirdParty/domain/modify_jump_domain.html
func (c *Client) ModifyWxaJumpDomain(ctx context.Context, req *ModifyWxaJumpDomainRequest) (*ModifyWxaJumpDomainResponse, error) {
	if req == nil || req.Action == "" {
		return nil, fmt.Errorf("action 不能为空（add/delete/set/get）")
	}

	token, err := c.GetComponentAccessToken(ctx, "")
	if err != nil {
		return nil, err
	}

	var result ModifyWxaJumpDomainResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    URLModifyJumpDomain,
		Query:  map[string]string{"access_token": token.AccessToken},
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
