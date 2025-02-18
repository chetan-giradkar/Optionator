package optionator

import (
	"crypto/tls"
	"testing"
	"time"
)

type NestedConfig struct {
	Port int    `default:"8080" required:"true"`
	Host string `default:"localhost" required:"true"`
}

type Server struct {
	Address   string        `default:"0.0.0.0" required:"true"`
	Timeout   time.Duration `default:"30s"`
	MaxConns  int           `default:"100"`
	TLSConfig *tls.Config   // no default provided.
	Nested    *NestedConfig // nested struct with defaults.
}

func TestNewDefaults(t *testing.T) {
	s, err := New(&Server{})
	if err != nil {
		t.Fatalf("Error creating server: %v", err)
	}
	if s.Address != "0.0.0.0" {
		t.Errorf("Expected Address to be '0.0.0.0', got '%s'", s.Address)
	}
	if s.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout to be 30s, got '%v'", s.Timeout)
	}
	if s.MaxConns != 100 {
		t.Errorf("Expected MaxConns to be 100, got %d", s.MaxConns)
	}
	// Validate nested defaults
	if s.Nested == nil {
		t.Errorf("Expected Nested to be allocated, got nil")
	} else {
		if s.Nested.Port != 8080 {
			t.Errorf("Expected Nested.Port to be 8080, got %d", s.Nested.Port)
		}
		if s.Nested.Host != "localhost" {
			t.Errorf("Expected Nested.Host to be 'localhost', got '%s'", s.Nested.Host)
		}
	}
}

func TestOverrideAndValidation(t *testing.T) {
	// Override Address and MaxConns.
	s, err := New(&Server{},
		With[*Server]("Address", "127.0.0.1"),
		With[*Server]("MaxConns", 200),
	)
	if err != nil {
		t.Fatalf("Error creating server: %v", err)
	}
	if s.Address != "127.0.0.1" {
		t.Errorf("Expected Address to be '127.0.0.1', got '%s'", s.Address)
	}
	if s.MaxConns != 200 {
		t.Errorf("Expected MaxConns to be 200, got %d", s.MaxConns)
	}
	// For nested, defaults should be applied and validated.
	if s.Nested.Port != 8080 {
		t.Errorf("Expected Nested.Port to be 8080 by default, got %d", s.Nested.Port)
	}
}

func TestRequiredValidationFailure(t *testing.T) {
	type TestStruct struct {
		Field1 string `default:"" required:"true"`
	}
	_, err := New(&TestStruct{})
	if err == nil {
		t.Errorf("Expected error due to required field Field1, but got none")
	}
}
