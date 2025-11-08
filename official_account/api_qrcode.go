// Package official_account 微信公众号API - 参数二维码
// 本接口支持「服务号（仅认证）」账号类型调用。其他账号类型如无特殊说明，均不可调用
package official_account

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jcbowen/wego/core"
)

const (
	ActionNameQRScene         = "QR_SCENE"           // 临时整型场景值
	ActionNameQRStrScene      = "QR_STR_SCENE"       // 临时字符串场景值
	ActionNameQRLimitScene    = "QR_LIMIT_SCENE"     // 永久整型场景值
	ActionNameQRLimitStrScene = "QR_LIMIT_STR_SCENE" // 永久字符串场景值
)

// QRCodeRequest 二维码请求
type QRCodeRequest struct {
	ExpireSeconds int64  `json:"expire_seconds,omitempty"` // 二维码有效时间（秒），最大2592000，仅临时二维码需要
	ActionName    string `json:"action_name"`              // 二维码类型：QR_SCENE(临时整型)/QR_STR_SCENE(临时字符串)/QR_LIMIT_SCENE(永久整型)/QR_LIMIT_STR_SCENE(永久字符串)
	ActionInfo    struct {
		Scene struct {
			SceneID  int    `json:"scene_id,omitempty"`  // 场景值ID，临时二维码时为32位非0整型，永久二维码时最大值为100000（目前参数只支持1--100000）
			SceneStr string `json:"scene_str,omitempty"` // 场景值ID（字符串形式的ID），字符串类型，长度限制为1到64
		} `json:"scene"` // 场景信息
	} `json:"action_info"` // 二维码详细信息
}

// QRCodeResponse 二维码响应
type QRCodeResponse struct {
	core.APIResponse
	Ticket        string `json:"ticket"`         // 获取的二维码ticket，凭借此ticket可以在有效时间内换取二维码
	ExpireSeconds int64  `json:"expire_seconds"` // 该二维码有效时间，以秒为单位。 最大不超过2592000（即30天）。
	URL           string `json:"url"`            // 二维码图片解析后的地址，开发者可根据该地址自行生成需要的二维码图片
}

type Qrcode struct {
	req *core.Request
}

func NewQrcode(req *core.Request) *Qrcode {
	return &Qrcode{req: req}
}

// Create 生成带参数的二维码
func (q *Qrcode) Create(ctx context.Context, qrCode *QRCodeRequest, accessToken string) (*QRCodeResponse, error) {
	// 验证参数
	if qrCode == nil {
		return nil, fmt.Errorf("二维码请求不能为空")
	}
	if qrCode.ActionName == "" {
		return nil, fmt.Errorf("二维码动作名称不能为空")
	}

	apiURL := fmt.Sprintf("%s?access_token=%s", APIQRCodeCreateURL, url.QueryEscape(accessToken))

	var result QRCodeResponse
	err := q.req.Make(ctx, "POST", apiURL, qrCode, &result)
	if err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result.APIResponse
	}

	return &result, nil
}

// ShowQrcode 通过ticket换取二维码
// ticket正确情况下，http 返回码是200，是一张图片，可以直接展示或者下载
func (q *Qrcode) ShowQrcode(ticket string) (string, error) {
	// 验证参数
	if ticket == "" {
		return "", fmt.Errorf("二维码ticket不能为空")
	}

	return fmt.Sprintf("https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=%s", url.QueryEscape(ticket)), nil
}
