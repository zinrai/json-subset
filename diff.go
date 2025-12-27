package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/theory/jsonpath/spec"
)

// DiffType represents the type of difference
type DiffType int

const (
	DiffMissingKey DiffType = iota
	DiffValueMismatch
	DiffTypeMismatch
	DiffElementNotFound
)

// Diff represents a single difference
type Diff struct {
	Path          spec.NormalizedPath
	Type          DiffType
	SubsetValue   interface{}
	SupersetValue interface{}
}

// Line represents a single line of output with its path
type Line struct {
	Content string
	Path    spec.NormalizedPath
}

// checkSubsetWithDiffs checks if subset is a subset of superset.
func checkSubsetWithDiffs(subset, superset interface{}) (bool, []Diff) {
	return checkSubsetPath(subset, superset, spec.NormalizedPath{})
}

func checkSubsetPath(subset, superset interface{}, path spec.NormalizedPath) (bool, []Diff) {
	if subset == nil {
		if superset == nil {
			return true, nil
		}
		return false, []Diff{{Path: copyPath(path), Type: DiffValueMismatch, SubsetValue: subset, SupersetValue: superset}}
	}

	subsetMap, subsetIsMap := subset.(map[string]interface{})
	supersetMap, supersetIsMap := superset.(map[string]interface{})

	subsetArr, subsetIsArr := subset.([]interface{})
	supersetArr, supersetIsArr := superset.([]interface{})

	if subsetIsMap && !supersetIsMap {
		return false, []Diff{{Path: copyPath(path), Type: DiffTypeMismatch, SubsetValue: subset, SupersetValue: superset}}
	}
	if subsetIsArr && !supersetIsArr {
		return false, []Diff{{Path: copyPath(path), Type: DiffTypeMismatch, SubsetValue: subset, SupersetValue: superset}}
	}

	if subsetIsMap {
		return checkObjectSubset(subsetMap, supersetMap, path)
	}
	if subsetIsArr {
		return checkArraySubset(subsetArr, supersetArr, path)
	}

	if subset == superset {
		return true, nil
	}
	if subsetFloat, ok := subset.(float64); ok {
		if supersetFloat, ok := superset.(float64); ok {
			if subsetFloat == supersetFloat {
				return true, nil
			}
		}
	}
	return false, []Diff{{Path: copyPath(path), Type: DiffValueMismatch, SubsetValue: subset, SupersetValue: superset}}
}

func checkObjectSubset(subset, superset map[string]interface{}, path spec.NormalizedPath) (bool, []Diff) {
	var diffs []Diff
	isSubset := true

	keys := make([]string, 0, len(subset))
	for k := range subset {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		subsetValue := subset[key]
		supersetValue, exists := superset[key]
		childPath := append(copyPath(path), spec.Name(key))

		if !exists {
			isSubset = false
			diffs = append(diffs, Diff{Path: childPath, Type: DiffMissingKey, SubsetValue: subsetValue})
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

func checkArraySubset(subset, superset []interface{}, path spec.NormalizedPath) (bool, []Diff) {
	var diffs []Diff
	isSubset := true

	for i, subsetElem := range subset {
		found := false
		for _, supersetElem := range superset {
			ok, _ := checkSubsetPath(subsetElem, supersetElem, spec.NormalizedPath{})
			if ok {
				found = true
				break
			}
		}
		if !found {
			isSubset = false
			childPath := append(copyPath(path), spec.Index(i))
			diffs = append(diffs, Diff{Path: childPath, Type: DiffElementNotFound, SubsetValue: subsetElem})
		}
	}

	return isSubset, diffs
}

// copyPath creates a copy of a NormalizedPath
func copyPath(path spec.NormalizedPath) spec.NormalizedPath {
	return append(spec.NormalizedPath{}, path...)
}

// FormatDiffOutput formats the subset JSON with diff markers
func FormatDiffOutput(subset interface{}, diffs []Diff) string {
	diffPaths := make(map[string]bool)
	for _, d := range diffs {
		diffPaths[d.Path.String()] = true
	}

	lines := generateLines(subset, spec.NormalizedPath{}, 0)
	return formatOutput(lines, diffPaths)
}

// generateLines generates lines from JSON value with path information
func generateLines(value interface{}, path spec.NormalizedPath, indent int) []Line {
	indentStr := strings.Repeat("  ", indent)

	switch v := value.(type) {
	case map[string]interface{}:
		return generateObjectLines(v, path, indent)

	case []interface{}:
		return generateArrayLines(v, path, indent)

	default:
		return []Line{{Content: indentStr + formatPrimitive(value), Path: copyPath(path)}}
	}
}

func generateObjectLines(obj map[string]interface{}, path spec.NormalizedPath, indent int) []Line {
	indentStr := strings.Repeat("  ", indent)
	var lines []Line

	lines = append(lines, Line{Content: indentStr + "{", Path: copyPath(path)})

	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, key := range keys {
		childPath := append(copyPath(path), spec.Name(key))
		childValue := obj[key]
		comma := ","
		if i == len(keys)-1 {
			comma = ""
		}

		childLines := generateKeyValueLines(key, childValue, childPath, indent+1, comma)
		lines = append(lines, childLines...)
	}

	lines = append(lines, Line{Content: indentStr + "}", Path: copyPath(path)})
	return lines
}

func generateArrayLines(arr []interface{}, path spec.NormalizedPath, indent int) []Line {
	indentStr := strings.Repeat("  ", indent)
	var lines []Line

	lines = append(lines, Line{Content: indentStr + "[", Path: copyPath(path)})

	for i, elem := range arr {
		childPath := append(copyPath(path), spec.Index(i))
		comma := ","
		if i == len(arr)-1 {
			comma = ""
		}

		childLines := generateLines(elem, childPath, indent+1)
		if len(childLines) > 0 {
			lastIdx := len(childLines) - 1
			childLines[lastIdx].Content += comma
		}
		lines = append(lines, childLines...)
	}

	lines = append(lines, Line{Content: indentStr + "]", Path: copyPath(path)})
	return lines
}

func generateKeyValueLines(key string, value interface{}, path spec.NormalizedPath, indent int, comma string) []Line {
	indentStr := strings.Repeat("  ", indent)

	switch v := value.(type) {
	case map[string]interface{}:
		var lines []Line
		lines = append(lines, Line{Content: indentStr + fmt.Sprintf("%q: {", key), Path: copyPath(path)})

		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i, childKey := range keys {
			childPath := append(copyPath(path), spec.Name(childKey))
			childComma := ","
			if i == len(keys)-1 {
				childComma = ""
			}
			childLines := generateKeyValueLines(childKey, v[childKey], childPath, indent+1, childComma)
			lines = append(lines, childLines...)
		}

		lines = append(lines, Line{Content: indentStr + "}" + comma, Path: copyPath(path)})
		return lines

	case []interface{}:
		var lines []Line
		lines = append(lines, Line{Content: indentStr + fmt.Sprintf("%q: [", key), Path: copyPath(path)})

		for i, elem := range v {
			childPath := append(copyPath(path), spec.Index(i))
			childComma := ","
			if i == len(v)-1 {
				childComma = ""
			}

			childLines := generateLines(elem, childPath, indent+1)
			if len(childLines) > 0 {
				lastIdx := len(childLines) - 1
				childLines[lastIdx].Content += childComma
			}
			lines = append(lines, childLines...)
		}

		lines = append(lines, Line{Content: indentStr + "]" + comma, Path: copyPath(path)})
		return lines

	default:
		content := indentStr + fmt.Sprintf("%q: %s%s", key, formatPrimitive(value), comma)
		return []Line{{Content: content, Path: copyPath(path)}}
	}
}

func formatPrimitive(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%v", v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatOutput formats lines with diff markers
func formatOutput(lines []Line, diffPaths map[string]bool) string {
	var sb strings.Builder

	for _, line := range lines {
		prefix := " "
		if shouldMarkAsDiff(line.Path, diffPaths) {
			prefix = "-"
		}
		sb.WriteString(prefix)
		sb.WriteString(line.Content)
		sb.WriteString("\n")
	}

	return sb.String()
}

// shouldMarkAsDiff checks if a line should be marked as diff
func shouldMarkAsDiff(path spec.NormalizedPath, diffPaths map[string]bool) bool {
	pathStr := path.String()

	// Exact match
	if diffPaths[pathStr] {
		return true
	}

	// Check if this path is a child of a diff path
	for diffPath := range diffPaths {
		if strings.HasPrefix(pathStr, diffPath) && len(pathStr) > len(diffPath) {
			return true
		}
	}
	return false
}
