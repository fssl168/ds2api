import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import type { Ds2ApiOpenClawOptions } from "./types.js";
import { OpenClawClient } from "./client.js";

const MODELS = [
  { id: "deepseek-chat", name: "DeepSeek Chat", engine: "deepseek" },
  { id: "deepseek-reasoner", name: "DeepSeek Reasoner", engine: "deepseek" },
  { id: "deepseek-chat-search", name: "DeepSeek Chat + Search", engine: "deepseek" },
  { id: "deepseek-reasoner-search", name: "DeepSeek Reasoner + Search", engine: "deepseek" },
  { id: "qwen-plus", name: "Qwen Plus", engine: "qwen" },
  { id: "qwen-max", name: "Qwen Max", engine: "qwen" },
  { id: "qwen-coder", name: "Qwen Coder", engine: "qwen" },
  { id: "qwen-flash", name: "Qwen Flash", engine: "qwen" },
  { id: "qwen3.5-plus", name: "Qwen 3.5 Plus", engine: "qwen" },
  { id: "qwen3.5-flash", name: "Qwen 3.5 Flash", engine: "qwen" },
];

function createOpenClawServer(options?: Ds2ApiOpenClawOptions): McpServer {
  const client = new OpenClawClient(options);

  const server = new McpServer({
    name: "ds2api-mcp-openclaw",
    version: "1.0.0",
  }, {
    capabilities: { tools: {} },
  });

  server.tool(
    "chat",
    `Send chat completion to ds2api via OpenClaw.
Supports DeepSeek and Qwen models with streaming support.

Available models:
${MODELS.map((m) => `- ${m.id} (${m.name})`).join("\n")}`,
    {
      model: z.string().describe("Model: deepseek-chat, qwen-plus, qwen-max, etc."),
      messages: z.array(z.object({
        role: z.enum(["user", "assistant", "system"]),
        content: z.string(),
      })).describe("Chat messages"),
      stream: z.boolean().optional().describe("Stream response"),
      temperature: z.number().optional().describe("Temperature 0-2"),
      max_tokens: z.number().int().optional().describe("Max tokens"),
    },
    async (params) => {
      try {
        const text = await client.chat(params.messages, params.model);
        return { content: [{ type: "text" as const, text }] };
      } catch (error: any) {
        return {
          content: [{ type: "text" as const, text: error.message || String(error) }],
          isError: true,
        };
      }
    }
  );

  server.tool("list_models", "List all available AI models (DeepSeek + Qwen)", {}, async () => {
    try {
      const models = await client.listModels();
      return {
        content: [{
          type: "text" as const,
          text: JSON.stringify(models, null, 2),
        }],
      };
    } catch (error: any) {
      return {
        content: [{ type: "text" as const, text: error.message || String(error) }],
        isError: true,
      };
    }
  });

  server.tool("get_status", "Check ds2api service health and pool status", {}, async () => {
    try {
      const status = await client.getStatus();
      return {
        content: [{ type: "text" as const, text: JSON.stringify(status, null, 2) }],
      };
    } catch (error: any) {
      return {
        content: [{ type: "text" as const, text: error.message || String(error) }],
        isError: true,
      };
    }
  });

  server.tool("get_pool_status", "Monitor account pool utilization", {
    pool_type: z.enum(["deepseek", "qwen", "all"]).optional(),
  }, async (params) => {
    try {
      const result = await client.callTool("get_pool_status", params);
      return { content: [result.content[0]] };
    } catch (error: any) {
      return {
        content: [{ type: "text" as const, text: error.message || String(error) }],
        isError: true,
      };
    }
  });

  server.tool("embeddings", "Generate text embeddings", {
    input: z.array(z.string()).describe("Texts to embed"),
    model: z.string().optional().describe("Embedding model"),
  }, async (params) => {
    try {
      const result = await client.callTool("embeddings", params);
      return { content: [result.content[0]] };
    } catch (error: any) {
      return {
        content: [{ type: "text" as const, text: error.message || String(error) }],
        isError: true,
      };
    }
  });

  return server;
}

export async function startOpenClawStdio(options?: Ds2ApiOpenClawOptions): Promise<void> {
  const server = createOpenClawServer(options);
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("[ds2api-openclaw] MCP Server ready for OpenClaw");
}

export { createOpenClawServer, OpenClawClient };
export type { Ds2ApiOpenClawOptions, OpenClawConfig, OpenClawToolCall, OpenClawToolResult } from "./types.js";
