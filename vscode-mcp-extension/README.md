# CodeEdgePro Nexus ⚡

**AI Avengers Protocol for Next-Gen Software Architecture and Engineering**

---

## 🚀 Features

### 💬 Expert Chat
Ask domain-specific questions to your AI-powered engineering experts. Each expert is trained on specific knowledge domains — from System Design to Product Management.

- **Token-Scoped Access** — Only experts granted in your MCP token are visible
- **File Context Attach** — Attach your current file or selected code to provide context to the expert
- **Citation-Backed Answers** — Every answer includes source citations from the knowledge base
- **Ctrl+Enter to Send** — Quick keyboard shortcut for fast workflows

### 🔧 Dynamic Tools
Access all MCP tools granted by your token directly from the Tools tab.

- **ask_expert** — Ask a domain expert a grounded question (RAG)
- **search_chunks** — Semantic search over expert knowledge chunks
- **search_course_content** — Search course materials via pgvector
- **search_repo_content** — Semantic search over connected GitHub repos
- **get_standards** — Fetch expert coding and quality standards
- **review_change** — Review a code diff against expert standards
- **get_usage_logs** — Fetch MCP token usage analytics and limits
- **list_experts** — List all active domain experts

### 📎 File Context
Attach your current file or selected code directly into the chat for context-aware expert responses.

- **Ctrl+Shift+A** — Quick attach current file
- **Selection Support** — Select specific code and attach only the selection
- **Auto-Context** — Attached file content is automatically included in your question

### 🔒 Token-Based Security
Every request is authenticated with your MCP Bearer token. Only experts and tools explicitly granted in your token scope are accessible.

### 🌐 Remote MCP Backend
Connect to any remote AI Avengers MCP backend via HTTP. Designed for production deployment with Docker-based infrastructure.

---

## ⚙️ Setup

1. **Set MCP URL**: Settings → CodeEdgePro → Mcp Url (e.g., `http://192.168.1.7:8081/mcp/v2`)
2. **Set Token**: `Ctrl+Shift+P` → `CodeEdgePro: Set MCP Token`
3. **Connect**: Click status bar or `Ctrl+Shift+P` → `CodeEdgePro: Connect`

## 🎨 Commands

| Command | Description |
|---------|-------------|
| `CodeEdgePro: Set MCP Token` | Set/update your Bearer auth token |
| `CodeEdgePro: Connect` | Connect to MCP backend |
| `CodeEdgePro: Disconnect` | Disconnect from MCP backend |
| `CodeEdgePro: Clear Chat` | Clear chat history |
| `CodeEdgePro: Attach Current File` | Attach active file to chat context |

## 🛠 Keybindings

| Shortcut | Action |
|----------|--------|
| `Ctrl+Shift+A` | Attach current file |
| `Ctrl+Enter` | Send message |

---

*Built by CodeEdgePro — AI Avengers Engineering Division*
