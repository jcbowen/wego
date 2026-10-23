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

第八批（P1/P2 遗留清尾）：
10. `TemplateMessageRequest.MiniProgram` 字段类型由值类型改为 `*TemplateMessageMiniProgram` 指针（配合 `omitempty`，nil 时不再序列化 `miniprogram` 键）——直接对 `request.MiniProgram.AppID` 赋值/取值的下游代码需改为指针写法或先判空
11. 删除 `official_account` 中 59 个已下线接口导出常量（device/wifi/shakearound/semantic/printer/invoice/merchant 等，见第七节第八批 #10）——引用这些常量的下游代码将编译失败，相关官方接口均已下线
12. 公众号订阅通知拆分为独立的 `SubscribeNotifyClient`（语义化方法名），旧 `SubscribeClient`/`GetSubscribeClient` 及其 `Wxa/New` 系列方法保留为兼容层并全部标 `Deprecated`，方法内部转发新实现，接口路径与序列化行为完全一致；建议迁移到 `GetSubscribeNotifyClient()`

> 2026-10-22 补充：经全量代码级复核，原报告仍有 12 项 P1/P2 问题未处理且未标记去向（P1 #16/#20/#21/#22/#23、P2 #24-#30），已集中登记为「第八批 —— 遗留清尾」（见第七节末尾）。
> 2026-10-22 追记：第八批 12 项已全部执行完毕（#10 用户决策直接删除、#11 用户决策采用拆分方案、#12 新增 `EventErrorCallback`），`go build`/`go vet`/`gofmt -l`/全量 `go test ./...` 通过，序列化改动经临时探针校验后删除。

---

## 一、总体结论

| 维度 | 评价 |
|---|---|
| payment 模块 | 质量最高：签名/验签、错误码、证书缓存、E 版构造函数、注释规范均已对齐官方 APIv3 规范，可作为全库标杆 |
| miniprogram 模块 | 较好：带 token 互斥锁 + 5 分钟过期余量，结构清晰，但能力覆盖度低 |
| storage 模块 | 较好：文件存储有 RWMutex + 临时文件原子替换，多后端抽象干净 |
| **official_account 模块** | **问题最集中**：存在多处「请求一定失败」的 P0 缺陷，且多为早期代码未随官方接口演进（审查时状态；P0~P8 批次已全部修复，能力也已补齐） |
| **openplatform 模块** | **次严重**：URL/参数名/请求体字段多处与官方不符，且存在 token 明文打日志的安全问题（审查时状态；已全部修复） |
| **message 模块** | **未完成**：`getCryptoInstance` 是硬编码占位实现，且 map 无锁并发读写（审查时状态；P0 #7 已修复：`CryptoConfigProvider` 注入 + `cryptoCache` 加锁，见修复进度表） |
| 工程化 | 审查时仓库缺 `go.sum` 无法构建，已修复（P0 #8，`go.sum` 因 `.gitignore` 约定不随库分发）；`go build`/`go vet`/`gofmt -l`/全量 `go test` 当前全绿 |
| 文档 | README 与代码存在多处实质性不一致（类型名、函数名、行为描述均对不上）（审查时状态；各批次已同步修正，最近一次为第八批同步订阅通知拆分与 EventErrorCallback） |

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

**第三批 —— 规范与清理：已完成 4 项（2026-10-22）；其余 P1/P2 项移交第八批**

| # | P2 问题 | 状态 | 主要改动 |
|---|---|---|---|
| 16 | 删除已下线接口常量 | ✅ 已完成 | 删除 `URLRefreshComponentAccessToken`（与 `URLAuthorizerToken` 同址，已确认全库无引用）；`URLSnsOAuth2ComponentRefreshToken` 注释补充两者不可混用的说明 |
| 17 | 抽离公共 opts 解析 | ✅ 已完成 | 新增 `core/opts.go` 的 `ApplyClientOptions`（回调式，支持模块扩展 handler）；`official_account`/`miniprogram`/`payment`/`openplatform` 四个模块构造函数的重复 switch 全部替换，openplatform 通过 extraHandler 处理 `EventHandler` |
| 18 | 修正 README 全部不一致项 | ✅ 已完成 | `MPConfig`→`OfficialAccountConfig`（3 处）、`NewWeGo`→`NewE`、删除不存在的「内存存储自动回退」描述、稳定版 token 说明更新（内置缓存 + `SetStableTokenStorage` 持久化注入 + force_refresh 限频）、依赖版本对齐 go.mod（Go 1.26.0 / jcbaseGo v0.16.3 / gorm v1.31.2）、新增「E 版构造函数（推荐）」章节、模块说明中过时类型名（`WegoClient`/`MPClient`）修正、示例代码与当前 API 对齐 |
| 19 | 补齐第五节缺失能力 | ✅ 已完成 | 方案 A（最小闭环）：① `GetMaterial` 视频素材 JSON 返回修复（新增 `GetMaterialResponse`，二进制走 Data、视频解析 title/description/down_url）；② `OpenPlatformOAuthE` 补全 `authorizerAppID` 参数（空值显式报错）；③ 新增 `UserClient`（用户信息/批量/关注列表/备注/标签 CRUD/打标/取标签粉丝/用户标签）；④ 公众号 `GetJSAPITicket`/`InvalidateJSAPITicket`（内存缓存 + 5 分钟余量）；⑤ 小程序订阅消息 `SendSubscribeMessage`、内容安全 `MsgSecCheck`/`MediaCheckAsync`（v2）；⑥ 小程序码 `GetWxaCode`/`GetWxaCodeUnlimit`/`CreateWxaQrcode`（二进制/JSON 混合响应处理）；⑦ `wego.go` 新增 `OfficialAccountUser(E)`/`OfficialAccount(E)` getter；⑧ 删除已下线的 `URLShortURL` 常量 |

> 编号说明：本批表内 #16-#19 为批次内序号，对应原问题分别为 P1 #20（已下线接口常量，本批仅删除 `URLRefreshComponentAccessToken` 一个常量，merchant/*、device/*、wifi/*、shakearound/*、semantic/*、printer/*、invoice/* 等其余已下线常量未清理）、P2 #31（opts 解析抽离）、第六节（README 一致性）、第五节（能力补齐）。原 P1 #16/#21/#22/#23 与 P2 #24-#30 未在本批处理，已移交「第八批 —— 遗留清尾」集中登记。

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

**第八批 —— P1/P2 遗留清尾：已完成 ✅（2026-10-22）**

> 2026-10-22 对全部批次做代码级复核：下列 12 项在 P0~P7 各批次中均未处理且此前未明确标记去向，现集中登记。其中 4 项需要决策取舍（#10/#11/#12 及可选能力），其余按既定方案执行。
> 2026-10-22 追记：经用户确认「全部执行」，#10 采用方案 a（直接删除）、#11 采用方案 a（拆分独立 Client）、#12 采用方案 a（新增错误回调），全部 12 项已完成，验证全部通过（`go build`/`go vet`/`gofmt -l`/全量 `go test ./...`，序列化改动经临时探针校验后删除）。
> 另：第五节保留的可选能力（miniprogram 云函数触发等、payment 旧版批量转账/商家券/消费卡/代金券/支付即服务/combine 合单支付）不属于缺陷修复，是否补齐由后续需求决定。

| # | 原问题编号 | 问题 | 核实位置 | 修复方案 | 风险/备注 |
|---|---|---|---|---|---|
| 1 | P1-21 | `Make` 对 GET 也设置 Content-Type | `core/req.go:145-147` | ✅ 仅 POST/PUT/PATCH 且 body 非空时设置 | 低风险 |
| 2 | P1-22 | `MakeRaw` 不校验状态码 | `core/req.go:240-246`；受影响调用方：official_account 各上传方法 | ✅ 新增公共 helper `core.CheckRawResponseStatus`，10 处 `MakeRaw` 调用点（上传/取素材/取二维码）对非 2xx 判错；`MakeRaw` 保持返回原始响应的语义 | 低风险 |
| 3 | P2-29 | `validateSignature` 用 `==` 比较 | `crypto/client.go:239` | ✅ 改用 `subtle.ConstantTimeCompare` 防时序侧信道 | 低风险 |
| 4 | P2-30 | nonceStr 用 `math/rand` | `openplatform/auth_client.go:13,1250-1251` | ✅ 改用 `crypto/rand` 生成 16 位字母数字（对齐 payment 的 generateNonce 实现），熵源失败回退时间戳 | 低风险 |
| 5 | P2-25 | `TemplateMessageRequest.MiniProgram` 值类型 + omitempty | `official_account/api_template.go:28` | ✅ 改为 `*TemplateMessageMiniProgram` 指针，空值不再序列化（经探针验证：nil 无 miniprogram 键、非 nil 正确输出） | 破坏性变更：字段类型变化 |
| 6 | P2-28 | wego.go 全注释 var 块 | `wego.go:604-629` | ✅ 删除废弃注释代码块；同批新增 `OfficialAccountSubscribeNotify/OfficialAccountSubscribeNotifyE` 便捷入口 | 无风险 |
| 7 | P1-23 | `url.QueryEscape(accessToken)` 滥用 | official_account/openplatform/miniprogram 全量 100+ 处 | ✅ 移除 token 类参数的 QueryEscape（共 14 个文件），非 token 参数转义全部保留，path 中 `PathEscape` 保留 | diff 大但无功能变化 |
| 8 | P2-24 | openplatform 反复 new Request | `openplatform/client.go` 共 13 处 `Make` 调用 | ✅ 全部替换为 `c.req`（`SetHTTPClient`/`SetLogger` 已同步重建 `c.req`）；另 3 处为构造/setter 合法初始化，保留 | 低风险 |
| 9 | P2-27 | `AuthorizerInfo` 字段大小写混用 | `openplatform/types.go:65,105-111` | ✅ 补充注释说明官方返回为首字母大写（历史原因），防止后人「修正」 | 无风险 |
| 10 | P1-20 | 已下线接口常量未清理 | `official_account/constants.go:113-245`：merchant/*、device/*、wifi/*、shakearound/*、semantic/*、printer/*、invoice/* | ✅ 删除 8 个分组 59 个常量（删除前 grep 确认全库零引用）；保留卡券/用户/数据分析/二维码等在用常量 | 破坏性变更：删除导出常量 |
| 11 | P1-16 | 公众号/小程序订阅消息体系混用 | `official_account/subscribe_client.go` 整体 | ✅ 拆分独立的 `SubscribeNotifyClient`（语义化方法名 GetNotifyCategory/GetPubTemplateTitles/... 等 8 个），旧 `SubscribeClient` 保留为兼容层转发新实现并全部标 `Deprecated`；接口路径与序列化行为零变化 | 破坏性变更：结构重构（兼容层保留） |
| 12 | P2-26 | 事件解析失败一律吞异常返回 success | `openplatform/client.go:512-552` | ✅ 新增可选 `EventErrorCallback`（`SetEventErrorCallback` 设置，不影响官方「返回 success」行为），13 个失败分支接入 `notifyEventError`（含 recover 保护），source 取值见 `openplatform/types.go` | 低风险 |

验证方式：`go build`、`go vet`、`gofmt -l`、全量 `go test ./...`；序列化类改动用临时探针校验请求体后删除。

---

## 八、附：官方文档核对要点（本次引用）

- **getStableAccessToken**：`POST /cgi-bin/stable_token`；仅支持 POST；频率 1 万次/分钟、50 万次/天；`force_refresh` 每天限 20 次且需间隔 30 秒；普通模式有效期内重复调用不刷新；**不支持第三方平台调用**
- **getAccessToken**：`GET /cgi-bin/token`，参数走 Query；错误码 40001/40013/40125/40164 等
- **clearQuota**：`POST /cgi-bin/clear_quota?access_token=ACCESS_TOKEN`，body 仅 `{ "appid": "..." }`（要被清空的账号 appid）；query 参数名固定为 `access_token`，**但 token 类型需与账号类型匹配**：清第三方平台用 `component_access_token`，代公众号/小程序清用 `authorizer_access_token`；每月 10 次；错误码 48006 表示次数用尽
- **component clear_quota/v2**（使用 AppSecret 重置）：`POST /cgi-bin/component/clear_quota/v2`，**无 query 参数**，body 为 `{"component_appid":...,"appsecret":...}`，代授权方时追加 `"appid"`；字段名为 `appsecret` 而非 `component_appsecret`；与 clear_quota 共用每月 10 次机会
- **quota/get**：`POST /cgi-bin/openapi/quota/get?access_token=ACCESS_TOKEN`，body 同样仅 `{"appid":"..."}`
- **微信支付 APIv3**：`WECHATPAY2-SHA256-RSA2048` 请求签名、应答验签串 `timestamp\nnonce\nbody\n`、平台证书与微信支付公钥双模式（以 `Wechatpay-Serial` 为准）—— 本库实现与之一致，无需改动

---

## 九、第九批 —— 官方文档全量验证（2026-10-23）

> 需求确认文件：项目根 `要求确认_第九批_官方文档全量验证.md`
> 核对基准：微信官方文档现行页面（公众号/开放平台走 `developers.weixin.qq.com/doc/service/**` 与 `/doc/oplatform/**` 服务端渲染版；小程序走 `/miniprogram/dev/server/API/**`；支付走 `pay.weixin.qq.com/doc/v3/merchant/**`）
> 处理方式：**只出报告，未改动任何代码**（P0 项待逐条确认后再修）
> 基线：`HEAD = 030b703`，`go build` / `go vet` / `gofmt -l` 全绿

**里程碑结论：前八批修复方向正确（客服消息内容体、群发按类型命名键、模板消息 miniprogram 指针、配额接口、datacube 字段均已复核通过），但本批仍发现 8 项 P0——其中 `crypto` 的 PKCS#7 填充问题是前八批从未触及的遗漏，且会导致消息加解密约半数失败。**

### 9.1 P0 —— 接口必然失败或数据错误

| # | 问题 | 位置 | 官方要求 | 当前实现 | 影响面 |
|---|---|---|---|---|---|
| 1 | **PKCS#7 填充块大小用了 16，官方为 32** | `crypto/aes.go:50`、`100` | 「PKCS#7：K 为秘钥字节数（**采用 32**）…KK….KK (K 个字节) if (N%K==0)」[技术方案](https://developers.weixin.qq.com/doc/oplatform/Third-party_Platforms/2.0/api/Before_Develop/Technical_Plan.html) | `blockSize := aes.BlockSize`（16）；去填充校验 `padLen > aes.BlockSize` | **探针实证：模拟微信服务端按 K=32 加密，本库解密失败率 52%**（100 次失败 52 次）。影响消息推送、授权事件、`component_verify_ticket` 全链路，且失败与消息长度相关，表现为「时好时坏」极难定位。加密侧 padLen∈[1,16] 落在服务端接受范围内，故单向可用 |
| 2 | 获取**永久**素材打到了**临时**素材路径 | `official_account/constants.go:81`、`material_client.go:260` | `POST /cgi-bin/material/get_material`，body `{media_id}` [文档](https://developers.weixin.qq.com/doc/service/api/material/permanent/api_getmaterial.html) | `URLGetMaterial = /cgi-bin/media/get` + POST + JSON body | 永久素材 100% 拿不到；该常量还被 `openplatform/auth_client.go:870` 复用（那边恰好 GET+query 歪打正着），常量归属本身错误 |
| 3 | 查询拦截模板消息接口整体错误 | `official_account/constants.go:58`、`template_client.go:34-54,166` | `POST /wxa/sec/queryblocktmplmsg`，body `{tmpl_msg_id, largest_id, limit}`，响应 `msginfo{id,tmpl_msg_id,title,content,send_timestamp,openid}` [文档](https://developers.weixin.qq.com/doc/service/api/notify/template/api_queryblocktmplmsg.html) | `/cgi-bin/message/template/query_block_tmpl_msg` + `begin_date/end_date/offset/count` + 响应 `total_count/list` | 接口完全不可用。**说明第六批那次「改成 query_block_tmpl_msg」的修复方向本身就是错的**（官方早已迁到 `/wxa/sec/`） |
| 4 | 自动回复规则响应字段名错（静默丢数据） | `official_account/message_client.go:615,619` | `add_friend_autoreply_info` / `message_default_autoreply_info` [文档](https://developers.weixin.qq.com/doc/service/api/notify/autoreplies/api_getcurrentautoreplyinfo.html) | `add_friend_reply_info` / `message_default_reply_info` | `errcode=0` 判定成功，但关注自动回复与消息自动回复两块**永远解析不到且不报错**，属静默数据丢失。（`keyword_autoreply_info` 及内部结构正确） |
| 5 | 网络通信检测响应类型错误 | `official_account/api_client.go:29-33` | `dns`/`ping` 均为 **objarray**：`dns[]{ip,real_operator}`、`ping[]{ip,from_operator,package_loss}` [文档](https://developers.weixin.qq.com/doc/service/api/base/api_callbackcheck.html) | `DNS []string` / `Ping []string` | 返回非空数组时 `json.Unmarshal` 直接报 `cannot unmarshal object into Go value of type string`，接口恒定报错 |
| 6 | 内容安全 `media_type` 常量值错误 | `miniprogram/subscribe_client.go:31-35` | **1:音频；2:图片**，错误码 40004 原文「目前支持多媒体类型 1 音频；2 图片」[文档](https://developers.weixin.qq.com/miniprogram/dev/server/API/sec-center/sec-check/api_mediacheckasync.html) | `Audio = 3`、`Image = 2`、另有官方不存在的 `Video = 4` | 调用方只要使用库自带常量提交音频，微信必然返回 40004 |
| 7 | `GetWXACode` 把图片二进制当 JSON 解析 | `openplatform/auth_client.go:1309-1351` | `POST /wxa/getwxacode` 成功时返回**图片二进制**（失败才是 JSON） | `var result WXACodeResponse`（json tag）+ `req.Make`，`core/req.go:233` 对非空响应体强制 JSON 反序列化 | 100% 报 `invalid character`。同包 `GetWxaTrialQrcode`、`official_account` 二维码接口都已用 `MakeRaw` + Content-Type 分流，此处是唯一漏网 |
| 8 | `CallAPIWithQuery` 二次拼接 `?access_token=` | `openplatform/auth_client.go:347,350` → `:276` | 代调用 `xxx?access_token=AUTHORIZER_ACCESS_TOKEN` | `fullURL` 已含 `?access_token=T1&type=xxx`，再传入 `CallAPI` 又拼一次 | 结果 `...?access_token=T1&type=xxx?access_token=T2`：有 query 时参数值被污染；**queryParams 为空时 access_token 值变成 `T1?access_token=T2`，必然 40001** |

### 9.2 P1 —— 偏离官方规范 / 功能不完整

**payment（5 项）**

| # | 问题 | 位置 |
|---|---|---|
| P1-1 | 商家转账加密 `user_name` 时未发 `Wechatpay-Serial`（官方标必填），平台证书模式下微信无法确认解密密钥；≥2000 元转账 `user_name` 强制必传 | `transfer.go:171,239`；对比 `profitsharing.go:71` 正确 |
| P1-2 | 请求分账 `Wechatpay-Serial` 仅条件发送，且未强制 `receivers.name`（MERCHANT_ID 时官方必传） | `profitsharing.go:61-68`、`types.go:370-376` |
| P1-3 | 退款不支持 `transaction_id`（官方与 `out_trade_no` 二选一），`amount.currency` 缺失；`api.go:421` 强制 `out_trade_no` | `types.go:287-294`、`api.go:417-445` |
| P1-4 | 非 2xx 应答不做验签（官方要求所有应答均验签，并会下发 `WECHATPAY/SIGNTEST/` 探测流量检验）→ 伪造失败响应可诱导重复退款 | `client.go:507-509` |
| P1-5 | `AppPayParams` 字段名为全小写（V2 遗留），官方 APIv3 为 `appId/partnerId/prepayId/package/nonceStr/timeStamp/sign`；签名串内容本身正确 | `types.go:236-244` |

**official_account（9 项）**

| # | 问题 | 位置 |
|---|---|---|
| P1-6 | 群发图片内容键应为 **`images`**（内含 `media_ids` 数组 + `recommend`/`title`/`need_open_comment`/`only_fans_can_comment`），`msgtype` 才是 `image`；官方示例已照抄确认 | `message_client.go:85-90,145-152` |
| P1-7 | 群发速度：官方为 **POST**；响应字段为 `realspeed`（无下划线），当前 `real_speed` 恒为 0 | `message_client.go:406`、`:49` |
| P1-8 | 客服会话查询 `getsession`/`getsessionlist` 官方为 **GET + Query**，实现为 POST + Body | `custom_client.go:663,692` |
| P1-9 | 删除个性化菜单 `menuid` 应在**请求体**（`{"menuid":"208379533"}`），实现走 Query | `menu_client.go:241` |
| P1-10 | 选用模板缺必填 `keyword_name_list`（官方：为空或不在模板库返回 40246） | `template_client.go:22-25,122-149` |
| P1-11 | 一次性订阅消息 `data` 官方为固定结构 `{"content":{"value":..,"color":..}}`，实现用 `map[string]` 导致键名任意 | `subscribe_notify_client.go:424` |
| P1-12 | `material/add_news`、`material/update_news` 在现行官方接口目录中已无此二接口（永久图文已被草稿箱 `/cgi-bin/draft/*` 取代） | `constants.go:83,86` |
| P1-13 | 永久视频素材上传缺官方 `description` 表单字段（`{"title":..,"introduction":..}`，注意是 introduction） | `material_client.go:550-622` |
| P1-14 | 网页授权链接多了官方不存在的 `connect_redirect=1`，且用 `url.Values.Encode()` 打乱参数顺序——官方明确「对授权链接做正则**强匹配**校验，参数顺序不对授权页无法正常访问」 | `oauth_client.go:47-57` |

**openplatform / miniprogram / 公共层（13 项）**

| # | 问题 | 位置 |
|---|---|---|
| P1-15 | 加密事件判定依赖 `encrypt_type` query 参数；若宿主网关未透传（常见），`component_verify_ticket` 被静默丢弃 → 所有接口在 ticket 过期后整体失效，日志仅一条 Warn | `client.go:536,568-574` |
| P1-16 | `EncodingAESKey` 变更事件：`SetPrevEncodingAESKey` 设在 `c.crypt`，但解密路径每次 `crypto.NewWXBizMsgCrypt` 重建 → PrevKey 恒空，属**死调用**；且回调中直接写 `c.config.EncodingAESKey` 有数据竞争 | `client.go:669-679,726` |
| P1-17 | `CallAPIWithRetry` 遇 40001 删除整条授权记录（含 refresh_token），重试时 refresh_token 已丢失 → 代调用永久不可用，只能重新授权 | `auth_client.go:419-443` |
| P1-18 | `GetRidInfo` 用 component_access_token 调公众号专用接口 `openapi/rid/get`（官方页 404，待复核） | `client.go:1129-1156` |
| P1-19 | `GetApiQuota` 空参时用 component_access_token 查自身额度（官方页 404，待复核） | `client.go:1086-1126` |
| P1-20 | `msg_sec_check` / `media_check_async` 的 `version` 官方为 **number**（示例 `"version": 2`），实现为 `string "2"` | `security_client.go:15,99` |
| P1-21 | `getwxacode` 的 `path` 官方**必填**（空返回 40159），实现用 `omitempty` 吞掉；注释「空则默认主页」是错的（该语义属 `getwxacodeunlimit` 的 `page`）；缺 `env_version` | `wxacode_client.go:25-31` |
| P1-22 | `getwxacodeunlimit` 的 `check_path` 官方**默认 true**，实现为 `bool` 无 `omitempty` → 零值发 `false`，与注释所述行为完全相反，未发布页面被静默放行 | `wxacode_client.go:46` |
| P1-23 | `createwxaqrcode` 的 `path` 官方必填（≤128 字符），实现用 `omitempty` | `wxacode_client.go:57-60` |
| P1-24 | 订阅消息 `miniprogram_state` / `lang` 无 `omitempty`，未设置时发空串（官方 lang 默认 zh_CN，另有 zh_HK/zh_TW 未定义常量） | `subscribe_client.go:50-51` |
| P1-25 | 密钥轮换回包未按官方「用哪个 AESKey 解密成功，就用此 AESKey 加密回包」处理；`message` 层实例未注入 storage → `PrevEncodingAESKey` 恒空，降级分支永不触发 | `crypto/client.go:187-201`、`message/client.go:200` |
| P1-26 | 脱敏正则未覆盖 `component_verify_ticket` / `authorizer_access_token` / `refresh_token`（键名精确匹配失败），且不处理 XML 形态 `<ComponentVerifyTicket>` | `core/req.go:28,39` |
| P1-27 | 非 2xx 的错误信息使用**未脱敏**原始响应体，与日志脱敏口径不一致，形成旁路泄露 | `core/req.go:208,276` |

### 9.3 P2 —— 一致性与字段完整性（择要）

- **payment**：4 个下单结构缺 `time_expire`/`support_fapiao`/`detail`/`settle_info`（`settle_info` 缺失会使「下单时声明分账」链路断裂，官方明确支付后无法补设）；H5 缺 `attach`；`SceneInfo` 缺 `device_id`/`store_info`；分账与退款查询多传 `mchid`；分账/回退应答缺 `create_time`/`finish_time`/`return_id`；撤销转账发空 `{}` body（官方无 body 参数）；回调成功应答体 `{"code":"SUCCESS"}` 非官方取值（官方仅 `FAIL`，成功应 204 空体）；`APIErrorDetail` 含非官方字段 `detail_list`；转账缺 `user_recv_style`、`fail_reason`；**10+ 处注释中的官方文档 doc id 失效或错位**
- **official_account**：`bizsend` data 多序列化 `name`；`getpubtemplatekeywords` 缺 `count`；`tid` 官方为 string；`mass/get` 的 `msg_id` 官方为 string（当前 int64，若返回字符串会反序列化失败）；`filter.is_to_all`/`tag_id` 官方为 string；`uploadnews`/预览缺 `clientmsgid`/`towxname`/`msg_data_id`；客服 `kf_id` 类型、`kf_list` 字段不全、缺 `customservice`/`mpnewsarticle`/`msgmenu` 类型；菜单缺 `article_id`；素材视频多出 `video_url`（属临时素材字段）；`getuserread` 的 `user_source` 枚举注释错误；5 个 `datacube/getupstreammsg*` 常量定义了但无方法实现；17 个 `/card/*` 卡券常量与 `/pay/unifiedorder`（V2 已下线）零引用；多组重复常量；稳定版 token 未提示「不支持第三方平台调用」
- **openplatform**：`config.Validate` 允许 EncodingAESKey 为空（第三方平台强制安全模式应必填 43 位）；事件 AppID 强校验 + 5 分钟时间戳窗口可能误杀合法事件（AppID 校验官方标「可选」）；component/authorizer token 无并发保护与过期余量（`miniprogram`/`official_account` 已对齐，仅此模块未对齐）；JS-SDK config 缓存按 7200s 算而非按 ticket 实际过期时间；`SendCustomMessage` 仅支持 text/image，`voice`/`video`/`music`/`news` 会产生非法请求体；死常量 `URLGetJumpDomainFile`/`URLGetUserInfo` 与死类型 `OAuthComponentAccessTokenRequest`；`AuthorizerInfo` 缺 `idc` 与部分网络域字段；pre_auth_code 复用缓存
- **miniprogram / 公共层**：`openid`/`scene` 必填未本地校验；`msgSecCheck` 请求体含官方未定义字段 `head_avatar`；响应 `detail` 缺 `keyword`，`label` 枚举注释漏 20013 且把 20012 注释错；`media_url` 注释「5MB」应为 10M；`getwxacode` 的 path 注释「不能带参数」已过时（官方示例带 query）；未暴露 `Num-Used` 响应头；`GetPhoneNumberResponse` 缺 `phone_info.watermark`；未提供稳定版 access_token；去填充未校验填充字节一致性；16 字节随机串取模偏置；`crypto/client.go:194` 错误包装取到被遮蔽的变量；`EventTimestampTolerance` 定义未使用；`WXBizMsgCrypt.DecryptMsg`/`VerifyURL` 无时间戳校验；`DecryptMiniProgramData` 注释与实现不符；`storage/file.go` 临时文件名固定（多进程互相踩踏）、未 `Sync()` 即 rename、权限 0644 明文落盘、Windows 失败路径残留 `.tmp`

### 9.4 已核对正确（复核通过）

- **payment 核心链路全部正确**：请求签名五行串、GET 空 body 处理、Authorization 头五字段、`serial_no` 用商户证书序列号、签名 URL 用 `EscapedPath()+RawQuery`、应答验签三行串、204 空 body 特例、时间戳 5 分钟双向校验、以 `Wechatpay-Serial` 前缀 `PUB_KEY_ID_` 判定公钥/平台证书模式、回调 AES-256-GCM（nonce 作 IV、associated_data 作 AAD）、敏感信息 RSAES-OAEP + SHA-1、平台证书下载跳过验签、账单下载跳过验签仍带签名；URL/方法/参数逐接口核对通过；未封装任何 V2 或已下线接口（付款码支付、撤销订单 V3 官方仍在维护）
- **official_account**：`token`/`stable_token`（含 force_refresh 20 次/天 + 30s 限频）、菜单 7 个接口、二维码、客服消息历史修复（内容体不含 msgtype）、`sendall` 按类型命名键的方向、模板消息 `miniprogram` 指针、订阅通知 7 个路径且 `miniprogram_state` 为 string 枚举、素材上传与草稿箱 6 个接口、用户管理与标签全套、网页授权 GET+Query 字段名、JSAPI ticket、datacube 15 个路径与跨度校验（第六批对齐结果复核通过）
- **openplatform**：`api_component_token`、`api_create_preauthcode`、`api_query_auth`、`api_authorizer_token`、`api_get_authorizer_info`、`api_start_push_ticket`（无 query、凭证走 body）、`clear_quota`、`component/clear_quota/v2`（`appsecret` 字段名正确）、代小程序 commit/体验版二维码/类目/提审/提审素材/审核状态/发布/回退、`modify_wxa_server_domain`/`modify_wxa_jump_domain`、代调用通用通道、`sns/oauth2/component/*`（第八批修复正确）、授权页 URL、JS-SDK 签名串与 ticket 缓存、事件 XML 结构、加解密算法本体
- **miniprogram**：`code2Session`、`getuserphonenumber`、`token`、订阅消息 send、三个小程序码 URL（v1/v2 **共用同一路径靠 body `version` 区分**，故 README 写 v2 而常量无 `/v2` 后缀**并不矛盾**）、二进制/JSON 分流、access_token 并发刷新
- **公共层**：GET/HEAD/DELETE 不加 Content-Type、Query 合并转义、AESKey 解码与 43 位校验、AES-256-CBC + IV 取前 16 字节、明文结构、网络字节序、签名 SHA1 字典序 + `ConstantTimeCompare`、`DecryptMiniProgramData`、message 回包复用请求 timestamp/nonce、`SecureMessageProcessor` 时间戳窗口与并发缓存

### 9.5 需人工复核（未臆断为错误）

1. 群发图片 `image` 与 `images` 微信是否双向兼容（旧文档为 `image/media_id`），建议真机验证后再决定保留兼容分支
2. `mass/get` 的 `msg_id`、`sendall` 的 `is_to_all`/`tag_id` 官方标 string，实际返回类型需抓包确认（若为字符串，`int64`/`bool` 会直接反序列化失败，P2 可能升级为 P0）
3. 授权链接 `connect_redirect=1` 是否真触发正则强匹配失败（需非测试号真机验证）
4. `/card/*` 17 个卡券常量：现行服务号 API 目录无卡券分类，但可能已迁独立文档，不能断定下线
5. `openapi/rid/get`、`openapi/quota/get` 是否支持 component_access_token（官方页 404）
6. `clear_quota/v2` 的 `appid`：官方参数表标必填，但示例 5.1 不传，二者矛盾
7. `version` 传字符串 `"2"` 是否被微信拒绝（若实测通过，P1-20 可降 P2）
8. `submit_audit` 回包 `auditid`：官方参数表写 String、示例写数字，当前 int64 可能解析失败
9. 退款 `amount.currency` 是否真必填（官方 SDK 模型中为可选）
10. 4xx/5xx 应答是否**一定**携带 `Wechatpay-*` 签名头（P1-4 修复需「头缺失时跳过」兜底）
11. `material/add_news`/`update_news` 是否已下线，建议用官方 API 诊断工具或直接调用确认返回码后再删
12. GZIP 账单的 `hash_value` 对应压缩前还是压缩后字节（官方语义未明确）

### 9.6 建议修复顺序

1. **P0-1**（PKCS#7 K=32）——唯一会造成「整站消息链路半数失败」的问题，且前八批从未触及，建议最先修并用官方 `cryptoDemo.zip` 示例数据做 round-trip 验证
2. **P0-2 / P0-3 / P0-4 / P0-5**（公众号 4 个必然失败/静默丢数据的接口）
3. **P0-6**（media_type 常量，改常量即可，成本极低）
4. **P0-7 / P0-8**（openplatform，前者可直接复用 `GetWxaTrialQrcode` 现成实现，后者改 query 拼接）
5. **P1-1 / P1-4**（支付资金安全：Wechatpay-Serial 与错误应答验签）
6. **P1-6 / P1-7 / P1-8 / P1-9**（公众号方法/参数位置与字段）
7. **P1-15 / P1-16 / P1-17**（开放平台加解密与代调用健壮性）+ **P1-25 / P1-26 / P1-27**（公共层安全）
8. 其余 P1 与 P2 随版本迭代清理

> 本批为**只读核查**，除临时探针（`crypto/zz_probe_test.go`，用于实证 P0-1，跑通后已删除）外，未修改、创建或删除任何文件。

---

## 十、第九批问题修复状态（2026-10-23 第十批执行）

> 依据：根目录《问题核实报告单_第九批.md》核实结论（P0 8 项属实、P1 25 项属实 + 2 项待复核、P2 择要属实）+ 官方文档二次抽查
> 用户决策：「全部处理，允许破坏性变更；无法真机验证的保守兼容处理」
> 需求确认文件：项目根 `要求确认_第十批_第九批核实问题修复.md`
> 验证：`go build ./...` / `go vet ./...` / `gofmt -l .` / `go test ./...` 全部通过

### 10.1 P0（8 项）全部修复

| # | 修复摘要 |
|---|---|
| P0-1 | `crypto/aes.go` PKCS#7 填充/去填充改用 K=32（`len(aesKey)`），去填充校验填充字节一致性；`generateRandomString` 拒绝采样修复取模偏置 |
| P0-2 | 永久素材接口改 `URLGetMaterialPermanent`（POST `/cgi-bin/material/get_material`），临时素材 `URLGetMaterial`（GET）保留；删除 `URLAddNews`/`URLUpdateNews`/`URLMaterialUploadImage` 及 `AddNews`/`UpdateMaterial` 方法 |
| P0-3 | 查询拦截模板消息改官方 `/wxa/sec/queryblocktmplmsg`，请求/响应结构与官方对齐 |
| P0-4 | 自动回恢复字段名 `add_friend_autoreply_info`/`message_default_autoreply_info` |
| P0-5 | 网络检测响应改为 `dns[]{ip,real_operator}`、`ping[]{ip,from_operator,package_loss}` 结构 |
| P0-6 | `media_type` 常量：音频 1、图片 2，删除官方不存在的 Video=4；补 `zh_HK`/`zh_TW` 语言常量 |
| P0-7 | `GetWXACode` 改 `MakeRaw` + Content-Type 分流（JSON 错误/图片二进制），`ContentType`/`Buffer` 改 `json:"-"` |
| P0-8 | `CallAPI` 检测 URL 已含 `?` 时用 `&` 追加；`CallAPIWithQuery` 不再自行注入 access_token，彻底消除二次拼接 |

### 10.2 P1（27 项）25 项修复、2 项待复核

| # | 修复摘要 |
|---|---|
| P1-1 | 商家转账加密字段时按需发送 `Wechatpay-Serial`（`CreateTransferBill`/`TransferAfterUserAuth`） |
| P1-3 | 退款支持 `transaction_id`（与 `out_trade_no` 二选一并本地校验）；`RefundAmount` 补 `currency` |
| P1-4 | 非 2xx 应答先验签再解析错误（带「签名头缺失时跳过」兜底） |
| P1-5 | `AppPayParams` 字段名改官方驼峰 `appId/partnerId/prepayId/...` |
| P1-6 | 群发图片新格式 `images`（`media_ids` 数组等），旧 `image` 格式标 `Deprecated`；`is_to_all` 用 `FlexBool` 双形态兼容 |
| P1-7 | 群发速度改 POST，响应字段 `realspeed` |
| P1-8 | 客服会话查询改 GET+Query |
| P1-9 | 删除个性化菜单 `menuid` 移入请求体 |
| P1-10 | 选用模板补必填 `keyword_name_list` |
| P1-11 | 一次性订阅 `data` 改官方固定结构 `{"content":{"value","color"}}` |
| P1-12 | 删除 `material/add_news`/`update_news` 常量（随 P0-2） |
| P1-13 | 新增 `AddMaterialVideo`（永久视频上传携带 `description` 表单字段 `{"title","introduction"}`）；`AddMaterial` 对 video 类型报错引导 |
| P1-14 | 授权链接删除 `connect_redirect=1`，参数按官方示例顺序拼接（不使用字典序 `Encode`） |
| P1-15 | 加密事件判定改为以 XML `Encrypt` 字段为准，不再依赖 `encrypt_type` query |
| P1-16 | `encoding_aes_key_changed` 改为将旧密钥持久化到 storage（`SavePrevEncodingAESKey`）；`config.EncodingAESKey` 读写加 `cryptMu` 互斥锁；删除死字段 `c.crypt` |
| P1-17 | 40001/42001 重试改为仅失效 access_token（保留 refresh_token），新增 `invalidateAuthorizerAccessToken` |
| P1-20 | `version` 改 number（int，固定 2） |
| P1-21 | `getwxacode` 的 `path` 去除 `omitempty`（官方必填，方法级校验）、补 `env_version`，修正注释 |
| P1-22 | `check_path` 改 `*bool` + `omitempty`（不传时官方默认 true） |
| P1-23 | `createwxaqrcode` 的 `path` 去除 `omitempty` 并校验 |
| P1-24 | 订阅消息 `miniprogram_state`/`lang` 加 `omitempty`；补 `zh_HK`/`zh_TW` |
| P1-25 | 新增 `DecryptMsgWithKey`/`EncryptMsgWithKey`（官方「用哪个 AESKey 解密成功就用它加密回包」）；`SecureMessageProcessor` 新增 `WithTokenStorage` 注入存储加载 PrevKey |
| P1-26 | 脱敏正则显式覆盖 `component_verify_ticket`/`authorizer_access_token`/`authorizer_refresh_token`/`refresh_token`/`session_key`，并新增 XML 形态脱敏 |
| P1-27 | 非 2xx 错误信息与 `CheckRawResponseStatus` 均脱敏后再输出 |
| P1-18 | **保留现状待复核**（`openapi/rid/get` 官方页 404，无法核实 component_access_token 支持，改动风险大于收益） |
| P1-19 | **保留现状待复核**（`openapi/quota/get` 同上） |

### 10.3 P2 择要（全部处理）

- **payment**：4 个下单结构补 `time_expire`/`support_fapiao`/`detail`/`settle_info`（`detail.goods_detail` 按官方补 `wechatpay_goods_id`，官方无 `invoice_amount`）；H5 补 `attach`；`SceneInfo` 补 `device_id`/`store_info`；分账/退款查询删除多传 `mchid`；分账/回退应答补 `create_time`/`finish_time`/`return_id`；撤销转账不再发空 `{}` body；回调成功应答改为 204 空体（新增 `NotificationSuccess`/`NotificationFail`，删除非官方的 `{"code":"SUCCESS"}` 应答）；`APIErrorDetail` 删除 `detail_list`；转账补 `user_recv_style`（官方为对象 `{type}`，按对象实现）与应答 `fail_reason`；**注释中的官方文档 doc id 失效问题未逐条修订**（需逐条联网核验新文档 ID，错误风险大于收益，待后续随接口注释更新）
- **official_account**：`bizsend` data 删除多序列化的 `name`；`getpubtemplatekeywords` 补 `count`；`tid` 改 string（破坏性，含兼容层签名）；群发/预览补 `clientmsgid`/`towxname`/`msg_data_id`；客服补 `customservice`/`msgmenu` 类型；菜单 `Button` 补 `article_id`；素材响应删除非官方 `video_url`；`getuserread` 的 `user_source` 枚举注释按官方文档修正；补齐 5 个 `getupstreammsg*` 无方法常量的实现；稳定版 token 注释补「不支持第三方平台代调用」
- **openplatform**：`config.Validate` 强制 EncodingAESKey 必填 43 位（破坏性：空配置将校验失败）；事件 AppID 校验放宽（空值跳过并 Warn）；component/authorizer token 获取对齐并发保护 + 5 分钟过期余量；JS-SDK config 缓存按 ticket 实际过期时间；`SendCustomMessage` 补 voice/video/music/news（video 含官方必填 `thumb_media_id`）；删除死常量 `URLGetJumpDomainFile`/`URLGetUserInfo` 与死类型 `OAuthComponentAccessTokenRequest`；`AuthorizerInfo` 补 `idc`；pre_auth_code 缓存复用（60 秒余量）
- **miniprogram/公共层**：`msgSecCheck`/`mediaCheckAsync` 补 openid/scene 本地校验；`head_avatar` 字段保守恢复保留（官方 v2 初版参数表含该字段用于头像画像检测，现行参数表已不再列出——按保守兼容原则恢复并新增 `MsgSecCheckFull` 完整参数方法，顺带补齐 title/nickname/signature 的透传入口，微信侧会忽略未知字段）；响应 `detail` 补 `keyword`、`label` 枚举注释按官方修正（20012 低俗、20013 版权）；`media_url` 注释改 10M；`WXACodeResponse` 暴露 `Num-Used` 响应头；`PhoneInfo` 补 `Watermark`；新增小程序稳定版 access_token（`GetStableAccessToken`，不写入默认缓存）；`EventTimestampTolerance` 接入 `DecryptMsgWithKey`/`VerifyURL`（时间戳为空/非数字时保守跳过）；`storage/file.go` 原子写修复（`CreateTemp` 唯一临时名防多进程互踩、`Sync()` 后 rename、权限 0600（已存在文件保留原模式）、消除双重 Close、失败路径清理临时文件）
- **待人工复核项保守处理**：`submit_audit` 回包 `auditid` 改 `core.FlexString` 双形态兼容（官方参数表 string、示例数字）；GZIP 账单 `hash_value` 校验改为「先按原始字节、失败再按解压后字节」双路径（官方语义未明确）；`mass/get` 的 `msg_id` 保持 int64、`tag_id` 保持 int（现行官方参数表与示例均为 number）；`clear_quota/v2` 的 `appid` 必填性官方自相矛盾，保持现状

### 10.4 本批破坏性变更清单

1. **删除**：`official_account` 的 `URLAddNews`/`URLUpdateNews`/`URLMaterialUploadImage` 常量与 `AddNews`/`UpdateMaterial` 方法及配套请求/响应类型；17 个 `/card/*` 卡券常量、`URLPayNotify`、`URLScanCodePush`、4 个零引用开放平台常量（随第九批核实）；`openplatform` 的 `URLGetJumpDomainFile`/`URLGetUserInfo` 常量与 `OAuthComponentAccessTokenRequest` 类型；`payment.APIErrorDetail.DetailList` 字段；`payment.NotificationAckJSON`/`NotificationAckCodeSuccess` 常量（成功应答改 `NotificationSuccess`）
2. **类型变更**：`TemplateSubscribeRequest.Data` 由 `map[string]TemplateSubscribeData` 改为固定结构 `TemplateSubscribeData{Content}`；`tid` 相关字段/参数由 int 改 string；`msg_sec_check`/`media_check_async` 的 `version` 由 string 改 int；`SubmitWxaAuditResponse.AuditID` 由 int64 改 `core.FlexString`；`GetWxaCodeUnlimitRequest.CheckPath` 由 `bool` 改 `*bool`；`MassFilter.IsToAll` 改 `core.FlexBool`
3. **行为变更**：`AddMaterial` 对 video 类型返回错误（引导使用 `AddMaterialVideo`）；`GetWxaCode`/`CreateWxaQrcode` 对空 `path` 返回错误；`MsgSecCheck`/`MediaCheckAsync` 对空 openid/非法 scene 返回错误；微信支付回调成功应答由 `{"code":"SUCCESS"}` JSON 改为 204 空体；`openplatform` 客户端构造时 EncodingAESKey 必填 43 位；授权方 token 40001 重试不再删除整条授权记录；`WXBizMsgCrypt.DecryptMsg`/`VerifyURL` 增加时间戳校验（±300 秒，空/非法时间戳跳过）；网页授权 URL 不再附加 `connect_redirect=1`
4. **新增导出**：`crypto.DecryptMsgWithKey`/`EncryptMsgWithKey`；`core.FlexString`/`FlexBool`；`message.WithTokenStorage`；`miniprogram.GetStableAccessToken`；`official_account.AddMaterialVideo`、`GetUpStreamMsg*` 5 个方法、客服 `customservice`/`msgmenu` 类型；`payment.NotificationSuccess`/`NotificationFail`；`openplatform.Client.DecryptMessage` 支持密钥轮换降级解密
