package httpx

import (
	"backend/internal/llm"
	"backend/internal/store"
	"encoding/json"
	"log"
	"net/http"
)

type annotateReq struct {
	Columns []string `json:"columns"`
}

type annotateResp struct {
	Column     string `json:"column"`
	Annotation string `json:"annotation"`
}

func RegisterHandlers(mux *http.ServeMux, llmClient *llm.Ollama, neo *store.Neo4j) {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// non-streaming: returns an array after finishing all columns
	mux.HandleFunc("/annotate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in annotateReq
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		out := make([]annotateResp, 0, len(in.Columns))
		for _, col := range in.Columns {
			log.Printf("🧠 Annotating: %s", col)
			text := llmClient.DescribeColumn(r.Context(), col)
			if text == "" {
				text = "⚠️ Empty model response"
			}
			if err := neo.SaveAnnotation(r.Context(), col, text); err != nil {
				log.Printf("❌ Neo4j save: %v", err)
			} else {
				log.Printf("✅ Saved: %s → %s", col, text)
			}
			out = append(out, annotateResp{Column: col, Annotation: text})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})

	// streaming: Server-Sent Events (SSE) – emits events per column and tokens
	mux.HandleFunc("/annotate/stream", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in annotateReq
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "stream unsupported", http.StatusInternalServerError)
			return
		}

		type chunk struct {
			Type   string `json:"type"` // start|token|end|saved
			Column string `json:"column"`
			Text   string `json:"text,omitempty"`
		}

		send := func(c chunk) {
			b, _ := json.Marshal(c)
			_, _ = w.Write([]byte("data: " + string(b) + "\n\n"))
			flusher.Flush()
		}

		for _, col := range in.Columns {
			send(chunk{Type: "start", Column: col})

			var builder string
			llmClient.StreamDescribeColumn(r.Context(), col, func(token string) {
				builder += token
				send(chunk{Type: "token", Column: col, Text: token})
			})

			_ = neo.SaveAnnotation(r.Context(), col, builder)
			send(chunk{Type: "saved", Column: col, Text: builder})
		}
	})
}

// Simple CORS wrapper
func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
