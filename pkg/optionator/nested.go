package optionator

import (
	"fmt"
	"reflect"
)

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
