package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/jcbowen/wego"
	"github.com/jcbowen/wego/crypto"
	"github.com/jcbowen/wego/message"
	"github.com/jcbowen/wego/storage"
)

func main() {
	fmt.Println("=== WeGo 微信开放平台示例 ===")

	// 配置微信开放平台参数
	config := &wego.WeGoConfig{
		ComponentAppID:     "your_component_app_id",
		ComponentAppSecret: "your_component_app_secret",
		ComponentToken:     "your_component_token",
		EncodingAESKey:     "your_encoding_aes_key",
		RedirectURI:        "https://yourdomain.com/callback",
	}

	// 创建WeGo实例
	wegoClient := wego.NewWeGo(config)

	// 使用各个功能模块
	demoAPI(wegoClient)
	demoAuth(wegoClient)
	demoMessage(wegoClient)
	demoStorage(wegoClient)
	demoCrypto(wegoClient)

	fmt.Println("=== 示例执行完成 ===")
}

// demoAPI 演示API功能
func demoAPI(wegoClient *wego.WeGo) {
	fmt.Println("\n--- API功能演示 ---")

	apiClient := wegoClient.API()
	
	// 获取组件访问令牌（需要verifyTicket参数）
	ctx := context.Background()
	token, err := apiClient.GetComponentAccessToken(ctx, "verify_ticket_here")
	if err != nil {
		log.Printf("获取组件访问令牌失败: %v", err)
	} else {
		fmt.Printf("组件访问令牌: %s (有效期: %d秒)\n", token.AccessToken, token.ExpiresIn)
	}

	// 获取预授权码
	preAuthCode, err := apiClient.GetPreAuthCode(ctx)
	if err != nil {
		log.Printf("获取预授权码失败: %v", err)
	} else {
		fmt.Printf("预授权码: %s (有效期: %d秒)\n", preAuthCode.PreAuthCode, preAuthCode.ExpiresIn)
	}
}

// demoAuth 演示授权功能
func demoAuth(wegoClient *wego.WeGo) {
	fmt.Println("\n--- 授权功能演示 ---")

	authClient := wegoClient.Auth()

	// 生成授权URL
	preAuthCodeStr := "your_pre_auth_code"
	redirectURI := "https://yourdomain.com/callback"
	authURL := fmt.Sprintf("https://mp.weixin.qq.com/cgi-bin/componentloginpage?component_appid=%s&pre_auth_code=%s&redirect_uri=%s",
		wegoClient.Config.ComponentAppID, preAuthCodeStr, url.QueryEscape(redirectURI))
	fmt.Printf("授权URL: %s\n", authURL)

	// 创建授权方客户端（模拟）
	authorizerClient := authClient.NewAuthorizerClient("authorizer_app_id")
	fmt.Printf("授权方客户端创建成功: %+v\n", authorizerClient)
}

// demoMessage 演示消息处理功能
func demoMessage(wegoClient *wego.WeGo) {
	fmt.Println("\n--- 消息处理功能演示 ---")

	messageClient := wegoClient.Message()

	// 创建消息处理器
	processor := message.NewMessageProcessor()
	fmt.Printf("消息处理器创建成功: %+v\n", processor)

	// 模拟处理component_verify_ticket事件
	componentVerifyTicketXML := `<?xml version="1.0" encoding="UTF-8"?>
<xml>
  <AppId><![CDATA[your_component_app_id]]></AppId>
  <CreateTime>1413192605</CreateTime>
  <InfoType><![CDATA[component_verify_ticket]]></InfoType>
  <ComponentVerifyTicket><![CDATA[ticket_value]]></ComponentVerifyTicket>
</xml>`

	result, err := processor.ProcessMessage([]byte(componentVerifyTicketXML))
	if err != nil {
		log.Printf("处理component_verify_ticket失败: %v", err)
	} else {
		fmt.Printf("component_verify_ticket处理结果: %+v\n", result)
	}
}

// demoStorage 演示存储功能
func demoStorage(wegoClient *wego.WeGo) {
	fmt.Println("\n--- 存储功能演示 ---")

	storageClient := wegoClient.Storage()

	// 创建内存存储实例
	memoryStorage := storage.NewMemoryStorage()
	fmt.Printf("内存存储实例创建成功: %+v\n", memoryStorage)

	// 演示保存组件令牌
	ctx := context.Background()
	componentToken := &storage.ComponentAccessToken{
		AccessToken: "test_component_access_token",
		ExpiresIn:   7200,
		ExpiresAt:   time.Now().Add(2 * time.Hour),
	}

	err := memoryStorage.SaveComponentToken(ctx, componentToken)
	if err != nil {
		log.Printf("保存组件令牌失败: %v", err)
	} else {
		fmt.Println("组件令牌保存成功")
	}

	// 演示读取组件令牌
	retrievedToken, err := memoryStorage.GetComponentToken(ctx)
	if err != nil {
		log.Printf("读取组件令牌失败: %v", err)
	} else if retrievedToken != nil {
		fmt.Printf("读取到的组件令牌: %s\n", retrievedToken.AccessToken)
	}

	// 演示文件存储
	fileStorage, err := storage.NewFileStorage("./storage_data")
	if err != nil {
		log.Printf("创建文件存储失败: %v", err)
	} else {
		fmt.Printf("文件存储实例创建成功: %+v\n", fileStorage)
		
		// 保存示例数据到文件存储
		preAuthCode := &storage.PreAuthCode{
			PreAuthCode: "test_pre_auth_code",
			ExpiresIn:   1800,
			ExpiresAt:   time.Now().Add(30 * time.Minute),
		}
		
		err = fileStorage.SavePreAuthCode(ctx, preAuthCode)
		if err != nil {
			log.Printf("保存预授权码到文件失败: %v", err)
		} else {
			fmt.Println("预授权码保存到文件成功")
		}
	}

	fmt.Printf("存储客户端: %+v\n", storageClient)
}

// demoCrypto 演示加密解密功能
func demoCrypto(wegoClient *wego.WeGo) {
	fmt.Println("\n--- 加密解密功能演示 ---")

	cryptoClient := wegoClient.Crypto()

	// 创建加密实例
	crypt := crypto.NewWXBizMsgCrypt(
		"your_token",
		"your_encoding_aes_key", 
		"your_app_id",
	)

	fmt.Printf("加密实例创建成功: %+v\n", crypt)

	// 演示加密
	text := "Hello, WeGo!"
	timestamp := "1413192605"
	nonce := "nonce"

	encrypted, err := crypt.EncryptMsg(text, timestamp, nonce)
	if err != nil {
		log.Printf("加密失败: %v", err)
	} else {
		fmt.Printf("加密结果: %s\n", encrypted)
	}

	// 演示解密（这里需要实际的加密数据才能演示）
	fmt.Println("解密功能需要实际的加密数据才能演示")
}