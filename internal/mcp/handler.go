package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ds2api/internal/config"
)

var (
	storeInstance *config.Store
)

func initStore(s *config.Store) {
	storeInstance = s
}

func handleChat(ctx context.Context, arguments json.RawMessage) (*ToolResult, error) {
	var params struct {
		Model       string `json:"model"`
		Messages    []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		Stream     bool    `json:"stream"`
		Temperature float64 `json:"temperature"`
		MaxTokens   int     `json:"max_tokens"`
	}
	if err := json.Unmarshal(arguments, &params); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	if params.Model == "" {
		return nil, fmt.Errorf("model is required")
	}
	if len(params.Messages) == 0 {
		return nil, fmt.Errorf("messages is required")
	}

	apiKey := ""
	if storeInstance != nil {
		keys := storeInstance.Keys()
		if len(keys) > 0 {
			apiKey = keys[0]
		}
	}
	baseURL := "http://127.0.0.1:5001"
	if v := config.GetEnv("DS2API_BASE_URL"); v != "" {
		baseURL = v
	}

	payload := map[string]interface{}{
		"model":      params.Model,
		"messages":   params.Messages,
		"stream":     params.Stream,
	}
	if params.Temperature > 0 {
		payload["temperature"] = params.Temperature
	}
	if params.MaxTokens > 0 {
		payload["max_tokens"] = params.MaxTokens
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &ToolResult{
			Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Request failed: %v", err)}},
			IsError: true,
		}, nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return &ToolResult{
			Content: []TextContent{{Type: "text", Text: fmt.Sprintf("API error (%d): %s", resp.StatusCode, string(respBody))}},
			IsError: true,
		}, nil
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	text := extractTextFromResponse(result)
	return &ToolResult{
		Content: []TextContent{{Type: "text", Text: text}},
	}, nil
}

func extractTextFromResponse(resp map[string]interface{}) string {
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := msg["content"].(string); ok {
					return content
				}
			}
		}
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return string(data)
}

func handleListModels(_ context.Context, _ json.RawMessage) (*ToolResult, error) {
	models := []map[string]string{
		{"id": "deepseek-chat", "name": "DeepSeek Chat", "engine": "deepseek"},
		{"id": "deepseek-reasoner", "name": "DeepSeek Reasoner (Thinking)", "engine": "deepseek"},
		{"id": "deepseek-chat-search", "name": "DeepSeek Chat + Search", "engine": "deepseek"},
		{"id": "deepseek-reasoner-search", "name": "DeepSeek Reasoner + Search", "engine": "deepseek"},
		{"id": "qwen-plus", "name": "通义千问 Qwen Plus", "engine": "qwen"},
		{"id": "qwen-max", "name": "通义千问 Qwen Max", "engine": "qwen"},
		{"id": "qwen-coder", "name": "通义千问 Qwen Coder", "engine": "qwen"},
		{"id": "qwen-flash", "name": "通义千问 Qwen Flash", "engine": "qwen"},
		{"id": "qwen3.5-plus", "name": "Qwen 3.5 Plus", "engine": "qwen"},
		{"id": "qwen3.5-flash", "name": "Qwen 3.5 Flash", "engine": "qwen"},
	}

	extraInfo := "\n\nEngine details:\n- **DeepSeek**: Uses DeepSeek web chat with account pool rotation and PoW protection\n- **Qwen (通义千问)**: Uses Qwen web API with ticket-based auth and Acquire/Release pool management"

	resultData, _ := json.Marshal(models)
	return &ToolResult{
		Content: []TextContent{{
			Type: "text",
			Text: string(resultData) + extraInfo,
		}},
		StructuredContent: models,
	}, nil
}

func handleGetStatus(_ context.Context, _ json.RawMessage) (*ToolResult, error) {
	status := map[string]interface{}{
		"service":  "ds2api",
		"status":   "running",
		"engines": []string{"deepseek", "qwen"},
		"endpoints": map[string]string{
			"openai_compatible": "/v1/chat/completions",
			"claude_compatible": "/anthropic/v1/messages",
			"gemini_compatible": "/v1beta/models/{model}:generateContent",
			"mcp":               "/mcp",
			"admin":             "/admin",
		},
		"version": "1.0.0",
	}
	if storeInstance != nil {
		snap := storeInstance.Snapshot()
		status["config"] = map[string]interface{}{
			"keys_count":         len(snap.Keys),
			"accounts_count":     len(snap.Accounts),
			"qwen_accounts_count": len(snap.QwenAccounts),
			"model_aliases":      snap.ModelAliases,
		}
	}
	data, _ := json.MarshalIndent(status, "", "  ")
	return &ToolResult{
		Content: []TextContent{{Type: "text", Text: string(data)}},
		StructuredContent: status,
	}, nil
}

func handlePoolStatus(_ context.Context, arguments json.RawMessage) (*ToolResult, error) {
	var params struct {
		PoolType string `json:"pool_type"`
	}
	json.Unmarshal(arguments, &params)
	if params.PoolType == "" {
		params.PoolType = "all"
	}

	poolStatus := make(map[string]interface{})
	if params.PoolType == "all" || params.PoolType == "deepseek" {
		dsStatus := map[string]interface{}{
			"type":           "deepseek",
			"description":    "DeepSeek account pool with native PoW protection",
			"features":       []string{"auto_token_refresh", "email_mobile_login", "pow_native", "concurrent_queue", "wait_queue"},
			"inflight_limit": 2,
		}
		poolStatus["deepseek"] = dsStatus
	}

	if params.PoolType == "all" || params.PoolType == "qwen" {
		qwStatus := map[string]interface{}{
			"type":             "qwen",
			"description":      "Qwen ticket-based pool with Acquire/Release pattern",
			"features":         []string{"ticket_auth", "acquire_release", "concurrent_control", "health_check", "auto_cooldown", "wait_queue"},
			"inflight_limit":   2,
		}
		poolStatus["qwen"] = qwStatus
	}

	data, _ := json.MarshalIndent(poolStatus, "", "  ")
	return &ToolResult{
		Content: []TextContent{{Type: "text", Text: string(data)}},
		StructuredContent: poolStatus,
	}, nil
}

func handleEmbeddings(ctx context.Context, arguments json.RawMessage) (*ToolResult, error) {
	var params struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}
	if err := json.Unmarshal(arguments, &params); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}
	if len(params.Input) == 0 {
		return nil, fmt.Errorf("input is required")
	}

	baseURL := "http://127.0.0.1:5001"
	if v := config.GetEnv("DS2API_BASE_URL"); v != "" {
		baseURL = v
	}

	apiKey := ""
	if storeInstance != nil {
		keys := storeInstance.Keys()
		if len(keys) > 0 {
			apiKey = keys[0]
		}
	}

	payload := map[string]interface{}{
		"model": params.Model,
		"input": params.Input,
	}
	if payload["model"] == "" {
		payload["model"] = "text-embedding-ada002"
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/embeddings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &ToolResult{
			Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Request failed: %v", err)}},
			IsError: true,
		}, nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return &ToolResult{
			Content: []TextContent{{Type: "text", Text: fmt.Sprintf("API error (%d): %s", resp.StatusCode, string(respBody))}},
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: []TextContent{{Type: "text", Text: string(respBody)}},
	}, nil
}
