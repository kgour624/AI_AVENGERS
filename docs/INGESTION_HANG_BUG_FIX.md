# Ingestion Job Hang Bug - Root Cause Analysis & Fix

**Date:** 2026-09-17  
**Status:** Fixed  
**Severity:** Critical (jobs hang forever, no recovery)

---

## Problem Statement

Ingestion job stuck in 'running' state for 30+ minutes with no progress:
- Job ID: visible in DB with `status='running'`, `started_at=2026-09-17 19:30:55`
- Stage: `charter_extraction` (CodeCraft LLM call)
- Expected: Job should pause after 3 LLM retries (max 6 minutes)
- Actual: Job hangs forever, never reaches 'paused' state
- No error message, no way to retry, no way to recover

---

## Root Cause Analysis

### Initial Hypothesis (User's Analysis)

**Correct:**
1. ✅ CodeCraftAPI issue started the problem
2. ✅ Job is NOT paused, it's **stuck/hanging**
3. ✅ `context.Background()` has no timeout → operations can hang forever

**Root Cause Chain:**

```
1. admin_handler.go:1021
   └─ Goroutine uses context.Background() (NO TIMEOUT)

2. CodeCraft API hangs
   └─ Connection open, no response

3. HTTP client timeout (120s) fires ✅
   └─ model_gateway.go:77 - Timeout: 120 * time.Second

4. gateway.Call() returns error ✅
   └─ Error propagates up the call stack

5. charter_extractor.Extract() returns error ✅
   └─ LLM call failed, return error

6. ingestion_pipeline.IngestTranscript() calls pauseOnLLMFailure() ✅
   └─ Tries to UPDATE ingestion_jobs SET status='paused'

7. pauseOnLLMFailure() DB update uses context.Background() ❌
   └─ NO DEADLINE on DB query

8. If DB connection is slow/hung, UPDATE blocks forever
   └─ Context has no timeout, query waits indefinitely

9. Goroutine hangs, job stays 'running' forever
   └─ No error logged, no status change, no recovery
```

### Why `context.Background()` is Dangerous

**Problem:**
- `context.Background()` has **NO deadline**
- If ANY step hangs (LLM timeout, DB deadlock, network partition), goroutine hangs forever
- No automatic recovery, no timeout, no error

**Impact:**
- Admin sees "running" forever with no error message
- No way to retry (job is not 'failed' or 'paused')
- Only fix: restart entire backend (kills all in-flight jobs)

**Why it happened:**
- Ingestion can take 10-30 minutes (chunking, embedding, LLM calls)
- Developer assumed all operations would complete or fail quickly
- Didn't account for: slow DB connections, network partitions, API hangs

---

## The Fix

### 1. Add 2-Hour Timeout to All Ingestion Goroutines

**Files Changed:**
- `backend-go/internal/admin/admin_handler.go`
  - Line 1021: Upload handler
  - Line 1428: Resume handler
  - Line 1543: Retry handler
  - Line 1251: RegenerateCharter handler (10-minute timeout)

**Before:**
```go
go func() {
    _, err := h.ingestion.IngestTranscript(
        context.Background(),  // ❌ NO TIMEOUT!
        jobID, expertID, expertName,
        string(content), header.Filename,
        false,
    )
    // ...
}()
```

**After:**
```go
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
    defer cancel()  // prevent context leak

    _, err := h.ingestion.IngestTranscript(
        ctx,  // ✅ 2-hour timeout
        jobID, expertID, expertName,
        string(content), header.Filename,
        false,
    )
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            h.logger.Error("ingestion timeout: exceeded 2-hour deadline", ...)
            return
        }
        // ...
    }
}()
```

### 2. Improve `pauseOnLLMFailure()` Error Handling

**File Changed:**
- `backend-go/internal/training/ingestion_pipeline.go`
  - Line 920-950: pauseOnLLMFailure() function

**Problem:**
- DB update failure was logged but ignored
- Job stayed 'running' if UPDATE failed
- No fallback strategy

**Fix:**
1. **Check context before DB update**
   - If context already cancelled, use fresh context

2. **Fallback with fresh context**
   - If primary UPDATE fails, try with fresh 5s timeout
   - Gives one last chance to update DB

3. **Mark as 'failed' if both attempts fail**
   - Better to fail explicitly than hang in 'running' state
   - Admin sees clear error, can investigate DB issue

**Code:**
```go
// Primary attempt with original context
_, err := p.db.Exec(ctx, updateQuery, ...)
if err != nil {
    // Fallback: try with fresh 5s timeout
    fallbackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    _, fallbackErr := p.db.Exec(fallbackCtx, updateQuery, ...)
    if fallbackErr != nil {
        // Both failed - mark as 'failed' instead of 'paused'
        failCtx, failCancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer failCancel()
        _, _ = p.db.Exec(failCtx,
            `UPDATE ingestion_jobs SET status='failed', error_message=$1 WHERE id=$2`,
            fmt.Sprintf("pauseOnLLMFailure: DB update failed twice: %s", fallbackErr.Error()),
            jobID,
        )
        return fmt.Errorf("pauseOnLLMFailure: failed to update job status: %w", fallbackErr)
    }
}
```

---

## Mental Model Verification

### Scenario 1: Normal Ingestion (5 minutes)
- Context timeout: 2 hours
- Ingestion completes in 5 minutes ✅
- Context cancelled via defer ✅
- No timeout error ✅

### Scenario 2: LLM Hangs (120s)
- HTTP client timeout fires after 120s ✅
- `gateway.Call()` returns error ✅
- `pauseOnLLMFailure()` called ✅
- DB update has 2-hour deadline ✅
- Job paused successfully ✅

### Scenario 3: DB Deadlock (infinite)
- `pauseOnLLMFailure()` tries UPDATE ❌
- DB query blocks on lock
- Context deadline (2h) fires ✅
- DB query cancelled ✅
- Fallback with fresh 5s context ✅
- Fallback succeeds (lock released) ✅
- Job paused successfully ✅

### Scenario 4: DB Down (both updates fail)
- Primary UPDATE fails (connection refused) ❌
- Fallback UPDATE also fails (DB still down) ❌
- Mark job as 'failed' (not 'paused') ✅
- Log clear error ✅
- Admin sees failure, can investigate DB ✅

### Scenario 5: Entire Pipeline Hangs (>2h)
- Context deadline fires after 2 hours ✅
- All in-flight operations cancelled ✅
- Job fails with timeout error ✅
- Admin sees clear error message ✅
- Can retry from checkpoint ✅

---

## Cross-Questions Answered

### Q: Why 2 hours, not 30 minutes?
**A:** Large transcripts (100k+ words) can take 30-60 minutes for embedding. 2 hours is generous but prevents infinite hangs.

### Q: What if legitimate ingestion takes >2 hours?
**A:** Checkpoint system saves progress every batch. Retry continues from last checkpoint. No data loss.

### Q: Why not use `c.Request.Context()` from HTTP request?
**A:** HTTP request context is cancelled when client disconnects. Ingestion must continue even if admin closes browser tab.

### Q: What about the 24-hour auto-fail checker?
**A:** Still needed for jobs paused by admin (intentional pause). This fix handles hung jobs that never reach 'paused' state.

### Q: Why 10 minutes for RegenerateCharter, not 2 hours?
**A:** RegenerateCharter only does charter extraction (single LLM call). HTTP timeout is 120s. 10min is 5x safety margin. If it takes >10min, something is broken (not just slow).

### Q: What if charter extraction legitimately takes >10 minutes?
**A:** Impossible. Charter extraction uses first 8000 chars of transcript (see `charter_extractor.go:56`). Even with slow LLM, 8000 chars takes <2 minutes. 10min timeout will never fire in normal operation.

---

## Testing Strategy

### Manual Test (Simulate Hang)

**Test 1: LLM Timeout**
```bash
# 1. Add sleep in charter_extractor.go
func (e *CharterExtractor) Extract(ctx context.Context, ...) (*Charter, error) {
    time.Sleep(3 * time.Minute)  // Simulate slow LLM
    // ...
}

# 2. Upload transcript
# 3. Wait 3 minutes
# 4. Check logs: should see "ingestion timeout: exceeded 2-hour deadline"
# 5. Check DB: status should be 'failed', error_message should mention timeout
```

**Test 2: DB Deadlock**
```bash
# 1. Start transaction in psql
BEGIN;
UPDATE ingestion_jobs SET status='running' WHERE id='<job_id>' FOR UPDATE;
# Don't commit - hold the lock

# 2. Upload transcript (different session)
# 3. Wait for charter LLM to fail
# 4. pauseOnLLMFailure() tries UPDATE - blocks on lock
# 5. Fallback should succeed after 5s
# 6. Check DB: status should be 'paused'
```

**Test 3: Context Deadline During UPDATE**
```bash
# 1. Set context timeout to 5 seconds in admin_handler.go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

# 2. Add sleep before DB update in pauseOnLLMFailure()
time.Sleep(10 * time.Second)

# 3. Upload transcript
# 4. Wait for charter LLM to fail
# 5. Context deadline fires during sleep
# 6. Fallback should succeed (no sleep in fallback)
# 7. Check DB: status should be 'paused'
```

### Unit Tests (TODO)

```go
// Test timeout fires
func TestIngestionTimeout(t *testing.T) {
    // Set 1-second timeout
    // Add 2-second sleep in IngestTranscript
    // Verify error is context.DeadlineExceeded
}

// Test context cancellation
func TestIngestionContextCancellation(t *testing.T) {
    // Start ingestion
    // Cancel context after 1 second
    // Verify goroutine exits cleanly
}

// Test pauseOnLLMFailure with cancelled context
func TestPauseOnLLMFailureWithCancelledContext(t *testing.T) {
    // Create cancelled context
    // Call pauseOnLLMFailure
    // Verify fallback succeeds
}

// Test pauseOnLLMFailure with DB error
func TestPauseOnLLMFailureWithDBError(t *testing.T) {
    // Mock DB to return error
    // Call pauseOnLLMFailure
    // Verify fallback is attempted
}
```

---

## References

### Go Context Package
- https://pkg.go.dev/context
- https://go.dev/blog/context

### Learning Materials Applied

**Arpit Bhiyani AI Masterclass:**
- "Always set timeouts on background goroutines"
- "Fail fast is better than hang forever"
- "Every critical operation needs a fallback"

**System Design Master Class:**
- "Timeouts prevent cascading failures"
- "Explicit failure is better than silent hang"
- "Always have a recovery path"

**Mental Model:**
- Cross-questioning: "What if this operation hangs?"
- Scenario verification: "What happens if DB is down?"
- Mental code execution: "Trace the error path"

---

## Impact

### Before Fix
- ❌ Jobs hang forever in 'running' state
- ❌ No error message, no way to retry
- ❌ Only recovery: restart backend (kills all jobs)
- ❌ Admin has no visibility into the failure

### After Fix
- ✅ Jobs timeout after 2 hours (generous but finite)
- ✅ Clear error message: "context deadline exceeded"
- ✅ Job marked as 'failed' (not stuck in 'running')
- ✅ Admin can retry from checkpoint (no data loss)
- ✅ Fallback strategy for DB update failures

---

## Lessons Learned

1. **Never use `context.Background()` in production goroutines**
   - Always use `context.WithTimeout()` or `context.WithDeadline()`
   - Even if you think operations will complete quickly

2. **Always have a fallback for critical operations**
   - DB updates, LLM calls, external API calls
   - Use fresh context with short timeout for fallback

3. **Fail explicitly, never hang silently**
   - Better to fail with clear error than hang forever
   - Admin can investigate and fix explicit failures

4. **Test timeout scenarios**
   - Simulate slow operations, DB deadlocks, network partitions
   - Verify goroutines exit cleanly on timeout

5. **Log context cancellation clearly**
   - Distinguish between normal completion and timeout
   - Include context in error messages

---

## Status

**Fixed:** 2026-09-17  
**Commits:**
- `fix(training): add timeout context to ingestion goroutine`
- `fix(training): improve pauseOnLLMFailure error handling`
- `fix(admin): add 2-hour timeout to resume and retry ingestion`
- `fix(admin): add 10-minute timeout to RegenerateCharter`

**Deployed:** Pending  
**Verified:** Manual testing required
