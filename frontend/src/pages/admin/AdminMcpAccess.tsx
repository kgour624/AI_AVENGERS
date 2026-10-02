import { useState, useEffect, useRef } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { createMcpToken, listMcpTokens, revokeMcpToken, listMcpV2Tools, createMcpV2Tool, generateMcpV2SQL, type McpTokenRecord, type ToolDefinition } from '@/api/mcp'
import { getAdminExperts } from '@/api/admin'
import type { Expert } from '@/types/expert'
import * as z from 'zod'

const toolSchema = z.object({
  name: z.string().min(1, "name required").regex(/^[a-z0-9_]+$/, "lowercase snake_case only"),
  display_name: z.string().min(1, "display_name required"),
  description: z.string().min(1, "description required"),
  input_schema: z.string().min(1, "input_schema required").refine((v) => { try { JSON.parse(v); return true } catch { return false } }, "must be valid JSON"),
})
type ToolForm = z.infer<typeof toolSchema>

/**
 * MCP Access — the tokens external coding agents (Claude Code and friends) use to
 * reach the domain experts.
 *
 * WHY this screen exists: a token can be minted from a CLI, which is fine for an
 * engineer and impossible for the person running the product. Everything an
 * operator needs is here: create, copy, see what a token may reach, and revoke.
 *
 * The token is shown EXACTLY ONCE, in the panel below the form. After that the
 * server only holds its hash, so the UI never pretends it can show it again.
 */
function AdminMcpAccess() {
  const queryClient = useQueryClient()
  
  const [toolName, setToolName] = useState("")
  const [toolDisplayName, setToolDisplayName] = useState("")
  const [toolDesc, setToolDesc] = useState("")
  const [toolSchemaStr, setToolSchemaStr] = useState('{\n  "type": "object",\n  "properties": {\n    "expert_id": {"type": "string"}\n  },\n  "required": ["expert_id"]\n}')
  const [toolErrors, setToolErrors] = useState<Record<string, string>>({})
  // Dynamic Engine — Phase 3: SQL_READ + API_CALL + LLM_PROMPT + COMPOSITE + Magic Generate 🪄 + Redis cache
  const [actionType, setActionType] = useState<'' | 'SQL_READ' | 'API_CALL' | 'LLM_PROMPT' | 'COMPOSITE'>('')
  const [sqlQuery, setSqlQuery] = useState("")
  const [generatePrompt, setGeneratePrompt] = useState("")
  const [isGenerating, setIsGenerating] = useState(false)
  const [generateError, setGenerateError] = useState("")
  const [apiUrl, setApiUrl] = useState("")
  const [apiMethod, setApiMethod] = useState("GET")
  const [apiHeadersStr, setApiHeadersStr] = useState('{}')
  const [apiBodyTemplate, setApiBodyTemplate] = useState("")
  const [llmPrompt, setLlmPrompt] = useState("")
  const [llmTier, setLlmTier] = useState("cheap")
  const [compositeStepsStr, setCompositeStepsStr] = useState('[\n  {"type":"SQL_READ","config":{"query":"SELECT name FROM experts LIMIT 10"}},\n  {"type":"LLM_PROMPT","config":{"prompt":"Summarize {{args._prev}} in Hindi","model_tier":"cheap"}}\n]')

  const createToolMutation = useMutation({
    mutationFn: async (data: ToolForm) => {
      const parsed = JSON.parse(data.input_schema)
      const body: any = { name: data.name, display_name: data.display_name, description: data.description, input_schema: parsed }
      if (actionType === 'SQL_READ' && sqlQuery.trim()) {
        body.action = { type: 'SQL_READ', config: { query: sqlQuery.trim(), allow_write: false }, timeout_ms: 5000 }
      } else if (actionType === 'API_CALL' && apiUrl.trim()) {
        let headers: any = {}
        try { headers = JSON.parse(apiHeadersStr || '{}') } catch { headers = {} }
        body.action = { type: 'API_CALL', config: { url: apiUrl.trim(), method: apiMethod, headers, body_template: apiBodyTemplate }, timeout_ms: 5000 }
      } else if (actionType === 'LLM_PROMPT' && llmPrompt.trim()) {
        body.action = { type: 'LLM_PROMPT', config: { prompt: llmPrompt.trim(), model_tier: llmTier }, timeout_ms: 10000 }
      } else if (actionType === 'COMPOSITE' && compositeStepsStr.trim()) {
        let steps: any = []
        try { steps = JSON.parse(compositeStepsStr) } catch { throw new Error("COMPOSITE steps must be valid JSON array") }
        body.action = { type: 'COMPOSITE', config: { steps }, timeout_ms: 10000 }
      }
      return createMcpV2Tool(body)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mcp-v2-tools'] })
      setToolName("")
      setToolDisplayName("")
      setToolDesc("")
      setToolSchemaStr('{\n  "type": "object",\n  "properties": {\n    "expert_id": {"type": "string"}\n  },\n  "required": ["expert_id"]\n}')
      setToolErrors({})
      setActionType('')
      setSqlQuery('')
      setGeneratePrompt('')
      setGenerateError('')
      setApiUrl('')
      setApiMethod('GET')
      setApiHeadersStr('{}')
      setApiBodyTemplate('')
      setLlmPrompt('')
      setCompositeStepsStr('[\n  {"type":"SQL_READ","config":{"query":"SELECT name FROM experts LIMIT 10"}},\n  {"type":"LLM_PROMPT","config":{"prompt":"Summarize {{args._prev}} in Hindi","model_tier":"cheap"}}\n]')
    },
  })

  const handleGenerateSQL = async () => {
    if (!generatePrompt.trim()) { setGenerateError("Prompt required"); return }
    setIsGenerating(true); setGenerateError("")
    try {
      const sql = await generateMcpV2SQL(generatePrompt.trim())
      setSqlQuery(sql)
    } catch (e: any) {
      setGenerateError(e?.response?.data?.error || e?.message || "Generate failed")
    } finally { setIsGenerating(false) }
  }

  const onCreateTool = (e: React.FormEvent) => {
    e.preventDefault()
    const result = toolSchema.safeParse({ name: toolName, display_name: toolDisplayName, description: toolDesc, input_schema: toolSchemaStr })
    if (!result.success) {
      const errors: Record<string, string> = {}
      for (const err of result.error.errors) {
        if (err.path[0]) errors[err.path[0].toString()] = err.message
      }
      setToolErrors(errors)
      return
    }
    if (actionType === 'SQL_READ' && !sqlQuery.trim()) {
      setToolErrors({ sql_query: "SQL query required for SQL_READ" })
      return
    }
    if (actionType === 'API_CALL' && !apiUrl.trim()) {
      setToolErrors({ api_url: "URL required for API_CALL" })
      return
    }
    if (actionType === 'API_CALL') {
      try { if (apiHeadersStr.trim()) JSON.parse(apiHeadersStr) } catch { setToolErrors({ api_headers: "headers must be valid JSON" }); return }
    }
    if (actionType === 'LLM_PROMPT' && !llmPrompt.trim()) {
      setToolErrors({ llm_prompt: "Prompt required for LLM_PROMPT" })
      return
    }
    if (actionType === 'COMPOSITE') {
      try { const v = JSON.parse(compositeStepsStr); if (!Array.isArray(v) || v.length===0) throw new Error(); } catch { setToolErrors({ composite_steps: "steps must be valid JSON array (1..5)" }); return }
    }
    setToolErrors({})
    createToolMutation.mutate(result.data)
  }

  const [label, setLabel] = useState('')
  const [selectedExpertIds, setSelectedExpertIds] = useState<string[]>([])
  const [selectedTools, setSelectedTools] = useState<string[]>([])
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const [expertsDropdownOpen, setExpertsDropdownOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)
  const expertsDropdownRef = useRef<HTMLDivElement>(null)
  const [newToken, setNewToken] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  const { data: tokens, isLoading } = useQuery({
    queryKey: ['admin', 'mcp-tokens'],
    queryFn: listMcpTokens,
  })

  const { data: tools = [], isLoading: toolsLoading, isError: toolsError } = useQuery({
    queryKey: ['mcp-v2-tools'],
    queryFn: listMcpV2Tools,
    staleTime: 5 * 60 * 1000,
  })

  const { data: experts = [], isLoading: expertsLoading, isError: expertsError } = useQuery({
    queryKey: ['admin', 'experts'],
    queryFn: getAdminExperts,
    staleTime: 5 * 60 * 1000,
  })

  const refresh = () => queryClient.invalidateQueries({ queryKey: ['admin', 'mcp-tokens'] })

  const createMutation = useMutation({
    mutationFn: () =>
      createMcpToken({
        label: label.trim(),
        expert_ids: selectedExpertIds,
        tools: selectedTools,
      }),
    onSuccess: (data) => {
      setNewToken(data.token)
      setCopied(false)
      setLabel('')
      setSelectedExpertIds([])
      setSelectedTools([])
      refresh()
    },
  })

  const revokeMutation = useMutation({
    mutationFn: (id: string) => revokeMcpToken(id),
    onSuccess: refresh,
  })

  // dropdown helpers
  const toggleTool = (name: string) => {
    setSelectedTools((prev) =>
      prev.includes(name) ? prev.filter((t) => t !== name) : [...prev, name]
    )
  }
  const handleSelectAll = () => {
    if (selectedTools.length === tools.length) {
      setSelectedTools([])
    } else {
      setSelectedTools(tools.map((t: ToolDefinition) => t.name))
    }
  }
  const handleClearAll = () => setSelectedTools([])
  const isAllSelected = tools.length > 0 && selectedTools.length === tools.length

  // experts dropdown helpers — granular per-expert scope
  const toggleExpert = (expertId: string) => {
    setSelectedExpertIds((prev) =>
      prev.includes(expertId) ? prev.filter((id) => id !== expertId) : [...prev, expertId]
    )
  }
  const handleSelectAllExperts = () => {
    if (selectedExpertIds.length === experts.length) {
      setSelectedExpertIds([])
    } else {
      setSelectedExpertIds(experts.map((e: Expert) => e.id))
    }
  }
  const handleClearAllExperts = () => setSelectedExpertIds([])
  const isAllExpertsSelected = experts.length > 0 && selectedExpertIds.length === experts.length

  useEffect(() => {
    const onClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false)
      }
      if (expertsDropdownRef.current && !expertsDropdownRef.current.contains(e.target as Node)) {
        setExpertsDropdownOpen(false)
      }
    }
    document.addEventListener("mousedown", onClickOutside)
    return () => document.removeEventListener("mousedown", onClickOutside)
  }, [])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-lg font-semibold text-text-primary">MCP Access</h1>
        <p className="mt-1 text-sm text-text-secondary">
          Tokens that let an external coding agent (Claude Code, Cursor) use your domain experts:
          read their standards, ask them questions, and have a change reviewed.
        </p>
      </div>

      <Card className="mb-6 space-y-4">
        <h2 className="text-sm font-semibold text-text-primary">Create New Tool — No-Code Builder</h2>
        <p className="text-xs text-text-secondary">Inserts into mcp_v2_tools and instantly refreshes dropdown below. Choose Action to make tool dynamic (no deploy).</p>
        <form onSubmit={onCreateTool} className="space-y-3">
          <input value={toolName} onChange={e => setToolName(e.target.value)} placeholder="name e.g. github_pr_review" className="w-full border rounded-lg px-3 py-2 text-sm bg-bg-primary border-border-default text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500" />
          {toolErrors.name && <p className="text-xs text-status-error">{toolErrors.name}</p>}
          <input value={toolDisplayName} onChange={e => setToolDisplayName(e.target.value)} placeholder="display_name e.g. GitHub PR Review" className="w-full border rounded-lg px-3 py-2 text-sm bg-bg-primary border-border-default text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500" />
          {toolErrors.display_name && <p className="text-xs text-status-error">{toolErrors.display_name}</p>}
          <textarea value={toolDesc} onChange={e => setToolDesc(e.target.value)} placeholder="description" rows={2} className="w-full border rounded-lg px-3 py-2 text-sm bg-bg-primary border-border-default text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500" />
          {toolErrors.description && <p className="text-xs text-status-error">{toolErrors.description}</p>}
          <textarea value={toolSchemaStr} onChange={e => setToolSchemaStr(e.target.value)} placeholder='{"type":"object","properties":{}}' rows={6} className="w-full border rounded-lg px-3 py-2 text-sm font-mono text-xs bg-bg-primary border-border-default text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500" />
          {toolErrors.input_schema && <p className="text-xs text-status-error">{toolErrors.input_schema}</p>}
          {/* Dynamic Engine Action */}
          <div className="border rounded-lg p-3 space-y-3 bg-white/5 border-border-subtle">
            <label className="text-xs font-medium text-text-primary">Action Type (optional — leave empty for static tool)</label>
            <select value={actionType} onChange={e => setActionType(e.target.value as any)} className="w-full border rounded-lg px-3 py-2 text-sm bg-bg-primary border-border-default text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500">
              <option value="">None (static)</option>
              <option value="SQL_READ">SQL_READ — SELECT only (Phase 1)</option>
              <option value="API_CALL">API_CALL — Webhook / External API (Phase 2)</option>
              <option value="LLM_PROMPT">LLM_PROMPT — AI Prompt (Phase 3)</option>
              <option value="COMPOSITE">COMPOSITE — SQL + API + LLM chain (Phase 3)</option>
            </select>
            {actionType === 'SQL_READ' && (
              <div className="space-y-2">
                <label className="text-xs font-medium text-text-primary">SQL Query (SELECT only, use {`{{args.xxx}}`} or $1)</label>
                <textarea value={sqlQuery} onChange={e => setSqlQuery(e.target.value)} placeholder="SELECT name, email FROM experts WHERE domain = {{args.domain}} LIMIT 100" rows={4} className="w-full border rounded-lg px-3 py-2 text-sm font-mono text-xs bg-bg-primary border-border-default text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500" />
                {toolErrors.sql_query && <p className="text-xs text-status-error">{toolErrors.sql_query}</p>}
                <div className="flex gap-2 items-start">
                  <input value={generatePrompt} onChange={e => setGeneratePrompt(e.target.value)} placeholder='e.g. experts ka naam aur email lao' className="flex-1 border rounded-lg px-3 py-2 text-sm bg-bg-primary border-border-default text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500" />
                  <Button type="button" disabled={isGenerating || !generatePrompt.trim()} onClick={handleGenerateSQL} variant="secondary" className="whitespace-nowrap">
                    {isGenerating ? "Generating..." : "🪄 Generate via AI"}
                  </Button>
                </div>
                {generateError && <p className="text-xs text-status-error">{generateError}</p>}
                <p className="text-[11px] text-text-disabled">Magic Button: English me likho, AI schema padh ke perfect SELECT bana dega. Bas Save daba do.</p>
              </div>
            )}
            {actionType === 'API_CALL' && (
              <div className="space-y-2">
                <label className="text-xs font-medium text-text-primary">Webhook URL (use {`{{args.xxx}}`} for dynamic)</label>
                <input value={apiUrl} onChange={e => setApiUrl(e.target.value)} placeholder="https://api.weather.com/v1/forecast?city={{args.city}}" className="w-full border rounded-lg px-3 py-2 text-sm font-mono text-xs bg-bg-primary border-border-default text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500" />
                {toolErrors.api_url && <p className="text-xs text-status-error">{toolErrors.api_url}</p>}
                <div className="grid grid-cols-2 gap-2">
                  <select value={apiMethod} onChange={e => setApiMethod(e.target.value)} className="border rounded-lg px-3 py-2 text-sm bg-bg-primary border-border-default text-text-primary">
                    <option value="GET">GET</option>
                    <option value="POST">POST</option>
                    <option value="PUT">PUT</option>
                    <option value="PATCH">PATCH</option>
                    <option value="DELETE">DELETE</option>
                  </select>
                  <input value={apiHeadersStr} onChange={e => setApiHeadersStr(e.target.value)} placeholder='{"Authorization":"Bearer {{args.token}}"}' className="border rounded-lg px-3 py-2 text-sm font-mono text-xs bg-bg-primary border-border-default text-text-primary" />
                </div>
                {toolErrors.api_headers && <p className="text-xs text-status-error">{toolErrors.api_headers}</p>}
                <label className="text-xs font-medium text-text-primary">Body Template (JSON, use {`{{args.xxx}}`})</label>
                <textarea value={apiBodyTemplate} onChange={e => setApiBodyTemplate(e.target.value)} placeholder='{"city":"{{args.city}}","units":"metric"}' rows={3} className="w-full border rounded-lg px-3 py-2 text-sm font-mono text-xs bg-bg-primary border-border-default text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500" />
                <p className="text-[11px] text-text-disabled">SSRF protected: 127.0.0.1/169.254.169.254/localhost blocked, timeout 5s, 1MB cap, no redirects.</p>
              </div>
            )}
            {actionType === 'LLM_PROMPT' && (
              <div className="space-y-2">
                <label className="text-xs font-medium text-text-primary">Prompt Template (only {`{{args.xxx}}`} allowed, no {`{{.Env}}`})</label>
                <textarea value={llmPrompt} onChange={e => setLlmPrompt(e.target.value)} placeholder="You are {{args.expert_name}}. Summarize {{args.topic}} in Hindi. Use only args." rows={4} className="w-full border rounded-lg px-3 py-2 text-sm font-mono text-xs bg-bg-primary border-border-default text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500" />
                {toolErrors.llm_prompt && <p className="text-xs text-status-error">{toolErrors.llm_prompt}</p>}
                <select value={llmTier} onChange={e => setLlmTier(e.target.value)} className="w-full border rounded-lg px-3 py-2 text-sm bg-bg-primary border-border-default text-text-primary">
                  <option value="cheap">cheap (default, 2k tokens)</option>
                  <option value="strong">strong</option>
                  <option value="fast">fast</option>
                </select>
                <p className="text-[11px] text-text-disabled">Prompt injection blocked, token limit 2k, tier cheap default. Strict {`{{args.xxx}}`} only.</p>
              </div>
            )}
            {actionType === 'COMPOSITE' && (
              <div className="space-y-2">
                <label className="text-xs font-medium text-text-primary">Steps JSON (array 1..5, SQL_READ/API_CALL/LLM_PROMPT) — use {`{{args.xxx}}`} / {`{{args._prev}}`}</label>
                <textarea value={compositeStepsStr} onChange={e => setCompositeStepsStr(e.target.value)} rows={6} className="w-full border rounded-lg px-3 py-2 text-sm font-mono text-xs bg-bg-primary border-border-default text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500" />
                {toolErrors.composite_steps && <p className="text-xs text-status-error">{toolErrors.composite_steps}</p>}
                <p className="text-[11px] text-text-disabled">Example: SQL → LLM chain. _prev holds previous step output. Each step isolated with timeout.</p>
              </div>
            )}
          </div>
          <Button type="submit" disabled={createToolMutation.isPending} variant="primary" className="w-full">
            {createToolMutation.isPending ? "Creating..." : "Create Tool"}
          </Button>
          {createToolMutation.isSuccess && <p className="text-xs text-status-success">Tool created — dropdown refreshed.</p>}
          {createToolMutation.isError && <p className="text-xs text-status-error">{(createToolMutation.error as Error).message}</p>}
        </form>
      </Card>

      <Card>
        <h2 className="text-sm font-semibold text-text-primary">New token</h2>
        <div className="mt-3 grid gap-3 sm:grid-cols-2">
          <Input
            label="Label"
            value={label}
            onChange={(e) => setLabel(e.target.value)}
            placeholder="e.g. Sneha laptop"
          />
          {/* Experts Multi-Select Dropdown — granular per-expert scope from GET /api/v1/admin/experts */}
          <div className="space-y-1">
            <label className="block text-xs font-medium text-text-primary">Experts (optional)</label>
            <div className="relative" ref={expertsDropdownRef}>
              <button
                type="button"
                onClick={() => setExpertsDropdownOpen((o) => !o)}
                className="w-full flex items-center justify-between rounded-md border border-border-subtle bg-transparent px-3 py-1.5 text-sm hover:border-brand focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand disabled:opacity-50"
                disabled={expertsLoading}
              >
                <span className={selectedExpertIds.length ? "text-text-primary" : "text-text-disabled"}>
                  {expertsLoading
                    ? "Loading experts..."
                    : selectedExpertIds.length === 0
                    ? "Select experts..."
                    : selectedExpertIds.length === experts.length
                    ? "All experts selected"
                    : `${selectedExpertIds.length} expert(s) selected`}
                </span>
              </button>
              {expertsDropdownOpen && (
                <div className="absolute z-50 mt-1 w-full rounded-md border border-border-subtle bg-[#1a1a1a] shadow-xl max-h-80 flex flex-col">
                  <div className="flex items-center justify-between px-3 py-2 border-b border-border-subtle text-xs bg-black/20">
                    <button type="button" onClick={handleSelectAllExperts} className="hover:text-brand font-medium text-text-primary transition">
                      {isAllExpertsSelected ? "Deselect All" : "Select All"}
                    </button>
                    <button type="button" onClick={handleClearAllExperts} className="hover:text-mode-refuse text-text-secondary transition">
                      Clear
                    </button>
                  </div>
                  <div className="overflow-y-auto flex-1 p-1">
                    {expertsError ? (
                      <div className="px-3 py-4 text-xs text-mode-refuse text-center">Failed to load experts.</div>
                    ) : experts.length === 0 ? (
                      <div className="px-3 py-4 text-xs text-text-disabled text-center">No experts found.</div>
                    ) : (
                      <ul className="space-y-0.5">
                        {experts.map((expert: Expert) => {
                          const checked = selectedExpertIds.includes(expert.id)
                          return (
                            <li key={expert.id}>
                              <label className={`flex items-start gap-2.5 px-2.5 py-2 rounded cursor-pointer transition ${checked ? "bg-brand/10" : "hover:bg-white/5"}`}>
                                <input type="checkbox" checked={checked} onChange={() => toggleExpert(expert.id)} className="mt-0.5" />
                                <div className="flex-1 min-w-0">
                                  <span className="block text-xs font-medium text-text-primary">{expert.name}</span>
                                  <span className="block text-[10px] text-text-disabled">{expert.domain} · {expert.slug}</span>
                                </div>
                              </label>
                            </li>
                          )
                        })}
                      </ul>
                    )}
                  </div>
                </div>
              )}
            </div>
            {selectedExpertIds.length > 0 && (
              <div className="flex flex-wrap gap-1.5 pt-2">
                {selectedExpertIds.map((id) => {
                  const expert = experts.find((e: Expert) => e.id === id)
                  return (
                    <span key={id} className="inline-flex items-center gap-1 bg-white/5 text-text-secondary text-[11px] px-2 py-0.5 rounded border border-border-subtle">
                      {expert?.name ?? id.slice(0, 8)}
                      <button type="button" onClick={() => toggleExpert(id)} className="hover:text-mode-refuse">×</button>
                    </span>
                  )
                })}
              </div>
            )}
          </div>
        </div>

        {/* Tools Dropdown Area */}
        <div className="mt-4 space-y-1">
          <label className="block text-xs font-medium text-text-primary">Tools (optional)</label>
          <div className="relative" ref={dropdownRef}>
            <button
              type="button"
              onClick={() => setDropdownOpen((o) => !o)}
              className="w-full flex items-center justify-between rounded-md border border-border-subtle bg-transparent px-3 py-1.5 text-sm hover:border-brand focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand disabled:opacity-50"
              disabled={toolsLoading}
            >
              <span className={selectedTools.length ? "text-text-primary" : "text-text-disabled"}>
                {toolsLoading
                  ? "Loading tools..."
                  : selectedTools.length === 0
                  ? "Select tools..."
                  : selectedTools.length === tools.length
                  ? "All tools selected"
                  : `${selectedTools.length} tool(s) selected`}
              </span>
            </button>
            
            {dropdownOpen && (
              <div className="absolute z-50 mt-1 w-full rounded-md border border-border-subtle bg-[#1a1a1a] shadow-xl max-h-80 flex flex-col">
                <div className="flex items-center justify-between px-3 py-2 border-b border-border-subtle text-xs bg-black/20">
                  <button type="button" onClick={handleSelectAll} className="hover:text-brand font-medium text-text-primary transition">
                    {isAllSelected ? "Deselect All" : "Select All"}
                  </button>
                  <button type="button" onClick={handleClearAll} className="hover:text-mode-refuse text-text-secondary transition">
                    Clear
                  </button>
                </div>
                <div className="overflow-y-auto flex-1 p-1">
                  {toolsError ? (
                    <div className="px-3 py-4 text-xs text-mode-refuse text-center">Failed to load tools.</div>
                  ) : tools.length === 0 ? (
                    <div className="px-3 py-4 text-xs text-text-disabled text-center">No tools found.</div>
                  ) : (
                    <ul className="space-y-0.5">
                      {tools.map((tool: ToolDefinition) => {
                        const checked = selectedTools.includes(tool.name)
                        return (
                          <li key={tool.id}>
                            <label className={`flex items-start gap-2.5 px-2.5 py-2 rounded cursor-pointer transition ${checked ? "bg-brand/10" : "hover:bg-white/5"}`}>
                              <input type="checkbox" checked={checked} onChange={() => toggleTool(tool.name)} className="mt-0.5" />
                              <div className="flex-1 min-w-0">
                                <span className="block text-xs font-medium text-text-primary">{tool.display_name}</span>
                                <span className="block text-[10px] text-text-disabled">{tool.description}</span>
                              </div>
                            </label>
                          </li>
                        )
                      })}
                    </ul>
                  )}
                </div>
              </div>
            )}
          </div>
          {selectedTools.length > 0 && (
            <div className="flex flex-wrap gap-1.5 pt-2">
              {selectedTools.map((name) => {
                const meta = tools.find((t: ToolDefinition) => t.name === name)
                return (
                  <span key={name} className="inline-flex items-center gap-1 bg-white/5 text-text-secondary text-[11px] px-2 py-0.5 rounded border border-border-subtle">
                    {meta?.display_name || name}
                    <button type="button" onClick={() => toggleTool(name)} className="hover:text-mode-refuse">×</button>
                  </span>
                )
              })}
            </div>
          )}
        </div>

        <p className="mt-3 text-xs text-text-disabled">
          Leave experts or tools empty to allow all of them. Restricting a token is what keeps one
          developer&apos;s agent out of experts it has no business in.
        </p>
        <div className="mt-4">
          <Button
            disabled={!label.trim() || createMutation.isPending}
            isLoading={createMutation.isPending}
            onClick={() => createMutation.mutate()}
          >
            Create token
          </Button>
        </div>

        {createMutation.isError && (
          <p className="mt-3 text-xs text-mode-refuse">
            Could not create the token. Check the label and try again.
          </p>
        )}

        {newToken && (
          <div className="mt-4 rounded-md border border-brand/40 bg-brand/5 p-3">
            <p className="text-xs font-medium text-text-primary">
              Copy this now — it is shown only once. The server stores only its hash.
            </p>
            <code className="mt-2 block break-all rounded bg-black/40 p-2 text-xs text-text-secondary">
              {newToken}
            </code>
            <div className="mt-2 flex items-center gap-3">
              <Button
                size="sm"
                variant="secondary"
                onClick={async () => {
                  await navigator.clipboard?.writeText(newToken)
                  setCopied(true)
                }}
              >
                {copied ? 'Copied' : 'Copy'}
              </Button>
              <button
                type="button"
                className="text-xs text-text-disabled hover:text-text-primary"
                onClick={() => setNewToken(null)}
              >
                Dismiss
              </button>
            </div>
          </div>
        )}
      </Card>

      <Card>
        <h2 className="text-sm font-semibold text-text-primary">Existing tokens</h2>
        {isLoading && <p className="mt-3 text-xs text-text-disabled">Loading tokens...</p>}
        {!isLoading && (tokens?.length ?? 0) === 0 && (
          <p className="mt-3 text-xs text-text-disabled">No tokens yet.</p>
        )}
        <div className="mt-3 space-y-2">
          {tokens?.map((token: McpTokenRecord) => (
            <div
              key={token.id}
              className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-border-subtle bg-white/5 px-3 py-2"
            >
              <div className="min-w-0 flex-1">
                <p className="truncate text-xs font-medium text-text-primary">
                  {token.label}
                  {token.revoked && <span className="ml-2 text-mode-refuse">revoked</span>}
                </p>
                <p className="mt-0.5 text-[11px] text-text-disabled">
                  experts: {token.expert_ids?.length ? token.expert_ids.slice(0,3).join(', ') + (token.expert_ids.length > 3 ? ` +${token.expert_ids.length - 3} more` : '') : token.domains?.length ? token.domains.join(', ') + ' (legacy domain)' : 'all'}
                  {' · '}tools: {token.tools?.length ? token.tools.join(', ') : 'all'}
                  {' · '}requests: {token.request_count}
                  {token.last_used_at ? ` · last used ${token.last_used_at.slice(0, 10)}` : ' · never used'}
                </p>
              </div>
              {!token.revoked && (
                <button
                  type="button"
                  disabled={revokeMutation.isPending}
                  onClick={() => revokeMutation.mutate(token.id)}
                  className="flex-shrink-0 rounded-md border border-border-subtle px-2.5 py-1 text-xs text-text-secondary transition hover:border-mode-refuse/40 hover:text-mode-refuse disabled:opacity-40"
                >
                  Revoke
                </button>
              )}
            </div>
          ))}
        </div>
        {revokeMutation.isError && (
          <p className="mt-2 text-xs text-mode-refuse">
            Revoke failed — the token may already be revoked.
          </p>
        )}
      </Card>
    </div>
  )
}

export const Component = AdminMcpAccess


