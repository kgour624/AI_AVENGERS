package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// openAIMessage is one message as Aider/litellm sends it.
//
// Content is json.RawMessage, not string: the OpenAI schema allows either
// a plain string or an array of content parts. Aider sends a string for
// text-only turns and an array when a turn carries an image. Binding it as
// string would fail the whole request on the array form.
type openAIMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// openAIContentPart is one element of the array form of message content.
type openAIContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// openAIProxyRequest is the OpenAI-compatible request body Aider sends.
type openAIProxyRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature *float64        `json:"temperature"`
	Stream      bool            `json:"stream"`
	Tools       json.RawMessage `json:"tools"`
}

// openAIProxyResponse is the OpenAI-compatible response Aider expects.
//
// The full envelope is required, not just Choices: litellm parses this body
// into the OpenAI SDK's ChatCompletion model, which needs id, object,
// created, model and a finish_reason on every choice. Usage is what Aider
// reads to report tokens and cost per iteration — omitting it made every
// iteration look free.
type openAIProxyResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
}

type openAIChoice struct {
	Index        int                `json:"index"`
	Message      openAIReplyMessage `json:"message"`
	FinishReason string             `json:"finish_reason"`
}

type openAIReplyMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// messageText flattens a message's content field to plain text.
// Array form keeps only the "text" parts — the gateway's provider contract
// is text-only, so image parts have nowhere to go.
func messageText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var parts []openAIContentPart
	if err := json.Unmarshal(raw, &parts); err == nil {
		var b strings.Builder
		for _, p := range parts {
			if p.Text == "" {
				continue
			}
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(p.Text)
		}
		return b.String()
	}
	return ""
}

// ProxyHandler serves OpenAI-compatible chat completions to AiderService.
//
// Routes (both registered, same handler):
//
//	POST /api/v1/llm/proxy
//	POST /api/v1/llm/proxy/chat/completions
//
// The second path exists because the OpenAI client library inside litellm
// appends "/chat/completions" to whatever base URL it is given. Aider's
// OPENAI_API_BASE is the first path, so the request actually lands on the
// second — without it every Aider call was a 404.
//
// Auth: service token (see middleware.ServiceTokenMiddleware), not JWT.
// AiderService is a backend process with no user session to borrow a JWT from.
//
// The whole conversation is forwarded to ModelGateway. req.Model is
// deliberately ignored: which real model runs is the admin's setting, so the
// gateway's "strong" tier is always used. Aider only needs the response shape
// to match.
func (g *ModelGateway) ProxyHandler(c *gin.Context) {
	var req openAIProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflowIDHeader := c.GetHeader("X-Workflow-ID")
	expertID := c.GetHeader("X-Expert-ID")

	// Streaming is rejected loudly instead of being silently answered with a
	// non-streaming body. litellm would hang waiting for SSE frames that
	// never arrive, and the failure would surface as an unexplained timeout
	// minutes later. AiderService builds its Coder with stream=False, so
	// this should never fire — if it does, the cause is named.
	if req.Stream {
		g.logger.Warn("llm proxy rejected streaming request",
			zap.String("workflow_id", workflowIDHeader),
			zap.String("expert_id", expertID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "streaming is not supported by /llm/proxy — send stream=false",
		})
		return
	}

	if len(req.Tools) > 0 {
		// Tool/function calling has no equivalent in the provider contract;
		// the reply would come back as prose and the caller would fail on a
		// missing tool_calls field. Named here rather than guessed at later.
		g.logger.Warn("llm proxy rejected tool-calling request",
			zap.String("workflow_id", workflowIDHeader),
			zap.String("expert_id", expertID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tool/function calling is not supported by /llm/proxy",
		})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages must not be empty"})
		return
	}

	// Forward every turn, in order, roles preserved.
	messages := make([]ProviderMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		text := messageText(m.Content)
		if text == "" {
			continue
		}
		role := m.Role
		if role == "" {
			role = "user"
		}
		messages = append(messages, ProviderMessage{Role: role, Content: text})
	}
	if len(messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages contain no text content"})
		return
	}

	// Aider does not set max_tokens, so this default is what every code
	// generation request gets. 4000 was too small to hold a first pass at a
	// backend package: the reply would be cut off mid SEARCH/REPLACE block and
	// discarded as malformed. The gateway clamps this down to whatever the
	// active provider actually allows (ModelGateway.Call), so asking for more
	// than a provider supports is safe.
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 8000
	}

	// Cost attribution: parse the header into the gateway's WorkflowID so
	// Call() adds this spend to workflows.cost_spent_usd. Previously the
	// header was only logged, so Aider iterations — the most expensive calls
	// in the system — never showed up in a workflow's budget.
	var workflowID *uuid.UUID
	if workflowIDHeader != "" {
		if parsed, err := uuid.Parse(workflowIDHeader); err == nil {
			workflowID = &parsed
		} else {
			g.logger.Warn("llm proxy: unparseable X-Workflow-ID, cost not attributed",
				zap.String("workflow_id", workflowIDHeader),
				zap.Error(err),
			)
		}
	}

	var temperature float64
	if req.Temperature != nil {
		temperature = *req.Temperature
	}

	g.logger.Info("llm proxy call",
		zap.String("workflow_id", workflowIDHeader),
		zap.String("expert_id", expertID),
		zap.String("requested_model", req.Model),
		zap.Int("messages", len(messages)),
		zap.Int("max_tokens", maxTokens),
	)

	resp, err := g.Call(c.Request.Context(), LLMRequest{
		Model:       ModelStrong,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		WorkflowID:  workflowID,
	})
	if err != nil {
		g.logger.Error("llm proxy call failed",
			zap.String("workflow_id", workflowIDHeader),
			zap.String("expert_id", expertID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	modelUsed := resp.ModelUsed
	if modelUsed == "" {
		modelUsed = req.Model
	}

	c.JSON(http.StatusOK, openAIProxyResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", uuid.NewString()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   modelUsed,
		Choices: []openAIChoice{
			{
				Index:        0,
				Message:      openAIReplyMessage{Role: "assistant", Content: resp.Content},
				FinishReason: "stop",
			},
		},
		Usage: openAIUsage{
			PromptTokens:     resp.InputTokens,
			CompletionTokens: resp.OutputTokens,
			TotalTokens:      resp.InputTokens + resp.OutputTokens,
		},
	})
}
