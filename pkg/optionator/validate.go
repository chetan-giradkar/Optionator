package optionator

import (
	"errors"
	"fmt"
	"reflect"
)

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
