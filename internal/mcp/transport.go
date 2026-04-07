package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"ds2api/internal/config"
)

var sessionMu sync.Mutex
var sessions = make(map[string]*Server)

func RegisterSession(id string, srv *Server) {
	sessionMu.Lock()
	sessions[id] = srv
	sessionMu.Unlock()
}

func GetSession(id string) *Server {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	return sessions[id]
}

func StreamableHTTPHandler(srv *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			initResp := InitializeResult{
				ProtocolVersion: "2025-03-26",
				Capabilities: ServerCapabilities{
					Tools: ToolServerCapabilities{ListChanged: false},
				},
				ServerInfo: srv.serverInfo,
				Instructions: "ds2api MCP Bridge - Use POST for JSON-RPC requests",
			}
			json.NewEncoder(w).Encode(initResp)
			return
		}
		srv.ServeHTTP(w, r)
	}
}

func StdioTransport(ctx context.Context, srv *Server) error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			sendStdio(NewJSONRPCError(nil, ParseError, "Parse error"))
			continue
		}
		resp := srv.HandleRequest(ctx, &req)
		sendStdio(resp)
	}
	return scanner.Err()
}

func sendStdio(resp *JSONRPCResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("[MCP stdio] marshal error: %v", err)
		return
	}
	fmt.Printf("%s\n", data)
}

func SSETransportHandler(srv *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("MCP-Session-ID")
		if sessionID == "" {
			sessionID = newSessionID()
			RegisterSession(sessionID, srv)
		} else if existing := GetSession(sessionID); existing == nil {
			RegisterSession(sessionID, srv)
		}

		if r.Method == http.MethodDelete || r.URL.Path == "/sse/endpoint" && r.Method == http.MethodDelete {
			deleteSession(sessionID)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Header.Get("Accept") == "text/event-stream" || strings.Contains(r.Header.Get("Accept"), "event-stream") {
			handleSSEConnection(w, r, srv, sessionID)
			return
		}

		body, _ := io.ReadAll(r.Body)
		var batchReq []JSONRPCRequest
		if err := json.Unmarshal(body, &batchReq); err == nil && len(batchReq) > 0 {
			handleBatchRPC(w, srv, batchReq)
			return
		}
		var req JSONRPCRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Invalid JSON-RPC", http.StatusBadRequest)
			return
		}
		resp := srv.HandleRequest(r.Context(), &req)
		w.Header().Set("Content-Type", "application/json")
		data, _ := json.Marshal(resp)
		w.Write(data)
	}
}

func handleSSEConnection(w http.ResponseWriter, r *http.Request, srv *Server, sessionID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("MCP-Session-ID", sessionID)
	w.Header().Set("Access-Control-Expose-Headers", "MCP-Session-ID")

	initData, _ := json.Marshal(InitializeResult{
		ProtocolVersion: "2025-03-26",
		Capabilities: ServerCapabilities{Tools: ToolServerCapabilities{ListChanged: false}},
		ServerInfo: srv.serverInfo,
	})
	fmt.Fprintf(w, "event: endpoint\ndata: /mcp?session_id=%s\n\n", sessionID)
	fmt.Fprintf(w, "data: %s\n\n", initData)
	flusher.Flush()

	ctx := r.Context()
	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		n, err := r.Body.Read(buf)
		if n > 0 {
			var req JSONRPCRequest
			if json.Unmarshal(buf[:n], &req) == nil {
				resp := srv.HandleRequest(ctx, &req)
				respData, _ := json.Marshal(resp)
				fmt.Fprintf(w, "data: %s\n\n", respData)
				flusher.Flush()
			}
		}
		if err != nil {
			if err != io.EOF {
				config.Logger.Warn("[MCP SSE] read error", "error", err)
			}
			return
		}
	}
}

func handleBatchRPC(w http.ResponseWriter, srv *Server, reqs []JSONRPCRequest) {
	resps := make([]*JSONRPCResponse, len(reqs))
	for i, req := range reqs {
		resps[i] = srv.HandleRequest(context.Background(), &req)
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	for _, resp := range resps {
		enc.Encode(resp)
	}
}

func deleteSession(id string) {
	sessionMu.Lock()
	delete(sessions, id)
	sessionMu.Unlock()
}

func newSessionID() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = "abcdefghijklmnopqrstuvwxyz0123456789"[i%36]
	}
	return string(b)
}

func ReadSSEMessages(r io.Reader) ([]json.RawMessage, error) {
	var messages []json.RawMessage
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 512*1024)
	var buf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if buf.Len() > 0 {
				messages = append(messages, json.RawMessage(buf.String()))
				buf.Reset()
			}
			continue
		}
		if strings.HasPrefix(line, "data: ") {
			buf.WriteString(strings.TrimPrefix(line, "data: "))
		}
	}
	if buf.Len() > 0 {
		messages = append(messages, json.RawMessage(buf.String()))
	}
	return messages, scanner.Err()
}
