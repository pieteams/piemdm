# PieMDM - Enterprise Master Data Management System

[![CI](https://github.com/pieteams/piemdm/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/pieteams/piemdm/actions/workflows/ci.yml)

**English**
| [简体中文](docs/readme/README_zh-CN.md)
| [繁體中文](docs/readme/README_zh-TW.md)
| [한국어](docs/readme/README_ko.md)
| [Русский](docs/readme/README_ru.md)
| [Tiếng Việt](docs/readme/README_vi.md)
| [日本語](docs/readme/README_ja.md)

PieMDM is a powerful and user-friendly open-source Master Data Management (MDM) system designed for enterprise data governance. Built with Go backend and Vue.js frontend, it provides comprehensive data management, governance, and integration capabilities.

**Project Website**: https://pieteams.github.io/piemdm/

## 🚀 Features

- Data Management & Integration
- Master Data Modeling
- Data Governance
- System Integration
- Access Control
- Workflow Management

## 📋 Requirements

- Go 1.24.12+
- Node.js 20+
- MySQL 8.0+
- Redis 6+

## 🚀 Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/pieteams/piemdm.git
cd piemdm
git config pull.rebase true
```

### 2. Backend Setup

```bash
cd backend
# Install dependencies
go mod tidy

# Copy configuration file
cp config/local.yml.example config/local.yml
# Edit config/local.yml to configure database and other settings

# Run database migrations
go run cmd/migration/main.go

# Start backend service
go run cmd/server/main.go
```

### 3. Frontend Setup

```bash
cd frontend
# Install dependencies
npm install
# or
pnpm install

# Start development server
npm run dev
# or
pnpm dev
```

### 4. Access Application

- Frontend: http://localhost:8081
- Backend API: http://localhost:8787
- API Documentation: http://localhost:8787/swagger/index.html

## 🐳 Docker Deployment

### Using Docker Compose

```bash
# Start all services
docker-compose -f deploy/docker-compose.yml up -d

# View logs
docker-compose -f deploy/docker-compose.yml logs -f

# Stop services
docker-compose -f deploy/docker-compose.yml down
```

### Manual Docker Build

```bash
# Build backend image
docker build -f backend/Dockerfile -t piemdm-api:latest .

# Build frontend image
cd frontend
docker build -t piemdm-web:latest .
```

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the root directory:

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

### Database Setup

```sql
CREATE DATABASE piemdm CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 📖 API Documentation

API documentation is automatically generated using Swagger and can be accessed at:

- Development: http://localhost:8787/swagger/index.html
- Production: https://your-domain.com/swagger/index.html

## 🧪 Testing

### Backend Tests

```bash
cd backend
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Generate mocks
make mock
```

### Frontend Tests

```bash
cd frontend
# Run unit tests
pnpm test

# Run e2e tests
npm run test:e2e
```

## 🚀 Deployment

### Production Build

```bash
# Build backend
cd backend
make build

# Build frontend
cd frontend
pnpm build
```

### Environment-specific Configurations

- Development: `config/local.yml`
- Production: `config/prod.yml`

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### Code Standards

- Follow Go best practices and conventions
- Use ESLint and Prettier for frontend code
- Write comprehensive tests
- Update documentation

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Support

- 📧 Email: [jasen215@gmail.com]
- 🐛 Issues: [GitHub Issues](https://github.com/pieteams/piemdm/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/pieteams/piemdm/discussions)

## 🙏 Acknowledgments

Thanks to all contributors and the open source community for making this project possible.

---

**Star ⭐ this repository if you find it helpful!**
