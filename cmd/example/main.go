package main

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/yourusername/optionator/pkg/optionator"
)

// Server represents a configurable HTTP server with defaults.
type Server struct {
	Address   string        `default:"0.0.0.0:8080"`
	Timeout   time.Duration `default:"30s"`
	MaxConns  int           `default:"100"`
	TLSConfig *tls.Config   // No default provided.
}

func main() {
	// Create a new Server with auto-generated options.
	srv, err := optionator.New(&Server{},
		optionator.With[*Server]("Address", "127.0.0.1:8081"),
		optionator.With[*Server]("Timeout", 60*time.Second),
		optionator.With[*Server]("MaxConns", 200),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Server Config: %+v\n", srv)
}
