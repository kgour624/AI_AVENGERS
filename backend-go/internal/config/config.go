package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
// Values are loaded from environment variables.
// No defaults for secrets — fail fast if missing.
type Config struct {
	Server              ServerConfig
	Database            DatabaseConfig
	Redis               RedisConfig
	JWT                 JWTConfig
	LLM                 LLMConfig
	ML                  MLConfig
	ChinaWall           ChinaWallConfig
	Context             ContextConfig
	Security            SecurityConfig
	OAuth               OAuthConfig
	Auth                AuthConfig
	RateLimit           RateLimitConfig
	Log                 LogConfig
	Provenance          ProvenanceConfig
	Versioning          VersioningConfig
	Tenant              TenantConfig
	CORSAllowedOrigins  []string // comma-separated in env: CORS_ALLOWED_ORIGINS
}

// TenantConfig controls C4 tenant isolation.
type TenantConfig struct {
	// IsolationEnabled: enforce tenant boundaries. Absent env → true
	// (IsSet pattern, same as B6/B8/C1/C2). Explicit false is the kill
	// switch — scopes resolve global and every assertion is a no-op, so a
	// bad rollout can be reverted without a redeploy. Migration 030
	// backfills all rows onto one 'default' tenant, so enabling this is
	// behaviour-neutral until a second tenant is used.
	IsolationEnabled bool
}

// VersioningConfig controls C2 expert versioning + drift detection.
type VersioningConfig struct {
	// Enabled: snapshot expert versions and detect drift. Absent env →
	// true (IsSet pattern, same as B6/B8/C1). Rollback switch.
	Enabled bool
	// DriftThreshold: topic-set (capability) Jaccard distance at/above
	// which drift is reported. Default 0.30.
	DriftThreshold float64
}

// ProvenanceConfig controls the C1 signed provenance chain.
type ProvenanceConfig struct {
	// Enabled: record a signed chain for answers/artifacts. Absent env →
	// true (IsSet pattern, same as B6/B8). Rollback switch.
	Enabled bool
	// SigningKey: HMAC key (any length). Empty → falls back to
	// Security.EncryptionKey in applyDefaults, so signing always has a
	// server secret without a new required env var.
	SigningKey string
}

// AuthConfig controls account-creation policy.
// SelfRegistrationEnabled=false (default) keeps POST /auth/register in the
// codebase but responds gracefully — accounts are provisioned by an admin.
type AuthConfig struct {
	SelfRegistrationEnabled bool
}

type ServerConfig struct {
	Host string
	Port int
	Env  string // development | production
}

type DatabaseConfig struct {
	URL                  string
	MaxConns             int32
	MinConns             int32
	MaxConnLifetime      time.Duration
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret              string
	AccessExpiryMinutes int
	RefreshExpiryDays   int
}

// LLMProvider identifies which LLM provider to use.
// openrouter = single key, multiple models (default)
// deepseek   = direct DeepSeek API
// anthropic  = direct Anthropic API
// gemini     = direct Google Gemini API
type LLMProvider string

const (
	ProviderOpenRouter   LLMProvider = "openrouter"
	ProviderDeepSeek     LLMProvider = "deepseek"
	ProviderAnthropic    LLMProvider = "anthropic"
	ProviderGemini       LLMProvider = "gemini"
	ProviderCodeCraftAPI LLMProvider = "codecraftapi"
	ProviderCavoti       LLMProvider = "cavoti" // OpenAI-compatible, sk- prefix, base: https://cavoti.com/v1
)

type LLMConfig struct {
	// OpenRouter (default multi-model gateway)
	OpenRouterAPIKey  string
	OpenRouterBaseURL string

	// Direct provider keys (optional — used when Provider != openrouter)
	DeepSeekAPIKey  string
	AnthropicAPIKey string
	GeminiAPIKey    string

	// CodeCraftAPI (multi-model aggregator with embeddings support)
	// Key prefix: cc_
	// Base URL default: https://codecraftapi.com/v1
	CodeCraftAPIKey     string // env: CODECRAFTAPI_KEY
	CodeCraftAPIBaseURL string // env: CODECRAFTAPI_BASE_URL

	// Cavoti (OpenAI-compatible, sk- prefix API key)
	// Base URL default: https://cavoti.com/v1
	CavotiAPIKey  string // env: CAVOTI_API_KEY
	CavotiBaseURL string // env: CAVOTI_BASE_URL

	// Active provider — can be overridden from admin panel via system_settings
	// Default: openrouter
	Provider LLMProvider

	// Model names per tier
	ModelCheap  string
	ModelStrong string
	ModelFast   string
}

type MLConfig struct {
	SidecarURL     string
	TimeoutSeconds int
}

type ChinaWallConfig struct {
	RerankerThreshold float64
	RelaxedThreshold  float64
	MaxRetries        int
	// QualityJudgeEnabled (B6): when true, flat-path answers are scored by
	// an LLM-as-judge after Layer 4. Off = zero extra calls (rollback switch).
	QualityJudgeEnabled bool
	// QualityFloor: minimum Overall score (0-1) to accept. Below this the
	// enforcer regenerates up to MaxQualityRetries times with judge feedback.
	QualityFloor float64
	// MaxQualityRetries: extra generate+judge cycles after the first. 0 =
	// score only, never regenerate. Default 1 (bounded, P10).
	MaxQualityRetries int
	// ClaimVerifyEnabled (B8): when true, flat-path answers run the
	// atomic claim→evidence pipeline after quality gate. Off = zero extra
	// calls (rollback switch). Absent env → true (same IsSet pattern as B6).
	ClaimVerifyEnabled bool
}

type ContextConfig struct {
	MaxTokens       int
	RecentMessages  int
	SemanticTopK    int
	CourseChunksTopK int
}

type SecurityConfig struct {
	EncryptionKey string // 32 bytes for AES-256
}

// OAuthConfig holds OAuth2 credentials for GitHub and GitLab repo integration.
// All fields are optional — empty means OAuth is disabled, PAT-only mode.
// WHY optional: PAT-based connect works without OAuth.
// OAuth is an enhancement, not a requirement for core functionality.
type OAuthConfig struct {
	GitHubClientID     string
	GitHubClientSecret string
	GitLabClientID     string
	GitLabClientSecret string
	BaseURL            string // e.g. https://api.yourdomain.com - used to build the redirect_uri GitHub/GitLab sends the browser BACK to (this backend's own /repo/callback route)
	// FrontendURL fix (feature #6, docs bug list): OAuthCallback
	// previously returned raw JSON as the response to a top-level
	// browser navigation (the OAuth provider redirects the browser
	// itself here, not an XHR call) - the user would see a bare JSON
	// page after authorizing, not land back in the SPA. This is where
	// the callback redirects the browser TO once the connection
	// succeeds or fails - a different URL than BaseURL above.
	FrontendURL string // e.g. https://app.yourdomain.com
}

type RateLimitConfig struct {
	PerIP   int
	PerUser int
}

type LogConfig struct {
	Level string // debug | info | warn | error
}

// Load reads configuration from environment variables.
// Panics if required values are missing — fail fast on startup.
func Load() (*Config, error) {
	v := viper.New()

	// Read from .env file if present (development only)
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Ignore error if .env not found — env vars may be set directly
	_ = v.ReadInConfig()

	cfg := &Config{
		Server: ServerConfig{
			Host: v.GetString("SERVER_HOST"),
			Port: v.GetInt("SERVER_PORT"),
			Env:  v.GetString("SERVER_ENV"),
		},
		Database: DatabaseConfig{
			URL:             v.GetString("DATABASE_URL"),
			MaxConns:        int32(v.GetInt("DB_MAX_CONNS")),
			MinConns:        int32(v.GetInt("DB_MIN_CONNS")),
			MaxConnLifetime: time.Duration(v.GetInt("DB_MAX_CONN_LIFETIME_MINUTES")) * time.Minute,
		},
		Redis: RedisConfig{
			URL:      v.GetString("REDIS_URL"),
			Password: v.GetString("REDIS_PASSWORD"),
			DB:       v.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			Secret:              v.GetString("JWT_SECRET"),
			AccessExpiryMinutes: v.GetInt("JWT_ACCESS_EXPIRY_MINUTES"),
			RefreshExpiryDays:   v.GetInt("JWT_REFRESH_EXPIRY_DAYS"),
		},
		// WHY defaults set here not in .env.example:
		//   viper.SetDefault applies when env var is missing.
		//   .env.example shows the value but doesn't set it.
		//   This ensures dev works out-of-box without editing .env.
		LLM: LLMConfig{
			OpenRouterAPIKey:    v.GetString("OPENROUTER_API_KEY"),
			OpenRouterBaseURL:   v.GetString("OPENROUTER_BASE_URL"),
			DeepSeekAPIKey:      v.GetString("DEEPSEEK_API_KEY"),
			AnthropicAPIKey:     v.GetString("ANTHROPIC_API_KEY"),
			GeminiAPIKey:        v.GetString("GEMINI_API_KEY"),
			CodeCraftAPIKey:     v.GetString("CODECRAFTAPI_KEY"),
			CodeCraftAPIBaseURL: v.GetString("CODECRAFTAPI_BASE_URL"),
			CavotiAPIKey:        v.GetString("CAVOTI_API_KEY"),
			CavotiBaseURL:       v.GetString("CAVOTI_BASE_URL"),
			Provider:            LLMProvider(v.GetString("LLM_PROVIDER")),
			ModelCheap:        v.GetString("LLM_MODEL_CHEAP"),
			ModelStrong:       v.GetString("LLM_MODEL_STRONG"),
			ModelFast:         v.GetString("LLM_MODEL_FAST"),
		},
		ML: MLConfig{
			SidecarURL:     v.GetString("ML_SIDECAR_URL"),
			TimeoutSeconds: v.GetInt("ML_SIDECAR_TIMEOUT_SECONDS"),
		},
		ChinaWall: ChinaWallConfig{
			RerankerThreshold:   v.GetFloat64("CHINA_WALL_RERANKER_THRESHOLD"),
			RelaxedThreshold:    v.GetFloat64("CHINA_WALL_RELAXED_THRESHOLD"),
			MaxRetries:          v.GetInt("CHINA_WALL_MAX_RETRIES"),
			QualityJudgeEnabled: true, // B6 default on; overridden below if env set
			QualityFloor:        v.GetFloat64("CHINA_WALL_QUALITY_FLOOR"),
			MaxQualityRetries:   v.GetInt("CHINA_WALL_MAX_QUALITY_RETRIES"),
			ClaimVerifyEnabled:  true, // B8 default on; overridden below if env set
		},
		Context: ContextConfig{
			MaxTokens:        v.GetInt("CONTEXT_MAX_TOKENS"),
			RecentMessages:   v.GetInt("CONTEXT_RECENT_MESSAGES"),
			SemanticTopK:     v.GetInt("CONTEXT_SEMANTIC_TOP_K"),
			CourseChunksTopK: v.GetInt("CONTEXT_COURSE_CHUNKS_TOP_K"),
		},
		Security: SecurityConfig{
			EncryptionKey: v.GetString("ENCRYPTION_KEY"),
		},
		OAuth: OAuthConfig{
			GitHubClientID:     v.GetString("GITHUB_CLIENT_ID"),
			GitHubClientSecret: v.GetString("GITHUB_CLIENT_SECRET"),
			GitLabClientID:     v.GetString("GITLAB_CLIENT_ID"),
			GitLabClientSecret: v.GetString("GITLAB_CLIENT_SECRET"),
			BaseURL:            v.GetString("BASE_URL"),
			FrontendURL:        v.GetString("FRONTEND_URL"),
		},
		// Default false: self-registration off unless explicitly enabled.
		// viper.GetBool returns false when the env var is absent, which is
		// exactly the desired default.
		Auth: AuthConfig{
			SelfRegistrationEnabled: v.GetBool("SELF_REGISTRATION_ENABLED"),
		},
		RateLimit: RateLimitConfig{
			PerIP:   v.GetInt("RATE_LIMIT_PER_IP"),
			PerUser: v.GetInt("RATE_LIMIT_PER_USER"),
		},
		Log: LogConfig{
			Level: v.GetString("LOG_LEVEL"),
		},
		Provenance: ProvenanceConfig{
			Enabled:    true, // C1 default on; overridden below if env set
			SigningKey: v.GetString("PROVENANCE_SIGNING_KEY"),
		},
		Versioning: VersioningConfig{
			Enabled:        true, // C2 default on; overridden below if env set
			DriftThreshold: v.GetFloat64("EXPERT_DRIFT_THRESHOLD"),
		},
		Tenant: TenantConfig{
			IsolationEnabled: true, // C4 default on; overridden below if env set
		},
	}

	// Parse CORS_ALLOWED_ORIGINS (comma-separated)
	if raw := v.GetString("CORS_ALLOWED_ORIGINS"); raw != "" {
		for _, o := range strings.Split(raw, ",") {
			if o = strings.TrimSpace(o); o != "" {
				cfg.CORSAllowedOrigins = append(cfg.CORSAllowedOrigins, o)
			}
		}
	}

	// B6: GetBool is false when the env is absent, so only honour an
	// explicit CHINA_WALL_QUALITY_JUDGE_ENABLED=false; absent → true.
	if v.IsSet("CHINA_WALL_QUALITY_JUDGE_ENABLED") {
		cfg.ChinaWall.QualityJudgeEnabled = v.GetBool("CHINA_WALL_QUALITY_JUDGE_ENABLED")
	}
	// MaxQualityRetries: absent → default 1 in applyDefaults; explicit 0
	// stays 0 (score-only, no regen). Tracked via a sentinel when unset.
	if !v.IsSet("CHINA_WALL_MAX_QUALITY_RETRIES") {
		cfg.ChinaWall.MaxQualityRetries = -1 // sentinel: applyDefaults fills 1
	}
	// B8: same IsSet pattern — absent → true; explicit false kills the gate.
	if v.IsSet("CHINA_WALL_CLAIM_VERIFY_ENABLED") {
		cfg.ChinaWall.ClaimVerifyEnabled = v.GetBool("CHINA_WALL_CLAIM_VERIFY_ENABLED")
	}
	// C1: same IsSet pattern — absent → true; explicit false disables.
	if v.IsSet("PROVENANCE_ENABLED") {
		cfg.Provenance.Enabled = v.GetBool("PROVENANCE_ENABLED")
	}
	// C2: same IsSet pattern — absent → true; explicit false disables.
	if v.IsSet("EXPERT_VERSIONING_ENABLED") {
		cfg.Versioning.Enabled = v.GetBool("EXPERT_VERSIONING_ENABLED")
	}
	// C4: same IsSet pattern — absent → true; explicit false is the kill switch.
	if v.IsSet("TENANT_ISOLATION_ENABLED") {
		cfg.Tenant.IsolationEnabled = v.GetBool("TENANT_ISOLATION_ENABLED")
	}

	// Validate required fields — fail fast
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Apply defaults for non-critical fields
	cfg.applyDefaults()

	return cfg, nil
}

// validate checks that all required configuration is present.
// Returns error with ALL missing fields — not just the first one.
func (c *Config) validate() error {
	var missing []string

	if c.Database.URL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.Redis.URL == "" {
		missing = append(missing, "REDIS_URL")
	}
	if c.JWT.Secret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if len(c.JWT.Secret) < 32 {
		missing = append(missing, "JWT_SECRET (minimum 32 characters)")
	}
	if c.LLM.OpenRouterAPIKey == "" {
		missing = append(missing, "OPENROUTER_API_KEY")
	}
	if c.Security.EncryptionKey == "" {
		missing = append(missing, "ENCRYPTION_KEY")
	}
	if len(c.Security.EncryptionKey) != 32 {
		missing = append(missing, "ENCRYPTION_KEY (must be exactly 32 bytes for AES-256)")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}

	return nil
}

// applyDefaults sets sensible defaults for optional fields.
func (c *Config) applyDefaults() {
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.Env == "" {
		c.Server.Env = "development"
	}
	if c.Database.MaxConns == 0 {
		c.Database.MaxConns = 20
	}
	if c.Database.MinConns == 0 {
		c.Database.MinConns = 5
	}
	if c.Database.MaxConnLifetime == 0 {
		c.Database.MaxConnLifetime = 60 * time.Minute
	}
	if c.JWT.AccessExpiryMinutes == 0 {
		// WHY 480 (8 hours) not 15 minutes:
		//   15-min tokens cause frequent logouts in dev:
		//   - Docker restart clears Redis → refresh fails
		//   - Page reload + cross-port cookie delay → bootstrap race
		//   - Concurrent requests on expiry → UX disruption every 15 min
		//   8 hours = one working session. Access token is in-memory
		//   (XSS safe) so longer expiry is acceptable for single-admin.
		//   Production: set JWT_ACCESS_EXPIRY_MINUTES=15 in .env explicitly.
		c.JWT.AccessExpiryMinutes = 480
	}
	if c.JWT.RefreshExpiryDays == 0 {
		c.JWT.RefreshExpiryDays = 30
	}
	if c.LLM.OpenRouterBaseURL == "" {
		c.LLM.OpenRouterBaseURL = "https://openrouter.ai/api/v1"
	}
	if c.LLM.CodeCraftAPIBaseURL == "" {
		// WHY this default: CodeCraftAPI's documented base URL.
		// Configurable via CODECRAFTAPI_BASE_URL env var if it ever changes.
		c.LLM.CodeCraftAPIBaseURL = "https://codecraftapi.com/v1"
	}
	if c.LLM.ModelCheap == "" {
		c.LLM.ModelCheap = "deepseek/deepseek-chat"
	}
	if c.LLM.ModelStrong == "" {
		c.LLM.ModelStrong = "anthropic/claude-3-5-sonnet"
	}
	if c.LLM.ModelFast == "" {
		c.LLM.ModelFast = "google/gemini-flash-1.5"
	}
	if c.ML.SidecarURL == "" {
		c.ML.SidecarURL = "http://localhost:8001"
	}
	if c.ML.TimeoutSeconds == 0 {
		// WHY 300 not 30:
		//   bge-base-en-v1.5 on CPU for 25 chunks ≈ 8-12s.
		//   Old default of 30s caused timeouts when batch size was 100
		//   (100 chunks × ~0.4s/chunk = 40s > 30s timeout).
		//   300s = 5 minutes. Even under heavy load, 25-chunk batches
		//   will never approach this limit.
		c.ML.TimeoutSeconds = 300
	}
	if c.ChinaWall.RerankerThreshold == 0 {
		c.ChinaWall.RerankerThreshold = 0.35
	}
	if c.ChinaWall.RelaxedThreshold == 0 {
		c.ChinaWall.RelaxedThreshold = 0.30
	}
	if c.ChinaWall.MaxRetries == 0 {
		c.ChinaWall.MaxRetries = 5
	}
	// B6 quality-judge numeric defaults. Bool default (true) is set in Load
	// so an absent env does not flip the feature off. MaxQualityRetries
	// sentinel -1 (absent) → 1; explicit 0 stays 0 (score-only).
	if c.ChinaWall.QualityFloor <= 0 || c.ChinaWall.QualityFloor > 1 {
		c.ChinaWall.QualityFloor = 0.7
	}
	if c.ChinaWall.MaxQualityRetries < 0 {
		c.ChinaWall.MaxQualityRetries = 1
	}
	// C1: signing key falls back to the AES key so provenance always has a
	// server secret without a new required env var. Dedicated override wins.
	if c.Provenance.SigningKey == "" {
		c.Provenance.SigningKey = c.Security.EncryptionKey
	}
	// C2: drift threshold default.
	if c.Versioning.DriftThreshold <= 0 || c.Versioning.DriftThreshold > 1 {
		c.Versioning.DriftThreshold = 0.30
	}
	if c.Context.MaxTokens == 0 {
		c.Context.MaxTokens = 10000
	}
	if c.Context.RecentMessages == 0 {
		c.Context.RecentMessages = 3
	}
	if c.Context.SemanticTopK == 0 {
		c.Context.SemanticTopK = 5
	}
	if c.Context.CourseChunksTopK == 0 {
		c.Context.CourseChunksTopK = 5
	}
	if c.OAuth.BaseURL == "" {
		c.OAuth.BaseURL = "http://localhost:8080"
	}
	if c.OAuth.FrontendURL == "" {
		c.OAuth.FrontendURL = "http://localhost:3000"
	}
	if c.RateLimit.PerIP == 0 {
		c.RateLimit.PerIP = 100
	}
	if c.RateLimit.PerUser == 0 {
		c.RateLimit.PerUser = 20
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}
