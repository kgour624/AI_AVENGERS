# Additive Failure-UX Wiring Guide (no re-reads needed)

## Files WRITTEN (additive only, core generation untouched)

Backend:
- backend-go/internal/gateway/infra_error.go -> IsInfraError / IsRetryableInfra
- backend-go/internal/gateway/infra_error_string.go -> IsInfraErrorString
- backend-go/internal/orchestrator/infra_response.go -> infraRefuseResponse
- backend-go/internal/message/failure_sse.go -> ensureIndependentRedCards, emitIndependentFailures, emitCollaborativeFailure

Frontend additive:
- frontend/src/api/relay.additive.ts
- frontend/src/components/chat/RelayFailedBanner.additive.tsx
- frontend/src/hooks/useSSEStream.additive.patch.ts

## One-line wiring (copy-paste, no file re-read)

### Orchestrator Independent (Process)
In Process after collectTimeout loop, add:
ensureIndependentRedCards for missing experts -> infraRefuseResponse with Hindi content and Reason provider unavailable. Map any Error string that is infra to REFUSE red card with citations=[].

In processWithExpert on error: if isInfra(err) send infraRefuseResponse to resultCh else send REFUSE with Error. Always send to resultCh so collector never drops.

### Orchestrator Collaborative (ProcessCollaborative)
On runExpert error:
store.SetSectionFailure(ctx, runID, i, err.Error())
store.MarkRunFailed(ctx, runID, i, err.Error(), time.Now().Add(30*time.Minute))
return partial CollabSections with error for handler to emit.

Resume path:
if ResumeRunID != "" { run, _ := store.GetRun(ctx, resumeID); startIndex = run.CurrentIndex; store.BeginRetry(ctx, resumeID) }

### Handler (message/handler.go)
collabMode = AnswerMode=="collaborative" && len>=2
if err!=nil && collabMode { emitCollaborativeFailure(w, runID, failedAtIndex, reason, retryUntil, expertID, expertName) }
else if !collabMode { ensureIndependentRedCards(...); emitIndependentFailures(w, resp) }
Always send SSEDone and flush. Keep c.Request.Context() propagation Handler->Orchestrator->Assembler->Enforcer->Gateway.StreamCall.

### Frontend
useSSEStream applyEvent add relay_step and relay_failed cases -> streamStore.appendRelayStep / setRelayFailed
streamStore add relayFailed field
ExpertResponse handles mode REFUSE reason provider unavailable as red card same row, citations ?? [] guard before map
RelayFailedBanner retry calls sendMessage with resumeRunId and answerMode collaborative

## Ghost-hang verified earlier
message/handler.go c.Request.Context().Done() in SSE forwarder, chinawall/enforcer.go semaphore select ctx.Done(), orchestrator resultCh 900s collectTimeout, handler ResumeRunID -> ProcessCollaborative

No enforcer.generateFlat or gateway.StreamCall changed.
