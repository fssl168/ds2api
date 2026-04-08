package limits

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

func EstimateTokens(text string, charsPerToken float64) int {
	if text == "" {
		return 0
	}
	charCount := utf8.RuneCountInString(text)
	if charsPerToken <= 0 {
		charsPerToken = 3.0
	}
	return int(math.Ceil(float64(charCount) / charsPerToken))
}

func EstimateMessagesTokens(messages []map[string]any, charsPerToken float64) int {
	total := 0
	for _, msg := range messages {
		content, _ := msg["content"].(string)
		total += EstimateTokens(content, charsPerToken)
		role, _ := msg["role"].(string)
		total += EstimateTokens(role, charsPerToken)
		total += 4
	}
	return total + 2
}

type ProtectionResult struct {
	Truncated      bool
	InputTokens    int
	OutputTokens   int
	ContextUsage   float64
	Warnings       []string
	ClampedMaxTok  int
	Messages       []map[string]any
}

func CheckAndProtect(model string, messages []map[string]any, maxTokensReq int) ProtectionResult {
	lim := Get(model)
	result := ProtectionResult{
		Messages:   messages,
		Warnings:  make([]string, 0),
		OutputTokens: maxTokensReq,
	}

	nMsgs := len(messages)
	if nMsgs > lim.MaxMessages {
		result.Truncated = true
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("message count %d exceeds limit %d, truncating older messages", nMsgs, lim.MaxMessages))
		keep := nMsgs - lim.MaxMessages
		result.Messages = messages[keep:]
		messages = result.Messages
	}

	for i, msg := range messages {
		content, _ := msg["content"].(string)
		contentLen := len(content)
		if contentLen > lim.MaxSingleMsgSize {
			result.Truncated = true
			runes := []rune(content)
			maxRunes := int(float64(lim.MaxSingleMsgSize) / lim.CharsPerToken)
			if maxRunes < 100 {
				maxRunes = 100
			}
			if len(runes) > maxRunes {
				truncatedContent := string(runes[:maxRunes]) + "\n\n[...truncated by context protection]"
				msg["content"] = truncatedContent
				result.Messages[i] = msg
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("message[%d] content %d chars exceeded limit %d, truncated", i, contentLen, lim.MaxSingleMsgSize))
			}
		}
	}

	inputTokens := EstimateMessagesTokens(messages, lim.CharsPerToken)
	result.InputTokens = inputTokens

	availableForOutput := lim.ContextWindow - inputTokens
	if availableForOutput < 256 {
		availableForOutput = 256
	}

	if maxTokensReq > 0 && maxTokensReq > lim.MaxOutputTokens {
		result.ClampedMaxTok = lim.MaxOutputTokens
		result.OutputTokens = lim.MaxOutputTokens
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("max_tokens %d clamped to model max %d", maxTokensReq, lim.MaxOutputTokens))
	} else if maxTokensReq > 0 && maxTokensReq > availableForOutput {
		result.ClampedMaxTok = availableForOutput
		result.OutputTokens = availableForOutput
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("max_tokens %d reduced to available context %d (input=%d, window=%d)",
				maxTokensReq, availableForOutput, inputTokens, lim.ContextWindow))
	} else if maxTokensReq <= 0 {
		defaultOut := lim.MaxOutputTokens
		if defaultOut > availableForOutput {
			defaultOut = availableForOutput
		}
		result.OutputTokens = defaultOut
	}

	result.ContextUsage = float64(inputTokens+result.OutputTokens) / float64(lim.ContextWindow) * 100
	if result.ContextUsage > 95 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("context usage %.1f%% is very high (input=%d tokens, window=%d)",
				result.ContextUsage, inputTokens, lim.ContextWindow))
	}

	return result
}

func SanitizeLog(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "...(truncated)"
}

func JoinWarnings(warnings []string) string {
	if len(warnings) == 0 {
		return ""
	}
	return "[context-protection] " + strings.Join(warnings, "; ")
}
