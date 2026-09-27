-- Public design sharing.
--
-- WHY token-keyed and not admin-keyed: the product has several admin accounts,
-- so a per-admin endpoint would force the landing page to integrate once per
-- admin. The token is the lookup key, which keeps exactly one public endpoint
-- for the whole product, serving every admin's publications.
CREATE TABLE IF NOT EXISTS public_designs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token       TEXT NOT NULL UNIQUE,
    title       TEXT NOT NULL,
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    created_by  UUID,
    created_at  TEXT NOT NULL DEFAULT to_char(NOW(), 'YYYY-MM-DD"T"HH24:MI:SSOF'),
    updated_at  TEXT NOT NULL DEFAULT to_char(NOW(), 'YYYY-MM-DD"T"HH24:MI:SSOF')
);

CREATE INDEX IF NOT EXISTS idx_public_designs_workflow ON public_designs(workflow_id);

-- One row per shared file. (design_id, path) is the unique key so re-pushing a
-- path overwrites it instead of duplicating it in the published array.
CREATE TABLE IF NOT EXISTS public_design_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    design_id  UUID NOT NULL REFERENCES public_designs(id) ON DELETE CASCADE,
    path       TEXT NOT NULL,
    name       TEXT NOT NULL,
    operation  TEXT,
    content    TEXT NOT NULL,
    added_at   TEXT NOT NULL DEFAULT to_char(NOW(), 'YYYY-MM-DD"T"HH24:MI:SSOF'),
    UNIQUE (design_id, path)
);
