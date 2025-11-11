package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jcbowen/wego/logger"
)

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
func (request *Request) Make(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("序列化请求体失败: %v", err)
		}

		// 添加详细的参数调试日志
		request.logger.Debug(fmt.Sprintf("请求参数详情 - URL: %s, Method: %s", url, method), map[string]interface{}{
			"client": "OpenPlatform",
			"method": "MakeRequest",
			"source": "request.params",
			"params": body,
		})
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	request.logger.Debug(map[string]interface{}{
		"request_method": req.Method,
		"url":            req.URL,
		"headers":        req.Header,
		"body":           string(reqBody),
	}, map[string]interface{}{
		"client": "OpenPlatform",
		"method": "MakeRequest",
		"source": "request.body",
	})

	// 发送请求前记录完整URL和参数
	request.logger.Debug(fmt.Sprintf("发送HTTP请求 - URL: %s, Method: %s", req.URL.String(), req.Method), map[string]interface{}{
		"client":   "OpenPlatform",
		"method":   "MakeRequest",
		"source":   "request.send",
		"full_url": req.URL.String(),
		"headers":  req.Header,
	})

	resp, err := request.httpClient.Do(req)
	if err != nil {
		request.logger.Error(fmt.Sprintf("发送请求失败: %v", err), map[string]interface{}{
			"client": "OpenPlatform",
			"method": "MakeRequest",
			"source": "request.error",
			"url":    req.URL.String(),
		})
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			request.logger.Error(fmt.Sprintf("关闭响应体失败: %v", err))
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		request.logger.Error(fmt.Sprintf("读取响应体失败: %v", err), map[string]interface{}{
			"client": "OpenPlatform",
			"method": "MakeRequest",
			"source": "response.read_error",
		})
		return fmt.Errorf("读取响应体失败: %v", err)
	}

	// 记录响应状态和完整响应内容
	request.logger.Debug(fmt.Sprintf("HTTP响应 - 状态码: %d, 内容长度: %d", resp.StatusCode, len(respBody)), map[string]interface{}{
		"client":          "OpenPlatform",
		"method":          "MakeRequest",
		"source":          "response.status",
		"status_code":     resp.StatusCode,
		"response_length": len(respBody),
	})

	request.logger.Debug(string(respBody), map[string]interface{}{
		"client": "OpenPlatform",
		"method": "MakeRequest",
		"source": "response.body",
	})

	if err = json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	return nil
}

// MakeRaw 发送原始HTTP请求，返回响应对象
func (request *Request) MakeRaw(req *http.Request) (*http.Response, error) {
	resp, err := request.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	return resp, nil
}
