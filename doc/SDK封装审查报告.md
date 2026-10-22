# WeGo 微信 SDK 封装审查报告

> 审查时间：2026-10-22
> 审查依据：微信官方开放文档（公众号 / 开放平台 / 小程序 / 微信支付 APIv3）
> 审查范围：`core`、`official_account`、`openplatform`、`miniprogram`、`message`、`crypto`、`storage`、`payment`、`wego.go`、`README.md`

---

## 零、修复进度（2026-10-22 更新）

第一批 P0（8 项）**已全部修复并验证**，`go build ./...` 与 `go vet ./...` 通过，已有测试套件全绿。

| # | P0 问题 | 状态 | 主要改动 |
|---|---|---|---|
| 1 | 网页授权 GET 传 Body + 无 json tag | ✅ 已修复 | `types.go` 四个 OAuth 请求结构体补 `appid/secret/code/grant_type/refresh_token/access_token/openid/lang` tag；`oauth_client.go` 三处 `Body` → `Query` |
| 2 | 客服消息多嵌套一层 msgtype | ✅ 已修复 | 新增 `CustomTextContent` 等 9 个内容体类型；`SendCustomMessageRequest` 字段改为内容体指针；`MessageXxx` 复用内容体类型 |
| 3 | 群发/预览请求体结构错误 | ✅ 已修复 | 新增 `MassContent` 接口与 6 个类型化内容体；`MassContentFields` 嵌入三个请求体；新增 `NewMassSendRequest/NewPreviewRequest/NewSendAllRequest` |
| 4 | `QueryBlockTmplMsg` 打到 get_industry | ✅ 已修复 | 新增常量 `URLQueryBlockTmplMsg = /cgi-bin/message/template/query_block_tmpl_msg` |
| 5 | 刷新网页授权 token 用错接口 | ✅ 已修复 | 新增 `URLSnsOAuth2ComponentRefreshToken`；`URLRefreshComponentAccessToken` 标记 Deprecated；顺带修复两处 token 明文日志 |
| 6 | clear_quota / quota 请求体字段错误 | ✅ 已修复 | body 统一为 `{"appid":...}`；`ClearComponentQuota` 改用 `/cgi-bin/component/clear_quota/v2` 且字段为 `appsecret`；新增 `ClearAuthorizerQuota` / `ClearAuthorizerQuotaBySecret` |
| 7 | message 模块占位实现 + map 无锁 | ✅ 已修复 | 新增 `CryptoConfigProvider` 与 `WithCryptoConfigProvider/WithDefaultCryptoConfig` 选项；`cryptoCache` 加 `sync.RWMutex` + 双重检查；新增 `ClearCryptoCache` |
| 8 | 仓库缺 go.sum | ✅ 已修复 | 已执行 `go mod tidy` 本地生成 `go.sum`；因 `.gitignore` 显式忽略 `/go.sum`（库项目约定：依赖校验由消费方自行生成），该文件不随仓库分发 |

**破坏性变更提示**（P0 修复引入，需在发布说明中注明）：

1. `MessageClient.MassSend/Preview/SendAll` 的 `(msgType string, content interface{})` 参数合并为 `content MassContent`
2. `SendCustomMessageRequest` 各字段类型由 `MessageXxx` 改为 `CustomXxxContent`
3. `openplatform.ClearQuotaRequest/GetApiQuotaRequest` 字段由 `ComponentAppID` 改为 `AppID`
4. `ClearComponentQuotaRequest.ComponentAppSecret` 改为 `AppSecret`，并新增可选 `AppID`
5. `NewSecureMessageProcessor()` 改为可变参数（原无参调用仍兼容）

---

## 一、总体结论

| 维度 | 评价 |
|---|---|
| payment 模块 | 质量最高：签名/验签、错误码、证书缓存、E 版构造函数、注释规范均已对齐官方 APIv3 规范，可作为全库标杆 |
| miniprogram 模块 | 较好：带 token 互斥锁 + 5 分钟过期余量，结构清晰，但能力覆盖度低 |
| storage 模块 | 较好：文件存储有 RWMutex + 临时文件原子替换，多后端抽象干净 |
| **official_account 模块** | **问题最集中**：存在多处「请求一定失败」的 P0 缺陷，且多为早期代码未随官方接口演进 |
| **openplatform 模块** | **次严重**：URL/参数名/请求体字段多处与官方不符，且存在 token 明文打日志的安全问题 |
| **message 模块** | **未完成**：`getCryptoInstance` 是硬编码占位实现，且 map 无锁并发读写 |
| 工程化 | 仓库缺 `go.sum`，当前无法 `go build` / `go vet` |
| 文档 | README 与代码存在多处实质性不一致（类型名、函数名、行为描述均对不上） |

**一句话总结**：支付链路可信，公众号/开放平台链路「看起来能跑，实际大量接口调不通」。优先修 P0，再补能力覆盖。

---

## 二、P0 —— 接口必然失败或数据错误（必修）

### 1. 网页授权：GET 请求传 JSON Body，且请求结构体无 json tag

- **文件**：`official_account/oauth_client.go:81-87`、`105-123`、`194-205`
- **文件**：`official_account/types.go:19-24`、`37-42`、`44-49`、`70-74`
- **根因**：
  - `GetAccessToken` / `RefreshAccessToken` / `ValidateAccessToken` 用 `Method: "GET"` 却把参数塞进 `Body`，`core/Request.Make` 会 `json.Marshal` 后写入请求体并加 `Content-Type: application/json`；而官方 `sns/*` 系列只接受 **Query String**。
  - `OAuthAccessTokenRequest` 等结构体**完全没有 json tag**，`json.Marshal` 产出 `{"AppID":"..","Secret":"..","Code":"..","GrantType":".."}`，与官方要求的 `appid/secret/code/grant_type` 完全不匹配。
- **影响面**：网页授权全链路不可用，必然返回 `41002 appid missing` / `40002 invalid grant_type`。
- **修复**：结构体补 `json:"appid"` 等 tag（或改为 `Query` 传参），并把 `Body` 改为 `Query`。
- 参考：`GetUserInfo`（`oauth_client.go:155-167`）虽用了 `Query`，但同样缺 tag，字段名会被转成 `AccessToken/OpenID/Lang`，同样错误。

### 2. 客服消息：请求体嵌套了一层多余的 msgtype

- **文件**：`official_account/custom_client.go:154-167`、`338-365`
- **根因**：`SendCustomMessageRequest` 的 `Text *MessageText` 字段指向的 `MessageText` 自身带 `MsgType string json:"msgtype"` 字段，导致实际序列化为：
  ```json
  {"touser":"oXX","msgtype":"text","text":{"msgtype":"text","text":{"content":"hello"}}}
  ```
  官方要求为 `{"touser":"oXX","msgtype":"text","text":{"content":"hello"}}`。
- **影响面**：所有客服消息发送失败（`40003 invalid openid` 或参数解析异常）。
- **修复**：`SendCustomMessageRequest` 的字段类型改为「内容体」结构（如 `Text *struct{ Content string }`），与 `msgtype` 平级。

### 3. 群发消息：请求体结构与官方完全不同

- **文件**：`official_account/message_client.go:64-101`
- **根因**：`MassSendRequest` / `SendAllRequest` / `PreviewRequest` 用了 `Content interface{} json:"content"`；官方格式是**按消息类型命名的键**：
  ```json
  {"filter":{"is_to_all":false,"tag_id":2},"text":{"content":"..."},"msgtype":"text"}
  ```
- **影响面**：群发、预览接口全部不可用。
- **修复**：改为 `map[string]interface{}` 动态组装，或为每种类型定义独立请求结构。

### 4. 模板消息「查询拦截消息」打到了错误的接口

- **文件**：`official_account/template_client.go:167`
- **根因**：`QueryBlockTmplMsg` 用的是 `URLTemplateGetIndustry`（`/cgi-bin/template/get_industry`），官方接口应为 `/cgi-bin/message/template/query_block_tmpl_msg`，`constants.go` 中也未定义该 URL。
- **影响面**：调用必然返回非预期结果（拿到行业信息结构，解析为空）。
- **修复**：新增 `URLQueryBlockTmplMsg = BaseAPIURL + "/cgi-bin/message/template/query_block_tmpl_msg"` 并替换。

### 5. 开放平台：刷新网页授权 token 用错接口

- **文件**：`openplatform/auth_client.go:1326` + `openplatform/constants.go:46`
- **根因**：`OAuthClient.RefreshToken` 用的是 `URLRefreshComponentAccessToken = /cgi-bin/component/api_authorizer_token`（这是「刷新授权方 access_token」接口），官方刷新**网页授权** token 应为 `/sns/oauth2/component/refresh_token`。
- **影响面**：代公众号网页授权的 token 刷新必然失败。
- **修复**：`constants.go` 中补 `URLSnsOAuth2ComponentRefreshToken`，`auth_client.go` 改引用。

### 6. 第三方平台 clear_quota / quota 请求体字段与官方不符

- **文件**：`openplatform/types.go:126-134`、`173-177`；`openplatform/client.go:951-978`、`1041-1064`
- **根因**：官方 `POST /cgi-bin/clear_quota?access_token=TOKEN` 的请求体只有一个字段 **`{"appid":"..."}`**（已核对官方文档，query 参数名确为 `access_token`，token 类型按场景区分，此处用法正确）。当前实现：
  - `ClearQuotaRequest` 发的是 `{"component_appid":"..."}`；
  - `GetApiQuotaRequest` 发的是 `{"component_appid":..,"authorizer_appid":..}`（官方 `/cgi-bin/openapi/quota/get` 同样只要 `appid`）；
  - `ClearComponentQuota` 打到 `/cgi-bin/component/clear_quota`，该路径官方不存在，且 body 传 `component_appid/component_appsecret`。
- **影响面**：三个配额相关接口全部无效。
- **修复**：统一改为 `{"appid": <目标账号appid>}`，并删除或修正 `ClearComponentQuota` 的 URL。

### 7. message 模块是未完成实现 + 并发数据竞争

- **文件**：`message/client.go:127-141`
- **根因**：`getCryptoInstance` 里 `token := "your_token"`、`encodingAESKey := "your_encoding_aes_key"` 是硬编码占位符；同时 `cryptoCache map[string]*crypto.WXBizMsgCrypt` 无任何同步保护。
- **影响面**：`SecureMessageProcessor` 完全不可用；即使补上配置，多协程并发收消息会触发 `fatal error: concurrent map writes` 直接崩进程。
- **修复**：从配置/存储注入 token 与 AESKey；改用 `sync.Map` 或 `sync.RWMutex` 保护缓存。

### 8. 仓库缺 go.sum，代码无法编译

- **根因**：根目录无 `go.sum`，`go vet ./...` 报 `missing go.sum entry`。
- **影响面**：无法编译、无法跑测试、无法 CI。
- **修复**：执行 `go mod tidy` 本地生成 `go.sum`。注：本仓库 `.gitignore` 显式忽略 `/go.sum`（库项目约定：依赖校验由消费方按其依赖图自行生成），故 `go.sum` 不随仓库提交，CI 若需复现构建应在流水线中执行 `go mod tidy`。

---

## 三、P1 —— 功能不完整 / 与官方规范偏差（强烈建议修）

| # | 问题 | 位置 | 说明 |
|---|---|---|---|
| 9 | 稳定版 access_token 实际未缓存 | `official_account/stable_token_client.go:79-95` | `getValidStableAccessToken` 恒返回 `nil,nil`，导致每次调用都请求微信。官方虽限 1 万次/分钟、50 万次/天，但不缓存会白白增加 RT 且逼近日限额；README 却声称「支持持久化存储」。另外官方明确：**强制刷新模式每天限 20 次且需间隔 30 秒**，当前代码无任何保护 |
| 10 | access_token 无并发保护、无过期余量 | `official_account/client.go:112-190` | `GetAccessToken`/`refreshAccessToken` 既无互斥锁也无提前刷新余量，高并发下会并发刷新互相覆盖（微信侧会返回 40001）。`miniprogram/client.go:129-199` 已实现 `tokenMu` + 5 分钟余量，应对齐 |
| 11 | 库内 `log.Panic` | `official_account/client.go:36-39`、`miniprogram/client.go:39-44`、`openplatform/client.go:42-46` | 依赖库不应终止宿主进程。`payment` 已有 `NewClientE` 返回 error 的范式，应统一推广 |
| 12 | `wego.go` 用 `panic` 报错 | `wego.go:267,275,283,291,299,307,314,323,332,342,351,363,374` | 同上，建议返回 `(*XxxClient, error)` 或提供 `GetXxx() (client, bool)` |
| 13 | token 明文写入日志 | `openplatform/auth_client.go:1006`、`1323` | `logger.Info("获取到ComponentAccessToken: %s", componentToken.AccessToken)` 是凭证泄漏，必须脱敏（同文件 1022 行已有 `[REDACTED]` 的正确写法） |
| 14 | 硬编码 URL 绕过常量 | `official_account/subscribe_client.go:134`、`163` | 直接写死 `https://api.weixin.qq.com/cgi-bin/template/get_category` 与 `get_pub_template_titles`，与 `constants.go` 中已定义的 `URLGetCategory`(`/wxaapi/newtmpl/getcategory`)、`URLGetPubTemplateTitles` 不一致，且前者路径本身就不是订阅消息接口 |
| 15 | 订阅消息字段类型错误 | `official_account/subscribe_client.go:114` | `Miniprogram *TemplateMessageMiniProgram json:"miniprogram_state,omitempty"` —— 官方 `miniprogram_state` 是 **string**（`developer`/`trial`/`formal`），不能是对象 |
| 16 | 公众号/小程序订阅消息体系混用 | `official_account/subscribe_client.go` 整体 | 公众号「订阅通知」应走 `/cgi-bin/message/subscribe/bizsend` + `/wxaapi/newtmpl/*` 的公众号版本；当前把小程序订阅消息接口塞进了 `official_account` 包，且 `SubscribeClient` 通过 `OfficialAccountSubscribe()` 暴露，语义混乱 |
| 17 | `CallAPI` 绕过注入的 HTTP 客户端 | `openplatform/auth_client.go:300` | 使用 `http.DefaultClient`，导致 `SetHTTPClient` 对该路径失效，自定义超时/代理/埋点全部丢失 |
| 18 | `CallAPIWithRetry` 的重试是空实现 | `openplatform/auth_client.go:410-430` | 注释写「Token过期，清除缓存并重试」，实际只 `continue`，既没清缓存也没刷新 token，重试无意义 |
| 19 | JSAPI Ticket 未缓存 | `openplatform/auth_client.go:1104-1138` | 每次生成 JSSDK 配置都请求一次 ticket，而 ticket 接口有独立调用配额。`JSSDKCacher` 只缓存了最终 config，没缓存 ticket |
| 20 | 大量早期/已下线接口常量 | `official_account/constants.go:113-245` | `merchant/*`（微信小店 V1）、`device/*`、`wifi/*`、`shakearound/*`、`semantic/*`、`printer/*`、`invoice/*` 等均为已下线或已迁移到微信支付/微信小店新 API 的历史接口，保留会误导使用者 |
| 21 | `core.Request.Make` 对 GET 也设置 Content-Type | `core/req.go:116-119` | 只要 body 非空就加 `application/json`，对 GET 属于多余且可能触发网关拦截；建议仅 POST/PUT/PATCH 设置 |
| 22 | `MakeRaw` 不校验状态码 | `core/req.go:209-215` | `UploadImage` / `UploadMaterial` / `AddMaterial` / `UploadMedia` 等走 `MakeRaw`，HTTP 500/302 不会被发现，错误被推迟到 JSON 解析阶段，报错信息失真 |
| 23 | `url.QueryEscape(accessToken)` 滥用 | `official_account/*` 全量 | access_token 只含安全字符，无需转义；统一用 `url.Values` 拼 query 更规范 |

---

## 四、P2 —— 一致性与可维护性

| # | 问题 | 位置 | 说明 |
|---|---|---|---|
| 24 | openplatform 反复 new Request | `openplatform/client.go:376,426,691,739,780,811,963,993,1024,1049,1082,1114,1140,1169,1196,1224` | 明明已有 `c.req` 成员，却每次 `core.NewRequest(c.httpClient, c.logger)`，无谓分配 |
| 25 | `TemplateMessageRequest.MiniProgram` 值类型 + omitempty | `official_account/api_template.go:28` | 结构体上 `omitempty` 无效，恒输出 `"miniprogram":{"appid":"","pagepath":""}`，应改为指针 |
| 26 | 事件解析失败一律吞异常返回 success | `openplatform/client.go:482,487,496,516,522` | 符合官方「必须返回 success」要求，但错误只打日志不落盘，线上排障困难；建议保留 error 供上层记录 |
| 27 | `AuthorizerInfo.MiniProgramInfo` 等字段大小写混用 | `openplatform/types.go:65,105-111` | 官方返回确为首字母大写（历史原因），但建议加注释说明，避免后人「修正」成下划线 |
| 28 | `wego.go` 注释块大段废弃代码 | `wego.go:389-415` | 全注释的 var 块无任何作用，建议删除 |
| 29 | `crypto.VerifySignature` 用 `==` 比较 | `crypto/client.go:237-240` | 建议改用 `hmac.Equal` / `subtle.ConstantTimeCompare` 防时序侧信道 |
| 30 | `miniprogram` 随机数用 `math/rand` | `openplatform/auth_client.go:1161-1179` | 生成 nonceStr 建议统一用 `crypto/rand`（`payment` 模块已有 `generateNonce` 的 crypto/rand 实现，可复用） |
| 31 | 各模块 opts 解析逻辑重复 | `official_account/client.go:59-74`、`miniprogram/client.go:64-73`、`payment/client.go:153-162`、`openplatform/client.go:59-77` | 四份几乎相同的 type-switch，建议抽到 `core` 统一处理 |

---

## 五、官方已提供、当前未封装的能力（待补齐）

### official_account
- 用户管理：`/cgi-bin/user/info`、`user/info/batchget`、`tags/*`（URL 已定义但**没有任何 client 方法**，属「有常量无实现」）
- 二维码：`URLQRCodeCreate` / `URLQRCodeShow` 常量已定义，但 `api_qrcode.go` 仅被 openplatform 间接使用，公众号侧无独立入口
- 短链接 `URLShortURL`、JSSDK ticket `URLGetTicket`（公众号侧无封装）
- 数据分析 `datacube/*`：常量已定义，无实现
- **新版客服消息**：`message/custom/typing` 已有；缺少 `kfaccount` 头像上传（当前只发 JSON，未做 multipart）
- 草稿箱/素材：`material_client.go` 已实现较完整，但 `GetMaterial` 对视频类返回 `{"video_url":..}` 的处理缺失（会把 JSON 当二进制返回）

### openplatform
- 代小程序：代码管理（`code/commit`、`get_qrcode`、`get_category`、`submit_audit`、`get_auditstatus`、`release`、`revertcoderelease`）全部缺失
- 小程序服务域/业务域设置：`modify_wxa_server_domain` / `modify_wxa_jump_domain` 常量已有、无实现
- `api_start_push_ticket`（`URLStartPushTicket`）无实现
- 代公众号网页授权：`wego.go:329-336` 的 `OpenPlatformOAuth` 传了空 `authorizerAppID`，是半成品

### miniprogram
- 登录：`code2Session` 已有
- 缺失：`getAccessToken`（内部有）、**小程序码**（`getwxacode`/`getwxacodeunlimit`/`createwxaqrcode`）、**订阅消息**、**内容安全检测**（`msg_sec_check`、`media_check_async`）、**统一服务消息**、**云函数触发**等
- 未在 `wego.go` 暴露任何语义化方法（仅有 `MiniProgram()` 返回裸 client）

### payment
- 覆盖度最好，可选补：转账（`transfer/batches`）、商家券、消费卡、代金券、支付即服务、`combine` 合单支付

---

## 六、文档与代码不一致（README）

| README 描述 | 实际情况 |
|---|---|
| 「文件存储创建失败会自动回退到**内存存储**并记录警告日志」 | `official_account`/`miniprogram`/`openplatform` 的 `NewClient` 直接 `log.Panic`；`storage.NewFileStorage` 是回退到**系统临时目录**，不是内存 |
| 「稳定版 token 支持持久化存储」 | 完全未实现（`getValidStableAccessToken` 恒 nil） |
| 示例代码 `&wego.MPConfig{...}` | 导出的是 `OfficialAccountConfig`，`MPConfig` 不存在，示例无法编译 |
| 「通过 `NewWeGoWithStorage` 指定自定义存储」 | 实际函数名为 `NewWithStorage` |
| 「Go 1.23.0+ / jcbaseGo v0.13.6」 | `go.mod` 为 `go 1.26.0` / `jcbaseGo v0.16.3` |
| 导出类型清单未包含 miniprogram / payment 新增类型 | `wego.go:461-504` 已导出，README 未同步 |
| 「`GetStableTokenClient()` 可用」 | 可调用，但每次都请求微信，README 描述的「优先使用缓存」不成立 |

---

## 七、建议修复顺序（Todolist）

**第一批 —— 止血（P0）：已完成 ✅**
1. ~~修 `oauth_client` 的 GET/Body 问题 + 补齐 json tag~~
2. ~~修 `SendCustomMessageRequest` 结构嵌套~~
3. ~~修群发/预览请求体结构~~
4. ~~修 `QueryBlockTmplMsg` URL~~
5. ~~修 `RefreshToken` 接口地址~~
6. ~~修 clear_quota / quota 请求体~~
7. ~~重写 `message.getCryptoInstance`（注入配置 + 加锁）~~
8. ~~`go mod tidy` 生成 go.sum，恢复可编译~~

**验证方式**：临时探针测试（`zz_probe_test.go`，已按惯例删除）校验了客服消息/群发/预览的序列化结构、OAuth Query 参数名、配额接口请求体、加解密实例并发获取。

**第二批 —— 稳定性（P1）：待执行**
9. stable_token 增加内存/存储缓存 + force_refresh 限频（20 次/天、间隔 30s）
10. `official_account.Client` 补 token 互斥锁与过期余量（对齐 miniprogram）
11. 全库移除 `log.Panic`，统一 E 版构造函数
12. 清理 token 明文日志
13. 统一订阅消息接口体系与 `miniprogram_state` 类型
14. `CallAPI` 改用注入的 httpClient；`CallAPIWithRetry` 补真实刷新逻辑
15. JSAPI ticket 缓存

**第三批 —— 规范与清理（P2）**
16. 删除已下线接口常量
17. 抽离公共 opts 解析
18. 修正 README 全部不一致项
19. 补齐第五节的缺失能力（按业务优先级）

---

## 八、附：官方文档核对要点（本次引用）

- **getStableAccessToken**：`POST /cgi-bin/stable_token`；仅支持 POST；频率 1 万次/分钟、50 万次/天；`force_refresh` 每天限 20 次且需间隔 30 秒；普通模式有效期内重复调用不刷新；**不支持第三方平台调用**
- **getAccessToken**：`GET /cgi-bin/token`，参数走 Query；错误码 40001/40013/40125/40164 等
- **clearQuota**：`POST /cgi-bin/clear_quota?access_token=ACCESS_TOKEN`，body 仅 `{ "appid": "..." }`（要被清空的账号 appid）；query 参数名固定为 `access_token`，**但 token 类型需与账号类型匹配**：清第三方平台用 `component_access_token`，代公众号/小程序清用 `authorizer_access_token`；每月 10 次；错误码 48006 表示次数用尽
- **component clear_quota/v2**（使用 AppSecret 重置）：`POST /cgi-bin/component/clear_quota/v2`，**无 query 参数**，body 为 `{"component_appid":...,"appsecret":...}`，代授权方时追加 `"appid"`；字段名为 `appsecret` 而非 `component_appsecret`；与 clear_quota 共用每月 10 次机会
- **quota/get**：`POST /cgi-bin/openapi/quota/get?access_token=ACCESS_TOKEN`，body 同样仅 `{"appid":"..."}`
- **微信支付 APIv3**：`WECHATPAY2-SHA256-RSA2048` 请求签名、应答验签串 `timestamp\nnonce\nbody\n`、平台证书与微信支付公钥双模式（以 `Wechatpay-Serial` 为准）—— 本库实现与之一致，无需改动
