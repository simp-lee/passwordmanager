package backend

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/atotto/clipboard"
	"github.com/simp-lee/passwordmanager/internal/crypto"
	"github.com/simp-lee/passwordmanager/internal/errors"
	"github.com/simp-lee/passwordmanager/internal/generator"
	"github.com/simp-lee/passwordmanager/internal/model"
	"github.com/simp-lee/passwordmanager/internal/storage"
)

type App struct {
	ctx        context.Context
	store      *storage.Storage
	dataDir    string
	isUnlocked bool
}

func NewApp() *App {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get home directory: %v", err))
	}
	dataDir := filepath.Join(homeDir, ".passwordmanager")

	store, err := storage.New(dataDir)
	if err != nil {
		panic(fmt.Sprintf("failed to create storage: %v", err))
	}

	return &App{
		store:      store,
		dataDir:    dataDir,
		isUnlocked: false,
	}
}

func (a *App) SetContext(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) IsVaultExists() bool {
	return a.store.IsVaultExists()
}

// UnlockVault unlocks the vault using the provided master password.
func (a *App) UnlockVault(masterPassword string) error {
	if err := a.store.UnlockVault(masterPassword); err != nil {
		return err
	}
	a.isUnlocked = true
	return nil
}

func (a *App) CreateVault(masterPassword string) error {
	if err := a.store.CreateVault(masterPassword); err != nil {
		return err
	}
	a.isUnlocked = true
	return nil
}

func (a *App) GetAccounts() ([]*model.Account, error) {
	if !a.isUnlocked {
		return nil, errors.ErrVaultLocked
	}
	return a.store.GetAccounts()
}

func (a *App) AddAccount(account *model.Account, password string) error {
	if !a.isUnlocked {
		return errors.ErrVaultLocked
	}

	now := time.Now()
	if account.ID == "" {
		account.ID = a.generateID()
	}
	account.CreatedAt = now
	account.UpdatedAt = now

	// Encrypt the password before storing it
	encrptedPassword, err := crypto.Encrypt([]byte(password), a.store.GetEncryptionKey())
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	account.EncryptedPassword = encrptedPassword

	return a.store.AddAccount(account)
}

func (a *App) GetAccountByID(id string) (*model.Account, error) {
	if !a.isUnlocked {
		return nil, errors.ErrVaultLocked
	}
	return a.store.GetAccountByID(id)
}

func (a *App) UpdateAccount(account *model.Account, password *string) error {
	if !a.isUnlocked {
		return errors.ErrVaultLocked
	}

	// Encrypt the password before storing it
	if password != nil {
		encrptedPassword, err := crypto.Encrypt([]byte(*password), a.store.GetEncryptionKey())
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		account.EncryptedPassword = encrptedPassword
	}

	account.UpdatedAt = time.Now()

	return a.store.UpdateAccount(account)
}

func (a *App) DeleteAccount(id string) error {
	if !a.isUnlocked {
		return errors.ErrVaultLocked
	}
	return a.store.DeleteAccount(id)
}

func (a *App) DecryptPassword(id string) (string, error) {
	if !a.isUnlocked {
		return "", errors.ErrVaultLocked
	}
	account, err := a.store.GetAccountByID(id)
	if err != nil {
		return "", err
	}
	if account == nil {
		return "", errors.ErrAccountNotFound
	}
	decryptedPassword, err := crypto.Decrypt(account.EncryptedPassword, a.store.GetEncryptionKey())
	if err != nil {
		return "", fmt.Errorf("failed to decrypt password: %w", err)
	}
	return string(decryptedPassword), nil
}

// CopyToClipboard copies the given content to the clipboard and clears it after the specified duration.
func (a *App) CopyToClipboard(content string, clearAfterSeconds int) error {
	if err := clipboard.WriteAll(content); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	go func() {
		time.Sleep(time.Duration(clearAfterSeconds) * time.Second)
		if err := clipboard.WriteAll(""); err != nil {
			fmt.Printf("failed to clear clipboard: %v\n", err)
		}
	}()

	return nil
}

func (a *App) GeneratePassword(opts generator.Options) (string, error) {
	return generator.GeneratePassword(opts)
}

func (a *App) ChangeMasterPassword(oldPassword, newPassword string) error {
	if !a.isUnlocked {
		return errors.ErrVaultLocked
	}

	if err := a.store.UnlockVault(oldPassword); err != nil {
		return errors.ErrInvalidPassword
	}

	return a.store.ChangeMasterPassword(newPassword)
}

func (a *App) SearchAccounts(query string) ([]*model.Account, error) {
	if !a.isUnlocked {
		return nil, errors.ErrVaultLocked
	}
	return a.store.SearchAccounts(query)
}

func (a *App) generateID() string {
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		panic(fmt.Sprintf("failed to generate ID: %v", err))
	}
	return hex.EncodeToString(buf)
}

// ShowExportDialog 显示导出文件对话框并返回选择的路径
func (a *App) ShowExportDialog() (string, error) {
	if !a.isUnlocked {
		return "", errors.ErrVaultLocked
	}

	now := time.Now().Format("2006-01-02")
	defaultFilename := fmt.Sprintf("passwordmanager-export-%s.encrypted", now)

	// 使用Wails运行时显示保存文件对话框
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
		Title:           "导出加密密码库",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "加密备份文件 (*.encrypted)",
				Pattern:     "*.encrypted",
			},
		},
	})

	if err != nil {
		return "", err
	}

	return path, nil
}

// ShowImportDialog 显示导入文件对话框并返回选择的路径
func (a *App) ShowImportDialog() (string, error) {
	// 使用Wails运行时显示打开文件对话框
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "导入密码库",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "加密备份文件 (*.encrypted)",
				Pattern:     "*.encrypted",
			},
		},
	})

	if err != nil {
		return "", err
	}

	return path, nil
}

// ExportVaultToPath 导出密码库到指定路径
func (a *App) ExportVaultToPath(path string) error {
	if !a.isUnlocked {
		return errors.ErrVaultLocked
	}

	// 调用存储模块的导出方法
	return a.store.ExportVault(path)
}

// ImportVaultFromPath 从指定路径导入密码库
func (a *App) ImportVaultFromPath(path string) error {
	if path == "" {
		return fmt.Errorf("未指定导入文件路径")
	}

	// 调用存储模块的导入方法
	return a.store.ImportVault(path)
}

// ExportVault 显示对话框并导出密码库
func (a *App) ExportVault() error {
	if !a.isUnlocked {
		return errors.ErrVaultLocked
	}

	// 显示文件对话框
	path, err := a.ShowExportDialog()
	if err != nil {
		return err
	}

	// 用户取消了对话框
	if path == "" {
		return nil
	}

	// 导出到选择的路径
	return a.ExportVaultToPath(path)
}

// ImportVault 显示对话框并导入密码库
func (a *App) ImportVault() error {
	// 显示文件对话框
	path, err := a.ShowImportDialog()
	if err != nil {
		return err
	}

	// 用户取消了对话框
	if path == "" {
		return nil
	}

	// 从选择的路径导入
	return a.ImportVaultFromPath(path)
}

// ExportCsv 显示对话框并导出密码库为CSV格式
func (a *App) ExportCsv() error {
	if !a.isUnlocked {
		return errors.ErrVaultLocked
	}

	now := time.Now().Format("2006-01-02")
	defaultFilename := fmt.Sprintf("passwordmanager-export-%s.csv", now)

	// 显示文件对话框
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
		Title:           "导出密码库为CSV",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "CSV文件 (*.csv)",
				Pattern:     "*.csv",
			},
		},
	})

	if err != nil {
		return err
	}

	if path == "" {
		return nil // 用户取消了对话框
	}

	if !strings.HasSuffix(strings.ToLower(path), ".csv") {
		path += ".csv" // 如果没有扩展名，则添加.csv
	}

	return a.store.ExportToCSV(path)
}
