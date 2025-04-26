package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"

	"github.com/simp-lee/passwordmanager/internal/errors"
	"golang.org/x/crypto/pbkdf2"
)

// Cryptographic constants
const (
	KeyLength  = 32     // AES-256 key length (bytes)
	SaltLength = 16     // PBKDF2 salt length (bytes)
	Iterations = 100000 // PBKDF2 iterations
)

// GenerateSalt creates a random salt for password hashing.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltLength)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate salt")
	}
	return salt, nil
}

// GenerateKey derives an encryption key from a master password and salt using PBKDF2.
func GenerateKey(masterPassword string, salt []byte) []byte {
	return pbkdf2.Key([]byte(masterPassword), salt, Iterations, KeyLength, sha256.New)
}

// HashPassword creates a password hash using PBKDF2 (used for verification).
func HashPassword(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, Iterations, KeyLength, sha256.New)
}

// Encrypt encrypts data using AES-GCM.
// Returns base64 string: nonce + ciphertext.
func Encrypt(plaintext, key []byte) (string, error) {
	if len(key) != KeyLength {
		return "", errors.Wrap(errors.ErrDataCorrupted, "invalid encryption key length")
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err, "failed to create cipher")
	}

	// Create GCM mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.Wrap(err, "failed to create GCM")
	}

	// Generate random nonce
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.Wrap(err, "failed to generate nonce")
	}

	// Encrypt: nonce is prepended to the ciphertext
	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts data encrypted by Encrypt.
// Expects base64 string (nonce + ciphertext) and the key.
func Decrypt(ciphertext string, key []byte) ([]byte, error) {
	if len(key) != KeyLength {
		return nil, errors.Wrap(errors.ErrDataCorrupted, "invalid decryption key length")
	}

	// Decode base64
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode base64")
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cipher")
	}

	// Create GCM mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GCM")
	}

	// Check length (must include nonce)
	nonceSize := aesgcm.NonceSize()
	if len(ciphertextBytes) < nonceSize {
		return nil, errors.Wrap(errors.ErrDataCorrupted, "ciphertext too short")
	}

	// Extract nonce and actual ciphertext
	nonce, actualCiphertext := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]

	// Decrypt and verify
	plaintext, err := aesgcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		// Error could be invalid key or corrupted data
		return nil, errors.Wrap(errors.ErrInvalidPassword, "decryption failed (invalid key or corrupted data)")
	}

	return plaintext, nil
}

// ClearBytes securely wipes sensitive byte slices (e.g., keys, passwords).
func ClearBytes(data []byte) {
	for i := range data {
		data[i] = 0
	}
}
