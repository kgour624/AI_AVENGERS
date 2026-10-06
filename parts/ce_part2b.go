 func (h *AdminHandler) GetExpertChunkTree(c *gin.Context) {
     expertID, err := uuid.Parse(c.Param("id"))
     if err != nil { response.BadRequest(c, "INVALID_ID", "invalid expert ID"); return }
     sourceFile := c.Query("source_file")
     parentWhere := `WHERE expert_id=$1`
     parentArgs := []interface{}{expertID}
     parentArgPos := 2
     if sourceFile != "" {
         parentWhere += fmt.Sprintf(` AND source_file=$%d`, parentArgPos)
         parentArgs = append(parentArgs, sourceFile); parentArgPos++
     }
     parentSQL := fmt.Sprintf(`SELECT ep.id, ep.expert_id, ep.page_index, ep.page_text, ep.section_path, ep.token_count, ep.source_file, ep.created_at, COALESCE(cc.cnt,0)
         FROM expert_pages ep LEFT JOIN (SELECT parent_id, COUNT(*) as cnt FROM course_chunks WHERE expert_id=$1 GROUP BY parent_id) cc ON cc.parent_id=ep.id
         %s ORDER BY ep.page_index ASC LIMIT 100`, parentWhere)
     rows, err := h.db.Query(c.Request.Context(), parentSQL, parentArgs...)
     if err != nil {
         h.logger.Error("chunk-tree parents query failed", zap.String("expert_id", expertID.String()), zap.Error(err))
         response.InternalError(c); return
     }
     parents := make([]parentExplorerRow, 0)
     for rows.Next() {
         var id, expertStr uuid.UUID
         var pageIndex, tokenCount, childCount int
         var pageText, sectionPath string
         var sourceFileVal *string
         var createdAt time.Time
         if err := rows.Scan(&id, &expertStr, &pageIndex, &pageText, &sectionPath, &tokenCount, &sourceFileVal, &createdAt, &childCount); err != nil {
             rows.Close(); h.logger.Error("scan tree parent failed", zap.Error(err)); response.InternalError(c); return
         }
         parents = append(parents, parentExplorerRow{
             ID: id.String(), ExpertID: expertStr.String(), PageIndex: pageIndex, PageText: pageText,
             SectionPath: sectionPath, TokenCount: tokenCount, SourceFile: sourceFileVal, CreatedAt: createdAt.Format(time.RFC3339), ChildCount: childCount,
         })
     }
     rows.Close()
     if err := rows.Err(); err != nil { h.logger.Error("parents rows err", zap.Error(err)); response.InternalError(c); return }
     childBase := `WHERE expert_id=$1 AND is_child=true`
     childArgs := []interface{}{expertID}
     childArgPos := 2
     if sourceFile != "" {
         childBase += fmt.Sprintf(` AND source_file=$%d`, childArgPos)
         childArgs = append(childArgs, sourceFile); childArgPos++
     }
     childSQL := `SELECT id, expert_id, chunk_index, chunk_text, topic, subtopic, source_file, chunk_hash, section_path, parent_id, parent_index, is_child, created_at FROM course_chunks ` + childBase + fmt.Sprintf(` ORDER BY COALESCE(parent_index, 999999), chunk_index ASC LIMIT 500`)
     cRows, err := h.db.Query(c.Request.Context(), childSQL, childArgs...)
     if err != nil {
         h.logger.Error("chunk-tree children query failed", zap.String("expert_id", expertID.String()), zap.Error(err))
         response.InternalError(c); return
     }
     defer cRows.Close()
     childrenByParent := make(map[string][]chunkExplorerRow)
     var orphans []chunkExplorerRow
     var children []chunkExplorerRow
     for cRows.Next() {
         var id, expertStr uuid.UUID
         var chunkIndex int
         var chunkText string
         var topic, subtopic, sourceFileVal, chunkHash *string
         var sectionPath string
         var parentID *uuid.UUID
         var parentIdx *int
         var isChild bool
         var createdAt time.Time
         if err := cRows.Scan(&id, &expertStr, &chunkIndex, &chunkText, &topic, &subtopic, &sourceFileVal, &chunkHash, &sectionPath, &parentID, &parentIdx, &isChild, &createdAt); err != nil {
             h.logger.Error("scan tree child failed", zap.Error(err)); response.InternalError(c); return
         }
         r := chunkExplorerRow{
             ID: id.String(), ExpertID: expertStr.String(), ChunkIndex: chunkIndex, ChunkText: chunkText,
             Topic: topic, Subtopic: subtopic, SourceFile: sourceFileVal, ChunkHash: chunkHash,
             SectionPath: sectionPath, IsChild: isChild, TokenEst: len(chunkText) / 4, CreatedAt: createdAt.Format(time.RFC3339),
         }
         if r.TokenEst < 1 && len(chunkText) > 0 { r.TokenEst = 1 }
         if parentID != nil { s := parentID.String(); r.ParentID = &s }
         if parentIdx != nil { r.ParentIndex = parentIdx }
         if r.ParentID == nil { orphans = append(orphans, r)
         } else { childrenByParent[*r.ParentID] = append(childrenByParent[*r.ParentID], r) }
         children = append(children, r)
     }
     if err := cRows.Err(); err != nil { h.logger.Error("children rows err", zap.Error(err)); response.InternalError(c); return }
     var totalChildren, totalParents int
     _ = h.db.QueryRow(c.Request.Context(), `SELECT COUNT(*) FROM course_chunks WHERE expert_id=$1 AND is_child=true`, expertID).Scan(&totalChildren)
     _ = h.db.QueryRow(c.Request.Context(), `SELECT COUNT(*) FROM expert_pages WHERE expert_id=$1`, expertID).Scan(&totalParents)
     type treeNode struct {
         Parent   parentExplorerRow  `json:"parent"`
         Children []chunkExplorerRow `json:"children"`
     }
     tree := make([]treeNode, 0, len(parents))
     for _, p := range parents {
         ch := childrenByParent[p.ID]
         if ch == nil { ch = []chunkExplorerRow{} }
         tree = append(tree, treeNode{Parent: p, Children: ch})
     }
     if orphans == nil { orphans = []chunkExplorerRow{} }
     response.OK(c, gin.H{
         "tree": tree, "orphans": orphans, "total_children": totalChildren, "total_parents": totalParents,
         "parents_shown": len(parents), "children_shown": len(children),
     })
 }
