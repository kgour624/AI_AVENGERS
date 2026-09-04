package training

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// Charter holds both reasoning and clarification charters for an expert.
// These are extracted from course transcripts and define the expert's
// personality, decision-making rules, and teaching style.
type Charter struct {
	ReasoningCharter     string              `json:"reasoning_charter"`
	ClarificationCharter map[string][]string `json:"clarification_charter"`
}

// CharterExtractor extracts expert charters from course transcripts.
// This is where the WHY principle is embedded into each expert.
//
// WHY charter extraction matters:
// Without charters, experts are just search engines.
// With charters, experts have opinions, push back on bad ideas,
// ask clarifying questions, and reason like the actual instructor.
// This is what makes AI Avengers different from generic AI.
type CharterExtractor struct {
	gateway *gateway.ModelGateway
	logger  *zap.Logger
}

// NewCharterExtractor creates a new charter extractor.
func NewCharterExtractor(gw *gateway.ModelGateway, logger *zap.Logger) *CharterExtractor {
	return &CharterExtractor{gateway: gw, logger: logger}
}

// Extract extracts both charters from a transcript.
// Uses strong model — charter quality is critical.
// WHY strong model: Charter is used in EVERY response.
// A bad charter = bad expert behavior forever.
// Cost: ~$0.04 per expert. Worth it.
func (e *CharterExtractor) Extract(ctx context.Context, transcript string, expertName string) (*Charter, error) {
	e.logger.Info("extracting charters",
		zap.String("expert", expertName),
		zap.Int("transcript_length", len(transcript)),
	)

	// Use first 8000 chars of transcript for charter extraction
	// WHY 8000: Enough to capture instructor's style and principles
	// Full transcript would exceed context window and add noise
	sample := transcript
	if len(sample) > 8000 {
		sample = sample[:8000]
	}

	// Extract reasoning charter
	reasoningCharter, err := e.extractReasoningCharter(ctx, sample, expertName)
	if err != nil {
		e.logger.Warn("reasoning charter extraction failed, using default",
			zap.Error(err),
		)
		reasoningCharter = defaultReasoningCharter(expertName)
	}

	// Validate WHY principle
	reasoningCharter = e.enforceWHYPrinciple(reasoningCharter)

	// Extract clarification charter
	clarificationCharter, err := e.extractClarificationCharter(ctx, sample, expertName)
	if err != nil {
		e.logger.Warn("clarification charter extraction failed, using default",
			zap.Error(err),
		)
		clarificationCharter = defaultClarificationCharter()
	}

	charter := &Charter{
		ReasoningCharter:     reasoningCharter,
		ClarificationCharter: clarificationCharter,
	}

	e.logger.Info("charter extraction complete",
		zap.String("expert", expertName),
		zap.Int("reasoning_length", len(reasoningCharter)),
		zap.Int("clarification_topics", len(clarificationCharter)),
	)

	return charter, nil
}

// extractReasoningCharter extracts the instructor's decision-making rules.
// CRITICAL: Every rule must include WHY (BECAUSE clause).
// This is Arpit's principle: humans are intelligent because they know WHY
// they do what they do. Same must apply to AI experts.
func (e *CharterExtractor) extractReasoningCharter(ctx context.Context, transcript, expertName string) (string, error) {
	prompt := fmt.Sprintf(`You are analyzing a course transcript to extract the instructor's core principles.

Instructor: %s

Transcript:
%s

Extract the instructor's reasoning charter — their decision-making rules, opinions, and principles.

CRITICAL RULE: Every principle MUST include a BECAUSE clause explaining WHY.
This is non-negotiable. If you cannot find the reason, do not include the rule.

NOT acceptable: "Never use microservices for small teams"
ACCEPTABLE: "Never use microservices for small teams BECAUSE operational overhead exceeds development velocity when team < 10 people"

Extract:
1. Strong opinions with reasoning ("I believe X BECAUSE Y")
2. Never/Always rules with reasoning ("Never do X BECAUSE it causes Y")
3. Decision rules with context ("When X, do Y BECAUSE Z")
4. Anti-patterns with explanation ("Avoid X BECAUSE it leads to Y")
5. Trade-off principles ("Choose X over Y when Z BECAUSE W")

Format as plain text with clear sections:
## Core Principles
## Decision Rules  
## Anti-Patterns
## Trade-offs

Write in first person as if the instructor is speaking.
Be specific and technical. Include numbers where the instructor mentioned them.
Do NOT include generic advice. Only what this specific instructor taught.`,
		expertName, transcript)

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelStrong,
		UserPrompt:  prompt,
		MaxTokens:   1500,
		Temperature: 0.3,
		UseCache:    false,
	})
	if err != nil {
		return "", fmt.Errorf("LLM call failed: %w", err)
	}

	return strings.TrimSpace(resp.Content), nil
}

// extractClarificationCharter extracts topic-specific questions the instructor asks.
// These questions are used in Gate 1 (Information Sufficiency) of the decision engine.
// WHY clarification questions matter:
// "Should I use sharding?" — cannot answer without knowing data size, team size, scale.
// The clarification charter tells the expert WHAT to ask before answering.
func (e *CharterExtractor) extractClarificationCharter(ctx context.Context, transcript, expertName string) (map[string][]string, error) {
	prompt := fmt.Sprintf(`You are analyzing a course transcript to extract clarification questions.

Instructor: %s

Transcript:
%s

For each major technical topic covered, identify the questions the instructor asks
BEFORE giving advice. These are the questions that determine WHICH answer is correct.

Examples:
- For "sharding": "What is your current data size?", "What is your read/write ratio?"
- For "caching": "What is your cache hit rate requirement?", "How often does data change?"
- For "microservices": "What is your team size?", "What is your current scale?"

Return ONLY valid JSON:
{
  "topic_name": ["Question 1?", "Question 2?", "Question 3?"],
  "another_topic": ["Question 1?", "Question 2?"]
}

Rules:
- Use lowercase_underscore for topic names
- Maximum 5 questions per topic
- Questions must be specific and technical
- Only include topics actually covered in the transcript
- Return ONLY JSON, no explanation`,
		expertName, transcript)

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelStrong,
		UserPrompt:  prompt,
		MaxTokens:   1000,
		Temperature: 0.2,
		UseCache:    false,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	return parseClarificationCharter(resp.Content)
}

// enforceWHYPrinciple validates and enhances the reasoning charter.
// Checks that rules contain BECAUSE clauses.
// Logs warnings for rules without WHY — does not remove them (LLM may have
// embedded the reason differently).
func (e *CharterExtractor) enforceWHYPrinciple(charter string) string {
	lines := strings.Split(charter, "\n")
	whyCount := 0
	totalRules := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			totalRules++
			lineLower := strings.ToLower(line)
			if strings.Contains(lineLower, "because") ||
				strings.Contains(lineLower, "since") ||
				strings.Contains(lineLower, "as it") ||
				strings.Contains(lineLower, "to avoid") ||
				strings.Contains(lineLower, "to prevent") {
				whyCount++
			}
		}
	}

	if totalRules > 0 {
		whyRatio := float64(whyCount) / float64(totalRules)
		e.logger.Info("WHY principle coverage",
			zap.Float64("ratio", whyRatio),
			zap.Int("rules_with_why", whyCount),
			zap.Int("total_rules", totalRules),
		)
		if whyRatio < 0.5 {
			e.logger.Warn("less than 50% of rules have WHY clause — charter quality may be low")
		}
	}

	return charter
}

// parseClarificationCharter parses LLM JSON response into map.
func parseClarificationCharter(response string) (map[string][]string, error) {
	// Strip markdown
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Find JSON object
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("no JSON object found")
	}
	response = response[start : end+1]

	var charter map[string][]string
	if err := json.Unmarshal([]byte(response), &charter); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	// Normalize topic names and limit questions
	normalized := make(map[string][]string)
	for topic, questions := range charter {
		normalizedTopic := normalizeTopicName(topic)
		if len(questions) > 5 {
			questions = questions[:5]
		}
		normalized[normalizedTopic] = questions
	}

	return normalized, nil
}

// defaultReasoningCharter returns a fallback charter when extraction fails.
func defaultReasoningCharter(expertName string) string {
	return fmt.Sprintf(`## Core Principles
- Always provide accurate, well-sourced information BECAUSE incorrect advice has real consequences
- Consider the user's context and constraints BECAUSE the right solution depends on the specific situation
- Recommend practical, proven solutions BECAUSE theoretical solutions that don't work in practice are useless

## Decision Rules
- Always ask for clarification when context is unclear BECAUSE assumptions lead to wrong recommendations
- Never recommend solutions outside my knowledge domain BECAUSE I can only vouch for what I was taught
- Warn about potential pitfalls and trade-offs BECAUSE every solution has costs

## Anti-Patterns
- Avoid making assumptions about user's requirements BECAUSE requirements vary widely
- Don't recommend technologies without understanding the use case BECAUSE technology choice depends on context

## Trade-offs
- Prefer simple solutions over complex ones BECAUSE complexity has maintenance costs
- Start with fundamentals before advanced topics BECAUSE foundations must be solid

Note: This is a default charter for %s. Upload transcripts to extract specific principles.`,
		expertName)
}

// defaultClarificationCharter returns fallback clarification questions.
func defaultClarificationCharter() map[string][]string {
	return map[string][]string{
		"general": {
			"What is your current setup?",
			"What are your specific requirements?",
			"What is your scale (users, data size, traffic)?",
			"What is your team size and experience level?",
		},
		"architecture": {
			"What is your expected scale?",
			"What are your performance requirements?",
			"What is your team size?",
			"Do you need high availability?",
		},
		"database": {
			"What is your data size?",
			"What is your read/write ratio?",
			"What are your consistency requirements?",
			"Do you need ACID guarantees?",
		},
	}
}
