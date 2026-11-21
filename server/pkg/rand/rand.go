package rand

import (
	crand "crypto/rand"
	"math/big"
)

// Random strings here end up in session IDs, reset tokens and client secrets,
// so they come from crypto/rand (never a time-seeded math/rand source).

const (
	letterSpecials = "~=+%^*/()[]{}/!@#$?|"
	letterDigits   = "0123456789"
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" + letterDigits
)

// intn returns a uniform random number in [0, n)
func intn(n int) int {
	v, err := crand.Int(crand.Reader, big.NewInt(int64(n)))
	if err != nil {
		// the system entropy source is unusable; never fall back to something predictable
		panic("crypto/rand: " + err.Error())
	}

	return int(v.Int64())
}

// Bytes generates a random alphanumeric string of n characters
func Bytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[intn(len(letterBytes))]
	}

	return b
}

// Password generates a random ASCII string with at least one digit and one special character
func Password(n int) string {
	b := make([]byte, n)

	for i := 0; i < n; i++ {
		var s string
		if i == 0 {
			s = letterDigits
		} else if i == 1 {
			s = letterSpecials
		} else {
			s = letterBytes
		}
		b[i] = s[intn(len(s))]
	}

	for i := len(b) - 1; i > 0; i-- {
		j := intn(i + 1)
		b[i], b[j] = b[j], b[i]
	}

	return string(b) // E.g. "3i[g0|)z"
}
