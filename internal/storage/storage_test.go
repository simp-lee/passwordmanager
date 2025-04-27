package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/simp-lee/passwordmanager/internal/crypto"
	"github.com/simp-lee/passwordmanager/internal/errors"
	"github.com/simp-lee/passwordmanager/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper functions for test setup and teardown
func setupTestStorage(t *testing.T) (*Storage, string) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "passwordmanager-test-*")
	require.NoError(t, err, "Failed to create temporary directory")

	storage, err := New(tmpDir)
	require.NoError(t, err, "Failed to create Storage instance")

	return storage, tmpDir
}

func cleanupTestStorage(tmpDir string) {
	os.RemoveAll(tmpDir)
}

// Helper to create a vault with test accounts
func setupVaultWithAccounts(t *testing.T, s *Storage, masterPassword string, numAccounts int) {
	// Create and unlock vault
	err := s.CreateVault(masterPassword)
	require.NoError(t, err, "Failed to create vault")

	err = s.UnlockVault(masterPassword)
	require.NoError(t, err, "Failed to unlock vault")

	// Add test accounts
	for i := 0; i < numAccounts; i++ {
		plainPassword := fmt.Sprintf("password-%d", i)
		encPassword, err := crypto.Encrypt([]byte(plainPassword), s.GetEncryptionKey())
		require.NoError(t, err, "Failed to encrypt password")

		account := &model.Account{
			ID:                fmt.Sprintf("test-id-%d", i),
			Platform:          fmt.Sprintf("Platform%d", i),
			Username:          fmt.Sprintf("user%d", i),
			Email:             fmt.Sprintf("user%d@example.com", i),
			URL:               fmt.Sprintf("https://platform%d.example.com", i),
			Notes:             fmt.Sprintf("Notes for account %d", i),
			EncryptedPassword: encPassword,
		}

		err = s.AddAccount(account)
		require.NoError(t, err, "Failed to add account")
	}
}

func TestNew(t *testing.T) {
	t.Run("EmptyDirectory", func(t *testing.T) {
		// Test empty directory argument
		_, err := New("")
		assert.ErrorIs(t, err, errors.ErrDirectoryRequired, "Should return ErrDirectoryRequired with empty directory")
	})

	t.Run("ValidDirectory", func(t *testing.T) {
		// Test normal creation
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		// Verify storage instance was created correctly
		assert.NotNil(t, s, "New should not return a nil Storage instance")

		expectedPath := filepath.Join(tmpDir, vaultFileName)
		assert.Equal(t, expectedPath, s.vaultPath, "Incorrect vault path")

		// Verify directory was actually created
		_, err := os.Stat(tmpDir)
		assert.NoError(t, err, "Data directory should be created")
	})
}

func TestVaultCreationAndUnlock(t *testing.T) {
	t.Run("CreateAndUnlock", func(t *testing.T) {
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		masterPassword := "test-password"

		// Initially, vault should not exist
		assert.False(t, s.IsVaultExists(), "Vault should not exist initially")

		// Create vault
		err := s.CreateVault(masterPassword)
		assert.NoError(t, err, "Failed to create vault")

		// Verify vault exists
		assert.True(t, s.IsVaultExists(), "Vault should exist after creation")

		// Unlock vault
		err = s.UnlockVault(masterPassword)
		assert.NoError(t, err, "Failed to unlock vault with correct password")
	})

	t.Run("CreateExistingVault", func(t *testing.T) {
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		masterPassword := "test-password"

		// Create vault
		err := s.CreateVault(masterPassword)
		require.NoError(t, err, "Failed to create vault")

		// Try creating again - should fail
		err = s.CreateVault(masterPassword)
		assert.ErrorIs(t, err, errors.ErrVaultExists, "Should return ErrVaultExists when creating vault again")
	})

	t.Run("UnlockWithWrongPassword", func(t *testing.T) {
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		masterPassword := "test-password"

		// Create vault
		err := s.CreateVault(masterPassword)
		require.NoError(t, err, "Failed to create vault")

		// Try unlocking with wrong password
		err = s.UnlockVault("wrong-password")
		assert.ErrorIs(t, err, errors.ErrInvalidPassword, "Should return ErrInvalidPassword with wrong password")
	})

	t.Run("UnlockNonExistentVault", func(t *testing.T) {
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		// Try unlocking non-existent vault
		err := s.UnlockVault("any-password")
		assert.ErrorIs(t, err, errors.ErrVaultNotExists, "Should return ErrVaultNotExists when vault doesn't exist")
	})
}

func TestAccountOperations(t *testing.T) {
	t.Run("BasicAccountCRUD", func(t *testing.T) {
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		masterPassword := "test-password"

		// Create and unlock vault
		err := s.CreateVault(masterPassword)
		require.NoError(t, err, "Failed to create vault")

		err = s.UnlockVault(masterPassword)
		require.NoError(t, err, "Failed to unlock vault")

		// Initially should have no accounts
		accounts, err := s.GetAccounts()
		assert.NoError(t, err, "Failed to get accounts")
		assert.Len(t, accounts, 0, "Should have 0 accounts initially")

		// Add account
		encryptedPassword, err := crypto.Encrypt([]byte("password123"), s.GetEncryptionKey())
		require.NoError(t, err, "Failed to encrypt password")

		account := &model.Account{
			ID:                "test-id",
			Platform:          "TestPlatform",
			Username:          "testuser",
			Email:             "test@example.com",
			EncryptedPassword: encryptedPassword,
			URL:               "https://example.com",
			Notes:             "Test notes",
		}

		err = s.AddAccount(account)
		assert.NoError(t, err, "Failed to add account")

		// Check created timestamp was set
		assert.False(t, account.CreatedAt.IsZero(), "Created timestamp should be set")
		assert.False(t, account.UpdatedAt.IsZero(), "Updated timestamp should be set")

		// Get accounts and verify
		accounts, err = s.GetAccounts()
		assert.NoError(t, err, "Failed to get accounts")
		assert.Len(t, accounts, 1, "Should have 1 account")
		assert.Equal(t, "test-id", accounts[0].ID, "Account ID mismatch")

		// Get account by ID
		retrievedAccount, err := s.GetAccountByID(account.ID)
		assert.NoError(t, err, "Failed to get account by ID")
		assert.Equal(t, account.Platform, retrievedAccount.Platform, "Platform mismatch")
		assert.Equal(t, account.Username, retrievedAccount.Username, "Username mismatch")
		assert.Equal(t, account.Email, retrievedAccount.Email, "Email mismatch")

		// Store the original timestamps for comparison
		originalCreatedAt := retrievedAccount.CreatedAt
		originalUpdatedAt := retrievedAccount.UpdatedAt

		// Wait longer to ensure time difference is detectable
		time.Sleep(100 * time.Millisecond)

		// Update account
		retrievedAccount.Platform = "UpdatedPlatform"
		retrievedAccount.Username = "updateduser"

		err = s.UpdateAccount(retrievedAccount)
		assert.NoError(t, err, "Failed to update account")

		// Verify update
		updatedAccount, err := s.GetAccountByID(account.ID)
		assert.NoError(t, err, "Failed to get updated account")
		assert.Equal(t, "UpdatedPlatform", updatedAccount.Platform, "Platform not updated")
		assert.Equal(t, "updateduser", updatedAccount.Username, "Username not updated")

		// Check timestamps
		assert.Equal(t, originalCreatedAt, updatedAccount.CreatedAt, "CreatedAt should not change on update")

		// Check that UpdatedAt is different and not earlier than original
		t.Logf("Original UpdatedAt: %v", originalUpdatedAt)
		t.Logf("New UpdatedAt: %v", updatedAccount.UpdatedAt)
		assert.NotEqual(t, originalUpdatedAt, updatedAccount.UpdatedAt, "UpdatedAt should change on update")
		assert.True(t, !updatedAccount.UpdatedAt.Before(originalUpdatedAt), "UpdatedAt should not be earlier than original timestamp")

		// Delete account
		err = s.DeleteAccount(account.ID)
		assert.NoError(t, err, "Failed to delete account")

		// Verify deletion
		accounts, err = s.GetAccounts()
		assert.NoError(t, err, "Failed to get accounts after deletion")
		assert.Len(t, accounts, 0, "Should have 0 accounts after deletion")

		// Try to get deleted account
		_, err = s.GetAccountByID(account.ID)
		assert.ErrorIs(t, err, errors.ErrAccountNotFound, "Should return ErrAccountNotFound for deleted account")
	})

	t.Run("AccountOperationsWithLockedVault", func(t *testing.T) {
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		// Without creating/unlocking vault, operations should fail
		_, err := s.GetAccounts()
		assert.ErrorIs(t, err, errors.ErrVaultLocked, "Should return ErrVaultLocked when vault is not unlocked")

		_, err = s.GetAccountByID("any-id")
		assert.ErrorIs(t, err, errors.ErrVaultLocked, "Should return ErrVaultLocked when vault is not unlocked")

		err = s.AddAccount(&model.Account{ID: "test-id"})
		assert.ErrorIs(t, err, errors.ErrVaultLocked, "Should return ErrVaultLocked when vault is not unlocked")

		err = s.UpdateAccount(&model.Account{ID: "test-id"})
		assert.ErrorIs(t, err, errors.ErrVaultLocked, "Should return ErrVaultLocked when vault is not unlocked")

		err = s.DeleteAccount("test-id")
		assert.ErrorIs(t, err, errors.ErrVaultLocked, "Should return ErrVaultLocked when vault is not unlocked")
	})

	t.Run("NonExistentAccount", func(t *testing.T) {
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		setupVaultWithAccounts(t, s, "test-password", 1)

		// Try to get non-existent account
		_, err := s.GetAccountByID("non-existent-id")
		assert.ErrorIs(t, err, errors.ErrAccountNotFound, "Should return ErrAccountNotFound for non-existent account")

		// Try to update non-existent account
		err = s.UpdateAccount(&model.Account{ID: "non-existent-id"})
		assert.ErrorIs(t, err, errors.ErrAccountNotFound, "Should return ErrAccountNotFound when updating non-existent account")

		// Try to delete non-existent account
		err = s.DeleteAccount("non-existent-id")
		assert.ErrorIs(t, err, errors.ErrAccountNotFound, "Should return ErrAccountNotFound when deleting non-existent account")
	})
}

func TestSearchAccounts(t *testing.T) {
	s, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	// Setup vault with multiple accounts for search testing
	setupVaultWithAccounts(t, s, "test-password", 5)

	// Additional account with specific searchable content
	specialAccount := &model.Account{
		ID:       "special-id",
		Platform: "SpecialPlatform",
		Username: "specialuser",
		Email:    "special@example.com",
		URL:      "https://special.example.com",
		Notes:    "This is a special account with unique content",
	}

	encPassword, err := crypto.Encrypt([]byte("special-password"), s.GetEncryptionKey())
	require.NoError(t, err, "Failed to encrypt password")
	specialAccount.EncryptedPassword = encPassword

	err = s.AddAccount(specialAccount)
	require.NoError(t, err, "Failed to add special account")

	t.Run("SearchEmpty", func(t *testing.T) {
		// Empty search should return all accounts
		results, err := s.SearchAccounts("")
		assert.NoError(t, err, "Search with empty query failed")
		assert.Len(t, results, 6, "Empty search should return all accounts")
	})

	t.Run("SearchPlatform", func(t *testing.T) {
		// Search by platform
		results, err := s.SearchAccounts("platform0")
		assert.NoError(t, err, "Search by platform failed")
		assert.Len(t, results, 1, "Should find one account by platform")
		assert.Equal(t, "Platform0", results[0].Platform, "Found wrong account")
	})

	t.Run("SearchEmail", func(t *testing.T) {
		// Search by email
		results, err := s.SearchAccounts("special@example")
		assert.NoError(t, err, "Search by email failed")
		assert.Len(t, results, 1, "Should find one account by email")
		assert.Equal(t, "special-id", results[0].ID, "Found wrong account")
	})

	t.Run("SearchNotes", func(t *testing.T) {
		// Search by notes
		results, err := s.SearchAccounts("unique content")
		assert.NoError(t, err, "Search by notes failed")
		assert.Len(t, results, 1, "Should find one account by notes")
		assert.Equal(t, "special-id", results[0].ID, "Found wrong account")
	})

	t.Run("SearchCaseInsensitive", func(t *testing.T) {
		// Case insensitive search
		results, err := s.SearchAccounts("SPECIALPLATFORM")
		assert.NoError(t, err, "Case insensitive search failed")
		assert.Len(t, results, 1, "Should find one account with case-insensitive search")
		assert.Equal(t, "special-id", results[0].ID, "Found wrong account")
	})

	t.Run("SearchNoResults", func(t *testing.T) {
		// Search with no results
		results, err := s.SearchAccounts("nonexistent")
		assert.NoError(t, err, "Search with no results failed")
		assert.Len(t, results, 0, "Should find no accounts")
	})

	t.Run("SearchWithLockedVault", func(t *testing.T) {
		// Create new storage with locked vault
		newStorage, newTmpDir := setupTestStorage(t)
		defer cleanupTestStorage(newTmpDir)

		// Search on locked vault should fail
		_, err := newStorage.SearchAccounts("anything")
		assert.ErrorIs(t, err, errors.ErrVaultLocked, "Search on locked vault should return ErrVaultLocked")
	})
}

func TestChangeMasterPassword(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		oldPassword := "old-password"
		newPassword := "new-password"

		// Setup vault with accounts
		setupVaultWithAccounts(t, s, oldPassword, 3)

		// Change master password
		err := s.ChangeMasterPassword(newPassword)
		assert.NoError(t, err, "Failed to change master password")

		// Lock vault and reset storage state
		s.vault = nil
		s.key = nil

		// Old password should no longer work
		err = s.UnlockVault(oldPassword)
		assert.ErrorIs(t, err, errors.ErrInvalidPassword, "Old password should no longer work")

		// New password should work
		err = s.UnlockVault(newPassword)
		assert.NoError(t, err, "Failed to unlock with new password")

		// Verify all account data is preserved
		accounts, err := s.GetAccounts()
		assert.NoError(t, err, "Failed to get accounts after password change")
		assert.Len(t, accounts, 3, "Account count mismatch after password change")

		// Verify passwords are correctly encrypted with new key
		for i, account := range accounts {
			expectedPassword := fmt.Sprintf("password-%d", i)
			decryptedBytes, err := crypto.Decrypt(account.EncryptedPassword, s.GetEncryptionKey())
			assert.NoError(t, err, "Failed to decrypt password for account %d", i)
			assert.Equal(t, expectedPassword, string(decryptedBytes), "Password mismatch for account %d", i)
		}
	})

	t.Run("WithLockedVault", func(t *testing.T) {
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		// Without unlocking vault, change password should fail
		err := s.ChangeMasterPassword("new-password")
		assert.ErrorIs(t, err, errors.ErrVaultLocked, "Should return ErrVaultLocked when vault is not unlocked")
	})
}

func TestExportAndImport(t *testing.T) {
	t.Run("ExportVault", func(t *testing.T) {
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		masterPassword := "test-password"

		// Create vault and add accounts
		setupVaultWithAccounts(t, s, masterPassword, 2)

		// Export vault
		exportPath := filepath.Join(tmpDir, "exported-vault.bin")
		err := s.ExportVault(exportPath)
		assert.NoError(t, err, "Failed to export vault")

		// Verify export file exists and is not empty
		fileInfo, err := os.Stat(exportPath)
		assert.NoError(t, err, "Export file does not exist")
		assert.Greater(t, fileInfo.Size(), int64(0), "Export file is empty")
	})

	t.Run("ExportNonExistentVault", func(t *testing.T) {
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		// Export non-existent vault should fail
		exportPath := filepath.Join(tmpDir, "export-fail.bin")
		err := s.ExportVault(exportPath)
		assert.ErrorIs(t, err, errors.ErrVaultNotExists, "Should fail when exporting non-existent vault")
	})

	t.Run("ImportVault", func(t *testing.T) {
		// Setup original storage with vault and accounts
		originalStorage, originalTmpDir := setupTestStorage(t)
		defer cleanupTestStorage(originalTmpDir)

		masterPassword := "test-password"
		setupVaultWithAccounts(t, originalStorage, masterPassword, 3)

		// Export original vault
		exportPath := filepath.Join(originalTmpDir, "vault-to-import.bin")
		err := originalStorage.ExportVault(exportPath)
		require.NoError(t, err, "Failed to export original vault")

		// Create new storage instance for import
		importStorage, importTmpDir := setupTestStorage(t)
		defer cleanupTestStorage(importTmpDir)

		// Import vault
		err = importStorage.ImportVault(exportPath)
		assert.NoError(t, err, "Failed to import vault")

		// Verify vault exists after import
		assert.True(t, importStorage.IsVaultExists(), "Vault should exist after import")

		// Unlock imported vault and verify data
		err = importStorage.UnlockVault(masterPassword)
		assert.NoError(t, err, "Failed to unlock imported vault")

		accounts, err := importStorage.GetAccounts()
		assert.NoError(t, err, "Failed to get accounts from imported vault")
		assert.Len(t, accounts, 3, "Imported vault has wrong number of accounts")
	})

	t.Run("ImportInvalidFile", func(t *testing.T) {
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		// Create invalid file for import
		invalidPath := filepath.Join(tmpDir, "invalid.bin")
		err := os.WriteFile(invalidPath, []byte("not a valid vault file"), 0600)
		require.NoError(t, err, "Failed to create invalid file")

		// Import should fail
		err = s.ImportVault(invalidPath)
		assert.Error(t, err, "Should fail when importing invalid file")
	})
}

func TestExportToCSV(t *testing.T) {
	t.Run("BasicExport", func(t *testing.T) {
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		// Setup vault with accounts
		setupVaultWithAccounts(t, s, "test-password", 2)

		// Export to CSV
		csvPath := filepath.Join(tmpDir, "export.csv")
		err := s.ExportToCSV(csvPath)
		assert.NoError(t, err, "CSV export failed")

		// Verify file exists and is not empty
		fileInfo, err := os.Stat(csvPath)
		assert.NoError(t, err, "CSV file does not exist")
		assert.Greater(t, fileInfo.Size(), int64(0), "CSV file is empty")

		// Read CSV file content to verify format, but not testing full content parsing
		content, err := os.ReadFile(csvPath)
		assert.NoError(t, err, "Failed to read CSV file")
		assert.Contains(t, string(content), "ID,Platform,Username,Email,Password", "CSV header missing")
	})

	t.Run("ExportWithLockedVault", func(t *testing.T) {
		s, tmpDir := setupTestStorage(t)
		defer cleanupTestStorage(tmpDir)

		// Try to export without unlocking vault
		csvPath := filepath.Join(tmpDir, "locked-export.csv")
		err := s.ExportToCSV(csvPath)
		assert.ErrorIs(t, err, errors.ErrVaultLocked, "Should fail when vault is locked")
	})
}

func TestGetEncryptionKey(t *testing.T) {
	s, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	// Initially (locked vault), key should be nil
	key := s.GetEncryptionKey()
	assert.Nil(t, key, "Key should be nil for locked vault")

	// Create and unlock vault
	err := s.CreateVault("test-password")
	require.NoError(t, err, "Failed to create vault")

	err = s.UnlockVault("test-password")
	require.NoError(t, err, "Failed to unlock vault")

	// After unlock, key should be available
	key = s.GetEncryptionKey()
	assert.NotNil(t, key, "Key should be available after unlock")
	assert.Len(t, key, crypto.KeyLength, "Key should have correct length")
}
