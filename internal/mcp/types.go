package mcp

import "encoding/json"

type JSONRPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      interface{}    `json:"id,omitempty"`
	Method  string         `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func NewJSONRPCResponse(id interface{}, result interface{}) (*JSONRPCResponse, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  data,
	}, nil
}

func NewJSONRPCError(id interface{}, code int, msg string) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: msg,
		},
	}
}

const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ClientCapabilities     `json:"capabilities"`
	ClientInfo      ImplementationInfo     `json:"clientInfo"`
}

type ClientCapabilities struct {
	Roots map[string]any `json:"roots,omitempty"`
	Sampling any        `json:"sampling,omitempty"`
}

type ServerCapabilities struct {
	Tools ToolServerCapabilities `json:"tools"`
}

type ToolServerCapabilities struct {
	ListChanged bool `json:"listChanged"`
}

type ImplementationInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string              `json:"protocolVersion"`
	Capabilities    ServerCapabilities  `json:"capabilities"`
	ServerInfo      ImplementationInfo  `json:"serverInfo"`
	Instructions    string              `json:"instructions"`
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
	Annotations *ToolAnnotations `json:"annotations,omitempty"`
}

type ToolAnnotations struct {
	ReadOnlyHint    bool `json:"readOnlyHint,omitempty"`
	DestructiveHint bool `json:"destructiveHint,omitempty"`
	IdempotentHint  bool `json:"idempotentHint,omitempty"`
	OpenWorldHint   bool `json:"openWorldHint,omitempty"`
}

type CallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
	Meta      *RequestMeta    `json:"_meta,omitempty"`
}

type RequestMeta struct {
	ProgressToken interface{} `json:"progressToken,omitempty"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ToolResult struct {
	Content           []TextContent `json:"content"`
	IsError           bool          `json:"isError,omitempty"`
	StructuredContent interface{}   `json:"structuredContent,omitempty"`
}

type ListToolsResult struct {
	Tools []Tool `json:"tools"`
	NextCursor string `json:"nextCursor,omitempty"`
}

type ListToolsParams struct {
	Cursor string `json:"cursor,omitempty"`
}

type PingResult struct{}
