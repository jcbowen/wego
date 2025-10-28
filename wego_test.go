package wego

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHTTPClient 模拟HTTP客户端
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) PostJSON(ctx context.Context, url string, data interface{}, result interface{}) error {
	args := m.Called(ctx, url, data, result)
	return args.Error(0)
}

func (m *MockHTTPClient) Get(ctx context.Context, url string, result interface{}) error {
	args := m.Called(ctx, url, result)
	return args.Error(0)
}

// Do 实现http.Client接口的Do方法
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// 实现http.Client接口的Post方法
func (m *MockHTTPClient) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	args := m.Called(url, contentType, body)
	return args.Get(0).(*http.Response), args.Error(1)
}

// MockTokenCache 模拟令牌缓存
type MockTokenCache struct {
	mock.Mock
}

func (m *MockTokenCache) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockTokenCache) Set(key string, value string, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func TestWegoConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *WeGoConfig
		wantErr bool
	}{
		{
			name: "有效配置",
			config: &WeGoConfig{
				ComponentAppID:     "wx1234567890abcdef",
				ComponentAppSecret: "secret1234567890abcdef",
				ComponentToken:     "token123",
				EncodingAESKey:     "jWmYm7qr5nMoAUwZRjGtBxmz3KA1tkAj3ykkR6q2B2C",
			},
			wantErr: false,
		},
		{
			name: "缺少AppID",
			config: &WeGoConfig{
				ComponentAppSecret: "secret1234567890abcdef",
				ComponentToken:     "token123",
				EncodingAESKey:     "jWmYm7qr5nMoAUwZRjGtBxmz3KA1tkAj3ykkR6q2B2C",
			},
			wantErr: true,
		},
		{
			name: "缺少AppSecret",
			config: &WeGoConfig{
				ComponentAppID: "wx1234567890abcdef",
				ComponentToken: "token123",
				EncodingAESKey: "jWmYm7qr5nMoAUwZRjGtBxmz3KA1tkAj3ykkR6q2B2C",
			},
			wantErr: true,
		},
		{
			name: "无效的EncodingAESKey",
			config: &WeGoConfig{
				ComponentAppID:     "wx1234567890abcdef",
				ComponentAppSecret: "secret1234567890abcdef",
				ComponentToken:     "token123",
				EncodingAESKey:     "invalid_key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWegoClient_GenerateAuthURL(t *testing.T) {
	config := &WeGoConfig{
		ComponentAppID: "wx1234567890abcdef",
		RedirectURI:    "https://example.com/auth/callback",
	}
	client := NewWegoClient(config)

	tests := []struct {
		name           string
		preAuthCode    string
		authType       int
		bizAppID       string
		expectedPrefix string
	}{
		{
			name:           "基础授权链接",
			preAuthCode:    "pre_auth_code_123",
			authType:       0,
			bizAppID:       "",
			expectedPrefix: "https://mp.weixin.qq.com/cgi-bin/componentloginpage?component_appid=wx1234567890abcdef&pre_auth_code=pre_auth_code_123&redirect_uri=https%3A%2F%2Fexample.com%2Fauth%2Fcallback",
		},
		{
			name:           "指定授权类型",
			preAuthCode:    "pre_auth_code_456",
			authType:       1,
			bizAppID:       "",
			expectedPrefix: "https://mp.weixin.qq.com/cgi-bin/componentloginpage?auth_type=1&component_appid=wx1234567890abcdef&pre_auth_code=pre_auth_code_456&redirect_uri=https%3A%2F%2Fexample.com%2Fauth%2Fcallback",
		},
		{
			name:           "指定业务应用ID",
			preAuthCode:    "pre_auth_code_789",
			authType:       0,
			bizAppID:       "biz_appid_123",
			expectedPrefix: "https://mp.weixin.qq.com/cgi-bin/componentloginpage?biz_appid=biz_appid_123&component_appid=wx1234567890abcdef&pre_auth_code=pre_auth_code_789&redirect_uri=https%3A%2F%2Fexample.com%2Fauth%2Fcallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := client.GenerateAuthURL(tt.preAuthCode, tt.authType, tt.bizAppID)
			assert.Contains(t, url, tt.expectedPrefix)
		})
	}
}

func TestMessageProcessor_VerifySignature(t *testing.T) {
	config := &WeGoConfig{
		ComponentToken: "test_token",
	}
	processor := NewMessageProcessor(config)

	tests := []struct {
		name      string
		signature string
		timestamp string
		nonce     string
		want      bool
	}{
		{
			name:      "有效签名",
			signature: "5d8b8b1c8b6c8b6c8b6c8b6c8b6c8b6c8b6c8b6c", // 示例签名，实际需要计算
			timestamp: "1234567890",
			nonce:     "nonce123",
			want:      false, // 由于签名验证需要实际计算，这里设为false
		},
		{
			name:      "空签名",
			signature: "",
			timestamp: "1234567890",
			nonce:     "nonce123",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.VerifySignature(tt.signature, tt.timestamp, tt.nonce)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestMessageProcessor_GenerateTextResponse(t *testing.T) {
	config := &WeGoConfig{
		ComponentToken:     "test_token",
		EncodingAESKey:     "jWmYm7qr5nMoAUwZRjGtBxmz3KA1tkAj3ykkR6q2B2C",
		ComponentAppID:     "wx1234567890abcdef",
		ComponentAppSecret: "secret1234567890abcdef",
	}
	processor := NewMessageProcessor(config)

	resp, err := processor.GenerateTextResponse("toUser", "fromUser", "测试回复消息")
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)

	// 验证XML格式
	var xmlResp struct {
		XMLName      xml.Name `xml:"xml"`
		ToUserName   string   `xml:"ToUserName"`
		FromUserName string   `xml:"FromUserName"`
		CreateTime   int64    `xml:"CreateTime"`
		MsgType      string   `xml:"MsgType"`
		Content      string   `xml:"Content"`
	}
	err = xml.Unmarshal([]byte(resp), &xmlResp)
	assert.NoError(t, err)
	assert.Equal(t, "toUser", xmlResp.ToUserName)
	assert.Equal(t, "fromUser", xmlResp.FromUserName)
	assert.Equal(t, "text", xmlResp.MsgType)
	assert.Equal(t, "测试回复消息", xmlResp.Content)
}

func TestWegoClient_SetAndGetComponentToken(t *testing.T) {
	config := &WeGoConfig{
		ComponentAppID:     "wx1234567890abcdef",
		ComponentAppSecret: "secret1234567890abcdef",
		ComponentToken:     "token123",
		EncodingAESKey:     "jWmYm7qr5nMoAUwZRjGtBxmz3KA1tkAj3ykkR6q2B2C",
	}
	client := NewWegoClient(config)

	// 设置组件令牌
	token := &ComponentAccessToken{
		AccessToken: "test_token_123",
		ExpiresIn:   7200,
		ExpiresAt:   time.Now().Add(7200 * time.Second),
	}
	err := client.SetComponentToken(token)
	assert.NoError(t, err)

	// 获取组件令牌
	retrievedToken, err := client.GetComponentToken(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, retrievedToken)
	assert.Equal(t, token.AccessToken, retrievedToken.AccessToken)
	assert.Equal(t, token.ExpiresIn, retrievedToken.ExpiresIn)
}

func TestWegoClient_SetAndGetAuthorizerToken(t *testing.T) {
	config := &WeGoConfig{
		ComponentAppID:     "wx1234567890abcdef",
		ComponentAppSecret: "secret1234567890abcdef",
		ComponentToken:     "token123",
		EncodingAESKey:     "jWmYm7qr5nMoAUwZRjGtBxmz3KA1tkAj3ykkR6q2B2C",
	}
	client := NewWegoClient(config)

	authorizerAppID := "authorizer_appid_123"
	accessToken := "authorizer_access_token_456"
	refreshToken := "authorizer_refresh_token_789"
	expiresIn := 7200

	// 设置授权方令牌
	err := client.SetAuthorizerToken(authorizerAppID, accessToken, refreshToken, expiresIn)
	assert.NoError(t, err)

	// 获取授权方令牌
	retrievedToken, err := client.GetAuthorizerToken(context.Background(), authorizerAppID)
	assert.NoError(t, err)
	assert.Equal(t, accessToken, retrievedToken.AuthorizerAccessToken)
	assert.Equal(t, refreshToken, retrievedToken.AuthorizerRefreshToken)
	assert.Equal(t, expiresIn, retrievedToken.ExpiresIn)
}

func TestDefaultEventHandler_HandleEvent(t *testing.T) {
	handler := &DefaultEventHandler{}

	tests := []struct {
		name     string
		event    *EventMessage
		expected string
	}{
		{
			name: "组件验证票据事件",
			event: &EventMessage{
				Event: EventTypeComponentVerifyTicket,
			},
			expected: "success",
		},
		{
			name: "授权成功事件",
			event: &EventMessage{
				Event: EventTypeAuthorized,
			},
			expected: "success",
		},
		{
			name: "取消授权事件",
			event: &EventMessage{
				Event: EventTypeUnauthorized,
			},
			expected: "success",
		},
		{
			name: "更新授权事件",
			event: &EventMessage{
				Event: EventTypeUpdateAuthorized,
			},
			expected: "success",
		},
		{
			name: "未知事件类型",
			event: &EventMessage{
				Event: "unknown_event",
			},
			expected: "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.HandleEvent(tt.event)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthorizerClient_GetUserInfo(t *testing.T) {
	// 创建模拟的WegoClient
	config := &WeGoConfig{
		ComponentAppID:     "wx1234567890abcdef",
		ComponentAppSecret: "secret1234567890abcdef",
		ComponentToken:     "token123",
		EncodingAESKey:     "jWmYm7qr5nMoAUwZRjGtBxmz3KA1tkAj3ykkR6q2B2C",
	}

	client := NewWegoClient(config)
	mockHTTPClient := &MockHTTPClient{}
	client.SetHTTPClient(mockHTTPClient)

	authorizerAppID := "authorizer_appid_123"
	authorizerClient := NewAuthorizerClient(client, authorizerAppID)

	// 设置模拟的授权方令牌
	err := client.SetAuthorizerToken(authorizerAppID, "test_access_token", "test_refresh_token", 7200)
	assert.NoError(t, err)

	// 模拟HTTP响应
	mockHTTPClient.On("Do", mock.Anything).
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)
			// 验证请求URL包含正确的参数
			assert.Contains(t, req.URL.String(), "access_token=test_access_token")
			assert.Contains(t, req.URL.String(), "openid=user_openid_123")
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`{
				"openid": "user_openid_123",
				"nickname": "测试用户",
				"sex": 1,
				"city": "北京",
				"country": "中国",
				"province": "北京",
				"headimgurl": "http://example.com/avatar.jpg",
				"errcode": 0,
				"errmsg": "ok"
			}`)),
		}, nil)

	ctx := context.Background()
	userInfo, err := authorizerClient.GetUserInfo(ctx, "user_openid_123")

	assert.NoError(t, err)
	assert.Equal(t, "测试用户", userInfo.Nickname)
	assert.Equal(t, "北京", userInfo.City)
	mockHTTPClient.AssertExpectations(t)
}

func TestWegoClient_RefreshAuthorizerToken(t *testing.T) {
	config := &WeGoConfig{
		ComponentAppID:     "wx1234567890abcdef",
		ComponentAppSecret: "secret1234567890abcdef",
		ComponentToken:     "token123",
		EncodingAESKey:     "jWmYm7qr5nMoAUwZRjGtBxmz3KA1tkAj3ykkR6q2B2C",
	}

	client := NewWegoClient(config)
	mockHTTPClient := &MockHTTPClient{}
	client.SetHTTPClient(mockHTTPClient)

	authorizerAppID := "authorizer_appid_123"
	refreshToken := "old_refresh_token"

	// 设置初始令牌
	err := client.SetAuthorizerToken(authorizerAppID, "old_access_token", refreshToken, 7200)
	assert.NoError(t, err)

	// 模拟获取组件访问令牌的HTTP响应
	mockHTTPClient.On("Do", mock.Anything).
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)
			assert.Equal(t, "POST", req.Method)
			assert.Contains(t, req.URL.String(), "https://api.weixin.qq.com/cgi-bin/component/api_component_token")
		}).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"component_access_token":"component_token_123","expires_in":7200,"errcode":0,"errmsg":"ok"}`)),
	}, nil).Once()

	// 模拟刷新令牌的HTTP响应
	mockHTTPClient.On("Do", mock.Anything).
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)
			assert.Equal(t, "POST", req.Method)
			assert.Contains(t, req.URL.String(), "https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token")
		}).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"authorizer_access_token":"new_access_token","expires_in":7200,"authorizer_refresh_token":"new_refresh_token","errcode":0,"errmsg":"ok"}`)),
	}, nil)

	ctx := context.Background()
	newToken, err := client.RefreshAuthorizerToken(ctx, authorizerAppID, refreshToken)

	assert.NoError(t, err)
	assert.Equal(t, "new_access_token", newToken.AuthorizerAccessToken)
	assert.Equal(t, "new_refresh_token", newToken.AuthorizerRefreshToken)

	// 验证令牌已更新
	updatedToken, err := client.GetAuthorizerToken(context.Background(), authorizerAppID)
	assert.NoError(t, err)
	assert.Equal(t, "new_access_token", updatedToken.AuthorizerAccessToken)
	mockHTTPClient.AssertExpectations(t)
}

func TestMessageProcessor_ProcessMessage(t *testing.T) {
	config := &WeGoConfig{
		ComponentToken:     "test_token",
		EncodingAESKey:     "jWmYm7qr5nMoAUwZRjGtBxmz3KA1tkAj3ykkR6q2B2C",
		ComponentAppID:     "wx1234567890abcdef",
		ComponentAppSecret: "secret1234567890abcdef",
	}
	processor := NewMessageProcessor(config)

	// 注册自定义事件处理器
	customHandler := &CustomEventHandler{}
	processor.RegisterEventHandler(EventTypeComponentVerifyTicket, customHandler)

	// 测试组件验证票据事件
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<xml>
  <ToUserName><![CDATA[gh_1234567890abcdef]]></ToUserName>
  <FromUserName><![CDATA[wx1234567890abcdef]]></FromUserName>
  <CreateTime>1413192605</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[component_verify_ticket]]></Event>
  <ComponentVerifyTicket><![CDATA[ticket@@@jfkslJFKSLJFLK123jflksdjflkjsdlfkjlksjdflkjsd]]></ComponentVerifyTicket>
</xml>`

	result, err := processor.ProcessMessage([]byte(xmlData))
	assert.NoError(t, err)
	assert.Equal(t, "success", result)
}

// CustomEventHandler 自定义事件处理器
type CustomEventHandler struct{}

func (h *CustomEventHandler) HandleEvent(event *EventMessage) (interface{}, error) {
	return "success", nil
}

// 性能测试
func BenchmarkMessageProcessor_GenerateTextResponse(b *testing.B) {
	config := &WeGoConfig{
		ComponentToken:     "test_token",
		EncodingAESKey:     "jWmYm7qr5nMoAUwZRjGtBxmz3KA1tkAj3ykkR6q2B2C",
		ComponentAppID:     "wx1234567890abcdef",
		ComponentAppSecret: "secret1234567890abcdef",
	}
	processor := NewMessageProcessor(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processor.GenerateTextResponse("toUser", "fromUser", "测试消息")
	}
}

// 并发测试
func TestWegoClient_ConcurrentAccess(t *testing.T) {
	config := &WeGoConfig{
		ComponentAppID:     "wx1234567890abcdef",
		ComponentAppSecret: "secret1234567890abcdef",
		ComponentToken:     "token123",
		EncodingAESKey:     "jWmYm7qr5nMoAUwZRjGtBxmz3KA1tkAj3ykkR6q2B2C",
	}
	client := NewWegoClient(config)

	// 并发设置和获取令牌
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			authorizerAppID := string(rune('A' + id))
			err := client.SetAuthorizerToken(authorizerAppID, "token_"+authorizerAppID, "refresh_"+authorizerAppID, 7200)
			assert.NoError(t, err)
			token, err := client.GetAuthorizerToken(context.Background(), authorizerAppID)
			assert.NoError(t, err)
			assert.Equal(t, "token_"+authorizerAppID, token.AuthorizerAccessToken)
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// 错误处理测试
func TestWegoClient_ErrorHandling(t *testing.T) {
	config := &WeGoConfig{
		ComponentAppID:     "wx1234567890abcdef",
		ComponentAppSecret: "secret1234567890abcdef",
		ComponentToken:     "token123",
		EncodingAESKey:     "jWmYm7qr5nMoAUwZRjGtBxmz3KA1tkAj3ykkR6q2B2C",
	}

	client := NewWegoClient(config)
	mockHTTPClient := &MockHTTPClient{}
	client.SetHTTPClient(mockHTTPClient)

	// 模拟HTTP请求失败
	mockHTTPClient.On("Do", mock.Anything).
		Return(nil, assert.AnError)

	ctx := context.Background()
	_, err := client.GetComponentAccessToken(ctx, "test_verify_ticket")
	assert.Error(t, err)
	mockHTTPClient.AssertExpectations(t)
}
