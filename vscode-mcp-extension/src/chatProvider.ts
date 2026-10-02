import * as vscode from 'vscode';
import { AuthStore } from './authStore';
import { mcpClient } from './mcpClient';

interface FileContext {
  name: string;
  content: string;
  selection: string | null;
}

export class ChatProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'aiAvengers.chatView';
  private _view?: vscode.WebviewView;
  private _files: FileContext[] = [];

  constructor(private ctx: vscode.ExtensionContext, private auth: AuthStore) {}

  public triggerReload() {
    this._view?.webview.postMessage({ type: 'reload' });
  }

  public setFileContext(name: string, content: string, selection: string | null) {
    // Avoid duplicates based on name
    if (!this._files.find(f => f.name === name)) {
      this._files.push({ name, content, selection });
    } else {
      // Update existing
      const idx = this._files.findIndex(f => f.name === name);
      this._files[idx] = { name, content, selection };
    }
    
    this._view?.webview.postMessage({ 
      type: 'filesUpdated', 
      files: this._files.map(f => ({ name: f.name.split(/[\\/]/).pop() || f.name, hasSelection: !!f.selection })) 
    });
  }

  resolveWebviewView(view: vscode.WebviewView) {
    this._view = view;
    view.webview.options = { enableScripts: true, localResourceRoots: [this.ctx.extensionUri] };
    view.webview.html = this.getHtml();

    view.webview.onDidReceiveMessage(async (msg) => {
      if (msg.type === 'loadExperts') {
        if (!mcpClient.isConnected) {
          for (const d of [2000, 4000, 6000]) { await new Promise(r => setTimeout(r, d)); if (mcpClient.isConnected) break; }
        }
        try {
          if (!mcpClient.isConnected) { view.webview.postMessage({ type: 'status', text: '⚡ Not connected — Ctrl+Shift+P → CodeEdgePro: Connect' }); return; }
          const experts = await mcpClient.listExperts();
          view.webview.postMessage({ type: 'experts', experts });
        } catch (e: any) { view.webview.postMessage({ type: 'error', text: e.message }); }
      }

      if (msg.type === 'loadTools') {
        try {
          if (!mcpClient.isConnected) return;
          const tools = await mcpClient.listTools();
          view.webview.postMessage({ type: 'tools', tools });
        } catch (e: any) { view.webview.postMessage({ type: 'error', text: 'Tools: ' + e.message }); }
      }

      if (msg.type === 'ask') {
        try {
          view.webview.postMessage({ type: 'thinking', value: true });
          
          let fileCtxStr = '';
          if (this._files.length > 0) {
            fileCtxStr = this._files.map(f => `File: ${f.name}\n\`\`\`\n${f.selection || f.content.substring(0, 8000)}\n\`\`\``).join('\n\n');
          }
          
          const result = await mcpClient.askExpert(msg.expertId, msg.question, fileCtxStr || undefined);
          const answer = typeof result === 'string' ? result : (result.answer || JSON.stringify(result, null, 2));
          const citations = typeof result === 'object' ? (result.citations || []) : [];
          view.webview.postMessage({ type: 'answer', text: answer, citations });
        } catch (e: any) {
          view.webview.postMessage({ type: 'error', text: e.message });
        } finally { view.webview.postMessage({ type: 'thinking', value: false }); }
      }

      if (msg.type === 'callTool') {
        try {
          view.webview.postMessage({ type: 'thinking', value: true });
          const result = await mcpClient.callTool(msg.toolName, msg.args || {});
          const text = typeof result === 'string' ? result : (result.answer || result.text || JSON.stringify(result, null, 2));
          view.webview.postMessage({ type: 'answer', text: '🔧 ' + msg.toolName + '\\n\\n' + text, citations: [] });
        } catch (e: any) {
          view.webview.postMessage({ type: 'error', text: e.message });
        } finally { view.webview.postMessage({ type: 'thinking', value: false }); }
      }

      if (msg.type === 'attachFile') {
        vscode.commands.executeCommand('aiAvengers.attachFile');
      }
      
      if (msg.type === 'removeFile') {
        this._files.splice(msg.index, 1);
        this._view?.webview.postMessage({ 
          type: 'filesUpdated', 
          files: this._files.map(f => ({ name: f.name.split(/[\\/]/).pop() || f.name, hasSelection: !!f.selection })) 
        });
      }
    });
  }

  private getHtml(): string {
    return `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<style>
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap');

  :root {
    --bg-primary: #0b1120;
    --bg-secondary: #111c2e;
    --bg-card: rgba(15, 25, 45, 0.85);
    --bg-glass: rgba(20, 35, 60, 0.6);
    --border: rgba(6, 182, 212, 0.2);
    --border-glow: rgba(6, 182, 212, 0.4);
    --accent: #06b6d4;
    --accent-bright: #22d3ee;
    --accent-purple: #8b5cf6;
    --accent-gradient: linear-gradient(135deg, #06b6d4, #8b5cf6);
    --text-primary: #e2e8f0;
    --text-secondary: #94a3b8;
    --text-muted: #64748b;
    --success: #34d399;
    --error: #f87171;
    --user-bg: linear-gradient(135deg, #0e3a5c, #164e63);
    --bot-bg: rgba(15, 25, 45, 0.9);
    --radius: 12px;
  }

  * { box-sizing: border-box; margin: 0; padding: 0; }

  body {
    font-family: 'Inter', -apple-system, sans-serif;
    background: var(--bg-primary);
    color: var(--text-primary);
    padding: 0;
    font-size: 13px;
    line-height: 1.5;
  }

  .header {
    padding: 16px 12px;
    background: linear-gradient(180deg, rgba(6,182,212,0.08) 0%, transparent 100%);
    border-bottom: 1px solid var(--border);
    text-align: center;
  }
  .header h1 {
    font-size: 16px;
    font-weight: 700;
    background: var(--accent-gradient);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    letter-spacing: 0.5px;
  }
  .header .sub { font-size: 10px; color: var(--text-muted); margin-top: 3px; letter-spacing: 1.5px; text-transform: uppercase; }

  .tabs {
    display: flex;
    border-bottom: 1px solid var(--border);
    background: var(--bg-secondary);
  }
  .tab {
    flex: 1;
    padding: 10px 0;
    text-align: center;
    font-size: 11px;
    font-weight: 600;
    color: var(--text-muted);
    cursor: pointer;
    transition: all 0.3s ease;
    border-bottom: 2px solid transparent;
    text-transform: uppercase;
    letter-spacing: 0.8px;
  }
  .tab:hover { color: var(--text-secondary); background: rgba(6,182,212,0.05); }
  .tab.active { color: var(--accent); border-bottom-color: var(--accent); background: rgba(6,182,212,0.08); }

  .panel { display: none; height: calc(100vh - 108px); flex-direction: column; }
  .panel.active { display: flex; }

  #chat {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    scroll-behavior: smooth;
  }
  #chat::-webkit-scrollbar { width: 4px; }
  #chat::-webkit-scrollbar-track { background: transparent; }
  #chat::-webkit-scrollbar-thumb { background: var(--border); border-radius: 4px; }

  .msg {
    margin: 8px 0;
    padding: 12px 14px;
    border-radius: var(--radius);
    font-size: 13px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-word;
    animation: fadeIn 0.3s ease;
  }
  @keyframes fadeIn { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: translateY(0); } }

  .msg.user {
    background: var(--user-bg);
    border: 1px solid rgba(6,182,212,0.25);
    margin-left: 20px;
    border-radius: var(--radius) var(--radius) 4px var(--radius);
  }
  .msg.bot {
    background: var(--bot-bg);
    border: 1px solid var(--border);
    margin-right: 20px;
    border-radius: var(--radius) var(--radius) var(--radius) 4px;
    backdrop-filter: blur(8px);
    max-height: 50vh; /* Scrolly output */
    overflow-y: auto;
  }
  .msg.bot::-webkit-scrollbar { width: 4px; }
  .msg.bot::-webkit-scrollbar-track { background: rgba(0,0,0,0.2); border-radius: 4px; }
  .msg.bot::-webkit-scrollbar-thumb { background: var(--accent); border-radius: 4px; }

  .msg.error {
    background: rgba(248,113,113,0.08);
    border: 1px solid rgba(248,113,113,0.25);
    color: var(--error);
    font-size: 12px;
  }
  .cit { font-size: 11px; color: var(--text-muted); margin-top: 8px; padding-top: 8px; border-top: 1px solid var(--border); }

  .thinking { display: none; padding: 12px 16px; margin: 8px 12px; border-radius: var(--radius); background: var(--bg-glass); border: 1px solid var(--border-glow); color: var(--accent); font-size: 12px; animation: pulse 1.5s ease-in-out infinite; }
  .thinking.show { display: block; }
  @keyframes pulse { 0%,100% { opacity: 0.5; } 50% { opacity: 1; } }

  .input-area {
    padding: 12px;
    border-top: 1px solid var(--border);
    background: linear-gradient(180deg, var(--bg-secondary) 0%, rgba(6,182,212,0.03) 100%);
  }

  /* Multiple file badges container */
  .files-container {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-bottom: 8px;
  }

  .file-badge {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 8px;
    background: rgba(6,182,212,0.12);
    border: 1px solid var(--border-glow);
    border-radius: 6px;
    font-size: 10px;
    color: var(--accent-bright);
    animation: fadeIn 0.2s ease;
  }
  .file-badge .close { cursor: pointer; opacity: 0.7; font-size: 12px; margin-left: 2px; }
  .file-badge .close:hover { opacity: 1; color: var(--error); }

  select {
    width: 100%;
    padding: 9px 12px;
    background: var(--bg-primary);
    color: var(--text-primary);
    border: 1px solid var(--border);
    border-radius: 8px;
    font-family: 'Inter', sans-serif;
    font-size: 12px;
    outline: none;
    cursor: pointer;
    transition: border-color 0.2s;
  }
  select:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(6,182,212,0.1); }

  textarea {
    width: 100%;
    padding: 10px 12px;
    margin-top: 8px;
    background: var(--bg-primary);
    color: var(--text-primary);
    border: 1px solid var(--border);
    border-radius: 8px;
    font-family: 'Inter', sans-serif;
    font-size: 13px;
    resize: none;
    outline: none;
    transition: border-color 0.2s;
  }
  textarea:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(6,182,212,0.1); }
  textarea::placeholder { color: var(--text-muted); }

  .btn-row { display: flex; gap: 6px; margin-top: 8px; }

  button {
    flex: 1;
    padding: 10px;
    background: var(--accent-gradient);
    color: white;
    border: none;
    border-radius: 8px;
    font-family: 'Inter', sans-serif;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    letter-spacing: 0.3px;
  }
  button:hover { opacity: 0.9; transform: translateY(-1px); box-shadow: 0 4px 16px rgba(6,182,212,0.25); }
  button:active { transform: translateY(0); }

  .btn-secondary {
    background: var(--bg-glass);
    border: 1px solid var(--border);
    color: var(--text-secondary);
    flex: 0 0 auto;
    padding: 10px 14px;
  }
  .btn-secondary:hover { border-color: var(--accent); color: var(--accent); background: rgba(6,182,212,0.08); }

  .btn-run {
    margin-top: 8px;
    padding: 6px 10px;
    background: rgba(6,182,212,0.15);
    border: 1px solid var(--accent);
    color: var(--accent-bright);
    font-size: 10px;
    border-radius: 6px;
    width: 100%;
  }
  .btn-run:hover { background: var(--accent); color: var(--bg-primary); }

  #status { font-size: 11px; margin-top: 8px; color: var(--text-muted); text-align: center; }

  .tools-grid { padding: 12px; overflow-y: auto; flex: 1; }
  .tool-card {
    padding: 14px;
    margin-bottom: 8px;
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    transition: all 0.25s;
    backdrop-filter: blur(8px);
  }
  .tool-card:hover { border-color: var(--accent); box-shadow: 0 6px 20px rgba(6,182,212,0.12); }
  .tool-name { font-size: 13px; font-weight: 600; color: var(--accent-bright); margin-bottom: 4px; }
  .tool-desc { font-size: 11px; color: var(--text-secondary); line-height: 1.5; }

  .empty-state { text-align: center; padding: 40px 20px; color: var(--text-muted); }
  .empty-state .icon { font-size: 36px; margin-bottom: 12px; }
  .empty-state .title { font-size: 14px; font-weight: 600; color: var(--text-secondary); margin-bottom: 4px; }
</style>
</head>
<body>
  <div class="header">
    <h1>⚡ CodeEdgePro Nexus</h1>
    <div class="sub">AI Avengers • Next-Gen Engineering</div>
  </div>

  <div class="tabs">
    <div class="tab active" data-tab="chat">💬 Chat</div>
    <div class="tab" data-tab="tools">🔧 Tools</div>
  </div>

  <div class="panel active" id="panel-chat">
    <div id="chat"></div>
    <div class="thinking" id="thinking">⚡ Neural processing...</div>
    <div class="input-area">
      <div class="files-container" id="filesContainer"></div>
      <select id="experts"><option value="">⏳ Loading experts...</option></select>
      <textarea id="q" rows="3" placeholder="Ask your expert anything..."></textarea>
      <div class="btn-row">
        <button class="btn-secondary" id="attachBtn" title="Attach current file (Ctrl+Shift+A)">📎 Add File</button>
        <button id="send">Send ⚡</button>
      </div>
      <div id="status">Ready</div>
    </div>
  </div>

  <div class="panel" id="panel-tools">
    <div class="tools-grid" id="toolsGrid">
      <div class="empty-state">
        <div class="icon">🔧</div>
        <div class="title">Loading tools...</div>
        <div>Connect to see available tools</div>
      </div>
    </div>
  </div>

<script>
const vscode = acquireVsCodeApi();
const chat = document.getElementById('chat');
const q = document.getElementById('q');
const sel = document.getElementById('experts');
const status = document.getElementById('status');
const thinking = document.getElementById('thinking');
const toolsGrid = document.getElementById('toolsGrid');
const filesContainer = document.getElementById('filesContainer');

document.querySelectorAll('.tab').forEach(tab => {
  tab.addEventListener('click', () => {
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.panel').forEach(p => p.classList.remove('active'));
    tab.classList.add('active');
    document.getElementById('panel-' + tab.dataset.tab).classList.add('active');
  });
});

document.getElementById('send').onclick = () => {
  const question = q.value.trim();
  if (!question) return;
  const expertId = sel.value;
  if (!expertId) { status.textContent = '⚠️ Select an expert first'; return; }
  chat.innerHTML += '<div class="msg user">' + escapeHtml(question) + '</div>';
  q.value = '';
  chat.scrollTop = chat.scrollHeight;
  vscode.postMessage({ type: 'ask', expertId, question });
};

q.addEventListener('keydown', e => { if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) document.getElementById('send').click(); });
document.getElementById('attachBtn').onclick = () => vscode.postMessage({ type: 'attachFile' });

function escapeHtml(t) { return t.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;'); }

function getField(obj, ...names) {
  for (const n of names) { if (obj[n] !== undefined && obj[n] !== null) return obj[n]; }
  return undefined;
}

window.addEventListener('message', e => {
  const m = e.data;

  if (m.type === 'thinking') {
    thinking.classList.toggle('show', m.value);
    status.textContent = m.value ? '⚡ Processing...' : 'Ready';
  }

  if (m.type === 'answer') {
    chat.innerHTML += '<div class="msg bot">' + escapeHtml(m.text) + (m.citations?.length ? '<div class="cit">📚 ' + m.citations.map(c => escapeHtml(String(c))).join('<br>') + '</div>' : '') + '</div>';
    chat.scrollTop = chat.scrollHeight;
  }

  if (m.type === 'error') {
    chat.innerHTML += '<div class="msg error">❌ ' + escapeHtml(m.text) + '</div>';
    chat.scrollTop = chat.scrollHeight;
  }

  if (m.type === 'experts') {
    sel.innerHTML = '';
    const list = Array.isArray(m.experts) ? m.experts : [];
    if (list.length === 0) { sel.innerHTML = '<option value="">No experts available</option>'; status.textContent = '⚠️ No experts'; return; }
    list.forEach(ex => {
      const id = getField(ex, 'id', 'ID', 'Id');
      const name = getField(ex, 'name', 'Name', 'display_name', 'DisplayName');
      const slug = getField(ex, 'slug', 'Slug');
      if (!id) return;
      const o = document.createElement('option');
      o.value = id;
      o.textContent = '🤖 ' + (name || slug || id);
      sel.appendChild(o);
    });
    status.textContent = '✅ ' + sel.options.length + ' expert' + (sel.options.length > 1 ? 's' : '') + ' loaded';
  }

  if (m.type === 'tools') {
    const tools = m.tools || [];
    if (tools.length === 0) { toolsGrid.innerHTML = '<div class="empty-state"><div class="icon">🔒</div><div class="title">No tools available</div><div>Your token may not have tool access</div></div>'; return; }
    toolsGrid.innerHTML = '';
    tools.forEach(t => {
      const tName = getField(t, 'name', 'Name', 'title');
      const tDesc = getField(t, 'description', 'Description') || 'No description';
      const card = document.createElement('div');
      card.className = 'tool-card';
      
      const btn = document.createElement('button');
      btn.className = 'btn-run';
      btn.textContent = '▶️ Run Tool';
      btn.onclick = () => {
        const args = prompt('Run ' + tName + '\\nArguments (JSON):', '{}');
        if (args === null) return;
        try { const parsed = JSON.parse(args); vscode.postMessage({ type: 'callTool', toolName: tName, args: parsed }); document.querySelector('[data-tab="chat"]').click(); }
        catch { alert('Invalid JSON'); }
      };

      card.innerHTML = '<div class="tool-name">🔧 ' + escapeHtml(tName || 'unnamed') + '</div><div class="tool-desc">' + escapeHtml(tDesc) + '</div>';
      card.appendChild(btn);
      toolsGrid.appendChild(card);
    });
  }

  if (m.type === 'filesUpdated') {
    filesContainer.innerHTML = '';
    const files = m.files || [];
    files.forEach((f, idx) => {
      const badge = document.createElement('div');
      badge.className = 'file-badge';
      badge.innerHTML = '<span>📎 ' + escapeHtml(f.name) + (f.hasSelection ? ' (sel)' : '') + '</span><span class="close" title="Remove">✕</span>';
      badge.querySelector('.close').onclick = () => vscode.postMessage({ type: 'removeFile', index: idx });
      filesContainer.appendChild(badge);
    });
  }

  if (m.type === 'reload') {
    vscode.postMessage({ type: 'loadExperts' });
    vscode.postMessage({ type: 'loadTools' });
  }

  if (m.type === 'status') { status.textContent = m.text; }
});

setTimeout(() => {
  vscode.postMessage({ type: 'loadExperts' });
  vscode.postMessage({ type: 'loadTools' });
}, 500);
</script>
</body></html>`;
  }
}
