package payment

import "time"

// Amount 订单金额
type Amount struct {
	Total    int    `json:"total"`              // 订单总金额，单位为分
	Currency string `json:"currency,omitempty"` // 货币类型，默认 CNY
}

// RefundAmount 退款金额
type RefundAmount struct {
	Refund int `json:"refund"` // 退款金额，单位为分
	Total  int `json:"total"`  // 原订单金额，单位为分
}

// Payer JSAPI 支付 payer 信息
type Payer struct {
	OpenID string `json:"openid"` // 用户在直连商户应用下的用户标示
}

// H5Info H5 支付场景信息
type H5Info struct {
	Type        string `json:"type"`                   // 场景类型，如 Wap、iOS、Android
	AppName     string `json:"app_name,omitempty"`     // 应用名称
	AppURL      string `json:"app_url,omitempty"`      // 网站 URL
	BundleID    string `json:"bundle_id,omitempty"`    // iOS Bundle ID
	PackageName string `json:"package_name,omitempty"` // Android 包名
}

// SceneInfo 支付场景信息
type SceneInfo struct {
	PayerClientIP string  `json:"payer_client_ip"`   // 用户终端 IP
	H5Info        *H5Info `json:"h5_info,omitempty"` // H5 场景信息
}

// JSAPIRequest JSAPI 统一下单请求
type JSAPIRequest struct {
	AppID       string     `json:"appid"`                // 应用 ID
	MchID       string     `json:"mchid"`                // 商户号
	Description string     `json:"description"`          // 商品描述
	OutTradeNo  string     `json:"out_trade_no"`         // 商户订单号
	NotifyURL   string     `json:"notify_url"`           // 通知地址
	Amount      Amount     `json:"amount"`               // 订单金额
	Payer       Payer      `json:"payer"`                // 支付者信息
	Attach      string     `json:"attach,omitempty"`     // 附加数据
	SceneInfo   *SceneInfo `json:"scene_info,omitempty"` // 场景信息
}

// NativeRequest Native 支付下单请求
type NativeRequest struct {
	AppID       string     `json:"appid"`                // 应用 ID
	MchID       string     `json:"mchid"`                // 商户号
	Description string     `json:"description"`          // 商品描述
	OutTradeNo  string     `json:"out_trade_no"`         // 商户订单号
	NotifyURL   string     `json:"notify_url"`           // 通知地址
	Amount      Amount     `json:"amount"`               // 订单金额
	Attach      string     `json:"attach,omitempty"`     // 附加数据
	GoodsTag    string     `json:"goods_tag,omitempty"`  // 订单优惠标记
	SceneInfo   *SceneInfo `json:"scene_info,omitempty"` // 场景信息
}

// AppRequest App 支付下单请求
type AppRequest struct {
	AppID       string     `json:"appid"`                // 应用 ID
	MchID       string     `json:"mchid"`                // 商户号
	Description string     `json:"description"`          // 商品描述
	OutTradeNo  string     `json:"out_trade_no"`         // 商户订单号
	NotifyURL   string     `json:"notify_url"`           // 通知地址
	Amount      Amount     `json:"amount"`               // 订单金额
	Attach      string     `json:"attach,omitempty"`     // 附加数据
	GoodsTag    string     `json:"goods_tag,omitempty"`  // 订单优惠标记
	SceneInfo   *SceneInfo `json:"scene_info,omitempty"` // 场景信息
}

// H5Request H5 支付请求
type H5Request struct {
	AppID       string     `json:"appid"`        // 应用 ID
	MchID       string     `json:"mchid"`        // 商户号
	Description string     `json:"description"`  // 商品描述
	OutTradeNo  string     `json:"out_trade_no"` // 商户订单号
	NotifyURL   string     `json:"notify_url"`   // 通知地址
	Amount      Amount     `json:"amount"`       // 订单金额
	SceneInfo   *SceneInfo `json:"scene_info"`   // H5 场景信息必填
}

// CodePayer 付款码支付支付者信息
type CodePayer struct {
	AuthCode string `json:"auth_code"` // 顾客的付款码（扫码枪读取，1~64 位数字，10~15 开头为微信付款码）
}

// CodePayStoreInfo 付款码支付门店信息
type CodePayStoreInfo struct {
	ID       string `json:"id,omitempty"`        // 门店编号
	Name     string `json:"name,omitempty"`      // 门店名称
	AreaCode string `json:"area_code,omitempty"` // 门店所在地行政区划码
	Address  string `json:"address,omitempty"`   // 门店详细地址
}

// CodePaySceneInfo 付款码支付场景信息
type CodePaySceneInfo struct {
	PayerClientIP string            `json:"payer_client_ip,omitempty"` // 调用支付接口的终端 IP
	StoreInfo     *CodePayStoreInfo `json:"store_info,omitempty"`      // 门店信息
}

// CodePayRequest 付款码支付下单请求（B扫C）
//
// 说明:
//   - amount.total 必须大于 0，付款码支付不支持 0 元下单
//   - 提交后同步返回订单结果（结构与查单一致）：trade_state=SUCCESS 表示扣款成功，
//     USERPAYING 表示用户输入密码中，需用 QueryOrder 轮询最终状态
//   - USERPAYING 超时不再等待时，需调用撤销接口终止交易，避免用户后续输入密码意外扣款
type CodePayRequest struct {
	AppID       string            `json:"appid"`                // 应用 ID
	MchID       string            `json:"mchid"`                // 商户号
	Description string            `json:"description"`          // 商品描述
	OutTradeNo  string            `json:"out_trade_no"`         // 商户订单号
	NotifyURL   string            `json:"notify_url"`           // 支付结果通知地址
	Amount      Amount            `json:"amount"`               // 订单金额
	Payer       CodePayer         `json:"payer"`                // 支付者信息（付款码）
	Attach      string            `json:"attach,omitempty"`     // 附加数据
	GoodsTag    string            `json:"goods_tag,omitempty"`  // 订单优惠标记
	SceneInfo   *CodePaySceneInfo `json:"scene_info,omitempty"` // 场景信息
}

// CloseOrderRequest 关闭订单请求
type CloseOrderRequest struct {
	MchID string `json:"mchid"` // 商户号
}

// PrepayResponse JSAPI 统一下单响应
type PrepayResponse struct {
	PrepayID string `json:"prepay_id"` // 预支付交易会话标识
}

// JSAPIPayParams 前端调起 JSAPI 支付所需的参数
//
// 字段说明:
//   - AppID: 应用 ID（公众号/小程序 appid）
//   - TimeStamp: 时间戳，秒级字符串
//   - NonceStr: 随机字符串
//   - Package: 固定格式 prepay_id=xxx
//   - SignType: 签名类型，固定 RSA
//   - PaySign: 使用商户私钥对上述字段签名后的结果
//
// 使用示例:
//   - 小程序 wx.requestPayment 与公众号 JS-SDK chooseWXPay 均可直接使用
type JSAPIPayParams struct {
	AppID     string `json:"appId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

// H5Response H5 支付响应
type H5Response struct {
	H5URL string `json:"h5_url"` // H5 支付跳转 URL
}

// NativeResponse Native 支付下单响应
type NativeResponse struct {
	CodeURL string `json:"code_url"` // 二维码链接，商户需将其转换为二维码图片
}

// AppPayParams App 端调起支付所需的参数
//
// 说明:
//   - 字段名大小写敏感，需与微信开放平台 SDK 要求一致
//   - Package 官方暂固定为 Sign=WXPay
type AppPayParams struct {
	AppID     string `json:"appid"`
	PartnerID string `json:"partnerid"`
	PrepayID  string `json:"prepayid"`
	Package   string `json:"package"`
	NonceStr  string `json:"noncestr"`
	Timestamp string `json:"timestamp"`
	Sign      string `json:"sign"`
}

// Order 查询订单响应
type Order struct {
	AppID          string    `json:"appid"`                  // 应用 ID
	MchID          string    `json:"mchid"`                  // 商户号
	OutTradeNo     string    `json:"out_trade_no"`           // 商户订单号
	TransactionID  string    `json:"transaction_id"`         // 微信支付订单号
	TradeType      string    `json:"trade_type,omitempty"`   // 交易类型
	TradeState     string    `json:"trade_state"`            // 交易状态
	TradeStateDesc string    `json:"trade_state_desc"`       // 交易状态描述
	BankType       string    `json:"bank_type,omitempty"`    // 付款银行
	Attach         string    `json:"attach,omitempty"`       // 附加数据
	SuccessTime    time.Time `json:"success_time,omitempty"` // 支付完成时间
	Amount         *Amount   `json:"amount,omitempty"`       // 订单金额
	Payer          *Payer    `json:"payer,omitempty"`        // 支付者信息
}

// RefundRequest 退款申请请求
type RefundRequest struct {
	OutTradeNo  string       `json:"out_trade_no"`         // 商户订单号
	OutRefundNo string       `json:"out_refund_no"`        // 商户退款单号
	Reason      string       `json:"reason,omitempty"`     // 退款原因
	NotifyURL   string       `json:"notify_url,omitempty"` // 退款结果通知地址
	Amount      RefundAmount `json:"amount"`               // 退款金额
}

// RefundResponse 退款申请与退款查询响应
//
// 说明:
//   - 申请退款接口仅返回部分字段，查询退款接口会额外返回 CreateTime、SuccessTime、FundsAccount
type RefundResponse struct {
	RefundID            string       `json:"refund_id"`               // 微信退款单号
	OutRefundNo         string       `json:"out_refund_no"`           // 商户退款单号
	TransactionID       string       `json:"transaction_id"`          // 微信支付订单号
	OutTradeNo          string       `json:"out_trade_no"`            // 商户订单号
	Channel             string       `json:"channel"`                 // 退款渠道
	UserReceivedAccount string       `json:"user_received_account"`   // 用户退款入账账户
	CreateTime          time.Time    `json:"create_time,omitempty"`   // 退款受理时间
	SuccessTime         time.Time    `json:"success_time,omitempty"`  // 退款成功时间
	FundsAccount        string       `json:"funds_account,omitempty"` // 退款出资账户
	Status              string       `json:"status"`                  // 退款状态
	Amount              RefundAmount `json:"amount"`                  // 退款金额
}

// EncryptCertificate 加密证书信息
type EncryptCertificate struct {
	Algorithm      string `json:"algorithm"`       // 加密算法
	Nonce          string `json:"nonce"`           // 加密使用的随机串
	AssociatedData string `json:"associated_data"` // 加密使用的附加数据
	Ciphertext     string `json:"ciphertext"`      // Base64 编码的密文
}

// CertificateItem 平台证书列表项
type CertificateItem struct {
	SerialNo           string             `json:"serial_no"`           // 证书序列号
	EffectiveTime      string             `json:"effective_time"`      // 生效时间
	ExpireTime         string             `json:"expire_time"`         // 过期时间
	EncryptCertificate EncryptCertificate `json:"encrypt_certificate"` // 加密证书信息
}

// CertificatesResponse 获取平台证书响应
type CertificatesResponse struct {
	Data []CertificateItem `json:"data"` // 证书列表
}

// Resource 回调通知资源对象
type Resource struct {
	Algorithm      string `json:"algorithm"`       // 加密算法
	Ciphertext     string `json:"ciphertext"`      // Base64 编码的密文
	Nonce          string `json:"nonce"`           // 加密随机串
	AssociatedData string `json:"associated_data"` // 加密附加数据
	OriginalType   string `json:"original_type"`   // 原始类型
}

// BillRequest 账单申请请求
//
// 字段说明:
//   - BillDate: 账单日期，格式 yyyy-MM-dd
//   - BillType: 账单类型，ALL / SUCCESS / REFUND，不填默认 ALL
//   - TarType: 压缩类型，GZIP，不填则不压缩
//   - AccountType: 资金账户类型，仅资金账单使用，BASIC / OPERATION / FEES
type BillRequest struct {
	BillDate    string `json:"bill_date"`              // 账单日期
	BillType    string `json:"bill_type,omitempty"`    // 账单类型
	TarType     string `json:"tar_type,omitempty"`     // 压缩类型
	AccountType string `json:"account_type,omitempty"` // 资金账户类型
}

// BillResponse 账单申请响应
type BillResponse struct {
	HashType    string `json:"hash_type"`    // 哈希类型，固定 SHA1
	HashValue   string `json:"hash_value"`   // 账单文件哈希值，用于校验完整性
	DownloadURL string `json:"download_url"` // 账单下载地址，5 分钟内有效
}

// ProfitSharingReceiver 分账接收方
//
// 说明:
//   - Name 为敏感字段，需使用微信支付公钥或平台证书加密后传入
//   - Type 取值：MERCHANT_ID（商户号）、PERSONAL_OPENID（个人 openid）
type ProfitSharingReceiver struct {
	Type        string `json:"type"`           // 分账接收方类型
	Account     string `json:"account"`        // 分账接收方账号
	Name        string `json:"name,omitempty"` // 分账接收方姓名（需加密）
	Amount      int    `json:"amount"`         // 分账金额，单位为分
	Description string `json:"description"`    // 分账描述
}

// ProfitSharingRequest 请求分账参数
type ProfitSharingRequest struct {
	AppID           string                  `json:"appid,omitempty"`  // 应用 ID，含个人接收方时必填
	TransactionID   string                  `json:"transaction_id"`   // 微信支付订单号
	OutOrderNo      string                  `json:"out_order_no"`     // 商户分账单号
	Receivers       []ProfitSharingReceiver `json:"receivers"`        // 分账接收方列表
	UnfreezeUnsplit bool                    `json:"unfreeze_unsplit"` // 是否解冻剩余未分账资金
}

// ProfitSharingReceiverResult 分账接收方结果
type ProfitSharingReceiverResult struct {
	Amount      int    `json:"amount"`                // 分账金额，单位为分
	Description string `json:"description"`           // 分账描述
	Type        string `json:"type"`                  // 接收方类型
	Account     string `json:"account"`               // 接收方账号
	Result      string `json:"result"`                // 分账结果，PENDING / SUCCESS / CLOSED
	FailReason  string `json:"fail_reason,omitempty"` // 分账失败原因
	DetailID    string `json:"detail_id,omitempty"`   // 分账明细单号
}

// ProfitSharingResponse 分账请求与查询结果响应
type ProfitSharingResponse struct {
	TransactionID string                        `json:"transaction_id"` // 微信支付订单号
	OutOrderNo    string                        `json:"out_order_no"`   // 商户分账单号
	OrderID       string                        `json:"order_id"`       // 微信分账单号
	State         string                        `json:"state"`          // 分账单状态，PROCESSING / FINISHED
	Receivers     []ProfitSharingReceiverResult `json:"receivers"`      // 分账接收方结果列表
}

// ProfitSharingReturnRequest 请求分账回退参数
type ProfitSharingReturnRequest struct {
	OrderID     string `json:"order_id,omitempty"`     // 微信分账单号，与 OutOrderNo 二选一
	OutOrderNo  string `json:"out_order_no,omitempty"` // 商户分账单号
	OutReturnNo string `json:"out_return_no"`          // 商户回退单号
	ReturnMchID string `json:"return_mchid"`           // 回退商户号
	Amount      int    `json:"amount"`                 // 回退金额，单位为分
	Description string `json:"description"`            // 回退描述
}

// ProfitSharingReturnResponse 分账回退请求与查询结果响应
type ProfitSharingReturnResponse struct {
	OrderID     string `json:"order_id,omitempty"`    // 微信分账单号
	OutOrderNo  string `json:"out_order_no"`          // 商户分账单号
	OutReturnNo string `json:"out_return_no"`         // 商户回退单号
	ReturnMchID string `json:"return_mchid"`          // 回退商户号
	Amount      int    `json:"amount"`                // 回退金额，单位为分
	Description string `json:"description"`           // 回退描述
	Result      string `json:"result"`                // 回退结果，PROCESSING / SUCCESS / FAILED
	FailReason  string `json:"fail_reason,omitempty"` // 回退失败原因
}

// ProfitSharingUnfreezeRequest 解冻剩余资金参数
type ProfitSharingUnfreezeRequest struct {
	TransactionID string `json:"transaction_id"` // 微信支付订单号
	OutOrderNo    string `json:"out_order_no"`   // 商户分账单号
	Description   string `json:"description"`    // 解冻描述
}

// ProfitSharingAmountsResponse 查询剩余待分金额响应
type ProfitSharingAmountsResponse struct {
	TransactionID string `json:"transaction_id"` // 微信支付订单号
	UnsplitAmount int    `json:"unsplit_amount"` // 剩余待分金额，单位为分
}

// Notification 回调通知结构
type Notification struct {
	ID           string   `json:"id"`            // 通知 ID
	CreateTime   string   `json:"create_time"`   // 通知创建时间
	EventType    string   `json:"event_type"`    // 通知类型
	ResourceType string   `json:"resource_type"` // 资源类型
	Resource     Resource `json:"resource"`      // 加密资源
	Summary      string   `json:"summary"`       // 回调摘要
}
