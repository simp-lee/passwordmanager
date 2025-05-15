document.addEventListener('alpine:init', () => {
    Alpine.data('passwordManager', () => ({
        // 保留原有的核心状态变量
        isUnlocked: false,
        isVaultExists: false,
        masterPassword: '',
        confirmMasterPassword: '',
        errorMessage: '',
        accounts: [],
        filteredAccounts: [],
        allAccounts: [], // 用于存储所有账户的原始数据
        selectedAccountId: null,
        selectedAccount: null,
        decryptedPassword: '',
        showPassword: false,
        isEditing: false,
        isNewAccount: false,
        editingAccount: {
            id: '',
            platform: '',
            username: '',
            email: '',
            url: '',
            notes: '',
            group: '',
        },
        editingPassword: '',
        showEditPassword: false,
        showPasswordOptions: false,
        searchQuery: '',
        notification: {
            show: false,
            message: '',
        },
        passwordStrengthInfo: {
            show: false,
            text: ''
        },

        // 更改主密码相关
        isChangingMasterPassword: false,
        currentMasterPassword: '',
        newMasterPassword: '',
        confirmNewMasterPassword: '',

        // 密码生成选项
        generateOptions: {
            Length: 16,
            UseLowercase: true,
            UseUppercase: true,
            UseDigits: true,
            UseSymbols: true,
            ExcludeSimilar: false,
            ExcludeAmbiguous: false,
        },

        // 排序相关变量
        sortable: null,

        // 分组相关变量
        groups: [],
        selectedGroup: null,
        showAllAccounts: true,

        // 初始化方法
        async init() {
            try {
                this.isVaultExists = await window.go.backend.App.IsVaultExists();
            } catch (error) {
                console.error('初始化错误:', error);
            }

            // 监听搜索输入变化
            this.$watch('searchQuery', (newValue, oldValue) => {
                // 如果输入从有到无，重新初始化排序
                if (oldValue && !newValue) {
                    // 根据当前模式选择加载方式
                    if (!this.showAllAccounts && this.selectedGroup !== null) {
                        this.loadAccountsByGroup(this.selectedGroup);
                    } else {
                        this.loadAccounts();
                    }
                }
            });

            // // 监听过滤列表变化（如果有的话）
            // this.$watch('filteredAccounts.length', () => {
            //     this.$nextTick(() => {
            //         this.initSortable();
            //     });
            // });
        },

        // 初始化拖拽排序
        initSortable() {
            // 销毁已有实例
            if (this.sortable) {
                this.sortable.destroy();
                this.sortable = null;
            }

            const el = document.getElementById('accounts-list');
            if (!el || !this.isUnlocked) return;

            // 分组排序模式标记
            if (!this.showAllAccounts && this.selectedGroup !== null) {
                el.classList.add('group-sort-mode');
            } else {
                el.classList.remove('group-sort-mode');
            }

            // 搜索模式下禁用排序并添加视觉提示
            if (this.searchQuery && this.searchQuery.trim()) {
                el.classList.add('search-mode');

                // 为所有拖拽手柄添加提示
                const dragHandles = document.querySelectorAll('.drag-handle');
                // 设置分组内搜索提示
                const tipText = !this.showAllAccounts && this.selectedGroup !== null ?
                    '分组内搜索中，清空搜索后可排序' : '清空搜索后可排序账户';

                dragHandles.forEach(handle => {
                    handle.setAttribute('title', tipText);
                });

                return; // 不初始化排序
            } else {
                el.classList.remove('search-mode');

                // 移除工具提示
                const dragHandles = document.querySelectorAll('.drag-handle');
                dragHandles.forEach(handle => {
                    handle.removeAttribute('title');
                });
            }

            // 初始化 Sortable
            this.sortable = new Sortable(el, {
                animation: 150, // 动画持续时间
                ghostClass: 'sortable-ghost', // 拖动时"影子"元素的样式
                chosenClass: 'sortable-chosen', // 被选中项的样式
                dragClass: "sortable-drag", // 拖动时元素的样式
                handle: '.drag-handle',
                scrollSensitivity: 30, // 滚动灵敏度
                scrollSpeed: 10,       // 滚动速度

                onStart: (evt) => {
                    // 开始拖动时，修改鼠标样式
                    document.body.classList.add('cursor-grabbing');
                },
                onEnd: (evt) => {
                    // 结束拖动时，恢复鼠标样式
                    document.body.classList.remove('cursor-grabbing');

                    // 仅在解锁状态下保存排序
                    if (this.isUnlocked && !this.searchQuery.trim()) {
                        this.saveAccountsOrder();
                    }
                }
            });
        },

        // 保存账户排序
        async saveAccountsOrder() {
            if (!this.isUnlocked) return;

            try {
                // 获取当前排序顺序
                const accountItems = document.querySelectorAll('.account-item');
                const newOrder = Array.from(accountItems).map(item => item.getAttribute('data-id'));

                // 如果是分组排序
                if (!this.showAllAccounts && this.selectedGroup !== null) {
                    await window.go.backend.App.UpdateAccountsOrderInGroup(newOrder, this.selectedGroup);

                    // 针对未分组情况显示特定消息
                    if (this.selectedGroup === '') {
                        this.showNotification('未分组账户排序已保存');
                    } else {
                        this.showNotification(`"${this.selectedGroup}"分组内排序已保存`);
                    }
                } else {
                    // 全局排序
                    await window.go.backend.App.UpdateAccountsOrder(newOrder);
                    this.showNotification('账户排序已保存');
                }
            } catch (error) {
                console.error('保存排序失败:', error);
                this.showNotification('排序保存失败');
            }
        },

        // 解锁后加载分组
        async loadGroups() {
            try {
                this.groups = await window.go.backend.App.GetAllGroups();
            } catch (error) {
                console.error('加载分组失败:', error);
            }
        },

        // 修改选择分组的方法
        selectGroup(group) {
            // 如果点击的是当前已选中的分组，不做任何操作
            if (this.selectedGroup === group) {
                return;
            }

            this.selectedGroup = group;
            this.showAllAccounts = false;
            this.searchQuery = '';
            this.loadAccountsByGroup(group);
        },

        // 显示所有账户
        viewAllAccounts() {
            // 如果已经在查看所有账户，不重新加载数据
            if (this.showAllAccounts && this.selectedGroup === null) {
                return;
            }

            this.showAllAccounts = true;
            this.selectedGroup = null;
            this.searchQuery = '';
            this.loadAccounts();
        },

        // 根据分组加载账户
        async loadAccountsByGroup(group) {
            try {
                // 销毁旧的sortable实例
                if (this.sortable) {
                    this.sortable.destroy();
                    this.sortable = null;
                }

                // 将 filteredAccounts 置空，通知 Alpine.js 数据已清空
                this.filteredAccounts = [];
                // 等待 Alpine.js 处理 DOM 更新（列表变空）
                await this.$nextTick();

                const loadedAccounts = await window.go.backend.App.GetAccountsByGroup(group);
                this.accounts = loadedAccounts; // 更新原始数据源 (如果你的逻辑依赖这个)
                this.filteredAccounts = [...loadedAccounts]; // 使用从后端加载的、已排序的数据填充

                // 在视图切换或账户修改后更新全局账户列表
                const allAccounts = await window.go.backend.App.GetAccounts();
                this.allAccounts = allAccounts;

                // 加载后（重新）初始化排序功能
                this.$nextTick(() => {
                    this.initSortable();
                });
            } catch (error) {
                console.error('加载分组账户失败:', error);
            }
        },

        async unlockVault() {
            try {
                await window.go.backend.App.UnlockVault(this.masterPassword);
                this.isUnlocked = true;
                this.errorMessage = '';
                this.masterPassword = '';
                await this.loadGroups();
                await this.loadAccounts();
                this.showNotification('密码库已解锁');
            } catch (error) {
                this.errorMessage = '密码错误，请重试';
                console.error('解锁错误:', error);
            }
        },

        async lockVault() {
            this.isUnlocked = false;
            this.selectedAccount = null;
            this.selectedAccountId = null;
            this.accounts = [];
            this.filteredAccounts = [];
            this.allAccounts = [];
            this.masterPassword = '';

            // 重新检查密码库是否存在
            try {
                this.isVaultExists = await window.go.backend.App.IsVaultExists();
            } catch (error) {
                console.error('检查密码库存在状态错误:', error);
            }

            this.showNotification('密码库已锁定');
        },

        async createVault() {
            if (this.masterPassword.length < 8) {
                this.errorMessage = '主密码至少需要8个字符';
                return;
            }

            if (this.masterPassword !== this.confirmMasterPassword) {
                this.errorMessage = '两次输入的密码不匹配';
                return;
            }

            try {
                await window.go.backend.App.CreateVault(this.masterPassword);
                this.isUnlocked = true;
                this.isVaultExists = true;
                this.errorMessage = '';
                this.masterPassword = '';
                this.confirmMasterPassword = '';
                this.accounts = [];
                this.filteredAccounts = [];
                this.showNotification('密码库已创建');
            } catch (error) {
                this.errorMessage = '创建密码库失败：' + error;
                console.error('创建密码库错误:', error);
            }
        },

        async loadAccounts() {
            try {
                // 销毁旧的sortable实例
                if (this.sortable) {
                    this.sortable.destroy();
                    this.sortable = null;
                }

                this.filteredAccounts = [];
                await this.$nextTick();

                // 从后端加载数据
                const loadedAccounts = await window.go.backend.App.GetAccounts();
                this.accounts = loadedAccounts;
                this.allAccounts = [...loadedAccounts];
                this.filteredAccounts = [...loadedAccounts];

                // 加载后（重新）初始化排序功能
                this.$nextTick(() => {
                    this.initSortable();
                });
            } catch (error) {
                console.error('加载账户错误:', error);
                this.showNotification('加载账户失败');
            }
        },

        async selectAccount(id) {
            this.selectedAccountId = id;
            this.showPassword = false;
            this.decryptedPassword = '';
            this.passwordStrengthInfo.show = false;
            this.isEditing = false;

            try {
                this.selectedAccount = await window.go.backend.App.GetAccountByID(id);
            } catch (error) {
                console.error('获取账户详情错误:', error);
                this.showNotification('加载账户详情失败');
                this.selectedAccount = null;
                this.selectedAccountId = null;
            }
        },

        async toggleShowPassword() {
            if (this.showPassword) {
                this.showPassword = false;
                return;
            }

            try {
                this.decryptedPassword = await window.go.backend.App.DecryptPassword(this.selectedAccountId);
                this.showPassword = true;

                // 显示密码强度
                this.passwordStrengthInfo.text = this.passwordStrength(this.decryptedPassword);
                this.passwordStrengthInfo.show = true;
            } catch (error) {
                console.error('显示密码错误:', error);
                this.showNotification('解密密码失败');
            }
        },

        async copyPassword() {
            try {
                const password = await window.go.backend.App.DecryptPassword(this.selectedAccountId);
                await window.go.backend.App.CopyToClipboard(password, 30);
                this.showNotification('密码已复制到剪贴板，将在30秒后清除');
            } catch (error) {
                console.error('复制密码错误:', error);
                this.showNotification('复制密码失败');
            }
        },

        newAccount() {
            this.isNewAccount = true;
            this.editingAccount = {
                id: '',
                platform: '',
                username: '',
                email: '',
                url: '',
                notes: '',
            };
            this.editingPassword = '';
            this.showEditPassword = false;
            this.showPasswordOptions = false;
            this.isEditing = true;
        },

        async editAccount() {
            if (!this.selectedAccount) return;

            this.isNewAccount = false;
            this.editingAccount = { ...this.selectedAccount };
            this.editingPassword = '';
            this.showEditPassword = false;
            this.showPasswordOptions = false;
            this.isEditing = true;
        },

        cancelEdit() {
            this.isEditing = false;
            this.editingAccount = {};
            this.editingPassword = '';
            this.showPasswordOptions = false;
        },

        async saveAccount() {
            if (!this.editingAccount.platform) {
                return;
            }

            try {
                const platformName = this.editingAccount.platform;
                const username = this.editingAccount.username;
                const isNew = this.isNewAccount;
                const originalId = this.editingAccount.id;

                if (isNew) {
                    await window.go.backend.App.AddAccount(this.editingAccount, this.editingPassword);
                    this.showNotification('账户已添加');
                } else {
                    const passwordToUpdate = this.editingPassword ? this.editingPassword : null;
                    await window.go.backend.App.UpdateAccount(this.editingAccount, passwordToUpdate);
                    this.showNotification('账户已更新');
                }

                // 关闭编辑界面，避免在加载期间出现问题
                this.isEditing = false;

                // 保存账户后刷新分组列表
                await this.loadGroups();

                // 重新加载所有账户
                // await this.loadAccounts();

                // 如果当前正在查看分组，刷新当前分组的账户
                if (!this.showAllAccounts && this.selectedGroup !== null) {
                    await this.loadAccountsByGroup(this.selectedGroup);
                } else {
                    await this.loadAccounts();
                }

                if (!isNew) {
                    // 编辑情况：选择原始ID的账户
                    this.selectAccount(originalId);
                } else {
                    // 根据平台名和用户名匹配新添加的账户
                    const matchedAccount = this.accounts.find(acc =>
                        acc.platform === platformName &&
                        acc.username === username
                    );

                    if (matchedAccount) {
                        this.selectAccount(matchedAccount.id);
                    }
                }
            } catch (error) {
                console.error('保存账户错误:', error);
                this.showNotification('保存账户失败');
            }
        },

        async confirmDeleteAccount() {
            if (!confirm('确定要删除这个账户吗？此操作不可恢复。')) {
                return;
            }

            try {
                await window.go.backend.App.DeleteAccount(this.selectedAccountId);
                this.showNotification('账户已删除');

                // 删除账户后重新加载分组
                await this.loadGroups();

                await this.loadAccounts();
                this.selectedAccount = null;
                this.selectedAccountId = null;
            } catch (error) {
                console.error('删除账户错误:', error);
                this.showNotification('删除账户失败');
            }
        },

        toggleShowEditPassword() {
            this.showEditPassword = !this.showEditPassword;
        },

        togglePasswordOptions() {
            this.showPasswordOptions = !this.showPasswordOptions;
        },

        async generatePasswordForEdit() {
            try {
                // 确保至少选择了一种字符集
                if (!this.generateOptions.UseLowercase &&
                    !this.generateOptions.UseUppercase &&
                    !this.generateOptions.UseDigits &&
                    !this.generateOptions.UseSymbols) {
                    this.showNotification('请至少选择一种字符集');
                    this.showPasswordOptions = true;
                    return;
                }

                this.editingPassword = await window.go.backend.App.GeneratePassword(this.generateOptions);
                this.showEditPassword = true;
            } catch (error) {
                console.error('生成密码错误:', error);
                this.showNotification('生成密码失败');
            }
        },

        passwordStrength(password) {
            if (!password) return '无';

            const length = password.length;
            const hasLower = /[a-z]/.test(password);
            const hasUpper = /[A-Z]/.test(password);
            const hasDigit = /\d/.test(password);
            const hasSymbol = /[^a-zA-Z0-9]/.test(password);

            const varietyCount = [hasLower, hasUpper, hasDigit, hasSymbol].filter(Boolean).length;

            if (length < 8) return '弱';
            if (length >= 12 && varietyCount >= 3) return '强';
            if (length >= 8 && varietyCount >= 2) return '中';
            return '弱';
        },

        async searchAccounts() {
            // 销毁现有排序实例
            if (this.sortable) {
                this.sortable.destroy();
                this.sortable = null;
            }

            if (!this.searchQuery.trim()) {
                // 清空搜索时，显示所有账户
                this.filteredAccounts = [...this.accounts];
            } else {
                try {
                    // 根据当前视图状态选择适当的搜索API
                    if (!this.showAllAccounts) {
                        // 分组内搜索（同时处理特定分组和未分组情况）
                        this.filteredAccounts = await window.go.backend.App.SearchAccountsInGroup(
                            this.searchQuery,
                            this.selectedGroup
                        );
                    } else {
                        // 全局搜索
                        this.filteredAccounts = await window.go.backend.App.SearchAccounts(this.searchQuery);
                    }
                } catch (error) {
                    console.error('搜索账户错误:', error);
                    this.showNotification('搜索账户失败');

                    // 降级到前端搜索
                    if (!this.showAllAccounts) {
                        // 如果在"未分组"模式下，只搜索未分组账户
                        if (this.selectedGroup === '') {
                            this.filteredAccounts = this.accounts.filter(account =>
                                (!account.group || account.group === '') &&
                                (account.platform.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                                    (account.username && account.username.toLowerCase().includes(this.searchQuery.toLowerCase())) ||
                                    (account.email && account.email.toLowerCase().includes(this.searchQuery.toLowerCase())) ||
                                    (account.url && account.url.toLowerCase().includes(this.searchQuery.toLowerCase())) ||
                                    (account.notes && account.notes.toLowerCase().includes(this.searchQuery.toLowerCase())))
                            );
                        } else {
                            // 其他特定分组搜索
                            this.filteredAccounts = this.accounts.filter(account =>
                                account.group === this.selectedGroup &&
                                (account.platform.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                                    (account.username && account.username.toLowerCase().includes(this.searchQuery.toLowerCase())) ||
                                    (account.email && account.email.toLowerCase().includes(this.searchQuery.toLowerCase())) ||
                                    (account.url && account.url.toLowerCase().includes(this.searchQuery.toLowerCase())) ||
                                    (account.notes && account.notes.toLowerCase().includes(this.searchQuery.toLowerCase())))
                            );
                        }
                    } else {
                        // 全局搜索 - 原有的前端搜索逻辑
                        this.filteredAccounts = this.accounts.filter(account =>
                            account.platform.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                            (account.username && account.username.toLowerCase().includes(this.searchQuery.toLowerCase())) ||
                            (account.email && account.email.toLowerCase().includes(this.searchQuery.toLowerCase())) ||
                            (account.url && account.url.toLowerCase().includes(this.searchQuery.toLowerCase())) ||
                            (account.notes && account.notes.toLowerCase().includes(this.searchQuery.toLowerCase()))
                        );
                    }
                }
            }

            // 在搜索结果更新后重新初始化排序功能
            this.$nextTick(() => {
                this.initSortable();
            });

            // 更新搜索模式的样式
            const el = document.getElementById('accounts-list');
            if (el) {
                if (this.searchQuery.trim()) {
                    el.classList.add('search-mode');
                } else {
                    el.classList.remove('search-mode');
                }
            }
        },

        // 更改主密码功能
        openChangeMasterPassword() {
            this.isChangingMasterPassword = true;
            this.currentMasterPassword = '';
            this.newMasterPassword = '';
            this.confirmNewMasterPassword = '';
            this.errorMessage = '';
        },

        cancelChangeMasterPassword() {
            this.isChangingMasterPassword = false;
        },

        async saveMasterPassword() {
            if (this.newMasterPassword.length < 8) {
                this.errorMessage = '主密码至少需要8个字符';
                return;
            }

            if (this.newMasterPassword !== this.confirmNewMasterPassword) {
                this.errorMessage = '两次输入的密码不匹配';
                return;
            }

            try {
                await window.go.backend.App.ChangeMasterPassword(
                    this.currentMasterPassword,
                    this.newMasterPassword
                );
                this.isChangingMasterPassword = false;
                this.showNotification('主密码已更改');
            } catch (error) {
                this.errorMessage = '更改主密码失败：' + error;
                console.error('更改主密码错误:', error);
            }
        },

        // 导入导出功能
        async exportVault() {
            if (!this.isUnlocked) {
                this.showNotification('请先解锁密码库');
                return;
            }

            try {
                await window.go.backend.App.ExportVault();
                this.showNotification('密码库已成功导出');
            } catch (error) {
                console.error('导出错误:', error);
                this.showNotification('导出失败: ' + error);
            }
        },

        async importVault() {
            if (!confirm('导入将覆盖当前密码库中的数据。请确认您已经备份了重要数据。是否继续？')) {
                return;
            }

            try {
                await window.go.backend.App.ImportVault();
                await this.loadGroups();
                await this.loadAccounts();
                this.showNotification('密码库已成功导入');
            } catch (error) {
                console.error('导入错误:', error);
                this.showNotification('导入失败: ' + error);
            }
        },

        // 导出为CSV功能
        async exportToCsv() {
            if (!this.isUnlocked) {
                this.showNotification('请先解锁密码库');
                return;
            }

            if (!confirm('警告：CSV文件将包含所有账户的明文密码，可能存在安全风险。确定要导出吗？')) {
                return;
            }

            try {
                await window.go.backend.App.ExportCsv();
                this.showNotification('密码库已成功导出为CSV格式');

                setTimeout(() => {
                    alert('安全提示：请确保妥善保管导出的CSV文件，完成使用后请删除。');
                }, 500);
            } catch (error) {
                console.error('导出CSV错误:', error);
                this.showNotification('导出CSV失败: ' + error);
            }
        },

        showNotification(message) {
            this.notification.message = message;
            this.notification.show = true;

            setTimeout(() => {
                this.notification.show = false;
            }, 3000);
        },

        openGitHubPage() {
            window.runtime.BrowserOpenURL("https://github.com/simp-lee/passwordmanager");
        }
    }));
});