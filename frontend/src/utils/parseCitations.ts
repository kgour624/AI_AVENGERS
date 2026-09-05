import type { Citation } from '@/types/expert'

/**
 * Splits an expert response's raw content string on inline citation
 * tokens, e.g. "Use UUID keys [CHUNK_abc123] because ... [CHUNK_def456]."
 *
 * WHY this exists: AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 8
 * (China Wall Layer 4, stripUncited) shows the backend's own regex for
 * finding citation tokens: `\[CHUNK_[a-f0-9-]+\]`. This confirms the
 * tokens are literally embedded inline in the content string the
 * backend sends - they are NOT a separate structured field the
 * frontend can render independently of the prose. FRONTEND_SYSTEM_
 * DESIGN.md's CitationChip doc comment (section 9) says it "renders
 * [CHUNK_abc] as a clickable chip" but never shows HOW that chip gets
 * inserted into the middle of markdown-formatted prose - this bridges
 * that gap.
 *
 * Approach: regex-split the raw string into text/citation segments
 * BEFORE handing anything to react-markdown, then render each text
 * segment as its own independent markdown fragment and each citation
 * segment as a CitationChip. KNOWN LIMITATION (documented rather than
 * silently accepted): if a citation token falls in the middle of a
 * multi-line markdown block construct (e.g. inside a numbered list
 * item or a code fence), splitting there and rendering two separate
 * markdown fragments can break that block's continuity, since each
 * fragment is parsed independently. This is an acceptable tradeoff
 * for now because in practice, per the backend's own generation
 * prompt (architecture doc section 11), citations are attached at
 * sentence boundaries, not mid-list-item or mid-code-fence - but if
 * that assumption changes, this function's approach needs revisiting.
 */
export type ContentSegment =
  | { type: 'text'; value: string }
  | { type: 'citation'; citation: Citation }

const CITATION_TOKEN_RE = /\[CHUNK_([a-f0-9-]+)\]/g

export function splitContentByCitations(content: string, citations: Citation[]): ContentSegment[] {
  const citationMap = new Map(citations.map((c) => [c.chunkId, c]))
  const segments: ContentSegment[] = []

  let lastIndex = 0
  let match: RegExpExecArray | null

  // WHY a fresh regex per call (lastIndex reset) rather than reusing
  // the module-level CITATION_TOKEN_RE directly with its own exec loop
  // state: a global regex's `.lastIndex` is mutable shared state. If
  // two calls to this function ever interleaved (unlikely given it's
  // synchronous, but not impossible if called from two components
  // rendering in the same tick), sharing lastIndex across calls would
  // corrupt both results. Cloning the regex per call avoids that
  // entirely rather than relying on "probably fine in practice."
  const re = new RegExp(CITATION_TOKEN_RE.source, CITATION_TOKEN_RE.flags)

  while ((match = re.exec(content)) !== null) {
    if (match.index > lastIndex) {
      segments.push({ type: 'text', value: content.slice(lastIndex, match.index) })
    }

    const chunkId = match[1]!
    const citation = citationMap.get(chunkId)

    if (citation) {
      segments.push({ type: 'citation', citation })
    } else {
      // Token present in the text but no matching Citation object was
      // sent alongside it - render the literal token as text rather
      // than silently dropping it. Dropping user-facing content is a
      // worse failure mode than showing a slightly odd-looking
      // unmatched token; this should also be rare enough in practice
      // to be worth investigating if it shows up.
      segments.push({ type: 'text', value: match[0] })
    }

    lastIndex = re.lastIndex
  }

  if (lastIndex < content.length) {
    segments.push({ type: 'text', value: content.slice(lastIndex) })
  }

  return segments
}
