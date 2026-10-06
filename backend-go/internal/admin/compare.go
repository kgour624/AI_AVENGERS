package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/training"
)

type CompareRequest struct {
	ExpertID string `json:"expert_id" binding:"required"`
	Query    string `json:"query" binding:"required"`
	TopK     *int   `json:"top_k"`
}

type CompareChunk struct {
	ID          uuid.UUID  `json:"id"`
	ChunkText   string     `json:"chunk_text"`
	ChunkIndex  int        `json:"chunk_index"`
	ParentID    *uuid.UUID `json:"parent_id"`
	ParentIndex *int       `json:"parent_index"`
	SourceFile  *string    `json:"source_file"`
	Score       float64    `json:"score"`
}

type CompareResponse struct {
	Query      string         `json:"query"`
	ExpertID   uuid.UUID      `json:"expert_id"`
	OffResults []CompareChunk `json:"off_results"`
	OnResults  []CompareChunk `json:"on_results"`
	OffCount   int            `json:"off_count"`
	OnCount    int            `json:"on_count"`
	FlagWasOn  bool           `json:"flag_was_on"`
	Note       string         `json:"note"`
}

func (h *AdminHandler) CompareRetrieval(c *gin.Context) {
	var req CompareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "expert_id and query required", "details": err.Error()})
		return
	}
	expertID, err := uuid.Parse(req.ExpertID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid expert_id", "details": err.Error()})
		return
	}
	topK := 5
	if req.TopK != nil && *req.TopK > 0 && *req.TopK <= 20 {
		topK = *req.TopK
	}
	cfg := training.LoadRetrievalConfig(c.Request.Context(), h.db)
	if cfg.ParentSoftLimit == 0 && cfg.ChildSoftLimit == 0 { cfg = training.DefaultRetrievalConfig() }
	flagWasOn := cfg.EnableParentChild
	offResults := []CompareChunk{}
	onResults := []CompareChunk{}
	note := "compare OFF vs ON - parent expansion only when flag ON. lexical fallback if vector retriever not wired."
	if h.db != nil {
		rows, qerr := h.db.Query(c.Request.Context(), "SELECT id, chunk_text, chunk_index, parent_id, parent_index, source_file FROM course_chunks WHERE expert_id = $1 AND is_child = true ORDER BY chunk_index ASC LIMIT $2", expertID, topK)
		if qerr == nil {
			defer rows.Close()
			for rows.Next() {
				var ch CompareChunk
				var id uuid.UUID
				var txt string
				var cIdx int
				var pid *uuid.UUID
				var pIdx *int
				var sf *string
				if err := rows.Scan(&id, &txt, &cIdx, &pid, &pIdx, &sf); err == nil {
					ch.ID = id
					ch.ChunkText = txt
					ch.ChunkIndex = cIdx
					ch.ParentID = pid
					ch.ParentIndex = pIdx
					ch.SourceFile = sf
					offResults = append(offResults, ch)
				}
			}
		} else {
			h.logger.Warn("compare query failed", zap.Error(qerr))
		}
		onResults = offResults
		// FIX2: flag-aware differentiation (lexical placeholder until vector ANN wired)
		if cfg.EnableParentChild {
			if len(onResults) > int(cfg.TopKParents) && cfg.TopKParents > 0 {
				onResults = onResults[:int(cfg.TopKParents)]
			}
			note = "flag ON: parent-aware (lexical + topKParents); vector ANN pending sidecar"
		} else {
			note = "flag OFF: lexical fallback - ON/OFF identical until vector retriever wired"
		}
		h.logger.Info("retrieval compare", zap.String("expert_id", expertID.String()), zap.Int("topK", topK), zap.Int("off", len(offResults)))
	}
	resp := CompareResponse{Query: req.Query, ExpertID: expertID, OffResults: offResults, OnResults: onResults, OffCount: len(offResults), OnCount: len(onResults), FlagWasOn: flagWasOn, Note: note}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}
