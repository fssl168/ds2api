package qwen

import (
	"fmt"
	"strings"
)

type L2Message struct {
	Content   string `json:"content"`
	MimeType  string `json:"mime_type"`
	MetaData  *L2MetaData `json:"meta_data,omitempty"`
}

type L2MetaData struct {
	OriQuery string `json:"ori_query"`
}

func BuildL2Messages(messages []map[string]any) []L2Message {
	if len(messages) == 0 {
		return nil
	}
	lastMsg := messages[len(messages)-1]
	content := extractTextFromMap(lastMsg)
	return []L2Message{
		{
			Content:  content,
			MimeType: "text/plain",
			MetaData: &L2MetaData{OriQuery: content},
		},
	}
}

func extractTextFromMap(msg map[string]any) string {
	val, ok := msg["content"]
	if !ok {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case []any:
		var parts []string
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				if t, ok := m["text"].(string); ok && (m["type"] == "text" || m["type"] == "") {
					parts = append(parts, t)
				} else if t, ok := m["content"].(string); ok {
					parts = append(parts, t)
				}
			}
		}
		return strings.Join(parts, " ")
	default:
		return fmt.Sprint(v)
	}
}

func resolveModelName(rawModel string) string {
	modelMap := map[string]string{
		"qwen/Qwen":              "Qwen",
		"qwen/qwen":              "Qwen",
		"qwen":                   "Qwen",
		"qwen/Qwen3-Max":         "Qwen3-Max",
		"qwen/qwen-max":          "Qwen3-Max",
		"qwen-max":               "Qwen3-Max",
		"qwen/Qwen3-Plus":        "Qwen3-Plus",
		"qwen/qwen-plus":         "Qwen3-Plus",
		"qwen-plus":              "Qwen3-Plus",
		"qwen/Qwen3-Coder":       "Qwen3-Coder",
		"qwen/qwen-coder":        "Qwen3-Coder",
		"qwen-coder":             "Qwen3-Coder",
		"qwen/Qwen3-Flash":       "Qwen3-Flash",
		"qwen/qwen-flash":        "Qwen3-Flash",
		"qwen-flash":             "Qwen3-Flash",
		"qwen/Qwen3.5-Plus":      "Qwen3.5-Plus",
		"qwen/qwen3.5-plus":      "Qwen3.5-Plus",
		"qwen/Qwen3.5-Flash":     "Qwen3.5-Flash",
		"qwen/qwen3.5-flash":     "Qwen3.5-Flash",
	}
	if m, ok := modelMap[rawModel]; ok {
		return m
	}
	return "Qwen"
}
