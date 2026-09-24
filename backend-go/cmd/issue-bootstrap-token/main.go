// issue-bootstrap-token generates a one-time admin bootstrap token.
// Usage (from backend-go, with DATABASE_URL set):
//
//	go run ./cmd/issue-bootstrap-token
//
// Print the raw token once and share out-of-band. User opens:
//
//	/bootstrap/<token>
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "db connect:", err)
		os.Exit(1)
	}
	defer pool.Close()

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	raw := hex.EncodeToString(buf)
	hash := sha256.Sum256([]byte(raw))
	tokenHash := hex.EncodeToString(hash[:])

	_, err = pool.Exec(ctx,
		`INSERT INTO admin_bootstrap_tokens (token_hash, expires_at)
		 VALUES ($1, $2)`,
		tokenHash, time.Now().Add(7*24*time.Hour),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "insert:", err)
		os.Exit(1)
	}
	fmt.Println(raw)
}
