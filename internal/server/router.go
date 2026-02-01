// Package server provides the HTTP router and middleware so main and tests can share it.
package server

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"habit-tracker-be/internal/handlers"
)

// responseWriter wraps http.ResponseWriter to capture status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// requestLogger logs every request to the terminal: method, path, status, size, duration.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrap := &responseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(wrap, r)
		dur := time.Since(start)
		log.Printf("[REQUEST] %s %s -> %d (%d bytes) in %v", r.Method, r.URL.Path, wrap.status, wrap.bytes, dur.Round(time.Millisecond))
	})
}

// Router returns the chi router with all middleware and routes mounted.
// Used by main for ListenAndServe and by tests for httptest.
func Router() http.Handler {
	allowedOrigins := []string{"http://localhost:5173"}
	if v := os.Getenv("ALLOWED_ORIGINS"); v != "" {
		allowedOrigins = strings.Split(v, ",")
		for i, o := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(o)
		}
	}

	r := chi.NewRouter()
	r.Use(requestLogger)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/habits", handlers.GetHabits)
		r.Post("/habits", handlers.CreateHabit)
		r.Post("/habits/{id}/log", handlers.ToggleHabitLog)
		r.Delete("/habits/{id}", handlers.DeleteHabit)
	})

	return r
}
