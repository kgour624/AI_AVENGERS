package receptionist

import (
	"context"
	"fmt"

	"ai_avengers/backend/internal/gateway"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SimpleLLMAdapter - mock/test adapter (kept for compatibility)
type SimpleLLMAdapter struct {
	Call func(ctx context.Context, system, user string) (string, error)
}

func (a *SimpleLLMAdapter) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if a.Call != nil {
		return a.Call(ctx, systemPrompt, userPrompt)
	}
	return fmt.Sprintf("[MOCK %s]\nSystem: %s\n\nUser: %s\n\n[End Mock]", "LLM", systemPrompt[:min(200, len(systemPrompt))], userPrompt[:min(800, len(userPrompt))]), nil
}

type SimpleExpertAdapter struct {
	Fetch func(ctx context.Context, expertID uuid.UUID, question string) (string, error)
	Names map[string]string
}

func (a *SimpleExpertAdapter) CallExpert(ctx context.Context, expertID uuid.UUID, questionEnglish string) (string, error) {
	if a.Fetch != nil {
		return a.Fetch(ctx, expertID, questionEnglish)
	}
	return fmt.Sprintf("[MOCK Expert %s] Answer to: %s\n\nThis is a mock expert response. Replace SimpleExpertAdapter.Fetch with real expert RAG call. Relevant extraction will run on this.", expertID.String()[:8], questionEnglish), nil
}

func (a *SimpleExpertAdapter) GetExpertName(ctx context.Context, expertID uuid.UUID) (string, error) {
	if n, ok := a.Names[expertID.String()]; ok {
		return n, nil
	}
	return "Expert-" + expertID.String()[:8], nil
}

// receptionistLLMAdapter - real gateway adapter used in main.go wiring (FIX 6)
type receptionistLLMAdapter struct {
	gateway *gateway.ModelGateway
}

func (r *receptionistLLMAdapter) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	resp, err := r.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelCheap,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.6,
		MaxTokens:    1500,
	})
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", fmt.Errorf("llm nil response")
	}
	return resp.Content, nil
}

var _ LLMClient = (*receptionistLLMAdapter)(nil)

// receptionistExpertAdapter - real expert RAG adapter for ExpertCaller (with db for RAG context)
type receptionistExpertAdapter struct {
	gateway *gateway.ModelGateway
	db      *pgxpool.Pool
}

func (r *receptionistExpertAdapter) CallExpert(ctx context.Context, expertID uuid.UUID, questionEnglish string) (string, error) {
	resp, err := r.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelStrong,
		SystemPrompt: fmt.Sprintf("You are expert %s. Answer in English markdown, verifiable.", expertID.String()),
		UserPrompt:   questionEnglish,
		Temperature:  0.2,
		MaxTokens:    3000,
	})
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", fmt.Errorf("expert llm nil response")
	}
	return resp.Content, nil
}

func (r *receptionistExpertAdapter) GetExpertName(ctx context.Context, expertID uuid.UUID) (string, error) {
	return "Expert-" + expertID.String()[:8], nil
}

var _ ExpertCaller = (*receptionistExpertAdapter)(nil)
var _ ExpertCaller = (*SimpleExpertAdapter)(nil)
