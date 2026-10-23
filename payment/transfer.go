// Package payment 微信支付 V3 - 商家转账（用户确认收款模式，单笔）
// 参考：
//   - 发起转账：https://pay.weixin.qq.com/doc/v3/merchant/4012716434
//   - 用户授权后转账：https://pay.weixin.qq.com/doc/v3/merchant/4014399371
//   - 撤销转账：https://pay.weixin.qq.com/doc/v3/merchant/4012716458
//   - 查询转账单（商户单号/微信单号）：https://pay.weixin.qq.com/doc/v3/merchant/4012716437 | 4012716457
//   - 电子回单（申请/查询）：https://pay.weixin.qq.com/doc/v3/merchant/4012716452 | 4012716455
package payment

import (
	"context"
	"fmt"
)

// TransferSceneReportInfo 商家转账场景报备信息项
type TransferSceneReportInfo struct {
	InfoType    string `json:"info_type"`    // 报备信息类型（场景 ID 对应的报备信息名称）
	InfoContent string `json:"info_content"` // 报备信息内容
}

// TransferBillCreateRequest 发起转账请求（POST /v3/fund-app/mch-transfer/transfer-bills）
//
// 注意：
//   - 用户确认收款模式：创建转账单后需引导用户确认收款（package_info 用于拉起收款确认页）
//   - user_name 为敏感字段，请先通过 EncryptSensitiveData 加密后再传入（金额 >= 2000 元必填）
//   - transfer_scene_report_infos 官方标注必填，需按转账场景准确填写（参考 https://pay.weixin.qq.com/doc/v3/merchant/4013774588）
type TransferBillCreateRequest struct {
	AppID                    string                    `json:"appid"`                                 // 商户 AppID（须与商户号绑定）
	OutBillNo                string                    `json:"out_bill_no"`                           // 商户单号（数字/大小写字母，商户系统内唯一）
	TransferSceneID          string                    `json:"transfer_scene_id"`                     // 转账场景 ID（如 1000 现金营销、1005 佣金报酬、1006 企业报销）
	OpenID                   string                    `json:"openid"`                                // 收款用户 OpenID（AppID 下）
	UserName                 string                    `json:"user_name,omitempty"`                   // 收款用户姓名（RSA 加密后的密文）
	TransferAmount           int64                     `json:"transfer_amount"`                       // 转账金额（单位：分）
	TransferRemark           string                    `json:"transfer_remark"`                       // 转账备注（UTF8，最长 32 字符，用户可见）
	NotifyURL                string                    `json:"notify_url,omitempty"`                  // 回调通知地址（https）
	UserRecvPerception       string                    `json:"user_recv_perception,omitempty"`        // 用户收款感知（如 "佣金报酬"、"现金奖励"）
	UserRecvStyle            *UserRecvStyle            `json:"user_recv_style,omitempty"`             // 用户收款样式（收款确认页/红包，不传默认收款确认页）
	TransferSceneReportInfos []TransferSceneReportInfo `json:"transfer_scene_report_infos,omitempty"` // 转账场景报备信息
}

// UserRecvStyle 用户收款样式（发起转账 user_recv_style）
//
// 说明:
//   - 官方为对象结构，仅含样式类型；不传时默认按收款确认页样式展示
//   - Type 取值：CONFIRM_PAGE 收款确认页样式 / RED_PACKET 红包样式
//     （红包单笔金额需 ≤200 元，仅部分转账场景可用，取值可能新增需兼容）
type UserRecvStyle struct {
	Type string `json:"type"` // 用户收款样式类型
}

// TransferBillCreateResponse 发起转账响应
type TransferBillCreateResponse struct {
	OutBillNo      string `json:"out_bill_no"`      // 商户单号
	TransferBillNo string `json:"transfer_bill_no"` // 微信转账单号
	State          string `json:"state"`            // 单据状态（见 TransferStateXxx 常量，创建后一般为 WAIT_USER_CONFIRM）
	PackageInfo    string `json:"package_info"`     // 跳转微信收款确认页的 package 信息
	CreateTime     string `json:"create_time"`      // 单据创建时间（rfc3339）
}

// TransferBillResponse 查询转账单响应
//
// user_name 为密文；成功/失败终态时 fail_reason 可能有值。
type TransferBillResponse struct {
	MchID          string `json:"mch_id"`                // 商户号
	OutBillNo      string `json:"out_bill_no"`           // 商户单号
	TransferBillNo string `json:"transfer_bill_no"`      // 微信转账单号
	AppID          string `json:"appid"`                 // 商户 AppID
	State          string `json:"state"`                 // 单据状态（见 TransferStateXxx 常量）
	TransferAmount int64  `json:"transfer_amount"`       // 转账金额（单位：分）
	TransferRemark string `json:"transfer_remark"`       // 转账备注
	OpenID         string `json:"openid,omitempty"`      // 收款用户 OpenID
	UserName       string `json:"user_name,omitempty"`   // 收款用户姓名（密文）
	FailReason     string `json:"fail_reason,omitempty"` // 失败原因（FAIL/已退资金时返回）
	CreateTime     string `json:"create_time"`           // 单据创建时间（rfc3339）
	UpdateTime     string `json:"update_time"`           // 最后一次状态变更时间（rfc3339）
}

// TransferBillCancelResponse 撤销转账响应
//
// 返回成功仅表示撤销请求已受理，系统异步处理，最终以查询单据状态为准。
type TransferBillCancelResponse struct {
	OutBillNo      string `json:"out_bill_no"`      // 商户单号
	TransferBillNo string `json:"transfer_bill_no"` // 微信转账单号
	State          string `json:"state"`            // 单据状态（CANCELING 撤销中 / CANCELLED 撤销完成）
	UpdateTime     string `json:"update_time"`      // 最后一次单据状态变更时间（rfc3339）
}

// TransferElecsignApplyRequest 申请电子回单请求
type TransferElecsignApplyRequest struct {
	OutBillNo      string `json:"out_bill_no,omitempty"`      // 商户单号（与微信转账单号二选一）
	TransferBillNo string `json:"transfer_bill_no,omitempty"` // 微信转账单号（与商户单号二选一）
}

// TransferElecsignResponse 电子回单申请/查询响应
//
// state=FINISHED 时返回 hash_type/hash_value/download_url；
// download_url 有效期 10 分钟，过期后重新查询获取，不要自行拼接。
type TransferElecsignResponse struct {
	State       string `json:"state"`                  // 申请单状态（见 TransferElecsignStateXxx 常量）
	CreateTime  string `json:"create_time"`            // 申请单创建时间（rfc3339）
	UpdateTime  string `json:"update_time"`            // 申请单更新时间（rfc3339）
	HashType    string `json:"hash_type,omitempty"`    // 回单文件摘要类型（SHA256/SM3），FINISHED 时返回
	HashValue   string `json:"hash_value,omitempty"`   // 回单文件摘要值（FINISHED 时返回，用于校验文件完整性）
	DownloadURL string `json:"download_url,omitempty"` // 回单文件下载地址（FINISHED 时返回，有效期 10 分钟）
}

// validateOutBillNo 校验商户单号格式（官方要求仅数字与大小写字母）
func validateOutBillNo(outBillNo string) error {
	if outBillNo == "" {
		return fmt.Errorf("商户单号不能为空")
	}
	for _, r := range outBillNo {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
			return fmt.Errorf("商户单号只能包含数字和大小写字母")
		}
	}
	return nil
}

// validateTransferSceneReportInfos 校验转账场景报备信息（官方标注必填，
// 客户端前置校验可避免等微信返回 400 才发现问题）
func validateTransferSceneReportInfos(infos []TransferSceneReportInfo) error {
	if len(infos) == 0 {
		return fmt.Errorf("transfer_scene_report_infos 官方标注必填，不能为空")
	}
	for i, info := range infos {
		if info.InfoType == "" || info.InfoContent == "" {
			return fmt.Errorf("transfer_scene_report_infos[%d] 的 info_type 与 info_content 均不能为空", i)
		}
	}
	return nil
}

// CreateTransferBill 发起商家转账（用户确认收款模式）
//
// 参数:
//   - ctx: 上下文
//   - req: 转账请求参数（UserName 如传入须为 EncryptSensitiveData 加密后的密文）
//
// 返回:
//   - *TransferBillCreateResponse: 含 package_info，用于引导用户确认收款
//   - error: 请求失败或微信返回错误时返回错误
//
// 使用示例:
//
//	userName, _ := client.EncryptSensitiveData(ctx, "张三")
//	resp, err := client.CreateTransferBill(ctx, &payment.TransferBillCreateRequest{
//		AppID: "wx123", OutBillNo: "B20261022001", TransferSceneID: "1005",
//		OpenID: "oXyz...", UserName: userName, TransferAmount: 100,
//		TransferRemark: "10月佣金", NotifyURL: "https://example.com/notify",
//	})
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012716434
func (c *Client) CreateTransferBill(ctx context.Context, req *TransferBillCreateRequest) (*TransferBillCreateResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if req.AppID == "" {
		return nil, fmt.Errorf("appid 不能为空")
	}
	if err := validateOutBillNo(req.OutBillNo); err != nil {
		return nil, err
	}
	if req.TransferSceneID == "" {
		return nil, fmt.Errorf("transfer_scene_id 不能为空")
	}
	if req.OpenID == "" {
		return nil, fmt.Errorf("openid 不能为空")
	}
	if req.TransferAmount <= 0 {
		return nil, fmt.Errorf("transfer_amount 必须大于 0")
	}
	if req.TransferRemark == "" {
		return nil, fmt.Errorf("transfer_remark 不能为空")
	}
	if err := validateTransferSceneReportInfos(req.TransferSceneReportInfos); err != nil {
		return nil, err
	}

	// user_name 为 RSA 加密字段，必须声明与加密密钥一致的 Wechatpay-Serial（官方要求）
	serial := ""
	if req.UserName != "" {
		s, err := c.EncryptSerialNo(ctx)
		if err != nil {
			return nil, err
		}
		serial = s
	}

	var result TransferBillCreateResponse
	if err := c.doRequestWithSerial(ctx, "POST", MchTransferCreateURL, req, &result, true, serial); err != nil {
		return nil, err
	}
	return &result, nil
}

// TransferAfterUserAuthRequest 用户授权后转账请求（POST /v3/fund-app/mch-transfer/transfer-bills/transfer）
//
// 适用场景：用户在转账授权页完成授权后，商户凭授权单据实际出款（请求体不含 openid/notify_url）。
//
// 注意：
//   - authorization_id 与 out_authorization_no 二选一必填（标识用户授权单据）
//   - user_name 为敏感字段，请先通过 EncryptSensitiveData 加密后再传入
//   - transfer_scene_report_infos 官方标注必填，需按转账场景准确填写
type TransferAfterUserAuthRequest struct {
	AppID                    string                    `json:"appid"`                                 // 商户 AppID（须与商户号绑定）
	OutBillNo                string                    `json:"out_bill_no"`                           // 商户单号（数字/大小写字母，商户系统内唯一）
	TransferSceneID          string                    `json:"transfer_scene_id"`                     // 转账场景 ID（如 1000 现金营销、1005 佣金报酬、1006 企业报销）
	UserName                 string                    `json:"user_name,omitempty"`                   // 收款用户姓名（RSA 加密后的密文）
	TransferAmount           int64                     `json:"transfer_amount"`                       // 转账金额（单位：分）
	TransferRemark           string                    `json:"transfer_remark"`                       // 转账备注（UTF8，最长 32 字符，用户可见）
	UserRecvPerception       string                    `json:"user_recv_perception,omitempty"`        // 用户收款感知（如 "佣金报酬"、"现金奖励"）
	TransferSceneReportInfos []TransferSceneReportInfo `json:"transfer_scene_report_infos,omitempty"` // 转账场景报备信息
	AuthorizationID          string                    `json:"authorization_id,omitempty"`            // 微信授权单据号（与 out_authorization_no 二选一）
	OutAuthorizationNo       string                    `json:"out_authorization_no,omitempty"`        // 商户授权单号（与 authorization_id 二选一）
}

// TransferAfterUserAuth 用户授权后转账（用户完成授权后实际出款）
//
// 参数:
//   - ctx: 上下文
//   - req: 转账请求参数（UserName 如传入须为加密后的密文；AuthorizationID 与 OutAuthorizationNo 二选一）
//
// 返回:
//   - *TransferBillResponse: 含最新单据状态
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4014399371
func (c *Client) TransferAfterUserAuth(ctx context.Context, req *TransferAfterUserAuthRequest) (*TransferBillResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}
	if req.AppID == "" {
		return nil, fmt.Errorf("appid 不能为空")
	}
	if err := validateOutBillNo(req.OutBillNo); err != nil {
		return nil, err
	}
	if req.TransferSceneID == "" {
		return nil, fmt.Errorf("transfer_scene_id 不能为空")
	}
	if req.TransferAmount <= 0 {
		return nil, fmt.Errorf("transfer_amount 必须大于 0")
	}
	if req.TransferRemark == "" {
		return nil, fmt.Errorf("transfer_remark 不能为空")
	}
	if err := validateTransferSceneReportInfos(req.TransferSceneReportInfos); err != nil {
		return nil, err
	}
	// 两者同空或同有均不合法，二选一必填
	if (req.AuthorizationID == "") == (req.OutAuthorizationNo == "") {
		return nil, fmt.Errorf("authorization_id 与 out_authorization_no 二选一必填")
	}

	// user_name 为 RSA 加密字段，必须声明与加密密钥一致的 Wechatpay-Serial（官方要求）
	serial := ""
	if req.UserName != "" {
		s, err := c.EncryptSerialNo(ctx)
		if err != nil {
			return nil, err
		}
		serial = s
	}

	var result TransferBillResponse
	if err := c.doRequestWithSerial(ctx, "POST", MchTransferTransferURL, req, &result, true, serial); err != nil {
		return nil, err
	}
	return &result, nil
}

// CancelTransferBill 撤销转账（仅用户确认收款前可撤销，撤销为异步处理）
//
// 参数:
//   - ctx: 上下文
//   - outBillNo: 商户单号
//
// 返回:
//   - *TransferBillCancelResponse: 返回成功仅表示撤销请求已受理，最终以查询单据状态为准
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012716458
func (c *Client) CancelTransferBill(ctx context.Context, outBillNo string) (*TransferBillCancelResponse, error) {
	if err := validateOutBillNo(outBillNo); err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf(MchTransferCancelURL, outBillNo)

	var result TransferBillCancelResponse
	// 官方撤销转账接口无请求体参数，不发送 body（此前误传空 JSON 对象）
	err := c.doRequest(ctx, "POST", apiURL, nil, &result, true)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryTransferBillByOutBillNo 按商户单号查询转账单
//
// 参数:
//   - ctx: 上下文
//   - outBillNo: 商户单号
//
// 返回:
//   - *TransferBillResponse: 单据详情（含状态、金额、失败原因等）
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012716437
func (c *Client) QueryTransferBillByOutBillNo(ctx context.Context, outBillNo string) (*TransferBillResponse, error) {
	if err := validateOutBillNo(outBillNo); err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf(MchTransferQueryByOutBillNoURL, outBillNo)

	var result TransferBillResponse
	err := c.doRequest(ctx, "GET", apiURL, nil, &result, true)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryTransferBillByBillNo 按微信转账单号查询转账单
//
// 参数:
//   - ctx: 上下文
//   - transferBillNo: 微信转账单号
//
// 返回:
//   - *TransferBillResponse: 单据详情
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012716457
func (c *Client) QueryTransferBillByBillNo(ctx context.Context, transferBillNo string) (*TransferBillResponse, error) {
	if transferBillNo == "" {
		return nil, fmt.Errorf("微信转账单号不能为空")
	}

	apiURL := fmt.Sprintf(MchTransferQueryByBillNoURL, transferBillNo)

	var result TransferBillResponse
	err := c.doRequest(ctx, "GET", apiURL, nil, &result, true)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ApplyForElecsignByOutBillNo 按商户单号申请转账电子回单
//
// 参数:
//   - ctx: 上下文
//   - outBillNo: 商户单号（须为转账成功的单据）
//
// 返回:
//   - *TransferElecsignResponse: 申请受理结果
//   - error: 请求失败或微信返回错误时返回错误
//
// 说明:
//   - 需传入收款用户姓名的转账单才支持电子回单（官方申请条件：单据状态 SUCCESS、传入收款用户姓名、六个月内）
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012716452
func (c *Client) ApplyForElecsignByOutBillNo(ctx context.Context, outBillNo string) (*TransferElecsignResponse, error) {
	if err := validateOutBillNo(outBillNo); err != nil {
		return nil, err
	}

	var result TransferElecsignResponse
	err := c.doRequest(ctx, "POST", MchTransferElecsignApplyByOutBillNoURL,
		TransferElecsignApplyRequest{OutBillNo: outBillNo}, &result, true)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ApplyForElecsignByBillNo 按微信转账单号申请转账电子回单
//
// 参数:
//   - ctx: 上下文
//   - transferBillNo: 微信转账单号
//
// 返回:
//   - *TransferElecsignResponse: 申请受理结果
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012716456
func (c *Client) ApplyForElecsignByBillNo(ctx context.Context, transferBillNo string) (*TransferElecsignResponse, error) {
	if transferBillNo == "" {
		return nil, fmt.Errorf("微信转账单号不能为空")
	}

	var result TransferElecsignResponse
	err := c.doRequest(ctx, "POST", MchTransferElecsignApplyByBillNoURL,
		TransferElecsignApplyRequest{TransferBillNo: transferBillNo}, &result, true)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryElecsignByOutBillNo 按商户单号查询电子回单申请进度
//
// 参数:
//   - ctx: 上下文
//   - outBillNo: 商户单号
//
// 返回:
//   - *TransferElecsignResponse: FINISHED 时返回 hash_type/hash_value/download_url
//   - error: 请求失败或微信返回错误时返回错误
//
// 使用示例:
//
//	resp, err := client.QueryElecsignByOutBillNo(ctx, "B20261022001")
//	if err == nil && resp.State == payment.TransferElecsignStateFinished {
//		// 下载 resp.DownloadURL 对应文件，并用 resp.HashType/resp.HashValue 校验完整性
//	}
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012716455
func (c *Client) QueryElecsignByOutBillNo(ctx context.Context, outBillNo string) (*TransferElecsignResponse, error) {
	if err := validateOutBillNo(outBillNo); err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf(MchTransferElecsignQueryByOutBillNoURL, outBillNo)

	var result TransferElecsignResponse
	err := c.doRequest(ctx, "GET", apiURL, nil, &result, true)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryElecsignByBillNo 按微信转账单号查询电子回单申请进度
//
// 参数:
//   - ctx: 上下文
//   - transferBillNo: 微信转账单号
//
// 返回:
//   - *TransferElecsignResponse: FINISHED 时返回 hash_type/hash_value/download_url
//   - error: 请求失败或微信返回错误时返回错误
//
// 文档:
//   - https://pay.weixin.qq.com/doc/v3/merchant/4012716455
func (c *Client) QueryElecsignByBillNo(ctx context.Context, transferBillNo string) (*TransferElecsignResponse, error) {
	if transferBillNo == "" {
		return nil, fmt.Errorf("微信转账单号不能为空")
	}

	apiURL := fmt.Sprintf(MchTransferElecsignQueryByBillNoURL, transferBillNo)

	var result TransferElecsignResponse
	err := c.doRequest(ctx, "GET", apiURL, nil, &result, true)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
