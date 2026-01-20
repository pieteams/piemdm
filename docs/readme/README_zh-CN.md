# PieMDM - 企业级主数据管理系统

[![CI](https://github.com/pieworks/piemdm/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/pieworks/piemdm/actions/workflows/ci.yml)

[English](../../README.md)
| **简体中文**
| [繁體中文](README_zh-TW.md)
| [한국어](README_ko.md)
| [Русский](README_ru.md)
| [Tiếng Việt](README_vi.md)
| [日本語](README_ja.md)

PieMDM 是一款功能强大且易于使用的开源主数据管理 (MDM) 系统，专为企业数据治理而设计。基于 Go 后端和 Vue.js 前端构建，提供全面的数据管理、治理和集成能力。

**项目官网**: https://pieworks.github.io/piemdm/

## 🚀 功能特性

- 数据管理与集成
- 主数据建模
- 数据治理
- 系统集成
- 访问控制
- 工作流管理

## 📋 环境要求

- Go 1.24.12+
- Node.js 20+
- MySQL 8.0+
- Redis 6+

## 🚀 快速开始

### 1. 克隆代码仓库

```bash
git clone https://github.com/pieworks/piemdm.git
cd piemdm
```

### 2. 后端设置

```bash
cd backend
# 安装依赖
go mod tidy

# 复制配置文件
cp config/local.yml.example config/local.yml
# 编辑 config/local.yml 配置数据库和其他设置

# 运行数据库迁移
go run cmd/migration/main.go

# 启动后端服务
go run cmd/server/main.go
```

### 3. 前端设置

```bash
cd frontend
# 安装依赖
npm install
# 或者
pnpm install

# 启动开发服务器
npm run dev
# 或者
pnpm dev
```

### 4. 访问应用

- 前端地址: http://localhost:8081
- 后端 API: http://localhost:8787
- API 文档: http://localhost:8787/swagger/index.html

## 🐳 Docker 部署

### 使用 Docker Compose

```bash
# 启动所有服务
docker-compose -f deploy/docker-compose.yml up -d

# 查看日志
docker-compose -f deploy/docker-compose.yml logs -f

# 停止服务
docker-compose -f deploy/docker-compose.yml down
```

### 手动构建 Docker

```bash
# 构建后端镜像
cd backend
docker build -t piemdm-api:latest -f scripts/build/Dockerfile .

# 构建前端镜像
cd frontend
docker build -t piemdm-web:latest .
```

## 🔧 配置

### 环境变量

在根目录下创建一个 `.env` 文件：

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

### 数据库设置

```sql
CREATE DATABASE piemdm CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 📖 API 文档

API 文档使用 Swagger 自动生成，访问地址如下：

- 开发环境: http://localhost:8787/swagger/index.html
- 生产环境: https://your-domain.com/swagger/index.html

## 🧪 测试

### 后端测试

```bash
cd backend
# 运行所有测试
make test

# 运行带覆盖率的测试
go test -cover ./...

# 生成 Mock 文件
make mock
```

### 前端测试

```bash
cd frontend
# 运行单元测试
pnpm test

# 运行 E2E 测试
npm run test:e2e
```

## 🚀 部署

### 生产环境构建

```bash
# 构建后端
cd backend
make build

# 构建前端
cd frontend
pnpm build
```

### 环境特定配置

- 开发环境: `config/local.yml`
- 生产环境: `config/prod.yml`

## 🤝 贡献

我们欢迎提交贡献！详情请参阅我们的 [贡献指南](CONTRIBUTING.md)。

### 开发工作流

1. Fork 本仓库
2. 创建特性分支
3. 提交更改
4. 添加测试
5. 提交 Pull Request

### 代码规范

- 遵循 Go 最佳实践和规范
- 前端代码使用 ESLint 和 Prettier
- 编写全面的测试
- 更新文档

## 📄 许可证

本项目采用 MIT 许可证 - 详情请见 [LICENSE](LICENSE) 文件。

## 📞 支持

- 📧 邮箱: [jasen215@gmail.com]
- 🐛 Issues: [GitHub Issues](https://github.com/pieworks/piemdm/issues)
- 💬 讨论: [GitHub Discussions](https://github.com/pieworks/piemdm/discussions)

## 🙏 致谢

感谢所有贡献者和开源社区对本项目的支持。

---

**如果您觉得本项目有帮助，请给个 Star ⭐！**
