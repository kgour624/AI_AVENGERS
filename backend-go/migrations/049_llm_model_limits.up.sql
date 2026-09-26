-- Per-model token limits, configurable from Admin → LLM Settings.
--
-- WHY a table and not more env vars: the limit belongs to the MODEL, not to the
-- deployment. A reasoning model (Claude thinking, DeepSeek R1, o-series) can
-- spend its whole completion budget on internal reasoning and return no visible
-- text at all — CodeCraftAPI's own docs say reasoning tokens count toward
-- completion_tokens and hard problems need 16000+. A model without reasoning
-- needs nothing like that. One deployment serves several providers and tiers
-- through different API keys, so the only correct place for the number is next
-- to the provider+tier it applies to, editable without a restart.
--
-- 0 on any column means "not configured": the provider's own maximum applies and
-- behaviour is byte-identical to before this table existed (the same 0-means-
-- default sentinel the domain profiles use for max_tokens_flat/structured).
--
-- tier is one of 'strong', 'fast', 'cheap' (the tiers the gateway selects
-- models by) or '*' for a provider-wide row that applies to any tier without its
-- own row.
CREATE TABLE IF NOT EXISTS llm_model_limits (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider                TEXT NOT NULL,
    tier                    TEXT NOT NULL CHECK (tier IN ('strong', 'fast', 'cheap', '*')),
    max_input_tokens        INTEGER NOT NULL DEFAULT 0 CHECK (max_input_tokens >= 0),
    max_output_tokens       INTEGER NOT NULL DEFAULT 0 CHECK (max_output_tokens >= 0),
    updated_by              UUID,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT llm_model_limits_provider_tier_key UNIQUE (provider, tier)
);
