package common

import "reflect"

func GetMapKeys(m interface{}) []string {
	// Use reflection to check if the input is a map
	val := reflect.ValueOf(m)
	if val.Kind() != reflect.Map {
		return nil
	}

	// Create a slice to store the keys
	keys := make([]string, 0, val.Len())

	// Iterate over the map and extract the keys
	for _, key := range val.MapKeys() {
		keys = append(keys, key.String())
	}

	return keys
}
