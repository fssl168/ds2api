import type { Ds2ApiOpenClawOptions, OpenClawToolResult } from "./types.js";

const DEFAULT_BASE_URL = "http://127.0.0.1:5001";
const DEFAULT_TIMEOUT = 120_000;

export class OpenClawClient {
  private baseURL: string;
  private apiKey: string;
  private timeout: number;

  constructor(options: Ds2ApiOpenClawOptions = {}) {
    this.baseURL = (options.baseURL || process.env.DS2API_BASE_URL || DEFAULT_BASE_URL).replace(/\/+$/, "");
    this.apiKey = options.apiKey || process.env.DS2API_API_KEY || "";
    this.timeout = options.timeout || DEFAULT_TIMEOUT;
  }

  async callTool(name: string, args: Record<string, unknown> = {}): Promise<OpenClawToolResult> {
    const body = {
      jsonrpc: "2.0",
      id: Date.now(),
      method: "tools/call",
      params: { name, arguments: args },
    };

    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), this.timeout);

    try {
      const resp = await fetch(`${this.baseURL}/mcp`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(this.apiKey ? { Authorization: `Bearer ${this.apiKey}` } : {}),
        },
        body: JSON.stringify(body),
        signal: controller.signal,
      });

      const data = await resp.json() as any;

      if (!resp.ok) {
        return {
          content: [{ type: "text", text: `HTTP ${resp.status}: ${JSON.stringify(data)}` }],
          isError: true,
        };
      }

      if (data.error) {
        return {
          content: [{ type: "text", text: `MCP Error (${data.error.code}): ${data.error.message}` }],
          isError: true,
        };
      }

      if (!data.result) {
        return { content: [{ type: "text", text: "Empty response" }], isError: true };
      }

      return data.result as OpenClawToolResult;
    } finally {
      clearTimeout(timer);
    }
  }

  async chat(messages: Array<{ role: string; content: string }>, model?: string): Promise<string> {
    const result = await this.callTool("chat", {
      model: model || "deepseek-chat",
      messages,
    });
    if (result.isError) throw new Error(result.content[0]?.text || "Chat failed");
    return result.content[0]?.text || "";
  }

  async listModels(): Promise<Array<{ id: string; name: string; engine: string }>> {
    const result = await this.callTool("list_models");
    if (result.isError) throw new Error(result.content[0]?.text || "List models failed");
    try {
      return JSON.parse(result.content[0]?.text || "[]");
    } catch {
      return [];
    }
  }

  async getStatus(): Promise<Record<string, unknown>> {
    const result = await this.callTool("get_status");
    if (result.isError) throw new Error(result.content[0]?.text || "Get status failed");
    try {
      return JSON.parse(result.content[0]?.text || "{}");
    } catch {
      return {};
    }
  }

  healthCheck(): Promise<boolean> {
    return this.callTool("ping", {}).then(
      () => true,
      () => false
    );
  }
}
