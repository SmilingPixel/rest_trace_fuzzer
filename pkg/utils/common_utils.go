// Package utils provides common utility functions for random data generation,
// type conversion, and string manipulation. These utilities are designed to
// support various operations such as generating random strings, mutating data,
// converting types to 64-bit equivalents, and handling edge cases for primitive types.
// The package also includes functions for formatting service names and generating
// random or default values for specific types.
package utils

import (
	"encoding/base64"
	"encoding/hex"
	"math"
	"math/rand/v2"
	"reflect"

	"github.com/rs/zerolog/log"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandStringBytes generates a random string of length n.
// See https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.IntN(len(letterBytes))]
	}
	return string(b)
}

// mutateRandBytes mutates a byte slice by randomly changing some bytes.
// It accepts three parameters:
// - b: the byte slice to mutate
// - mutationProbability: the probability of mutating each byte
// - maxMutations: the maximum number of bytes to mutate
// The function returns nothing.
func mutateRandBytes(b []byte, mutationProbability float64, maxMutations int) {
	if maxMutations == 0 {
		return
	}
	if mutationProbability < 0 || mutationProbability > 1 {
		log.Warn().Msgf("[mutateRandBytes] Invalid mutation probability: %f, skip mutation", mutationProbability)
		return
	}
	mutations := 0
	for i := range b {
		if rand.Float64() < mutationProbability && mutations < maxMutations {
			b[i] = letterBytes[rand.IntN(len(letterBytes))]
			mutations++
		}
	}
}

// MutateRandBytesForString mutates a string by converting it to a byte slice, mutating it, and converting it back to a string.
// It accepts three parameters:
// - s: the string to mutate
// - mutationProbability: the probability of mutating each byte
// - maxMutations: the maximum number of bytes to mutate
// The function returns the mutated string.
func MutateRandBytesForString(s string, mutationProbability float64, maxMutations int) string {
	b := []byte(s)
	mutateRandBytes(b, mutationProbability, maxMutations)
	return string(b)
}

// ConvertIntTo64BitType converts a given integer to its 64-bit equivalent type.
// It accepts an integer as input and returns an int64 as output.
// The function handles various integer types and converts them to int64.
// If the input type is not recognized, it logs a warning and returns 0.
//
// Supported conversions:
// - int, int8, int16, int32 -> int64
// - int64 -> no conversion
//
// Example usage:
//
//	var num = 42
//	result := ConvertIntTo64BitType(num) // result will be int64(42)
//
// Parameters:
// - num: the integer to be converted, of type any
//
// Returns:
// - the converted integer as an int64, or 0 if the input type is not recognized.
func ConvertIntTo64BitType(num any) int64 {
	switch num2 := num.(type) {
	case int:
		return int64(num2)
	case int8:
		return int64(num2)
	case int16:
		return int64(num2)
	case int32:
		return int64(num2)
	case int64:
		return num2
	default:
		log.Warn().Msgf("[ConvertIntTo64BitType] Unknown type: %T", num)
		return 0
	}
}

// ConvertFloatTo64BitType converts a given float to its 64-bit equivalent type.
// It accepts an float number as input and returns a float64 as output.
// The function handles various float types and converts them to float64.
// If the input type is not recognized, it logs a warning and returns 0.0.
//
// Supported conversions:
// - float32 -> float64
// - float64 -> no conversion
//
// Example usage:
//
//	var num = float32(42.0)
//	result := ConvertFloatTo64BitType(num) // result will be float64(42.0)
//
// Parameters:
// - num: the float to be converted, of type any
//
// Returns:
// - the converted float as a float64, or 0.0 if the input type is not recognized.
func ConvertFloatTo64BitType(num any) float64 {
	switch num2 := num.(type) {
	case float32:
		return float64(num2)
	case float64:
		return num2
	default:
		log.Warn().Msgf("[ConvertFloatTo64BitType] Unknown type: %T", num)
		return 0.0
	}
}

// NormInt64 generates a normally distributed random int64 value.
// It accepts two parameters:
// - mean: the mean value of the distribution
// - stdDev: the standard deviation of the distribution
// The function returns a normally distributed random int64 value.
func NormInt64(mean, stdDev int64) int64 {
	return int64(math.Round(rand.NormFloat64()*float64(stdDev) + float64(mean)))
}

// DefaultValueForPrimitiveTypeKind returns the default value for a given primitive type kind.
// A primitive type is a basic data type that is not composed of other types, such as integers, floats, strings, and booleans.
// We assume integer types are int64, and float types are float64.
func DefaultValueForPrimitiveTypeKind(kind reflect.Kind) any {
	switch kind {
	case reflect.Int64:
		return int64(114514)
	case reflect.Float64:
		return 114.514
	case reflect.Bool:
		return true
	case reflect.String:
		return "114-514"
	default:
		log.Warn().Msgf("[DefaultValueForPrimitiveTypeKind] Unsupported kind: %v", kind)
		return nil
	}
}

// RandomValueForPrimitiveTypeKind generates a random value for a given primitive type.
// A primitive type is a basic data type that is not composed of other types, such as integers, floats, strings, and booleans.
// We assume integer types are int64, and float types are float64.
func RandomValueForPrimitiveTypeKind(kind reflect.Kind) any {
	switch kind {
	case reflect.Int64:
		return rand.Int64N(114514)
	case reflect.Float64:
		return rand.Float64() + float64(rand.IntN(114514))
	case reflect.Bool:
		return rand.IntN(2) == 1
	case reflect.String:
		randLength := rand.IntN(114) + 1
		return RandStringBytes(randLength)
	default:
		log.Warn().Msgf("[RandomValueForPrimitiveTypeKind] Unsupported kind: %v", kind)
		return nil
	}
}

// EdgeCaseValueForPrimitiveTypeKind generates an edge case value for a given primitive type.
// A primitive type is a basic data type that is not composed of other types, such as integers, floats, strings, and booleans.
// We assume integer types are int64, and float types are float64.
// The function returns a value that is close to the edge of the type's range.
func EdgeCaseValueForPrimitiveTypeKind(kind reflect.Kind) any {
	var (
		intEdgeCase    = []int64{0, 1, -1, math.MaxInt64, math.MinInt64, math.MaxInt32, math.MinInt32}
		floatEdgeCase  = []float64{0.0, 1.0, -1.0, math.MaxFloat64, math.SmallestNonzeroFloat64, math.MaxFloat32, math.SmallestNonzeroFloat32}
		stringEdgeCase = []string{"", " ", "%20", ".*"}
		boolEdgeCase   = []bool{true, false}
	)
	switch kind {
	case reflect.Int64:
		return intEdgeCase[rand.IntN(len(intEdgeCase))]
	case reflect.Float64:
		return floatEdgeCase[rand.IntN(len(floatEdgeCase))]
	case reflect.Bool:
		return boolEdgeCase[rand.IntN(len(boolEdgeCase))]
	case reflect.String:
		return stringEdgeCase[rand.IntN(len(stringEdgeCase))]
	default:
		log.Warn().Msgf("[EdgeCaseValueForPrimitiveTypeKind] Unsupported kind: %v", kind)
		return nil
	}
}

// FormatServiceName formats the service name.
// It does the following:
//  1. Convert the name to "standard case".(See [resttracefuzzer/pkg/utils.ConvertToStandardCase])
//  2. remove the suffix "service" if exists.
func FormatServiceName(name string) string {
	name = ConvertToStandardCase(name)
	if len(name) > 7 && name[len(name)-7:] == "service" {
		name = name[:len(name)-7]
	}
	return name
}

// Base64ToHex converts a base64 string to a hex string.
// It accepts a base64 string as input and returns a hex string as output.
// The function uses the standard library's base64 and encoding/hex packages to perform the conversion.
// If the input string is not valid base64, it logs a warning and returns an empty string.
// Base64ToHex converts a Base64-encoded string to its hexadecimal representation.
// It first decodes the input Base64 string into raw bytes and then encodes those bytes
// into a hexadecimal string.
//
// Parameters:
//   - base64Str: A string containing the Base64-encoded data.
//
// Returns:
//   - A string containing the hexadecimal representation of the decoded data.
//   - An error if the input string is not a valid Base64-encoded string or if decoding fails.
//
// Example:
//   hexStr, err := Base64ToHex("SGVsbG8gd29ybGQ=")
//   // hexStr will contain "48656c6c6f20776f726c64"
//
// Note:
//   Ensure that the input string is a valid Base64-encoded string to avoid errors.
func Base64ToHex(base64Str string) (string, error) {
	// Decode the Base64 string
    decoded, err := base64.StdEncoding.DecodeString(base64Str)
    if err != nil {
        log.Err(err).Msgf("[Base64ToHex] Failed to decode Base64 string: %v", err)
		return "", err
    }

    // Convert the decoded bytes to a hexadecimal string
    hexStr := hex.EncodeToString(decoded)
	return hexStr, nil
}
