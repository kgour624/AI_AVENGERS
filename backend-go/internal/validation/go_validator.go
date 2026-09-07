package validation

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// ValidationResult is the output of any language validator.
type ValidationResult struct {
	Passed     bool   // true if all stages passed
	Stage      string // which stage failed: syntax | vet | format | typecheck
	Error      string // human-readable error message
	Output     string // raw stdout+stderr from the failing tool
	DurationMs int64  // total validation time
}

// GoValidator validates Go source code.
// Stages: syntax (go/parser) → vet (go vet) → format (gofmt).
type GoValidator struct {
	cfg SandboxConfig
}

// NewGoValidator creates a new Go validator.
func NewGoValidator() *GoValidator {
	return &GoValidator{cfg: DefaultSandboxConfig()}
}

// Validate runs all Go validation stages on the provided source code.
//
// filename: the .go filename to use (e.g. "handler.go").
//   Used for go/parser error messages and gofmt invocation.
// code: the full Go source code string.
func (v *GoValidator) Validate(ctx context.Context, filename, code string) *ValidationResult {
	start := nowMs()

	// Stage 1: Syntax check via go/parser (native, no subprocess)
	// WHY first: fastest check, no I/O needed.
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, filename, code, parser.AllErrors)
	if err != nil {
		return &ValidationResult{
			Passed:     false,
			Stage:      "syntax",
			Error:      fmt.Sprintf("Go syntax error: %s", err.Error()),
			Output:     err.Error(),
			DurationMs: elapsedMs(start),
		}
	}

	// Write to temp dir for subprocess stages
	tmpDir, err := MakeTempDir()
	if err != nil {
		return &ValidationResult{
			Passed: false, Stage: "setup",
			Error:  fmt.Sprintf("failed to create temp dir: %s", err.Error()),
		}
	}
	defer os.RemoveAll(tmpDir)

	// Write a minimal go.mod so go vet works without a module context
	goMod := "module ai_avengers_validation\n\ngo 1.22\n"
	if _, err := WriteTempFile(tmpDir, "go.mod", goMod); err != nil {
		return &ValidationResult{
			Passed: false, Stage: "setup",
			Error:  fmt.Sprintf("failed to write go.mod: %s", err.Error()),
		}
	}

	if _, err := WriteTempFile(tmpDir, filename, code); err != nil {
		return &ValidationResult{
			Passed: false, Stage: "setup",
			Error:  fmt.Sprintf("failed to write source file: %s", err.Error()),
		}
	}

	// Stage 2: go vet
	vetResult, err := Run(ctx, v.cfg, tmpDir, []string{"go", "vet", "./..."}, nil)
	if err != nil {
		// System error (go not found, etc.) — treat as validation failure
		return &ValidationResult{
			Passed: false, Stage: "vet",
			Error:  fmt.Sprintf("go vet unavailable: %s", err.Error()),
			DurationMs: elapsedMs(start),
		}
	}
	if vetResult.TimedOut {
		return &ValidationResult{
			Passed: false, Stage: "vet",
			Error:  "go vet timed out",
			DurationMs: elapsedMs(start),
		}
	}
	if vetResult.ExitCode != 0 {
		return &ValidationResult{
			Passed:     false,
			Stage:      "vet",
			Error:      "go vet found issues",
			Output:     vetResult.Stderr,
			DurationMs: elapsedMs(start),
		}
	}

	// Stage 3: gofmt -l (list files that need formatting)
	// Exit code 0 always; non-empty stdout means file needs formatting.
	fmtResult, err := Run(ctx, v.cfg, tmpDir, []string{"gofmt", "-l", filename}, nil)
	if err != nil {
		// gofmt not found — skip format check, don't fail
		// WHY: format check is a quality gate, not a correctness gate.
		// If gofmt is missing (unusual), we still want vet to count.
		return &ValidationResult{
			Passed:     true,
			Stage:      "format_skipped",
			Error:      "gofmt not available",
			DurationMs: elapsedMs(start),
		}
	}
	if strings.TrimSpace(fmtResult.Stdout) != "" {
		return &ValidationResult{
			Passed:     false,
			Stage:      "format",
			Error:      fmt.Sprintf("file %s needs gofmt formatting", filename),
			Output:     fmtResult.Stdout,
			DurationMs: elapsedMs(start),
		}
	}

	return &ValidationResult{
		Passed:     true,
		DurationMs: elapsedMs(start),
	}
}
