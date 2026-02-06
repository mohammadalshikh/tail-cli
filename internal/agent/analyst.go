package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/mohammadalshikh/tail-cli/internal/analyzer"
)

type LogAnalyst interface {
	Analyze(entry analyzer.LogEntry, context []string) (string, error)
}

type OpenAIClient struct {
	apiKey     string
	httpClient http.Client
	model      string
	apiURL     string
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float32       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewOpenAIClient() (*OpenAIClient, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("Error: OPENAI_API_KEY environment variable not set")
	}
	return &OpenAIClient{
		apiKey:     key,
		httpClient: http.Client{},
		model:      "gpt-4o-mini",
		apiURL:     "https://api.openai.com/v1/chat/completions",
	}, nil
}

func buildPrompt(entry analyzer.LogEntry, context []string) string {
	contextStr := ""
	if len(context) > 0 {
		for i, line := range context {
			contextStr += fmt.Sprintf("%d. %s\n", i+1, line)
		}
	} else {
		contextStr = "No context available"
	}

	return fmt.Sprintf(
		"Analyze this log entry that shows P99 latency spike:\n\n"+
			"Latency: %.0fms\n"+
			"Entry: %s\n\n"+
			"Context (previous logs):\n%s\n"+
			"Explain briefly why this is slow.",
		entry.Latency,
		entry.Data,
		contextStr,
	)
}

func buildChatRequest(model, prompt string) ChatRequest {
	return ChatRequest{
		Model: model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You are a senior SRE analyzing production logs. Give a 2-3 sentence root cause analysis focusing on: what failed, why it failed, and immediate action. Be technical and concise.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.3,
		MaxTokens:   150,
	}
}

func (c *OpenAIClient) Analyze(
	entry analyzer.LogEntry,
	context []string) (string, error) {

	prompt := buildPrompt(entry, context)

	chatRequest := buildChatRequest(c.model, prompt)

	requestJSON, err := json.Marshal(chatRequest)
	if err != nil {
		return "", fmt.Errorf("Error (failed to marshal request JSON): %w", err)
	}

	req, err := http.NewRequest("POST", c.apiURL, bytes.NewBuffer(requestJSON))
	if err != nil {
		return "", fmt.Errorf("Error (failed to create request): %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error (API request failed): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errBody)
		return "", fmt.Errorf("Error (API response failed): %d: %v", resp.StatusCode, errBody)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("Error (failed to parse response): %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("Error (no response from AI)")
	}

	return chatResp.Choices[0].Message.Content, nil
}
