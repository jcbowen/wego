package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FlexString 兼容 JSON 字符串/数字双形态的字段
//
// 背景：微信部分接口文档与实际返回不一致（如 queryblocktmplmsg 的 id 参数表为
// string 而示例为数字；mass/get 的 msg_id 文档标注 string）。使用本类型可同时
// 接受两种形态，统一以字符串形式存储，避免反序列化失败。
//
// 使用示例:
//
//	type Resp struct {
//		ID FlexString `json:"id"`
//	}
type FlexString string

// UnmarshalJSON 反序列化 JSON 值为 FlexString
//
// 参数:
//   - data: 原始 JSON 字节（字符串或数字字面量）
//
// 异常:
//   - JSON 字符串非法时返回反序列化错误
func (s *FlexString) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		*s = ""
		return nil
	}

	// 字符串形态：走标准解码以正确处理转义
	if trimmed[0] == '"' {
		var v string
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*s = FlexString(v)
		return nil
	}

	// 数字/其他字面量形态：原样保留字面量
	*s = FlexString(trimmed)
	return nil
}

// MarshalJSON 序列化 FlexString 为 JSON 字符串
func (s FlexString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// String 返回字符串值
func (s FlexString) String() string {
	return string(s)
}

// Int64 将字符串值解析为 int64
//
// 返回:
//   - int64: 解析结果（空串返回 0）
//   - error: 值不是合法整数时返回错误
func (s FlexString) Int64() (int64, error) {
	if s == "" {
		return 0, nil
	}
	var v int64
	if _, err := fmt.Sscanf(string(s), "%d", &v); err != nil {
		return 0, fmt.Errorf("值 %q 不是合法整数: %w", string(s), err)
	}
	return v, nil
}

// FlexBool 兼容 JSON 布尔/字符串双形态的字段
//
// 背景：微信部分接口文档将布尔语义字段标注为 string（如 sendall 的
// filter.is_to_all）。使用本类型可同时接受 true/false、"true"/"false"/"0"/"1"。
//
// 使用示例:
//
//	type Filter struct {
//		IsToAll FlexBool `json:"is_to_all"`
//	}
type FlexBool bool

// UnmarshalJSON 反序列化 JSON 值为 FlexBool
//
// 参数:
//   - data: 原始 JSON 字节（布尔或字符串字面量）
//
// 异常:
//   - 值形态不被识别时返回错误
func (b *FlexBool) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		*b = false
		return nil
	}

	// 布尔形态
	if trimmed == "true" {
		*b = true
		return nil
	}
	if trimmed == "false" {
		*b = false
		return nil
	}

	// 字符串形态
	var v string
	if err := json.Unmarshal(data, &v); err == nil {
		switch strings.ToLower(v) {
		case "true", "1":
			*b = true
		case "false", "0", "":
			*b = false
		default:
			return fmt.Errorf("无法将 %q 解析为布尔值", v)
		}
		return nil
	}

	return fmt.Errorf("无法将 %s 解析为布尔值", trimmed)
}

// MarshalJSON 序列化 FlexBool 为 JSON 布尔值
func (b FlexBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(bool(b))
}

// Bool 返回布尔值
func (b FlexBool) Bool() bool {
	return bool(b)
}
