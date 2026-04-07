import * as vscode from 'vscode';
import { ModelsProvider } from './modelsProvider';

let statusBarItem: vscode.StatusBarItem;

export function activate(context: vscode.ExtensionContext) {
    const config = vscode.workspace.getConfiguration('ds2api-mcp');
    const url = config.get<string>('url', 'http://127.0.0.1:5001/mcp');
    const apiKey = config.get<string>('apiKey', '');

    const modelsProvider = new ModelsProvider(url, apiKey);
    context.subscriptions.push(
        vscode.window.registerTreeDataProvider('ds2api-mcp.models', modelsProvider),
        vscode.commands.registerCommand('ds2api-mcp.showStatus', async () => {
            const ok = await checkConnection(url);
            vscode.window.showInformationMessage(
                ok ? `✅ ds2api connected at ${url}` : `❌ Cannot reach ds2api at ${url}`
            );
            modelsProvider.refresh();
        }),
        vscode.commands.registerCommand('ds2api-mcp.listModels', async () => {
            modelsProvider.refresh();
        }),
        vscode.commands.registerCommand('ds2api-mcp.openSettings', () => {
            vscode.commands.executeCommand('workbench.action.openSettings', 'ds2api-mcp');
        })
    );

    statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    statusBarItem.text = '$(plug) ds2api';
    statusBarItem.tooltip = 'ds2api MCP Bridge';
    statusBarItem.command = 'ds2api-mcp.showStatus';
    statusBarItem.show();

    context.subscriptions.push(statusBarItem);

    setInterval(async () => {
        const ok = await checkConnection(url);
        statusBarItem.text = ok ? '$(check) ds2api' : '$(error) ds2api';
    }, 30000);
}

async function checkConnection(url: string): Promise<boolean> {
    try {
        const resp = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ jsonrpc: '2.0', id: 1, method: 'ping' }),
            signal: AbortSignal.timeout(5000),
        });
        return resp.ok;
    } catch {
        return false;
    }
}

export function deactivate() {
    statusBarItem?.dispose();
}
