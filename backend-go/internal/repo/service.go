package repo

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/ml"
	"ai_avengers/backend/internal/response"
	"ai_avengers/backend/internal/tenant"
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
	embedder      ml.Embedder // ml.Embedder interface: sidecar or CodeCraftAPI, resolved at call time
	chunker       *training.TextChunker
	encryptionKey []byte
	githubOAuth   OAuthConfig
	gitlabOAuth   OAuthConfig
	httpClient    *http.Client
	redis         *redis.Client // for OAuth state CSRF protection
	// frontendURL fix (feature #6, docs bug list): OAuthCallback used to
	// return raw JSON to what is actually a top-level browser navigation
	// (the OAuth provider redirects the browser here directly) - the
	// user saw a bare JSON page instead of landing back in the SPA. This
	// is the base URL OAuthCallback redirects back to.
	frontendURL string
	logger      *zap.Logger
}

// NewService creates a new repo integration service.
// embedder satisfies ml.Embedder — either *ml.SidecarClient (default) or
// *ml.DynamicEmbedder (when CodeCraftAPI embeddings are enabled).
func NewService(
	db *pgxpool.Pool,
	embedder ml.Embedder,
	redisClient *redis.Client,
	encryptionKey string,
	githubClientID, githubClientSecret string,
	gitlabClientID, gitlabClientSecret string,
	baseURL string,
	frontendURL string,
	logger *zap.Logger,
) *Service {
	return &Service{
		db:       db,
		embedder: embedder,
		chunker:  training.NewTextChunker(training.DefaultChunkerConfig()),
		encryptionKey: []byte(encryptionKey),
		redis:    redisClient,
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
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		frontendURL: frontendURL,
		logger:      logger,
	}
}

// oauthStateTTL is how long an OAuth state is valid.
// 10 minutes is generous for a human to complete the OAuth flow.
const oauthStateTTL = 10 * time.Minute

// oauthStateKey builds the Redis key for an OAuth state.
func oauthStateKey(state string) string {
	return "oauth_state:" + state
}

// GetOAuthURL returns the OAuth authorization URL for a provider.
// Stores state in Redis for CSRF validation in OAuthCallback.
//
// Mental execution:
// 1. Validate provider
// 2. Store state -> provider in Redis (TTL 10min)
// 3. Return authorization URL with state param
func (s *Service) GetOAuthURL(ctx context.Context, provider, state, projectID string) (string, error) {
	var authURL string
	switch provider {
	case ProviderGitHub:
		authURL = fmt.Sprintf(
			"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=repo&state=%s",
			s.githubOAuth.ClientID,
			url.QueryEscape(s.githubOAuth.RedirectURL),
			state,
		)
	case ProviderGitLab:
		authURL = fmt.Sprintf(
			"https://gitlab.com/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=read_repository&state=%s",
			s.gitlabOAuth.ClientID,
			url.QueryEscape(s.gitlabOAuth.RedirectURL),
			state,
		)
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}

	// Store state metadata in Redis for CSRF validation
	// Value: "provider:projectID" so callback knows both
	stateValue := provider + ":" + projectID
	if err := s.redis.Set(ctx, oauthStateKey(state), stateValue, oauthStateTTL).Err(); err != nil {
		s.logger.Warn("failed to store OAuth state in Redis", zap.Error(err))
		// Don't fail — OAuth can still work without CSRF check
		// but log it so we know Redis is having issues
	}

	return authURL, nil
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
	var status, repoURL, repoName, errorMsg string
	var totalChunks int
	var lastSync *time.Time

	err := s.db.QueryRow(ctx,
		`SELECT sync_status, repo_url, COALESCE(repo_name,''),
		        total_chunks, last_sync_at, COALESCE(error_message, '')
		 FROM repo_connections
		 WHERE project_id=$1
		 ORDER BY created_at DESC LIMIT 1`,
		projectID,
	).Scan(&status, &repoURL, &repoName, &totalChunks, &lastSync, &errorMsg)
	if err != nil {
		return map[string]interface{}{"connected": false}, nil
	}

	result := map[string]interface{}{
		"connected":    true,
		"status":       status,
		"repo_url":     repoURL,
		"repo_name":    repoName,
		"total_chunks": totalChunks,
		"last_sync_at": lastSync,
	}

	// Include error message if sync failed
	if errorMsg != "" {
		result["error_message"] = errorMsg
	}

	return result, nil
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
			embeddings, err := s.embedder.Embed(ctx, texts)
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

// ErrNoRepoConnection means the project has no git provider connected, so
// there is no token to act with. A distinct error because the caller's answer
// to the client is a specific instruction ("connect a repo first"), not a
// generic failure.
var ErrNoRepoConnection = errors.New("no repo connection for this project")

// ErrProviderMismatch means a token exists, but for a different provider than
// the caller asked for. Returning the wrong provider's token would send a
// GitHub token to GitLab (or the reverse) — a credential leak to a third
// party, not merely a failed request.
var ErrProviderMismatch = errors.New("the connected provider does not match the requested one")

// AccessTokenForProject returns the DECRYPTED access token stored for a
// project's git connection.
//
// Why this exists as an exported method:
//
// Two features outside this package need to act on a client's git remote with
// the client's own credentials: the harness git-push export (§18 of
// docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md) and the code-feedback ingest,
// which clones the repo the client built (§17). Both need exactly one thing
// from here — the token — and neither should own a second copy of the
// encryption key or of the decrypt routine. Exporting the narrowest possible
// accessor keeps encrypt/decrypt and s.encryptionKey private to this package,
// which is the property migration 001 stored the token encrypted to protect.
//
// The returned token is for immediate in-memory use. Callers must never write
// it to disk, into a git config, or into a log line.
func (s *Service) AccessTokenForProject(ctx context.Context, projectID uuid.UUID, provider string) (string, error) {
	var encrypted, storedProvider string
	err := s.db.QueryRow(ctx,
		`SELECT COALESCE(access_token, ''), provider
		   FROM repo_connections
		  WHERE project_id = $1
		  ORDER BY created_at DESC
		  LIMIT 1`,
		projectID,
	).Scan(&encrypted, &storedProvider)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoRepoConnection
	}
	if err != nil {
		return "", fmt.Errorf("load repo connection: %w", err)
	}
	if encrypted == "" {
		// A row can exist with a NULL token: access_token is nullable and
		// ConnectRepo is not the only way a row is created.
		return "", ErrNoRepoConnection
	}
	if storedProvider != provider {
		return "", fmt.Errorf("%w: connected %s, asked for %s", ErrProviderMismatch, storedProvider, provider)
	}

	token, err := s.decrypt(encrypted)
	if err != nil {
		// Deliberately does not wrap the decrypt error's text into something
		// that could carry ciphertext into a log. The cause is almost always a
		// changed ENCRYPTION_KEY, and that is what the message should say.
		return "", fmt.Errorf("stored token could not be decrypted — was ENCRYPTION_KEY changed after the repo was connected?")
	}
	if strings.TrimSpace(token) == "" {
		return "", ErrNoRepoConnection
	}
	return token, nil
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
	tenant *tenant.Service
	logger *zap.Logger
}

// NewHandler creates a new repo handler.
//
// WHY tenant is injected: every repo endpoint takes a project_id from the
// request. Without an ownership/tenant check a caller who knows (or guesses)
// another project's UUID can read its repo status — the same IDOR class the
// message handler closed with AssertProject (A2).
func NewHandler(svc *Service, tenantSvc *tenant.Service, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, tenant: tenantSvc, logger: logger}
}

// assertProjectAccess resolves the caller's tenant scope and verifies the
// project is visible to them. It writes the HTTP error itself and returns
// false so every caller reads as `if !h.assertProjectAccess(...) { return }`.
// Mirrors message.Handler's C4 ownership check.
func (h *Handler) assertProjectAccess(c *gin.Context, projectID uuid.UUID) bool {
	clientID, ok := c.MustGet("user_id").(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "invalid session")
		return false
	}
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	scope, err := h.tenant.Resolve(c.Request.Context(), clientID, roleStr)
	if err != nil {
		response.Forbidden(c, "Tenant scope could not be resolved")
		return false
	}
	if err := h.tenant.AssertProject(c.Request.Context(), scope, projectID); err != nil {
		response.Forbidden(c, err.Error())
		return false
	}
	return true
}

// GetOAuthURL GET /repo/oauth/:provider?project_id=<uuid>
// Returns the OAuth authorization URL. Frontend redirects user there.
//
// Mental execution:
// Client calls GET /repo/oauth/github?project_id=abc-123
// -> state UUID generated
// -> state stored in Redis: oauth_state:{state} = "github:abc-123"
// -> returns {url: "https://github.com/login/oauth/authorize?...", state: "..."}
// -> frontend redirects browser to url
func (h *Handler) GetOAuthURL(c *gin.Context) {
	provider := c.Param("provider")
	projectID := c.Query("project_id")
	if projectID == "" {
		response.BadRequest(c, "MISSING_PROJECT_ID", "project_id query param required")
		return
	}
	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		response.BadRequest(c, "INVALID_PROJECT_ID", "project_id must be a valid UUID")
		return
	}
	// A2: only the project's owner may start an OAuth handshake. This also
	// bounds OAuthCallback (public, no JWT): its state can only have been
	// minted here, by an authenticated caller who already passed this check.
	if !h.assertProjectAccess(c, parsedProjectID) {
		return
	}

	state := uuid.New().String() // CSRF protection
	oauthURL, err := h.svc.GetOAuthURL(c.Request.Context(), provider, state, projectID)
	if err != nil {
		response.BadRequest(c, "INVALID_PROVIDER", err.Error())
		return
	}
	response.OK(c, map[string]string{"url": oauthURL, "state": state})
}

// OAuthCallback GET /repo/callback/:provider?code=...&state=...
// Called by GitHub/GitLab after user authorizes the app.
// No JWT — this is a browser redirect from the OAuth provider.
//
// Feature #6 fix (docs bug list): every exit path below now issues an
// HTTP redirect (c.Redirect) back into the frontend SPA instead of
// response.* JSON helpers. WHY: this URL is what the OAuth PROVIDER
// redirects the user's BROWSER to directly (a top-level navigation),
// not an XHR/fetch call from React - returning JSON left the user
// staring at a bare JSON page after authorizing, with no way back
// into the app. Failure paths redirect to the project page (once the
// project ID is known from state) or the app root (before that),
// each with a `repoError` query param the frontend can read and show
// as a toast/banner; success redirects to the project page with
// `repoConnected=true`.
func (h *Handler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	frontendBase := h.svc.frontendURL

	if code == "" || state == "" {
		c.Redirect(http.StatusFound, frontendBase+"/?repoError=missing_params")
		return
	}

	// Validate state (CSRF check)
	stateValue, err := h.svc.redis.Get(c.Request.Context(), oauthStateKey(state)).Result()
	if err != nil {
		c.Redirect(http.StatusFound, frontendBase+"/?repoError=invalid_state")
		return
	}

	// Delete state — one-time use
	_ = h.svc.redis.Del(c.Request.Context(), oauthStateKey(state))

	// Parse state value: "provider:projectID"
	parts := strings.SplitN(stateValue, ":", 2)
	if len(parts) != 2 || parts[0] != provider {
		c.Redirect(http.StatusFound, frontendBase+"/?repoError=state_mismatch")
		return
	}
	projectID, err := uuid.Parse(parts[1])
	if err != nil {
		c.Redirect(http.StatusFound, frontendBase+"/?repoError=invalid_project")
		return
	}
	// From here on the project is known - send failures back to that
	// specific project page rather than the generic app root, so the
	// user lands where they started instead of the project list.
	projectURL := fmt.Sprintf("%s/projects/%s", frontendBase, projectID)

	// Exchange code for access token
	accessToken, err := h.svc.ExchangeCode(c.Request.Context(), provider, code)
	if err != nil {
		h.logger.Error("OAuth code exchange failed",
			zap.String("provider", provider),
			zap.Error(err),
		)
		c.Redirect(http.StatusFound, projectURL+"?repoError=exchange_failed")
		return
	}

	// We don't have clientID here (no JWT) — look it up from the project
	// WHY: OAuth callback has no JWT. We trust the state param (CSRF validated)
	// to identify the project, and we look up the project owner.
	var clientID uuid.UUID
	err = h.svc.db.QueryRow(c.Request.Context(),
		`SELECT client_id FROM projects WHERE id=$1 AND deleted_at IS NULL`,
		projectID,
	).Scan(&clientID)
	if err != nil {
		c.Redirect(http.StatusFound, frontendBase+"/?repoError=project_not_found")
		return
	}

	// Connect repo with the obtained token
	// repo_url and default_branch will be fetched during sync
	// For now store a placeholder URL — sync will update it
	connID, err := h.svc.ConnectRepo(
		c.Request.Context(), projectID, clientID,
		provider, "", accessToken, "main",
	)
	if err != nil {
		h.logger.Error("ConnectRepo failed after OAuth", zap.Error(err))
		c.Redirect(http.StatusFound, projectURL+"?repoError=connect_failed")
		return
	}

	// Start background sync
	_ = h.svc.SyncRepo(c.Request.Context(), connID)

	h.logger.Info("OAuth repo connected",
		zap.String("provider", provider),
		zap.String("project_id", projectID.String()),
	)

	c.Redirect(http.StatusFound, projectURL+"?repoConnected=true")
}

// ConnectRepo POST /projects/:id/repo
func (h *Handler) ConnectRepo(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid project ID")
		return
	}
	if !h.assertProjectAccess(c, projectID) {
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
	if !h.assertProjectAccess(c, projectID) {
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
	if !h.assertProjectAccess(c, projectID) {
		return
	}
	status, err := h.svc.GetSyncStatus(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, status)
}
