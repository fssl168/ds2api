# ds2api MCP Bridge for Cursor

[![VS Code](https://img.shields.io/badge/Cursor-Compatible-green)](https://cursor.sh)

Connect **DeepSeek** & **Qwen (通义千问)** AI models to [Cursor](https://cursor.sh) IDE via Model Context Protocol.

## Features

- **Agent Mode Integration**: DeepSeek/Qwen models work with Cursor's Agent mode
- **Chat Panel**: Built-in chat UI in sidebar (select text → right-click → Chat)
- **Model Browser**: Tree view of all 10 available models
- **Auto .cursor/rules**: Automatically injects ds2api MCP rules
- **Status Bar**: Real-time connection status
- **Health Check**: 30-second auto heartbeat

## Installation

### From VSIX

```bash
# Download from releases or build:
cd publish/cursor
npm install && npm run compile
npx vsce package
```

Then in Cursor: `Extensions: Install from VSIX...` → select `ds2api-mcp-cursor-1.0.0.vix`

### From Marketplace

Search `ds2api-mcp-cursor` in Cursor extensions marketplace.

## Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `ds2api-mcp-cursor.url` | `http://127.0.0.1:5001/mcp` | MCP endpoint |
| `ds2api-mcp-cursor.apiKey` | `` | API key |
| `ds2api-mcp-cursor.defaultModel` | `deepseek-chat` | Default model |
| `ds2api-mcp-cursor.agentMode` | `true` | Enable agent mode support |
| `ds2api-mcp-cursor.autoInjectRules` | `true` | Auto-inject `.cursor/rules` |

## Usage

### Chat with Selection

1. Select code in editor
2. Right-click → **ds2api: Chat with DeepSeek** or **Chat with Qwen**
3. View response in sidebar chat panel

### Agent Mode

When Cursor's Agent mode is enabled, the injected `.cursor/rules/ds2api.mdc` file guides the AI to use ds2api MCP tools:

> User: "Explain this async Rust pattern using deepseek-reasoner"
> 
> Agent will automatically call the `chat` tool with the selected text

### Model Browser

Open sidebar → **ds2api** panel → **Models** tab shows all 10 models.

## Available Models

| Model | Engine | Best In |
|-------|--------|---------|
| deepseek-chat | DeepSeek | General coding |
| deepseek-reasoner | DeepSeek | Complex reasoning |
| qwen-max | Qwen | Chinese tasks |
| qwen-coder | Qwen | Code gen |
| ... | ... | ... |

## Requirements

- Cursor IDE (based on VS Code 1.97+)
- ds2api running at `localhost:5001`
