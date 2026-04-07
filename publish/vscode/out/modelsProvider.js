import * as vscode from 'vscode';
export class ModelsProvider {
    url;
    apiKey;
    _onDidChangeTreeData = new vscode.EventEmitter();
    onDidChangeTreeData = this._onDidChangeTreeData.event;
    models = [];
    constructor(url, apiKey) {
        this.url = url;
        this.apiKey = apiKey;
    }
    refresh() {
        this._onDidChangeTreeData.fire();
    }
    getTreeItem(element) {
        return element;
    }
    getChildren(element) {
        if (element) {
            return Promise.resolve([]);
        }
        return this.fetchModels();
    }
    async fetchModels() {
        const controller = new AbortController();
        let timer;
        try {
            timer = setTimeout(() => controller.abort(), 10000);
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
                    params: { name: 'list_models', arguments: {} }
                }),
                signal: controller.signal,
            });
            const data = await resp.json();
            if (data.error || data.result?.isError) {
                return [new ModelItem(`Error: ${data.error?.message || data.result?.content?.[0]?.text}`, vscode.TreeItemCollapsibleState.None)];
            }
            const text = data.result?.content?.[0]?.text || '';
            let modelList = [];
            try {
                modelList = JSON.parse(text);
            }
            catch {
                return [];
            }
            this.models = modelList.map(m => new ModelItem(`${m.name} (${m.id})`, vscode.TreeItemCollapsibleState.None, m.engine === 'deepseek' ? '$(symbol-class)' : '$(symbol-color)', { engine: m.engine, id: m.id }));
            return this.models;
        }
        catch (err) {
            return [new ModelItem(`Connection failed: ${err.message}`, vscode.TreeItemCollapsibleState.None)];
        }
        finally {
            if (timer !== undefined)
                clearTimeout(timer);
        }
    }
}
class ModelItem extends vscode.TreeItem {
    label;
    collapsibleState;
    metadata;
    constructor(label, collapsibleState, iconPath, metadata) {
        super(label, collapsibleState);
        this.label = label;
        this.collapsibleState = collapsibleState;
        this.metadata = metadata;
        if (iconPath)
            this.iconPath = iconPath;
        this.tooltip = `${label}\nEngine: ${metadata?.engine || 'unknown'}\nID: ${metadata?.id || ''}`;
    }
}
//# sourceMappingURL=modelsProvider.js.map