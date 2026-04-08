# Qwen (通义千问) Web 转 API 集成方案

## 一、项目背景与目标

### 1.1 现状分析

**ds2api 项目架构**（基于代码深度分析）：
- **语言**: Go 1.24+ 全量实现，无 Python 运行时依赖
- **核心模式**: 将 DeepSeek Web 对话能力转换为 OpenAI/Claude/Gemini 兼容 API
- **适配器层**: `internal/adapter/{openai,claude,gemini}/` 三套协议适配
- **客户端层**: `internal/deepseek/` 封装 DeepSeek Web API 的登录、会话、PoW、流式调用
- **账号池**: `internal/account/` 实现多账号轮询 + 并发队列控制
- **鉴权系统**: `internal/auth/` 支持 Bearer Token 直通和托管账号双模式

### 1.2 目标

将通义千问 (qianwen.com) 的 Web 聊天能力以相同的架构模式集成到 ds2api 中，实现：
- **统一调度**: 复用 ds2api 的账号池、并发控制、请求分发机制
- **多协议兼容**: 通过已有 OpenAI/Claude/Gemini 适配器暴露 Qwen 能力
- **用户资源统一**: 一个 config.json 管理 DeepSeek + Qwen 双平台账号

---

## 二、通义千问 Web API 技术分析

### 2.1 认证机制

| 项目 | 值 |
|---|---|
| **Token 来源** | Cookie `login_tongyi_ticket` / `tongyi_sso_ticket` |
| **认证方式** | `Authorization: Bearer <ticket>` |
| **XSRF 防护** | 服务端 Set-Cookie 返回 `XSRF-TOKEN`，请求需带 `x-xsrf-token` header |
| **设备标识** | `x-deviceid: <UUID>` （浏览器生成一次，后续复用） |
| **平台标识** | `x-platform: pc_tongyi` |
| **多 Token 轮询** | 逗号分隔多个 ticket：`Bearer tok1,tok2,tok3` |

### 2.2 核心 API 端点（从 qianwen.com 实际抓包获取）

```
API 基础地址: https://chat2-api.qianwen.com

┌─────────────────────────────────────────────────────────────┐
│ 接口                    │ 方法   │ 用途                      │
├─────────────────────────────────────────────────────────────┤
│ /api/v1/model/list      │ GET    │ 获取可用模型列表          │
│ /api/v1/session/top/list│ POST   │ 获取置顶会话列表          │
│ /api/v2/session/page/list│ POST  │ 获取分页会话列表          │
│ /api/v1/session/group/list│ POST │ 获取会话分组              │
│ /api/v1/growth/benefit/query │ POST │ 查询用户权益          │
└─────────────────────────────────────────────────────────────┘

OAuth 地址: https://api.qianwen.com/qianwen/havana/oauth/hide
登录地址: https://passport.qianwen.com/havanaone/login/login.htm
```

### 2.3 可用模型列表（2026-04 从 `/api/v1/model/list` 实时获取）

| modelCode | 显示名称 | 特性 | legacyModelCode |
|---|---|---|---|
| `Qwen` | Qwen3.5-千问 | 默认模型，综合AI助手 | `Qwen` |
| `Qwen3.5-Plus` | Qwen3.5-Plus | 多模态 | `tongyi-qwen35-plus-model` |
| `Qwen3.5-Flash` | Qwen3.5-Flash | 快速响应 | `tongyi-qwen3.5-flash` |
| `Qwen3-Max` | Qwen3-Max | 通用均衡 | `tongyi-qwen3-max-model` |
| `Qwen3-Max-Thinking-Preview` | Qwen3-Max-Thinking | 深度推理 | `tongyi-qwen3-max-thinking` |
| `Qwen3-Coder` | Qwen3-Coder | 代码生成 | `tongyi-qwen3-coder` |
| `Qwen3-Flash` | Qwen3-Flash | 简单任务快速 | `tongyi-qwen-flash` |
| `Qwen3-Plus` | Qwen3-Plus | 均衡全能 | `tongyi-qwen-plus-latest` |
| `Qwen3-VL-Plus` | Qwen3-VL-Plus | 视觉理解 | `tongyi-qwen3-vl-plus` |
| `Qwen3-Coder-Flash` | Qwen3-Coder-Flash | 快速代码 | `tongyi-qwen3-coder-flash` |
| `Qwen3-VL-235B-A22B` | Qwen3-VL-235B-A22B | 多模态MoE | `tongyi-qwen3-vl-235b-a22b` |
| `Qwen3-VL-32B` | Qwen3-VL-32B | 视觉稠密 | `tongyi-qwen3-vl-32b` |
| `Qwen3-VL-30B-A3B` | Qwen3-VL-30B-A3B | 视觉MoE | `tongyi-qwen3-vl-30b-a3b-instruct` |
| `Qwen3-235B-A22B-2507` | Qwen3-235B-A22B-2507 | 最强MoE | `tongyi-qwen3-235b-a22b-instruct-2507` |
| `Qwen3-Omni-Flash` | Qwen3-Omni-Flash | 全模态 | `tongyi-qwen3-omni-flash` |
| `Qwen3-Next-80B-A3B` | Qwen3-Next-80B-A3B | 下一代稀疏MoE | `tongyi-qwen3-next-80b-a3b` |
| `Qwen3-30B-A3B-2507` | Qwen3-30B-A3B-2507 | 紧凑MoE | `tongyi-qwen3-30b-a3b-instruct-2507` |

### 2.4 参考开源项目技术要点

来自 `qwen-free-api`(LLM-Red-Team) 和 `qwen3api`(ibootz) 的关键实现：

1. **聊天接口**: 使用 SSE 流式输出，兼容 OpenAI `/v1/chat/completions`
2. **会话管理**: 自动创建 chat_id/session_id，支持自动清理
3. **Token 轮询**: 多 token 逗号分隔，每次请求随机选取
4. **思考模式**: 支持 `thinking_mode.enabled=true` 或模型名后缀 `-thinking`
5. **搜索模式**: 模型名后缀 `-search` 启用联网搜索
6. **无需 PoW**: 与 DeepSeek 不同，Qwen 无 PoW 验证门槛

---

## 三、集成架构设计

### 3.1 整体架构图

```
                        Client (OpenAI/Claude/Gemini SDK)
                                    │
                    ┌───────────────┼───────────────┐
                    ▼               ▼               ▼
             ┌──────────┐    ┌──────────┐    ┌──────────┐
             │ OpenAI   │    │ Claude   │    │ Gemini   │
             │ Adapter  │    │ Adapter  │    │ Adapter  │
             └────┬─────┘    └────┬─────┘    └────┬─────┘
                  │               │               │
                  └───────────────┼───────────────┘
                                  ▼
                       ┌────────────────────┐
                       │  Model Router       │
                       │  (按模型名分发)      │
                       └──┬─────────────┬───┘
                          │             │
                ┌─────────▼──┐  ┌──────▼─────────┐
                │ DeepSeek   │  │  Qwen           │
                │ Client     │  │  Client         │
                │ (现有)      │  │  (新增)         │
                └──────┬─────┘  └──────┬─────────┘
                       │               │
              ┌────────▼──┐    ┌──────▼─────────┐
              │ Account   │    │  Qwen Account  │
              │ Pool      │    │  Pool (新增)   │
              │ (现有)     │    │                │
              └───────────┘    └────────────────┘
```

### 3.2 核心设计原则

1. **最小侵入**: 新增 `internal/qwen/` 包，不修改现有 deepseek 包的任何代码
2. **接口对齐**: 实现 `DeepSeekCaller` 兼容接口，使现有 adapter 可无缝路由到 Qwen
3. **配置统一**: 在 `config.json` 中新增 `qwen_accounts` 字段，复用 keys 鉴权
4. **账号池独立**: Qwen 使用独立的 Account Pool，并发参数独立可配

---

## 四、详细实施方案

### 4.1 新增文件清单

```
internal/
├── qwen/                          # ← 新增包
│   ├── client.go                  # 核心客户端结构体
│   ├── client_auth.go             # 认证/Token管理
│   ├── client_session.go          # 会话创建与管理
│   ├── client_completion.go       # 聊天完成接口(SSE流式)
│   ├── constants.go               # API地址/常量/默认Header
│   ├── prompt.go                  # 消息格式转换(prompt构建)
│   ├── transport/                 # HTTP传输层(可选复用deepseek或独立)
│   │   └── transport.go
│   └── models.go                  # Qwen模型定义与映射
├── config/
│   └── config.go                  # ← 修改：增加 Qwen 配置字段
├── auth/
│   └── request.go                 # ← 修改：增加 Qwen 路由逻辑
├── server/
│   └── router.go                  # ← 修改：注册 Qwen handler
└── adapter/
    └── openai/
        └── handler_chat.go        # ← 修改：增加 Qwen 模型路由
```

### 4.2 配置扩展 (`config.json`)

```json
{
  "keys": ["sk-existing"],
  "accounts": [
    { "email": "deepseek@example.com", "password": "***" }
  ],
  "qwen_accounts": [
    {
      "ticket": "login_tongyi_ticket_value_from_cookie",
      "label": "qwen-account-1"
    },
    {
      "ticket": "another_ticket_value",
      "label": "qwen-account-2"
    }
  ],
  "model_aliases": {
    "gpt-4o": "deepseek-chat",
    "qwen-max": "qwen/Qwen3-Max",
    "qwen-coder": "qwen/Qwen3-Coder"
  },
  "runtime": {
    "account_max_inflight": 2,
    "qwen_account_max_inflight": 2,
    "qwen_account_max_queue": 0,
    "qwen_global_max_inflight": 0
  }
}
```

### 4.3 核心模块设计

#### 4.3.1 Qwen Client (`internal/qwen/client.go`)

```go
type Client struct {
    Store     *config.Store
    Auth      *auth.Resolver
    capture   *devcapture.Store
    regular   trans.Doer
    stream    trans.Doer
    fallback *http.Client
    fallbackS *http.Client
    maxRetries int
}

func NewClient(store *config.Store, resolver *auth.Resolver) *Client
func (c *Client) Preload(ctx context.Context) error
```

**需要实现的接口方法**（对齐 `DeepSeekCaller`）：

```go
// 对齐 DeepSeekCaller 接口 - 使 OpenAI adapter 可直接路由
func (c *Client) CreateSession(ctx context.Context, a *auth.RequestAuth, maxAttempts int) (string, error)
func (c *Client) GetPow(ctx context.Context, a *auth.RequestAuth, maxAttempts int) (string, error)
func (c *Client) CallCompletion(ctx context.Context, a *auth.RequestAuth, payload map[string]any, powResp string, maxAttempts int) (*http.Response, error)
func (c *Client) DeleteAllSessionsForToken(ctx context.Context, token string) error
```

> **注意**: Qwen 无 PoW 机制，`GetPow` 返回空字符串即可。

#### 4.3.2 认证模块 (`internal/qwen/client_auth.go`)

```go
// Token 来自 login_tongyi_ticket cookie，非登录流程获取
// 支持多 token 轮询（通过 account pool）
// Token 存储在 config.QwenAccount.Ticket 字段

func (c *Client) Login(ctx context.Context, acc config.QwenAccount) (string, error)
// 验证 token 有效性（调用轻量接口检测）
func (c *Client) ValidateToken(ctx context.Context, token string) bool
// 刷新 token（如支持的话，否则标记失效切换账号）
func (c *Client) RefreshToken(ctx context.Context, a *auth.RequestAuth) bool
```

#### 4.3.3 会话管理 (`internal/qwen/client_session.go`)

```go
func (c *Client) CreateSession(ctx context.Context, a *auth.RequestAuth, maxAttempts int) (string, error)
func (c *Client) DeleteSession(ctx context.Context, sessionID string, token string) error
func (c *Client) DeleteAllSessions(ctx context.Context, token string) error
```

#### 4.3.4 流式聊天 (`internal/qwen/client_completion.go`)

```go
func (c *Client) CallCompletion(ctx context.Context, a *auth.RequestAuth, payload map[string]any, powResp string, maxAttempts int) (*http.Response, error)
// 内部 SSE 流解析 → 转换为 ds2api 统一的流格式
```

**关键差异点**（vs DeepSeek）：
- Qwen 使用标准 SSE (`data: {...}\n\n`)
- 无需 PoW header
- 响应格式更接近原生 OpenAI 格式（减少转换工作量）
- 可能需要处理 `` 思考标签

#### 4.3.5 常量定义 (`internal/qwen/constants.go`)

```go
const (
    QwenHost           = "chat2-api.qianwen.com"
    QwenLoginURL       = "https://passport.qianwen.com/..."
    QwenModelListURL   = "https://chat2-api.qianwen.com/api/v1/model/list"
    QwenSessionCreateURL = "https://chat2-api.qianwen.com/api/v1/session/create" // 待确认
    QwenChatURL        = "https://chat2-api.qianwen.com/..." // 聊天完成接口（待逆向确认）
)

var QwenBaseHeaders = map[string]string{
    "Host":           "chat2-api.qianwen.com",
    "User-Agent":     "Mozilla/5.0 ...",
    "Content-Type":   "application/json",
    "Accept":         "*/*",
    "x-platform":     "pc_tongyi",
    "x-client-version": "9999.0.0",
}
```

### 4.4 模型路由策略

在现有的 `ResolveModel()` 函数中扩展判断逻辑：

```
请求模型名 → 判断前缀/关键字:
  ├─ "qwen-" 或 "qwen/" 前缀  → 路由到 Qwen Client
  ├─ "deepseek-" 前缀          → 路由到 DeepSeek Client (现有)
  └─ model_aliases 映射         → 按映射结果选择后端
```

**新增 Qwen 模型定义**:

```go
var QwenModels = []ModelInfo{
    {ID: "qwen/Qwen", Object: "model", Created: ..., OwnedBy: "qwen"},
    {ID: "qwen/Qwen3.5-Plus", Object: "model", Created: ..., OwnedBy: "qwen"},
    {ID: "qwen/Qwen3-Max", Object: "model", Created: ..., OwnedBy: "qwen"},
    {ID: "qwen/Qwen3-Max-Thinking", Object: "model", Created: ..., OwnedBy: "qwen"},
    {ID: "qwen/Qwen3-Coder", Object: "model", Created: ..., OwnedBy: "qwen"},
    {ID: "qwen/Qwen3.5-Flash", Object: "model", Created: ..., OwnedBy: "qwen"},
    // ...
}
```

### 4.5 修改现有文件的具体变更

#### `internal/config/config.go` 变更

```go
type Config struct {
    // ... 现有字段 ...
    QwenAccounts []QwenAccount `json:"qwen_accounts,omitempty"`
}

type QwenAccount struct {
    Ticket string `json:"ticket"`           // login_tongyi_ticket
    Label  string `json:"label,omitempty"`  // 可选标识
}

// Store 扩展方法
func (s *Store) QwenAccounts() []QwenAccount
func (s *Store) RuntimeQwenAccountMaxInflight() int
func (s *Store) RuntimeQwenAccountMaxQueue(recommended int) int
func (s *Store) RuntimeQwenGlobalMaxInflight(recommended int) int
```

#### `internal/server/router.go` 变更

```go
func NewApp() *App {
    // ... 现有初始化 ...

    // 新增: Qwen Client 初始化
    qwenClient := qwen.NewClient(store, resolver)

    // 传入 adapter handlers (或使用策略模式选择后端)
    openaiHandler := &openai.Handler{
        Store: store, Auth: resolver,
        DS: dsClient,       // DeepSeek 后端
        QW: qwenClient,     // Qwen 后端 (新增)
    }
    // ...
}
```

#### `internal/adapter/openai/deps.go` 扩展

```go
type Handler struct {
    Store ConfigReader
    Auth  AuthResolver
    DS    DeepSeekCaller     // DeepSeek 后端
    QW    QwenCaller         // Qwen 后端 (新增接口)
    // ...
}

// 新增接口
type QwenCaller interface {
    CreateSession(ctx context.Context, a *auth.RequestAuth, maxAttempts int) (string, error)
    GetPow(ctx context.Context, a *auth.RequestAuth, maxAttempts int) (string, error)
    CallCompletion(ctx context.Context, a *auth.RequestAuth, payload map[string]any, powResp string, maxAttempts int) (*http.Response, error)
    DeleteAllSessionsForToken(ctx context.Context, token string) error
}
```

#### `internal/adapter/openai/handler_chat.go` 变更

```go
func (h *Handler) ChatCompletions(w http.ResponseWriter, r *http.Request) {
    // ... 现有鉴权逻辑 ...

    var req map[string]any
    json.NewDecoder(r.Body).Decode(&req)

    // 新增: 根据模型选择后端
    backend := h.selectBackend(req["model"])

    sessionID, err := backend.CreateSession(...)
    pow, err := backend.GetPow(...)  // Qwen 返回 ""
    resp, err := backend.CallCompletion(...)

    // ... 流式/非流式处理不变 ...
}

func (h *Handler) selectBackend(model any) BackendCaller {
    modelName, _ := model.(string)
    if isQwenModel(modelName) {
        return h.QW
    }
    return h.DS
}
```

### 4.6 Admin API 扩展

在 Admin API 中增加 Qwen 账号管理：

- `POST /admin/accounts/qwen` - 添加 Qwen 账号 (ticket)
- `GET /admin/accounts/qwen` - 列出 Qwen 账号及状态
- `DELETE /admin/accounts/qwen/{label}` - 删除 Qwen 账号
- `POST /admin/test/qwen` - 测试 Qwen 账号有效性
- `GET /admin/queue/status` - 扩展返回 Qwen 池状态

### 4.7 WebUI 扩展

- 设置页面增加「Qwen 账号」Tab
- 支持批量导入 ticket
- 显示 Qwen 模型列表选择
- 账号测试功能

---

## 五、实施步骤（按优先级排序）

### Phase 1: 核心框架搭建（必须）

| 步骤 | 任务 | 涉及文件 |
|------|------|----------|
| 1.1 | 创建 `internal/qwen/` 包骨架 | `qwen/*.go` |
| 1.2 | 定义常量和 API 端点 | `qwen/constants.go` |
| 1.3 | 实现 Qwen Account 配置结构 | `config/config.go` |
| 1.4 | 实现 Qwen Client 基本结构 | `qwen/client.go` |
| 1.5 | 实现 Token 验证逻辑 | `qwen/client_auth.go` |
| 1.6 | 实现会话创建/删除 | `qwen/client_session.go` |

### Phase 2: 聊天能力对接（必须）

| 步骤 | 任务 | 涉及文件 |
|------|------|----------|
| 2.1 | 逆向确认聊天完成接口 URL 和协议 | 研究 |
| 2.2 | 实现 CallCompletion (SSE 流式) | `qwen/client_completion.go` |
| 2.3 | 实现 Prompt 消息格式转换 | `qwen/prompt.go` |
| 2.4 | 实现 Qwen 模型定义和映射 | `qwen/models.go`, `config/models.go` |
| 2.5 | 修改 OpenAI adapter 增加后端路由 | `openai/handler_chat.go`, `deps.go` |
| 2.6 | 修改 Router 注册 Qwen Client | `server/router.go` |

### Phase 3: 账号池与管理（重要）

| 步骤 | 任务 | 涉及文件 |
|------|------|----------|
| 3.1 | 为 Qwen 创建独立的 Account Pool | `account/pool_core.go` 扩展或新建 |
| 3.2 | 修改 Auth Resolver 支持 Qwen 路由 | `auth/request.go` |
| 3.3 | Admin API 增加 Qwen 账号 CRUD | `admin/handler_accounts_crud.go` |
| 3.4 | Admin API 增加 Qwen 账号测试 | `admin/handler_accounts_testing.go` |

### Phase 4: 完善与优化（可选）

| 步骤 | 任务 |
|------|------|
| 4.1 | WebUI 增加 Qwen 账号管理界面 |
| 4.2 | Claude/Gemini adapter 也支持 Qwen 路由 |
| 4.3 | 支持 Qwen 思考模式 (Thinking) 的透传 |
| 4.4 | 支持 Qwen 联网搜索模式 |
| 4.5 | 单元测试和端到端测试 |
| 4.6 | Vercel Node.js 流式路径的 Qwen 适配 |

---

## 六、风险与注意事项

### 6.1 技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 聊天完成接口 URL 未确认 | 阻塞开发 | 需要登录状态下抓包确认实际 chat stream 接口 |
| Token 格式/有效期变化 | 认证失败 | 实现 token 自动验证 + 失效切换 |
| 接口反爬升级 (签名/WAF) | 请求被拦截 | 参考 deepseek 的 utls 方案，可能需要 JS 反混淆 |
| SSE 格式与 DeepSeek 不完全一致 | 解析错误 | 编写专门的 Qwen SSE parser |
| 并发限制未知 | 429 错误 | 先用保守值 (每账号 1-2 并发)，后续调优 |

### 6.2 开发前置条件

1. **必须有有效的 qianwen.com 登录账号** — 用于抓包确认完整聊天接口
2. **需要在已登录状态下发送一条消息** — 抓取实际的 stream 请求/响应
3. **参考 qwen-free-api 源码** — 其 Node.js 实现包含了完整的接口对接细节

### 6.3 关键待确认项（✅ 已全部确认！2026-04-07 实际抓包）

- [x] ~~聊天完成接口的完整 URL path~~ → **`POST https://chat2.qianwen.com/api/v2/chat`**
- [x] ~~请求体的完整 JSON 结构~~ → **已获取完整结构（见下方附录）**
- [x] ~~SSE 响应的 data 字段具体格式~~ → **已获取完整格式（含增量流式）**
- [x] ~~是否有额外的签名参数~~ → **⚠️ 有复杂安全层：nonce+timestamp+bx签名+eo-clt反爬**
- [ ] Token 有效期和多账号轮询的实际限制

---

## 七、🔬 实际抓包数据（2026-04-07 从 qianwen.com 登录状态抓取）

> ⚠️ **重要发现**: Qwen 的反爬机制比 DeepSeek **复杂得多**，存在多层安全校验！

### 7.1 核心聊天接口 ✅ 已确认

```
POST https://chat2.qianwen.com/api/v2/chat
  ?biz_id=ai_qwen
  &chat_client=h5
  &device=pc
  &fr=pc
  &pr=qwen
  &ut=<device_uuid>
  &la=zh-CN
  &tz=Asia/Shanghai
  &nonce=<random_8char>
  &timestamp=<unix_ms>
```

**Response Content-Type**: `text/event-stream;charset=UTF-8` (标准 SSE)

### 7.2 完整请求体 (JSON)

```json
{
  "deep_search": "0",
  "req_id": "962930be419e41e0a795f062ecaa7877",
  "model": "Qwen3-Coder",
  "scene": "chat",
  "session_id": "6c330dfa5257472b958a119c99b634b4",
  "sub_scene": "chat",
  "temporary": false,
  "messages": [
    {
      "content": "AI你好，请用一句话介绍你",
      "mime_type": "text/plain",
      "meta_data": {
        "ori_query": "AI你好，请用一句话介绍你"
      }
    }
  ],
  "from": "default",
  "topic_id": "29e8d4897d444ddcb20c8a88ff2b026a",
  "parent_req_id": "0",
  "scene_param": "first_turn",
  "chat_client": "h5",
  "client_tm": "1775579161064",
  "protocol_version": "v2",
  "biz_id": "ai_qwen"
}
```

### 7.3 SSE 响应格式 (逐帧)

```
event:message
data:{"communication":{"reqid":"...","resid":0,"sessionid":"..."},"data":{"messages":[
  {"mime_type":"signal/post","status":"complete"},
  {"mime_type":"bar/progress","status":"processing"},
  {"content":"我是通义千问，阿里巴巴","mime_type":"multi_load/iframe","status":"processing"}
],"status":"processing"},"success":true}

event:message
data:{"communication":{"resid":1,...},"data":{"messages":[...,
  {"content":"我是通义千问，阿里巴巴集团旗下的超大规模",...,"status":"processing"}
],...}

event:message  (最后一帧)
data:{"communication":{"resid":5,...},"data":{"messages":[...,
  {"content":"我是通义千问，阿里巴巴集团旗下的超大规模语言模型，能够帮助你回答问题、创作文字、表达观点和玩游戏等。",...,"status":"complete"}
],...,"status":"complete"},...}

event:complete
data:true
```

**关键字段解析**:
| 字段 | 含义 |
|---|---|
| `resid` | 响应序号，从 0 递增 |
| `data.messages[0]` | signal/post - 控制信号 |
| `data.messages[1]` | bar/progress - 进度条 |
| `data.messages[2].content` | **增量文本内容**（逐帧追加） |
| `status` | `processing`(进行中) / `complete`(完成) |
| 最终 `event:complete\ndata:true` | 流结束标记 |

### 7.4 ⚠️ 安全层分析（最大技术挑战）

Qwen 使用了 **三层安全防护**，远比 DeepSeek 复杂：

#### 层级 1: 时间戳校准

```
GET https://sec.qianwen.com/api/calibration/getMillisTimeStamp?t=<client_time>
→ 返回: {"data":{"millisTimeStamp":"1775579162598","exLoadBxTimeout":1000}}
```

#### 层级 2: 安全注册（浏览器指纹）

```
POST https://sec.qianwen.com/security/external/access/register?chid=<chid>

Body: {
  "screenResolution": "2560x1440",
  "cookieEnabled": true,
  "localStorageEnabled": true,
  "timezoneOffset": "Asia/Shanghai",
  "fontList": ["Calibri","SimHei",...],
  "pluginList": [{"name":"PDF Viewer",...}],
  "language": ["zh-CN"],
  "fingerprint": "c47c2aa12b59bd9d4797c4792f4e84dd",
  "businessScene": "qwen_web",
  "chid": "..."
}
```

**返回关键 Token**:
- `eo-clt-actkn`: 访问令牌（后续请求必须携带）
- `eo-clt-bacsft`: 加密安全令牌列表（Base64 数组）
- `eo-clt-dvidn`: 设备 ID 编码
- `eo-clt-actkn-dl`: 令牌截止时间戳

#### 层级 3: Quark Web 安全配置

```
GET https://px.wpk.quark.cn/api/v1/jconfig?wpk-header=app=quark-web-security&tm=...&ud=...&sver=2.3.27&sign=...
→ 返回 Quark 反爬配置（流量限制、采样率等）
```

#### 聊天请求中的安全 Headers（必须全部携带）

| Header | 示例值 | 来源 |
|--------|--------|------|
| `bx_et` | `gLIZkF2uNlEZPBPlUTx48sMh1Bx...` (超长) | 百行(Baixing)加密Token |
| `bx-ua` | `231!SIp3d4mUqJ3+jmaA5A3xJq4j...` (超长) | 浏览器指纹UA |
| `bx-umidtoken` | `T2gAk80bo7TTFudPBwzCTXifLYN...` | 设备UMID |
| `eo-clt-acs-ve` | `1.0.0` | 安全协议版本 |
| `eo-clt-actkn` | `1kCqCva6+fiU/4z0cnkFxCtPAk2hLxA1drkFt2P/N36FDONSQ...` | 访问令牌(层级2返回) |
| `eo-clt-sacsft` | `aUBXR0B2Rn9ZX11CNUV+eFsxa3YeHXocXWBiTGpja3cxODA3MDYwMQ==` | 安全令牌(层级2返回) |
| `eo-clt-snver` | `lv` | 签名版本 |
| `eo-clt-dvidn` | `eCy#AAOPBaBB2aJjjojVfrMUwODyNTRB0hBeXXA5nnTiiu10s...` | 设备ID编码 |
| `eo-clt-acs-kp` | `tytk_hash:7751679eeca29dfe39a03e2c9e16c16e` | Ticket Hash |
| `clt-acs-caer` | `vrad` | 反爬引擎标识 |
| `clt-acs-request-params` | `biz_id,chat_client,...` | 签名参数列表 |
| `clt-acs-sign` | `GB10ITWiUDEngkw9zkQSA+TBz+0TEQx2LxiJg8+2dZ8=` | 请求签名 |
| `clt-acs-reqt` | `1775579162059` | 请求时间戳 |
| `x-xsrf-token` | `68a57b92-30ca-4bd1-ba9a-99ea40bde0a4` | XSRF Token(Cookie) |
| `x-chat-id` | `962930be419e41e0a795f062ecaa7877` | 聊天会话UUID |
| `x-deviceid` | `79047695-70cb-c666-f5c7-243765b508df` | 设备UUID |
| `x-platform` | `pc_tongyi` | 平台标识 |

### 7.5 Cookie 中的认证信息

```
tongyi_sso_ticket=1ogROHqapzX0CcdoyjQj$klhtV2M3wGtaHs5PPt9Kx_F0ty14Q2igUFFPW4ybaNzpf8Gz1g_Byns0
tongyi_sso_ticket_hash=tytk_hash:7751679eeca29dfe39a03e2c9e16c16e
XSRF-TOKEN=68a57b92-30ca-4bd1-ba9a-99ea40bde0a4
```

**核心 Token**: `tongyi_sso_ticket` — 这就是用户的登录凭证，对应 qwen-free-api 中提到的 `login_tongyi_ticket`。

---

## 八、实施策略调整（基于抓包结果）

### 8.1 技术难度评估

| 维度 | DeepSeek | Qwen | 差异 |
|------|----------|------|------|
| **认证方式** | 邮箱/手机号+密码登录 | Cookie ticket | Qwen 更简单 |
| **反爬机制** | PoW (WASM计算) | **多层加密+浏览器指纹+签名** | Qwen **显著更复杂** |
| **PoW/WAF** | 需要 WASM 计算 | **无需 PoW，但需要 JS 逆向** | 不同类型困难 |
| **SSE 格式** | 自定义格式 | 标准 SSE + 嵌套 JSON | Qwen 更规范 |
| **会话管理** | 手动创建 session | 自动关联 session_id | 类似 |

### 8.2 推荐实施方案（调整后）

**方案 A: 参考已有开源项目（推荐）**

直接参考 `qwen-free-api` (Node.js) 或 `qwen3-reverse` (Python) 的实现，它们已经解决了：
- bx token 的生成逻辑
- eo-clt 安全头的构造
- 签名算法的逆向
- 完整的会话生命周期管理

**策略**: 在 Go 中复现其核心逻辑，或通过 FFI/子进程调用已有的 Node.js 实现。

**方案 B: 最小化实现（快速验证）**

如果安全层的 JS 逆向过于复杂，可以采用：
1. 使用 **headless browser** (如 chromedp) 模拟完整浏览器环境
2. 通过 CDP 协议直接操控已登录的浏览器实例发送请求
3. 拦截网络响应转换为 API 格式

这种方式牺牲性能但能绕过所有反爬。

### 8.3 文件结构更新（调整后）

```
internal/
├── qwen/
│   ├── client.go               # 核心客户端
│   ├── constants.go            # API地址/默认Header
│   ├── models.go               # 模型定义与映射
│   ├── prompt.go               # OpenAI → Qwen 消息转换
│   ├── security/               # ← 新增：安全层（核心难点）
│   │   ├── register.go         # sec.qianwen.com 安全注册
│   │   ├── timestamp.go        # 时间戳校准
│   │   ├── bx_token.go         # bx token 生成/模拟
│   │   └── signature.go        # 请求签名计算
│   ├── session.go              # 会话管理
│   ├── completion.go           # SSE 流式调用 + 解析
│   └── sse_parser.go           # Qwen SSE 专用解析器
└── ...
```

---

## 九、🔓 安全层深度破解分析（基于源码研究）

> 本章节基于对 `qwen-free-api`（LLM-Red-Team, TypeScript）和 `qwen3-reverse`（wwwzhouhui, Python FastAPI）
> 两个开源项目的**完整源码分析**，结合实际抓包数据，给出可执行的破解方案。

### 9.1 三代 Qwen API 对比

| 维度 | qwen-free-api (旧) | qwen3-reverse (中) | qianwen.com 当前 (新) |
|------|-------------------|-------------------|---------------------|
| **端点** | `qianwen.biz.aliyun.com/dialog/conversation` | `chat.qwen.ai/api/v2/chats/new` | `chat2.qianwen.com/api/v2/chat` |
| **协议** | HTTP/2 (h2) | HTTP/1.1 (requests) | HTTP/2 (浏览器原生) |
| **认证** | Cookie: `tongyi_sso_ticket=xxx` | Cookie + Bearer Token | Cookie + `eo-clt-*` 安全头 |
| **安全头** | ❌ 无 | ⚠️ 硬编码 bx 头 (3个) | 🔒 动态生成 (~15个安全头) |
| **SSE格式** | 自定义 (`[DONE]`) | OpenAI 兼容 | 嵌套JSON (`event:message\ndata:{...}`) |
| **会话管理** | 手动 sessionId | REST API 创建 chat_id | 自动 session_id |
| **状态** | 🚫 已归档(2025-11) | ✅ 活跃 | 🔄 最新版 |

### 9.2 关键发现：bx 头可以硬编码！

**`qwen3-reverse` 的核心突破**：它直接在代码中**硬编码了 bx 安全头**：

```python
# 来自 qwen3-reverse 源码 - QwenClient.__init__()
self.session.headers.update({
    "bx-v": "2.5.31",                              # 百行SDK版本号
    "bx-umidtoken": "T2gAcn1glXMhITXikmXs0OiYrFufhNZzPNYm5sbNWFmnuLgP8Ow4muZZWWKLkXctGU8=",
    # ↑ 设备UMID Token - 可以复用或留空
    "bx-ua": "231!pap3gkmUoC3+j3rAf43qmN4jUq/YvqY2leOxacSC80vTPuB9lMZY9mRWFzrwLEVl7FnY1roS2IxpF9PC+tT6OPC/V1abyEyFxAEaUkxrQ0vccA/tzKw3glZGZSmZh59aXfU4Y5MMXwxnTVZ+/jC4BeXFncDsBa28ZBehEUtIQXxk0ipMY2r/FgC6Na/HA+Uj9Qp+qujynhFxWF7CugwWdsBgD+B34gRr+MNep4Dqk+8t67MMbpXQHJlCok+++4mWYi++6Pamo76GFxBDj+ITHFtd3m4G4R7CN5sgbbtPQepaUgeliRgmWUMcw/rzjJisKKIE3oFnHj5npIyP4H0w2xFthQbQuC/1LAQ2Iq+lrvL6xCS3CI5Giy2exk4LwMJdsmiTpm03B1Cjib62vLA2gk0bsHCo9KykoTD41HO/oqAKDPx5erm9boNvAlKSz6OxdcQ19KSRz/rfuwb8IKhL0zcjcFhl3",
    # ↑ 浏览器指纹UA - 包含屏幕分辨率、字体列表、插件等信息的哈希
})
```

**这意味着**：
- ✅ `bx-v` 是固定版本号，可直接使用
- ✅ `bx-ua` 是浏览器指纹哈希，**可以硬编码复用**（来自某台真实浏览器的指纹）
- ✅ `bx-umidtoken` 是设备标识，可以留空或使用固定值
- ⚠️ 但当前最新版 `chat2.qianwen.com` 还有额外的 `eo-clt-*` 和 `clt-acs-*` 头

### 9.3 qwen-free-api 认证机制（最简方案）

```typescript
// 来自 qwen-free-api 源码 - generateCookie()
function generateCookie(ticket: string) {
  return [
    `${ticket.length > 100 ? 'login_aliyunid_ticket' : 'tongyi_sso_ticket'}=${ticket}`,
    'aliyun_choice=intl',
    "_samesite_flag_=true",
    `t=${uuid(false)}`,  // 随机UUID作为t参数
  ].join("; ");
}

// FAKE_HEADERS 极其简单 - 只有基本浏览器伪装
const FAKE_HEADERS = {
  Accept: "application/json, text/plain, */*",
  "X-Platform": "pc_tongyi",
  "X-Xsrf-Token": "48b9ee49-a184-45e2-9f67-fa87213edcdc",
  Referer: "https://tongyi.aliyun.com/",
  "User-Agent": "Mozilla/5.0 ...",
};
```

**关键发现**：旧版 API 只需要 **1 个 cookie** (`tongyi_sso_ticket`) 即可工作！无需任何复杂的安全头。

### 9.4 安全层分级破解策略

根据以上分析，我们采用**渐进式策略**：

#### Level 0: 直接用旧版 API（最快验证）

```
端点: https://qianwen.biz.aliyun.com/dialog/conversation
认证: Cookie: tongyi_sso_ticket=<ticket>
安全头: 仅基础 UA + X-Platform
复杂度: ★☆☆☆☆☆
```

→ **推荐先实现此版本验证可行性**，参考 qwen-free-api 的完整源码。

#### Level 1: 使用 chat.qwen.ai + 硬编码 bx 头

```
端点: https://chat.qwen.ai/api/v2/chats/new (创建会话)
      https://chat.qwen.ai/api/v2/chats/completions (发送消息)
认证: Cookie + Bearer Token
安全头: 硬编码 bx-v/bx-ua/bx-umidtoken (从 qwen3-reverse 提取)
复杂度: ★★☆☆☆☆
```

→ 参考 `qwen3-reverse` 的 Python 实现。

#### Level 2: 完整适配 chat2.qianwen.com（最终目标）

```
端点: https://chat2.qianwen.com/api/v2/chat
认证: tongyi_sso_ticket cookie + eo-clt 安全令牌
安全头: 完整15+个动态安全头
复杂度: ★★★★☆☆
```

→ 需要实现 sec.qianwen.com 安全注册流程或 JS 逆向。

### 9.5 推荐实施路径（调整后）

```
Phase 1 (Level 0): 适配旧版 API ──────────┐
    ↓ 复用 qwen-free-api 逻辑                │ 预计 2-3 小时
    ↓ 验证 Qwen Web→API 可行性               │
                                              ↓
Phase 2 (Level 1): 升级到 chat.qwen.ai ───────┤ 核心功能可用
    ↓ 复用 qwen3-reverse 的 bx 头            │
    ↓ 支持更多模型                           │ 预计 3-5 小时
                                              ↓
Phase 3 (Level 2): 完整 chat2 适配 ───────────┘ 最终目标
    ↓ 实现 sec.qianwen.com 注册流程           │
    ↓ 或 headless browser 方案               │ 预估 1-2 天
```

### 9.6 Go 实现核心伪代码（基于 Level 0/1 方案）

```go
// internal/qwen/client.go
package qwen

type Client struct {
    baseURL    string
    httpClient *http.Client
    tickets    []string  // 多 ticket 轮询
    bxHeaders  map[string]string // 硬编码的 bx 安全头
}

func NewClient(tickets []string) *Client {
    return &Client{
        baseURL:    "https://qianwen.biz.aliyun.com",  // Level 0
        httpClient: &http.Client{Timeout: 120 * time.Second},
        tickets:    tickets,
        bxHeaders: map[string]string{
            "bx-v":         "2.5.31",
            "bx-ua":        hardcodedBXUA,   // 从 qwen3-reverse 提取
            "bx-umidtoken": hardcodedUMID,
        },
    }
}

func (c *Client) buildHeaders(ticket string) http.Header {
    hdr := http.Header{}
    hdr.Set("Content-Type", "application/json")
    hdr.Set("Accept", "text/event-stream")
    hdr.Set("X-Platform", "pc_tongyi")
    hdr.Set("Cookie", c.buildCookie(ticket))
    // 复制 bx 安全头
    for k, v := range c.bxHeaders {
        hdr.Set(k, v)
    }
    return hdr
}

func (c *Client) buildCookie(ticket string) string {
    key := "tongyi_sso_ticket"
    if len(ticket) > 100 {
        key = "login_aliyunid_ticket"
    }
    return fmt.Sprintf("%s=%s; aliyun_choice=intl; _samesite_flag_=true; t=%s",
        key, ticket, uuidNoDash())
}
```

---

## 十、预期成果

完成后，用户可以通过以下方式使用 Qwen：

```bash
# 通过 OpenAI 兼容接口使用 Qwen
curl http://localhost:5001/v1/chat/completions \
  -H "Authorization: Bearer your-ds2api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen/Qwen3-Max",
    "messages": [{"role": "user", "content": "你好"}],
    "stream": true
  }'

# 通过 Claude 兼容接口使用 Qwen（映射后）
curl http://localhost:5001/anthropic/v1/messages \
  -H "x-api-key: your-ds2api-key" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-sonnet-4-5",  # 映射到 qwen/Qwen3-Max
    "messages": [{"role": "user", "content": "你好"}]
  }'
```

**核心价值**: 用户只需维护一个 ds2api 实例、一个 config.json、一套 API Key，即可同时调度 DeepSeek 和 Qwen 两平台的免费 Web 资源。
