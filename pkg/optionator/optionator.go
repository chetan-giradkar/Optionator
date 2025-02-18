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

// New creates a new configuration object from a pointer to a struct,
// sets any default values specified via struct tags, and applies the provided options.
func New[T any](target T, opts ...Option[T]) (T, error) {
	return NewWithConfig(target, defaultConfig, opts...)
}

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
// It now accepts fieldType from metadata for enhanced type handling.
func parseAndSetDefault(field reflect.Value, defaultTag string, fieldType reflect.Type) error {
	if fieldType == reflect.TypeOf(time.Duration(0)) {
		d, err := time.ParseDuration(defaultTag)
		if err != nil {
			return err
		}
		// Set the underlying int64 value of the duration.
		field.SetInt(int64(d))
		return nil
	}

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
		return fmt.Errorf("unsupported field type: %v", fieldType)
	}
	return nil
}

// isZeroValue checks if a value is zero.
func isZeroValue(v reflect.Value) bool {
	zero := reflect.Zero(v.Type())
	return reflect.DeepEqual(v.Interface(), zero.Interface())
}
