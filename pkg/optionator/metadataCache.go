package optionator

import (
	"reflect"
	"sync"
)

var metadataCache sync.Map // map[reflect.Type][]fieldMetadata

type fieldMetadata struct {
	Index      []int
	Name       string
	DefaultTag string
	Required   bool
	Type       reflect.Type
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
