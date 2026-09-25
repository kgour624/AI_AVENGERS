package repo

import "testing"

// The incremental rule decides whether a sync re-downloads and re-embeds a
// file, so each branch is pinned here rather than being left to be discovered
// in production by watching a bill or a sync log.

func TestClassifyTreeReusesOnlyMatchingContentBackedFiles(t *testing.T) {
	size := 42
	entries := []repoTreeItem{
		{Path: "same.go", BlobSHA: "sha-a", SizeBytes: &size},
		{Path: "modified.go", BlobSHA: "sha-new"},
		{Path: "metadata_only.go", BlobSHA: "sha-m"},
		{Path: "no_provider_sha.go"},
		{Path: "brand_new.go", BlobSHA: "sha-n"},
	}
	previous := map[string]prevRepoFile{
		"same.go":            {BlobSHA: "sha-a", HasContent: true, Language: "go", SizeBytes: &size},
		"modified.go":        {BlobSHA: "sha-old", HasContent: true},
		"metadata_only.go":   {BlobSHA: "sha-m", HasContent: false},
		"no_provider_sha.go": {BlobSHA: "sha-p", HasContent: true},
	}

	unchanged, paths := classifyTree(entries, previous)

	if !unchanged["same.go"] {
		t.Fatal("same.go: identical blob id and stored content must be reused")
	}
	if len(paths) != 1 || paths[0] != "same.go" {
		t.Fatalf("unchanged paths = %v, want exactly [same.go]", paths)
	}

	for _, changed := range []string{"modified.go", "metadata_only.go", "no_provider_sha.go", "brand_new.go"} {
		if unchanged[changed] {
			t.Fatalf("%s must NOT be treated as unchanged", changed)
		}
	}
}

func TestClassifyTreeTreatsMetadataOnlyFileAsChanged(t *testing.T) {
	// The trap this guards: a file whose content was never stored (binary or
	// oversized) reports the same blob id forever. Reusing it would make the
	// file permanently unindexable, so content availability must be required.
	entries := []repoTreeItem{{Path: "logo.png", BlobSHA: "sha-png"}}
	previous := map[string]prevRepoFile{
		"logo.png": {BlobSHA: "sha-png", HasContent: false},
	}

	unchanged, paths := classifyTree(entries, previous)
	if unchanged["logo.png"] || len(paths) != 0 {
		t.Fatalf("metadata-only file must be re-attempted, got unchanged=%v paths=%v", unchanged, paths)
	}
}

func TestClassifyTreeHandlesEmptyPreviousIndex(t *testing.T) {
	entries := []repoTreeItem{{Path: "a.go", BlobSHA: "sha-a"}, {Path: "b.go", BlobSHA: "sha-b"}}
	unchanged, paths := classifyTree(entries, map[string]prevRepoFile{})
	if len(unchanged) != 0 || len(paths) != 0 {
		t.Fatalf("first sync must treat every file as new, got unchanged=%v paths=%v", unchanged, paths)
	}
}
