package logger

import (
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

// LoggerInterface 日志记录器接口
// 直接使用debugger.LoggerInterface定义，确保完全兼容debugger的logger实现
// 支持不同级别的日志记录，可以在控制器中直接使用
//
// 接口方法说明：
// - Debug: 记录调试级别日志
// - Info: 记录信息级别日志  
// - Warn: 记录警告级别日志
// - Error: 记录错误级别日志
// - WithFields: 创建带有字段的日志记录器
// - GetLevel: 获取当前日志记录器的日志级别
//
// 使用示例：
//   logger.Debug("调试信息", map[string]interface{}{"key": "value"})
//   logger.Info("普通信息")
//   logger.WithFields(map[string]interface{}{"user_id": 123}).Info("用户操作")
type LoggerInterface = debugger.LoggerInterface