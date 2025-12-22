package main

import (
	"strings"
	"testing"
)

func TestCheckSubset(t *testing.T) {
	tests := []struct {
		name       string
		subset     interface{}
		superset   interface{}
		wantSubset bool
		wantInDiff string
	}{
		{
			name:       "identical objects",
			subset:     map[string]interface{}{"a": float64(1), "b": float64(2)},
			superset:   map[string]interface{}{"a": float64(1), "b": float64(2)},
			wantSubset: true,
		},
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
			wantInDiff: "missing key",
		},
		{
			name:       "object not subset - value mismatch",
			subset:     map[string]interface{}{"a": float64(999)},
			superset:   map[string]interface{}{"a": float64(1), "b": float64(2)},
			wantSubset: false,
			wantInDiff: "value mismatch",
		},
		{
			name:       "nested object subset",
			subset:     map[string]interface{}{"user": map[string]interface{}{"name": "alice"}},
			superset:   map[string]interface{}{"user": map[string]interface{}{"name": "alice", "age": float64(30)}},
			wantSubset: true,
		},
		{
			name:       "nested object not subset",
			subset:     map[string]interface{}{"user": map[string]interface{}{"name": "alice", "role": "admin"}},
			superset:   map[string]interface{}{"user": map[string]interface{}{"name": "alice", "age": float64(30)}},
			wantSubset: false,
			wantInDiff: "missing key",
		},
		{
			name:       "array subset (set mode) - same order",
			subset:     []interface{}{float64(1), float64(2)},
			superset:   []interface{}{float64(1), float64(2), float64(3)},
			wantSubset: true,
		},
		{
			name:       "array subset (set mode) - different order",
			subset:     []interface{}{float64(2), float64(1)},
			superset:   []interface{}{float64(1), float64(2), float64(3)},
			wantSubset: true,
		},
		{
			name:       "array not subset - missing element",
			subset:     []interface{}{float64(1), float64(4)},
			superset:   []interface{}{float64(1), float64(2), float64(3)},
			wantSubset: false,
			wantInDiff: "element not found",
		},
		{
			name: "array of objects subset (set mode)",
			subset: []interface{}{
				map[string]interface{}{"id": float64(1)},
				map[string]interface{}{"id": float64(2)},
			},
			superset: []interface{}{
				map[string]interface{}{"id": float64(2)},
				map[string]interface{}{"id": float64(1)},
				map[string]interface{}{"id": float64(3)},
			},
			wantSubset: true,
		},
		{
			name: "array of objects not subset",
			subset: []interface{}{
				map[string]interface{}{"id": float64(1)},
				map[string]interface{}{"id": float64(4)},
			},
			superset: []interface{}{
				map[string]interface{}{"id": float64(1)},
				map[string]interface{}{"id": float64(2)},
			},
			wantSubset: false,
			wantInDiff: "element not found",
		},
		{
			name: "array of objects with extra fields subset",
			subset: []interface{}{
				map[string]interface{}{"key": "a", "value": "1"},
			},
			superset: []interface{}{
				map[string]interface{}{"key": "a", "value": "1", "metadata": "info"},
			},
			wantSubset: true,
		},
		{
			name:       "empty object is subset of any object",
			subset:     map[string]interface{}{},
			superset:   map[string]interface{}{"a": float64(1), "b": float64(2)},
			wantSubset: true,
		},
		{
			name:       "empty array is subset of any array",
			subset:     []interface{}{},
			superset:   []interface{}{float64(1), float64(2)},
			wantSubset: true,
		},
		{
			name:       "null values",
			subset:     map[string]interface{}{"a": nil},
			superset:   map[string]interface{}{"a": nil, "b": float64(2)},
			wantSubset: true,
		},
		{
			name:       "type mismatch",
			subset:     map[string]interface{}{"a": "1"},
			superset:   map[string]interface{}{"a": float64(1)},
			wantSubset: false,
			wantInDiff: "type mismatch",
		},
		{
			name:       "string values",
			subset:     map[string]interface{}{"name": "alice"},
			superset:   map[string]interface{}{"name": "alice", "role": "admin"},
			wantSubset: true,
		},
		{
			name:       "boolean values",
			subset:     map[string]interface{}{"active": true},
			superset:   map[string]interface{}{"active": true, "verified": false},
			wantSubset: true,
		},
		{
			name:       "deeply nested subset",
			subset:     map[string]interface{}{"a": map[string]interface{}{"b": map[string]interface{}{"c": float64(1)}}},
			superset:   map[string]interface{}{"a": map[string]interface{}{"b": map[string]interface{}{"c": float64(1), "d": float64(2)}}},
			wantSubset: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSubset, gotDiffs := checkSubset(tt.subset, tt.superset)

			if gotSubset != tt.wantSubset {
				t.Errorf("checkSubset() = %v, want %v\ndiffs: %v", gotSubset, tt.wantSubset, gotDiffs)
			}

			if !gotSubset && tt.wantInDiff != "" {
				found := false
				for _, diff := range gotDiffs {
					if strings.Contains(diff, tt.wantInDiff) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected diff containing %q, got: %v", tt.wantInDiff, gotDiffs)
				}
			}
		})
	}
}

func TestCheckSubsetPath(t *testing.T) {
	tests := []struct {
		name     string
		subset   interface{}
		superset interface{}
		wantPath string
	}{
		{
			name:     "reports correct path for nested mismatch",
			subset:   map[string]interface{}{"user": map[string]interface{}{"profile": map[string]interface{}{"age": float64(99)}}},
			superset: map[string]interface{}{"user": map[string]interface{}{"profile": map[string]interface{}{"age": float64(30)}}},
			wantPath: "$.user.profile.age",
		},
		{
			name:     "reports correct path for array element",
			subset:   []interface{}{float64(1), float64(999)},
			superset: []interface{}{float64(1), float64(2)},
			wantPath: "$[1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, diffs := checkSubset(tt.subset, tt.superset)

			found := false
			for _, diff := range diffs {
				if strings.Contains(diff, tt.wantPath) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected path %q in diffs, got: %v", tt.wantPath, diffs)
			}
		})
	}
}
