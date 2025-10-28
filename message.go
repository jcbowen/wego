package wego

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"time"
)

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

// MessageProcessor 消息处理器
type MessageProcessor struct {
	config        *WeGoConfig
	handlers      map[string]MessageHandler
	eventHandlers map[string]EventHandler
}

// NewMessageProcessor 创建消息处理器
func NewMessageProcessor(config *WeGoConfig) *MessageProcessor {
	return &MessageProcessor{
		config:        config,
		handlers:      make(map[string]MessageHandler),
		eventHandlers: make(map[string]EventHandler),
	}
}

// RegisterMessageHandler 注册消息处理器
func (p *MessageProcessor) RegisterMessageHandler(msgType string, handler MessageHandler) {
	p.handlers[msgType] = handler
}

// RegisterEventHandler 注册事件处理器
func (p *MessageProcessor) RegisterEventHandler(eventType string, handler EventHandler) {
	p.eventHandlers[eventType] = handler
}

// ProcessMessage 处理消息
func (p *MessageProcessor) ProcessMessage(xmlData []byte) (interface{}, error) {
	// 解析基础消息类型
	var baseMsg Message
	if err := xml.Unmarshal(xmlData, &baseMsg); err != nil {
		return nil, fmt.Errorf("解析XML消息失败: %v", err)
	}

	// 根据消息类型进行具体解析
	switch baseMsg.MsgType {
	case MessageTypeEvent:
		return p.processEventMessage(xmlData)
	case MessageTypeText:
		return p.processTextMessage(xmlData)
	case MessageTypeImage:
		return p.processImageMessage(xmlData)
	default:
		return nil, fmt.Errorf("不支持的消息类型: %s", baseMsg.MsgType)
	}
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

	handler, exists := p.handlers[MessageTypeText]
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

	handler, exists := p.handlers[MessageTypeImage]
	if !exists {
		return nil, fmt.Errorf("未注册的图片消息处理器")
	}

	return handler.HandleMessage(&imageMsg.Message)
}

// VerifySignature 验证消息签名
func (p *MessageProcessor) VerifySignature(signature, timestamp, nonce string) bool {
	return p.generateSignature(timestamp, nonce) == signature
}

// generateSignature 生成消息签名
func (p *MessageProcessor) generateSignature(timestamp, nonce string) string {
	strs := []string{p.config.ComponentToken, timestamp, nonce}
	sort.Strings(strs)
	str := strings.Join(strs, "")
	h := sha1.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// EncryptMessage 加密消息
func (p *MessageProcessor) EncryptMessage(plainText string) (string, error) {
	if p.config.EncodingAESKey == "" {
		return plainText, nil
	}

	aesKey, err := DecodeAESKey(p.config.EncodingAESKey)
	if err != nil {
		return "", fmt.Errorf("解码AES密钥失败: %v", err)
	}

	encrypted, err := EncryptMsg(plainText, p.config.ComponentAppID, aesKey)
	if err != nil {
		return "", fmt.Errorf("加密消息失败: %v", err)
	}

	return encrypted, nil
}

// DecryptMessage 解密消息
func (p *MessageProcessor) DecryptMessage(encryptedMsg string) (string, error) {
	if p.config.EncodingAESKey == "" {
		return encryptedMsg, nil
	}

	aesKey, err := DecodeAESKey(p.config.EncodingAESKey)
	if err != nil {
		return "", fmt.Errorf("解码AES密钥失败: %v", err)
	}

	decrypted, err := DecryptMsg(encryptedMsg, aesKey)
	if err != nil {
		return "", fmt.Errorf("解密消息失败: %v", err)
	}

	return decrypted, nil
}

// GenerateTextResponse 生成文本回复消息
func (p *MessageProcessor) GenerateTextResponse(toUser, fromUser, content string) (string, error) {
	resp := TextMessage{
		Message: Message{
			ToUserName:   toUser,
			FromUserName: fromUser,
			CreateTime:   time.Now().Unix(),
			MsgType:      MessageTypeText,
		},
		Content: content,
	}

	xmlData, err := xml.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("序列化回复消息失败: %v", err)
	}

	return string(xmlData), nil
}

// GenerateSuccessResponse 生成成功回复
func (p *MessageProcessor) GenerateSuccessResponse() string {
	return "success"
}

// DefaultMessageHandler 默认消息处理器
type DefaultMessageHandler struct{}

// HandleMessage 处理消息
func (h *DefaultMessageHandler) HandleMessage(msg *Message) (interface{}, error) {
	return "success", nil
}

// DefaultEventHandler 默认事件处理器
type DefaultEventHandler struct{}

// HandleEvent 处理事件
func (h *DefaultEventHandler) HandleEvent(event *EventMessage) (interface{}, error) {
	switch event.Event {
	case EventTypeComponentVerifyTicket:
		return h.handleComponentVerifyTicket(event)
	case EventTypeAuthorized:
		return h.handleAuthorized(event)
	case EventTypeUnauthorized:
		return h.handleUnauthorized(event)
	case EventTypeUpdateAuthorized:
		return h.handleUpdateAuthorized(event)
	default:
		return "success", nil
	}
}

// handleComponentVerifyTicket 处理验证票据事件
func (h *DefaultEventHandler) handleComponentVerifyTicket(event *EventMessage) (interface{}, error) {
	// 这里应该保存component_verify_ticket
	// 实际使用时需要实现具体的存储逻辑
	return "success", nil
}

// handleAuthorized 处理授权成功事件
func (h *DefaultEventHandler) handleAuthorized(event *EventMessage) (interface{}, error) {
	// 这里应该处理授权成功逻辑
	// 实际使用时需要实现具体的业务逻辑
	return "success", nil
}

// handleUnauthorized 处理取消授权事件
func (h *DefaultEventHandler) handleUnauthorized(event *EventMessage) (interface{}, error) {
	// 这里应该处理取消授权逻辑
	// 实际使用时需要实现具体的业务逻辑
	return "success", nil
}

// handleUpdateAuthorized 处理授权更新事件
func (h *DefaultEventHandler) handleUpdateAuthorized(event *EventMessage) (interface{}, error) {
	// 这里应该处理授权更新逻辑
	// 实际使用时需要实现具体的业务逻辑
	return "success", nil
}

// ParseComponentVerifyTicket 解析验证票据事件
func ParseComponentVerifyTicket(xmlData []byte) (*ComponentVerifyTicketEvent, error) {
	var event ComponentVerifyTicketEvent
	if err := xml.Unmarshal(xmlData, &event); err != nil {
		return nil, fmt.Errorf("解析验证票据事件失败: %v", err)
	}
	return &event, nil
}

// ParseAuthorizedEvent 解析授权成功事件
func ParseAuthorizedEvent(xmlData []byte) (*AuthorizedEvent, error) {
	var event AuthorizedEvent
	if err := xml.Unmarshal(xmlData, &event); err != nil {
		return nil, fmt.Errorf("解析授权成功事件失败: %v", err)
	}
	return &event, nil
}

// ParseUnauthorizedEvent 解析取消授权事件
func ParseUnauthorizedEvent(xmlData []byte) (*UnauthorizedEvent, error) {
	var event UnauthorizedEvent
	if err := xml.Unmarshal(xmlData, &event); err != nil {
		return nil, fmt.Errorf("解析取消授权事件失败: %v", err)
	}
	return &event, nil
}

// ParseUpdateAuthorizedEvent 解析授权更新事件
func ParseUpdateAuthorizedEvent(xmlData []byte) (*UpdateAuthorizedEvent, error) {
	var event UpdateAuthorizedEvent
	if err := xml.Unmarshal(xmlData, &event); err != nil {
		return nil, fmt.Errorf("解析授权更新事件失败: %v", err)
	}
	return &event, nil
}
