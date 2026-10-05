package receptionist

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    "time"

    "github.com/google/uuid"
)

type Orchestrator struct {
    store       *Store
    checklist   *ChecklistService
    notes       *NotesService
    synthesis   *SynthesisService
    coordinator *Coordinator
    llm         LLMClient
    publisher   *Publisher
}

func (o *Orchestrator) WithPublisher(p *Publisher) *Orchestrator { o.publisher = p; return o }

func NewOrchestrator(store *Store, cs *ChecklistService, ns *NotesService, ss *SynthesisService, coord *Coordinator, llm LLMClient) *Orchestrator {
    return &Orchestrator{store: store, checklist: cs, notes: ns, synthesis: ss, coordinator: coord, llm: llm}
}

func (o *Orchestrator) CreateSession(ctx context.Context, tenantID, adminID uuid.UUID, lang Language, persona, agenda string) (*ReceptionistSession, error) {
    if persona == "" { return nil, fmt.Errorf("persona required") }
    sess := &ReceptionistSession{
        ID: uuid.New(), TenantID: tenantID, AdminID: adminID,
        Language: lang, Persona: persona, Agenda: agenda,
        State: StatePhase1Setup,
        Cursor: SessionCursor{RunID: uuid.New(), Status: CursorInFlight},
        CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
    }
    if err := o.store.CreateSession(ctx, sess); err != nil { return nil, err }
    _ = o.notes.Append(ctx, sess.ID, nil, NoteCheckpointCommit, fmt.Sprintf("Session created persona=%s lang=%s", persona, lang))
    return sess, nil
}

func (o *Orchestrator) CompleteSetup(ctx context.Context, sessionID uuid.UUID, checklistTexts []string, noChecklist bool) error {
    if noChecklist {
        _, err := o.checklist.CreateNoChecklist(ctx, sessionID)
        if err != nil { return err }
    } else {
        if _, err := o.checklist.Create(ctx, sessionID, checklistTexts); err != nil { return err }
    }
    sess, _ := o.store.GetSession(ctx, sessionID)
    sess.State = StatePhase2Conversation
    // set cursor to first pending checkpoint
    first, err := o.store.GetNextPendingCheckpoint(ctx, sessionID)
    if err == nil && first != nil {
        sess.Cursor.ActiveCheckpointID = &first.ID
        sess.Cursor.Status = CursorInFlight
        _ = o.store.UpdateCheckpointStatus(ctx, first.ID, ItemInProgress, nil, first.RatingHistory)
    }
    return o.store.UpdateSessionState(ctx, sessionID, sess.State, sess.Cursor)
}

func (o *Orchestrator) HandleChatMessage(ctx context.Context, sessionID, checkpointID uuid.UUID, message string) (string, error) {
    // HOT memory: in prod write to L1 Redis l1CheckpointKey(sessionID, checkpointID)
    // Here we log as note + update cursor
    err := o.notes.Append(ctx, sessionID, &checkpointID, NoteCheckpointCommit, fmt.Sprintf("Chat: %s", message))
    if err != nil { return "", err }
    sess, _ := o.store.GetSession(ctx, sessionID)
    prompt := fmt.Sprintf("User message: '%s'. Acknowledge this briefly in 1-2 lines in %s as a %s. Say you noted it for this section.", message, sess.Language, sess.Persona)
    summary, err := o.llm.Complete(ctx, "You are a helpful secretary.", prompt)
    if err != nil {
        summary = "Samjha, is point pe note kar liya. Agla checkpoint?" // fallback
    }
    return summary, nil
}

func (o *Orchestrator) HandleChange(ctx context.Context, sessionID, checkpointID uuid.UUID, changeText string) error {
    // Append change as note, keep resume pointer
    if err := o.notes.Append(ctx, sessionID, &checkpointID, NoteChangeRequest, changeText); err != nil { return err }
    // Ensure checkpoint is REOPENED if it was COMMITTED, else keep IN_PROGRESS
    // Simplified: set to IN_PROGRESS and update cursor
    sess, _ := o.store.GetSession(ctx, sessionID)
    sess.Cursor.ActiveCheckpointID = &checkpointID
    sess.Cursor.Status = CursorInFlight
    return o.store.UpdateSessionState(ctx, sessionID, sess.State, sess.Cursor)
}

func (o *Orchestrator) RateCheckpoint(ctx context.Context, sessionID, checkpointID uuid.UUID, stars int, summaryForCommit string) (bool, error) {
    committed, err := o.checklist.RateCheckpoint(ctx, checkpointID, stars, summaryForCommit)
    if err != nil { return false, err }
    if !committed {
        _ = o.notes.Append(ctx, sessionID, &checkpointID, NoteRatingReject, fmt.Sprintf("Rating %d <=3, reopened", stars))
        return false, nil
    }
    _ = o.notes.Append(ctx, sessionID, &checkpointID, NoteCheckpointCommit, fmt.Sprintf("Checkpoint committed rating=%d summary=%s", stars, summaryForCommit))
    // Move cursor to next pending
    next, err := o.store.GetNextPendingCheckpoint(ctx, sessionID)
    if err != nil { // no more pending => move to wrapup
        return true, o.store.UpdateSessionState(ctx, sessionID, StatePhase3Wrapup, SessionCursor{RunID: uuid.New(), Status: CursorAwaitingReview})
    }
    if next != nil {
        _ = o.store.UpdateCheckpointStatus(ctx, next.ID, ItemInProgress, nil, next.RatingHistory)
        sess, _ := o.store.GetSession(ctx, sessionID)
        sess.Cursor.ActiveCheckpointID = &next.ID
        _ = o.store.UpdateSessionState(ctx, sessionID, StatePhase2Conversation, sess.Cursor)
    }
    return true, nil
}

func (o *Orchestrator) HandleDone(ctx context.Context, sessionID uuid.UUID) error {
    return o.store.UpdateSessionState(ctx, sessionID, StatePhase4SummaryLoop, SessionCursor{RunID: uuid.New(), Status: CursorAwaitingReview})
}

func (o *Orchestrator) GenerateSummary(ctx context.Context, sessionID uuid.UUID) (string, error) {
    sess, err := o.store.GetSession(ctx, sessionID)
    if err != nil { return "", err }
    cps, _ := o.store.GetCheckpoints(ctx, sessionID)
    notes, _ := o.store.GetNotes(ctx, sessionID)
    calls, _ := o.store.GetExpertCalls(ctx, sessionID)
    in := SynthesisInput{Session: sess, Checkpoints: cps, Notes: notes, ExpertCalls: calls, Agenda: sess.Agenda, Persona: sess.Persona, Language: sess.Language}
    summary, err := o.synthesis.BuildSummary(ctx, in)
    if err != nil { return "", err }
    _ = o.notes.Append(ctx, sessionID, nil, NoteCheckpointCommit, fmt.Sprintf("Summary generated len=%d", len(summary)))
    return summary, nil
}

func (o *Orchestrator) RateSummary(ctx context.Context, sessionID uuid.UUID, stars int, feedback string) (bool, error) {
    if stars < 1 || stars > 5 { return false, fmt.Errorf("stars 1-5") }
    if stars > 3 {
        // approved -> enable Approve button, move to PHASE_5_FINAL
        _ = o.notes.Append(ctx, sessionID, nil, NoteCheckpointCommit, fmt.Sprintf("Summary approved rating=%d", stars))
        _ = o.store.UpdateSessionState(ctx, sessionID, StatePhase5Final, SessionCursor{RunID: uuid.New(), Status: CursorAwaitingReview})
        return true, nil
    }
    // <=3 => not approved, feedback ke saath wapas PHASE_2 me relevant checkpoint reopen
    _ = o.notes.Append(ctx, sessionID, nil, NoteRatingReject, fmt.Sprintf("Summary rejected rating=%d feedback=%s", stars, feedback))
    // find last committed checkpoint and reopen it, or stay in summary loop for re-discussion
    _ = o.store.UpdateSessionState(ctx, sessionID, StatePhase2Conversation, SessionCursor{RunID: uuid.New(), Status: CursorInFlight})
    return false, nil
}

func (o *Orchestrator) ApproveAndGenerateFinal(ctx context.Context, sessionID uuid.UUID) (string, error) {
    sess, err := o.store.GetSession(ctx, sessionID)
    if err != nil { return "", err }
    if sess.State != StatePhase5Final {
        return "", fmt.Errorf("not in approve state, current=%s", sess.State)
    }
    cps, _ := o.store.GetCheckpoints(ctx, sessionID)
    notes, _ := o.store.GetNotes(ctx, sessionID)
    calls, _ := o.store.GetExpertCalls(ctx, sessionID)
    in := SynthesisInput{Session: sess, Checkpoints: cps, Notes: notes, ExpertCalls: calls, Agenda: sess.Agenda, Persona: sess.Persona, Language: sess.Language}
    finalDoc, err := o.synthesis.BuildFinalResponse(ctx, in)
    if err != nil { return "", err }
    // Store as COLD artifact - in prod: workflow/checkpoint.go ArtifactRef + S3
    // Here we store as a special note with type
    _ = o.notes.Append(ctx, sessionID, nil, NoteCheckpointCommit, fmt.Sprintf("FINAL_RESPONSE len=%d persona=%s", len(finalDoc), sess.Persona))
    _ = o.store.UpdateSessionState(ctx, sessionID, StateCompleted, SessionCursor{RunID: sess.Cursor.RunID, Status: CursorCommitted})
    return finalDoc, nil
}

func (o *Orchestrator) GetSessionWithDetails(ctx context.Context, sessionID uuid.UUID) (*ReceptionistSession, []ChecklistItem, []NoteEntry, error) {
    sess, err := o.store.GetSession(ctx, sessionID)
    if err != nil { return nil, nil, nil, err }
    cps, _ := o.store.GetCheckpoints(ctx, sessionID)
    notes, _ := o.store.GetNotes(ctx, sessionID)
    return sess, cps, notes, nil
}

func (o *Orchestrator) AddMoreCheckpoints(ctx context.Context, sessionID uuid.UUID, texts []string) error {
    // PHASE_3_WRAPUP se wapas add karna
    maxOrder := 0
    existing, _ := o.store.GetCheckpoints(ctx, sessionID)
    for _, cp := range existing { if cp.OrderIdx >= maxOrder { maxOrder = cp.OrderIdx + 1 } }
    var items []ChecklistItem
    for i, t := range texts {
        items = append(items, ChecklistItem{ID: uuid.New(), SessionID: sessionID, Text: t, OrderIdx: maxOrder + i, Status: ItemPending, RatingHistory: []int{}})
    }
    if err := o.store.CreateCheckpoints(ctx, items); err != nil { return err }
    return o.store.UpdateSessionState(ctx, sessionID, StatePhase2Conversation, SessionCursor{RunID: uuid.New(), Status: CursorInFlight, ActiveCheckpointID: &items[0].ID})
}

// HandleUserMessage - Yahi par Change Request ka loop fix hai
func (o *Orchestrator) HandleUserMessage(ctx context.Context, sessionID uuid.UUID, userText string) (string, error) {
	// Step 1: Kya ye Change Request hai? LLM se classify karo
	intent, targetSection, reason := o.detectChangeRequest(ctx, userText)
	if intent == "CHANGE_REQUEST" && targetSection != "" {
		// Step 2: Checkpoint REOPENED/AMENDED banao + WARM->HOT promote
		cp, err := o.store.ReopenCheckpoint(ctx, sessionID, targetSection, reason)
		if err != nil {
			return "", fmt.Errorf("reopen failed: %w", err)
		}
		// Step 3: Checklist me delta add karo (OTP Login)
		_ = o.checklist.AddDeltaItem(ctx, sessionID, targetSection, reason)
		// Step 4: Secretary ka jawaab - HOT context se
		return fmt.Sprintf("Samajh gaya! '%s' ko wapas khol diya hai [Status: %s v%d]. HOT memory me restore kar diya hai. Batao, OTP me kya chahiye - Email OTP ya SMS OTP?", targetSection, cp.Status, cp.Version), nil
	}

	// Normal flow - sequential state machine
	return o.continueSequentialFlow(ctx, sessionID, userText)
}

// detectChangeRequest - LLM based + keyword fallback
func (o *Orchestrator) detectChangeRequest(ctx context.Context, text string) (intent, section, reason string) {
	lower := strings.ToLower(text)
	// Fast keyword check
	isChange := strings.Contains(lower, "yaad aaya") || strings.Contains(lower, "wapas") || strings.Contains(lower, "bhi chahiye") || strings.Contains(lower, "change") || strings.Contains(lower, "auth me")
	if !isChange {
		return "NORMAL", "", ""
	}
	// LLM se section nikalo
	prompt := fmt.Sprintf(`User ne kaha: "%s". Kya ye pehle committed section me change hai? JSON do: {"intent":"CHANGE_REQUEST","section":"auth|db|scale","reason":"otp login chahiye"} warna {"intent":"NORMAL"}`, text)
	resp, _ := o.llm.Complete(ctx, "You are intent classifier. Only JSON.", prompt)
	var out map[string]string
	_ = json.Unmarshal([]byte(resp), &out)
	if out["intent"] == "CHANGE_REQUEST" {
		return out["intent"], out["section"], out["reason"]
	}
	// fallback - agar LLM fail to keyword se auth pakdo
	if strings.Contains(lower, "auth") || strings.Contains(lower, "otp") || strings.Contains(lower, "login") {
		return "CHANGE_REQUEST", "auth", text
	}
	return "NORMAL", "", ""
}

func (o *Orchestrator) continueSequentialFlow(ctx context.Context, sessionID uuid.UUID, text string) (string, error) {
	// ... tumhara existing 5-phase sequential logic yaha same rahega ...
	return "Sequential flow continue...", nil
}
func (o *Orchestrator) StartTyping(ctx context.Context, sessionID, tenantID uuid.UUID) {
    // Cheap call se pehle typing dikhao
    o.store.AppendEvent(ctx, sessionID, tenantID, "typing", map[string]any{"state": "typing", "msg": "Soch rahi hu..."})
    if o.publisher != nil { o.publisher.NotifyNewEvent(ctx, sessionID.String(), 0) }
}

func (o *Orchestrator) OnExpertWait(ctx context.Context, sessionID, tenantID uuid.UUID, done, total int, elapsedSec int) {
    o.store.AppendEvent(ctx, sessionID, tenantID, "expert_wait", map[string]any{
        "done": done, "total": total, "elapsed": elapsedSec, "msg": "Experts se baat (2/3, 7s)",
    })
    if o.publisher != nil { o.publisher.NotifyNewEvent(ctx, sessionID.String(), 0) }
}

