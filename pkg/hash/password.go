package hash

import (
	"crypto"
	"fmt"
)

// PasswordHasher provides hashing logic to securely store passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
}

// SHA1Hasher uses SHA1 to hash passwords with provided salt.
type SHA256Hasher struct {
	salt string
}

func NewSHA256Hasher(salt string) *SHA256Hasher {
	return &SHA256Hasher{salt: salt}
}

// Hash creates SHA1 hash of given password.
func (h *SHA256Hasher) Hash(password string) (string, error) {
	hash := crypto.SHA256.New()

	if _, err := hash.Write([]byte(password)); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum([]byte(h.salt))), nil
}
