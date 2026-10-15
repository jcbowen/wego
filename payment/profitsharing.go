package payment

import (
	"context"
	"fmt"
	"net/url"
)

// ProfitSharing 请求分账
//
// 参数:
//   - ctx: 上下文
//   - req: 分账请求参数，Receivers 中的 Name 需为加密后的密文
//
// 返回:
//   - *ProfitSharingResponse: 分账单受理结果，State 为 PROCESSING 时需继续查询
//   - error: 参数缺失、加密失败或请求失败时返回错误
//
// 说明:
//   - 接口为异步处理，最终分账结果需用 ProfitSharingQuery 查询
//   - 请求头 Wechatpay-Serial 会自动填充为实际加密所用的公钥 ID 或平台证书序列号
//
// 使用示例:
//   - name, _ := client.EncryptSensitiveData(ctx, "张三")
//   - resp, err := client.ProfitSharing(ctx, &payment.ProfitSharingRequest{
//   - TransactionID: "4200001234", OutOrderNo: "PS001", UnfreezeUnsplit: true,
//   - Receivers: []payment.ProfitSharingReceiver{{Type: "MERCHANT_ID", Account: "1900000109", Name: name, Amount: 100, Description: "分账"}},
//   - })
//
// 文档:
//   - https://pay.wechatpay.cn/doc/v3/merchant/4012524936
func (c *Client) ProfitSharing(ctx context.Context, req *ProfitSharingRequest) (*ProfitSharingResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if req.TransactionID == "" {
		return nil, fmt.Errorf("微信支付订单号不能为空")
	}
	if req.OutOrderNo == "" {
		return nil, fmt.Errorf("商户分账单号不能为空")
	}
	if len(req.Receivers) == 0 {
		return nil, fmt.Errorf("分账接收方列表不能为空")
	}
	for i, receiver := range req.Receivers {
		if receiver.Type == "" {
			return nil, fmt.Errorf("第 %d 个分账接收方的类型不能为空", i+1)
		}
		if receiver.Account == "" {
			return nil, fmt.Errorf("第 %d 个分账接收方的账号不能为空", i+1)
		}
		if receiver.Amount <= 0 {
			return nil, fmt.Errorf("第 %d 个分账接收方的金额必须大于 0", i+1)
		}
	}

	// 请求头 Wechatpay-Serial 仅在报文中含加密字段时才需要声明
	//
	// 接收方类型为 MERCHANT_ID 时无需加密姓名，若此时仍强制获取加密密钥，
	// 仅配置平台证书但证书下载失败的商户会被无谓阻断
	serial := ""
	if hasEncryptedReceiverName(req.Receivers) {
		s, err := c.EncryptSerialNo(ctx)
		if err != nil {
			return nil, err
		}
		serial = s
	}

	var result ProfitSharingResponse
	if err := c.doRequestWithSerial(ctx, "POST", ProfitSharingOrderURL, req, &result, true, serial); err != nil {
		return nil, err
	}

	return &result, nil
}

// hasEncryptedReceiverName 判断分账接收方中是否含需加密的姓名
//
// 参数:
//   - receivers: 分账接收方列表
//
// 返回:
//   - bool: 存在非空 Name 时返回 true
//
// 说明:
//   - 接收方类型为 PERSONAL_OPENID（个人 openid）时，姓名为必填且需加密
//   - 接收方类型为 MERCHANT_ID（商户号）时，姓名无需填写也无需加密
func hasEncryptedReceiverName(receivers []ProfitSharingReceiver) bool {
	for _, receiver := range receivers {
		if receiver.Name != "" {
			return true
		}
	}
	return false
}

// ProfitSharingQuery 查询分账结果
//
// 参数:
//   - ctx: 上下文
//   - outOrderNo: 商户分账单号
//   - transactionID: 微信支付订单号
//
// 返回:
//   - *ProfitSharingResponse: 分账单结果，State 为 FINISHED 表示处理完毕
//   - error: 请求失败时返回错误
//
// 使用示例:
//   - resp, err := client.ProfitSharingQuery(ctx, "PS001", "4200001234")
//
// 文档:
//   - https://pay.wechatpay.cn/doc/v3/merchant/4012524940
func (c *Client) ProfitSharingQuery(ctx context.Context, outOrderNo, transactionID string) (*ProfitSharingResponse, error) {
	if outOrderNo == "" {
		return nil, fmt.Errorf("商户分账单号不能为空")
	}
	if transactionID == "" {
		return nil, fmt.Errorf("微信支付订单号不能为空")
	}

	apiURL := fmt.Sprintf(ProfitSharingQueryURL, url.PathEscape(outOrderNo)) +
		"?mchid=" + url.QueryEscape(c.config.MchID) +
		"&transaction_id=" + url.QueryEscape(transactionID)

	var result ProfitSharingResponse
	if err := c.doRequest(ctx, "GET", apiURL, nil, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// ProfitSharingUnfreeze 解冻剩余未分账资金
//
// 参数:
//   - ctx: 上下文
//   - req: 解冻请求参数
//
// 返回:
//   - *ProfitSharingResponse: 解冻结果
//   - error: 参数缺失或请求失败时返回错误
//
// 使用示例:
//   - resp, err := client.ProfitSharingUnfreeze(ctx, &payment.ProfitSharingUnfreezeRequest{TransactionID: "4200001234", OutOrderNo: "PS001", Description: "解冻"})
//
// 文档:
//   - https://pay.wechatpay.cn/doc/v3/merchant/4012524945
func (c *Client) ProfitSharingUnfreeze(ctx context.Context, req *ProfitSharingUnfreezeRequest) (*ProfitSharingResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if req.TransactionID == "" {
		return nil, fmt.Errorf("微信支付订单号不能为空")
	}
	if req.OutOrderNo == "" {
		return nil, fmt.Errorf("商户分账单号不能为空")
	}

	var result ProfitSharingResponse
	if err := c.doRequest(ctx, "POST", ProfitSharingUnfreezeURL, req, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// ProfitSharingReturn 请求分账回退
//
// 参数:
//   - ctx: 上下文
//   - req: 回退请求参数，OrderID 与 OutOrderNo 至少填一个
//
// 返回:
//   - *ProfitSharingReturnResponse: 回退受理结果
//   - error: 参数缺失或请求失败时返回错误
//
// 使用示例:
//   - resp, err := client.ProfitSharingReturn(ctx, &payment.ProfitSharingReturnRequest{OutOrderNo: "PS001", OutReturnNo: "RT001", ReturnMchID: "1900000109", Amount: 100, Description: "分账回退"})
//
// 文档:
//   - https://pay.wechatpay.cn/doc/v3/merchant/4012524949
func (c *Client) ProfitSharingReturn(ctx context.Context, req *ProfitSharingReturnRequest) (*ProfitSharingReturnResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if req.OrderID == "" && req.OutOrderNo == "" {
		return nil, fmt.Errorf("微信分账单号与商户分账单号不能同时为空")
	}
	if req.OutReturnNo == "" {
		return nil, fmt.Errorf("商户回退单号不能为空")
	}
	if req.ReturnMchID == "" {
		return nil, fmt.Errorf("回退商户号不能为空")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("回退金额必须大于 0，当前为 %d 分", req.Amount)
	}

	var result ProfitSharingReturnResponse
	if err := c.doRequest(ctx, "POST", ProfitSharingReturnURL, req, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// ProfitSharingReturnQuery 查询分账回退结果
//
// 参数:
//   - ctx: 上下文
//   - outReturnNo: 商户回退单号
//   - outOrderNo: 商户分账单号
//
// 返回:
//   - *ProfitSharingReturnResponse: 回退结果，Result 为 SUCCESS 表示回退成功
//   - error: 请求失败时返回错误
//
// 使用示例:
//   - resp, err := client.ProfitSharingReturnQuery(ctx, "RT001", "PS001")
//
// 文档:
//   - https://pay.wechatpay.cn/doc/v3/merchant/4012524953
func (c *Client) ProfitSharingReturnQuery(ctx context.Context, outReturnNo, outOrderNo string) (*ProfitSharingReturnResponse, error) {
	if outReturnNo == "" {
		return nil, fmt.Errorf("商户回退单号不能为空")
	}
	if outOrderNo == "" {
		return nil, fmt.Errorf("商户分账单号不能为空")
	}

	apiURL := fmt.Sprintf(ProfitSharingReturnQueryURL, url.PathEscape(outReturnNo)) +
		"?out_order_no=" + url.QueryEscape(outOrderNo)

	var result ProfitSharingReturnResponse
	if err := c.doRequest(ctx, "GET", apiURL, nil, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// ProfitSharingAmounts 查询订单剩余待分金额
//
// 参数:
//   - ctx: 上下文
//   - transactionID: 微信支付订单号
//
// 返回:
//   - *ProfitSharingAmountsResponse: 含剩余待分金额的响应
//   - error: 请求失败时返回错误
//
// 使用示例:
//   - resp, err := client.ProfitSharingAmounts(ctx, "4200001234")
//
// 文档:
//   - https://pay.wechatpay.cn/doc/v3/merchant/4012524957
func (c *Client) ProfitSharingAmounts(ctx context.Context, transactionID string) (*ProfitSharingAmountsResponse, error) {
	if transactionID == "" {
		return nil, fmt.Errorf("微信支付订单号不能为空")
	}

	apiURL := fmt.Sprintf(ProfitSharingAmountsURL, url.PathEscape(transactionID))

	var result ProfitSharingAmountsResponse
	if err := c.doRequest(ctx, "GET", apiURL, nil, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}
