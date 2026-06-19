package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"system-health-monitor/collectors"
)

//go:embed dashboard/index.html
var dashboardFS embed.FS

func main() {
	port := envOrDefault("PORT", "9090")
	runner := collectors.ExecRunner{}
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/status", statusHandler(runner))
	mux.HandleFunc("/", dashboardHandler)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("system-health-monitor listening on http://127.0.0.1:%s", port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func statusHandler(runner collectors.Runner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		type result struct {
			name string
			set  func(*collectors.StatusResponse)
		}

		resultCh := make(chan result, 5)
		go func() {
			section := collectors.CollectProfiles(ctx, runner)
			resultCh <- result{name: "profiles", set: func(resp *collectors.StatusResponse) { resp.Profiles = section }}
		}()
		go func() {
			section := collectors.CollectKanban(ctx, runner)
			resultCh <- result{name: "kanban", set: func(resp *collectors.StatusResponse) { resp.Kanban = section }}
		}()
		go func() {
			section := collectors.CollectMCP(ctx, runner)
			resultCh <- result{name: "mcp", set: func(resp *collectors.StatusResponse) { resp.MCP = section }}
		}()
		go func() {
			section := collectors.CollectGateway(ctx, runner)
			resultCh <- result{name: "gateway", set: func(resp *collectors.StatusResponse) { resp.Gateway = section }}
		}()
		go func() {
			section := collectors.CollectSystem(ctx, runner)
			resultCh <- result{name: "system", set: func(resp *collectors.StatusResponse) { resp.System = section }}
		}()

		response := collectors.StatusResponse{Updated: time.Now().Format(time.RFC3339)}
		for i := 0; i < 5; i++ {
			select {
			case <-ctx.Done():
				markUnfinishedSections(&response)
				writeJSON(w, http.StatusOK, response)
				return
			case item := <-resultCh:
				item.set(&response)
			}
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func markUnfinishedSections(resp *collectors.StatusResponse) {
	if resp.Profiles.Data == nil && resp.Profiles.Error == "" {
		resp.Profiles.Error = "command timed out"
	}
	if resp.Kanban.Data == nil && resp.Kanban.Error == "" {
		resp.Kanban.Error = "command timed out"
	}
	if resp.MCP.Data == nil && resp.MCP.Error == "" {
		resp.MCP.Error = "command timed out"
	}
	if resp.Gateway.Data == nil && resp.Gateway.Error == "" {
		resp.Gateway.Error = "command timed out"
	}
	if resp.System.Data == nil && resp.System.Error == "" {
		resp.System.Error = "command timed out"
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	page, err := dashboardFS.ReadFile("dashboard/index.html")
	if err != nil {
		http.Error(w, "dashboard unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(page)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("json encode error: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	if _, err := strconv.Atoi(value); err != nil {
		return fallback
	}
	return value
}
