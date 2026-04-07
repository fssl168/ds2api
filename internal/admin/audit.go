package admin

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"ds2api/internal/config"
)

type AuditEntry struct {
	Time   time.Time `json:"time"`
	Action string    `json:"action"`
	Detail string    `json:"detail"`
	Remote string    `json:"remote"`
}

type AuditLog struct {
	mu      sync.RWMutex
	entries []AuditEntry
	cap     int
	next    int
	full    bool
}

var globalAudit = &AuditLog{
	entries: make([]AuditEntry, 200),
	cap:     200,
}

func AuditLogAppend(action, detail string, r *http.Request) {
	globalAudit.Append(action, detail, r)
	config.Logger.Info("[admin:audit]",
		"action", action,
		"detail", detail,
		"remote", r.RemoteAddr,
	)
}

func (l *AuditLog) Append(action, detail string, r *http.Request) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries[l.next] = AuditEntry{
		Time:   time.Now(),
		Action: action,
		Detail: detail,
		Remote: r.RemoteAddr,
	}
	l.next++
	if l.next >= l.cap {
		l.next = 0
		l.full = true
	}
}

func (l *AuditLog) List(limit int) []AuditEntry {
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
	start := size - limit
	if start < 0 {
		start = 0
	}
	result := make([]AuditEntry, 0, size-start)
	for i := start; i < size; i++ {
		idx := i
		if l.full {
			idx = (l.next + i) % l.cap
		}
		result = append(result, l.entries[idx])
	}
	for i := len(result)/2 - 1; i >= 0; i-- {
		j := len(result) - 1 - i
		result[i], result[j] = result[j], result[i]
	}
	return result
}

func (h *Handler) listAuditLog(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := chi.URLParam(r, "limit"); v != "" {
		var n int
		if err := json.Unmarshal([]byte(v), &n); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	entries := globalAudit.List(limit)
	writeJSON(w, http.StatusOK, map[string]any{"entries": entries, "total": len(entries)})
}
