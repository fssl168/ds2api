package admin

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"ds2api/internal/config"
)

type SessionStatus string

const (
	SessionSuccess SessionStatus = "success"
	SessionError   SessionStatus = "error"
)

type SessionEntry struct {
	Time        time.Time     `json:"time"`
	Model       string        `json:"model"`
	Engine      string        `json:"engine"`
	CallerID    string        `json:"caller_id"`
	MessageCount int          `json:"message_count"`
	IsStream    bool          `json:"is_stream"`
	Status      SessionStatus `json:"status"`
	LatencyMs   int64         `json:"latency_ms"`
	Remote      string        `json:"remote"`
	ErrorMsg    string        `json:"error,omitempty"`
}

type SessionLog struct {
	mu      sync.RWMutex
	entries []SessionEntry
	cap     int
	next    int
	full    bool
}

var globalSessionLog = &SessionLog{
	entries: make([]SessionEntry, 500),
	cap:     500,
}

func SessionLogAppend(model, engine, callerID string, msgCount int, isStream bool, status SessionStatus, latencyMs int64, r *http.Request, errMsg string) {
	globalSessionLog.Append(SessionEntry{
		Time:         time.Now(),
		Model:        model,
		Engine:       engine,
		CallerID:     callerID,
		MessageCount: msgCount,
		IsStream:     isStream,
		Status:       status,
		LatencyMs:    latencyMs,
		Remote:       r.RemoteAddr,
		ErrorMsg:     errMsg,
	})
	config.Logger.Info("[session]",
		"model", model, "engine", engine, "caller", maskCallerID(callerID),
		"messages", msgCount, "stream", isStream, "status", status,
		"latency_ms", latencyMs)
}

func maskCallerID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:8] + "***"
}

func (l *SessionLog) Append(entry SessionEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries[l.next] = entry
	l.next++
	if l.next >= l.cap {
		l.next = 0
		l.full = true
	}
}

func (l *SessionLog) List(limit int, engineFilter, statusFilter string) []SessionEntry {
	if limit <= 0 {
		limit = 100
	}
	if limit > l.cap {
		limit = l.cap
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	size := l.next
	if l.full {
		size = l.cap
	}
	if size == 0 {
		return nil
	}
	raw := make([]SessionEntry, 0, size)
	for i := 0; i < size; i++ {
		idx := i
		if l.full {
			idx = (l.next + i) % l.cap
		}
		raw = append(raw, l.entries[idx])
	}
	engineFilter = strings.ToLower(strings.TrimSpace(engineFilter))
	statusFilter = strings.ToLower(strings.TrimSpace(statusFilter))
	filtered := raw[:0]
	for _, e := range raw {
		if engineFilter != "" && !strings.Contains(strings.ToLower(e.Engine), engineFilter) {
			continue
		}
		if statusFilter != "" && !strings.Contains(string(e.Status), statusFilter) {
			continue
		}
		filtered = append(filtered, e)
	}
	if len(filtered) <= limit {
		reverseInPlace(filtered)
		return filtered
	}
	result := make([]SessionEntry, limit)
	for i := 0; i < limit; i++ {
		result[i] = filtered[len(filtered)-1-i]
	}
	return result
}

func reverseInPlace(s []SessionEntry) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func (h *Handler) listSessionLogs(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		var n int
		if err := json.Unmarshal([]byte(v), &n); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	engineFilter := r.URL.Query().Get("engine")
	statusFilter := r.URL.Query().Get("status")
	entries := globalSessionLog.List(limit, engineFilter, statusFilter)
	writeJSON(w, http.StatusOK, map[string]any{"entries": entries, "total": len(entries)})
}
