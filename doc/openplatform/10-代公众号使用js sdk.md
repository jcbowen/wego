# 11-代公众号使用js sdk

## 概述

JS-SDK是微信公众平台面向网页开发者提供的基于微信内网页开发工具包。通过使用JS-SDK，网页开发者可借助微信高效地使用拍照、选图、语音、位置等手机系统的能力，同时可以直接使用微信分享、扫一扫、卡券、支付等微信特有的能力，为微信用户提供更优质的网页体验。

## JS-SDK功能列表

### 基础接口
- `wx.config` - 初始化配置
- `wx.ready` - 配置成功回调
- `wx.error` - 配置失败回调

### 分享接口
- `wx.updateAppMessageShareData` - 分享给朋友
- `wx.updateTimelineShareData` - 分享到朋友圈
- `wx.onMenuShareAppMessage` - 分享给朋友（旧版）
- `wx.onMenuShareTimeline` - 分享到朋友圈（旧版）

### 图像接口
- `wx.chooseImage` - 拍照或从手机相册中选图
- `wx.previewImage` - 预览图片
- `wx.uploadImage` - 上传图片
- `wx.downloadImage` - 下载图片

### 音频接口
- `wx.startRecord` - 开始录音
- `wx.stopRecord` - 停止录音
- `wx.onVoiceRecordEnd` - 录音时间超过一分钟没有停止
- `wx.playVoice` - 播放语音
- `wx.pauseVoice` - 暂停播放
- `wx.stopVoice` - 停止播放
- `wx.onVoicePlayEnd` - 播放语音完毕
- `wx.uploadVoice` - 上传语音
- `wx.downloadVoice` - 下载语音

### 智能接口
- `wx.translateVoice` - 识别音频并返回识别结果

### 设备信息
- `wx.getNetworkType` - 获取网络状态

### 地理位置
- `wx.openLocation` - 使用微信内置地图查看位置
- `wx.getLocation` - 获取地理位置

### 界面操作
- `wx.hideOptionMenu` - 隐藏右上角菜单
- `wx.showOptionMenu` - 显示右上角菜单
- `wx.closeWindow` - 关闭当前网页窗口
- `wx.hideMenuItems` - 批量隐藏功能按钮
- `wx.showMenuItems` - 批量显示功能按钮
- `wx.hideAllNonBaseMenuItem` - 隐藏所有非基础按钮
- `wx.showAllNonBaseMenuItem` - 显示所有功能按钮

### 微信扫一扫
- `wx.scanQRCode` - 调起微信扫一扫

### 微信小店
- `wx.openProductSpecificView` - 跳转微信商品页

### 微信卡券
- `wx.addCard` - 批量添加卡券
- `wx.chooseCard` - 调起适用于门店的卡券列表
- `wx.openCard` - 查看微信卡包中的卡券

### 微信支付
- `wx.chooseWXPay` - 发起一个微信支付请求

## wxopen组件实现

### JS-SDK配置生成

```go
// JSSDKConfig JS-SDK配置结构
type JSSDKConfig struct {
    AppID     string   `json:"appId" ini:"app_id"`          // 公众号appid
    Timestamp int64    `json:"timestamp" ini:"timestamp"`   // 时间戳
    NonceStr  string   `json:"nonceStr" ini:"nonce_str"`    // 随机字符串
    Signature string   `json:"signature" ini:"signature"`   // 签名
    JSAPIList []string `json:"jsApiList" ini:"js_api_list"` // JS-SDK调用权限列表
}

// JSSDKManager JS-SDK管理器
type JSSDKManager struct {
    authorizerClient *AuthorizerClient
    cacher           *JSSDKCacher
}

// 创建JS-SDK管理器
func (ac *AuthorizerClient) GetJSSDKManager() *JSSDKManager {
    return &JSSDKManager{
        authorizerClient: ac,
        cacher:           &JSSDKCacher{},
    }
}

// 生成JS-SDK配置
func (jm *JSSDKManager) GetConfig(ctx context.Context, url string, jsAPIList []string) (*JSSDKConfig, error) {
    // 验证参数
    if url == "" {
        return nil, fmt.Errorf("URL不能为空")
    }
    if len(jsAPIList) == 0 {
        return nil, fmt.Errorf("JSAPI列表不能为空")
    }
    if jm.authorizerClient.authorizerAppID == "" {
        return nil, fmt.Errorf("授权方AppID不能为空")
    }
    
    // 验证URL
    if err := jm.validateURL(url); err != nil {
        return nil, err
    }

    // 获取授权方AccessToken
    accessToken, err := jm.authorizerClient.authClient.client.GetAuthorizerAccessToken(ctx, jm.authorizerClient.authorizerAppID)
    if err != nil {
        return nil, fmt.Errorf("获取AccessToken失败: %v", err)
    }
    
    // 获取JSAPI Ticket
    ticket, err := jm.getJSAPITicket(ctx, accessToken)
    if err != nil {
        return nil, fmt.Errorf("获取JSAPI Ticket失败: %v", err)
    }
    
    // 生成签名
    config := jm.generateSignature(url, ticket, jsAPIList)
    
    return config, nil
}

// 获取JSAPI Ticket
func (jm *JSSDKManager) getJSAPITicket(ctx context.Context, accessToken string) (string, error) {
    apiURL := official_account.URLGetTicket

    params := map[string]interface{}{
        "access_token": accessToken,
        "type":         "jsapi",
    }

    var result struct {
        ErrCode   int    `json:"errcode"`
        ErrMsg    string `json:"errmsg"`
        Ticket    string `json:"ticket"`
        ExpiresIn int    `json:"expires_in"`
    }

    err := jm.authorizerClient.authClient.client.req.Make(ctx, &core.ReqMakeOpt{
        Method: "GET",
        URL:    apiURL,
        Query:  params,
        Result: &result,
    })
    if err != nil {
        return "", fmt.Errorf("调用微信API获取JSAPI Ticket失败: %v", err)
    }

    if result.ErrCode != 0 {
        return "", fmt.Errorf("获取JSAPI Ticket失败，错误码: %d, 错误信息: %s", result.ErrCode, result.ErrMsg)
    }

    if result.Ticket == "" {
        return "", fmt.Errorf("获取JSAPI Ticket失败，返回票据为空")
    }

    return result.Ticket, nil
}

// 验证URL
func (jm *JSSDKManager) validateURL(url string) error {
    // 验证URL格式
    parsedURL, err := url.Parse(url)
    if err != nil {
        return fmt.Errorf("URL格式错误: %v", err)
    }
    
    if parsedURL.Host == "" {
        return fmt.Errorf("URL缺少主机名")
    }
    
    // 验证协议
    // 注意：这里没有硬编码强制HTTPS，因为开发环境可能使用HTTP
    // 生产环境建议在配置中设置强制HTTPS
    
    return nil
}

// 生成签名
func (jm *JSSDKManager) generateSignature(url, ticket string, jsAPIList []string) *JSSDKConfig {
    nonceStr := generateNonceStr()
    timestamp := time.Now().Unix()
    
    // 签名算法
    signStr := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%d&url=%s", 
        ticket, nonceStr, timestamp, url)
    
    signature := sha1.Sum([]byte(signStr))
    signatureHex := hex.EncodeToString(signature[:])
    
    return &JSSDKConfig{
        AppID:     jm.authorizerClient.authorizerAppID,
        Timestamp: timestamp,
        NonceStr:  nonceStr,
        Signature: signatureHex,
        JSAPIList: jsAPIList,
    }
}

// 生成随机字符串
func generateNonceStr() string {
    const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    bytes := make([]byte, 16)
    for i := range bytes {
        bytes[i] = letters[rand.Intn(len(letters))]
    }
    return string(bytes)
}
```

### JS-SDK配置缓存

```go
// JSSDKCacher JS-SDK配置缓存器
type JSSDKCacher struct {
    cache sync.Map
}

// GetCachedConfig 获取缓存的JS-SDK配置
func (c *JSSDKCacher) GetCachedConfig(url string, jsAPIList []string) (*JSSDKConfig, error) {
    cacheKey := c.generateCacheKey(url, jsAPIList)

    if cached, exists := c.cache.Load(cacheKey); exists {
        if config, ok := cached.(*JSSDKConfig); ok {
            // 检查配置是否过期（7200秒有效期）
            if time.Now().Unix()-config.Timestamp < 7200 {
                return config, nil
            }
        }
    }

    return nil, fmt.Errorf("缓存未命中或已过期")
}

// CacheConfig 缓存JS-SDK配置
func (c *JSSDKCacher) CacheConfig(url string, jsAPIList []string, config *JSSDKConfig) {
    cacheKey := c.generateCacheKey(url, jsAPIList)
    c.cache.Store(cacheKey, config)
}

// generateCacheKey 生成缓存键
func (c *JSSDKCacher) generateCacheKey(url string, jsAPIList []string) string {
    sortedAPIs := make([]string, len(jsAPIList))
    copy(sortedAPIs, jsAPIList)
    sort.Strings(sortedAPIs)

    keyData := url + "|" + strings.Join(sortedAPIs, ",")

    hash := sha1.Sum([]byte(keyData))
    return hex.EncodeToString(hash[:])
}
```

## 完整示例

### 1. 后端JS-SDK配置接口

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/jcbowen/wego"
)

func main() {
    config := &wego.WxOpenConfig{
        ComponentAppID:     "your_component_appid",
        ComponentAppSecret: "your_component_appsecret",
        ComponentToken:     "your_component_token",
        EncodingAESKey:     "your_encoding_aes_key",
    }
    
    client := wego.NewWxOpenClient(config)
    
    // JS-SDK配置接口
    http.HandleFunc("/jssdk/config", func(w http.ResponseWriter, r *http.Request) {
        // 获取参数
        authorizerAppID := r.URL.Query().Get("authorizer_appid")
        url := r.URL.Query().Get("url")
        jsAPIListStr := r.URL.Query().Get("js_api_list")
        
        if authorizerAppID == "" || url == "" {
            http.Error(w, "参数缺失", http.StatusBadRequest)
            return
        }
        
        // 解析JSAPI列表
        var jsAPIList []string
        if jsAPIListStr != "" {
            if err := json.Unmarshal([]byte(jsAPIListStr), &jsAPIList); err != nil {
                http.Error(w, "JSAPI列表格式错误", http.StatusBadRequest)
                return
            }
        } else {
            // 默认JSAPI列表
            jsAPIList = []string{
                "updateAppMessageShareData",
                "updateTimelineShareData",
                "onMenuShareWeibo",
                "chooseImage",
                "previewImage",
                "uploadImage",
                "downloadImage",
                "getNetworkType",
                "openLocation",
                "getLocation",
                "hideOptionMenu",
                "showOptionMenu",
                "hideMenuItems",
                "showMenuItems",
                "hideAllNonBaseMenuItem",
                "showAllNonBaseMenuItem",
                "closeWindow",
            }
        }
        
        // 获取授权方客户端
        authorizerClient := client.GetAuthorizerClient(authorizerAppID)
        
        // 获取JS-SDK管理器
        jsSDKManager := authorizerClient.GetJSSDKManager()
        
        // 生成JS-SDK配置
        ctx := context.Background()
        jsConfig, err := jsSDKManager.GetConfig(ctx, url, jsAPIList)
        if err != nil {
            http.Error(w, "生成JS-SDK配置失败: "+err.Error(), http.StatusInternalServerError)
            return
        }
        
        // 返回配置
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(jsConfig)
    })
    
    fmt.Println("JS-SDK服务启动在 :8080")
    http.ListenAndServe(":8080", nil)
}
```

### 2. 前端JS-SDK使用示例

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>JS-SDK示例</title>
    <script src="https://res.wx.qq.com/open/js/jweixin-1.6.0.js"></script>
</head>
<body>
    <h1>JS-SDK功能演示</h1>
    
    <div>
        <button onclick="shareToFriend()">分享给朋友</button>
        <button onclick="shareToTimeline()">分享到朋友圈</button>
        <button onclick="chooseImage()">选择图片</button>
        <button onclick="getLocation()">获取位置</button>
        <button onclick="scanQRCode()">扫一扫</button>
    </div>
    
    <div id="result"></div>
    
    <script>
        // 从后端获取JS-SDK配置
        async function initJSSDK() {
            try {
                const currentURL = window.location.href.split('#')[0];
                
                const response = await fetch(`/jssdk/config?authorizer_appid=YOUR_AUTHORIZER_APPID&url=${encodeURIComponent(currentURL)}`);
                const config = await response.json();
                
                // 配置JS-SDK
                wx.config({
                    debug: false, // 开启调试模式
                    appId: config.appId,
                    timestamp: config.timestamp,
                    nonceStr: config.nonceStr,
                    signature: config.signature,
                    jsApiList: [
                        'updateAppMessageShareData',
                        'updateTimelineShareData',
                        'onMenuShareWeibo',
                        'chooseImage',
                        'previewImage',
                        'uploadImage',
                        'downloadImage',
                        'getNetworkType',
                        'openLocation',
                        'getLocation',
                        'hideOptionMenu',
                        'showOptionMenu',
                        'hideMenuItems',
                        'showMenuItems',
                        'hideAllNonBaseMenuItem',
                        'showAllNonBaseMenuItem',
                        'closeWindow',
                        'scanQRCode'
                    ]
                });
                
                // 配置成功回调
                wx.ready(function() {
                    console.log('JS-SDK配置成功');
                    
                    // 设置分享内容
                    setShareData();
                });
                
                // 配置失败回调
                wx.error(function(res) {
                    console.error('JS-SDK配置失败:', res);
                    document.getElementById('result').innerHTML = 'JS-SDK配置失败，请刷新页面重试';
                });
                
            } catch (error) {
                console.error('初始化JS-SDK失败:', error);
                document.getElementById('result').innerHTML = '初始化失败，请检查网络连接';
            }
        }
        
        // 设置分享内容
        function setShareData() {
            // 分享给朋友
            wx.updateAppMessageShareData({
                title: 'JS-SDK功能演示',
                desc: '这是一个JS-SDK功能演示页面',
                link: window.location.href,
                imgUrl: 'https://example.com/logo.png',
                success: function () {
                    console.log('分享给朋友设置成功');
                }
            });
            
            // 分享到朋友圈
            wx.updateTimelineShareData({
                title: 'JS-SDK功能演示',
                link: window.location.href,
                imgUrl: 'https://example.com/logo.png',
                success: function () {
                    console.log('分享到朋友圈设置成功');
                }
            });
        }
        
        // 分享给朋友
        function shareToFriend() {
            wx.updateAppMessageShareData({
                title: '自定义分享标题',
                desc: '自定义分享描述',
                link: window.location.href,
                imgUrl: 'https://example.com/share.png',
                success: function () {
                    document.getElementById('result').innerHTML = '分享给朋友设置成功';
                }
            });
        }
        
        // 分享到朋友圈
        function shareToTimeline() {
            wx.updateTimelineShareData({
                title: '朋友圈分享标题',
                link: window.location.href,
                imgUrl: 'https://example.com/share.png',
                success: function () {
                    document.getElementById('result').innerHTML = '分享到朋友圈设置成功';
                }
            });
        }
        
        // 选择图片
        function chooseImage() {
            wx.chooseImage({
                count: 1, // 默认9
                sizeType: ['original', 'compressed'], // 可以指定是原图还是压缩图，默认二者都有
                sourceType: ['album', 'camera'], // 可以指定来源是相册还是相机，默认二者都有
                success: function (res) {
                    var localIds = res.localIds; // 返回选定照片的本地ID列表，localId可以作为img标签的src属性显示图片
                    document.getElementById('result').innerHTML = '选择了图片: ' + localIds[0];
                }
            });
        }
        
        // 获取位置
        function getLocation() {
            wx.getLocation({
                type: 'wgs84', // 默认为wgs84的gps坐标，如果要返回直接给openLocation用的火星坐标，可传入'gcj02'
                success: function (res) {
                    var latitude = res.latitude; // 纬度，浮点数，范围为90 ~ -90
                    var longitude = res.longitude; // 经度，浮点数，范围为180 ~ -180
                    var speed = res.speed; // 速度，以米/每秒计
                    var accuracy = res.accuracy; // 位置精度
                    
                    document.getElementById('result').innerHTML = 
                        `纬度: ${latitude}<br>经度: ${longitude}<br>速度: ${speed}<br>精度: ${accuracy}`;
                }
            });
        }
        
        // 扫一扫
        function scanQRCode() {
            wx.scanQRCode({
                needResult: 1, // 默认为0，扫描结果由微信处理，1则直接返回扫描结果
                scanType: ["qrCode", "barCode"], // 可以指定扫二维码还是一维码，默认二者都有
                success: function (res) {
                    var result = res.resultStr; // 当needResult 为 1 时，扫码返回的结果
                    document.getElementById('result').innerHTML = '扫描结果: ' + result;
                }
            });
        }
        
        // 页面加载完成后初始化JS-SDK
        window.onload = initJSSDK;
    </script>
</body>
</html>
```

### 3. 高级JS-SDK功能示例

```javascript
// 网络状态检测
function checkNetwork() {
    wx.getNetworkType({
        success: function (res) {
            var networkType = res.networkType; // 返回网络类型2g，3g，4g，wifi
            alert('当前网络类型: ' + networkType);
        }
    });
}

// 打开地图
function openMap() {
    wx.getLocation({
        type: 'gcj02', // 返回可以用于wx.openLocation的经纬度
        success: function (res) {
            var latitude = res.latitude;
            var longitude = res.longitude;
            
            wx.openLocation({
                latitude: latitude, // 纬度，浮点数，范围为90 ~ -90
                longitude: longitude, // 经度，浮点数，范围为180 ~ -180
                name: '当前位置', // 位置名
                address: '详细地址', // 地址详情说明
                scale: 28, // 地图缩放级别,整形值,范围从1~28。默认为最大
                infoUrl: '' // 在查看位置界面底部显示的超链接,可点击跳转
            });
        }
    });
}

// 隐藏菜单项
function hideMenuItems() {
    wx.hideMenuItems({
        menuList: [
            'menuItem:share:appMessage',
            'menuItem:share:timeline',
            'menuItem:share:qq',
            'menuItem:share:weiboApp',
            'menuItem:favorite',
            'menuItem:share:facebook',
            'menuItem:share:QZone'
        ] // 要隐藏的菜单项
    });
}

// 显示菜单项
function showMenuItems() {
    wx.showMenuItems({
        menuList: [
            'menuItem:share:appMessage',
            'menuItem:share:timeline'
        ] // 要显示的菜单项
    });
}
```

## 最佳实践

### 1. 配置缓存优化

```go
// 优化后的JS-SDK配置获取
func (jm *JSSDKManager) GetConfigOptimized(ctx context.Context, url string, jsAPIList []string) (*JSSDKConfig, error) {
    // 先检查缓存
    if cachedConfig, err := jm.cacher.GetCachedConfig(url, jsAPIList); err == nil && cachedConfig != nil {
        return cachedConfig, nil
    }
    
    // 缓存不存在或已过期，重新生成
    config, err := jm.GetConfig(ctx, url, jsAPIList)
    if err != nil {
        return nil, err
    }
    
    // 缓存新配置
    jm.cacher.CacheConfig(url, jsAPIList, config)
    
    return config, nil
}
```

### 2. 错误处理与重试

```go
// 带重试的JS-SDK配置获取
func (jm *JSSDKManager) GetConfigWithRetry(ctx context.Context, url string, jsAPIList []string, maxRetries int) (*JSSDKConfig, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        config, err := jm.GetConfigOptimized(ctx, url, jsAPIList)
        if err == nil {
            return config, nil
        }
        
        lastErr = err
        
        // 检查context是否取消
        select {
        case <-ctx.Done():
            return nil, fmt.Errorf("context cancelled: %v, last error: %v", ctx.Err(), lastErr)
        case <-time.After(time.Duration(i+1) * time.Second):
            continue // 指数退避重试
        }
    }
    
    return nil, fmt.Errorf("获取JS-SDK配置失败，重试%d次后仍然失败: %v", maxRetries, lastErr)
}
```

### 3. 缓存统计和监控

```go
// 获取缓存统计信息
stats := jsSDKManager.cacher.GetStats()
fmt.Printf("缓存命中率: %.2f%%\n", float64(stats.Hits)/float64(stats.Hits+stats.Misses)*100)
fmt.Printf("总缓存项: %d\n", stats.TotalItems)
fmt.Printf("缓存淘汰: %d\n", stats.Evictions)

// 定期清理过期缓存
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        jsSDKManager.cacher.CleanupExpired()
        log.Printf("缓存清理完成，当前统计: %+v", jsSDKManager.cacher.GetStats())
    }
}()
```

### 4. 错误处理增强

```go
// 使用标准错误类型进行错误判断
config, err := jsSDKManager.GetConfig(ctx, url, jsAPIList)
if err != nil {
    switch {
    case errors.Is(err, ErrInvalidURL):
        log.Printf("URL格式错误: %v", err)
    case errors.Is(err, ErrEmptyJSAPIList):
        log.Printf("JSAPI列表为空")
    case errors.Is(err, ErrCacheMiss):
        log.Printf("缓存未命中，重新获取")
    default:
        log.Printf("获取配置失败: %v", err)
    }
}
```

## 注意事项

### 1. 签名算法
- URL必须与当前页面URL完全一致
- 签名参数必须按字典序排序
- 签名算法使用SHA1

### 2. 缓存策略
- JSAPI Ticket有效期为7200秒
- 建议实现本地缓存减少API调用
- 注意缓存失效时的处理

### 3. 安全考虑
- 生产环境必须使用HTTPS
- 验证请求来源防止恶意调用
- 限制JSAPI列表防止权限滥用

### 4. 兼容性
- 注意新版和旧版API的兼容性
- 测试不同微信版本的兼容性
- 提供降级方案

通过wxopen组件的JS-SDK功能，您可以轻松实现微信网页的各种高级功能，为用户提供原生应用般的体验。