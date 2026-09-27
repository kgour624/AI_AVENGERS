package mcp

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
)

// AnswerRequest is one expert question from a coding agent.
type AnswerRequest struct {
	// Domain the question belongs to (already authorised by the caller's scope).
	Domain string
	// ExpertID is the resolved expert to ask; resolution is the caller's job so
	// this type stays a plain data carrier.
	ExpertID string
	// Question is what the agent wants to know.
	Question string
}

// AnswerResult is the expert's answer.
type AnswerResult struct {
	Answer    string
	Mode      string
	Citations int
}

// ExpertAnswerer is how the MCP tools reach an expert's answer.
//
// WHY an interface at this boundary: the tools must not know whether the answer
// comes from a live API call, an in-process pipeline, or a fake in a test. That
// is also what keeps this phase small — no gateway, embedder or enforcer wiring
// is duplicated into the MCP binary.
type ExpertAnswerer interface {
	Answer(ctx context.Context, req AnswerRequest) (AnswerResult, error)
}

// HTTPAnswerer asks the running API server, which already owns the whole answer
// pipeline (China Wall, citations, self-learning). Reusing it over HTTP means
// MCP cannot drift from what the product actually answers — there is exactly one
// answer implementation in this system, not two.
//
// It talks to POST {base}/api/v1/chats/{chat}/messages and reads the same SSE
// stream the UI reads, because that endpoint's payload is the one the eval
// harness already consumes. ChatID must be a dedicated chat for MCP traffic, so
// agent questions never land in a user's conversation.
type HTTPAnswerer struct {
	BaseURL string
	Token   string
	ChatID  string
	Client  *http.Client
}

// NewHTTPAnswerer builds the answerer over the API server.
func NewHTTPAnswerer(baseURL, token, chatID string) *HTTPAnswerer {
	return &HTTPAnswerer{BaseURL: baseURL, Token: token, ChatID: chatID}
}

// Answer posts one question and folds the streamed reply into one answer.
func (h *HTTPAnswerer) Answer(ctx context.Context, req AnswerRequest) (AnswerResult, error) {
	if h == nil || h.BaseURL == "" || h.Token == "" || h.ChatID == "" {
		return AnswerResult{}, NewToolError("NOT_CONFIGURED", "ask_expert is not configured on this server")
	}
	if strings.TrimSpace(req.Question) == "" {
		return AnswerResult{}, NewToolError("INVALID_ARGS", "question is required")
	}

	body, err := json.Marshal(map[string]any{
		"message":    req.Question,
		"expert_ids": []string{req.ExpertID},
	})
	if err != nil {
		return AnswerResult{}, err
	}

	url := strings.TrimRight(h.BaseURL, "/") + "/api/v1/chats/" + h.ChatID + "/messages"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return AnswerResult{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+h.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	client := h.Client
	if client == nil {
		// Generous but finite: an answer that never arrives must fail the tool
		// call rather than hold the agent forever (same ceiling the collector
		// uses server-side).
		client = &http.Client{Timeout: 15 * time.Minute}
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return AnswerResult{}, fmt.Errorf("mcp: ask the API server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		// A rejected token is the caller's problem to fix, not a server fault.
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return AnswerResult{}, NewToolError("FORBIDDEN", "the configured MCP API token was rejected")
		}
		return AnswerResult{}, fmt.Errorf("mcp: api server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}

	result, err := readSSEAnswer(resp.Body)
	if err != nil {
		return AnswerResult{}, err
	}
	if strings.TrimSpace(result.Answer) == "" {
		return AnswerResult{}, fmt.Errorf("mcp: the expert returned an empty answer")
	}
	return result, nil
}

// readSSEAnswer consumes the chat SSE stream until the answer completes.
//
// Wire format (message.sendSSE): `data: {"type":"complete","data":{...}}`.
// Unparseable lines are skipped rather than failing the call: the stream also
// carries progress events this tool does not care about.
func readSSEAnswer(r io.Reader) (AnswerResult, error) {
	var out AnswerResult
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

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
				if out.Answer != "" {
					out.Answer += "\n"
				}
				out.Answer += data.Content
			}
			if data.Mode != "" {
				out.Mode = data.Mode
			}
			if len(data.Citations) > 0 && string(data.Citations) != "null" {
				var cites []json.RawMessage
				if json.Unmarshal(data.Citations, &cites) == nil {
					out.Citations += len(cites)
				}
			}
		case "error":
			var data struct {
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			_ = json.Unmarshal(evt.Data, &data)
			msg := data.Message
			if msg == "" {
				msg = data.Error
			}
			if msg == "" {
				msg = "the answer failed"
			}
			return AnswerResult{}, fmt.Errorf("mcp: %s", msg)
		}
	}
	if err := sc.Err(); err != nil {
		return AnswerResult{}, fmt.Errorf("mcp: read answer stream: %w", err)
	}
	return out, nil
}
