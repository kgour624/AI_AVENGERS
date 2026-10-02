import * as vscode from 'vscode';

let requestCounter = 0;

async function rpcCall(url: string, token: string, method: string, params: any = {}): Promise<any> {
  requestCounter++;
  const body = { jsonrpc: "2.0", id: requestCounter, method, params };

  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json", "Authorization": `Bearer ${token}` },
    body: JSON.stringify(body)
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`HTTP ${res.status}: ${text}`);
  }

  const data: any = await res.json();
  if (data.error) throw new Error(`RPC ${data.error.code}: ${data.error.message}`);
  return data.result;
}

export class AvengerMCPClient {
  private _isConnected = false;
  private _url = "";
  private _token = "";

  get isConnected() { return this._isConnected; }

  async connect(mcpUrl: string, token: string): Promise<void> {
    this._url = mcpUrl.replace(/\/sse\/?$/, "");
    this._token = token;
    const result = await rpcCall(this._url, this._token, "initialize", {
      protocolVersion: "2024-11-05",
      capabilities: {},
      clientInfo: { name: "codeedgepro-vscode", version: "1.0.0" }
    });
    if (!result?.serverInfo) throw new Error("Invalid MCP server response");
    this._isConnected = true;
  }

  async disconnect(): Promise<void> {
    this._isConnected = false;
    this._url = "";
    this._token = "";
  }

  async listExperts(): Promise<any[]> {
    if (!this._isConnected) throw new Error("Not connected");
    const result = await rpcCall(this._url, this._token, "tools/call", { name: "list_experts", arguments: {} });
    if (result?.structuredContent?.experts) return result.structuredContent.experts;
    const text = result?.content?.[0]?.text || "";
    const experts: any[] = [];
    const regex = /- \*\*(.*?)\*\* — id: `(.*?)` — slug: `(.*?)`/g;
    let m;
    while ((m = regex.exec(text)) !== null) experts.push({ name: m[1], id: m[2], slug: m[3] });
    return experts;
  }

  async listTools(): Promise<any[]> {
    if (!this._isConnected) throw new Error("Not connected");
    const result = await rpcCall(this._url, this._token, "tools/list", {});
    return result?.tools || [];
  }

  async callTool(name: string, args: Record<string, any>): Promise<any> {
    if (!this._isConnected) throw new Error("Not connected");
    const result = await rpcCall(this._url, this._token, "tools/call", { name, arguments: args });
    if (result?.structuredContent) return result.structuredContent;
    const text = result?.content?.[0]?.text || "";
    try { return JSON.parse(text); } catch { return { answer: text }; }
  }

  async askExpert(expertId: string, question: string, fileContext?: string): Promise<any> {
    if (!this._isConnected) throw new Error("Not connected");
    const args: any = { expert_id: expertId, question };
    if (fileContext) args.question = `[Attached File Context]\n${fileContext}\n\n[Question]\n${question}`;
    const result = await rpcCall(this._url, this._token, "tools/call", { name: "ask_expert", arguments: args });
    if (result?.structuredContent) return result.structuredContent;
    const text = result?.content?.[0]?.text || "";
    return { answer: text, citations: [] };
  }

  async searchChunks(expertId: string, query: string, topK?: number): Promise<any> {
    if (!this._isConnected) throw new Error("Not connected");
    const result = await rpcCall(this._url, this._token, "tools/call", { name: "search_chunks", arguments: { expert_id: expertId, query, topK: topK || 10 } });
    if (result?.structuredContent) return result.structuredContent;
    return { chunks: [], text: result?.content?.[0]?.text || "" };
  }
}

export const mcpClient = new AvengerMCPClient();
