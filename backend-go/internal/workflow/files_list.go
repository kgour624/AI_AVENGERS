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

// artifactPayload mirrors the content written by publishCodeArtifacts.
type artifactPayload struct {
	FilePath  string `json:"file_path"`
	Filename  string `json:"filename"`
	Content   string `json:"content"`
	Operation string `json:"operation"`
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
	events, err := h.store.GetByType(ctx, workflowID, []string{"code_artifact_produced"}, 0)
	if err != nil {
		return nil, err
	}

	out := []WorkflowFile{}
	seen := make(map[string]bool, len(events))
	for _, e := range events {
		var p artifactPayload
		if len(e.Content) > 0 {
			if err := json.Unmarshal(e.Content, &p); err != nil {
				continue
			}
		}
		path := strings.TrimSpace(p.FilePath)
		if path == "" {
			path = strings.TrimSpace(p.Filename)
		}
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		out = append(out, WorkflowFile{
			Path:       path,
			ExpertID:   e.PostedByExpertID,
			Operation:  p.Operation,
			HasContent: strings.TrimSpace(p.Content) != "",
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
	events, err := store.GetByType(ctx, workflowID, []string{"code_artifact_produced"}, 0)
	if err != nil {
		return "", nil, false
	}
	want := strings.TrimSpace(path)
	if want == "" {
		return "", nil, false
	}
	for _, e := range events {
		var p artifactPayload
		if len(e.Content) > 0 {
			if err := json.Unmarshal(e.Content, &p); err != nil {
				continue
			}
		}
		got := strings.TrimSpace(p.FilePath)
		if got == "" {
			got = strings.TrimSpace(p.Filename)
		}
		if got == want && strings.TrimSpace(p.Content) != "" {
			return p.Content, e.PostedByExpertID, true
		}
	}
	return "", nil, false
}
