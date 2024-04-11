// Package server abstraction layer to make it easier to handle server related tasks.
package server

import (
	"net/http"

	"github.com/robotjoosen/usvc-message-consumer/pkg/server"
)

type Server struct {
	server.Server
}

func New(port int, routes map[string]http.HandlerFunc) *Server {
	srv := &Server{
		Server: server.Server{
			Port: port,
		},
	}

	srv.InitialiseRoutes(routes)

	return srv
}
