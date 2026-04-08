package config

import "strings"

type ModelInfo struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Created    int64  `json:"created"`
	OwnedBy    string `json:"owned_by"`
	Permission []any  `json:"permission,omitempty"`
}

type ModelAliasReader interface {
	ModelAliases() map[string]string
}

var DeepSeekModels = []ModelInfo{
	{ID: "deepseek-chat", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-reasoner", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-chat-search", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-reasoner-search", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-expert-chat", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-expert-reasoner", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-expert-chat-search", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-expert-reasoner-search", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-vision-chat", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-vision-reasoner", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-vision-chat-search", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
	{ID: "deepseek-vision-reasoner-search", Object: "model", Created: 1677610602, OwnedBy: "deepseek", Permission: []any{}},
}

var QwenModels = []ModelInfo{
	// 千问 Max 系列
	{ID: "qwen3-max", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-max", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-max-latest", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},

	// 千问 Plus 系列
	{ID: "qwen3.5-plus", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-plus", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-plus-latest", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},

	// 千问 Flash 系列
	{ID: "qwen3.5-flash", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-flash", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},

	// 千问 Turbo 系列
	{ID: "qwen-turbo", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-turbo-latest", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},

	// 千问 Long 系列
	{ID: "qwen-long", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-long-latest", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},

	// QwQ 推理系列
	{ID: "qwq-plus", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwq-plus-latest", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},

	// Qwen-VL 视觉系列
	{ID: "qwen-vl-max", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-vl-max-latest", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-vl-plus", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-vl-plus-latest", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},

	// Qwen-Audio 语音系列
	{ID: "qwen-audio-turbo", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-audio-turbo-latest", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},

	// Qwen-Omni 多模态系列
	{ID: "qwen-omni-turbo", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-omni-turbo-latest", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-omni-turbo-realtime", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-omni-turbo-realtime-latest", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},

	// Qwen-MT 翻译系列
	{ID: "qwen-mt-plus", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-mt-plus-latest", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-mt-turbo", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-mt-turbo-latest", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},

	// Qwen-Coder 代码生成系列
	{ID: "qwen3-coder-next", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen3-coder-plus", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen3-coder-flash", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen3-coder-480B", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-coder-turbo", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-coder-turbo-latest", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen2.5-coder-7b-instruct", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen2.5-coder-14b-instruct", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen2.5-coder-32b-instruct", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},

	// Qwen 其他系列
	{ID: "qwen3.6", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen3-max-thinking", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-image-2.0", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
	{ID: "qwen-image-2.0-2026-03-03", Object: "model", Created: 1677610602, OwnedBy: "qwen", Permission: []any{}},
}

var ClaudeModels = []ModelInfo{
	// Current aliases
	{ID: "claude-opus-4-6", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-sonnet-4-5", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-haiku-4-5", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},

	// Current snapshots
	{ID: "claude-opus-4-5-20251101", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-opus-4-1", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-opus-4-1-20250805", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-opus-4-0", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-opus-4-20250514", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-sonnet-4-5-20250929", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-sonnet-4-0", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-sonnet-4-20250514", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-haiku-4-5-20251001", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},

	// Claude 3.x (legacy/deprecated snapshots and aliases)
	{ID: "claude-3-7-sonnet-latest", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-7-sonnet-20250219", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-5-sonnet-latest", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-5-sonnet-20240620", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-5-sonnet-20241022", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-opus-20240229", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-sonnet-20240229", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-5-haiku-latest", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-5-haiku-20241022", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-3-haiku-20240307", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},

	// Claude 2.x and 1.x (retired but accepted for compatibility)
	{ID: "claude-2.1", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-2.0", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-1.3", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-1.2", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-1.1", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-1.0", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-instant-1.2", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-instant-1.1", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
	{ID: "claude-instant-1.0", Object: "model", Created: 1715635200, OwnedBy: "anthropic"},
}

func GetModelConfig(model string) (thinking bool, search bool, ok bool) {
	switch lower(model) {
	case "deepseek-chat":
		return false, false, true
	case "deepseek-reasoner":
		return true, false, true
	case "deepseek-chat-search":
		return false, true, true
	case "deepseek-reasoner-search":
		return true, true, true
	case "deepseek-expert-chat":
		return false, false, true
	case "deepseek-expert-reasoner":
		return true, false, true
	case "deepseek-expert-chat-search":
		return false, true, true
	case "deepseek-expert-reasoner-search":
		return true, true, true
	case "deepseek-vision-chat":
		return false, false, true
	case "deepseek-vision-reasoner":
		return true, false, true
	case "deepseek-vision-chat-search":
		return false, true, true
	case "deepseek-vision-reasoner-search":
		return true, true, true
	default:
		return false, false, false
	}
}

func IsSupportedDeepSeekModel(model string) bool {
	_, _, ok := GetModelConfig(model)
	return ok
}

func DefaultModelAliases() map[string]string {
	return map[string]string{
		"gpt-4o":                 "deepseek-chat",
		"gpt-4.1":                "deepseek-chat",
		"gpt-4.1-mini":           "deepseek-chat",
		"gpt-4.1-nano":           "deepseek-chat",
		"gpt-5":                  "deepseek-chat",
		"gpt-5-mini":             "deepseek-chat",
		"gpt-5-codex":            "deepseek-reasoner",
		"o1":                     "deepseek-reasoner",
		"o1-mini":                "deepseek-reasoner",
		"o3":                     "deepseek-reasoner",
		"o3-mini":                "deepseek-reasoner",
		"claude-sonnet-4-5":      "deepseek-chat",
		"claude-haiku-4-5":       "deepseek-chat",
		"claude-opus-4-6":        "deepseek-reasoner",
		"claude-3-5-sonnet":      "deepseek-chat",
		"claude-3-5-haiku":       "deepseek-chat",
		"claude-3-opus":          "deepseek-reasoner",
		"gemini-2.5-pro":         "deepseek-chat",
		"gemini-2.5-flash":       "deepseek-chat",
		"llama-3.1-70b-instruct": "deepseek-chat",
		"qwen-max":               "deepseek-chat",
	}
}

func ResolveModel(store ModelAliasReader, requested string) (string, bool) {
	model := lower(strings.TrimSpace(requested))
	if model == "" {
		return "", false
	}
	if IsSupportedDeepSeekModel(model) {
		return model, true
	}
	aliases := DefaultModelAliases()
	if store != nil {
		for k, v := range store.ModelAliases() {
			aliases[lower(strings.TrimSpace(k))] = lower(strings.TrimSpace(v))
		}
	}
	if mapped, ok := aliases[model]; ok && IsSupportedDeepSeekModel(mapped) {
		return mapped, true
	}
	if strings.HasPrefix(model, "deepseek-") {
		return "", false
	}

	knownFamily := false
	for _, prefix := range []string{
		"gpt-", "o1", "o3", "claude-", "gemini-", "llama-", "qwen-", "mistral-", "command-",
	} {
		if strings.HasPrefix(model, prefix) {
			knownFamily = true
			break
		}
	}
	if !knownFamily {
		return "", false
	}

	useReasoner := strings.Contains(model, "reason") ||
		strings.Contains(model, "reasoner") ||
		strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3") ||
		strings.Contains(model, "opus") ||
		strings.Contains(model, "r1")
	useSearch := strings.Contains(model, "search")
	useExpert := strings.Contains(model, "expert")
	useVision := strings.Contains(model, "vision")

	switch {
	case useVision && useReasoner && useSearch:
		return "deepseek-vision-reasoner-search", true
	case useVision && useReasoner:
		return "deepseek-vision-reasoner", true
	case useVision && useSearch:
		return "deepseek-vision-chat-search", true
	case useVision:
		return "deepseek-vision-chat", true
	case useExpert && useReasoner && useSearch:
		return "deepseek-expert-reasoner-search", true
	case useExpert && useReasoner:
		return "deepseek-expert-reasoner", true
	case useExpert && useSearch:
		return "deepseek-expert-chat-search", true
	case useExpert:
		return "deepseek-expert-chat", true
	case useReasoner && useSearch:
		return "deepseek-reasoner-search", true
	case useReasoner:
		return "deepseek-reasoner", true
	case useSearch:
		return "deepseek-chat-search", true
	default:
		return "deepseek-chat", true
	}
}

func lower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func OpenAIModelsResponse() map[string]any {
	data := make([]ModelInfo, 0, len(DeepSeekModels)+len(QwenModels))
	data = append(data, DeepSeekModels...)
	data = append(data, QwenModels...)
	return map[string]any{"object": "list", "data": data}
}

func OpenAIModelByID(store ModelAliasReader, id string) (ModelInfo, bool) {
	canonical, ok := ResolveModel(store, id)
	if !ok {
		return ModelInfo{}, false
	}
	// Check DeepSeek models
	for _, model := range DeepSeekModels {
		if model.ID == canonical {
			return model, true
		}
	}
	// Check Qwen models
	for _, model := range QwenModels {
		if model.ID == canonical {
			return model, true
		}
	}
	return ModelInfo{}, false
}

func ClaudeModelsResponse() map[string]any {
	resp := map[string]any{"object": "list", "data": ClaudeModels}
	if len(ClaudeModels) > 0 {
		resp["first_id"] = ClaudeModels[0].ID
		resp["last_id"] = ClaudeModels[len(ClaudeModels)-1].ID
	} else {
		resp["first_id"] = nil
		resp["last_id"] = nil
	}
	resp["has_more"] = false
	return resp
}
