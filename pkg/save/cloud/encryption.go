// Package cloud provides cloud save synchronization with encryption at rest.
package cloud

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// KeySize is the AES-256 key size in bytes.
	KeySize = 32
	// SaltSize is the salt size for PBKDF2 in bytes.
	SaltSize = 32
	// NonceSize is the AES-GCM nonce size in bytes.
	NonceSize = 12
	// PBKDF2Iterations is the number of iterations for key derivation.
	PBKDF2Iterations = 100000
)

var (
	// ErrInvalidCiphertext is returned when ciphertext is too short or malformed.
	ErrInvalidCiphertext = errors.New("invalid ciphertext: too short or malformed")
	// ErrDecryptionFailed is returned when authentication tag verification fails.
	ErrDecryptionFailed = errors.New("decryption failed: authentication tag mismatch")
)

// deriveKey derives an AES-256 key from a password and salt using PBKDF2-HMAC-SHA256.
func deriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, KeySize, sha256.New)
}

// Encrypt encrypts plaintext using AES-256-GCM with a password-derived key.
// Returns ciphertext in format: [salt(32)][nonce(12)][ciphertext+tag].
func Encrypt(plaintext []byte, password string) ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	key := deriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	result := make([]byte, 0, SaltSize+NonceSize+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)
	return result, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with a password-derived key.
// Expects ciphertext in format: [salt(32)][nonce(12)][ciphertext+tag].
func Decrypt(ciphertext []byte, password string) ([]byte, error) {
	minSize := SaltSize + NonceSize + 16
	if len(ciphertext) < minSize {
		return nil, ErrInvalidCiphertext
	}

	salt := ciphertext[:SaltSize]
	nonce := ciphertext[SaltSize : SaltSize+NonceSize]
	encrypted := ciphertext[SaltSize+NonceSize:]

	key := deriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}
