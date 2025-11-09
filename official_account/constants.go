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
	URLConnectOAuth2Authorize      = core.OpenBaseURL + "/connect/oauth2/authorize"
	URLSnsOAuth2AccessToken        = core.BaseAPIURL + "/sns/oauth2/access_token"
	URLSnsOAuth2RefreshToken       = core.BaseAPIURL + "/sns/oauth2/refresh_token"
	URLSnsUserInfo                 = core.BaseAPIURL + "/sns/userinfo"
	URLSnsAuth                     = core.BaseAPIURL + "/sns/auth"
	
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
	URLDelWxaNewTemplate         = core.BaseAPIURL + "/wxaapi/newtmpl/deltemplate"
	URLGetCategory               = core.BaseAPIURL + "/wxaapi/newtmpl/getcategory"
	URLGetPubTemplateKeywords    = core.BaseAPIURL + "/wxaapi/newtmpl/getpubtemplatekeywords"
	URLGetPubTemplateTitles      = core.BaseAPIURL + "/wxaapi/newtmpl/getpubtemplatetitles"
	URLGetWxaPubTemplate         = core.BaseAPIURL + "/wxaapi/newtmpl/gettemplate"
	URLAddWxaNewTemplate         = core.BaseAPIURL + "/wxaapi/newtmpl/addtemplate"
	URLSendNewSubscribeMsg       = core.BaseAPIURL + "/cgi-bin/message/subscribe/bizsend"
	URLTemplateSubscribe         = core.BaseAPIURL + "/cgi-bin/message/template/subscribe"

	// 二维码
	URLQRCodeCreate = core.BaseAPIURL + "/cgi-bin/qrcode/create"
	URLQRCodeShow   = core.MPBaseURL + "/cgi-bin/showqrcode"

	// 短链接
	URLShortURL = core.BaseAPIURL + "/cgi-bin/shorturl"

	// 设备功能
	URLDeviceAuthorize   = core.BaseAPIURL + "/device/authorize"
	URLDeviceCreateQRCode = core.BaseAPIURL + "/device/create_qrcode"
	URLDeviceGetQRCode   = core.BaseAPIURL + "/device/getqrcode"
	URLDeviceGetStatus   = core.BaseAPIURL + "/device/get_stat"
	URLDeviceOpen        = core.BaseAPIURL + "/device/opendevice"
	URLDeviceSend        = core.BaseAPIURL + "/device/transmsg"
	URLDeviceUnbind      = core.BaseAPIURL + "/device/unbind"
	URLDeviceUpdateStatus = core.BaseAPIURL + "/device/updatestatus"

	// 用户管理
	URLUserInfo       = core.BaseAPIURL + "/cgi-bin/user/info"
	URLGetUserInfoBatch = core.BaseAPIURL + "/cgi-bin/user/info/batchget"
	URLGetUserList    = core.BaseAPIURL + "/cgi-bin/user/get"
	URLUpdateRemark   = core.BaseAPIURL + "/cgi-bin/user/info/updateremark"
	URLSetUserTag     = core.BaseAPIURL + "/cgi-bin/tags/members/batchtagging"
	URLGetUserTags    = core.BaseAPIURL + "/cgi-bin/tags/getidlist"
	URLCancelUserTag  = core.BaseAPIURL + "/cgi-bin/tags/members/batchuntagging"
	URLGetTagUsers    = core.BaseAPIURL + "/cgi-bin/user/tag/get"

	// 用户标签
	URLCreateTag       = core.BaseAPIURL + "/cgi-bin/tags/create"
	URLDeleteTag       = core.BaseAPIURL + "/cgi-bin/tags/delete"
	URLGetTags         = core.BaseAPIURL + "/cgi-bin/tags/get"
	URLUpdateTag       = core.BaseAPIURL + "/cgi-bin/tags/update"

	// 数据分析
	URLDataCubeGetUserSummary   = core.BaseAPIURL + "/datacube/getusersummary"
	URLDataCubeGetUserCumulate  = core.BaseAPIURL + "/datacube/getusercumulate"
	URLDataCubeGetArticleSummary = core.BaseAPIURL + "/datacube/getarticlesummary"
	URLDataCubeGetArticleTotal  = core.BaseAPIURL + "/datacube/getarticletotal"
	URLDataCubeGetUserRead      = core.BaseAPIURL + "/datacube/getuserread"
	URLDataCubeGetUserReadHour  = core.BaseAPIURL + "/datacube/getuserreadhour"
	URLDataCubeGetUserShare     = core.BaseAPIURL + "/datacube/getusershare"
	URLDataCubeGetUserShareHour = core.BaseAPIURL + "/datacube/getusersharehour"
	URLDataCubeGetUpStreamMsg   = core.BaseAPIURL + "/datacube/getupstreammsg"
	URLDataCubeGetUpStreamMsgHour = core.BaseAPIURL + "/datacube/getupstreammsghour"
	URLDataCubeGetUpStreamMsgDist = core.BaseAPIURL + "/datacube/getupstreammsgdist"
	URLDataCubeGetUpStreamMsgDistWeekly = core.BaseAPIURL + "/datacube/getupstreammsgdistweek"
	URLDataCubeGetUpStreamMsgDistMonthly = core.BaseAPIURL + "/datacube/getupstreammsgdistmonth"
	URLDataCubeGetInterfaceSummary = core.BaseAPIURL + "/datacube/getinterfacesummary"
	URLDataCubeGetInterfaceSummaryHour = core.BaseAPIURL + "/datacube/getinterfacesummaryhour"

	// 语义理解
	URLSemanticSearch = core.BaseAPIURL + "/semantic/semproxy/search"

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

	// 微信小店
	URLProductAdd              = core.BaseAPIURL + "/merchant/create"
	URLProductCategoryGet      = core.BaseAPIURL + "/merchant/category/get"
	URLProductDel              = core.BaseAPIURL + "/merchant/del"
	URLProductGet              = core.BaseAPIURL + "/merchant/get"
	URLProductUpdate           = core.BaseAPIURL + "/merchant/update"
	URLProductStockGet         = core.BaseAPIURL + "/merchant/stock/get"
	URLProductStockUpdate      = core.BaseAPIURL + "/merchant/stock/update"
	URLProductPropertyGet      = core.BaseAPIURL + "/merchant/property/get"
	URLProductSkuGet           = core.BaseAPIURL + "/merchant/sku/get"
	URLProductGroupGet         = core.BaseAPIURL + "/merchant/group/get"
	URLProductGroupAdd         = core.BaseAPIURL + "/merchant/group/add"
	URLProductGroupUpdate      = core.BaseAPIURL + "/merchant/group/update"
	URLProductGroupDel         = core.BaseAPIURL + "/merchant/group/del"
	URLProductGroupProductAdd  = core.BaseAPIURL + "/merchant/group/addproduct"
	URLProductGroupProductDel  = core.BaseAPIURL + "/merchant/group/delproduct"
	URLOrderGetByID            = core.BaseAPIURL + "/merchant/order/getbyid"
	URLOrderGetByStatus        = core.BaseAPIURL + "/merchant/order/getbyfilter"
	URLOrderDelivery           = core.BaseAPIURL + "/merchant/order/setdelivery"
	URLOrderClose              = core.BaseAPIURL + "/merchant/order/close"
	URLUploadShopProductImage  = core.BaseAPIURL + "/merchant/common/upload_img"

	// 微信小店图片
	URLUploadShopImage = core.BaseAPIURL + "/merchant/common/upload_img"

	// 微信支付
	// TODO: 后续移至payment包或公共常量文件
	URLPayNotify = core.PayAPIURL + "/pay/unifiedorder"

	// 微信扫一扫
	URLScanCodePush = core.BaseAPIURL + "/cgi-bin/message/custom/send"

	// 微信发票
	URLInvoiceTitleGet = core.BaseAPIURL + "/card/invoice/reimburse/gettitle"

	// 微信打印
	URLAddPrinter   = core.BaseAPIURL + "/device/printer/addprinter"
	URLDelPrinter   = core.BaseAPIURL + "/device/printer/delprinter"
	URLGetPrinter   = core.BaseAPIURL + "/device/printer/getprinter"
	URLPrint        = core.BaseAPIURL + "/device/printer/print"
	URLGetPrintStatus = core.BaseAPIURL + "/device/printer/getprintstatus"

	// 微信连Wi-Fi
	URLWifiGetQrCode        = core.BaseAPIURL + "/wifi/qrcode/get"
	URLWifiSetHomepage      = core.BaseAPIURL + "/wifi/homepage/set"
	URLWifiGetHomepage      = core.BaseAPIURL + "/wifi/homepage/get"
	URLWifiAddShop          = core.BaseAPIURL + "/wifi/shop/add"
	URLWifiGetShop          = core.BaseAPIURL + "/wifi/shop/get"
	URLWifiListShop         = core.BaseAPIURL + "/wifi/shop/list"
	URLWifiUpdateShop       = core.BaseAPIURL + "/wifi/shop/update"
	URLWifiAddDevice        = core.BaseAPIURL + "/wifi/device/add"
	URLWifiDeleteDevice     = core.BaseAPIURL + "/wifi/device/delete"
	URLWifiGetDevice        = core.BaseAPIURL + "/wifi/device/get"
	URLWifiListDevice       = core.BaseAPIURL + "/wifi/device/list"
	URLWifiUpdateDevice     = core.BaseAPIURL + "/wifi/device/update"
	URLWifiGetStatistics    = core.BaseAPIURL + "/wifi/statistics/get"

	// 摇一摇周边
	URLBeaconGetDevice        = core.BaseAPIURL + "/shakearound/device/search"
	URLBeaconRegisterDevice   = core.BaseAPIURL + "/shakearound/device/register"
	URLBeaconUpdateDevice     = core.BaseAPIURL + "/shakearound/device/update"
	URLBeaconDeleteDevice     = core.BaseAPIURL + "/shakearound/device/delete"
	URLBeaconBindPage         = core.BaseAPIURL + "/shakearound/device/bindpage"
	URLBeaconGetPage          = core.BaseAPIURL + "/shakearound/page/search"
	URLBeaconAddPage          = core.BaseAPIURL + "/shakearound/page/add"
	URLBeaconUpdatePage       = core.BaseAPIURL + "/shakearound/page/update"
	URLBeaconDeletePage       = core.BaseAPIURL + "/shakearound/page/delete"
	URLBeaconGetStatistics    = core.BaseAPIURL + "/shakearound/statistics/devicelist"

	// 开放平台相关
	// TODO: 后续移至openplatform包
	URLComponentAccessToken = core.BaseAPIURL + "/cgi-bin/component/api_component_token"
	URLAuthorizerToken      = core.BaseAPIURL + "/cgi-bin/component/api_authorizer_token"
	URLGetAuthorizerInfo    = core.BaseAPIURL + "/cgi-bin/component/api_get_authorizer_info"
	URLGetPreAuthCode       = core.BaseAPIURL + "/cgi-bin/component/api_create_preauthcode"

	// 事件处理相关
	// 已在core/constants.go中定义，此处保留仅供兼容性使用
	EventResponseSuccess = core.EventResponseSuccess
	EventTimestampTolerance = core.EventTimestampTolerance
)
