# Password Manager User Guide

[English](#introduction) | [简体中文](#简介)

## Introduction

This is a command-line password manager for securely storing and managing account information for various websites and applications. It uses strong encryption to protect your password data and provides a rich set of management features. This program supports both Chinese (zh) and English (en) language interfaces.

## Features

- AES-256 encryption to protect your password data
- Generation of high-strength random passwords
- Secure storage of account information including usernames, passwords, URLs, etc.
- Import/export of vault data
- Master password changing
- Account search and filtering
- Password copying to clipboard with automatic clearing
- Bilingual interface supporting Chinese and English

## Installation

### Compile from Source

```bash
git clone https://github.com/simp-lee/passwordmanager.git
cd passwordmanager
go build -o passwordmanager cmd/main.go
```

After compilation, move the binary to your PATH:

```bash
sudo mv passwordmanager /usr/local/bin/
```

## Language Settings

This password manager uses English (en) interface by default. You can switch the language using the -L or --lang parameter:

```bash
# Use Chinese interface
passwordmanager -L zh list

# Use English interface
passwordmanager -L en list

# Default is English
passwordmanager list
```

Language settings apply to all commands, and you can switch languages anytime based on your preference. For example:

```bash
# Initialize with Chinese interface
passwordmanager -L zh init

# Add an account with English interface
passwordmanager -L en add
```

## Basic Usage

### Initialize Vault

Before first use, you need to create a password vault and set a master password:

```bash
passwordmanager init
```

The system will prompt you to enter a master password. Please use a strong password and make sure you don't forget it, as the master password cannot be recovered.

### Add New Account

```bash
passwordmanager add
```

Follow the prompts to enter account information, including platform name, username, email, etc. You can choose to manually enter a password or generate a random one.

### Generate Random Password

```bash
passwordmanager generate
```

By default, this generates a 16-character strong password that includes uppercase and lowercase letters, numbers, and special symbols. You can customize the generated password using the following options:

- `-l, --length` - Specify password length (default 16)
- `--no-lowercase` - Don't use lowercase letters
- `-U, --no-uppercase` - Don't use uppercase letters
- `-D, --no-digits` - Don't use numbers
- `-S, --no-symbols` - Don't use special symbols
- `-e, --exclude-similar` - Exclude similar characters (like l, 1, I, O, 0)
- `-a, --exclude-ambiguous` - Exclude potentially confusing characters (like {}, [], (), /)
- `-c, --copy` - Copy directly to clipboard

Example:
```bash
# Generate a 20-character password without special symbols and similar characters
passwordmanager generate -l 20 -S -e
```

### List All Accounts

```bash
passwordmanager list
```

Displays basic information for all accounts in the vault, including ID (16 hexadecimal character unique identifier), platform, username, and email.

### View Account Details

```bash
passwordmanager get <platform-name>
```

Displays detailed information for all accounts containing the specified platform name, but doesn't include passwords. The search is based on inclusion, not exact matching.

### View and Copy Password

```bash
passwordmanager password <account-ID>
```

Displays the password for the specified account ID. For security reasons, it will first ask if you are sure you want to display the password. If you choose not to display it, the system will offer the option to copy it to the clipboard.

Use the `-c` or `--copy` option to copy the password directly to the clipboard without displaying it:

```bash
passwordmanager password <account-ID> -c
```

Passwords copied to the clipboard will be automatically cleared after 30 seconds to enhance security.

### Update Account Information

```bash
passwordmanager update <account-ID>
```

Updates information for the account with the specified ID, including platform, username, email, password, etc. For fields you don't want to modify, simply press Enter to keep the original value.

### Delete Account

```bash
passwordmanager delete <account-ID>
```

Removes the account with the specified ID from the vault. Confirmation will be required before deletion.

### Search Accounts

```bash
passwordmanager search <query>
```

Searches for accounts where the platform name, username, or email contains the specified query.

## Advanced Features

### Change Master Password

```bash
passwordmanager change-master-password
```

Changes the master password for the vault. You'll need to enter your current master password for verification first.

### Export Vault

```bash
passwordmanager export <export-path>
```

Exports the encrypted vault to the specified file. The exported file remains encrypted and requires the master password to open.

### Import Vault

```bash
passwordmanager import <import-path>
```

Imports vault data from the specified file. This operation will overwrite the current vault data.

### Export to CSV File

```bash
passwordmanager export-csv <export-path>
```

Exports account data to CSV format. **Warning: Passwords in CSV files are stored in plain text. Handle the exported file with care.**

## Security Features

1. **Strong Encryption**: Uses AES-256-GCM encryption and PBKDF2 key derivation with 100,000 iterations
2. **Memory Safety**: All sensitive data (such as plaintext passwords) is securely cleared from memory immediately after use
3. **Clipboard Protection**: Passwords copied to the clipboard are automatically cleared after 30 seconds
4. **Input Protection**: Password input is not displayed on the screen
5. **Security Confirmation**: Dangerous operations (such as deleting accounts or exporting plain text) require confirmation

## Security Recommendations

1. **Master Password Security**: Use a strong password as your master password, including uppercase and lowercase letters, numbers, and special characters.
2. **Regular Backups**: Use the `export` command to back up your vault regularly.
3. **Regular Password Updates**: Periodically update passwords for important accounts.
4. **Avoid Plain Text Export**: Unless necessary, avoid using `export-csv` to export passwords in plain text.

## Troubleshooting

### Cannot Open Vault

If you forget your master password, vault data cannot be recovered. This is to ensure security.

### Vault File Corruption

If the vault file is corrupted, you can try to recover using a recent backup file:

```bash
passwordmanager import <backup-file-path>
```

### Data Directory Location

By default, vault data is stored in the `.passwordmanager` folder in the user's home directory:

- Windows: `C:\Users\<username>\.passwordmanager\`
- Linux/macOS: `~/.passwordmanager/`

## Command Reference

| Command                   | Description                                          |
|---------------------------|------------------------------------------------------|
| `init`                    | Initialize the password manager                      |
| `add`                     | Add a new account                                    |
| `generate`                | Generate a new password                              |
| `list`                    | List all accounts                                    |
| `get [platform]`          | Get details for accounts with the specified platform |
| `password [ID]`           | Show or copy the password for a specific account     |
| `update [ID]`             | Update an existing account                           |
| `delete [ID]`             | Delete an account by ID                              |
| `search [query]`          | Search accounts by platform, username, or email      |
| `change-master-password`  | Change the master password                           |
| `export [path]`           | Export the encrypted vault to a file                 |
| `import [path]`           | Import a vault from a file                           |
| `export-csv [path]`       | Export accounts to a CSV file (plain text passwords!)|

## Example Scenarios

### Create and Store a New Bank Account

```bash
# Initialize the vault
passwordmanager init

# Add a bank account
passwordmanager add
# Platform: MyBank
# Username: myusername
# Password: Choose "Generate random password"
# ...
```

### Find and Use a Password

```bash
# Search for the bank account
passwordmanager search MyBank

# Copy the password to clipboard (assuming ID is 4373126014574e97)
passwordmanager password 4373126014574e97 -c
```

### Update an Expired Password

```bash
# Generate a new password
passwordmanager generate -l 20 -e

# Update the account password (assuming ID is 4373126014574e97)
passwordmanager update 4373126014574e97
# Choose "Update password?: y"
# Choose "Generate random password?: y"
```

### Backup Vault

```bash
# Export encrypted vault
passwordmanager export ~/backup/vault-backup-2025-04-26.encrypted
```

---

# 密码管理器使用说明

[English](#introduction) | [简体中文](#简介)

## 简介

这是一个基于命令行的密码管理器，用于安全地存储和管理您的各种网站和应用的账户信息。它使用强加密技术保护您的密码数据，并提供了丰富的管理功能。本程序支持中文(zh)和英文(en)两种语言界面。

## 功能特点

- 使用 AES-256 加密保护您的密码数据
- 生成高强度随机密码
- 安全存储账户的用户名、密码、URL等信息
- 支持导入/导出保险库数据
- 支持修改主密码
- 支持搜索和过滤账户信息
- 支持密码复制到剪贴板并自动清除
- 支持中英文双语界面

## 安装

### 从源码编译

```bash
git clone https://github.com/simp-lee/passwordmanager.git
cd passwordmanager
go build -o passwordmanager cmd/main.go
```

安装后，将编译好的二进制文件移动到您的 PATH 路径中：

```bash
sudo mv passwordmanager /usr/local/bin/
```

## 语言设置

本密码管理器默认使用英文(en)界面。您可以通过 -L 或 --lang 参数切换语言：

```bash
# 使用中文界面
passwordmanager -L zh list

# 使用英文界面
passwordmanager -L en list

# 默认为英文
passwordmanager list
```

语言设置适用于所有命令，您可以根据个人偏好随时切换。例如：

```bash
# 使用中文界面初始化
passwordmanager -L zh init

# 使用英文界面添加账户
passwordmanager -L en add
```

## 基本使用

### 初始化保险库

首次使用前，需要创建密码保险库并设置主密码：

```bash
passwordmanager init
```

系统会提示您输入主密码。请使用强密码并确保不会忘记，因为主密码无法恢复。

### 添加新账户

```bash
passwordmanager add
```

按照提示输入账户信息，包括平台名称、用户名、邮箱等。您可以选择手动输入密码或生成随机密码。

### 生成随机密码

```bash
passwordmanager generate
```

默认生成一个16位的强密码，包含大小写字母、数字和特殊符号。您可以使用以下选项自定义生成的密码：

- `-l, --length` - 指定密码长度（默认16）
- `--no-lowercase` - 不使用小写字母
- `-U, --no-uppercase` - 不使用大写字母
- `-D, --no-digits` - 不使用数字
- `-S, --no-symbols` - 不使用特殊符号
- `-e, --exclude-similar` - 排除相似字符（如 l, 1, I, O, 0）
- `-a, --exclude-ambiguous` - 排除可能混淆的字符（如 {}, [], (), /）
- `-c, --copy` - 直接复制到剪贴板

示例：
```bash
# 生成一个20位的密码，不包含特殊符号和相似字符
passwordmanager generate -l 20 -S -e
```

### 列出所有账户

```bash
passwordmanager list
```

显示保险库中所有账户的基本信息，包括ID（16个十六进制字符的唯一标识符）、平台、用户名和邮箱。

### 查看账户详情

```bash
passwordmanager get <平台名称>
```

显示包含指定平台名称的所有账户详细信息，但不包括密码。搜索是基于包含关系，而非精确匹配。

### 查看和复制密码

```bash
passwordmanager password <账户ID>
```

显示指定ID账户的密码。出于安全考虑，会先询问是否确定要显示密码。如果选择不显示，系统会提供复制到剪贴板的选项。

使用 `-c` 或 `--copy` 选项可以直接复制密码到剪贴板而不显示：

```bash
passwordmanager password <账户ID> -c
```

复制到剪贴板的密码将在30秒后自动清除，以增强安全性。

### 更新账户信息

```bash
passwordmanager update <账户ID>
```

更新指定ID账户的信息，包括平台、用户名、邮箱、密码等。对于不需要修改的字段，直接按回车键保持原值。

### 删除账户

```bash
passwordmanager delete <账户ID>
```

从保险库中删除指定ID的账户。操作前会要求确认。

### 搜索账户

```bash
passwordmanager search <关键词>
```

搜索平台名称、用户名或邮箱中包含指定关键词的账户。

## 高级功能

### 更改主密码

```bash
passwordmanager change-master-password
```

更改保险库的主密码。需要先输入当前主密码进行验证。

### 导出保险库

```bash
passwordmanager export <导出路径>
```

将加密的保险库导出到指定文件。导出的文件仍然是加密的，需要主密码才能打开。

### 导入保险库

```bash
passwordmanager import <导入路径>
```

从指定文件导入保险库数据。此操作会覆盖当前的保险库数据。

### 导出为CSV文件

```bash
passwordmanager export-csv <导出路径>
```

将账户数据导出为CSV格式。**警告：CSV文件中的密码是明文存储的，请妥善保管导出的文件**。

## 安全特性

1. **强加密**：使用AES-256-GCM加密和PBKDF2密钥派生，迭代次数为100,000
2. **内存安全**：所有敏感数据（如明文密码）在使用后会立即从内存中安全清除
3. **剪贴板保护**：复制到剪贴板的密码会在30秒后自动清除
4. **输入保护**：密码输入不会显示在屏幕上
5. **安全确认**：危险操作（如删除账户、导出明文）前需要确认

## 安全性建议

1. **主密码安全**：使用强密码作为主密码，包含大小写字母、数字和特殊字符。
2. **定期备份**：使用 `export` 命令定期备份您的保险库。
3. **定期更新密码**：定期为重要账户更新密码。
4. **避免明文导出**：除非必要，避免使用 `export-csv` 导出明文密码。

## 故障排除

### 无法打开保险库

如果忘记了主密码，保险库数据无法恢复。这是为了保证安全性。

### 保险库文件损坏

如果保险库文件损坏，可以尝试使用最近的备份文件恢复：

```bash
passwordmanager import <备份文件路径>
```

### 数据目录位置

默认情况下，保险库数据存储在用户主目录下的 `.passwordmanager` 文件夹中：

- Windows: `C:\Users\<用户名>\.passwordmanager\`
- Linux/macOS: `~/.passwordmanager/`

## 命令参考

| 命令                      | 描述                        |
|--------------------------|-----------------------------|
| `init`                   | 初始化密码管理器               |
| `add`                    | 添加新账户                    |
| `generate`               | 生成新密码                    |
| `list`                   | 列出所有账户                  |
| `get [platform]`         | 获取包含指定平台名称的账户详情    |
| `password [ID]`          | 显示特定账户的密码或复制到剪贴板  |
| `update [ID]`            | 更新现有账户                  |
| `delete [ID]`            | 通过ID删除账户                |
| `search [query]`         | 按平台、用户名或邮箱搜索账户     |
| `change-master-password` | 更改主密码                    |
| `export [path]`          | 导出加密的保险库到文件          |
| `import [path]`          | 从文件导入保险库               |
| `export-csv [path]`      | 导出账户到CSV文件（明文密码！）  |

## 示例场景

### 创建新的银行账户并存储

```bash
# 初始化保险库
passwordmanager init

# 添加银行账户
passwordmanager add
# 平台名称: MyBank
# 用户名: myusername
# 密码: 选择"生成随机密码"
# ...
```

### 查找并使用密码

```bash
# 查找银行账户
passwordmanager search MyBank

# 复制密码到剪贴板（假设ID为 4373126014574e97）
passwordmanager password 4373126014574e97 -c
```

### 更新过期的密码

```bash
# 生成新密码
passwordmanager generate -l 20 -e

# 更新账户密码（假设ID为 4373126014574e97）
passwordmanager update 4373126014574e97
# 选择"是否更新密码?: y"
# 选择"是否生成随机密码?: y"
```

### 备份保险库

```bash
# 导出加密保险库
passwordmanager export ~/backup/vault-backup-2025-04-26.encrypted
```