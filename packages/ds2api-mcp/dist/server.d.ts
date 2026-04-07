import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { Ds2ApiClientOptions } from "./types.js";
import { Ds2ApiClient } from "./client.js";
declare function createServer(options?: Ds2ApiClientOptions): McpServer;
export declare function startStdioServer(options?: Ds2ApiClientOptions): Promise<void>;
export { createServer, Ds2ApiClient };
export type { Ds2ApiClientOptions, ChatParams, ChatMessage, ModelInfo, ServiceStatus, PoolStatus } from "./types.js";
//# sourceMappingURL=server.d.ts.map