package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher struct{}

// NewPasswordHasher creates new password hasher
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{}
}

// Hash returns a hashed version of the plaintext password or an error if fails
func (ph *PasswordHasher) Hash(plaintextPassword string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

// Matches returns true if the plaintext password matches the hashed password or an error if fails
func (ph *PasswordHasher) Matches(plaintextPassword, hashedPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}
