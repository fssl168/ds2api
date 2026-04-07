import * as vscode from 'vscode';
export declare class ChatPanelProvider implements vscode.WebviewViewProvider {
    private url;
    private apiKey;
    private defaultModel;
    private _view?;
    private _messages;
    constructor(url: string, apiKey: string, defaultModel: string);
    resolveWebviewView(webviewView: vscode.WebviewView): void;
    sendChat(model: string, message: string): void;
    private handleChat;
    private getHtmlContent;
}
