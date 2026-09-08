import { useState } from 'react'
import type { Citation } from '@/types/expert'
import { Tooltip } from '@/components/ui/Tooltip'
import { Modal } from '@/components/ui/Modal'

export interface CitationChipProps {
  citation: Citation
  /** 1-based index for human-readable label: Source 1, Source 2, etc. */
  index: number
}

/**
 * Renders a citation as a human-readable [Source N] chip.
 * Hover: shows chunk text preview in tooltip.
 * Click: opens modal with full chunk text + relevance score.
 *
 * WHY [Source N] not [CHUNK_hash]:
 *   UUID hashes are machine identifiers — meaningless to students.
 *   "Source 1" is immediately understandable: "this claim came from
 *   the first source the expert cited."
 *   The full chunkId is still shown in the modal for admin/debug use.
 */
export function CitationChip({ citation, index }: CitationChipProps) {
  const [isModalOpen, setIsModalOpen] = useState(false)

  // Truncate tooltip preview to first 120 chars
  const preview = citation.text.length > 120
    ? citation.text.slice(0, 120) + '...'
    : citation.text

  return (
    <>
      <Tooltip content={preview}>
        <button
          type="button"
          onClick={() => setIsModalOpen(true)}
          className="rounded border border-glass-border bg-surface-overlay px-1.5 py-0.5 text-xs font-medium text-glow-cyan transition-[border-color,box-shadow] duration-150 ease-arc hover:border-glow-cyan/50 hover:bg-glow-cyan/10 hover:shadow-[0_0_12px_-4px_var(--glow-cyan)]"
          aria-label={`View source ${index}`}
        >
          Source {index}
        </button>
      </Tooltip>

      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)}>
        <h3 className="mb-3 text-sm font-semibold text-glow-cyan">Source {index}</h3>
        <div className="rounded-lg border border-surface-border bg-surface-void p-3">
          <p className="text-sm leading-relaxed text-text-primary">{citation.text}</p>
        </div>
        <div className="mt-3 flex items-center justify-between">
          <span className="text-xs text-text-disabled">
            Relevance: {(citation.score * 100).toFixed(1)}%
          </span>
          <span className="font-mono text-[10px] text-text-disabled">
            {citation.chunkId.slice(0, 8)}
          </span>
        </div>
      </Modal>
    </>
  )
}
