import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import type { Ds2ApiOpenClawOptions } from "./types.js";
import { OpenClawClient } from "./client.js";
declare function createOpenClawServer(options?: Ds2ApiOpenClawOptions): McpServer;
export declare function startOpenClawStdio(options?: Ds2ApiOpenClawOptions): Promise<void>;
export { createOpenClawServer, OpenClawClient };
export type { Ds2ApiOpenClawOptions, OpenClawConfig, OpenClawToolCall, OpenClawToolResult } from "./types.js";
//# sourceMappingURL=server.d.ts.map