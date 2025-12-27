package openplatform

import (
	"context"
	"testing"
)

// BenchmarkGenerateNonceStr 基准测试随机字符串生成性能
func BenchmarkGenerateNonceStr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateNonceStr()
	}
}

// BenchmarkCacheKeyGeneration 基准测试缓存键生成性能
func BenchmarkCacheKeyGeneration(b *testing.B) {
	cacher := &JSSDKCacher{}
	url := "http://example.com/path/to/page"
	jsAPIList := []string{"updateAppMessageShareData", "updateTimelineShareData", "chooseImage", "previewImage", "uploadImage"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cacher.generateCacheKey(url, jsAPIList)
	}
}

// BenchmarkCacheOperations 基准测试缓存操作性能
func BenchmarkCacheOperations(b *testing.B) {
	cacher := &JSSDKCacher{}
	config := &JSSDKConfig{
		AppID:     "wx1234567890",
		Timestamp: 1234567890,
		NonceStr:  "test_nonce_string",
		Signature: "test_signature_hash",
		JSAPIList: []string{"api1", "api2", "api3"},
	}
	
	b.Run("Cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cacher.CacheConfig("http://test.com", []string{"api1", "api2"}, config)
		}
	})
	
	b.Run("GetCached", func(b *testing.B) {
		// 先缓存一些数据
		for i := 0; i < 100; i++ {
			cacher.CacheConfig("http://test.com", []string{"api1", "api2"}, config)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cacher.GetCachedConfig("http://test.com", []string{"api1", "api2"})
		}
	})
}

// BenchmarkContextPropagation 基准测试Context传递性能
func BenchmarkContextPropagation(b *testing.B) {
	// 模拟一个JSSDKManager（注意：这需要模拟依赖）
	// 这里仅测试Context本身的传递开销
	ctx := context.Background()
	
	b.Run("WithContext", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// 模拟Context传递
			select {
			case <-ctx.Done():
				b.Error("Context不应该被取消")
			default:
				// 正常处理
			}
		}
	})
}