import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { ClaudeCodeOptions } from "./types.js";
import { ClaudeCodeClient } from "./client.js";
declare function createClaudeCodeServer(options?: ClaudeCodeOptions): McpServer;
export declare function startClaudeCodeStdio(options?: ClaudeCodeOptions): Promise<void>;
export { createClaudeCodeServer, ClaudeCodeClient };
export type { ClaudeCodeOptions, ClaudeCodeConfig, ClaudeCodeToolCall } from "./types.js";
//# sourceMappingURL=server.d.ts.map