 func (h *AdminHandler) ListExpertParents(c *gin.Context) {
     expertID, err := uuid.Parse(c.Param("id"))
     if err != nil { response.BadRequest(c, "INVALID_ID", "invalid expert ID"); return }
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
     sourceFile := c.Query("source_file")
     baseWhere := `WHERE expert_id=$1`
     args := []interface{}{expertID}
     argPos := 2
     if sourceFile != "" {
         baseWhere += fmt.Sprintf(` AND source_file=$%d`, argPos)
         args = append(args, sourceFile); argPos++
     }
     var total int
     if err := h.db.QueryRow(c.Request.Context(), `SELECT COUNT(*) FROM expert_pages `+baseWhere, args...).Scan(&total); err != nil {
         h.logger.Error("list parents count failed", zap.String("expert_id", expertID.String()), zap.Error(err))
         response.InternalError(c); return
     }
     selArgs := append(append([]interface{}{}, args...), limit, offset)
     selSQL := fmt.Sprintf(`SELECT ep.id, ep.expert_id, ep.page_index, ep.page_text, ep.section_path, ep.token_count, ep.source_file, ep.created_at, COALESCE(cc.cnt,0) as child_count
         FROM expert_pages ep
         LEFT JOIN (SELECT parent_id, COUNT(*) as cnt FROM course_chunks WHERE expert_id=$1 GROUP BY parent_id) cc ON cc.parent_id=ep.id
         %s ORDER BY ep.page_index ASC LIMIT $%d OFFSET $%d`, baseWhere, argPos, argPos+1)
     rows, err := h.db.Query(c.Request.Context(), selSQL, selArgs...)
     if err != nil {
         h.logger.Error("list parents query failed", zap.String("expert_id", expertID.String()), zap.Error(err))
         response.InternalError(c); return
     }
     defer rows.Close()
     out := make([]parentExplorerRow, 0)
     for rows.Next() {
         var id, expertStr uuid.UUID
         var pageIndex, tokenCount, childCount int
         var pageText, sectionPath string
         var sourceFileVal *string
         var createdAt time.Time
         if err := rows.Scan(&id, &expertStr, &pageIndex, &pageText, &sectionPath, &tokenCount, &sourceFileVal, &createdAt, &childCount); err != nil {
             h.logger.Error("scan parent row failed", zap.Error(err))
             response.InternalError(c); return
         }
         out = append(out, parentExplorerRow{
             ID: id.String(), ExpertID: expertStr.String(), PageIndex: pageIndex, PageText: pageText,
             SectionPath: sectionPath, TokenCount: tokenCount, SourceFile: sourceFileVal, CreatedAt: createdAt.Format(time.RFC3339), ChildCount: childCount,
         })
     }
     if err := rows.Err(); err != nil { h.logger.Error("parent rows iteration failed", zap.Error(err)); response.InternalError(c); return }
     response.OK(c, gin.H{"items": out, "total": total, "limit": limit, "offset": offset})
 }
