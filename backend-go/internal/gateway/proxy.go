package gateway

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// openAIMessage mirrors the OpenAI chat message format.
// Aider sends requests in this format.
type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIProxyRequest is the OpenAI-compatible request body Aider sends.
type openAIProxyRequest struct {
	Model     string          `json:"model"`
	Messages  []openAIMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
}

// openAIProxyResponse is the OpenAI-compatible response Aider expects.
type openAIProxyResponse struct {
	Choices []openAIChoice `json:"choices"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

// ProxyHandler handles OpenAI-compatible LLM requests from AiderService.
// Aider sends requests in OpenAI chat completions format.
// This handler translates to LLMRequest, calls ModelGateway, returns response.
// Headers X-Workflow-ID and X-Expert-ID are logged for cost attribution.
//
// Route: POST /api/v1/llm/proxy
// Auth: JWT required (protected route)
func (g *ModelGateway) ProxyHandler(c *gin.Context) {
	var req openAIProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log cost attribution headers
	workflowID := c.GetHeader("X-Workflow-ID")
	expertID := c.GetHeader("X-Expert-ID")
	g.logger.Info("llm proxy call",
		zap.String("workflow_id", workflowID),
		zap.String("expert_id", expertID),
		zap.String("model", req.Model),
		zap.Int("max_tokens", req.MaxTokens),
	)

	// Extract system and user messages from OpenAI messages array
	var systemPrompt, userPrompt string
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			systemPrompt = msg.Content
		case "user":
			userPrompt = msg.Content
		}
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4000
	}

	// Call ModelGateway — preserves cost tracking, retry, provider switching
	resp, err := g.Call(c.Request.Context(), LLMRequest{
		Model:        ModelStrong,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    maxTokens,
	})
	if err != nil {
		g.logger.Error("llm proxy call failed",
			zap.String("workflow_id", workflowID),
			zap.String("expert_id", expertID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return OpenAI-compatible response format
	c.JSON(http.StatusOK, openAIProxyResponse{
		Choices: []openAIChoice{
			{
				Message: openAIMessage{
					Role:    "assistant",
					Content: resp.Content,
				},
			},
		},
	})
}
