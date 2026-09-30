package mcpv2

import "sync/atomic"

// KillSwitch is Global MCP Kill Switch — when ON, tools/list returns [] + 403 mcp_disabled ( §3.15 ).
// Backed by atomic bool + DB flag for persistence across restarts.
var killSwitch atomic.Bool

// IsKilled returns true if global kill is ON.
func IsKilled() bool { return killSwitch.Load() }

// SetKilled toggles global kill — Admin Control Center calls this + writes L3 master_event_log.
func SetKilled(on bool) { killSwitch.Store(on) }