# Phase 3 Task 2 - Step 1: Projector Analysis

**Date:** 2026-09-18  
**Status:** Analysis Complete  
**Next:** Step 2 - Add dedicated handler

---

## Findings

### Current Projector Structure

**File:** `backend-go/internal/workflow/projector.go` (236 lines)

**Key Functions:**
- `Run()` - Main projection loop, subscribes to blackboard events
- `project()` - Routes events to handlers (line 93-199)
- `PostTaskStatus()` - Helper to post task status changes
- `PostTaskFailed()` - Helper to post task failures

### Current Event Handlers

1. **`task_plan_ready`** (line 100-130)
   - Parses plan with tasks array
   - INSERTs workflow_tasks rows (one per task)
   - Status = 'todo'

2. **`task_status_changed`** (line 132-152)
   - Parses {expert_id, status}
   - UPDATEs workflow_tasks.status
   - Sets started_at, completed_at timestamps

3. **`task_failed`** (line 154-170)
   - Parses {expert_id, reason}
   - UPDATEs workflow_tasks.status = 'failed'

4. **`default` (artifact events)** (line 172-199)
   - Handles: architecture_decision, data_model_proposed, api_contract_proposed, module_design_proposed, **code_artifact_produced**, test_case_proposed, requirement_captured
   - UPDATEs workflow_tasks.produced_artifact_event_id

---

## Issue Identified

### Problem: Multiple Code Files Overwrite Same Row

**Current Behavior:**
```go
// Event 1: auth.go posted
UPDATE workflow_tasks 
  SET produced_artifact_event_id = 123 
  WHERE workflow_id = X AND assigned_expert_id = backend_expert

// Event 2: auth_test.go posted
UPDATE workflow_tasks 
  SET produced_artifact_event_id = 124  // OVERWRITES 123!
  WHERE workflow_id = X AND assigned_expert_id = backend_expert

// Event 3: middleware.go posted
UPDATE workflow_tasks 
  SET produced_artifact_event_id = 125  // OVERWRITES 124!
  WHERE workflow_id = X AND assigned_expert_id = backend_expert

// Result: Only middleware.go is tracked ❌
```

### Mental Model Verification

**Scenario:** Aider creates 3 files
- `publishCodeArtifacts()` posts 3 `code_artifact_produced` events
- Projector receives 3 events
- Current handler UPDATEs same workflow_tasks row 3 times
- Only last event ID is stored
- **Expected:** Track all 3 files ✅
- **Actual:** Track only last file ❌

### Cross-Questions

**Q: Should we track all code files or just the latest?**  
A: Track ALL files - each file is a separate artifact that should be visible in Kanban

**Q: Should we UPDATE existing row or INSERT new rows?**  
A: INSERT new rows - each code file is a separate task/artifact

**Q: What should the task title be?**  
A: "Generated: {filename}" (e.g., "Generated: auth.go")

**Q: What should the status be?**  
A: 'done' - file already created by Aider

**Q: Should we store file metadata?**  
A: Yes - store in artifact_data JSON column (language, lines_of_code, file_path)

---

## Solution Design

### Approach: Dedicated Handler for code_artifact_produced

**Instead of:**
```go
default:
    // Handles all artifact types the same way
    artifactTypes := map[string]bool{
        "code_artifact_produced": true,
        ...
    }
    UPDATE workflow_tasks SET produced_artifact_event_id = event.ID
```

**Do this:**
```go
case "code_artifact_produced":
    // Dedicated handler for code artifacts
    // Parse event data: {filename, file_path, content, language, lines_of_code}
    // INSERT new workflow_tasks row for this file
    // Status = 'done'
    // Store metadata in artifact_data column

default:
    // Other artifact types (architecture_decision, etc.)
    // Continue using UPDATE approach
```

### Implementation Plan

**Step 2:** Add `case "code_artifact_produced":` handler
- Parse event.Data (filename, file_path, language, lines_of_code)
- INSERT workflow_tasks row
- Title = "Generated: {filename}"
- Status = 'done'
- artifact_type = 'code_file'
- artifact_data = JSON with metadata
- produced_artifact_event_id = event.ID

**Step 3:** Test the handler
- Create workflow with implementation phase
- Aider generates 3 files
- Verify 3 workflow_tasks rows created
- Verify each row has correct metadata

---

## Database Schema Check

**Table:** `workflow_tasks`

**Relevant Columns:**
- `id` - UUID primary key
- `workflow_id` - UUID (which workflow)
- `assigned_expert_id` - UUID (which expert created it)
- `title` - TEXT (task title)
- `description` - TEXT (task description)
- `status` - TEXT (todo, in_progress, done, failed)
- `artifact_type` - TEXT (code_file, architecture, etc.)
- `artifact_data` - JSONB (metadata)
- `produced_artifact_event_id` - UUID (link to blackboard event)
- `started_at` - TIMESTAMP
- `completed_at` - TIMESTAMP
- `created_at` - TIMESTAMP
- `updated_at` - TIMESTAMP

**Perfect for our use case!** ✅

---

## Next Steps

**Step 2:** Add dedicated handler for `code_artifact_produced`
- Location: `backend-go/internal/workflow/projector.go`
- Add before `default:` case (around line 172)
- Parse event data
- INSERT workflow_tasks row

**Step 3:** Test the implementation
- Manual test with real workflow
- Verify multiple files create multiple rows
- Verify metadata stored correctly

---

## References

- **AIDER_INTEGRATION_ARCHITECTURE.md** §6.3 - Code Publishing
- **Projector Design** - "Single Write Path" pattern
- **Blackboard Pattern** - Events are source of truth
