// assemble.go: pure Markdown join of the ordered sections into the single
// message content a Collaborative Relay answer is saved/streamed as.
//
// WHY a plain join and not another LLM call: assembly is pure formatting —
// the sections are already final content, in final order. Spending an LLM
// call here would add latency/cost for zero benefit and risk the model
// paraphrasing (and so subtly altering) an expert's already-approved
// answer. consistency.go (a separate, optional LLM pass) is where any
// cross-section judgement belongs, not here.
package collab

import "strings"

// AssembleMarkdown joins the relay's finished sections into one Markdown
// document: each section becomes a level-2 heading (its title, with the
// answering expert's name) followed by that section's content. Sections
// with empty Content (should not normally happen — relay.go always fills
// in at least a placeholder for a failed expert) are still rendered with
// their heading so the output always accounts for every planned section.
func AssembleMarkdown(sections []Section) string {
	var sb strings.Builder
	for i, section := range sections {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("## ")
		sb.WriteString(section.SectionTitle)
		sb.WriteString(" (")
		sb.WriteString(section.ExpertName)
		sb.WriteString(")\n\n")
		sb.WriteString(strings.TrimSpace(section.Content))
	}
	return sb.String()
}
