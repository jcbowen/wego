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

// 微信支付商家转账（用户确认收款模式）相关 API 地址
// 参考：https://pay.weixin.qq.com/doc/v3/merchant/4012716434
const (
	// MchTransferCreateURL 发起转账（需用户确认收款）地址
	MchTransferCreateURL = "https://api.mch.weixin.qq.com/v3/fund-app/mch-transfer/transfer-bills"

	// MchTransferTransferURL 用户授权后转账地址（用户完成授权后实际出款）
	MchTransferTransferURL = "https://api.mch.weixin.qq.com/v3/fund-app/mch-transfer/transfer-bills/transfer"

	// MchTransferCancelURL 撤销转账地址（需格式化 out_bill_no，仅用户确认收款前可撤销）
	MchTransferCancelURL = "https://api.mch.weixin.qq.com/v3/fund-app/mch-transfer/transfer-bills/out-bill-no/%s/cancel"

	// MchTransferQueryByOutBillNoURL 按商户单号查询转账单地址（需格式化 out_bill_no）
	MchTransferQueryByOutBillNoURL = "https://api.mch.weixin.qq.com/v3/fund-app/mch-transfer/transfer-bills/out-bill-no/%s"

	// MchTransferQueryByBillNoURL 按微信转账单号查询转账单地址（需格式化 transfer_bill_no）
	MchTransferQueryByBillNoURL = "https://api.mch.weixin.qq.com/v3/fund-app/mch-transfer/transfer-bills/transfer-bill-no/%s"

	// MchTransferElecsignApplyByOutBillNoURL 按商户单号申请电子回单地址
	MchTransferElecsignApplyByOutBillNoURL = "https://api.mch.weixin.qq.com/v3/fund-app/mch-transfer/elecsign/out-bill-no"

	// MchTransferElecsignQueryByOutBillNoURL 按商户单号查询电子回单地址（需格式化 out_bill_no）
	MchTransferElecsignQueryByOutBillNoURL = "https://api.mch.weixin.qq.com/v3/fund-app/mch-transfer/elecsign/out-bill-no/%s"

	// MchTransferElecsignApplyByBillNoURL 按微信转账单号申请电子回单地址
	MchTransferElecsignApplyByBillNoURL = "https://api.mch.weixin.qq.com/v3/fund-app/mch-transfer/elecsign/transfer-bill-no"

	// MchTransferElecsignQueryByBillNoURL 按微信转账单号查询电子回单地址（需格式化 transfer_bill_no）
	MchTransferElecsignQueryByBillNoURL = "https://api.mch.weixin.qq.com/v3/fund-app/mch-transfer/elecsign/transfer-bill-no/%s"
)

// 转账单据状态（TransferBillResponse.State / TransferBillCreateResponse.State）
const (
	TransferStateAccepted        = "ACCEPTED"          // 转账已受理（非终态，可原单重试）
	TransferStateProcessing      = "PROCESSING"        // 转账处理中，结果尚未明确（建议检查账户余额）
	TransferStateWaitUserConfirm = "WAIT_USER_CONFIRM" // 待收款用户确认（可拉起收款确认页）
	TransferStateTransfering     = "TRANSFERING"       // 转账结果尚未明确（可拉起收款确认页重试确认）
	TransferStateSuccess         = "SUCCESS"           // 转账成功
	TransferStateFail            = "FAIL"              // 转账失败
	TransferStateCanceling       = "CANCELING"         // 撤销请求受理成功，撤销中
	TransferStateCancelled       = "CANCELLED"         // 撤销完成
)

// 电子回单申请单状态（TransferElecsignResponse.State）
const (
	TransferElecsignStateGenerating = "GENERATING" // 电子回单受理成功，处理中
	TransferElecsignStateFinished   = "FINISHED"   // 电子回单处理完成（返回 hash 与下载地址）
	TransferElecsignStateFailed     = "FAILED"     // 电子回单生成失败（fail_reason 返回原因）
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
//   - 商户处理回调成功后必须应答 HTTP 204 + 空响应体（官方规范，勿返回 {"code":"SUCCESS"} 之类 JSON），
//     否则微信会按 15s/15s/30s/3m/10m/20m/30m/30m/30m/60m/3h/3h/3h/6h/6h/6h 的频率重试 15 次，
//     应答构造见 NotificationSuccess / NotificationFail
//   - 处理失败需返回 4xx/5xx，响应体仅允许 {"code":"FAIL","message":"..."}（FAIL 为唯一取值）
const (
	// NotificationAckCodeFail 处理失败应答的 code 值（官方唯一取值）
	NotificationAckCodeFail = "FAIL"
)
