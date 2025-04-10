package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func SliceContains[T comparable](slice []T, element T) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}

func ConvertToJSONAndBack(value any, target any) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("error marshalling value: %v", err)
	}

	targetType := reflect.TypeOf(target)

	if targetType.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer to a variable")
	}

	err = json.Unmarshal(jsonData, target)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON into target: %v", err)
	}

	return nil
}
