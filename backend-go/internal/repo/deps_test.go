package repo

import (
	"container/heap"
	"sort"
	"testing"
)

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x := append([]string(nil), a...)
	y := append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

func specs(refs []importRef) []string {
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		out = append(out, ref.Spec)
	}
	return out
}

func TestExtractGoImportsHandlesGroupedAndAliased(t *testing.T) {
	src := `package main

import (
	"fmt"
	"ai_avengers/backend/internal/repo"
	alias "github.com/google/uuid"
)

import "os"

func main() { fmt.Println(repo.ProviderGitHub, alias.Nil, os.Args) }
`
	got := specs(extractGoImports(src))
	want := []string{"fmt", "ai_avengers/backend/internal/repo", "github.com/google/uuid", "os"}
	if !sameStrings(got, want) {
		t.Fatalf("extractGoImports() = %v, want %v", got, want)
	}
}

func TestExtractGoImportsIgnoresUnparseableAndCgo(t *testing.T) {
	if got := extractGoImports("this is not go at all"); len(got) != 0 {
		t.Fatalf("extractGoImports() = %v, want none for unparseable input", got)
	}
	if got := specs(extractGoImports("package p\n\nimport \"C\"\n")); len(got) != 0 {
		t.Fatalf("extractGoImports() = %v, want cgo import skipped", got)
	}
}

func TestExtractJSImportsCoversStaticRequireAndDynamic(t *testing.T) {
	src := `
import React from 'react'
import { helper } from './helper'
const legacy = require('../legacy/util')
import('./lazy')
`
	got := specs(extractJSImports(src))
	want := []string{"react", "./helper", "../legacy/util", "./lazy"}
	if !sameStrings(got, want) {
		t.Fatalf("extractJSImports() = %v, want %v", got, want)
	}
}

func TestExtractJSImportsDeduplicates(t *testing.T) {
	got := specs(extractJSImports("import a from './x'\nimport b from './x'\n"))
	if len(got) != 1 || got[0] != "./x" {
		t.Fatalf("extractJSImports() = %v, want one ./x", got)
	}
}

func TestExtractPythonImportsCoversAbsoluteAndRelative(t *testing.T) {
	src := `
import os
import a.b.c
from pkg.mod import thing
from . import sibling
from ..parent import up
`
	got := specs(extractPythonImports(src))
	want := []string{"os", "a.b.c", "pkg.mod", ".", "..parent"}
	if !sameStrings(got, want) {
		t.Fatalf("extractPythonImports() = %v, want %v", got, want)
	}
}

func TestResolveGoImportToPackageFiles(t *testing.T) {
	index := newRepoPathIndex([]string{
		"cmd/app/main.go",
		"internal/repo/service.go",
		"internal/repo/deps.go",
	})

	got := index.resolve("cmd/app/main.go", "go", importRef{Spec: "ai_avengers/backend/internal/repo"})
	want := []string{"internal/repo/service.go", "internal/repo/deps.go"}
	if !sameStrings(got, want) {
		t.Fatalf("resolve() = %v, want the package's files %v", got, want)
	}
}

func TestResolveJSRelativeWithExtensionAndIndex(t *testing.T) {
	index := newRepoPathIndex([]string{
		"src/api/client.ts",
		"src/lib/helper.ts",
		"src/lib/index.ts",
		"src/components/Button.tsx",
	})

	if got := index.resolve("src/api/client.ts", "javascript", importRef{Spec: "./../lib/helper"}); !sameStrings(got, []string{"src/lib/helper.ts"}) {
		t.Fatalf("resolve helper = %v, want src/lib/helper.ts", got)
	}
	if got := index.resolve("src/api/client.ts", "javascript", importRef{Spec: "../lib"}); !sameStrings(got, []string{"src/lib/index.ts"}) {
		t.Fatalf("resolve index = %v, want src/lib/index.ts", got)
	}
	// A bare specifier is an npm package; nothing in the tree should match.
	if got := index.resolve("src/api/client.ts", "javascript", importRef{Spec: "react"}); len(got) != 0 {
		t.Fatalf("resolve react = %v, want no repo file", got)
	}
}

func TestResolvePythonRelativeLevels(t *testing.T) {
	index := newRepoPathIndex([]string{
		"pkg/sub/mod.py",
		"pkg/sub/sibling.py",
		"pkg/parent.py",
	})

	if got := index.resolve("pkg/sub/mod.py", "python", importRef{Spec: ".sibling"}); !sameStrings(got, []string{"pkg/sub/sibling.py"}) {
		t.Fatalf("resolve .sibling = %v, want pkg/sub/sibling.py", got)
	}
	if got := index.resolve("pkg/sub/mod.py", "python", importRef{Spec: "..parent"}); !sameStrings(got, []string{"pkg/parent.py"}) {
		t.Fatalf("resolve ..parent = %v, want pkg/parent.py", got)
	}
}

func TestBuildRepoEdgesSkipsSelfImportsAndFindsCycles(t *testing.T) {
	contents := map[string]string{
		"a.go": "package a\n\nimport \"example/b\"\n",
		"b.go": "package b\n\nimport \"example/a\"\n",
		"c.go": "package c\n\nimport \"example/c\"\n", // self-import: must be dropped
	}
	languages := map[string]string{"a.go": "go", "b.go": "go", "c.go": "go"}

	edges := buildRepoEdges(contents, languages)

	var pairs []string
	for _, edge := range edges {
		pairs = append(pairs, edge.Src+"->"+edge.Dst)
		if edge.Src == edge.Dst {
			t.Fatalf("self-import retained: %s", edge.Src)
		}
	}
	// a<->b is the cycle the graph walk must survive; the file-level resolver
	// maps each import to the other file's package directory, which here is the
	// repository root, so both directions exist.
	if !sameStrings(pairs, []string{"a.go->b.go", "b.go->a.go"}) {
		t.Fatalf("edges = %v, want a<->b only", pairs)
	}
}

func TestBuildRepoEdgesIgnoresUnsupportedLanguage(t *testing.T) {
	contents := map[string]string{"main.rs": "use crate::foo;"}
	languages := map[string]string{"main.rs": "rust"}
	if edges := buildRepoEdges(contents, languages); len(edges) != 0 {
		t.Fatalf("edges = %v, want none for an unsupported language", edges)
	}
}

func TestRequirementTokensDropStopWordsAndShortWords(t *testing.T) {
	got := requirementTokens("The user should update the Invoicing and PDF export")
	want := []string{"update", "invoicing", "pdf", "export"}
	if !sameStrings(got, want) {
		t.Fatalf("requirementTokens() = %v, want %v", got, want)
	}
}

func TestScoreRepoFilePrefersExactFileName(t *testing.T) {
	tokens := []string{"invoicing"}

	nameScore, nameReason := scoreRepoFile("src/billing/invoicing.ts", tokens, 0, 0)
	dirScore, _ := scoreRepoFile("src/invoicing/other.ts", tokens, 0, 0)

	if nameScore <= dirScore {
		t.Fatalf("exact file-name match (%v) should outrank a directory match (%v)", nameScore, dirScore)
	}
	if nameReason == "" {
		t.Fatal("scoreRepoFile() returned no reason for a match")
	}
}

func TestScoreRepoFileUsesHubBonusWhenNothingMatches(t *testing.T) {
	score, reason := scoreRepoFile("src/core/registry.ts", []string{"invoicing"}, 10, 20)
	if score <= 0 {
		t.Fatalf("score = %v, want the hub bonus to keep a widely imported file", score)
	}
	if reason == "" {
		t.Fatal("scoreRepoFile() returned no reason for a hub file")
	}
}

func TestSuggestionHeapKeepsTopK(t *testing.T) {
	best := &suggestionHeap{}
	for _, s := range []float64{1, 5, 3, 9, 2, 7} {
		candidate := RepoFileSuggestion{Score: s}
		if best.Len() < 3 {
			heap.Push(best, candidate)
			continue
		}
		if (*best)[0].Score < candidate.Score {
			heap.Pop(best)
			heap.Push(best, candidate)
		}
	}
	if best.Len() != 3 {
		t.Fatalf("heap size = %d, want 3", best.Len())
	}
	got := map[float64]bool{}
	for _, item := range *best {
		got[item.Score] = true
	}
	for _, want := range []float64{9, 7, 5} {
		if !got[want] {
			t.Fatalf("top-K heap = %v, missing %v", *best, want)
		}
	}
}
