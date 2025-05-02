package parser

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrInvalidMapKey    = errors.New("invalid key")
	ErrInvalidFieldType = errors.New("invalid field type")
)

// ParseJSONMapInt64 parses a map[string]any and returns the value as an int64.
// Since json.Unmarshal() converts all numbers to float64, we need to convert them back to int64.
// Caller should ensure that castint from float64 to int64 is safe.
func ParseJSONMapInt64(m map[string]any, key string) (int64, error) {
	value, ok := m[key].(float64)
	if !ok {
		return 0, ErrInvalidMapKey
	}

	return int64(value), nil
}

// ParseJSONMapString parses a map[string]any and returns the value as a string.
func ParseJSONMapString(m map[string]any, key string) (string, error) {
	value, ok := m[key].(string)
	if !ok {
		return "", ErrInvalidMapKey
	}

	return value, nil
}

// StructsToFilteredMaps converts a slice of structs to filtered maps using dot notation.
func StructsToFilteredMaps(v interface{}, includeFields []string) ([]map[string]interface{}, error) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, fmt.Errorf("%w: expected slice or array, got %T", ErrInvalidFieldType, v)
	}

	// Parse includeFields into a hierarchical structure
	fieldTree, err := buildFieldTree(includeFields)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, val.Len())

	for i := range val.Len() {
		elem := indirect(val.Index(i))

		if elem.Kind() != reflect.Struct {
			return nil, fmt.Errorf("%w: element %d is not a struct", ErrInvalidFieldType, i)
		}

		filteredMap, err := filterStruct(elem, fieldTree)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}

		result[i] = filteredMap
	}

	return result, nil
}

func indirect(val reflect.Value) reflect.Value {
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return val
}

// buildFieldTree converts dot-notation fields into a tree structure.
func buildFieldTree(fields []string) (map[string]interface{}, error) {
	tree := make(map[string]interface{})

	for _, field := range fields {
		parts := strings.Split(field, ".")
		current := tree

		for i, part := range parts {
			if i == len(parts)-1 {
				current[part] = true // Mark as leaf node
			} else {
				if _, exists := current[part]; !exists {
					current[part] = make(map[string]interface{})
				}

				if next, ok := current[part].(map[string]interface{}); ok {
					current = next
				} else {
					return nil, fmt.Errorf("%w: expected map for part %q, got %T", ErrInvalidFieldType, part, current[part])
				}
			}
		}
	}

	return tree, nil
}

func filterStruct(val reflect.Value, tree map[string]interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	typ := val.Type()

	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if fieldType.PkgPath != "" {
			continue // unexported
		}

		name := getFieldName(fieldType)

		subtree, exists := tree[name]
		if !exists {
			continue
		}

		filteredVal, err := processField(field, subtree)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}

		if filteredVal != nil {
			result[name] = filteredVal
		}
	}

	return result, nil
}

func processField(field reflect.Value, subtree interface{}) (interface{}, error) { //nolint: ireturn
	//nolint: exhaustive
	switch node := subtree.(type) {
	case bool:
		return field.Interface(), nil
	case map[string]interface{}:
		switch field.Kind() {
		case reflect.Struct:
			return filterStruct(field, node)
		case reflect.Slice, reflect.Array:
			return filterSlice(field, node)
		default:
			return nil, fmt.Errorf("%w: unsupported nested type", ErrInvalidFieldType)
		}
	default:
		return nil, fmt.Errorf("%w: invalid field tree structure", ErrInvalidFieldType)
	}
}

func filterSlice(val reflect.Value, tree map[string]interface{}) ([]interface{}, error) {
	//nolint: prealloc
	var result []interface{}

	for i := range val.Len() {
		elem := indirect(val.Index(i))
		if elem.Kind() != reflect.Struct {
			return nil, fmt.Errorf("%w: slice element %d is not a struct", ErrInvalidFieldType, i)
		}

		filtered, err := filterStruct(elem, tree)
		if err != nil {
			return nil, fmt.Errorf("slice element %d: %w", i, err)
		}

		result = append(result, filtered)
	}

	return result, nil
}

func getFieldName(field reflect.StructField) string {
	if tag := field.Tag.Get("json"); tag != "" {
		if name := strings.Split(tag, ",")[0]; name != "" {
			return name
		}
	}

	return field.Name
}
