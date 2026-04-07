package admin

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ds2api/internal/config"
)

func (h *Handler) configImport(w http.ResponseWriter, r *http.Request) {
	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": "invalid json"})
		return
	}

	mode := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("mode")))
	if mode == "" {
		mode = strings.TrimSpace(strings.ToLower(fieldString(req, "mode")))
	}
	if mode == "" {
		mode = "merge"
	}
	if mode != "merge" && mode != "replace" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": "mode must be merge or replace"})
		return
	}

	payload := req
	if raw, ok := req["config"].(map[string]any); ok && len(raw) > 0 {
		payload = raw
	}
	rawJSON, err := json.Marshal(payload)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": "invalid config payload"})
		return
	}
	var incoming config.Config
	if err := json.Unmarshal(rawJSON, &incoming); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": err.Error()})
		return
	}
	incoming.ClearAccountTokens()

	importedKeys, importedAccounts, importedQwenAccounts := 0, 0, 0
	err = h.Store.Update(func(c *config.Config) error {
		next := c.Clone()
		if mode == "replace" {
			next = incoming.Clone()
			next.Accounts = normalizeAndDedupeAccounts(next.Accounts)
			next.QwenAccounts = normalizeAndDedupeQwenAccounts(next.QwenAccounts)
			next.VercelSyncHash = c.VercelSyncHash
			next.VercelSyncTime = c.VercelSyncTime
			importedKeys = len(next.Keys)
			importedAccounts = len(next.Accounts)
			importedQwenAccounts = len(next.QwenAccounts)
		} else {
			existingKeys := map[string]struct{}{}
			for _, k := range next.Keys {
				existingKeys[k] = struct{}{}
			}
			for _, k := range incoming.Keys {
				key := strings.TrimSpace(k)
				if key == "" {
					continue
				}
				if _, ok := existingKeys[key]; ok {
					continue
				}
				existingKeys[key] = struct{}{}
				next.Keys = append(next.Keys, key)
				importedKeys++
			}

			existingAccounts := map[string]struct{}{}
			for _, acc := range next.Accounts {
				acc = normalizeAccountForStorage(acc)
				key := accountDedupeKey(acc)
				if key != "" {
					existingAccounts[key] = struct{}{}
				}
			}
			for _, acc := range incoming.Accounts {
				acc = normalizeAccountForStorage(acc)
				key := accountDedupeKey(acc)
				if key == "" {
					continue
				}
				if _, ok := existingAccounts[key]; ok {
					continue
				}
				existingAccounts[key] = struct{}{}
				next.Accounts = append(next.Accounts, acc)
				importedAccounts++
			}

			if len(incoming.ClaudeMapping) > 0 {
				if next.ClaudeMapping == nil {
					next.ClaudeMapping = map[string]string{}
				}
				for k, v := range incoming.ClaudeMapping {
					next.ClaudeMapping[k] = v
				}
			}
			if len(incoming.ClaudeModelMap) > 0 {
				if next.ClaudeModelMap == nil {
					next.ClaudeModelMap = map[string]string{}
				}
				for k, v := range incoming.ClaudeModelMap {
					next.ClaudeModelMap[k] = v
				}
			}

			if len(incoming.ModelAliases) > 0 {
				if next.ModelAliases == nil {
					next.ModelAliases = map[string]string{}
				}
				for k, v := range incoming.ModelAliases {
					next.ModelAliases[k] = v
				}
			}

			existingQwenAccounts := map[string]struct{}{}
			for _, qa := range next.QwenAccounts {
				key := qwenAccountDedupeKey(qa)
				if key != "" {
					existingQwenAccounts[key] = struct{}{}
				}
			}
			for _, qa := range incoming.QwenAccounts {
				qa = normalizeQwenAccountForStorage(qa)
				key := qwenAccountDedupeKey(qa)
				if key == "" {
					continue
				}
				if _, ok := existingQwenAccounts[key]; ok {
					continue
				}
				existingQwenAccounts[key] = struct{}{}
				next.QwenAccounts = append(next.QwenAccounts, qa)
				importedQwenAccounts++
			}
			if incoming.Responses.StoreTTLSeconds > 0 {
				next.Responses.StoreTTLSeconds = incoming.Responses.StoreTTLSeconds
			}
			if strings.TrimSpace(incoming.Embeddings.Provider) != "" {
				next.Embeddings.Provider = incoming.Embeddings.Provider
			}
			if strings.TrimSpace(incoming.Admin.PasswordHash) != "" {
				next.Admin.PasswordHash = incoming.Admin.PasswordHash
			}
			if incoming.Admin.JWTExpireHours > 0 {
				next.Admin.JWTExpireHours = incoming.Admin.JWTExpireHours
			}
			if incoming.Admin.JWTValidAfterUnix > 0 {
				next.Admin.JWTValidAfterUnix = incoming.Admin.JWTValidAfterUnix
			}
			if incoming.Runtime.AccountMaxInflight > 0 {
				next.Runtime.AccountMaxInflight = incoming.Runtime.AccountMaxInflight
			}
			if incoming.Runtime.AccountMaxQueue > 0 {
				next.Runtime.AccountMaxQueue = incoming.Runtime.AccountMaxQueue
			}
			if incoming.Runtime.GlobalMaxInflight > 0 {
				next.Runtime.GlobalMaxInflight = incoming.Runtime.GlobalMaxInflight
			}
			if incoming.Runtime.TokenRefreshIntervalHours > 0 {
				next.Runtime.TokenRefreshIntervalHours = incoming.Runtime.TokenRefreshIntervalHours
			}
		}

		normalizeSettingsConfig(&next)
		if err := validateSettingsConfig(next); err != nil {
			return newRequestError(err.Error())
		}

		*c = next
		return nil
	})
	if err != nil {
		if detail, ok := requestErrorDetail(err); ok {
			writeJSON(w, http.StatusBadRequest, map[string]any{"detail": detail})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"detail": err.Error()})
		return
	}

	h.Pool.Reset()
	if h.QW != nil {
		h.QW.ResetTickets()
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success":                true,
		"mode":                   mode,
		"imported_keys":          importedKeys,
		"imported_accounts":      importedAccounts,
		"imported_qwen_accounts": importedQwenAccounts,
		"message":                "config imported",
	})
}

func (h *Handler) computeSyncHash() string {
	snap := h.Store.Snapshot().Clone()
	snap.ClearAccountTokens()
	snap.VercelSyncHash = ""
	snap.VercelSyncTime = 0
	b, _ := json.Marshal(snap)
	sum := md5.Sum(b)
	return fmt.Sprintf("%x", sum)
}

func normalizeAndDedupeQwenAccounts(accounts []config.QwenAccount) []config.QwenAccount {
	seen := map[string]struct{}{}
	result := make([]config.QwenAccount, 0, len(accounts))
	for _, qa := range accounts {
		qa = normalizeQwenAccountForStorage(qa)
		key := qwenAccountDedupeKey(qa)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, qa)
	}
	return result
}

func normalizeQwenAccountForStorage(qa config.QwenAccount) config.QwenAccount {
	qa.Ticket = strings.TrimSpace(qa.Ticket)
	if qa.Label == "" {
		if qa.Ticket != "" {
			qa.Label = "qwen-" + qa.Ticket[:min(8, len(qa.Ticket))]
		} else {
			qa.Label = fmt.Sprintf("qwen-%d", time.Now().UnixNano()%10000)
		}
	} else {
		qa.Label = strings.TrimSpace(qa.Label)
	}
	return qa
}

func qwenAccountDedupeKey(qa config.QwenAccount) string {
	ticket := strings.TrimSpace(qa.Ticket)
	label := strings.TrimSpace(qa.Label)
	if ticket == "" && label == "" {
		return ""
	}
	if ticket != "" {
		return "ticket:" + ticket
	}
	return "label:" + label
}
