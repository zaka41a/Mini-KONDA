package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpx "backend/internal/http"
	"backend/internal/llm"
	"backend/internal/store"
)

func main() {
	// ---- Config with sensible defaults
	ollamaURL := getenv("OLLAMA_URL", "http://127.0.0.1:11434")
	model := getenv("OLLAMA_MODEL", "mistral")
	neoURL := getenv("NEO4J_URL", "neo4j://localhost:7687")
	neoUser := getenv("NEO4J_USER", "neo4j")
	neoPass := getenv("NEO4J_PASS", "password")
	addr := getenv("HTTP_ADDR", ":8080")

	// ---- Clients
	neo, err := store.NewNeo4j(neoURL, neoUser, neoPass)
	if err != nil {
		log.Fatalf("❌ Neo4j init: %v", err)
	}
	defer neo.Close(context.Background())

	llmClient := llm.NewOllama(ollamaURL, model)

	// ---- HTTP server
	mux := http.NewServeMux()
	httpx.RegisterHandlers(mux, llmClient, neo)

	srv := &http.Server{
		Addr:              addr,
		Handler:           httpx.WithCORS(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("🚀 Backend running on http://localhost%s", addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("👋 Shutdown complete")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
