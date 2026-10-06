// AdminMcpControlCenter — MCP Control Center (Ghar Baithe Baccha Control) 🤖
// Theme: cn() + slate-900 dark + FloatingChat.tsx pattern — Experts as Robots
import { useState } from "react"
import { cn } from "@/utils/cn"

type ExpertRow = {
  id: string
  name: string
  slug: string
  enabled: boolean
  status: "idle" | "in-use"
  platform: "claude" | "cursor" | "generic"
  dailyLimit: number
  monthlyLimit: number
  genericPercent: 0 | 5 | 10 | 20
}

export default function AdminMcpControlCenter() {
  const [experts] = useState<ExpertRow[]>([
    { id: "1", name: "Arpit System Design", slug: "arpit-system-design", enabled: true, status: "idle", platform: "claude", dailyLimit: 100000, monthlyLimit: 2000000, genericPercent: 0 },
  ])
  const [killSwitch, setKillSwitch] = useState(false)

  return (
    <div className={cn("min-h-screen bg-slate-900 text-slate-100 p-6")}>
      <div className="max-w-7xl mx-auto space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <span>🤖</span> MCP Control Center
            <span className="text-sm font-normal text-slate-400">— Experts as Robots</span>
          </h1>
          <label className="flex items-center gap-2 bg-red-950 border border-red-800 px-4 py-2 rounded-lg cursor-pointer">
            <input type="checkbox" checked={killSwitch} onChange={(e) => setKillSwitch(e.target.checked)} />
            <span className="text-sm font-semibold text-red-300">Global Kill Switch</span>
          </label>
        </div>

        {killSwitch && (
          <div className="bg-red-900/30 border border-red-700 rounded-lg p-3 text-sm text-red-300">
            Kill Switch ON — all tools/list returns [] + 403 mcp_disabled
          </div>
        )}

        <div className="bg-slate-800 rounded-xl border border-slate-700 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-slate-800 border-b border-slate-700 text-slate-400">
                <tr>
                  <th className="text-left p-3">Robot</th>
                  <th className="text-left p-3">Status</th>
                  <th className="text-left p-3">Enable</th>
                  <th className="text-left p-3">Daily / Monthly</th>
                  <th className="text-left p-3">Generic %</th>
                  <th className="text-left p-3">Provider (cheap|strong|fast)</th>
                  <th className="text-left p-3">Actions</th>
                </tr>
              </thead>
              <tbody>
                {experts.map((ex) => (
                  <tr key={ex.id} className="border-b border-slate-700/50 hover:bg-slate-700/30">
                    <td className="p-3 flex items-center gap-2">
                      <span className="w-8 h-8 rounded-full bg-slate-700 flex items-center justify-center text-lg">🤖</span>
                      <div>
                        <div className="font-medium">{ex.name}</div>
                        <div className="text-xs text-slate-400">{ex.slug}</div>
                      </div>
                    </td>
                    <td className="p-3">
                      <span className={cn("inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs", ex.status === "in-use" ? "bg-green-900/50 text-green-300" : "bg-slate-700 text-slate-400")}>
                        <span className={cn("w-2 h-2 rounded-full", ex.status === "in-use" ? "bg-green-400 animate-pulse" : "bg-slate-400")} />
                        {ex.status}
                      </span>
                    </td>
                    <td className="p-3">
                      <label className="relative inline-flex items-center cursor-pointer">
                        <input type="checkbox" checked={ex.enabled} readOnly className="sr-only peer" />
                        <div className="w-9 h-5 bg-slate-600 peer-checked:bg-emerald-600 rounded-full transition" />
                      </label>
                    </td>
                    <td className="p-3 text-xs text-slate-300">
                      {ex.dailyLimit.toLocaleString()} / {ex.monthlyLimit.toLocaleString()}
                    </td>
                    <td className="p-3">
                      <select value={ex.genericPercent} onChange={() => {}} className="bg-slate-700 border border-slate-600 rounded px-2 py-1 text-xs">
                        <option value={0}>0%</option>
                        <option value={5}>5%</option>
                        <option value={10}>10%</option>
                        <option value={20}>20% (red badge)</option>
                      </select>
                    </td>
                    <td className="p-3 text-xs text-slate-400">claude: strong — codecraftapi/gpt-6</td>
                    <td className="p-3">
                      <button className="text-xs bg-slate-700 hover:bg-slate-600 px-2 py-1 rounded">View</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
          <h3 className="font-semibold mb-2">Live — SSE GET /admin/mcp/live</h3>
          <p className="text-sm text-slate-400">Kaunsa robot kaunse platform (Claude/Cursor) pe abhi jawab de raha hai — yaha live dikhega (mcp_usage_log where duration is null).</p>
          <div className="mt-3 h-20 bg-slate-900 rounded border border-slate-700 flex items-center justify-center text-sm text-slate-500">
            No active runs — all robots idle 🤖💤
          </div>
        </div>
      </div>
    </div>
  )
}

export const Component = AdminMcpControlCenter;