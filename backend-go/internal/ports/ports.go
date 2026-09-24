// Package ports defines cross-domain interfaces owned by consumers.
//
// WHY a dedicated package (not living inside each consumer):
//   Stage 0 of the scalable architecture keeps the modular monolith but
//   makes seams explicit so a later service split is an adapter swap, not
//   a rewrite. Ports live here so neither side of a boundary needs to
//   import the other's concrete package (DIP).
//
// SOLID:
//   ISP — each interface is small and one-purpose.
//   DIP — domains depend on these interfaces, not on Postgres/Redis adapters.
//
// Go idiom (Ultimate Go): interfaces are defined by the consumer and
// take values in. Adapters implement them in their own packages.
package ports

import (
	"context"

	"github.com/google/uuid"
)

// Capability is an action an account may perform against an expert.
// Additive: new capabilities do not break existing callers.
type Capability string

const (
	CapabilityChat     Capability = "chat"
	CapabilityWorkflow Capability = "workflow"
	CapabilityAPI      Capability = "api"
	CapabilityProject  Capability = "project"
)

// AccessDecision is the result of an entitlement check.
type AccessDecision struct {
	Allowed bool
	Reason  string
}

// EntitlementCheck answers "may this account use this expert for this capability?".
// Implemented by internal/entitlement. Seeded by user_expert_grants (migration 022).
// Future: plans/subscriptions/API keys without changing call sites.
type EntitlementCheck interface {
	// Can checks a single (account, expert, capability) triple.
	Can(ctx context.Context, accountID uuid.UUID, role string, expertID uuid.UUID, cap Capability) (AccessDecision, error)

	// FilterAllowed splits requested expert IDs into allowed vs denied.
	// admin/client: all requested pass. domain_expert: only granted IDs.
	FilterAllowed(ctx context.Context, accountID uuid.UUID, role string, requested []uuid.UUID) (allowed, denied []uuid.UUID, err error)

	// MustAllow returns a non-nil error when any requested expert is denied.
	MustAllow(ctx context.Context, accountID uuid.UUID, role string, expertIDs []uuid.UUID) error

	// GrantedSet returns the set of expert IDs granted to accountID.
	GrantedSet(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]struct{}, error)
}

// ExpertSummary is the minimal public view a consumer needs from Expert Knowledge.
// Kept small intentionally (Farley: translate at boundaries; local copy of immutable facts).
type ExpertSummary struct {
	ID     uuid.UUID
	Name   string
	Slug   string
	Domain string
	Active bool
}

// KnowledgeReader is the read-side port for the Expert Knowledge context.
// Conversation and Workflow call this instead of SQL on experts/course_chunks.
// Stage 0 ships the interface + Postgres adapter; call-site migration is incremental.
type KnowledgeReader interface {
	GetExperts(ctx context.Context, ids []uuid.UUID) ([]ExpertSummary, error)
	ListTrainedActive(ctx context.Context) ([]ExpertSummary, error)
}

// UsageEvent is one metered action (future billing / quota).
type UsageEvent struct {
	AccountID uuid.UUID
	ExpertID  *uuid.UUID
	Dimension string // e.g. "llm_tokens", "chat_message", "workflow_run"
	Quantity  float64
	Unit      string // e.g. "tokens", "count", "usd"
	Meta      map[string]string
}

// UsageRecorder records metered usage. Noop is valid (Null Object).
type UsageRecorder interface {
	Record(ctx context.Context, event UsageEvent) error
}

// DomainEvent is a cross-context fact written to the transactional outbox.
type DomainEvent struct {
	// AggregateType + AggregateID identify the source entity (e.g. "expert", id).
	AggregateType string
	AggregateID   uuid.UUID
	// EventType is a stable name: "entitlement.granted", "expert.published".
	EventType string
	// Payload is JSON-serializable domain data (keep small and versionable).
	Payload map[string]interface{}
}

// EventPublisher appends domain events to the outbox inside the caller's
// business transaction (or immediately after). Dispatcher ships them later.
type EventPublisher interface {
	Publish(ctx context.Context, events ...DomainEvent) error
}

// NoopUsageRecorder is a Null Object — safe default when metering is off.
type NoopUsageRecorder struct{}

func (NoopUsageRecorder) Record(context.Context, UsageEvent) error { return nil }

// NoopEventPublisher drops events (tests / early boot).
type NoopEventPublisher struct{}

func (NoopEventPublisher) Publish(context.Context, ...DomainEvent) error { return nil }
