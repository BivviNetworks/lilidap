package map_helpers_test

import (
	"lilidap/internal/testutils/map_helpers"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEqualMaps(t *testing.T) {
	map1 := map[string]interface{}{
		"name": "John",
		"age":  30,
		"address": map[string]interface{}{
			"city":    "New York",
			"country": "USA",
		},
	}

	map2 := map[string]interface{}{
		"name": "John",
		"age":  30,
		"address": map[string]interface{}{
			"city":    "New York",
			"country": "USA",
		},
	}

	map3 := map[string]interface{}{
		"name": "Alice",
		"age":  25,
		"address": map[string]interface{}{
			"city":    "Los Angeles",
			"country": "USA",
		},
	}

	// Compare map1 and map2 for equality
	require.True(t, map_helpers.Equal(map1, map2))

	// Compare map1 and map3 for equality
	require.False(t, map_helpers.Equal(map1, map3))
}

func TestCopyMap(t *testing.T) {
	map1 := map[string]interface{}{
		"name": "John",
		"age":  30,
		"address": map[string]interface{}{
			"city":    "New York",
			"country": "USA",
		},
	}
	// Compare map1 and map2 for equality
	require.True(t, map_helpers.Equal(map1, map_helpers.Copy(map1)))
}

func TestMergeMaps(t *testing.T) {
	map1 := map[string]interface{}{
		"name": "John",
		"age":  30,
		"address": map[string]interface{}{
			"city":    "New York",
			"country": "USA",
		},
	}

	map2 := map[string]interface{}{
		"age": 40,
		"address": map[string]interface{}{
			"state": "NY",
		},
	}

	map3 := map[string]interface{}{
		"name": "John",
		"age":  40,
		"address": map[string]interface{}{
			"city":    "New York",
			"state":   "NY",
			"country": "USA",
		},
	}

	omap1 := map_helpers.Copy(map1)
	omap2 := map_helpers.Copy(map2)

	// Create a new merged map without modifying the original maps.
	merged := map_helpers.Merge(map1, map2)

	require.True(t, map_helpers.Equal(omap1, map1))
	require.True(t, map_helpers.Equal(omap2, map2))
	require.True(t, map_helpers.Equal(map3, merged))

}
