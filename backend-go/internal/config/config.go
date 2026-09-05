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
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	LLM       LLMConfig
	ML        MLConfig
	ChinaWall ChinaWallConfig
	Context   ContextConfig
	Security  SecurityConfig
	OAuth     OAuthConfig
	RateLimit RateLimitConfig
	Log       LogConfig
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

type LLMConfig struct {
	OpenRouterAPIKey string
	OpenRouterBaseURL string
	ModelCheap       string
	ModelStrong      string
	ModelFast        string
}

type MLConfig struct {
	SidecarURL     string
	TimeoutSeconds int
}

type ChinaWallConfig struct {
	RerankerThreshold float64
	RelaxedThreshold  float64
	MaxRetries        int
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
	BaseURL            string // e.g. https://api.yourdomain.com
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
		LLM: LLMConfig{
			OpenRouterAPIKey:  v.GetString("OPENROUTER_API_KEY"),
			OpenRouterBaseURL: v.GetString("OPENROUTER_BASE_URL"),
			ModelCheap:        v.GetString("LLM_MODEL_CHEAP"),
			ModelStrong:       v.GetString("LLM_MODEL_STRONG"),
			ModelFast:         v.GetString("LLM_MODEL_FAST"),
		},
		ML: MLConfig{
			SidecarURL:     v.GetString("ML_SIDECAR_URL"),
			TimeoutSeconds: v.GetInt("ML_SIDECAR_TIMEOUT_SECONDS"),
		},
		ChinaWall: ChinaWallConfig{
			RerankerThreshold: v.GetFloat64("CHINA_WALL_RERANKER_THRESHOLD"),
			RelaxedThreshold:  v.GetFloat64("CHINA_WALL_RELAXED_THRESHOLD"),
			MaxRetries:        v.GetInt("CHINA_WALL_MAX_RETRIES"),
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
		RateLimit: RateLimitConfig{
			PerIP:   v.GetInt("RATE_LIMIT_PER_IP"),
			PerUser: v.GetInt("RATE_LIMIT_PER_USER"),
		},
		Log: LogConfig{
			Level: v.GetString("LOG_LEVEL"),
		},
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
		c.JWT.AccessExpiryMinutes = 15
	}
	if c.JWT.RefreshExpiryDays == 0 {
		c.JWT.RefreshExpiryDays = 7
	}
	if c.LLM.OpenRouterBaseURL == "" {
		c.LLM.OpenRouterBaseURL = "https://openrouter.ai/api/v1"
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
		c.ML.TimeoutSeconds = 30
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
