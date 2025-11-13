package logger

import (
    "fmt"
    "log"
    "os"
    "runtime"
    "strings"
    "time"

    "github.com/jcbowen/jcbaseGo/component/debugger"
)

// DefaultLogger 默认日志记录器实现
// 提供简单的控制台输出，不依赖debugger组件
// 支持不同日志级别和字段附加功能
type DefaultLogger struct {
    level  debugger.LogLevel
    fields map[string]interface{}
    logs   []LoggerLog
}

// NewDefaultLogger 创建默认日志记录器实例
// @return *DefaultLogger 默认日志记录器实例
func NewDefaultLogger() *DefaultLogger {
    return &DefaultLogger{
        level:  debugger.LevelInfo,
        fields: make(map[string]interface{}),
        logs:   []LoggerLog{},
    }
}

// Debug 方法已移除，jcbaseGo接口不包含该方法

// Info 记录信息级别日志
// @param msg 日志消息，可以是任意类型
// @param fields 附加字段，可选参数
func (l *DefaultLogger) Info(msg any, fields ...map[string]interface{}) {
    if !l.shouldLog(debugger.LevelInfo) {
        return
    }
    l.log(debugger.LevelInfo, msg, fields)
}

// Warn 记录警告级别日志
// @param msg 日志消息，可以是任意类型
// @param fields 附加字段，可选参数
func (l *DefaultLogger) Warn(msg any, fields ...map[string]interface{}) {
    if !l.shouldLog(debugger.LevelWarn) {
        return
    }
    l.log(debugger.LevelWarn, msg, fields)
}

// Error 记录错误级别日志
// @param msg 日志消息，可以是任意类型
// @param fields 附加字段，可选参数
func (l *DefaultLogger) Error(msg any, fields ...map[string]interface{}) {
    if !l.shouldLog(debugger.LevelError) {
        return
    }
    l.log(debugger.LevelError, msg, fields)
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
        logs:   l.logs,
    }
}

// GetLevel 获取当前日志记录器的日志级别
// @return string 日志级别
func (l *DefaultLogger) GetLevel() debugger.LogLevel {
	return l.level
}

// SetLevel 设置日志级别
// @param level 日志级别：debug, info, warn, error
func (l *DefaultLogger) SetLevel(level debugger.LogLevel) {
	l.level = level
}

// shouldLog 判断是否应该记录指定级别的日志
// @param level 要记录的日志级别
// @return bool 是否应该记录
func (l *DefaultLogger) shouldLog(level debugger.LogLevel) bool {
    switch l.GetLevel() {
    case debugger.LevelInfo:
        return true
    case debugger.LevelWarn:
        return level == debugger.LevelWarn || level == debugger.LevelError
    case debugger.LevelError:
        return level == debugger.LevelError
    default:
        return false
    }
}

// log 实际执行日志记录
// @param level 日志级别
// @param msg 日志消息
// @param fields 附加字段
func (l *DefaultLogger) log(level debugger.LogLevel, msg any, fields []map[string]interface{}) {
    // 处理消息
    message := l.formatMessage(msg)

    // 合并字段
    allFields := make(map[string]interface{})
    for k, v := range l.fields {
        allFields[k] = v
    }
    if len(fields) > 0 {
        for k, v := range fields[0] {
            allFields[k] = v
        }
    }

    // 位置信息
    fileName, line, function := getCallerInfo()

    // 记录到内部logs
    l.logs = append(l.logs, LoggerLog{
        Timestamp: time.Now(),
        Level:     level,
        Message:   message,
        Fields:    allFields,
        FileName:  fileName,
        Line:      line,
        Function:  function,
    })

    // 可点击格式输出：[LEVEL] file:line - message
    logMsg := fmt.Sprintf("[%s] %s:%d - %s", level.String(), fileName, line, message)
    if len(allFields) > 0 {
        logMsg += " " + l.formatFields(allFields)
    }

    switch level {
    case debugger.LevelError, debugger.LevelWarn:
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

// LoggerLog 日志记录结构
type LoggerLog struct {
    Timestamp time.Time
    Level     debugger.LogLevel
    Message   string
    Fields    map[string]interface{}
    FileName  string
    Line      int
    Function  string
}

func getCallerInfo() (string, int, string) {
    pc, file, line, ok := runtime.Caller(3)
    if !ok {
        return "unknown", 0, "unknown"
    }
    fn := runtime.FuncForPC(pc)
    function := "unknown"
    if fn != nil {
        function = fn.Name()
    }
    parts := strings.Split(file, "/")
    if len(parts) > 0 {
        file = parts[len(parts)-1]
    }
    return file, line, function
}

// levelLabel 已不再需要，使用 debugger.LogLevel.String()
