package receptionist

import (
"bytes"
"fmt"
"io"
"strings"
"time"

"github.com/gin-gonic/gin"
"github.com/google/uuid"
)

func nowUTC() time.Time { return time.Now().UTC() }

// Upload - POST /sessions/:id/upload (multipart: file, optional checkpoint_id)
// FIX: chi -> gin (chi is not a dependency of this module).
func (h *Handler) Upload(c *gin.Context) {
sid, _ := uuid.Parse(c.Param("id"))
if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
writeJSON(c, 400, gin.H{"error": "file too large max 10MB"})
return
}
file, header, err := c.Request.FormFile("file")
if err != nil {
writeJSON(c, 400, gin.H{"error": "file required"})
return
}
defer file.Close()
var buf bytes.Buffer
if _, err := io.Copy(&buf, file); err != nil {
writeJSON(c, 500, gin.H{"error": err.Error()})
return
}
checkpointID := c.Request.FormValue("checkpoint_id")
var cidPtr *uuid.UUID
if checkpointID != "" {
if u, err := uuid.Parse(checkpointID); err == nil {
cidPtr = &u
}
}

text := ""
name := strings.ToLower(header.Filename)
switch {
case strings.HasSuffix(name, ".md"), strings.HasSuffix(name, ".txt"):
text = buf.String()
case strings.HasSuffix(name, ".pdf"):
text = fmt.Sprintf("[PDF %s bytes=%d] %s", header.Filename, buf.Len(), buf.String()[:min(2000, buf.Len())])
default:
text = fmt.Sprintf("[Image %s bytes=%d] (OCR placeholder)", header.Filename, buf.Len())
}
if len(text) > 8000 {
text = text[:8000] + "\n...[truncated]"
}

_ = h.store.AppendNote(c.Request.Context(), NoteEntry{
ID:          uuid.New(),
SessionID:   sid,
CheckpointID: cidPtr,
Type:        NoteFileUpload,
SummaryText: fmt.Sprintf("File %s: %s", header.Filename, text[:min(500, len(text))]),
Timestamp:   nowUTC(),
})
if cidPtr != nil {
_, _ = h.orch.HandleChatMessage(c.Request.Context(), sid, *cidPtr, fmt.Sprintf("[FILE %s]\n%s", header.Filename, text))
}
writeJSON(c, 200, gin.H{"status": "uploaded", "filename": header.Filename, "extracted_chars": len(text)})
}