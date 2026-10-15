package payment

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
)

// buildBillURL 拼接账单申请地址与查询参数
//
// 参数:
//   - baseURL: 账单接口基础地址
//   - req: 账单申请参数
//
// 返回:
//   - string: 带查询参数的完整 URL
func buildBillURL(baseURL string, req *BillRequest) string {
	query := url.Values{}
	query.Set("bill_date", req.BillDate)
	if req.BillType != "" {
		query.Set("bill_type", req.BillType)
	}
	if req.TarType != "" {
		query.Set("tar_type", req.TarType)
	}
	if req.AccountType != "" {
		query.Set("account_type", req.AccountType)
	}
	return baseURL + "?" + query.Encode()
}

// ApplyTradeBill 申请交易账单
//
// 参数:
//   - ctx: 上下文
//   - req: 账单申请参数，BillDate 必填，格式 yyyy-MM-dd
//
// 返回:
//   - *BillResponse: 含下载链接与哈希值的响应
//   - error: 参数缺失或请求失败时返回错误
//
// 说明:
//   - 微信在次日 9 点开始生成前一天的对账单，建议次日 10 点后获取
//
// 使用示例:
//   - bill, err := client.ApplyTradeBill(ctx, &payment.BillRequest{BillDate: "2026-10-14", BillType: "ALL"})
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012791894
func (c *Client) ApplyTradeBill(ctx context.Context, req *BillRequest) (*BillResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if req.BillDate == "" {
		return nil, fmt.Errorf("账单日期不能为空")
	}

	var result BillResponse
	if err := c.doRequest(ctx, "GET", buildBillURL(BillTradeURL, req), nil, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// ApplyFundFlowBill 申请资金账单
//
// 参数:
//   - ctx: 上下文
//   - req: 账单申请参数，BillDate 必填，AccountType 常用 BASIC / OPERATION / FEES
//
// 返回:
//   - *BillResponse: 含下载链接与哈希值的响应
//   - error: 参数缺失或请求失败时返回错误
//
// 使用示例:
//   - bill, err := client.ApplyFundFlowBill(ctx, &payment.BillRequest{BillDate: "2026-10-14", AccountType: "BASIC"})
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012791896
func (c *Client) ApplyFundFlowBill(ctx context.Context, req *BillRequest) (*BillResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if req.BillDate == "" {
		return nil, fmt.Errorf("账单日期不能为空")
	}

	var result BillResponse
	if err := c.doRequest(ctx, "GET", buildBillURL(BillFundFlowURL, req), nil, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

// DownloadBill 下载账单文件并校验完整性
//
// 参数:
//   - ctx: 上下文
//   - downloadURL: 申请账单接口返回的下载地址，5 分钟内有效
//   - hashValue: 申请账单接口返回的哈希值，传空字符串表示跳过校验
//
// 返回:
//   - []byte: 账单文件原始字节，tar_type 为 GZIP 时为压缩流
//   - error: 下载失败或哈希校验不一致时返回错误
//
// 说明:
//   - 官方下载账单响应不含签名头，故跳过验签
//
// 使用示例:
//   - data, err := client.DownloadBill(ctx, bill.DownloadURL, bill.HashValue)
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012791889
func (c *Client) DownloadBill(ctx context.Context, downloadURL, hashValue string) ([]byte, error) {
	if downloadURL == "" {
		return nil, fmt.Errorf("下载地址不能为空")
	}

	body, err := c.doRequestRaw(ctx, "GET", downloadURL, false)
	if err != nil {
		return nil, fmt.Errorf("下载账单失败: %w", err)
	}

	if hashValue == "" {
		return body, nil
	}

	actual := sha1.Sum(body)
	if !strings.EqualFold(hex.EncodeToString(actual[:]), hashValue) {
		return nil, fmt.Errorf("账单哈希校验失败，期望 %s，实际 %s", hashValue, hex.EncodeToString(actual[:]))
	}

	return body, nil
}
