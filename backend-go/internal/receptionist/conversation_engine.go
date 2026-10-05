package receptionist

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ConversationEngine struct {
	store   *Store
	search  SearchProvider
	pub     *Publisher
	llm     LLMClient
	logger  *zap.Logger
}

func NewConversationEngine(s *Store, search SearchProvider, pub *Publisher, llm LLMClient, l *zap.Logger) *ConversationEngine {
	return &ConversationEngine{store: s, search: search, pub: pub, llm: llm, logger: l}
}

func (e *ConversationEngine) HandleChat(ctx context.Context, sessionID uuid.UUID, userMsg string) (string, error) {
	return "", nil
}

func (e *ConversationEngine) IsAllDiscussed(pts []Point) bool {
	if len(pts) == 0 {
		return false
	}
	for _, p := range pts {
		if p.Status == "" || p.Status == "draft" || p.Status == "discussing" {
			return false
		}
	}
	return true
}

func (e *ConversationEngine) HandleAnswer(ctx context.Context, sessionID, checkpointID, tenantID uuid.UUID, pointID, answer string) ([]Point, string, error) {
	pts, _, err := e.store.GetTemplate(ctx, checkpointID)
	if err != nil {
		return nil, "", err
	}
	var explain string
	updated := false
	for i, p := range pts {
		if p.ID == pointID || p.QText == pointID {
			if strings.EqualFold(answer, "Haan") || strings.EqualFold(answer, "Yes") {
				pts[i].Status = "approved"
			} else if strings.EqualFold(answer, "Nahi") || strings.EqualFold(answer, "No") {
				pts[i].Status = "dismissed"
			} else if strings.Contains(strings.ToLower(answer), "samjhao") || strings.Contains(strings.ToLower(answer), "explain") {
				pts[i].Status = "judge_explained"
				pts[i].JudgeExplained = true
				if e.llm != nil {
					explain, _ = e.llm.Complete(ctx, "Explain in simple terms for receptionist checklist", p.QText)
				}
				if explain == "" {
					explain = "Yeh point aapke overall agenda ke context aur 2026 industry standard baseline ko align karne ke liye zaroori hai."
				}
			} else {
				pts[i].Status = "approved"
			}
			updated = true
			break
		}
	}
	if !updated && len(pts) > 0 {
		pts[0].Status = "approved"
	}
	_ = e.store.UpsertTemplate(ctx, checkpointID, sessionID, tenantID, 0, "", pts)
	return pts, explain, nil
}

func (e *ConversationEngine) AddMorePoints(ctx context.Context, sessionID, checkpointID, tenantID uuid.UUID, agenda string, existing []Point) ([]Point, error) {
	newPts := append([]Point{}, existing...)
	if e.search != nil {
		additional, err := e.search.Search(ctx, agenda+" extra points")
		if err == nil {
			for _, p := range additional {
				newPts = append(newPts, Point{
					ID:     uuid.New().String(),
					QText:  p.QText,
					Status: "draft",
					Source: "search",
				})
			}
		}
	}
	if len(newPts) == len(existing) {
		newPts = append(newPts, Point{
			ID:     uuid.New().String(),
			QText:  "Additional validation check for " + agenda,
			Status: "draft",
			Source: "ai",
		})
	}
	_ = e.store.UpsertTemplate(ctx, checkpointID, sessionID, tenantID, 0, agenda, newPts)
	return newPts, nil
}
