package test

import (
	"testing"

	"resttracefuzzer/pkg/utils"

	"github.com/stretchr/testify/assert"
)

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
	}

	for _, test := range tests {
		result := utils.MatchVariableNames(test.name1, test.name2)
		assert.Equal(t, test.expected, result)
	}
}

func TestCalculateSimilarity(t *testing.T) {
	calculator := utils.NewLevenshteinSimilarityCalculator()

	tests := []struct {
		str1     string
		str2     string
		expected float64
	}{
		{"kitten", "sitting", 0.5714285714285714},
		{"flaw", "lawn", 0.75},
		{"intention", "execution", 0.5555555555555556},
		{"", "", 1.0},
		{"", "nonempty", 0.0},
	}

	for _, test := range tests {
		result := calculator.CalculateSimilarity(test.str1, test.str2)
		assert.InDelta(t, test.expected, result, 0.0001)
	}
}

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

func TestExtractLastSegment(t *testing.T) {
	tests := []struct {
		input      string
		delimiters string
		expected   string
	}{
		{"/oteldemo.CartService/GetCart", "/.", "GetCart"},
		{"com.example.ClassName.methodName", ".", "methodName"},
		{"path/to/file.txt", "/", "file.txt"},
		{"path-to-file.txt", "-", "file.txt"},
	}

	for _, test := range tests {
		result := utils.ExtractLastSegment(test.input, test.delimiters)
		assert.Equal(t, test.expected, result)
	}
}
