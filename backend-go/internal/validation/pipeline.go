package validation

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// Language constants for supported validators.
const (
	LangGo         = "go"
	LangTypeScript = "typescript"
	LangJavaScript = "javascript"
	LangUnknown    = "unknown"
)

// maxRevisionRounds is the maximum OTA revision rounds before giving up.
// WHY defined here (not imported from workflow package):
//   validation package must not import workflow — circular dependency.
//   Value matches workflow.MaxRevisionRounds = 3 by design.
const maxRevisionRounds = 3

// ArtifactRequest is the input to the validation pipeline.
type ArtifactRequest struct {
	Filename string // e.g. "handler.go", "api.ts"
	Code     string // full source code
	ExpertID string // for logging
}

// PipelineResult is the output of the full validation pipeline.
type PipelineResult struct {
	Passed         bool
	Language       string
	FinalCode      string // may differ from input if revision loop fixed it
	RevisionRounds int    // how many OTA rounds were needed (0 = passed first try)
	LastResult     *ValidationResult
	Error          string
}

// Pipeline orchestrates validation across all supported languages.
// Implements the OTA revision loop from DOMAIN_EXPERT_COLLABORATION_DESIGN.md §11.3.
type Pipeline struct {
	goValidator *GoValidator
	tsValidator *TSValidator
	gateway     *gateway.ModelGateway
	logger      *zap.Logger
}

// NewPipeline creates a new validation pipeline.
func NewPipeline(gw *gateway.ModelGateway, logger *zap.Logger) *Pipeline {
	return &Pipeline{
		goValidator: NewGoValidator(),
		tsValidator: NewTSValidator(),
		gateway:     gw,
		logger:      logger,
	}
}

// Validate runs the full validation pipeline for a code artifact.
// If validation fails, enters the OTA revision loop (max 3 rounds).
//
// Mental execution (Go file, first try passes):
//   Input: filename="handler.go", code="package main...\n"
//   1. detectLanguage("handler.go") -> "go"
//   2. goValidator.Validate() -> {Passed: true}
//   3. Return PipelineResult{Passed: true, RevisionRounds: 0}
//
// Mental execution (TS file, needs 1 revision):
//   Input: filename="api.ts", code="const x = 1" (missing semicolon for prettier)
//   1. detectLanguage("api.ts") -> "typescript"
//   2. tsValidator.Validate() -> {Passed: false, Stage: "format", Error: "needs prettier"}
//   3. OTA round 1: LLM fixes code -> "const x = 1;\n"
//   4. tsValidator.Validate() -> {Passed: true}
//   5. Return PipelineResult{Passed: true, RevisionRounds: 1, FinalCode: "const x = 1;\n"}
func (p *Pipeline) Validate(ctx context.Context, req ArtifactRequest) *PipelineResult {
	lang := detectLanguage(req.Filename)

	if lang == LangUnknown {
		p.logger.Warn("validation: unsupported language, skipping",
			zap.String("filename", req.Filename),
			zap.String("expert_id", req.ExpertID),
		)
		return &PipelineResult{
			Passed:    true, // pass-through for unsupported languages
			Language:  lang,
			FinalCode: req.Code,
		}
	}

	currentCode := req.Code

	for round := 0; round <= maxRevisionRounds; round++ {
		// Run validator for this language
		var result *ValidationResult
		switch lang {
		case LangGo:
			result = p.goValidator.Validate(ctx, req.Filename, currentCode)
		case LangTypeScript, LangJavaScript:
			result = p.tsValidator.Validate(ctx, req.Filename, currentCode)
		}

		if result.Passed {
			p.logger.Info("validation passed",
				zap.String("filename", req.Filename),
				zap.String("language", lang),
				zap.Int("revision_rounds", round),
				zap.Int64("duration_ms", result.DurationMs),
			)
			return &PipelineResult{
				Passed:         true,
				Language:       lang,
				FinalCode:      currentCode,
				RevisionRounds: round,
				LastResult:     result,
			}
		}

		p.logger.Warn("validation failed",
			zap.String("filename", req.Filename),
			zap.String("stage", result.Stage),
			zap.String("error", result.Error),
			zap.Int("round", round),
		)

		// All revision rounds exhausted
		if round == MaxRevisionRounds {
			return &PipelineResult{
				Passed:         false,
				Language:       lang,
				FinalCode:      currentCode,
				RevisionRounds: round,
				LastResult:     result,
				Error: fmt.Sprintf("validation failed after %d revision rounds: [%s] %s",
					round, result.Stage, result.Error),
			}
		}

		// OTA revision: ask LLM to fix the code
		// Observe: result.Error + result.Output
		// Think+Act: LLM returns fixed code
		fixedCode, fixErr := p.requestFix(ctx, req.Filename, currentCode, result)
		if fixErr != nil {
			p.logger.Warn("validation: LLM fix request failed",
				zap.Int("round", round),
				zap.Error(fixErr),
			)
			// Can't get a fix — stop retrying
			return &PipelineResult{
				Passed:         false,
				Language:       lang,
				FinalCode:      currentCode,
				RevisionRounds: round,
				LastResult:     result,
				Error:          fmt.Sprintf("LLM fix unavailable: %s", fixErr.Error()),
			}
		}
		currentCode = fixedCode
	}

	// Should not reach here (loop exits via return above)
	return &PipelineResult{Passed: false, Language: lang, FinalCode: currentCode}
}

// requestFix asks the LLM to fix a validation error.
// Uses ModelCheap (deterministic, low temperature) because this is
// a mechanical fix, not creative reasoning.
//
// Mental execution:
//   Input: filename="handler.go", code="...", result={Stage:"format", Error:"needs gofmt"}
//   Prompt: "Fix this Go file. Validation error: [format] needs gofmt\n\nCode:\n..."
//   Output: fixed Go code (extracted from LLM response)
func (p *Pipeline) requestFix(
	ctx context.Context,
	filename, code string,
	result *ValidationResult,
) (string, error) {
	prompt := fmt.Sprintf(
		`You are a code fixer. Fix the following file to pass validation.

Filename: %s
Validation stage that failed: %s
Error: %s
Validator output:
%s

Return ONLY the fixed code, no explanation, no markdown fences.`,
		filename, result.Stage, result.Error, result.Output,
	)

	resp, err := p.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelCheap,
		SystemPrompt: "You are a precise code fixer. Output only valid source code.",
		UserPrompt:   prompt + "\n\nOriginal code:\n" + code,
		MaxTokens:    2000,
		Temperature:  0.1, // deterministic: mechanical fix
		UseCache:     false,
	})
	if err != nil {
		return "", fmt.Errorf("LLM call failed: %w", err)
	}

	// Strip any accidental markdown fences the LLM might add
	fixed := stripCodeFences(resp.Content)
	if strings.TrimSpace(fixed) == "" {
		return "", fmt.Errorf("LLM returned empty response")
	}
	return fixed, nil
}

// detectLanguage returns the language constant for a filename.
func detectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		return LangGo
	case ".ts", ".tsx":
		return LangTypeScript
	case ".js", ".jsx":
		return LangJavaScript
	default:
		return LangUnknown
	}
}

// stripCodeFences removes markdown code fences from LLM output.
// LLMs sometimes wrap code in ```go ... ``` even when told not to.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	// Remove opening fence (```go, ```typescript, ``` etc.)
	if strings.HasPrefix(s, "```") {
		newline := strings.Index(s, "\n")
		if newline != -1 {
			s = s[newline+1:]
		}
	}
	// Remove closing fence
	if strings.HasSuffix(s, "```") {
		s = s[:len(s)-3]
	}
	return strings.TrimSpace(s)
}
