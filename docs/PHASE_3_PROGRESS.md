# Phase 3 Progress Tracker

**Last Updated:** 2026-09-18  
**Current Status:** Task 2 Complete, Task 3 Ready

---

## Overview

Phase 3 adds code artifact publishing to the Aider integration. When Aider generates code files, they are published to the blackboard and projected into workflow_tasks for visibility in the Kanban UI.

---

## Task Breakdown

### Task 1: Implement `publishCodeArtifacts()` ✅ COMPLETE

**Goal:** Publish generated code from Aider workspaces to blackboard

**What Was Done:**
- ✅ Added `publishCodeArtifacts()` function in `aider_runner.go`
- ✅ Walks workspace recursively
- ✅ Skips .git/, design artifacts
- ✅ Detects language from file extension (18+ languages)
- ✅ Counts lines of code
- ✅ Posts `code_artifact_produced` events to blackboard
- ✅ Added `detectLanguage()` helper
- ✅ Added `countLines()` helper
- ✅ Integrated in `Run()` method (non-fatal errors)

**Files Modified:**
- `backend-go/internal/workflow/aider_runner.go` (+200 lines)

**Commits:**
- `feat(aider): implement publishCodeArtifacts()`
- `docs(aider): Phase 3 Task 1 Complete`

**Documentation:**
- `docs/PHASE_3_TASK_1_COMPLETE.md`

---

### Task 2: Update Projector ✅ COMPLETE

**Goal:** Handle `code_artifact_produced` events in Projector

**What Was Done:**
- ✅ Added dedicated handler for `code_artifact_produced` events
- ✅ INSERTs new workflow_tasks row for each code file
- ✅ Each file is a separate task visible in Kanban
- ✅ Status = 'done' (already created by Aider)
- ✅ Validates required fields (filename, expert_id)
- ✅ Fallback values for optional fields (language, file_path)
- ✅ Removed from default case (prevents double-handling)
- ✅ Comprehensive logging (debug, warn, info)

**Files Modified:**
- `backend-go/internal/workflow/projector.go` (+100 lines)

**Commits (Small, Frequent):**
1. `docs(phase3-task2): Step 1 - Analyzed Projector structure`
2. `feat(phase3-task2): Step 2 - Add dedicated handler`
3. `refactor(phase3-task2): Step 3 - Remove from default case`
4. `feat(phase3-task2): Step 4 - Add validation`
5. `feat(phase3-task2): Step 5 - Add comprehensive logging`
6. `docs(phase3-task2): Step 6 - Document code artifact projection`
7. `docs(phase3-task2): Step 7 - Phase 3 Task 2 Complete`

**Documentation:**
- `docs/PHASE_3_TASK_2_COMPLETE.md`
- `docs/PHASE_3_TASK_2_STEP_1.md`

---

### Task 3: Add Artifact Validation ⏳ READY TO START

**Goal:** Validate code before publishing

**What Needs to Be Done:**
- ❌ Run build before publishing
- ❌ Run tests before publishing
- ❌ Only publish if validation passes
- ❌ Log validation errors
- ❌ Add retry logic for transient failures

**Files to Modify:**
- `backend-go/internal/workflow/aider_runner.go`

**Approach:**
```go
publishCodeArtifacts() {
  // Step 1: Validate code
  if err := runBuild(workspacePath); err != nil {
    return fmt.Errorf("build failed: %w", err)
  }
  
  if err := runTests(workspacePath); err != nil {
    return fmt.Errorf("tests failed: %w", err)
  }
  
  // Step 2: Publish (only if validation passed)
  filepath.Walk(workspacePath, ...)
}
```

**Mental Model:**
- Cross-question: Should we publish if build fails?
  - Answer: No, broken code shouldn't be published
- Cross-question: Should we publish if tests fail?
  - Answer: Configurable (some projects have flaky tests)
- Cross-question: What if validation takes too long?
  - Answer: Add timeout (5 minutes for build, 10 minutes for tests)

---

## Architecture Overview

### Event Flow

```
Aider creates files
  ↓
publishCodeArtifacts() posts events
  ↓
Blackboard stores events
  ↓
Projector receives events
  ↓
code_artifact_produced handler
  ↓
INSERT workflow_tasks rows
  ↓
Kanban UI shows tasks
```

### Data Flow

```
Workspace:
  - auth.go (150 lines, go)
  - auth_test.go (80 lines, go)
  - middleware.go (120 lines, go)

Blackboard Events:
  1. {type: "code_artifact_produced", data: {filename: "auth.go", ...}}
  2. {type: "code_artifact_produced", data: {filename: "auth_test.go", ...}}
  3. {type: "code_artifact_produced", data: {filename: "middleware.go", ...}}

Workflow Tasks:
  1. {title: "Generated: auth.go", status: "done", artifact_type: "code_file"}
  2. {title: "Generated: auth_test.go", status: "done", artifact_type: "code_file"}
  3. {title: "Generated: middleware.go", status: "done", artifact_type: "code_file"}

Kanban UI:
  Done Column:
    - Generated: auth.go
    - Generated: auth_test.go
    - Generated: middleware.go
```

---

## Testing Status

### Manual Testing

- ❌ Task 1: Not tested yet (requires Aider CLI setup)
- ❌ Task 2: Not tested yet (requires Task 1 working)
- ❌ Task 3: Not started

### Unit Testing

- ❌ Task 1: No unit tests yet
- ❌ Task 2: No unit tests yet
- ❌ Task 3: Not started

### Integration Testing

- ❌ End-to-end flow: Not tested yet

**TODO:** Add comprehensive test suite after Task 3 complete

---

## Metrics

### Code Changes

- **Task 1:** +200 lines (aider_runner.go)
- **Task 2:** +100 lines (projector.go)
- **Task 3:** TBD
- **Total:** ~300 lines added

### Commits

- **Task 1:** 3 commits
- **Task 2:** 7 commits (small, frequent)
- **Task 3:** TBD
- **Total:** 10 commits so far

### Documentation

- **Task 1:** 1 doc (PHASE_3_TASK_1_COMPLETE.md)
- **Task 2:** 2 docs (PHASE_3_TASK_2_COMPLETE.md, PHASE_3_TASK_2_STEP_1.md)
- **Task 3:** TBD
- **Total:** 3 docs + this progress tracker

---

## Next Steps

1. **Start Task 3:** Add artifact validation
   - Implement build validation
   - Implement test validation
   - Add timeout handling
   - Add retry logic

2. **Testing:** Add comprehensive test suite
   - Unit tests for publishCodeArtifacts()
   - Unit tests for Projector handler
   - Integration test for end-to-end flow

3. **Documentation:** Update architecture docs
   - Update AIDER_INTEGRATION_ARCHITECTURE.md
   - Add troubleshooting guide
   - Add deployment guide

---

## References

- **AIDER_INTEGRATION_ARCHITECTURE.md** - Overall design
- **PHASE_1_COMPLETE.md** - Foundation
- **PHASE_2_COMPLETE.md** - OTA loop
- **PHASE_3_TASK_1_COMPLETE.md** - Code publishing
- **PHASE_3_TASK_2_COMPLETE.md** - Projector updates
- **INGESTION_HANG_BUG_FIX.md** - Bug fix analysis

---

## Status Summary

**Phase 1**: ✅ Complete (Foundation)  
**Phase 2**: ✅ Complete (OTA loop)  
**Phase 3 Task 1**: ✅ Complete (Code publishing)  
**Phase 3 Task 2**: ✅ Complete (Projector updates)  
**Phase 3 Task 3**: ⏳ Ready to start (Artifact validation)
