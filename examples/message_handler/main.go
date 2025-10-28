package main

import (
	"fmt"
	"log"

	"github.com/jcbowen/wego"
	"github.com/jcbowen/wego/message"
)

// TextMessageHandler 文本消息处理器
type TextMessageHandler struct{}

// HandleMessage 处理文本消息
func (h *TextMessageHandler) HandleMessage(msg *message.Message) (interface{}, error) {
	fmt.Printf("收到文本消息，发送方: %s, 接收方: %s\n", msg.FromUserName, msg.ToUserName)
	
	// 返回回复消息
	reply := &message.TextMessage{
		Message: message.Message{
			ToUserName:   msg.FromUserName,
			FromUserName: msg.ToUserName,
			CreateTime:   msg.CreateTime,
			MsgType:      message.MessageTypeText,
		},
		Content: "收到您的消息: " + "[文本内容]",
	}
	
	return reply, nil
}

// EventMessageHandler 事件消息处理器
type EventMessageHandler struct{}

// HandleEvent 处理事件消息
func (h *EventMessageHandler) HandleEvent(event *message.EventMessage) (interface{}, error) {
	fmt.Printf("收到事件消息，事件类型: %s\n", event.Event)
	
	switch event.Event {
	case message.EventTypeComponentVerifyTicket:
		fmt.Println("处理component_verify_ticket事件")
	case message.EventTypeAuthorized:
		fmt.Println("处理授权成功事件")
	case message.EventTypeUnauthorized:
		fmt.Println("处理取消授权事件")
	case message.EventTypeUpdateAuthorized:
		fmt.Println("处理授权更新事件")
	default:
		fmt.Printf("未知事件类型: %s\n", event.Event)
	}
	
	return nil, nil
}

func main() {
	fmt.Println("=== WeGo 消息处理器示例 ===")

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
	messageClient := wegoClient.Message()

	// 创建消息处理器
	processor := message.NewMessageProcessor()

	// 注册消息处理器
	textHandler := &TextMessageHandler{}
	eventHandler := &EventMessageHandler{}

	processor.RegisterMessageHandler(message.MessageTypeText, textHandler)
	processor.RegisterEventHandler(message.EventTypeComponentVerifyTicket, eventHandler)
	processor.RegisterEventHandler(message.EventTypeAuthorized, eventHandler)
	processor.RegisterEventHandler(message.EventTypeUnauthorized, eventHandler)
	processor.RegisterEventHandler(message.EventTypeUpdateAuthorized, eventHandler)

	fmt.Println("消息处理器注册完成")

	// 模拟处理各种消息
	demoTextMessage(processor)
	demoComponentVerifyTicket(processor)
	demoAuthorizedEvent(processor)
	demoUnauthorizedEvent(processor)
	demoUpdateAuthorizedEvent(processor)

	fmt.Println("=== 消息处理器示例执行完成 ===")
}

// demoTextMessage 演示处理文本消息
func demoTextMessage(processor *message.MessageProcessor) {
	fmt.Println("\n--- 处理文本消息 ---")

	textMessageXML := `<?xml version="1.0" encoding="UTF-8"?>
<xml>
  <ToUserName><![CDATA[toUser]]></ToUserName>
  <FromUserName><![CDATA[fromUser]]></FromUserName>
  <CreateTime>1348831860</CreateTime>
  <MsgType><![CDATA[text]]></MsgType>
  <Content><![CDATA[这是一条测试消息]]></Content>
  <MsgId>1234567890123456</MsgId>
</xml>`

	result, err := processor.ProcessMessage([]byte(textMessageXML))
	if err != nil {
		log.Printf("处理文本消息失败: %v", err)
	} else {
		fmt.Printf("文本消息处理结果: %+v\n", result)
	}
}

// demoComponentVerifyTicket 演示处理component_verify_ticket事件
func demoComponentVerifyTicket(processor *message.MessageProcessor) {
	fmt.Println("\n--- 处理component_verify_ticket事件 ---")

	componentVerifyTicketXML := `<?xml version="1.0" encoding="UTF-8"?>
<xml>
  <AppId><![CDATA[your_component_app_id]]></AppId>
  <CreateTime>1413192605</CreateTime>
  <InfoType><![CDATA[component_verify_ticket]]></InfoType>
  <ComponentVerifyTicket><![CDATA[ticket_value]]></ComponentVerifyTicket>
</xml>`

	result, err := processor.ProcessMessage([]byte(componentVerifyTicketXML))
	if err != nil {
		log.Printf("处理component_verify_ticket事件失败: %v", err)
	} else {
		fmt.Printf("component_verify_ticket事件处理结果: %+v\n", result)
	}
}

// demoAuthorizedEvent 演示处理授权成功事件
func demoAuthorizedEvent(processor *message.MessageProcessor) {
	fmt.Println("\n--- 处理授权成功事件 ---")

	authorizedEventXML := `<?xml version="1.0" encoding="UTF-8"?>
<xml>
  <AppId><![CDATA[your_component_app_id]]></AppId>
  <CreateTime>1413192605</CreateTime>
  <InfoType><![CDATA[authorized]]></InfoType>
  <AuthorizerAppid><![CDATA[authorizer_app_id]]></AuthorizerAppid>
  <AuthorizationCode><![CDATA[auth_code]]></AuthorizationCode>
  <AuthorizationCodeExpiredTime>1413192605</AuthorizationCodeExpiredTime>
  <PreAuthCode><![CDATA[pre_auth_code]]></PreAuthCode>
</xml>`

	result, err := processor.ProcessMessage([]byte(authorizedEventXML))
	if err != nil {
		log.Printf("处理授权成功事件失败: %v", err)
	} else {
		fmt.Printf("授权成功事件处理结果: %+v\n", result)
	}
}

// demoUnauthorizedEvent 演示处理取消授权事件
func demoUnauthorizedEvent(processor *message.MessageProcessor) {
	fmt.Println("\n--- 处理取消授权事件 ---")

	unauthorizedEventXML := `<?xml version="1.0" encoding="UTF-8"?>
<xml>
  <AppId><![CDATA[your_component_app_id]]></AppId>
  <CreateTime>1413192605</CreateTime>
  <InfoType><![CDATA[unauthorized]]></InfoType>
  <AuthorizerAppid><![CDATA[authorizer_app_id]]></AuthorizerAppid>
</xml>`

	result, err := processor.ProcessMessage([]byte(unauthorizedEventXML))
	if err != nil {
		log.Printf("处理取消授权事件失败: %v", err)
	} else {
		fmt.Printf("取消授权事件处理结果: %+v\n", result)
	}
}

// demoUpdateAuthorizedEvent 演示处理授权更新事件
func demoUpdateAuthorizedEvent(processor *message.MessageProcessor) {
	fmt.Println("\n--- 处理授权更新事件 ---")

	updateAuthorizedEventXML := `<?xml version="1.0" encoding="UTF-8"?>
<xml>
  <AppId><![CDATA[your_component_app_id]]></AppId>
  <CreateTime>1413192605</CreateTime>
  <InfoType><![CDATA[updateauthorized]]></InfoType>
  <AuthorizerAppid><![CDATA[authorizer_app_id]]></AuthorizerAppid>
  <AuthorizationCode><![CDATA[auth_code]]></AuthorizationCode>
  <AuthorizationCodeExpiredTime>1413192605</AuthorizationCodeExpiredTime>
  <PreAuthCode><![CDATA[pre_auth_code]]></PreAuthCode>
</xml>`

	result, err := processor.ProcessMessage([]byte(updateAuthorizedEventXML))
	if err != nil {
		log.Printf("处理授权更新事件失败: %v", err)
	} else {
		fmt.Printf("授权更新事件处理结果: %+v\n", result)
	}
}