package admin

import (
	"testing"

	"github.com/gin-gonic/gin"
)

// Gin panics at REGISTRATION time when two routes conflict (a static segment
// fighting a wildcard at the same position, or a duplicate path). That panic
// happens during startup, so without a test the first signal is a crash in a
// deployed environment — the worst possible place to learn about it.
//
// This test registers the expert route tree exactly as cmd/server/main.go does,
// with the same segment shapes. It is deliberately a copy of those paths: the
// point is to fail loudly HERE when a new sibling route is added that collides
// with a neighbour, and the neighbours are the thing the test names.
func TestExpertRouteTreeRegistersWithoutConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	group := engine.Group("/api/v1/admin")

	noop := func(*gin.Context) {}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("gin route registration panicked — a conflicting route was added: %v", recovered)
		}
	}()

	// Existing tree (before Phase D): 'ingest' and 'ingestion' share a prefix at
	// the same position, and the jobs subtree mixes static and wildcard children.
	group.POST("/experts/:id/ingest", noop)
	group.GET("/experts/:id/jobs", noop)
	group.GET("/experts/:id/jobs/stream", noop)
	group.GET("/experts/:id/jobs/events", noop)
	group.POST("/experts/:id/jobs/:jobID/resume", noop)
	group.POST("/experts/:id/jobs/:jobID/retry", noop)
	group.GET("/experts/:id/versions", noop)
	group.POST("/experts/:id/versions/:versionId/pin", noop)

	// Phase D additions.
	group.GET("/experts/:id/ingestion/audit", noop)
	group.GET("/experts/:id/ingestion/diagnostics", noop)
	group.POST("/experts/:id/ingestion/reconcile", noop)

	// I2 additions: the same static segment on two verbs.
	group.POST("/experts/:id/capability-eval", noop)
	group.GET("/experts/:id/capability-eval", noop)

	// Route count is asserted too: a silently dropped registration would not
	// panic, and "no panic" must not be mistaken for "all of them registered".
	if got := len(engine.Routes()); got != 13 {
		t.Fatalf("registered routes = %d, want 13", got)
	}
}
