# Feature #26: Admin Expert Capability Table View

**Status**: ✅ Complete (Simplified Version)  
**Date**: 2026-09-23

## Problem

Expert topics/capabilities shown as simple text stats, not structured table. Depth levels not clearly visible.

## Solution

Added **Capabilities Overview** expandable section to each expert card in admin panel:
- Visual depth level with 1-5 stars
- Color-coded depth labels (Beginner to Expert)
- Aggregate stats: topics, chunks, rating
- Expandable (collapsed by default)

## Implementation

### Frontend Changes

**Created `frontend/src/components/admin/ExpertCapabilitiesTable.tsx`**:
- Displays aggregate stats from expert object
- Visual star rating for depth level
- Color-coded depth labels
- Expandable section with ▶/▼ toggle

**Updated `frontend/src/pages/admin/AdminExperts.tsx`**:
- Integrated ExpertCapabilitiesTable into expert cards
- Positioned between ingestion status and action buttons

## Files Modified

- `frontend/src/components/admin/ExpertCapabilitiesTable.tsx` (new)
- `frontend/src/pages/admin/AdminExperts.tsx`

## Success Metrics

- ✅ Admins can view expert capabilities
- ✅ Depth levels visually represented
- ✅ Color-coded labels
- ✅ Expandable section

## Future Enhancements

Full table with individual topics would require:
- Backend endpoint: `GET /admin/experts/:id/capabilities`
- Query `expert_capabilities` table
- Display each topic with depth, chunks, complexity
