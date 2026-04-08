# ds2api MCP 插件发布指南

## 发布矩阵

| 平台 | 包类型 | 发布目标 | 发布命令 | 状态 |
|------|--------|----------|----------|------|
| **npm (SDK)** | Node.js 包 | [npmjs.com/ds2api-mcp](https://www.npmjs.com/package/ds2api-mcp) | `npm publish` (packages/ds2api-mcp) | ✅ 已就绪 |
| **OpenClaw** | Node.js 包 | [npmjs.com/ds2api-mcp-openclaw](https://www.npmjs.com/package/ds2api-mcp-openclaw) | `npm publish` (publish/openclaw) | ✅ 已就绪 |
| **Claude Code** | Node.js 包 | [npmjs.com/ds2api-mcp-claude-code](https://www.npmjs.com/package/ds2api-mcp-claude-code) | `npm publish` (publish/claude-code) | ✅ **新增** |
| **Cursor** | `.vsix` 扩展 | Cursor Marketplace / VSIX | `vsce package` (publish/cursor) | ✅ **新增** |
| **VS Code** | `.vsix` 扩展 | [marketplace.visualstudio.com](https://marketplace.visualstudio.com) | `vsce package` + `vsce publish` (publish/vscode) | ✅ 已就绪 |
| **JetBrains** | 插件 zip | [plugins.jetbrains.com](https://plugins.jetbrains.com) | `./gradlew buildPlugin` (publish/jetbrains) | ✅ 已就绪 |
| **OpenCode** | 配置文档 | GitHub README | — | 文档即可 |

---

## 1. npm 发布（Node.js 生态）

### 前置条件
- npm 账号 ([npmjs.com 注册](https://www.npmjs.com/signup))
- 已登录: `npm login`

### 发布步骤

```bash
cd packages/ds2api-mcp

# 1. 构建
npm run build

# 2. 预览包内容
npm pack --dry-run

# 3. 发布
npm publish --access public
```

### 用户安装使用

```bash
# 全局安装 CLI
npm install -g ds2api-mcp
ds2api-mcp --version

# 或在项目中使用 SDK
npm install ds2api-mcp
```

### Claude Code 集成（用户侧）

```jsonc
// ~/.claude/settings.json
{
  "mcpServers": {
    "ds2api": {
      "command": "npx",
      "args": ["ds2api-mcp"],
      "env": { "DS2API_BASE_URL": "http://127.0.0.1:5001" }
    }
  }
}
```

---

## 1.5. OpenClaw 发布（Agent 平台）

### 前置条件
- npm 账号（与 npm SDK 共用）
- [OpenClaw](https://github.com/openclaw) 已安装

### 发布步骤

```bash
cd publish/openclaw

# 1. 安装依赖
npm install

# 2. 构建 TypeScript
npm run build

# 3. 生成配置模板（可选）
node dist/cli.js --config > openclaw.json

# 4. 发布到 npm
npm publish --access public
```

### 用户安装使用

```bash
npm install -g ds2api-mcp-openclaw
ds2api-openclaw --version
```

### OpenClaw 集成配置（用户侧）

**方式 A: Streamable HTTP**
```jsonc
// openclaw.json
{
  "mcpServers": {
    "ds2api": {
      "type": "streamable_http",
      "url": "http://127.0.0.1:5001/mcp"
    }
  }
}
```

**方式 B: Stdio 模式**
```jsonc
{
  "mcpServers": {
    "ds2api": {
      "command": "npx",
      "args": ["ds2api-mcp-openclaw"]
    }
  }
}
```

### 包结构
```
publish/openclaw/
├── package.json              ← npm 包配置 (bin: ds2api-openclaw)
├── tsconfig.json
├── openclaw.json             ← 配置模板
├── README.md                 ← OpenClaw 专用文档
└── src/
    ├── index.ts              ← 主入口
    ├── client.ts             ← OpenClawClient (healthCheck + callTool)
    ├── server.ts             ← MCP Server (5 tools, OpenClaw 优化)
    ├── cli.ts                ← CLI (--config / --help / --version)
    └── types.ts              ← 类型定义
```

### 功能特性
- ✅ `OpenClawClient` 带 `healthCheck()` 方法
- ✅ 自动重试和超时控制
- ✅ `--config` 一键生成 openclaw.json 配置
- ✅ 双模式: streamable_http + stdio
- ✅ 完整类型定义 (OpenClawConfig/OpenClawToolCall/OpenClawToolResult)

---

## 2. VS Code Marketplace 发布

### 前置条件
- Azure DevOps 组织 (免费创建)
- VS Code Marketplace 发布者身份 (`vsce login`)

### 发布步骤

```bash
cd publish/vscode

# 1. 安装依赖
npm install

# 2. 编译 TypeScript
npm run compile

# 3. 打包 .vsix
npx vsce package

# 4. 发布到 Marketplace
npx vsce publish
```

### 打包产物
```
publish/vscode/
├── ds2api-mcp-1.0.0.vsix    ← 上传此文件
├── package.json               ← 扩展清单
└── src/
    ├── extension.ts           ← 激活逻辑：状态栏、命令、TreeDataProvider
    └── modelsProvider.ts      ← 模型列表树视图
```

### 功能特性
- ✅ 状态栏显示连接状态（绿色=已连接，红色=断开）
- ✅ 侧边栏模型浏览器（10 个模型，按引擎分组）
- ✅ 命令面板集成（`Ctrl+Shift+P` → ds2api:）
- ✅ 设置页面（URL / API Key / Log Level）
- ✅ 自动心跳检测（30 秒间隔）

### 用户安装方式
- Marketplace: 搜索 `ds2api-mcp`
- 或手动: 下载 `.vsix` → `Code: Install Extension VSIX...`
- 安装后自动连接 `http://127.0.0.1:5001/mcp`

---

## 2.5. Cursor IDE 发布（Agent 模式增强）

### 前置条件
- [Cursor IDE](https://cursor.sh) 已安装
- VS Code 扩展兼容（Cursor 基于 VS Code）

### 发布步骤

```bash
cd publish/cursor

# 1. 安装依赖
npm install

# 2. 编译
npm run compile

# 3. 打包 .vsix
npx vsce package

# 4. 发布
# 方式 A: Cursor Marketplace
npx vsce publish
# 方式 B: 手动分发 ds2api-mcp-cursor-1.0.0.vsix
```

### 用户安装

```bash
# 在 Cursor 中: Extensions → "..." → Install from VSIX
# 选择 ds2api-mcp-cursor-1.0.0.vsix
```

### 包结构
```
publish/cursor/
├── package.json              ← 扩展清单 (含 Agent Mode 配置)
├── src/
│   ├── extension.ts          ← 激活: 状态栏 + 6命令 + .cursor/rules 注入
│   ├── modelsProvider.ts     ← 侧边栏模型树
│   └── chatPanel.ts          ← 内置 Chat 面板 Webview
└── README.md
```

### Cursor 特有功能
- ✅ **Agent Mode 集成**: 自动注入 `.cursor/rules/ds2api.mdc`
- ✅ **右键 Chat**: 选中代码 → `ds2api: Chat with DeepSeek/Qwen`
- ✅ **Chat Panel**: 侧边栏内置聊天 UI（10 模型选择器）
- ✅ **模型浏览器**: TreeView 展示全部 10 个模型
- ✅ **30s 心跳**: 自动检测连接状态

---

## 2.6. Claude Code 发布（独立 npm 包）

### 前置条件
- npm 账号
- Claude Code CLI 已安装

### 发布步骤

```bash
cd publish/claude-code

# 1. 安装依赖 & 构建
npm install && npm run build

# 2. 发布到 npm
npm publish --access public
```

### 用户安装使用

```bash
npm install -g ds2api-mcp-claude

# 一键配置 Claude Code
ds2api-claude --config >> ~/.claude/settings.json

# 生成 CLAUDE.md 规则
ds2api-claude --rules > .claude/CLAUDE.md
```

### 包结构
```
publish/claude-code/
├── package.json              ← bin: ds2api-claude
├── claude.json               ← 配置模板 (含 thinking/maxTokens)
├── README.md
└── src/
    ├── index.ts / client.ts  ← ClaudeCodeClient (healthCheck)
    ├── server.ts             ← MCP Server (Claude Code 优化描述)
    └── cli.ts                ← CLI (--config/--rules/--help/--version)
```

### Claude Code 特有功能
- ✅ `--config` 一键生成 `~/.claude/settings.json`
- ✅ `--rules` 生成 CLAUDE.md 规则文件
- ✅ Thinking mode 兼容 (`maxTokens` 参数支持)
- ✅ 工具描述针对 Claude Code 场景定制
- ✅ 双模式: stdio (推荐) + streamable_http

---

## 3. JetBrains Marketplace 发布

### 前置条件
- JetBrains 账号 ([account.jetbrains.com](https://account.jetbrains.com))
- IntelliJ IDEA / Android Studio（用于构建插件）

### 发布步骤

```bash
cd publish/jetbrains

# 方式 A: 使用 Gradle（推荐）
./gradlew buildPlugin
# 产物: build/distributions/ds2api-mcp-1.0.0.zip

# 方式 B: 手动打包为 zip
zip -r ds2api-mcp.zip src/ META-INF/

# 上传到 https://plugins.jetbrains.com/plugin/publish
```

### 插件结构
```
publish/jetbrains/
└── src/main/
    ├── resources/META-INF/plugin.xml   ← 插件清单
    └── kotlin/ds2api/mcp/
        ├── PluginMain.kt              ← 设置持久化 + Action
        └── Ds2ApiMcpConfigurable.kt     ← Settings 页面 UI
```

### 功能特性
- ✅ Settings → Tools → ds2api MCP Bridge 配置页
- ✅ Tools 菜单 → Show Connection Status
- ✅ Tools 菜单 → List Available Models
- ✅ 连接状态通知（Balloon）
- ✅ 设置持久化（IDE 重启后保留配置）

### 用户安装方式
- JetBrains IDE: Settings → Plugins → Marketplace → 搜索 `ds2api-mcp`
- 或从磁盘安装: Settings → Plugins → ⚙️ → Install Plugin from Disk...

---

## 4. Claude Code / OpenCode（社区发布）

这两个平台没有正式的插件市场，通过 **README + 配置示例** 发布：

### Claude Code 推荐配置

```bash
# 一行安装体验
npx -y ds2api-mcp
```

```jsonc
// ~/.claude/settings.json（推荐方式）
{
  "mcpServers": {
    "ds2api": {
      "command": "npx",
      "args": ["ds2api-mcp"],
      "env": {
        "DS2API_BASE_URL": "http://127.0.0.1:5001"
      }
    }
  }
}
```

### OpenCode 推荐配置

```jsonc
// ~/.config/opencode/config.json
{
  "mcp": {
    "servers": [{
      "name": "ds2api",
      "command": "npx",
      "args": ["ds2api-mcp"]
    }]
  },
  "openai_compat": {
    "base_url": "http://127.0.0.1:5001",
    "api_key": ""
  }
}
```

---

## 5. 发布检查清单

### 发布前

- [ ] `go build ./...` 编译零错误
- [ ] `go vet ./...` 无警告
- [ ] MCP 测试全部通过（10/10）
- [ ] npm 包 `npm run build` 成功
- [ ] VS Code 扩展 `vsce package` 成功
- [ ] JetBrains 插件 `buildPlugin` 成功
- [ ] README.MD 更新各平台安装说明

### 发布操作

```bash
# ① npm
cd packages/ds2api-mcp && npm publish --access public

# ② VS Code
cd publish/vscode && npx vsce publish

# ③ JetBrains
# 手动上传 zip 到 https://plugins.jetbrains.com/plugin/publish

# ④ Git Tag
git tag v1.0.0-mcp
git push origin v1.0.0-mcp
```

### 发布后验证

- [ ] npm: `npm install -g ds2api-mcp && ds2api-mcp --version`
- [ ] VS Code: 在新窗口安装扩展，确认侧边栏和状态栏出现
- [ ] JetBrains: 在新 IDE 实例中安装，确认 Settings 页面出现
- [ ] Claude Code: `claude` 启动后 `/mcp` 可见 ds2api 工具

---

## 各平台链接汇总

| 平台 | 链接 |
|------|------|
| **GitHub 仓库** | https://github.com/fssl168/ds2api |
| **npm SDK 包** | https://www.npmjs.com/package/ds2api-mcp |
| **npm OpenClaw 包** | https://www.npmjs.com/package/ds2api-mcp-openclaw |
| **npm Claude Code 包** | https://www.npmjs.com/package/ds2api-mcp-claude-code |
| **Cursor 扩展** | (VSIX / Marketplace) |
| **VS Code Marketplace** | https://marketplace.visualstudio.com/items?itemName=fssl168.ds2api-mcp |
| **JetBrains Marketplace** | https://plugins.jetbrains.com/plugin/[id] |
| **Issues** | https://github.com/fssl168/ds2api/issues |
