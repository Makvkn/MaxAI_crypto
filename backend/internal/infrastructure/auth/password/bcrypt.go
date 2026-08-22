// Package password hashes and verifies credentials with bcrypt (§14).
package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	appauth "github.com/maxaicrypto/backend/internal/application/auth"
)

const cost = bcrypt.DefaultCost

// Hasher implements auth.PasswordHasher.
type Hasher struct{}

// NewHasher returns a bcrypt password hasher.
func NewHasher() *Hasher { return &Hasher{} }

// Hash returns a bcrypt hash of password.
func (Hasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// Verify reports whether password matches hash.
func (Hasher) Verify(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

var _ appauth.PasswordHasher = (*Hasher)(nil)
