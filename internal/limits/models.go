package limits

import (
	"os"
	"strings"
)

var Enabled bool

func init() {
	v := strings.TrimSpace(os.Getenv("CONTEXT_PROTECTION"))
	Enabled = v == "true" || v == "1" || v == "on"
	if Enabled {
		println("[limits] CONTEXT_PROTECTION enabled")
	}
}

type ModelLimits struct {
	ContextWindow    int
	MaxOutputTokens  int
	CharsPerToken    float64
	MaxMessages      int
	MaxSingleMsgSize int
}

func Get(model string) ModelLimits {
	m := strings.ToLower(strings.TrimSpace(model))
	switch {
	case strings.Contains(m, "qwen3.5-plus") || strings.Contains(m, "qwen-3.5-plus"):
		return ModelLimits{ContextWindow: 131072, MaxOutputTokens: 8192, CharsPerToken: 2.0, MaxMessages: 100, MaxSingleMsgSize: 65536}
	case strings.Contains(m, "qwen3.5-flash") || strings.Contains(m, "qwen-3.5-flash"):
		return ModelLimits{ContextWindow: 131072, MaxOutputTokens: 8192, CharsPerToken: 2.0, MaxMessages: 100, MaxSingleMsgSize: 65536}
	case strings.Contains(m, "qwen-max"):
		return ModelLimits{ContextWindow: 32768, MaxOutputTokens: 2048, CharsPerToken: 1.8, MaxMessages: 80, MaxSingleMsgSize: 32000}
	case strings.Contains(m, "qwen-coder"):
		return ModelLimits{ContextWindow: 32768, MaxOutputTokens: 4096, CharsPerToken: 1.8, MaxMessages: 80, MaxSingleMsgSize: 32000}
	case strings.Contains(m, "qwen-flash"):
		return ModelLimits{ContextWindow: 128000, MaxOutputTokens: 2048, CharsPerToken: 2.0, MaxMessages: 100, MaxSingleMsgSize: 65536}
	case strings.Contains(m, "qwen-plus"):
		return ModelLimits{ContextWindow: 32768, MaxOutputTokens: 2048, CharsPerToken: 1.8, MaxMessages: 80, MaxSingleMsgSize: 32000}
	case strings.HasPrefix(m, "deepseek-reasoner"):
		return ModelLimits{ContextWindow: 65536, MaxOutputTokens: 8192, CharsPerToken: 3.0, MaxMessages: 100, MaxSingleMsgSize: 48000}
	default:
		return ModelLimits{ContextWindow: 65536, MaxOutputTokens: 8192, CharsPerToken: 3.0, MaxMessages: 100, MaxSingleMsgSize: 48000}
	}
}
