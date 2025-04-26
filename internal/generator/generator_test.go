package generator

import (
	"strings"
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Length != 16 {
		t.Errorf("Expected default length 16, got %d", opts.Length)
	}
	if !opts.UseLowercase {
		t.Error("Expected default use of lowercase letters")
	}
	if !opts.UseUppercase {
		t.Error("Expected default use of uppercase letters")
	}
	if !opts.UseDigits {
		t.Error("Expected default use of digits")
	}
	if !opts.UseSymbols {
		t.Error("Expected default use of symbols")
	}
	if opts.ExcludeSimilar {
		t.Error("Expected default not to exclude similar characters")
	}
	if opts.ExcludeAmbiguous {
		t.Error("Expected default not to exclude ambiguous characters")
	}
}

func TestGeneratePassword(t *testing.T) {
	tests := []struct {
		name          string
		options       Options
		expectedLen   int
		expectedChars string // Base charset before exclusions
		expectErr     bool
	}{
		{
			name: "Default options",
			options: Options{
				Length:       16,
				UseLowercase: true,
				UseUppercase: true,
				UseDigits:    true,
				UseSymbols:   true,
			},
			expectedLen:   16,
			expectedChars: lowercase + uppercase + digits + symbols,
			expectErr:     false,
		},
		{
			name: "Lowercase only",
			options: Options{
				Length:       10,
				UseLowercase: true,
				UseUppercase: false,
				UseDigits:    false,
				UseSymbols:   false,
			},
			expectedLen:   10,
			expectedChars: lowercase,
			expectErr:     false,
		},
		{
			name: "Exclude similar",
			options: Options{
				Length:         12,
				UseLowercase:   true,
				UseUppercase:   true,
				UseDigits:      true,
				UseSymbols:     true,
				ExcludeSimilar: true,
			},
			expectedLen:   12,
			expectedChars: lowercase + uppercase + digits + symbols,
			expectErr:     false,
		},
		{
			name: "Exclude ambiguous",
			options: Options{
				Length:           12,
				UseLowercase:     true,
				UseUppercase:     true,
				UseDigits:        true,
				UseSymbols:       true,
				ExcludeAmbiguous: true,
			},
			expectedLen:   12,
			expectedChars: lowercase + uppercase + digits + symbols,
			expectErr:     false,
		},
		{
			name: "Invalid length",
			options: Options{
				Length:       0,
				UseLowercase: true,
				UseUppercase: true,
				UseDigits:    true,
				UseSymbols:   true,
			},
			expectErr: true,
		},
		{
			name: "Empty charset",
			options: Options{
				Length:       10,
				UseLowercase: false,
				UseUppercase: false,
				UseDigits:    false,
				UseSymbols:   false,
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			password, err := GeneratePassword(test.options)

			if test.expectErr {
				if err == nil {
					t.Error("Expected an error, but got none")
				}
				return // Stop test case if error was expected
			}

			if err != nil {
				t.Errorf("Unexpected error generating password: %v", err)
				return
			}

			if len(password) != test.expectedLen {
				t.Errorf("Expected password length %d, got %d", test.expectedLen, len(password))
			}

			// Check each character: it should not be in the excluded sets,
			// and if not excluded, it should be in the base expected set.
			for _, char := range password {
				charStr := string(char)

				if test.options.ExcludeSimilar && strings.ContainsRune(similarChars, char) {
					t.Errorf("Password contains excluded similar character: %s", charStr)
				}

				if test.options.ExcludeAmbiguous && strings.ContainsRune(ambiguousChars, char) {
					t.Errorf("Password contains excluded ambiguous character: %s", charStr)
				}

				// Build the actual allowed charset for this test case
				var allowedCharsetBuilder strings.Builder
				if test.options.UseLowercase {
					allowedCharsetBuilder.WriteString(lowercase)
				}
				if test.options.UseUppercase {
					allowedCharsetBuilder.WriteString(uppercase)
				}
				if test.options.UseDigits {
					allowedCharsetBuilder.WriteString(digits)
				}
				if test.options.UseSymbols {
					allowedCharsetBuilder.WriteString(symbols)
				}
				allowedCharset := allowedCharsetBuilder.String()

				// Filter the allowed charset based on exclusions
				if test.options.ExcludeSimilar || test.options.ExcludeAmbiguous {
					var filteredAllowedCharset strings.Builder
					for _, allowedChar := range allowedCharset {
						if test.options.ExcludeSimilar && strings.ContainsRune(similarChars, allowedChar) {
							continue
						}
						if test.options.ExcludeAmbiguous && strings.ContainsRune(ambiguousChars, allowedChar) {
							continue
						}
						filteredAllowedCharset.WriteRune(allowedChar)
					}
					allowedCharset = filteredAllowedCharset.String()
				}

				// Check if the generated character is in the final allowed set
				if !strings.ContainsRune(allowedCharset, char) {
					t.Errorf("Password contains character '%s' which is not in the expected allowed set '%s'", charStr, allowedCharset)
				}
			}
		})
	}
}

func TestPasswordUniqueness(t *testing.T) {
	// Test if generated passwords are random enough
	opts := DefaultOptions()
	passwords := make(map[string]bool)
	iterations := 100 // Number of passwords to generate for uniqueness check

	for i := 0; i < iterations; i++ {
		password, err := GeneratePassword(opts)
		if err != nil {
			t.Fatalf("Failed to generate password: %v", err)
		}
		if passwords[password] {
			t.Errorf("Generated duplicate password: %s", password)
		}
		passwords[password] = true
	}
}
