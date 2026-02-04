# GitHub 仓库合并策略配置指南

## 如何修改 GitHub 仓库的合并策略

### 步骤 1：进入仓库设置

1. 打开你的 GitHub 仓库：`https://github.com/pieteams/piemdm`
2. 点击仓库页面右上角的 **Settings**（设置）标签

### 步骤 2：找到 Pull Requests 设置

在左侧导航栏中，找到并点击 **General**（常规设置）

向下滚动到 **Pull Requests**（拉取请求）部分

### 步骤 3：配置合并策略

在 **Pull Requests** 部分，你会看到三个选项：

```
☐ Allow merge commits
☐ Allow squash merging
☐ Allow rebase merging
```

**推荐配置：**

✅ **勾选 "Allow rebase merging"**（推荐）
- 这将启用 Rebase and merge 选项

✅ **勾选 "Allow merge commits"**（可选）
- 如果需要保留完整合并历史，可以同时启用

❌ **取消勾选 "Allow squash merging"**（推荐）
- 或者保持勾选但不在 PR 中使用

### 步骤 4：设置默认合并策略（可选）

GitHub 还允许设置**默认合并按钮**的行为：

1. 在同一个设置页面，找到 **"Default to squash merging"** 选项
2. **取消勾选**此选项（如果已勾选）
3. 这样默认按钮就不会是 "Squash and merge"

### 步骤 5：保存设置

点击页面底部的 **Save changes**（保存更改）按钮

## 配置后的效果

配置完成后，在 Pull Request 页面：

### 之前（只允许 Squash and merge）：
```
[Squash and merge]  ← 只有这个选项
```

### 之后（允许 Rebase and merge）：
```
[Rebase and merge]  ← 推荐使用这个
[Create a merge commit]  ← 可选
[Squash and merge]  ← 如果启用的话
```

## 注意事项

### 1. 权限要求

- 只有仓库的 **Owner**（所有者）或具有 **Admin**（管理员）权限的用户才能修改这些设置
- 如果你没有权限，需要联系仓库管理员

### 2. 现有 PR 不受影响

- 修改设置后，**现有的 PR 不会自动改变**
- 新的 PR 会显示新的合并选项
- 对于现有 PR，合并时可以选择新的策略

### 3. 团队通知

修改合并策略后，建议：
- 在团队中通知这个变更
- 更新项目文档（如 CONTRIBUTING.md）
- 确保团队成员了解新的合并策略

## 验证配置

配置完成后，可以：

1. 创建一个测试 PR
2. 查看合并按钮是否显示 "Rebase and merge"
3. 确认可以正常使用 Rebase and merge

## 完整配置示例

**推荐配置（针对你的项目）：**

```
✅ Allow merge commits
❌ Allow squash merging（取消勾选，或勾选但不使用）
✅ Allow rebase merging（推荐）
❌ Default to squash merging（取消勾选）
```

这样配置后：
- PR 页面会显示 "Rebase and merge" 按钮
- 团队成员可以使用 Rebase and merge
- 保持历史清晰，本地分支可以对齐

## 相关文档

- [GitHub 官方文档：配置合并方法](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/about-merge-methods-on-github)
- 项目内部文档：`docs/git-merge-strategy.md`
