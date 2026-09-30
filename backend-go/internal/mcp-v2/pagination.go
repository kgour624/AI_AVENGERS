package mcpv2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// Cursor is opaque base64(JSON) — never exposes raw offset ( §4.5 ).
type Cursor struct {
	Offset   int    `json:"offset"`
	Limit    int    `json:"limit"`
	Qhash    string `json:"qhash,omitempty"`
	ExpertID string `json:"expertId,omitempty"`
}

// EncodeCursor creates opaque cursor for next page.
func EncodeCursor(c Cursor) string {
	b, _ := json.Marshal(c)
	return base64.StdEncoding.EncodeToString(b)
}

// DecodeCursor validates and decodes opaque cursor. Invalid → ErrNotFound style error for 400 invalid_cursor.
func DecodeCursor(s string) (Cursor, error) {
	if s == "" {
		return Cursor{Offset: 0, Limit: 20}, nil
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, fmt.Errorf("invalid_cursor: %w", err)
	}
	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return Cursor{}, fmt.Errorf("invalid_cursor: %w", err)
	}
	if c.Limit <= 0 {
		c.Limit = 20
	}
	if c.Limit > 50 {
		c.Limit = 50
	}
	if c.Offset < 0 {
		return Cursor{}, fmt.Errorf("invalid_cursor: negative offset")
	}
	return c, nil
}

// NextCursor builds nextCursor if offset+limit < total.
func NextCursor(cur Cursor, total int) *string {
	next := cur.Offset + cur.Limit
	if next >= total {
		return nil
	}
	s := EncodeCursor(Cursor{Offset: next, Limit: cur.Limit, Qhash: cur.Qhash, ExpertID: cur.ExpertID})
	return &s
}