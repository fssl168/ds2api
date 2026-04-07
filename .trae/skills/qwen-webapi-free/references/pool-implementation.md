# QwenPool 账号池实现详解

## 1. 设计目标

对标 DeepSeek 的 AccountPool，为通义千问提供相同级别的多账号管理能力：

- 多 ticket 轮询分配
- 并发控制（每 ticket 最大 in-flight 数）
- 全局并发上限
- 等待队列（池满时不立即拒绝）
- 健康检查（成功/失败追踪）
- 自动冷却（连续失败后临时移除）
- 热重载（配置变更后无需重启）

## 2. 核心数据结构

### 2.1 QwenPoolEntry — 池条目

```go
type QwenPoolEntry struct {
    Label         string     // 显示标签（如 "qwen-account-1"）
    Ticket        string     // 通义千问凭证
    FailCount     int        // 连续失败计数
    CooldownUntil time.Time // 冷却截止时间（零值表示未冷却）
}
```

### 2.2 QwenPool — 池管理器

```go
type QwenPool struct {
    store                  *config.Store      // 配置存储引用
    mu                     sync.RWMutex       // 读写锁
    entries                []*QwenPoolEntry   // 所有账号条目
    inUse                  map[string]int      // label → 当前使用次数
    waiters                []chan struct{}     // 等待队列
    maxInflightPerTicket   int                 // 每 ticket 最大并发（默认 2）
    globalMaxInflight      int                 // 全局最大并发
    maxQueueSize           int                 // 等待队列上限
    recommendedConcurrency int                 // 建议并发值
}
```

## 3. 生命周期

### 3.1 初始化

```
NewQwenPool(store)
  → Reset()
    → 从 store.Snapshot().QwenAccounts 读取配置
    → 按 ticket 非空过滤 + 排序（有 ticket 优先，按 label 字典序）
    → 构建 entries 列表
    → 计算 recommendedConcurrency = len(entries) × maxInflightPerTicket
    → maxQueueSize = recommendedConcurrency × 2（最少 4）
    → globalMaxInflight = recommendedConcurrency
```

### 3.2 重载

当配置变更（新增/删除/修改 qwen_accounts）时调用 `Reset()`：

- 保留已有 entry 的 FailCount 和 CooldownUntil 状态
- 重新计算并发参数
- 清空等待队列（drainWaitersLocked）

### 3.3 Acquire — 获取账号

```
Acquire(ctx):
  1. 加写锁
  2. 调用 acquireLocked() 尝试直接获取
     → 遍历 entries：
       - 跳过 CooldownUntil > now 的
       - 跳过 inUse[label] >= maxInflightPerTicket 的
       - 跳过 global in-flight >= globalMaxInflight 的
       - 找到：inUse[label]++，bumpQueue（移到队尾优化轮询），返回 entry
     → 未找到且 canQueue：创建 waiter channel，加入队列，阻塞 select 等待
       - ctx.Done → 移除 waiter，返回 ctx.Err()
       - waiter 被 close → 重新 acquireLocked()
     → 未找到且队列满：返回 ErrQwenPoolExhausted
```

### 3.4 Release — 释放账号

```
Release(label):
  1. 加写锁
  2. 查找 inUse[label]
  3. count <= 0 → 直接返回（防止重复释放）
  4. count == 1 → delete from map，notifyWaiterLocked() 唤醒一个等待者
  5. count > 1 → count--，notifyWaiterLocked() 唤醒一个等待者
```

### 3.5 MarkSuccess / MarkFailed — 健康追踪

```
MarkSuccess(label):
  → 找到 entry，FailCount > 0 则递减（恢复健康）

MarkFailed(label):
  → 找到 entry，FailCount++
  → FailCount >= maxFailureCountBeforeCooldown (3):
      cooldown = 30s + failCount * 10s
      设置 CooldownUntil = now + cooldown
      日志警告：label, fail_count, cooldown_seconds
```

**冷却时间公式**：`30 + failCount * 10` 秒

| 连续失败次数 | 冷却时间 |
|------------|---------|
| 3 次 | 60 秒 |
| 4 次 | 70 秒 |
| 5 次 | 80 秒 |
| ... | ... |

### 3.6 RandomEntry — 无状态随机选取

用于不需要 Acquire/Release 的场景（如 pickTicket 获取凭证用于 SM Register）：

```
RandomEntry():
  → 读锁
  → 过滤掉 CooldownUntil > now 的 entries
  → 从可用 entries 中随机选一个返回
  → 无可用 → 返回 nil
```

## 4. 等待队列机制

### 4.1 数据结构

使用 `[]chan struct{}` 作为 FIFO 队列：

```go
waiters []chan struct{}  // 等待者 channel 列表
```

### 4.2 入队

```go
waiter := make(chan struct{})
p.waiters = append(p.waiters, waiter)
// 在 goroutine 中阻塞: <-waiter
```

### 4.3 出队（通知）

```go
func (p *QwenPool) notifyWaiterLocked() {
    if len(p.waiters) == 0 { return }
    waiter := p.waiters[0]           // 取队首
    p.waiters = p.waiters[1:]        // 出队
    close(waiter)                   // close channel 唤醒等待者
}
```

### 4.4 取消等待

ctx 取消时从队列中移除：

```go
func (p *QwenPool) removeWaiterLocked(waiter chan struct{}) bool {
    for i, w := range p.waiters {
        if w == waiter {
            p.waiters = append(p.waiters[:i], p.waiters[i+1:]...)
            return true
        }
    }
    return false
}
```

### 4.5 排空（Reset 时）

```go
func (p *QwenPool) drainWaitersLocked() {
    for _, waiter := range p.waiters {
        close(waiter)  // 全部唤醒（会得到 ErrQwenPoolExhausted）
    }
    p.waiters = nil
}
```

## 5. 调度优化：bumpQueue

每次 Acquire 成功后，将使用的 entry 移到 entries 列表末尾：

```go
func (p *QwenPool) bumpQueue(label string) {
    for i, e := range p.entries {
        if e.Label != label { continue }
        entry := p.entries[i]
        p.entries = append(p.entries[:i], p.entries[i+1:]...)
        p.entries = append(p.entries, entry)  // 移到末尾
        return
    }
}
```

**效果**：最近使用的账号被排到后面，下次 Acquire 优先选择其他账号，实现**近似 LRU 的均匀分配**。

## 6. 并发控制参数

### 6.1 默认值

```go
const (
    defaultMaxInflightPerTicket    = 2     // 每 ticket 默认最大并发
    defaultCooldownSeconds         = 30    // 基础冷却秒数
    maxFailureCountBeforeCooldown = 3     // 触发冷却的连续失败次数
)
```

### 6.2 自动计算

```go
func defaultRecommendedQwenConcurrency(entryCount, maxInflight int) int {
    if entryCount <= 0 { return 0 }
    if maxInflight <= 0 { maxInflight = defaultMaxInflightPerTicket }
    return entryCount * maxInflight
}
```

| tickets 数 | maxInflight | recommended | queue size | global limit |
|-----------|-------------|-------------|-----------|-------------|
| 1 | 2 | 2 | 4 | 2 |
| 2 | 2 | 4 | 8 | 4 |
| 3 | 2 | 6 | 12 | 6 |
| 5 | 2 | 10 | 20 | 10 |
| 5 | 5 | 25 | 50 | 25 |

### 6.3 运行时调整

通过 `SetLimits()` 可热更新参数：

```go
pool.SetLimits(maxInflightPerTicket, maxQueueSize, globalMaxInflight)
```

调整后会自动 `notifyWaitersLocked()` 唤醒可能满足条件的等待者。

## 7. 状态监控

### 7.1 Status() 返回值

```go
map[string]any{
    "available":                 int    // 可用账号数（未冷却 & 未达并发上限）
    "in_use":                    int    // 当前占用槽位数
    "total":                     int    // 总账号数
    "available_accounts":        []string  // 可用账号 label 列表
    "in_use_accounts":           []string  // 使用中账号 label 列表
    "cooldown_accounts":         int       // 冷却中账号数
    "max_inflight_per_ticket":   int       // 每 ticket 并发上限
    "global_max_inflight":       int       // 全局并发上限
    "recommended_concurrency":   int       // 建议并发值
    "waiting":                   int       // 等待队列中的请求数
    "max_queue_size":            int       // 队列容量上限
}
```

### 7.2 Admin API 端点

```
GET /admin/qwen-pool/status  →  调用 pool.Status()
```

前端展示：可用 / 使用中 / 冷却中 / 等待中 各状态的数量和详情。

## 8. 与 DeepSeek Pool 的对比

| 特性 | DeepSeek Pool | QwenPool |
|------|---------------|----------|
| 凭证类型 | email/password → token | ticket（直接使用） |
| 登录流程 | 需要模拟登录获取 token | 无需登录，ticket 即凭证 |
| PoW 计算 | 需要 WASM PoW | 不需要 |
| Session 管理 | CreateSession/DeleteSession | 空实现（L2 协议无 session 概念） |
| Acquire/Release | ✅ | ✅ |
| 等待队列 | ✅ | ✅ |
| 健康检查 | ✅ | ✅ |
| 自动冷却 | ✅ | ✅ |
| 并发控制 | per-ticket + global | per-ticket + global |
| bumpQueue 调度 | ✅ | ✅ |
| RandomEntry | ✅ | ✅ |
| Reset/Reload | ✅ | ✅ |
| Status 监控 | ✅ | ✅ |

## 9. 集成到 CallCompletion 的完整流程

```go
func (c *Client) CallCompletion(ctx context.Context, ...) (*http.Response, error) {
    // 1. Acquire
    entry, err := c.pool.Acquire(ctx)
    if err != nil { return nil, err }
    ticket := entry.Ticket
    label := entry.Label
    defer c.pool.Release(label)

    // 2. 构造请求
    // ... (build payload, generate IDs, etc.)

    // 3. 带重试的调用
    for attempt := 0; attempt < maxAttempts; attempt++ {
        if attempt > 0 {
            // 重试时重新 acquire（可能换一个 ticket）
            retryEntry, _ := c.pool.Acquire(ctx)
            ticket = retryEntry.Ticket
            defer c.pool.Release(retryEntry.Label)
        }

        resp, err := c.doPostChat(ctx, ticket, chatID, jsonBody)
        if err != nil {
            c.pool.MarkFailed(currentLabel)  // 失败标记
            continue                       // 继续重试
        }
        c.pool.MarkSuccess(currentLabel)     // 成功标记
        return resp, nil
    }
    return nil, fmt.Errorf("all attempts failed")
}
```

**关键设计**：
- `defer Release` 确保无论成功失败都释放
- 重试时重新 Acquire（可能获得不同的 ticket）
- 每次 API 调用结果都反馈给池（MarkSuccess/MarkFailed）
