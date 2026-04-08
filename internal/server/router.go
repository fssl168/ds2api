package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"ds2api/internal/account"
	"ds2api/internal/adapter/claude"
	"ds2api/internal/adapter/gemini"
	"ds2api/internal/adapter/openai"
	"ds2api/internal/admin"
	"ds2api/internal/auth"
	"ds2api/internal/config"
	"ds2api/internal/deepseek"
	"ds2api/internal/mcp"
	"ds2api/internal/qwen"
	"ds2api/internal/webui"
)

type App struct {
	Store    *config.Store
	Pool     *account.Pool
	Resolver *auth.Resolver
	DS       *deepseek.Client
	QW       *qwen.Client
	Router   http.Handler
}

func NewApp() *App {
	store := config.LoadStore()
	pool := account.NewPool(store)
	var dsClient *deepseek.Client
	resolver := auth.NewResolver(store, pool, func(ctx context.Context, acc config.Account) (string, error) {
		return dsClient.Login(ctx, acc)
	})
	dsClient = deepseek.NewClient(store, resolver)
	qwenClient := qwen.NewClient(store)
	if err := qwenClient.Preload(context.Background()); err != nil {
		config.Logger.Warn("[QWEN] preload failed", "error", err)
	} else {
		config.Logger.Info("[QWEN] client initialized", "tickets", len(qwenClient.Tickets()))
	}
	if err := dsClient.PreloadPow(context.Background()); err != nil {
		config.Logger.Warn("[PoW] solver init failed", "error", err)
	} else {
		config.Logger.Info("[PoW] native Go solver active")
	}

	openaiHandler := &openai.Handler{Store: store, Auth: resolver, DS: dsClient, QW: qwenClient}
	claudeHandler := &claude.Handler{Store: store, Auth: resolver, DS: dsClient}
	geminiHandler := &gemini.Handler{Store: store, Auth: resolver, DS: dsClient}
	adminHandler := &admin.Handler{Store: store, Pool: pool, DS: dsClient, QW: qwenClient}
	webuiHandler := webui.NewHandler()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors)
	rl := newRateLimiter(time.Minute, 120)
	r.Use(rl.Middleware())
	r.Use(timeout(0))

	healthzHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
	readyzHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	}
	r.Get("/healthz", healthzHandler)
	r.Head("/healthz", healthzHandler)
	r.Get("/readyz", readyzHandler)
	r.Head("/readyz", readyzHandler)
	openai.RegisterRoutes(r, openaiHandler)
	claude.RegisterRoutes(r, claudeHandler)
	gemini.RegisterRoutes(r, geminiHandler)
	r.Route("/admin", func(ar chi.Router) {
		admin.RegisterRoutes(ar, adminHandler)
	})
	webui.RegisterRoutes(r, webuiHandler)
	registerMCPRoutes(r, dsClient, qwenClient, store)
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/admin/") && webuiHandler.HandleAdminFallback(w, req) {
			return
		}
		http.NotFound(w, req)
	})

	return &App{Store: store, Pool: pool, Resolver: resolver, DS: dsClient, QW: qwenClient, Router: r}
}

func timeout(d time.Duration) func(http.Handler) http.Handler {
	if d <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	return middleware.Timeout(d)
}

type rateLimiter struct {
	mu      sync.Mutex
	clients map[string][]time.Time
	window  time.Duration
	limit   int
}

func newRateLimiter(window time.Duration, limit int) *rateLimiter {
	rl := &rateLimiter{
		clients: make(map[string][]time.Time),
		window:  window,
		limit:   limit,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				ip = strings.TrimSpace(strings.Split(fwd, ",")[0])
			}
			if !rl.allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"type": "rate_limit_error", "message": "Rate limit exceeded"}})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	times := rl.clients[ip]
	var valid []time.Time
	for _, t := range times {
		if now.Sub(t) < rl.window {
			valid = append(valid, t)
		}
	}
	if len(valid) >= rl.limit {
		return false
	}
	valid = append(valid, now)
	rl.clients[ip] = valid
	return true
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, times := range rl.clients {
			var valid []time.Time
			for _, t := range times {
				if now.Sub(t) < rl.window {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.clients, ip)
			} else {
				rl.clients[ip] = valid
			}
		}
		rl.mu.Unlock()
	}
}

func cors(next http.Handler) http.Handler {
	allowedOrigin := strings.TrimSpace(os.Getenv("CORS_ORIGIN"))
	if allowedOrigin == "" {
		allowedOrigin = "*"
	} else {
		config.Logger.Warn("[security] CORS origin restricted", "origin", allowedOrigin)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Ds2-Target-Account, X-Vercel-Protection-Bypass")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func WriteUnhandledError(w http.ResponseWriter, err error) {
	config.Logger.Error("[unhandled]", "error", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"type": "api_error", "message": "Internal Server Error"}})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func registerMCPRoutes(r *chi.Mux, ds *deepseek.Client, qw *qwen.Client, store *config.Store) {
	registry := mcp.NewPluginRegistry(store, ds, qw)

	baseURL := "http://127.0.0.1:" + strings.TrimSpace(os.Getenv("PORT"))
	if baseURL == "http://127.0.0.1:" {
		baseURL = "http://127.0.0.1:5001"
	}

	openclawCfg := mcp.DefaultOpenClawConfig()
	claudeCodeCfg := mcp.DefaultClaudeCodeConfig()
	jetbrainsCfg := mcp.DefaultJetBrainsConfig()
	opencodeCfg := mcp.DefaultOpenCodeConfig()
	vscodeCfg := mcp.DefaultVSCodeConfig()

	mcp.RegisterOpenClaw(r, registry.Server(), openclawCfg)
	mcp.RegisterClaudeCode(r, registry.Server(), claudeCodeCfg)
	mcp.RegisterJetBrains(r, registry.Server(), jetbrainsCfg)
	mcp.RegisterOpenCode(r, registry.Server(), opencodeCfg)
	mcp.RegisterVSCode(r, registry.Server(), vscodeCfg)

	r.Get("/mcp/guides", func(w http.ResponseWriter, _ *http.Request) {
		guides := mcp.AllBuiltinGuides(baseURL)
		writeJSON(w, http.StatusOK, map[string]any{"platforms": guides})
	})

	r.Get("/mcp/guide/{platform}", func(w http.ResponseWriter, req *http.Request) {
		platform := chi.URLParam(req, "platform")
		guide, err := mcp.BuiltinGuide(platform, baseURL)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"platform": platform, "guide": guide})
	})

	config.Logger.Info("[MCP] plugin bridge initialized",
		"platforms", len(mcp.AllBuiltinPlatforms()),
		"transport", mcp.MCPTransport(),
	)

	if mcp.IsMCPMode() && mcp.MCPTransport() == "stdio" {
		go func() {
			ctx := context.Background()
			if err := registry.RunStdioMode(ctx); err != nil {
				config.Logger.Error("[MCP] stdio mode error", "error", err)
			}
		}()
	}
}
