package official_account

import (
	"github.com/jcbowen/wego/core"
)

// 微信公众号API地址常量
const (
	// 基础接口
	URLAccessToken       = core.BaseAPIURL + "/cgi-bin/token"
	URLStableAccessToken = core.BaseAPIURL + "/cgi-bin/stable_token"
	URLCallbackCheck     = core.BaseAPIURL + "/cgi-bin/callback/check"
	URLGetApiDomainIp    = core.BaseAPIURL + "/cgi-bin/get_api_domain_ip"
	URLGetCallbackIp     = core.BaseAPIURL + "/cgi-bin/getcallbackip"
	URLClearQuota        = core.BaseAPIURL + "/cgi-bin/clear_quota"
	URLGetTicket         = core.BaseAPIURL + "/cgi-bin/ticket/getticket"

	// 网页授权
	URLConnectOAuth2Authorize = core.OpenBaseURL + "/connect/oauth2/authorize"
	URLSnsOAuth2AccessToken   = core.BaseAPIURL + "/sns/oauth2/access_token"
	URLSnsOAuth2RefreshToken  = core.BaseAPIURL + "/sns/oauth2/refresh_token"
	URLSnsUserInfo            = core.BaseAPIURL + "/sns/userinfo"
	URLSnsAuth                = core.BaseAPIURL + "/sns/auth"

	// 开放平台相关（临时保留，后续应移动到openplatform包）
	URLSnsOAuth2ComponentAccessToken  = core.BaseAPIURL + "/sns/oauth2/component/access_token"
	URLSnsOAuth2ComponentRefreshToken = core.BaseAPIURL + "/sns/oauth2/component/refresh_token"

	// 自定义菜单
	URLCreateMenu            = core.BaseAPIURL + "/cgi-bin/menu/create"
	URLGetCurrentMenu        = core.BaseAPIURL + "/cgi-bin/get_current_selfmenu_info"
	URLGetMenu               = core.BaseAPIURL + "/cgi-bin/menu/get"
	URLDeleteMenu            = core.BaseAPIURL + "/cgi-bin/menu/delete"
	URLAddConditionalMenu    = core.BaseAPIURL + "/cgi-bin/menu/addconditional"
	URLDeleteConditionalMenu = core.BaseAPIURL + "/cgi-bin/menu/delconditional"
	URLTryMatchMenu          = core.BaseAPIURL + "/cgi-bin/menu/trymatch"

	// 消息管理
	URLUploadImage   = core.BaseAPIURL + "/cgi-bin/media/uploadimg"
	URLDeleteMassMsg = core.BaseAPIURL + "/cgi-bin/message/mass/delete"
	URLGetSpeed      = core.BaseAPIURL + "/cgi-bin/message/mass/speed/get"
	URLMassMsgGet    = core.BaseAPIURL + "/cgi-bin/message/mass/get"
	URLMassSend      = core.BaseAPIURL + "/cgi-bin/message/mass/send"
	URLPreview       = core.BaseAPIURL + "/cgi-bin/message/mass/preview"
	URLSendAll       = core.BaseAPIURL + "/cgi-bin/message/mass/sendall"
	URLSetSpeed      = core.BaseAPIURL + "/cgi-bin/message/mass/speed/set"
	URLUploadNewsMsg = core.BaseAPIURL + "/cgi-bin/media/uploadnews"

	// 模板消息
	URLMessageTemplateSend           = core.BaseAPIURL + "/cgi-bin/message/template/send"
	URLTemplateAddTemplate           = core.BaseAPIURL + "/cgi-bin/template/api_add_template"
	URLTemplateGetIndustry           = core.BaseAPIURL + "/cgi-bin/template/get_industry"
	URLTemplateDelPrivateTemplate    = core.BaseAPIURL + "/cgi-bin/template/del_private_template"
	URLTemplateGetAllPrivateTemplate = core.BaseAPIURL + "/cgi-bin/template/get_all_private_template"
	URLTemplateSetIndustry           = core.BaseAPIURL + "/cgi-bin/template/api_set_industry"
	// URLQueryBlockTmplMsg 查询被拦截的模板消息记录
	// 参考：https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Template_Message_Interface.html
	URLQueryBlockTmplMsg = core.BaseAPIURL + "/cgi-bin/message/template/query_block_tmpl_msg"

	// 客服消息
	URLMessageCustomSend       = core.BaseAPIURL + "/cgi-bin/message/custom/send"
	URLAddCustomAccount        = core.BaseAPIURL + "/customservice/kfaccount/add"
	URLUpdateCustomAccount     = core.BaseAPIURL + "/customservice/kfaccount/update"
	URLDeleteCustomAccount     = core.BaseAPIURL + "/customservice/kfaccount/del"
	URLSetCustomAccountHeadImg = core.BaseAPIURL + "/customservice/kfaccount/uploadheadimg"
	URLGetAllCustomAccounts    = core.BaseAPIURL + "/cgi-bin/customservice/getkflist"
	URLGetOnlineCustomAccounts = core.BaseAPIURL + "/cgi-bin/customservice/getonlinekflist"
	URLCreateCustomSession     = core.BaseAPIURL + "/customservice/kfsession/create"
	URLCloseCustomSession      = core.BaseAPIURL + "/customservice/kfsession/close"
	URLGetCustomSession        = core.BaseAPIURL + "/customservice/kfsession/getsession"
	URLGetCustomSessionList    = core.BaseAPIURL + "/customservice/kfsession/getsessionlist"
	URLGetWaitCase             = core.BaseAPIURL + "/customservice/kfsession/getwaitcase"
	URLGetMsgRecord            = core.BaseAPIURL + "/customservice/msgrecord/getmsglist"
	URLTyping                  = core.BaseAPIURL + "/cgi-bin/message/custom/typing"

	// 自动回复
	URLGetCurrentAutoreplyInfo = core.BaseAPIURL + "/cgi-bin/get_current_autoreply_info"

	// 素材管理
	URLUploadMaterial      = core.BaseAPIURL + "/cgi-bin/media/upload"
	URLGetMaterial         = core.BaseAPIURL + "/cgi-bin/media/get"
	URLDeleteMaterial      = core.BaseAPIURL + "/cgi-bin/material/del_material"
	URLUpdateNews          = core.BaseAPIURL + "/cgi-bin/material/update_news"
	URLGetMaterialCount    = core.BaseAPIURL + "/cgi-bin/material/get_materialcount"
	URLBatchGetMaterial    = core.BaseAPIURL + "/cgi-bin/material/batchget_material"
	URLAddNews             = core.BaseAPIURL + "/cgi-bin/material/add_news"
	URLMaterialUploadImage = core.BaseAPIURL + "/cgi-bin/media/uploadimg"
	URLUploadVideo         = core.BaseAPIURL + "/cgi-bin/material/add_material"
	URLGetHDVoice          = core.BaseAPIURL + "/cgi-bin/media/get/jssdk"

	// 草稿管理
	URLAddDraft      = core.BaseAPIURL + "/cgi-bin/draft/add"
	URLGetDraft      = core.BaseAPIURL + "/cgi-bin/draft/get"
	URLDeleteDraft   = core.BaseAPIURL + "/cgi-bin/draft/delete"
	URLGetDraftCount = core.BaseAPIURL + "/cgi-bin/draft/count"
	URLBatchGetDraft = core.BaseAPIURL + "/cgi-bin/draft/batchget"
	URLUpdateDraft   = core.BaseAPIURL + "/cgi-bin/draft/update"

	// 订阅通知
	URLDelWxaNewTemplate      = core.BaseAPIURL + "/wxaapi/newtmpl/deltemplate"
	URLGetCategory            = core.BaseAPIURL + "/wxaapi/newtmpl/getcategory"
	URLGetPubTemplateKeywords = core.BaseAPIURL + "/wxaapi/newtmpl/getpubtemplatekeywords"
	URLGetPubTemplateTitles   = core.BaseAPIURL + "/wxaapi/newtmpl/getpubtemplatetitles"
	URLGetWxaPubTemplate      = core.BaseAPIURL + "/wxaapi/newtmpl/gettemplate"
	URLAddWxaNewTemplate      = core.BaseAPIURL + "/wxaapi/newtmpl/addtemplate"
	URLSendNewSubscribeMsg    = core.BaseAPIURL + "/cgi-bin/message/subscribe/bizsend"
	URLTemplateSubscribe      = core.BaseAPIURL + "/cgi-bin/message/template/subscribe"

	// 二维码
	URLQRCodeCreate = core.BaseAPIURL + "/cgi-bin/qrcode/create"
	URLQRCodeShow   = core.MPBaseURL + "/cgi-bin/showqrcode"

	// 用户管理
	URLUserInfo         = core.BaseAPIURL + "/cgi-bin/user/info"
	URLGetUserInfoBatch = core.BaseAPIURL + "/cgi-bin/user/info/batchget"
	URLGetUserList      = core.BaseAPIURL + "/cgi-bin/user/get"
	URLUpdateRemark     = core.BaseAPIURL + "/cgi-bin/user/info/updateremark"
	URLSetUserTag       = core.BaseAPIURL + "/cgi-bin/tags/members/batchtagging"
	URLGetUserTags      = core.BaseAPIURL + "/cgi-bin/tags/getidlist"
	URLCancelUserTag    = core.BaseAPIURL + "/cgi-bin/tags/members/batchuntagging"
	URLGetTagUsers      = core.BaseAPIURL + "/cgi-bin/user/tag/get"

	// 用户标签
	URLCreateTag = core.BaseAPIURL + "/cgi-bin/tags/create"
	URLDeleteTag = core.BaseAPIURL + "/cgi-bin/tags/delete"
	URLGetTags   = core.BaseAPIURL + "/cgi-bin/tags/get"
	URLUpdateTag = core.BaseAPIURL + "/cgi-bin/tags/update"

	// 数据分析
	URLDataCubeGetUserSummary            = core.BaseAPIURL + "/datacube/getusersummary"
	URLDataCubeGetUserCumulate           = core.BaseAPIURL + "/datacube/getusercumulate"
	URLDataCubeGetArticleSummary         = core.BaseAPIURL + "/datacube/getarticlesummary"
	URLDataCubeGetArticleTotal           = core.BaseAPIURL + "/datacube/getarticletotal"
	URLDataCubeGetUserRead               = core.BaseAPIURL + "/datacube/getuserread"
	URLDataCubeGetUserReadHour           = core.BaseAPIURL + "/datacube/getuserreadhour"
	URLDataCubeGetUserShare              = core.BaseAPIURL + "/datacube/getusershare"
	URLDataCubeGetUserShareHour          = core.BaseAPIURL + "/datacube/getusersharehour"
	URLDataCubeGetUpStreamMsg            = core.BaseAPIURL + "/datacube/getupstreammsg"
	URLDataCubeGetUpStreamMsgHour        = core.BaseAPIURL + "/datacube/getupstreammsghour"
	URLDataCubeGetUpStreamMsgDist        = core.BaseAPIURL + "/datacube/getupstreammsgdist"
	URLDataCubeGetUpStreamMsgDistWeekly  = core.BaseAPIURL + "/datacube/getupstreammsgdistweek"
	URLDataCubeGetUpStreamMsgDistMonthly = core.BaseAPIURL + "/datacube/getupstreammsgdistmonth"
	URLDataCubeGetInterfaceSummary       = core.BaseAPIURL + "/datacube/getinterfacesummary"
	URLDataCubeGetInterfaceSummaryHour   = core.BaseAPIURL + "/datacube/getinterfacesummaryhour"

	// 微信卡券
	URLCardCodeDecrypt        = core.BaseAPIURL + "/card/code/decrypt"
	URLCardCodeGet            = core.BaseAPIURL + "/card/code/get"
	URLCardCodeUnavailable    = core.BaseAPIURL + "/card/code/unavailable"
	URLCardCodeUpdate         = core.BaseAPIURL + "/card/code/update"
	URLCardCreateQRCode       = core.BaseAPIURL + "/card/qrcode/create"
	URLCardCreate             = core.BaseAPIURL + "/card/create"
	URLCardDelete             = core.BaseAPIURL + "/card/delete"
	URLCardGetHTML            = core.BaseAPIURL + "/card/gethtml"
	URLCardGetUserCardList    = core.BaseAPIURL + "/card/user/getcardlist"
	URLCardOrderInfo          = core.BaseAPIURL + "/card/pay/getorderinfo"
	URLCardSetPayCell         = core.BaseAPIURL + "/card/pay/setcell"
	URLCardSetSelfConsumeCell = core.BaseAPIURL + "/card/pay/setselfconsumecell"
	URLCardStockGetCount      = core.BaseAPIURL + "/card/stock/getcount"
	URLCardStockBatchGet      = core.BaseAPIURL + "/card/batchget"
	URLCardTestWhiteListSet   = core.BaseAPIURL + "/card/testwhitelist/set"
	URLCardUpdate             = core.BaseAPIURL + "/card/update"
	URLCardQrCodeCreate       = core.BaseAPIURL + "/card/qrcode/create"

	// 微信支付
	// TODO: 后续移至payment包或公共常量文件
	URLPayNotify = core.PayAPIURL + "/pay/unifiedorder"

	// 微信扫一扫
	URLScanCodePush = core.BaseAPIURL + "/cgi-bin/message/custom/send"

	// 开放平台相关
	// TODO: 后续移至openplatform包
	URLComponentAccessToken = core.BaseAPIURL + "/cgi-bin/component/api_component_token"
	URLAuthorizerToken      = core.BaseAPIURL + "/cgi-bin/component/api_authorizer_token"
	URLGetAuthorizerInfo    = core.BaseAPIURL + "/cgi-bin/component/api_get_authorizer_info"
	URLGetPreAuthCode       = core.BaseAPIURL + "/cgi-bin/component/api_create_preauthcode"

	// 事件处理相关
	// 已在core/constants.go中定义，此处保留仅供兼容性使用
	EventResponseSuccess    = core.EventResponseSuccess
	EventTimestampTolerance = core.EventTimestampTolerance
)
