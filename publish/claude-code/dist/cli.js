#!/usr/bin/env node
import { startClaudeCodeStdio } from "./server.js";
import { createRequire } from "node:module";
const require = createRequire(import.meta.url);
const pkg = require("../package.json");
const args = process.argv.slice(2);
if (args.includes("--help") || args.includes("-h")) {
    console.log(`
ds2api-claude - MCP Bridge for Claude Code

Usage:
  ds2api-claude                    Start stdio mode (for Claude Code)
  ds2api-claude --help             Show help
  ds2api-claude --version          Show version
  ds2api-claude --config           Generate ~/.claude/settings.json config
  ds2api-claude --rules            Generate CLAUDE.md rules

Claude Code Integration:

  Method 1: Stdio (Recommended)
  Add to ~/.claude/settings.json:
  {
    "mcpServers": {
      "ds2api": {
        "command": "npx",
        "args": ["ds2api-mcp-claude"],
        "env": { "DS2API_BASE_URL": "http://127.0.0.1:5001" }
      }
    }
  }

  Method 2: Remote HTTP
  {
    "mcpServers": {
      "ds2api": {
        "type": "streamable_http",
        "url": "http://127.0.0.1:5001/mcp"
      }
    }
  }

Environment Variables:
  DS2API_BASE_URL    ds2api URL (default: http://127.0.0.1:5001)
  DS2API_API_KEY     API key

Examples in Claude Code:
  > Use deepseek-reasoner to analyze this Rust code
  > List all available models
  > Check if the API is healthy
  > How many accounts are in the pool?
`);
    process.exit(0);
}
if (args.includes("--version") || args.includes("-v")) {
    console.log(`ds2api-claude v${pkg.version}`);
    process.exit(0);
}
if (args.includes("--config")) {
    const config = {
        mcpServers: {
            ds2api: {
                command: "npx",
                args: ["ds2api-mcp-claude"],
                env: { DS2API_BASE_URL: "http://127.0.0.1:5001" },
            },
        },
        ds2api: {
            baseURL: "http://127.0.0.1:5001",
            defaultModel: "deepseek-chat",
            maxTokens: 16384,
            thinking: false,
            preferredEngine: "auto",
        },
    };
    console.log(JSON.stringify(config, null, 2));
    process.exit(0);
}
if (args.includes("--rules")) {
    const rules = `# ds2api MCP Rules for Claude Code

You have access to ds2api MCP tools providing DeepSeek and Qwen AI models.

## Available Tools
- **chat**: Send messages to DeepSeek or Qwen models
- **list_models**: List all 10 available models
- **get_status**: Check service health
- **get_pool_status**: Monitor account pools  
- **embeddings**: Generate text embeddings

## Model Selection Guide
| Task | Model | Why |
|------|-------|-----|
| General coding | deepseek-chat | Fast, reliable |
| Complex reasoning | deepseek-reasoner | Thinking mode |
| Chinese tasks | qwen-max / qwen3.5-plus | Best Chinese quality |
| Code generation | qwen-coder | Specialized for code |
| Quick answers | qwen-flash / qwen3.5-flash | Fastest |

## Workflow Tips
1. For code analysis, prefer deepseek-reasoner for step-by-step reasoning
2. For Chinese user queries, default to qwen-max or qwen3.5-plus
3. Always check get_status before heavy operations
4. Use list_models to confirm availability when suggesting models
5. ds2api runs locally at http://127.0.0.1:5001
`;
    console.log(rules);
    process.exit(0);
}
startClaudeCodeStdio().catch((err) => {
    console.error("[ds2api-claude] Fatal:", err);
    process.exit(1);
});
//# sourceMappingURL=cli.js.map