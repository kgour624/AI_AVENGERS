package workflow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
)

// PublishedDesign is a share token a client can hand to the landing-page team.
//
// WHY a token instead of /admin/<adminId>/...: this product has several admin
// accounts, and an endpoint keyed by admin would mean one integration per admin.
// The token is the primary key of the lookup, so there is exactly ONE public
// endpoint for the whole product: the landing page holds a base URL plus a token
// per published design. A second admin publishing a second design does not add
// an endpoint — it adds a row.
type PublishedDesign struct {
	ID         uuid.UUID `json:"id"`
	Token      string    `json:"token"`
	Title      string    `json:"title"`
	WorkflowID uuid.UUID `json:"workflow_id"`
	CreatedAt  string    `json:"created_at"`
}

// PublishedFile is one file inside a publication. Content is returned by the
// single-file read only, so a listing stays small no matter how big the design.
type PublishedFile struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Operation string `json:"operation,omitempty"`
}

const publicTokenPrefix = "av_pub_"

// newPublicToken returns an unguessable token. Random (not a readable slug) so
// nobody can walk published designs by guessing names; a human title is carried
// separately purely for display.
func newPublicToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return publicTokenPrefix + hex.EncodeToString(buf), nil
}

// fileContentsForWorkflow returns path -> (content, operation) for every file the
// workflow produced. Reuses workflowFiles() so the FILES tab, the chat file
// picker and publishing can never disagree about which files exist.
func (h *Handler) fileContentsForWorkflow(ctx context.Context, workflowID uuid.UUID) (map[string]publishedContent, []WorkflowFile, error) {
	files, err := h.workflowFiles(ctx, workflowID)
	if err != nil {
		return nil, nil, err
	}
	events, err := h.store.GetSince(ctx, workflowID, 0, 500)
	if err != nil {
		return nil, nil, err
	}
	contents := make(map[string]publishedContent, len(files))
	// Later events win: if a file was rewritten after a redesign or by a later
	// code-generation run, the published copy is the newest one, not the first.
	for _, e := range events {
		if e.PostedByExpertID == nil || structuralEvents[e.EventType] {
			continue
		}
		path, operation, hasBody := artifactPathAndBody(e.Content)
		if !hasBody {
			continue
		}
		if path == "" {
			path = e.EventType
		}
		body, ok := artifactBody(e.Content)
		if !ok {
			continue
		}
		contents[path] = publishedContent{Body: body, Operation: operation}
	}
	return contents, files, nil
}

type publishedContent struct {
	Body      string
	Operation string
}

// CreatePublication POST /admin/workflows/:id/publications
//
// Publishes the selected files under a fresh token. Selecting nothing is a
// request error rather than an empty publication: an empty share looks like a
// broken landing page.
func (h *Handler) CreatePublication(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}
	var req struct {
		Title string   `json:"title"`
		Paths []string `json:"paths"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Paths) == 0 {
		response.BadRequest(c, "INVALID_REQUEST", "title and at least one file path are required")
		return
	}
	userID, _ := c.Get("user_id")
	ctx := context.Background()

	contents, _, err := h.fileContentsForWorkflow(ctx, workflowID)
	if err != nil {
		h.logger.Error("publish: read workflow files failed", zap.String("workflow_id", workflowID.String()), zap.Error(err))
		response.InternalError(c)
		return
	}

	token, err := newPublicToken()
	if err != nil {
		response.InternalError(c)
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "Design"
	}

	var designID uuid.UUID
	err = h.engine.DB().QueryRow(ctx,
		`INSERT INTO public_designs (token, title, workflow_id, created_by)
		 VALUES ($1,$2,$3,$4) RETURNING id`,
		token, title, workflowID, userID).Scan(&designID)
	if err != nil {
		h.logger.Error("publish: insert design failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	added := h.addPublishedItems(ctx, designID, req.Paths, contents)
	response.OK(c, gin.H{"id": designID, "token": token, "title": title, "files_added": added})
}

// PushToPublication POST /admin/publications/:id/push
//
// Adds files to an existing publication. This is the "push" half: the landing
// page already points at the token, so pushing more files never changes the URL
// the landing page uses. Re-pushing the same path overwrites it (the unique key
// is (design_id, path)), which keeps the array free of duplicates.
func (h *Handler) PushToPublication(c *gin.Context) {
	designID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid publication ID")
		return
	}
	var req struct {
		Paths []string `json:"paths"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Paths) == 0 {
		response.BadRequest(c, "INVALID_REQUEST", "at least one file path is required")
		return
	}
	ctx := context.Background()

	var workflowID uuid.UUID
	if err := h.engine.DB().QueryRow(ctx,
		`SELECT workflow_id FROM public_designs WHERE id=$1`, designID).Scan(&workflowID); err != nil {
		response.NotFound(c, "publication not found")
		return
	}
	contents, _, err := h.fileContentsForWorkflow(ctx, workflowID)
	if err != nil {
		response.InternalError(c)
		return
	}
	added := h.addPublishedItems(ctx, designID, req.Paths, contents)
	response.OK(c, gin.H{"files_added": added})
}

// addPublishedItems upserts the requested paths and reports how many of them
// actually carried content. Paths the workflow never produced are skipped rather
// than stored empty.
func (h *Handler) addPublishedItems(ctx context.Context, designID uuid.UUID, paths []string, contents map[string]publishedContent) int {
	added := 0
	for _, path := range paths {
		pc, ok := contents[path]
		if !ok {
			continue
		}
		name := path
		if i := strings.LastIndex(path, "/"); i >= 0 {
			name = path[i+1:]
		}
		if _, err := h.engine.DB().Exec(ctx,
			`INSERT INTO public_design_items (design_id, path, name, operation, content)
			 VALUES ($1,$2,$3,$4,$5)
			 ON CONFLICT (design_id, path)
			 DO UPDATE SET content = EXCLUDED.content,
			               operation = EXCLUDED.operation,
			               added_at = NOW()`,
			designID, path, name, pc.Operation, pc.Body); err != nil {
			h.logger.Warn("publish: insert item failed", zap.String("path", path), zap.Error(err))
			continue
		}
		added++
	}
	return added
}

// ListPublications GET /admin/workflows/:id/publications
func (h *Handler) ListPublications(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}
	ctx := context.Background()
	rows, err := h.engine.DB().Query(ctx,
		`SELECT d.id, d.token, d.title, d.workflow_id, d.created_at,
		        (SELECT COUNT(*) FROM public_design_items i WHERE i.design_id = d.id)
		   FROM public_designs d
		  WHERE d.workflow_id = $1
		  ORDER BY d.created_at DESC`, workflowID)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()

	type publication struct {
		ID        uuid.UUID `json:"id"`
		Token     string    `json:"token"`
		Title     string    `json:"title"`
		FileCount int       `json:"file_count"`
		CreatedAt string    `json:"created_at"`
	}
	out := []publication{}
	for rows.Next() {
		var p publication
		var wid uuid.UUID
		if err := rows.Scan(&p.ID, &p.Token, &p.Title, &wid, &p.CreatedAt, &p.FileCount); err != nil {
			continue
		}
		out = append(out, p)
	}
	response.OK(c, out)
}

// GetPublicDesign GET /public/designs/:token
//
// The ONE integration point for the landing page. No authentication, no admin
// id: the token selects the publication, so the same URL shape serves every
// admin account. Read-only by construction — this handler has no write path.
func (h *Handler) GetPublicDesign(c *gin.Context) {
	token := c.Param("token")
	ctx := context.Background()

	var id uuid.UUID
	var title string
	if err := h.engine.DB().QueryRow(ctx,
		`SELECT id, title FROM public_designs WHERE token=$1`, token).Scan(&id, &title); err != nil {
		response.NotFound(c, "published design not found")
		return
	}

	rows, err := h.engine.DB().Query(ctx,
		`SELECT path, name, operation, added_at
		   FROM public_design_items
		  WHERE design_id=$1
		  ORDER BY path ASC`, id)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()

	files := []gin.H{}
	for rows.Next() {
		var path, name, operation, addedAt string
		if err := rows.Scan(&path, &name, &operation, &addedAt); err != nil {
			continue
		}
		files = append(files, gin.H{"path": path, "name": name, "operation": operation, "updated_at": addedAt})
	}
	response.OK(c, gin.H{"title": title, "files": files})
}

// GetPublicDesignFile GET /public/designs/:token/file?path=...
//
// Content is fetched per file so the landing page can render a viewer without
// downloading the whole design, and the listing above stays cheap.
func (h *Handler) GetPublicDesignFile(c *gin.Context) {
	token := c.Param("token")
	path := c.Query("path")
	if strings.TrimSpace(path) == "" {
		response.BadRequest(c, "INVALID_REQUEST", "path query parameter is required")
		return
	}
	ctx := context.Background()

	var content string
	if err := h.engine.DB().QueryRow(ctx,
		`SELECT i.content
		   FROM public_design_items i
		   JOIN public_designs d ON d.id = i.design_id
		  WHERE d.token=$1 AND i.path=$2`, token, path).Scan(&content); err != nil {
		response.NotFound(c, "file not found in this publication")
		return
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", `inline; filename="`+path+`"`)
	c.String(200, content)
}

// GetWorkflowFileContent GET /workflows/:id/file-content?path=...
//
// The admin-side download. Same source as publishing (workflowFiles), so the
// FILES tab, the download button and a published share never disagree about
// what a file contains — and a file produced later (redesign, code generation)
// is included because the map is rebuilt from the full event log each time.
func (h *Handler) GetWorkflowFileContent(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}
	path := strings.TrimSpace(c.Query("path"))
	if path == "" {
		response.BadRequest(c, "INVALID_REQUEST", "path query parameter is required")
		return
	}
	contents, _, err := h.fileContentsForWorkflow(context.Background(), workflowID)
	if err != nil {
		response.InternalError(c)
		return
	}
	pc, ok := contents[path]
	if !ok {
		response.NotFound(c, "file not found in this workflow")
		return
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filepath.Base(path)+`"`)
	c.String(200, pc.Body)
}

// GetCombinedDesignDoc GET /workflows/:id/design-doc
//
// Serves the current unified design document (final_design.md, or the latest
// redesign_N.md after a redesign) straight off disk — the "Download Unified
// Design" button's one endpoint.
//
// WHY read the file from disk rather than from the blackboard event's own
// content: writeCombinedDesignDoc (design_doc.go) does not put the markdown
// body into the event it posts — only file_path/file_name/version, the same
// choice export_git.go and the ZIP download already made for everything else
// under the workspace main/ directory. The event is the pointer; the
// workspace file is still the one source of truth for content, exactly like
// GetWorkflowFileContent above treats blackboard content as truth for
// per-expert artifacts (a deliberate difference in *where* truth lives for a
// system-generated document versus an expert-authored one, not an
// inconsistency).
func (h *Handler) GetCombinedDesignDoc(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}
	ctx := c.Request.Context()

	events, err := h.store.GetSince(ctx, workflowID, 0, 500)
	if err != nil {
		h.logger.Error("get combined design doc: list events failed",
			zap.String("workflow_id", workflowID.String()), zap.Error(err))
		response.InternalError(c)
		return
	}
	filePath, fileName, _, found := lookupCombinedDesignDoc(events)
	if !found {
		response.NotFound(c, "no combined design document has been generated for this workflow yet")
		return
	}

	content, readErr := os.ReadFile(filePath)
	if readErr != nil {
		// The event exists but the file is gone (workspace cleanup, moved
		// volume, etc). This is a real server-side problem, not a 404 for a
		// document that was simply never generated — logged distinctly so
		// the two cases are not confused when triaging.
		h.logger.Error("get combined design doc: file missing on disk",
			zap.String("workflow_id", workflowID.String()),
			zap.String("file_path", filePath), zap.Error(readErr))
		response.InternalError(c)
		return
	}

	if fileName == "" {
		fileName = filepath.Base(filePath)
	}
	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+fileName+`"`)
	c.String(200, string(content))
}
