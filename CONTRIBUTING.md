# Contributing to PieMDM

Thank you for your interest in contributing to PieMDM! This document provides guidelines and information for contributors.

## 🤝 How to Contribute

### Reporting Issues

1. **Search existing issues** first to avoid duplicates
2. **Use issue templates** when available
3. **Provide detailed information**:
   - Environment details (OS, Go version, Node.js version)
   - Steps to reproduce
   - Expected vs actual behavior
   - Screenshots or logs if applicable

### Suggesting Features

1. **Check existing feature requests** to avoid duplicates
2. **Describe the use case** and why it's valuable
3. **Provide implementation ideas** if you have them
4. **Consider backward compatibility**

### Code Contributions

1. **Fork the repository**
2. **Create a feature branch** from `main`
3. **Make your changes** following our coding standards
4. **Add tests** for new functionality
5. **Update documentation** as needed
6. **Submit a pull request**

## 🛠 Development Setup

### Prerequisites

- Go 1.24.12+
- Node.js 18+
- MySQL 8.0+
- Redis 6+
- Git

### Local Development

1. **Clone your fork**:

   ```bash
   git clone https://github.com/your-username/piemdm.git
   cd piemdm
   ```

2. **Set up backend**:

   ```bash
   cd backend
   make init
   cp config/local.yml.example config/local.yml
   # Edit config/local.yml with your database settings
   make run
   ```

3. **Set up frontend**:

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

4. **Set up database**:
   ```sql
   CREATE DATABASE piemdm CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

### Using Docker

```bash
# Copy environment file
cp env.example .env
# Edit .env with your settings

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f
```

## 📝 Coding Standards

### Go Code Standards

1. **Follow Go conventions**:

   - Use `gofmt` for formatting
   - Follow effective Go guidelines
   - Use meaningful variable and function names

2. **Code organization**:

   - Place handlers in `internal/handler/`
   - Business logic in `internal/service/`
   - Data access in `internal/repository/`
   - Models in `internal/model/`

3. **Error handling**:

   - Always handle errors appropriately
   - Use wrapped errors with context
   - Provide meaningful error messages

4. **Testing**:
   - Write unit tests for all new functions
   - Use table-driven tests when appropriate
   - Mock external dependencies

### Frontend Code Standards

1. **Vue.js conventions**:

   - Use Composition API
   - Follow Vue style guide
   - Use TypeScript when possible

2. **Component structure**:

   - Keep components focused and reusable
   - Use proper prop validation
   - Emit events for parent communication

3. **Styling**:
   - Use scoped styles
   - Follow BEM naming convention
   - Use Tailwind CSS utilities

### API Design

1. **RESTful principles**:

   - Use appropriate HTTP methods
   - Follow resource-based URLs
   - Return consistent response formats

2. **Documentation**:
   - Add Swagger annotations
   - Include request/response examples
   - Document error codes

## 🧪 Testing

### Backend Testing

```bash
cd backend

# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detection
make test-race

# Generate mocks
make mock
```

### Frontend Testing

```bash
cd frontend

# Run unit tests
npm run test

# Run e2e tests
npm run test:e2e

# Run tests in watch mode
npm run test:watch
```

### Integration Testing

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
make test-integration

# Cleanup
docker-compose -f docker-compose.test.yml down
```

## 📋 Pull Request Process

### Before Submitting

1. **Ensure tests pass**:

   ```bash
   make test
   npm run test
   ```

2. **Check code quality**:

   ```bash
   make lint
   make format
   ```

3. **Update documentation**:

   - Update README if needed
   - Add/update API documentation
   - Update CHANGELOG.md

4. **Commit message format**:

   ```
   type(scope): description

   Longer description if needed

   Fixes #issue-number
   ```

   Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### Pull Request Template

When creating a PR, please include:

- **Description**: What changes were made and why
- **Type of change**: Bug fix, feature, documentation, etc.
- **Testing**: How the changes were tested
- **Screenshots**: If UI changes were made
- **Checklist**: Confirm all requirements are met

### Review Process

1. **Automated checks** must pass (CI/CD, tests, linting)
2. **Code review** by at least one maintainer
3. **Manual testing** if needed
4. **Documentation review** for user-facing changes
5. **Merge** after approval

## 🏗 Project Structure

```
piemdm/
├── backend/               # Backend Go application
│   ├── cmd/               # Application entry points
│   ├── internal/          # Private application code
│   ├── pkg/               # Public packages
│   ├── config/            # Configuration files
│   ├── scripts/           # Build and deployment scripts
│   └── test/              # Test files
├── frontend/              # Frontend Vue.js application
│   ├── src/               # Source code
│   ├── public/            # Static assets
│   └── config/            # Configuration files
├── docs/                  # Documentation
├── scripts/               # Shared scripts
└── docker-compose.yml     # Docker services
```

## 🎯 Development Workflow

### Feature Development

1. **Create issue** describing the feature
2. **Create branch** from `main`:
   ```bash
   git checkout -b feature/issue-number-description
   ```
3. **Implement feature** with tests
4. **Update documentation**
5. **Submit pull request**

### Bug Fixes

1. **Create issue** describing the bug
2. **Create branch** from `main`:
   ```bash
   git checkout -b fix/issue-number-description
   ```
3. **Fix bug** with regression tests
4. **Submit pull request**

### Release Process

1. **Update version** in relevant files
2. **Update CHANGELOG.md**
3. **Create release branch**
4. **Tag release** after merge
5. **Deploy to production**

## 📚 Resources

### Documentation

- [Go Documentation](https://golang.org/doc/)
- [Vue.js Guide](https://vuejs.org/guide/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)

### Tools

- [Air](https://github.com/cosmtrek/air) - Live reload for Go
- [Swagger](https://swagger.io/) - API documentation
- [golangci-lint](https://golangci-lint.run/) - Go linter
- [ESLint](https://eslint.org/) - JavaScript linter

## 🆘 Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Email**: [jasen215@gmail.com] for private inquiries

## 📄 License

By contributing to PieMDM, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to PieMDM! 🎉
