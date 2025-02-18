package optionator

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

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

// New creates a new configuration object from a pointer to a struct,
// sets any default values specified via struct tags, and applies the provided options.
func New[T any](target T, opts ...Option[T]) (T, error) {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return target, errors.New("target must be a pointer to a struct")
	}

	s := v.Elem()
	t := s.Type()

	// Set default values from struct tags.
	for i := 0; i < t.NumField(); i++ {
		field := s.Field(i)
		if !field.CanSet() {
			continue
		}
		fieldType := t.Field(i)
		defaultTag := fieldType.Tag.Get("default")
		if defaultTag != "" {
			if err := parseAndSetDefault(field, defaultTag); err != nil {
				return target, fmt.Errorf("error setting default for field %s: %w", fieldType.Name, err)
			}
		}
	}

	// Apply provided options to override defaults.
	for _, opt := range opts {
		if err := opt(target); err != nil {
			return target, err
		}
	}

	return target, nil
}
