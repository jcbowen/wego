package core

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jcbowen/jcbaseGo/component/helper"
)

// HTTPClient HTTP客户端接口
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// 日志级别常量
const (
	LevelDebug = "debug" // 调试级别：记录所有详细信息
	LevelInfo  = "info"  // 信息级别：只记录基本信息
	LevelWarn  = "warn"  // 警告级别：记录警告信息
	LevelError = "error" // 错误级别：记录错误信息
)

// LoggerInterface 日志接口
type LoggerInterface interface {
	// Debug 记录调试级别日志
	Debug(msg any, fields ...map[string]interface{})

	// Info 记录信息级别日志
	Info(msg any, fields ...map[string]interface{})

	// Warn 记录警告级别日志
	Warn(msg any, fields ...map[string]interface{})

	// Error 记录错误级别日志
	Error(msg any, fields ...map[string]interface{})

	// WithFields 创建带有字段的日志记录器
	WithFields(fields map[string]interface{}) LoggerInterface
}

// LoggerLog 记录通过logger打印的日志信息
type LoggerLog struct {
	Timestamp time.Time              `json:"timestamp"` // 日志时间戳
	Level     string                 `json:"level"`     // 日志级别：debug/info/warn/error
	Message   string                 `json:"message"`   // 日志消息
	Fields    map[string]interface{} `json:"fields"`    // 日志附加字段
}

// DefaultLogger 默认日志实现
type DefaultLogger struct {
	fields map[string]interface{}
	logs   []LoggerLog // 存储收集的日志
}

// Debug 记录调试级别日志
func (l *DefaultLogger) Debug(msg any, fields ...map[string]interface{}) {
	l.log(LevelDebug, msg, fields...)
}

// Info 记录信息级别日志
func (l *DefaultLogger) Info(msg any, fields ...map[string]interface{}) {
	l.log(LevelInfo, msg, fields...)
}

// Warn 记录警告级别日志
func (l *DefaultLogger) Warn(msg any, fields ...map[string]interface{}) {
	l.log(LevelWarn, msg, fields...)
}

// Error 记录错误级别日志
func (l *DefaultLogger) Error(msg any, fields ...map[string]interface{}) {
	l.log(LevelError, msg, fields...)
}

// WithFields 创建带有字段的日志记录器
func (l *DefaultLogger) WithFields(fields map[string]interface{}) LoggerInterface {
	// 合并现有字段和新字段
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &DefaultLogger{
		fields: newFields,
		logs:   l.logs, // 继承父logger的日志
	}
}

// log 内部日志记录方法
// - level: 日志级别（debug/info/warn/error）
// - msg: 日志消息（字符串、结构体、map、数组、实现了Stringer接口的类型等）
// - fields: 可选的附加字段（键值对）
func (l *DefaultLogger) log(level string, msg any, fields ...map[string]interface{}) {
	// 检查日志级别是否启用
	if !l.shouldLog(level) {
		return
	}

	// 处理msg参数
	var message string
	switch v := msg.(type) {
	case string:
		message = v
	case fmt.Stringer:
		message = v.String()
	default:
		helper.Json(msg).ToString(&message)
	}

	// 合并所有字段
	allFields := make(map[string]interface{})

	// 添加基础字段
	allFields["level"] = level
	allFields["message"] = message
	allFields["timestamp"] = time.Now().Format(time.RFC3339)

	// 添加实例字段
	for k, v := range l.fields {
		allFields[k] = v
	}

	// 添加调用方传入的字段
	if len(fields) > 0 {
		for k, v := range fields[0] {
			allFields[k] = v
		}
	}

	// 收集日志信息到logs字段
	loggerLog := LoggerLog{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    allFields,
	}
	l.logs = append(l.logs, loggerLog)

	// log.Println("[" + level + "] " + message)

	// 格式化日志输出
	logEntry := map[string]interface{}{
		"debug_log": allFields,
	}

	// 转换为JSON格式输出
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		// 如果JSON转换失败，使用简单格式输出
		log.Printf("[%s] %s: %s", level, time.Now().Format("2006-01-02 15:04:05"), msg)
		if len(l.fields) > 0 {
			log.Printf(" fields=%v", l.fields)
		}
		if len(fields) > 0 {
			log.Printf(" extra_fields=%v", fields[0])
		}
		return
	}

	log.Println(string(jsonData))
}

// shouldLog 检查是否应该记录指定级别的日志
func (l *DefaultLogger) shouldLog(level string) bool {
	return true
	// 根据配置的日志级别决定是否记录
	//switch l.config.Level {
	//case LevelDebug:
	//	// 调试级别记录所有日志
	//	return true
	//case LevelInfo:
	//	// 信息级别记录info、warn、error
	//	return level == LevelInfo || level == LevelWarn || level == LevelError
	//case LevelWarn:
	//	// 警告级别记录warn、error
	//	return level == LevelWarn || level == LevelError
	//case LevelError:
	//	// 错误级别只记录error
	//	return level == LevelError
	//default:
	//	// 默认记录所有日志
	//	return true
	//}
}
