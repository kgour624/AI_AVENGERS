import * as vscode from 'vscode';
import { AuthStore } from './authStore';
import { mcpClient } from './mcpClient';
import { ChatProvider } from './chatProvider';

export async function activate(ctx: vscode.ExtensionContext) {
  const auth = new AuthStore(ctx);
  const chatProvider = new ChatProvider(ctx, auth);
  ctx.subscriptions.push(vscode.window.registerWebviewViewProvider(ChatProvider.viewType, chatProvider));

  const statusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
  statusBar.text = "$(robot) CodeEdgePro: Disconnected";
  statusBar.command = "aiAvengers.connect";
  statusBar.show();
  ctx.subscriptions.push(statusBar);

  const updateStatus = (connected: boolean) => {
    statusBar.text = connected ? "$(check) CodeEdgePro: Connected" : "$(debug-disconnect) CodeEdgePro: Disconnected";
    statusBar.backgroundColor = connected ? undefined : new vscode.ThemeColor('statusBarItem.errorBackground');
  };

  const doConnect = async () => {
    const token = await auth.getToken();
    if (!token) {
      const sel = await vscode.window.showErrorMessage("Token not set.", "Set Token");
      if (sel === "Set Token") vscode.commands.executeCommand("aiAvengers.setToken");
      return;
    }
    const mcpUrl = auth.getMcpUrl();
    try {
      await vscode.window.withProgress({ location: vscode.ProgressLocation.Notification, title: "Connecting to CodeEdgePro..." }, async () => {
        await mcpClient.connect(mcpUrl, token);
      });
      updateStatus(true);
      vscode.window.showInformationMessage("CodeEdgePro MCP Connected!");
      chatProvider.triggerReload();
    } catch (e: any) {
      updateStatus(false);
      if (e.message?.includes("401") || e.message?.includes("Unauthorized")) {
        vscode.window.showErrorMessage("401 — Token Invalid.", "Reset Token").then(s => {
          if (s === "Reset Token") vscode.commands.executeCommand("aiAvengers.setToken");
        });
      } else {
        vscode.window.showErrorMessage(`Connect failed: ${e.message}`);
      }
    }
  };

  ctx.subscriptions.push(vscode.commands.registerCommand("aiAvengers.setToken", async () => {
    const token = await vscode.window.showInputBox({ prompt: "Enter MCP Bearer Token", password: true, ignoreFocusOut: true, placeHolder: "eyJhbG..." });
    if (!token) return;
    await auth.setToken(token.trim());
    vscode.window.showInformationMessage("Token saved. Connecting...");
    await doConnect();
  }));

  ctx.subscriptions.push(vscode.commands.registerCommand("aiAvengers.connect", doConnect));
  ctx.subscriptions.push(vscode.commands.registerCommand("aiAvengers.disconnect", async () => {
    await mcpClient.disconnect();
    updateStatus(false);
    vscode.window.showInformationMessage("CodeEdgePro Disconnected");
  }));
  ctx.subscriptions.push(vscode.commands.registerCommand("aiAvengers.clearChat", async () => {
    await vscode.commands.executeCommand("workbench.action.webview.reloadWebviewAction");
  }));

  // Attach current file context
  ctx.subscriptions.push(vscode.commands.registerCommand("aiAvengers.attachFile", async () => {
    const editor = vscode.window.activeTextEditor;
    if (!editor) { vscode.window.showWarningMessage("No active file open."); return; }
    const doc = editor.document;
    const fileName = doc.fileName;
    const content = doc.getText();
    const selection = editor.selection.isEmpty ? null : doc.getText(editor.selection);
    chatProvider.setFileContext(fileName, content, selection);
    vscode.window.showInformationMessage(`Attached: ${fileName.split(/[\\/]/).pop()}`);
  }));

  const existingToken = await auth.getToken();
  if (existingToken) setTimeout(() => doConnect(), 1500);
}

export async function deactivate() { await mcpClient.disconnect(); }
