// Package docextract turns an uploaded document into plain text for ingestion.
//
// WHY this exists: transcript ingestion was built for plain text — the API read
// the upload's raw bytes and treated them as a UTF-8 string, so a PDF produced
// garbage (or "empty after cleaning") and every other office format was
// unusable. Admins train experts from lecture PDFs, slide decks, spreadsheets
// and Word docs, so the bytes must be converted to text BEFORE the pipeline
// runs.
//
// WHY the conversion happens in the ML sidecar (see the /extract endpoint):
//   * parsing untrusted documents is a hostile-input job; the sidecar is
//     non-root and holds no DB credentials,
//   * the mature parsers are Python (pypdf, python-docx, openpyxl, ...),
//   * it reuses an existing deployment unit instead of adding a new service.
//
// WHY plain-text formats are decoded in-process instead of calling the sidecar:
// .txt/.md are what the system was originally built for, so they must keep
// working even when the sidecar is down. Only formats that need real parsing
// depend on it.
//
// The allowlist below deliberately mirrors ml-sidecar/extractor.py. The sidecar
// re-validates rather than trusting this list (defense in depth), and this list
// is authoritative for the fast 400 the admin sees before a job is created.
package docextract

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"go.uber.org/zap"
)

// formatByExtension maps a lowercase extension to the parser family used by the
// sidecar. One family may cover several extensions (docm → docx).
var formatByExtension = map[string]string{
	"pdf":      "pdf",
	"docx":     "docx",
	"docm":     "docx",
	"doc":      "legacy-word",
	"xlsx":     "xlsx",
	"xlsm":     "xlsx",
	"xls":      "xls",
	"csv":      "csv",
	"tsv":      "csv",
	"pptx":     "pptx",
	"ppt":      "legacy-slides",
	"rtf":      "rtf",
	"epub":     "epub",
	"html":     "html",
	"htm":      "html",
	"xhtml":    "html",
	"odt":      "odf",
	"ods":      "odf",
	"odp":      "odf",
	"txt":      "text",
	"text":     "text",
	"md":       "text",
	"markdown": "text",
	"json":     "text",
	"log":      "text",
	"xml":      "text",
	"srt":      "subtitles",
	"vtt":      "subtitles",
}

// plainTextExtensions need no parsing: the bytes ARE the text. Decoded locally
// so ingestion of the original text formats never depends on the sidecar.
var plainTextExtensions = map[string]bool{
	"txt": true, "text": true, "md": true, "markdown": true,
	"json": true, "log": true, "xml": true,
}

// maxUploadBytes mirrors the API's 50MB guard with headroom, so the API's
// friendlier error wins while the sidecar still has a hard backstop.
const maxUploadBytes = 60 * 1024 * 1024

// SupportedExtensions returns the allowlist, sorted, for error messages and docs.
func SupportedExtensions() []string {
	out := make([]string, 0, len(formatByExtension))
	for ext := range formatByExtension {
		out = append(out, ext)
	}
	sort.Strings(out)
	return out
}

// SupportedList renders the allowlist as ".pdf, .docx, ..." for a user-facing
// error message.
func SupportedList() string {
	exts := SupportedExtensions()
	for i, ext := range exts {
		exts[i] = "." + ext
	}
	return strings.Join(exts, ", ")
}

// Extension returns the lowercase extension of a filename ("" when absent).
func Extension(filename string) string {
	return strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
}

// IsSupported reports whether the file extension is accepted for ingestion.
func IsSupported(filename string) bool {
	_, ok := formatByExtension[Extension(filename)]
	return ok
}

// DetectBinaryFamily classifies signatures that must never be passed through
// the plain-text fast path. A .txt carrying a PDF/ZIP/legacy Office payload is
// rejected before it can become garbage transcript text.
func DetectBinaryFamily(data []byte) string {
	switch {
	case bytes.HasPrefix(data, []byte("%PDF-")):
		return "pdf"
	case bytes.HasPrefix(data, []byte("{\\rtf")):
		return "rtf"
	case len(data) >= 8 && bytes.Equal(data[:8], []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}):
		return "legacy-office"
	case len(data) >= 4 && bytes.Equal(data[:4], []byte{'P', 'K', 3, 4}):
		return detectZipFamily(data)
	default:
		return ""
	}
}

func detectZipFamily(data []byte) string {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "zip"
	}
	for _, file := range reader.File {
		name := strings.ToLower(file.Name)
		switch {
		case strings.HasPrefix(name, "word/"):
			return "docx"
		case strings.HasPrefix(name, "xl/"):
			return "xlsx"
		case strings.HasPrefix(name, "ppt/"):
			return "pptx"
		case name == "[content_types].xml":
			// Continue; this file is present in DOCX/XLSX/PPTX but doesn't
			// identify which Office Open XML family the archive belongs to.
			continue
		case name == "mimetype":
			return "odf-or-epub"
		case strings.HasPrefix(name, "meta-inf/"):
			return "epub"
		}
	}
	return "zip"
}

// parseError is the sidecar's error body: {"detail": {"reason": ..., "message": ...}}.
type parseError struct {
	Detail struct {
		Reason  string `json:"reason"`
		Message string `json:"message"`
	} `json:"detail"`
}

// Error is a document-conversion failure with a stable machine-readable reason.
//
// WHY a dedicated type: the admin UI shows Message verbatim while logs and
// metrics group by Reason (unsupported_format, pdf_encrypted, pdf_no_text,
// too_large, corrupt, timeout, extractor_unavailable...). A bare error string
// would force the UI to parse prose.
type Error struct {
	Reason  string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("document extraction failed (%s): %s", e.Reason, e.Message)
}

// Result is the extracted document.
type Result struct {
	Text     string
	Format   string
	Chars    int
	Pages    int
	Warnings []string
}

// Extractor converts documents to text. A nil *Extractor is usable: it only
// handles plain-text formats (sidecar not configured).
type Extractor struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewExtractor builds an extractor pointed at the ML sidecar.
//
// baseURL is the same ML_SIDECAR_URL the embedder uses. A separate HTTP client
// (rather than reusing ml.SidecarClient) because document conversion needs its
// own, much longer timeout than an embedding call and should not be capped by
// it. Empty baseURL disables the sidecar-backed formats.
func NewExtractor(baseURL string, timeoutSeconds int, logger *zap.Logger) *Extractor {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 300
	}
	trimmed := strings.TrimRight(baseURL, "/")
	if trimmed == "" {
		logger.Warn("document extraction: ML sidecar URL is empty — " +
			"office/PDF formats are disabled, plain text still works")
		return nil
	}
	return &Extractor{
		baseURL:    trimmed,
		httpClient: &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second},
		logger:     logger,
	}
}

// Extract converts an uploaded file to text.
//
// Mental execution:
//   Input:  "lecture.pdf", 12MB of PDF bytes
//   1. Allowlist the extension (fast, before any network call).
//   2. Plain text (.txt/.md/...) → decode here, no sidecar needed.
//   3. Anything else → POST the file to the sidecar's /extract.
//
// Error cases:
//   * *Error{unsupported_format} — extension not in the allowlist
//   * *Error{extractor_unavailable} — sidecar unreachable/misconfigured
//   * *Error{<sidecar reason>} — passthrough of the sidecar's own diagnosis
func (e *Extractor) Extract(ctx context.Context, filename string, data []byte) (*Result, error) {
	ext := Extension(filename)
	family, supported := formatByExtension[ext]
	if !supported {
		return nil, &Error{
			Reason: "unsupported_format",
			Message: fmt.Sprintf(".%s is not a supported format. Supported: %s",
				ext, SupportedList()),
		}
	}
	if len(data) == 0 {
		return nil, &Error{Reason: "empty", Message: "The uploaded file is empty."}
	}
	if len(data) > maxUploadBytes {
		return nil, &Error{
			Reason:  "too_large",
			Message: fmt.Sprintf("File is %d bytes, above the %d limit.", len(data), maxUploadBytes),
		}
	}
	if family == "text" {
		if binaryFamily := DetectBinaryFamily(data); binaryFamily != "" {
			return nil, &Error{
				Reason: "extension_mismatch",
				Message: fmt.Sprintf("This file is named .%s but its content looks like %s. " +
					"Rename it with the correct extension and retry.", ext, binaryFamily),
			}
		}
	}
	if ext == "xls" && DetectBinaryFamily(data) == "xlsx" {
		return nil, &Error{
			Reason:  "extension_mismatch",
			Message: "This file is named .xls but its content is an XLSX workbook. Rename it to .xlsx and upload again.",
		}
	}

	// Plain text never needs the sidecar. This is what keeps the original
	// .txt/.md workflow working during a sidecar outage.
	if plainTextExtensions[ext] {
		text := DecodeText(data)
		if strings.TrimSpace(text) == "" {
			return nil, &Error{Reason: "empty", Message: "The uploaded file has no readable text."}
		}
		return &Result{Text: text, Format: "text", Chars: utf8.RuneCountInString(text)}, nil
	}

	if e == nil {
		return nil, &Error{
			Reason: "extractor_unavailable",
			Message: fmt.Sprintf("Documents in .%s format need the document extractor, "+
				"which is not configured in this deployment. Upload .txt or .md instead.", ext),
		}
	}
	return e.extractViaSidecar(ctx, filename, family, data)
}

// sidecarResponse mirrors the /extract success body.
type sidecarResponse struct {
	Text     string   `json:"text"`
	Format   string   `json:"format"`
	Chars    int      `json:"chars"`
	Pages    int      `json:"pages"`
	Warnings []string `json:"warnings"`
}

func (e *Extractor) extractViaSidecar(ctx context.Context, filename, family string, data []byte) (*Result, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("build extract request: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return nil, fmt.Errorf("build extract request: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("build extract request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/extract", &body)
	if err != nil {
		return nil, fmt.Errorf("build extract request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := e.httpClient.Do(req)
	if err != nil {
		// Unreachable sidecar is an infrastructure problem, not a bad document.
		// Surfacing it distinctly lets the admin retry instead of re-uploading.
		return nil, &Error{
			Reason: "extractor_unavailable",
			Message: "Document extractor is unreachable: " + err.Error() +
				". Nothing was ingested — retry once the ML sidecar is healthy.",
		}
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read extract response: %w", err)
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		var parsed parseError
		if json.Unmarshal(payload, &parsed) == nil && parsed.Detail.Message != "" {
			e.logger.Info("document extraction rejected",
				zap.String("filename", filename),
				zap.String("reason", parsed.Detail.Reason),
			)
			return nil, &Error{Reason: parsed.Detail.Reason, Message: parsed.Detail.Message}
		}
		return nil, &Error{Reason: "corrupt", Message: strings.TrimSpace(string(payload))}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &Error{
			Reason: "extractor_error",
			Message: fmt.Sprintf("Document extractor returned status %d: %s",
				resp.StatusCode, truncate(string(payload), 300)),
		}
	}

	var parsed sidecarResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return nil, fmt.Errorf("decode extract response: %w", err)
	}
	if strings.TrimSpace(parsed.Text) == "" {
		return nil, &Error{
			Reason:  "empty",
			Message: "No text could be extracted from this document.",
		}
	}

	format := parsed.Format
	if format == "" {
		format = family
	}
	return &Result{
		Text:     parsed.Text,
		Format:   format,
		Chars:    parsed.Chars,
		Pages:    parsed.Pages,
		Warnings: parsed.Warnings,
	}, nil
}

// DecodeText decodes a text upload.
//
// WHY not just string(bytes): real transcripts arrive as UTF-16 (Windows
// exports) or windows-1252/latin-1 often enough that a strict UTF-8 assumption
// silently produces mojibake ("cafÃ©"), which then gets embedded as-is.
func DecodeText(data []byte) string {
	// Byte-order marks first — they are unambiguous.
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		return string(data[3:])
	}
	if bytes.HasPrefix(data, []byte{0xFF, 0xFE}) {
		return decodeUTF16(data[2:], true)
	}
	if bytes.HasPrefix(data, []byte{0xFE, 0xFF}) {
		return decodeUTF16(data[2:], false)
	}
	if utf8.Valid(data) {
		return string(data)
	}
	// Not valid UTF-8 and no BOM: fall back to latin-1, which maps every byte
	// to a rune and therefore never fails (same posture as the sidecar).
	var builder strings.Builder
	builder.Grow(len(data))
	for _, b := range data {
		builder.WriteRune(rune(b))
	}
	return builder.String()
}

func decodeUTF16(data []byte, littleEndian bool) string {
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}
	units := make([]uint16, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		if littleEndian {
			units = append(units, uint16(data[i])|uint16(data[i+1])<<8)
		} else {
			units = append(units, uint16(data[i])<<8|uint16(data[i+1]))
		}
	}
	return string(utf16.Decode(units))
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}
