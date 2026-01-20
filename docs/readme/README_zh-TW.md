# PieMDM - 企業級主數據管理系統

[![CI](https://github.com/pieworks/piemdm/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/pieworks/piemdm/actions/workflows/ci.yml)

[English](../../README.md)
| [简体中文](README_zh-CN.md)
| **繁體中文**
| [한국어](README_ko.md)
| [Русский](README_ru.md)
| [Tiếng Việt](README_vi.md)
| [日本語](README_ja.md)

PieMDM 是一款功能強大且易於使用的開源主數據管理 (MDM) 系統，專為企業數據治理而設計。基於 Go 後端和 Vue.js 前端構建，提供全面的數據管理、治理和集成能力。

**專案官網**: https://pieworks.github.io/piemdm/

## 🚀 功能特性

- 數據管理與集成
- 主數據建模
- 數據治理
- 系統集成
- 訪問控制
- 工作流管理

## 📋 環境要求

- Go 1.24.12+
- Node.js 20+
- MySQL 8.0+
- Redis 6+

## 🚀 快速開始

### 1. 克隆代碼倉庫

```bash
git clone https://github.com/pieworks/piemdm.git
cd piemdm
```

### 2. 後端設置

```bash
cd backend
# 安裝依賴
go mod tidy

# 複製配置文件
cp config/local.yml.example config/local.yml
# 編輯 config/local.yml 配置數據庫和其他設置

# 運行數據庫遷移
go run cmd/migration/main.go

# 啟動後端服務
go run cmd/server/main.go
```

### 3. 前端設置

```bash
cd frontend
# 安裝依賴
npm install
# 或者
pnpm install

# 啟動開發服務器
npm run dev
# 或者
pnpm dev
```

### 4. 訪問應用

- 前端地址: http://localhost:8081
- 後端 API: http://localhost:8787
- API 文檔: http://localhost:8787/swagger/index.html

## 🐳 Docker 部署

### 使用 Docker Compose

```bash
# 啟動所有服務
docker-compose -f deploy/docker-compose.yml up -d

# 查看日誌
docker-compose -f deploy/docker-compose.yml logs -f

# 停止服務
docker-compose -f deploy/docker-compose.yml down
```

### 手動構建 Docker

```bash
# 構建後端鏡像
cd backend
docker build -t piemdm-api:latest -f scripts/build/Dockerfile .

# 構建前端鏡像
cd frontend
docker build -t piemdm-web:latest .
```

## 🔧 配置

### 環境變量

在根目錄下創建一個 `.env` 文件：

```env
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=piemdm

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_password

# JWT
JWT_SECRET=your_jwt_secret_key

# Application
APP_ENV=development
APP_PORT=8787
```

### 數據庫設置

```sql
CREATE DATABASE piemdm CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 📖 API 文檔

API 文檔使用 Swagger 自動生成，訪問地址如下：

- 開發環境: http://localhost:8787/swagger/index.html
- 生產環境: https://your-domain.com/swagger/index.html

## 🧪 測試

### 後端測試

```bash
cd backend
# 運行所有測試
make test

# 運行帶覆蓋率的測試
go test -cover ./...

# 生成 Mock 文件
make mock
```

### 前端測試

```bash
cd frontend
# 運行單元測試
pnpm test

# 運行 E2E 測試
npm run test:e2e
```

## 🚀 部署

### 生產環境構建

```bash
# 構建後端
cd backend
make build

# 構建前端
cd frontend
pnpm build
```

### 環境特定配置

- 開發環境: `config/local.yml`
- 生產環境: `config/prod.yml`

## 🤝 貢獻

我們歡迎提交貢獻！詳情請參閱我們的 [貢獻指南](CONTRIBUTING.md)。

### 開發工作流

1. Fork 本倉庫
2. 創建特性分支
3. 提交更改
4. 添加測試
5. 提交 Pull Request

### 代碼規範

- 遵循 Go 最佳實踐和規範
- 前端代碼使用 ESLint 和 Prettier
- 編寫全面的測試
- 更新文檔

## 📄 許可證

本項目採用 MIT 許可證 - 詳情請見 [LICENSE](LICENSE) 文件。

## 📞 支持

- 📧 郵箱: [jasen215@gmail.com]
- 🐛 Issues: [GitHub Issues](https://github.com/pieworks/piemdm/issues)
- 💬 討論: [GitHub Discussions](https://github.com/pieworks/piemdm/discussions)

## 🙏 致謝

感謝所有貢獻者和開源社區對本項目的支持。

---

**如果您覺得本項目有幫助，請給個 Star ⭐！**
