package official_account

// 微信公众号API地址常量
const (
	// 基础接口
	APIAccessTokenURL       = "https://api.weixin.qq.com/cgi-bin/token"
	APIStableAccessTokenURL = "https://api.weixin.qq.com/cgi-bin/stable_token"
	APICallbackCheckURL     = "https://api.weixin.qq.com/cgi-bin/callback/check"
	APIGetApiDomainIpURL    = "https://api.weixin.qq.com/cgi-bin/get_api_domain_ip"
	APIGetCallbackIpURL     = "https://api.weixin.qq.com/cgi-bin/getcallbackip"
	APIClearQuotaURL        = "https://api.weixin.qq.com/cgi-bin/clear_quota"

	// 自定义菜单
	APICreateMenuURL            = "https://api.weixin.qq.com/cgi-bin/menu/create"
	APIGetCurrentMenuURL        = "https://api.weixin.qq.com/cgi-bin/get_current_selfmenu_info"
	APIGetMenuURL               = "https://api.weixin.qq.com/cgi-bin/menu/get"
	APIDeleteMenuURL            = "https://api.weixin.qq.com/cgi-bin/menu/delete"
	APIAddConditionalMenuURL    = "https://api.weixin.qq.com/cgi-bin/menu/addconditional"
	APIDeleteConditionalMenuURL = "https://api.weixin.qq.com/cgi-bin/menu/delconditional"
	APITryMatchMenuURL          = "https://api.weixin.qq.com/cgi-bin/menu/trymatch"

	// 消息管理
	APIUploadImageURL   = "https://api.weixin.qq.com/cgi-bin/media/uploadimg"
	APIDeleteMassMsgURL = "https://api.weixin.qq.com/cgi-bin/message/mass/delete"
	APIGetSpeedURL      = "https://api.weixin.qq.com/cgi-bin/message/mass/speed/get"
	APIMassMsgGetURL    = "https://api.weixin.qq.com/cgi-bin/message/mass/get"
	APIMassSendURL      = "https://api.weixin.qq.com/cgi-bin/message/mass/send"
	APIPreviewURL       = "https://api.weixin.qq.com/cgi-bin/message/mass/preview"
	APISendAllURL       = "https://api.weixin.qq.com/cgi-bin/message/mass/sendall"
	APISetSpeedURL      = "https://api.weixin.qq.com/cgi-bin/message/mass/speed/set"
	APIUploadNewsMsgURL = "https://api.weixin.qq.com/cgi-bin/media/uploadnews"

	// 模板消息
	APIMessageTemplateSendURL           = "https://api.weixin.qq.com/cgi-bin/message/template/send"
	APITemplateApiAddTemplateURL        = "https://api.weixin.qq.com/cgi-bin/template/api_add_template"
	APITemplateGetIndustryURL           = "https://api.weixin.qq.com/cgi-bin/template/get_industry"
	APITemplateDelPrivateTemplateURL    = "https://api.weixin.qq.com/cgi-bin/template/del_private_template"
	APITemplateGetAllPrivateTemplateURL = "https://api.weixin.qq.com/cgi-bin/template/get_all_private_template"
	APITemplateApiSetIndustryURL        = "https://api.weixin.qq.com/cgi-bin/template/api_set_industry"

	// 客服消息
	APIMessageCustomSendURL       = "https://api.weixin.qq.com/cgi-bin/message/custom/send"
	APIAddCustomAccountURL        = "https://api.weixin.qq.com/customservice/kfaccount/add"
	APIUpdateCustomAccountURL     = "https://api.weixin.qq.com/customservice/kfaccount/update"
	APIDeleteCustomAccountURL     = "https://api.weixin.qq.com/customservice/kfaccount/del"
	APISetCustomAccountHeadImgURL = "https://api.weixin.qq.com/customservice/kfaccount/uploadheadimg"
	APIGetAllCustomAccountsURL    = "https://api.weixin.qq.com/cgi-bin/customservice/getkflist"
	APIGetOnlineCustomAccountsURL = "https://api.weixin.qq.com/cgi-bin/customservice/getonlinekflist"
	APICreateCustomSessionURL     = "https://api.weixin.qq.com/customservice/kfsession/create"
	APICloseCustomSessionURL      = "https://api.weixin.qq.com/customservice/kfsession/close"
	APIGetCustomSessionURL        = "https://api.weixin.qq.com/customservice/kfsession/getsession"
	APIGetCustomSessionListURL    = "https://api.weixin.qq.com/customservice/kfsession/getsessionlist"
	APIGetWaitCaseURL             = "https://api.weixin.qq.com/customservice/kfsession/getwaitcase"
	APIGetMsgRecordURL            = "https://api.weixin.qq.com/customservice/msgrecord/getmsglist"
	APITypingURL                  = "https://api.weixin.qq.com/cgi-bin/message/custom/typing"

	// 自动回复
	APIGetCurrentAutoreplyInfoURL = "https://api.weixin.qq.com/cgi-bin/get_current_autoreply_info"

	// 素材管理
	APIUploadMaterialURL      = "https://api.weixin.qq.com/cgi-bin/media/upload"
	APIGetMaterialURL         = "https://api.weixin.qq.com/cgi-bin/media/get"
	APIDeleteMaterialURL      = "https://api.weixin.qq.com/cgi-bin/material/del_material"
	APIUpdateNewsURL          = "https://api.weixin.qq.com/cgi-bin/material/update_news"
	APIGetMaterialCountURL    = "https://api.weixin.qq.com/cgi-bin/material/get_materialcount"
	APIBatchGetMaterialURL    = "https://api.weixin.qq.com/cgi-bin/material/batchget_material"
	APIAddNewsURL             = "https://api.weixin.qq.com/cgi-bin/material/add_news"
	APIMaterialUploadImageURL = "https://api.weixin.qq.com/cgi-bin/media/uploadimg"
	APIUploadVideoURL         = "https://api.weixin.qq.com/cgi-bin/material/add_material"
	APIGetHDVoiceURL          = "https://api.weixin.qq.com/cgi-bin/media/get/jssdk"

	// 草稿管理
	APIAddDraftURL      = "https://api.weixin.qq.com/cgi-bin/draft/add"
	APIGetDraftURL      = "https://api.weixin.qq.com/cgi-bin/draft/get"
	APIDeleteDraftURL   = "https://api.weixin.qq.com/cgi-bin/draft/delete"
	APIGetDraftCountURL = "https://api.weixin.qq.com/cgi-bin/draft/count"
	APIBatchGetDraftURL = "https://api.weixin.qq.com/cgi-bin/draft/batchget"
	APIUpdateDraftURL   = "https://api.weixin.qq.com/cgi-bin/draft/update"

	// 订阅通知
	APIDelWxaNewTemplateURL         = "https://api.weixin.qq.com/wxaapi/newtmpl/deltemplate"
	APIGetCategoryURL               = "https://api.weixin.qq.com/wxaapi/newtmpl/getcategory"
	APIGetPubNewTemplateKeywordsURL = "https://api.weixin.qq.com/wxaapi/newtmpl/getpubtemplatekeywords"
	APIGetPubNewTemplateTitlesURL   = "https://api.weixin.qq.com/wxaapi/newtmpl/getpubtemplatetitles"
	APIGetWxaPubNewTemplateURL      = "https://api.weixin.qq.com/wxaapi/newtmpl/gettemplate"
	APIAddWxaNewTemplateURL         = "https://api.weixin.qq.com/wxaapi/newtmpl/addtemplate"
	APISendNewSubscribeMsgURL       = "https://api.weixin.qq.com/cgi-bin/message/subscribe/bizsend"
	APITemplateSubscribeURL         = "https://api.weixin.qq.com/cgi-bin/message/template/subscribe"

	// 二维码
	APIQRCodeCreateURL = "https://api.weixin.qq.com/cgi-bin/qrcode/create"
	APIQRCodeShowURL   = "https://mp.weixin.qq.com/cgi-bin/showqrcode"
)
