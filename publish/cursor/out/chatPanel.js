export class ChatPanelProvider {
    url;
    apiKey;
    defaultModel;
    _view;
    _messages = [];
    constructor(url, apiKey, defaultModel) {
        this.url = url;
        this.apiKey = apiKey;
        this.defaultModel = defaultModel;
    }
    resolveWebviewView(webviewView) {
        this._view = webviewView;
        webviewView.webview.options = { enableScripts: true };
        webviewView.webview.html = this.getHtmlContent();
        webviewView.webview.onDidReceiveMessage(async (msg) => {
            if (msg.type === 'sendChat') {
                await this.handleChat(msg.model, msg.message);
            }
        });
    }
    sendChat(model, message) {
        this._view?.webview.postMessage({ type: 'setInput', model, message });
        this.handleChat(model, message);
    }
    async handleChat(model, userMessage) {
        if (!this._view)
            return;
        this._messages.push({ role: 'user', content: userMessage });
        this._view.webview.postMessage({ type: 'addMessage', role: 'user', content: userMessage });
        this._view.webview.postMessage({ type: 'setLoading', loading: true });
        try {
            const controller = new AbortController();
            let timer;
            try {
                timer = setTimeout(() => controller.abort(), 120000);
                const resp = await fetch(this.url, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        ...(this.apiKey ? { Authorization: `Bearer ${this.apiKey}` } : {})
                    },
                    body: JSON.stringify({
                        jsonrpc: '2.0',
                        id: Date.now(),
                        method: 'tools/call',
                        params: {
                            name: 'chat',
                            arguments: {
                                model: model || this.defaultModel,
                                messages: this._messages
                            }
                        }
                    }),
                    signal: controller.signal,
                });
                const data = await resp.json();
                if (data.error || data.result?.isError) {
                    const errText = data.error?.message || data.result?.content?.[0]?.text || 'Unknown error';
                    this._view.webview.postMessage({ type: 'addMessage', role: 'assistant', content: `[Error] ${errText}`, isError: true });
                }
                else {
                    const reply = data.result?.content?.[0]?.text || '(empty response)';
                    this._messages.push({ role: 'assistant', content: reply });
                    this._view.webview.postMessage({ type: 'addMessage', role: 'assistant', content: reply });
                }
            }
            finally {
                if (timer !== undefined)
                    clearTimeout(timer);
            }
        }
        catch (err) {
            this._view.webview.postMessage({ type: 'addMessage', role: 'assistant', content: `[Connection Error] ${err.message}`, isError: true });
        }
        finally {
            this._view.webview.postMessage({ type: 'setLoading', loading: false });
        }
    }
    getHtmlContent() {
        return /*html*/ `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: var(--vscode-font-family); margin: 0; padding: 8px; color: var(--vscode-foreground); background: var(--vscode-editor-background); }
        #container { display: flex; flex-direction: column; height: 100%; }
        #messages { flex: 1; overflow-y: auto; padding: 4px 0; }
        .msg { padding: 6px 10px; margin: 4px 0; border-radius: 6px; max-width: 100%; word-wrap: break-word; font-size: 13px; }
        .user-msg { background: var(--vscode-button-background); margin-left: 16px; border-left: 3px solid var(--vscode-button-foreground); }
        .assistant-msg { background: var(--vscode-editor-inactiveSelectionBackground); margin-right: 16px; border-left: 3px solid var(--vscode-textLink-foreground); }
        .error-msg { border-left-color: var(--vscode-errorForeground); }
        #input-area { display: flex; gap: 6px; padding-top: 8px; border-top: 1px solid var(--vscode-panel-border); }
        select { padding: 4px 8px; border-radius: 4px; background: var(--vscode-input-background); color: var(--vscode-input-foreground); border: 1px solid var(--vscode-input-border); font-size: 12px; width: 140px; }
        input { flex: 1; padding: 4px 8px; border-radius: 4px; background: var(--vscode-input-background); color: var(--vscode-input-foreground); border: 1px solid var(--vscode-input-border); font-size: 13px; }
        button { padding: 4px 14px; border-radius: 4px; background: var(--vscode-button-background); color: var(--vscode-button-foreground); border: none; cursor: pointer; font-size: 13px; }
        button:hover { background: var(--vscode-button-hoverBackground); }
        button:disabled { opacity: 0.5; cursor: not-allowed; }
        #loading { font-size: 12px; color: var(--vscode-descriptionForeground); padding: 4px 10px; display: none; }
        .model-tag { font-size: 10px; color: var(--vscode-descriptionForeground); margin-bottom: 2px; }
    </style>
</head>
<body>
<div id="container">
    <div id="messages"></div>
    <div id="loading">⏳ Thinking...</div>
    <div id="input-area">
        <select id="modelSelect">
            <option value="deepseek-chat">deepseek-chat</option>
            <option value="deepseek-reasoner">deepseek-reasoner</option>
            <option value="qwen-plus">qwen-plus</option>
            <option value="qwen-max">qwen-max</option>
            <option value="qwen-coder">qwen-coder</option>
            <option value="qwen-flash">qwen-flash</option>
            <option value="qwen3.5-plus">qwen3.5-plus</option>
            <option value="qwen3.5-flash">qwen3.5-flash</option>
        </select>
        <input type="text" id="msgInput" placeholder="Ask ds2api..." />
        <button id="sendBtn" onclick="send()">Send</button>
    </div>
</div>
<script>
const vscode = acquireVsCodeApi();
function send() {
    const input = document.getElementById('msgInput');
    const model = document.getElementById('modelSelect').value;
    const msg = input.value.trim();
    if (!msg) return;
    input.value = '';
    vscode.postMessage({ type: 'sendChat', model, message: msg });
}
document.getElementById('msgInput').addEventListener('keydown', e => { if (e.key === 'Enter') send(); });
window.addEventListener('message', event => {
    const d = event.data;
    if (d.type === 'setInput') {
        document.getElementById('modelSelect').value = d.model;
        document.getElementById('msgInput').value = d.message;
        send();
    } else if (d.type === 'addMessage') {
        const div = document.createElement('div');
        div.className = 'msg ' + (d.role === 'user' ? 'user-msg' : ('assistant-msg' + (d.isError ? ' error-msg' : '')));
        div.innerHTML = '<div class="model-tag">' + (d.role === 'user' ? 'You' : 'ds2api') + '</div>' + d.content.replace(/\n/g, '<br>');
        document.getElementById('messages').appendChild(div);
        document.getElementById('messages').scrollTop = document.getElementById('messages').scrollHeight;
    } else if (d.type === 'setLoading') {
        document.getElementById('loading').style.display = d.loading ? 'block' : 'none';
        document.getElementById('sendBtn').disabled = d.loading;
    }
});
</script>
</body>
</html>`;
    }
}
//# sourceMappingURL=chatPanel.js.map