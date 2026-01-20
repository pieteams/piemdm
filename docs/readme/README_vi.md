# PieMDM - Hệ thống Quản lý Dữ liệu Chủ Doanh nghiệp

[![CI](https://github.com/pieworks/piemdm/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/pieworks/piemdm/actions/workflows/ci.yml)

[English](../../README.md)
| [简体中文](README_zh-CN.md)
| [繁體中文](README_zh-TW.md)
| [한국어](README_ko.md)
| [Русский](README_ru.md)
| **Tiếng Việt**
| [日本語](README_ja.md)

PieMDM là một hệ thống Quản lý Dữ liệu Chủ (MDM) mã nguồn mở mạnh mẽ và thân thiện với người dùng, được thiết kế cho quản trị dữ liệu doanh nghiệp. Được xây dựng với backend Go và frontend Vue.js, nó cung cấp khả năng quản lý, quản trị và tích hợp dữ liệu toàn diện.

**Trang web dự án**: https://pieworks.github.io/piemdm/

## 🚀 Tính năng

- Quản lý & Tích hợp Dữ liệu
- Mô hình hóa Dữ liệu Chủ
- Quản trị Dữ liệu
- Tích hợp Hệ thống
- Kiểm soát Truy cập
- Quản lý Quy trình

## 📋 Yêu cầu

- Go 1.24.12+
- Node.js 20+
- MySQL 8.0+
- Redis 6+

## 🚀 Bắt đầu nhanh

### 1. Clone Kho lưu trữ

```bash
git clone https://github.com/pieworks/piemdm.git
cd piemdm
```

### 2. Thiết lập Backend

```bash
cd backend
# Cài đặt các phụ thuộc
go mod tidy

# Sao chép tệp cấu hình
cp config/local.yml.example config/local.yml
# Chỉnh sửa config/local.yml để cấu hình cơ sở dữ liệu và các cài đặt khác

# Chạy di chuyển cơ sở dữ liệu (migration)
go run cmd/migration/main.go

# Khởi động dịch vụ backend
go run cmd/server/main.go
```

### 3. Thiết lập Frontend

```bash
cd frontend
# Cài đặt các phụ thuộc
npm install
# hoặc
pnpm install

# Khởi động máy chủ phát triển
npm run dev
# hoặc
pnpm dev
```

### 4. Truy cập Ứng dụng

- Frontend: http://localhost:8081
- Backend API: http://localhost:8787
- Tài liệu API: http://localhost:8787/swagger/index.html

## 🐳 Triển khai Docker

### Sử dụng Docker Compose

```bash
# Khởi động tất cả dịch vụ
docker-compose -f deploy/docker-compose.yml up -d

# Xem log
docker-compose -f deploy/docker-compose.yml logs -f

# Dừng dịch vụ
docker-compose -f deploy/docker-compose.yml down
```

### Xây dựng Docker Thủ công

```bash
# Xây dựng image backend
cd backend
docker build -t piemdm-api:latest -f scripts/build/Dockerfile .

# Xây dựng image frontend
cd frontend
docker build -t piemdm-web:latest .
```

## 🔧 Cấu hình

### Biến môi trường

Tạo tệp `.env` trong thư mục gốc:

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

### Thiết lập Cơ sở dữ liệu

```sql
CREATE DATABASE piemdm CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 📖 Tài liệu API

Tài liệu API được tạo tự động bằng Swagger và có thể truy cập tại:

- Phát triển: http://localhost:8787/swagger/index.html
- Sản xuất: https://your-domain.com/swagger/index.html

## 🧪 Kiểm thử

### Kiểm thử Backend

```bash
cd backend
# Chạy tất cả các test
make test

# Chạy test với coverage
go test -cover ./...

# Tạo mock
make mock
```

### Kiểm thử Frontend

```bash
cd frontend
# Chạy unit test
pnpm test

# Chạy e2e test
npm run test:e2e
```

## 🚀 Triển khai

### Xây dựng bản Sản xuất

```bash
# Xây dựng backend
cd backend
make build

# Xây dựng frontend
cd frontend
pnpm build
```

### Cấu hình theo Môi trường

- Phát triển: `config/local.yml`
- Sản xuất: `config/prod.yml`

## 🤝 Đóng góp

Chúng tôi hoan nghênh mọi đóng góp! Vui lòng xem [Hướng dẫn Đóng góp](CONTRIBUTING.md) để biết chi tiết.

### Quy trình Phát triển

1. Fork kho lưu trữ
2. Tạo nhánh tính năng (feature branch)
3. Thực hiện thay đổi
4. Thêm test
5. Gửi Pull Request

### Tiêu chuẩn Mã

- Tuân thủ các quy tắc và thực tiễn tốt nhất của Go
- Sử dụng ESLint và Prettier cho mã frontend
- Viết test đầy đủ
- Cập nhật tài liệu

## 📄 Giấy phép

Dự án này được cấp phép theo Giấy phép MIT - xem tệp [LICENSE](LICENSE) để biết chi tiết.

## 📞 Hỗ trợ

- 📧 Email: [jasen215@gmail.com]
- 🐛 Issues: [GitHub Issues](https://github.com/pieworks/piemdm/issues)
- 💬 Thảo luận: [GitHub Discussions](https://github.com/pieworks/piemdm/discussions)

## 🙏 Lời cảm ơn

Cảm ơn tất cả những người đóng góp và cộng đồng nguồn mở đã làm cho dự án này trở nên khả thi.

---

**Hãy tặng Star ⭐ cho kho lưu trữ này nếu bạn thấy nó hữu ích!**
