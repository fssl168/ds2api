import type { ClaudeCodeOptions } from "./types.js";
export declare class ClaudeCodeClient {
    private baseURL;
    private apiKey;
    private timeout;
    constructor(options?: ClaudeCodeOptions);
    callTool(name: string, args?: Record<string, unknown>): Promise<any>;
    chat(messages: Array<{
        role: string;
        content: string;
    }>, model?: string): Promise<string>;
    listModels(): Promise<Array<{
        id: string;
        name: string;
        engine: string;
    }>>;
    getStatus(): Promise<Record<string, unknown>>;
    healthCheck(): Promise<boolean>;
}
//# sourceMappingURL=client.d.ts.map