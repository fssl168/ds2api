# ds2api MCP Bridge for OpenClaw

[![npm version](https://img.shields.io/npm/v/ds2api-mcp-openclaw.svg)](https://www.npmjs.com/package/ds2api-mcp-openclaw)

Connect **DeepSeek** & **Qwen (通义千问)** AI models to [OpenClaw](https://github.com/openclaw) via Model Context Protocol.

## Features

- 5 MCP Tools optimized for OpenClaw agent workflows
- 10 AI Models: DeepSeek (4) + Qwen/通义千问 (6)
- Dual transport: Streamable HTTP + Stdio
- Auto-retry and health check built-in
- Zero-config: auto-discovers ds2api at `localhost:5001`

## Quick Install

```bash
npm install -g ds2api-mcp-openclaw
```

## Configuration

### Method 1: Streamable HTTP (Recommended)

Copy `openclaw.json` to your OpenClaw config directory:

```json
{
  "mcpServers": {
    "ds2api": {
      "type": "streamable_http",
      "url": "http://127.0.0.1:5001/mcp",
      "headers": {
        "Authorization": "Bearer YOUR_API_KEY"
      }
    }
  }
}
```

### Method 2: Stdio Mode

```json
{
  "mcpServers": {
    "ds2api": {
      "command": "npx",
      "args": ["ds2api-mcp-openclaw"],
      "env": {
        "DS2API_BASE_URL": "http://127.0.0.1:5001"
      }
    }
  }
}
```

### Generate Config Template

```bash
ds2api-openclaw --config > openclaw.json
```

## Usage in OpenClaw

Once configured, OpenClaw agents can use these tools:

| Tool | Description | Example |
|------|-------------|---------|
| **chat** | Send message to AI model | `"Use deepseek-reasoner to analyze this code"` |
| **list_models** | List available models | `"What models are available?"` |
| **get_status** | Check service health | `"Is ds2api running?"` |
| **get_pool_status** | Monitor account pools | `"How many accounts are free?"` |
| **embeddings** | Generate embeddings | `"Embed this document"` |

## Available Models

### DeepSeek Engine

| Model ID | Best For |
|----------|----------|
| `deepseek-chat` | General coding, chat |
| `deepseek-reasoner` | Complex reasoning (thinking) |
| `deepseek-chat-search` | Chat with web search |
| `deepseek-reasoner-search` | Reasoning with web search |

### Qwen (通义千问) Engine

| Model ID | Best For |
|----------|----------|
| `qwen-plus` | Balanced Chinese/English |
| `qwen-max` | High-quality Chinese tasks |
| `qwen-coder` | Code generation |
| `qwen-flash` | Fast responses, low cost |
| `qwen3.5-plus` | Latest Qwen, best quality |
| `qwen3.5-flash` | Latest Qwen, fast |

## Programmatic API

```typescript
import { OpenClawClient } from "ds2api-mcp-openclaw";

const client = new OpenClawClient({
  baseURL: "http://127.0.0.1:5001",
  apiKey: "your-key",
});

// Chat with DeepSeek
const reply = await client.chat([
  { role: "user", content: "Explain async/await" }
], "deepseek-chat");

// List models
const models = await client.listModels();

// Health check
const healthy = await client.healthCheck();
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DS2API_BASE_URL` | `http://127.0.0.1:5001` | ds2api server URL |
| `DS2API_API_KEY` | `` | Authentication key |

## Publishing

```bash
cd publish/openclaw
npm run build
npm publish --access public
```

## License

MIT © [fssl168](https://github.com/fssl168)
