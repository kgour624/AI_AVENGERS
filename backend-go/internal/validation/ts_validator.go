package validation

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// TSValidator validates TypeScript and JavaScript source code.
// Stages: typecheck (tsc) -> format (prettier).
type TSValidator struct {
	cfg SandboxConfig
}

// NewTSValidator creates a new TypeScript/JavaScript validator.
func NewTSValidator() *TSValidator {
	return &TSValidator{cfg: DefaultSandboxConfig()}
}

// Validate runs all TS/JS validation stages on the provided source code.
func (v *TSValidator) Validate(ctx context.Context, filename, code string) *ValidationResult {
	start := nowMs()

	tmpDir, err := MakeTempDir()
	if err != nil {
		return &ValidationResult{Passed: false, Stage: "setup",
			Error: fmt.Sprintf("failed to create temp dir: %s", err.Error())}
	}
	defer os.RemoveAll(tmpDir)

	if _, err := WriteTempFile(tmpDir, filename, code); err != nil {
		return &ValidationResult{Passed: false, Stage: "setup",
			Error: fmt.Sprintf("failed to write source file: %s", err.Error())}
	}

	tsconfig := "{\"compilerOptions\":{\"strict\":true,\"noEmit\":true,\"allowJs\":true,\"checkJs\":true,\"target\":\"ES2020\",\"module\":\"commonjs\"},\"include\":[\"" + filename + "\"]}"
	if _, err := WriteTempFile(tmpDir, "tsconfig.json", tsconfig); err != nil {
		return &ValidationResult{Passed: false, Stage: "setup",
			Error: fmt.Sprintf("failed to write tsconfig: %s", err.Error())}
	}

	// Stage 1: tsc --noEmit
	tscResult, err := Run(ctx, v.cfg, tmpDir,
		[]string{"tsc", "--noEmit", "--project", "tsconfig.json"}, nil)
	if err != nil {
		return &ValidationResult{Passed: false, Stage: "typecheck",
			Error: fmt.Sprintf("tsc not available: %s", err.Error()),
			DurationMs: elapsedMs(start)}
	}
	if tscResult.TimedOut {
		return &ValidationResult{Passed: false, Stage: "typecheck",
			Error: "tsc timed out", DurationMs: elapsedMs(start)}
	}
	if tscResult.ExitCode != 0 {
		return &ValidationResult{Passed: false, Stage: "typecheck",
			Error:      "TypeScript type errors found",
			Output:     tscResult.Stdout + tscResult.Stderr,
			DurationMs: elapsedMs(start)}
	}

	// Stage 2: prettier --check
	prettierResult, err := Run(ctx, v.cfg, tmpDir,
		[]string{"prettier", "--check", filename}, nil)
	if err != nil {
		// prettier not found - skip format check, don't fail
		return &ValidationResult{Passed: true, Stage: "format_skipped",
			Error: "prettier not available", DurationMs: elapsedMs(start)}
	}
	if prettierResult.ExitCode != 0 {
		output := prettierResult.Stdout + prettierResult.Stderr
		if strings.Contains(output, "Code style issues") || prettierResult.ExitCode == 1 {
			return &ValidationResult{Passed: false, Stage: "format",
				Error:      fmt.Sprintf("file %s needs prettier formatting", filename),
				Output:     output,
				DurationMs: elapsedMs(start)}
		}
	}

	return &ValidationResult{Passed: true, DurationMs: elapsedMs(start)}
}
