import * as vscode from 'vscode';
const TOKEN_KEY = 'aiAvengers.token';
export class AuthStore {
  constructor(private ctx: vscode.ExtensionContext) {}
  async getToken(): Promise<string | undefined> {
    return await this.ctx.secrets.get(TOKEN_KEY);
  }
  async setToken(token: string): Promise<void> {
    await this.ctx.secrets.store(TOKEN_KEY, token);
  }
  async clearToken(): Promise<void> {
    await this.ctx.secrets.delete(TOKEN_KEY);
  }
  getMcpUrl(): string {
    return vscode.workspace.getConfiguration('aiAvengers').get<string>('mcpUrl') || 'https://your-domain/mcp/v2/sse';
  }
}
