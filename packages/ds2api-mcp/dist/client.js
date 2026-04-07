const DEFAULT_BASE_URL = "http://127.0.0.1:5001";
const DEFAULT_TIMEOUT = 120_000;
export class Ds2ApiClient {
    baseURL;
    apiKey;
    timeout;
    constructor(options = {}) {
        this.baseURL = (options.baseURL || process.env.DS2API_BASE_URL || DEFAULT_BASE_URL).replace(/\/+$/, "");
        this.apiKey = options.apiKey || process.env.DS2API_API_KEY || "";
        this.timeout = options.timeout || DEFAULT_TIMEOUT;
    }
    async request(path, body) {
        const url = `${this.baseURL}${path}`;
        const headers = {
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
            return data;
        }
        finally {
            clearTimeout(timer);
        }
    }
    async mcpCall(toolName, args = {}) {
        const body = {
            jsonrpc: "2.0",
            id: Date.now(),
            method: "tools/call",
            params: {
                name: toolName,
                arguments: args,
            },
        };
        const resp = await this.request("/mcp", body);
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
            return JSON.parse(text);
        }
        catch {
            return text;
        }
    }
    async chat(params) {
        return this.mcpCall("chat", params);
    }
    async listModels() {
        return this.mcpCall("list_models");
    }
    async getStatus() {
        return this.mcpCall("get_status");
    }
    async getPoolStatus(params) {
        return this.mcpCall("get_pool_status", params);
    }
    async embeddings(params) {
        return this.mcpCall("embeddings", params);
    }
}
//# sourceMappingURL=client.js.map