import type { BlackboardEvent } from '@/api/workflows'

/**
 * downloadDesignPackage — converts blackboard artifact events into a
 * structured Markdown document and triggers a browser file-save.
 *
 * WHY client-side: the backend has no dedicated download endpoint
 * (delivery.ts only exposes git-push and code-feedback). All artifact
 * data is already fetched by KanbanPage via getBlackboard(), so we
 * generate the file from that in-memory data — zero extra network
 * round-trips, zero new dependencies (native Blob + createObjectURL).
 *
 * WHY Markdown: every artifact's content is already prose / JSON that
 * renders well as fenced code blocks. A single .md file is universally
 * readable without tooling.
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

function renderValue(value: unknown): string {
  if (typeof value === 'string') return value
  return JSON.stringify(value, null, 2)
}

function toHeading(eventType: string): string {
  return eventType
    .split('_')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ')
}

function toLabel(key: string): string {
  return key
    .replace(/([A-Z])/g, ' $1')
    .replace(/_/g, ' ')
    .trim()
    .replace(/^./, (c) => c.toUpperCase())
}

export function downloadDesignPackage(
  workflowTitle: string,
  events: BlackboardEvent[],
  expertNames: Map<string, string>
): void {
  const artifacts = events.filter((e) => ARTIFACT_TYPES.has(e.eventType))
  if (artifacts.length === 0) return

  const grouped = new Map<string, BlackboardEvent[]>()
  for (const artifact of artifacts) {
    const list = grouped.get(artifact.eventType) ?? []
    list.push(artifact)
    grouped.set(artifact.eventType, list)
  }

  const lines: string[] = []

  lines.push(`# Design Package — ${workflowTitle}`)
  lines.push('')
  lines.push(
    `Generated on ${new Date().toUTCString()} · ${artifacts.length} artifact${
      artifacts.length === 1 ? '' : 's'
    }`
  )
  lines.push('')
  lines.push('---')
  lines.push('')

  lines.push('## Table of Contents')
  lines.push('')
  for (const [eventType] of grouped) {
    lines.push(`- [${toHeading(eventType)}](#${eventType.replace(/_/g, '-')})`)
  }
  lines.push('')
  lines.push('---')
  lines.push('')

  for (const [eventType, group] of grouped) {
    lines.push(`## ${toHeading(eventType)}`)
    lines.push('')

    for (const artifact of group) {
      const expertName =
        (artifact.postedByExpertId && expertNames.get(artifact.postedByExpertId)) || 'system'

      lines.push(`### ${expertName}`)
      lines.push('')

      for (const [key, value] of Object.entries(artifact.content ?? {})) {
        const rendered = renderValue(value)
        lines.push(`**${toLabel(key)}**`)
        lines.push('')
        if (rendered.includes('\n') || rendered.length > 120) {
          lines.push('```')
          lines.push(rendered)
          lines.push('```')
        } else {
          lines.push(rendered)
        }
        lines.push('')
      }

      lines.push('---')
      lines.push('')
    }
  }

  const markdown = lines.join('\n')
  const blob = new Blob([markdown], { type: 'text/markdown;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  const safeName = workflowTitle.replace(/[^a-z0-9\-_. ]/gi, '_').trim() || 'workflow'
  anchor.download = `${safeName}-design-package.md`
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
