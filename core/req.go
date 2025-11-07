package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jcbowen/jcbaseGo/component/debugger"
)

type Request struct {
	httpClient HTTPClient
	logger     debugger.LoggerInterface
}

func NewRequest(httpClient HTTPClient, logger debugger.LoggerInterface) *Request {
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

	resp, err := request.httpClient.Do(req)
	if err != nil {
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
		return fmt.Errorf("读取响应体失败: %v", err)
	}

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
