package storage

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"slices"

	"github.com/simp-lee/passwordmanager/internal/crypto"
	"github.com/simp-lee/passwordmanager/internal/errors"
	"github.com/simp-lee/passwordmanager/internal/model"
)

const (
	vaultFileName = "vault.encrypted" // Encrypted vault filename
	hashLength    = 32                // SHA-256 hash length
)

// Storage manages vault file operations (read, write, encrypt).
type Storage struct {
	vaultPath string       // Path to vault file
	vault     *model.Vault // In-memory vault data
	key       []byte       // Encryption key (derived from master password)
	mu        sync.RWMutex // Mutex for concurrent access
}

// New creates a storage instance for the given data directory.
// Creates the directory if it doesn't exist.
func New(dataDir string) (*Storage, error) {
	if dataDir == "" {
		return nil, errors.ErrDirectoryRequired
	}
	// Create data directory if needed (0700 permissions)
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, errors.Wrap(err, "error creating data directory")
	}
	vaultPath := filepath.Join(dataDir, vaultFileName)
	return &Storage{
		vaultPath: vaultPath,
		vault:     &model.Vault{},
	}, nil
}

// IsVaultExists checks if the vault file exists.
func (s *Storage) IsVaultExists() bool {
	_, err := os.Stat(s.vaultPath)
	return err == nil
}

// CreateVault creates a new, empty vault with the master password.
// Generates salt, hash, key, initializes, and saves the vault.
// Requires exclusive lock.
func (s *Storage) CreateVault(masterPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.IsVaultExists() {
		return errors.ErrVaultExists
	}

	// Generate salt
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return err
	}

	// Hash master password (for verification)
	masterKeyHash := crypto.HashPassword(masterPassword, salt)
	defer crypto.ClearBytes(masterKeyHash)

	// Derive encryption key
	key := crypto.GenerateKey(masterPassword, salt)

	// Initialize in-memory vault
	s.vault = &model.Vault{
		Salt:     salt,
		Accounts: []*model.Account{},
	}
	s.key = key // Store key in memory

	// Save the new vault to disk
	return s.saveVault(masterKeyHash)
}

// UnlockVault unlocks the vault using the master password.
// Reads file, verifies hash, derives key, decrypts, and loads data.
// Requires exclusive lock.
func (s *Storage) UnlockVault(masterPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.IsVaultExists() {
		return errors.ErrVaultNotExists
	}

	// Read vault file
	fileData, err := os.ReadFile(s.vaultPath)
	if err != nil {
		return errors.Wrap(err, "error reading vault file")
	}

	// Basic format check
	if len(fileData) < hashLength+crypto.SaltLength {
		return errors.ErrDataCorrupted
	}

	// Extract hash, salt, encrypted data
	storedHash := fileData[:hashLength]
	salt := fileData[hashLength : hashLength+crypto.SaltLength]
	encryptedVaultData := fileData[hashLength+crypto.SaltLength:]

	// Verify master password hash
	currentHash := crypto.HashPassword(masterPassword, salt)
	defer crypto.ClearBytes(currentHash)
	if !bytes.Equal(storedHash, currentHash) {
		return errors.ErrInvalidPassword
	}

	// Derive decryption key
	key := crypto.GenerateKey(masterPassword, salt)

	// Decrypt vault data
	decryptedData, err := crypto.Decrypt(string(encryptedVaultData), key)
	if err != nil {
		crypto.ClearBytes(key)
		return errors.Wrap(err, "failed to decrypt vault data")
	}

	// Parse JSON data into Vault struct
	vault := &model.Vault{}
	if err := json.Unmarshal(decryptedData, vault); err != nil {
		crypto.ClearBytes(key)
		crypto.ClearBytes(decryptedData)
		return errors.Wrap(err, "failed to unmarshal vault data")
	}
	crypto.ClearBytes(decryptedData)

	// Update in-memory state
	s.vault = vault
	s.vault.Salt = salt
	s.key = key

	return nil // Unlocked successfully
}

// saveVault persists the in-memory vault to disk (encrypted).
// Serializes accounts, encrypts, prepends hash/salt, and writes atomically.
// Caller must hold the lock. Requires masterKeyHash.
func (s *Storage) saveVault(masterKeyHash []byte) error {
	// Pre-conditions
	if s.vault == nil || s.key == nil || len(s.vault.Salt) == 0 || len(masterKeyHash) == 0 {
		return errors.Wrap(errors.ErrVaultLocked, "cannot save vault, missing state")
	}

	// Marshal only accounts to JSON
	tempVault := &model.Vault{Accounts: s.vault.Accounts}
	vaultData, err := json.Marshal(tempVault)
	if err != nil {
		return errors.Wrap(err, "failed to marshal vault data")
	}

	// Encrypt JSON data
	encryptedDataString, err := crypto.Encrypt(vaultData, s.key)
	if err != nil {
		return errors.Wrap(err, "failed to encrypt vault data")
	}
	encryptedDataBytes := []byte(encryptedDataString)

	// Prepare file data: hash | salt | encrypted_data
	fileData := bytes.Join([][]byte{masterKeyHash, s.vault.Salt, encryptedDataBytes}, nil)

	// Atomic write (write to temp file, then rename)
	tempFile := s.vaultPath + ".tmp"
	if err := os.WriteFile(tempFile, fileData, 0600); err != nil {
		return errors.Wrap(err, "failed to write temporary file")
	}
	if err := os.Rename(tempFile, s.vaultPath); err != nil {
		os.Remove(tempFile) // Clean up temp file on failure
		return errors.Wrap(err, "failed to rename temporary file")
	}

	return nil // Saved successfully
}

// AddAccount adds a new account to the vault and saves it.
// Requires exclusive lock.
func (s *Storage) AddAccount(account *model.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.vault == nil || s.key == nil {
		return errors.ErrVaultLocked
	}

	// Set timestamps
	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now

	// Append account
	s.vault.Accounts = append(s.vault.Accounts, account)

	// Get current hash for saving
	hash, err := s.getMasterKeyHashForSave()
	if err != nil {
		return err
	}
	defer crypto.ClearBytes(hash)

	// Save updated vault
	return s.saveVault(hash)
}

// GetAccounts retrieves all accounts from the unlocked vault.
// Requires read lock.
func (s *Storage) GetAccounts() ([]*model.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.vault == nil {
		return nil, errors.ErrVaultLocked
	}
	return s.vault.Accounts, nil
}

// GetAccountByID retrieves a specific account by ID from the unlocked vault.
// Requires read lock.
func (s *Storage) GetAccountByID(id string) (*model.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.vault == nil {
		return nil, errors.ErrVaultLocked
	}

	// Find account by ID
	for _, account := range s.vault.Accounts {
		if account.ID == id {
			return account, nil
		}
	}
	return nil, errors.Wrap(errors.ErrAccountNotFound, fmt.Sprintf("account ID %s not found", id))
}

// UpdateAccount updates an existing account in the vault and saves it.
// Requires exclusive lock.
func (s *Storage) UpdateAccount(account *model.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.vault == nil {
		return errors.ErrVaultLocked
	}

	found := false
	// Find and replace account
	for i, acc := range s.vault.Accounts {
		if acc.ID == account.ID {
			account.CreatedAt = acc.CreatedAt // Preserve original creation time
			account.UpdatedAt = time.Now()    // Update modification time
			s.vault.Accounts[i] = account
			found = true
			break
		}
	}

	if !found {
		return errors.Wrap(errors.ErrAccountNotFound, fmt.Sprintf("account ID %s not found for update", account.ID))
	}

	// Get current hash for saving
	hash, err := s.getMasterKeyHashForSave()
	if err != nil {
		return err
	}
	defer crypto.ClearBytes(hash)

	// Save updated vault
	return s.saveVault(hash)
}

// DeleteAccount removes an account by ID from the vault and saves it.
// Requires exclusive lock.
func (s *Storage) DeleteAccount(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.vault == nil {
		return errors.ErrVaultLocked
	}

	found := false
	// Find and remove account
	for i, account := range s.vault.Accounts {
		if account.ID == id {
			s.vault.Accounts = slices.Delete(s.vault.Accounts, i, i+1)
			found = true
			break
		}
	}

	if !found {
		return errors.Wrap(errors.ErrAccountNotFound, fmt.Sprintf("account ID %s not found for deletion", id))
	}

	// Get current hash for saving
	hash, err := s.getMasterKeyHashForSave()
	if err != nil {
		return err
	}
	defer crypto.ClearBytes(hash)

	// Save updated vault
	return s.saveVault(hash)
}

// GetEncryptionKey retrieves the current encryption key (nil if locked).
// Handle with care. Requires read lock.
func (s *Storage) GetEncryptionKey() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.key
}

// ExportVault exports the raw encrypted vault file to exportPath.
// Requires read lock.
func (s *Storage) ExportVault(exportPath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.IsVaultExists() {
		return errors.ErrVaultNotExists
	}

	// Read current vault file
	data, err := os.ReadFile(s.vaultPath)
	if err != nil {
		return errors.Wrap(err, "failed to read vault file for export")
	}

	// Write data to export path (0600 permissions)
	return os.WriteFile(exportPath, data, 0600)
}

// ImportVault replaces the current vault with one from importPath.
// Backs up the existing vault first. Resets in-memory state (locks vault).
// Requires exclusive lock.
func (s *Storage) ImportVault(importPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Backup existing vault
	backupPath := s.vaultPath + ".bak"
	if s.IsVaultExists() {
		if err := os.Rename(s.vaultPath, backupPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to backup existing vault: %v\n", err)
		} else {
			fmt.Printf("Existing vault backed up to %s\n", backupPath)
		}
	}

	// Read import file
	data, err := os.ReadFile(importPath)
	if err != nil {
		// Try to restore backup on read failure
		if _, backupErr := os.Stat(backupPath); backupErr == nil {
			os.Rename(backupPath, s.vaultPath)
		}
		return errors.Wrap(err, "failed to read import file")
	}

	// Basic format validation (check minimum size)
	// This is a simplified check; a full validation would require attempting decryption.
	if len(data) <= hashLength+crypto.SaltLength {
		if _, backupErr := os.Stat(backupPath); backupErr == nil {
			os.Rename(backupPath, s.vaultPath)
		}
		return errors.Wrap(errors.ErrDataCorrupted, "import file format invalid or too short")
	}

	// Atomic write for import
	tempFile := s.vaultPath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		if _, backupErr := os.Stat(backupPath); backupErr == nil {
			os.Rename(backupPath, s.vaultPath)
		}
		return errors.Wrap(err, "failed to write imported data to temporary file")
	}
	if err := os.Rename(tempFile, s.vaultPath); err != nil {
		os.Remove(tempFile)
		if _, backupErr := os.Stat(backupPath); backupErr == nil {
			os.Rename(backupPath, s.vaultPath)
		}
		return errors.Wrap(err, "failed to finalize import operation")
	}

	// Reset in-memory state (lock)
	s.vault = nil
	s.key = nil

	fmt.Println("Vault imported successfully. Unlock with its original master password.")
	return nil
}

// ChangeMasterPassword changes the master password for the unlocked vault.
// Generates new salt/hash/key, re-encrypts data, and saves.
// Requires exclusive lock.
func (s *Storage) ChangeMasterPassword(newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.vault == nil || s.key == nil {
		return errors.ErrVaultLocked
	}

	// Preserve current accounts
	accounts := s.vault.Accounts

	// Generate new salt
	newSalt, err := crypto.GenerateSalt()
	if err != nil {
		return err
	}

	// Calculate new hash
	newMasterKeyHash := crypto.HashPassword(newPassword, newSalt)
	defer crypto.ClearBytes(newMasterKeyHash)

	// Derive new key
	newKey := crypto.GenerateKey(newPassword, newSalt)

	// Update in-memory state with new salt and key
	s.vault.Salt = newSalt
	s.key = newKey
	s.vault.Accounts = accounts // Ensure accounts are still set

	// Save vault (will use new key for encryption and write new hash/salt)
	err = s.saveVault(newMasterKeyHash)
	if err != nil {
		crypto.ClearBytes(newKey) // Clear new key if save failed
		// Note: In-memory state might be inconsistent if save fails.
		return err
	}

	return nil // Password changed successfully
}

// ExportToCSV exports decrypted vault data to a CSV file.
// WARNING: Insecure. Requires read lock.
func (s *Storage) ExportToCSV(exportPath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.vault == nil || s.key == nil {
		return errors.ErrVaultLocked
	}

	fmt.Println("WARNING: Exporting passwords to CSV is insecure.")

	// Create CSV file
	file, err := os.Create(exportPath)
	if err != nil {
		return errors.Wrap(err, "failed to create CSV file")
	}
	defer file.Close()

	// Write UTF-8 BOM
	_, err = file.Write([]byte{0xEF, 0xBB, 0xBF})
	if err != nil {
		return errors.Wrap(err, "failed to write UTF-8 BOM")
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "Platform", "Username", "Email", "Password", "URL", "Notes", "CreatedAt", "UpdatedAt"}
	if err := writer.Write(header); err != nil {
		return errors.Wrap(err, "failed to write CSV header")
	}

	// Write account records
	for _, account := range s.vault.Accounts {
		// Decrypt password
		decryptedPassword, err := crypto.Decrypt(account.EncryptedPassword, s.key)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to decrypt password for account ID %s", account.ID))
		}

		// Prepare record
		record := []string{
			account.ID, account.Platform, account.Username, account.Email,
			string(decryptedPassword), // Plain text password!
			account.URL, account.Notes,
			account.CreatedAt.Format(time.RFC3339),
			account.UpdatedAt.Format(time.RFC3339),
		}
		crypto.ClearBytes(decryptedPassword) // Clear decrypted password immediately

		// Write record
		if err := writer.Write(record); err != nil {
			return errors.Wrap(err, "failed to write CSV record")
		}
	}

	return nil // Export successful
}

// SearchAccounts searches unlocked accounts by query (case-insensitive).
// Checks Platform, Username, Email, URL, Notes. Returns all if query is empty.
// Requires read lock.
func (s *Storage) SearchAccounts(query string) ([]*model.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.vault == nil {
		return nil, errors.ErrVaultLocked
	}

	if query == "" {
		return s.vault.Accounts, nil // Return all if query is empty
	}

	query = strings.ToLower(query)
	var results []*model.Account

	// Iterate and check fields
	for _, account := range s.vault.Accounts {
		if strings.Contains(strings.ToLower(account.Platform), query) ||
			strings.Contains(strings.ToLower(account.Username), query) ||
			strings.Contains(strings.ToLower(account.Email), query) ||
			strings.Contains(strings.ToLower(account.URL), query) ||
			strings.Contains(strings.ToLower(account.Notes), query) {
			results = append(results, account)
		}
	}

	return results, nil
}

// getMasterKeyHashForSave reads the current master key hash directly from the vault file.
// Used internally during save operations to ensure the correct hash is preserved.
// Assumes vault file exists and is readable. Caller should hold lock.
func (s *Storage) getMasterKeyHashForSave() ([]byte, error) {
	if s.vault == nil || s.key == nil || len(s.vault.Salt) == 0 {
		return nil, errors.ErrVaultLocked // Vault must be unlocked
	}

	// Read raw vault file
	fileData, err := os.ReadFile(s.vaultPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read vault file for hash")
	}

	// Check length
	if len(fileData) < hashLength {
		return nil, errors.ErrDataCorrupted
	}

	// Extract hash (first part of the file)
	hash := make([]byte, hashLength)
	copy(hash, fileData[:hashLength])
	return hash, nil
}
