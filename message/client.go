package message

import (
	"encoding/xml"
	"fmt"

	"github.com/jcbowen/wego/openplatform"
)

// MessageClient 消息客户端
type MessageClient struct {
	Client *openplatform.OpenPlatformClient
}

// NewMessageClient 创建新的消息客户端
func NewMessageClient(client *openplatform.OpenPlatformClient) *MessageClient {
	return &MessageClient{
		Client: client,
	}
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

// MessageProcessor 消息处理器
type MessageProcessor struct {
	handlers      map[string]MessageHandler
	eventHandlers map[string]EventHandler
}

// NewMessageProcessor 创建消息处理器
func NewMessageProcessor() *MessageProcessor {
	return &MessageProcessor{
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

// processVoiceMessage 处理语音消息
func (p *MessageProcessor) processVoiceMessage(xmlData []byte) (interface{}, error) {
	var voiceMsg VoiceMessage
	if err := xml.Unmarshal(xmlData, &voiceMsg); err != nil {
		return nil, fmt.Errorf("解析语音消息失败: %v", err)
	}

	handler, exists := p.handlers[MessageTypeVoice]
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

	handler, exists := p.handlers[MessageTypeVideo]
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

	handler, exists := p.handlers[MessageTypeLocation]
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

	handler, exists := p.handlers[MessageTypeLink]
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