import type { ExpertResponse as ExpertResponseType, TemplateSectionResult, Citation } from "../types/expert";

/**
 * Builds a sanitized, filesystem-safe filename from an expert/role name and timestamp.
 */
export function buildMarkdownFilename(roleName: string, timestamp?: string | number | Date): string {
  const safeRole = roleName
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "") || "response";

  const date = timestamp ? new Date(timestamp) : new Date();
  const pad = (n: number) => String(n).padStart(2, "0");
  const stamp = `${date.getFullYear()}${pad(date.getMonth() + 1)}${pad(date.getDate())}-${pad(
    date.getHours()
  )}${pad(date.getMinutes())}${pad(date.getSeconds())}`;

  return `${safeRole}-${stamp}.md`;
}

/**
 * Formats citations into a Markdown list. Falls back gracefully if fields are missing.
 */
function formatCitations(citations?: Citation[]): string {
  if (!citations || citations.length === 0) return "";

  const lines = citations.map((c, idx) => {
    const label = c.sourceName || c.chunkId || `Source ${idx + 1}`;
    const extra = c.chunkIndex !== undefined ? ` (Chunk #${c.chunkIndex + 1})` : "";
    return `${idx + 1}. ${label}${extra}`;
  });

  return `\n\n## Citations\n\n${lines.join("\n")}`;
}

/**
 * Flattens templateSections into Markdown headings + body text, matching the
 * same precedence order used by handleCopyResponse: templateSections first,
 * falling back to the flat `content` string.
 */
function formatSections(sections: TemplateSectionResult[]): string {
  return sections
    .map((section) => {
      const heading = section.title?.trim() || section.sectionId || "Section";
      const body = (section.content ?? "").trim();
      return `## ${heading}\n\n${body}`;
    })
    .join("\n\n");
}

/**
 * Builds the full Markdown document for a single expert response, including
 * role name, verdict/confidence metadata, body content, and citations.
 */
export function buildResponseMarkdown(response: ExpertResponseType): string {
  const title = `# ${response.roleName ?? "Response"}`;

  const metaParts: string[] = [];
  if (response.verdict) metaParts.push(`**Verdict:** ${response.verdict}`);
  if (typeof response.confidence === "number") {
    metaParts.push(`**Confidence:** ${Math.round(response.confidence)}%`);
  }
  const meta = metaParts.length > 0 ? metaParts.join("  \n") : "";

  const hasSections = Array.isArray(response.templateSections) && response.templateSections.length > 0;
  const body = hasSections
    ? formatSections(response.templateSections as TemplateSectionResult[])
    : (response.content ?? "").trim();

  const citationsBlock = formatCitations(response.citations);

  return [title, meta, body].filter(Boolean).join("\n\n") + citationsBlock + "\n";
}

/**
 * Triggers a browser download of the given markdown text as a .md file,
 * using Blob + a temporary <a download> link (no external dependency).
 */
export function downloadMarkdown(markdown: string, filename: string): void {
  const blob = new Blob([markdown], { type: "text/markdown;charset=utf-8" });
  const url = URL.createObjectURL(blob);

  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  URL.revokeObjectURL(url);
}
