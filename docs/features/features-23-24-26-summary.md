# Features #23, #24, #26 Implementation Summary

## Feature #23: Citation Source Preview ✅ COMPLETE

**Status**: Fully implemented and tested

### What Was Done

1. **Backend Changes**
   - Added `SourceName` and `ChunkIndex` fields to `Citation` struct
   - Added `SourceFile` and `ChunkIndex` fields to `CourseChunk` struct
   - Updated `getCourseChunks()` to fetch `source_file` and `chunk_index` from database
   - Updated `extractCitations()` to populate new fields

2. **Frontend Changes**
   - Added `sourceName?` and `chunkIndex?` to `Citation` interface
   - Enhanced `CitationChip` modal to display source info
   - Shows "From: [transcript_name.txt] • Chunk #42"
   - Graceful fallback for missing data

3. **Files Modified**
   - `backend-go/internal/chinawall/enforcer.go`
   - `backend-go/internal/context/assembler.go`
   - `frontend/src/types/expert.ts`
   - `frontend/src/components/chat/CitationChip.tsx`

4. **Documentation**
   - Created `docs/features/feature-23-citation-source-preview.md`

---

## Feature #24: Project Tech Stack Tags ⏳ PENDING

**Status**: Not yet implemented

### Problem

- `tech_stack` JSONB field exists in `projects` table
- No UI to add/edit/remove tech stack tags
- Users can't specify "React", "Node.js", "PostgreSQL", etc.

### Implementation Plan

#### Backend (Verify First)

1. **Check if backend already returns tech_stack**
   - Find project handler (likely in `backend-go/internal/project/`)
   - Verify GET `/api/v1/projects/:id` returns `tech_stack` field
   - Verify PATCH `/api/v1/projects/:id` accepts `tech_stack` field

2. **If backend doesn't return tech_stack**
   - Add `TechStack []string` to project struct
   - Update SELECT query to include `tech_stack`
   - Update UPDATE query to accept `tech_stack`
   - Use `json.Marshal/Unmarshal` for JSONB conversion

#### Frontend

1. **Update Project Type** (`frontend/src/types/project.ts`)
   ```typescript
   export interface Project {
     // ... existing fields
     techStack?: string[]  // NEW: array of tech stack tags
   }
   ```

2. **Create TechStackEditor Component** (`frontend/src/components/project/TechStackEditor.tsx`)
   ```typescript
   interface TechStackEditorProps {
     techStack: string[]
     onChange: (techStack: string[]) => void
     readOnly?: boolean
   }
   
   // Features:
   // - Display tags as chips with × remove button
   // - Input field to add new tags (Enter to add)
   // - Prevent duplicate tags
   // - Trim whitespace
   // - Max 20 tags
   ```

3. **Add to ProjectPage** (`frontend/src/pages/projects/ProjectPage.tsx`)
   ```typescript
   // Add section after "Architecture Type"
   <div className="space-y-2">
     <h3 className="text-sm font-semibold">Tech Stack</h3>
     <TechStackEditor
       techStack={project.techStack || []}
       onChange={handleTechStackChange}
       readOnly={!canEdit}
     />
   </div>
   ```

4. **Update EditProjectModal** (`frontend/src/components/project/EditProjectModal.tsx`)
   - Add TechStackEditor to modal form
   - Include techStack in form state
   - Send techStack in PATCH request

5. **Update API** (`frontend/src/api/projects.ts`)
   ```typescript
   export interface UpdateProjectRequest {
     // ... existing fields
     techStack?: string[]
   }
   ```

### UI Design

```
Tech Stack
┌─────────────────────────────────────────────┐
│ [React ×] [Node.js ×] [PostgreSQL ×]        │
│ [+ Add tag]                                  │
└─────────────────────────────────────────────┘
```

### Testing Checklist

- [ ] Add new tag (Enter key)
- [ ] Remove tag (× button)
- [ ] Prevent duplicate tags
- [ ] Trim whitespace
- [ ] Max 20 tags limit
- [ ] Save to backend
- [ ] Reload page shows saved tags
- [ ] Empty tech stack shows empty state

---

## Feature #26: Admin Expert Capability Table View ⏳ PENDING

**Status**: Not yet implemented

### Problem

- Expert topics/capabilities shown as simple list
- No structured table view
- Depth levels not clearly visible
- Chunk counts and complexity not highlighted

### Implementation Plan

#### Backend (Verify First)

1. **Check if endpoint exists**
   - Verify GET `/api/v1/experts/:id/topics` or `/api/v1/admin/experts/:id/topics`
   - Should return array of `ExpertTopic` objects
   - Fields: `topic`, `depthLevel`, `chunkCount`, `complexityCeiling`

2. **If endpoint doesn't exist**
   - Create handler in `backend-go/internal/expert/` or `backend-go/internal/admin/`
   - Query `expert_capabilities` table
   - Return JSON array

#### Frontend

1. **Verify ExpertTopic Type** (`frontend/src/types/expert.ts`)
   ```typescript
   export interface ExpertTopic {
     topic: string
     depthLevel: 1 | 2 | 3 | 4 | 5
     chunkCount: number
     complexityCeiling: 'basic' | 'intermediate' | 'advanced' | 'expert' | 'master'
   }
   ```

2. **Create ExpertCapabilitiesTable Component** (`frontend/src/components/admin/ExpertCapabilitiesTable.tsx`)
   ```typescript
   interface ExpertCapabilitiesTableProps {
     expertId: string
   }
   
   // Features:
   // - Fetch topics from API
   // - Display in table format
   // - Sortable columns (topic, depth, chunks)
   // - Visual depth indicator (1-5 stars or progress bar)
   // - Color-coded complexity badges
   // - Loading state
   // - Empty state
   ```

3. **Add to AdminExperts** (`frontend/src/pages/admin/AdminExperts.tsx`)
   ```typescript
   // Add expandable section to expert cards
   <Collapsible>
     <CollapsibleTrigger>
       <Button variant="ghost">View Capabilities</Button>
     </CollapsibleTrigger>
     <CollapsibleContent>
       <ExpertCapabilitiesTable expertId={expert.id} />
     </CollapsibleContent>
   </Collapsible>
   ```

### UI Design

```
┌─────────────────────────────────────────────────────────────┐
│ Topic              │ Depth │ Chunks │ Complexity           │
├─────────────────────────────────────────────────────────────┤
│ React Hooks        │ ★★★★☆ │ 23     │ [Advanced]           │
│ State Management   │ ★★★★★ │ 45     │ [Expert]             │
│ API Integration    │ ★★★☆☆ │ 18     │ [Intermediate]       │
│ Performance        │ ★★★★☆ │ 31     │ [Advanced]           │
└─────────────────────────────────────────────────────────────┘
```

### Visual Indicators

1. **Depth Level** (1-5 stars)
   - 1: ★☆☆☆☆ (Basic)
   - 2: ★★☆☆☆ (Beginner)
   - 3: ★★★☆☆ (Intermediate)
   - 4: ★★★★☆ (Advanced)
   - 5: ★★★★★ (Expert)

2. **Complexity Badges**
   - Basic: Gray
   - Intermediate: Blue
   - Advanced: Purple
   - Expert: Orange
   - Master: Red

3. **Chunk Count**
   - Color-coded by count
   - < 10: Gray (low coverage)
   - 10-30: Blue (good coverage)
   - > 30: Green (excellent coverage)

### Testing Checklist

- [ ] Fetch topics from API
- [ ] Display in table format
- [ ] Sort by topic (alphabetical)
- [ ] Sort by depth level (1-5)
- [ ] Sort by chunk count (ascending/descending)
- [ ] Visual depth indicators render correctly
- [ ] Complexity badges color-coded
- [ ] Loading state shows spinner
- [ ] Empty state shows "No capabilities found"
- [ ] Expandable section works

---

## Next Steps

### For Feature #24 (Tech Stack Tags)

1. **Backend Investigation**
   ```bash
   # Find project handler
   find backend-go/internal -name "*project*" -type f
   
   # Search for tech_stack in code
   grep -r "tech_stack" backend-go/internal/
   ```

2. **Frontend Implementation**
   - Start with TechStackEditor component
   - Add to ProjectPage
   - Update EditProjectModal
   - Test end-to-end

### For Feature #26 (Expert Capabilities Table)

1. **Backend Investigation**
   ```bash
   # Find expert topics endpoint
   grep -r "expert_capabilities" backend-go/internal/
   grep -r "/topics" backend-go/internal/
   ```

2. **Frontend Implementation**
   - Create ExpertCapabilitiesTable component
   - Add to AdminExperts page
   - Implement sorting
   - Add visual indicators
   - Test with real data

---

## Files to Create/Modify

### Feature #24
- `frontend/src/types/project.ts` (modify)
- `frontend/src/components/project/TechStackEditor.tsx` (create)
- `frontend/src/pages/projects/ProjectPage.tsx` (modify)
- `frontend/src/components/project/EditProjectModal.tsx` (modify)
- `frontend/src/api/projects.ts` (modify)
- Backend project handler (find and modify)

### Feature #26
- `frontend/src/components/admin/ExpertCapabilitiesTable.tsx` (create)
- `frontend/src/pages/admin/AdminExperts.tsx` (modify)
- `frontend/src/api/experts.ts` (verify/modify)
- Backend expert topics endpoint (verify exists)

---

## Estimated Time

- **Feature #24**: 2-3 hours
  - Backend verification: 30 min
  - TechStackEditor component: 1 hour
  - Integration: 1 hour
  - Testing: 30 min

- **Feature #26**: 3-4 hours
  - Backend verification: 30 min
  - ExpertCapabilitiesTable component: 2 hours
  - Visual indicators: 1 hour
  - Integration: 30 min
  - Testing: 30 min

---

## Success Criteria

### Feature #24
- ✅ Users can add tech stack tags
- ✅ Users can remove tech stack tags
- ✅ Tags persist to database
- ✅ Tags display on project page
- ✅ Tags editable in modal

### Feature #26
- ✅ Expert capabilities shown in table
- ✅ Depth levels visualized (stars)
- ✅ Complexity badges color-coded
- ✅ Sortable columns
- ✅ Expandable section in admin panel

---

**Feature #23 Complete** ✅  
**Feature #24 Pending** ⏳  
**Feature #26 Pending** ⏳
