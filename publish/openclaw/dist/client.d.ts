import type { Ds2ApiOpenClawOptions, OpenClawToolResult } from "./types.js";
export declare class OpenClawClient {
    private baseURL;
    private apiKey;
    private timeout;
    constructor(options?: Ds2ApiOpenClawOptions);
    callTool(name: string, args?: Record<string, unknown>): Promise<OpenClawToolResult>;
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