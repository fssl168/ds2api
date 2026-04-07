---
name: qwen-webapi-free
description: >
  通义千问（Qwen）Web API 逆向工程与免费接入完整方案。
  包含：API 协议逆向分析、认证机制破解（Cookie/Security Headers/签名算法）、
  SSE 流式解析、多账号池管理（Acquire/Release 模式）、前端集成、Admin CRUD。
  适用场景：(1) 将 Qwen 接入 OpenAI 兼容 API 网关 (2) 实现 Qwen 多账号轮询与并发控制
  (3) 复现 Qwen Web 聊天协议 (4) 构建 Qwen 账号池管理系统。
  触发关键词：qwen, 通义千问, 千问, qwen-webapi, qwen pool, qwen reverse, qwen 逆向,
  tongyi, 千问账号池, qwen free api, qwen web chat protocol.
---

# Qwen Web API Free — 通义千问逆向工程完整方案

## 概述

本 Skill 记录了将阿里通义千问（Qwen）Web 聊天接口逆向工程并集成到 ds2api（OpenAI 兼容 API 网关）的**完整过程**。涵盖协议分析、认证破解、流式解析、账号池实现、前后端集成全链路。

## 架构总览

```
Client (OpenAI SDK)
  → /v1/chat/completions (model: qwen-plus)
    → handler_chat.go (智能路由: qwen-* 前缀 → DetermineCaller)
      → qwen/client.go:CallCompletion()
        → QwenPool.Acquire()     # 从池中获取 ticket
        → doPostChat()            # 构造请求 + 安全头 + Cookie
          → POST https://chat2.qianwen.com/api/v2/chat
        → sse_scanner.go          # 解析 SSE 流
        → QwenPool.Release()      # 归还 ticket
        → MarkSuccess/MarkFailed  # 更新健康状态
```

## 快速索引

| 主题 | 文件 | 说明 |
|------|------|------|
| **逆向分析过程** | [references/reverse-engineering.md](references/reverse-engineering.md) | 完整逆向步骤：抓包→分析→还原 |
| **API 协议详解** | [references/api-protocol.md](references/api-protocol.md) | 认证、签名、请求体、响应格式 |
| **账号池实现** | [references/pool-implementation.md](references/pool-implementation.md) | Acquire/Release、并发控制、冷却机制 |

## 核心文件清单

### 后端 (`internal/qwen/`)

| 文件 | 行数 | 职责 |
|------|------|------|
| `client.go` | ~286 | 客户端入口、设备指纹生成（bx-et/bx-ua/umid）、ticket 管理 |
| `client_auth.go` | ~121 | Cookie 构建、安全头设置、HMAC 签名、时间校准、nonce 生成 |
| `client_completion.go` | ~139 | 对话补全主流程：Acquire→POST→SSE 解析→Release→重试 |
| `client_sec.go` | ~448 | SecurityManager：注册/刷新 ACS token、UMID 获取、浏览器特征收集 |
| `client_session.go` | ~19 | Session 存根（Qwen 无需 session，空实现兼容接口） |
| `pool.go` | ~395 | **QwenPool 完整实现**：Acquire/Release、等待队列、健康检查、自动冷却 |
| `prompt.go` | ~84 | L2 消息格式转换、模型名映射（qwen-* → 内部模型名） |
| `sse_scanner.go` | ~122 | SSE 流解析器：data/event 行解析、JSON 反序列化 |
| `constants.go` | ~12 | API 端点常量、重试参数 |
| `models.go` | ~20 | 模型列表注册 |

### 集成层

| 文件 | 变更内容 |
|------|----------|
| `internal/adapter/openai/handler_chat.go` | 智能路由：`qwen-*` 前缀 → `DetermineCaller(r)` → Qwen Caller |
| `internal/adapter/openai/handler_routes.go` | 注册 qwen caller 到 deps 注入 |
| `internal/admin/deps.go` | 新增 `QwenCaller` 接口（Pool/ResetTickets） |
| `internal/admin/handler.go` | 新增 `QW` 字段 + `/admin/qwen-pool/status` 路由 |
| `internal/admin/handler_qwen_accounts_crud.go` | Qwen 账号 CRUD + 池状态 API |
| `internal/admin/handler_config_import.go` | 批量导入支持 qwen_accounts（merge/replace + 去重） |
| `internal/config/codec.go` | JSON 编解码支持 QwenAccount |
| `internal/config/store.go` | 导出包含 QwenAccounts、tickets 提取逻辑 |
| `internal/server/router.go` | 传递 QW client 到 admin handler |

### 前端 (`webui/src/`)

| 文件 | 职责 |
|------|------|
| `features/account/QwenAccountsTable.jsx` | 千问账号表格展示 |
| `features/account/AddQwenAccountModal.jsx` | 新增千问账号弹窗 |
| `components/BatchImport.jsx` | 批量导入含 qwen_only 模板 + 全量模板 |
| `features/apiTester/*` | API 测试器支持 Qwen 模型选择 |

## 关键技术要点

### 1. 认证体系（三重防护）

```
层级1: Cookie    → tongyi_sso_ticket={ticket} + XSRF-TOKEN
层级2: bx-* 头   → bx-et (反爬 token) + bx-ua (浏览器指纹) + bx-umidtoken
层级3: ACS 签名  → eo-clt-actkn/dvidn/sacsft/snver + clt-acs-sign (HMAC-SHA256)
```

详见 [api-protocol.md](references/api-protocol.md#认证体系)。

### 2. 签名算法

```go
// 核心签名（无 SecurityManager 时）
key = "qwen_chat_sign_key_v1"
msg = "{chatID}{timestamp}{ticket_hash}{params}"
sign = Base64(HMAC-SHA256(key, msg))[:32]

// 增强（有 SecurityManager 时）
key = ":{ticket_hash}"
msg = "{chatID}{timestamp}{kp}{params}"
sign = Base64(HMAC-SHA256(key, msg))
```

### 3. 模型名映射

| 用户请求 | 实际发送 |
|----------|----------|
| `qwen-plus`, `qwen/qwen-plus`, `qwen/Qwen3-Plus` | `Qwen3-Plus` |
| `qwen-max`, `qwen/qwen-max` | `Qwen3-Max` |
| `qwen-coder`, `qwen/qwen-coder` | `Qwen3-Coder` |
| `qwen-flash`, `qwen/qwen-flash` | `Qwen3-Flash` |
| `qwen3.5-plus`, `qwen/qwen3.5-plus` | `Qwen3.5-Plus` |
| `qwen3.5-flash`, `qwen/qwen3.5-flash` | `Qwen3.5-Flash` |
| 其他 qwen-* | `Qwen`（默认） |

### 4. QwenPool 账号池核心流程

```
请求到达
  → Acquire(ctx)
    → 遍历 entries，跳过 cooldown 中
    → 检查 inUse[label] < maxInflightPerTicket
    → 检查 global in-flight < globalMaxInflight
    → 找到：inUse++，返回 entry
    → 未找到且队列未满：创建 waiter channel，阻塞等待
    → 未找到且队列已满：返回 ErrQwenPoolExhausted (429)
  → 使用 ticket 调用 API
  → 成功 → Release(label) + MarkSuccess(label)
  → 失败 → Release(label) + MarkFailed(label)
    → failCount++ ≥ 3 → 冷却 30+failCount*10 秒
```

详见 [pool-implementation.md](references/pool-implementation.md)。

### 5. SSE 流格式

```
event: complete\ndata: {"success":true,"data":{"messages":[{"content":"...","mime_type":"text/plain"}],"status":"complete"}}\n\n
或
data: {"success":true,"data":{"messages":[{...}],"status":"running"}}\n\n
event: complete\ndata: true\n\n
```

解析规则：
- `event: complete` + `data: true` / `data: [DONE]` → 流结束
- `data:` JSON → 提取 `data.messages[0].content`（过滤 `multi_load/iframe` 和 `text/plain`）

### 6. 请求体结构（L2 协议）

```json
{
  "deep_search": "0",
  "req_id": "{uuid}",
  "model": "Qwen3-Plus",
  "scene": "chat",
  "session_id": "{uuid}",
  "sub_scene": "chat",
  "temporary": false,
  "messages": [{"content": "用户消息", "mime_type": "text/plain", "meta_data": {"ori_query": "用户消息"}}],
  "from": "default",
  "topic_id": "{uuid}",
  "parent_req_id": "0",
  "scene_param": "first_turn",
  "chat_client": "h5",
  "client_tm": "{timestamp_ms}",
  "protocol_version": "v2",
  "biz_id": "ai_qwen"
}
```

注意：只发送最后一条用户消息（非完整历史），这是 L2 协议的特点。

## 逆向步骤速查

完整的抓包→分析→还原过程见 [reverse-engineering.md](references/reverse-engineering.md)，速查版：

1. **F12 抓包** `https://www.qianwen.com` 聊天 Network → Filter: `chat2.qianwen.com`
2. **提取 Cookie**: `tongyi_sso_ticket=xxxxx`（这就是你的 ticket 凭证）
3. **记录所有 Request Headers**：特别是 `bx-*`、`eo-clt-*`、`clt-acs-*`、`x-*` 开头的
4. **分析签名**：`clt-acs-sign` 是 HMAC-SHA256，key 含 `qwen_chat_sign_key_v1`
5. **分析 bx-et**：~492 字符 base64，前 368 字节随机 + 13 字节时间戳 + 32 字节 HMAC-SHA256 签名
6. **分析 bx-ua**：`231!` + gzip(base64(浏览器指纹 JSON))
7. **分析 Security Register**：POST `sec.qianwen.com/security/external/access/register` 返回 actkn/dvidn/sacsft/snver
8. **分析 UMID**：POST `ynuf.aliapp.org/service/um.json` 返回 tn/id
9. **构造最小可用请求**：先不启用 SecurityManager，用固定签名 key 即可调通
10. **增强安全性**：实现 SecurityManager 自动注册/刷新 ACS token

## 常见问题排查

| 问题 | 原因 | 解决 |
|------|------|------|
| 401/403 | ticket 过期或无效 | 重新从浏览器获取新 ticket |
| 429 Too Many Requests | 账号池耗尽 | 增加 qwen_accounts 数量或调整并发限制 |
| 签名被拒 | 时间戳偏差大 | 先调用 calibrateTime 校准服务器时间 |
| SSE 解析空 | 格式变化 | 检查 `parseL2Line` 是否匹配实际响应 |
| SecurityManager 注册失败 | umid 获取失败 | 可降级为无 SM 模式（固定签名 key） |
