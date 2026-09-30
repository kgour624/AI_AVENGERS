// Package repo — GitHub/GitLab sync for MCP ( §4.3 End-to-End).
// Shallow clone -> chunk + bge embed -> mcp_repo_chunks vector(768) + ivfflat.
package repo

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// WebhookVerifier checks X-Hub-Signature-256 (GitHub) or GitLab token.
func VerifyGitHubSignature(secret, payload []byte, signature string) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(signature, prefix) {
		return false
	}
	sig, err := hex.DecodeString(strings.TrimPrefix(signature, prefix))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	expected := mac.Sum(nil)
	return hmac.Equal(sig, expected)
}

// SyncJob enqueues shallow clone + embed — WaitGroup+semaphore(8) + Contiguous Checkpoint (RULE 8-D:38-39).
type SyncJob struct {
	RepoConnectionID string
	RepoURL          string
	Branch           string
	AccessTokenRef   string // encrypted at rest, decrypt via vault port
}

// Validate checks repo url + branch.
func (j SyncJob) Validate() error {
	if j.RepoURL == "" || j.Branch == "" {
		return fmt.Errorf("repo_url and branch required")
	}
	return nil
}

// ChunkFile simulates chunking — real impl calls ml-sidecar bge-base-en-v1.5 + pgvector insert.
func ChunkFile(ctx context.Context, filePath, content string) []string {
	// Simple 500-char sliding window placeholder — real uses training/chunker.go pattern
	if len(content) <= 500 {
		return []string{content}
	}
	var chunks []string
	for i := 0; i < len(content); i += 500 {
		end := i + 500
		if end > len(content) {
			end = len(content)
		}
		chunks = append(chunks, content[i:end])
		if ctx.Err() != nil {
			break
		}
	}
	return chunks
}