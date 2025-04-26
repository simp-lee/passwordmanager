package i18n

import (
	"fmt"
	"strings"
)

// 语言类型
type Lang string

const (
	LangZH Lang = "zh"
	LangEN Lang = "en"
)

// 当前语言
var currentLang = LangEN

// 翻译映射
var translations = map[Lang]map[string]string{
	LangZH: {
		// 通用
		"error":                     "错误: %s",
		"warning":                   "警告: %s",
		"success":                   "成功: %s",
		"yes_or_no":                 "(y/n): ",
		"confirm_operation":         "确定要继续吗?",
		"operation_canceled":        "已取消操作",
		"press_enter_keep_original": "提示: 对于不需要修改的字段，直接按回车键保持原值",

		// 主密码相关
		"enter_master_password":         "请输入主密码: ",
		"enter_new_master_password":     "请输入新主密码（至少8个字符）: ",
		"enter_current_master_password": "请输入当前主密码: ",
		"confirm_master_password":       "请确认主密码: ",
		"passwords_dont_match":          "两次输入的密码不匹配",
		"password_min_length":           "密码长度至少需要 %d 个字符",
		"master_password_changed":       "主密码已成功更改",
		"invalid_password":              "验证当前密码失败: %v",

		// 初始化保险库
		"vault_exists":     "保险库已存在。如需重置，请先删除数据目录: %s",
		"vault_created":    "保险库已成功创建！",
		"vault_not_exists": "保险库不存在，请运行 'passwordmanager init' 创建新保险库",

		// 输入字段
		"platform":          "输入平台名称*",
		"platform_required": "错误: 平台名称为必填项",
		"username":          "输入用户名",
		"email":             "输入邮箱",
		"url":               "输入网址",
		"notes":             "输入备注",
		"input_password":    "输入密码: ",
		"confirm_password":  "确认密码: ",

		// 密码生成
		"generate_random_password": "是否生成随机密码?",
		"generated_password":       "已生成密码: %s",
		"password_length":          "密码长度",
		"copy_to_clipboard":        "是否复制到剪贴板?",
		"password_copied":          "密码已复制到剪贴板，%d秒后自动清除",
		"password_copy_failed":     "复制到剪贴板失败: %v",
		"must_select_charset":      "错误: 至少需要选择一种字符类型",

		// 账户操作
		"account_added":           "账户 '%s' 添加成功 (ID: %s)",
		"account_updated":         "账户更新成功",
		"updating_account":        "正在更新账户: %s (ID: %s)",
		"account_deleted":         "账户已成功删除",
		"confirm_delete_account":  "确定要删除此账户吗?",
		"add_account_failed":      "添加账户失败: %v",
		"update_account_failed":   "更新账户失败: %v",
		"delete_account_failed":   "删除账户失败: %v",
		"get_account_failed":      "获取账户失败: %v",
		"encrypt_password_failed": "加密密码失败: %v",
		"decrypt_password_failed": "解密密码失败: %v",

		// 显示密码
		"password_display_warning": "警告：显示密码可能会被屏幕记录或被他人看到",
		"confirm_show_password":    "确定要显示密码吗?",
		"canceled_show_password":   "已取消显示密码",
		"account_password":         "账户 '%s' 的密码: %s",
		"account_password_copied":  "账户 '%s' 的密码已复制到剪贴板，%d秒后自动清除",
		"update_password":          "是否更新密码?",

		// 账户列表和搜索
		"no_accounts":             "保险库中没有账户记录",
		"total_accounts":          "共 %d 个账户",
		"password_hint":           "提示: 使用 'passwordmanager password <ID>' 查看账户密码",
		"not_found_platform":      "未找到平台为 '%s' 的账户",
		"found_platform_accounts": "找到 %d 个平台为 '%s' 的账户:",
		"not_found_query":         "未找到匹配 '%s' 的账户",
		"found_query_accounts":    "找到 %d 个匹配 '%s' 的账户:",

		// 导入导出
		"export_success":     "保险库已成功导出到: %s",
		"export_warning":     "注意: 导出文件包含所有账户数据，请妥善保管",
		"export_exists":      "导出文件已存在，是否覆盖?",
		"import_warning":     "警告: 导入操作将覆盖现有保险库数据!",
		"import_success":     "保险库导入成功。请使用其原始主密码解锁。",
		"import_failed":      "导入保险库失败: %v",
		"csv_export_warning": "警告: CSV导出将包含明文密码！这是一个安全风险。",
		"csv_post_export":    "安全提示: 请在使用完毕后删除该文件",
		"csv_export_success": "账户数据已成功导出到: %s",

		// 表格标题
		"id_header":       "ID",
		"platform_header": "平台",
		"username_header": "用户名",
		"email_header":    "邮箱",
		"created_at":      "创建时间",
		"updated_at":      "更新时间",

		// 应用名称和描述
		"app_name":        "密码管理器",
		"app_description": "一个简单的密码管理器，用于安全地存储和管理您的密码。",

		// 命令名称和描述
		"cmd_init_short":            "初始化密码管理器",
		"cmd_add_short":             "添加新账户",
		"cmd_generate_short":        "生成新密码",
		"cmd_list_short":            "列出所有账户",
		"cmd_get_short":             "获取特定平台的账户详情",
		"cmd_delete_short":          "通过ID删除账户",
		"cmd_password_short":        "显示特定账户的密码并复制到剪贴板",
		"cmd_change_password_short": "更改主密码",
		"cmd_export_short":          "导出加密的保险库到文件",
		"cmd_import_short":          "从文件导入保险库（会覆盖现有数据！）",
		"cmd_update_short":          "更新现有账户",
		"cmd_search_short":          "按平台、用户名或邮箱搜索账户",
		"cmd_export_csv_short":      "导出账户到CSV文件（明文密码！）",

		// 选项描述
		"opt_lang":              "语言 (zh/en)",
		"opt_length":            "密码长度",
		"opt_no_lowercase":      "不使用小写字母",
		"opt_no_uppercase":      "不使用大写字母",
		"opt_no_digits":         "不使用数字",
		"opt_no_symbols":        "不使用特殊符号",
		"opt_exclude_similar":   "排除相似字符 (例如 l, 1, I, O, 0)",
		"opt_exclude_ambiguous": "排除可能混淆的字符 (例如 {}, [], (), /)",
		"opt_copy":              "直接复制到剪贴板",

		// 查看账户后缀提示
		"view_password_hint":       "\n要查看某个账户的密码，请使用命令:\npasswordmanager password <ID>",
		"view_account_detail_hint": "\n要查看账户详情，请使用命令:\npasswordmanager get <platform>",
	},
	LangEN: {
		// 通用
		"error":                     "Error: %s",
		"warning":                   "Warning: %s",
		"success":                   "Success: %s",
		"yes_or_no":                 "(y/n): ",
		"confirm_operation":         "Are you sure you want to continue?",
		"operation_canceled":        "Operation canceled",
		"press_enter_keep_original": "Tip: For fields you don't want to change, just press Enter to keep the original value",

		// 主密码相关
		"enter_master_password":         "Enter master password: ",
		"enter_new_master_password":     "Enter new master password (at least 8 characters): ",
		"enter_current_master_password": "Enter current master password: ",
		"confirm_master_password":       "Confirm master password: ",
		"passwords_dont_match":          "Passwords don't match",
		"password_min_length":           "Password must be at least %d characters",
		"master_password_changed":       "Master password changed successfully",
		"invalid_password":              "Failed to verify current password: %v",

		// 初始化保险库
		"vault_exists":     "Vault already exists. To reset, delete the data directory first: %s",
		"vault_created":    "Vault created successfully!",
		"vault_not_exists": "Vault does not exist. Run 'passwordmanager init' to create a new vault",

		// 输入字段
		"platform":          "Enter platform name*",
		"platform_required": "Error: Platform name is required",
		"username":          "Enter username",
		"email":             "Enter email",
		"url":               "Enter URL",
		"notes":             "Enter notes",
		"input_password":    "Enter password: ",
		"confirm_password":  "Confirm password: ",

		// 密码生成
		"generate_random_password": "Generate random password?",
		"generated_password":       "Generated password: %s",
		"password_length":          "Password length",
		"copy_to_clipboard":        "Copy to clipboard?",
		"password_copied":          "Password copied to clipboard, will be cleared in %d seconds",
		"password_copy_failed":     "Failed to copy to clipboard: %v",
		"must_select_charset":      "Error: You must select at least one character type",

		// 账户操作
		"account_added":           "Account '%s' added successfully (ID: %s)",
		"account_updated":         "Account updated successfully",
		"updating_account":        "Updating account: %s (ID: %s)",
		"account_deleted":         "Account deleted successfully",
		"confirm_delete_account":  "Are you sure you want to delete this account?",
		"add_account_failed":      "Failed to add account: %v",
		"update_account_failed":   "Failed to update account: %v",
		"delete_account_failed":   "Failed to delete account: %v",
		"get_account_failed":      "Failed to get account: %v",
		"encrypt_password_failed": "Failed to encrypt password: %v",
		"decrypt_password_failed": "Failed to decrypt password: %v",

		// 显示密码
		"password_display_warning": "Warning: Displaying password may be recorded on screen or seen by others",
		"confirm_show_password":    "Are you sure you want to display the password?",
		"canceled_show_password":   "Password display canceled",
		"account_password":         "Password for account '%s': %s",
		"account_password_copied":  "Password for account '%s' copied to clipboard, will be cleared in %d seconds",
		"update_password":          "Update password?",

		// 账户列表和搜索
		"no_accounts":             "No accounts in the vault",
		"total_accounts":          "Total: %d accounts",
		"password_hint":           "Tip: Use 'passwordmanager password <ID>' to view an account's password",
		"not_found_platform":      "No accounts found for platform '%s'",
		"found_platform_accounts": "Found %d accounts for platform '%s':",
		"not_found_query":         "No accounts found matching '%s'",
		"found_query_accounts":    "Found %d accounts matching '%s':",

		// 导入导出
		"export_success":     "Vault exported successfully to: %s",
		"export_warning":     "Note: The export file contains all account data, keep it safe",
		"export_exists":      "Export file already exists, overwrite?",
		"import_warning":     "Warning: Importing will overwrite your existing vault data!",
		"import_success":     "Vault imported successfully. Use its original master password to unlock.",
		"import_failed":      "Failed to import vault: %v",
		"csv_export_warning": "Warning: CSV export will contain plaintext passwords! This is a security risk.",
		"csv_post_export":    "Security tip: Delete the file after use",
		"csv_export_success": "Account data exported successfully to: %s",

		// 表格标题
		"id_header":       "ID",
		"platform_header": "Platform",
		"username_header": "Username",
		"email_header":    "Email",
		"created_at":      "Created At",
		"updated_at":      "Updated At",

		// 应用名称和描述
		"app_name":        "Password Manager",
		"app_description": "A simple password manager for securely storing and managing your passwords.",

		// 命令名称和描述
		"cmd_init_short":            "Initialize password manager",
		"cmd_add_short":             "Add a new account",
		"cmd_generate_short":        "Generate a new password",
		"cmd_list_short":            "List all accounts",
		"cmd_get_short":             "Get account details for a specific platform",
		"cmd_delete_short":          "Delete an account by ID",
		"cmd_password_short":        "Show or copy password for a specific account",
		"cmd_change_password_short": "Change master password",
		"cmd_export_short":          "Export encrypted vault to a file",
		"cmd_import_short":          "Import vault from a file (overwrites existing data!)",
		"cmd_update_short":          "Update an existing account",
		"cmd_search_short":          "Search accounts by platform, username, or email",
		"cmd_export_csv_short":      "Export accounts to CSV file (plaintext passwords!)",

		// 选项描述
		"opt_lang":              "Language (zh/en)",
		"opt_length":            "Password length",
		"opt_no_lowercase":      "Don't use lowercase letters",
		"opt_no_uppercase":      "Don't use uppercase letters",
		"opt_no_digits":         "Don't use digits",
		"opt_no_symbols":        "Don't use symbols",
		"opt_exclude_similar":   "Exclude similar characters (e.g. l, 1, I, O, 0)",
		"opt_exclude_ambiguous": "Exclude ambiguous characters (e.g. {}, [], (), /)",
		"opt_copy":              "Copy directly to clipboard",

		// 查看账户后缀提示
		"view_password_hint":       "\nTo view an account's password, use command:\npasswordmanager password <ID>",
		"view_account_detail_hint": "\nTo view account details, use command:\npasswordmanager get <platform>",
	},
}

// SetLanguage 设置当前语言
func SetLanguage(lang string) {
	switch lang {
	case "zh":
		currentLang = LangZH
	default:
		currentLang = LangEN
	}
}

// T 获取翻译文本
func T(key string) string {
	if text, ok := translations[currentLang][key]; ok {
		return text
	}
	// 如果没有找到翻译，返回键名
	return key
}

// Tf 格式化翻译文本
func Tf(key string, args ...any) string {
	text := T(key)
	// 如果有格式化参数，使用fmt.Sprintf格式化文本
	if len(args) > 0 {
		return fmt.Sprintf(text, args...)
	}
	return text
}

// ReadConfirmation 读取确认输入
func ReadConfirmation(prompt string) bool {
	fmt.Print(prompt + " " + T("yes_or_no"))
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// GetCurrentLanguage 获取当前语言
func GetCurrentLanguage() string {
	return string(currentLang)
}
