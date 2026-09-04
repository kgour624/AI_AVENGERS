-- ============================================================
-- Migration 001: Initial Schema
-- AI Avengers — All tables
-- Run: go run cmd/migrate/main.go up
-- ============================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector;

-- ============================================================
-- USERS
-- ============================================================
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) UNIQUE NOT NULL,
    hashed_password VARCHAR(255) NOT NULL,
    full_name       VARCHAR(255) NOT NULL,
    role            VARCHAR(20) NOT NULL DEFAULT 'client',
    is_active       BOOLEAN DEFAULT TRUE,
    totp_secret     VARCHAR(255),
    totp_enabled    BOOLEAN DEFAULT FALSE,
    preferences     JSONB DEFAULT '{}',
    last_login      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    CONSTRAINT users_role_check CHECK (role IN ('admin', 'client'))
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users(role) WHERE deleted_at IS NULL;

-- ============================================================
-- EXPERTS
-- ============================================================
CREATE TABLE experts (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                  VARCHAR(255) NOT NULL,
    slug                  VARCHAR(255) UNIQUE NOT NULL,
    domain                VARCHAR(100) NOT NULL,
    description           TEXT,
    avatar_url            VARCHAR(500),
    reasoning_charter     TEXT NOT NULL DEFAULT '',
    clarification_charter JSONB NOT NULL DEFAULT '{}',
    capability_summary    JSONB DEFAULT '{}',
    total_chunks          INTEGER DEFAULT 0,
    total_topics          INTEGER DEFAULT 0,
    avg_depth_level       DECIMAL(3,2),
    avg_rating            DECIMAL(3,2) DEFAULT 0,
    total_ratings         INTEGER DEFAULT 0,
    is_active             BOOLEAN DEFAULT TRUE,
    is_training           BOOLEAN DEFAULT FALSE,
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    updated_at            TIMESTAMPTZ DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

CREATE INDEX idx_experts_slug ON experts(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_experts_domain ON experts(domain) WHERE is_active = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_experts_active ON experts(is_active) WHERE deleted_at IS NULL;

-- ============================================================
-- COURSE CHUNKS
-- ============================================================
CREATE TABLE course_chunks (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id        UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    chunk_text       TEXT NOT NULL,
    chunk_index      INTEGER NOT NULL,
    topic            VARCHAR(255),
    subtopic         VARCHAR(255),
    source_file      VARCHAR(500),
    embedding        vector(768) NOT NULL,
    prev_chunk_id    UUID REFERENCES course_chunks(id),
    next_chunk_id    UUID REFERENCES course_chunks(id),
    times_retrieved  INTEGER DEFAULT 0,
    times_cited      INTEGER DEFAULT 0,
    boost_factor     DECIMAL(5,4) DEFAULT 1.0,
    created_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_chunks_expert ON course_chunks(expert_id);
CREATE INDEX idx_chunks_topic ON course_chunks(expert_id, topic);
CREATE INDEX idx_chunks_embedding ON course_chunks
    USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- ============================================================
-- EXPERT CAPABILITIES
-- ============================================================
CREATE TABLE expert_capabilities (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id         UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    topic             VARCHAR(255) NOT NULL,
    depth_level       INTEGER,
    chunk_count       INTEGER NOT NULL DEFAULT 0,
    complexity_ceiling VARCHAR(50),
    can_handle        TEXT[],
    cannot_handle     TEXT[],
    example_questions TEXT[],
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(expert_id, topic),
    CONSTRAINT depth_level_check CHECK (depth_level BETWEEN 1 AND 5)
);

CREATE INDEX idx_capabilities_expert ON expert_capabilities(expert_id);

-- ============================================================
-- PROJECTS
-- ============================================================
CREATE TABLE projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    status          VARCHAR(50) DEFAULT 'active',
    repo_url        VARCHAR(500),
    repo_provider   VARCHAR(20),
    repo_branch     VARCHAR(255) DEFAULT 'main',
    repo_connected  BOOLEAN DEFAULT FALSE,
    repo_last_sync  TIMESTAMPTZ,
    tech_stack      JSONB DEFAULT '{}',
    architecture_type VARCHAR(100),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    CONSTRAINT projects_status_check CHECK (status IN ('active', 'completed', 'archived'))
);

CREATE INDEX idx_projects_client ON projects(client_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_status ON projects(status) WHERE deleted_at IS NULL;

-- ============================================================
-- PROJECT EXPERTS
-- ============================================================
CREATE TABLE project_experts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    expert_id   UUID NOT NULL REFERENCES experts(id),
    added_at    TIMESTAMPTZ DEFAULT NOW(),
    is_active   BOOLEAN DEFAULT TRUE,
    UNIQUE(project_id, expert_id)
);

CREATE INDEX idx_project_experts_project ON project_experts(project_id);
CREATE INDEX idx_project_experts_expert ON project_experts(expert_id);

-- ============================================================
-- CHATS
-- ============================================================
CREATE TABLE chats (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    client_id       UUID NOT NULL REFERENCES users(id),
    title           VARCHAR(500) NOT NULL,
    message_count   INTEGER DEFAULT 0,
    is_archived     BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_chats_project ON chats(project_id);
CREATE INDEX idx_chats_client ON chats(client_id);
CREATE INDEX idx_chats_updated ON chats(updated_at DESC);

-- ============================================================
-- MESSAGES
-- ============================================================
CREATE TABLE messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id         UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL,
    content         TEXT NOT NULL,
    turn_number     INTEGER NOT NULL,
    expert_id       UUID REFERENCES experts(id),
    decision_mode   VARCHAR(20),
    citations       JSONB,
    gate_stopped    INTEGER,
    confidence      DECIMAL(5,4),
    tokens_used     INTEGER,
    cost_usd        DECIMAL(10,6),
    model_used      VARCHAR(100),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT messages_role_check CHECK (role IN ('user', 'assistant')),
    CONSTRAINT messages_mode_check CHECK (
        decision_mode IS NULL OR
        decision_mode IN ('ASK', 'WARN', 'PUSH_BACK', 'REFUSE', 'ADVISE')
    )
);

CREATE INDEX idx_messages_chat ON messages(chat_id, turn_number);
CREATE INDEX idx_messages_created ON messages(created_at DESC);

-- ============================================================
-- CHAT INDEX
-- ============================================================
CREATE TABLE chat_index (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id         UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    message_id      UUID REFERENCES messages(id),
    turn_number     INTEGER NOT NULL,
    tags            TEXT[],
    topic           VARCHAR(255),
    one_line_summary TEXT NOT NULL,
    importance      INTEGER,
    key_decisions   JSONB DEFAULT '[]',
    embedding       vector(768) NOT NULL,
    superseded_by   UUID REFERENCES chat_index(id),
    is_superseded   BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT importance_check CHECK (importance BETWEEN 1 AND 5)
);

CREATE INDEX idx_chat_index_chat ON chat_index(chat_id);
CREATE INDEX idx_chat_index_importance ON chat_index(chat_id, importance DESC);
CREATE INDEX idx_chat_index_superseded ON chat_index(chat_id, is_superseded);
CREATE INDEX idx_chat_index_embedding ON chat_index
    USING ivfflat (embedding vector_cosine_ops) WITH (lists = 50);

-- ============================================================
-- CHAT SUMMARIES
-- ============================================================
CREATE TABLE chat_summaries (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id          UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    summary_text     TEXT NOT NULL,
    turn_range_start INTEGER NOT NULL,
    turn_range_end   INTEGER NOT NULL,
    key_decisions    JSONB DEFAULT '[]',
    open_questions   TEXT[],
    created_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_summaries_chat ON chat_summaries(chat_id, turn_range_end DESC);

-- ============================================================
-- L2 GROUP MEMORY
-- ============================================================
CREATE TABLE project_memory_l2 (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    expert_id       UUID NOT NULL REFERENCES experts(id),
    memory_type     VARCHAR(50) NOT NULL,
    content         TEXT NOT NULL,
    context         TEXT,
    turn_reference  INTEGER,
    embedding       vector(768),
    importance      INTEGER,
    is_superseded   BOOLEAN DEFAULT FALSE,
    superseded_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT l2_importance_check CHECK (importance BETWEEN 1 AND 5),
    CONSTRAINT l2_memory_type_check CHECK (
        memory_type IN ('decision', 'code', 'error', 'fix', 'recommendation', 'requirement')
    )
);

CREATE INDEX idx_l2_project ON project_memory_l2(project_id);
CREATE INDEX idx_l2_expert ON project_memory_l2(project_id, expert_id);
CREATE INDEX idx_l2_active ON project_memory_l2(project_id, is_superseded);
CREATE INDEX idx_l2_embedding ON project_memory_l2
    USING ivfflat (embedding vector_cosine_ops) WITH (lists = 30);

-- ============================================================
-- L3 MASTER EVENT LOG (append-only, immutable)
-- ============================================================
CREATE TABLE master_event_log (
    id              BIGSERIAL PRIMARY KEY,
    project_id      UUID NOT NULL REFERENCES projects(id),
    expert_id       UUID REFERENCES experts(id),
    client_id       UUID NOT NULL REFERENCES users(id),
    chat_id         UUID REFERENCES chats(id),
    message_id      UUID REFERENCES messages(id),
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL DEFAULT '{}',
    reasoning       TEXT,
    decision_made   TEXT,
    alternatives_considered JSONB DEFAULT '[]',
    created_at      TIMESTAMPTZ DEFAULT NOW()
    -- NO updated_at, NO deleted_at — this table is IMMUTABLE
);

CREATE INDEX idx_l3_project ON master_event_log(project_id, created_at DESC);
CREATE INDEX idx_l3_expert ON master_event_log(expert_id, created_at DESC);
CREATE INDEX idx_l3_event_type ON master_event_log(event_type);
CREATE INDEX idx_l3_client ON master_event_log(client_id, created_at DESC);

-- ============================================================
-- RATINGS
-- ============================================================
CREATE TABLE ratings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id      UUID NOT NULL REFERENCES messages(id),
    client_id       UUID NOT NULL REFERENCES users(id),
    expert_id       UUID NOT NULL REFERENCES experts(id),
    project_id      UUID NOT NULL REFERENCES projects(id),
    score           INTEGER NOT NULL,
    feedback        TEXT,
    feedback_type   VARCHAR(50),
    code_executed   BOOLEAN DEFAULT FALSE,
    execution_success BOOLEAN,
    error_message   TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT ratings_score_check CHECK (score BETWEEN 1 AND 5),
    CONSTRAINT ratings_feedback_type_check CHECK (
        feedback_type IS NULL OR
        feedback_type IN ('accepted', 'rejected', 'modified', 'ignored')
    )
);

CREATE INDEX idx_ratings_expert ON ratings(expert_id);
CREATE INDEX idx_ratings_message ON ratings(message_id);
CREATE INDEX idx_ratings_project ON ratings(project_id);

-- ============================================================
-- REPO CONNECTIONS
-- ============================================================
CREATE TABLE repo_connections (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id       UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    client_id        UUID NOT NULL REFERENCES users(id),
    provider         VARCHAR(20) NOT NULL,
    repo_url         VARCHAR(500) NOT NULL,
    repo_name        VARCHAR(255),
    default_branch   VARCHAR(255) DEFAULT 'main',
    access_token     TEXT,
    refresh_token    TEXT,
    token_expires_at TIMESTAMPTZ,
    last_sync_at     TIMESTAMPTZ,
    sync_status      VARCHAR(50) DEFAULT 'pending',
    total_chunks     INTEGER DEFAULT 0,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT provider_check CHECK (provider IN ('github', 'gitlab')),
    CONSTRAINT sync_status_check CHECK (
        sync_status IN ('pending', 'syncing', 'complete', 'failed')
    )
);

CREATE INDEX idx_repo_connections_project ON repo_connections(project_id);

-- ============================================================
-- INGESTION JOBS
-- ============================================================
CREATE TABLE ingestion_jobs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id        UUID REFERENCES experts(id),
    project_id       UUID REFERENCES projects(id),
    job_type         VARCHAR(50) NOT NULL,
    status           VARCHAR(50) DEFAULT 'pending',
    source_path      VARCHAR(500),
    total_chunks     INTEGER DEFAULT 0,
    processed_chunks INTEGER DEFAULT 0,
    error_message    TEXT,
    started_at       TIMESTAMPTZ,
    completed_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT job_type_check CHECK (job_type IN ('transcript', 'repo')),
    CONSTRAINT job_status_check CHECK (
        status IN ('pending', 'running', 'complete', 'failed')
    )
);

CREATE INDEX idx_ingestion_jobs_expert ON ingestion_jobs(expert_id);
CREATE INDEX idx_ingestion_jobs_status ON ingestion_jobs(status);

-- ============================================================
-- SYSTEM SETTINGS
-- ============================================================
CREATE TABLE system_settings (
    key         VARCHAR(255) PRIMARY KEY,
    value       JSONB NOT NULL,
    description TEXT,
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_by  UUID REFERENCES users(id)
);

-- Default settings
INSERT INTO system_settings (key, value, description) VALUES
(
    'china_wall',
    '{"reranker_threshold": 0.35, "relaxed_threshold": 0.30, "max_retries": 5, "strip_uncited": true}',
    'China Wall enforcement configuration'
),
(
    'context',
    '{"max_tokens": 10000, "recent_messages": 3, "semantic_top_k": 5, "course_chunks_top_k": 5}',
    'Context assembly configuration'
),
(
    'models',
    '{"cheap": "deepseek/deepseek-chat", "strong": "anthropic/claude-3-5-sonnet", "fast": "google/gemini-flash-1.5"}',
    'LLM model routing'
),
(
    'cost_budget',
    '{"monthly_limit_usd": 1000, "alert_threshold": 0.8, "current_month_spend": 0}',
    'Cost monitoring and budget alerts'
),
(
    'rate_limits',
    '{"per_ip": 100, "per_user": 20, "window_seconds": 60}',
    'Rate limiting configuration'
);
