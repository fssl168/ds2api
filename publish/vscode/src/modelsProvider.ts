import * as vscode from 'vscode';

interface ModelInfo {
    id: string;
    name: string;
    engine: string;
}

export class ModelsProvider implements vscode.TreeDataProvider<ModelItem> {
    private _onDidChangeTreeData = new vscode.EventEmitter<void>();
    readonly onDidChangeTreeData = this._onDidChangeTreeData.event;
    private models: ModelItem[] = [];

    constructor(private url: string, private apiKey: string) {}

    refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: ModelItem): vscode.TreeItem {
        return element;
    }

    getChildren(element?: ModelItem): Thenable<ModelItem[]> {
        if (element) {
            return Promise.resolve([]);
        }
        return this.fetchModels();
    }

    private async fetchModels(): Promise<ModelItem[]> {
        const controller = new AbortController();
        let timer: ReturnType<typeof setTimeout> | undefined;
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
            const data = await resp.json() as any;
            if (data.error || data.result?.isError) {
                return [new ModelItem(`Error: ${data.error?.message || data.result?.content?.[0]?.text}`, vscode.TreeItemCollapsibleState.None)];
            }
            const text = data.result?.content?.[0]?.text || '';
            let modelList: ModelInfo[] = [];
            try { modelList = JSON.parse(text); } catch { return []; }
            
            this.models = modelList.map(m => new ModelItem(
                `${m.name} (${m.id})`,
                vscode.TreeItemCollapsibleState.None,
                m.engine === 'deepseek' ? '$(symbol-class)' : '$(symbol-color)',
                { engine: m.engine, id: m.id }
            ));
            return this.models;
        } catch (err: any) {
            return [new ModelItem(`Connection failed: ${err.message}`, vscode.TreeItemCollapsibleState.None)];
        } finally {
            if (timer !== undefined) clearTimeout(timer);
        }
    }
}

class ModelItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState,
        iconPath?: string | vscode.ThemeIcon,
        public readonly metadata?: { engine: string; id: string }
    ) {
        super(label, collapsibleState);
        if (iconPath) this.iconPath = iconPath;
        this.tooltip = `${label}\nEngine: ${metadata?.engine || 'unknown'}\nID: ${metadata?.id || ''}`;
    }
}
