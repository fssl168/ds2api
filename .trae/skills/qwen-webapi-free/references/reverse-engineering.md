# 通义千问 Web API 逆向工程过程

## 一、环境准备

### 1.1 目标站点
- **主站**: `https://www.qianwen.com`（通义千问网页版）
- **API 域名**: `chat2.qianwen.com`
- **安全域**: `sec.qianwen.com`
- **UMID 域**: `ynuf.aliapp.org`

### 1.2 工具链
- Chrome DevTools（F12 → Network）
- 浏览器登录通义千问账号
- Postman / curl（用于验证还原结果）

## 二、抓包分析

### 2.1 抓取聊天请求

1. 打开 `https://www.qianwen.com`，登录账号
2. F12 → Network 面板，过滤 `chat2`
3. 发送一条消息，观察请求

**关键发现**：

```
POST https://chat2.qianwen.com/api/v2/chat?biz_id=ai_qwen&chat_client=h5&device=pc&fr=pc&pr=qwen&ut={deviceID}&la=zh-CN&tz=Asia/Shanghai&nonce={random8}&timestamp={ms}
```

### 2.2 Cookie 分析

```
theme-mode=light; _samesite_flag_=true;
tongyi_sso_ticket={YOUR_TICKET};           ← 核心凭证！
tongyi_sso_ticket_hash=tytk_hash:{sha256前16位};
XSRF-TOKEN={xsrf_uuid};
```

**结论**：`tongyi_sso_ticket` 就是用户的唯一凭证（类似 API Key），从浏览器 Cookie 中直接提取即可使用。

### 2.3 Request Headers 全量记录

按类别分类所有自定义 Header：

#### 类别 A：bx-* 反爬系列

| Header | 格式 | 作用 |
|--------|------|------|
| `bx-v` | `2.5.31` | bx SDK 版本号 |
| `bx-et` | ~492 字符 base64 | 反爬 token（随机+时间戳+HMAC签名） |
| `bx-ua` | `231!{base64_gzip}` | 浏览器指纹（gzip 压缩的 JSON） |
| `bx-umidtoken` | base64(44字节) | UMID 设备标识 |

#### 类别 B：eo-clt-* ACS 安全令牌

| Header | 来源 | 作用 |
|--------|------|------|
| `eo-clt-acs-ve` | 固定 `1.0.0` | ACS 版本 |
| `eo-clt-acs-kp` | SHA256(ticket)[:16] 的 hash 前缀 | 密钥指针 |
| `eo-clt-sacsft` | Security Register 返回 | 签名密钥片段 |
| `eo-clt-snver` | Security Register 返回 | 签名版本 |
| `eo-clt-dvidn` | Security Register 返回 | 设备 ID |
| `eo-clt-actkn` | Security Register 返回 | 访问令牌 |

#### 类别 C：clt-acs-* 客户端签名

| Header | 格式 | 作用 |
|--------|------|------|
| `clt-acs-caer` | `vrad` | 加密算法标识 |
| `clt-acs-request-params` | 参数列表字符串 | 签名覆盖参数 |
| `clt-acs-sign` | Base64(HMAC-SHA256)[:32] | 请求签名 |
| `clt-acs-reqt` | 时间戳毫秒 | 签名时间戳 |

#### 类别 D：x-* 自定义头

| Header | 值 | 作用 |
|--------|-----|------|
| `x-xsrf-token` | UUID | CSRF 保护 |
| `x-chat-id` | UUID (=req_id) | 聊话会话 ID |
| `x-deviceid` | UUID | 设备标识 |
| `x-platform` | `pc_tongyi` | 平台标识 |

## 三、逐层逆向还原

### 3.1 第一层：Cookie 认证（最简可用）

**发现**：即使只设置 Cookie + 基础浏览器 Header，不设置任何 bx-*/ACS 签名头，**部分情况下也能成功调用**。

最小可行请求：
```http
POST /api/v2/chat?biz_id=ai_qwen&chat_client=h5&device=pc&fr=pc&pr=qwen
Cookie: tongyi_sso_ticket={ticket}; tongyi_sso_ticket_hash={hash}; XSRF-TOKEN={xsrf}
Content-Type: application/json
User-Agent: Mozilla/5.0 ...
Referer: https://www.qianwen.com/chat/{chatId}
Origin: https://www.qianwen.com
```

**但稳定性差**：高频调用或新 IP 会触发 401/403。

### 3.2 第二层：bx-et + bx-ua 反爬

#### bx-et 还原

抓包观察格式：
- 长度固定 ~492 字符（base64url 编码）
- 以 `.` 结尾

**逆向分析**：
```python
# 伪代码还原
payload = random_bytes(368)          # 368 字节随机填充
timestamp = str(current_time_ms)      # 13 位毫秒时间戳
combined = payload + timestamp.encode()
signature = HMAC-SHA256("baxia_et_salt", combined)  # 32 字节签名
full = combined + signature            # 368 + 13 + 32 = 413 字节
encoded = base64url_encode(full)       # ~550 字符
result = encoded[:492] + "."           # 截断到 492 + 补齐到 490 + "."
```

Go 实现：见 `client.go` → `generateBxET()`

#### bx-ua 还原

抓包观察格式：
- 以 `231!` 开头
- 后跟 base64 编码数据（解码后是 gzip 数据）

**逆向分析**：
```python
fingerprint = {
    "s": {"w":1920,"h":1080,...},     # 屏幕
    "n": {"pl":"Win32","la":"zh-CN",...}, # Navigator
    "h": {"cores":8,"mem":8,...},        # 硬件
    "g": {"v":"Google Inc. (NVIDIA)",...},# WebGL
    "c": {"hw":"a1b2c3...","hh":"7890..."},# Canvas hash
    "f": ["Arial","Microsoft YaHei",...], # 字体列表
    "p": [{"n":"PDF Viewer",...}],        # 插件
    "st": {"idb":True,"ls":True,...},     # 存储
    "t": timestamp_ms,                    # 时间戳
    "e": random_hex(32),                  # 熵值
}
json_bytes = json.dumps(fingerprint)
gzipped = gzip.compress(json_bytes)
result = "231!" + base64(gzipped)
```

Go 实现：见 `client.go` → `generateBxUA()` + `buildBrowserFingerprint()`

### 3.3 第三层：HMAC 签名（clt-acs-sign）

#### 签名算法还原

抓包多组请求，对比 `clt-acs-sign` 和其他 header 的关系：

**发现规律**：
```
sign = Base64(HMAC-SHA256(key, message))[:32]
```

其中：
- **基础模式 key** = `"qwen_chat_sign_key_v1"`（硬编码）
- **增强模式 key** = `":" + ticket_hash`（SecurityManager 注册后）
- **message** = `{chatID}{timestamp}{kp}{params}`
- **kp** = `tytk_hash:{sha256(ticket)[:16]}`（ticket 的短哈希）
- **params** = `"biz_id,chat_client,device,fr,pr,ut,la,tz,nonce,timestamp"`

Go 实现：见 `client_auth.go` → `computeRealSign()`

### 3.4 第四层：SecurityManager ACS 令牌

#### 发现注册流程

在 Network 中搜索 `sec.qianwen.com`，发现两个关键请求：

**1. UMID 获取**
```
POST https://ynuf.aliapp.org/service/um.json
Body: data={base64_fingerprint}&xa=wagbridgead-sm-nginx-quarkpc-security-calibration-time-web&xt=
Response: {"tn":"{umid_token}","id":"{device_id}"}
```

UMID fingerprint 格式：`107!{base64("screen=1920*1080|language=zh-CN|...")}`

**2. Security Register**
```
POST https://sec.qianwen.com/security/external/access/register?chid={chid}
Headers: Cookie, bx-umidtoken, bx-v, eo-clt-acs-bx-intss=1
Body: {
  "features": {屏幕/语言/WebGL/字体/插件/...},
  "fingerprint": {user_agent/platform/vendor/...},
  "businessScene": "qwen_cha",
  "chid": "{chid}",
  "unifyRelateGenerate": []
}
Response: {
  "data": {
    "eo-clt-actkn": "{access_token}",
    "eo-clt-dvidn": "{device_id_new}",
    "eo-clt-bacsft": ["{sign_key_fragment}"],
    "eo-clt-snver": "{sign_version}",
    "eo-clt-actkn-dl": {expire_timestamp},
    "expireTime": {expire_ms}
  }
}
```

**返回的 actkn/dvidn/sacsft/snver 用于后续聊天请求的 eo-clt-* 头**，且签名 key 从固定值变为动态 `:{ticket_hash}`。

Go 实现：见 `client_sec.go` → `SecurityManager.Register()`

## 四、请求体协议分析

### 4.1 L2 协议结构

抓包分析 POST body JSON：

```json
{
  "deep_search": "0",
  "req_id": "uuid-无横线",
  "model": "Qwen3-Plus",
  "scene": "chat",
  "session_id": "同 req_id",
  "sub_scene": "chat",
  "temporary": false,
  "messages": [{
    "content": "用户输入的消息文本",
    "mime_type": "text/plain",
    "meta_data": {"ori_query": "用户输入的消息文本"}
  }],
  "from": "default",
  "topic_id": "uuid-无横线",
  "parent_req_id": "0",
  "scene_param": "first_turn",
  "chat_client": "h5",
  "client_tm": "当前时间戳毫秒",
  "protocol_version": "v2",
  "biz_id": "ai_qwen"
}
```

**关键特点**：
1. 只发送最后一条用户消息（不是完整对话历史）—— 这是 L2 协议的特点
2. `session_id` 每次请求可不同（服务端有状态管理）
3. `model` 使用内部名称（Qwen/Qwen3-Plus 等），非 OpenAI 兼容名
4. `messages[0].meta_data.ori_query` 是消息原文副本

## 五、响应格式分析

### 5.1 SSE 流格式

响应 Content-Type: `text/event-stream`

```
data: {"success":true,"data":{"messages":[{"content":"你好","mime_type":"text/plain"}],"status":"running"}}\n\n
data: {"success":true,"data":{"messages":[{"content":"！我是","mime_type":"text/plain"}],"status":"running"}}\n\n
data: {"success":true,"data":{"messages":[{"content":"通义","mime_type":"text/plain"}],"status":"complete"}}\n\n
event: complete\ndata: true\n\n
```

### 5.2 解析规则

1. 跳过空行
2. `event: complete` 行 → 标记流结束
3. `data:` 行：
   - `true` 或 `[DONE]` → 结束标记
   - JSON 对象 → 提取 `data.messages[]` 中 `mime_type` 为 `text/plain` 或 `multi_load/iframe` 的 `content` 字段
4. `data.status === "complete"` → 最后一个有效数据包

## 六、模型名映射表

通过多次抓包不同模型的请求，整理出映射关系：

| 页面选择 | 请求 model 值 | 能力 |
|----------|-------------|------|
| 通义千问（默认） | `Qwen` | 基础对话 |
| Qwen-Max | `Qwen3-Max` | 最强推理 |
| Qwen-Plus | `Qwen3-Plus` | 增强版 |
| Qwen-Coder | `Qwen3-Coder` | 代码专用 |
| Qwen-Flash | `Qwen3-Flash` | 轻量快速 |
| Qwen3.5-Plus | `Qwen3.5-Plus` | 新一代增强 |
| Qwen3.5-Flash | `Qwen3.5-Flash` | 新一代轻量 |

## 七、时间校准

API 要求客户端时间戳与服务器同步。校准端点：

```
GET https://sec.qianwen.com/api/calibration/getMillisTimeStamp
Response: {"data": {"millisTimeStamp": "1743xxx..."}}
```

用于生成 `clt-acs-reqt` 和 URL query 中的 `timestamp` 参数。
