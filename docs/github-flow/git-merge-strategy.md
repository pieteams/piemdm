# GitHub 合并策略与本地分支同步

## 1. 本仓库的策略：只用 squash

ruleset `Protect main` 里 `allowed_merge_methods: ["squash"]` 且开启 `Require linear history`，所以主干上永远是「一个 PR 一条单父提交」。

- 提交标题：PR 标题，GitHub 自动补 `(#N)`
- 提交正文：PR 内每条提交的消息按顺序拼接
- 提交作者：PR 作者；committer 是 GitHub

## 2. 本地提交的 hash 会消失，这是定义不是故障

squash 之后主干上是一条全新的提交，本地分支的 `abc123`、`def456` 不会进入主干历史。原来的处理办法是换成 rebase 或 merge commit 来保住 hash，代价是主干失去「一个 PR 一条提交」和「整块 revert」。

正确做法是接受 hash 变化，合并后直接丢掉本地分支。本地分支不是资产，主干才是。

## 3. 标准流程

```bash
# 1. 从最新主干拉分支
git fetch origin
git checkout -b feat/xxx origin/main

# 2. 本地随意提交几次，消息写成能独立看懂的一句话
git commit -m "add xxx"

# 3. 开 PR，标题按 <type>(<scope>)!: <subject> 写
gh pr create --fill

# 4. 合并（仓库只有这一个入口）
gh pr merge --squash

# 5. 丢掉本地分支，重新对齐主干
git checkout main
git fetch --prune
git reset --hard origin/main
git branch -D feat/xxx
```

第 5 步是关键的顺序：**先删分支，不要先 `git pull`**。在带着旧提交的分支上 `git pull`，Git 会认为 squash 提交和本地提交是两条不同历史，于是再造一个 merge commit，还把同一份改动带回来一次。

分支已开启 `Automatically delete head branches`，远端不用手动删。

## 4. 合并后本地出现重复改动的补救

如果已经在旧分支上 pull 出 merge commit，不要推到主干（ruleset 也会拦）。直接放弃本地分支：

```bash
git checkout main
git fetch origin
git reset --hard origin/main
git branch -D feat/xxx
```

改动本身已经在主干上，工作区不需要保留。

## 5. 三种合并方式的差别

| 方式 | 主干形态 | 本地怎么对齐 | 何时用 |
|---|---|---|---|
| **Squash and merge** | 一个 PR 一条提交 | 合并后删本地分支 | 本仓库唯一方式 |
| Rebase and merge | 逐条提交线性落库 | `git pull --rebase` 可对齐 | 每条提交都要独立评审、独立回滚 |
| Create a merge commit | 双父提交，保留分支 | `git pull` 直接对齐 | 上游是发布分支流，或仓库是制品镜像 |

后两种在本仓库没有入口，也不用为此配置 `git config pull.rebase`。`pull.rebase` 只影响你自己拉代码的方式，和 PR 怎么落库无关。

## 6. 什么时候值得改成逐条落库

同时满足才考虑：每条提交都能单独评审和单独回滚、需要保留作者签名链或 `Signed-off-by`、上游是补丁流（按 patch 应用而非按 PR 合并）。一次性修 bug 或加小功能的 PR，三条通常都不成立。
