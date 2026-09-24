package eval

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ExpertResolver maps a golden-set expert_slug to a live expert UUID.
// Returns ErrSkip when the slug is absent in this environment.
type ExpertResolver func(ctx context.Context, slug string) (uuid.UUID, error)

// HTTPAnswerer calls the live chat SSE endpoint for each case.
//
// Env prerequisites (wired by cmd/eval):
//   - BaseURL: e.g. http://localhost:8080/api/v1
//   - Token: Bearer access_token for a user that owns ChatID
//   - ChatID: an existing chat the token can write to
//   - Resolve: slug → expert_id (typically a DB lookup)
type HTTPAnswerer struct {
	BaseURL string
	Token   string
	ChatID  uuid.UUID
	Resolve ExpertResolver
	Client  *http.Client
}

// Answer posts one case to POST /chats/:id/messages and aggregates the
// SSE "complete" payload into an Observed answer.
func (h *HTTPAnswerer) Answer(ctx context.Context, c Case) (Observed, error) {
	if h == nil || h.BaseURL == "" || h.Token == "" || h.ChatID == uuid.Nil {
		return Observed{}, fmt.Errorf("eval: HTTPAnswerer not configured")
	}
	if h.Resolve == nil {
		return Observed{}, fmt.Errorf("eval: HTTPAnswerer missing expert resolver")
	}
	expertID, err := h.Resolve(ctx, c.ExpertSlug)
	if err != nil {
		return Observed{}, err
	}

	body, err := json.Marshal(map[string]interface{}{
		"message":    c.Input,
		"expert_ids": []string{expertID.String()},
	})
	if err != nil {
		return Observed{}, err
	}

	url := strings.TrimRight(h.BaseURL, "/") + "/chats/" + h.ChatID.String() + "/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return Observed{}, err
	}
	req.Header.Set("Authorization", "Bearer "+h.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	client := h.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Minute}
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return Observed{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return Observed{}, fmt.Errorf("eval: chat HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}

	obs, err := readSSEAnswer(resp.Body)
	if err != nil {
		return Observed{}, err
	}
	obs.LatencyMs = time.Since(start).Milliseconds()
	return obs, nil
}

// readSSEAnswer consumes the chat SSE stream until "done"/"error" and
// folds complete-event payloads into one Observed. Pure enough to unit
// test via a bytes.Reader. SSE wire format (message.sendSSE):
//
//	data: {"type":"complete","data":{...}}\n\n
func readSSEAnswer(r io.Reader) (Observed, error) {
	var obs Observed
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	gotComplete := false
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" {
			continue
		}
		var evt struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal([]byte(payload), &evt); err != nil {
			continue
		}
		switch evt.Type {
		case "complete":
			var data struct {
				Content   string          `json:"content"`
				Mode      string          `json:"mode"`
				Citations json.RawMessage `json:"citations"`
			}
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				continue
			}
			if data.Content != "" {
				if obs.Content != "" {
					obs.Content += "\n"
				}
				obs.Content += data.Content
			}
			if data.Mode != "" {
				obs.Mode = data.Mode
			}
			if len(data.Citations) > 0 && string(data.Citations) != "null" {
				var cites []json.RawMessage
				if json.Unmarshal(data.Citations, &cites) == nil {
					obs.Citations += len(cites)
				}
			}
			gotComplete = true
		case "error":
			var data struct {
				Message string `json:"message"`
			}
			_ = json.Unmarshal(evt.Data, &data)
			msg := data.Message
			if msg == "" {
				msg = "sse error"
			}
			return obs, fmt.Errorf("eval: %s", msg)
		case "done":
			// stream finished
			goto done
		}
	}
done:
	if err := sc.Err(); err != nil {
		return obs, err
	}
	if !gotComplete && strings.TrimSpace(obs.Content) == "" {
		return obs, fmt.Errorf("eval: no complete event in SSE stream")
	}
	if strings.EqualFold(obs.Mode, "REFUSE") {
		obs.Refused = true
	}
	return obs, nil
}
