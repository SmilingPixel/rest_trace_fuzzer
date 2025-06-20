package test

import (
	"testing"

	"resttracefuzzer/pkg/utils"

	"github.com/stretchr/testify/assert"
)

// TestSplitIntoWords tests the SplitIntoWords function from the utils package.
// It verifies that various input strings are correctly split into words.
func TestSplitIntoWords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"petStore", []string{"pet", "store"}},
		{"pet_store", []string{"pet", "store"}},
		{"PetStore", []string{"pet", "store"}},
		{"petStoreExample", []string{"pet", "store", "example"}},
		{"pet_store_example", []string{"pet", "store", "example"}},
	}

	for _, test := range tests {
		result := utils.SplitIntoWords(test.input)
		assert.Equal(t, test.expected, result)
	}
}

// TestMatchVariableNames tests the MatchVariableNames function from the utils package.
// It verifies that various pairs of variable names are correctly matched based on their similarity.
// We use the IdentitySimilarityCalculator to calculate the similarity between words, and set the threshold to 1.0.
func TestMatchVariableNames(t *testing.T) {
	tests := []struct {
		name1    string
		name2    string
		expected bool
	}{
		{"petStore", "pet_store", true},
		{"PetStore", "pet_store", true},
		{"petStoreExample", "pet_store_example", true},
		{"petStore", "petStoreExample", false},
		{"pet_store", "pet_store_example", false},
		{"id", "id", false},               // Common field names should not match
		{"createdAt", "updatedAt", false}, // Common field names should not match
		{"petId", "pet_id", true},         // Non-common prefix should still match
	}

	similarityCalculator := utils.NewIdentitySimilarityCalculator()
	threshold := 1.0
	for _, test := range tests {
		result := utils.MatchVariableNames(test.name1, test.name2, similarityCalculator, threshold)
		assert.Equal(t, test.expected, result)
	}
}

// TestIdentitySimilarityCalculator tests the CalculateSimilarity function of the IdentitySimilarityCalculator.
// It verifies that the similarity between various pairs of strings is correctly calculated based on identity.
func TestIdentitySimilarityCalculator(t *testing.T) {
	calculator := utils.NewIdentitySimilarityCalculator()

	tests := []struct {
		str1     string
		str2     string
		expected float64
	}{
		{"identical", "identical", 1.0},
		{"different", "identical", 0.0},
		{"", "", 1.0},
		{"", "nonempty", 0.0},
	}

	for _, test := range tests {
		result := calculator.CalculateSimilarity(test.str1, test.str2)
		assert.Equal(t, test.expected, result)
	}
}

// TestLevenshteinSimilarityCalculator tests the CalculateSimilarity function of the LevenshteinSimilarityCalculator.
// It verifies that the similarity between various pairs of strings is correctly calculated using the Levenshtein algorithm.
func TestLevenshteinSimilarityCalculator(t *testing.T) {
	testCalculateSimilarity(t, utils.NewLevenshteinSimilarityCalculator())
}

// TestJaccardSimilarityCalculator tests the CalculateSimilarity function of the JaccardSimilarityCalculator.
// It verifies that the similarity between various pairs of strings is correctly calculated using the Jaccard similarity algorithm.
func TestJaccardSimilarityCalculator(t *testing.T) {
	testCalculateSimilarity(t, utils.NewJaccardSimilarityCalculator())
}

// testCalculateSimilarity tests the CalculateSimilarity function from the utils package.
// It verifies that the similarity between various pairs of strings is correctly calculated using the Levenshtein algorithm.
func testCalculateSimilarity(t *testing.T, calculator utils.SimilarityCalculator) {

	tests := []struct {
		str1     string
		str2     string
		expected float64
	}{
		{"kitten", "sitting", 0.5714285714285714},
		{"flaw", "lawn", 0.5},
		{"intention", "execution", 0.4444444444444444},
		{"", "", 1.0},
		{"", "nonempty", 0.0},
	}

	for _, test := range tests {
		result := calculator.CalculateSimilarity(test.str1, test.str2)
		assert.InDelta(t, test.expected, result, 0.0001)
	}
}

// TestConvertToStandardCase tests the ConvertToStandardCase function from the utils package.
// It verifies that various input strings are correctly converted to a standard case format.
func TestConvertToStandardCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"petStore", "petstore"},
		{"pet_store", "petstore"},
		{"pet-store", "petstore"},
		{"PetStoreExample", "petstoreexample"},
		{"pet_store_example", "petstoreexample"},
	}

	for _, test := range tests {
		result := utils.ConvertToStandardCase(test.input)
		assert.Equal(t, test.expected, result)
	}
}

// TestExtractLastSegment tests the ExtractLastSegment function from the utils package.
// It verifies that the last segment of various input strings is correctly extracted based on the provided delimiters.
func TestExtractLastSegment(t *testing.T) {
	tests := []struct {
		input      string
		delimiters []string
		expected   string
	}{
		{"/oteldemo.CartService/GetCart", []string{"/", "."}, "GetCart"},
		{"com.example.ClassName.methodName", []string{".", " "}, "methodName"},
		{"path/to/file.txt", []string{"/"}, "file.txt"},
		{"path-to-file.txt", []string{"-"}, "file.txt"},
	}

	for _, test := range tests {
		result := utils.ExtractLastSegment(test.input, test.delimiters)
		assert.Equal(t, test.expected, result)
	}
}

// TestSplitByDelimiters tests the SplitByDelimiters function from the utils package.
// It verifies that various input strings are correctly split into segments based on the provided delimiters.
func TestSplitByDelimiters(t *testing.T) {
	tests := []struct {
		input      string
		delimiters []string
		expected   []string
	}{
		{"/oteldemo.CartService/GetCart", []string{"/", "."}, []string{"", "oteldemo", "CartService", "GetCart"}},
		{"com.example.ClassName.methodName", []string{".", " "}, []string{"com", "example", "ClassName", "methodName"}},
		{"path/to/file.txt", []string{"/"}, []string{"path", "to", "file.txt"}},
		{"path-to-file.txt", []string{"-"}, []string{"path", "to", "file.txt"}},
		{"a,b;c|d", []string{",", ";", "|"}, []string{"a", "b", "c", "d"}},
		{"singleword", []string{"/"}, []string{"singleword"}},
		{"", []string{"/"}, []string{""}},
	}

	for _, test := range tests {
		result := utils.SplitByDelimiters(test.input, test.delimiters)
		assert.Equal(t, test.expected, result)
	}
}
