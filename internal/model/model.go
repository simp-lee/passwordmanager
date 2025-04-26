package model

import "time"

// Account represents a password entry in the vault
// Contains all necessary information for a stored credential
type Account struct {
	ID                string    `json:"id"`
	Platform          string    `json:"platform"` // Platform/service name (e.g., "GitHub", "Gmail")
	Username          string    `json:"username,omitempty"`
	Email             string    `json:"email,omitempty"`
	EncryptedPassword string    `json:"encrypted_password"` // Password encrypted with the master key
	URL               string    `json:"url,omitempty"`
	Notes             string    `json:"notes,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Vault represents the password vault containing all accounts
// The master password is not stored, only a hash for verification
type Vault struct {
	MasterKeyHash []byte     `json:"master_key_hash,omitempty"` // Hash of the master password
	Salt          []byte     `json:"salt"`                      // Salt used for master password hashing
	Accounts      []*Account `json:"accounts"`                  // List of stored accounts
}
