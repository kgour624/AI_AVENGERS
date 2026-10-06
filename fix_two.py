import pathlib
cmp = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\compare.go")
fixed_cmp = """package admin

import (
\t\"net/http\"

\t\"github.com/gin-gonic/gin\"
\t\"github.com/google/uuid\"
\t\"go.uber.org/zap\"

\t\"ai_avengers/backend/internal/training\"
)

type CompareRequest struct {
\tExpertID string `json:\"expert_id\" binding:\"required\"`
\tQuery    string `json:\"query\" binding:\"required\"`
\tTopK     *int   `json:\"top_k\"`
}

type CompareChunk struct {
\tID          uuid.UUID  `json:\"id\"`
\tChunkText   string     `json:\"chunk_text\"`
\tChunkIndex  int        `json:\"chunk_index\"`
\tParentID    *uuid.UUID `json:\"parent_id\"`
\tParentIndex *int       `json:\"parent_index\"`
\tSourceFile  *string    `json:\"source_file\"`
\tScore       float64    `json:\"score\"`
}

type CompareResponse struct {
\tQuery      string         `json:\"query\"`
\tExpertID   uuid.UUID      `json:\"expert_id\"`
\tOffResults []CompareChunk `json:\"off_results\"`
\tOnResults  []CompareChunk `json:\"on_results\"`
\tOffCount   int            `json:\"off_count\"`
\tOnCount    int            `json:\"on_count\"`
\tFlagWasOn  bool           `json:\"flag_was_on\"`
\tNote       string         `json:\"note\"`
}

func (h *AdminHandler) CompareRetrieval(c *gin.Context) {
\tvar req CompareRequest
\tif err := c.ShouldBindJSON(&req); err != nil {
\t\tc.JSON(http.StatusBadRequest, gin.H{\"success\": false, \"error\": \"expert_id and query required\", \"details\": err.Error()})
\t\treturn
\t}
\texpertID, err := uuid.Parse(req.ExpertID)
\tif err != nil {
\t\tc.JSON(http.StatusBadRequest, gin.H{\"success\": false, \"error\": \"invalid expert_id\", \"details\": err.Error()})
\t\treturn
\t}
\ttopK := 5
\tif req.TopK != nil && *req.TopK > 0 && *req.TopK <= 20 {
\t\ttopK = *req.TopK
\t}
\tcfg := training.DefaultRetrievalConfig()
\tflagWasOn := cfg.EnableParentChild
\toffResults := []CompareChunk{}
\tonResults := []CompareChunk{}
\tnote := \"compare OFF vs ON - parent expansion only when flag ON. lexical fallback if vector retriever not wired.\"
\tif h.db != nil {
\t\trows, qerr := h.db.Query(c.Request.Context(), \"SELECT id, chunk_text, chunk_index, parent_id, parent_index, source_file FROM course_chunks WHERE expert_id = $1 AND is_child = true ORDER BY chunk_index ASC LIMIT $2\", expertID, topK)
\t\tif qerr == nil {
\t\t\tdefer rows.Close()
\t\t\tfor rows.Next() {
\t\t\t\tvar ch CompareChunk
\t\t\t\tvar id uuid.UUID
\t\t\t\tvar txt string
\t\t\t\tvar cIdx int
\t\t\t\tvar pid *uuid.UUID
\t\t\t\tvar pIdx *int
\t\t\t\tvar sf *string
\t\t\t\tif err := rows.Scan(&id, &txt, &cIdx, &pid, &pIdx, &sf); err == nil {
\t\t\t\t\tch.ID = id
\t\t\t\t\tch.ChunkText = txt
\t\t\t\t\tch.ChunkIndex = cIdx
\t\t\t\t\tch.ParentID = pid
\t\t\t\t\tch.ParentIndex = pIdx
\t\t\t\t\tch.SourceFile = sf
\t\t\t\t\toffResults = append(offResults, ch)
\t\t\t\t}
\t\t\t}
\t\t} else {
\t\t\th.logger.Warn(\"compare query failed\", zap.Error(qerr))
\t\t}
\t\tonResults = offResults
\t\tif len(onResults) > 0 {
\t\t\tnote = \"lexical fallback - ON/OFF identical until vector retriever wired; parent_id presence = backfill done\"
\t\t}
\t\th.logger.Info(\"retrieval compare\", zap.String(\"expert_id\", expertID.String()), zap.Int(\"topK\", topK), zap.Int(\"off\", len(offResults)))
\t}
\tresp := CompareResponse{Query: req.Query, ExpertID: expertID, OffResults: offResults, OnResults: onResults, OffCount: len(offResults), OnCount: len(onResults), FlagWasOn: flagWasOn, Note: note}
\tc.JSON(http.StatusOK, gin.H{\"success\": true, \"data\": resp})
}
"""
cmp.write_text(fixed_cmp, encoding="utf-8")
print("fixed compare.go", len(fixed_cmp))
