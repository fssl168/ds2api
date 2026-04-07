const DEFAULT_BASE_URL = "http://127.0.0.1:5001";
const DEFAULT_TIMEOUT = 120_000;
export class OpenClawClient {
    baseURL;
    apiKey;
    timeout;
    constructor(options = {}) {
        this.baseURL = (options.baseURL || process.env.DS2API_BASE_URL || DEFAULT_BASE_URL).replace(/\/+$/, "");
        this.apiKey = options.apiKey || process.env.DS2API_API_KEY || "";
        this.timeout = options.timeout || DEFAULT_TIMEOUT;
    }
    async callTool(name, args = {}) {
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
            const data = await resp.json();
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
            return data.result;
        }
        finally {
            clearTimeout(timer);
        }
    }
    async chat(messages, model) {
        const result = await this.callTool("chat", {
            model: model || "deepseek-chat",
            messages,
        });
        if (result.isError)
            throw new Error(result.content[0]?.text || "Chat failed");
        return result.content[0]?.text || "";
    }
    async listModels() {
        const result = await this.callTool("list_models");
        if (result.isError)
            throw new Error(result.content[0]?.text || "List models failed");
        try {
            return JSON.parse(result.content[0]?.text || "[]");
        }
        catch {
            return [];
        }
    }
    async getStatus() {
        const result = await this.callTool("get_status");
        if (result.isError)
            throw new Error(result.content[0]?.text || "Get status failed");
        try {
            return JSON.parse(result.content[0]?.text || "{}");
        }
        catch {
            return {};
        }
    }
    healthCheck() {
        return this.callTool("ping", {}).then(() => true, () => false);
    }
}
//# sourceMappingURL=client.js.map