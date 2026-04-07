export interface OpenClawConfig {
  mcpServers: {
    ds2api: {
      type?: "streamable_http" | "stdio";
      url?: string;
      command?: string;
      args?: string[];
      env?: Record<string, string>;
      headers?: Record<string, string>;
    };
  };
  ds2api?: {
    baseURL?: string;
    apiKey?: string;
    defaultModel?: string;
    timeout?: number;
    retryCount?: number;
  };
}

export interface OpenClawToolCall {
  name: string;
  arguments: Record<string, unknown>;
}

export interface OpenClawToolResult {
  content: Array<{ type: "text"; text: string }>;
  isError?: boolean;
}

export interface Ds2ApiOpenClawOptions {
  baseURL?: string;
  apiKey?: string;
  defaultModel?: string;
  timeout?: number;
}
