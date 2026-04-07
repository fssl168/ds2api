# ds2api-mcp

[![npm version](https://img.shields.io/npm/v/ds2api-mcp.svg)](https://www.npmjs.com/package/ds2api-mcp)
[![license](https://img.shields.io/npm/l/ds2api-mcp.svg)](https://github.com/fssl168/ds2api/blob/main/LICENSE)

**ds2api MCP Bridge** - Connect [DeepSeek](https://deepseek.com) & [Qwen (通义千问)](https://qwenlm) AI models to any **MCP-compatible** client via Model Context Protocol.

## Features

- 5 MCP Tools: `chat`, `list_models`, `get_status`, `get_pool_status`, `embeddings`
- 10 AI Models: DeepSeek (4) + Qwen/通义千问 (6)
- Official [@modelcontextprotocol/sdk](https://www.npmjs.com/package/@modelcontextprotocol/sdk) based
- Stdio transport for Claude Code, OpenCode, and other stdio-based clients
- HTTP client SDK for programmatic use
- TypeScript first with full type definitions
- Zero-config: auto-connects to ds2api at `http://127.0.0.1:5001`

## Quick Start

### Prerequisites

- Node.js >= 18
- [ds2api](https://github.com/fssl168/ds2api) running locally or remotely

### Install

```bash
npm install -g ds2api-mcp
```

### Usage as CLI (Stdio Mode)

```bash
ds2api-mcp
```

This starts the MCP server in stdio mode, ready for any MCP client to connect.

## Platform Integration

### Claude Code

Edit `~/.claude/settings.json`:

```json
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

Then restart Claude Code. Ask it:

> "What models are available?" → calls `list_models`
> "Use deepseek-reasoner to explain Rust async" → calls `chat`

### VS Code

Create `.vscode/mcp.json` in your project:

```json
{
  "servers": {
    "ds2api": {
      "type": "streamable-http",
      "url": "http://127.0.0.1:5001/mcp"
    }
  }
}
```

VS Code will auto-connect when you open the project.

### OpenCode

Add to your opencode config (`~/.config/opencode/config.json`):

```json
{
  "mcp": {
    "servers": [
      {
        "name": "ds2api",
        "command": "npx",
        "args": ["ds2api-mcp"]
      }
    ]
  },
  "openai_compat": {
    "base_url": "http://127.0.0.1:5001",
    "api_key": "your-api-key"
  }
}
```

### JetBrains IDEs

Settings → Tools → Model Context Protocol → Servers → Add:

| Field | Value |
|-------|-------|
| Name | ds2api |
| Type | Streamable HTTP |
| URL | `http://127.0.0.1:5001/mcp` |

### OpenClaw

```json
{
  "mcpServers": {
    "ds2api": {
      "type": "streamable_http",
      "url": "http://127.0.0.1:5001/mcp"
    }
  }
}
```

## Programmatic Usage (SDK)

### Client API

```typescript
import { Ds2ApiClient } from "ds2api-mcp";

const client = new Ds2ApiClient({
  baseURL: "http://127.0.0.1:5001",
  apiKey: "your-key",
});

// Chat completion
const reply = await client.chat({
  model: "qwen-plus",
  messages: [{ role: "user", content: "Hello!" }],
});
console.log(reply); // "HiHi"

// List models
const models = await client.listModels();
console.log(models);
// [{ id: "deepseek-chat", name: "DeepSeek Chat", engine: "deepseek" }, ...]

// Service status
const status = await client.getStatus();

// Pool status
const pools = await client.getPoolStatus({ pool_type: "all" });

// Embeddings
const emb = await client.embeddings({ input: ["hello world"] });
```

### Custom Server (Advanced)

```typescript
import { createServer, StdioServerTransport } from "ds2api-mcp";

const server = createServer({ baseURL: "http://your-ds2api:5001" });
const transport = new StdioServerTransport();
await server.connect(transport);
```

## Available Models

| Model ID | Engine | Best For |
|----------|--------|----------|
| `deepseek-chat` | DeepSeek | General coding & chat |
| `deepseek-reasoner` | DeepSeek | Complex reasoning (thinking mode) |
| `deepseek-chat-search` | DeepSeek | Chat + web search |
| `deepseek-reasoner-search` | DeepSeek | Reasoning + web search |
| `qwen-plus` | Qwen | Balanced Chinese/English |
| `qwen-max` | Qwen | High-quality Chinese tasks |
| `qwen-coder` | Qwen | Code generation |
| `qwen-flash` | Qwen | Fast responses, low cost |
| `qwen3.5-plus` | Qwen | Latest Qwen 3.5, best quality |
| `qwen3.5-flash` | Qwen | Latest Qwen 3.5, fast |

## MCP Tools Reference

### chat

Send a chat completion request.

```typescript
{
  model: string;        // required - model name
  messages: Array<{     // required
    role: "user" | "assistant" | "system";
    content: string;
  }>;
  stream?: boolean;     // optional - stream response
  temperature?: number; // optional - 0-2
  max_tokens?: number;  // optional - max tokens to generate
}
```

### list_models

List all available models. No parameters required.

### get_status

Get service health, config, and endpoint info. No parameters required.

### get_pool_status

Get account pool utilization details.

```typescript
{
  pool_type?: "deepseek" | "qwen" | "all";  // default: "all"
}
```

### embeddings

Generate text embeddings.

```typescript
{
  input: string[];       // required - array of texts
  model?: string;        // optional - embedding model
}
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DS2API_BASE_URL` | `http://127.0.0.1:5001` | ds2api server URL |
| `DS2API_API_KEY` | `` | API key for authentication |

## Architecture

```
┌──────────────┐     ┌─────────────────┐     ┌──────────────────┐
│  MCP Client   │────▶│  ds2api-mcp npm │────▶│   ds2api Server  │
│ (Claude Code │ SDK │  (this package)  │ MCP │  (Go backend)     │
│  / VS Code / │     │                  │     │                  │
│  OpenCode)   │     │  5 tools         │     │  DeepSeek Engine  │
│              │     │  10 models       │     │  Qwen Engine     │
└──────────────┘     └─────────────────┘     └──────────────────┘
```

## Publishing to npm

```bash
cd packages/ds2api-mcp
npm run build
npm publish --access public
```

## License

MIT © [fssl168](https://github.com/fssl168)
