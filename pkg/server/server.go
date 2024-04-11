// Package server is a http router wrapper for basic server functionality.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	Port   int
	mux    *http.ServeMux
	server *http.Server
}

type RFC7808 struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func (s *Server) InitialiseRoutes(routeHandlers map[string]http.HandlerFunc) *http.ServeMux {
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/", s.NotFoundResponse)

	for pattern, routeFunc := range routeHandlers {
		s.mux.HandleFunc(pattern, routeFunc)
	}

	return s.mux
}

func (s *Server) Run() {
	slog.Info("starting server",
		slog.Int("port", s.Port),
	)

	go func() {
		s.server = &http.Server{
			Addr:              fmt.Sprintf(":%d", s.Port),
			Handler:           s.mux,
			ReadHeaderTimeout: 5 * time.Second,
		}

		if err := s.server.ListenAndServe(); err != nil {
			return
		}
	}()
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return
	}
}

func (s *Server) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	slog.Warn("no response available",
		slog.Int("status_code", http.StatusNotFound),
		slog.String("path", r.RequestURI),
	)

	ErrorResponse(w, "not found", "no handler defined for path")
}

func SuccessResponse(w http.ResponseWriter, content string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, content)
}

func ErrorResponse(w http.ResponseWriter, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusInternalServerError)

	msg, err := json.Marshal(RFC7808{Type: "about:blank", Title: title, Detail: detail})
	if err != nil {
		slog.Error(err.Error())

		return
	}

	fmt.Fprint(w, string(msg))
}
