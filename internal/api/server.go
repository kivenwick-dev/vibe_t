package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"banner-fingerprint/internal/fingerprint"
	"banner-fingerprint/internal/model"
)

type Server struct {
	engine *fingerprint.Engine
	logger *slog.Logger
}

func NewServer(engine *fingerprint.Engine, logger *slog.Logger) *Server {
	return &Server{engine: engine, logger: logger}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("POST /fingerprint", s.fingerprint)
	return requestLogger(s.logger, mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) fingerprint(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	var inputs []model.BannerInput
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&inputs); err != nil {
		writeError(w, http.StatusBadRequest, "request body must be a JSON array")
		return
	}
	if len(inputs) > 10000 {
		writeError(w, http.StatusRequestEntityTooLarge, "maximum batch size is 10000")
		return
	}

	results := make([]model.FingerprintResult, 0, len(inputs))
	for _, input := range inputs {
		results = append(results, s.engine.Identify(input))
	}
	writeJSON(w, http.StatusOK, results)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start).String())
	})
}
