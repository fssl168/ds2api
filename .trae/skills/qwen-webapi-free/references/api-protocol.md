# 通义千问 API 协议详解

## 1. 端点总览

| 用途 | 方法 | URL | 认证 |
|------|------|-----|------|
| **聊天补全** | POST | `https://chat2.qianwen.com/api/v2/chat?{query}` | Cookie + Headers |
| **时间校准** | GET | `https://sec.qianwen.com/api/calibration/getMillisTimeStamp` | 无 |
| **安全注册** | POST | `https://sec.qianwen.com/security/external/access/register?chid={chid}` | Cookie |
| **UMID 获取** | POST | `https://ynuf.aliapp.org/service/um.json` | 无 |

## 2. 认证体系详解

### 2.1 Cookie 结构

```
theme-mode=light                                    # UI 主题
_samesite_flag_=true                                # SameSite 标志
tongyi_sso_ticket={TICKET}                          # ★ 核心凭证（从浏览器获取）
tongyi_sso_ticket_hash=tytk_hash:{SHA256(TICKET)[:16]} # Ticket 哈希（前缀固定）
XSRF-TOKEN={UUID}                                   # CSRF Token
```

**Ticket 获取方式**：
1. 浏览器登录 `https://www.qianwen.com`
2. F12 → Application → Cookies → `qianwen.com` → 找 `tongyi_sso_ticket`
3. 该值即为 ticket 凭证，有效期较长（通常数天到数周）

### 2.2 三层认证优先级

```
┌─────────────────────────────────┐
│  Layer 3: SecurityManager ACS   │ ← 最高安全性（自动注册刷新）
│  - eo-clt-actkn (访问令牌)       │
│  - eo-clt-dvidn (设备ID)         │
│  - eo-clt-sacsft (签名密钥片段)   │
│  - eo-clt-snver (签名版本)        │
│  - 签名 key = ":" + ticket_hash  │
├─────────────────────────────────┤
│  Layer 2: bx-* 反爬指纹          │ ← 中等安全性（伪造浏览器特征）
│  - bx-et (~492字符 base64)       │
│  - bx-ua (gzip 浏览器指纹)        │
│  - bx-umidtoken (设备标识)        │
├─────────────────────────────────┤
│  Layer 1: Cookie + 基础签名        │ ← 最低安全性（可快速调通）
│  - Cookie (ticket)               │
│  - clt-acs-sign (固定key签名)     │
└─────────────────────────────────┘
```

**降级策略**：当 SecurityManager 注册失败时，自动回退到 Layer 2 或 Layer 1。

## 3. 签名算法

### 3.1 基础签名（Layer 1/2 fallback）

```go
func computeRealSign(chatID string, timestamp int64) string {
    key := []byte("qwen_chat_sign_key_v1")                          // 固定密钥
    msg := fmt.Sprintf("%s%d%s%s", chatID, timestamp, "", params) // msg 中 kp 为空
    mac := hmac.New(sha256.New, key)
    mac.Write([]byte(msg))
    return base64.StdEncoding.EncodeToString(mac.Sum(nil))[:32]     // 截取前 32 字符
}
```

参数说明：
- `chatID`: 同 req_id（UUID 无横线）
- `timestamp`: 毫秒时间戳（建议先调 calibrateTime 校准）
- `""`: 第三个参数（基础模式下为空字符串）
- `params`: 固定值 `"biz_id,chat_client,device,fr,pr,ut,la,tz,nonce,timestamp"`

### 3.2 增强签名（Layer 3 SecurityManager）

```go
func (sm *SecurityManager) BuildChatHeaders(...) http.Header {
    kp := sm.client.ticketHash(ticket)                               // tytk_hash:{sha256[:16]}
    sigInput := fmt.Sprintf("%s%d%s%s", chatID, timestamp, kp, params) // msg 中 kp 有值
    signValue := computeHMAC256(":"+kp, sigInput)                     // key 动态化
    // ...
}
```

差异点：
- **key** 从 `"qwen_chat_sign_key_v1"` 变为 `":" + ticket_hash`
- **msg** 中第三个参数从空字符串变为 `kp`（ticket 哈希前缀）

### 3.3 Ticket Hash 计算

```go
func (c *Client) ticketHash(ticket string) string {
    h := sha256.Sum256([]byte(ticket))
    return fmt.Sprintf("tytk_hash:%x", h[:16])    // 取前 16 字节 hex
}
```

## 4. 请求头完整清单

### 4.1 必需头（所有层级）

| Header | 示例值 | 说明 |
|--------|--------|------|
| `Content-Type` | `application/json;charset=UTF-8` | 请求体编码 |
| `Accept` | `application/json, text/event-stream, text/plain, */*` | 接受 SSE 流 |
| `User-Agent` | `Mozilla/5.0 ... Chrome/147.0.0.0 ...` | 浏览器 UA |
| `Origin` | `https://www.qianwen.com` | 来源 origin |
| `Referer` | `https://www.qianwen.com/chat/{chatID}` | 引用页面 |
| `Cookie` | 见 2.1 节 | 认证凭证 |

### 4.2 安全头（Layer 2+）

| Header | 生成方式 | 说明 |
|--------|----------|------|
| `bx-v` | 固定 `2.5.31` | bx SDK 版本 |
| `bx-et` | `generateBxET()` | 反爬 token（~492字符） |
| `bx-ua` | `generateBxUA()` | 浏览器指纹（gzip+base64） |
| `bx-umidtoken` | `generateFakeUMIDToken()` 或 UMID API 返回 | 设备 umid |

### 4.3 ACS 头（Layer 3，SecurityManager 有效时）

| Header | 来源 | 说明 |
|--------|------|------|
| `eo-clt-acs-ve` | 固定 `1.0.0` | ACS 协议版本 |
| `eo-clt-acs-kp` | `ticketHash(ticket)` | 密钥指针 |
| `eo-clt-sacsft` | Register 返回 `bacsfts[0]` | 签名密钥片段 |
| `eo-clt-snver` | Register 返回 `snver` 或默认 `lv` | 签名版本 |
| `eo-clt-dvidn` | Register 返回 `dvidn` | 设备 ID |
| `eo-clt-actkn` | Register 返回 `actkn` | 访问令牌 |

### 4.4 签证头（所有层级）

| Header | 生成方式 | 说明 |
|--------|----------|------|
| `clt-acs-caer` | 固定 `vrad` | 加密算法 |
| `clt-acs-request-params` | 固定参数字符串 | 签名覆盖字段 |
| `clt-acs-sign` | `computeRealSign()` 或 `computeHMAC256()` | 请求签名 |
| `clt-acs-reqt` | `fmt.Sprintf("%d", serverTime)` | 签名时间戳 |

### 4.5 自定义头（所有层级）

| Header | 生成方式 | 说明 |
|--------|----------|------|
| `x-xsrf-token` | `uuid.New().String()` | CSRF Token |
| `x-chat-id` | 同 req_id | 聊话 ID |
| `x-deviceid` | `uuid.New().String()`（Client 初始化时生成） | 设备 ID |
| `x-platform` | 固定 `pc_tongyi` | 平台标识 |

### 4.6 Sec-Fetch 头（伪装浏览器）

```
Sec-Ch-Ua: "Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"
Sec-Ch-Ua-Mobile: ?0
Sec-Ch-Ua-Platform: "Windows"
Sec-Fetch-Dest: empty
Sec-Fetch-Mode: cors
Sec-Fetch-Site: same-site
```

## 5. URL Query 参数

聊天请求 URL 携带查询参数：

```
?biz_id=ai_qwen              # 业务 ID
&chat_client=h5             # 客户端类型（h5=网页）
&device=pc                  # 设备类型
&fr=pc                     # 来源
&pr=qwen                   # 产品
&ut={deviceID}             # 用户/设备 token
&la=zh-CN                  # 语言
&tz=Asia/Shanghai          # 时区
&nonce={8位随机字母数字}     # 随机数
&timestamp={毫秒时间戳}      # 服务校准时间
```

## 6. 请求体详细字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `deep_search` | string | 是 | 固定 `"0"`（深度搜索开关） |
| `req_id` | string | 是 | 请求 UUID（无横线） |
| `model` | string | 是 | 内部模型名（Qwen/Qwen3-Plus 等） |
| `scene` | string | 是 | 固定 `"chat"` |
| `session_id` | string | 是 | 会话 ID（可与 req_id 相同） |
| `sub_scene` | string | 是 | 固定 `"chat"` |
| `temporary` | bool | 是 | 固定 `false` |
| `messages` | array | 是 | 消息数组（仅含最后一条用户消息） |
| `from` | string | 是 | 固定 `"default"` |
| `topic_id` | string | 是 | 话题 UUID（无横线） |
| `parent_req_id` | string | 是 | 固定 `"0"` |
| `scene_param` | string | 是 | 固定 `"first_turn"` |
| `chat_client` | string | 是 | 固定 `"h5"` |
| `client_tm` | string | 是 | 客户端时间戳（毫秒字符串） |
| `protocol_version` | string | 是 | 固定 `"v2"` |
| `biz_id` | string | 是 | 固定 `"ai_qwen"` |

### messages 元素结构

```json
{
  "content": "用户消息文本",
  "mime_type": "text/plain",
  "meta_data": {
    "ori_query": "用户消息文本（与 content 相同）"
  }
}
```

## 7. 响应格式

### 7.1 成功响应（SSE 流）

Content-Type: `text/event-stream; charset=utf-8`

每个 SSE 事件格式：

```
data: {JSON}\n\n
```

JSON 结构：

```typescript
interface QwenSSEEvent {
  success: boolean
  data: {
    messages: Array<{
      content: string        // 文本内容片段
      mime_type: string      // "text/plain" | "multi_load/iframe" | ...
      status?: string        // "running" | "complete"
    }>
    status: "running" | "complete" | "error"
  }
}
```

### 7.2 流结束标志

两种结束信号（任一出现即表示流结束）：

1. **event 行**:
   ```
   event: complete
   data: true
   ```

2. **data 值**:
   ```
   data: [DONE]
   ```

### 7.3 错误响应

| HTTP Status | 含义 |
|-------------|------|
| 401 | Ticket 无效或过期 |
| 403 | 签名被拒 / IP 封禁 / 频率限制 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |

错误响应体示例：
```json
{"code":"AUTH_FAILED","message":"用户未登录","success":false}
```

## 8. chid 生成规则

`chid`（Client Hardware ID）用于 Security Register：

```
格式: {10位时间戳后缀}{17位随机小写字母}

示例: 4506123456abcdefghijklmnopq
```

Go 实现：
```go
func generateChid() string {
    b := make([]byte, 17)
    for i := range b {
        b[i] = letterBytes[rand.Intn(len(letterBytes))]  // a-z
    }
    ts := time.Now().UnixMilli()
    return fmt.Sprintf("%d%s", ts%10000000000, string(b))
}
```

## 9. nonce 生成规则

URL query 中的 nonce：8 位随机小写字母数字

```go
func generateNonce() string {
    const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
    b := make([]byte, 8)
    for i := range b {
        b[i] = chars[rand.Intn(len(chars))]
    }
    return string(b)
}
```
