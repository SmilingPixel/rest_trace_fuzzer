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


// SimilarityCalculator is an interface that defines a method to calculate the similarity
// between two strings. The result is a number between 0 and 1.
type SimilarityCalculator interface {
	CalculateSimilarity(str1, str2 string) float64
}

// LevenshteinSimilarityCalculator implements the calculation of Levenshtein distance and similarity.
type LevenshteinSimilarityCalculator struct{}

// CalculateSimilarity calculates the similarity between two strings based on normalized Levenshtein distance.
func (l LevenshteinSimilarityCalculator) CalculateSimilarity(str1, str2 string) float64 {
	distance := levenshteinDistance(str1, str2)
	maxLength := max(len(str1), len(str2))
	if maxLength == 0 {
		return 1.0 // Both strings are empty
	}
	// Calculate normalized similarity score
	return 1.0 - float64(distance)/float64(maxLength)
}

// levenshteinDistance computes the Levenshtein distance between two strings.
func levenshteinDistance(a, b string) int {
	lenA := len(a)
	lenB := len(b)
	if lenA == 0 {
		return lenB
	}
	if lenB == 0 {
		return lenA
	}

	// Create a matrix to store distances
	matrix := make([][]int, lenA+1)
	for i := range matrix {
		matrix[i] = make([]int, lenB+1)
	}
	// Initialize the matrix
	for i := 0; i <= lenA; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= lenB; j++ {
		matrix[0][j] = j
	}
	// Compute the distance
	for i := 1; i <= lenA; i++ {
		for j := 1; j <= lenB; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			// Calculate the minimum of deletion, insertion, or substitution
			matrix[i][j] = min(
				matrix[i-1][j]+1,   // Deletion
				matrix[i][j-1]+1,   // Insertion
				matrix[i-1][j-1]+cost, // Substitution
			)
		}
	}
	return matrix[lenA][lenB]
}
