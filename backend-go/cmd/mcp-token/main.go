// Command mcp-token mints an MCP access token.
//
// WHY a separate command and not an admin endpoint: minting a credential is an
// operator action, not a product feature. Keeping it out of the API means the
// running server has no code path that creates tokens — one less thing exposed.
//
// The token is printed exactly once. Only its hash is stored, so it cannot be
// recovered later; a lost token is revoked and replaced, not looked up.
//
// Usage:
//
//	DATABASE_URL=... mcp-token -user <user-uuid> -label "sneha laptop" \
//	    -domains "frontend,system design" -tools "get_standards,ask_expert"
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"ai_avengers/backend/internal/mcp"
)

func main() {
		userID := flag.String("user", "", "user UUID the token acts as (required)")
	label := flag.String("label", "", "human label shown in the audit log (required)")
	domains := flag.String("domains", "", "comma-separated allowed domains (empty = all) — legacy, prefer -expert-ids")
	expertIDs := flag.String("expert-ids", "", "comma-separated allowed expert IDs (empty = all)")
	tools := flag.String("tools", "", "comma-separated allowed tools (empty = all)")
	flag.Parse()

	if err := run(*userID, *label, *domains, *expertIDs, *tools); err != nil {
		fmt.Fprintln(os.Stderr, "mcp-token:", err)
		os.Exit(1)
	}
}

func run(userID, label, domains, expertIDs, tools string) error {
	userID = strings.TrimSpace(userID)
	label = strings.TrimSpace(label)
	if userID == "" || label == "" {
		return fmt.Errorf("-user and -label are required")
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	token, err := mcp.NewToken()
	if err != nil {
		return err
	}

		_, err = pool.Exec(ctx,
		`INSERT INTO mcp_tokens (user_id, token_hash, label, domains, expert_ids, tools)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		userID, mcp.HashToken(token), label, splitList(domains), splitList(expertIDs), splitList(tools))
	if err != nil {
		return fmt.Errorf("store token: %w", err)
	}

	fmt.Println(token)
	fmt.Fprintln(os.Stderr, "Store it now — only its hash is kept, so it cannot be shown again.")
	return nil
}

// splitList turns "a, b ,c" into ["a","b","c"]; empty means "no restriction".
func splitList(raw string) []string {
	out := []string{}
	for _, part := range strings.Split(raw, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
