import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import { ClaudeCodeClient } from "./client.js";
function createClaudeCodeServer(options) {
    const client = new ClaudeCodeClient(options);
    const server = new McpServer({
        name: "ds2api-mcp-claude-code",
        version: "1.0.0",
    }, {
        capabilities: { tools: {} },
    });
    server.tool("chat", `Send chat completion to ds2api for Claude Code.
Supports DeepSeek and Qwen models with optional streaming.

Recommended for Claude Code:
- Use deepseek-reasoner for complex reasoning tasks
- Use qwen-max or qwen3.5-plus for Chinese language tasks
- Use qwen-coder for code generation`, {
        model: z.string().describe("Model: deepseek-chat, deepseek-reasoner, qwen-plus, qwen-max, etc."),
        messages: z.array(z.object({
            role: z.enum(["user", "assistant", "system"]),
            content: z.string(),
        })).describe("Chat conversation"),
        stream: z.boolean().optional().describe("Stream response (default: false)"),
        temperature: z.number().optional().describe("Sampling temperature 0-2"),
        max_tokens: z.number().int().optional().describe("Max tokens to generate"),
    }, async (params) => {
        try {
            const text = await client.chat(params.messages, params.model);
            return { content: [{ type: "text", text }] };
        }
        catch (error) {
            return { content: [{ type: "text", text: error.message || String(error) }], isError: true };
        }
    });
    server.tool("list_models", "List all AI models available through ds2api (DeepSeek + Qwen)", {}, async () => {
        try {
            const models = await client.listModels();
            return { content: [{ type: "text", text: JSON.stringify(models, null, 2) }] };
        }
        catch (error) {
            return { content: [{ type: "text", text: error.message || String(error) }], isError: true };
        }
    });
    server.tool("get_status", "Check ds2api service health, account pool status, and endpoint info", {}, async () => {
        try {
            const status = await client.getStatus();
            return { content: [{ type: "text", text: JSON.stringify(status, null, 2) }] };
        }
        catch (error) {
            return { content: [{ type: "text", text: error.message || String(error) }], isError: true };
        }
    });
    server.tool("get_pool_status", "Monitor ds2api account pool utilization (DeepSeek PoW/WASM + Qwen Acquire/Release)", {
        pool_type: z.enum(["deepseek", "qwen", "all"]).optional(),
    }, async (params) => {
        try {
            const result = await client.callTool("get_pool_status", params);
            return { content: [result.content[0]] };
        }
        catch (error) {
            return { content: [{ type: "text", text: error.message || String(error) }], isError: true };
        }
    });
    server.tool("embeddings", "Generate text embeddings via ds2api", {
        input: z.array(z.string()).describe("Text strings to embed"),
        model: z.string().optional().describe("Embedding model"),
    }, async (params) => {
        try {
            const result = await client.callTool("embeddings", params);
            return { content: [result.content[0]] };
        }
        catch (error) {
            return { content: [{ type: "text", text: error.message || String(error) }], isError: true };
        }
    });
    return server;
}
export async function startClaudeCodeStdio(options) {
    const server = createClaudeCodeServer(options);
    const transport = new StdioServerTransport();
    await server.connect(transport);
    console.error("[ds2api-claude] MCP Server ready for Claude Code");
    console.error("[ds2api-claude] Connected to:", options?.baseURL || process.env.DS2API_BASE_URL || "http://127.0.0.1:5001");
}
export { createClaudeCodeServer, ClaudeCodeClient };
//# sourceMappingURL=server.js.map