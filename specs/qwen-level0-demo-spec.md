# Spec: Qwen Web-to-API Level 0 Demo 集成

## 1. 目标
在 ds2api 项目中实现通义千问(Qwen) Level 0 方案的 Web 转 API Demo，验证技术可行性。
使用 `qianwen.biz.aliyun.com` 旧版 API（参考 qwen-free-api 源码），通过 `tongyi_sso_ticket` cookie 认证，
将 Qwen 的聊天能力以 OpenAI 兼容格式暴露，复用 ds2api 现有的账号池、并发控制、请求分发体系。

## 2. 技术方案概述

### 2.1 API 端点（Level 0 - 旧版）
- **聊天**: `POST https://qianwen.biz.aliyun.com/dialog/conversation`
- **会话删除**: `POST https://qianwen.biz.aliyun.com/dialog/session/delete`
- **协议**: HTTP/2 (h2)，SSE 流式响应

### 2.2 认证方式
- Cookie: `tongyi_sso_ticket=<ticket值>; aliyun_choice=intl; _samesite_flag_=true; t=<uuid>`
- 无需额外安全头（bx/eo-clt 等均为旧版不需要）
- 多 ticket 支持: 逗号分隔，每次请求随机选取一个

### 2.3 SSE 响应格式（来自 qwen-free-api 源码分析）
```
event: data (隐式)
data: {"sessionId":"xxx","msgId":"xxx","msgStatus":"processing","contents":[{"contentType":"text","role":"assistant","content":"增量文本"}]}
... (多帧)
data: {"sessionId":"xxx","msgId":"xxx","msgStatus":"finished","contents":[...]}
event: data (隐式)
data: [DONE]
```

### 2.4 请求体格式
```json
{
  "mode": "chat",
  "model": "",
  "action": "next",
  "userAction": "chat",
  "requestId": "<uuid>",
  "sessionId": "<existing_or_empty>",
  "sessionType": "text_chat",
  "parentMsgId": "<parent_or_empty>",
  "params": {"fileUploadBatchId": "<uuid>"},
  "contents": [
    {
      "content": "合并后的用户消息文本",
      "contentType": "text",
      "role": "user"
    }
  ]
}
```

## 3. 新增文件清单

```
internal/qwen/
├── constants.go          # API 地址、默认 Header、模型定义
├── client.go             # 核心客户端结构体 + NewClient()
├── client_auth.go        # Cookie 构建、Token 管理/验证
├── client_session.go     # 会话创建/删除（Demo 中 session 由服务端自动管理）
├── client_completion.go  # 流式聊天调用 + SSE 解析
├── prompt.go             # OpenAI messages → Qwen contents 格式转换
└── models.go            # 模型列表定义（用于 /v1/models 接口）

修改现有文件:
├── internal/config/config.go          # 增加 QwenAccounts 字段
├── internal/config/account.go         # 增加 QwenAccount 结构体
├── internal/server/router.go          # 注册 Qwen handler
├── internal/adapter/openai/deps.go   # 增加 QW 字段 + QwenCaller 接口
└── internal/adapter/openai/handler_chat.go  # 模型路由到 Qwen 后端
```

## 4. 配置扩展

config.json 新增字段:
```json
{
  "qwen_accounts": [
    {
      "ticket": "从 qianwen.com cookie 获取的 tongyi_sso_ticket 值",
      "label": "qwen-account-1"
    }
  ]
}
```

## 5. 模型路由规则
- 请求 model 名以 `qwen/` 或 `qwen-` 开头 → 路由到 Qwen Client
- 可用模型名: `qwen/Qwen`, `qwen/Qwen-Max`, `qwen/qwen-turbo` 等（Demo 统一映射为空字符串 model，由 Qwen 服务端决定实际模型）
- 不指定 model 或匹配别名时走 DeepSeek（保持向后兼容）

## 6. 接口契约

### 6.1 QwenCaller 接口（对齐 DeepSeekCaller）
```go
type QwenCaller interface {
    CreateSession(ctx context.Context, a *auth.RequestAuth, maxAttempts int) (string, error)
    GetPow(ctx context.Context, a *auth.RequestAuth, maxAttempts int) (string, error)  // 返回空字符串
    CallCompletion(ctx context.Context, a *auth.RequestAuth, payload map[string]any, powResp string, maxAttempts int) (*http.Response, error)
    DeleteAllSessionsForToken(ctx context.Context, token string) error
}
```

### 6.2 错误处理
- Token 无效 → 返回 401 + 切换下一个 ticket
- 网络超时 → 重试（最多 3 次）
- SSE 解析错误 → 记录日志并返回已接收的部分内容
- 会话删除失败 → 静默忽略（不阻塞主流程）

## 7. Demo 验收标准
1. `GET /v1/models` 返回包含 Qwen 模型的列表
2. `POST /v1/chat/completions` 使用 `model: "qwen/Qwen"` 能成功调用 Qwen 并返回流式响应
3. 流式响应格式兼容 OpenAI SDK（可被 curl / OpenAI Python 库正确消费）
4. 多 ticket 配置时能轮询切换
5. 编译通过 (`go build ./...`) 且无新增 lint 错误

## 8. 非功能需求
- 不修改任何现有 deepseek 包代码
- 复用现有的 account pool 并发控制模式
- 保持与现有 OpenAI/Claude/Gemini adapter 相同的错误处理风格
- 所有新代码遵循项目现有的 Go 代码规范
