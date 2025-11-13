package logger

import (
    "github.com/jcbowen/jcbaseGo/component/debugger"
)

// DebuggerLoggerAdapter debugger日志记录器适配器
// 用于将debugger的logger实例适配为wego的logger接口
// 确保在迁移过程中可以无缝使用现有的debugger logger
type DebuggerLoggerAdapter struct {
    debuggerLogger debugger.LoggerInterface
}

// NewDebuggerLoggerAdapter 创建debugger日志记录器适配器
//
// @param debuggerLogger debugger.LoggerInterface实例
// @return *DebuggerLoggerAdapter 适配器实例
//
// 使用示例：
//
//	debuggerLogger := debugger.NewDefaultLogger(someDebugger)
//	adapter := logger.NewDebuggerLoggerAdapter(debuggerLogger)
//	adapter.Info("使用debugger logger")
func NewDebuggerLoggerAdapter(debuggerLogger debugger.LoggerInterface) *DebuggerLoggerAdapter {
	return &DebuggerLoggerAdapter{
		debuggerLogger: debuggerLogger,
	}
}

// Debug 方法已移除，jcbaseGo接口不包含该方法

// Info 记录信息级别日志
// @param msg 日志消息
// @param fields 附加字段
func (a *DebuggerLoggerAdapter) Info(msg any, fields ...map[string]interface{}) {
    a.debuggerLogger.Info(msg, fields...)
}

// Warn 记录警告级别日志
// @param msg 日志消息
// @param fields 附加字段
func (a *DebuggerLoggerAdapter) Warn(msg any, fields ...map[string]interface{}) {
    a.debuggerLogger.Warn(msg, fields...)
}

// Error 记录错误级别日志
// @param msg 日志消息
// @param fields 附加字段
func (a *DebuggerLoggerAdapter) Error(msg any, fields ...map[string]interface{}) {
    a.debuggerLogger.Error(msg, fields...)
}

// WithFields 创建带有字段的日志记录器
// @param fields 附加字段
// @return LoggerInterface 新的日志记录器实例
func (a *DebuggerLoggerAdapter) WithFields(fields map[string]interface{}) LoggerInterface {
    return &DebuggerLoggerAdapter{
        debuggerLogger: a.debuggerLogger.WithFields(fields),
    }
}

func (a *DebuggerLoggerAdapter) GetLevel() debugger.LogLevel {
    return a.debuggerLogger.GetLevel()
}

// IsDebuggerLogger 检查是否为debugger logger实例
// 用于在运行时判断logger实例的类型
//
// @param logger LoggerInterface实例
// @return bool 是否为debugger logger
func IsDebuggerLogger(logger LoggerInterface) bool {
	_, ok := logger.(*DebuggerLoggerAdapter)
	return ok
}

// GetUnderlyingDebuggerLogger 获取底层的debugger logger实例
// 如果logger是DebuggerLoggerAdapter，则返回底层的debugger logger
// 否则返回nil
//
// @param logger LoggerInterface实例
// @return debugger.LoggerInterface 底层的debugger logger实例
func GetUnderlyingDebuggerLogger(logger LoggerInterface) debugger.LoggerInterface {
    if adapter, ok := logger.(*DebuggerLoggerAdapter); ok {
        return adapter.debuggerLogger
    }
    return nil
}

// ConvertToDebuggerLogger 将wego logger转换为debugger logger
// 如果logger已经是debugger logger，则直接返回
// 否则创建一个适配器包装wego logger
//
// @param logger LoggerInterface实例
// @return debugger.LoggerInterface debugger logger实例
func ConvertToDebuggerLogger(logger LoggerInterface) debugger.LoggerInterface {
	if debuggerLogger := GetUnderlyingDebuggerLogger(logger); debuggerLogger != nil {
		return debuggerLogger
	}

	// 创建适配器包装wego logger
	return &wegoLoggerToDebuggerAdapter{
		wegoLogger: logger,
	}
}

// wegoLoggerToDebuggerAdapter wego logger到debugger logger的适配器
type wegoLoggerToDebuggerAdapter struct {
    wegoLogger LoggerInterface
}

func (a *wegoLoggerToDebuggerAdapter) Info(msg any, fields ...map[string]interface{}) {
    a.wegoLogger.Info(msg, fields...)
}

func (a *wegoLoggerToDebuggerAdapter) Warn(msg any, fields ...map[string]interface{}) {
    a.wegoLogger.Warn(msg, fields...)
}

func (a *wegoLoggerToDebuggerAdapter) Error(msg any, fields ...map[string]interface{}) {
    a.wegoLogger.Error(msg, fields...)
}

func (a *wegoLoggerToDebuggerAdapter) WithFields(fields map[string]interface{}) debugger.LoggerInterface {
    return &wegoLoggerToDebuggerAdapter{
        wegoLogger: a.wegoLogger.WithFields(fields),
    }
}

func (a *wegoLoggerToDebuggerAdapter) GetLevel() debugger.LogLevel {
    return a.wegoLogger.GetLevel()
}
