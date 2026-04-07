import * as vscode from 'vscode';
export declare class ModelsProvider implements vscode.TreeDataProvider<ModelItem> {
    private url;
    private apiKey;
    private _onDidChangeTreeData;
    readonly onDidChangeTreeData: vscode.Event<void>;
    private models;
    constructor(url: string, apiKey: string);
    refresh(): void;
    getTreeItem(element: ModelItem): vscode.TreeItem;
    getChildren(element?: ModelItem): Thenable<ModelItem[]>;
    private fetchModels;
}
declare class ModelItem extends vscode.TreeItem {
    readonly label: string;
    readonly collapsibleState: vscode.TreeItemCollapsibleState;
    readonly metadata?: {
        engine: string;
        id: string;
    } | undefined;
    constructor(label: string, collapsibleState: vscode.TreeItemCollapsibleState, iconPath?: string | vscode.ThemeIcon, metadata?: {
        engine: string;
        id: string;
    } | undefined);
}
export {};
