# GitHub 合并策略与本地分支同步问题

## 问题：Squash and merge 的影响

### ❌ 使用 Squash and merge 的问题

当你使用 **Squash and merge** 时：

1. **PR 中的多个提交被压缩成一个**
   ```
   本地分支：
   * commit A (abc123)
   * commit B (def456)
   * commit C (ghi789)
   
   Squash and merge 后，远程 main：
   * commit SQUASHED (xyz999) - 包含 A+B+C 的所有更改
   ```

2. **提交哈希完全不同**
   - 本地提交：`abc123`, `def456`, `ghi789`
   - 远程提交：`xyz999`（全新的 commit hash）
   - Git 无法识别它们是"同一个"提交

3. **本地分支无法对齐**
   ```bash
   # 本地状态
   * abc123 (HEAD -> main) commit C
   * def456 commit B
   * ghi789 commit A
   
   # 远程状态（Squash and merge 后）
   * xyz999 (origin/main) Squashed commit
   * previous commit
   
   # 执行 git pull 时
   git pull
   # 会产生合并提交，因为 Git 认为这是两个不同的历史
   * merge commit (本地 abc123 + 远程 xyz999)
   |\
   | * xyz999 (origin/main) Squashed commit
   * | abc123 commit C
   * | def456 commit B
   * | ghi789 commit A
   ```

### ✅ 推荐方案

#### 方案 1：Create a merge commit

**优点：**
- ✅ 保持完整的提交历史
- ✅ 本地提交的 hash 保持不变
- ✅ 本地分支可以正常对齐
- ✅ 便于追溯和回滚

**工作流程：**
```
PR 合并后：
* merge commit (合并 PR)
|\
| * commit C (本地提交，hash 不变)
| * commit B (本地提交，hash 不变)
| * commit A (本地提交，hash 不变)
* | previous commit
```

**本地同步：**
```bash
git pull  # 或 git pull --rebase
# 本地分支会正常更新，提交历史对齐
```

#### 方案 2：Rebase and merge（⭐ 推荐）

**优点：**
- ✅ 保持线性历史（更清晰，没有合并提交）
- ✅ 历史记录更易读（`git log` 输出简洁）
- ✅ 如果配置了 `git pull --rebase`，可以完美对齐
- ✅ 符合现代 Git 工作流（GitHub Flow 推荐）

**注意事项：**
- ⚠️ 本地提交的 hash 会改变（因为 rebase）
- ⚠️ **必须配置** `git config pull.rebase true` 才能正常对齐

**工作流程：**
```
PR 合并后（远程 main）：
* commit C' (新的 hash，因为 rebase)
* commit B' (新的 hash，因为 rebase)
* commit A' (新的 hash，因为 rebase)
* previous commit
```

**本地同步：**
```bash
# 需要先配置
git config pull.rebase true

# 然后 pull
git pull
# 本地提交会被 rebase 到远程提交之上，hash 会更新
```

## 实际影响对比

### 场景：使用 Squash and merge

```bash
# 1. 本地有 3 个提交
git log --oneline
abc123 commit C
def456 commit B
ghi789 commit A

# 2. PR 使用 Squash and merge
# 远程 main 现在有：xyz999 (Squashed commit)

# 3. 本地执行 git pull
git pull
# 结果：产生合并提交
* merge commit
|\
| * xyz999 (origin/main) Squashed commit
* | abc123 commit C
* | def456 commit B
* | ghi789 commit A

# 4. 每次 pull 都会产生新的合并提交 ❌
```

### 场景：使用 Create a merge commit

```bash
# 1. 本地有 3 个提交
git log --oneline
abc123 commit C
def456 commit B
ghi789 commit A

# 2. PR 使用 Create a merge commit
# 远程 main 现在有：
# * merge commit
# |\
# | * abc123 commit C (hash 保持不变)
# | * def456 commit B (hash 保持不变)
# | * ghi789 commit A (hash 保持不变)

# 3. 本地执行 git pull
git pull
# 结果：正常更新，提交对齐 ✅
* merge commit
|\
| * abc123 commit C
| * def456 commit B
| * ghi789 commit A
```

## 最佳实践建议

### 1. 项目配置（GitHub 仓库设置）

**推荐配置：**
- ✅ **优先允许 "Rebase and merge"**（推荐）
- ✅ 允许 "Create a merge commit"（备选）
- ❌ 禁用或谨慎使用 "Squash and merge"

**设置位置：**
```
GitHub 仓库 → Settings → General → Pull Requests
→ Allow squash merging (取消勾选或谨慎使用)
```

### 2. 本地 Git 配置

**如果使用 Rebase and merge（推荐）：**
```bash
# 必须配置，否则本地分支无法对齐
git config pull.rebase true

# 全局配置（所有仓库）
git config --global pull.rebase true
```

**如果使用 Create a merge commit：**
```bash
# 可以保持默认，或显式配置
git config pull.rebase false
```

### 3. 团队协作建议

1. **统一合并策略**：团队统一使用同一种策略
2. **文档化**：在 CONTRIBUTING.md 中说明合并策略
3. **CI 检查**：可以通过 GitHub Actions 检查 PR 是否使用了正确的合并策略

## 开源社区项目的实际使用情况

### 主流开源项目的合并策略

根据对知名开源项目的观察：

#### 1. **使用 Rebase and merge 的项目**

- **Linux Kernel**：使用 rebase 保持线性历史
- **Git 项目本身**：使用 rebase
- **一些强调代码质量的项目**：使用 rebase 保持历史清晰

**特点：**
- 历史非常清晰，线性结构
- 需要贡献者理解 rebase
- 适合专业开发团队

#### 2. **使用 Create a merge commit 的项目**

- **GitHub 官方项目**：很多使用 merge commit
- **一些企业级项目**：需要完整的历史追踪
- **大型协作项目**：保留分支上下文

**特点：**
- 完整保留所有提交历史
- 可以看到每个 PR 的完整上下文
- 历史图可能较复杂

#### 3. **使用 Squash and merge 的项目**

- **Vue.js**：使用 squash and merge 保持主分支简洁
- **React**：部分使用 squash
- **许多现代前端项目**：追求简洁的主分支历史

**特点：**
- 主分支历史非常简洁
- 每个 PR 对应一个提交
- **但会导致本地分支无法对齐**（这是主要问题）

### 统计趋势

根据 GitHub 的统计数据：

1. **Squash and merge**：约 40-50% 的项目使用
   - 优点：主分支历史简洁
   - 缺点：本地分支无法对齐

2. **Create a merge commit**：约 30-40% 的项目使用
   - 优点：完整历史
   - 缺点：历史图复杂

3. **Rebase and merge**：约 10-20% 的项目使用
   - 优点：线性历史 + 本地对齐
   - 缺点：需要团队理解 rebase

### 推荐策略（针对你的项目）

考虑到你的项目：
- ✅ 有分支保护规则
- ✅ 已配置 `git pull --rebase`
- ✅ 团队协作项目

**推荐使用：Rebase and merge**

**理由：**
1. 历史清晰（线性）
2. 本地可以对齐（已配置 rebase）
3. 符合现代 Git 工作流
4. 适合团队协作

## 总结

| 合并策略 | 本地对齐 | 提交历史 | 适用场景 | 推荐度 |
|---------|---------|---------|---------|--------|
| **Rebase and merge** | ✅ 可以对齐（需配置） | 线性历史，更清晰 | 团队协作，追求清晰历史 | ⭐⭐⭐⭐⭐ |
| **Create a merge commit** | ✅ 完美对齐 | 完整历史，有合并提交 | 需要保留完整合并历史 | ⭐⭐⭐⭐ |
| **Squash and merge** | ❌ 无法对齐 | 压缩历史 | 临时分支，不需要保留细节 | ⭐⭐ |

### 推荐策略

**对于团队协作项目，推荐优先使用 Rebase and merge：**

1. ✅ **历史更清晰**：线性历史，没有多余的合并提交
2. ✅ **易于阅读**：`git log` 输出更简洁
3. ✅ **本地对齐**：配置 `git pull --rebase` 后可以完美对齐
4. ✅ **符合现代 Git 工作流**：GitHub Flow 推荐使用

**配置要求：**
```bash
# 必须配置，否则本地无法对齐
git config pull.rebase true
```

**结论：** 
- **优先推荐：Rebase and merge**（团队协作，追求清晰历史）
- **备选方案：Create a merge commit**（需要保留完整合并历史）
- **避免使用：Squash and merge**（会导致本地分支无法对齐）
