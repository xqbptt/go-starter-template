package utils

import (
	"math/rand"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateRandomString generates a random string of length n.
func GenerateRandomString(n int64) string {
	source := rand.NewSource(time.Now().UnixNano() + int64(rand.Intn(1000))) // Unique seed for each call
	r := rand.New(source)

	result := make([]byte, n)
	for i := range result {
		result[i] = letters[r.Intn(len(letters))] // Pick a random character
	}

	return string(result)
}
