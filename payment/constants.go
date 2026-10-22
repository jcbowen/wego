package payment

// 微信支付 V3 相关 API 地址
const (
	// PayJSAPIURL JSAPI 统一下单地址
	PayJSAPIURL = "https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi"

	// PayH5URL H5 支付地址
	PayH5URL = "https://api.mch.weixin.qq.com/v3/pay/transactions/h5"

	// PayNativeURL Native 支付下单地址
	PayNativeURL = "https://api.mch.weixin.qq.com/v3/pay/transactions/native"

	// PayAppURL App 支付下单地址
	PayAppURL = "https://api.mch.weixin.qq.com/v3/pay/transactions/app"

	// PayCodePayURL 付款码支付下单地址（收银员扫顾客付款码，B扫C）
	// 提交后立即返回订单结果：trade_state=SUCCESS 为扣款成功；USERPAYING 表示用户支付中，
	// 需轮询查单确认最终状态；USERPAYING 长期未成功时用撤销接口终止交易
	PayCodePayURL = "https://api.mch.weixin.qq.com/v3/pay/transactions/codepay"

	// PayQueryOrderURL 按商户订单号查询订单地址（需格式化 out_trade_no）
	PayQueryOrderURL = "https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/%s"

	// PayQueryOrderByTransactionIDURL 按微信支付订单号查询订单地址（需格式化 transaction_id）
	PayQueryOrderByTransactionIDURL = "https://api.mch.weixin.qq.com/v3/pay/transactions/id/%s"

	// PayCloseOrderURL 关闭订单地址（需格式化 out_trade_no）
	PayCloseOrderURL = "https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/%s/close"

	// PayReverseOrderURL 撤销订单地址（需格式化 out_trade_no，付款码支付配套：
	// USERPAYING 在途交易无法用关单接口终止，只能撤销）
	PayReverseOrderURL = "https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/%s/reverse"

	// PayRefundURL 退款申请地址
	PayRefundURL = "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds"

	// PayRefundQueryURL 按商户退款单号查询退款地址（需格式化 out_refund_no）
	PayRefundQueryURL = "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds/%s"

	// PayCertificatesURL 获取平台证书地址
	PayCertificatesURL = "https://api.mch.weixin.qq.com/v3/certificates"
)

// 微信支付账单相关 API 地址与常量
const (
	// BillTradeURL 申请交易账单地址
	BillTradeURL = "https://api.mch.weixin.qq.com/v3/bill/tradebill"

	// BillFundFlowURL 申请资金账单地址
	BillFundFlowURL = "https://api.mch.weixin.qq.com/v3/bill/fundflowbill"

	// BillHashTypeSHA1 账单哈希类型，官方固定为 SHA1
	BillHashTypeSHA1 = "SHA1"
)

// 微信支付分账相关 API 地址
const (
	// ProfitSharingOrderURL 请求分账地址
	ProfitSharingOrderURL = "https://api.mch.weixin.qq.com/v3/profitsharing/orders"

	// ProfitSharingQueryURL 查询分账结果地址（需格式化 out_order_no）
	ProfitSharingQueryURL = "https://api.mch.weixin.qq.com/v3/profitsharing/orders/%s"

	// ProfitSharingUnfreezeURL 解冻剩余资金地址
	ProfitSharingUnfreezeURL = "https://api.mch.weixin.qq.com/v3/profitsharing/orders/unfreeze"

	// ProfitSharingReturnURL 请求分账回退地址
	ProfitSharingReturnURL = "https://api.mch.weixin.qq.com/v3/profitsharing/return-orders"

	// ProfitSharingReturnQueryURL 查询分账回退结果地址（需格式化 out_return_no）
	ProfitSharingReturnQueryURL = "https://api.mch.weixin.qq.com/v3/profitsharing/return-orders/%s"

	// ProfitSharingAmountsURL 查询剩余待分金额地址（需格式化 transaction_id）
	ProfitSharingAmountsURL = "https://api.mch.weixin.qq.com/v3/profitsharing/transactions/%s/amounts"
)

// 微信支付验签模式相关常量
const (
	// PublicKeyIDPrefix 微信支付公钥 ID 前缀
	//
	// 说明:
	//   - 微信支付应答与回调的响应头 Wechatpay-Serial 携带该前缀时，表示本次签名使用微信支付公钥
	//   - 未携带该前缀时表示使用平台证书，商户需据此选择对应的验签密钥
	PublicKeyIDPrefix = "PUB_KEY_ID_"

	// PublicKeyHeader 告知微信支付本次请求期望使用的验签/加密密钥的请求头
	PublicKeyHeader = "Wechatpay-Serial"

	// JSAPISignType JSAPI 调起支付的签名类型
	JSAPISignType = "RSA"

	// AppPayPackageValue App 调起支付的 package 固定值
	AppPayPackageValue = "Sign=WXPay"
)

// 支付订单状态（Order.TradeState）
const (
	TradeStateSuccess    = "SUCCESS"    // 支付成功
	TradeStateRefund     = "REFUND"     // 转入退款
	TradeStateNotPay     = "NOTPAY"     // 未支付
	TradeStateClosed     = "CLOSED"     // 已关闭
	TradeStateRevoked    = "REVOKED"    // 已撤销（付款码支付）
	TradeStateUserPaying = "USERPAYING" // 用户支付中（付款码支付）
	TradeStatePayError   = "PAYERROR"   // 支付失败
)

// 退款状态（RefundResponse.Status）
const (
	RefundStatusSuccess  = "SUCCESS"  // 退款成功
	RefundStatusClosed   = "CLOSED"   // 退款关闭
	RefundStatusAbnormal = "ABNORMAL" // 退款异常
)

// 分账状态（ProfitSharingResponse.State）
const (
	ProfitSharingStateProcessing = "PROCESSING" // 分账处理中
	ProfitSharingStateFinished   = "FINISHED"   // 分账处理完成
)

// 分账接收方结果（ProfitSharingReceiverResult.Result）
const (
	ProfitSharingResultPending = "PENDING" // 待分账
	ProfitSharingResultSuccess = "SUCCESS" // 分账成功
	ProfitSharingResultClosed  = "CLOSED"  // 分账失败已关闭
)

// 分账回退结果（ProfitSharingReturnResponse.Result）
const (
	ProfitSharingReturnProcessing = "PROCESSING" // 回退处理中
	ProfitSharingReturnSuccess    = "SUCCESS"    // 回退成功
	ProfitSharingReturnFailed     = "FAILED"     // 回退失败
)

// 回调应答相关常量
//
// 说明:
//   - 商户处理回调成功后必须应答 200/204，否则微信会按 15s/15s/30s/3m/10m/20m/30m/30m/30m/60m/3h/3h/3h/6h/6h/6h 的频率重试 15 次
//   - 处理失败需返回 5xx 或 4xx 并带上 code/message，微信会据此重试
const (
	// NotificationAckJSON 处理成功时推荐返回的应答体
	NotificationAckJSON = `{"code":"SUCCESS","message":"成功"}`

	// NotificationAckCodeSuccess 处理成功应答的 code 值
	NotificationAckCodeSuccess = "SUCCESS"

	// NotificationAckCodeFail 处理失败应答的 code 值
	NotificationAckCodeFail = "FAIL"
)
