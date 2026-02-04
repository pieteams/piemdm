# GitHub Flow 进阶指南 (Enterprise Pro)

本文档是 [GitHub Flow 标准规范](github-flow.md) 的**企业级增强版**。它结合了 **Monorepo** 架构特点和 **企业级交付标准**，在坚持 GitHub Flow 简单性的同时，强化了环境管理、质量门禁和发布稳定性。

---

## 🏗 一、 分支策略全景图

我们坚持 **GitHub Flow** 的核心原则：**以 `main` 为中心，移除 `develop` 分支**。

```mermaid
graph LR
    subgraph Upstream[主仓库 (pieteams/piemdm)]
    M[main] -->|Tag: v1.0.0| Rel[Release]
    M -->|CI Check| M
    end

    subgraph Contribution[贡献流程]
    F[feat/user-api] -->|Pull Request| M
    B[fix/login-bug] -->|Pull Request| M
    end

    subgraph Environment[环境交付]
    M -.->|Auto Deploy| Staging[预发布环境]
    Rel -.->|Manual Approval| Prod[生产环境]
    end
```

### ✅ 分支清单

| 分支类型 | 命名规范 | 生命周期 | 说明 |
| :--- | :--- | :--- | :--- |
| **Main** | `main` | **永久** | 唯一可发布代码源，对应**预发布环境 (Staging)**。 |
| **Feature** | `feat/<issue-id>-<desc>` | 临时 (合并即删) | 功能开发。如 `feat/102-batch-import`。 |
| **Fix** | `fix/<issue-id>-<desc>` | 临时 (合并即删) | Bug 修复。如 `fix/105-npe-error`。 |
| **Release** | `release/vX.Y.Z` | 临时 | *(可选)* 仅用于大型版本封板测试，完成后合并回 main 打 Tag。 |

---

## 🌍 二、 环境管理与交付流程

虽然分支很简单，但企业级流程需要严格的环境流转。

| 环境 | 对应状态 | 部署内容 | 触发机制 |
| :--- | :--- | :--- | :--- |
| **Dev** | Feature Branch | 当前分支代码 | 开发者本地启动 / CI 临时构建 |
| **Staging** | Main (Head) | 最新 `main` 代码 | PR 合并进 `main` 后自动部署 |
| **Prod** | Tag (vX.Y.Z) | `v*` 标签版本 | 推送 Tag -> CI 构建 -> **人工审批** -> 部署 |

> **关键差异**：GitHub Flow 不使用 `dev` 分支，而是用 `main` 充当集成测试源（Staging Codebase）。

---

## 📦 三、 Monorepo 组件发布策略

### 1. 后端 (Go)
*   **开发**：`go.work` 本地引用，全源码编译。
*   **发布**：
    1.  **Shared Pkg**：先给 `packages/go/openapi` 打 Tag (v1.0.0)。
    2.  **App**：`backend` 的 `go.mod` 必须引用远程 Tag 版本，**禁止依赖 `go.work`**。

### 2. 前端 (Vue)
*   **开发**：`pnpm workspace` 软链引用。
*   **发布**：
    *   前端库 (`packages/web/api-client`) 不需要发布 npm 包。
    *   CI 构建时，直接将 monorepo 源码打包成静态资源 (dist)。
    *   **产物**：Docker 镜像 (包含 Nginx + dist) 或上传 CDN。

---

## 🛡 四、 CI/CD 质量关卡 (Quality Gates)

PR 合并到 `main` 必须通过以下自动化检查：

1.  **Lint & Style**：
    *   Go: `golangci-lint`
    *   Vue: `eslint` + `prettier`
2.  **Unit Test**：
    *   Backend: `go test -race ./...` (覆盖率 > 60%)
    *   Frontend: `vitest` 组件测试
3.  **Build Check**：
    *   模拟生产环境构建 (验证无 `replace` / `go.work` 依赖)

---

## 📝 五、 FAQ

### Q1: 为什么不用 Git Flow (develop 分支)?
**A**: 在配合完善的 CI/CD 下，`main` 就是最稳定的集成点。引入 `develop` 会导致 "代码合并地狱" 和 "部署等待"，降低迭代效率。

### Q2: 生产环境发现 Bug 怎么办？
**A**: 标准 Hotfix 流程：
1. 从 `main` (或对应的 Tag) 切出 `fix/xxx`。
2. 修复并合并回 `main`。
3. 从 `main` 打出新的 Patch Tag (v1.0.1) 进行发布。

### Q3: 前端包为什么不发 npm?
**A**: 在 Monorepo 中，源码级依赖 (Workspace) 比 npm 依赖更高效，避免了 "修改库 -> 发布 -> 更新依赖" 的繁琐循环。仅当该库需要被**仓库之外**的项目使用时，才考虑发布 npm。
