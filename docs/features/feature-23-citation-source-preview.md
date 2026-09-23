# Feature #23: Citation Source Preview

**Status**: ✅ Complete (2026-09-23)

## Problem

When clicking `[Source N]` citation chips in expert responses, the modal showed:
- ✅ Chunk text (200 chars preview)
- ✅ Relevance score
- ✅ Chunk ID (for debugging)
- ❌ **Source transcript name** (missing)
- ❌ **Chunk position/number** (missing)

Users couldn't tell which transcript a citation came from. Was it from "React Hooks Part 1" or "State Management Advanced"? No way to know.

## Solution

Added source transcript name and chunk index to citations throughout the entire stack:

### Backend Changes

1. **Citation struct** (`backend-go/internal/chinawall/enforcer.go`)
   - Added `SourceName string` field
   - Added `ChunkIndex int` field
   - Both fields are `omitempty` in JSON (backward compatible)

2. **CourseChunk struct** (`backend-go/internal/chinawall/enforcer.go`)
   - Added `SourceFile string` field
   - Added `ChunkIndex int` field
   - Populated from database during hybrid search

3. **Hybrid search** (`backend-go/internal/context/assembler.go`)
   - Updated vector search query: `SELECT ... COALESCE(source_file,''), chunk_index`
   - Updated keyword search query: same fields
   - Updated `rawChunk` struct to include new fields
   - Updated both reranked and fallback paths to populate fields

4. **Citation extraction** (`backend-go/internal/chinawall/enforcer.go`)
   - `extractCitations()` now populates `SourceName` and `ChunkIndex`
   - Pulls from `chunk.SourceFile` and `chunk.ChunkIndex`

### Frontend Changes

1. **Citation type** (`frontend/src/types/expert.ts`)
   - Added `sourceName?: string` (optional, empty string fallback)
   - Added `chunkIndex?: number` (optional, 0-based)

2. **CitationChip modal** (`frontend/src/components/chat/CitationChip.tsx`)
   - Added "From:" section above chunk text
   - Displays: `From: [transcript_name.txt] • Chunk #42`
   - Graceful fallback: "Unknown Source" when `sourceName` is empty
   - Hides "Chunk #N" when `chunkIndex` is undefined
   - Converts 0-based index to 1-based display (index 41 → "Chunk #42")

## User Experience

### Before
```
Source 1
─────────────────────────────────
[Chunk text preview...]

Relevance: 87.3%
Chunk ID: a3f2b1c4
```

### After
```
Source 1
─────────────────────────────────
From: react_hooks_part1.txt • Chunk #42

[Chunk text preview...]

Relevance: 87.3%
Chunk ID: a3f2b1c4
```

## Technical Details

### Database Schema

The `course_chunks` table already had these fields:
- `source_file VARCHAR(500)` - Original transcript filename
- `chunk_index INTEGER` - 0-based position in transcript

No migration needed! We just weren't fetching or displaying them.

### Backward Compatibility

- **Backend**: `omitempty` JSON tags mean old clients ignore new fields
- **Frontend**: Optional fields (`sourceName?`, `chunkIndex?`) mean old data works
- **Fallback**: Empty `sourceName` displays as "Unknown Source"
- **Graceful**: Missing `chunkIndex` hides the "Chunk #N" part

### Edge Cases Handled

1. **Legacy chunks** (no `source_file`)
   - Backend: `COALESCE(source_file,'')` returns empty string
   - Frontend: Displays "Unknown Source"

2. **Repo chunks** (from connected repositories)
   - Currently don't have `source_file` populated
   - Will show "Unknown Source" until repo sync adds filenames

3. **Missing chunk index**
   - Frontend checks `chunkIndex !== undefined`
   - Hides "Chunk #N" part if missing

## Testing

### Manual Testing

1. Ask an expert a question that triggers citations
2. Click any `[Source N]` chip
3. Verify modal shows:
   - "From: [transcript_name.txt]" (or "Unknown Source")
   - "Chunk #42" (if available)
   - Chunk text preview
   - Relevance score
   - Chunk ID

### Test Cases

- ✅ Citation with source_file and chunk_index
- ✅ Citation with empty source_file (shows "Unknown Source")
- ✅ Citation with missing chunk_index (hides "Chunk #N")
- ✅ Multiple citations from same transcript
- ✅ Multiple citations from different transcripts

## Files Modified

### Backend
- `backend-go/internal/chinawall/enforcer.go` (Citation struct, CourseChunk struct, extractCitations)
- `backend-go/internal/context/assembler.go` (getCourseChunks hybrid search)

### Frontend
- `frontend/src/types/expert.ts` (Citation interface)
- `frontend/src/components/chat/CitationChip.tsx` (modal UI)

## Future Enhancements

1. **Clickable source names**: Link to full transcript view
2. **Chunk navigation**: "Previous/Next chunk" buttons in modal
3. **Transcript preview**: Show surrounding chunks for context
4. **Source filtering**: "Show all citations from this transcript"
5. **Repo chunk support**: Add source_file to repo_chunks table

## Related Issues

This feature addresses the citation preview gap identified in the bug list:
- Users couldn't identify which transcript a citation came from
- No way to locate the citation in the original training material
- Citation modal was missing critical context

## Lessons Learned

1. **Database fields existed**: We had `source_file` and `chunk_index` all along, just never used them
2. **Additive changes**: No breaking changes, all fields optional
3. **Graceful degradation**: Old data still works, new data enhances UX
4. **Full-stack coordination**: Backend, database, and frontend all needed updates

---

**Feature #23 Complete** ✅

Citations now show source transcript names and chunk positions, making it easy for users to understand where expert knowledge comes from.
