package qwen

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"

	"ds2api/internal/config"
)

type qwenSSEScanner struct {
	scanner      *bufio.Scanner
	event        QwenStreamResult
	err          error
	done         bool
	completeSeen bool
}

type QwenStreamResult struct {
	ID   string
	Text string
	Done bool
	Err  error
}

func NewQwenSSEScanner(r io.Reader) *qwenSSEScanner {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	return &qwenSSEScanner{
		scanner: sc,
	}
}

func (s *qwenSSEScanner) Next() bool {
	if s.done || s.err != nil || s.completeSeen {
		return false
	}
	for s.scanner.Scan() {
		line := s.scanner.Bytes()
		s.event = parseL2Line(line)
		if s.event.Done && s.event.Text == "" {
			s.completeSeen = true
			return false
		}
		if s.event.Done {
			s.done = true
			return s.event.Text != ""
		}
		if s.event.Text != "" {
			return true
		}
	}
	s.err = s.scanner.Err()
	return false
}

func (s *qwenSSEScanner) Event() QwenStreamResult {
	return s.event
}

func (s *qwenSSEScanner) Err() error {
	return s.err
}

type l2Event struct {
	Data struct {
		Messages []l2Msg `json:"messages"`
		Status   string  `json:"status"`
	} `json:"data"`
	Success bool `json:"success"`
}

type l2Msg struct {
	Content  string `json:"content"`
	MimeType string `json:"mime_type"`
	Status   string `json:"status"`
}

func parseL2Line(line []byte) QwenStreamResult {
	str := strings.TrimSpace(string(line))
	if str == "" {
		return QwenStreamResult{}
	}

	if strings.HasPrefix(str, "event:") {
		eventType := strings.TrimSpace(strings.TrimPrefix(str, "event:"))
		if eventType == "complete" {
			return QwenStreamResult{Done: true}
		}
		return QwenStreamResult{}
	}

	if !strings.HasPrefix(str, "data:") {
		return QwenStreamResult{}
	}

	data := strings.TrimSpace(strings.TrimPrefix(str, "data:"))
	if data == "true" || data == "[DONE]" {
		return QwenStreamResult{Done: true}
	}

	var evt l2Event
	if err := json.Unmarshal([]byte(data), &evt); err != nil {
		config.Logger.Warn("[qwen-sse] parse error", "raw", truncateStr(data, 200), "error", err)
		return QwenStreamResult{}
	}

	text := ""
	for _, msg := range evt.Data.Messages {
		if msg.MimeType == "multi_load/iframe" || msg.MimeType == "text/plain" {
			text = msg.Content
		}
	}

	done := evt.Data.Status == "complete"

	return QwenStreamResult{
		Text: text,
		Done: done,
	}
}
