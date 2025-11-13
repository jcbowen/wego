package logger

import (
    "github.com/jcbowen/jcbaseGo/component/debugger"
)

// LoggerInterface 日志记录器接口
// 直接使用 debugger.LoggerInterface 定义，确保与 jcbaseGo 完全兼容
// 接口方法：Info/Warn/Error/WithFields/GetLevel（返回 debugger.LogLevel）
// 使用示例：
//
//  logger.Info("普通信息")
//  logger.WithFields(map[string]interface{}{"user_id": 123}).Warn("用户操作告警")
//  if logger.GetLevel() == debugger.LevelError { /* ... */ }
type LoggerInterface = debugger.LoggerInterface
