package main

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {

	testCases := []struct {
		Name     string
		Input    string
		Expected []string
	}{
		{
			"no delimiters",
			"localhost:8080",
			[]string{"localhost:8080"},
		},
		{
			"only dots",
			"test.example.com",
			[]string{"test", "example", "com"},
		},
		{
			"only hyphens",
			"eu-east-1",
			[]string{"eu", "east", "1"},
		},
		{
			"mix dots-hyphens",
			"eu-east-1.example.com",
			[]string{"eu", "east", "1", "example", "com"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := Tokenize(tc.Input)
			if !reflect.DeepEqual(tc.Expected, actual) {
				t.Errorf("expected %v but got %v", tc.Expected, actual)
			}
		})
	}
}

func TestMatchSlots(t *testing.T) {
	slots := map[string]map[string]bool{
		"env": {
			"dev": true, "uat": true, "stg": true, "prd": true,
		},
		"locality": {
			"local": true, "remote": true,
		},
	}

	testCases := []struct {
		Name             string
		Input            string
		WantMatches      []string
		WantIndices      []int
		WantCombinations int
	}{
		{
			"dev-local",
			"dev-local.example.com",
			[]string{"env", "locality"},
			[]int{0, 1},
			8,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tokens := Tokenize(tc.Input)
			gotMatches, gotIndices, gotCombinations := MatchSlots(tokens, slots)

			if !reflect.DeepEqual(tc.WantMatches, gotMatches) {
				t.Errorf("want %v but got %v", tc.WantMatches, gotMatches)
			}

			if !reflect.DeepEqual(tc.WantIndices, gotIndices) {
				t.Errorf("want %v indices but got %d", tc.WantIndices, gotIndices)
			}

			if tc.WantCombinations != gotCombinations {
				t.Errorf("want %v combinations but got %d", tc.WantCombinations, gotCombinations)
			}
		})
	}
}
