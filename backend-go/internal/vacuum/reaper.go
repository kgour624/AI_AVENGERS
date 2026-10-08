package vacuum

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// OrphanGrace — how long a job may sit in a busy status (picking/cleaning/
// verifying) with no DB write at all before it is treated as orphaned: 45 minutes.
//
// WHY 45 minutes: the HTTP run path bounds itself with this same window (see
// ExecuteJob), and the picker path only touches file_jobs at stage boundaries,
// so phase 2 — the slow neck of the pipeline, two model calls per chunk — can
// legitimately leave a row silent for a long time. Anything shorter risks
// declaring a live run dead; anything longer keeps a dead row on the admin
// screen, where it renders as a spinner + Cancel and never as a Retry.
const OrphanGrace = 45 * time.Minute

// orphanSweepInterval — how often the sweep runs. One minute is far below the
// grace window, so a healed row shows up on the admin screen within a minute of
// becoming eligible.
const orphanSweepInterval = time.Minute

// orphanCandidatesSQL selects the ids of busy jobs whose last write is older
// than the grace cutoff ($1). The sweep then skips the ones this process still
// owns (see isRunning).
//
// WHY verifying is in the list: phase 5 (hash verify + LLM classification) writes
// file_jobs exactly twice — once on entry (phase=5) and once when the job closes
// as done — so a run that dies between those two writes leaves the row sitting in
// verifying, which the admin UI renders as a live spinner with Cancel and no
// Retry. That is precisely the stuck row this reaper exists to close.
const orphanCandidatesSQL = `SELECT id::text FROM file_jobs
 WHERE status IN ('picking','cleaning','verifying') AND updated_at < $1`

// orphanFailSQL marks one orphaned job failed.
//
// WHY status='failed' and not 'canceled' or 'scheduled': failed is the only
// non-destructive state the admin UI already knows how to act on — it renders
// Retry and Retry-last-stage for it, and RetryJob accepts it (attempts < 3).
// So the healing needs no frontend change and throws nothing away, unlike
// Cancel, which deletes the row, the chunks and the stored files.
const orphanFailSQL = `UPDATE file_jobs SET
     status='failed',
     phase=6,
     error_code=$2,
     last_error=$3,
     retry_after=NULL,
     finished_at=NOW(),
     duration_ms=COALESCE(duration_ms, (EXTRACT(EPOCH FROM (NOW()-COALESCE(picked_at, created_at)))*1000)::int),
     updated_at=NOW()
 WHERE id=$1`

// HealOrphansOnBoot fails every job that is still marked busy, because a process
// that has just started provably owns no live run: whatever left those rows in
// picking/cleaning died with the previous process (restart, crash, redeploy).
//
// WHY no grace window here: waiting 45 minutes after a restart would leave the
// admin staring at a spinner that can never move again. The restart itself is
// the liveness proof, so this pass uses a zero window.
func HealOrphansOnBoot(ctx context.Context, db *pgxpool.Pool, log *zap.Logger) int {
	n := sweep(ctx, db, 0, "ORPHANED_RESTART", "run lost: backend restarted while this job was in flight", log)
	if n > 0 && log != nil {
		log.Warn("vacuum orphan reaper: healed jobs left busy by a previous process", zap.Int("healed", n))
	}
	return n
}

// StartOrphanReaper heals orphans once on boot and then sweeps every minute
// until ctx is cancelled. Run it in a goroutine from process startup.
func StartOrphanReaper(ctx context.Context, db *pgxpool.Pool, log *zap.Logger) {
	if db == nil {
		return
	}
	HealOrphansOnBoot(ctx, db, log)
	ticker := time.NewTicker(orphanSweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if n := sweep(ctx, db, OrphanGrace, "ORPHANED", "no progress for 45m: run died without reporting", log); n > 0 && log != nil {
				log.Warn("vacuum orphan reaper: healed silent jobs", zap.Int("healed", n))
			}
		}
	}
}

// sweep is the shared body of both passes: select silent busy jobs, skip the
// ones with a live run registered in this process, and fail the rest. It
// returns how many rows were healed.
func sweep(ctx context.Context, db *pgxpool.Pool, grace time.Duration, code, message string, log *zap.Logger) int {
	rows, err := db.Query(ctx, orphanCandidatesSQL, time.Now().Add(-grace))
	if err != nil {
		if log != nil {
			log.Warn("vacuum orphan reaper: candidate query failed", zap.Error(err))
		}
		return 0
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		if log != nil {
			log.Warn("vacuum orphan reaper: candidate scan failed", zap.Error(err))
		}
		return 0
	}
	healed := 0
	for _, id := range ids {
		// A registered run is proof of life: a phase-2 scan can be silent for
		// longer than the grace window, and failing a live row would offer the
		// admin a Retry that starts a second run on the same job.
		if isRunning(id) {
			continue
		}
		tag, err := db.Exec(ctx, orphanFailSQL, id, code, message)
		if err != nil {
			if log != nil {
				log.Warn("vacuum orphan reaper: update failed", zap.String("job_id", id), zap.Error(err))
			}
			continue
		}
		if tag.RowsAffected() == 0 {
			continue
		}
		healed++
		// Tell anyone still watching that the run is over. Without this an open
		// Live pipeline panel keeps re-showing the ghost of the dead run (the
		// hub replays its retained snapshot on every reattach), which is exactly
		// what made an orphaned job look alive on screen.
		progressBus.Forget(id)
		publishProgress(id, "failed", "failed", message, 6, 100, 0, 0)
		if log != nil {
			log.Info("vacuum orphan reaper: job reaped", zap.String("job_id", id), zap.String("error_code", code))
		}
	}
	return healed
}
