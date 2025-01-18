package utils

import (
	"strings"
	"unicode"
)

// splitIntoWords takes a variable name formatted in either camelCase or snake_case and
// breaks it down into its constituent words. It returns a slice of those words all in lowercase.
//
// For example:
//   - "petStore" would be split into ["pet", "store"]
//   - "pet_store" would be split into ["pet", "store"]
func splitIntoWords(name string) []string {
	var words []string
	var currentWord strings.Builder

	for _, r := range name {
		if unicode.IsUpper(r) {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
			currentWord.WriteRune(unicode.ToLower(r))
		} else if r == '_' {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		} else {
			currentWord.WriteRune(r)
		}
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}


// matchVariableNames determines if two variable names represent the same concept by
// comparing their respective word slices. It returns true if they match and false otherwise.
//
// Two variable names are considered a match if, after splitting them into words, they yield identical slices.
// Comparison is case-insensitive, and underscores are treated as word boundaries.
func MatchVariableNames(name1, name2 string) bool {
	words1 := splitIntoWords(name1)
	words2 := splitIntoWords(name2)

	if len(words1) != len(words2) {
		return false
	}

	for i := range words1 {
		if words1[i] != words2[i] {
			return false
		}
	}

	return true
}
