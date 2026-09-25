package workflow

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"ai_avengers/backend/internal/validation"
)

// Toolchain-aware workspace verification.
//
// This file exists because of one wrong assumption in the original design: that
// the api container can compile and test the code Aider writes. At the time it
// could not — the runtime image was alpine plus ca-certificates, tzdata, git and
// rsync, with no Go toolchain and no node. So exec.Command("go", "build")
// returned "executable file not found", and because the old code could not tell
// that apart from a compile error, three things were permanently broken:
//
//   - observeWorkspace fed the model "Build Errors:" with an EMPTY body on every
//     iteration. exec's not-found error produces no output, so the model was
//     handed a heading with nothing under it — noise where the signal should be.
//   - taskComplete was (buildErr == nil && testErr == nil), which could never be
//     true, so every task burned all five iterations and always reported
//     Completed=false.
//   - publishCodeArtifacts returns early when the build fails, so no
//     code_artifact_produced event ever reached the blackboard — the Deliverables
//     panel could not show generated code even when Aider wrote working code.
//
// A missing tool and broken code are different facts. VerifyStatus keeps them
// apart, and every caller now treats them differently: broken code blocks,
// a missing tool does not — but it never counts as success either.
//
// Verification is also per-project, not Go-only. Running "go build ./..." in a
// React expert's workspace was guaranteed to fail and told a frontend expert to
// go fix a Go build error.
//
// STATUS OF THE TOOLCHAIN TODAY: commit 31becac added `go nodejs npm` to the
// runtime image, so these checks can run. Node dependencies are installed on
// demand with lifecycle scripts disabled and a timeout; network/install failure
// remains "unavailable", not a claim that the code itself is broken.

// VerifyStatus is the outcome of checking one project in a workspace.
type VerifyStatus string

const (
	// VerifyPassed means the project's build and tests ran and succeeded.
	VerifyPassed VerifyStatus = "passed"
	// VerifyFailed means the project's own code is broken. Actionable.
	VerifyFailed VerifyStatus = "failed"
	// VerifyUnavailable means the check could not run at all — missing
	// toolchain, missing dependencies, timeout. Says nothing about the code.
	VerifyUnavailable VerifyStatus = "unavailable"
)

// VerifyResult is the outcome for one project inside a workspace.
type VerifyResult struct {
	// Project is the ecosystem checked: "go" or "node".
	Project string
	Status  VerifyStatus
	// Reason explains an unavailable result. Empty otherwise.
	Reason string
	// Output is the failing command and its output. Empty unless failed.
	Output string
}

// WorkspaceVerification is every project found in one workspace.
type WorkspaceVerification struct {
	Results []VerifyResult
}

// Verified reports whether at least one project was actually checked.
func (v WorkspaceVerification) Verified() bool {
	for _, r := range v.Results {
		if r.Status == VerifyPassed || r.Status == VerifyFailed {
			return true
		}
	}
	return false
}

// AllPassed reports whether every project that could be checked passed, and at
// least one was checked.
//
// "Nothing could be checked" is deliberately not success. Returning true there
// would mark a task complete because the workspace was unverifiable — the exact
// false green this file exists to prevent.
func (v WorkspaceVerification) AllPassed() bool {
	checked := false
	for _, r := range v.Results {
		switch r.Status {
		case VerifyFailed:
			return false
		case VerifyPassed:
			checked = true
		}
	}
	return checked
}

// Failures returns the projects whose code is broken.
func (v WorkspaceVerification) Failures() []VerifyResult {
	var out []VerifyResult
	for _, r := range v.Results {
		if r.Status == VerifyFailed {
			out = append(out, r)
		}
	}
	return out
}

// Report renders the verification for the model.
//
// The unavailable case says outright not to treat it as a code problem. Without
// that line the model spends its next iteration hunting for a compile error
// that does not exist.
func (v WorkspaceVerification) Report() string {
	if len(v.Results) == 0 {
		return "BUILD CHECK: no project manifest found in this workspace yet, so nothing could be " +
			"built. Create the manifest alongside the source files — go.mod for Go, " +
			"package.json for a Node/React project — so the build can be verified.\n"
	}

	var b strings.Builder
	b.WriteString("BUILD CHECK:\n")
	for _, r := range v.Results {
		switch r.Status {
		case VerifyPassed:
			fmt.Fprintf(&b, "  %s: build and tests passed.\n", r.Project)
		case VerifyFailed:
			fmt.Fprintf(&b, "  %s: FAILED — fix this before anything else.\n%s\n",
				r.Project, indentLines(r.Output, "    "))
		case VerifyUnavailable:
			fmt.Fprintf(&b, "  %s: could not be verified (%s). This is an environment "+
				"limitation, NOT a problem in your code — do not try to fix it.\n",
				r.Project, r.Reason)
		}
	}
	return b.String()
}

func indentLines(s, prefix string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}

// verifyWorkspace checks every project it can find in the workspace.
//
// Detection is by manifest on disk, not by the expert's declared domain: the
// workspace is the ground truth, later waves receive merged code from other
// experts, and a manifest that appears in iteration 2 is picked up in iteration
// 3 with no extra wiring.
func (a *AiderRunner) verifyWorkspace(ctx context.Context, workspacePath string) WorkspaceVerification {
	var v WorkspaceVerification

	// --- Go ---
	switch {
	case hasWorkspaceFile(workspacePath, "go.mod"):
		v.Results = append(v.Results, runProjectChecks(ctx, workspacePath, "go", [][]string{
			{"go", "build", "./..."},
			{"go", "test", "./..."},
		}))
	case hasSourceWithExt(workspacePath, ".go"):
		// A real, actionable defect rather than a missing tool: the sources
		// cannot be compiled by anyone in this state.
		v.Results = append(v.Results, VerifyResult{
			Project: "go",
			Status:  VerifyFailed,
			Output:  "Go source files exist but there is no go.mod, so the package cannot be built. Add go.mod with a module path and the Go version.",
		})
	}

	// --- Node / React ---
	switch {
	case hasWorkspaceFile(workspacePath, "package.json"):
		// A11: dependencies are installed first so the build can actually
		// run. The installer checks the manifest fingerprint, so existing
		// node_modules is reused only when package.json/lockfiles are unchanged.
		// Install failure stays "unavailable", never pass or code failure.
		res, ok := a.installNodeDependencies(ctx, workspacePath)
		if !ok {
			v.Results = append(v.Results, res)
			break
		}
		// --if-present: a project with no build or test script is not a failure.
		v.Results = append(v.Results, runProjectChecks(ctx, workspacePath, "node", [][]string{
			{"npm", "run", "build", "--if-present"},
			{"npm", "run", "test", "--if-present"},
		}))
	case hasSourceWithExt(workspacePath, ".ts", ".tsx", ".js", ".jsx"):
		v.Results = append(v.Results, VerifyResult{
			Project: "node",
			Status:  VerifyFailed,
			Output:  "JavaScript/TypeScript source files exist but there is no package.json, so nothing can be built or run. Add package.json with the dependencies and a build script.",
		})
	}

	return v
}

// nodeInstallTimeout bounds a dependency install. Fetching arbitrary packages
// over the network must never hang a workflow.
const nodeInstallTimeout = 5 * time.Minute
const nodeInstallFingerprintFile = ".ai-avengers-deps.sha256"

// installNodeDependencies installs a Node project's dependencies so its build
// and tests can actually run.
//
// WHY --ignore-scripts: `npm install` runs lifecycle scripts (preinstall,
// install, postinstall) declared by every dependency by default, and since the
// §17 ingest that package.json can belong to a CLIENT's cloned repository.
// Ignoring scripts removes that remote-code-execution path while still
// resolving the module graph the build needs. What remains is the exposure
// runProjectChecks already accepts: `npm run build` executes the project's own
// scripts with validation.MinimalEnv, i.e. no platform secrets.
//
// ok is false when the install could not complete; the caller must report that
// as unavailable, never as a pass and never as a code failure.
func (a *AiderRunner) installNodeDependencies(ctx context.Context, workspacePath string) (VerifyResult, bool) {
	if _, err := exec.LookPath("npm"); err != nil {
		return VerifyResult{
			Project: "node",
			Status:  VerifyUnavailable,
			Reason:  `"npm" is not installed in this image`,
		}, false
	}
	fingerprint, err := nodeDependencyFingerprint(workspacePath)
	if err != nil {
		return VerifyResult{Project: "node", Status: VerifyUnavailable, Reason: "could not fingerprint package manifests: " + err.Error()}, false
	}
	if hasWorkspaceDir(workspacePath, "node_modules") && nodeInstallFingerprintMatches(workspacePath, fingerprint) {
		return VerifyResult{}, true
	}

	// `npm ci` is exact and lockfile-driven; without a lockfile, `npm install`
	// is the only option. Both skip lifecycle scripts.
	args := []string{"install", "--ignore-scripts", "--no-audit", "--no-fund", "--package-lock=false"}
	if hasWorkspaceFile(workspacePath, "package-lock.json") || hasWorkspaceFile(workspacePath, "npm-shrinkwrap.json") {
		args = []string{"ci", "--ignore-scripts", "--no-audit", "--no-fund"}
	}

	installCtx, cancel := context.WithTimeout(ctx, nodeInstallTimeout)
	defer cancel()

	cmd := exec.CommandContext(installCtx, "npm", args...)
	cmd.Dir = workspacePath
	cmd.Env = validation.MinimalEnv()
	out, err := cmd.CombinedOutput()
	if err == nil {
		// Keep the marker inside node_modules: rsync excludes that directory,
		// so this internal cache metadata can never leak into main/ or a client
		// deliverable.
		if err := os.WriteFile(filepath.Join(workspacePath, "node_modules", nodeInstallFingerprintFile), []byte(fingerprint), 0600); err != nil {
			return VerifyResult{Project: "node", Status: VerifyUnavailable, Reason: "dependencies installed but install fingerprint could not be saved: " + err.Error()}, false
		}
		return VerifyResult{}, true
	}
	if installCtx.Err() != nil {
		return VerifyResult{
			Project: "node",
			Status:  VerifyUnavailable,
			Reason:  fmt.Sprintf("dependency install did not finish within %s", nodeInstallTimeout),
		}, false
	}
	return VerifyResult{
		Project: "node",
		Status:  VerifyUnavailable,
		Reason:  "dependency install failed: " + strings.TrimSpace(string(out)),
	}, false
}

// nodeDependencyFingerprint hashes package.json and available lockfiles so
// node_modules is reused only while the declared dependency set is unchanged.
func nodeDependencyFingerprint(workspacePath string) (string, error) {
	h := sha256.New()
	for _, name := range []string{"package.json", "package-lock.json", "npm-shrinkwrap.json"} {
		content, err := os.ReadFile(filepath.Join(workspacePath, name))
		if err != nil {
			if os.IsNotExist(err) && name != "package.json" {
				continue
			}
			return "", err
		}
		_, _ = h.Write([]byte(name))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write(content)
		_, _ = h.Write([]byte{0})
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func nodeInstallFingerprintMatches(workspacePath, expected string) bool {
	content, err := os.ReadFile(filepath.Join(workspacePath, "node_modules", nodeInstallFingerprintFile))
	return err == nil && string(content) == expected
}

// runProjectChecks runs commands in order and stops at the first problem.
//
// A missing executable is reported as unavailable, never as a code failure.
// That single distinction is what this file is for.
//
// The environment is not inherited, and that is the point:
//
// cmd.Env used to be left nil, which means "inherit the parent's environment".
// The parent is the api process, whose environment holds ENCRYPTION_KEY — the
// key that decrypts clients' git tokens — plus JWT_SECRET, every provider API
// key, AIDER_PROXY_TOKEN, both OAuth client secrets, and DATABASE_URL with its
// password (docker-compose.yml).
//
// What runs here is not our code. `go test ./...` runs tests a model generated.
// `npm run build` runs whatever scripts a package.json declares, including its
// dependencies' — and since the code-feedback ingest (§17) that package.json can
// belong to a CLIENT's cloned repository. One line of `process.env` in any of
// that would have read every credential the platform holds.
//
// validation.MinimalEnv is the allowlist, shared with the validation pipeline so
// there is one answer to "what may a subprocess see" rather than two that drift.
// Note what it does NOT do: there is no network restriction and no memory limit
// (see sandbox_linux.go for why the latter is a documented no-op). The wall-clock
// bound is the caller's context, and for the §17 path that is a per-command
// timeout.
func runProjectChecks(ctx context.Context, dir, project string, cmds [][]string) VerifyResult {
	env := validation.MinimalEnv()

	for _, argv := range cmds {
		if _, err := exec.LookPath(argv[0]); err != nil {
			return VerifyResult{
				Project: project,
				Status:  VerifyUnavailable,
				Reason:  fmt.Sprintf("%q is not installed in this image", argv[0]),
			}
		}

		cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
		cmd.Dir = dir
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err == nil {
			continue
		}

		if ctx.Err() != nil {
			return VerifyResult{
				Project: project,
				Status:  VerifyUnavailable,
				Reason:  fmt.Sprintf("%q did not finish in time", strings.Join(argv, " ")),
			}
		}
		return VerifyResult{
			Project: project,
			Status:  VerifyFailed,
			Output:  fmt.Sprintf("$ %s\n%s", strings.Join(argv, " "), strings.TrimSpace(string(out))),
		}
	}
	return VerifyResult{Project: project, Status: VerifyPassed}
}

// goCoverage returns the Go statement coverage of the workspace.
//
// ok is false when coverage could not be measured at all — no go.mod, no Go
// toolchain, or the test run could not start. Callers must not read 0.0 as
// "zero coverage" in that case; it means "unknown".
func (a *AiderRunner) goCoverage(ctx context.Context, workspacePath string) (coverage float64, ok bool) {
	if !hasWorkspaceFile(workspacePath, "go.mod") {
		return 0, false
	}
	if _, err := exec.LookPath("go"); err != nil {
		return 0, false
	}
	output, cov, err := a.runTestsWithCoverage(ctx, workspacePath)
	if err != nil && strings.TrimSpace(output) == "" {
		// No output at all means the command never really ran.
		return 0, false
	}
	return cov, true
}

// hasWorkspaceFile reports whether name is a file at the workspace root.
func hasWorkspaceFile(workspacePath, name string) bool {
	st, err := os.Stat(filepath.Join(workspacePath, name))
	return err == nil && !st.IsDir()
}

// hasWorkspaceDir reports whether name is a directory at the workspace root.
func hasWorkspaceDir(workspacePath, name string) bool {
	st, err := os.Stat(filepath.Join(workspacePath, name))
	return err == nil && st.IsDir()
}

// hasSourceWithExt reports whether the workspace holds at least one file with
// one of the given extensions.
//
// .git is skipped for the obvious reason. design/ is skipped because it holds
// the seeded design documents, which are reference material and must never make
// the workspace look like a project of that language.
func hasSourceWithExt(workspacePath string, exts ...string) bool {
	found := false
	_ = filepath.Walk(workspacePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // unreadable entry: ignore, keep walking
		}
		if info.IsDir() {
			switch info.Name() {
			case ".git", "design", "node_modules", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		for _, want := range exts {
			if ext == want {
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found
}
