import { useState } from 'react'
import { Highlight, themes, type Language } from 'prism-react-renderer'

/**
 * Syntax-highlighted code block, used as react-markdown's `code`
 * component override. Source: FRONTEND_SYSTEM_DESIGN.md section 2
 * ("prism-react-renderer - Code syntax highlighting in responses").
 *
 * WHY the language guard: react-markdown passes className like
 * "language-sql" for fenced code blocks, or no className at all for
 * inline code (single backticks). Highlight requires a valid Language
 * value - passing an empty string or an unrecognized language would
 * throw at render time rather than degrading gracefully. Falls back
 * to 'tsx' as a reasonable default rather than crashing the whole
 * message on an unrecognized/missing language tag.
 */
export function CodeBlock({ className, children }: { className?: string; children?: React.ReactNode }) {
  const match = /language-(\w+)/.exec(className ?? '')
  const language = (match?.[1] ?? 'tsx') as Language
  const code = String(children).replace(/\n$/, '')
  const [copied, setCopied] = useState(false)

  if (!className) {
    // Inline code (single backtick) - no fence, no block styling.
    return <code className="rounded bg-surface-overlay px-1 py-0.5 font-mono text-sm">{code}</code>
  }

  return (
    <div className="group relative my-2">
      <button
        type="button"
        onClick={() => {
          navigator.clipboard
            .writeText(code)
            .then(() => {
              setCopied(true)
              setTimeout(() => setCopied(false), 1500)
            })
            .catch((err) => {
              console.warn('copy to clipboard failed', err)
            })
        }}
        aria-label="Copy code"
        className="absolute right-2 top-2 rounded-md border border-surface-border bg-surface-overlay/80 px-2 py-1 text-xs text-text-secondary opacity-0 backdrop-blur-sm transition-opacity hover:text-text-primary group-hover:opacity-100"
      >
        {copied ? 'Copied!' : 'Copy'}
      </button>
      <Highlight theme={themes.vsDark} code={code} language={language}>
        {({ style, tokens, getLineProps, getTokenProps }) => (
          <pre style={style} className="overflow-x-auto rounded-md p-3 text-sm">
            {tokens.map((line, i) => (
              <div key={i} {...getLineProps({ line })}>
                {line.map((token, key) => (
                  <span key={key} {...getTokenProps({ token })} />
                ))}
              </div>
            ))}
          </pre>
        )}
      </Highlight>
    </div>
  )
}
