package openplatform

import (
	"testing"
	"time"
)

// TestJSSDKNonceStr 测试随机字符串生成的线程安全性
func TestJSSDKNonceStr(t *testing.T) {
	// 并发测试随机字符串生成
	results := make(chan string, 100)
	
	for i := 0; i < 100; i++ {
		go func() {
			results <- generateNonceStr()
		}()
	}
	
	// 收集结果并检查唯一性
	uniqueResults := make(map[string]bool)
	for i := 0; i < 100; i++ {
		select {
		case nonce := <-results:
			if uniqueResults[nonce] {
				t.Errorf("发现重复的nonce字符串: %s", nonce)
			}
			uniqueResults[nonce] = true
			
			// 检查长度
			if len(nonce) != 16 {
				t.Errorf("nonce字符串长度不正确: got %d, want 16", len(nonce))
			}
		case <-time.After(time.Second):
			t.Error("随机字符串生成超时")
		}
	}
}

// TestJSSDKCacher 测试缓存功能
func TestJSSDKCacher(t *testing.T) {
	cacher := &JSSDKCacher{}
	
	// 测试缓存未命中
	config := &JSSDKConfig{
		AppID:     "test_appid",
		Timestamp: time.Now().Unix(),
		NonceStr:  "test_nonce",
		Signature: "test_signature",
		JSAPIList: []string{"api1", "api2"},
	}
	
	// 缓存配置
	cacher.CacheConfig("http://test.com", []string{"api1", "api2"}, config)
	
	// 验证缓存命中
	cached, err := cacher.GetCachedConfig("http://test.com", []string{"api1", "api2"})
	if err != nil {
		t.Errorf("缓存应该命中: %v", err)
	}
	if cached.AppID != config.AppID {
		t.Errorf("缓存配置不匹配: got %s, want %s", cached.AppID, config.AppID)
	}
	
	// 验证缓存统计
	stats := cacher.GetStats()
	if stats.Hits != 1 {
		t.Errorf("缓存命中统计不正确: got %d, want 1", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("缓存未命中统计不正确: got %d, want 0", stats.Misses)
	}
	if stats.TotalItems != 1 {
		t.Errorf("缓存项统计不正确: got %d, want 1", stats.TotalItems)
	}
	
	// 测试缓存过期
	oldConfig := &JSSDKConfig{
		AppID:     "old_appid",
		Timestamp: time.Now().Unix() - 7201, // 过期时间
		NonceStr:  "old_nonce",
		Signature: "old_signature",
		JSAPIList: []string{"api3"},
	}
	cacher.CacheConfig("http://old.com", []string{"api3"}, oldConfig)
	
	_, err = cacher.GetCachedConfig("http://old.com", []string{"api3"})
	if err == nil {
		t.Error("过期缓存应该返回错误")
	}
	
	// 验证缓存未命中统计更新
	stats = cacher.GetStats()
	if stats.Misses != 1 {
		t.Errorf("缓存未命中统计不正确: got %d, want 1", stats.Misses)
	}
}

// TestJSSDKCacheKey 测试缓存键生成
func TestJSSDKCacheKey(t *testing.T) {
	cacher := &JSSDKCacher{}
	
	// 相同参数应该生成相同的缓存键
	key1 := cacher.generateCacheKey("http://test.com", []string{"api1", "api2"})
	key2 := cacher.generateCacheKey("http://test.com", []string{"api1", "api2"})
	if key1 != key2 {
		t.Error("相同参数应该生成相同的缓存键")
	}
	
	// 不同参数应该生成不同的缓存键
	key3 := cacher.generateCacheKey("http://other.com", []string{"api1", "api2"})
	if key1 == key3 {
		t.Error("不同URL应该生成不同的缓存键")
	}
	
	// API列表顺序不影响缓存键（排序后应该相同）
	key4 := cacher.generateCacheKey("http://test.com", []string{"api2", "api1"})
	if key1 != key4 {
		t.Error("API列表顺序不应该影响缓存键")
	}
}

// TestJSSDKConfig 测试配置结构
func TestJSSDKConfig(t *testing.T) {
	// 验证配置结构
	config := &JSSDKConfig{
		AppID:     "wx1234567890",
		Timestamp: 1234567890,
		NonceStr:  "test_nonce_string",
		Signature: "test_signature_hash",
		JSAPIList: []string{"updateAppMessageShareData", "chooseImage"},
	}
	
	// 验证JSON标签
	if config.AppID != "wx1234567890" {
		t.Errorf("AppID不匹配: got %s", config.AppID)
	}
	if len(config.JSAPIList) != 2 {
		t.Errorf("JSAPIList长度不正确: got %d", len(config.JSAPIList))
	}
}