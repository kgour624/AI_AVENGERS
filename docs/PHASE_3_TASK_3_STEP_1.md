# Phase 3 Task 3 - Step 1: Build/Test Analysis

**Date:** 2026-09-18  
**Status:** Analysis Complete  
**Next:** Step 2 - Add validation logic

---

## Findings

### Existing Functions

#### `runBuild()` (line 470-485)

```go
func (a *AiderRunner) runBuild(
    ctx context.Context,
    workspacePath string,
) (string, error) {
    cmd := exec.CommandContext(ctx, "go", "build", "./...")
    cmd.Dir = workspacePath
    output, err := cmd.CombinedOutput()
    return string(output), err
}
```

**Behavior:**
- Executes: `go build ./...`
- Compiles all packages in workspace
- Returns: (output, error)
  - `error != nil` → build failed (compile errors)
  - `error == nil` → build succeeded

#### `runTests()` (line 487-510)

```go
func (a *AiderRunner) runTests(
    ctx context.Context,
    workspacePath string,
) (string, error) {
    cmd := exec.CommandContext(ctx, "go", "test", "./...")
    cmd.Dir = workspacePath
    output, err := cmd.CombinedOutput()
    return string(output), err
}
```

**Behavior:**
- Executes: `go test ./...`
- Runs all tests in workspace
- Returns: (output, error)
  - `error != nil` → tests failed
  - `error == nil` → tests passed

### Current Usage

#### In `runAiderIteration()` (line 590-592)

```go
// Check if task complete
// Heuristic: build + tests pass
_, buildErr := a.runBuild(ctx, workspacePath)
_, testErr := a.runTests(ctx, workspacePath)
taskComplete = (buildErr == nil && testErr == nil)
```

**Purpose:**
- Determine if Aider's iteration succeeded
- If both pass → task complete, exit loop
- If either fails → continue iterating (max 5 iterations)

#### In `Run()` method (line 143)

```go
// Publish code artifacts to blackboard
if err := a.publishCodeArtifacts(ctx, req, workspacePath, commitSHAs); err != nil {
    a.logger.Error("failed to publish code artifacts",
        zap.Error(err),
        zap.String("workflow_id", req.WorkflowID.String()),
    )
    // Non-fatal: code already committed to git
}
```

**Current Behavior:**
- Called AFTER Aider loop completes
- Called regardless of build/test status
- No validation before publishing

---

## Issue Identified

### Problem: Broken Code Gets Published

**Scenario:**

```
1. Aider iteration 1: Creates code → build fails
2. Aider iteration 2: Fixes code → tests fail
3. Aider iteration 3: Fixes tests → build fails again
4. Aider iteration 4: Fixes build → tests fail again
5. Aider iteration 5: Fixes tests → build still fails
6. Max iterations reached (5) → loop exits
7. publishCodeArtifacts() called → broken code published ❌
```

**Result:**
- Broken code in blackboard
- Kanban shows "Generated: broken.go" as done
- Frontend may try to use broken code
- Downstream experts see broken artifacts

### Mental Model Verification

**Scenario 1: Valid Code**
```
Aider creates valid code:
  - auth.go (compiles, tests pass)
  - auth_test.go (tests pass)

runAiderIteration():
  - runBuild() → error == nil ✅
  - runTests() → error == nil ✅
  - taskComplete = true ✅

publishCodeArtifacts():
  - Files published to blackboard ✅
  - Kanban shows tasks as done ✅
```

**Scenario 2: Broken Code**
```
Aider creates broken code:
  - auth.go (compile error: undefined variable)
  - auth_test.go (tests can't run, build fails)

runAiderIteration() x5:
  - Iteration 1: runBuild() → error != nil ❌
  - Iteration 2: runBuild() → error != nil ❌
  - Iteration 3: runBuild() → error != nil ❌
  - Iteration 4: runBuild() → error != nil ❌
  - Iteration 5: runBuild() → error != nil ❌
  - Max iterations reached

publishCodeArtifacts():
  - Files published to blackboard ❌ (SHOULD NOT PUBLISH)
  - Kanban shows tasks as done ❌ (MISLEADING)
```

### Cross-Questions

**Q: Should we publish broken code?**  
A: No, only publish if validation passes. Broken code is not a deliverable.

**Q: What if code never passes validation?**  
A: Don't publish, log error, mark task as failed. Expert needs to investigate.

**Q: Should validation be optional?**  
A: Yes, configurable per workflow. Some projects have flaky tests or no tests.

**Q: What if build passes but tests fail?**  
A: Configurable. Some teams accept code with failing tests (WIP), others don't.

**Q: What if validation takes too long?**  
A: Add timeout (5 minutes for build, 10 minutes for tests). Fail if timeout.

**Q: Should we retry validation?**  
A: No, validation already ran in Aider loop. If it failed 5 times, retrying won't help.

---

## Solution Design

### Approach: Add Validation to `publishCodeArtifacts()`

**Current Flow:**
```go
publishCodeArtifacts(ctx, req, workspacePath, commitSHAs) {
  // Walk workspace
  // Post events to blackboard
}
```

**New Flow:**
```go
publishCodeArtifacts(ctx, req, workspacePath, commitSHAs) {
  // Step 1: Validate code
  buildOutput, buildErr := a.runBuild(ctx, workspacePath)
  if buildErr != nil {
    return fmt.Errorf("build validation failed: %w\nOutput: %s", buildErr, buildOutput)
  }
  
  testOutput, testErr := a.runTests(ctx, workspacePath)
  if testErr != nil {
    return fmt.Errorf("test validation failed: %w\nOutput: %s", testErr, testOutput)
  }
  
  // Step 2: Publish (only if validation passed)
  // Walk workspace
  // Post events to blackboard
}
```

### Configuration (Future Enhancement)

**Workflow-level config:**
```go
type ValidationConfig struct {
  RequireBuildPass bool // Default: true
  RequireTestsPass bool // Default: true
  BuildTimeout     time.Duration // Default: 5 minutes
  TestTimeout      time.Duration // Default: 10 minutes
}
```

**Usage:**
```go
if config.RequireBuildPass {
  if buildErr != nil {
    return fmt.Errorf("build validation failed")
  }
}

if config.RequireTestsPass {
  if testErr != nil {
    return fmt.Errorf("test validation failed")
  }
}
```

---

## Implementation Plan

### Step 2: Add Validation Logic

**Location:** `backend-go/internal/workflow/aider_runner.go`  
**Function:** `publishCodeArtifacts()` (line 665-770)

**Changes:**
1. Add validation at start of function (before walking workspace)
2. Call `a.runBuild(ctx, workspacePath)`
3. If build fails, return error with output
4. Call `a.runTests(ctx, workspacePath)`
5. If tests fail, return error with output
6. Continue with existing walk logic

**Mental Model:**
```go
publishCodeArtifacts() {
  // CROSS-QUESTION: Should we validate?
  // ANSWER: Yes, only publish valid code
  
  // Validate build
  buildOutput, buildErr := a.runBuild(ctx, workspacePath)
  if buildErr != nil {
    // CROSS-QUESTION: Should we log output?
    // ANSWER: Yes, helps debug why validation failed
    a.logger.Error("build validation failed",
      zap.String("output", buildOutput),
      zap.Error(buildErr),
    )
    return fmt.Errorf("build validation failed: %w", buildErr)
  }
  
  // Validate tests
  testOutput, testErr := a.runTests(ctx, workspacePath)
  if testErr != nil {
    a.logger.Error("test validation failed",
      zap.String("output", testOutput),
      zap.Error(testErr),
    )
    return fmt.Errorf("test validation failed: %w", testErr)
  }
  
  // CROSS-QUESTION: Should we log success?
  // ANSWER: Yes, confirms validation passed
  a.logger.Info("validation passed",
    zap.String("workflow_id", req.WorkflowID.String()),
  )
  
  // Continue with existing walk logic...
}
```

### Step 3: Add Timeout Handling

**Problem:** Build/tests may hang forever

**Solution:** Add context timeout

```go
// Create timeout context for validation
validationCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
defer cancel()

// Run build with timeout
buildOutput, buildErr := a.runBuild(validationCtx, workspacePath)
if buildErr != nil {
  if validationCtx.Err() == context.DeadlineExceeded {
    return fmt.Errorf("build validation timeout (15 minutes)")
  }
  return fmt.Errorf("build validation failed: %w", buildErr)
}

// Run tests with timeout
testOutput, testErr := a.runTests(validationCtx, workspacePath)
if testErr != nil {
  if validationCtx.Err() == context.DeadlineExceeded {
    return fmt.Errorf("test validation timeout (15 minutes)")
  }
  return fmt.Errorf("test validation failed: %w", testErr)
}
```

---

## Next Steps

**Step 2:** Add validation logic to `publishCodeArtifacts()`
- Add build validation
- Add test validation
- Return error if validation fails

**Step 3:** Add timeout handling
- Create timeout context (15 minutes)
- Check for timeout errors
- Log timeout separately from validation failure

**Step 4:** Add detailed error logging
- Log build output on failure
- Log test output on failure
- Log validation success

**Step 5:** Add metrics tracking
- Track validation success rate
- Track validation duration
- Track timeout frequency

---

## References

- **AIDER_INTEGRATION_ARCHITECTURE.md** §6.3 - Code Publishing
- **Phase 2 Complete** - OTA loop with build/test checks
- **System Design Master Class** - "Validate before commit" pattern
