import { useState } from 'react'
import { Button } from '@/components/ui/Button'
import { downloadDesignPackage } from '@/utils/download'
import type { BlackboardEvent } from '@/api/workflows'

/**
 * DownloadDesignPackageButton — shown on completed workflows.
 *
 * Converts all blackboard artifact events into a structured Markdown
 * document and triggers a browser file-save. No backend endpoint needed:
 * the data is already in memory from KanbanPage's blackboard query.
 *
 * WHY a separate component (not inline in KanbanPage):
 * - Keeps KanbanPage's JSX readable — it is already long.
 * - Isolates the download state (isDownloading) from the page's state.
 * - Matches the pattern of WorkflowChatPanel / AmendmentsPanel / DeliveryPanel
 *   — each panel is its own component, imported and placed in KanbanPage.
 */

const ARTIFACT_TYPES = new Set([
  'architecture_decision',
  'data_model_proposed',
  'api_contract_proposed',
  'module_design_proposed',
  'code_artifact_produced',
  'test_case_proposed',
  'requirement_captured',
])

export function DownloadDesignPackageButton({
  workflowTitle,
  events,
  expertNames,
}: {
  workflowTitle: string
  events: BlackboardEvent[]
  expertNames: Map<string, string>
}) {
  const [isDownloading, setIsDownloading] = useState(false)

  const artifactCount = events.filter((e) => ARTIFACT_TYPES.has(e.eventType)).length

  // Nothing to download yet — render nothing rather than a broken button.
  if (artifactCount === 0) return null

  const handleDownload = () => {
    setIsDownloading(true)
    try {
      downloadDesignPackage(workflowTitle, events, expertNames)
    } finally {
      // Reset after a short delay so the user sees the loading state briefly.
      setTimeout(() => setIsDownloading(false), 800)
    }
  }

  return (
    <Button
      variant="secondary"
      size="sm"
      isLoading={isDownloading}
      onClick={handleDownload}
      title={`Download all ${artifactCount} design artifact${
        artifactCount === 1 ? '' : 's'
      } as Markdown`}
    >
      {!isDownloading && (
        <span aria-hidden="true" className="mr-1">
          ⬇
        </span>
      )}
      Download Design Package
    </Button>
  )
}
