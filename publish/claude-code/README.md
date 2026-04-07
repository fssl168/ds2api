# ds2api MCP Bridge for Claude Code

[![npm](https://img.shields.io/npm/v/ds2api-mcp-claude.svg)](https://www.npmjs.com/package/ds2api-mcp-claude)

Connect **DeepSeek** & **Qwen (通义千问)** AI models to [Claude Code](https://claude.ai/code) via MCP.

## Features

- 5 MCP tools optimized for Claude Code workflows
- 10 AI Models: DeepSeek (4) + Qwen (6)
- Thinking mode compatible (`max_tokens` support)
- Auto `.claude/settings.json` config generation
- CLAUDE.md rules generation
- Dual transport: stdio + streamable HTTP

## Install

```bash
npm install -g ds2api-mcp-claude
```

## Quick Setup

### One-line setup (stdio):

```bash
ds2api-claude --config >> ~/.claude/settings.json
```

Or manually add to `~/.claude/settings.json`:

```jsonc
{
  "mcpServers": {
    "ds2api": {
      "command": "npx",
      "args": ["ds2api-mcp-claude"],
      "env": { "DS2API_BASE_URL": "http://127.0.0.1:5001" }
    }
  }
}
```

### Generate CLAUDE.md rules:

```bash
ds2api-claude --rules > .claude/CLAUDE.md
```

## Usage Examples in Claude Code

After connecting, ask Claude Code:

> **"Use deepseek-reasoner to explain this Rust async pattern"**
> → Calls `chat` tool with `deepseek-reasoner` model

> **"What models are available?"** 
> → Calls `list_models`, shows all 10 models

> **"Is the API healthy? How many accounts?"**
> → Calls `get_status` + `get_pool_status`

> **"Explain this in Chinese using best quality model"**
> → Uses `qwen3.5-plus` automatically

## Available Models

| Model | Engine | Best For |
|-------|--------|----------|
| `deepseek-chat` | DeepSeek | General coding |
| `deepseek-reasoner` | DeepSeek | Complex reasoning |
| `qwen-plus` | Qwen | Balanced CN/EN |
| `qwen-max` | Qwen | High-quality CN |
| `qwen-coder` | Qwen | Code gen |
| `qwen3.5-plus` | Qwen | Latest, best |
| `qwen3.5-flash` | Qwen | Latest, fast |

## Programmatic API

```typescript
import { ClaudeCodeClient } from "ds2api-mcp-claude";

const client = new ClaudeCodeClient({ baseURL: "http://127.0.0.1:5001" });
await client.chat([{ role: "user", content: "Hello!" }]);
const models = await client.listModels();
```

## Publish

```bash
cd publish/claude-code
npm install && npm run build
npm publish --access public
```
