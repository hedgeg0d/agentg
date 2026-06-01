package auth

import (
	"crypto/subtle"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type secret struct {
	set    bool
	hashed bool
	value  string
}

func newSecret(s string) secret {
	if s == "" {
		return secret{}
	}
	return secret{set: true, hashed: isBcrypt(s), value: s}
}

func (s secret) verify(input string) bool {
	switch {
	case !s.set:
		return false
	case s.hashed:
		return bcrypt.CompareHashAndPassword([]byte(s.value), []byte(input)) == nil
	default:
		return subtle.ConstantTimeCompare([]byte(s.value), []byte(input)) == 1
	}
}

func isBcrypt(s string) bool {
	return strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$") || strings.HasPrefix(s, "$2y$")
}

func HashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(h), err
}
