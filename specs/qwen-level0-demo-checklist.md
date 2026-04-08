# Checklist: Qwen Level 0 Demo

## 代码质量
- [ ] 所有新文件遵循项目 Go 编码规范
- [ ] 无硬编码密钥或 token
- [ ] 错误处理完整（网络错误、认证失败、SSE 解析错误）
- [ ] 日志记录关键操作（请求发送、响应接收、token 切换）

## 功能完整性
- [ ] GET /v1/models 包含 Qwen 模型
- [ ] POST /v1/chat/completions model=qwen/Qwen 能成功调用
- [ ] 流式 SSE 响应格式正确（data: {...}\n\n 格式）
- [ ] 非流式模式也能正常工作
- [ ] 多 ticket 轮询切换正常
- [ ] Token 失效时自动切换到下一个

## 兼容性
- [ ] 不修改现有 deepseek 包任何代码
- [ ] 现有 DeepSeek 功能不受影响
- [ ] config.json 向后兼容（qwen_accounts 为可选字段）
- [ ] OpenAI Python/JS SDK 可正常消费

## 安全
- [ ] Ticket 不在日志中明文输出
- [ ] 输入参数做基本校验
- [ ] 无 SQL 注入 / XSS 风险
