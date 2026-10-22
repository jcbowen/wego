package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/wego/logger"
)

// VerifyResponseFunc 响应验证回调函数类型
type VerifyResponseFunc func(ctx context.Context, statusCode int, headers http.Header, body []byte) error

// ErrorParserFunc 非 2xx 响应的错误解析回调类型
//
// 说明:
//   - 返回 nil 表示响应体不是该服务的错误格式，由调用方回退到通用错误
//   - 微信支付 V3 等接口在 4xx/5xx 时会返回结构化错误 JSON，需解析后供业务判断
type ErrorParserFunc func(statusCode int, headers http.Header, body []byte) error

// sensitiveQueryPattern 匹配 URL query 中的敏感参数（token/secret/ticket 等）
var sensitiveQueryPattern = regexp.MustCompile(`(?i)(access_token|component_access_token|secret|ticket)=([^&]*)`)

// redactSensitiveQuery 脱敏 URL 中的敏感参数，避免凭证进入日志
func redactSensitiveQuery(rawURL string) string {
	if rawURL == "" {
		return rawURL
	}
	return sensitiveQueryPattern.ReplaceAllString(rawURL, "$1=[REDACTED]")
}

// sensitiveBodyPattern 匹配响应体 JSON 中敏感字段（token/secret/ticket 等）的字符串值
var sensitiveBodyPattern = regexp.MustCompile(`(?i)("(?:access_token|component_access_token|secret|ticket|api_ticket|jsapi_ticket|encrypt_key|session_key)"\s*:\s*")[^"]*(")`)

// redactSensitiveBody 脱敏响应体 JSON 中的敏感字段值，保留键名便于排障；
// 非 JSON 文本（如账单文件流、加密 XML）不含上述键名模式，原样返回
//
// 参数：body 原始响应体
// 返回：脱敏后的字符串
// 使用示例：r.logger.Info(..., "response_body": redactSensitiveBody(respBody))
func redactSensitiveBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	return sensitiveBodyPattern.ReplaceAllString(string(body), "${1}[REDACTED]${2}")
}

type ReqMakeOpt struct {
	Method         string             // HTTP方法，如GET、POST等
	URL            string             // 请求URL
	Query          interface{}        // 查询参数，支持map[string]string或结构体
	Body           interface{}        // 请求体
	Result         interface{}        // 响应体，需要传入指针
	RawResult      *[]byte            // 原始响应体接收器，非 nil 时跳过 JSON 解析（如下载账单文件）
	Headers        map[string]string  // 自定义请求头，可选
	VerifyResponse VerifyResponseFunc // 响应验证回调，可选
	ErrorParser    ErrorParserFunc    // 非 2xx 响应的错误解析回调，可选
}

type Request struct {
	httpClient HTTPClient
	logger     logger.LoggerInterface
}

func NewRequest(httpClient HTTPClient, logger logger.LoggerInterface) *Request {
	return &Request{
		httpClient: httpClient,
		logger:     logger,
	}
}

// Make 发送HTTP请求的通用方法
//
// 参数:
//   - ctx: 请求上下文，用于控制请求超时和取消
//   - options: 请求配置选项，包含方法、URL、查询参数、请求体和结果接收器
//
// 返回:
//   - error: 请求执行过程中的错误，包含详细的错误信息
//
// 功能:
//   - 支持GET、POST等常用HTTP方法
//   - 自动处理查询参数拼接，支持map[string]string和结构体
//   - 支持JSON格式的请求体和响应体
//   - 自动验证HTTP响应状态码
//   - 提供详细的日志记录
func (r *Request) Make(ctx context.Context, options *ReqMakeOpt) error {
	method := options.Method
	requestURL := options.URL
	query := options.Query
	body := options.Body
	result := options.Result

	// 处理查询参数
	if query != nil {
		queryMap := helper.Convert{Value: query}.ToMapString()
		if len(queryMap) > 0 {
			parsedURL, err := url.Parse(requestURL)
			if err != nil {
				return fmt.Errorf("解析URL失败: %w", err)
			}

			// 构建查询参数
			q := parsedURL.Query()
			for key, value := range queryMap {
				q.Set(key, value)
			}
			parsedURL.RawQuery = q.Encode()
			requestURL = parsedURL.String()
		}
	}

	var reqBody []byte
	if body != nil {
		switch v := body.(type) {
		case []byte:
			reqBody = v
		case io.Reader:
			var err error
			reqBody, err = io.ReadAll(v)
			if err != nil {
				return fmt.Errorf("读取请求体失败: %w", err)
			}
		default:
			var err error
			reqBody, err = json.Marshal(body)
			if err != nil {
				return fmt.Errorf("序列化请求体失败: %w", err)
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置默认 Content-Type：仅对携带请求体的写方法（POST/PUT/PATCH）生效；
	// GET/HEAD/DELETE 等语义上无请求体，不设置该头，避免多余头触发网关拦截
	switch req.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		if len(reqBody) > 0 {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// 设置自定义请求头（覆盖默认值）
	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}

	// 记录请求日志（URL 中的 token/secret 等凭证需脱敏后落日志；
	// headers/body 可能含 Authorization、密钥等凭证，仅记录规模信息，需完整审计时请自行记录）
	safeURL := redactSensitiveQuery(req.URL.String())
	r.logger.Info(fmt.Sprintf("发送HTTP请求 - Method: %s URL: %s", req.Method, safeURL), map[string]interface{}{
		"request_method":       req.Method,
		"request_url":          safeURL,
		"request_header_count": len(req.Header),
		"request_body_size":    len(reqBody),
	})

	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error(fmt.Sprintf("发送请求失败: %v", err), map[string]interface{}{
			"url": redactSensitiveQuery(requestURL),
		})
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer func(Body io.ReadCloser) {
		if closeErr := Body.Close(); closeErr != nil {
			r.logger.Error(fmt.Sprintf("关闭响应体失败: %v", closeErr))
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		r.logger.Error(fmt.Sprintf("读取响应体失败: %v", err), map[string]interface{}{
			"status_code": resp.StatusCode,
		})
		return fmt.Errorf("读取响应体失败: %w", err)
	}

	// 记录响应日志（响应体可能含 access_token 等凭证，脱敏后落日志）
	r.logger.Info(fmt.Sprintf("HTTP响应 - 状态码: %d, 内容长度: %d", resp.StatusCode, len(respBody)), map[string]interface{}{
		"status_code":     resp.StatusCode,
		"response_length": len(respBody),
		"response_body":   redactSensitiveBody(respBody),
	})

	// 检查HTTP状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		safeBody := redactSensitiveBody(respBody)
		r.logger.Error(fmt.Sprintf("HTTP请求失败 - 状态码: %d, 响应: %s", resp.StatusCode, safeBody), map[string]interface{}{
			"status_code": resp.StatusCode,
		})
		// 优先使用业务方提供的错误解析，便于区分错误码（如微信支付 ORDER_PAID）
		if options.ErrorParser != nil {
			if parsed := options.ErrorParser(resp.StatusCode, resp.Header, respBody); parsed != nil {
				return parsed
			}
		}
		return fmt.Errorf("HTTP请求失败 - 状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 执行响应验证
	if options.VerifyResponse != nil {
		if err := options.VerifyResponse(ctx, resp.StatusCode, resp.Header, respBody); err != nil {
			return fmt.Errorf("响应验证失败: %w", err)
		}
	}

	// 原始响应体优先，供非 JSON 响应（如账单文件流）使用
	if options.RawResult != nil {
		*options.RawResult = respBody
		return nil
	}

	// 处理空响应
	if len(respBody) == 0 {
		if result != nil {
			r.logger.Info("响应体为空，跳过JSON解析")
		}
		return nil
	}

	// 解析JSON响应
	if result != nil {
		if err = json.Unmarshal(respBody, result); err != nil {
			r.logger.Error(fmt.Sprintf("解析响应失败: %v, 响应内容: %s", err, redactSensitiveBody(respBody)))
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}

	return nil
}

// MakeRaw 发送原始HTTP请求，返回响应对象
func (r *Request) MakeRaw(req *http.Request) (*http.Response, error) {
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	return resp, nil
}

// CheckRawResponseStatus 校验 MakeRaw 返回的响应状态码（MakeRaw 保持返回原始响应的语义，校验由调用方按需调用）。
// 非 2xx 时返回错误：若响应体可解析出微信 errcode/errmsg 则附带业务错误信息，否则附加响应体前 512 字节原文，
// 避免把网关/代理层错误（如 500/404 + HTML 页面）当成功响应体处理导致报错信息失真。
//
// 参数:
//   - statusCode: HTTP 状态码
//   - body: 已读取的响应体
//
// 返回:
//   - error: 非 2xx 时返回错误，2xx 返回 nil
//
// 使用示例:
//
//	resp, _ := req.MakeRaw(httpReq)
//	body, _ := io.ReadAll(resp.Body)
//	if err := core.CheckRawResponseStatus(resp.StatusCode, body); err != nil { ... }
func CheckRawResponseStatus(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}
	var apiResp APIResponse
	if json.Unmarshal(body, &apiResp) == nil && apiResp.ErrCode != 0 {
		return fmt.Errorf("HTTP请求失败，状态码: %d, errcode: %d, errmsg: %s", statusCode, apiResp.ErrCode, apiResp.ErrMsg)
	}
	return fmt.Errorf("HTTP请求失败，状态码: %d, 响应体: %.512s", statusCode, body)
}
