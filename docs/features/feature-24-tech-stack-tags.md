# Feature #24: Project Tech Stack Tags

**Status**: ✅ Complete  
**Date**: 2026-09-23

## Problem

The `tech_stack` JSONB field exists in the `projects` table, but there's no UI to add or edit tech stack tags like "React", "Node.js", "PostgreSQL".

## Solution

Added a **Tech Stack** section to ProjectPage with tag editor:
- Display tags as chips with × remove button
- Input field + "Add Tag" button
- Auto-save on add/remove
- Tags stored as JSONB object: `{"React": true, "Node.js": true}`

## Implementation

### Backend Changes

**Updated `backend-go/internal/project/service.go`**:
1. Added `tech_stack` to SELECT queries (List, GetByID)
2. Added `TechStack *map[string]interface{}` to UpdateProjectRequest
3. Added tech_stack update logic in Update() method
4. Scan tech_stack into Project.TechStack field

### Frontend Changes

**1. Updated `frontend/src/types/project.ts`**:
- Added `techStack?: string[] | Record<string, any>` to Project interface

**2. Updated `frontend/src/api/projects.ts`**:
- Added `tech_stack?: Record<string, any>` to UpdateProjectRequest

**3. Created `frontend/src/components/project/TechStackEditor.tsx`**:
- Displays tags as chips with × remove button
- Input field + Add button for new tags
- Auto-save on add/remove
- Format conversion: array ↔ object

**4. Updated `frontend/src/pages/projects/ProjectPage.tsx`**:
- Integrated TechStackEditor between Experts and Chats sections

## Files Modified

- `backend-go/internal/project/service.go`
- `frontend/src/types/project.ts`
- `frontend/src/api/projects.ts`
- `frontend/src/components/project/TechStackEditor.tsx` (new)
- `frontend/src/pages/projects/ProjectPage.tsx`

## Success Metrics

- ✅ Users can add tech stack tags
- ✅ Users can remove tech stack tags
- ✅ Tags persist across reloads
- ✅ Auto-save works
- ✅ Duplicate prevention
- ✅ Error handling
