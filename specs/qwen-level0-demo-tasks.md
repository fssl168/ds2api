# Tasks: Qwen Level 0 Demo

## T1: 项目骨架与常量定义
- [x] T1.1 创建 `internal/qwen/` 目录
- [ ] T1.2 实现 `constants.go` - API 地址、默认 Header、模型列表
- [ ] T1.3 实现 `models.go` - Qwen 模型定义（用于 /v1/models）

## T2: 核心客户端
- [ ] T2.1 实现 `client.go` - Client 结构体 + NewClient() 初始化
- [ ] T2.2 实现 `client_auth.go` - Cookie 构建 (generateCookie)、Token 验证 (getTokenLiveStatus)

## T3: 会话管理
- [ ] T3.1 实现 `client_session.go` - CreateSession (返回空字符串，由服务端自动创建)
- [ ] T3.2 实现 `client_session.go` - DeleteSession (removeConversation 调用)
- [ ] T3.3 实现 `client_session.go` - DeleteAllSessionsForToken

## T4: 流式聊天（核心）
- [ ] T4.1 实现 `client_completion.go` - CallCompletion (POST /dialog/conversation, HTTP/2)
- [ ] T4.2 实现 SSE 解析器 - 解析 qwen-free-api 格式的 SSE 事件流
- [ ] T4.3 实现非流式模式 (receiveStream) - 等待完整响应后返回

## T5: 消息格式转换
- [ ] T5.1 实现 `prompt.go` - OpenAI messages[] → Qwen contents[] 格式转换
- [ ] T5.2 多轮对话合并逻辑（参考 qwen-free-api 的 messagesPrepare）

## T6: 配置与集成
- [ ] T6.1 修改 `config/config.go` - 增加 QwenAccounts 字段和访问方法
- [ ] T6.2 修改 `config/account.go` - 增加 QwenAccount 结构体
- [ ] T6.3 修改 `adapter/openai/deps.go` - 增加 QW 字段 + QwenCaller 接口定义
- [ ] T6.4 修改 `adapter/openai/handler_chat.go` - 增加模型路由逻辑 (selectBackend)
- [ ] T6.5 修改 `server/router.go` - 初始化 QwenClient 并注入 handler

## T7: 编译验证与测试
- [ ] T7.1 `go build ./...` 编译通过
- [ ] T7.2 `go vet ./...` 无警告
- [ ] T7.3 手动端到端测试: curl 发送请求验证流式输出
