package qwen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"ds2api/internal/auth"
	"ds2api/internal/config"
)

func (c *Client) CallCompletion(ctx context.Context, _ *auth.RequestAuth, payload map[string]any, _ string, maxAttempts int) (*http.Response, error) {
	entry, err := c.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("qwen pool acquire failed: %w", err)
	}
	ticket := entry.Ticket
	label := entry.Label
	defer c.pool.Release(label)

	messagesRaw, _ := payload["messages"].([]map[string]any)
	rawModel, _ := payload["model"].(string)
	modelName := resolveModelName(rawModel)

	l2Messages := BuildL2Messages(messagesRaw)
	reqID := strings.ReplaceAll(uuid.New().String(), "-", "")
	chatID := reqID
	sessionID := chatID
	topicID := strings.ReplaceAll(uuid.New().String(), "-", "")
	clientTm := time.Now().UnixMilli()

	bodyPayload := map[string]any{
		"deep_search":      "0",
		"req_id":           reqID,
		"model":            modelName,
		"scene":            "chat",
		"session_id":       sessionID,
		"sub_scene":        "chat",
		"temporary":        false,
		"messages":         l2Messages,
		"from":             "default",
		"topic_id":         topicID,
		"parent_req_id":    "0",
		"scene_param":      "first_turn",
		"chat_client":      "h5",
		"client_tm":        fmt.Sprintf("%d", clientTm),
		"protocol_version": "v2",
		"biz_id":           "ai_qwen",
	}

	if mt, ok := payload["max_tokens"]; ok {
		if v, ok := mt.(float64); ok && v > 0 {
			bodyPayload["max_tokens"] = int(v)
		} else if v, ok := mt.(int); ok && v > 0 {
			bodyPayload["max_tokens"] = v
		}
	}

	jsonBody, err := json.Marshal(bodyPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	var lastErr error
	var currentLabel = label
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(RetryDelayMs) * time.Millisecond):
			}
			retryEntry, err := c.pool.Acquire(ctx)
			if err != nil {
				return nil, fmt.Errorf("qwen pool acquire failed on retry: %w", err)
			}
			ticket = retryEntry.Ticket
			currentLabel = retryEntry.Label
			defer c.pool.Release(currentLabel)
		}

		resp, err := c.doPostChat(ctx, ticket, reqID, jsonBody)
		if err != nil {
			lastErr = err
			c.pool.MarkFailed(currentLabel)
			config.Logger.Warn("[qwen-l2] call failed", "attempt", attempt+1, "error", err)
			continue
		}
		c.pool.MarkSuccess(currentLabel)
		return resp, nil
	}
	return nil, fmt.Errorf("all %d attempts failed, last: %w", maxAttempts, lastErr)
}

func (c *Client) doPostChat(ctx context.Context, ticket string, chatID string, jsonBody []byte) (*http.Response, error) {
	serverTime, _ := c.calibrateTime(ctx)
	nonce := c.generateNonce()
	queryParams := fmt.Sprintf(
		"?biz_id=ai_qwen&chat_client=h5&device=pc&fr=pc&pr=qwen&ut=%s&la=zh-CN&tz=Asia/Shanghai&nonce=%s&timestamp=%d",
		c.deviceID, nonce, serverTime,
	)
	url := QwenChatURL + queryParams

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setSecurityHeaders(req, ticket, chatID, serverTime)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		config.Logger.Warn("[qwen-l2] auth failed response",
			"status", resp.StatusCode,
			"body", truncateStr(string(body), 1000),
		)
		return nil, fmt.Errorf("auth failed (status %d), ticket may be invalid or security headers rejected, body: %s", resp.StatusCode, truncateStr(string(body), 500))
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, truncateStr(string(b), 500))
	}

	return resp, nil
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
