
 // ============================================================
 // CHUNK EXPLORER (Phase 1.5 Observability) - read-only parent-child inspector
 // ============================================================
 type chunkExplorerRow struct {
     ID          string  `json:"id"`
     ExpertID    string  `json:"expert_id"`
     ChunkIndex  int     `json:"chunk_index"`
     ChunkText   string  `json:"chunk_text"`
     Topic       *string `json:"topic,omitempty"`
     Subtopic    *string `json:"subtopic,omitempty"`
     SourceFile  *string `json:"source_file,omitempty"`
     ChunkHash   *string `json:"chunk_hash,omitempty"`
     SectionPath string  `json:"section_path"`
     ParentID    *string `json:"parent_id,omitempty"`
     ParentIndex *int    `json:"parent_index,omitempty"`
     IsChild     bool    `json:"is_child"`
     TokenEst    int     `json:"token_estimate"`
     CreatedAt   string  `json:"created_at"`
 }
 type parentExplorerRow struct {
     ID          string  `json:"id"`
     ExpertID    string  `json:"expert_id"`
     PageIndex   int     `json:"page_index"`
     PageText    string  `json:"page_text"`
     SectionPath string  `json:"section_path"`
     TokenCount  int     `json:"token_count"`
     SourceFile  *string `json:"source_file,omitempty"`
     CreatedAt   string  `json:"created_at"`
     ChildCount  int     `json:"child_count"`
 }
 func (h *AdminHandler) ListExpertChunks(c *gin.Context) {
     expertID, err := uuid.Parse(c.Param("id"))
     if err != nil {
         response.BadRequest(c, "INVALID_ID", "invalid expert ID")
         return
     }
     limit := 20
     if v := c.Query("limit"); v != "" {
         if n, err := strconv.Atoi(v); err == nil {
             if n < 1 { n = 1 }
             if n > 100 { n = 100 }
             limit = n
         }
     }
     offset := 0
     if v := c.Query("offset"); v != "" {
         if n, err := strconv.Atoi(v); err == nil && n >= 0 { offset = n }
     }
     var isChildFilter *bool
     if v := c.Query("is_child"); v != "" {
         b, err := strconv.ParseBool(v)
         if err != nil { response.BadRequest(c, "INVALID_FILTER", "is_child must be true or false"); return }
         isChildFilter = &b
     }
     var parentIDFilter *uuid.UUID
     if v := c.Query("parent_id"); v != "" {
         pid, err := uuid.Parse(v)
         if err != nil { response.BadRequest(c, "INVALID_FILTER", "parent_id must be a valid UUID"); return }
         parentIDFilter = &pid
     }
     sourceFile := c.Query("source_file")
     base := `FROM course_chunks WHERE expert_id=$1`
     args := []interface{}{expertID}
     argPos := 2
     if isChildFilter != nil {
         base += fmt.Sprintf(` AND is_child=$%d`, argPos)
         args = append(args, *isChildFilter); argPos++
     }
     if parentIDFilter != nil {
         base += fmt.Sprintf(` AND parent_id=$%d`, argPos)
         args = append(args, *parentIDFilter); argPos++
     }
     if sourceFile != "" {
         base += fmt.Sprintf(` AND source_file=$%d`, argPos)
         args = append(args, sourceFile); argPos++
     }
     var total int
     countSQL := `SELECT COUNT(*) ` + base
     if err := h.db.QueryRow(c.Request.Context(), countSQL, args...).Scan(&total); err != nil {
         h.logger.Error("list expert chunks count failed", zap.String("expert_id", expertID.String()), zap.Error(err))
         response.InternalError(c); return
     }
     selectSQL := `SELECT id, expert_id, chunk_index, chunk_text, topic, subtopic, source_file, chunk_hash, section_path, parent_id, parent_index, is_child, created_at ` + base + fmt.Sprintf(` ORDER BY chunk_index ASC LIMIT $%d OFFSET $%d`, argPos, argPos+1)
     args = append(args, limit, offset)
     rows, err := h.db.Query(c.Request.Context(), selectSQL, args...)
     if err != nil {
         h.logger.Error("list expert chunks query failed", zap.String("expert_id", expertID.String()), zap.Error(err))
         response.InternalError(c); return
     }
     defer rows.Close()
     out := make([]chunkExplorerRow, 0)
     for rows.Next() {
         var r chunkExplorerRow
         var id, expertStr uuid.UUID
         var topic, subtopic, sourceFileVal, chunkHash *string
         var parentID *uuid.UUID
         var parentIdx *int
         var sectionPath string
         var isChild bool
         var chunkText string
         var chunkIndex int
         var createdAt time.Time
         if err := rows.Scan(&id, &expertStr, &chunkIndex, &chunkText, &topic, &subtopic, &sourceFileVal, &chunkHash, &sectionPath, &parentID, &parentIdx, &isChild, &createdAt); err != nil {
             h.logger.Error("scan chunk row failed", zap.Error(err))
             response.InternalError(c); return
         }
         r.ID = id.String(); r.ExpertID = expertStr.String(); r.ChunkIndex = chunkIndex; r.ChunkText = chunkText
         r.Topic = topic; r.Subtopic = subtopic; r.SourceFile = sourceFileVal; r.ChunkHash = chunkHash; r.SectionPath = sectionPath
         if parentID != nil { s := parentID.String(); r.ParentID = &s }
         if parentIdx != nil { r.ParentIndex = parentIdx }
         r.IsChild = isChild; r.TokenEst = len(chunkText) / 4
         if r.TokenEst < 1 && len(chunkText) > 0 { r.TokenEst = 1 }
         r.CreatedAt = createdAt.Format(time.RFC3339)
         out = append(out, r)
     }
     if err := rows.Err(); err != nil { h.logger.Error("rows iteration failed", zap.Error(err)); response.InternalError(c); return }
     response.OK(c, gin.H{"items": out, "total": total, "limit": limit, "offset": offset})
 }

