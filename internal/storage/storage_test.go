package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/simp-lee/passwordmanager/internal/crypto"
	"github.com/simp-lee/passwordmanager/internal/errors"
	"github.com/simp-lee/passwordmanager/internal/model"
)

// Test helper function
func setupTestStorage(t *testing.T) (*Storage, string) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "passwordmanager-test-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	storage, err := New(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create Storage instance: %v", err)
	}

	return storage, tmpDir
}

func cleanupTestStorage(tmpDir string) {
	os.RemoveAll(tmpDir)
}

func TestNew(t *testing.T) {
	// Test empty directory argument
	_, err := New("")
	if !errors.Is(err, errors.ErrDirectoryRequired) {
		t.Errorf("Expected ErrDirectoryRequired when passing empty directory, got: %v", err)
	}

	// Test normal creation
	s, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	if s == nil {
		t.Fatal("New returned a nil Storage instance")
	}

	expectedPath := filepath.Join(tmpDir, vaultFileName)
	if s.vaultPath != expectedPath {
		t.Errorf("Expected vaultPath to be %s, got %s", expectedPath, s.vaultPath)
	}

	// Verify directory was actually created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Errorf("Data directory was not created: %v", err)
	}
}

func TestCreateUnlockVault(t *testing.T) {
	s, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	masterPassword := "test-password"

	// Initially, vault should not exist
	if s.IsVaultExists() {
		t.Error("Expected vault not to exist initially")
	}

	// Create vault
	err := s.CreateVault(masterPassword)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	// Now vault should exist
	if !s.IsVaultExists() {
		t.Error("Vault should exist after creation")
	}

	// Creating again should fail
	err = s.CreateVault(masterPassword)
	if !errors.Is(err, errors.ErrVaultExists) {
		t.Errorf("Expected ErrVaultExists when creating vault again, got: %v", err)
	}

	// Unlock vault
	err = s.UnlockVault(masterPassword)
	if err != nil {
		t.Fatalf("Failed to unlock vault: %v", err)
	}

	// Unlock with wrong password
	err = s.UnlockVault("wrong-password")
	if !errors.Is(err, errors.ErrInvalidPassword) {
		t.Errorf("Expected ErrInvalidPassword when unlocking with wrong password, got: %v", err)
	}
}

func TestAccountOperations(t *testing.T) {
	s, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	masterPassword := "test-password"

	// Create and unlock vault
	if err := s.CreateVault(masterPassword); err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := s.UnlockVault(masterPassword); err != nil {
		t.Fatalf("Failed to unlock vault: %v", err)
	}

	// Initially should have no accounts
	accounts, err := s.GetAccounts()
	if err != nil {
		t.Fatalf("Failed to get accounts: %v", err)
	}
	if len(accounts) != 0 {
		t.Errorf("Expected 0 accounts initially, got %d", len(accounts))
	}

	// Add account
	encryptedPassword, _ := crypto.Encrypt([]byte("password123"), s.key)
	account := &model.Account{
		ID:                "test-id",
		Platform:          "TestPlatform",
		Username:          "testuser",
		Email:             "test@example.com",
		EncryptedPassword: encryptedPassword,
		URL:               "https://example.com",
		Notes:             "Test notes",
	}

	if err := s.AddAccount(account); err != nil {
		t.Fatalf("Failed to add account: %v", err)
	}

	// Get accounts and verify
	accounts, err = s.GetAccounts()
	if err != nil {
		t.Fatalf("Failed to get accounts: %v", err)
	}
	if len(accounts) != 1 {
		t.Errorf("Expected 1 account, got %d", len(accounts))
	}
	if accounts[0].ID != account.ID {
		t.Errorf("Expected account ID %s, got %s", account.ID, accounts[0].ID)
	}

	// Get account by ID
	retrievedAccount, err := s.GetAccountByID(account.ID)
	if err != nil {
		t.Fatalf("Failed to get account by ID: %v", err)
	}
	if retrievedAccount.Platform != account.Platform {
		t.Errorf("Expected platform %s, got %s", account.Platform, retrievedAccount.Platform)
	}

	// Update account
	retrievedAccount.Platform = "UpdatedPlatform"
	retrievedAccount.Username = "updateduser"
	if err := s.UpdateAccount(retrievedAccount); err != nil {
		t.Fatalf("Failed to update account: %v", err)
	}

	// Verify update
	updatedAccount, err := s.GetAccountByID(account.ID)
	if err != nil {
		t.Fatalf("Failed to get updated account: %v", err)
	}
	if updatedAccount.Platform != "UpdatedPlatform" {
		t.Errorf("Expected updated platform 'UpdatedPlatform', got %s", updatedAccount.Platform)
	}

	// Test search
	searchResults, err := s.SearchAccounts("update")
	if err != nil {
		t.Fatalf("Failed to search accounts: %v", err)
	}
	if len(searchResults) != 1 {
		t.Errorf("Expected 1 account in search results, got %d", len(searchResults))
	}

	// Delete account
	if err := s.DeleteAccount(account.ID); err != nil {
		t.Fatalf("Failed to delete account: %v", err)
	}

	// Verify deletion
	accounts, err = s.GetAccounts()
	if err != nil {
		t.Fatalf("Failed to get accounts: %v", err)
	}
	if len(accounts) != 0 {
		t.Errorf("Expected 0 accounts after deletion, got %d", len(accounts))
	}

	// Try to get non-existent account
	_, err = s.GetAccountByID(account.ID)
	if !errors.Is(err, errors.ErrAccountNotFound) {
		t.Errorf("Expected ErrAccountNotFound when getting non-existent account, got: %v", err)
	}
}

func TestChangeMasterPassword(t *testing.T) {
	s, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	oldPassword := "old-password"
	newPassword := "new-password"

	// Create and unlock vault
	if err := s.CreateVault(oldPassword); err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := s.UnlockVault(oldPassword); err != nil {
		t.Fatalf("Failed to unlock vault: %v", err)
	}

	// Add an account
	encryptedPassword, _ := crypto.Encrypt([]byte("account-password"), s.key)
	account := &model.Account{
		ID:                "test-id",
		Platform:          "TestPlatform",
		EncryptedPassword: encryptedPassword,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if err := s.AddAccount(account); err != nil {
		t.Fatalf("Failed to add account: %v", err)
	}

	// Change master password
	if err := s.ChangeMasterPassword(newPassword); err != nil {
		t.Fatalf("Failed to change master password: %v", err)
	}

	// Unlock with new password
	if err := s.UnlockVault(newPassword); err != nil {
		t.Fatalf("Failed to unlock with new password: %v", err)
	}

	// Ensure account data still exists
	accounts, err := s.GetAccounts()
	if err != nil {
		t.Fatalf("Failed to get accounts: %v", err)
	}
	if len(accounts) != 1 {
		t.Errorf("Expected 1 account to remain, got %d", len(accounts))
	}

	// Try to unlock with old password, should fail
	if err := s.UnlockVault(oldPassword); !errors.Is(err, errors.ErrInvalidPassword) {
		t.Errorf("Expected ErrInvalidPassword when unlocking with old password, got: %v", err)
	}
}
