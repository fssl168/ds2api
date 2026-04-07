export interface Ds2ApiClientOptions {
    baseURL?: string;
    apiKey?: string;
    timeout?: number;
}
export interface ChatMessage {
    role: "user" | "assistant" | "system";
    content: string;
}
export interface ChatParams {
    model: string;
    messages: ChatMessage[];
    stream?: boolean;
    temperature?: number;
    max_tokens?: number;
}
export interface EmbeddingsParams {
    model?: string;
    input: string[];
}
export interface ModelInfo {
    id: string;
    name: string;
    engine: "deepseek" | "qwen";
}
export interface ServiceStatus {
    service: string;
    status: string;
    engines: string[];
    endpoints: Record<string, string>;
    version: string;
    config?: {
        keys_count: number;
        accounts_count: number;
        qwen_accounts_count: number;
        model_aliases: Record<string, string>;
    };
}
export interface PoolStatus {
    [poolType: string]: {
        type: string;
        description: string;
        features: string[];
        inflight_limit: number;
    };
}
export interface PoolStatusParams {
    pool_type?: "deepseek" | "qwen" | "all";
}
//# sourceMappingURL=types.d.ts.map