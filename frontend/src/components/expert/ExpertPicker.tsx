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
  onChange: (selectedIds: Set<string>) => void
}

export function ExpertPicker({ experts, selectedIds, onChange }: ExpertPickerProps) {
  function toggle(expertId: string) {
    const next = new Set(selectedIds)
    if (next.has(expertId)) {
      next.delete(expertId)
    } else {
      next.add(expertId)
    }
    onChange(next)
  }

  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-xs text-text-secondary">Experts:</span>
      {experts.map((expert) => (
        <ExpertBadge
          key={expert.expertId}
          expert={expert}
          isSelected={selectedIds.has(expert.expertId)}
          onClick={() => toggle(expert.expertId)}
        />
      ))}
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
