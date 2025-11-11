package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// DefaultLogger 默认日志记录器实现
// 提供简单的控制台输出，不依赖debugger组件
// 支持不同日志级别和字段附加功能
type DefaultLogger struct {
	level  string                 // 日志级别：debug, info, warn, error
	fields map[string]interface{} // 附加字段
}

// NewDefaultLogger 创建默认日志记录器实例
// @return *DefaultLogger 默认日志记录器实例
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		level:  "info", // 默认级别为info
		fields: make(map[string]interface{}),
	}
}

// Debug 记录调试级别日志
// @param msg 日志消息，可以是任意类型
// @param fields 附加字段，可选参数
func (l *DefaultLogger) Debug(msg any, fields ...map[string]interface{}) {
	if !l.shouldLog("debug") {
		return
	}
	l.log("DEBUG", msg, fields)
}

// Info 记录信息级别日志
// @param msg 日志消息，可以是任意类型
// @param fields 附加字段，可选参数
func (l *DefaultLogger) Info(msg any, fields ...map[string]interface{}) {
	if !l.shouldLog("info") {
		return
	}
	l.log("INFO", msg, fields)
}

// Warn 记录警告级别日志
// @param msg 日志消息，可以是任意类型
// @param fields 附加字段，可选参数
func (l *DefaultLogger) Warn(msg any, fields ...map[string]interface{}) {
	if !l.shouldLog("warn") {
		return
	}
	l.log("WARN", msg, fields)
}

// Error 记录错误级别日志
// @param msg 日志消息，可以是任意类型
// @param fields 附加字段，可选参数
func (l *DefaultLogger) Error(msg any, fields ...map[string]interface{}) {
	if !l.shouldLog("error") {
		return
	}
	l.log("ERROR", msg, fields)
}

// WithFields 创建带有字段的日志记录器
// @param fields 附加字段
// @return LoggerInterface 新的日志记录器实例
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
		level:  l.level,
		fields: newFields,
	}
}

// GetLevel 获取当前日志记录器的日志级别
// @return string 日志级别
func (l *DefaultLogger) GetLevel() string {
	return l.level
}

// SetLevel 设置日志级别
// @param level 日志级别：debug, info, warn, error
func (l *DefaultLogger) SetLevel(level string) {
	l.level = strings.ToLower(level)
}

// shouldLog 判断是否应该记录指定级别的日志
// @param level 要记录的日志级别
// @return bool 是否应该记录
func (l *DefaultLogger) shouldLog(level string) bool {
	levelOrder := map[string]int{
		"debug": 1,
		"info":  2,
		"warn":  3,
		"error": 4,
	}

	currentLevel, ok := levelOrder[strings.ToLower(l.level)]
	if !ok {
		currentLevel = 2 // 默认info级别
	}

	logLevel, ok := levelOrder[strings.ToLower(level)]
	if !ok {
		return false
	}

	return logLevel >= currentLevel
}

// log 实际执行日志记录
// @param level 日志级别
// @param msg 日志消息
// @param fields 附加字段
func (l *DefaultLogger) log(level string, msg any, fields []map[string]interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 构建日志消息
	logMsg := fmt.Sprintf("[%s] %s %s", timestamp, level, l.formatMessage(msg))

	// 合并所有字段
	allFields := make(map[string]interface{})
	for k, v := range l.fields {
		allFields[k] = v
	}
	for _, fieldSet := range fields {
		for k, v := range fieldSet {
			allFields[k] = v
		}
	}

	// 如果有附加字段，添加到日志消息中
	if len(allFields) > 0 {
		logMsg += " " + l.formatFields(allFields)
	}

	// 根据级别输出到不同的标准输出
	switch strings.ToUpper(level) {
	case "ERROR", "WARN":
		log.New(os.Stderr, "", 0).Println(logMsg)
	default:
		log.New(os.Stdout, "", 0).Println(logMsg)
	}
}

// formatMessage 格式化日志消息
// @param msg 日志消息
// @return string 格式化后的消息
func (l *DefaultLogger) formatMessage(msg any) string {
	switch v := msg.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatFields 格式化附加字段
// @param fields 附加字段
// @return string 格式化后的字段字符串
func (l *DefaultLogger) formatFields(fields map[string]interface{}) string {
	if len(fields) == 0 {
		return ""
	}

	var parts []string
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	return "[" + strings.Join(parts, " ") + "]"
}
