// Line 1
package receptionist

import (
    "context"
    "fmt"
    "strings"
)

// Line 8
type SynthesisInput struct {
    Session     *ReceptionistSession
    Checkpoints []ChecklistItem
    Notes       []NoteEntry
    ExpertCalls []map[string]interface{}
    Agenda      string
    Persona     string
    Language    Language // client language for summary, but final is always English
}

// Line 19
type SynthesisService struct {
    llm LLMClient
}

// Line 23
type LLMClient interface {
    Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

// Line 27
func NewSynthesisService(llm LLMClient) *SynthesisService { return &SynthesisService{llm: llm} }

// Line 29
func (s *SynthesisService) BuildSummary(ctx context.Context, in SynthesisInput) (string, error) {
    // Summary is in Client Language (requirement: client se usi bhasha me baat)
    sys := fmt.Sprintf("You are a brilliant secretary. Summarize clearly in %s. Be concise, cover all checkpoints, expert inputs, changes. Ask for rating.", in.Language)
    var b strings.Builder
    b.WriteString(fmt.Sprintf("Agenda: %s\n", in.Agenda))
    b.WriteString(fmt.Sprintf("Persona: %s\n", in.Persona))
    for i, cp := range in.Checkpoints {
        status := string(cp.Status)
        summary := ""
        if cp.CommittedSummary != nil { summary = *cp.CommittedSummary }
        b.WriteString(fmt.Sprintf("\nCheckpoint %d [%s]: %s\nSummary: %s\nRatingHistory: %v\n", i+1, status, cp.Text, summary, cp.RatingHistory))
    }
    b.WriteString("\nNotes:\n")
    for _, n := range in.Notes { b.WriteString(fmt.Sprintf("- [%s] %s\n", n.Type, n.SummaryText)) }
    b.WriteString("\nExpert Consultations:\n")
    for _, ec := range in.ExpertCalls { b.WriteString(fmt.Sprintf("- %v\n", ec)) }
    return s.llm.Complete(ctx, sys, b.String())
}

// Line 50
func (s *SynthesisService) BuildFinalResponse(ctx context.Context, in SynthesisInput) (string, error) {
    // Final is ALWAYS ENGLISH + Persona voice (requirement)
    sys := fmt.Sprintf(`You are acting as a "%s". Generate a final professional document in proper English.
Rules:
- Start with Original Requirement
- Then Agenda & Final Goal
- Then Per-Checkpoint Journey (what was discussed, what was decided)
- Then Expert Consultations Referenced (synthesize, don't just copy)
- Then Changes Incorporated
- Finally Final Architecture/Response in %s voice
- Use structured headings, bullet points, tables where needed.
- Do not invent beyond provided context.`, in.Persona, in.Persona)
    var b strings.Builder
    b.WriteString(fmt.Sprintf("ORIGINAL REQUIREMENT & AGENDA:\n%s\n\n", in.Agenda))
    b.WriteString("CHECKPOINTS (WARM Summaries):\n")
    for i, cp := range in.Checkpoints {
        summ := ""
        if cp.CommittedSummary != nil { summ = *cp.CommittedSummary }
        b.WriteString(fmt.Sprintf("%d. %s -> %s (ratings: %v)\n", i+1, cp.Text, summ, cp.RatingHistory))
    }
    b.WriteString("\nNOTES LEDGER:\n")
    for _, n := range in.Notes { b.WriteString(fmt.Sprintf("[%s] %s\n", n.Type, n.SummaryText)) }
    b.WriteString("\nEXPERT CALLS (Relevant Extracts):\n")
    for _, ec := range in.ExpertCalls { b.WriteString(fmt.Sprintf("%v\n", ec)) }
    return s.llm.Complete(ctx, sys, b.String())
}
