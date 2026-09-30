// Package business — Core domain types (Business model — RULE 8-B:33, Ultimate Go §7).
// Business defines its own types; App imports Business (imports down only — §1).
package business

// Expert is Business model (core). No JSON tags.
type Expert struct {
	ID      string
	Name    string
	Slug    string
	Charter string
}
type Chunk struct {
	ID       string
	ExpertID string
	Text     string
	Score    float32
}
type RepoChunk struct {
	ID       string
	FilePath string
	Text     string
	Score    float32
}
type UsageLog struct {
	ExpertID       string
	Platform       string
	Provider       string
	Model          string
	Tier           string
	InputTokens    int
	OutputTokens   int
	TotalTokens    int
	CostUSD        float64
	GenericUsed    bool
	GenericPercent int
	DurationMs     int
}
type ExpertLimits struct {
	ExpertID         string
	Platform         string
	DailyLimit       int
	MonthlyLimit     int
	CurrentDaily     int
	CurrentMonthly   int
	BlockedUntilUnix *int64
}
type Message struct{ Role, Content string }
type GatewayUsage struct{ InputTokens, OutputTokens int; CostUSD float64 }
type GateResult struct{ GateStopped int; Allowed bool }