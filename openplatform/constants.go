package openplatform

import (
	"github.com/jcbowen/wego/core"
)

// 微信开放平台API地址常量
const (
	// 组件OAuth
	URLComponentOAuth2AccessToken = core.BaseAPIURL + "/sns/oauth2/component/access_token"
	URLComponentOAuth2UserInfo    = core.BaseAPIURL + "/sns/userinfo"

	// 组件基础接口
	URLComponentToken      = core.BaseAPIURL + "/cgi-bin/component/api_component_token"
	URLPreAuthCode         = core.BaseAPIURL + "/cgi-bin/component/api_create_preauthcode"
	URLQueryAuth           = core.BaseAPIURL + "/cgi-bin/component/api_query_auth"
	URLAuthorizerToken     = core.BaseAPIURL + "/cgi-bin/component/api_authorizer_token"
	URLGetAuthorizerInfo   = core.BaseAPIURL + "/cgi-bin/component/api_get_authorizer_info"
	URLGetAuthorizerList   = core.BaseAPIURL + "/cgi-bin/component/api_get_authorizer_list"
	URLGetAuthorizerOption = core.BaseAPIURL + "/cgi-bin/component/api_get_authorizer_option"
	URLSetAuthorizerOption = core.BaseAPIURL + "/cgi-bin/component/api_set_authorizer_option"
	URLStartPushTicket     = core.BaseAPIURL + "/cgi-bin/component/api_start_push_ticket"
	// URLClearComponentQuota 使用AppSecret重置第三方平台API调用次数（v2 版本，无 query 参数）
	// 参考：https://developers.weixin.qq.com/doc/oplatform/openApi/openapi/api_componentclearquota_v2.html
	URLClearComponentQuota = core.BaseAPIURL + "/cgi-bin/component/clear_quota/v2"
	URLModifyServerDomain  = core.BaseAPIURL + "/cgi-bin/component/modify_wxa_server_domain"
	URLGetJumpDomainFile   = core.BaseAPIURL + "/cgi-bin/component/get_domain_confirmfile"
	URLModifyJumpDomain    = core.BaseAPIURL + "/cgi-bin/component/modify_wxa_jump_domain"

	// 通用接口
	URLClearQuota  = core.BaseAPIURL + "/cgi-bin/clear_quota"
	URLGetApiQuota = core.BaseAPIURL + "/cgi-bin/openapi/quota/get"
	URLGetRidInfo  = core.BaseAPIURL + "/cgi-bin/openapi/rid/get"

	// 小程序
	URLWxaGetTemplateDraftList = core.BaseAPIURL + "/wxa/gettemplatedraftlist"
	URLWxaAddToTemplate        = core.BaseAPIURL + "/wxa/addtotemplate"
	URLWxaGetTemplateList      = core.BaseAPIURL + "/wxa/gettemplatelist"
	URLWxaDeleteTemplate       = core.BaseAPIURL + "/wxa/deletetemplate"

	// 小程序码
	URLGetWxaCode = core.BaseAPIURL + "/wxa/getwxacode"

	// 用户信息
	URLGetUserInfo = core.BaseAPIURL + "/cgi-bin/user/info"

	// URLSnsOAuth2ComponentRefreshToken 第三方平台代公众号刷新**网页授权** access_token
	// 注意：刷新「授权方 access_token」请使用 URLAuthorizerToken（api_authorizer_token），两者不可混用
	// 参考：https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/official_account/refresh_token.html
	URLSnsOAuth2ComponentRefreshToken = core.BaseAPIURL + "/sns/oauth2/component/refresh_token"
)

// 授权类型常量
const (
	// AuthTypeOfficialAccount 公众号授权
	AuthTypeOfficialAccount = 1

	// AuthTypeMiniProgram 小程序授权
	AuthTypeMiniProgram = 2

	// AuthTypeBoth 公众号和小程序都展示
	AuthTypeBoth = 3

	// AuthTypeMiniProgramPromoter 小程序推客账号
	AuthTypeMiniProgramPromoter = 4

	// AuthTypeChannels 视频号账号
	AuthTypeChannels = 5

	// AuthTypeAll 全部
	AuthTypeAll = 6

	// AuthTypeLiveStreamingAssistant 带货助手账号
	AuthTypeLiveStreamingAssistant = 8
)
