import { useState } from 'react'
import type { Citation } from '@/types/expert'
import { Tooltip } from '@/components/ui/Tooltip'
import { Modal } from '@/components/ui/Modal'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 9
 * ("Renders [CHUNK_abc] as clickable chip. On hover: shows chunk text
 * in tooltip. On click: opens citation detail modal. WHY: Citations
 * are the trust mechanism of AI Avengers.")
 *
 * WHY this component gets its own click-to-open-modal state instead
 * of relying on a shared "which citation is open" store: each
 * CitationChip instance is independent - a message can render several
 * citation chips, and opening one should never affect another's
 * state. Lifting this to a shared store (like streamStore) would add
 * complexity for zero benefit, since nothing outside this component
 * needs to know or control which citation modal is open - the
 * opposite reasoning from ExpertPicker's lifted state, and
 * deliberately so.
 */
export function CitationChip({ citation }: { citation: Citation }) {
  const [isModalOpen, setIsModalOpen] = useState(false)

  return (
    <>
      <Tooltip content={citation.text}>
        <button
          type="button"
          onClick={() => setIsModalOpen(true)}
          className="rounded border border-surface-border bg-surface-overlay px-1.5 py-0.5 text-xs font-mono text-brand hover:bg-brand/10"
        >
          [CHUNK_{citation.chunkId.slice(0, 8)}]
        </button>
      </Tooltip>

      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)}>
        <h3 className="mb-2 text-sm font-medium text-text-secondary">Citation detail</h3>
        <p className="text-sm text-text-primary">{citation.text}</p>
        <p className="mt-3 text-xs text-text-disabled">
          Chunk ID: {citation.chunkId} \u00b7 Relevance score: {citation.score.toFixed(3)}
        </p>
      </Modal>
    </>
  )
}
