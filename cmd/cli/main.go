package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
	"golang.org/x/term"

	"github.com/olekukonko/tablewriter"
	"github.com/simp-lee/passwordmanager/internal/crypto"
	"github.com/simp-lee/passwordmanager/internal/errors"
	"github.com/simp-lee/passwordmanager/internal/generator"
	"github.com/simp-lee/passwordmanager/internal/i18n"
	"github.com/simp-lee/passwordmanager/internal/model"
	"github.com/simp-lee/passwordmanager/internal/storage"
	"github.com/spf13/cobra"
)

var (
	dataDir    string
	store      *storage.Storage
	isUnlocked bool
)

func main() {
	// Get user home directory as the base path for data storage
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get user home directory: %v\n", err)
		os.Exit(1)
	}
	dataDir = filepath.Join(homeDir, ".passwordmanager")

	// Initialize the storage system
	store, err = storage.New(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	// Create the root command
	rootCmd := &cobra.Command{
		Use:     "passwordmanager",
		Short:   i18n.T("app_name"),
		Long:    i18n.T("app_description"),
		Version: "0.1.0",
	}

	rootCmd.PersistentFlags().StringP("lang", "L", "en", i18n.T("opt_lang"))

	// Set up language and unlock logic for all commands
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		lang, _ := cmd.Flags().GetString("lang")
		i18n.SetLanguage(lang)

		// For commands that don't require unlocking, return directly
		switch cmd.Name() {
		case "init", "import", "help", "version":
			return
		}

		// If already unlocked, no need to operate
		if isUnlocked {
			return
		}

		// Check if the vault exists
		if !store.IsVaultExists() {
			fmt.Fprintf(os.Stderr, i18n.T("vault_not_exists")+"\n")
			os.Exit(1)
		}

		// Perform unlock operation
		masterPassword, err := readPassword(i18n.T("enter_master_password"))
		if err != nil {
			fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
			os.Exit(1)
		}

		if err := store.UnlockVault(masterPassword); err != nil {
			fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
			os.Exit(1)
		}

		crypto.ClearBytes([]byte(masterPassword))
		isUnlocked = true
	}

	// --- Command Definitions ---

	initCmd := &cobra.Command{
		Use:   "init",
		Short: i18n.T("cmd_init_short"),
		Run:   initVault,
	}

	addCmd := &cobra.Command{
		Use:   "add",
		Short: i18n.T("cmd_add_short"),
		Run:   addAccount,
	}

	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: i18n.T("cmd_generate_short"),
		Run:   generatePassword,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: i18n.T("cmd_list_short"),
		Run:   listAccounts,
	}

	getCmd := &cobra.Command{
		Use:   "get [platform]",
		Short: i18n.T("cmd_get_short"),
		Args:  cobra.MinimumNArgs(1),
		Run:   getAccount,
	}

	deleteCmd := &cobra.Command{
		Use:   "delete [ID]",
		Short: i18n.T("cmd_delete_short"),
		Args:  cobra.ExactArgs(1),
		Run:   deleteAccount,
	}

	showPasswordCmd := &cobra.Command{
		Use:   "password [ID]",
		Short: i18n.T("cmd_password_short"),
		Args:  cobra.ExactArgs(1),
		Run:   showPassword,
	}
	// Update copy option description
	showPasswordCmd.Flags().BoolP("copy", "c", false, i18n.T("opt_copy"))

	changePasswordCmd := &cobra.Command{
		Use:   "change-master-password",
		Short: i18n.T("cmd_change_password_short"),
		Run:   changeMasterPassword,
	}

	exportCmd := &cobra.Command{
		Use:   "export [path]",
		Short: i18n.T("cmd_export_short"),
		Args:  cobra.ExactArgs(1),
		Run:   exportVault,
	}

	importCmd := &cobra.Command{
		Use:   "import [path]",
		Short: i18n.T("cmd_import_short"),
		Args:  cobra.ExactArgs(1),
		Run:   importVault,
	}

	updateCmd := &cobra.Command{
		Use:   "update [ID]",
		Short: i18n.T("cmd_update_short"),
		Args:  cobra.ExactArgs(1),
		Run:   updateAccount,
	}

	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: i18n.T("cmd_search_short"),
		Args:  cobra.ExactArgs(1),
		Run:   searchAccount,
	}

	exportCsvCmd := &cobra.Command{
		Use:   "export-csv [path]",
		Short: i18n.T("cmd_export_csv_short"),
		Args:  cobra.ExactArgs(1),
		Run:   exportToCsv,
	}

	rootCmd.AddCommand(
		initCmd, addCmd, generateCmd, listCmd, getCmd,
		deleteCmd, showPasswordCmd, changePasswordCmd,
		exportCmd, importCmd, updateCmd, searchCmd, exportCsvCmd,
	)

	// Add flags for generate command
	generateCmd.Flags().IntP("length", "l", 16, i18n.T("opt_length"))
	//generateCmd.Flags().BoolP("no-lowercase", "L", false, i18n.T("opt_no_lowercase"))
	generateCmd.Flags().Bool("no-lowercase", false, i18n.T("opt_no_lowercase"))
	generateCmd.Flags().BoolP("no-uppercase", "U", false, i18n.T("opt_no_uppercase"))
	generateCmd.Flags().BoolP("no-digits", "D", false, i18n.T("opt_no_digits"))
	generateCmd.Flags().BoolP("no-symbols", "S", false, i18n.T("opt_no_symbols"))
	generateCmd.Flags().BoolP("exclude-similar", "e", false, i18n.T("opt_exclude_similar"))
	generateCmd.Flags().BoolP("exclude-ambiguous", "a", false, i18n.T("opt_exclude_ambiguous"))
	generateCmd.Flags().BoolP("copy", "c", false, i18n.T("opt_copy"))

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
		os.Exit(1)
	}
}

// 读取密码（不回显）
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", errors.Wrap(err, "读取密码失败")
	}
	return string(passwordBytes), nil
}

// readAndConfirmPassword reads and confirms a password, ensuring minimum length.
func readAndConfirmPassword(prompt, confirmPrompt string, minLength int) (string, error) {
	password, err := readPassword(prompt)
	if err != nil {
		return "", err
	}

	if len(password) < minLength {
		return "", fmt.Errorf("%s", i18n.Tf("password_min_length", minLength))
	}

	confirmPassword, err := readPassword(confirmPrompt)
	if err != nil {
		return "", err
	}

	if password != confirmPassword {
		return "", fmt.Errorf("%s", i18n.T("passwords_dont_match"))
	}

	return password, nil
}

// readUserInput reads a line of input from the user.
func readUserInput(prompt string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}

// readConfirmation reads a yes/no confirmation from the user.
func readConfirmation(prompt string) bool {
	return i18n.ReadConfirmation(prompt)
}

// generateID creates a random 16-character hex ID (8 bytes).
func generateID() string {
	buf := make([]byte, 8) // 8 bytes = 16 hex chars
	_, err := rand.Read(buf)
	if err != nil {
		// This should not happen in production
		panic(fmt.Sprintf("Failed to generate random ID: %v", err))
	}
	return hex.EncodeToString(buf)
}

// initVault handles the 'init' command.
func initVault(cmd *cobra.Command, args []string) {
	if store.IsVaultExists() {
		fmt.Println(i18n.Tf("vault_exists", dataDir))
		return
	}

	// Prompt user for new master password
	masterPassword, err := readAndConfirmPassword(
		i18n.T("enter_new_master_password"),
		i18n.T("confirm_master_password"),
		8)
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
		return
	}

	// Create the vault
	if err := store.CreateVault(masterPassword); err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
		return
	}

	crypto.ClearBytes([]byte(masterPassword))
	fmt.Println(i18n.T("vault_created"))
}

// addAccount handles the 'add' command.
func addAccount(cmd *cobra.Command, args []string) {
	// Read necessary info
	platform := readUserInput(i18n.T("platform"), "")
	if platform == "" {
		fmt.Fprintf(os.Stderr, i18n.T("platform_required")+"\n")
		return
	}

	username := readUserInput(i18n.T("username"), "")
	email := readUserInput(i18n.T("email"), "")
	url := readUserInput(i18n.T("url"), "")
	notes := readUserInput(i18n.T("notes"), "")

	var password string
	var err error

	// Ask to generate or input password
	if readConfirmation(i18n.T("generate_random_password")) {
		// Generate password with default options
		opts := generator.DefaultOptions()
		password, err = generator.GeneratePassword(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
			return
		}

		fmt.Println(i18n.Tf("generated_password", password))

		// Ask to copy to clipboard
		if readConfirmation(i18n.T("copy_to_clipboard")) {
			if err := clipboard.WriteAll(password); err != nil {
				fmt.Fprintf(os.Stderr, i18n.Tf("password_copy_failed", err)+"\n")
			} else {
				fmt.Println(i18n.Tf("password_copied", 60))
				// Start goroutine to clear clipboard
				go func() {
					time.Sleep(60 * time.Second)
					clipboard.WriteAll("")
				}()
			}
		}
	} else {
		// Manually input password
		password, err = readAndConfirmPassword(i18n.T("input_password"), i18n.T("confirm_password"), 1)
		if err != nil {
			fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
			return
		}
	}

	// Create account object
	account := &model.Account{
		ID:       generateID(),
		Platform: platform,
		Username: username,
		Email:    email,
		URL:      url,
		Notes:    notes,
	}

	// Encrypt password
	encryptedPassword, err := crypto.Encrypt([]byte(password), store.GetEncryptionKey())
	if err != nil {
		crypto.ClearBytes([]byte(password))
		fmt.Fprintf(os.Stderr, i18n.Tf("encrypt_password_failed", err)+"\n")
		return
	}
	account.EncryptedPassword = encryptedPassword
	crypto.ClearBytes([]byte(password))

	// Save account
	if err := store.AddAccount(account); err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("add_account_failed", err)+"\n")
		return
	}

	fmt.Println(i18n.Tf("account_added", account.Platform, account.ID))
}

// generatePassword handles the 'generate' command.
func generatePassword(cmd *cobra.Command, args []string) {
	// Get flags
	length, _ := cmd.Flags().GetInt("length")
	noLowercase, _ := cmd.Flags().GetBool("no-lowercase")
	noUppercase, _ := cmd.Flags().GetBool("no-uppercase")
	noDigits, _ := cmd.Flags().GetBool("no-digits")
	noSymbols, _ := cmd.Flags().GetBool("no-symbols")
	excludeSimilar, _ := cmd.Flags().GetBool("exclude-similar")
	excludeAmbiguous, _ := cmd.Flags().GetBool("exclude-ambiguous")
	copyToClipboard, _ := cmd.Flags().GetBool("copy")

	// Ensure at least one character set is selected
	if noLowercase && noUppercase && noDigits && noSymbols {
		fmt.Fprintf(os.Stderr, i18n.T("must_select_charset")+"\n")
		return
	}

	// Set generation options
	opts := generator.Options{
		Length:           length,
		UseLowercase:     !noLowercase,
		UseUppercase:     !noUppercase,
		UseDigits:        !noDigits,
		UseSymbols:       !noSymbols,
		ExcludeSimilar:   excludeSimilar,
		ExcludeAmbiguous: excludeAmbiguous,
	}

	password, err := generator.GeneratePassword(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
		return
	}

	fmt.Println(i18n.Tf("generated_password", password))

	// Copy to clipboard if requested
	if copyToClipboard || readConfirmation(i18n.T("copy_to_clipboard")) {
		if err := clipboard.WriteAll(password); err != nil {
			fmt.Fprintf(os.Stderr, i18n.Tf("password_copy_failed", err)+"\n")
		} else {
			fmt.Println(i18n.Tf("password_copied", 60))
			// Start goroutine to clear clipboard
			go func() {
				time.Sleep(60 * time.Second)
				clipboard.WriteAll("")
			}()
		}
	}
}

// listAccounts handles the 'list' command.
func listAccounts(cmd *cobra.Command, args []string) {
	accounts, err := store.GetAccounts()
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
		return
	}

	if len(accounts) == 0 {
		fmt.Println(i18n.T("no_accounts"))
		return
	}

	// Create table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		i18n.T("id_header"),
		i18n.T("platform_header"),
		i18n.T("username_header"),
		i18n.T("email_header"),
	})

	// Configure table style
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("+")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")

	// Add data rows
	for _, account := range accounts {
		table.Append([]string{
			account.ID,
			account.Platform,
			account.Username,
			account.Email,
		})
	}

	table.Render()

	fmt.Printf("\n" + i18n.Tf("total_accounts", len(accounts)) + "\n")

	fmt.Println(i18n.T("view_password_hint"))
	fmt.Println(i18n.T("view_account_detail_hint"))
}

// getAccount handles the 'get' command.
func getAccount(cmd *cobra.Command, args []string) {
	platform := args[0]
	accounts, err := store.GetAccounts()
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("get_account_failed", err)+"\n")
		return
	}

	var foundAccounts []*model.Account
	for _, account := range accounts {
		if strings.Contains(strings.ToLower(account.Platform), strings.ToLower(platform)) {
			foundAccounts = append(foundAccounts, account)
		}
	}

	if len(foundAccounts) == 0 {
		fmt.Println(i18n.Tf("not_found_platform", platform))
		return
	}

	fmt.Println(i18n.Tf("found_platform_accounts", len(foundAccounts), platform))
	fmt.Println()

	for _, account := range foundAccounts {
		fmt.Printf("ID: %s\n", account.ID)
		fmt.Printf("%s: %s\n", i18n.T("platform_header"), account.Platform)

		if account.Username != "" {
			fmt.Printf("%s: %s\n", i18n.T("username_header"), account.Username)
		}

		if account.Email != "" {
			fmt.Printf("%s: %s\n", i18n.T("email_header"), account.Email)
		}

		if account.URL != "" {
			fmt.Printf("URL: %s\n", account.URL)
		}

		if account.Notes != "" {
			fmt.Printf("%s: %s\n", i18n.T("notes"), account.Notes)
		}

		fmt.Printf("%s: %s\n", i18n.T("created_at"), account.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("%s: %s\n", i18n.T("updated_at"), account.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("----------------------------------------")
	}

	fmt.Println("\n" + i18n.T("password_hint"))
}

// deleteAccount handles the 'delete' command.
func deleteAccount(cmd *cobra.Command, args []string) {
	id := args[0]

	// Get account details for confirmation prompt
	account, err := store.GetAccountByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("get_account_failed", err)+"\n")
		return
	}

	// Show confirmation info
	fmt.Printf(i18n.T("platform_header")+": %s", account.Platform)
	if account.Username != "" {
		fmt.Printf(" (%s)", account.Username)
	} else if account.Email != "" {
		fmt.Printf(" (%s)", account.Email)
	}
	fmt.Println()

	// Confirm deletion
	if !readConfirmation(i18n.T("confirm_delete_account")) {
		fmt.Println(i18n.T("operation_canceled"))
		return
	}

	// Perform deletion
	if err := store.DeleteAccount(id); err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("delete_account_failed", err)+"\n")
		return
	}

	fmt.Println(i18n.T("account_deleted"))
}

// showPassword handles the 'password' command.
func showPassword(cmd *cobra.Command, args []string) {
	id := args[0]
	copyToClipboard, _ := cmd.Flags().GetBool("copy")

	account, err := store.GetAccountByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("get_account_failed", err)+"\n")
		return
	}

	// Decrypt the password
	decryptedPasswordBytes, err := crypto.Decrypt(account.EncryptedPassword, store.GetEncryptionKey())
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("decrypt_password_failed", err)+"\n")
		return
	}
	decryptedPassword := string(decryptedPasswordBytes)
	defer crypto.ClearBytes(decryptedPasswordBytes)

	// Handle display or copy
	if copyToClipboard {
		// Copy directly to clipboard
		if err := clipboard.WriteAll(decryptedPassword); err != nil {
			fmt.Fprintf(os.Stderr, i18n.Tf("password_copy_failed", err)+"\n")
			return
		}

		fmt.Println(i18n.Tf("account_password_copied", account.Platform, 30))

		// Auto-clear clipboard
		go func() {
			time.Sleep(30 * time.Second)
			clipboard.WriteAll("")
		}()
	} else {
		// Show warning before displaying
		fmt.Println(i18n.T("password_display_warning"))
		if !readConfirmation(i18n.T("confirm_show_password")) {
			fmt.Println(i18n.T("canceled_show_password"))

			// Offer to copy instead
			if readConfirmation(i18n.T("copy_to_clipboard")) {
				if err := clipboard.WriteAll(decryptedPassword); err != nil {
					fmt.Fprintf(os.Stderr, i18n.Tf("password_copy_failed", err)+"\n")
					return
				}

				fmt.Println(i18n.Tf("account_password_copied", account.Platform, 30))

				// Auto-clear clipboard
				go func() {
					time.Sleep(30 * time.Second)
					clipboard.WriteAll("")
				}()
			}
			return
		}

		fmt.Println(i18n.Tf("account_password", account.Platform, decryptedPassword))

		// Ask if user also wants to copy it
		if readConfirmation(i18n.T("copy_to_clipboard")) {
			if err := clipboard.WriteAll(decryptedPassword); err != nil {
				fmt.Fprintf(os.Stderr, i18n.Tf("password_copy_failed", err)+"\n")
				return
			}

			fmt.Println(i18n.Tf("password_copied", 30))

			// Auto-clear clipboard
			go func() {
				time.Sleep(30 * time.Second)
				clipboard.WriteAll("")
			}()
		}
	}
}

// changeMasterPassword handles the 'change-master-password' command.
func changeMasterPassword(cmd *cobra.Command, args []string) {
	// Verify current password by attempting to unlock again
	// (Assumes vault is already unlocked by PersistentPreRun)
	oldPassword, err := readPassword(i18n.T("enter_current_master_password"))
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
		return
	}

	// Re-verify the current password (unlock is idempotent if already unlocked with same key)
	if err := store.UnlockVault(oldPassword); err != nil {
		crypto.ClearBytes([]byte(oldPassword))
		fmt.Fprintf(os.Stderr, i18n.Tf("invalid_password", err)+"\n")
		return
	}

	crypto.ClearBytes([]byte(oldPassword))

	// Enter and confirm new password
	newPassword, err := readAndConfirmPassword(
		i18n.T("enter_new_master_password"),
		i18n.T("confirm_master_password"),
		8)
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
		return
	}

	// Change the master password in storage
	if err := store.ChangeMasterPassword(newPassword); err != nil {
		crypto.ClearBytes([]byte(newPassword))
		fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
		return
	}

	crypto.ClearBytes([]byte(newPassword))
	isUnlocked = false // Reset unlocked state, forcing re-unlock with new password

	fmt.Println(i18n.T("master_password_changed"))
}

// updateAccount handles the 'update' command.
func updateAccount(cmd *cobra.Command, args []string) {
	id := args[0]

	// Get existing account data
	account, err := store.GetAccountByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("get_account_failed", err)+"\n")
		return
	}

	fmt.Println(i18n.Tf("updating_account", account.Platform, account.ID))
	fmt.Println(i18n.T("press_enter_keep_original"))

	// Update basic account info, using current value as default
	account.Platform = readUserInput(i18n.T("platform_header"), account.Platform)
	account.Username = readUserInput(i18n.T("username_header"), account.Username)
	account.Email = readUserInput(i18n.T("email_header"), account.Email)
	account.URL = readUserInput("URL", account.URL)
	account.Notes = readUserInput(i18n.T("notes"), account.Notes)

	// Ask if password should be updated
	if readConfirmation(i18n.T("update_password")) {
		var newPassword string

		if readConfirmation(i18n.T("generate_random_password")) {
			// Generate random password
			opts := generator.DefaultOptions()
			generatedPassword, err := generator.GeneratePassword(opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
				return
			}

			newPassword = generatedPassword
			fmt.Println(i18n.Tf("generated_password", newPassword))
		} else {
			// Manually input password
			manualPassword, err := readAndConfirmPassword(
				i18n.T("input_password"),
				i18n.T("confirm_password"),
				1)
			if err != nil {
				fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
				return
			}

			newPassword = manualPassword
		}

		// Encrypt the new password
		encryptedPassword, err := crypto.Encrypt([]byte(newPassword), store.GetEncryptionKey())
		if err != nil {
			crypto.ClearBytes([]byte(newPassword))
			fmt.Fprintf(os.Stderr, i18n.Tf("encrypt_password_failed", err)+"\n")
			return
		}

		account.EncryptedPassword = encryptedPassword
		crypto.ClearBytes([]byte(newPassword))

		// Ask to copy the new password
		if readConfirmation(i18n.T("copy_to_clipboard")) {
			if err := clipboard.WriteAll(newPassword); err != nil {
				fmt.Fprintf(os.Stderr, i18n.Tf("password_copy_failed", err)+"\n")
			} else {
				fmt.Println(i18n.Tf("password_copied", 60))

				go func() {
					time.Sleep(60 * time.Second)
					clipboard.WriteAll("")
				}()
			}
		}
	}

	account.UpdatedAt = time.Now()

	if err := store.UpdateAccount(account); err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("update_account_failed", err)+"\n")
		return
	}

	fmt.Println(i18n.T("account_updated"))
}

// exportVault handles the 'export' command.
func exportVault(cmd *cobra.Command, args []string) {
	exportPath := args[0]

	// Check if export file already exists and confirm overwrite
	if _, err := os.Stat(exportPath); err == nil {
		if !readConfirmation(i18n.T("export_exists")) {
			fmt.Println(i18n.T("operation_canceled"))
			return
		}
	}

	// Perform the export
	if err := store.ExportVault(exportPath); err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
		return
	}

	fmt.Println(i18n.Tf("export_success", exportPath))
	fmt.Println(i18n.T("export_warning"))
}

// importVault handles the 'import' command.
func importVault(cmd *cobra.Command, args []string) {
	importPath := args[0]

	// Check if import file exists
	if _, err := os.Stat(importPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, i18n.Tf("error", "导入文件不存在: "+importPath)+"\n")
		return
	}

	// Show warning and confirm
	fmt.Println(i18n.T("import_warning"))
	if !readConfirmation(i18n.T("confirm_operation")) {
		fmt.Println(i18n.T("operation_canceled"))
		return
	}

	if err := store.ImportVault(importPath); err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("import_failed", err)+"\n")
		return
	}

	isUnlocked = false // Reset unlocked state after successful import
	fmt.Println(i18n.T("import_success"))
}

// searchAccount handles the 'search' command.
func searchAccount(cmd *cobra.Command, args []string) {
	query := args[0]

	results, err := store.SearchAccounts(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
		return
	}

	if len(results) == 0 {
		fmt.Println(i18n.Tf("not_found_query", query))
		return
	}

	fmt.Println(i18n.Tf("found_query_accounts", len(results), query))
	fmt.Println()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		i18n.T("id_header"),
		i18n.T("platform_header"),
		i18n.T("username_header"),
		i18n.T("email_header"),
	})

	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("+")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")

	for _, account := range results {
		table.Append([]string{
			account.ID,
			account.Platform,
			account.Username,
			account.Email,
		})
	}

	table.Render()

	fmt.Println(i18n.T("view_password_hint"))
	fmt.Println(i18n.T("view_account_detail_hint"))
}

// exportToCsv handles the 'export-csv' command.
func exportToCsv(cmd *cobra.Command, args []string) {
	exportPath := args[0]

	// Show security warnings
	fmt.Println(i18n.T("csv_export_warning"))
	fmt.Println(i18n.T("csv_post_export"))

	if !readConfirmation(i18n.T("confirm_operation")) {
		fmt.Println(i18n.T("operation_canceled"))
		return
	}

	// Check if export file exists and confirm overwrite
	if _, err := os.Stat(exportPath); err == nil {
		if !readConfirmation(i18n.T("export_exists")) {
			fmt.Println(i18n.T("operation_canceled"))
			return
		}
	}

	// Perform CSV export
	if err := store.ExportToCSV(exportPath); err != nil {
		fmt.Fprintf(os.Stderr, i18n.Tf("error", err.Error())+"\n")
		return
	}

	fmt.Println(i18n.Tf("csv_export_success", exportPath))
	fmt.Println(i18n.T("csv_post_export"))
}
