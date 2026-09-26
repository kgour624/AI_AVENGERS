// Package capability maps a domain to WHAT KIND OF WORK it may do, so a
// workflow can route a task to the right expert instead of handing every phase
// to whichever experts happen to exist.
//
// WHY this exists (production incident): a workflow whose experts were all
// "system design" produced Go code artifacts — the implementation phase used the
// same designers because nothing declared that code needs a programming expert.
// The fix is a declaration of capability, not a special case for Go:
//
//	design         — system design, HLD/LLD, architecture
//	implementation — writes code in a language (go, java, csharp, python, ...)
//	testing        — writes/owns tests and quality gates
//	data           — data engineering / data science / ML pipelines
//	product        — product management, requirements
//	other          — anything not yet classified (never silently treated as code)
//
// WHERE the declaration lives: domain_profiles.config (JSONB), the same row the
// China Wall already reads. New domains (.NET, data engineer, Rust, ...) are
// onboarded by adding a profile row with "capability"/"language" — there is no
// language list in Go code to update, which is what keeps this from being
// hardcoded to today's domains.
package capability

import (
	"strings"
)

// Kind is what a piece of work needs. Empty kind means "unclassified", which
// must never satisfy a code or test requirement.
type Kind string

const (
	KindUnclassified   Kind = ""
	KindDesign         Kind = "design"
	KindImplementation Kind = "implementation"
	KindTesting        Kind = "testing"
	KindData           Kind = "data"
	KindProduct        Kind = "product"
	KindOther          Kind = "other"
)

// Declaration is one domain's declared capability. It mirrors the optional
// "capability"/"language" keys of domain_profiles.config; both are optional so
// existing rows keep working (they simply stay unclassified until edited).
type Declaration struct {
	Kind     Kind
	Language string // lowercased, e.g. "go", "java", "csharp", "python"; "" when N/A
}

// NormName canonicalises a domain/language token: case, surrounding and inner
// whitespace, hyphens and underscores all fold together, so "System Design",
// "system design" and "system_design" are one identity. This is the same rule
// the China Wall's profile lookup uses; keeping one rule avoids the class of bug
// where a profile exists but never matches.
func NormName(s string) string {
	v := strings.ToLower(strings.TrimSpace(s))
	v = strings.ReplaceAll(v, "-", " ")
	v = strings.ReplaceAll(v, "_", " ")
	return strings.Join(strings.Fields(v), " ")
}

// LanguageAliases folds the ways one language is commonly written so a mapping
// still matches when the admin types "C#" or ".NET". Only aliases that mean the
// SAME language belong here; it is a spelling table, not a skill table, and an
// unknown language is passed through unchanged (never guessed into another).
func LanguageAliases(raw string) string {
	v := NormName(raw)
	switch v {
	case "c#", "csharp", "c sharp", ".net", "dotnet", "dot net", "asp net", "aspnet":
		return "csharp"
	case "golang", "go":
		return "go"
	case "js", "javascript", "node", "nodejs", "node js":
		return "javascript"
	case "ts", "typescript":
		return "typescript"
	case "py", "python", "python3":
		return "python"
	}
	return v
}
