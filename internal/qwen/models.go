package qwen

import "time"

var QwenModels = []ModelInfo{
	{ID: "qwen/Qwen", Object: "model", Created: time.Unix(1700000000, 0), OwnedBy: "qwen"},
	{ID: "qwen/Qwen3-Max", Object: "model", Created: time.Unix(1700000000, 0), OwnedBy: "qwen"},
	{ID: "qwen/Qwen3-Max-Thinking", Object: "model", Created: time.Unix(1700000000, 0), OwnedBy: "qwen"},
	{ID: "qwen/Qwen3-Plus", Object: "model", Created: time.Unix(1700000000, 0), OwnedBy: "qwen"},
	{ID: "qwen/Qwen3-Coder", Object: "model", Created: time.Unix(1700000000, 0), OwnedBy: "qwen"},
	{ID: "qwen/Qwen3-Flash", Object: "model", Created: time.Unix(1700000000, 0), OwnedBy: "qwen"},
	{ID: "qwen/Qwen3.5-Plus", Object: "model", Created: time.Unix(1700000000, 0), OwnedBy: "qwen"},
	{ID: "qwen/Qwen3.5-Flash", Object: "model", Created: time.Unix(1700000000, 0), OwnedBy: "qwen"},
	{ID: "qwen/Qwen3.6-Plus", Object: "model", Created: time.Unix(1700000000, 0), OwnedBy: "qwen"},
	{ID: "qwen/Qwen3.6-Plus-2026-04-02", Object: "model", Created: time.Unix(1700000000, 0), OwnedBy: "qwen"},
}

type ModelInfo struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created time.Time `json:"created"`
	OwnedBy string    `json:"owned_by"`
}
