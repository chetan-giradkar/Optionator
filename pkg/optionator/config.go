package optionator

import (
	"errors"
	"reflect"
)

// Config holds customizable tag names for defaults and required fields.
type Config struct {
	DefaultTag  string
	RequiredTag string
}

var defaultConfig = Config{
	DefaultTag:  "default",
	RequiredTag: "required",
}

// NewWithConfig creates a new configuration object using the provided config.
func NewWithConfig[T any](target T, config Config, opts ...Option[T]) (T, error) {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return target, errors.New("target must be a pointer to a struct")
	}
	// Set defaults recursively.
	if err := setDefaultRecursively(v.Elem(), config); err != nil {
		return target, err
	}
	// Apply provided options to override defaults.
	for _, opt := range opts {
		if err := opt(target); err != nil {
			return target, err
		}
	}
	// Validate required fields.
	if err := validateRequiredFields(v.Elem(), config); err != nil {
		return target, err
	}
	return target, nil
}
