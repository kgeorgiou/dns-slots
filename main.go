package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type slots map[string]map[string]bool

func main() {

	slots, err := readSlotsFile("slots.yml")
	if err != nil {
		panic("oooops.")
	}

	s := "dev-local.example.com"
	tokens := Tokenize(s)
	matches, indices, _ := MatchSlots(tokens, slots)
	Spin(tokens, matches, indices, slots, map[string]bool{})
}

func readSlotsFile(filename string) (slots, error) {
	yamlMap := map[string][]string{}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(content, yamlMap)
	if err != nil {
		return nil, err
	}

	res := map[string]map[string]bool{}

	for k, v := range yamlMap {
		mm := map[string]bool{}
		for _, vv := range v {
			mm[vv] = true
		}
		res[k] = mm
	}

	return res, nil
}

// Tokenize splits a DNS entry on "." and "-"
func Tokenize(s string) []string {
	result := []string{}

	dot := strings.Split(s, ".")
	for _, d := range dot {
		hyphen := strings.Split(d, "-")
		result = append(result, hyphen...)
	}

	return result
}

// MatchSlots .
func MatchSlots(tokens []string, slots slots) (matches []string, indices []int, combinations int) {
	matches = []string{}
	indices = []int{}
	combinations = 1

	for tokenIdx, token := range tokens {
		for slotName, slot := range slots {
			if _, ok := slot[token]; ok {
				matches = append(matches, slotName)
				indices = append(indices, tokenIdx)
				combinations *= len(slot)
			}
		}
	}

	return matches, indices, combinations
}

// Spin 2 Win!
func Spin(tokens []string, matches []string, indices []int, slots slots, seen map[string]bool) {
	outcome := strings.Join(tokens, "-")
	if _, ok := seen[outcome]; ok {
		return
	}

	seen[outcome] = true
	fmt.Println(outcome)

	if len(matches) == 0 || len(indices) == 0 {
		return
	}

	for i := 0; i < len(indices); i++ {

		slot, ok := slots[matches[i]]
		if !ok {
			panic("oops.")
		}

		for v := range slot {
			newTokens := make([]string, len(tokens))
			copy(newTokens, tokens)

			newTokens[indices[i]] = v

			Spin(newTokens, matches[1:], indices[1:], slots, seen)
		}
	}

}
