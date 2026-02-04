# Git 自动清理已删除远程分支配置

## 问题

当远程分支被删除后（例如 PR 合并后删除 feature 分支），本地仍然保留这些分支的引用，导致：
- 本地分支列表混乱
- `git branch -a` 显示很多已删除的远程分支
- 需要手动清理

## 解决方案

### 方案 1：全局配置（推荐）

配置 Git 在每次 `fetch` 时自动清理已删除的远程分支：

```bash
# 全局配置（所有仓库生效）
git config --global fetch.prune true
```

**优点：**
- ✅ 一次配置，所有仓库生效
- ✅ 每次 `git fetch` 或 `git pull` 时自动清理
- ✅ 无需手动操作

### 方案 2：仅当前仓库配置

如果只想对当前仓库生效：

```bash
# 仅当前仓库
git config fetch.prune true
```

### 方案 3：仅特定远程配置

如果只想对特定远程（如 origin）生效：

```bash
# 仅对 origin 远程生效
git config remote.origin.prune true
```

## 验证配置

检查配置是否生效：

```bash
# 检查全局配置
git config --global --get fetch.prune

# 检查当前仓库配置
git config --get fetch.prune

# 检查特定远程配置
git config --get remote.origin.prune
```

## 工作原理

配置后，当你执行以下命令时，Git 会自动清理已删除的远程分支：

```bash
git fetch
# 或
git pull
# 或
git fetch origin
```

**清理过程：**
1. Git 从远程获取最新的分支信息
2. 检测到远程分支已被删除
3. 自动删除本地的远程分支引用（如 `origin/feature-branch`）
4. 本地分支（如果存在）不会被删除，需要手动删除

## 手动清理（无需配置时）

如果不想配置自动清理，也可以手动执行：

```bash
# 清理已删除的远程分支引用
git fetch --prune

# 或者使用简写
git fetch -p
```

## 删除本地分支

自动清理只会删除**远程分支引用**（如 `origin/feature-branch`），不会删除**本地分支**。

如果需要删除本地分支：

```bash
# 删除单个本地分支
git branch -d feature-branch

# 强制删除（即使未合并）
git branch -D feature-branch

# 删除所有已合并到 main 的本地分支
git branch --merged main | grep -v "main" | xargs git branch -d

# 删除所有远程已删除的本地分支（需要先 fetch --prune）
git fetch --prune
git branch -vv | grep ': gone]' | awk '{print $1}' | xargs git branch -d
```

## 完整配置示例

结合之前配置的 `pull.rebase`，完整的 Git 配置应该是：

```bash
# 配置 pull 使用 rebase
git config --global pull.rebase true

# 配置自动清理已删除的远程分支
git config --global fetch.prune true
```

## 查看所有远程分支引用

```bash
# 查看所有远程分支引用
git branch -r

# 查看所有分支（本地 + 远程）
git branch -a

# 查看分支跟踪关系
git branch -vv
```

## 注意事项

1. **本地分支不会被自动删除**：自动清理只删除远程分支引用，本地分支需要手动删除
2. **安全性**：`fetch.prune` 是安全的，只会删除远程已不存在的分支引用
3. **性能**：自动清理不会影响性能，只是额外的检查步骤

## 推荐配置

对于团队协作项目，推荐同时配置：

```bash
# 配置 pull 使用 rebase（保持历史清晰）
git config --global pull.rebase true

# 配置自动清理已删除的远程分支（保持分支列表整洁）
git config --global fetch.prune true
```

这样配置后：
- ✅ `git pull` 使用 rebase，保持历史清晰
- ✅ `git fetch` 自动清理已删除的远程分支，保持分支列表整洁
- ✅ 一次配置，所有仓库生效
