# Optionator

Optionator is a Go library that simplifies object configuration by automatically applying default values from struct tags and allowing functional options to override them.

## Features

- **Reflection Efficiency:** Caches field metadata for faster default value application.
- **Nested Struct Support:** Recursively applies defaults to nested or embedded structs.
- **Customizable Tag Names:** Configure which struct tags to use for defaults and required fields.
- **Validation:** Automatically validates that required fields (tagged with `required:"true"`) are non-zero.
- **Type-Safe Options:** Uses Go generics for a type-safe API.

## Example Usage

```go
package main

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/chetan-giradkar/Optionator/pkg/optionator"
)

type Server struct {
	Address   string        `default:"0.0.0.0" required:"true"`
	Timeout   time.Duration `default:"30s"`
	MaxConns  int           `default:"100"`
	TLSConfig *tls.Config
}

func main() {
	srv, err := optionator.New(&Server{},
		optionator.With[*Server]("Address", "127.0.0.1"),
		optionator.With[*Server]("MaxConns", 200),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Server Config: %+v\n", srv)
}
```
