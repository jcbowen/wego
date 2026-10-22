package core

import (
	"fmt"

	"github.com/jcbowen/jcbaseGo/component/debugger"
)

// ApplyClientOptions 遍历各模块构造函数的可选参数并分发。
//
// 内置支持两种通用类型：
//   - debugger.LoggerInterface: 自定义日志器（回调 onLogger）
//   - HTTPClient: 自定义HTTP客户端（回调 onHTTPClient）
//
// extraHandlers 允许模块注册额外类型的处理逻辑（如 openplatform.EventHandler），
// 返回 true 表示该参数已被消费，不再进入通用分支。
// 未被任何分支处理的类型通过 warn 回调记录警告（warn 为 nil 时忽略）。
//
// 注意：warn 使用闭包而非 logger 值，以便在 SetLogger 生效后仍能取到当前日志器，
// 与重构前各模块循环内 client.logger.Warn(...) 的行为保持一致。
//
// 参数:
//   - opts: 可选参数切片
//   - warn: 未知参数类型的警告回调（接收已格式化的消息），可为 nil
//   - onLogger: 处理 debugger.LoggerInterface 的回调，可为 nil
//   - onHTTPClient: 处理 HTTPClient 的回调，可为 nil
//   - extraHandlers: 模块自定义类型的处理器，可为空
//
// 返回值: 无
// 异常: 不主动panic；回调内部的行为由调用方决定
//
// 使用示例:
//
//	core.ApplyClientOptions(opts, func(msg string) { client.logger.Warn(msg) },
//	    client.SetLogger, client.SetHTTPClient)
//
//	// 带模块扩展类型:
//	core.ApplyClientOptions(opt, warn, client.SetLogger, client.SetHTTPClient,
//	    func(option any) (handled bool) {
//	        if h, ok := option.(EventHandler); ok {
//	            client.SetEventHandler(h)
//	            return true
//	        }
//	        return false
//	    })
func ApplyClientOptions(
	opts []any,
	warn func(msg string),
	onLogger func(debugger.LoggerInterface),
	onHTTPClient func(HTTPClient),
	extraHandlers ...func(option any) (handled bool),
) {
	for _, option := range opts {
		if handledByExtra(option, extraHandlers...) {
			continue
		}
		applyBuiltinOption(option, warn, onLogger, onHTTPClient)
	}
}

// handledByExtra 依次尝试 extraHandlers 消费参数，任一返回 true 即视为已处理
func handledByExtra(option any, extraHandlers ...func(option any) (handled bool)) bool {
	for _, handler := range extraHandlers {
		if handler != nil && handler(option) {
			return true
		}
	}
	return false
}

// applyBuiltinOption 将参数分发到内置的日志器/HTTP客户端分支，未识别类型输出警告
func applyBuiltinOption(
	option any,
	warn func(msg string),
	onLogger func(debugger.LoggerInterface),
	onHTTPClient func(HTTPClient),
) {
	switch v := option.(type) {
	case debugger.LoggerInterface:
		if onLogger != nil {
			onLogger(v)
		}
	case HTTPClient:
		if onHTTPClient != nil {
			onHTTPClient(v)
		}
	default:
		if warn != nil {
			warn(fmt.Sprintf("未知的可选参数类型: %T", v))
		}
	}
}
