package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ai_avengers/backend/internal/training"
)

func main() {
	var expertIDStr string
	var sourceFile string
	var dryRun bool
	flag.StringVar(&expertIDStr, "expert-id", "", "expert UUID to backfill (required)")
	flag.StringVar(&sourceFile, "source-file", "", "filter by source_file (optional, empty = all files for expert)")
	flag.BoolVar(&dryRun, "dry-run", false, "preview without writing")
	flag.Parse()

	if expertIDStr == "" {
		fmt.Fprintln(os.Stderr, "usage: rechunk --expert-id <uuid> [--source-file name] [--dry-run]")
		os.Exit(2)
	}
	expertID, err := uuid.Parse(expertIDStr)
	if err != nil {
		log.Fatalf("invalid expert-id: %v", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL required")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Load distinct source files for this expert when not filtered
	files := []string{sourceFile}
	if sourceFile == "" {
		rows, err := pool.Query(ctx, `SELECT DISTINCT source_file FROM course_chunks WHERE expert_id=$1 AND source_file IS NOT NULL`, expertID)
		if err != nil {
			log.Fatalf("query source files: %v", err)
		}
		defer rows.Close()
		files = nil
		for rows.Next() {
			var f string
			_ = rows.Scan(&f)
			if f != "" {
				files = append(files, f)
			}
		}
		_ = rows.Err()
		if len(files) == 0 {
			fmt.Println("no source files found for expert, nothing to do")
			return
		}
	}

	chunker := training.NewTextChunker(training.DefaultChunkerConfig())
	parentCfg := training.DefaultParentConfig()
	childCfg := training.DefaultChildConfig()
	totalParents, totalChildren := 0, 0

	for _, sf := range files {
		// Reconstruct raw text by concatenating chunk texts ordered by chunk_index
		rows, err := pool.Query(ctx, `SELECT text, section_path, chunk_index FROM course_chunks WHERE expert_id=$1 AND source_file=$2 ORDER BY chunk_index ASC`, expertID, sf)
		if err != nil {
			log.Printf("skip %s: %v", sf, err)
			continue
		}
		var rawParts []string
		for rows.Next() {
			var t, sp string
			var idx int
			_ = rows.Scan(&t, &sp, &idx)
			rawParts = append(rawParts, t)
		}
		rows.Close()
		if len(rawParts) == 0 {
			continue
		}
		raw := ""
		for i, p := range rawParts {
			if i > 0 {
				raw += "\n\n"
			}
			raw += p
		}
		parents, children := chunker.ChunkMarkdown(ctx, raw, parentCfg, childCfg)
		fmt.Printf("file %s: %d old pieces -> %d parents, %d children (atomic fence verified)\n", sf, len(rawParts), len(parents), len(children))
		totalParents += len(parents)
		totalChildren += len(children)
		if dryRun {
			continue
		}
		ids, err := training.StoreParents(ctx, pool, expertID, sf, parents)
		if err != nil {
			log.Printf("store parents failed for %s: %v", sf, err)
			continue
		}
		// Backfill parent_id on existing children by chunk_hash matching re-chunked hashes
		// New children that did not exist before are inserted with parent_id
		children = training.AssignParentIDs(children, parents, ids)
		for _, ch := range children {
			pidStr := ch.ParentID
			if pidStr == "" {
				continue
			}
			pid, _ := uuid.Parse(pidStr)
			_, _ = pool.Exec(ctx, `UPDATE course_chunks SET parent_id=$1, parent_index=$2, is_child=true WHERE expert_id=$3 AND chunk_hash=$4 AND parent_id IS NULL`, pid, ch.ParentIndex, expertID, ch.ChunkHash)
		}
		fmt.Printf("  -> stored %d parents, linked children for %s\n", len(ids), sf)
	}
	fmt.Printf("DONE dry_run=%v total parents=%d children=%d\n", dryRun, totalParents, totalChildren)
}
