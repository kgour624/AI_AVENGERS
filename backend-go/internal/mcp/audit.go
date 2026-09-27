package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// AuditEntry is one recorded tool call.
type AuditEntry struct {
	TokenID     string
	TokenLabel  string
	Tool        string
	Domain      string
	Status      string // ok | tool_error | server_error
	ErrorCode   string
	LatencyMs   int
	InputBytes  int
	InputSHA256 string
}

// Auditor records tool calls without ever blocking a call on the database.
//
// WHY a queue and one writer goroutine: audit is bookkeeping, the tool call is
// the user's work. Writing inline would put a database round trip in front of
// every answer, and a slow database would show up as a slow expert. The parent
// (Close) owns the goroutine, so nothing outlives the server.
type Auditor struct {
	db     *pgxpool.Pool
	logger *zap.Logger
	ch     chan AuditEntry
	wg     sync.WaitGroup
}

// auditQueueSize is generous: entries are tiny and the writer keeps up with any
// human-paced agent. If it ever fills, dropping a record is preferable to
// delaying the answer.
const auditQueueSize = 256

// NewAuditor starts the writer and returns the auditor.
func NewAuditor(ctx context.Context, db *pgxpool.Pool, logger *zap.Logger) *Auditor {
	if logger == nil {
		logger = zap.NewNop()
	}
	a := &Auditor{db: db, logger: logger, ch: make(chan AuditEntry, auditQueueSize)}
	a.wg.Add(1)
	go a.run(ctx)
	return a
}

// Record queues one entry. Never blocks: if the queue is full the entry is
// dropped and said so, because losing a log line is better than stalling a call.
func (a *Auditor) Record(entry AuditEntry) {
	if a == nil {
		return
	}
	select {
	case a.ch <- entry:
	default:
		a.logger.Warn("mcp audit queue full — entry dropped", zap.String("tool", entry.Tool))
	}
}

// Close stops the writer after draining what is queued.
func (a *Auditor) Close() {
	if a == nil {
		return
	}
	close(a.ch)
	a.wg.Wait()
}

func (a *Auditor) run(ctx context.Context) {
	defer a.wg.Done()
	for entry := range a.ch {
		// A cancelled server context still lets the final entries through: the
		// drain path uses a short background context, so shutdown does not lose
		// the last calls.
		writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		_, err := a.db.Exec(writeCtx,
			`INSERT INTO mcp_audit
			   (token_id, token_label, tool, domain, status, error_code,
			    latency_ms, input_bytes, input_sha256)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			nullUUID(entry.TokenID), entry.TokenLabel, entry.Tool, entry.Domain,
			entry.Status, entry.ErrorCode, entry.LatencyMs, entry.InputBytes, entry.InputSHA256)
		cancel()
		if err != nil {
			a.logger.Warn("mcp audit write failed", zap.String("tool", entry.Tool), zap.Error(err))
		}
	}
}

// HashInput fingerprints a tool payload. The product records THAT a question was
// asked and can prove which one it was, without storing the client's code.
func HashInput(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// nullUUID keeps an empty token id out of the foreign key.
func nullUUID(id string) any {
	if id == "" {
		return nil
	}
	return id
}
