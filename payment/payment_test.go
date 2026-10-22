package payment

import (
	"bytes"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/logger"
)

// ---------- 测试辅助 ----------

// mockHTTPClient 用于拦截 HTTP 请求，避免测试发起真实网络调用
type mockHTTPClient struct {
	statusCode int
	body       []byte
	headers    http.Header
	lastReq    *http.Request
	lastBody   []byte
}

// Do 实现 core.HTTPClient 接口
func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.lastReq = req
	if req.Body != nil {
		m.lastBody, _ = io.ReadAll(req.Body)
	}
	if m.headers == nil {
		m.headers = http.Header{}
	}
	return &http.Response{
		StatusCode: m.statusCode,
		Header:     m.headers,
		Body:       io.NopCloser(bytes.NewReader(m.body)),
	}, nil
}

// testKeyMaterial 测试用密钥材料
type testKeyMaterial struct {
	privPEM string
	pubPEM  string
	priv    *rsa.PrivateKey
	pub     *rsa.PublicKey
}

// genKeyMaterial 生成测试用 RSA 密钥对及 PEM 文本
func genKeyMaterial(t *testing.T) testKeyMaterial {
	t.Helper()
	key, err := rsa.GenerateKey(crand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}
	privDER, _ := x509.MarshalPKCS8PrivateKey(key)
	pubDER, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	return testKeyMaterial{
		privPEM: string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})),
		pubPEM:  string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})),
		priv:    key,
		pub:     &key.PublicKey,
	}
}

// newTestClient 构造使用 mock HTTP 客户端的支付客户端（公钥模式）
func newTestClient(t *testing.T, mock core.HTTPClient) (*Client, testKeyMaterial) {
	t.Helper()
	km := genKeyMaterial(t)
	cfg := &Config{
		AppID:        "wx1234567890",
		MchID:        "1900000001",
		APIV3Key:     strings.Repeat("a", 32),
		CertSerialNo: "MERCHANT_CERT_SERIAL",
		PrivateKey:   km.privPEM,
		PublicKeyID:  "PUB_KEY_ID_TEST",
		PublicKey:    km.pubPEM,
	}
	c := &Client{
		config:     cfg,
		httpClient: mock,
		logger:     logger.NewDefaultLoggerInterface(),
		privateKey: km.priv,
		publicKey:  km.pub,
		certs:      map[string]*x509.Certificate{},
	}
	c.req = core.NewRequest(c.httpClient, c.logger)
	return c, km
}

// signedHeaders 用指定私钥生成可通过验签的响应头
func signedHeaders(t *testing.T, key *rsa.PrivateKey, serial, body string) http.Header {
	t.Helper()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := "testnonce1234"
	message := timestamp + "\n" + nonce + "\n" + body + "\n"
	hash := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPKCS1v15(crand.Reader, key, crypto.SHA256, hash[:])
	if err != nil {
		t.Fatalf("生成响应签名失败: %v", err)
	}
	h := http.Header{}
	h.Set("Wechatpay-Serial", serial)
	h.Set("Wechatpay-Timestamp", timestamp)
	h.Set("Wechatpay-Nonce", nonce)
	h.Set("Wechatpay-Signature", base64.StdEncoding.EncodeToString(sig))
	return h
}

// ---------- 密钥与配置 ----------

func TestParsePrivateKey(t *testing.T) {
	km := genKeyMaterial(t)

	got, err := parsePrivateKey(km.privPEM)
	if err != nil {
		t.Fatalf("PKCS8 私钥解析失败: %v", err)
	}
	if !got.Equal(km.priv) {
		t.Fatal("PKCS8 私钥解析结果不匹配")
	}

	pkcs1 := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(km.priv)}))
	if _, err := parsePrivateKey(pkcs1); err != nil {
		t.Fatalf("PKCS1 私钥解析失败: %v", err)
	}

	if _, err := parsePrivateKey("not-a-pem"); err == nil {
		t.Fatal("非法私钥 PEM 应报错")
	}
}

func TestParsePublicKey(t *testing.T) {
	km := genKeyMaterial(t)

	got, err := parsePublicKey(km.pubPEM)
	if err != nil {
		t.Fatalf("PKIX 公钥解析失败: %v", err)
	}
	if !got.Equal(km.pub) {
		t.Fatal("PKIX 公钥解析结果不匹配")
	}

	pkcs1 := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(km.pub)}))
	if _, err := parsePublicKey(pkcs1); err != nil {
		t.Fatalf("PKCS1 公钥解析失败: %v", err)
	}

	if _, err := parsePublicKey("not-a-pem"); err == nil {
		t.Fatal("非法公钥 PEM 应报错")
	}
}

func TestConfigValidate(t *testing.T) {
	km := genKeyMaterial(t)
	base := Config{AppID: "wx1", MchID: "1900000001", APIV3Key: strings.Repeat("a", 32), CertSerialNo: "SER", PrivateKey: km.privPEM}

	if err := base.Validate(); err != nil {
		t.Fatalf("仅平台证书模式应通过校验: %v", err)
	}
	if base.UsePublicKey() {
		t.Fatal("未配置公钥时不应进入公钥模式")
	}

	both := base
	both.PublicKeyID = "PUB_KEY_ID_123"
	both.PublicKey = km.pubPEM
	if err := both.Validate(); err != nil {
		t.Fatalf("公钥与证书同时配置应通过校验: %v", err)
	}
	if !both.UsePublicKey() {
		t.Fatal("配置公钥后应进入公钥模式")
	}

	onlyID := base
	onlyID.PublicKeyID = "PUB_KEY_ID_123"
	if err := onlyID.Validate(); err == nil {
		t.Fatal("只配 PublicKeyID 应报错")
	}

	onlyKey := base
	onlyKey.PublicKey = km.pubPEM
	if err := onlyKey.Validate(); err == nil {
		t.Fatal("只配 PublicKey 应报错")
	}

	shortKey := base
	shortKey.APIV3Key = "short"
	if err := shortKey.Validate(); err == nil {
		t.Fatal("APIV3Key 长度不足应报错")
	}
}

// ---------- 验签路由 ----------

func TestResolveVerifyPublicKey(t *testing.T) {
	km := genKeyMaterial(t)
	c := &Client{config: &Config{PublicKeyID: "PUB_KEY_ID_TEST"}, publicKey: km.pub, certs: map[string]*x509.Certificate{}}

	got, err := c.resolveVerifyPublicKey(context.Background(), "PUB_KEY_ID_TEST")
	if err != nil {
		t.Fatalf("公钥 ID 应路由到公钥: %v", err)
	}
	if !got.Equal(km.pub) {
		t.Fatal("应返回配置的微信支付公钥")
	}

	if _, err := c.resolveVerifyPublicKey(context.Background(), ""); err == nil {
		t.Fatal("空 Wechatpay-Serial 应报错")
	}

	certOnly := &Client{config: &Config{}, certs: map[string]*x509.Certificate{}}
	_, err = certOnly.resolveVerifyPublicKey(context.Background(), "PUB_KEY_ID_TEST")
	if err == nil {
		t.Fatal("未配置公钥却收到公钥签名应报错")
	}
	if !strings.Contains(err.Error(), "PublicKeyID") {
		t.Fatalf("错误信息应提示补齐配置，实际: %v", err)
	}
}

func TestVerifyResponseSignatureRejectsStaleTimestamp(t *testing.T) {
	km := genKeyMaterial(t)
	c := &Client{config: &Config{PublicKeyID: "PUB_KEY_ID_TEST"}, publicKey: km.pub, certs: map[string]*x509.Certificate{}}

	timestamp := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)
	headers := http.Header{}
	headers.Set("Wechatpay-Serial", "PUB_KEY_ID_TEST")
	headers.Set("Wechatpay-Timestamp", timestamp)
	headers.Set("Wechatpay-Nonce", "n")
	headers.Set("Wechatpay-Signature", base64.StdEncoding.EncodeToString([]byte("x")))

	if err := c.verifyResponseSignature(context.Background(), 200, headers, []byte("{}")); err == nil {
		t.Fatal("时间戳偏差超过 5 分钟应拒绝")
	}
}

// ---------- 调起支付参数 ----------

func TestBuildJSAPIPayParams(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 204}
	c, km := newTestClient(t, mock)

	params, err := c.BuildJSAPIPayParams("wx2912")
	if err != nil {
		t.Fatalf("生成 JSAPI 参数失败: %v", err)
	}
	if params.Package != "prepay_id=wx2912" || params.SignType != "RSA" || params.AppID != "wx1234567890" {
		t.Fatalf("JSAPI 参数内容异常: %+v", params)
	}

	message := strings.Join([]string{params.AppID, params.TimeStamp, params.NonceStr, params.Package}, "\n") + "\n"
	hash := sha256.Sum256([]byte(message))
	sig, err := base64.StdEncoding.DecodeString(params.PaySign)
	if err != nil {
		t.Fatalf("paySign 非 base64: %v", err)
	}
	if err := rsa.VerifyPKCS1v15(km.pub, crypto.SHA256, hash[:], sig); err != nil {
		t.Fatalf("paySign 验签失败: %v", err)
	}

	if _, err := c.BuildJSAPIPayParams(""); err == nil {
		t.Fatal("空 prepay_id 应报错")
	}
}

func TestBuildAppPayParams(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 204}
	c, km := newTestClient(t, mock)

	params, err := c.BuildAppPayParams("WX1217752501201407033233368018")
	if err != nil {
		t.Fatalf("生成 App 参数失败: %v", err)
	}
	if params.Package != AppPayPackageValue || params.PartnerID != "1900000001" || params.AppID != "wx1234567890" {
		t.Fatalf("App 参数内容异常: %+v", params)
	}

	// 官方签名串为四行：appid\ntimestamp\nnoncestr\nprepayid\n
	message := strings.Join([]string{params.AppID, params.Timestamp, params.NonceStr, params.PrepayID}, "\n") + "\n"
	hash := sha256.Sum256([]byte(message))
	sig, err := base64.StdEncoding.DecodeString(params.Sign)
	if err != nil {
		t.Fatalf("sign 非 base64: %v", err)
	}
	if err := rsa.VerifyPKCS1v15(km.pub, crypto.SHA256, hash[:], sig); err != nil {
		t.Fatalf("App sign 验签失败: %v", err)
	}

	if _, err := c.BuildAppPayParams(""); err == nil {
		t.Fatal("空 prepay_id 应报错")
	}
}

// ---------- 敏感信息加密 ----------

func TestEncryptSensitiveData(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 204}
	c, km := newTestClient(t, mock)

	cipher, err := c.EncryptSensitiveData(context.Background(), "张三")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	raw, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		t.Fatalf("密文非 base64: %v", err)
	}
	plain, err := rsa.DecryptOAEP(sha1.New(), nil, km.priv, raw, nil)
	if err != nil {
		t.Fatalf("OAEP 解密失败: %v", err)
	}
	if string(plain) != "张三" {
		t.Fatalf("解密结果不匹配，实际: %s", plain)
	}

	serial, err := c.EncryptSerialNo(context.Background())
	if err != nil {
		t.Fatalf("获取加密序列号失败: %v", err)
	}
	if serial != "PUB_KEY_ID_TEST" {
		t.Fatalf("公钥模式下应返回公钥 ID，实际: %s", serial)
	}

	if _, err := EncryptSensitiveWithPublicKey(nil, "x"); err == nil {
		t.Fatal("公钥为空应报错")
	}
	if _, err := c.EncryptSensitiveData(context.Background(), ""); err == nil {
		t.Fatal("明文为空应报错")
	}
}

// ---------- 下单 / 关单 / 退款 ----------

func TestNativePay(t *testing.T) {
	body := `{"code_url":"weixin://wxpay/bizpayurl?pr=abcd"}`
	mock := &mockHTTPClient{statusCode: 200, body: []byte(body)}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", body)

	resp, err := c.NativePay(context.Background(), &NativeRequest{
		Description: "测试商品",
		OutTradeNo:  "ORDER_001",
		Amount:      Amount{Total: 100},
	})
	if err != nil {
		t.Fatalf("NativePay 失败: %v", err)
	}
	if resp.CodeURL == "" {
		t.Fatal("code_url 不应为空")
	}
	if mock.lastReq.URL.String() != PayNativeURL {
		t.Fatalf("请求地址错误: %s", mock.lastReq.URL.String())
	}
	if mock.lastReq.Header.Get("Wechatpay-Serial") != "PUB_KEY_ID_TEST" {
		t.Fatal("公钥模式应携带 Wechatpay-Serial 请求头")
	}

	var sent map[string]interface{}
	if err := json.Unmarshal(mock.lastBody, &sent); err != nil {
		t.Fatalf("请求体解析失败: %v", err)
	}
	if sent["appid"] != "wx1234567890" || sent["mchid"] != "1900000001" {
		t.Fatalf("appid/mchid 未自动填充: %+v", sent)
	}
}

func TestAppPay(t *testing.T) {
	body := `{"prepay_id":"WX1217752501201407033233368018"}`
	mock := &mockHTTPClient{statusCode: 200, body: []byte(body)}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", body)

	resp, err := c.AppPay(context.Background(), &AppRequest{
		Description: "测试商品",
		OutTradeNo:  "ORDER_001",
		Amount:      Amount{Total: 100},
	})
	if err != nil {
		t.Fatalf("AppPay 失败: %v", err)
	}
	if resp.PrepayID == "" {
		t.Fatal("prepay_id 不应为空")
	}
	if mock.lastReq.URL.String() != PayAppURL {
		t.Fatalf("请求地址错误: %s", mock.lastReq.URL.String())
	}
}

func TestCodePay(t *testing.T) {
	// 扣款成功：同步返回订单结果
	body := `{"appid":"wx1234567890","mchid":"1900000001","out_trade_no":"ORDER_001","transaction_id":"1009665485201508060766242864","trade_type":"MICROPAY","trade_state":"SUCCESS","trade_state_desc":"支付成功","amount":{"total":100,"payer_total":100,"currency":"CNY","payer_currency":"CNY"}}`
	mock := &mockHTTPClient{statusCode: 200, body: []byte(body)}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", body)

	order, err := c.CodePay(context.Background(), &CodePayRequest{
		Description: "测试商品",
		OutTradeNo:  "ORDER_001",
		Amount:      Amount{Total: 100},
		Payer:       CodePayer{AuthCode: "1345235315201154982642655744"},
		SceneInfo:   &CodePaySceneInfo{StoreInfo: &CodePayStoreInfo{ID: "0001"}},
	})
	if err != nil {
		t.Fatalf("CodePay 失败: %v", err)
	}
	if order.TradeState != TradeStateSuccess || order.TransactionID == "" {
		t.Fatalf("扣款成功结果异常: %+v", order)
	}
	if order.Amount == nil || order.Amount.PayerTotal != 100 {
		t.Fatalf("实付金额解析异常: %+v", order.Amount)
	}
	if mock.lastReq.URL.String() != PayCodePayURL {
		t.Fatalf("请求地址错误: %s", mock.lastReq.URL.String())
	}

	var sent map[string]interface{}
	if err := json.Unmarshal(mock.lastBody, &sent); err != nil {
		t.Fatalf("请求体解析失败: %v", err)
	}
	if sent["appid"] != "wx1234567890" || sent["mchid"] != "1900000001" {
		t.Fatalf("appid/mchid 未自动填充: %+v", sent)
	}
	if payer, ok := sent["payer"].(map[string]interface{}); !ok || payer["auth_code"] != "1345235315201154982642655744" {
		t.Fatalf("auth_code 未携带: %+v", sent["payer"])
	}
	if _, exists := sent["notify_url"]; exists {
		t.Fatalf("官方 codepay 参数表无 notify_url，不应携带: %+v", sent)
	}

	// USERPAYING：用户支付中，返回订单结果由业务轮询查单
	payingBody := `{"out_trade_no":"ORDER_002","trade_state":"USERPAYING","trade_state_desc":"用户支付中"}`
	mock.body = []byte(payingBody)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", payingBody)
	paying, err := c.CodePay(context.Background(), &CodePayRequest{
		Description: "测试商品",
		OutTradeNo:  "ORDER_002",
		Amount:      Amount{Total: 100},
		Payer:       CodePayer{AuthCode: "1345235315201154982642655745"},
		SceneInfo:   &CodePaySceneInfo{StoreInfo: &CodePayStoreInfo{OutID: "SHOP_A1"}},
	})
	if err != nil {
		t.Fatalf("CodePay(USERPAYING) 失败: %v", err)
	}
	if paying.TradeState != TradeStateUserPaying {
		t.Fatalf("USERPAYING 结果异常: %+v", paying)
	}

	// 可选字段：detail / settle_info / support_fapiao / scene_info 序列化
	payingBody2 := `{"out_trade_no":"ORDER_003","trade_state":"USERPAYING"}`
	mock.body = []byte(payingBody2)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", payingBody2)
	_, err = c.CodePay(context.Background(), &CodePayRequest{
		Description:   "测试商品",
		OutTradeNo:    "ORDER_003",
		Amount:        Amount{Total: 100},
		Payer:         CodePayer{AuthCode: "1345235315201154982642655746"},
		SupportFapiao: true,
		Detail: &OrderDetail{
			CostPrice: 608800,
			InvoiceID: "微信123",
			GoodsDetail: []GoodsDetail{{
				MerchantGoodsID: "1246464644",
				WxpayGoodsID:    "1001",
				GoodsName:       "iPhoneX 256G",
				Quantity:        1,
				UnitPrice:       528800,
			}},
		},
		SettleInfo: &SettleInfo{ProfitSharing: true},
		SceneInfo: &CodePaySceneInfo{
			DeviceID:  "013467007045764",
			StoreInfo: &CodePayStoreInfo{ID: "0001", Name: "腾讯大厦分店", AreaCode: "440305", Address: "广东省深圳市南山区科技中一道10000号"},
		},
	})
	if err != nil {
		t.Fatalf("CodePay(可选字段) 失败: %v", err)
	}
	if err := json.Unmarshal(mock.lastBody, &sent); err != nil {
		t.Fatalf("请求体解析失败: %v", err)
	}
	if sent["support_fapiao"] != true {
		t.Fatalf("support_fapiao 未携带: %+v", sent)
	}
	if detail, ok := sent["detail"].(map[string]interface{}); !ok || detail["cost_price"] != float64(608800) {
		t.Fatalf("detail 未携带: %+v", sent["detail"])
	} else if goods, ok := detail["goods_detail"].([]interface{}); !ok || len(goods) != 1 {
		t.Fatalf("goods_detail 未携带: %+v", detail)
	} else if item, ok := goods[0].(map[string]interface{}); !ok || item["merchant_goods_id"] != "1246464644" || item["quantity"] != float64(1) {
		t.Fatalf("goods_detail 元素异常: %+v", goods[0])
	} else if item["wxpay_goods_id"] != "1001" {
		t.Fatalf("wxpay_goods_id 未携带: %+v", item)
	}
	if settle, ok := sent["settle_info"].(map[string]interface{}); !ok || settle["profit_sharing"] != true {
		t.Fatalf("settle_info 未携带: %+v", sent["settle_info"])
	}
	if scene, ok := sent["scene_info"].(map[string]interface{}); !ok || scene["device_id"] != "013467007045764" {
		t.Fatalf("scene_info 未携带: %+v", sent["scene_info"])
	} else if store, ok := scene["store_info"].(map[string]interface{}); !ok || store["id"] != "0001" {
		t.Fatalf("store_info 未携带: %+v", scene["store_info"])
	}
}

func TestReverseOrder(t *testing.T) {
	// 官方应答为 200 OK + JSON（appid/mchid/out_trade_no）
	reverseBody := `{"appid":"wx1234567890","mchid":"1900000001","out_trade_no":"ORDER_001"}`
	mock := &mockHTTPClient{statusCode: 200, body: []byte(reverseBody)}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", reverseBody)

	if err := c.ReverseOrder(context.Background(), "ORDER_001"); err != nil {
		t.Fatalf("ReverseOrder 失败: %v", err)
	}
	wantURL := strings.Replace(PayReverseOrderURL, "%s", "ORDER_001", 1)
	if mock.lastReq.URL.String() != wantURL {
		t.Fatalf("请求地址错误: %s", mock.lastReq.URL.String())
	}
	var sent map[string]interface{}
	if err := json.Unmarshal(mock.lastBody, &sent); err != nil {
		t.Fatalf("请求体解析失败: %v", err)
	}
	if sent["appid"] != "wx1234567890" || sent["mchid"] != "1900000001" {
		t.Fatalf("appid/mchid 未自动填充: %+v", sent)
	}

	// 商户订单号为空必须拒绝
	if err := c.ReverseOrder(context.Background(), ""); err == nil {
		t.Fatal("空订单号未报错")
	}
}

func TestCloseOrder(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 204}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", "")

	if err := c.CloseOrder(context.Background(), "ORDER_001"); err != nil {
		t.Fatalf("CloseOrder 失败: %v", err)
	}
	if got := mock.lastReq.URL.Path; got != "/v3/pay/transactions/out-trade-no/ORDER_001/close" {
		t.Fatalf("关单路径错误: %s", got)
	}
	if mock.lastReq.Method != http.MethodPost {
		t.Fatalf("关单应为 POST，实际: %s", mock.lastReq.Method)
	}

	var sent map[string]interface{}
	if err := json.Unmarshal(mock.lastBody, &sent); err != nil {
		t.Fatalf("请求体解析失败: %v", err)
	}
	if sent["mchid"] != "1900000001" {
		t.Fatalf("请求体应携带 mchid: %+v", sent)
	}
}

func TestQueryOrderURLs(t *testing.T) {
	body := `{"out_trade_no":"ORDER_001","trade_state":"SUCCESS"}`
	mock := &mockHTTPClient{statusCode: 200, body: []byte(body)}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", body)

	if _, err := c.QueryOrder(context.Background(), "ORDER_001"); err != nil {
		t.Fatalf("QueryOrder 失败: %v", err)
	}
	if got := mock.lastReq.URL.Path; got != "/v3/pay/transactions/out-trade-no/ORDER_001" {
		t.Fatalf("查单路径错误: %s", got)
	}
	if got := mock.lastReq.URL.Query().Get("mchid"); got != "1900000001" {
		t.Fatalf("查单应带 mchid 参数: %s", got)
	}

	if _, err := c.QueryOrderByTransactionID(context.Background(), "4200001234"); err != nil {
		t.Fatalf("QueryOrderByTransactionID 失败: %v", err)
	}
	if got := mock.lastReq.URL.Path; got != "/v3/pay/transactions/id/4200001234" {
		t.Fatalf("按微信单号查单路径错误: %s", got)
	}
}

func TestRefundQuery(t *testing.T) {
	body := `{"refund_id":"5000000001","out_refund_no":"REFUND_001","status":"SUCCESS","amount":{"refund":100,"total":100}}`
	mock := &mockHTTPClient{statusCode: 200, body: []byte(body)}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", body)

	resp, err := c.RefundQuery(context.Background(), "REFUND_001")
	if err != nil {
		t.Fatalf("RefundQuery 失败: %v", err)
	}
	if resp.Status != "SUCCESS" || resp.Amount.Refund != 100 {
		t.Fatalf("退款查询结果异常: %+v", resp)
	}
	if got := mock.lastReq.URL.Path; got != "/v3/refund/domestic/refunds/REFUND_001" {
		t.Fatalf("退款查询路径错误: %s", got)
	}
}

// ---------- 账单 ----------

func TestApplyTradeBill(t *testing.T) {
	body := `{"hash_type":"SHA1","hash_value":"abc","download_url":"https://api.mch.weixin.qq.com/v3/billdownload/file?token=x"}`
	mock := &mockHTTPClient{statusCode: 200, body: []byte(body)}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", body)

	resp, err := c.ApplyTradeBill(context.Background(), &BillRequest{BillDate: "2026-10-14", BillType: "ALL", TarType: "GZIP"})
	if err != nil {
		t.Fatalf("ApplyTradeBill 失败: %v", err)
	}
	if resp.HashType != BillHashTypeSHA1 || resp.DownloadURL == "" {
		t.Fatalf("账单响应异常: %+v", resp)
	}

	q := mock.lastReq.URL.Query()
	if q.Get("bill_date") != "2026-10-14" || q.Get("bill_type") != "ALL" || q.Get("tar_type") != "GZIP" {
		t.Fatalf("账单查询参数异常: %v", q)
	}
}

func TestApplyFundFlowBill(t *testing.T) {
	body := `{"hash_type":"SHA1","hash_value":"abc","download_url":"https://x"}`
	mock := &mockHTTPClient{statusCode: 200, body: []byte(body)}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", body)

	if _, err := c.ApplyFundFlowBill(context.Background(), &BillRequest{BillDate: "2026-10-14", AccountType: "BASIC"}); err != nil {
		t.Fatalf("ApplyFundFlowBill 失败: %v", err)
	}
	if got := mock.lastReq.URL.Path; got != "/v3/bill/fundflowbill" {
		t.Fatalf("资金账单路径错误: %s", got)
	}
	if got := mock.lastReq.URL.Query().Get("account_type"); got != "BASIC" {
		t.Fatalf("account_type 参数错误: %s", got)
	}
}

func TestDownloadBill(t *testing.T) {
	content := []byte("交易时间,订单号\n2026-10-14,ORDER_001\n")
	sum := sha1.Sum(content)
	hashValue := hex.EncodeToString(sum[:])

	mock := &mockHTTPClient{statusCode: 200, body: content}
	c, _ := newTestClient(t, mock)

	got, err := c.DownloadBill(context.Background(), "https://api.mch.weixin.qq.com/v3/billdownload/file?token=x", hashValue)
	if err != nil {
		t.Fatalf("DownloadBill 失败: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("账单内容不一致: %s", got)
	}

	if _, err := c.DownloadBill(context.Background(), "https://x", "wronghash"); err == nil {
		t.Fatal("哈希不一致应报错")
	}

	if _, err := c.DownloadBill(context.Background(), "", ""); err == nil {
		t.Fatal("下载地址为空应报错")
	}
}

// ---------- 分账 ----------

func TestProfitSharing(t *testing.T) {
	body := `{"transaction_id":"4200001234","out_order_no":"PS001","order_id":"3001","state":"PROCESSING","receivers":[{"amount":100,"type":"MERCHANT_ID","account":"1900000109","result":"PENDING"}]}`
	mock := &mockHTTPClient{statusCode: 200, body: []byte(body)}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", body)

	name, err := c.EncryptSensitiveData(context.Background(), "张三")
	if err != nil {
		t.Fatalf("加密姓名失败: %v", err)
	}

	resp, err := c.ProfitSharing(context.Background(), &ProfitSharingRequest{
		TransactionID:   "4200001234",
		OutOrderNo:      "PS001",
		UnfreezeUnsplit: true,
		Receivers: []ProfitSharingReceiver{
			{Type: "MERCHANT_ID", Account: "1900000109", Name: name, Amount: 100, Description: "分账给商户"},
		},
	})
	if err != nil {
		t.Fatalf("ProfitSharing 失败: %v", err)
	}
	if resp.State != "PROCESSING" || len(resp.Receivers) != 1 {
		t.Fatalf("分账响应异常: %+v", resp)
	}
	if got := mock.lastReq.Header.Get("Wechatpay-Serial"); got != "PUB_KEY_ID_TEST" {
		t.Fatalf("分账请求头应带加密所用公钥 ID，实际: %s", got)
	}
}

func TestProfitSharingQuery(t *testing.T) {
	body := `{"out_order_no":"PS001","order_id":"3001","state":"FINISHED"}`
	mock := &mockHTTPClient{statusCode: 200, body: []byte(body)}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", body)

	resp, err := c.ProfitSharingQuery(context.Background(), "PS001", "4200001234")
	if err != nil {
		t.Fatalf("ProfitSharingQuery 失败: %v", err)
	}
	if resp.State != "FINISHED" {
		t.Fatalf("分账状态异常: %s", resp.State)
	}
	if got := mock.lastReq.URL.Path; got != "/v3/profitsharing/orders/PS001" {
		t.Fatalf("查询分账路径错误: %s", got)
	}
	if got := mock.lastReq.URL.Query().Get("transaction_id"); got != "4200001234" {
		t.Fatalf("transaction_id 参数错误: %s", got)
	}
}

func TestProfitSharingReturn(t *testing.T) {
	body := `{"out_order_no":"PS001","out_return_no":"RT001","result":"PROCESSING","amount":100}`
	mock := &mockHTTPClient{statusCode: 200, body: []byte(body)}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", body)

	resp, err := c.ProfitSharingReturn(context.Background(), &ProfitSharingReturnRequest{
		OutOrderNo:  "PS001",
		OutReturnNo: "RT001",
		ReturnMchID: "1900000109",
		Amount:      100,
		Description: "分账回退",
	})
	if err != nil {
		t.Fatalf("ProfitSharingReturn 失败: %v", err)
	}
	if resp.Result != "PROCESSING" {
		t.Fatalf("回退结果异常: %s", resp.Result)
	}
}

func TestProfitSharingReturnQuery(t *testing.T) {
	body := `{"out_return_no":"RT001","out_order_no":"PS001","result":"SUCCESS"}`
	mock := &mockHTTPClient{statusCode: 200, body: []byte(body)}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", body)

	if _, err := c.ProfitSharingReturnQuery(context.Background(), "RT001", "PS001"); err != nil {
		t.Fatalf("ProfitSharingReturnQuery 失败: %v", err)
	}
	if got := mock.lastReq.URL.Path; got != "/v3/profitsharing/return-orders/RT001" {
		t.Fatalf("回退查询路径错误: %s", got)
	}
}

func TestProfitSharingUnfreezeAndAmounts(t *testing.T) {
	body := `{"out_order_no":"PS001","state":"FINISHED"}`
	mock := &mockHTTPClient{statusCode: 200, body: []byte(body)}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", body)

	if _, err := c.ProfitSharingUnfreeze(context.Background(), &ProfitSharingUnfreezeRequest{
		TransactionID: "4200001234", OutOrderNo: "PS001", Description: "解冻",
	}); err != nil {
		t.Fatalf("ProfitSharingUnfreeze 失败: %v", err)
	}
	if got := mock.lastReq.URL.Path; got != "/v3/profitsharing/orders/unfreeze" {
		t.Fatalf("解冻路径错误: %s", got)
	}

	amountBody := `{"transaction_id":"4200001234","unsplit_amount":50}`
	mock.body = []byte(amountBody)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", amountBody)
	amounts, err := c.ProfitSharingAmounts(context.Background(), "4200001234")
	if err != nil {
		t.Fatalf("ProfitSharingAmounts 失败: %v", err)
	}
	if amounts.UnsplitAmount != 50 {
		t.Fatalf("剩余待分金额异常: %d", amounts.UnsplitAmount)
	}
}

// ---------- 参数校验 ----------

func TestParameterValidation(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 204}
	c, _ := newTestClient(t, mock)
	ctx := context.Background()

	cases := []struct {
		name string
		fn   func() error
	}{
		{"NativePay 空请求", func() error { _, err := c.NativePay(ctx, nil); return err }},
		{"AppPay 空请求", func() error { _, err := c.AppPay(ctx, nil); return err }},
		{"CodePay 空请求", func() error { _, err := c.CodePay(ctx, nil); return err }},
		{"CodePay 空付款码", func() error {
			_, err := c.CodePay(ctx, &CodePayRequest{Description: "测试商品", OutTradeNo: "ORDER_001", Amount: Amount{Total: 100}})
			return err
		}},
		{"CodePay 缺场景信息", func() error {
			_, err := c.CodePay(ctx, &CodePayRequest{Description: "测试商品", OutTradeNo: "ORDER_001", Amount: Amount{Total: 100}, Payer: CodePayer{AuthCode: "1345235315201154982642655744"}})
			return err
		}},
		{"CodePay 缺门店信息", func() error {
			_, err := c.CodePay(ctx, &CodePayRequest{
				Description: "测试商品", OutTradeNo: "ORDER_001", Amount: Amount{Total: 100},
				Payer:     CodePayer{AuthCode: "1345235315201154982642655744"},
				SceneInfo: &CodePaySceneInfo{},
			})
			return err
		}},
		{"CodePay 门店编号缺失", func() error {
			_, err := c.CodePay(ctx, &CodePayRequest{
				Description: "测试商品", OutTradeNo: "ORDER_001", Amount: Amount{Total: 100},
				Payer:     CodePayer{AuthCode: "1345235315201154982642655744"},
				SceneInfo: &CodePaySceneInfo{StoreInfo: &CodePayStoreInfo{Name: "门店"}},
			})
			return err
		}},
		{"CloseOrder 空订单号", func() error { return c.CloseOrder(ctx, "") }},
		{"QueryOrder 空订单号", func() error { _, err := c.QueryOrder(ctx, ""); return err }},
		{"QueryOrderByTransactionID 空单号", func() error { _, err := c.QueryOrderByTransactionID(ctx, ""); return err }},
		{"RefundQuery 空退款单号", func() error { _, err := c.RefundQuery(ctx, ""); return err }},
		{"ApplyTradeBill 空请求", func() error { _, err := c.ApplyTradeBill(ctx, nil); return err }},
		{"ApplyTradeBill 空日期", func() error { _, err := c.ApplyTradeBill(ctx, &BillRequest{}); return err }},
		{"ApplyFundFlowBill 空日期", func() error { _, err := c.ApplyFundFlowBill(ctx, &BillRequest{}); return err }},
		{"DownloadBill 空地址", func() error { _, err := c.DownloadBill(ctx, "", ""); return err }},
		{"ProfitSharing 空请求", func() error { _, err := c.ProfitSharing(ctx, nil); return err }},
		{"ProfitSharing 空接收方", func() error {
			_, err := c.ProfitSharing(ctx, &ProfitSharingRequest{TransactionID: "t", OutOrderNo: "o"})
			return err
		}},
		{"ProfitSharingQuery 空参数", func() error { _, err := c.ProfitSharingQuery(ctx, "", ""); return err }},
		{"ProfitSharingReturn 空单号", func() error {
			_, err := c.ProfitSharingReturn(ctx, &ProfitSharingReturnRequest{OutReturnNo: "r", ReturnMchID: "m"})
			return err
		}},
		{"ProfitSharingReturnQuery 空参数", func() error { _, err := c.ProfitSharingReturnQuery(ctx, "", ""); return err }},
		{"ProfitSharingUnfreeze 空参数", func() error { _, err := c.ProfitSharingUnfreeze(ctx, nil); return err }},
		{"ProfitSharingAmounts 空单号", func() error { _, err := c.ProfitSharingAmounts(ctx, ""); return err }},
	}

	for _, tc := range cases {
		if err := tc.fn(); err == nil {
			t.Errorf("%s 应返回错误", tc.name)
		}
	}
}

// ---------- 证书解密 ----------

func TestDecryptCertificate(t *testing.T) {
	apiV3Key := strings.Repeat("k", 32)
	certPEM := "-----BEGIN CERTIFICATE-----\nTEST\n-----END CERTIFICATE-----\n"

	block, err := newAESEncryptHelper(t, []byte(apiV3Key), []byte(certPEM))
	if err != nil {
		t.Fatalf("构造密文失败: %v", err)
	}

	plain, err := decryptCertificate(apiV3Key, &EncryptCertificate{
		Algorithm:      "AEAD_AES_256_GCM",
		Nonce:          block.nonce,
		AssociatedData: block.aad,
		Ciphertext:     block.ciphertext,
	})
	if err != nil {
		t.Fatalf("解密证书失败: %v", err)
	}
	if plain != certPEM {
		t.Fatalf("解密结果不匹配: %s", plain)
	}

	if _, err := decryptCertificate("short", &EncryptCertificate{Algorithm: "AEAD_AES_256_GCM"}); err == nil {
		t.Fatal("APIV3Key 长度不足应报错")
	}
	if _, err := decryptCertificate(apiV3Key, &EncryptCertificate{Algorithm: "UNKNOWN"}); err == nil {
		t.Fatal("不支持的算法应报错")
	}
}

// aesEncryptResult AES-GCM 加密结果，用于证书解密测试
type aesEncryptResult struct {
	nonce      string
	aad        string
	ciphertext string
}

// newAESEncryptHelper 使用 API V3 密钥加密明文，模拟微信支付下发的证书密文
func newAESEncryptHelper(t *testing.T, key, plaintext []byte) (aesEncryptResult, error) {
	t.Helper()
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return aesEncryptResult{}, err
	}
	block, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return aesEncryptResult{}, err
	}
	nonce := make([]byte, block.NonceSize())
	if _, err := crand.Read(nonce); err != nil {
		return aesEncryptResult{}, err
	}
	aad := "certificate"
	ciphertext := block.Seal(nil, nonce, plaintext, []byte(aad))
	return aesEncryptResult{
		nonce:      string(nonce),
		aad:        aad,
		ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}, nil
}

// ---------- URL 构造 ----------

func TestBuildBillURL(t *testing.T) {
	got := buildBillURL(BillTradeURL, &BillRequest{BillDate: "2026-10-14", BillType: "SUCCESS"})
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("URL 解析失败: %v", err)
	}
	if parsed.Path != "/v3/bill/tradebill" {
		t.Fatalf("路径错误: %s", parsed.Path)
	}
	if parsed.Query().Get("bill_date") != "2026-10-14" || parsed.Query().Get("bill_type") != "SUCCESS" {
		t.Fatalf("查询参数错误: %v", parsed.Query())
	}
	if strings.Contains(got, "tar_type") {
		t.Fatal("未设置的可选参数不应出现在 URL 中")
	}
}
