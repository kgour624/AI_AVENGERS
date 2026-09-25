package repo

import (
	"go/parser"
	"go/token"
	"path"
	"regexp"
	"strings"
)

// This file turns a file's source text into resolved dependency edges.
//
// DESIGN: extraction and resolution are separate steps on purpose.
//
//   extractImports  — language-aware, text in, raw specifiers out. It never
//                     touches the repository, so it is trivially unit-testable.
//   repoPathIndex   — the repository's known paths, and the rules that turn a
//                     specifier into real file paths.
//
// Splitting them keeps "what does this file say it imports" (a parsing problem)
// apart from "which stored file does that mean" (a repository problem), which
// is where essentially all of the imprecision lives.

// Edge kinds. Stored on the row so a future extractor can be added, or one
// extractor's output invalidated, without guessing what wrote an edge.
const (
	EdgeKindImport  = "import"
	EdgeKindRequire = "require"
	EdgeKindInclude = "include"
)

// maxImportsPerFile bounds how many outgoing edges one file can contribute.
// A generated or minified file can contain thousands of require() calls; letting
// it through would let one file dominate the graph and the expansion budget.
const maxImportsPerFile = 200

// importRef is one specifier as written in a source file.
type importRef struct {
	Spec string
	Kind string
}

// canExtractImports reports whether edges can be derived for a language. An
// unsupported language yields no edges rather than guessed ones.
func canExtractImports(language string) bool {
	switch language {
	case "go", "typescript", "javascript", "python":
		return true
	}
	return false
}

var (
	// Matches all three JS/TS forms: static import, dynamic import(), require().
	jsImportRe = regexp.MustCompile(`(?m)(?:import\s+(?:[^'"()]*?\s+from\s+)?|require\s*\(\s*|import\s*\(\s*)['"]([^'"]+)['"]`)
	// Python: "from a.b import c" / "from . import c" / "from ..pkg import c".
	pyFromRe = regexp.MustCompile(`(?m)^\s*from\s+([.\w]+)\s+import\s+`)
	// Python: "import a.b.c".
	pyImportRe = regexp.MustCompile(`(?m)^\s*import\s+([\w.]+)`)
)

// extractImports returns the import specifiers written in the file. It is
// deliberately tolerant: an unparseable file yields no specs rather than an
// error that would fail a whole repository sync.
func extractImports(language, content string) []importRef {
	switch language {
	case "go":
		return extractGoImports(content)
	case "typescript", "javascript":
		return extractJSImports(content)
	case "python":
		return extractPythonImports(content)
	}
	return nil
}

func extractGoImports(content string) []importRef {
	// ImportsOnly parses just the import block. WHY not a full parse: a file
	// that is mid-edit or uses a build tag can fail to type-check while its
	// import block is perfectly readable, and edges are all we need.
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "source.go", content, parser.ImportsOnly)
	if err != nil {
		return nil
	}

	var refs []importRef
	for _, imp := range file.Imports {
		spec := strings.Trim(imp.Path.Value, `"`)
		if spec == "" || spec == "C" {
			continue // blank, malformed, or cgo — none resolve to a repo file
		}
		refs = append(refs, importRef{Spec: spec, Kind: EdgeKindImport})
		if len(refs) >= maxImportsPerFile {
			break
		}
	}
	return refs
}

func extractJSImports(content string) []importRef {
	var refs []importRef
	seen := map[string]bool{}

	for _, match := range jsImportRe.FindAllStringSubmatch(content, maxImportsPerFile) {
		spec := strings.TrimSpace(match[1])
		if spec == "" || seen[spec] {
			continue
		}
		seen[spec] = true
		// The kind is decided from the matched text: the same specifier can be
		// reached through require() or a static import, and knowing which tells
		// a reader whether the edge is a bundler-visible dependency.
		kind := EdgeKindImport
		if strings.Contains(match[0], "require") {
			kind = EdgeKindRequire
		}
		refs = append(refs, importRef{Spec: spec, Kind: kind})
	}
	return refs
}

func extractPythonImports(content string) []importRef {
	var refs []importRef
	seen := map[string]bool{}

	add := func(spec, kind string) {
		spec = strings.TrimSpace(spec)
		if spec == "" || seen[spec] {
			return
		}
		seen[spec] = true
		refs = append(refs, importRef{Spec: spec, Kind: kind})
	}

	for _, match := range pyFromRe.FindAllStringSubmatch(content, maxImportsPerFile) {
		add(match[1], EdgeKindImport)
	}
	for _, match := range pyImportRe.FindAllStringSubmatch(content, maxImportsPerFile) {
		// "import a, b" — the regex only captures the first name; splitting
		// here would guess at commas inside a single-line statement, so the
		// remainder is intentionally left to the other import statements.
		add(match[1], EdgeKindImport)
	}
	return refs
}

// repoPathIndex answers "which stored file does this specifier name?".
type repoPathIndex struct {
	isFile   map[string]bool
	dirFiles map[string][]string
}

func newRepoPathIndex(paths []string) *repoPathIndex {
	idx := &repoPathIndex{
		isFile:   make(map[string]bool, len(paths)),
		dirFiles: make(map[string][]string, len(paths)),
	}
	for _, p := range paths {
		idx.isFile[p] = true
		dir := path.Dir(p)
		idx.dirFiles[dir] = append(idx.dirFiles[dir], p)
	}
	return idx
}

// resolve turns one specifier from srcPath into the files it names.
//
// It returns every match because a Go import names a package, not a file, and
// because a relative JS import can resolve to more than one candidate. Emitting
// all of them is the honest file-level answer; the caller de-duplicates.
func (idx *repoPathIndex) resolve(srcPath, language string, ref importRef) []string {
	switch language {
	case "go":
		return idx.resolveBySuffixDir(ref.Spec, ".go")
	case "typescript", "javascript":
		return idx.resolveJS(srcPath, ref.Spec)
	case "python":
		return idx.resolvePython(srcPath, ref.Spec)
	}
	return nil
}

// resolveBySuffixDir resolves a module path by matching its longest suffix
// that names a directory in the repository.
//
// WHY suffix matching: the repository's module prefix is not known here (it
// lives in go.mod / tsconfig / setup.py, which are not necessarily ingested),
// so "ai_avengers/backend/internal/repo" is matched by trying
// "ai_avengers/backend/internal/repo", then "backend/internal/repo", then
// "internal/repo" — the last one exists, so its files are the target.
func (idx *repoPathIndex) resolveBySuffixDir(spec, onlyExt string) []string {
	spec = strings.Trim(spec, "/")
	if spec == "" {
		return nil
	}

	parts := strings.Split(spec, "/")
	for i := 0; i < len(parts); i++ {
		candidate := strings.Join(parts[i:], "/")
		files, ok := idx.dirFiles[candidate]
		if !ok {
			continue
		}
		if onlyExt == "" {
			return files
		}
		var filtered []string
		for _, f := range files {
			if strings.HasSuffix(f, onlyExt) {
				filtered = append(filtered, f)
			}
		}
		if len(filtered) > 0 {
			return filtered
		}
	}

	// A specifier can also name a file directly ("pkg/mod" -> pkg/mod.go).
	return idx.resolveRelativeFile(path.Dir(spec), path.Base(spec))
}

// resolveJS resolves a JS/TS specifier, relative or not.
func (idx *repoPathIndex) resolveJS(srcPath, spec string) []string {
	if strings.HasPrefix(spec, ".") {
		target := path.Clean(path.Join(path.Dir(srcPath), spec))
		return idx.resolveRelativeFile(path.Dir(target), path.Base(target))
	}
	// Bare specifiers are npm packages, but a repo using path aliases maps
	// them onto its own tree — so try to match a suffix directory first and
	// fall back to nothing (a real node_modules dependency has no stored file).
	return idx.resolveBySuffixDir(spec, "")
}

// resolvePython resolves a Python import, including leading-dot relative form.
func (idx *repoPathIndex) resolvePython(srcPath, spec string) []string {
	if strings.HasPrefix(spec, ".") {
		// "from .x import y" -> x is relative to this file's package directory.
		dots := len(spec) - len(strings.TrimLeft(spec, "."))
		rest := strings.TrimLeft(spec, ".")
		base := path.Dir(srcPath)
		for i := 1; i < dots; i++ {
			base = path.Dir(base)
		}
		if rest == "" {
			// "from . import x" names the current package.
			if files, ok := idx.dirFiles[base]; ok {
				return files
			}
			return nil
		}
		target := path.Join(base, strings.ReplaceAll(rest, ".", "/"))
		return idx.resolveBySuffixDir(target, ".py")
	}
	return idx.resolveBySuffixDir(strings.ReplaceAll(spec, ".", "/"), ".py")
}

// resolveRelativeFile finds files named by a directory plus a base name, trying
// each language extension and an /index file, which covers how each of these
// ecosystems actually resolves a partial path.
func (idx *repoPathIndex) resolveRelativeFile(dir, base string) []string {
	var out []string
	candidates := []string{
		path.Join(dir, base),
		path.Join(dir, base+".go"),
		path.Join(dir, base+".py"),
		path.Join(dir, base+".ts"),
		path.Join(dir, base+".tsx"),
		path.Join(dir, base+".js"),
		path.Join(dir, base+".jsx"),
		path.Join(dir, base, "index.ts"),
		path.Join(dir, base, "index.tsx"),
		path.Join(dir, base, "index.js"),
		path.Join(dir, base, "index.jsx"),
		path.Join(dir, base, "__init__.py"),
	}
	for _, candidate := range candidates {
		if idx.isFile[candidate] {
			out = append(out, candidate)
		}
	}
	return out
}

// buildRepoEdges derives every edge reachable from the fetched files.
//
// Only files whose content was stored can be parsed; a metadata-only file
// contributes no edges and is linked to by nothing except by name.
func buildRepoEdges(contents map[string]string, languages map[string]string) []repoEdge {
	paths := make([]string, 0, len(contents))
	for p := range contents {
		paths = append(paths, p)
	}
	index := newRepoPathIndex(paths)

	var edges []repoEdge
	for srcPath, content := range contents {
		language := languages[srcPath]
		if !canExtractImports(language) {
			continue
		}
		for _, ref := range extractImports(language, content) {
			for _, dst := range index.resolve(srcPath, language, ref) {
				if dst == srcPath {
					continue // a self-import is a cycle of length 1, always noise
				}
				edges = append(edges, repoEdge{Src: srcPath, Dst: dst, Kind: ref.Kind})
			}
		}
	}
	return edges
}

// repoEdge is one directed dependency between two stored files.
type repoEdge struct {
	Src  string
	Dst  string
	Kind string
}
