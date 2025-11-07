package core

import (
	"fmt"
)

// APIResponse 微信API通用响应结构
type APIResponse struct {
	ErrCode int    `json:"errcode"` // 错误码，0表示成功
	ErrMsg  string `json:"errmsg"`  // 错误描述
}

// Error 实现error接口
func (r *APIResponse) Error() string {
	return fmt.Sprintf("微信API错误[%d]: %s", r.ErrCode, r.ErrMsg)
}

// IsSuccess 检查API响应是否成功
func (r *APIResponse) IsSuccess() bool {
	return r.ErrCode == 0
}
