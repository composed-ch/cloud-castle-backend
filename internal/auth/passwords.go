package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func RandomPasswordAlnum(n uint) (string, error) {
	buf := make([]rune, n)
	alphabet := make([]rune, 0)
	for c := '0'; c <= '9'; c++ {
		alphabet = append(alphabet, c)
	}
	for c := 'A'; c <= 'Z'; c++ {
		alphabet = append(alphabet, c)
	}
	for c := 'a'; c <= 'z'; c++ {
		alphabet = append(alphabet, c)
	}
	for i := range n {
		x, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", fmt.Errorf("get random number: %v", err)
		}
		buf[i] = alphabet[x.Int64()]
	}
	return string(buf), nil
}

// SufficientlyStrong returns true if the password has at least eight different characters.
func SufficientlyStrong(password string) bool {
	if len(password) < 8 {
		return false
	}
	chars := make(map[rune]bool)
	for _, c := range password {
		chars[c] = true
	}
	return len(chars) >= 8
}
