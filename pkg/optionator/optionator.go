package optionator

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"
)

var metadataCache sync.Map // map[reflect.Type][]fieldMetadata

type fieldMetadata struct {
	Index      []int
	Name       string
	DefaultTag string
	Required   bool
	Type       reflect.Type
}

// Option defines a function that modifies a target configuration object.
type Option[T any] func(target T) error

// With returns an Option that sets a specific field to a given value.
func With[T any](fieldName string, value interface{}) Option[T] {
	return func(target T) error {
		v := reflect.ValueOf(target)
		// Ensure target is a pointer to a struct.
		if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
			return errors.New("target must be a pointer to a struct")
		}
		elem := v.Elem()
		field := elem.FieldByName(fieldName)
		if !field.IsValid() {
			return fmt.Errorf("no such field: %s", fieldName)
		}
		if !field.CanSet() {
			return fmt.Errorf("cannot set field: %s", fieldName)
		}
		val := reflect.ValueOf(value)
		// Ensure the provided value is convertible to the field's type.
		if !val.Type().ConvertibleTo(field.Type()) {
			return fmt.Errorf("cannot convert %v to %v", val.Type(), field.Type())
		}
		field.Set(val.Convert(field.Type()))
		return nil
	}
}

// parseAndSetDefault sets the default value on the field based on its kind.
func parseAndSetDefault(field reflect.Value, defaultTag string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(defaultTag)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(defaultTag, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ui, err := strconv.ParseUint(defaultTag, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(ui)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(defaultTag, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(defaultTag)
		if err != nil {
			return err
		}
		field.SetBool(b)
	default:
		// Special handling for time.Duration.
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(defaultTag)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(d))
		}
	}
	return nil
}

// Config holds customizable tag names for defaults and required fields.
type Config struct {
	DefaultTag  string
	RequiredTag string
}

var defaultConfig = Config{
	DefaultTag:  "default",
	RequiredTag: "required",
}

// getTypeMetadata now accepts a Config parameter to use the correct tag names.
func getTypeMetadata(t reflect.Type, config Config) []fieldMetadata {
	if cached, ok := metadataCache.Load(t); ok {
		return cached.([]fieldMetadata)
	}
	var metadata []fieldMetadata
	// Iterate over struct fields.
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		// Only exportable fields.
		if sf.PkgPath != "" {
			continue
		}
		fm := fieldMetadata{
			Index:      sf.Index,
			Name:       sf.Name,
			DefaultTag: sf.Tag.Get(config.DefaultTag),
			Required:   sf.Tag.Get(config.RequiredTag) == "true",
			Type:       sf.Type,
		}
		metadata = append(metadata, fm)
	}
	metadataCache.Store(t, metadata)
	return metadata
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

// New creates a new configuration object from a pointer to a struct,
// sets any default values specified via struct tags, and applies the provided options.
func New[T any](target T, opts ...Option[T]) (T, error) {
	return NewWithConfig(target, defaultConfig, opts...)
}

// setDefaultRecursively applies default values recursively for nested structs.
func setDefaultRecursively(v reflect.Value, config Config) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			// Allocate new value if pointer is nil.
			v.Set(reflect.New(v.Type().Elem()))
		}
		return setDefaultRecursively(v.Elem(), config)
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	metadata := getTypeMetadata(t, config)
	for _, fm := range metadata {
		field := v.FieldByIndex(fm.Index)
		// If field is a struct or pointer to struct, apply defaults recursively.
		if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct) {
			if err := setDefaultRecursively(field, config); err != nil {
				return err
			}
		}
		// Only set default if field is zero and a default tag is provided.
		if isZeroValue(field) && fm.DefaultTag != "" {
			if err := parseAndSetDefault(field, fm.DefaultTag, fm.Type); err != nil {
				return fmt.Errorf("error setting default for field %s: %w", fm.Name, err)
			}
		}
	}
	return nil
}

// isZeroValue checks if a value is zero.
func isZeroValue(v reflect.Value) bool {
	zero := reflect.Zero(v.Type())
	return reflect.DeepEqual(v.Interface(), zero.Interface())
}

// validateRequiredFields checks if required fields are non-zero.
func validateRequiredFields(v reflect.Value, config Config) error {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return errors.New("nil pointer encountered in validation")
		}
		return validateRequiredFields(v.Elem(), config)
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	metadata := getTypeMetadata(t, config)
	for _, fm := range metadata {
		field := v.FieldByIndex(fm.Index)
		// For nested structs, validate recursively.
		if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct) {
			if err := validateRequiredFields(field, config); err != nil {
				return err
			}
		}
		if fm.Required && isZeroValue(field) {
			return fmt.Errorf("required field %s is zero", fm.Name)
		}
	}
	return nil
}
