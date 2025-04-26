package generator

import (
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/simp-lee/passwordmanager/internal/errors"
)

const (
	lowercase = "abcdefghijklmnopqrstuvwxyz"
	uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits    = "0123456789"
	symbols   = "!@#$%^&*()-_=+[]{}|;:,.<>?/"

	similarChars   = "lI1O0"
	ambiguousChars = "{}()[]<>/\\"
)

// Options defines password generation parameters
type Options struct {
	Length           int  // Length of the generated password
	UseLowercase     bool // Include lowercase letters
	UseUppercase     bool // Include uppercase letters
	UseDigits        bool // Include digits
	UseSymbols       bool // Include symbols
	ExcludeSimilar   bool // Exclude similar characters (e.g., 'l', '1', 'I', 'O', '0')
	ExcludeAmbiguous bool // Exclude potentially ambiguous characters (e.g., '{}', '()', '[]', '<>', '/')
}

func DefaultOptions() Options {
	return Options{
		Length:           16,
		UseLowercase:     true,
		UseUppercase:     true,
		UseDigits:        true,
		UseSymbols:       true,
		ExcludeSimilar:   false,
		ExcludeAmbiguous: false,
	}
}

// GeneratePassword creates a new random password based on specified options
// It uses crypto/rand for cryptographically secure random generation
func GeneratePassword(options Options) (string, error) {
	if options.Length <= 0 {
		return "", errors.ErrInvalidLength
	}

	var charset string

	// Build the character set based on selected options
	if options.UseLowercase {
		charset += lowercase
	}
	if options.UseUppercase {
		charset += uppercase
	}
	if options.UseDigits {
		charset += digits
	}
	if options.UseSymbols {
		charset += symbols
	}

	// Filter out characters based on exclusion options
	if options.ExcludeSimilar || options.ExcludeAmbiguous {
		var filteredCharset strings.Builder
		filteredCharset.Grow(len(charset))

		for _, char := range charset {
			if options.ExcludeSimilar && strings.ContainsRune(similarChars, char) {
				continue
			}
			if options.ExcludeAmbiguous && strings.ContainsRune(ambiguousChars, char) {
				continue
			}
			filteredCharset.WriteRune(char)
		}

		charset = filteredCharset.String()
	}

	if len(charset) == 0 {
		return "", errors.ErrEmptyCharset
	}

	// Generate the password using cryptographically secure random numbers
	var password strings.Builder
	password.Grow(options.Length) // Pre-allocate memory for performance
	max := big.NewInt(int64(len(charset)))

	for i := 0; i < options.Length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", errors.Wrap(err, "error generating random number")
		}
		password.WriteByte(charset[n.Int64()])
	}

	return password.String(), nil
}
