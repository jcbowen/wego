// Package official_account 微信公众号API - 客服账号管理（头像上传）
// 参考：https://developers.weixin.qq.com/doc/service/api/customer/servicermanage/api_uploadkfheadimg.html
package official_account

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/url"

	"github.com/jcbowen/wego/core"
)

// UploadKfAccountHeadImage 上传客服账号头像（multipart/form-data）
//
// 参数:
//   - ctx: 上下文
//   - kfAccount: 完整客服账号，格式为 "账号前缀@公众号微信号"（如 "test1@test"）
//   - filename: 文件名（建议带扩展名，如 "avatar.jpg"）
//   - imageData: 图片二进制内容（jpg/png，大小 5M 以内）
//
// 返回:
//   - *core.APIResponse: 接口响应
//   - error: 请求失败或微信返回错误时返回错误
//
// 使用示例:
//
//	img, _ := os.ReadFile("avatar.jpg")
//	_, err := client.UploadKfAccountHeadImage(ctx, "kf1@wxaccount", "avatar.jpg", img)
//
// 文档:
//   - https://developers.weixin.qq.com/doc/service/api/customer/servicermanage/api_uploadkfheadimg.html
func (c *Client) UploadKfAccountHeadImage(ctx context.Context, kfAccount, filename string, imageData []byte) (*core.APIResponse, error) {
	if kfAccount == "" {
		return nil, fmt.Errorf("客服账号不能为空")
	}
	if len(imageData) == 0 {
		return nil, fmt.Errorf("图片内容不能为空")
	}

	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	body, contentType, err := buildKfMultipartFileBody("media", filename, imageData)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s?access_token=%s&kf_account=%s",
		URLSetCustomAccountHeadImg, url.QueryEscape(token), url.QueryEscape(kfAccount))

	var result core.APIResponse
	err = c.req.Make(ctx, &core.ReqMakeOpt{
		Method: "POST",
		URL:    apiURL,
		Body:   body,
		Headers: map[string]string{
			"Content-Type": contentType,
		},
		Result: &result,
	})
	if err != nil {
		return nil, err
	}
	if !result.IsSuccess() {
		return nil, &result
	}
	return &result, nil
}

// buildKfMultipartFileBody 构造单文件 multipart/form-data 请求体
//
// 参数:
//   - fieldName: 表单字段名（微信素材类接口固定为 media）
//   - filename: 文件名
//   - content: 文件二进制内容
//
// 返回:
//   - []byte: 请求体内容
//   - string: 需要设置的 Content-Type（含 boundary）
//   - error: 构造失败时返回错误
func buildKfMultipartFileBody(fieldName, filename string, content []byte) ([]byte, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return nil, "", fmt.Errorf("创建multipart表单失败: %w", err)
	}
	if _, err := part.Write(content); err != nil {
		return nil, "", fmt.Errorf("写入multipart表单失败: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("关闭multipart writer失败: %w", err)
	}

	return buf.Bytes(), writer.FormDataContentType(), nil
}
