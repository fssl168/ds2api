#!/usr/bin/env node

import { startStdioServer } from "./server.js";
import { createRequire } from "node:module";

const require = createRequire(import.meta.url);
const pkg = require("../package.json");

const args = process.argv.slice(2);

if (args.includes("--help") || args.includes("-h")) {
  console.log(`
ds2api-mcp - MCP Bridge for ds2api

Usage:
  ds2api-mcp                    Start in stdio mode (for Claude Code, OpenCode, etc.)
  ds2api-mcp --help             Show this help
  ds2api-mcp --version          Show version

Environment Variables:
  DS2API_BASE_URL    ds2api server URL (default: http://127.0.0.1:5001)
  DS2API_API_KEY     API key for authentication

Platform Configuration:

  Claude Code (~/.claude/settings.json):
  {
    "mcpServers": {
      "ds2api": {
        "command": "npx",
        "args": ["ds2api-mcp"],
        "env": { "DS2API_BASE_URL": "http://127.0.0.1:5001" }
      }
    }
  }

  VS Code (.vscode/mcp.json):
  {
    "servers": {
      "ds2api": {
        "type": "streamable-http",
        "url": "http://127.0.0.1:5001/mcp"
      }
    }
  }

  OpenCode (opencode.json):
  {
    "mcp": {
      "servers": [{
        "name": "ds2api",
        "command": "npx",
        "args": ["ds2api-mcp"]
      }]
    }
  }
`);
  process.exit(0);
}

if (args.includes("--version") || args.includes("-v")) {
  console.log(`ds2api-mcp v${pkg.version}`);
  process.exit(0);
}

startStdioServer().catch((err) => {
  console.error("[ds2api-mcp] Fatal error:", err);
  process.exit(1);
});
