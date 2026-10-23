package payment

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// 微信支付模块预定义错误，业务可使用 errors.Is 判断
var (
	// ErrNoEncryptKey 未配置微信支付公钥且无法获取平台证书，无法加密敏感信息
	ErrNoEncryptKey = errors.New("未配置微信支付公钥且无法获取平台证书，无法加密敏感信息")

	// ErrCertDownloadFailed 平台证书下载或解密失败
	ErrCertDownloadFailed = errors.New("平台证书下载失败")

	// ErrSignatureHeaderMissing 响应缺少微信支付签名头
	ErrSignatureHeaderMissing = errors.New("缺少微信支付响应签名头")
)

// APIError 微信支付 V3 接口返回的业务错误
//
// 字段说明:
//   - StatusCode: HTTP 状态码，如 400、401、403、404、429、500
//   - Code: 微信支付错误码，如 ORDER_PAID、SIGN_ERROR、NO_AUTH
//   - Message: 错误描述
//   - Detail: 详细的错误明细，可能为空
//
// 说明:
//   - 微信支付 V3 的错误响应体形如 {"code":"ORDER_PAID","message":"订单已支付","detail":{...}}
//   - 通过 IsAPIError / AsAPIError 判断与提取，便于业务层做差异化处理
//
// 使用示例:
//   - _, err := client.CloseOrder(ctx, "ORDER_001")
//   - if apiErr, ok := payment.AsAPIError(err); ok && apiErr.Code == "ORDERPAIDCLOSE" { ... }
type APIError struct {
	StatusCode int             `json:"status_code"`
	Code       string          `json:"code"`
	Message    string          `json:"message"`
	Detail     *APIErrorDetail `json:"detail,omitempty"`
	RawBody    string          `json:"raw_body,omitempty"`
	Header     http.Header     `json:"-"`
}

// APIErrorDetail 微信支付错误明细
type APIErrorDetail struct {
	Field    string `json:"field,omitempty"`    // 出错字段名
	Value    string `json:"value,omitempty"`    // 出错字段值
	Issue    string `json:"issue,omitempty"`    // 具体问题
	Location string `json:"location,omitempty"` // 出错位置
}

// Error 实现 error 接口
func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	msg := fmt.Sprintf("微信支付接口错误 - 状态码: %d, 错误码: %s, 描述: %s", e.StatusCode, e.Code, e.Message)
	if e.Detail != nil && e.Detail.Issue != "" {
		msg += ", 明细: " + e.Detail.Issue
	}
	return msg
}

// Is 支持 errors.Is 按错误码比较
//
// 使用示例:
//   - if errors.Is(err, &payment.APIError{Code: "ORDER_PAID"}) { ... }
func (e *APIError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	t, ok := target.(*APIError)
	if !ok {
		return false
	}
	if t.Code != "" && t.Code != e.Code {
		return false
	}
	if t.StatusCode != 0 && t.StatusCode != e.StatusCode {
		return false
	}
	return true
}

// IsAPIError 判断错误是否为微信支付业务错误
//
// 参数:
//   - err: 待判断的错误
//
// 返回:
//   - bool: 是微信支付 APIError 时返回 true
func IsAPIError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr)
}

// AsAPIError 从错误链中提取微信支付业务错误
//
// 参数:
//   - err: 待提取的错误
//
// 返回:
//   - *APIError: 提取到的业务错误，非微信支付错误时返回 nil
//   - bool: 是否提取成功
//
// 使用示例:
//   - if apiErr, ok := payment.AsAPIError(err); ok { log.Println(apiErr.Code) }
func AsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// parseWechatAPIError 解析微信支付错误响应体
//
// 参数:
//   - statusCode: HTTP 状态码
//   - headers: HTTP 响应头，用于记录 Request-ID 便于排查
//   - body: 响应体
//
// 返回:
//   - error: 解析成功返回 *APIError，响应体不是微信错误格式时返回 nil
//
// 说明:
//   - 微信支付在 4xx/5xx 时同样返回带 code/message 的 JSON，业务需据此区分处理
//   - 非微信格式（如网关返回 HTML）时返回 nil，由调用方回退到通用错误
func parseWechatAPIError(statusCode int, headers http.Header, body []byte) error {
	if len(body) == 0 {
		return nil
	}

	var payload struct {
		Code    string          `json:"code"`
		Message string          `json:"message"`
		Detail  *APIErrorDetail `json:"detail"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	if payload.Code == "" && payload.Message == "" {
		return nil
	}

	apiErr := &APIError{
		StatusCode: statusCode,
		Code:       payload.Code,
		Message:    payload.Message,
		Detail:     payload.Detail,
		RawBody:    string(body),
	}
	if headers != nil {
		apiErr.Header = headers.Clone()
	}
	return apiErr
}
