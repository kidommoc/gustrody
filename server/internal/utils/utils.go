package utils

import (
	"math/rand"
)

var charset = []byte("0123456789abcdef")

func GenerateRamdonHexString(n uint) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(16)]
	}
	return string(b)
}
