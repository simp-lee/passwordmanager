package passwordmanager_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/simp-lee/passwordmanager/internal/crypto"
	"github.com/simp-lee/passwordmanager/internal/generator"
	"github.com/simp-lee/passwordmanager/internal/model"
	"github.com/simp-lee/passwordmanager/internal/storage"
)

// Integration Test: Test the basic workflow
func TestBasicFlow(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "passwordmanager-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir) // Clean up the temporary directory afterwards

	// 1. Initialize storage
	store, err := storage.New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// 2. Create vault
	masterPassword := "test-master-password"
	if err := store.CreateVault(masterPassword); err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	// 3. Unlock vault
	if err := store.UnlockVault(masterPassword); err != nil {
		t.Fatalf("Failed to unlock vault: %v", err)
	}

	// 4. Generate password
	passwordOpts := generator.DefaultOptions()
	password, err := generator.GeneratePassword(passwordOpts)
	if err != nil {
		t.Fatalf("Failed to generate password: %v", err)
	}

	// 5. Encrypt password
	encryptedPassword, err := crypto.Encrypt([]byte(password), store.GetEncryptionKey())
	if err != nil {
		t.Fatalf("Failed to encrypt password: %v", err)
	}

	// 6. Add account
	account := &model.Account{
		ID:                "test-account-id",
		Platform:          "TestPlatform",
		Username:          "test-username",
		Email:             "test@example.com",
		EncryptedPassword: encryptedPassword,
		URL:               "https://test.example.com",
		Notes:             "Test account notes",
	}

	if err := store.AddAccount(account); err != nil {
		t.Fatalf("Failed to add account: %v", err)
	}

	// 7. Get and verify account
	retrievedAccount, err := store.GetAccountByID(account.ID)
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if retrievedAccount.Platform != account.Platform {
		t.Errorf("Expected platform %s, got %s", account.Platform, retrievedAccount.Platform)
	}

	// 8. Decrypt password
	decryptedPasswordBytes, err := crypto.Decrypt(retrievedAccount.EncryptedPassword, store.GetEncryptionKey())
	if err != nil {
		t.Fatalf("Failed to decrypt password: %v", err)
	}

	if string(decryptedPasswordBytes) != password {
		t.Errorf("Decrypted password does not match original password")
	}

	// 9. Change master password
	newMasterPassword := "new-master-password"
	if err := store.ChangeMasterPassword(newMasterPassword); err != nil {
		t.Fatalf("Failed to change master password: %v", err)
	}

	// 10. Unlock with new password
	if err := store.UnlockVault(newMasterPassword); err != nil {
		t.Fatalf("Failed to unlock with new password: %v", err)
	}

	// 11. Verify data is still accessible
	accounts, err := store.GetAccounts()
	if err != nil {
		t.Fatalf("Failed to get accounts after password change: %v", err)
	}

	if len(accounts) != 1 {
		t.Errorf("Expected 1 account after password change, got %d", len(accounts))
	}

	// 12. Export and Import Test
	exportPath := filepath.Join(tmpDir, "exported-vault")
	if err := store.ExportVault(exportPath); err != nil {
		t.Fatalf("Failed to export vault: %v", err)
	}

	// Create a new storage instance in a different directory
	newStore, err := storage.New(filepath.Join(tmpDir, "new-data"))
	if err != nil {
		t.Fatalf("Failed to create new storage instance: %v", err)
	}

	// Import the exported vault
	if err := newStore.ImportVault(exportPath); err != nil {
		t.Fatalf("Failed to import vault: %v", err)
	}

	// Unlock the imported vault with the new password
	if err := newStore.UnlockVault(newMasterPassword); err != nil {
		t.Fatalf("Failed to unlock imported vault: %v", err)
	}

	// Verify imported data
	importedAccounts, err := newStore.GetAccounts()
	if err != nil {
		t.Fatalf("Failed to get accounts from imported vault: %v", err)
	}

	if len(importedAccounts) != 1 {
		t.Errorf("Expected 1 account after import, got %d", len(importedAccounts))
	}

	if importedAccounts[0].ID != account.ID {
		t.Errorf("Imported account ID mismatch: expected %s, got %s", account.ID, importedAccounts[0].ID)
	}
}
