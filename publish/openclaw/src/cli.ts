#!/usr/bin/env node

import { startOpenClawStdio } from "./server.js";
import { createRequire } from "node:module";

const require = createRequire(import.meta.url);
const pkg = require("../package.json");

const args = process.argv.slice(2);

if (args.includes("--help") || args.includes("-h")) {
  console.log(`
ds2api-openclaw - MCP Bridge for OpenClaw Agent Platform

Usage:
  ds2api-openclaw                  Start stdio mode (for OpenClaw subprocess)
  ds2api-openclaw --help           Show help
  ds2api-openclaw --version        Show version

OpenClaw Integration:

  Add to your openclaw configuration (openclaw.json):
  {
    "mcpServers": {
      "ds2api": {
        "type": "streamable_http",
        "url": "http://127.0.0.1:5001/mcp"
      }
    }
  }

  Or use stdio mode:
  {
    "mcpServers": {
      "ds2api": {
        "command": "npx",
        "args": ["ds2api-mcp-openclaw"]
      }
    }
  }

Environment Variables:
  DS2API_BASE_URL    ds2api URL (default: http://127.0.0.1:5001)
  DS2API_API_KEY     API key

For more info: https://github.com/fssl168/ds2api
`);
  process.exit(0);
}

if (args.includes("--version") || args.includes("-v")) {
  console.log(`ds2api-openclaw v${pkg.version}`);
  process.exit(0);
}

if (args.includes("--config")) {
  const configTemplate = {
    mcpServers: {
      ds2api: {
        type: "streamable_http",
        url: "http://127.0.0.1:5001/mcp",
        headers: { Authorization: "Bearer YOUR_API_KEY" },
      },
    },
    ds2api: {
      baseURL: "http://127.0.0.1:5001",
      defaultModel: "deepseek-chat",
      timeout: 120000,
    },
  };
  console.log(JSON.stringify(configTemplate, null, 2));
  process.exit(0);
}

startOpenClawStdio().catch((err) => {
  console.error("[ds2api-openclaw] Fatal:", err);
  process.exit(1);
});
