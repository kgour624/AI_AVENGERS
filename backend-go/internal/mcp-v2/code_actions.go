package mcpv2

import (
	"context"
	"fmt"
)

// CodeAction handles review_code / write_tests / generate_docs / fix_error — Power Pack ( §3.11 )
// Never runs code inside MCP server — returns structuredContent {patch} + ui://code-preview for Host to apply via sendMcpMessage("tool", apply_patch).

type ReviewRequest struct {
	ExpertID string `json:"expertId"`
	FilePath string `json:"filePath"`
	RepoURL  string `json:"repoUrl,omitempty"`
	Diff     string `json:"diff,omitempty"`
}

type ReviewResult struct {
	Review    string   `json:"review"`
	Citations []Chunk  `json:"citations"`
	Patch     string   `json:"patch,omitempty"` // diff for Host apply_patch
}

func (s *Service) ReviewCode(ctx context.Context, req ReviewRequest) (ReviewResult, error) {
	if req.ExpertID == "" {
		return ReviewResult{}, fmt.Errorf("expertId required")
	}
	// 1. Expert-scoped search: course_chunks + repo_chunks
	q := req.Diff
	if q == "" {
		q = req.FilePath
	}
	if q == "" {
		q = "review code"
	}
	chunks, _, err := s.SearchChunks(ctx, req.ExpertID, q, 5, "")
	if err != nil {
		return ReviewResult{}, err
	}
	repoChunks, _ := s.deps.DB.SearchRepoChunks(ctx, req.ExpertID, nil, 5)
	for _, rc := range repoChunks {
		chunks = append(chunks, Chunk{ID: rc.ID, ExpertID: req.ExpertID, Text: rc.Text, Score: rc.Score})
	}
	// 2. Sampling to generate cited review (borrows Host LLM, includeContext none)
	msgs := []Message{{Role: "user", Content: "Review this code as expert " + req.ExpertID + " citing chunks:\nFile: " + req.FilePath + "\nDiff: " + req.Diff}}
	for _, c := range chunks {
		msgs = append(msgs, Message{Role: "user", Content: "CITATION: " + c.Text})
	}
	output, _, err := s.deps.Gateway.Call(ctx, "strong", msgs)
	if err != nil {
		return ReviewResult{}, fmt.Errorf("gateway: %w", err)
	}
	return ReviewResult{Review: output, Citations: chunks, Patch: ""}, nil
}

type WriteTestsRequest struct {
	ExpertID  string `json:"expertId"`
	FilePath  string `json:"filePath"`
	Framework string `json:"framework"` // jest|go-test
}

func (s *Service) WriteTests(ctx context.Context, req WriteTestsRequest) (ReviewResult, error) {
	if req.ExpertID == "" || req.FilePath == "" {
		return ReviewResult{}, fmt.Errorf("expertId and filePath required")
	}
	if req.Framework == "" {
		req.Framework = "jest"
	}
	return s.ReviewCode(ctx, ReviewRequest{ExpertID: req.ExpertID, FilePath: req.FilePath, Diff: "write_tests framework=" + req.Framework})
}