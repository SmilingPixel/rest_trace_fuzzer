// Package utils provides utility functions for natural language processing (NLP)
// and string manipulation. These utilities are designed to assist with tasks such as
// splitting variable names into words, comparing variable names for similarity, and
// converting strings between different casing styles. The package also includes
// implementations of various similarity calculators, such as Levenshtein and Jaccard,
// to support flexible and robust string comparison.
package utils

import (
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"
)

// SplitIntoWords takes a variable name formatted in either camelCase or snake_case and
// breaks it down into its constituent words. It returns a slice of those words all in lowercase.
//
// For example:
//   - "petStore" would be split into ["pet", "store"]
//   - "pet_store" would be split into ["pet", "store"]
func SplitIntoWords(name string) []string {
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
// Two variable names are considered a match if, after splitting them into words, they yield "similiar" slices.
// Comparison is case-insensitive, and underscores are treated as word boundaries.
//
// In sprcific, we do the following:
//  1. Convert arrays to singular form using GetArrayElementNameHeuristic.
//  2. Split the variable names into words using SplitIntoWords. For example, "petStore" -> ["pet", "store"].
//  3. Remove some common field names, e.g., "id". If the words list is empty after this step, we return false.
//  4. "Ignore" the prefixes, truncating the longer one if necessary. For example, if name1 and name2 are ["example", "pet", "store"] and ["app", "store"], respectively, we would compare ["pet", "store"] and ["app", "store"].
//  5. Compare the words in the two slices. If the similiarity reaches a certain threshold, we consider the variable names a match. We use [resttracefuzzer/pkg/utils.SimilarityCalculator] to calculate the similarity.
//  6. Return true if the average similarity is above the threshold, and false otherwise.
//
// Parameters:
//   - name1: The first variable name to compare.
//   - name2: The second variable name to compare.
//   - similarityCalculator: A similarity calculator to use for comparing the words in the two slices. If not provided (nil), the identity similarity calculator is used.
//   - threshold: The threshold above or equal to which the average similarity is considered a match.
func MatchVariableNames(name1, name2 string, similarityCalculator SimilarityCalculator, threshold float64) bool {
	name1 = GetSingularFormNameHeuristic(name1)
	name2 = GetSingularFormNameHeuristic(name2)

	words1 := SplitIntoWords(name1)
	words2 := SplitIntoWords(name2)

	// Remove common field names
	filteredWords1 := make([]string, 0)
	filteredWords2 := make([]string, 0)
	for _, word := range words1 {
		if !IsCommonFieldName(word) {
			filteredWords1 = append(filteredWords1, word)
		}
	}
	for _, word := range words2 {
		if !IsCommonFieldName(word) {
			filteredWords2 = append(filteredWords2, word)
		}
	}
	words1 = filteredWords1
	words2 = filteredWords2
	// If either list is empty after filtering, return false
	if len(words1) == 0 || len(words2) == 0 {
		log.Debug().Msgf("[MatchVariableNames] Filtered words are empty: %v, %v", words1, words2)
		return false
	}

	// Truncate the longer slice if necessary
	if len(words1) != len(words2) {
		truncatedLength := min(len(words1), len(words2))
		words1 = words1[len(words1)-truncatedLength:]
		words2 = words2[len(words2)-truncatedLength:]
	}

	// Calculate the average similarity between the two word slices
	if similarityCalculator == nil {
		log.Warn().Msg("[MatchVariableNames] No similarity calculator provided. Using identity similarity calculator.")
		similarityCalculator = NewIdentitySimilarityCalculator()
	}
	similaritySum := 0.0
	for i := range words1 {
		similaritySum += similarityCalculator.CalculateSimilarity(words1[i], words2[i])
	}
	averageSimilarity := similaritySum / float64(len(words1))
	return averageSimilarity >= threshold
}

// SimilarityCalculator is an interface that defines a method to calculate the similarity
// between two strings. The result is a number between 0 and 1.
type SimilarityCalculator interface {
	CalculateSimilarity(str1, str2 string) float64
}

// IdentitySimilarityCalculator implements the calculation of similarity based on identity.
// It returns 1.0 if the two strings are equal and 0.0 otherwise.
type IdentitySimilarityCalculator struct{}

// CalculateSimilarity calculates the similarity between two strings based on identity.
func NewIdentitySimilarityCalculator() *IdentitySimilarityCalculator {
	return &IdentitySimilarityCalculator{}
}

// CalculateSimilarity calculates the similarity between two strings based on identity.
func (i *IdentitySimilarityCalculator) CalculateSimilarity(str1, str2 string) float64 {
	if str1 == str2 {
		return 1.0
	}
	return 0.0
}

// LevenshteinSimilarityCalculator implements the calculation of Levenshtein distance and similarity.
type LevenshteinSimilarityCalculator struct{}

// NewLevenshteinSimilarityCalculator creates a new LevenshteinSimilarityCalculator.
func NewLevenshteinSimilarityCalculator() *LevenshteinSimilarityCalculator {
	return &LevenshteinSimilarityCalculator{}
}

// CalculateSimilarity calculates the similarity between two strings based on normalized Levenshtein distance.
func (l *LevenshteinSimilarityCalculator) CalculateSimilarity(str1, str2 string) float64 {
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
				matrix[i-1][j]+1,      // Deletion
				matrix[i][j-1]+1,      // Insertion
				matrix[i-1][j-1]+cost, // Substitution
			)
		}
	}
	return matrix[lenA][lenB]
}

// JaccardSimilarityCalculator implements the calculation of Jaccard similarity.
type JaccardSimilarityCalculator struct{}

// NewJaccardSimilarityCalculator creates a new JaccardSimilarityCalculator.
func NewJaccardSimilarityCalculator() *JaccardSimilarityCalculator {
	return &JaccardSimilarityCalculator{}
}

// CalculateSimilarity calculates the similarity between two strings based on Jaccard similarity.
func (j *JaccardSimilarityCalculator) CalculateSimilarity(str1, str2 string) float64 {
	set1 := make(map[rune]struct{})
	set2 := make(map[rune]struct{})

	for _, r := range str1 {
		set1[r] = struct{}{}
	}
	for _, r := range str2 {
		set2[r] = struct{}{}
	}

	intersectionSize := 0
	for r := range set1 {
		if _, exists := set2[r]; exists {
			intersectionSize++
		}
	}

	unionSize := len(set1) + len(set2) - intersectionSize

	if unionSize == 0 {
		return 1.0 // Both strings are empty
	}

	return float64(intersectionSize) / float64(unionSize)
}

// ConvertToStandardCase transforms a variable's name from various casing styles
// (e.g., camelCase, snake_case, snake-case) into a standardized lowercase format
// without any separators. This function is useful for ensuring uniform processing
// and comparison of variable names across different conventions.
// For example, 'petStore', 'pet_store', and 'pet-store' would all be converted to 'petstore'.
func ConvertToStandardCase(input string) string {
	// Remove underscores and hyphens
	removedSeparators := strings.ReplaceAll(strings.ReplaceAll(input, "_", ""), "-", "")

	// Convert to lowercase
	lowercaseResult := strings.ToLower(removedSeparators)

	return lowercaseResult
}

// ExtractLastSegment extracts the last segment from a string using customizable delimiters.
// Delimiters are provided as a string, where each character is considered a distinct delimiter.
//
// Parameters:
// - input: A string that includes segments separated by various delimiters.
// - delimiters: A list of string where each string is a delimiter to use for splitting the input.
//
// Returns:
// - A string representing the last segment in the input after splitting by the specified delimiters.
//
// Example:
//
//	input: "/oteldemo.CartService/GetCart", delimiters: []string{"/", "."}
//	output: "GetCart"
func ExtractLastSegment(input string, delimiters []string) string {
	lastSegment := input

	// Loop through each delimiter string and split/reduce the input accordingly
	for _, delimiter := range delimiters {
		segments := strings.Split(lastSegment, delimiter)
		lastSegment = segments[len(segments)-1]
	}

	return lastSegment
}

// SplitByDelimiters splits a string into segments based on a list of delimiters.
//
// Parameters:
// - input: The string to be split.
// - delimiters: A slice of strings, where each string is treated as a delimiter.
//
// Returns:
// - A slice of strings representing the segments of the input string after splitting by the specified delimiters.
//
// Example:
//
//	input: "/oteldemo.CartService/GetCart", delimiters: []string{"/", "."}
//	output: []string{"", "oteldemo", "CartService", "GetCart"}
func SplitByDelimiters(input string, delimiters []string) []string {
	segments := []string{input}

	// Loop through each delimiter string and split the input accordingly
	for _, delimiter := range delimiters {
		var newSegments []string
		for _, segment := range segments {
			newSegments = append(newSegments, strings.Split(segment, delimiter)...)
		}
		segments = newSegments
	}

	return segments
}

// GetSingularFormNameHeuristic returns a singular form of an array name or name in plural form by applying simple heuristics.
//   - At a basic level, it removes the trailing 's' or 'es' character(s) from the name if present.
//   - If the name ends with 'List', 'Array', or 'Collection', it removes the suffix.
func GetSingularFormNameHeuristic(name string) string {
	if name == "" {
		return name
	}
	// handle "es" before "s" to avoid incorrect removal
	if strings.HasSuffix(name, "es") {
		if strings.HasSuffix(name, "es") {
			return strings.TrimSuffix(name, "es")
		}
		return strings.TrimSuffix(name, "s")
	}
	if strings.HasSuffix(name, "s") {
		return strings.TrimSuffix(name, "s")
	}
	if strings.HasSuffix(name, "List") {
		return strings.TrimSuffix(name, "List")
	}
	if strings.HasSuffix(name, "Array") {
		return strings.TrimSuffix(name, "Array")
	}
	if strings.HasSuffix(name, "Collection") {
		return strings.TrimSuffix(name, "Collection")
	}
	return name
}
