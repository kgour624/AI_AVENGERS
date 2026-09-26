package workflow

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/response"
)

// WorkflowFile is one file a workflow actually produced, with the expert that
// produced it.
//
// WHY this exists (single source of truth): the FILES tab, the chat's
// list_sections tool and the on-disk workspace could disagree — a file could be
// listed that was never written, and the chat had no way to tell which expert
// owned a file. The code_artifact_produced event is written by the producer
// itself and carries file_path, the content, and posted_by_expert_id, so it is
// the one place where "this file exists, this expert made it, this is what is in
// it" is recorded together.
type WorkflowFile struct {
	Path       string     `json:"path"`
	ExpertID   *uuid.UUID `json:"expert_id,omitempty"`
	Operation  string     `json:"operation,omitempty"`
	HasContent bool       `json:"has_content"`
}

// artifactPathAndBody reads the path/operation when the producer supplied them
// and the body any artifact carries. Generic on purpose: the same function serves
// code files, documents, SQL, specs and anything else an expert posts, so no
// event type is special-cased.
func artifactPathAndBody(raw []byte) (path string, operation string, hasBody bool) {
	if len(raw) == 0 {
		return "", "", false
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		body, ok := artifactBody(raw)
		return "", "", ok && body != ""
	}
	if v, ok := m["file_path"].(string); ok {
		path = strings.TrimSpace(v)
	}
	if path == "" {
		if v, ok := m["filename"].(string); ok {
			path = strings.TrimSpace(v)
		}
	}
	if v, ok := m["operation"].(string); ok {
		operation = strings.TrimSpace(v)
	}
	body, ok := artifactBody(raw)
	return path, operation, ok && strings.TrimSpace(body) != ""
}

// ListFiles GET /workflows/:id/files
//
// Lists the produced files (path + owning expert + whether content is stored).
// The chat uses this to offer a file picker and to route a question about a file
// to the expert who actually wrote it.
func (h *Handler) ListFiles(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}
	files, err := h.workflowFiles(c.Request.Context(), workflowID)
	if err != nil {
		h.logger.Error("list workflow files failed",
			zap.String("workflow_id", workflowID.String()), zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, files)
}

// workflowFiles reads code_artifact_produced events and returns one entry per
// path, first producer winning (later waves that merely modify the file do not
// steal ownership from the expert who created it).
func (h *Handler) workflowFiles(ctx context.Context, workflowID uuid.UUID) ([]WorkflowFile, error) {
	events, err := h.store.GetSince(ctx, workflowID, 0, 500)
	if err != nil {
		return nil, err
	}

	out := []WorkflowFile{}
	seen := make(map[string]bool, len(events))
	for _, e := range events {
		// ANY artifact an expert produced counts here — not only code. A data
		// workflow's SQL/YAML, a product workflow's spec, and a design workflow's
		// document all appear, because the rule is "expert posted it and it has a
		// body", not "its event type is code_artifact_produced".
		if e.PostedByExpertID == nil || structuralEvents[e.EventType] {
			continue
		}
		path, operation, hasBody := artifactPathAndBody(e.Content)
		if !hasBody {
			continue
		}
		if path == "" {
			path = e.EventType
		}
		if seen[path] {
			continue
		}
		seen[path] = true
		out = append(out, WorkflowFile{
			Path:       path,
			ExpertID:   e.PostedByExpertID,
			Operation:  operation,
			HasContent: hasBody,
		})
	}
	return out, nil
}

// lookupArtifact returns the stored content of one produced file and the expert
// that produced it.
//
// WHY a free function over the store rather than a Handler method: the workflow
// chat service needs exactly this lookup when a question is scoped to a file,
// and it holds the store, not the Handler. One implementation means the picker
// (Handler) and the answer path (service) can never disagree about what a file
// contains or who owns it.
func lookupArtifact(ctx context.Context, store *blackboard.Store, workflowID uuid.UUID, path string) (string, *uuid.UUID, bool) {
	if store == nil {
		return "", nil, false
	}
	events, err := store.GetSince(ctx, workflowID, 0, 500)
	if err != nil {
		return "", nil, false
	}
	want := strings.TrimSpace(path)
	if want == "" {
		return "", nil, false
	}
	for _, e := range events {
		if e.PostedByExpertID == nil || structuralEvents[e.EventType] {
			continue
		}
		got, _, hasBody := artifactPathAndBody(e.Content)
		if got == "" {
			got = e.EventType
		}
		if got == want && hasBody {
			body, _ := artifactBody(e.Content)
			return body, e.PostedByExpertID, true
		}
	}
	return "", nil, false
}
