package repo

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/ml"
	"ai_avengers/backend/internal/response"
	"ai_avengers/backend/internal/training"
)

// Provider constants
const (
	ProviderGitHub = "github"
	ProviderGitLab = "gitlab"
)

// Supported file extensions for code chunking
var supportedExtensions = map[string]string{
	".go":   "go",
	".py":   "python",
	".ts":   "typescript",
	".tsx":  "typescript",
	".js":   "javascript",
	".jsx":  "javascript",
	".java": "java",
	".md":   "markdown",
	".yaml": "yaml",
	".yml":  "yaml",
	".json": "json",
	".sql":  "sql",
	".rs":   "rust",
	".cpp":  "cpp",
	".c":    "c",
}

// Directories to skip during repo sync
var skipDirs = map[string]bool{
	"node_modules": true, ".git": true, "vendor": true,
	"dist": true, "build": true, ".next": true,
	"__pycache__": true, ".venv": true, "venv": true,
	"target": true, "bin": true, "obj": true,
}

// OAuthConfig holds OAuth2 configuration for a provider.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// RepoFile represents a file in a repository.
type RepoFile struct {
	Path     string
	Content  string
	Size     int
	Language string
}

// Service handles GitHub/GitLab integration.
//
// Flow:
// 1. Client initiates OAuth -> redirect to provider
// 2. Provider redirects back with code
// 3. Exchange code for access token
// 4. Store encrypted token in DB
// 5. Client triggers sync -> background job
// 6. Fetch file tree via API
// 7. Fetch file contents
// 8. Chunk + embed + store in repo_chunks
type Service struct {
	db            *pgxpool.Pool
	mlClient      *ml.SidecarClient
	chunker       *training.TextChunker
	encryptionKey []byte
	githubOAuth   OAuthConfig
	gitlabOAuth   OAuthConfig
	httpClient    *http.Client
	logger        *zap.Logger
}

// NewService creates a new repo integration service.
func NewService(
	db *pgxpool.Pool,
	mlClient *ml.SidecarClient,
	encryptionKey string,
	githubClientID, githubClientSecret string,
	gitlabClientID, gitlabClientSecret string,
	baseURL string,
	logger *zap.Logger,
) *Service {
	return &Service{
		db:       db,
		mlClient: mlClient,
		chunker:  training.NewTextChunker(training.DefaultChunkerConfig()),
		encryptionKey: []byte(encryptionKey),
		githubOAuth: OAuthConfig{
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			RedirectURL:  baseURL + "/api/v1/repo/callback/github",
		},
		gitlabOAuth: OAuthConfig{
			ClientID:     gitlabClientID,
			ClientSecret: gitlabClientSecret,
			RedirectURL:  baseURL + "/api/v1/repo/callback/gitlab",
		},
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger,
	}
}

// GetOAuthURL returns the OAuth authorization URL for a provider.
func (s *Service) GetOAuthURL(provider, state string) (string, error) {
	switch provider {
	case ProviderGitHub:
		return fmt.Sprintf(
			"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=repo&state=%s",
			s.githubOAuth.ClientID,
			url.QueryEscape(s.githubOAuth.RedirectURL),
			state,
		), nil
	case ProviderGitLab:
		return fmt.Sprintf(
			"https://gitlab.com/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=read_repository&state=%s",
			s.gitlabOAuth.ClientID,
			url.QueryEscape(s.gitlabOAuth.RedirectURL),
			state,
		), nil
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

// ExchangeCode exchanges OAuth code for access token.
func (s *Service) ExchangeCode(ctx context.Context, provider, code string) (string, error) {
	switch provider {
	case ProviderGitHub:
		return s.exchangeGitHubCode(ctx, code)
	case ProviderGitLab:
		return s.exchangeGitLabCode(ctx, code)
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

// ConnectRepo saves a repo connection with encrypted token.
func (s *Service) ConnectRepo(
	ctx context.Context,
	projectID, clientID uuid.UUID,
	provider, repoURL, accessToken, defaultBranch string,
) (uuid.UUID, error) {
	// Encrypt token before storing
	encryptedToken, err := s.encrypt(accessToken)
	if err != nil {
		return uuid.Nil, fmt.Errorf("encrypt token failed: %w", err)
	}

	// Extract repo name from URL
	repoName := extractRepoName(repoURL)

	var connID uuid.UUID
	err = s.db.QueryRow(ctx,
		`INSERT INTO repo_connections
			(project_id, client_id, provider, repo_url, repo_name, default_branch, access_token)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (project_id) DO UPDATE SET
			provider=EXCLUDED.provider,
			repo_url=EXCLUDED.repo_url,
			repo_name=EXCLUDED.repo_name,
			default_branch=EXCLUDED.default_branch,
			access_token=EXCLUDED.access_token,
			sync_status='pending',
			updated_at=NOW()
		 RETURNING id`,
		projectID, clientID, provider, repoURL, repoName, defaultBranch, encryptedToken,
	).Scan(&connID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("save connection failed: %w", err)
	}

	// Update project repo fields
	_, _ = s.db.Exec(ctx,
		`UPDATE projects SET
			repo_url=$1, repo_provider=$2, repo_branch=$3,
			repo_connected=TRUE, updated_at=NOW()
		 WHERE id=$4`,
		repoURL, provider, defaultBranch, projectID,
	)

	return connID, nil
}

// SyncRepo starts background repo sync.
// Returns immediately — sync runs in goroutine.
func (s *Service) SyncRepo(ctx context.Context, connectionID uuid.UUID) error {
	// Update status to syncing
	_, err := s.db.Exec(ctx,
		`UPDATE repo_connections SET sync_status='syncing', updated_at=NOW() WHERE id=$1`,
		connectionID,
	)
	if err != nil {
		return err
	}

	// Start background sync
	go s.runSync(context.Background(), connectionID)
	return nil
}

// GetSyncStatus returns the current sync status.
func (s *Service) GetSyncStatus(ctx context.Context, projectID uuid.UUID) (map[string]interface{}, error) {
	var status, repoURL, repoName string
	var totalChunks int
	var lastSync *time.Time

	err := s.db.QueryRow(ctx,
		`SELECT sync_status, repo_url, COALESCE(repo_name,''),
		        total_chunks, last_sync_at
		 FROM repo_connections
		 WHERE project_id=$1
		 ORDER BY created_at DESC LIMIT 1`,
		projectID,
	).Scan(&status, &repoURL, &repoName, &totalChunks, &lastSync)
	if err != nil {
		return map[string]interface{}{"connected": false}, nil
	}

	return map[string]interface{}{
		"connected":    true,
		"status":       status,
		"repo_url":     repoURL,
		"repo_name":    repoName,
		"total_chunks": totalChunks,
		"last_sync_at": lastSync,
	}, nil
}

// runSync performs the actual repo sync in background.
//
// Mental execution:
// 1. Load connection + decrypt token
// 2. Fetch file tree from GitHub/GitLab API
// 3. Filter relevant files (by extension, size)
// 4. Fetch file contents in batches
// 5. Chunk each file
// 6. Generate embeddings (batch)
// 7. Store in repo_chunks
// 8. Update sync status
func (s *Service) runSync(ctx context.Context, connectionID uuid.UUID) {
	s.logger.Info("repo sync started", zap.String("connection_id", connectionID.String()))

	// Load connection
	var projectID uuid.UUID
	var provider, repoURL, encryptedToken, branch string
	err := s.db.QueryRow(ctx,
		`SELECT project_id, provider, repo_url, access_token, default_branch
		 FROM repo_connections WHERE id=$1`,
		connectionID,
	).Scan(&projectID, &provider, &repoURL, &encryptedToken, &branch)
	if err != nil {
		s.updateSyncStatus(ctx, connectionID, "failed", "connection not found", 0)
		return
	}

	// Decrypt token
	accessToken, err := s.decrypt(encryptedToken)
	if err != nil {
		s.updateSyncStatus(ctx, connectionID, "failed", "token decryption failed", 0)
		return
	}

	// Fetch file tree
	files, err := s.fetchFileTree(ctx, provider, repoURL, branch, accessToken)
	if err != nil {
		s.logger.Error("fetch file tree failed", zap.Error(err))
		s.updateSyncStatus(ctx, connectionID, "failed", err.Error(), 0)
		return
	}

	s.logger.Info("files to process", zap.Int("count", len(files)))

	// Delete existing repo chunks for this connection
	_, _ = s.db.Exec(ctx,
		`DELETE FROM repo_chunks WHERE repo_connection_id=$1`,
		connectionID,
	)

	totalChunks := 0

	// Process files in batches of 5
	for i := 0; i < len(files); i += 5 {
		end := i + 5
		if end > len(files) {
			end = len(files)
		}
		batch := files[i:end]

		// Fetch content for batch
		var repoFiles []RepoFile
		for _, filePath := range batch {
			content, err := s.fetchFileContent(ctx, provider, repoURL, branch, filePath, accessToken)
			if err != nil {
				s.logger.Warn("fetch file content failed",
					zap.String("file", filePath), zap.Error(err))
				continue
			}
			lang := getLanguage(filePath)
			repoFiles = append(repoFiles, RepoFile{
				Path: filePath, Content: content,
				Size: len(content), Language: lang,
			})
		}

		// Chunk and embed
		for _, rf := range repoFiles {
			chunks := s.chunker.Chunk(rf.Content)
			if len(chunks) == 0 {
				continue
			}

			// Generate embeddings
			texts := make([]string, len(chunks))
			for j, ch := range chunks {
				texts[j] = ch.Text
			}
			embeddings, err := s.mlClient.Embed(ctx, texts)
			if err != nil {
				s.logger.Warn("embed failed for file", zap.String("file", rf.Path))
				continue
			}

			// Store chunks
			for j, chunk := range chunks {
				embedding := pgvector.NewVector(embeddings[j])
				_, err := s.db.Exec(ctx,
					`INSERT INTO repo_chunks
						(project_id, repo_connection_id, chunk_text, chunk_index,
						 file_path, file_language, file_size, embedding)
					 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
					projectID, connectionID, chunk.Text, chunk.Index,
					rf.Path, rf.Language, rf.Size, embedding,
				)
				if err != nil {
					s.logger.Warn("store chunk failed", zap.Error(err))
					continue
				}
				totalChunks++
			}
		}
	}

	s.updateSyncStatus(ctx, connectionID, "complete", "", totalChunks)
	s.logger.Info("repo sync complete",
		zap.Int("total_chunks", totalChunks),
		zap.String("connection_id", connectionID.String()),
	)
}

// fetchFileTree fetches the list of relevant files from a repo.
func (s *Service) fetchFileTree(ctx context.Context, provider, repoURL, branch, token string) ([]string, error) {
	switch provider {
	case ProviderGitHub:
		return s.fetchGitHubTree(ctx, repoURL, branch, token)
	case ProviderGitLab:
		return s.fetchGitLabTree(ctx, repoURL, branch, token)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// fetchGitHubTree fetches file tree from GitHub API.
func (s *Service) fetchGitHubTree(ctx context.Context, repoURL, branch, token string) ([]string, error) {
	// Extract owner/repo from URL
	ownerRepo := extractOwnerRepo(repoURL)
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/git/trees/%s?recursive=1", ownerRepo, branch)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var result struct {
		Tree []struct {
			Path string `json:"path"`
			Type string `json:"type"`
			Size int    `json:"size"`
		} `json:"tree"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var files []string
	for _, item := range result.Tree {
		if item.Type != "blob" {
			continue
		}
		if item.Size > 100*1024 { // Skip files > 100KB
			continue
		}
		if shouldSkipPath(item.Path) {
			continue
		}
		if !isSupportedFile(item.Path) {
			continue
		}
		files = append(files, item.Path)
	}
	return files, nil
}

// fetchGitLabTree fetches file tree from GitLab API.
func (s *Service) fetchGitLabTree(ctx context.Context, repoURL, branch, token string) ([]string, error) {
	projectPath := extractGitLabPath(repoURL)
	apiURL := fmt.Sprintf(
		"https://gitlab.com/api/v4/projects/%s/repository/tree?recursive=true&ref=%s&per_page=100",
		url.QueryEscape(projectPath), branch,
	)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitLab API request failed: %w", err)
	}
	defer resp.Body.Close()

	var items []struct {
		Path string `json:"path"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	var files []string
	for _, item := range items {
		if item.Type != "blob" {
			continue
		}
		if shouldSkipPath(item.Path) || !isSupportedFile(item.Path) {
			continue
		}
		files = append(files, item.Path)
	}
	return files, nil
}

// fetchFileContent fetches a single file's content.
func (s *Service) fetchFileContent(ctx context.Context, provider, repoURL, branch, filePath, token string) (string, error) {
	switch provider {
	case ProviderGitHub:
		ownerRepo := extractOwnerRepo(repoURL)
		apiURL := fmt.Sprintf(
			"https://api.github.com/repos/%s/contents/%s?ref=%s",
			ownerRepo, url.QueryEscape(filePath), branch,
		)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github.v3.raw")
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		content, err := io.ReadAll(resp.Body)
		return string(content), err

	case ProviderGitLab:
		projectPath := extractGitLabPath(repoURL)
		apiURL := fmt.Sprintf(
			"https://gitlab.com/api/v4/projects/%s/repository/files/%s/raw?ref=%s",
			url.QueryEscape(projectPath), url.QueryEscape(filePath), branch,
		)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		content, err := io.ReadAll(resp.Body)
		return string(content), err
	}
	return "", fmt.Errorf("unsupported provider")
}

// exchangeGitHubCode exchanges OAuth code for GitHub access token.
func (s *Service) exchangeGitHubCode(ctx context.Context, code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", s.githubOAuth.ClientID)
	data.Set("client_secret", s.githubOAuth.ClientSecret)
	data.Set("code", code)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://github.com/login/oauth/access_token",
		strings.NewReader(data.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("GitHub OAuth error: %s", result.Error)
	}
	return result.AccessToken, nil
}

// exchangeGitLabCode exchanges OAuth code for GitLab access token.
func (s *Service) exchangeGitLabCode(ctx context.Context, code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", s.gitlabOAuth.ClientID)
	data.Set("client_secret", s.gitlabOAuth.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", s.gitlabOAuth.RedirectURL)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://gitlab.com/oauth/token",
		strings.NewReader(data.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("GitLab OAuth error: %s", result.Error)
	}
	return result.AccessToken, nil
}

// updateSyncStatus updates the sync status in DB.
func (s *Service) updateSyncStatus(ctx context.Context, connectionID uuid.UUID, status, errMsg string, totalChunks int) {
	var lastSync interface{}
	if status == "complete" || status == "failed" {
		lastSync = time.Now()
	}
	_, _ = s.db.Exec(ctx,
		`UPDATE repo_connections SET
			sync_status=$1,
			total_chunks=$2,
			last_sync_at=$3,
			updated_at=NOW()
		 WHERE id=$4`,
		status, totalChunks, lastSync, connectionID,
	)
}

// encrypt encrypts a string using AES-256-GCM.
// WHY AES-256-GCM: Authenticated encryption — prevents tampering.
// OAuth tokens are sensitive — must be encrypted at rest.
func (s *Service) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts an AES-256-GCM encrypted string.
func (s *Service) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// Helper functions

func extractOwnerRepo(repoURL string) string {
	// https://github.com/owner/repo -> owner/repo
	repoURL = strings.TrimSuffix(repoURL, ".git")
	parts := strings.Split(repoURL, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	return repoURL
}

func extractGitLabPath(repoURL string) string {
	// https://gitlab.com/group/project -> group/project
	repoURL = strings.TrimSuffix(repoURL, ".git")
	if idx := strings.Index(repoURL, "gitlab.com/"); idx >= 0 {
		return repoURL[idx+len("gitlab.com/"):]
	}
	return repoURL
}

func extractRepoName(repoURL string) string {
	repoURL = strings.TrimSuffix(repoURL, ".git")
	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return repoURL
}

func shouldSkipPath(path string) bool {
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if skipDirs[part] {
			return true
		}
	}
	return false
}

func isSupportedFile(path string) bool {
	for ext := range supportedExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

func getLanguage(path string) string {
	for ext, lang := range supportedExtensions {
		if strings.HasSuffix(path, ext) {
			return lang
		}
	}
	return "unknown"
}

// Handler handles HTTP requests for repo integration.
type Handler struct {
	svc    *Service
	logger *zap.Logger
}

// NewHandler creates a new repo handler.
func NewHandler(svc *Service, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// GetOAuthURL GET /repo/oauth/:provider
func (h *Handler) GetOAuthURL(c *gin.Context) {
	provider := c.Param("provider")
	state := uuid.New().String() // CSRF protection
	oauthURL, err := h.svc.GetOAuthURL(provider, state)
	if err != nil {
		response.BadRequest(c, "INVALID_PROVIDER", err.Error())
		return
	}
	response.OK(c, map[string]string{"url": oauthURL, "state": state})
}

// ConnectRepo POST /projects/:id/repo
func (h *Handler) ConnectRepo(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid project ID")
		return
	}
	var req struct {
		Provider     string `json:"provider" binding:"required"`
		RepoURL      string `json:"repo_url" binding:"required"`
		AccessToken  string `json:"access_token" binding:"required"`
		DefaultBranch string `json:"default_branch"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if req.DefaultBranch == "" {
		req.DefaultBranch = "main"
	}
	connID, err := h.svc.ConnectRepo(
		c.Request.Context(), projectID, clientID,
		req.Provider, req.RepoURL, req.AccessToken, req.DefaultBranch,
	)
	if err != nil {
		h.logger.Error("connect repo failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.Created(c, map[string]interface{}{
		"connection_id": connID,
		"status":        "connected",
	})
}

// SyncRepo POST /projects/:id/repo/sync
func (h *Handler) SyncRepo(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid project ID")
		return
	}
	// Get connection ID
	var connID uuid.UUID
	err = h.svc.db.QueryRow(c.Request.Context(),
		`SELECT id FROM repo_connections WHERE project_id=$1 ORDER BY created_at DESC LIMIT 1`,
		projectID,
	).Scan(&connID)
	if err != nil {
		response.NotFound(c, "repo connection")
		return
	}
	if err := h.svc.SyncRepo(c.Request.Context(), connID); err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, map[string]string{"status": "sync started"})
}

// GetSyncStatus GET /projects/:id/repo/status
func (h *Handler) GetSyncStatus(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid project ID")
		return
	}
	status, err := h.svc.GetSyncStatus(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, status)
}
