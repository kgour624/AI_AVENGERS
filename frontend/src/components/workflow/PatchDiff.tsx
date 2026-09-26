/**
 * PatchDiff — the workflow's changes, on screen, as a diff.
 *
 * WHY this exists: the patch could be generated and downloaded, and that was all.
 * Reviewing what an expert actually changed meant leaving the app, opening a
 * .patch file and hoping the editor's diff view understood it. Reviewing is the
 * step that decides whether the work is accepted, so it belongs next to the
 * button that produces it.
 *
 * What this deliberately does NOT do: syntax-highlight, collapse hunks, or hide
 * anything. A review tool that quietly hides part of a diff is worse than no
 * review tool, and a "−0 +0" summary over a diff nobody scrolled is how a wrong
 * change gets approved.
 */

function lineClass(line: string): string {
  // Order matters: the +++/--- file headers also start with + and -, so they are
  // matched first or every file header would be coloured as a change.
  if (line.startsWith('+++') || line.startsWith('---')) return 'text-text-primary font-medium'
  if (line.startsWith('@@')) return 'text-brand'
  if (line.startsWith('diff --git') || line.startsWith('index ') || line.startsWith('new file') || line.startsWith('deleted file'))
    return 'text-text-disabled'
  if (line.startsWith('+')) return 'bg-mode-advise/10 text-mode-advise'
  if (line.startsWith('-')) return 'bg-mode-refuse/10 text-mode-refuse'
  return 'text-text-secondary'
}

export function PatchDiff({ patch }: { patch: string }) {
  const lines = patch.split('\n')

  // Counts are computed from the patch text itself, so the summary cannot
  // disagree with the diff below it.
  let additions = 0
  let deletions = 0
  let files = 0
  for (const line of lines) {
    if (line.startsWith('diff --git')) files++
    else if (line.startsWith('+') && !line.startsWith('+++')) additions++
    else if (line.startsWith('-') && !line.startsWith('---')) deletions++
  }

  if (patch.trim() === '') {
    return (
      <p className="text-xs text-text-disabled">
        The patch is empty — this workflow has changed nothing yet.
      </p>
    )
  }

  return (
    <div>
      <p className="mb-2 text-xs text-text-secondary">
        {files} file{files === 1 ? '' : 's'} · <span className="text-mode-advise">+{additions}</span>{' '}
        <span className="text-mode-refuse">−{deletions}</span>
      </p>
      <pre className="max-h-96 overflow-auto rounded border border-surface-border bg-surface-secondary/40 p-2 text-[11px] leading-5">
        {lines.map((line, i) => (
          <div key={i} className={lineClass(line)}>
            {line === '' ? '\u00a0' : line}
          </div>
        ))}
      </pre>
    </div>
  )
}
