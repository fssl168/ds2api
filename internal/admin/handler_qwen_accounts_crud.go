package admin

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"ds2api/internal/config"
)

func (h *Handler) qwenPoolStatus(w http.ResponseWriter, r *http.Request) {
	if h.QW == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"detail": "Qwen backend not configured"})
		return
	}
	pool := h.QW.Pool()
	status := pool.Status()
	writeJSON(w, http.StatusOK, status)
}

func (h *Handler) listQwenAccounts(w http.ResponseWriter, r *http.Request) {
	qwenAccounts := h.Store.Snapshot().QwenAccounts
	items := make([]map[string]any, 0, len(qwenAccounts))
	for _, qa := range qwenAccounts {
		ticketPreview := qa.Ticket
		if len(ticketPreview) > 30 {
			ticketPreview = ticketPreview[:30] + "..."
		}
		items = append(items, map[string]any{
			"label":          qa.Label,
			"ticket":         qa.Ticket,
			"ticket_preview": ticketPreview,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"total": len(items),
	})
}

func (h *Handler) addQwenAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Ticket string `json:"ticket"`
		Label  string `json:"label,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": "无效的 JSON"})
		return
	}
	req.Ticket = strings.TrimSpace(req.Ticket)
	req.Label = strings.TrimSpace(req.Label)
	if req.Ticket == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": "ticket 不能为空"})
		return
	}
	if req.Label == "" {
		req.Label = fmt.Sprintf("qwen-%s", req.Ticket[:min(8, len(req.Ticket))])
	}
	err := h.Store.Update(func(c *config.Config) error {
		for _, existing := range c.QwenAccounts {
			if existing.Label == req.Label {
				return fmt.Errorf("标签已存在: %s", req.Label)
			}
		}
		c.QwenAccounts = append(c.QwenAccounts, config.QwenAccount{
			Ticket: req.Ticket,
			Label:  req.Label,
		})
		return nil
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "total": len(h.Store.Snapshot().QwenAccounts)})
}

func (h *Handler) deleteQwenAccount(w http.ResponseWriter, r *http.Request) {
	label := chi.URLParam(r, "label")
	if label == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": "缺少 label 参数"})
		return
	}
	err := h.Store.Update(func(c *config.Config) error {
		idx := -1
		for i, qa := range c.QwenAccounts {
			if qa.Label == label {
				idx = i
				break
			}
		}
		if idx < 0 {
			return fmt.Errorf("千问账号不存在: %s", label)
		}
		c.QwenAccounts = append(c.QwenAccounts[:idx], c.QwenAccounts[idx+1:]...)
		return nil
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "total": len(h.Store.Snapshot().QwenAccounts)})
}

func (h *Handler) testQwenAccount(w http.ResponseWriter, r *http.Request) {
	label := chi.URLParam(r, "label")
	if label == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": "缺少 label 参数"})
		return
	}
	var targetQA config.QwenAccount
	qwenAccounts := h.Store.Snapshot().QwenAccounts
	found := false
	for _, qa := range qwenAccounts {
		if qa.Label == label {
			targetQA = qa
			found = true
			break
		}
	}
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]any{"detail": "千问账号不存在: " + label})
		return
	}

	testResult := testQwenTicket(r.Context(), targetQA.Ticket)
	writeJSON(w, http.StatusOK, map[string]any{
		"label":  label,
		"status": testResult.Status,
		"detail": testResult.Detail,
	})
}

type qwenTestResult struct {
	Status string
	Detail string
}

func testQwenTicket(ctx context.Context, ticket string) qwenTestResult {
	deviceID := uuid.New().String()
	xsrf := uuid.New().String()
	chatID := uuid.New().String()
	hashSum := sha256.Sum256([]byte(ticket))
	ticketHash := fmt.Sprintf("tytk_hash:%x", hashSum[:16])

	cookie := fmt.Sprintf(
		"theme-mode=light; _samesite_flag_=true; tongyi_sso_ticket=%s; tongyi_sso_ticket_hash=%s; XSRF-TOKEN=%s",
		ticket, ticketHash, xsrf,
	)

	reqBody, _ := json.Marshal(map[string]any{
		"deep_search":      "0",
		"req_id":           strings.ReplaceAll(uuid.New().String(), "-", ""),
		"model":            "qwen-plus",
		"scene":            "chat",
		"session_id":       chatID,
		"sub_scene":        "chat",
		"temporary":        false,
		"messages":         []map[string]any{{"role": "user", "content": "hi"}},
		"from":             "default",
		"topic_id":         strings.ReplaceAll(uuid.New().String(), "-", ""),
		"parent_req_id":    "0",
		"scene_param":      "first_turn",
		"chat_client":      "h5",
		"client_tm":        fmt.Sprintf("%d", time.Now().UnixMilli()),
		"protocol_version": "v2",
		"biz_id":           "ai_qwen",
	})

	url := "https://chat2.qianwen.com/api/v2/chat?biz_id=ai_qwen&chat_client=h5&device=pc&fr=pc&pr=qwen&ut=" + deviceID + "&la=zh-CN&tz=Asia/Shanghai&nonce=test8bits&timestamp=" + fmt.Sprintf("%d", time.Now().UnixMilli())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(reqBody)))
	if err != nil {
		return qwenTestResult{Status: "error", Detail: fmt.Sprintf("创建请求失败: %v", err)}
	}

	req.Header.Set("Accept", "application/json, text/event-stream, text/plain, */*")
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Origin", "https://www.qianwen.com")
	req.Header.Set("Referer", "https://www.qianwen.com/chat/"+chatID)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")
	req.Header.Set("x-xsrf-token", xsrf)
	req.Header.Set("x-chat-id", chatID)
	req.Header.Set("x-deviceid", deviceID)
	req.Header.Set("x-platform", "pc_tongyi")
	req.Header.Set("Cookie", cookie)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return qwenTestResult{Status: "error", Detail: fmt.Sprintf("请求失败: %v", err)}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	switch resp.StatusCode {
	case http.StatusOK:
		if strings.Contains(bodyStr, "\"error\"") || strings.Contains(bodyStr, "\"errorCode\"") {
			return qwenTestResult{Status: "error", Detail: fmt.Sprintf("API 返回错误 (200): %s", truncate(bodyStr, 300))}
		}
		return qwenTestResult{Status: "ok", Detail: fmt.Sprintf("ticket 有效! 响应长度: %d, 标签验证通过", len(bodyStr))}
	case http.StatusUnauthorized, http.StatusForbidden:
		return qwenTestResult{Status: "error", Detail: fmt.Sprintf("ticket 无效或已过期 (HTTP %d): %s", resp.StatusCode, truncate(bodyStr, 300))}
	default:
		return qwenTestResult{Status: "error", Detail: fmt.Sprintf("意外状态码 %d: %s", resp.StatusCode, truncate(bodyStr, 300))}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
