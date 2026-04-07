package mcp

import (
	"context"
	"fmt"
	"os"

	"ds2api/internal/config"
	"ds2api/internal/deepseek"
	"ds2api/internal/qwen"
)

type PluginRegistry struct {
	server      *Server
	store       *config.Store
	dsClient    *deepseek.Client
	qwenClient  *qwen.Client
	platforms   map[string]PlatformConfig
}

type PlatformConfig struct {
	Name        string
	Description string
	Enabled     bool
	Guide       func(baseURL string) string
}

func NewPluginRegistry(store *config.Store, ds *deepseek.Client, qw *qwen.Client) *PluginRegistry {
	srv := NewServer(store)

	reg := &PluginRegistry{
		server:     srv,
		store:      store,
		dsClient:   ds,
		qwenClient: qw,
		platforms:  make(map[string]PlatformConfig),
	}
	return reg
}

func (r *PluginRegistry) Server() *Server {
	return r.server
}

func (r *PluginRegistry) RegisterPlatform(name, description string, enabled bool, guideFn func(baseURL string) string) {
	r.platforms[name] = PlatformConfig{
		Name:        name,
		Description: description,
		Enabled:     enabled,
		Guide:       guideFn,
	}
}

func (r *PluginRegistry) Platforms() map[string]PlatformConfig {
	return r.platforms
}

func (r *PluginRegistry) GetConnectionGuide(platformName, baseURL string) (string, error) {
	p, ok := r.platforms[platformName]
	if !ok {
		return "", fmt.Errorf("unknown platform: %s", platformName)
	}
	if p.Guide == nil {
		return "", fmt.Errorf("no connection guide for platform: %s", platformName)
	}
	return p.Guide(baseURL), nil
}

func (r *PluginRegistry) AllGuides(baseURL string) map[string]string {
	guides := make(map[string]string)
	for name, p := range r.platforms {
		if p.Guide != nil && p.Enabled {
			guides[name] = p.Guide(baseURL)
		}
	}
	return guides
}

func (r *PluginRegistry) RunStdioMode(ctx context.Context) error {
	config.Logger.Info("[MCP] starting stdio transport mode")
	fmt.Fprintln(os.Stderr, "[MCP] ds2api MCP Bridge running in stdio mode")
	fmt.Fprintln(os.Stderr, "[MCP] Press Ctrl+C to stop")
	return StdioTransport(ctx, r.server)
}

func IsMCPMode() bool {
	mode := os.Getenv("DS2API_MCP_MODE")
	return mode == "stdio" || mode == "streamable" || mode == "sse"
}

func MCPTransport() string {
	t := os.Getenv("DS2API_MCP_TRANSPORT")
	if t == "" {
		if IsMCPMode() {
			return "stdio"
		}
		return "streamable_http"
	}
	return t
}
