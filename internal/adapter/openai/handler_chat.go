package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"ds2api/internal/admin"
	"ds2api/internal/auth"
	"ds2api/internal/config"
	openaifmt "ds2api/internal/format/openai"
	"ds2api/internal/limits"
	"ds2api/internal/qwen"
	"ds2api/internal/sse"
	streamengine "ds2api/internal/stream"
	"ds2api/internal/util"
)

func (h *Handler) ChatCompletions(w http.ResponseWriter, r *http.Request) {
	if isVercelStreamReleaseRequest(r) {
		h.handleVercelStreamRelease(w, r)
		return
	}
	if isVercelStreamPrepareRequest(r) {
		h.handleVercelStreamPrepare(w, r)
		return
	}

	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "invalid json")
		return
	}
	rawModel, _ := req["model"].(string)
	useQwen := isQwenModel(rawModel)

	var a *auth.RequestAuth
	var err error
	if useQwen {
		a, err = h.Auth.DetermineCaller(r)
	} else {
		a, err = h.Auth.Determine(r)
	}
	if err != nil {
		status := http.StatusUnauthorized
		detail := err.Error()
		if err == auth.ErrNoAccount {
			status = http.StatusTooManyRequests
		}
		writeOpenAIError(w, status, detail)
		return
	}
	defer func() {
		if h.Store.AutoDeleteSessions() && a.DeepSeekToken != "" {
			deleteCtx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()
			err := h.DS.DeleteAllSessionsForToken(deleteCtx, a.DeepSeekToken)
			if err != nil {
				config.Logger.Warn("[auto_delete_sessions] failed", "account", a.AccountID, "error", err)
			} else {
				config.Logger.Debug("[auto_delete_sessions] success", "account", a.AccountID)
			}
		}
		h.Auth.Release(a)
	}()

	r = r.WithContext(auth.WithAuth(r.Context(), a))

	msgsRaw := extractMessagesFromRequest(req)
	stream := util.ToBool(req["stream"])
	engineType := "deepseek"
	if useQwen {
		engineType = "qwen"
	}
	startTime := time.Now()
	var sessionErr error
	var sessionStatus admin.SessionStatus = admin.SessionSuccess
	defer func() {
		latency := time.Since(startTime).Milliseconds()
		errMsg := ""
		if sessionErr != nil {
			errMsg = sessionErr.Error()
			sessionStatus = admin.SessionError
		}
		admin.SessionLogAppend(rawModel, engineType, a.CallerID, len(msgsRaw), stream, sessionStatus, latency, r, errMsg)
	}()

	if limits.Enabled {
		msgsRaw := extractMessagesFromRequest(req)
		maxTok := 0
		if v, ok := req["max_tokens"].(float64); ok && v > 0 {
			maxTok = int(v)
		} else if v, ok := req["max_tokens"].(int); ok && v > 0 {
			maxTok = v
		}
		prot := limits.CheckAndProtect(rawModel, msgsRaw, maxTok)
		if len(prot.Warnings) > 0 {
			config.Logger.Warn(limits.JoinWarnings(prot.Warnings),
				"model", rawModel, "input_tokens", prot.InputTokens,
				"output_tokens", prot.OutputTokens, "context_usage_pct", prot.ContextUsage)
		}
		if prot.Truncated {
			req["messages"] = prot.Messages
		}
		if prot.OutputTokens > 0 {
			req["max_tokens"] = prot.OutputTokens
		}
	}

	if useQwen && h.QW == nil {
		sessionErr = fmt.Errorf("qwen backend not configured")
		writeOpenAIError(w, http.StatusServiceUnavailable, "Qwen backend not configured. Add qwen_accounts to config.json.")
		return
	}
	if useQwen {
		sessionErr = h.handleQwenChatRaw(w, r, a, req)
		return
	}

	stdReq, err := normalizeOpenAIChatRequest(h.Store, req, requestTraceID(r))
	if err != nil {
		sessionErr = err
		writeOpenAIError(w, http.StatusBadRequest, err.Error())
		return
	}

	sessionID, err := h.DS.CreateSession(r.Context(), a, 3)
	if err != nil {
		sessionErr = err
		if a.UseConfigToken {
			writeOpenAIError(w, http.StatusUnauthorized, "Account token is invalid. Please re-login the account in admin.")
		} else {
			writeOpenAIError(w, http.StatusUnauthorized, "Invalid token. If this should be a DS2API key, add it to config.keys first.")
		}
		return
	}
	pow, err := h.DS.GetPow(r.Context(), a, 3)
	if err != nil {
		sessionErr = err
		writeOpenAIError(w, http.StatusUnauthorized, "Failed to get PoW (invalid token or unknown error).")
		return
	}
	payload := stdReq.CompletionPayload(sessionID)
	resp, err := h.DS.CallCompletion(r.Context(), a, payload, pow, 3)
	if err != nil {
		sessionErr = err
		writeOpenAIError(w, http.StatusInternalServerError, "Failed to get completion.")
		return
	}
	if stdReq.Stream {
		h.handleStream(w, r, resp, sessionID, stdReq.ResponseModel, stdReq.FinalPrompt, stdReq.Thinking, stdReq.Search, stdReq.ToolNames)
		return
	}
	h.handleNonStream(w, r.Context(), resp, sessionID, stdReq.ResponseModel, stdReq.FinalPrompt, stdReq.Thinking, stdReq.ToolNames)
}

func (h *Handler) handleQwenChatRaw(w http.ResponseWriter, r *http.Request, a *auth.RequestAuth, req map[string]any) error {
	messagesRaw := extractMessagesFromRequest(req)

	sessionID, _ := h.QW.CreateSession(r.Context(), a, 3)
	pow, _ := h.QW.GetPow(r.Context(), a, 3)

	rawModel, _ := req["model"].(string)
	if rawModel == "" {
		rawModel = "qwen/qwen-plus"
	}

	payload := map[string]any{
		"model":         rawModel,
		"messages":      messagesRaw,
		"session_id":    sessionID,
		"parent_msg_id": "",
	}
	resp, err := h.QW.CallCompletion(r.Context(), a, payload, pow, qwen.MaxRetryCount)
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "Qwen completion failed: "+err.Error())
		return err
	}

	stream := util.ToBool(req["stream"])
	completionID := "qwenchatcmpl-" + strings.ReplaceAll(uuid.New().String(), "-", "")[:24]
	if stream {
		h.handleQwenStream(w, r, resp, completionID, rawModel, "")
		return nil
	}
	h.handleQwenNonStream(w, r.Context(), resp, completionID, rawModel, "")
	return nil
}

func (h *Handler) handleQwenChat(w http.ResponseWriter, r *http.Request, a *auth.RequestAuth, stdReq util.StandardRequest) {
	messagesRaw := extractMessagesFromRequest(map[string]any{"messages": stdReq.Messages})

	sessionID, _ := h.QW.CreateSession(r.Context(), a, 3)
	pow, _ := h.QW.GetPow(r.Context(), a, 3)

	payload := map[string]any{
		"messages":      messagesRaw,
		"session_id":    sessionID,
		"parent_msg_id": "",
	}
	resp, err := h.QW.CallCompletion(r.Context(), a, payload, pow, qwen.MaxRetryCount)
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "Qwen completion failed: "+err.Error())
		return
	}

	completionID := "qwenchatcmpl-" + strings.ReplaceAll(uuid.New().String(), "-", "")[:24]
	if stdReq.Stream {
		h.handleQwenStream(w, r, resp, completionID, stdReq.ResponseModel, stdReq.FinalPrompt)
		return
	}
	h.handleQwenNonStream(w, r.Context(), resp, completionID, stdReq.ResponseModel, stdReq.FinalPrompt)
}

func (h *Handler) handleQwenNonStream(w http.ResponseWriter, ctx context.Context, resp *http.Response, completionID, model, finalPrompt string) {
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		writeOpenAIError(w, resp.StatusCode, string(body))
		return
	}
	result := collectQwenStream(resp, true)
	finalText := sanitizeLeakedOutput(result.Text)
	respBody := openaifmt.BuildChatCompletion(completionID, model, finalPrompt, "", finalText, nil)
	writeJSON(w, http.StatusOK, respBody)
}

func (h *Handler) handleQwenStream(w http.ResponseWriter, r *http.Request, resp *http.Response, completionID, model, finalPrompt string) {
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		writeOpenAIError(w, resp.StatusCode, string(body))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	rc := http.NewResponseController(w)
	flusher, canFlush := w.(http.Flusher)
	if !canFlush {
		config.Logger.Warn("[qwen-stream] response writer does not support flush")
	}

	created := time.Now().Unix()
	emitInitialChunk(w, flusher, completionID, created, model)

	scanner := qwen.NewQwenSSEScanner(resp.Body)
	for scanner.Next() {
		event := scanner.Event()
		if event.Done {
			break
		}
		if event.Text == "" {
			continue
		}
		chunk := buildStreamTextChunk(completionID, created, model, event.Text)
		w.Write([]byte("data: "))
		w.Write(chunk)
		w.Write([]byte("\n\n"))
		if canFlush {
			flusher.Flush()
		}
		rc.SetWriteDeadline(time.Now().Add(30 * time.Second))
	}
	if scanner.Err() != nil {
		config.Logger.Warn("[qwen-stream] scan error", "error", scanner.Err())
	}

	doneChunk := buildDoneChunk(completionID, created, model)
	w.Write([]byte("data: "))
	w.Write(doneChunk)
	w.Write([]byte("\n\ndata: [DONE]\n\n"))
	if canFlush {
		flusher.Flush()
	}
}

func emitInitialChunk(w http.ResponseWriter, flusher http.Flusher, id string, created int64, model string) {
	initial := buildStreamTextChunk(id, created, model, "")
	w.Write([]byte("data: "))
	w.Write(initial)
	w.Write([]byte("\n\n"))
	if flusher != nil {
		flusher.Flush()
	}
}

func collectQwenStream(resp *http.Response, closeBody bool) sse.CollectResult {
	if closeBody {
		defer resp.Body.Close()
	}
	text := ""
	scanner := qwen.NewQwenSSEScanner(resp.Body)
	for scanner.Next() {
		text += scanner.Event().Text
	}
	return sse.CollectResult{Text: text}
}

func isQwenModel(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(model, "qwen/") ||
		strings.HasPrefix(model, "qwen-") ||
		strings.HasPrefix(model, "qwen3.") ||
		model == "qwen"
}

func extractMessagesFromRequest(payload map[string]any) []map[string]any {
	msgs, ok := payload["messages"].([]map[string]any)
	if !ok {
		var raw []any
		if raw, ok = payload["messages"].([]any); ok {
			msgs = make([]map[string]any, 0, len(raw))
			for _, r := range raw {
				if m, ok := r.(map[string]any); ok {
					msgs = append(msgs, m)
				}
			}
		}
	}
	return msgs
}

func (h *Handler) handleNonStream(w http.ResponseWriter, ctx context.Context, resp *http.Response, completionID, model, finalPrompt string, thinkingEnabled bool, toolNames []string) {
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		writeOpenAIError(w, resp.StatusCode, string(body))
		return
	}
	_ = ctx
	result := sse.CollectStream(resp, thinkingEnabled, true)

	finalThinking := result.Thinking
	finalText := sanitizeLeakedOutput(result.Text)
	respBody := openaifmt.BuildChatCompletion(completionID, model, finalPrompt, finalThinking, finalText, toolNames)
	writeJSON(w, http.StatusOK, respBody)
}

func (h *Handler) handleStream(w http.ResponseWriter, r *http.Request, resp *http.Response, completionID, model, finalPrompt string, thinkingEnabled, searchEnabled bool, toolNames []string) {
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		writeOpenAIError(w, resp.StatusCode, string(body))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	rc := http.NewResponseController(w)
	_, canFlush := w.(http.Flusher)
	if !canFlush {
		config.Logger.Warn("[stream] response writer does not support flush; streaming may be buffered")
	}

	created := time.Now().Unix()
	bufferToolContent := len(toolNames) > 0
	emitEarlyToolDeltas := h.toolcallFeatureMatchEnabled() && h.toolcallEarlyEmitHighConfidence()
	initialType := "text"
	if thinkingEnabled {
		initialType = "thinking"
	}

	streamRuntime := newChatStreamRuntime(
		w,
		rc,
		canFlush,
		completionID,
		created,
		model,
		finalPrompt,
		thinkingEnabled,
		searchEnabled,
		toolNames,
		bufferToolContent,
		emitEarlyToolDeltas,
	)

	streamengine.ConsumeSSE(streamengine.ConsumeConfig{
		Context:             r.Context(),
		Body:                resp.Body,
		ThinkingEnabled:     thinkingEnabled,
		InitialType:         initialType,
		KeepAliveInterval:   time.Duration(30) * time.Second,
		IdleTimeout:         time.Duration(30) * time.Second,
		MaxKeepAliveNoInput: 5,
	}, streamengine.ConsumeHooks{
		OnKeepAlive: func() {
			streamRuntime.sendKeepAlive()
		},
		OnParsed: streamRuntime.onParsed,
		OnFinalize: func(reason streamengine.StopReason, _ error) {
			if string(reason) == "content_filter" {
				streamRuntime.finalize("content_filter")
				return
			}
			streamRuntime.finalize("stop")
		},
	})
}

func buildStreamTextChunk(id string, created int64, model, text string) []byte {
	choices := []map[string]any{
		{
			"delta": map[string]any{"content": text},
			"index": 0,
		},
	}
	chunk := openaifmt.BuildChatStreamChunk(id, created, model, choices, nil)
	out, _ := json.Marshal(chunk)
	return out
}

func buildDoneChunk(id string, created int64, model string) []byte {
	choices := []map[string]any{
		{
			"delta":         map[string]any{},
			"index":         0,
			"finish_reason": "stop",
		},
	}
	chunk := openaifmt.BuildChatStreamChunk(id, created, model, choices, nil)
	out, _ := json.Marshal(chunk)
	return append(out, '\n', '[', 'D', 'O', 'N', 'E', ']', '\n')
}
