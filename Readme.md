# Optionator

Optionator is a Go library that simplifies object configuration by automatically applying default values from struct tags and allowing functional options to override them.

## Overview

The library uses reflection to:
- Set default values on struct fields using the `default` tag.
- Provide a type-safe mechanism to override those defaults using functional options.

## Example Usage

```go
package main

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/chetan-giradkar/Optionator/pkg/optionator"
)

// Server represents a configurable HTTP server with defaults.
type Server struct {
	Address   string        `default:"0.0.0.0:8080"`
	Timeout   time.Duration `default:"30s"`
	MaxConns  int           `default:"100"`
	TLSConfig *tls.Config   // No default provided.
}

func main() {
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
