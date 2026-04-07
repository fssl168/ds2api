import type { Ds2ApiClientOptions, ChatParams, EmbeddingsParams, ModelInfo, ServiceStatus, PoolStatus, PoolStatusParams } from "./types.js";
export declare class Ds2ApiClient {
    private baseURL;
    private apiKey;
    private timeout;
    constructor(options?: Ds2ApiClientOptions);
    private request;
    private mcpCall;
    chat(params: ChatParams): Promise<string>;
    listModels(): Promise<ModelInfo[]>;
    getStatus(): Promise<ServiceStatus>;
    getPoolStatus(params?: PoolStatusParams): Promise<PoolStatus>;
    embeddings(params: EmbeddingsParams): Promise<unknown>;
}
//# sourceMappingURL=client.d.ts.map