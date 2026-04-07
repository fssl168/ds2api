import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import type { Ds2ApiClientOptions } from "./types.js";
import { Ds2ApiClient } from "./client.js";

const MODELS: Array<{ id: string; name: string; engine: "deepseek" | "qwen" }> = [
  { id: "deepseek-chat", name: "DeepSeek Chat", engine: "deepseek" },
  { id: "deepseek-reasoner", name: "DeepSeek Reasoner (Thinking)", engine: "deepseek" },
  { id: "deepseek-chat-search", name: "DeepSeek Chat + Search", engine: "deepseek" },
  { id: "deepseek-reasoner-search", name: "DeepSeek Reasoner + Search", engine: "deepseek" },
  { id: "qwen-plus", name: "Qwen Plus", engine: "qwen" },
  { id: "qwen-max", name: "Qwen Max", engine: "qwen" },
  { id: "qwen-coder", name: "Qwen Coder", engine: "qwen" },
  { id: "qwen-flash", name: "Qwen Flash", engine: "qwen" },
  { id: "qwen3.5-plus", name: "Qwen 3.5 Plus", engine: "qwen" },
  { id: "qwen3.5-flash", name: "Qwen 3.5 Flash", engine: "qwen" },
];

function createServer(options?: Ds2ApiClientOptions): McpServer {
  const client = new Ds2ApiClient(options);

  const server = new McpServer({
    name: "ds2api-mcp",
    version: "1.0.0",
  }, {
    capabilities: {
      tools: {},
    },
  });

  server.tool(
    "chat",
    `Send a chat completion request to ds2api.
Supports DeepSeek models (deepseek-chat, deepseek-reasoner, etc.) and Qwen models (qwen-plus, qwen-max, qwen-coder, qwen-flash, qwen3.5-plus, qwen3.5-flash).
Returns the model's response text.`,
    {
      model: z.string().describe("Model name (e.g., deepseek-chat, qwen-plus, qwen-max, deepseek-reasoner, qwen-coder, qwen-flash)"),
      messages: z.array(z.object({
        role: z.enum(["user", "assistant", "system"]),
        content: z.string(),
      })).describe("Array of chat messages"),
      stream: z.boolean().optional().describe("Whether to stream the response (default: false)"),
      temperature: z.number().optional().describe("Sampling temperature (0-2, default: 1)"),
      max_tokens: z.number().int().optional().describe("Maximum tokens to generate"),
    },
    async (params) => {
      try {
        const result = await client.chat({
          model: params.model,
          messages: params.messages,
          stream: params.stream,
          temperature: params.temperature,
          max_tokens: params.max_tokens,
        });
        return {
          content: [{ type: "text" as const, text: result }],
        };
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        return {
          content: [{ type: "text" as const, text: message }],
          isError: true,
        };
      }
    }
  );

  server.tool(
    "list_models",
    "List all available models supported by ds2api, including DeepSeek and Qwen models.",
    {},
    async () => {
      try {
        const models = await client.listModels();
        return {
          content: [{ type: "text" as const, text: JSON.stringify(models, null, 2) }],
        };
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        return {
          content: [{ type: "text" as const, text: message }],
          isError: true,
        };
      }
    }
  );

  server.tool(
    "get_status",
    "Get ds2api service status including health check and account pool information.",
    {},
    async () => {
      try {
        const status = await client.getStatus();
        return {
          content: [{ type: "text" as const, text: JSON.stringify(status, null, 2) }],
        };
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        return {
          content: [{ type: "text" as const, text: message }],
          isError: true,
        };
      }
    }
  );

  server.tool(
    "get_pool_status",
    "Get detailed account pool status for both DeepSeek and Qwen pools.",
    {
      pool_type: z.enum(["deepseek", "qwen", "all"]).optional().describe("Which pool to query (default: all)"),
    },
    async (params) => {
      try {
        const status = await client.getPoolStatus({ pool_type: params.pool_type });
        return {
          content: [{ type: "text" as const, text: JSON.stringify(status, null, 2) }],
        };
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        return {
          content: [{ type: "text" as const, text: message }],
          isError: true,
        };
      }
    }
  );

  server.tool(
    "embeddings",
    "Generate embeddings for text using ds2api.",
    {
      model: z.string().optional().describe("Embedding model (default: text-embedding-ada002)"),
      input: z.array(z.string()).describe("Array of strings to embed"),
    },
    async (params) => {
      try {
        const result = await client.embeddings({ model: params.model, input: params.input });
        return {
          content: [{ type: "text" as const, text: typeof result === "string" ? result : JSON.stringify(result, null, 2) }],
        };
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        return {
          content: [{ type: "text" as const, text: message }],
          isError: true,
        };
      }
    }
  );

  return server;
}

export async function startStdioServer(options?: Ds2ApiClientOptions): Promise<void> {
  const server = createServer(options);
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("[ds2api-mcp] MCP Server running in stdio mode");
  console.error("[ds2api-mcp] Connected to ds2api at:", options?.baseURL || process.env.DS2API_BASE_URL || "http://127.0.0.1:5001");
}

export { createServer, Ds2ApiClient };
export type { Ds2ApiClientOptions, ChatParams, ChatMessage, ModelInfo, ServiceStatus, PoolStatus } from "./types.js";
