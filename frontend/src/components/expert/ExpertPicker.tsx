import { useState } from 'react'
import type { ProjectExpert } from '@/types/project'
import { ExpertBadge } from './ExpertBadge'

/**
 * Multi-select panel for choosing which experts to send a message to.
 * Source: FRONTEND_SYSTEM_DESIGN.md section 1 ("Experts: [SD \u2713] [DB \u2713] [Backend] [DSA]"
 * row above the message input) and section 3 (components/expert/ExpertPicker.tsx).
 *
 * WHY selection state is controlled (lifted to the parent via
 * selectedIds + onChange) rather than owned internally: MessageInput
 * (Phase 3) needs the currently-selected expert ids at send time to
 * populate SendMessageOptions.expertIds. An uncontrolled ExpertPicker
 * would have no way to expose that without an imperative ref API,
 * which is more brittle than just lifting the state up - this is the
 * same "local vs lifted state" principle from Transcripts/Frontend/
 * reactjs2.md's counter-reset example (state lives at the lowest
 * common ancestor that needs it, which here is the chat input area,
 * not inside the picker itself).
 */
export interface ExpertPickerProps {
  experts: ProjectExpert[]
  selectedIds: Set<string>
  orderedIds?: string[]
  onChange: (selectedIds: Set<string>, newOrderedIds?: string[]) => void
  onReorder?: (newOrderedIds: string[]) => void
}

export function ExpertPicker({
  experts,
  selectedIds,
  orderedIds,
  onChange,
  onReorder,
}: ExpertPickerProps) {
  // Compute effective ordered array of selected expert IDs
  const activeOrder = orderedIds && orderedIds.length > 0
    ? orderedIds.filter((id) => selectedIds.has(id))
    : Array.from(selectedIds)

  function toggle(expertId: string) {
    const nextSet = new Set(selectedIds)
    let nextOrder = [...activeOrder]

    if (nextSet.has(expertId)) {
      nextSet.delete(expertId)
      nextOrder = nextOrder.filter((id) => id !== expertId)
    } else {
      nextSet.add(expertId)
      nextOrder.push(expertId)
    }

    onChange(nextSet, nextOrder)
    if (onReorder) onReorder(nextOrder)
  }

  function moveExpert(index: number, direction: 'left' | 'right') {
    const targetIndex = direction === 'left' ? index - 1 : index + 1
    if (targetIndex < 0 || targetIndex >= activeOrder.length) return
    const nextOrder = [...activeOrder]
    const temp = nextOrder[index]!
    nextOrder[index] = nextOrder[targetIndex]!
    nextOrder[targetIndex] = temp
    if (onReorder) onReorder(nextOrder)
  }

  const expertMap = new Map(experts.map((e) => [e.expertId, e]))

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-xs text-text-secondary">Experts:</span>
        {experts.map((expert) => {
          const isSelected = selectedIds.has(expert.expertId)
          const orderIdx = activeOrder.indexOf(expert.expertId)
          return (
            <ExpertBadge
              key={expert.expertId}
              expert={expert}
              isSelected={isSelected}
              orderIndex={isSelected && orderIdx !== -1 ? orderIdx + 1 : undefined}
              onClick={() => toggle(expert.expertId)}
            />
          )
        })}
      </div>

      {/* Interactive Reordering Strip when 2+ experts selected */}
      {activeOrder.length > 1 && (
        <div className="flex flex-wrap items-center gap-2 rounded-lg border border-brand/30 bg-brand/5 p-2 text-xs">
          <span className="font-semibold text-brand">Execution Order (Sequence):</span>
          <div className="flex flex-wrap items-center gap-1.5">
            {activeOrder.map((id, idx) => {
              const exp = expertMap.get(id)
              if (!exp) return null
              return (
                <div
                  key={id}
                  className="flex items-center gap-1 rounded border border-surface-border bg-surface-overlay px-2 py-1 text-text-primary"
                >
                  <span className="font-bold text-brand">#{idx + 1}</span>
                  <span>{exp.expertName}</span>
                  <div className="ml-1 flex items-center gap-0.5">
                    <button
                      type="button"
                      disabled={idx === 0}
                      onClick={() => moveExpert(idx, 'left')}
                      className="rounded px-1 text-[10px] font-bold text-text-secondary hover:bg-brand/20 hover:text-brand disabled:opacity-30"
                      title="Move earlier in execution sequence"
                    >
                      ◀
                    </button>
                    <button
                      type="button"
                      disabled={idx === activeOrder.length - 1}
                      onClick={() => moveExpert(idx, 'right')}
                      className="rounded px-1 text-[10px] font-bold text-text-secondary hover:bg-brand/20 hover:text-brand disabled:opacity-30"
                      title="Move later in execution sequence"
                    >
                      ▶
                    </button>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}

// Convenience hook for the common case (own the Set locally when no
// parent needs to read/control it directly, e.g. a standalone demo or
// a page that only needs the picker for its side effect of updating a
// ref). Kept separate from the controlled component above rather than
// making selectedIds/onChange optional on ExpertPicker itself, which
// would make the controlled-vs-uncontrolled contract ambiguous - the
// exact anti-pattern the useState video in TyeScript Simplified.md's
// React+TS section warns against.
export function useExpertSelection(initial: string[] = []) {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set(initial))
  return { selectedIds, setSelectedIds }
}
