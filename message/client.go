package message

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/jcbowen/wego/crypto"
	"github.com/jcbowen/wego/openplatform"
)

// MessageClient 消息客户端
type MessageClient struct {
	Client *openplatform.APIClient
}

// NewMessageClient 创建新的消息客户端
func NewMessageClient(client *openplatform.APIClient) *MessageClient {
	return &MessageClient{
		Client: client,
	}
}

// SecureMessageProcessor 安全消息处理器（支持加解密）
type SecureMessageProcessor struct {
	processor   *MessageProcessor
	cryptoCache map[string]*crypto.WXBizMsgCrypt // 按授权方AppID缓存加解密实例
}

// NewSecureMessageProcessor 创建安全消息处理器
func NewSecureMessageProcessor() *SecureMessageProcessor {
	return &SecureMessageProcessor{
		processor:   NewMessageProcessor(),
		cryptoCache: make(map[string]*crypto.WXBizMsgCrypt),
	}
}

// ProcessSecureMessage 处理安全消息（包含加解密，符合微信官方规范）
func (p *SecureMessageProcessor) ProcessSecureMessage(
	authorizerAppID string,
	msgSignature string,
	timestamp string,
	nonce string,
	encryptedMsg string,
) (interface{}, error) {
	// 验证时间戳（防止重放攻击，微信官方建议5分钟时间窗口）
	if err := p.validateTimestamp(timestamp); err != nil {
		return nil, fmt.Errorf("时间戳验证失败: %v", err)
	}

	// 获取加解密实例
	cryptoInstance, err := p.getCryptoInstance(authorizerAppID)
	if err != nil {
		return nil, fmt.Errorf("获取加解密实例失败: %v", err)
	}

	// 验证消息签名（必须使用msg_signature参数）
	valid := cryptoInstance.VerifySignature(msgSignature, timestamp, nonce, encryptedMsg)
	if !valid {
		return nil, fmt.Errorf("消息签名验证不通过，请检查msg_signature参数")
	}

	// 解密消息
	decryptedMsg, err := cryptoInstance.DecryptMsg(msgSignature, timestamp, nonce, encryptedMsg)
	if err != nil {
		return nil, fmt.Errorf("消息解密失败: %v", err)
	}

	// 处理消息
	reply, err := p.processor.ProcessMessage([]byte(decryptedMsg))
	if err != nil {
		return nil, err
	}

	// 如果回复是"success"，直接返回（微信官方要求）
	if replyStr, ok := reply.(string); ok && replyStr == "success" {
		return "success", nil
	}

	// 将回复转换为XML
	replyXML, err := p.convertReplyToXML(reply)
	if err != nil {
		return nil, fmt.Errorf("转换回复为XML失败: %v", err)
	}

	// 加密回复（使用相同的timestamp和nonce）
	encryptedReply, err := cryptoInstance.EncryptMsg(replyXML, timestamp, nonce)
	if err != nil {
		return nil, fmt.Errorf("回复加密失败: %v", err)
	}

	return encryptedReply, nil
}

// EncryptReply 加密回复消息
func (p *SecureMessageProcessor) EncryptReply(
	authorizerAppID string,
	reply interface{},
	timestamp string,
	nonce string,
) (string, error) {
	// 获取加解密实例
	cryptoInstance, err := p.getCryptoInstance(authorizerAppID)
	if err != nil {
		return "", fmt.Errorf("获取加解密实例失败: %v", err)
	}

	// 将回复转换为XML
	replyXML, err := p.convertReplyToXML(reply)
	if err != nil {
		return "", fmt.Errorf("转换回复为XML失败: %v", err)
	}

	// 加密回复
	encryptedReply, err := cryptoInstance.EncryptMsg(replyXML, timestamp, nonce)
	if err != nil {
		return "", fmt.Errorf("回复加密失败: %v", err)
	}

	return encryptedReply, nil
}

// getCryptoInstance 获取加解密实例
func (p *SecureMessageProcessor) getCryptoInstance(authorizerAppID string) (*crypto.WXBizMsgCrypt, error) {
	if cryptoInstance, exists := p.cryptoCache[authorizerAppID]; exists {
		return cryptoInstance, nil
	}

	// 从授权方信息中获取Token和EncodingAESKey
	// 这里需要实现获取授权方配置的逻辑
	token := "your_token"                     // 需要从配置或数据库中获取
	encodingAESKey := "your_encoding_aes_key" // 需要从配置或数据库中获取

	cryptoInstance := crypto.NewWXBizMsgCrypt(token, encodingAESKey, authorizerAppID)
	p.cryptoCache[authorizerAppID] = cryptoInstance

	return cryptoInstance, nil
}

// validateTimestamp 验证时间戳（防止重放攻击，符合微信官方规范）
func (p *SecureMessageProcessor) validateTimestamp(timestamp string) error {
	// 解析时间戳
	ts, err := parseTimestamp(timestamp)
	if err != nil {
		return fmt.Errorf("时间戳格式错误: %v", err)
	}

	// 获取当前时间
	now := time.Now().Unix()

	// 检查时间戳是否在合理的时间范围内（前后5分钟内）
	// 微信官方建议的时间窗口为5分钟
	const timeWindow = 5 * 60 // 5分钟，单位秒

	if ts < now-timeWindow || ts > now+timeWindow {
		return fmt.Errorf("时间戳超出有效范围，当前时间戳: %d, 服务器时间: %d", ts, now)
	}

	return nil
}

// parseTimestamp 解析时间戳字符串
func parseTimestamp(timestamp string) (int64, error) {
	// 尝试解析为整数
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("时间戳必须是有效的整数: %v", err)
	}

	// 验证时间戳范围（不能是未来太远的时间）
	if ts < 0 {
		return 0, fmt.Errorf("时间戳不能为负数")
	}

	// 检查时间戳是否在合理范围内（不能超过当前时间10年）
	maxFuture := time.Now().Add(10 * 365 * 24 * time.Hour).Unix()
	if ts > maxFuture {
		return 0, fmt.Errorf("时间戳超出合理范围")
	}

	return ts, nil
}

// VerifyURL 验证URL（符合微信官方规范）
func (p *SecureMessageProcessor) VerifyURL(
	authorizerAppID string,
	msgSignature string,
	timestamp string,
	nonce string,
	echostr string,
) (string, error) {
	// 获取加解密实例
	cryptoInstance, err := p.getCryptoInstance(authorizerAppID)
	if err != nil {
		return "", fmt.Errorf("获取加解密实例失败: %v", err)
	}

	// 验证URL
	decryptedEchostr, err := cryptoInstance.VerifyURL(msgSignature, timestamp, nonce, echostr)
	if err != nil {
		return "", fmt.Errorf("URL验证失败: %v", err)
	}

	return decryptedEchostr, nil
}

// convertReplyToXML 将回复转换为XML
func (p *SecureMessageProcessor) convertReplyToXML(reply interface{}) (string, error) {
	switch v := reply.(type) {
	case string:
		// 如果是字符串，直接返回
		return v, nil
	case *TextMessage:
		// 文本消息回复
		return p.convertTextMessageToXML(v)
	case *ImageMessage:
		// 图片消息回复
		return p.convertImageMessageToXML(v)
	case *VoiceMessage:
		// 语音消息回复
		return p.convertVoiceMessageToXML(v)
	case *VideoMessage:
		// 视频消息回复
		return p.convertVideoMessageToXML(v)
	default:
		return "", fmt.Errorf("不支持的回复类型: %T", reply)
	}
}

// convertTextMessageToXML 将文本消息转换为XML
func (p *SecureMessageProcessor) convertTextMessageToXML(msg *TextMessage) (string, error) {
	type TextReply struct {
		XMLName      xml.Name `xml:"xml"`
		ToUserName   CDATA    `xml:"ToUserName"`
		FromUserName CDATA    `xml:"FromUserName"`
		CreateTime   int64    `xml:"CreateTime"`
		MsgType      CDATA    `xml:"MsgType"`
		Content      CDATA    `xml:"Content"`
	}

	reply := TextReply{
		ToUserName:   CDATA{Value: msg.ToUserName},
		FromUserName: CDATA{Value: msg.FromUserName},
		CreateTime:   msg.CreateTime,
		MsgType:      CDATA{Value: "text"},
		Content:      CDATA{Value: msg.Content},
	}

	output, err := xml.MarshalIndent(reply, "", "  ")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// convertImageMessageToXML 将图片消息转换为XML
func (p *SecureMessageProcessor) convertImageMessageToXML(msg *ImageMessage) (string, error) {
	type ImageReply struct {
		XMLName      xml.Name `xml:"xml"`
		ToUserName   CDATA    `xml:"ToUserName"`
		FromUserName CDATA    `xml:"FromUserName"`
		CreateTime   int64    `xml:"CreateTime"`
		MsgType      CDATA    `xml:"MsgType"`
		Image        struct {
			MediaID CDATA `xml:"MediaId"`
		} `xml:"Image"`
	}

	reply := ImageReply{
		ToUserName:   CDATA{Value: msg.ToUserName},
		FromUserName: CDATA{Value: msg.FromUserName},
		CreateTime:   msg.CreateTime,
		MsgType:      CDATA{Value: "image"},
	}
	reply.Image.MediaID = CDATA{Value: msg.MediaID}

	output, err := xml.MarshalIndent(reply, "", "  ")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// convertVoiceMessageToXML 将语音消息转换为XML
func (p *SecureMessageProcessor) convertVoiceMessageToXML(msg *VoiceMessage) (string, error) {
	type VoiceReply struct {
		XMLName      xml.Name `xml:"xml"`
		ToUserName   CDATA    `xml:"ToUserName"`
		FromUserName CDATA    `xml:"FromUserName"`
		CreateTime   int64    `xml:"CreateTime"`
		MsgType      CDATA    `xml:"MsgType"`
		Voice        struct {
			MediaID CDATA `xml:"MediaId"`
		} `xml:"Voice"`
	}

	reply := VoiceReply{
		ToUserName:   CDATA{Value: msg.ToUserName},
		FromUserName: CDATA{Value: msg.FromUserName},
		CreateTime:   msg.CreateTime,
		MsgType:      CDATA{Value: "voice"},
	}
	reply.Voice.MediaID = CDATA{Value: msg.MediaID}

	output, err := xml.MarshalIndent(reply, "", "  ")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// convertVideoMessageToXML 将视频消息转换为XML
func (p *SecureMessageProcessor) convertVideoMessageToXML(msg *VideoMessage) (string, error) {
	type VideoReply struct {
		XMLName      xml.Name `xml:"xml"`
		ToUserName   CDATA    `xml:"ToUserName"`
		FromUserName CDATA    `xml:"FromUserName"`
		CreateTime   int64    `xml:"CreateTime"`
		MsgType      CDATA    `xml:"MsgType"`
		Video        struct {
			MediaID     CDATA `xml:"MediaId"`
			Title       CDATA `xml:"Title"`
			Description CDATA `xml:"Description"`
		} `xml:"Video"`
	}

	reply := VideoReply{
		ToUserName:   CDATA{Value: msg.ToUserName},
		FromUserName: CDATA{Value: msg.FromUserName},
		CreateTime:   msg.CreateTime,
		MsgType:      CDATA{Value: "video"},
	}
	reply.Video.MediaID = CDATA{Value: msg.MediaID}
	reply.Video.Title = CDATA{Value: ""}
	reply.Video.Description = CDATA{Value: ""}

	output, err := xml.MarshalIndent(reply, "", "  ")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// CDATA XML CDATA类型
type CDATA struct {
	Value string `xml:",cdata"`
}

// MessageType 消息类型常量
const (
	MessageTypeText       = "text"
	MessageTypeImage      = "image"
	MessageTypeVoice      = "voice"
	MessageTypeVideo      = "video"
	MessageTypeShortVideo = "shortvideo"
	MessageTypeLocation   = "location"
	MessageTypeLink       = "link"
	MessageTypeEvent      = "event"
)

// EventType 事件类型常量
const (
	EventTypeComponentVerifyTicket = "component_verify_ticket"
	EventTypeUnauthorized          = "unauthorized"
	EventTypeAuthorized            = "authorized"
	EventTypeUpdateAuthorized      = "updateauthorized"
)

// Message 基础消息结构
type Message struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
}

// TextMessage 文本消息
type TextMessage struct {
	Message
	Content string `xml:"Content"`
	MsgID   int64  `xml:"MsgId,omitempty"`
}

// ImageMessage 图片消息
type ImageMessage struct {
	Message
	PicURL  string `xml:"PicUrl"`
	MediaID string `xml:"MediaId"`
	MsgID   int64  `xml:"MsgId,omitempty"`
}

// EventMessage 事件消息
type EventMessage struct {
	Message
	Event string `xml:"Event"`
}

// ComponentVerifyTicketEvent 验证票据事件
type ComponentVerifyTicketEvent struct {
	EventMessage
	ComponentVerifyTicket string `xml:"ComponentVerifyTicket"`
}

// AuthorizedEvent 授权成功事件
type AuthorizedEvent struct {
	EventMessage
	AuthorizerAppID              string `xml:"AuthorizerAppid"`
	AuthorizationCode            string `xml:"AuthorizationCode"`
	AuthorizationCodeExpiredTime int64  `xml:"AuthorizationCodeExpiredTime"`
	PreAuthCode                  string `xml:"PreAuthCode"`
}

// UnauthorizedEvent 取消授权事件
type UnauthorizedEvent struct {
	EventMessage
	AuthorizerAppID string `xml:"AuthorizerAppid"`
}

// UpdateAuthorizedEvent 授权更新事件
type UpdateAuthorizedEvent struct {
	EventMessage
	AuthorizerAppID              string `xml:"AuthorizerAppid"`
	AuthorizationCode            string `xml:"AuthorizationCode"`
	AuthorizationCodeExpiredTime int64  `xml:"AuthorizationCodeExpiredTime"`
	PreAuthCode                  string `xml:"PreAuthCode"`
}

// MessageHandler 消息处理器接口
type MessageHandler interface {
	HandleMessage(msg *Message) (interface{}, error)
}

// EventHandler 事件处理器接口
type EventHandler interface {
	HandleEvent(event *EventMessage) (interface{}, error)
}

// ComponentVerifyTicketHandler 第三方平台component_verify_ticket事件处理器接口
type ComponentVerifyTicketHandler interface {
	HandleComponentVerifyTicket(event *ComponentVerifyTicketEvent) error
}

// AuthorizeEventHandler 第三方平台授权事件处理器接口
type AuthorizeEventHandler interface {
	HandleAuthorizeEvent(event interface{}) error
}

// MessageProcessor 消息处理器
type MessageProcessor struct {
	messageHandlers               map[string]MessageHandler
	eventHandlers                 map[string]EventHandler
	componentVerifyTicketHandlers []ComponentVerifyTicketHandler
	authorizeEventHandlers        []AuthorizeEventHandler
}

// NewMessageProcessor 创建消息处理器
func NewMessageProcessor() *MessageProcessor {
	return &MessageProcessor{
		messageHandlers:               make(map[string]MessageHandler),
		eventHandlers:                 make(map[string]EventHandler),
		componentVerifyTicketHandlers: make([]ComponentVerifyTicketHandler, 0),
		authorizeEventHandlers:        make([]AuthorizeEventHandler, 0),
	}
}

// RegisterMessageHandler 注册消息处理器
func (p *MessageProcessor) RegisterMessageHandler(msgType string, handler MessageHandler) {
	p.messageHandlers[msgType] = handler
}

// RegisterEventHandler 注册事件处理器
func (p *MessageProcessor) RegisterEventHandler(eventType string, handler EventHandler) {
	p.eventHandlers[eventType] = handler
}

// RegisterComponentVerifyTicketHandler 注册component_verify_ticket事件处理器
func (p *MessageProcessor) RegisterComponentVerifyTicketHandler(handler ComponentVerifyTicketHandler) {
	p.componentVerifyTicketHandlers = append(p.componentVerifyTicketHandlers, handler)
}

// RegisterAuthorizeEventHandler 注册授权事件处理器
func (p *MessageProcessor) RegisterAuthorizeEventHandler(handler AuthorizeEventHandler) {
	p.authorizeEventHandlers = append(p.authorizeEventHandlers, handler)
}

// ProcessMessage 处理消息
func (p *MessageProcessor) ProcessMessage(xmlData []byte) (interface{}, error) {
	// 解析基础消息类型
	var baseMsg Message
	if err := xml.Unmarshal(xmlData, &baseMsg); err != nil {
		return nil, fmt.Errorf("解析XML消息失败: %v", err)
	}

	// 检查是否为第三方平台特殊事件
	// 第三方平台事件通常有特定的Event类型
	if baseMsg.MsgType == MessageTypeEvent {
		var eventMsg EventMessage
		if err := xml.Unmarshal(xmlData, &eventMsg); err == nil {
			// 检查是否为第三方平台特定事件
			switch eventMsg.Event {
			case EventTypeComponentVerifyTicket, EventTypeAuthorized, EventTypeUpdateAuthorized, EventTypeUnauthorized:
				return p.handleThirdPartyMessage(xmlData, &baseMsg)
			}
		}
	}

	// 根据消息类型进行具体解析
	switch baseMsg.MsgType {
	case MessageTypeEvent:
		return p.processEventMessage(xmlData)
	case MessageTypeText:
		return p.processTextMessage(xmlData)
	case MessageTypeImage:
		return p.processImageMessage(xmlData)
	case MessageTypeVoice:
		return p.processVoiceMessage(xmlData)
	case MessageTypeVideo:
		return p.processVideoMessage(xmlData)
	case MessageTypeLocation:
		return p.processLocationMessage(xmlData)
	case MessageTypeLink:
		return p.processLinkMessage(xmlData)
	default:
		return nil, fmt.Errorf("不支持的消息类型: %s", baseMsg.MsgType)
	}
}

// handleThirdPartyMessage 处理第三方平台消息
func (p *MessageProcessor) handleThirdPartyMessage(xmlData []byte, msg *Message) (interface{}, error) {
	// 解析第三方平台事件
	var event EventMessage
	err := xml.Unmarshal(xmlData, &event)
	if err != nil {
		return nil, fmt.Errorf("解析第三方平台事件失败: %v", err)
	}

	// 根据事件类型进行处理
	switch event.Event {
	case EventTypeComponentVerifyTicket:
		return p.handleComponentVerifyTicketEvent(xmlData)
	case EventTypeAuthorized, EventTypeUpdateAuthorized, EventTypeUnauthorized:
		return p.handleAuthorizeEvent(xmlData)
	default:
		// 如果不是第三方平台特定事件，则按普通事件处理
		return p.processEventMessage(xmlData)
	}
}

// handleComponentVerifyTicketEvent 处理component_verify_ticket事件
func (p *MessageProcessor) handleComponentVerifyTicketEvent(xmlData []byte) (interface{}, error) {
	var event ComponentVerifyTicketEvent
	err := xml.Unmarshal(xmlData, &event)
	if err != nil {
		return nil, fmt.Errorf("解析component_verify_ticket事件失败: %v", err)
	}

	// 调用所有注册的处理器
	for _, handler := range p.componentVerifyTicketHandlers {
		err := handler.HandleComponentVerifyTicket(&event)
		if err != nil {
			return nil, fmt.Errorf("处理component_verify_ticket事件失败: %v", err)
		}
	}

	// 返回成功响应
	return "success", nil
}

// handleAuthorizeEvent 处理授权事件
func (p *MessageProcessor) handleAuthorizeEvent(xmlData []byte) (interface{}, error) {
	// 解析XML获取事件类型
	var baseEvent EventMessage
	err := xml.Unmarshal(xmlData, &baseEvent)
	if err != nil {
		return nil, fmt.Errorf("解析授权事件失败: %v", err)
	}

	// 根据事件类型解析具体的事件
	switch baseEvent.Event {
	case "authorized":
		var event AuthorizedEvent
		err := xml.Unmarshal(xmlData, &event)
		if err != nil {
			return nil, fmt.Errorf("解析授权成功事件失败: %v", err)
		}
		// 调用所有注册的处理器
		for _, handler := range p.authorizeEventHandlers {
			err := handler.HandleAuthorizeEvent(&event)
			if err != nil {
				return nil, fmt.Errorf("处理授权事件失败: %v", err)
			}
		}
	case "unauthorized":
		var event UnauthorizedEvent
		err := xml.Unmarshal(xmlData, &event)
		if err != nil {
			return nil, fmt.Errorf("解析取消授权事件失败: %v", err)
		}
		// 调用所有注册的处理器
		for _, handler := range p.authorizeEventHandlers {
			err := handler.HandleAuthorizeEvent(&event)
			if err != nil {
				return nil, fmt.Errorf("处理授权事件失败: %v", err)
			}
		}
	case "updateauthorized":
		var event UpdateAuthorizedEvent
		err := xml.Unmarshal(xmlData, &event)
		if err != nil {
			return nil, fmt.Errorf("解析授权更新事件失败: %v", err)
		}
		// 调用所有注册的处理器
		for _, handler := range p.authorizeEventHandlers {
			err := handler.HandleAuthorizeEvent(&event)
			if err != nil {
				return nil, fmt.Errorf("处理授权事件失败: %v", err)
			}
		}
	default:
		return nil, fmt.Errorf("不支持的授权事件类型: %s", baseEvent.Event)
	}

	// 返回成功响应
	return "success", nil
}

// processEventMessage 处理事件消息
func (p *MessageProcessor) processEventMessage(xmlData []byte) (interface{}, error) {
	var event EventMessage
	if err := xml.Unmarshal(xmlData, &event); err != nil {
		return nil, fmt.Errorf("解析事件消息失败: %v", err)
	}

	handler, exists := p.eventHandlers[event.Event]
	if !exists {
		return nil, fmt.Errorf("未注册的事件处理器: %s", event.Event)
	}

	return handler.HandleEvent(&event)
}

// processTextMessage 处理文本消息
func (p *MessageProcessor) processTextMessage(xmlData []byte) (interface{}, error) {
	var textMsg TextMessage
	if err := xml.Unmarshal(xmlData, &textMsg); err != nil {
		return nil, fmt.Errorf("解析文本消息失败: %v", err)
	}

	handler, exists := p.messageHandlers[MessageTypeText]
	if !exists {
		return nil, fmt.Errorf("未注册的文本消息处理器")
	}

	return handler.HandleMessage(&textMsg.Message)
}

// processImageMessage 处理图片消息
func (p *MessageProcessor) processImageMessage(xmlData []byte) (interface{}, error) {
	var imageMsg ImageMessage
	if err := xml.Unmarshal(xmlData, &imageMsg); err != nil {
		return nil, fmt.Errorf("解析图片消息失败: %v", err)
	}

	handler, exists := p.messageHandlers[MessageTypeImage]
	if !exists {
		return nil, fmt.Errorf("未注册的图片消息处理器")
	}

	return handler.HandleMessage(&imageMsg.Message)
}

// processVoiceMessage 处理语音消息
func (p *MessageProcessor) processVoiceMessage(xmlData []byte) (interface{}, error) {
	var voiceMsg VoiceMessage
	if err := xml.Unmarshal(xmlData, &voiceMsg); err != nil {
		return nil, fmt.Errorf("解析语音消息失败: %v", err)
	}

	handler, exists := p.messageHandlers[MessageTypeVoice]
	if !exists {
		return nil, fmt.Errorf("未注册的语音消息处理器")
	}

	return handler.HandleMessage(&voiceMsg.Message)
}

// processVideoMessage 处理视频消息
func (p *MessageProcessor) processVideoMessage(xmlData []byte) (interface{}, error) {
	var videoMsg VideoMessage
	if err := xml.Unmarshal(xmlData, &videoMsg); err != nil {
		return nil, fmt.Errorf("解析视频消息失败: %v", err)
	}

	handler, exists := p.messageHandlers[MessageTypeVideo]
	if !exists {
		return nil, fmt.Errorf("未注册的视频消息处理器")
	}

	return handler.HandleMessage(&videoMsg.Message)
}

// processLocationMessage 处理位置消息
func (p *MessageProcessor) processLocationMessage(xmlData []byte) (interface{}, error) {
	var locationMsg LocationMessage
	if err := xml.Unmarshal(xmlData, &locationMsg); err != nil {
		return nil, fmt.Errorf("解析位置消息失败: %v", err)
	}

	handler, exists := p.messageHandlers[MessageTypeLocation]
	if !exists {
		return nil, fmt.Errorf("未注册的位置消息处理器")
	}

	return handler.HandleMessage(&locationMsg.Message)
}

// processLinkMessage 处理链接消息
func (p *MessageProcessor) processLinkMessage(xmlData []byte) (interface{}, error) {
	var linkMsg LinkMessage
	if err := xml.Unmarshal(xmlData, &linkMsg); err != nil {
		return nil, fmt.Errorf("解析链接消息失败: %v", err)
	}

	handler, exists := p.messageHandlers[MessageTypeLink]
	if !exists {
		return nil, fmt.Errorf("未注册的链接消息处理器")
	}

	return handler.HandleMessage(&linkMsg.Message)
}

// VoiceMessage 语音消息
type VoiceMessage struct {
	Message
	MediaID string `xml:"MediaId"`
	Format  string `xml:"Format"`
	MsgID   int64  `xml:"MsgId,omitempty"`
}

// VideoMessage 视频消息
type VideoMessage struct {
	Message
	MediaID      string `xml:"MediaId"`
	ThumbMediaID string `xml:"ThumbMediaId"`
	MsgID        int64  `xml:"MsgId,omitempty"`
}

// LocationMessage 位置消息
type LocationMessage struct {
	Message
	LocationX float64 `xml:"Location_X"`
	LocationY float64 `xml:"Location_Y"`
	Scale     int     `xml:"Scale"`
	Label     string  `xml:"Label"`
	MsgID     int64   `xml:"MsgId,omitempty"`
}

// LinkMessage 链接消息
type LinkMessage struct {
	Message
	Title       string `xml:"Title"`
	Description string `xml:"Description"`
	URL         string `xml:"Url"`
	MsgID       int64  `xml:"MsgId,omitempty"`
}

// SecureMessageProcessor 方法实现

// RegisterMessageHandler 注册消息处理器
func (p *SecureMessageProcessor) RegisterMessageHandler(msgType string, handler MessageHandler) {
	p.processor.RegisterMessageHandler(msgType, handler)
}

// RegisterEventHandler 注册事件处理器
func (p *SecureMessageProcessor) RegisterEventHandler(eventType string, handler EventHandler) {
	p.processor.RegisterEventHandler(eventType, handler)
}

// RegisterComponentVerifyTicketHandler 注册component_verify_ticket事件处理器
func (p *SecureMessageProcessor) RegisterComponentVerifyTicketHandler(handler ComponentVerifyTicketHandler) {
	p.processor.RegisterComponentVerifyTicketHandler(handler)
}

// RegisterAuthorizeEventHandler 注册授权事件处理器
func (p *SecureMessageProcessor) RegisterAuthorizeEventHandler(handler AuthorizeEventHandler) {
	p.processor.RegisterAuthorizeEventHandler(handler)
}

// ProcessMessage 处理明文消息
func (p *SecureMessageProcessor) ProcessMessage(xmlData []byte) (interface{}, error) {
	return p.processor.ProcessMessage(xmlData)
}