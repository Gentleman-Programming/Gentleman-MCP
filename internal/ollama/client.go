package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type GenerateRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	Stream   bool                   `json:"stream"`
	Context  []int                  `json:"context,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
	System   string                 `json:"system,omitempty"`
	Template string                 `json:"template,omitempty"`
	Raw      bool                   `json:"raw,omitempty"`
}

type GenerateResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

type ChatSession struct {
	Model   string
	Context []int
	System  string
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // Gemma3 may take a little while
		},
	}
}

func (c *Client) Generate(ctx context.Context, model, prompt string) (string, error) {
	return c.GenerateWithOptions(ctx, model, prompt, nil)
}

func (c *Client) GenerateWithOptions(ctx context.Context, model, prompt string, options map[string]interface{}) (string, error) {
	// 1. Prepare the request
	req := GenerateRequest{
		Model:   model,
		Prompt:  prompt,
		Stream:  false,
		Options: options,
	}

	// 2. Serialize the request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// 3. Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 4. Set the headers
	httpReq.Header.Set("Content-Type", "application/json")

	// 5. Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call ollama: %w", err)
	}

	defer resp.Body.Close()

	// 6. Check the response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned non-200 status: %s", resp.Status)
	}

	// 7. Decode the response
	var ollamaResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// 8. Return the generated text
	return ollamaResp.Response, nil
}

func (c *Client) GenerateWithContext(ctx context.Context, model, prompt string, contextData []int, system string) (*GenerateResponse, error) {
	// 1. Prepare the request with context
	req := GenerateRequest{
		Model:   model,
		Prompt:  prompt,
		Stream:  false,
		Context: contextData,
		System:  system,
	}

	// 2. Serialize the request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 3. Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 4. Set the headers
	httpReq.Header.Set("Content-Type", "application/json")

	// 5. Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call ollama: %w", err)
	}

	defer resp.Body.Close()

	// 6. Check the response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned non-200 status: %s", resp.Status)
	}

	// 7. Decode the response
	var ollamaResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 8. Return the full response with context
	return &ollamaResp, nil
}

func (c *Client) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/version", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ollama health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama health check returned status: %s", resp.Status)
	}

	return nil
}

func NewChatSession(model, system string) *ChatSession {
	return &ChatSession{
		Model:   model,
		Context: make([]int, 0),
		System:  system,
	}
}

func (s *ChatSession) GenerateResponse(ctx context.Context, client *Client, prompt string) (string, error) {
	resp, err := client.GenerateWithContext(ctx, s.Model, prompt, s.Context, s.System)
	if err != nil {
		return "", err
	}

	// Update context for conversation continuity
	if len(resp.Context) > 0 {
		s.Context = resp.Context
	}

	return resp.Response, nil
}
