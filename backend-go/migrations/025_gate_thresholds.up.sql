-- Migration 025: per-domain Gate 1 threshold config (B4)
--
-- WHY: Gate 1 usable/strong cutoffs lived as package constants in
-- gate_system.go (0.40 / 0.70). Ratings feedback was recorded but never
-- fed back into those cutoffs, so over-refusing and over-answering domains
-- shared the same bar forever. P8 of the reliability roadmap: thresholds
-- live in config, not code. Defaults stay equal to the historical
-- constants so behaviour is unchanged until a row is written.
--
-- source column:
--   'default'     — never written (code-side fallback when no row)
--   'calibrated'  — produced by ProposeGateThresholds from ratings
--   'applied'     — admin explicitly accepted a calibrated proposal
--   'manual'      — admin hand-set via PATCH
--
-- GateSystem only consumes source IN ('applied','manual'). Calibrated
-- proposals sit until an admin applies them (P7: no blind auto-tune;
-- eval harness is not yet in the repo).

CREATE TABLE IF NOT EXISTS gate_thresholds (
    domain       VARCHAR(100) PRIMARY KEY,
    usable       DOUBLE PRECISION NOT NULL,
    strong       DOUBLE PRECISION NOT NULL,
    source       VARCHAR(30)  NOT NULL DEFAULT 'manual',
    sample_size  INTEGER      NOT NULL DEFAULT 0,
    notes        TEXT         NOT NULL DEFAULT '',
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by   UUID         REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT gate_thresholds_usable_range  CHECK (usable  >= 0.20 AND usable  <= 0.70),
    CONSTRAINT gate_thresholds_strong_range  CHECK (strong  >= 0.40 AND strong  <= 0.95),
    CONSTRAINT gate_thresholds_order         CHECK (strong  >  usable),
    CONSTRAINT gate_thresholds_source_check  CHECK (
        source IN ('calibrated', 'applied', 'manual')
    )
);

CREATE INDEX IF NOT EXISTS idx_gate_thresholds_source
    ON gate_thresholds (source);

COMMENT ON TABLE gate_thresholds IS
    'Per-domain Gate 1 usable/strong cutoffs. Applied/manual rows drive GateSystem; calibrated rows are proposals awaiting admin apply (B4/P7/P8).';
