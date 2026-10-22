package payment

import (
	"context"
	"crypto"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/wego/core"
	"github.com/jcbowen/wego/logger"
	"github.com/jcbowen/wego/storage"
)

// Client 微信支付 V3 客户端
type Client struct {
	config     *Config
	httpClient core.HTTPClient
	storage    storage.TokenStorage
	logger     logger.LoggerInterface
	req        *core.Request
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	certs      map[string]*x509.Certificate
	certMu     sync.RWMutex
	// certRefreshMu 串行化平台证书下载，避免并发首次请求触发重复下载被微信限流
	certRefreshMu   sync.Mutex
	certRefreshFail time.Time
}

// NewClient 创建新的微信支付客户端（使用默认文件存储）
//
// 参数:
//   - config: 支付配置信息
//   - opts: 可选参数，支持 debugger.LoggerInterface 与 core.HTTPClient
//
// 返回:
//   - *Client: 支付客户端实例
//
// Deprecated: 配置或私钥非法时会直接 log.Panic 终止进程，依赖库不应越权决定宿主进程生死。
// 请改用 NewClientE 显式处理错误。
func NewClient(config *Config, opts ...any) *Client {
	client, err := NewClientE(config, opts...)
	if err != nil {
		log.Panic("创建微信支付客户端失败: " + err.Error())
	}
	return client
}

// NewClientE 创建新的微信支付客户端并返回错误（使用默认文件存储）
//
// 参数:
//   - config: 支付配置信息
//   - opts: 可选参数，支持 debugger.LoggerInterface 与 core.HTTPClient
//
// 返回:
//   - *Client: 支付客户端实例
//   - error: 配置校验、存储创建或密钥解析失败时返回错误
//
// 说明:
//   - 默认存储目录为 ./runtime/wego_storage，需确保进程对该目录有读写权限
//
// 使用示例:
//   - client, err := payment.NewClientE(&payment.Config{AppID: "wx123", ...})
func NewClientE(config *Config, opts ...any) (*Client, error) {
	fileStorage, err := storage.NewFileStorage("./runtime/wego_storage")
	if err != nil {
		return nil, fmt.Errorf("文件存储创建失败: %w", err)
	}
	return NewClientWithStorageE(config, fileStorage, opts...)
}

// NewClientWithStorage 创建新的微信支付客户端（使用自定义存储）
//
// 参数:
//   - config: 支付配置信息
//   - storage: 自定义存储实例
//   - opts: 可选参数，支持 debugger.LoggerInterface 与 core.HTTPClient
//
// 返回:
//   - *Client: 支付客户端实例
//
// Deprecated: 配置或私钥非法时会直接 log.Panic 终止进程，请改用 NewClientWithStorageE。
func NewClientWithStorage(config *Config, storage storage.TokenStorage, opts ...any) *Client {
	client, err := NewClientWithStorageE(config, storage, opts...)
	if err != nil {
		log.Panic("创建微信支付客户端失败: " + err.Error())
	}
	return client
}

// NewClientWithStorageE 创建新的微信支付客户端并返回错误（使用自定义存储）
//
// 参数:
//   - config: 支付配置信息
//   - storage: 自定义存储实例
//   - opts: 可选参数，支持 debugger.LoggerInterface 与 core.HTTPClient
//
// 返回:
//   - *Client: 支付客户端实例
//   - error: 配置校验或密钥解析失败时返回错误
//
// 使用示例:
//   - client, err := payment.NewClientWithStorageE(cfg, myStorage, myHTTPClient)
func NewClientWithStorageE(config *Config, storage storage.TokenStorage, opts ...any) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("支付配置不能为空")
	}
	if storage == nil {
		return nil, fmt.Errorf("存储实例不能为空")
	}
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("微信支付配置校验失败: %w", err)
	}

	privateKey, err := parsePrivateKey(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("微信支付私钥解析失败: %w", err)
	}

	// 公钥为可选项，未配置时保持平台证书模式
	var publicKey *rsa.PublicKey
	if config.UsePublicKey() {
		publicKey, err = parsePublicKey(config.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("微信支付公钥解析失败: %w", err)
		}
	}

	client := &Client{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		storage:    storage,
		logger:     logger.NewDefaultLoggerInterface(),
		privateKey: privateKey,
		publicKey:  publicKey,
		certs:      make(map[string]*x509.Certificate),
	}

	core.ApplyClientOptions(opts, func(msg string) { client.logger.Warn(msg) }, client.SetLogger, client.SetHTTPClient)

	client.req = core.NewRequest(client.httpClient, client.logger)
	return client, nil
}

// SetLogger 设置自定义日志器
//
// 参数:
//   - log: 自定义日志器，nil 则使用默认日志器
func (c *Client) SetLogger(log debugger.LoggerInterface) {
	if log == nil {
		c.logger = logger.NewDefaultLoggerInterface()
	} else {
		c.logger = logger.NewDebuggerLoggerAdapter(log)
	}
	if c.req != nil {
		c.req = core.NewRequest(c.httpClient, c.logger)
	}
}

// SetHTTPClient 设置自定义 HTTP 客户端
//
// 参数:
//   - client: 自定义 HTTP 客户端
func (c *Client) SetHTTPClient(client core.HTTPClient) {
	c.httpClient = client
	if c.req != nil {
		c.req = core.NewRequest(client, c.logger)
	}
}

// GetConfig 获取支付配置
//
// 返回:
//   - *Config: 支付配置副本指针
func (c *Client) GetConfig() *Config {
	return c.config
}

// GetLogger 获取日志器
//
// 返回:
//   - logger.LoggerInterface: 日志器实例
func (c *Client) GetLogger() logger.LoggerInterface {
	return c.logger
}

// parsePrivateKey 解析 RSA 私钥 PEM
//
// 参数:
//   - privateKeyPEM: PEM 格式私钥字符串
//
// 返回:
//   - *rsa.PrivateKey: 解析后的私钥
//   - error: 解析失败时返回错误
func parsePrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("无法解析私钥PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("私钥类型不是RSA")
	}

	return rsaKey, nil
}

// parsePublicKey 解析微信支付公钥 PEM
//
// 参数:
//   - publicKeyPEM: PEM 格式公钥字符串，支持 PKIX（BEGIN PUBLIC KEY）与 PKCS1（BEGIN RSA PUBLIC KEY）
//
// 返回:
//   - *rsa.PublicKey: 解析后的公钥
//   - error: 解析失败时返回错误
//
// 使用示例:
//   - pub, err := parsePublicKey(string(pemBytes))
func parsePublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("无法解析公钥PEM")
	}

	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("公钥类型不是RSA")
		}
		return rsaKey, nil
	}

	rsaKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %w", err)
	}
	return rsaKey, nil
}

// signSHA256WithRSA 使用商户私钥对消息做 SHA256withRSA 签名
//
// 参数:
//   - message: 待签名原文
//
// 返回:
//   - string: Base64 编码的签名值
//   - error: 签名失败时返回错误
//
// 使用示例:
//   - sign, err := client.signSHA256WithRSA("appId\ntimeStamp\n...")
func (c *Client) signSHA256WithRSA(message string) (string, error) {
	hash := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(crand.Reader, c.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// extractSignURLPath 从完整 URL 中提取用于微信 V3 签名的绝对路径（含查询参数）
//
// 参数:
//   - rawURL: 请求完整 URL
//
// 返回:
//   - string: 绝对路径，如 /v3/pay/transactions/jsapi
//   - error: URL 解析失败时返回错误
//
// 说明:
//   - 必须使用 EscapedPath() 而非 Path：Path 是解码后的结果，而微信服务端按原始转义路径计算签名
//   - 订单号、退款单号等路径参数若含 /、空格或中文，用 Path 会导致签名与服务端不一致而报 401
func extractSignURLPath(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	signURL := parsedURL.EscapedPath()
	if parsedURL.RawQuery != "" {
		signURL += "?" + parsedURL.RawQuery
	}
	return signURL, nil
}

// generateNonce 生成指定长度的十六进制随机字符串
//
// 参数:
//   - length: 字符串长度
//
// 返回:
//   - string: 随机字符串
//
// 说明:
//   - 使用 crypto/rand 生成，避免取模映射带来的分布偏差
func generateNonce(length int) string {
	if length <= 0 {
		return ""
	}
	// 每 2 个十六进制字符对应 1 字节随机数
	byteLen := (length + 1) / 2
	b := make([]byte, byteLen)
	if _, err := crand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b)[:length]
}

// authorization 构造微信支付 V3 Authorization 请求头
//
// 参数:
//   - method: HTTP 方法
//   - url: 请求完整 URL（含查询参数）
//   - body: 请求体字节数组
//
// 返回:
//   - string: Authorization 头内容
//   - error: 签名失败时返回错误
//
// 说明:
//   - 微信官方要求签名串中的 URL 为去除域名后的绝对路径（含查询参数）
func (c *Client) authorization(method, url string, body []byte) (string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce(32)

	signURL, err := extractSignURLPath(url)
	if err != nil {
		return "", fmt.Errorf("提取请求签名 URL 失败: %w", err)
	}

	message := method + "\n" + signURL + "\n" + timestamp + "\n" + nonce + "\n" + string(body) + "\n"
	hash := sha256.Sum256([]byte(message))

	signature, err := rsa.SignPKCS1v15(crand.Reader, c.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("请求签名失败: %w", err)
	}

	auth := fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%s",serial_no="%s"`,
		c.config.MchID,
		nonce,
		base64.StdEncoding.EncodeToString(signature),
		timestamp,
		c.config.CertSerialNo,
	)
	return auth, nil
}

// doRequest 发送带 V3 签名的 HTTP 请求
//
// 参数:
//   - ctx: 上下文
//   - method: HTTP 方法
//   - url: 请求完整 URL
//   - body: 请求体结构体
//   - result: 响应反序列化目标指针
//   - verifyResponse: 是否校验应答签名
//
// 返回:
//   - error: 请求或反序列化失败时返回错误
func (c *Client) doRequest(ctx context.Context, method, url string, body interface{}, result interface{}, verifyResponse bool) error {
	return c.doRequestWithSerial(ctx, method, url, body, result, verifyResponse, "")
}

// doRequestWithSerial 发送带 V3 签名并可指定 Wechatpay-Serial 的 HTTP 请求
//
// 参数:
//   - ctx: 上下文
//   - method: HTTP 方法
//   - url: 请求完整 URL
//   - body: 请求体结构体
//   - result: 响应反序列化目标指针
//   - verifyResponse: 是否校验应答签名
//   - serial: 强制指定的 Wechatpay-Serial，空字符串表示使用默认策略
//
// 返回:
//   - error: 请求或反序列化失败时返回错误
//
// 说明:
//   - 涉及敏感信息加密的接口（如分账）必须传入与实际加密密钥一致的序列号
func (c *Client) doRequestWithSerial(ctx context.Context, method, url string, body interface{}, result interface{}, verifyResponse bool, serial string) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("请求体序列化失败: %w", err)
		}
	}

	auth, err := c.authorization(method, url, bodyBytes)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"Authorization": auth,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	// 敏感信息加密场景以加密所用密钥为准，其余场景在公钥模式下声明公钥 ID
	if serial != "" {
		headers[PublicKeyHeader] = serial
	} else if c.config.UsePublicKey() {
		headers[PublicKeyHeader] = c.config.PublicKeyID
	}

	opt := &core.ReqMakeOpt{
		Method:      method,
		URL:         url,
		Body:        bodyBytes,
		Result:      result,
		Headers:     headers,
		ErrorParser: c.parseAPIError,
	}
	if verifyResponse {
		opt.VerifyResponse = c.verifyResponseSignature
	}

	return c.req.Make(ctx, opt)
}

// doRequestRaw 发送带 V3 签名的请求并返回原始响应体
//
// 参数:
//   - ctx: 上下文
//   - method: HTTP 方法
//   - url: 请求完整 URL
//   - verifyResponse: 是否校验应答签名
//
// 返回:
//   - []byte: 原始响应体
//   - error: 请求失败时返回错误
//
// 说明:
//   - 适用于非 JSON 响应，如账单文件流
func (c *Client) doRequestRaw(ctx context.Context, method, url string, verifyResponse bool) ([]byte, error) {
	auth, err := c.authorization(method, url, nil)
	if err != nil {
		return nil, err
	}

	// 账单等非 JSON 响应不能用 application/json，否则可能被网关拒绝
	headers := map[string]string{
		"Authorization": auth,
		"Accept":        "*/*",
	}
	if c.config.UsePublicKey() {
		headers[PublicKeyHeader] = c.config.PublicKeyID
	}

	var raw []byte
	opt := &core.ReqMakeOpt{
		Method:      method,
		URL:         url,
		RawResult:   &raw,
		Headers:     headers,
		ErrorParser: c.parseAPIError,
	}
	if verifyResponse {
		opt.VerifyResponse = c.verifyResponseSignature
	}

	if err := c.req.Make(ctx, opt); err != nil {
		return nil, err
	}

	return raw, nil
}

// parseAPIError 将非 2xx 响应解析为 *APIError
//
// 参数:
//   - statusCode: HTTP 状态码
//   - headers: HTTP 响应头
//   - body: 响应体
//
// 返回:
//   - error: 微信格式错误返回 *APIError，否则返回 nil
//
// 说明:
//   - 4xx/5xx 响应不参与验签（微信同样会签名，但本库只解析错误码，不做信任判断）
func (c *Client) parseAPIError(statusCode int, headers http.Header, body []byte) error {
	return parseWechatAPIError(statusCode, headers, body)
}

// verifyResponseSignature 验证微信支付 V3 响应签名
//
// 参数:
//   - ctx: 上下文
//   - statusCode: HTTP 状态码（签名串中不使用，仅用于日志）
//   - headers: HTTP 响应头
//   - body: 响应体字节数组
//
// 返回:
//   - error: 验证失败时返回错误
//
// 说明:
//   - 微信官方应答验签串格式为：应答时间戳\n应答随机串\n应答报文主体\n
//   - 建议校验时间戳与当前时间偏差不超过 5 分钟，防止重放攻击
//   - 验签密钥由响应头 Wechatpay-Serial 决定：公钥 ID 走公钥，否则走平台证书
func (c *Client) verifyResponseSignature(ctx context.Context, statusCode int, headers http.Header, body []byte) error {
	serial := headers.Get("Wechatpay-Serial")
	timestamp := headers.Get("Wechatpay-Timestamp")
	nonce := headers.Get("Wechatpay-Nonce")
	signature := headers.Get("Wechatpay-Signature")

	if serial == "" || timestamp == "" || nonce == "" || signature == "" {
		return fmt.Errorf("%w: Wechatpay-Serial/Timestamp/Nonce/Signature 存在空值", ErrSignatureHeaderMissing)
	}

	// 校验时间戳，允许 5 分钟偏差
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("响应时间戳格式错误: %w", err)
	}
	if diff := ts - time.Now().Unix(); diff > 300 || diff < -300 {
		return fmt.Errorf("响应时间戳偏差过大: %d 秒", diff)
	}

	// 构造验签名串：应答时间戳\n应答随机串\n应答报文主体\n
	message := timestamp + "\n" + nonce + "\n" + string(body) + "\n"
	hash := sha256.Sum256([]byte(message))

	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("响应签名Base64解码失败: %w", err)
	}

	pubKey, err := c.resolveVerifyPublicKey(ctx, serial)
	if err != nil {
		return err
	}

	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], sigBytes); err != nil {
		return fmt.Errorf("响应签名验证失败: %w", err)
	}

	return nil
}
