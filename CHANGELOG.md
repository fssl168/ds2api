# DS2API 开发日志 (CHANGELOG)

所有 notable changes 都记录在此文件。

格式基于 [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

---

## [Unreleased]

### 新增

#### 客户端侧上下文保护系统 (`internal/limits/`)

- **新增包** `internal/limits/`，提供完整的客户端侧上下文窗口管理
- **环境变量开关**: `CONTEXT_PROTECTION=true|1|on`（默认关闭，向后兼容）
- **模型限制定义** ([models.go](internal/limits/models.go)):
  - 7 个模型族的上下文窗口、最大输出、字符/Token 比
  - Qwen3.5-plus/flash: 128K / 8192 out
  - Qwen-max: 32K / 2048 out
  - Qwen-coder: 32K / 4096 out
  - Qwen-flash: 128K / 2048 out
  - Qwen-plus: 32K / 2048 out
  - DeepSeek-reasoner: 64K / 8192 out
  - DeepSeek-default: 64K / 8192 out

- **5 项保护功能** ([protection.go](internal/limits/protection.go)):

  | # | 功能 | 说明 |
  |---|------|------|
  | 1 | **消息数截断** | 超过 `MaxMessages` 时保留最新消息，丢弃最旧的 |
  | 2 | **单消息截断** | 单条消息超过 `MaxSingleMsgSize`（字符）时 rune 安全截断，附加 `[...truncated]` 标记 |
  | 3 | **Token 估算** | 基于 `charsPerToken` 比率进行字符→Token 粗估算（UTF-8 rune 计数） |
  | 4 | **max_tokens 钳制** | 用户请求的 max_tokens 被限制在 `[min(请求值, 模型最大值, 可用上下文), 256]` 区间内 |
  | 5 | **使用率告警** | 当 `(input_tokens + output_tokens) / context_window > 95%` 时输出警告日志 |

- **集成点**: [handler_chat.go](internal/adapter/openai/handler_chat.go) — OpenAI 统一入口，鉴权后、引擎分发前执行保护检查，对 Qwen 和 DeepSeek 双路径均生效
- **Qwen max_tokens 透传**: [client_completion.go](internal/qwen/client_completion.go) — 将钳制后的 max_tokens 写入通义千问 API payload

#### WebUI 审计日志页面

- **后端**: [audit.go](internal/admin/audit.go) — 环形缓冲区（200 条），线程安全 RWMutex
  - 全局单例 `globalAudit`
  - `GET /admin/audit-log?limit=100` API 端点
  - 9 个关键操作全覆盖审计：add/delete key/account/qwen-account/password/config/import
- **前端**: [AuditLogContainer.jsx](webui/src/features/audit/AuditLogContainer.jsx)
  - 自动刷新（10s 可切换）
  - 操作类型图标 + 颜色编码
  - 时间本地化 + 来源 IP 脱敏显示
  - 侧边栏 Shield 图标导航入口

### 安全加固

| # | 类别 | 修复 | 文件 |
|---|------|------|------|
| S-01 | DoS 防护 | MCP Server/Transport 请求体添加 `MaxBytesReader(10MB)` | [mcp/server.go](internal/mcp/server.go), [mcp/transport.go](internal/mcp/transport.go) |
| S-02 | 信息泄露 | 500 错误响应移除 `err.Error()` 内部细节，仅返回通用消息 | [router.go](internal/server/router.go) |
| S-03 | CORS 策略 | 支持环境变量 `CORS_ORIGIN` 配置白名单（非空时记录安全日志） | [router.go](internal/server/router.go) |
| S-04 | Session ID | 从固定可预测序列改为 `crypto/rand` + hex 编码 | [mcp/transport.go](internal/mcp/transport.go) |
| S-05 | 速率限制 | 新增 `rateLimiter` 中间件（120 req/min/IP，自动清理 goroutine） | [router.go](internal/server/router.go) |
| S-06 | API Key 安全 | query 参数传递 api_key 时输出 `[security]` 警告日志 | [auth/request.go](internal/auth/request.go) |
| S-07 | 审计日志 | Admin 9 个关键操作写入结构化审计日志 | [handler.go](internal/admin/handler.go) + 4 handler 文件 |

### 其他变更

- MCP 目录迁移: `internal/plugins/mcp/` → `internal/mcp/`
- README.MD 重写: 注明原作者 CJackHwang/ds2api，移除 PUBLISH.md 章节
- `.gitignore`: 排除 `.trae/`, `publish/`, `packages/`

---

## [2026-04-08] 初始版本

### 核心

- DeepSeek 多账号池（邮箱/手机登录、自动 token 刷新、PoW WASM 计算）
- 通义千问 Qwen 账号池（Acquire/Release 并发控制、健康检查）
- OpenAI 兼容接口 (`/v1/*`)
- Claude 兼容接口 (`/anthropic/*`)
- Gemini 兼容接口 (`/v1beta/*`)
- Tool Calling 防泄漏处理（多格式解析）

### MCP 桥接层

- JSON-RPC 2.0 Server（initialize/tools/list/tools/call/ping）
- Streamable HTTP / SSE / Stdio 三种传输模式
- 5 个工具：chat / list_models / get_status / get_pool_status / embeddings
- 5 平台注册：OpenClaw / Claude Code / JetBrains / OpenCode / VS Code

### 管理

- WebUI 管理台（中英文双语、深色模式）
- Admin API（配置 CRUD / 账号测试 / 批量导入导出）
- 运维探针（healthz / readyz）
