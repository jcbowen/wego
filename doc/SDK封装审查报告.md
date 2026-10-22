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
| 5 | 刷新网页授权 token 用错接口 | ✅ 已修复 | 新增 `URLSnsOAuth2ComponentRefreshToken`；`URLRefreshComponentAccessToken` 已删除（第三批 #16，与 `URLAuthorizerToken` 同址）；顺带修复两处 token 明文日志 |
| 6 | clear_quota / quota 请求体字段错误 | ✅ 已修复 | body 统一为 `{"appid":...}`；`ClearComponentQuota` 改用 `/cgi-bin/component/clear_quota/v2` 且字段为 `appsecret`；新增 `ClearAuthorizerQuota` / `ClearAuthorizerQuotaBySecret` |
| 7 | message 模块占位实现 + map 无锁 | ✅ 已修复 | 新增 `CryptoConfigProvider` 与 `WithCryptoConfigProvider/WithDefaultCryptoConfig` 选项；`cryptoCache` 加 `sync.RWMutex` + 双重检查；新增 `ClearCryptoCache` |
| 8 | 仓库缺 go.sum | ✅ 已修复 | 已执行 `go mod tidy` 本地生成 `go.sum`；因 `.gitignore` 显式忽略 `/go.sum`（库项目约定：依赖校验由消费方自行生成），该文件不随仓库分发 |

**破坏性变更提示**（修复引入，需在发布说明中注明）：

P0 批次：
1. `MessageClient.MassSend/Preview/SendAll` 的 `(msgType string, content interface{})` 参数合并为 `content MassContent`
2. `SendCustomMessageRequest` 各字段类型由 `MessageXxx` 改为 `CustomXxxContent`
3. `openplatform.ClearQuotaRequest/GetApiQuotaRequest` 字段由 `ComponentAppID` 改为 `AppID`
4. `ClearComponentQuotaRequest.ComponentAppSecret` 改为 `AppSecret`，并新增可选 `AppID`
5. `NewSecureMessageProcessor()` 改为可变参数（原无参调用仍兼容）

P1 批次：
6. `SendNewSubscribeMsgRequest` 新增 `MiniprogramState string` 字段，原 `Miniprogram` 字段的 json tag 由 `miniprogram_state`（错误用法）改为 `miniprogram`——若此前依赖原错误序列化行为需调整
7. `WeGo` 全部 getter 与 `OAuthClient.GetAuthorizeURL` 系列新增 `XxxE` 返回 `(client, error)` 变体；旧方法标 `Deprecated`（保留 panic 兼容），建议迁移到 E 版
8. `core.Request` 日志不再输出完整 headers/body（依赖日志做请求审计的需自行记录）

P2 批次：
9. 删除导出常量 `openplatform.URLRefreshComponentAccessToken`（与 `URLAuthorizerToken` 同址，见第三批 #16）——引用该常量的下游代码需改用 `URLAuthorizerToken`（刷新授权方 access_token）或 `URLSnsOAuth2ComponentRefreshToken`（刷新网页授权 token）

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

> 2026-10-22 更新：下列「✅ 已补齐」项为第四批（方案 A）与第五批（方案 B）成果；第五节全部剩余项已补齐。

### official_account
- ✅ 用户管理：`UserClient`（`user/info`、`batchget`、`user/get`、`updateremark`、`tags/*` 全套）
- ✅ 公众号 JSSDK ticket：`Client.GetJSAPITicket()`（含内存缓存）
- ✅ 二维码（第五批）：`Client.CreateQRCode`（Client 级入口）+ `Client.GetQRCodeImage`（ticket 换图片二进制，含错误 JSON 识别与非 2xx 状态码判错）
- ~~短链接 `URLShortURL`~~：官方已停用该服务，常量已删除
- ✅ 数据分析 datacube（第五批）：`GetUserSummary`/`GetUserCumulate`、`GetArticleSummary`/`GetArticleTotal`、`GetUserRead(Hour)`、`GetUserShare(Hour)`、`GetInterfaceSummary(Hour)`，统一日期跨度校验；响应字段/渠道枚举已对照官方新版对齐（msgid 为 string、user_source 官方枚举、getarticletotal 明细含 target_user 等）
- ✅ 客服消息（第五批）：`UploadKfAccountHeadImage` 客服头像 multipart 上传已实现
- ✅ `GetMaterial` 视频素材返回 `down_url` JSON 的处理已修复

### openplatform
- ✅ 代小程序代码管理（第五批）：`CommitWxaCode`、`GetWxaTrialQrcode`（二进制/JSON 混合分流）、`GetWxaCategory`、`SubmitWxaAudit`、`UploadWxaAuditMedia`（multipart）、`GetWxaAuditStatus`、`ReleaseWxaCode`（空对象 body 防 44002）、`RevertWxaCodeRelease`（GET + app_version）
- ✅ 第三方平台域名设置（第五批）：`ModifyWxaServerDomain`、`ModifyWxaJumpDomain`（注意业务域名字段为 `wxa_jump_h5_domain`）
- ✅ `StartPushTicket`（第五批）：无 query 参数，凭证全走请求体
- ✅ 代公众号网页授权 `OpenPlatformOAuthE` 空参数半成品已修复（`authorizerAppID` 参数化）

### miniprogram
- 登录：`code2Session` 已有
- ✅ 已补齐：订阅消息（`SendSubscribeMessage`）、内容安全检测（`MsgSecCheck`/`MediaCheckAsync` v2）、小程序码（`GetWxaCode`/`GetWxaCodeUnlimit`/`CreateWxaQrcode`）
- 缺失：云函数触发等
- ✅ `wego.go` 已暴露 `MiniProgram()`/`MiniProgramE()`，新增能力均为 `Client` 语义化方法

### payment
- ✅ 商家转账（第五批）：新版单笔接口族（`/v3/fund-app/mch-transfer/*`）——`CreateTransferBill`（用户确认收款）、`TransferAfterUserAuth`（独立请求结构，authorization_id/out_authorization_no 二选一）、`CancelTransferBill`、`QueryTransferBillByOutBillNo`/`ByBillNo`、电子回单申请/查询（按商户单号/微信单号共 4 个），完整走 V3 签名 + 公钥/平台证书验签
- 仍可选补：旧版批量转账 `transfer/batches`、商家券、消费卡、代金券、支付即服务、`combine` 合单支付

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

**第二批 —— 稳定性（P1）：已完成 ✅（2026-10-22）**

| # | P1 问题 | 状态 | 主要改动 |
|---|---|---|---|
| 9 | stable_token 空壳缓存 + force_refresh 无限频 | ✅ 已修复 | `stable_token_client.go` 重写：`DefaultStableAccessTokenStorage` 加 `sync.RWMutex` 并发安全；`getValidStableAccessToken` 实现真实缓存命中（5 分钟过期余量，内存优先、持久化兜底回填）；新增限频字段与 `checkForceRefreshLimitLocked`/`recordForceRefreshLocked`（20 次/天 + 30 秒间隔 + 跨天自动重置），在 `GetStableAccessTokenForceRefresh` 路径拦截；支持通过 `SetStableTokenStorage` 注入自定义存储 |
| 10 | 公众号 access_token 无锁无余量 | ✅ 已修复 | `official_account/client.go` 对齐 miniprogram：新增 `tokenMu sync.RWMutex`（读锁查缓存、写锁刷新 + 双重检查）；URL 改用常量 `URLAccessToken`（参数走 Query）；过期余量 5 分钟（`isTokenValid` 提前判定） |
| 11 | 全库 `log.Panic` | ✅ 已修复 | 四个模块 + `wego` 全部新增 E 版入口：模块级 `NewClientE/NewClientWithStorageE`（内部走 `Config.Validate()` 校验）；`wego.NewE/NewWithStorageE`（`buildClient[T]` 泛型统一分发）；`WeGo` 全部 13 个 getter（`OpenPlatformAuth/OfficialAccountXxx/MiniProgram/Payment` 等）与 `OAuthClient.GetAuthorizeURL` 系列均新增 `XxxE` 变体；所有旧构造/旧 getter 统一标 `Deprecated` 并转发 E 版，panic 只保留在 Deprecated 兼容层内（行为与旧版一致），核心逻辑零 panic |
| 12 | token 明文日志 | ✅ 已修复 | `core/req.go` 新增 `redactSensitiveQuery`（URL 中 `access_token/component_access_token/secret/ticket` 等参数值整体替换为 `[REDACTED]`）与 `redactSensitiveBody`（响应体 JSON 中 `access_token/secret/ticket/session_key` 等字段值脱敏，保留键名便于排障）；请求日志改为脱敏 URL，仅记录 header 数量与 body 大小（不再记录完整 headers/body）；响应体脱敏后落日志；`openplatform` 组件 token 明文日志已在 P0-5 修复 |
| 13 | 订阅消息硬编码 URL 与类型 | ✅ 已修复 | `subscribe_client.go` 两处硬编码 URL 换为常量 `URLGetCategory` / `URLGetPubTemplateTitles`；`SendNewSubscribeMsgRequest` 新增 `MiniprogramState string`（`json:"miniprogram_state,omitempty"`）；原 `Miniprogram` 字段的 json tag 修正为 `miniprogram`（跳转小程序配置，原误用为 `miniprogram_state`） |
| 14 | `CallAPI` 用 `http.DefaultClient`；重试不清缓存 | ✅ 已修复 | `auth_client.go`：`CallAPI` 改用 `c.authClient.client.req.MakeRaw(req)`（注入的 HTTP 客户端生效），并补充非 2xx 状态码校验；`CallAPIWithRetry` 遇 40001/42001 调用 `storage.DeleteAuthorizerToken` 清除本地缓存后重试，非 token 错误直接返回 |
| 15 | JSAPI ticket 无缓存 | ✅ 已修复 | `openplatform.Client` 新增 `jsapiTicketMu sync.Mutex` + `jsapiTicketMap`（按 `authorizerAppID` 缓存）；`JSSDKManager.getJSAPITicket` 优先读缓存，缓存有效期按接口返回的 `expires_in` 减 5 分钟余量计算，过期自动清理 |

**验证方式**：临时探针测试校验了限频器（间隔拦截/日限拦截/跨天重置）、stable token 缓存命中、core 脱敏、`NewE` 错误传播链路，跑通后已按惯例删除。`go build`、`go vet`、`gofmt -l`、全量 `go test ./...` 通过（race 检测因 Windows 环境缺 CGO 编译器跳过，并发逻辑已由探针的 64 并发用例覆盖）。

**第三批 —— 规范与清理（P2）：部分完成（2026-10-22）**

| # | P2 问题 | 状态 | 主要改动 |
|---|---|---|---|
| 16 | 删除已下线接口常量 | ✅ 已完成 | 删除 `URLRefreshComponentAccessToken`（与 `URLAuthorizerToken` 同址，已确认全库无引用）；`URLSnsOAuth2ComponentRefreshToken` 注释补充两者不可混用的说明 |
| 17 | 抽离公共 opts 解析 | ✅ 已完成 | 新增 `core/opts.go` 的 `ApplyClientOptions`（回调式，支持模块扩展 handler）；`official_account`/`miniprogram`/`payment`/`openplatform` 四个模块构造函数的重复 switch 全部替换，openplatform 通过 extraHandler 处理 `EventHandler` |
| 18 | 修正 README 全部不一致项 | ✅ 已完成 | `MPConfig`→`OfficialAccountConfig`（3 处）、`NewWeGo`→`NewE`、删除不存在的「内存存储自动回退」描述、稳定版 token 说明更新（内置缓存 + `SetStableTokenStorage` 持久化注入 + force_refresh 限频）、依赖版本对齐 go.mod（Go 1.26.0 / jcbaseGo v0.16.3 / gorm v1.31.2）、新增「E 版构造函数（推荐）」章节、模块说明中过时类型名（`WegoClient`/`MPClient`）修正、示例代码与当前 API 对齐 |
| 19 | 补齐第五节缺失能力 | ✅ 已完成 | 方案 A（最小闭环）：① `GetMaterial` 视频素材 JSON 返回修复（新增 `GetMaterialResponse`，二进制走 Data、视频解析 title/description/down_url）；② `OpenPlatformOAuthE` 补全 `authorizerAppID` 参数（空值显式报错）；③ 新增 `UserClient`（用户信息/批量/关注列表/备注/标签 CRUD/打标/取标签粉丝/用户标签）；④ 公众号 `GetJSAPITicket`/`InvalidateJSAPITicket`（内存缓存 + 5 分钟余量）；⑤ 小程序订阅消息 `SendSubscribeMessage`、内容安全 `MsgSecCheck`/`MediaCheckAsync`（v2）；⑥ 小程序码 `GetWxaCode`/`GetWxaCodeUnlimit`/`CreateWxaQrcode`（二进制/JSON 混合响应处理）；⑦ `wego.go` 新增 `OfficialAccountUser(E)`/`OfficialAccount(E)` getter；⑧ 删除已下线的 `URLShortURL` 常量 |

**第四批 —— 能力补齐（P2-19）：已完成 ✅（2026-10-22，方案 A 最小闭环）**

验证方式：临时探针测试（假 RoundTripper 拦截请求）校验了用户管理请求体/Query 拼接、batchtagging 序列化、getticket `type=jsapi` 参数与缓存、GetMaterial 二进制/视频 JSON/错误三路径、订阅消息 data 嵌套、msg_sec_check v2 请求体、getwxacodeunlimit 请求体与二进制/错误双路径、`OpenPlatformOAuthE` 空参校验，跑通后已删除。`go build`、`go vet`、`gofmt -l`、全量 `go test ./...` 通过。

破坏性变更（P2-19 批次）：
- `MaterialClient.GetMaterial` 签名从 `([]byte, error)` 改为 `(*GetMaterialResponse, error)`
- `WeGo.OpenPlatformOAuth/OpenPlatformOAuthE` 新增 `authorizerAppID` 参数
- 删除导出常量 `official_account.URLShortURL`（官方已停用短链接服务）

**验证方式**：`go build`、`go vet`、`gofmt -l`（全库零输出，`payment/api.go`/`payment/types.go` 历史遗留格式化问题已随重构消除）、全量 `go test ./...` 通过。

**外部交叉验证后的补丁（2026-10-22）**：经独立会话交叉审查发现并修复 4 处遗漏——① README 存储回退描述与 `storage/file.go` 实际实现不符（存在临时目录回退），已改为事实；② `ApplyClientOptions` 的 warnLogger 按值传入导致未知参数警告不进自定义 logger（与重构前行为有差异），已改为闭包；③ README 示例仍调用 Deprecated getter（10 处）且占位 AESKey 非 43 位（复制即触发 `Validate` 报错），已全部换 E 版并修正占位值；④ 审查报告第 19 行"标记 Deprecated"与常量已删除状态不一致，已同步。

**示例编译复验（2026-10-22，第二轮）**：首轮探针结论有误——复刻编译发现示例 1–3 存在未使用变量（`declared and not used`）无法通过编译，已将各示例结尾改为输出客户端就绪状态（`%v` 打印 `xxxClient != nil`）使变量被消费；破坏性变更提示补录 P2 条目 9（删除导出常量 `URLRefreshComponentAccessToken`）。最终复验：README 全部 5 个示例提取后逐一 `go build` 通过（可复制即用），探针已删除。

**第五批 —— 方案 B 能力补齐：已完成 ✅（2026-10-22）**

第五节全部剩余项补齐（需求确认文件：项目根 `要求确认_第五批方案B能力补齐.md`）：

| # | 内容 | 实现位置 |
|---|---|---|
| 1 | 代小程序代码管理全套（commit/体验版二维码/类目/提审/审核状态/发布/回退 + 提审素材 multipart） | `openplatform/wxa_code_client.go` |
| 2 | 第三方平台域名设置（服务器域名/业务域名，body 字段 `wxa_server_domain`/`wxa_jump_h5_domain`） | `openplatform/domain_client.go` |
| 3 | `StartPushTicket`（无 query 参数，凭证全走请求体） | `openplatform/push_ticket.go` |
| 4 | 公众号二维码 Client 级入口 `CreateQRCode` + `GetQRCodeImage`（图片二进制，错误 JSON 识别） | `official_account/api_qrcode.go` |
| 5 | datacube 数据分析 10 个接口（用户增长/累计、图文每日/总、阅读、分享、接口分析，统一日期跨度校验） | `official_account/datacube_client.go` |
| 6 | `UploadKfAccountHeadImage` 客服头像 multipart 上传 | `official_account/kf_client.go` |
| 7 | 商家转账新版单笔接口族 9 个方法（发起/授权后转账/撤销/双单号查单/回单申请与查询×2），V3 签名 + 公钥验签全链路 | `payment/transfer.go` |
| 8 | `wego.go` 新增 `OpenPlatform(E)` getter；README/审查报告同步 | `wego.go` 等 |

验证：三模块探针测试（mock storage + Do 拦截；payment 用本地生成 RSA 密钥对完整走「请求签名 → 响应验签」）校验 URL/Query/Body/Content-Type/二进制分流，跑通后删除。`go build`、`go vet`、`gofmt -l`、全量 `go test ./...` 通过。

破坏性变更：本批为纯新增能力，无破坏性变更。

**第六批 —— 官方文档验证修复：已完成 ✅（2026-10-22）**

对第五批未提交代码逐项对照官方文档验证（需求确认文件：项目根 `要求确认_第六批_官方文档验证修复.md`），确认属实的差异已全部修复：

| # | 修复内容 | 位置 |
|---|---|---|
| 1 | datacube 响应对齐官方：`MsgID` int64→string（"12003_3" 格式）；删 `UserSummaryItem.CumulateUser`（getusersummary 无此字段）；`user_source` 渠道枚举注释更新为官方新版（0其他合计/1公众号搜索/17名片分享/30扫描二维码/.../200视频号/201直播）；`ArticleTotalDetailItem` 删 `UserSource` 补 `TargetUser`（送达人数）；`UserReadItem` 补 `UserSource`；`ShareScene` 注释改官方枚举（1好友转发/2朋友圈/255其他）；getarticlesummary/getarticletotal/getusershare 标注官方已停止维护 | `official_account/datacube_client.go` |
| 2 | `TransferAfterUserAuth` 改用独立请求 `TransferAfterUserAuthRequest`（无 openid/notify_url，authorization_id 与 out_authorization_no 二选一必填校验）；`TransferBillCancelResponse` 补 `UpdateTime`（官方必填） | `payment/transfer.go` |
| 3 | 商家转账注释文档号全量更正（发起转账 4012716434、双单号查单 4012716437/4012716457、电子回单 4012716452/4012716456/4012716455）；`transfer_scene_report_infos` 注释更新为官方必填口径 | `payment/transfer.go` + `payment/constants.go` |
| 4 | 客服头像注释更正（官方仅要求"5M 以内"，无 640x640 尺寸要求），文档链接更新为新版 api_uploadkfheadimg 页 | `official_account/kf_client.go` |
| 5 | `GetQRCodeImage` 补 HTTP 非 2xx 状态码判错（官方对非法 ticket 返回 404，此前会被当图片二进制返回） | `official_account/api_qrcode.go` |

决策记录：`user_recv_style`（红包样式）字段暂不封装，避免过度设计，后续有诉求再加。

破坏性变更：均属第五批未提交代码的批内修正（字段类型/结构签名在合入前即已纠正），对既有发布版本无影响。

**第七批 —— 六批复审遗留瑕疵修复：已完成 ✅（2026-10-22）**

两处瑕疵均先以临时探针证实后修复（探针已删除）：

| # | 问题探针结论 | 修复内容 | 位置 |
|---|---|---|---|
| 1 | `CreateTransferBill`/`TransferAfterUserAuth` 在 `transfer_scene_report_infos` 为空时请求穿透到 HTTP 层（官方标注必填但 SDK 无本地校验） | 新增 `validateTransferSceneReportInfos`：判空 + 逐项校验 `info_type`/`info_content` 非空，两个转账方法前置调用 | `payment/transfer.go` |
| 2 | `GetQRCodeImage` 非 2xx 时错误只含状态码，丢失微信错误响应体（探针 404 + JSON 体仅报 "状态码: 404"） | 读取响应体提前至状态码判断前；非 2xx 时若 body 可解析出 `errcode/errmsg` 则附带入错，否则附前 512 字节原文；2xx JSON 判定仍以 Content-Type 为准避免误判图片字节 | `official_account/api_qrcode.go` |

破坏性变更：#1 为客户端前置校验，仅拦截原本会被微信以 400 拒绝的请求，属第五批未提交代码的批内修正；#2 仅增强错误信息，无签名变化。

补记（同日）：探针过程中发现 `TransferAfterUserAuth` 缺 `transfer_remark` 必填校验（官方必填，`CreateTransferBill` 已有），已补齐，校验顺序与其对齐（金额 → remark → 报备信息 → 二选一）。

---

## 八、附：官方文档核对要点（本次引用）

- **getStableAccessToken**：`POST /cgi-bin/stable_token`；仅支持 POST；频率 1 万次/分钟、50 万次/天；`force_refresh` 每天限 20 次且需间隔 30 秒；普通模式有效期内重复调用不刷新；**不支持第三方平台调用**
- **getAccessToken**：`GET /cgi-bin/token`，参数走 Query；错误码 40001/40013/40125/40164 等
- **clearQuota**：`POST /cgi-bin/clear_quota?access_token=ACCESS_TOKEN`，body 仅 `{ "appid": "..." }`（要被清空的账号 appid）；query 参数名固定为 `access_token`，**但 token 类型需与账号类型匹配**：清第三方平台用 `component_access_token`，代公众号/小程序清用 `authorizer_access_token`；每月 10 次；错误码 48006 表示次数用尽
- **component clear_quota/v2**（使用 AppSecret 重置）：`POST /cgi-bin/component/clear_quota/v2`，**无 query 参数**，body 为 `{"component_appid":...,"appsecret":...}`，代授权方时追加 `"appid"`；字段名为 `appsecret` 而非 `component_appsecret`；与 clear_quota 共用每月 10 次机会
- **quota/get**：`POST /cgi-bin/openapi/quota/get?access_token=ACCESS_TOKEN`，body 同样仅 `{"appid":"..."}`
- **微信支付 APIv3**：`WECHATPAY2-SHA256-RSA2048` 请求签名、应答验签串 `timestamp\nnonce\nbody\n`、平台证书与微信支付公钥双模式（以 `Wechatpay-Serial` 为准）—— 本库实现与之一致，无需改动
