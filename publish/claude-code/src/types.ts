export interface ClaudeCodeConfig {
  mcpServers: {
    ds2api: {
      command?: string;
      args?: string[];
      env?: Record<string, string>;
      type?: "streamable_http" | "stdio";
      url?: string;
    };
  };
  ds2api?: {
    baseURL?: string;
    apiKey?: string;
    defaultModel?: string;
    maxTokens?: number;
    thinking?: boolean;
    preferredEngine?: "deepseek" | "qwen" | "auto";
  };
}

export interface ClaudeCodeToolCall {
  name: string;
  arguments: Record<string, unknown>;
}

export interface ClaudeCodeOptions {
  baseURL?: string;
  apiKey?: string;
  defaultModel?: string;
  maxTokens?: number;
  thinking?: boolean;
  timeout?: number;
}
