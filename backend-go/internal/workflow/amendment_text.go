package workflow

// The deterministic core of §7.5 step 3: turning an approved {old_text,
// new_text} into new file content, and keeping DECISIONS.md in order.
//
// Split from amendment.go for the same two reasons client_remote.go is split
// from client_repo.go:
//
//  1. This is the code that REPLACES the Aider session the doc called for. Its
//     whole justification is that it has a defined answer for every input and an
//     LLM does not. Code making that claim should be readable on its own, with
//     no db, no git and no filesystem in the way.
//  2. Depending on nothing but the standard library, it can be exercised
//     directly — which matters in an environment where the full package cannot
//     be compiled at all.

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ErrAmendmentStale means the design text the proposal was written against is no
// longer there, or is there more than once. The proposal has to be made again;
// it cannot be applied by guessing.
var ErrAmendmentStale = errors.New("the design moved since this was proposed")

// applyAmendmentText produces the amended file content.
//
// The three outcomes, and why each one is what it is:
//
//	oldText == ""          append. This is how a new acceptance criterion, or a
//	                       new statement row, is added: there is nothing to
//	                       replace.
//	oldText appears once   replace. The only case that can be applied safely.
//	oldText appears 0 or   refuse. Zero means the design changed since the
//	>1 times               proposal was written; more than one means the proposal
//	                       does not say WHICH occurrence it meant. Replacing the
//	                       first would be a guess at the contract the code is
//	                       built against, so this is an error with a count in it
//	                       rather than a best effort.
//
// Deliberately the same rule Aider's own SEARCH/REPLACE blocks follow, which is
// why no model call is needed to apply an amendment.
func applyAmendmentText(content, oldText, newText string) (string, error) {
	if oldText == "" {
		// Exactly one blank line between the existing content and the addition,
		// whatever the file ended with. Markdown needs the separation, and a
		// growing tail of blank lines makes every later diff noisy.
		trimmed := strings.TrimRight(content, "\n")
		return trimmed + "\n\n" + strings.TrimRight(newText, "\n") + "\n", nil
	}

	switch n := strings.Count(content, oldText); {
	case n == 0:
		return "", fmt.Errorf("%w: the text this amendment replaces is no longer in the file", ErrAmendmentStale)
	case n > 1:
		return "", fmt.Errorf("%w: the text this amendment replaces appears %d times, so it is not clear which one was meant", ErrAmendmentStale, n)
	}

	if oldText == newText {
		return "", fmt.Errorf("this amendment changes nothing — the old and new text are identical")
	}
	return strings.Replace(content, oldText, newText, 1), nil
}

// decisionIDRe finds an existing decision id in DECISIONS.md.
var decisionIDRe = regexp.MustCompile(`\bDEC-(\d+)\b`)

// nextDecisionID returns the next id: the highest DEC-nnn already present, plus
// one.
//
// Highest-plus-one, not count-plus-one, for the same reason DesignSectionStore
// uses MAX+10 instead of (count+1)*10: counting reissues an id an existing entry
// already holds the moment anything is removed or written out of order, and a
// decision log with two DEC-007 entries is a log nobody can cite.
func nextDecisionID(content string) string {
	highest := 0
	for _, m := range decisionIDRe.FindAllStringSubmatch(content, -1) {
		if n, ok := atoiBounded(m[1]); ok && n > highest {
			highest = n
		}
	}
	return fmt.Sprintf("DEC-%03d", highest+1)
}

// insertNewestFirst puts entry above the existing entries but below the file's
// header, honouring the "Newest first." rule in the DECISIONS.md template.
//
// The insertion point is the first "## " heading, which is the newest existing
// entry. With no entries yet the entry goes at the end — after the header, which
// is where the first one belongs.
func insertNewestFirst(existing, entry string) string {
	const marker = "\n## "
	if idx := strings.Index(existing, marker); idx >= 0 {
		return existing[:idx+1] + entry + existing[idx+1:]
	}
	return strings.TrimRight(existing, "\n") + "\n\n" + entry
}

// acceptanceIDsForSection returns the highest n already used by an AC-<section>-n
// id, or 0 if the section has none.
func acceptanceIDsForSection(content string, sectionNo int) int {
	re := regexp.MustCompile(fmt.Sprintf(`\bAC-%d-(\d+)\b`, sectionNo))
	highest := 0
	for _, m := range re.FindAllStringSubmatch(content, -1) {
		if n, ok := atoiBounded(m[1]); ok && n > highest {
			highest = n
		}
	}
	return highest
}

// formatAcceptanceBlock renders one criterion in the exact shape §3.4 defines
// and documentCheck (authoring.go) validates.
//
// The format is not re-invented here: acceptanceLineRe, in authoring.go, is what
// parses it, and a block this function produces MUST match that regexp or the
// next authoring turn's document check fails on a criterion the client approved.
// That coupling is the reason the owner argument has to be a single token — see
// ownerTokenFromSectionPath.
func formatAcceptanceBlock(acID, owner, sectionPath, statement, verify, doneWhen string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## %s   owner: %s   section: %s\n", acID, owner, sectionPath)
	fmt.Fprintf(&b, "Statement: %s\n", strings.TrimSpace(statement))
	fmt.Fprintf(&b, "Verify:    %s\n", strings.TrimSpace(verify))
	fmt.Fprintf(&b, "Done when: %s\n", strings.TrimSpace(doneWhen))
	return b.String()
}

// ownerTokenFromSectionPath extracts the expert slug from an assigned section
// path: "design/20-unlimited-go.md" -> "unlimited-go".
//
// Why the slug and not experts.name: acceptanceLineRe captures the owner as \S+,
// so a two-word name like "Unlimited Go" parses as "Unlimited" and the
// criterion's owner is silently wrong. experts.slug is UNIQUE NOT NULL
// (migration 001) and has no spaces, and DesignSectionStore already built this
// path from it — so the correct single-token owner is sitting in the path and
// needs no extra query.
func ownerTokenFromSectionPath(sectionPath string) string {
	base := strings.TrimSuffix(filepath.Base(sectionPath), ".md")
	if i := strings.Index(base, "-"); i >= 0 && i+1 < len(base) {
		return base[i+1:]
	}
	return base
}

// firstLine returns the first line of s, bounded, for a commit subject and a
// decision-log heading.
func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, "\n\r"); i >= 0 {
		s = s[:i]
	}
	if s == "" {
		return "design amendment"
	}
	const limit = 100
	if len(s) > limit {
		return s[:limit] + "…"
	}
	return s
}

// atoiBounded parses a small non-negative integer, refusing anything that is not
// all digits or is absurdly large.
//
// Not strconv.Atoi because both callers want "skip this match" rather than an
// error value, and the regexp has already guaranteed the digits — the bound is
// the only real check left.
func atoiBounded(s string) (int, bool) {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
		if n > 1000000 {
			return 0, false
		}
	}
	return n, len(s) > 0
}
