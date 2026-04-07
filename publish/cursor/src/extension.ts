import * as vscode from 'vscode';
import { ModelsProvider } from './modelsProvider';
import { ChatPanelProvider } from './chatPanel';

let statusBarItem: vscode.StatusBarItem;
let healthCheckInterval: ReturnType<typeof setInterval>;

export function activate(context: vscode.ExtensionContext) {
    const config = vscode.workspace.getConfiguration('ds2api-mcp-cursor');
    const url = config.get<string>('url', 'http://127.0.0.1:5001/mcp');
    const apiKey = config.get<string>('apiKey', '');
    const defaultModel = config.get<string>('defaultModel', 'deepseek-chat');

    const modelsProvider = new ModelsProvider(url, apiKey);
    const chatPanelProvider = new ChatPanelProvider(url, apiKey, defaultModel);

    context.subscriptions.push(
        vscode.window.registerTreeDataProvider('ds2api-mcp-cursor.models', modelsProvider),
        vscode.window.registerWebviewViewProvider('ds2api-mcp-cursor.chat', chatPanelProvider),
        vscode.commands.registerCommand('ds2api-mcp-cursor.showStatus', async () => {
            const ok = await checkConnection(url);
            vscode.window.showInformationMessage(
                ok ? `✅ ds2api connected at ${url}` : `❌ Cannot reach ds2api at ${url}`
            );
            modelsProvider.refresh();
        }),
        vscode.commands.registerCommand('ds2api-mcp-cursor.listModels', () => {
            modelsProvider.refresh();
            vscode.commands.executeCommand('ds2api-mcp-cursor.sidebar.focus');
        }),
        vscode.commands.registerCommand('ds2api-mcp-cursor.chatWithDeepSeek', async () => {
            const editor = vscode.window.activeTextEditor;
            const selection = editor?.document.getText(editor.selection);
            if (!selection) {
                vscode.window.showWarningMessage('Please select some text first.');
                return;
            }
            chatPanelProvider.sendChat('deepseek-chat', selection);
            vscode.commands.executeCommand('ds2api-mcp-cursor.sidebar.focus');
        }),
        vscode.commands.registerCommand('ds2api-mcp-cursor.chatWithQwen', async () => {
            const editor = vscode.window.activeTextEditor;
            const selection = editor?.document.getText(editor.selection);
            if (!selection) {
                vscode.window.showWarningMessage('Please select some text first.');
                return;
            }
            chatPanelProvider.sendChat('qwen-plus', selection);
            vscode.commands.executeCommand('ds2api-mcp-cursor.sidebar.focus');
        }),
        vscode.commands.registerCommand('ds2api-mcp-cursor.injectCursorRules', async () => {
            await injectCursorRules();
            vscode.window.showInformationMessage('✅ .cursor/rules injected for ds2api');
        }),
        vscode.commands.registerCommand('ds2api-mcp-cursor.openSettings', () => {
            vscode.commands.executeCommand('workbench.action.openSettings', 'ds2api-mcp-cursor');
        })
    );

    statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    statusBarItem.text = '$(plug) ds2api';
    statusBarItem.tooltip = 'ds2api MCP Bridge for Cursor';
    statusBarItem.command = 'ds2api-mcp-cursor.showStatus';
    statusBarItem.show();
    context.subscriptions.push(statusBarItem);

    healthCheckInterval = setInterval(async () => {
        const ok = await checkConnection(url);
        statusBarItem.text = ok ? '$(check) dsapi' : '$(error) ds2api';
    }, 30000);

    if (config.get<boolean>('autoInjectRules', true)) {
        injectCursorRules().catch(() => {});
    }
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

async function injectCursorRules(): Promise<void> {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) return;

    const rulesDir = vscode.Uri.joinPath(workspaceFolder.uri, '.cursor', 'rules');
    try {
        await vscode.workspace.fs.createDirectory(rulesDir);
    } catch { }

    const ruleFile = vscode.Uri.joinPath(rulesDir, 'ds2api.mdc');
    const ruleContent = `# ds2api MCP Rules

You have access to ds2api MCP tools that provide DeepSeek and Qwen AI models.

## Available Tools
- **chat**: Send messages to DeepSeek or Qwen models
- **list_models**: List all 10 available models
- **get_status**: Check service health
- **get_pool_status**: Monitor account pools
- **embeddings**: Generate text embeddings

## Model Selection Guide
| Use Case | Recommended Model |
|----------|-------------------|
| General coding | deepseek-chat |
| Complex reasoning | deepseek-reasoner |
| Chinese content | qwen-max or qwen3.5-plus |
| Code generation | qwen-coder |
| Fast responses | qwen-flash |

## Usage
When user asks about AI models, code explanation, or needs LLM assistance:
1. Use list_models to show available options
2. Use chat with appropriate model for the task
3. For Chinese language tasks, prefer Qwen models
4. For reasoning-heavy tasks, use deepseek-reasoner

## Important
- Always confirm model availability before suggesting one
- Check get_status if experiencing issues
- ds2api runs locally at http://127.0.0.1:5001
`;

    await vscode.workspace.fs.writeFile(ruleFile, Buffer.from(ruleContent, 'utf-8'));
}

export function deactivate() {
    statusBarItem?.dispose();
    if (healthCheckInterval) clearInterval(healthCheckInterval as any);
}
