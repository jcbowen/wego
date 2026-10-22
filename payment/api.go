package payment

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// validateOrderRequest 校验下单公共必填参数
//
// 参数:
//   - outTradeNo: 商户订单号
//   - description: 商品描述
//   - total: 订单总金额，单位为分
//
// 返回:
//   - error: 参数不合法时返回错误
//
// 说明:
//   - 提前在本地拦截明显非法的参数，可避免一次无效网络往返
//   - 商户订单号建议仅使用大小写字母、数字、下划线与短横线，避免 URL 转义带来的排查成本
func validateOrderRequest(outTradeNo, description string, total int) error {
	if outTradeNo == "" {
		return fmt.Errorf("商户订单号不能为空")
	}
	if description == "" {
		return fmt.Errorf("商品描述不能为空")
	}
	if total <= 0 {
		return fmt.Errorf("订单总金额必须大于 0，当前为 %d 分", total)
	}
	return nil
}

// JSAPIPay JSAPI 统一下单
//
// 参数:
//   - ctx: 上下文
//   - req: JSAPI 下单请求参数
//
// 返回:
//   - *PrepayResponse: 包含 prepay_id 的响应
//   - error: 请求失败时返回错误
//
// 文档:
//   - https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_1_1.shtml
func (c *Client) JSAPIPay(ctx context.Context, req *JSAPIRequest) (*PrepayResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if err := validateOrderRequest(req.OutTradeNo, req.Description, req.Amount.Total); err != nil {
		return nil, err
	}
	if req.Payer.OpenID == "" {
		return nil, fmt.Errorf("支付者openid不能为空")
	}

	req.AppID = defaultString(req.AppID, c.config.AppID)
	req.MchID = defaultString(req.MchID, c.config.MchID)
	req.NotifyURL = defaultString(req.NotifyURL, c.config.NotifyURL)

	var result PrepayResponse
	if err := c.doRequest(ctx, "POST", PayJSAPIURL, req, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// H5Pay H5 支付统一下单
//
// 参数:
//   - ctx: 上下文
//   - req: H5 下单请求参数
//
// 返回:
//   - *H5Response: 包含 h5_url 的响应
//   - error: 请求失败时返回错误
//
// 文档:
//   - https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_3_1.shtml
func (c *Client) H5Pay(ctx context.Context, req *H5Request) (*H5Response, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if req.SceneInfo == nil || req.SceneInfo.H5Info == nil {
		return nil, fmt.Errorf("H5场景信息不能为空")
	}
	if err := validateOrderRequest(req.OutTradeNo, req.Description, req.Amount.Total); err != nil {
		return nil, err
	}

	req.AppID = defaultString(req.AppID, c.config.AppID)
	req.MchID = defaultString(req.MchID, c.config.MchID)
	req.NotifyURL = defaultString(req.NotifyURL, c.config.NotifyURL)

	var result H5Response
	if err := c.doRequest(ctx, "POST", PayH5URL, req, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// NativePay Native 支付下单（扫码支付）
//
// 参数:
//   - ctx: 上下文
//   - req: Native 下单请求参数
//
// 返回:
//   - *NativeResponse: 包含 code_url 的响应，商户需将其转换为二维码
//   - error: 请求失败时返回错误
//
// 使用示例:
//   - resp, err := client.NativePay(ctx, &payment.NativeRequest{OutTradeNo: "ORDER_001", Amount: payment.Amount{Total: 100}, Description: "商品"})
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012791880
func (c *Client) NativePay(ctx context.Context, req *NativeRequest) (*NativeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if err := validateOrderRequest(req.OutTradeNo, req.Description, req.Amount.Total); err != nil {
		return nil, err
	}

	req.AppID = defaultString(req.AppID, c.config.AppID)
	req.MchID = defaultString(req.MchID, c.config.MchID)
	req.NotifyURL = defaultString(req.NotifyURL, c.config.NotifyURL)

	var result NativeResponse
	if err := c.doRequest(ctx, "POST", PayNativeURL, req, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// AppPay App 支付下单
//
// 参数:
//   - ctx: 上下文
//   - req: App 下单请求参数
//
// 返回:
//   - *PrepayResponse: 包含 prepay_id 的响应，需再用 BuildAppPayParams 生成调起参数
//   - error: 请求失败时返回错误
//
// 使用示例:
//   - prepay, err := client.AppPay(ctx, &payment.AppRequest{OutTradeNo: "ORDER_001", Amount: payment.Amount{Total: 100}, Description: "商品"})
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012791876
func (c *Client) AppPay(ctx context.Context, req *AppRequest) (*PrepayResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if err := validateOrderRequest(req.OutTradeNo, req.Description, req.Amount.Total); err != nil {
		return nil, err
	}

	req.AppID = defaultString(req.AppID, c.config.AppID)
	req.MchID = defaultString(req.MchID, c.config.MchID)
	req.NotifyURL = defaultString(req.NotifyURL, c.config.NotifyURL)

	var result PrepayResponse
	if err := c.doRequest(ctx, "POST", PayAppURL, req, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// CodePay 付款码支付下单（B扫C：收银员用扫码设备读取顾客展示的付款码收款）
//
// 参数:
//   - ctx: 上下文
//   - req: 付款码支付下单请求（Payer.AuthCode 为顾客付款码，必填）
//
// 返回:
//   - *Order: 订单结果，与查单接口同构；TradeState=SUCCESS 表示扣款成功（TransactionID 有值），
//     TradeState=USERPAYING 表示用户支付中，需用 QueryOrder 轮询最终状态
//   - error: 请求失败时返回错误；扣款失败（如余额不足）返回 *APIError，业务码可据此判断是否换码重扫
//
// 使用示例:
//   - order, err := client.CodePay(ctx, &payment.CodePayRequest{OutTradeNo: "ORDER_001", Amount: payment.Amount{Total: 100}, Description: "商品", Payer: payment.CodePayer{AuthCode: "1345235315201154982642655744"}})
//   - if order.TradeState == payment.TradeStateUserPaying { /* 轮询查单 */ }
//
// 文档:
//   - https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_5_1.shtml
func (c *Client) CodePay(ctx context.Context, req *CodePayRequest) (*Order, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if err := validateOrderRequest(req.OutTradeNo, req.Description, req.Amount.Total); err != nil {
		return nil, err
	}
	if req.Payer.AuthCode == "" {
		return nil, fmt.Errorf("付款码不能为空")
	}

	req.AppID = defaultString(req.AppID, c.config.AppID)
	req.MchID = defaultString(req.MchID, c.config.MchID)
	req.NotifyURL = defaultString(req.NotifyURL, c.config.NotifyURL)

	var result Order
	if err := c.doRequest(ctx, "POST", PayCodePayURL, req, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// BuildAppPayParams 生成 App 端调起支付所需的参数
//
// 参数:
//   - prepayID: AppPay 返回的预支付交易会话标识
//
// 返回:
//   - *AppPayParams: 含 sign 的调起支付参数，可直接交给客户端 SDK
//   - error: prepayID 为空或签名失败时返回错误
//
// 说明:
//   - 官方签名串为四行：appid\ntimestamp\nnoncestr\nprepayid\n，不包含 partnerid 与 package
//   - 客户端字段名为 appid、partnerid、prepayid、package、noncestr、timestamp、sign，大小写敏感
//
// 使用示例:
//   - prepay, _ := client.AppPay(ctx, req)
//   - params, err := client.BuildAppPayParams(prepay.PrepayID)
//
// 文档:
//   - https://pay.weixin.qq.com/doc/global/v3/zh/4012354778
func (c *Client) BuildAppPayParams(prepayID string) (*AppPayParams, error) {
	if prepayID == "" {
		return nil, fmt.Errorf("prepay_id不能为空")
	}

	appID := c.config.AppID
	if appID == "" {
		return nil, fmt.Errorf("AppID不能为空")
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce(32)

	message := strings.Join([]string{appID, timestamp, nonce, prepayID}, "\n") + "\n"
	sign, err := c.signSHA256WithRSA(message)
	if err != nil {
		return nil, err
	}

	return &AppPayParams{
		AppID:     appID,
		PartnerID: c.config.MchID,
		PrepayID:  prepayID,
		Package:   AppPayPackageValue,
		NonceStr:  nonce,
		Timestamp: timestamp,
		Sign:      sign,
	}, nil
}

// QueryOrder 按商户订单号查询订单
//
// 参数:
//   - ctx: 上下文
//   - outTradeNo: 商户订单号
//
// 返回:
//   - *Order: 订单信息
//   - error: 请求失败时返回错误
//
// 文档:
//   - https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_1_2.shtml
func (c *Client) QueryOrder(ctx context.Context, outTradeNo string) (*Order, error) {
	if outTradeNo == "" {
		return nil, fmt.Errorf("商户订单号不能为空")
	}

	apiURL := fmt.Sprintf(PayQueryOrderURL, url.PathEscape(outTradeNo)) + "?mchid=" + url.QueryEscape(c.config.MchID)

	var result Order
	if err := c.doRequest(ctx, "GET", apiURL, nil, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// QueryOrderByTransactionID 按微信支付订单号查询订单
//
// 参数:
//   - ctx: 上下文
//   - transactionID: 微信支付订单号
//
// 返回:
//   - *Order: 订单信息
//   - error: 请求失败时返回错误
//
// 使用示例:
//   - order, err := client.QueryOrderByTransactionID(ctx, "4200001234202401011234567890")
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012791872
func (c *Client) QueryOrderByTransactionID(ctx context.Context, transactionID string) (*Order, error) {
	if transactionID == "" {
		return nil, fmt.Errorf("微信支付订单号不能为空")
	}

	apiURL := fmt.Sprintf(PayQueryOrderByTransactionIDURL, url.PathEscape(transactionID)) + "?mchid=" + url.QueryEscape(c.config.MchID)

	var result Order
	if err := c.doRequest(ctx, "GET", apiURL, nil, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// CloseOrder 关闭未支付的订单
//
// 参数:
//   - ctx: 上下文
//   - outTradeNo: 商户订单号
//
// 返回:
//   - error: 关单失败时返回错误
//
// 说明:
//   - 适用于用户取消订单或订单超时未支付的场景
//   - 接口成功时微信返回 204 No Content，无响应体
//
// 使用示例:
//   - if err := client.CloseOrder(ctx, "ORDER_20240101_001"); err != nil { return err }
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012526915
func (c *Client) CloseOrder(ctx context.Context, outTradeNo string) error {
	if outTradeNo == "" {
		return fmt.Errorf("商户订单号不能为空")
	}

	apiURL := fmt.Sprintf(PayCloseOrderURL, url.PathEscape(outTradeNo))
	req := &CloseOrderRequest{MchID: c.config.MchID}

	return c.doRequest(ctx, "POST", apiURL, req, nil, true)
}

// ReverseOrder 撤销订单（付款码支付配套）
//
// 参数:
//   - ctx: 上下文
//   - outTradeNo: 商户订单号
//
// 返回:
//   - error: 撤销失败时返回错误
//
// 说明:
//   - 付款码交易处于 USERPAYING（用户支付中）时无法用 CloseOrder 关闭，只能撤销以终止在途扣款
//   - 对已支付订单撤销会失败（如 ORDER_PAID 类错误），撤销不退款——退钱必须走 Refund
//   - 接口成功时微信返回 204 No Content，无响应体
//
// 使用示例:
//   - if err := client.ReverseOrder(ctx, "ORDER_20240101_001"); err != nil { return err }
//
// 文档:
//   - https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_5_6.shtml
func (c *Client) ReverseOrder(ctx context.Context, outTradeNo string) error {
	if outTradeNo == "" {
		return fmt.Errorf("商户订单号不能为空")
	}

	apiURL := fmt.Sprintf(PayReverseOrderURL, url.PathEscape(outTradeNo))
	req := &CloseOrderRequest{MchID: c.config.MchID}

	return c.doRequest(ctx, "POST", apiURL, req, nil, true)
}

// Refund 申请退款
//
// 参数:
//   - ctx: 上下文
//   - req: 退款请求参数
//
// 返回:
//   - *RefundResponse: 退款结果
//   - error: 请求失败时返回错误
//
// 文档:
//   - https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_1_10.shtml
func (c *Client) Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if req.OutTradeNo == "" {
		return nil, fmt.Errorf("商户订单号不能为空")
	}
	if req.OutRefundNo == "" {
		return nil, fmt.Errorf("商户退款单号不能为空")
	}
	if req.Amount.Refund <= 0 {
		return nil, fmt.Errorf("退款金额必须大于 0，当前为 %d 分", req.Amount.Refund)
	}
	if req.Amount.Total <= 0 {
		return nil, fmt.Errorf("原订单金额必须大于 0，当前为 %d 分", req.Amount.Total)
	}
	if req.Amount.Refund > req.Amount.Total {
		return nil, fmt.Errorf("退款金额(%d分)不能大于原订单金额(%d分)", req.Amount.Refund, req.Amount.Total)
	}

	req.NotifyURL = defaultString(req.NotifyURL, c.config.NotifyURL)

	var result RefundResponse
	if err := c.doRequest(ctx, "POST", PayRefundURL, req, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// RefundQuery 按商户退款单号查询退款结果
//
// 参数:
//   - ctx: 上下文
//   - outRefundNo: 商户退款单号
//
// 返回:
//   - *RefundResponse: 退款结果，Status 为 SUCCESS 表示退款成功
//   - error: 请求失败时返回错误
//
// 使用示例:
//   - refund, err := client.RefundQuery(ctx, "REFUND_20240101_001")
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012791893
func (c *Client) RefundQuery(ctx context.Context, outRefundNo string) (*RefundResponse, error) {
	if outRefundNo == "" {
		return nil, fmt.Errorf("商户退款单号不能为空")
	}

	apiURL := fmt.Sprintf(PayRefundQueryURL, url.PathEscape(outRefundNo)) + "?mchid=" + url.QueryEscape(c.config.MchID)

	var result RefundResponse
	if err := c.doRequest(ctx, "GET", apiURL, nil, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// BuildJSAPIPayParams 生成前端调起 JSAPI 支付所需的参数
//
// 参数:
//   - prepayID: JSAPIPay 返回的预支付交易会话标识
//
// 返回:
//   - *JSAPIPayParams: 含 paySign 的调起支付参数，可直接交给小程序 wx.requestPayment 或公众号 JS-SDK
//   - error: prepayID 为空或签名失败时返回错误
//
// 说明:
//   - 签名串格式为：appId\ntimeStamp\nnonceStr\npackage\n
//   - 参数大小写需与微信要求一致，前端字段名为 appId、timeStamp、nonceStr、package、signType、paySign
//
// 使用示例:
//   - prepay, _ := client.JSAPIPay(ctx, req)
//   - params, err := client.BuildJSAPIPayParams(prepay.PrepayID)
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012791874
func (c *Client) BuildJSAPIPayParams(prepayID string) (*JSAPIPayParams, error) {
	if prepayID == "" {
		return nil, fmt.Errorf("prepay_id不能为空")
	}

	appID := c.config.AppID
	if appID == "" {
		return nil, fmt.Errorf("AppID不能为空")
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce(32)
	pkg := "prepay_id=" + prepayID

	message := strings.Join([]string{appID, timestamp, nonce, pkg}, "\n") + "\n"
	paySign, err := c.signSHA256WithRSA(message)
	if err != nil {
		return nil, err
	}

	return &JSAPIPayParams{
		AppID:     appID,
		TimeStamp: timestamp,
		NonceStr:  nonce,
		Package:   pkg,
		SignType:  JSAPISignType,
		PaySign:   paySign,
	}, nil
}

// defaultString 返回第一个非空字符串
func defaultString(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
