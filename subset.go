package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

// checkSubset checks if subset is a subset of superset.
// Returns true if subset is contained in superset, along with any differences found.
func checkSubset(subset, superset interface{}) (bool, []string) {
	return checkSubsetPath(subset, superset, "$")
}

func checkSubsetPath(subset, superset interface{}, path string) (bool, []string) {
	// Handle nil cases
	if subset == nil {
		return superset == nil, diffIfNotEqual(path, subset, superset)
	}

	// Type check
	subsetType := reflect.TypeOf(subset)
	supersetType := reflect.TypeOf(superset)
	if subsetType != supersetType {
		return false, []string{fmt.Sprintf("%s: type mismatch (subset: %T, superset: %T)", path, subset, superset)}
	}

	switch subsetVal := subset.(type) {
	case map[string]interface{}:
		return checkObjectSubset(subsetVal, superset.(map[string]interface{}), path)
	case []interface{}:
		return checkArraySubset(subsetVal, superset.([]interface{}), path)
	default:
		// Primitive values: must be equal
		if reflect.DeepEqual(subset, superset) {
			return true, nil
		}
		return false, []string{fmt.Sprintf("%s: value mismatch (subset: %s, superset: %s)", path, formatValue(subset), formatValue(superset))}
	}
}

func checkObjectSubset(subset, superset map[string]interface{}, path string) (bool, []string) {
	var diffs []string
	isSubset := true

	// Get sorted keys for consistent output
	keys := make([]string, 0, len(subset))
	for k := range subset {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		subsetValue := subset[key]
		supersetValue, exists := superset[key]

		childPath := fmt.Sprintf("%s.%s", path, key)

		if !exists {
			isSubset = false
			diffs = append(diffs, fmt.Sprintf("%s: missing key in superset", childPath))
			continue
		}

		ok, childDiffs := checkSubsetPath(subsetValue, supersetValue, childPath)
		if !ok {
			isSubset = false
			diffs = append(diffs, childDiffs...)
		}
	}

	return isSubset, diffs
}

func checkArraySubset(subset, superset []interface{}, path string) (bool, []string) {
	// Set mode: every element in subset must exist in superset (order ignored)
	var diffs []string
	isSubset := true

	for i, subsetElem := range subset {
		found := false
		for _, supersetElem := range superset {
			if isSubsetOf(subsetElem, supersetElem) {
				found = true
				break
			}
		}
		if !found {
			isSubset = false
			diffs = append(diffs, fmt.Sprintf("%s[%d]: element not found in superset array: %s", path, i, formatValue(subsetElem)))
		}
	}

	return isSubset, diffs
}

// isSubsetOf checks if a is a subset of b (without collecting diffs)
func isSubsetOf(a, b interface{}) bool {
	ok, _ := checkSubsetPath(a, b, "")
	return ok
}

func diffIfNotEqual(path string, subset, superset interface{}) []string {
	if reflect.DeepEqual(subset, superset) {
		return nil
	}
	return []string{fmt.Sprintf("%s: value mismatch (subset: %s, superset: %s)", path, formatValue(subset), formatValue(superset))}
}

func formatValue(v interface{}) string {
	if v == nil {
		return "null"
	}

	bytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}

	s := string(bytes)
	// Truncate long values
	if len(s) > 50 {
		return s[:47] + "..."
	}
	return s
}
