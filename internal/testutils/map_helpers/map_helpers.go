package map_helpers

import (
	"reflect"
)

func Equal(map1, map2 map[string]interface{}) bool {
	// Check if the lengths of the maps are equal.
	if len(map1) != len(map2) {
		return false
	}

	// Iterate through the keys and values of map1.
	for key, value1 := range map1 {
		// Check if the key exists in map2.
		value2, exists := map2[key]
		if !exists {
			return false
		}

		// Use reflection to compare the values.
		if !reflect.DeepEqual(value1, value2) {
			// Values are not equal, return false.
			return false
		}

		// If both values are maps, recursively compare them.
		if subMap1, isMap1 := value1.(map[string]interface{}); isMap1 {
			if subMap2, isMap2 := value2.(map[string]interface{}); isMap2 {
				if !Equal(subMap1, subMap2) {
					// Recursive comparison failed, return false.
					return false
				}
			} else {
				// One value is a map and the other is not, return false.
				return false
			}
		}
	}

	// All keys and values have been compared and are equal.
	return true
}

func Merge(target, source map[string]interface{}) map[string]interface{} {
	// Create a new map to store the merged values.
	merged := make(map[string]interface{})

	// Copy values from the target map to the merged map.
	for key, targetValue := range target {
		merged[key] = targetValue
	}

	// Merge values from the source map into the merged map.
	for key, sourceValue := range source {
		targetValue, exists := merged[key]
		if !exists {
			// Key doesn't exist in the merged map, add it.
			merged[key] = sourceValue
			continue
		}

		// Check if both values are maps. If so, recursively merge them.
		if sourceMap, sourceIsMap := sourceValue.(map[string]interface{}); sourceIsMap {
			if targetMap, targetIsMap := targetValue.(map[string]interface{}); targetIsMap {
				// Recursive call to merge maps.
				merged[key] = Merge(targetMap, sourceMap)
				continue
			}
		}

		// For other types or when a recursive merge isn't possible, overwrite with the source value.
		merged[key] = sourceValue
	}

	return merged
}

func Copy(source map[string]interface{}) map[string]interface{} {
	return Merge(map[string]interface{}{}, source)
}
