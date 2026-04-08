package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"ds2api/internal/config"
)

type ToolHandler func(ctx context.Context, arguments json.RawMessage) (*ToolResult, error)

type Server struct {
	mu         sync.RWMutex
	handlers   map[string]ToolHandler
	tools      []Tool
	store      *config.Store
	serverInfo ImplementationInfo
}

func NewServer(store *config.Store) *Server {
	s := &Server{
		handlers: make(map[string]ToolHandler),
		tools:    make([]Tool, 0),
		store:    store,
		serverInfo: ImplementationInfo{
			Name:    "ds2api-mcp",
			Version: "1.0.0",
		},
	}
	initStore(store)
	s.registerBuiltinTools()
	return s
}

func (s *Server) RegisterTool(name, description string, inputSchema json.RawMessage, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[name] = handler
	s.tools = append(s.tools, Tool{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
	})
}

func (s *Server) registerBuiltinTools() {
	s.RegisterTool("chat", "Send a chat completion request to ds2api (supports DeepSeek and Qwen models). Returns the model's response text.", json.RawMessage(`{"type":"object","properties":{"model":{"type":"string","description":"Model name (e.g., deepseek-chat, qwen-plus, qwen-max, deepseek-reasoner, qwen-coder, qwen-flash)"},"messages":{"type":"array","description":"Array of chat messages","items":{"type":"object","properties":{"role":{"type":"string","enum":["user","assistant","system"]},"content":{"type":"string"}},"required":["role","content"]}},"stream":{"type":"boolean","description":"Whether to stream the response (default: false)"},"temperature":{"type":"number","description":"Sampling temperature (0-2, default: 1)"},"max_tokens":{"type":"integer","description":"Maximum tokens to generate"}},"required":["model","messages"]}`), handleChat)

	s.RegisterTool("list_models", "List all available models supported by ds2api, including DeepSeek and Qwen (通义千问) models.", json.RawMessage(`{"type":"object","additionalProperties":false}`), handleListModels)

	s.RegisterTool("get_status", "Get ds2api service status including health check and account pool information.", json.RawMessage(`{"type":"object","additionalProperties":false}`), handleGetStatus)

	s.RegisterTool("get_pool_status", "Get detailed account pool status for both DeepSeek and Qwen pools (in-flight counts, available accounts, cooldown state).", json.RawMessage(`{"type":"object","properties":{"pool_type":{"type":"string","enum":["deepseek","qwen","all"],"description":"Which pool to query (default: all)"}},"additionalProperties":false}`), handlePoolStatus)

	s.RegisterTool("embeddings", "Generate embeddings for text using ds2api.", json.RawMessage(`{"type":"object","properties":{"model":{"type":"string","description":"Embedding model (default: text-embedding-ada002)"},"input":{"type":"array","description":"Array of strings to embed","items":{"type":"string"}}},"required":["input"]}`), handleEmbeddings)
}

func (s *Server) HandleRequest(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(ctx, req.ID, req.Params)
	case "tools/list":
		return s.handleListTools(ctx, req.ID, req.Params)
	case "tools/call":
		return s.handleCallTool(ctx, req.ID, req.Params)
	case "ping":
		return s.handlePing(ctx, req.ID)
	default:
		return NewJSONRPCError(req.ID, MethodNotFound, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

func (s *Server) handleInitialize(_ context.Context, id interface{}, _ json.RawMessage) *JSONRPCResponse {
	result := InitializeResult{
		ProtocolVersion: "2025-03-26",
		Capabilities: ServerCapabilities{
			Tools: ToolServerCapabilities{ListChanged: false},
		},
		ServerInfo: s.serverInfo,
		Instructions: `ds2api MCP Bridge - Universal AI API Gateway

This MCP server exposes ds2api capabilities as tools:
- chat: Send messages to DeepSeek or Qwen (通义千问) models via OpenAI-compatible API
- list_models: Discover all available models (DeepSeek + Qwen)
- get_status: Check service health and configuration
- get_pool_status: Monitor account pool utilization
- embeddings: Generate text embeddings

Supported model prefixes:
- deepseek-* → DeepSeek engine (chat, reasoner, search variants)
- qwen-* / qwen3.* → 通义千问 engine (plus, max, coder, flash, 3.5 series)`,
	}
	resp, _ := NewJSONRPCResponse(id, result)
	return resp
}

func (s *Server) handleListTools(_ context.Context, id interface{}, params json.RawMessage) *JSONRPCResponse {
	var p ListToolsParams
	if len(params) > 0 && string(params) != "{}" && string(params) != "null" {
		json.Unmarshal(params, &p)
	}
	s.mu.RLock()
	tools := make([]Tool, len(s.tools))
	copy(tools, s.tools)
	s.mu.RUnlock()
	result := ListToolsResult{Tools: tools}
	resp, _ := NewJSONRPCResponse(id, result)
	return resp
}

func (s *Server) handleCallTool(ctx context.Context, id interface{}, params json.RawMessage) *JSONRPCResponse {
	var p CallToolParams
	if err := json.Unmarshal(params, &p); err != nil {
		return NewJSONRPCError(id, InvalidParams, fmt.Sprintf("Invalid params: %v", err))
	}
	s.mu.RLock()
	handler, ok := s.handlers[p.Name]
	s.mu.RUnlock()
	if !ok {
		return NewJSONRPCError(id, MethodNotFound, fmt.Sprintf("Unknown tool: %s", p.Name))
	}
	result, err := handler(ctx, p.Arguments)
	if err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: mustMarshal(ToolResult{
				Content: []TextContent{{Type: "text", Text: err.Error()}},
				IsError: true,
			}),
		}
	}
	resp, _ := NewJSONRPCResponse(id, result)
	return resp
}

func (s *Server) handlePing(_ context.Context, id interface{}) *JSONRPCResponse {
	resp, _ := NewJSONRPCResponse(id, PingResult{})
	return resp
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, MCP-Session-ID")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}
	var req JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		config.Logger.Error("[MCP] invalid request", "error", err)
		resp := NewJSONRPCError(nil, ParseError, "Parse error")
		data, _ := json.Marshal(resp)
		w.Write(data)
		return
	}
	resp := s.HandleRequest(r.Context(), &req)
	data, _ := json.Marshal(resp)
	w.Write(data)
}

func mustMarshal(v interface{}) json.RawMessage {
	d, _ := json.Marshal(v)
	return d
}
