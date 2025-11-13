package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/wego/logger"
)

type ReqMakeOpt struct {
	Method string      // HTTP方法，如GET、POST等
	URL    string      // 请求URL
	Query  interface{} // 查询参数，支持map[string]string或结构体
	Body   interface{} // 请求体
	Result interface{} // 响应体，需要传入指针
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
				return fmt.Errorf("解析URL失败: %v", err)
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
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("序列化请求体失败: %v", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	if len(reqBody) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	// 记录请求日志
    r.logger.Info(fmt.Sprintf("发送HTTP请求 - Method: %s URL: %s", req.Method, req.URL.String()), map[string]interface{}{
		"request_method":  req.Method,
		"request_url":     req.URL.String(),
		"request_headers": req.Header,
		"request_body":    string(reqBody),
	})

	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.logger.Error(fmt.Sprintf("发送请求失败: %v", err), map[string]interface{}{
			"url": requestURL,
		})
		return fmt.Errorf("发送请求失败: %v", err)
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
		return fmt.Errorf("读取响应体失败: %v", err)
	}

	// 记录响应日志
    r.logger.Info(fmt.Sprintf("HTTP响应 - 状态码: %d, 内容长度: %d", resp.StatusCode, len(respBody)), map[string]interface{}{
		"status_code":     resp.StatusCode,
		"response_length": len(respBody),
		"response_body":   string(respBody),
	})

	// 检查HTTP状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		r.logger.Error(fmt.Sprintf("HTTP请求失败 - 状态码: %d, 响应: %s", resp.StatusCode, string(respBody)), map[string]interface{}{
			"status_code": resp.StatusCode,
		})
		return fmt.Errorf("HTTP请求失败 - 状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
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
			r.logger.Error(fmt.Sprintf("解析响应失败: %v, 响应内容: %s", err, string(respBody)))
			return fmt.Errorf("解析响应失败: %v", err)
		}
	}

	return nil
}

// MakeRaw 发送原始HTTP请求，返回响应对象
func (r *Request) MakeRaw(req *http.Request) (*http.Response, error) {
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	return resp, nil
}
