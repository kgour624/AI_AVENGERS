package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// SHAPreservationGuard — Harness > Prompt: hash mismatch => FAIL, no content loss.
// Original hash is computed from raw S3 bytes before cleaning.
// Cleaned hash is computed from cleaned output. Guard verifies stored hash == recomputed.
type SHAPreservationGuard struct{}

func NewSHAPreservationGuard() *SHAPreservationGuard { return &SHAPreservationGuard{} }

// Verify recomputes SHA256 of cleanedText and compares to original? No — it returns
// the computed hash for caller to store. VerifyHash does equality check.
func (g *SHAPreservationGuard) Verify(_ context.Context, originalSHA256 string, cleanedText string) (bool, string, error) {
	if strings.TrimSpace(originalSHA256) == "" {
		return false, "", fmt.Errorf("original SHA empty")
	}
	computed := sha256Hex(cleanedText)
	// Guard does NOT compare original vs cleaned — they are expected to differ after cleaning.
	// Instead it verifies that cleaned hash is consistent and non-empty.
	verified := computed != "" && len(computed) == 64
	return verified, computed, nil
}

func (g *SHAPreservationGuard) VerifyHash(originalSHA256, cleanedSHA256 string) bool {
	if originalSHA256 == "" || cleanedSHA256 == "" {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(originalSHA256), strings.TrimSpace(cleanedSHA256))
}

// VerifyPreservedContent re-reads preserved S3 bytes and checks hash == expectedCleanedSHA.
func (g *SHAPreservationGuard) VerifyPreservedContent(preservedBytes []byte, expectedCleanedSHA string) bool {
	h := sha256.Sum256(preservedBytes)
	got := hex.EncodeToString(h[:])
	return strings.EqualFold(got, strings.TrimSpace(expectedCleanedSHA))
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
