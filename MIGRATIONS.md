# MIGRATIONS.md — AI Avengers Database Migration Registry

> **Single Source of Truth** for all database migrations.
> **MANDATORY:** Har naya migration add karne se pehle yeh file padho.
> **MANDATORY:** Har naya migration add karne ke baad yeh file update karo.

---

## ⚠️ Critical Rules (Pehle Padho)

1. **Next migration number = last number + 1.** Kabhi bhi existing number reuse mat karo.
2. **Har migration ke 2 files hone chahiye:** `NNN_name.up.sql` + `NNN_name.down.sql`
3. **Column add karne se pehle check karo** ki woh column already kisi migration mein exist toh nahi karta.
4. **`golang-migrate` same version number pe 2 files hone par crash karta hai** — duplicate number = production outage.
5. **`IF NOT EXISTS` / `ON CONFLICT DO NOTHING` use karo** — migrations idempotent honi chahiye.
6. **Existing columns rename/drop mat karo** — additive only.

---

## 📊 Current Migration Registry

| # | File Name | Tables/Columns Changed | Kya Karta Hai |
|---|-----------|------------------------|---------------|
| 001 | `001_initial_schema` | `users`, `projects`, `experts`, `chats`, `messages`, `course_chunks`, `chat_index`, `chat_summaries`, `master_event_log`, `l1_memory`, `l2_memory`, `system_settings`, `ingestion_jobs` | Base schema. pgvector extension. Sab core tables. |
| 002 | `002_hybrid_search` | `course_chunks` (tsvector, GIN index) | Hybrid search: vector + full-text search ke liye indexes. |
| 003 | `003_repo_chunks` | `repo_chunks` (new table), `experts` (git columns) | Code repository chunking. Git commit reference columns. |
| 004 | `004_messages_warning_fields` | `messages` (confidence, decision_mode, warning_text, clarifying_questions, citations, gate_stopped) | China Wall metadata fields on messages. |
| 005 | `005_seed_admin` | `users`, `system_settings` (INSERT) | Default admin user + system initialization data. |
| 006 | `006_collaboration_layer` | `experts` (model_tier, temperature, top_p, loop_pattern, max_loop_iterations, allowed_tools, training_status), `course_chunks` (chunk_hash), `workflows`, `blackboard_events`, `workflow_tasks`, `workflow_checkpoints`, `approval_requests`, `messages` (workflow_id), `system_settings` | Multi-agent workflow engine. Blackboard event log. Kanban projection. Client approval gates. |
| 007 | `007_ingestion_resume` | `ingestion_jobs` (resume fields, checkpoint columns) | Ingestion pipeline resume/checkpoint support. |
| 008 | `008_chunk_hash_unique` | `course_chunks` (UNIQUE constraint on expert_id + chunk_hash) | Duplicate chunk prevention. |
| 009 | `009_job_transcript_content` | `ingestion_jobs` (transcript_content) | Raw input text storage in ingestion jobs. |
| 010 | `010_expert_categories` | `expert_categories` (new table), `experts` (category_id), `messages` (reply_to_message_id), `system_settings` | Expert category layer. Structured answer templates. Reply threading. |
| 011 | `011_ingestion_pause` | `ingestion_jobs` (status CHECK updated to include 'paused', paused_at) | Ingestion pause/resume. Charter LLM failure handling. |
| 012 | `012_ingestion_jobs_updated_at` | `ingestion_jobs` (updated_at column + trigger) | Auto-updating updated_at. Required by admin resume/retry queries. |
| 013 | `013_message_template_sections` | `messages` (template_sections JSONB) | Structured section-wise answers persist karne ke liye. Page reload pe blank answer bug fix. |
| 014 | `014_workflow_recovery` | `workflows` (runner_state JSONB, current_task_cursor BIGINT), `workflow_tasks` (UNIQUE constraint on workflow_id + assigned_expert_id) | WorkflowRunner pod-restart recovery. Projector ON CONFLICT support. |

**Next migration number: `015`**

---

## 📍 Column Quick-Reference (Kahan Kya Hai)

Yeh section isliye hai taaki duplicate column add karne ki galti na ho.

### `messages` table
| Column | Migration | Type |
|--------|-----------|------|
| id, chat_id, role, content, turn_number, expert_id, created_at | 001 | base |
| confidence, decision_mode, warning_text, clarifying_questions, citations, gate_stopped | 004 | JSONB/text |
| workflow_id | 006 | UUID FK |
| reply_to_message_id | **010** | UUID FK (self-ref) |
| template_sections | **013** | JSONB |

### `workflows` table
| Column | Migration | Type |
|--------|-----------|------|
| id, client_id, project_id, title, status, current_phase, phase_started_at, phase_completed_at, selected_expert_ids, cost_* | 006 | base |
| runner_state | **014** | JSONB |
| current_task_cursor | **014** | BIGINT |

### `workflow_tasks` table
| Column | Migration | Type |
|--------|-----------|------|
| id, workflow_id, assigned_expert_id, title, description, status, produced_artifact_event_id, cost_usd, started_at, completed_at | 006 | base |
| UNIQUE(workflow_id, assigned_expert_id) | **014** | constraint |

### `experts` table
| Column | Migration | Type |
|--------|-----------|------|
| id, name, domain, reasoning_charter, clarification_charter, is_active, is_training, deleted_at | 001 | base |
| (git columns) | 003 | text |
| model_tier, temperature, top_p, loop_pattern, max_loop_iterations, allowed_tools, training_status | 006 | various |
| category_id | **010** | UUID FK |

### `ingestion_jobs` table
| Column | Migration | Type |
|--------|-----------|------|
| id, expert_id, status, source_file, error_message, created_at | 001 | base |
| (resume fields) | 007 | various |
| transcript_content | 009 | TEXT |
| paused_at | **011** | TIMESTAMPTZ |
| updated_at | **012** | TIMESTAMPTZ |

---

## 🛠️ Naya Migration Add Karne Ka Process

```
Step 1: Yeh file padho — next number confirm karo (currently: 015)
Step 2: Column Quick-Reference check karo — column already exist toh nahi?
Step 3: Files banao:
          backend-go/migrations/015_your_name.up.sql
          backend-go/migrations/015_your_name.down.sql
Step 4: up.sql mein IF NOT EXISTS use karo (idempotent)
Step 5: down.sql mein exact reverse likho
Step 6: MIGRATIONS.md update karo:
          - Registry table mein row add karo
          - Column Quick-Reference update karo
          - "Next migration number" update karo (015 -> 016)
```

---

## 🚫 Galtiyan Jo Ho Chuki Hain (Seekho Inse)

### Galti 1: Duplicate Migration Number (2026-09-16)
**Kya hua:** `011_workflow_recovery` create kiya jabki `011_ingestion_pause` already exist karta tha.
**Impact:** `golang-migrate` same version number pe 2 files hone par crash karta hai.
**Fix:** `011_workflow_recovery` delete karke `014_workflow_recovery` banaya.
**Lesson:** Hamesha `ls migrations/` ya yeh file dekho pehle.

### Galti 2: Duplicate Columns (2026-09-16)
**Kya hua:** `011_workflow_recovery` mein `template_sections` aur `reply_to_message_id` add kiye — dono already 013 aur 010 mein the.
**Impact:** Migration fail hoti ya silently no-op hoti (IF NOT EXISTS ki wajah se).
**Fix:** Sirf actual new columns rakhe — `runner_state` aur `current_task_cursor`.
**Lesson:** Column Quick-Reference section hamesha check karo.

---

*Last updated: 2026-09-16 | Maintainer: System Architect*
