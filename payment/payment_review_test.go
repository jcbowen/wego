package payment

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/logger"
	"github.com/jcbowen/wego/storage"
)

// ---------- 审查修复相关测试 ----------

// mockTokenStorage 仅实现平台证书相关方法的存储，其余方法由嵌入接口兜底
type mockTokenStorage struct {
	storage.TokenStorage

	getCertErr    error // GetWechatPayCert 返回的错误，用于模拟"不存在即报错"的自定义存储
	latestCert    *storage.WechatPayCertificate
	saved         []*storage.WechatPayCertificate
	downloadCalls int
}

func (m *mockTokenStorage) GetWechatPayCert(_ context.Context, _, _ string) (*storage.WechatPayCertificate, error) {
	return nil, m.getCertErr
}

func (m *mockTokenStorage) GetLatestWechatPayCert(_ context.Context, _ string) (*storage.WechatPayCertificate, error) {
	return m.latestCert, nil
}

func (m *mockTokenStorage) SaveWechatPayCert(_ context.Context, _ string, cert *storage.WechatPayCertificate) error {
	m.saved = append(m.saved, cert)
	return nil
}

// newCertModeClient 构造仅使用平台证书模式（未配置公钥）的测试客户端
func newCertModeClient(t *testing.T, mock core.HTTPClient, st storage.TokenStorage) (*Client, testKeyMaterial) {
	t.Helper()
	km := genKeyMaterial(t)
	cfg := &Config{
		AppID:        "wx1234567890",
		MchID:        "1900000001",
		APIV3Key:     strings.Repeat("a", 32),
		CertSerialNo: "MERCHANT_CERT_SERIAL",
		PrivateKey:   km.privPEM,
	}
	c := &Client{
		config:     cfg,
		httpClient: mock,
		logger:     logger.NewDefaultLoggerInterface(),
		privateKey: km.priv,
		storage:    st,
		certs:      map[string]*x509.Certificate{},
	}
	c.req = core.NewRequest(c.httpClient, c.logger)
	return c, km
}

// seedPlatformCert 向客户端内存缓存注入一张测试平台证书，避免测试触发证书下载
//
// 返回:
//   - string: 注入证书的序列号，可作为响应头 Wechatpay-Serial
func seedPlatformCert(t *testing.T, c *Client, km testKeyMaterial) string {
	t.Helper()
	certPEM, serialNo := selfSignedCertPEM(t, km)
	cert, err := parseCertificate(certPEM)
	if err != nil {
		t.Fatalf("解析注入证书失败: %v", err)
	}
	c.certMu.Lock()
	c.certs[serialNo] = cert
	c.certMu.Unlock()
	return serialNo
}

// ---------- 1. 签名 URL 路径转义 ----------

// TestExtractSignURLPathEscaped 订单号含特殊字符时，签名用的路径必须是未解码的原始转义路径
func TestExtractSignURLPathEscaped(t *testing.T) {
	cases := []struct {
		name   string
		rawURL string
		want   string
	}{
		{
			name:   "普通订单号",
			rawURL: "https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/ORDER_001?mchid=1900000001",
			want:   "/v3/pay/transactions/out-trade-no/ORDER_001?mchid=1900000001",
		},
		{
			name:   "含斜杠的订单号",
			rawURL: "https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/ORDER%2F1?mchid=1900000001",
			want:   "/v3/pay/transactions/out-trade-no/ORDER%2F1?mchid=1900000001",
		},
		{
			name:   "含中文的订单号",
			rawURL: "https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/" + url.PathEscape("订单1") + "?mchid=1900000001",
			want:   "/v3/pay/transactions/out-trade-no/" + url.PathEscape("订单1") + "?mchid=1900000001",
		},
	}

	for _, tc := range cases {
		got, err := extractSignURLPath(tc.rawURL)
		if err != nil {
			t.Fatalf("%s: 解析失败: %v", tc.name, err)
		}
		if got != tc.want {
			t.Errorf("%s: 签名路径不符\n期望: %s\n实际: %s", tc.name, tc.want, got)
		}
	}

	// 回归保护：使用 url.URL.Path 会得到解码结果，与微信端计算不一致
	u, _ := url.Parse("https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/ORDER%2F1")
	if u.Path == u.EscapedPath() {
		t.Fatalf("测试前置条件不成立：Path 与 EscapedPath 相同，无法验证该修复")
	}
}

// TestQueryOrderSignURLMatchesRequestURL 实际发出的请求 URL 与签名内容中的 URL 必须一致
func TestQueryOrderSignURLMatchesRequestURL(t *testing.T) {
	mock := &mockHTTPClient{
		statusCode: 200,
		body:       []byte(`{"trade_state":"SUCCESS"}`),
		headers:    http.Header{},
	}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", string(mock.body))

	// 订单号含斜杠，会被 PathEscape 转义
	if _, err := c.QueryOrder(context.Background(), "ORDER/2026/001"); err != nil {
		t.Fatalf("查询订单失败: %v", err)
	}

	requestURL := mock.lastReq.URL.String()
	wantPath := "/v3/pay/transactions/out-trade-no/ORDER%2F2026%2F001?mchid=1900000001"
	if !strings.HasSuffix(requestURL, wantPath) {
		t.Fatalf("请求 URL 路径不符，实际: %s", requestURL)
	}

	// 从 Authorization 无法直接读 URL，改为直接核对签名路径函数对请求 URL 的输出
	signPath, _ := extractSignURLPath(requestURL)
	if signPath != wantPath {
		t.Errorf("签名路径应为原始转义路径\n期望: %s\n实际: %s", wantPath, signPath)
	}
	parsed, _ := url.Parse(requestURL)
	if signPath == parsed.Path+"?"+parsed.RawQuery {
		t.Errorf("签名路径不应使用解码后的 Path")
	}
}

// ---------- 2. 结构化错误响应 ----------

// TestParseWechatAPIError 解析微信支付错误响应体
func TestParseWechatAPIError(t *testing.T) {
	body := []byte(`{"code":"ORDER_PAID","message":"订单已支付","detail":{"issue":"该订单已支付，不能关闭"}}`)
	err := parseWechatAPIError(400, nil, body)
	if err == nil {
		t.Fatalf("应解析出 APIError")
	}
	apiErr, ok := AsAPIError(err)
	if !ok {
		t.Fatalf("应能通过 AsAPIError 提取")
	}
	if apiErr.StatusCode != 400 || apiErr.Code != "ORDER_PAID" {
		t.Errorf("错误码解析错误: %+v", apiErr)
	}
	if apiErr.Detail == nil || apiErr.Detail.Issue == "" {
		t.Errorf("错误明细未解析: %+v", apiErr)
	}
	if !strings.Contains(apiErr.Error(), "ORDER_PAID") {
		t.Errorf("错误信息应包含错误码，实际: %s", apiErr.Error())
	}
}

// TestParseWechatAPIErrorNonWechatFormat 非微信错误格式应返回 nil，交由通用错误处理
func TestParseWechatAPIErrorNonWechatFormat(t *testing.T) {
	cases := [][]byte{
		[]byte(`<html><body>502 Bad Gateway</body></html>`),
		[]byte(``),
		[]byte(`{"foo":"bar"}`),
	}
	for _, body := range cases {
		if err := parseWechatAPIError(502, nil, body); err != nil {
			t.Errorf("非微信格式不应解析为 APIError，body=%s", string(body))
		}
	}
}

// TestDoRequestReturnsAPIError 请求失败时应返回可提取的 *APIError
func TestDoRequestReturnsAPIError(t *testing.T) {
	mock := &mockHTTPClient{
		statusCode: 403,
		body:       []byte(`{"code":"NO_AUTH","message":"商户无权限"}`),
		headers:    http.Header{},
	}
	c, _ := newTestClient(t, mock)

	err := c.CloseOrder(context.Background(), "ORDER_001")
	if err == nil {
		t.Fatalf("应返回错误")
	}
	if !IsAPIError(err) {
		t.Fatalf("应识别为微信支付业务错误，实际: %v", err)
	}
	apiErr, _ := AsAPIError(err)
	if apiErr.Code != "NO_AUTH" {
		t.Errorf("错误码不符: %s", apiErr.Code)
	}
	// 支持 errors.Is 按错误码比较
	if !errors.Is(err, &APIError{Code: "NO_AUTH"}) {
		t.Errorf("errors.Is 按错误码比较失败")
	}
	if errors.Is(err, &APIError{Code: "OTHER"}) {
		t.Errorf("errors.Is 不应匹配不同错误码")
	}
}

// ---------- 3. 参数校验 ----------

// TestOrderAmountValidation 下单金额与必填项校验
func TestOrderAmountValidation(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 200, body: []byte(`{"prepay_id":"wx1"}`), headers: http.Header{}}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", string(mock.body))

	if _, err := c.JSAPIPay(context.Background(), &JSAPIRequest{
		OutTradeNo: "O1", Description: "商品", Amount: Amount{Total: 0},
		Payer: Payer{OpenID: "o1"},
	}); err == nil {
		t.Errorf("金额 0 应报错")
	}
	if _, err := c.JSAPIPay(context.Background(), &JSAPIRequest{
		Description: "商品", Amount: Amount{Total: 100}, Payer: Payer{OpenID: "o1"},
	}); err == nil {
		t.Errorf("缺少商户订单号应报错")
	}
	if _, err := c.JSAPIPay(context.Background(), &JSAPIRequest{
		OutTradeNo: "O1", Description: "商品", Amount: Amount{Total: 100},
	}); err == nil {
		t.Errorf("缺少 openid 应报错")
	}
	if _, err := c.NativePay(context.Background(), &NativeRequest{
		OutTradeNo: "O1", Description: "商品", Amount: Amount{Total: -1},
	}); err == nil {
		t.Errorf("负数金额应报错")
	}
	if _, err := c.AppPay(context.Background(), &AppRequest{
		OutTradeNo: "O1", Amount: Amount{Total: 100}, Description: "商品",
	}); err != nil {
		t.Errorf("合法参数不应报错: %v", err)
	}
}

// TestRefundAmountValidation 退款金额校验
func TestRefundAmountValidation(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 200, body: []byte(`{"status":"SUCCESS"}`), headers: http.Header{}}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", string(mock.body))

	base := RefundRequest{OutTradeNo: "O1", OutRefundNo: "R1"}

	req := base
	req.Amount = RefundAmount{Refund: 0, Total: 100}
	if _, err := c.Refund(context.Background(), &req); err == nil {
		t.Errorf("退款金额 0 应报错")
	}

	req = base
	req.Amount = RefundAmount{Refund: 200, Total: 100}
	if _, err := c.Refund(context.Background(), &req); err == nil {
		t.Errorf("退款金额大于订单金额应报错")
	}

	req = base
	req.Amount = RefundAmount{Refund: 50, Total: 100}
	if _, err := c.Refund(context.Background(), &req); err != nil {
		t.Errorf("合法退款参数不应报错: %v", err)
	}
}

// TestProfitSharingReceiverValidation 分账接收方校验
func TestProfitSharingReceiverValidation(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 200, body: []byte(`{"state":"PROCESSING"}`), headers: http.Header{}}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", string(mock.body))

	_, err := c.ProfitSharing(context.Background(), &ProfitSharingRequest{
		TransactionID: "4200001",
		OutOrderNo:    "PS1",
		Receivers:     []ProfitSharingReceiver{{Type: "MERCHANT_ID", Account: "1900000109", Amount: 0}},
	})
	if err == nil {
		t.Errorf("分账金额 0 应报错")
	}

	_, err = c.ProfitSharing(context.Background(), &ProfitSharingRequest{
		TransactionID: "4200001",
		OutOrderNo:    "PS1",
		Receivers:     []ProfitSharingReceiver{{Type: "", Account: "1900000109", Amount: 100}},
	})
	if err == nil {
		t.Errorf("缺少接收方类型应报错")
	}
}

// ---------- 4. 分账 Wechatpay-Serial 请求头 ----------

// TestProfitSharingSerialHeaderOnlyWhenEncrypted 无加密姓名时不强制获取加密密钥，也不声明序列号
func TestProfitSharingSerialHeaderOnlyWhenEncrypted(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 200, body: []byte(`{"state":"PROCESSING"}`), headers: http.Header{}}
	// 仅平台证书模式且存储不可用：修复前会因强制获取加密序列号而失败
	st := &mockTokenStorage{getCertErr: errors.New("cert not found")}
	c, km := newCertModeClient(t, mock, st)
	serial := seedPlatformCert(t, c, km)
	mock.headers = signedHeaders(t, km.priv, serial, string(mock.body))

	_, err := c.ProfitSharing(context.Background(), &ProfitSharingRequest{
		TransactionID: "4200001",
		OutOrderNo:    "PS1",
		Receivers:     []ProfitSharingReceiver{{Type: "MERCHANT_ID", Account: "1900000109", Amount: 100, Description: "分账"}},
	})
	if err != nil {
		t.Fatalf("无加密字段时不应因加密密钥失败而阻断: %v", err)
	}
	if got := mock.lastReq.Header.Get(PublicKeyHeader); got != "" {
		t.Errorf("无加密字段时不应声明 Wechatpay-Serial，实际: %s", got)
	}
}

// TestProfitSharingSerialHeaderWithEncryptedName 含加密姓名时声明实际加密所用的公钥 ID
func TestProfitSharingSerialHeaderWithEncryptedName(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 200, body: []byte(`{"state":"PROCESSING"}`), headers: http.Header{}}
	c, km := newTestClient(t, mock)
	mock.headers = signedHeaders(t, km.priv, "PUB_KEY_ID_TEST", string(mock.body))

	_, err := c.ProfitSharing(context.Background(), &ProfitSharingRequest{
		TransactionID: "4200001",
		OutOrderNo:    "PS1",
		Receivers: []ProfitSharingReceiver{
			{Type: "PERSONAL_OPENID", Account: "oABC", Name: "ENCRYPTED_NAME", Amount: 100, Description: "分账"},
		},
	})
	if err != nil {
		t.Fatalf("分账失败: %v", err)
	}
	if got := mock.lastReq.Header.Get(PublicKeyHeader); got != "PUB_KEY_ID_TEST" {
		t.Errorf("应声明加密所用公钥 ID，实际: %s", got)
	}
}

// ---------- 5. 证书获取健壮性 ----------

// TestGetPlatformCertFallsBackOnStorageError 存储读取报错时应回退到重新下载
func TestGetPlatformCertFallsBackOnStorageError(t *testing.T) {
	km := genKeyMaterial(t)
	certPEM, serialNo := selfSignedCertPEM(t, km)
	ciphertext := encryptCertForTest(t, strings.Repeat("a", 32), certPEM)

	mock := &mockHTTPClient{
		statusCode: 200,
		body:       []byte(`{"data":[{"serial_no":"` + serialNo + `","encrypt_certificate":{"algorithm":"AEAD_AES_256_GCM","nonce":"0123456789ab","associated_data":"certificate","ciphertext":"` + ciphertext + `"}}]}`),
		headers:    http.Header{},
	}
	st := &mockTokenStorage{getCertErr: errors.New("storage unavailable")}
	c, _ := newCertModeClient(t, mock, st)

	cert, err := c.getPlatformCert(context.Background(), serialNo)
	if err != nil {
		t.Fatalf("存储报错时应回退下载，实际失败: %v", err)
	}
	if cert == nil {
		t.Fatalf("应返回证书")
	}
}

// TestDownloadPlatformCertsThrottle 下载失败后短时间内不重复请求，避免被微信限流
func TestDownloadPlatformCertsThrottle(t *testing.T) {
	mock := &mockHTTPClient{
		statusCode: 500,
		body:       []byte(`{"code":"SYSTEM_ERROR","message":"系统错误"}`),
		headers:    http.Header{},
	}
	c, _ := newCertModeClient(t, mock, &mockTokenStorage{})

	if err := c.downloadPlatformCerts(context.Background()); err == nil {
		t.Fatalf("首次下载应失败")
	}
	// 记录首次失败后的请求计数，再次调用不应发起新的网络请求
	mock.lastReq = nil
	if err := c.downloadPlatformCerts(context.Background()); err == nil {
		t.Fatalf("节流期内应直接返回失败")
	}
	if mock.lastReq != nil {
		t.Errorf("节流期内不应重复发起下载请求")
	}
}

// TestDownloadPlatformCertsEmptyData 接口返回空证书列表时应报错而非静默成功
func TestDownloadPlatformCertsEmptyData(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 200, body: []byte(`{"data":[]}`), headers: http.Header{}}
	c, _ := newCertModeClient(t, mock, &mockTokenStorage{})

	err := c.downloadPlatformCerts(context.Background())
	if err == nil {
		t.Fatalf("空证书列表应报错")
	}
	if !errors.Is(err, ErrCertDownloadFailed) {
		t.Errorf("应返回 ErrCertDownloadFailed，实际: %v", err)
	}
}

// ---------- 6. 构造函数返回错误 ----------

// TestNewClientWithStorageEReturnsError 非法配置应返回错误而非 panic
func TestNewClientWithStorageEReturnsError(t *testing.T) {
	km := genKeyMaterial(t)
	st := &mockTokenStorage{}

	cases := []struct {
		name string
		cfg  *Config
	}{
		{"配置为空", nil},
		{"缺少AppID", &Config{MchID: "1", APIV3Key: strings.Repeat("a", 32), CertSerialNo: "S", PrivateKey: km.privPEM}},
		{"APIV3Key长度错误", &Config{AppID: "wx1", MchID: "1", APIV3Key: "short", CertSerialNo: "S", PrivateKey: km.privPEM}},
		{"私钥非法", &Config{AppID: "wx1", MchID: "1", APIV3Key: strings.Repeat("a", 32), CertSerialNo: "S", PrivateKey: "not-a-pem"}},
		{"公钥只配一半", &Config{AppID: "wx1", MchID: "1", APIV3Key: strings.Repeat("a", 32), CertSerialNo: "S", PrivateKey: km.privPEM, PublicKeyID: "PUB_KEY_ID_X"}},
	}

	for _, tc := range cases {
		if _, err := NewClientWithStorageE(tc.cfg, st); err == nil {
			t.Errorf("%s: 应返回错误", tc.name)
		}
	}

	if _, err := NewClientWithStorageE(&Config{
		AppID: "wx1", MchID: "1", APIV3Key: strings.Repeat("a", 32),
		CertSerialNo: "S", PrivateKey: km.privPEM,
	}, st); err != nil {
		t.Errorf("合法配置不应报错: %v", err)
	}

	if _, err := NewClientWithStorageE(&Config{
		AppID: "wx1", MchID: "1", APIV3Key: strings.Repeat("a", 32),
		CertSerialNo: "S", PrivateKey: km.privPEM,
	}, nil); err == nil {
		t.Errorf("存储为空应报错")
	}
}

// ---------- 7. 账单下载 Accept 头 ----------

// TestDownloadBillUsesWildcardAccept 账单等非 JSON 响应不应声明 application/json
func TestDownloadBillUsesWildcardAccept(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 200, body: []byte("账单内容"), headers: http.Header{}}
	c, _ := newTestClient(t, mock)

	if _, err := c.DownloadBill(context.Background(), "https://api.mch.weixin.qq.com/v3/billdownload/file?token=x", ""); err != nil {
		t.Fatalf("下载账单失败: %v", err)
	}
	if got := mock.lastReq.Header.Get("Accept"); got == "application/json" {
		t.Errorf("账单下载不应声明 application/json")
	}
	if string(mock.lastBody) != "" {
		t.Errorf("GET 请求不应带请求体")
	}
}

// ---------- 测试辅助 ----------

// selfSignedCertPEM 生成一张测试用自签名证书，返回 PEM 与序列号（十六进制）
func selfSignedCertPEM(t *testing.T, km testKeyMaterial) (string, string) {
	t.Helper()
	tpl := &x509.Certificate{
		SerialNumber: big.NewInt(20261015001),
		Subject:      pkix.Name{CommonName: "Tenpay.com Test Cert"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(crand.Reader, tpl, tpl, &km.priv.PublicKey, km.priv)
	if err != nil {
		t.Fatalf("生成测试证书失败: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("解析测试证书失败: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return string(pemBytes), cert.SerialNumber.Text(16)
}

// encryptCertForTest 使用 API V3 密钥按微信 AEAD_AES_256_GCM 方式加密证书内容
func encryptCertForTest(t *testing.T, apiV3Key, certPEM string) string {
	t.Helper()
	block, err := aes.NewCipher([]byte(apiV3Key))
	if err != nil {
		t.Fatalf("创建 AES 失败: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("创建 GCM 失败: %v", err)
	}
	nonce := []byte("0123456789ab")
	aad := []byte("certificate")
	sealed := gcm.Seal(nil, nonce, []byte(certPEM), aad)
	return base64.StdEncoding.EncodeToString(sealed)
}
