import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\training\chunker.go")
t = p.read_text(encoding="utf-8")
old="""type TextChunk struct {
\tText      string
\tIndex     int
\tStartChar int
\tEndChar   int
\t// TokenCount is the estimated size used by the size targets.
\tTokenCount int
\tChunkHash  string
\t// SectionPath is the heading trail this chunk came from ("RAG > Chunking").
\t//
\t// Empty when the source had no headings (a plain transcript) or the chunk is
\t// preamble. Stored so a later retrieval step can filter or cite by section \u2014
\t// the design doc's metadata-filter stage runs before retrieval, and it can
\t// only do that if the metadata is kept at ingest time.
\tSectionPath string
}"""
new="""type TextChunk struct {
\tText      string
\tIndex     int
\tStartChar int
\tEndChar   int
\t// TokenCount is the estimated size used by the size targets.
\tTokenCount int
\tChunkHash  string
\t// SectionPath is the heading trail this chunk came from ("RAG > Chunking").
\t//
\t// Empty when the source had no headings (a plain transcript) or the chunk is
\t// preamble. Stored so a later retrieval step can filter or cite by section \u2014
\t// the design doc's metadata-filter stage runs before retrieval, and it can
\t// only do that if the metadata is kept at ingest time.
\tSectionPath string
\tParentIndex int
\tParentID string
}"""
if old in t:
    t=t.replace(old,new)
    print("patched TextChunk")
else:
    print("old not found")
if '"context"' not in t:
    t=t.replace('import (','import (\n\t"context"')
    print("added context")
if "type ParentChunk" not in t:
    add=open(r"C:\Users\sharm\AI_AVENGERS-1\parent_add.go","encoding="utf-8").read()
    t=t.rstrip()+"\n"+add+"\n"
    print("appended")
else:
    print("already has ParentChunk")
p.write_text(t,encoding="utf-8")
print("done",len(t.splitlines()))
