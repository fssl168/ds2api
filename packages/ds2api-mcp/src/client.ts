import type { Ds2ApiClientOptions, ChatParams, EmbeddingsParams, ModelInfo, ServiceStatus, PoolStatus, PoolStatusParams } from "./types.js";

const DEFAULT_BASE_URL = "http://127.0.0.1:5001";
const DEFAULT_TIMEOUT = 120_000;

export class Ds2ApiClient {
  private baseURL: string;
  private apiKey: string;
  private timeout: number;

  constructor(options: Ds2ApiClientOptions = {}) {
    this.baseURL = (options.baseURL || process.env.DS2API_BASE_URL || DEFAULT_BASE_URL).replace(/\/+$/, "");
    this.apiKey = options.apiKey || process.env.DS2API_API_KEY || "";
    this.timeout = options.timeout || DEFAULT_TIMEOUT;
  }

  private async request<T>(path: string, body?: unknown): Promise<T> {
    const url = `${this.baseURL}${path}`;
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    if (this.apiKey) {
      headers["Authorization"] = `Bearer ${this.apiKey}`;
    }

    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), this.timeout);

    try {
      const resp = await fetch(url, {
        method: body ? "POST" : "GET",
        headers,
        body: body ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      });

      const data = await resp.json();

      if (!resp.ok) {
        throw new Error(`API error (${resp.status}): ${JSON.stringify(data)}`);
      }

      return data as T;
    } finally {
      clearTimeout(timer);
    }
  }

  private async mcpCall<T>(toolName: string, args: Record<string, unknown> = {}): Promise<T> {
    const body = {
      jsonrpc: "2.0",
      id: Date.now(),
      method: "tools/call",
      params: {
        name: toolName,
        arguments: args,
      },
    };

    const resp = await this.request<{
      jsonrpc: string;
      result?: { content: Array<{ type: string; text: string }>; isError?: boolean };
      error?: { code: number; message: string };
    }>("/mcp", body);

    if (resp.error) {
      throw new Error(`MCP Error (${resp.error.code}): ${resp.error.message}`);
    }

    if (!resp.result) {
      throw new Error("MCP response missing result");
    }

    if (resp.result.isError) {
      const text = resp.result.content?.[0]?.text || "Unknown error";
      throw new Error(text);
    }

    const text = resp.result.content?.[0]?.text || "";
    try {
      return JSON.parse(text) as T;
    } catch {
      return text as unknown as T;
    }
  }

  async chat(params: ChatParams): Promise<string> {
    return this.mcpCall<string>("chat", params as unknown as Record<string, unknown>);
  }

  async listModels(): Promise<ModelInfo[]> {
    return this.mcpCall<ModelInfo[]>("list_models");
  }

  async getStatus(): Promise<ServiceStatus> {
    return this.mcpCall<ServiceStatus>("get_status");
  }

  async getPoolStatus(params?: PoolStatusParams): Promise<PoolStatus> {
    return this.mcpCall<PoolStatus>("get_pool_status", params as unknown as Record<string, unknown>);
  }

  async embeddings(params: EmbeddingsParams): Promise<unknown> {
    return this.mcpCall<unknown>("embeddings", params as unknown as Record<string, unknown>);
  }
}
