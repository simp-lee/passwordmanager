package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("Failed to generate salt: %v", err)
	}

	if len(salt1) != SaltLength {
		t.Errorf("Expected salt length %d, got %d", SaltLength, len(salt1))
	}

	// Ensure generated salts are different each time
	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("Failed to generate second salt: %v", err)
	}

	if bytes.Equal(salt1, salt2) {
		t.Error("Generated salts are identical, expected different")
	}
}

func TestGenerateKey(t *testing.T) {
	password := "test-password"
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("Failed to generate salt: %v", err)
	}

	key := GenerateKey(password, salt)

	if len(key) != KeyLength {
		t.Errorf("Expected key length %d, got %d", KeyLength, len(key))
	}

	// Same password and salt should produce the same key
	key2 := GenerateKey(password, salt)
	if !bytes.Equal(key, key2) {
		t.Error("Same password and salt produced different keys")
	}

	// Different password should produce a different key
	key3 := GenerateKey("different-password", salt)
	if bytes.Equal(key, key3) {
		t.Error("Different passwords produced the same key")
	}

	// Different salt should produce a different key
	salt2, _ := GenerateSalt()
	key4 := GenerateKey(password, salt2)
	if bytes.Equal(key, key4) {
		t.Error("Different salts produced the same key")
	}
}

func TestHashPassword(t *testing.T) {
	password := "test-password"
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("Failed to generate salt: %v", err)
	}

	hash := HashPassword(password, salt)

	if len(hash) != KeyLength {
		t.Errorf("Expected hash length %d, got %d", KeyLength, len(hash))
	}

	// Same password and salt should produce the same hash
	hash2 := HashPassword(password, salt)
	if !bytes.Equal(hash, hash2) {
		t.Error("Same password and salt produced different hashes")
	}

	// Different password should produce a different hash
	hash3 := HashPassword("different-password", salt)
	if bytes.Equal(hash, hash3) {
		t.Error("Different passwords produced the same hash")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, KeyLength)
	for i := range key {
		key[i] = byte(i) // Use a predictable key for testing
	}

	tests := []struct {
		name      string
		plaintext string
		key       []byte
		expectErr bool
	}{
		{
			name:      "Normal encrypt/decrypt",
			plaintext: "Test plaintext data",
			key:       key,
			expectErr: false,
		},
		{
			name:      "Empty plaintext",
			plaintext: "",
			key:       key,
			expectErr: false,
		},
		{
			name:      "Long plaintext",
			plaintext: string(bytes.Repeat([]byte("a"), 1000)),
			key:       key,
			expectErr: false,
		},
		{
			name:      "Invalid key length",
			plaintext: "Test data",
			key:       []byte{1, 2, 3}, // Invalid length
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plaintext := []byte(test.plaintext)

			// Encrypt
			ciphertext, err := Encrypt(plaintext, test.key)
			if test.expectErr {
				if err == nil {
					t.Error("Expected encryption error, but got none")
				}
				return // Stop test case if error was expected
			}
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Ensure subsequent encryptions are different (due to nonce)
			if test.plaintext != "" && len(test.key) == KeyLength {
				ciphertext2, err := Encrypt(plaintext, test.key)
				if err != nil {
					t.Fatalf("Second encryption failed: %v", err)
				}
				if ciphertext == ciphertext2 {
					t.Error("Same plaintext and key produced identical ciphertext on second encryption, expected different (due to random nonce)")
				}
			}

			// Decrypt
			decrypted, err := Decrypt(ciphertext, test.key)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			// Verify decryption result
			if !bytes.Equal(plaintext, decrypted) {
				t.Errorf("Decrypted data does not match original data\nOriginal: %s\nDecrypted: %s", plaintext, decrypted)
			}

			// Test decryption with wrong key
			if len(test.key) == KeyLength && test.plaintext != "" {
				wrongKey := make([]byte, KeyLength)
				copy(wrongKey, test.key)
				wrongKey[0] ^= 1 // Flip one bit

				_, err = Decrypt(ciphertext, wrongKey)
				if err == nil {
					t.Error("Decryption with wrong key should fail, but it succeeded")
				}
			}
		})
	}
}

func TestClearBytes(t *testing.T) {
	data := []byte("sensitive-data")
	ClearBytes(data)

	for _, b := range data {
		if b != 0 {
			t.Error("ClearBytes should zero out all bytes")
			break
		}
	}
}
