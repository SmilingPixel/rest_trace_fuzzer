package utils

import (
	"math"
	"math/rand/v2"

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
// It accepts an interface{} as input and returns an int64 as output.
// The function handles various integer types and converts them to int64.
// If the input type is not recognized, it logs a warning and returns 0.
//
// Supported conversions:
// - int, int8, int16, int32 -> int64
// - int64 -> no conversion
//
// Example usage:
//     var num interface{} = 42
//     result := ConvertIntTo64BitType(num) // result will be int64(42)
//
// Parameters:
// - num: the integer to be converted, of type interface{}
//
// Returns:
// - the converted integer as an int64, or 0 if the input type is not recognized.
func ConvertIntTo64BitType(num interface{}) int64 {
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
// It accepts an interface{} as input and returns a float64 as output.
// The function handles various float types and converts them to float64.
// If the input type is not recognized, it logs a warning and returns 0.0.
//
// Supported conversions:
// - float32 -> float64
// - float64 -> no conversion
//
// Example usage:
//     var num interface{} = float32(42.0)
//     result := ConvertFloatTo64BitType(num) // result will be float64(42.0)
//
// Parameters:
// - num: the float to be converted, of type interface{}
//
// Returns:
// - the converted float as a float64, or 0.0 if the input type is not recognized.
func ConvertFloatTo64BitType(num interface{}) float64 {
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
