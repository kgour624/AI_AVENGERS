"""Document extraction for transcript ingestion.

WHY this lives in the ML sidecar and not in the Go API:
  1. Parsing untrusted documents (PDF/DOCX/XLSX/...) is a hostile-input job.
     Running it in the sidecar keeps it out of the API process — the sidecar is
     non-root, holds no DB credentials and no request/response state.
  2. The mature parsers are Python (pypdf, python-docx, openpyxl, ...). The Go
     equivalents are either weak or licence-encumbered.
  3. It reuses the existing sidecar HTTP contract (see /embed and /rerank), so
     no new service, port or deployment unit is introduced.

Every parser is imported LAZILY inside its handler: one missing optional
dependency must disable that one format with a clear message, never stop the
sidecar from booting (the same process still serves embeddings).

Output contract: plain text with lightweight structure markers ("# Page 3",
"# Sheet: Sheet1", "# Slide 2", "# Chapter 4"). The markers matter because the
chunker downstream splits on text alone — without them every chunk from a
spreadsheet would be an anonymous row of numbers.
"""

from __future__ import annotations

import csv
import io
import os
import re
import shutil
import subprocess
import tempfile
from dataclasses import dataclass, field
from typing import Callable, Dict, List, Optional, Tuple

# Hard cap on extracted characters. WHY a cap: DOCX/XLSX/PPTX/EPUB are ZIP
# containers, so a few KB of uploaded bytes can expand to gigabytes of text
# (zip bomb). Failing loudly is deliberate: silently truncating would train the
# expert on a partial corpus while the UI reported success.
MAX_EXTRACTED_CHARS = int(os.getenv("DOC_MAX_EXTRACTED_CHARS", "5000000"))

# Cap on pages/sheets/slides processed, for the same reason.
MAX_PAGES = int(os.getenv("DOC_MAX_PAGES", "2000"))

# Rows read per spreadsheet sheet.
MAX_ROWS = int(os.getenv("DOC_MAX_ROWS", "20000"))

# A line is considered page furniture (header/footer) when it repeats as the
# first or last non-empty line of at least this share of pages.
_FURNITURE_RATIO = 0.6
_FURNITURE_MIN_PAGES = 4

# LibreOffice is the fallback for formats without a native parser (.doc, .ppt,
# .wpd, .pages, ...). Optional: absent means those formats report a clear error.
_SOFFICE = shutil.which("soffice") or shutil.which("libreoffice")


class ExtractionError(Exception):
    """A document could not be turned into text, with a machine-readable reason.

    `reason` is a stable code the Go layer maps to admin-facing copy; `message`
    is already human-readable and is shown verbatim.
    """

    def __init__(self, reason: str, message: str) -> None:
        super().__init__(message)
        self.reason = reason
        self.message = message


@dataclass
class ExtractionResult:
    text: str
    format: str
    chars: int
    pages: int = 0
    warnings: List[str] = field(default_factory=list)
    truncated: bool = False


# Extension -> family. One family may cover several extensions.
_FORMAT_BY_EXT: Dict[str, str] = {
    "pdf": "pdf",
    "docx": "docx",
    "doc": "legacy-word",
    "docm": "docx",
    "xlsx": "xlsx",
    "xlsm": "xlsx",
    "xls": "xls",
    "csv": "csv",
    "tsv": "csv",
    "pptx": "pptx",
    "ppt": "legacy-slides",
    "rtf": "rtf",
    "epub": "epub",
    "html": "html",
    "htm": "html",
    "xhtml": "html",
    "xml": "text",
    "odt": "odf",
    "ods": "odf",
    "odp": "odf",
    "txt": "text",
    "text": "text",
    "md": "text",
    "markdown": "text",
    "json": "text",
    "log": "text",
    "srt": "subtitles",
    "vtt": "subtitles",
}

# Extensions the API is allowed to accept. Kept in sync with the Go allowlist
# (backend-go/internal/docextract). The sidecar re-validates rather than
# trusting the caller — defense in depth for a hostile-input path.
SUPPORTED_EXTENSIONS = sorted(_FORMAT_BY_EXT.keys())

# Magic-byte signatures, used to catch a file whose extension lies (a .txt that
# is really a PDF). Only the formats where a mismatch is plausible are checked.
_MAGIC: List[Tuple[bytes, str]] = [
    (b"%PDF-", "pdf"),
    (b"PK\x03\x04", "zip"),  # docx/xlsx/pptx/epub/odt are ZIP containers
    (b"{\\rtf", "rtf"),
]


def _extension(filename: str) -> str:
    _, _, ext = filename.rpartition(".")
    return ext.strip().lower()


def _sniff(data: bytes) -> Optional[str]:
    head = data[:8]
    for signature, name in _MAGIC:
        if head.startswith(signature):
            return name
    return None


def _decode_text(data: bytes) -> str:
    """Decode a text file, trying the common encodings before giving up.

    WHY not just utf-8: real transcripts arrive as windows-1252/latin-1 often
    enough that a strict utf-8 decode would reject usable files.
    """
    for encoding in ("utf-8-sig", "utf-8", "utf-16"):
        try:
            return data.decode(encoding)
        except (UnicodeDecodeError, UnicodeError):
            continue
    try:
        from charset_normalizer import from_bytes

        best = from_bytes(data).best()
        if best is not None:
            return str(best)
    except ImportError:
        # charset-normalizer is optional: latin-1 never fails, so it is the
        # last-resort decode rather than an error.
        pass
    return data.decode("latin-1", errors="replace")


def _strip_page_furniture(page_texts: List[str]) -> List[str]:
    """Remove repeated headers/footers from PDF pages.

    Slide decks exported to PDF repeat the deck title and the page number on
    every page. Left in, that furniture becomes the dominant text in every chunk
    and dilutes retrieval. A line is only dropped when it repeats on most pages,
    so genuine repeated content in a short document survives.
    """
    if len(page_texts) < _FURNITURE_MIN_PAGES:
        return page_texts

    first_lines: Dict[str, int] = {}
    last_lines: Dict[str, int] = {}
    for text in page_texts:
        lines = [ln.strip() for ln in text.splitlines() if ln.strip()]
        if not lines:
            continue
        first_lines[lines[0]] = first_lines.get(lines[0], 0) + 1
        last_lines[lines[-1]] = last_lines.get(lines[-1], 0) + 1

    threshold = max(2, int(len(page_texts) * _FURNITURE_RATIO))
    furniture = {ln for ln, count in first_lines.items() if count >= threshold}
    furniture |= {ln for ln, count in last_lines.items() if count >= threshold}
    # A page-number-only line ("12", "Page 12 of 40", "12/40") is furniture even
    # when it is not identical across pages.
    page_number = re.compile(r"^(page\s*)?\d+(\s*(of|/)\s*\d+)?$", re.IGNORECASE)

    cleaned: List[str] = []
    for text in page_texts:
        kept = []
        for line in text.splitlines():
            stripped = line.strip()
            if stripped and (stripped in furniture or page_number.match(stripped)):
                continue
            kept.append(line)
        cleaned.append("\n".join(kept))
    return cleaned


def _join_sections(sections: List[str], warnings: List[str]) -> str:
    text = "\n\n".join(s for s in sections if s and s.strip())
    if len(text) > MAX_EXTRACTED_CHARS:
        raise ExtractionError(
            "too_large",
            f"Document expands to {len(text):,} characters, above the "
            f"{MAX_EXTRACTED_CHARS:,} limit. Split it into smaller files and "
            f"ingest them one by one.",
        )
    if not text.strip():
        warnings.append("extraction produced no text")
    return text


# ---------------------------------------------------------------------------
# PDF
# ---------------------------------------------------------------------------
def _extract_pdf(data: bytes) -> Tuple[str, int, List[str]]:
    from pypdf import PdfReader

    warnings: List[str] = []
    try:
        reader = PdfReader(io.BytesIO(data))
    except Exception as exc:  # noqa: BLE001 - surfaced to the admin verbatim
        raise ExtractionError("corrupt", f"Could not read PDF: {exc}") from exc

    if getattr(reader, "is_encrypted", False):
        # A password-protected file cannot be parsed without the password, and
        # guessing is not our job. Empty-password PDFs (owner-locked only) are
        # common though, so try that once before failing.
        try:
            if reader.decrypt("") == 0:
                raise ExtractionError(
                    "pdf_encrypted",
                    "PDF is password-protected. Remove the password and upload again.",
                )
        except ExtractionError:
            raise
        except Exception as exc:  # noqa: BLE001
            raise ExtractionError(
                "pdf_encrypted",
                "PDF is password-protected. Remove the password and upload again.",
            ) from exc

    pages = reader.pages
    if len(pages) > MAX_PAGES:
        raise ExtractionError(
            "too_many_pages",
            f"PDF has {len(pages)} pages, above the {MAX_PAGES} limit. "
            f"Split it and ingest the parts separately.",
        )

    page_texts: List[str] = []
    empty_pages = 0
    for page in pages:
        try:
            extracted = page.extract_text() or ""
        except Exception:  # noqa: BLE001 - one bad page must not kill the file
            extracted = ""
        if not extracted.strip():
            empty_pages += 1
        page_texts.append(extracted)

    page_texts = _strip_page_furniture(page_texts)
    sections = [
        f"# Page {index + 1}\n{text.strip()}"
        for index, text in enumerate(page_texts)
        if text.strip()
    ]

    if not sections:
        raise ExtractionError(
            "pdf_no_text",
            "No text layer found in this PDF (it looks scanned). OCR is not "
            "enabled — upload a text-based PDF or the original document.",
        )
    if empty_pages:
        warnings.append(f"{empty_pages} of {len(pages)} pages had no text layer")
    return _join_sections(sections, warnings), len(pages), warnings


# ---------------------------------------------------------------------------
# Word (DOCX) / legacy DOC
# ---------------------------------------------------------------------------
def _extract_docx(data: bytes) -> Tuple[str, int, List[str]]:
    from docx import Document

    warnings: List[str] = []
    try:
        document = Document(io.BytesIO(data))
    except Exception as exc:  # noqa: BLE001
        raise ExtractionError("corrupt", f"Could not read DOCX: {exc}") from exc

    parts: List[str] = []
    for paragraph in document.paragraphs:
        text = paragraph.text.strip()
        if text:
            parts.append(text)

    for index, table in enumerate(document.tables):
        rows = []
        for row in table.rows:
            cells = [cell.text.strip() for cell in row.cells]
            if any(cells):
                rows.append(" | ".join(cells))
        if rows:
            parts.append(f"# Table {index + 1}\n" + "\n".join(rows))

    return _join_sections(parts, warnings), 0, warnings


# ---------------------------------------------------------------------------
# Spreadsheets
# ---------------------------------------------------------------------------
def _rows_to_text(rows: List[List[str]]) -> List[str]:
    """Render rows as "header: value | header: value" lines.

    WHY not raw tab-separated rows: retrieval compares a question against a
    chunk's text. "Latency: 200ms | Limit: 1000 rps" carries the column meaning
    into every chunk, while "200ms\\t1000" does not.
    """
    rows = [[("" if cell is None else str(cell)).strip() for cell in row] for row in rows]
    rows = [row for row in rows if any(row)]
    if not rows:
        return []

    header = rows[0]
    body = rows[1:]
    # A "header" that is all numbers is really data, not a header row.
    if header and all(re.fullmatch(r"-?\d+(\.\d+)?", cell or "") for cell in header if cell):
        header, body = [], rows

    lines: List[str] = []
    for row in body:
        if header and len(header) >= len(row):
            pairs = [
                f"{header[i]}: {value}"
                for i, value in enumerate(row)
                if value and i < len(header)
            ]
        else:
            pairs = [value for value in row if value]
        if pairs:
            lines.append(" | ".join(pairs))
    return lines


def _extract_xlsx(data: bytes) -> Tuple[str, int, List[str]]:
    from openpyxl import load_workbook

    warnings: List[str] = []
    try:
        workbook = load_workbook(io.BytesIO(data), read_only=True, data_only=True)
    except Exception as exc:  # noqa: BLE001
        raise ExtractionError("corrupt", f"Could not read XLSX: {exc}") from exc

    sections: List[str] = []
    sheet_count = 0
    for sheet in workbook.worksheets:
        sheet_count += 1
        rows: List[List[str]] = []
        for index, row in enumerate(sheet.iter_rows(values_only=True)):
            if index >= MAX_ROWS:
                warnings.append(
                    f"sheet '{sheet.title}' truncated at {MAX_ROWS} rows"
                )
                break
            rows.append(list(row))
        lines = _rows_to_text(rows)
        if lines:
            sections.append(f"# Sheet: {sheet.title}\n" + "\n".join(lines))
    workbook.close()

    if not sections:
        raise ExtractionError("empty", "Workbook contains no readable rows.")
    return _join_sections(sections, warnings), sheet_count, warnings


def _extract_xls(data: bytes) -> Tuple[str, int, List[str]]:
    warnings: List[str] = []
    try:
        import xlrd
    except ImportError:
        return _extract_legacy(data, "xls")

    try:
        book = xlrd.open_workbook(file_contents=data)
    except Exception:  # noqa: BLE001 - fall through to LibreOffice
        return _extract_legacy(data, "xls")

    sections: List[str] = []
    for sheet in book.sheets():
        rows = [
            [sheet.cell_value(r, c) for c in range(sheet.ncols)]
            for r in range(min(sheet.nrows, MAX_ROWS))
        ]
        if sheet.nrows > MAX_ROWS:
            warnings.append(f"sheet '{sheet.name}' truncated at {MAX_ROWS} rows")
        lines = _rows_to_text(rows)
        if lines:
            sections.append(f"# Sheet: {sheet.name}\n" + "\n".join(lines))

    if not sections:
        raise ExtractionError("empty", "Workbook contains no readable rows.")
    return _join_sections(sections, warnings), len(book.sheets()), warnings


def _extract_csv(data: bytes) -> Tuple[str, int, List[str]]:
    warnings: List[str] = []
    text = _decode_text(data)
    sample = text[:4096]
    try:
        dialect = csv.Sniffer().sniff(sample, delimiters=",;\t|")
    except csv.Error:
        dialect = csv.excel

    reader = csv.reader(io.StringIO(text), dialect)
    rows: List[List[str]] = []
    for index, row in enumerate(reader):
        if index >= MAX_ROWS:
            warnings.append(f"file truncated at {MAX_ROWS} rows")
            break
        rows.append(row)

    lines = _rows_to_text(rows)
    if not lines:
        raise ExtractionError("empty", "CSV contains no readable rows.")
    return _join_sections(lines, warnings), 1, warnings


# ---------------------------------------------------------------------------
# Slides
# ---------------------------------------------------------------------------
def _extract_pptx(data: bytes) -> Tuple[str, int, List[str]]:
    from pptx import Presentation

    warnings: List[str] = []
    try:
        presentation = Presentation(io.BytesIO(data))
    except Exception as exc:  # noqa: BLE001
        raise ExtractionError("corrupt", f"Could not read PPTX: {exc}") from exc

    sections: List[str] = []
    for index, slide in enumerate(presentation.slides):
        if index >= MAX_PAGES:
            warnings.append(f"presentation truncated at {MAX_PAGES} slides")
            break
        parts: List[str] = []
        for shape in slide.shapes:
            if getattr(shape, "has_text_frame", False):
                for paragraph in shape.text_frame.paragraphs:
                    text = "".join(run.text for run in paragraph.runs).strip()
                    if text:
                        parts.append(text)
            if getattr(shape, "has_table", False):
                for row in shape.table.rows:
                    cells = [cell.text.strip() for cell in row.cells]
                    if any(cells):
                        parts.append(" | ".join(cells))
        try:
            notes = slide.notes_slide.notes_text_frame.text.strip()
            if notes:
                parts.append(f"Notes: {notes}")
        except Exception:  # noqa: BLE001 - slides without a notes master
            pass
        if parts:
            sections.append(f"# Slide {index + 1}\n" + "\n".join(parts))

    if not sections:
        raise ExtractionError("empty", "Presentation contains no readable text.")
    return _join_sections(sections, warnings), len(presentation.slides), warnings


# ---------------------------------------------------------------------------
# HTML / EPUB / RTF / ODF
# ---------------------------------------------------------------------------
def _html_to_text(markup: str) -> str:
    from bs4 import BeautifulSoup

    soup = BeautifulSoup(markup, "lxml")
    for tag in soup(["script", "style", "noscript", "nav", "footer", "header"]):
        tag.decompose()
    for level in range(1, 7):
        for heading in soup.find_all(f"h{level}"):
            heading.insert_before(soup.new_string(f"\n\n{'#' * level} "))
            heading.insert_after(soup.new_string("\n\n"))
    text = soup.get_text(separator="\n")
    return re.sub(r"\n{3,}", "\n\n", text)


def _extract_html(data: bytes) -> Tuple[str, int, List[str]]:
    warnings: List[str] = []
    text = _join_sections([_html_to_text(_decode_text(data))], warnings)
    return text, 0, warnings


def _extract_epub(data: bytes) -> Tuple[str, int, List[str]]:
    import ebooklib
    from ebooklib import epub

    warnings: List[str] = []
    try:
        book = epub.read_epub(io.BytesIO(data))
    except Exception as exc:  # noqa: BLE001
        raise ExtractionError("corrupt", f"Could not read EPUB: {exc}") from exc

    sections: List[str] = []
    for index, item in enumerate(book.get_items_of_type(ebooklib.ITEM_DOCUMENT)):
        if index >= MAX_PAGES:
            warnings.append(f"book truncated at {MAX_PAGES} chapters")
            break
        chapter = _html_to_text(item.get_content().decode("utf-8", errors="replace"))
        if chapter.strip():
            sections.append(f"# Chapter {index + 1}\n{chapter.strip()}")

    if not sections:
        raise ExtractionError("empty", "EPUB contains no readable text.")
    return _join_sections(sections, warnings), len(sections), warnings


def _extract_rtf(data: bytes) -> Tuple[str, int, List[str]]:
    from striprtf.striprtf import rtf_to_text

    warnings: List[str] = []
    try:
        text = rtf_to_text(data.decode("latin-1", errors="replace"), errors="ignore")
    except Exception as exc:  # noqa: BLE001
        raise ExtractionError("corrupt", f"Could not read RTF: {exc}") from exc
    return _join_sections([text], warnings), 0, warnings


def _extract_odf(data: bytes) -> Tuple[str, int, List[str]]:
    from odf import teletype
    from odf.opendocument import load
    from odf.text import P

    warnings: List[str] = []
    try:
        document = load(io.BytesIO(data))
    except Exception as exc:  # noqa: BLE001
        raise ExtractionError("corrupt", f"Could not read ODF file: {exc}") from exc

    paragraphs = [
        teletype.extractText(paragraph).strip()
        for paragraph in document.getElementsByType(P)
    ]
    text = "\n".join(p for p in paragraphs if p)
    if not text.strip():
        raise ExtractionError("empty", "Document contains no readable text.")
    return _join_sections([text], warnings), 0, warnings


# ---------------------------------------------------------------------------
# Subtitles / plain text
# ---------------------------------------------------------------------------
_TIMECODE = re.compile(
    r"^\s*(\d+\s*$|WEBVTT|NOTE\s|\d{1,2}:\d{2}:\d{2}[.,]\d{1,3}\s*-->)"
)


def _extract_subtitles(data: bytes) -> Tuple[str, int, List[str]]:
    """Strip cue numbers and timecodes, keep the spoken text.

    WHY: a 2-hour .srt is ~40% timestamps and index lines. They are pure noise
    for retrieval and they inflate the chunk count.
    """
    warnings: List[str] = []
    text = _decode_text(data)
    kept: List[str] = []
    for line in text.splitlines():
        if not line.strip() or _TIMECODE.match(line):
            continue
        if line.strip() and kept and kept[-1] == line.strip():
            continue  # rolling-caption duplicates
        if line.strip():
            kept.append(line.strip())
    return _join_sections(kept, warnings), 0, warnings


def _extract_text(data: bytes) -> Tuple[str, int, List[str]]:
    warnings: List[str] = []
    return _join_sections([_decode_text(data)], warnings), 0, warnings


# ---------------------------------------------------------------------------
# Legacy binary formats (.doc / .ppt / .xls-fallback / exotic)
# ---------------------------------------------------------------------------
_CATDOC = shutil.which("catdoc")
_CATPPT = shutil.which("catppt")
_XLS2CSV = shutil.which("xls2csv")


def _run_converter(command: List[str], timeout: int = 180) -> bytes:
    try:
        completed = subprocess.run(command, capture_output=True, timeout=timeout, check=False)
    except subprocess.TimeoutExpired as exc:
        raise ExtractionError(
            "timeout", "Conversion timed out. The file may be corrupt or too large."
        ) from exc
    if completed.returncode != 0:
        detail = (completed.stderr or b"").decode("utf-8", errors="replace")[:300]
        raise ExtractionError(
            "corrupt", f"Converter could not read this file. {detail}".strip()
        )
    return completed.stdout or b""


def _extract_legacy(data: bytes, extension: str) -> Tuple[str, int, List[str]]:
    """Extract a legacy binary format.

    Two tiers, cheapest first:
      1. catdoc tools (catdoc / catppt / xls2csv) — ~2MB, installed by default.
         They cover the legacy formats admins actually upload (.doc, .ppt).
      2. LibreOffice, if the image was built with WITH_LIBREOFFICE=true. It
         handles the long tail (.wpd, .pages, .key, .numbers, ...) at the cost
         of a ~500MB dependency, which is why it is opt-in.

    WHY not LibreOffice only: it needs a writable profile per invocation and
    dominates the image. WHY not catdoc only: some formats have no catdoc tool.
    """
    with tempfile.TemporaryDirectory() as workdir:
        source = os.path.join(workdir, f"input.{extension}")
        with open(source, "wb") as handle:
            handle.write(data)

        warnings: List[str] = []
        tier1: Dict[str, Optional[str]] = {
            "doc": _CATDOC,
            "ppt": _CATPPT,
            "xls": _XLS2CSV,
        }
        tool = tier1.get(extension)
        if tool:
            text = _decode_text(_run_converter([tool, source]))
            if text.strip():
                warnings.append(f"converted with {os.path.basename(tool)}")
                return _join_sections([text], warnings), 0, warnings

        if _SOFFICE:
            completed_ok = False
            try:
                subprocess.run(
                    [
                        _SOFFICE,
                        "--headless",
                        "--norestore",
                        f"-env:UserInstallation=file://{workdir}/profile",
                        "--convert-to",
                        "txt:Text (encoded):UTF8",
                        "--outdir",
                        workdir,
                        source,
                    ],
                    capture_output=True,
                    timeout=180,
                    check=True,
                )
                completed_ok = True
            except (subprocess.TimeoutExpired, subprocess.CalledProcessError):
                completed_ok = False

            produced = os.path.join(workdir, "input.txt")
            if completed_ok and os.path.exists(produced):
                with open(produced, "rb") as handle:
                    converted = handle.read()
                if converted.strip():
                    warnings.append("converted with LibreOffice")
                    return _join_sections([_decode_text(converted)], warnings), 0, warnings

    raise ExtractionError(
        "no_parser",
        f"No converter available for .{extension} in this deployment. Convert the "
        f"file to PDF or DOCX and upload it again.",
    )


# ---------------------------------------------------------------------------
# Dispatch
# ---------------------------------------------------------------------------
_HANDLERS: Dict[str, Callable[[bytes], Tuple[str, int, List[str]]]] = {
    "pdf": _extract_pdf,
    "docx": _extract_docx,
    "xlsx": _extract_xlsx,
    "xls": _extract_xls,
    "csv": _extract_csv,
    "pptx": _extract_pptx,
    "html": _extract_html,
    "epub": _extract_epub,
    "rtf": _extract_rtf,
    "odf": _extract_odf,
    "subtitles": _extract_subtitles,
    "text": _extract_text,
}

_LEGACY_EXTENSIONS = {"legacy-word": "doc", "legacy-slides": "ppt"}
_ODF_MIMETYPES = {
    "application/vnd.oasis.opendocument.text": "odt",
    "application/vnd.oasis.opendocument.spreadsheet": "ods",
    "application/vnd.oasis.opendocument.presentation": "odp",
}


def _validate_container(extension: str, data: bytes) -> None:
    """Check that the extension and binary container agree before parsing.

    WHY ZIP member sniffing instead of just the PK signature: DOCX/XLSX/PPTX,
    EPUB and ODF are all ZIP files. A .txt renamed from any of those must not
    fall through to a latin-1 decode and get trained as garbage.
    """
    sniffed = _sniff(data)
    if sniffed == "zip":
        try:
            import zipfile

            with zipfile.ZipFile(io.BytesIO(data)) as archive:
                names = {name.lower() for name in archive.namelist()[:5000]}
        except Exception as exc:  # noqa: BLE001
            raise ExtractionError("corrupt", f"Could not read ZIP-based document: {exc}") from exc

        if any(name.startswith("word/") for name in names):
            actual = "docx"
        elif any(name.startswith("xl/") for name in names):
            actual = "xlsx"
        elif any(name.startswith("ppt/") for name in names):
            actual = "pptx"
        elif "mimetype" in names:
            try:
                with zipfile.ZipFile(io.BytesIO(data)) as archive:
                    mimetype = archive.read("mimetype").decode("ascii", errors="ignore").strip()
            except Exception as exc:  # noqa: BLE001
                raise ExtractionError("corrupt", f"Could not read document mimetype: {exc}") from exc
            actual = _ODF_MIMETYPES.get(mimetype, "unknown-zip")
            if mimetype == "application/epub+zip":
                actual = "epub"
        elif any(name.startswith("meta-inf/") for name in names):
            actual = "epub"
        else:
            actual = "zip"

        compatible = {
            "docx": {"docx", "docm"},
            "xlsx": {"xlsx", "xlsm"},
            "pptx": {"pptx"},
            "epub": {"epub"},
            "odt": {"odt"},
            "ods": {"ods"},
            "odp": {"odp"},
            "zip": set(),
        }
        if extension not in compatible.get(actual, set()):
            raise ExtractionError(
                "extension_mismatch",
                f"This file is named .{extension} but its content looks like {actual}. "
                f"Rename it with the correct extension and retry.",
            )
        return

    expected = {
        "pdf": {"pdf"},
        "rtf": {"rtf"},
        "legacy-office": {"doc", "ppt", "xls"},
    }
    if sniffed in expected and extension not in expected[sniffed]:
        raise ExtractionError(
            "extension_mismatch",
            f"This file is named .{extension} but its content looks like {sniffed}. "
            f"Rename it with the correct extension and retry.",
        )
    if sniffed and extension in {"txt", "text", "md", "markdown", "json", "log", "xml"}:
        raise ExtractionError(
            "extension_mismatch",
            f"This file is named .{extension} but its content looks like {sniffed}. "
            f"Rename it with the correct extension and retry.",
        )


def extract_document(filename: str, data: bytes) -> ExtractionResult:
    """Turn an uploaded document into plain text.

    Raises ExtractionError with a stable reason code on any failure — the Go
    layer surfaces `message` to the admin and logs `reason`.
    """
    if not data:
        raise ExtractionError("empty", "The uploaded file is empty.")

    extension = _extension(filename)
    family = _FORMAT_BY_EXT.get(extension)
    if family is None:
        raise ExtractionError(
            "unsupported_format",
            f".{extension or 'unknown'} is not a supported format. Supported: "
            + ", ".join(f".{ext}" for ext in SUPPORTED_EXTENSIONS),
        )

    # .xls is a legacy OLE file; .xlsx is ZIP. Accept an XLSX payload named
    # .xls only after correcting the extension is not enough—the UI should
    # fail clearly rather than feed it to xlrd's legacy parser.
    if extension == "xls" and data[:4] == b"PK\x03\x04":
        raise ExtractionError(
            "extension_mismatch",
            "This file is named .xls but its content is an XLSX ZIP workbook. "
            "Rename it to .xlsx and upload again.",
        )

    # Extension-vs-content check: e.g. a .txt that is really a PDF/ZIP must not
    # be decoded as latin-1 garbage and silently trained on.
    _validate_container(extension, data)

    if family in _LEGACY_EXTENSIONS:
        warnings: List[str] = []
        text, pages, warnings = _extract_legacy(data, _LEGACY_EXTENSIONS[family])
        return ExtractionResult(
            text=text,
            format=family,
            chars=len(text),
            pages=pages,
            warnings=warnings,
        )

    handler = _HANDLERS[family]
    try:
        text, pages, warnings = handler(data)
    except ExtractionError:
        raise
    except ImportError as exc:
        raise ExtractionError(
            "parser_unavailable",
            f"No parser available for .{extension} in this deployment ({exc}).",
        ) from exc

    return ExtractionResult(
        text=text,
        format=family,
        chars=len(text),
        pages=pages,
        warnings=warnings,
        truncated=False,
    )
