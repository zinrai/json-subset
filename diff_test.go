package main

import (
	"strings"
	"testing"
)

func TestSubsetCheck(t *testing.T) {
	tests := []struct {
		name       string
		subset     interface{}
		superset   interface{}
		wantSubset bool
	}{
		{
			name:       "object subset",
			subset:     map[string]interface{}{"a": float64(1)},
			superset:   map[string]interface{}{"a": float64(1), "b": float64(2)},
			wantSubset: true,
		},
		{
			name:       "object not subset - missing key",
			subset:     map[string]interface{}{"a": float64(1), "c": float64(3)},
			superset:   map[string]interface{}{"a": float64(1), "b": float64(2)},
			wantSubset: false,
		},
		{
			name:       "array subset - set mode ignores order",
			subset:     []interface{}{float64(2), float64(1)},
			superset:   []interface{}{float64(1), float64(2), float64(3)},
			wantSubset: true,
		},
		{
			name:       "array not subset - element not found",
			subset:     []interface{}{float64(1), float64(4)},
			superset:   []interface{}{float64(1), float64(2), float64(3)},
			wantSubset: false,
		},
		{
			name:       "nested object subset",
			subset:     map[string]interface{}{"user": map[string]interface{}{"name": "alice"}},
			superset:   map[string]interface{}{"user": map[string]interface{}{"name": "alice", "age": float64(30)}},
			wantSubset: true,
		},
		{
			name:       "nested object not subset",
			subset:     map[string]interface{}{"user": map[string]interface{}{"name": "alice", "email": "alice@example.com"}},
			superset:   map[string]interface{}{"user": map[string]interface{}{"name": "alice"}},
			wantSubset: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := checkSubsetWithDiffs(tt.subset, tt.superset)
			if got != tt.wantSubset {
				t.Errorf("checkSubsetWithDiffs() = %v, want %v", got, tt.wantSubset)
			}
		})
	}
}

func TestDiffOutput(t *testing.T) {
	subset := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "alice",
			"email": "alice@example.com",
		},
	}
	superset := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "alice",
		},
	}

	_, diffs := checkSubsetWithDiffs(subset, superset)
	output := FormatDiffOutput(subset, diffs)

	if !strings.Contains(output, "-") {
		t.Error("diff output should contain '-' marker")
	}
	if !strings.Contains(output, "email") {
		t.Error("diff output should contain missing key 'email'")
	}
}
