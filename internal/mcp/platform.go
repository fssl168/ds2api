package mcp

import (
	"fmt"
	"net/http"

	"ds2api/internal/config"
)

type BaseConfig struct {
	Enabled   bool   `json:"enabled"`
	Transport string `json:"transport"`
	Port      int    `json:"port"`
	Path      string `json:"path"`
}

func DefaultBaseConfig() BaseConfig {
	return BaseConfig{
		Enabled:   true,
		Transport: "streamable_http",
		Port:      5001,
		Path:      "/mcp",
	}
}

func RegisterTransport(r interface{ Handle(string, http.Handler) }, srv *Server, cfg BaseConfig, platformName string) {
	if !cfg.Enabled {
		return
	}
	switch cfg.Transport {
	case "streamable_http":
		r.Handle(cfg.Path, StreamableHTTPHandler(srv))
		config.Logger.Info("[MCP]["+platformName+"] registered streamable HTTP", "path", cfg.Path)
	case "sse":
		r.Handle(cfg.Path+"/sse", SSETransportHandler(srv))
		config.Logger.Info("[MCP]["+platformName+"] registered SSE transport", "path", cfg.Path+"/sse")
	case "stdio":
		config.Logger.Info("[MCP]["+platformName+"] stdio mode ready")
	default:
		r.Handle(cfg.Path, StreamableHTTPHandler(srv))
		config.Logger.Info("[MCP]["+platformName+"] registered default (streamable HTTP)", "path", cfg.Path)
	}
}

type PlatformDef struct {
	Name        string
	Description string
	Guide       func(baseURL string) string
	DefaultCfg  func() BaseConfig
}

var builtinPlatforms = map[string]PlatformDef{}

func RegisterBuiltinPlatform(def PlatformDef) {
	builtinPlatforms[def.Name] = def
}

func GetBuiltinPlatform(name string) (PlatformDef, bool) {
	p, ok := builtinPlatforms[name]
	return p, ok
}

func AllBuiltinPlatforms() map[string]PlatformDef {
	return builtinPlatforms
}

func BuiltinGuide(platformName, baseURL string) (string, error) {
	p, ok := builtinPlatforms[platformName]
	if !ok {
		return "", fmt.Errorf("unknown platform: %s", platformName)
	}
	if p.Guide == nil {
		return "", fmt.Errorf("no guide for platform: %s", platformName)
	}
	return p.Guide(baseURL), nil
}

func AllBuiltinGuides(baseURL string) map[string]string {
	guides := make(map[string]string)
	for name, p := range builtinPlatforms {
		if p.Guide != nil {
			guides[name] = p.Guide(baseURL)
		}
	}
	return guides
}

const (
	PlatformOpenClaw   = "openclaw"
	PlatformClaudeCode = "claude-code"
	PlatformJetBrains  = "jetbrains"
	PlatformOpenCode   = "opencode"
	PlatformVSCode     = "vscode"
)

type OpenClawConfig struct {
	BaseConfig
	APIKey      string `json:"api_key"`
	SessionAuth bool   `json:"session_auth"`
}

func DefaultOpenClawConfig() OpenClawConfig {
	return OpenClawConfig{
		BaseConfig:  DefaultBaseConfig(),
		SessionAuth: true,
	}
}

func RegisterOpenClaw(r interface{ Handle(string, http.Handler) }, srv *Server, cfg OpenClawConfig) {
	RegisterTransport(r, srv, cfg.BaseConfig, PlatformOpenClaw)
}

type ClaudeCodeConfig struct {
	BaseConfig
	MaxTokens int  `json:"max_tokens"`
	Thinking  bool `json:"thinking"`
}

func DefaultClaudeCodeConfig() ClaudeCodeConfig {
	base := DefaultBaseConfig()
	base.Transport = "stdio"
	return ClaudeCodeConfig{
		BaseConfig: base,
		MaxTokens:  16384,
		Thinking:   false,
	}
}

func RegisterClaudeCode(r interface{ Handle(string, http.Handler) }, srv *Server, cfg ClaudeCodeConfig) {
	RegisterTransport(r, srv, cfg.BaseConfig, PlatformClaudeCode)
}

type JetBrainsConfig struct {
	BaseConfig
}

func DefaultJetBrainsConfig() JetBrainsConfig {
	return JetBrainsConfig{BaseConfig: DefaultBaseConfig()}
}

func RegisterJetBrains(r interface{ Handle(string, http.Handler) }, srv *Server, cfg JetBrainsConfig) {
	RegisterTransport(r, srv, cfg.BaseConfig, PlatformJetBrains)
}

type OpenCodeConfig struct {
	BaseConfig
}

func DefaultOpenCodeConfig() OpenCodeConfig {
	base := DefaultBaseConfig()
	base.Transport = "stdio"
	return OpenCodeConfig{BaseConfig: base}
}

func RegisterOpenCode(r interface{ Handle(string, http.Handler) }, srv *Server, cfg OpenCodeConfig) {
	RegisterTransport(r, srv, cfg.BaseConfig, PlatformOpenCode)
}

type VSCodeConfig struct {
	BaseConfig
	AutoStart bool `json:"auto_start"`
	DebugMode bool `json:"debug_mode"`
}

func DefaultVSCodeConfig() VSCodeConfig {
	return VSCodeConfig{
		BaseConfig: DefaultBaseConfig(),
		AutoStart:  true,
		DebugMode:  false,
	}
}

func RegisterVSCode(r interface{ Handle(string, http.Handler) }, srv *Server, cfg VSCodeConfig) {
	RegisterTransport(r, srv, cfg.BaseConfig, PlatformVSCode)
}

func init() {
	RegisterBuiltinPlatform(PlatformDef{
		Name:        PlatformOpenClaw,
		Description: "OpenClaw MCP integration - supports both stdio and streamable HTTP transports",
		DefaultCfg: func() BaseConfig { return DefaultBaseConfig() },
		Guide:       openclawGuide,
	})
	RegisterBuiltinPlatform(PlatformDef{
		Name:        PlatformClaudeCode,
		Description: "Claude Code MCP integration - optimized for Claude Code CLI with stdio and SSE support",
		DefaultCfg: func() BaseConfig {
			c := DefaultBaseConfig()
			c.Transport = "stdio"
			return c
		},
		Guide: claudeCodeGuide,
	})
	RegisterBuiltinPlatform(PlatformDef{
		Name:        PlatformJetBrains,
		Description: "JetBrains IDEs (IntelliJ IDEA, PyCharm, WebStorm, GoLand) MCP integration via streamable HTTP",
		DefaultCfg: func() BaseConfig { return DefaultBaseConfig() },
		Guide:       jetbrainsGuide,
	})
	RegisterBuiltinPlatform(PlatformDef{
		Name:        PlatformOpenCode,
		Description: "OpenCode terminal-based AI coding assistant MCP integration via stdio transport",
		DefaultCfg: func() BaseConfig {
			c := DefaultBaseConfig()
			c.Transport = "stdio"
			return c
		},
		Guide: opencodeGuide,
	})
	RegisterBuiltinPlatform(PlatformDef{
		Name:        PlatformVSCode,
		Description: "Visual Studio Code MCP integration via streamable HTTP with native VS Code extension support",
		DefaultCfg: func() BaseConfig { return DefaultBaseConfig() },
		Guide:       vscodeGuide,
	})
}
