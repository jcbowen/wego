package logger

import (
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

// NewLogger 创建日志记录器实例
// 支持直接传入debugger的logger实现，确保完全兼容
//
// @param logger 可选的debugger.LoggerInterface实例
//   如果传入非nil的debugger logger，则直接使用该实例
//   如果传入nil，则创建独立的默认logger实现
//
// @return LoggerInterface 日志记录器实例
//
// 使用示例：
//   // 使用默认logger
//   logger := logger.NewLogger(nil)
//   
//   // 使用debugger的logger（如果可用）
//   debuggerLogger := debugger.NewDefaultLogger(someDebugger)
//   logger := logger.NewLogger(debuggerLogger)
//
// 注意：此函数设计用于支持平滑迁移，可以逐步替换debugger依赖
func NewLogger(logger debugger.LoggerInterface) LoggerInterface {
	if logger != nil {
		return logger // 直接使用传入的debugger logger
	}
	return NewDefaultLoggerInterface() // 否则使用独立的默认实现
}

// NewDefaultLoggerInterface 创建默认日志记录器接口实例
// 提供不依赖debugger组件的简单控制台输出实现
//
// @return LoggerInterface 默认日志记录器接口实例
//
// 使用示例：
//   logger := logger.NewDefaultLoggerInterface()
//   logger.Info("应用启动")
func NewDefaultLoggerInterface() LoggerInterface {
	return &DefaultLogger{
		level:  "info", // 默认级别为info
		fields: make(map[string]interface{}),
	}
}

// NewLoggerWithLevel 创建指定级别的日志记录器实例
//
// @param level 日志级别：debug, info, warn, error
// @return *DefaultLogger 默认日志记录器实例
//
// 使用示例：
//   logger := logger.NewLoggerWithLevel("debug")
//   logger.Debug("调试信息")
func NewLoggerWithLevel(level string) *DefaultLogger {
	logger := NewDefaultLogger()
	logger.SetLevel(level)
	return logger
}

// NewLoggerWithFields 创建带有初始字段的日志记录器实例
//
// @param fields 初始字段
// @return *DefaultLogger 默认日志记录器实例
//
// 使用示例：
//   logger := logger.NewLoggerWithFields(map[string]interface{}{"app": "wego"})
//   logger.Info("应用启动")
func NewLoggerWithFields(fields map[string]interface{}) *DefaultLogger {
	logger := NewDefaultLogger()
	logger.fields = fields
	return logger
}