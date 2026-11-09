# GitHub 仓库合并策略配置

本仓库只允许 **Squash and merge**。配置分两层：Settings 决定有哪些按钮，ruleset 决定能不能绕过。

## 1. Settings

仓库 → Settings → General → Pull Requests：

- ✅ 勾选 **Allow squash merging**
- ✅ Default commit message 选 **Pull request title and commit details**
- ❌ 取消 **Allow merge commits**
- ❌ 取消 **Allow rebase merging**
- ✅ 勾选 **Allow auto-merge**（可选）
- ✅ 勾选 **Automatically delete head branches**

保存后，PR 页面只剩一个合并按钮。

## 2. ruleset（真正的强制层）

只改 Settings 不够：有写权限的人仍可从命令行或 API 落 merge commit。强制在 Settings → Rules → 现有 ruleset `Protect main`（id `12391537`）里：

| 规则 | 取值 |
|---|---|
| Require a pull request before merging | `allowed_merge_methods: ["squash"]`、`required_approving_review_count: 0` |
| Require linear history | 开 |
| Require status checks | `ci-backend`、`ci-frontend`、`docs-build`，`strict: true` |
| Restrict deletions / Block force pushes | 开 |

必检名取自 workflow 里的 job `name`，改 job 名要同步改 ruleset，否则 PR 永远等一个不存在的检查。

## 3. 验证

1. 新 PR 的合并按钮只有一个 **Squash and merge**。
2. 合并后主干上是单父提交，标题形如 `ci(release): publish (#19)`。
3. `git log --merges origin/main` 输出为空。
4. 设成必检的检查在 PR 的 Checks 里出现过一次。

## 4. 常见错误

- 只改 Settings 不改 ruleset，管理员仍能点 merge commit。
- 保留 Allow rebase merging，再指望主干线性。
- 合并后先 `git pull` 再删本地分支，会把本地提交和 squash 提交再合并一次，主干外的分支多出一条 merge commit（见 `./git-merge-strategy.md`）。
