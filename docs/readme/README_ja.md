# PieMDM - エンタープライズマスターデータ管理システム

[![CI](https://github.com/pieworks/piemdm/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/pieworks/piemdm/actions/workflows/ci.yml)

[English](../../README.md)
| [简体中文](README_zh-CN.md)
| [繁體中文](README_zh-TW.md)
| [한국어](README_ko.md)
| [Русский](README_ru.md)
| [Tiếng Việt](README_vi.md)
| **日本語**

PieMDMは、企業のデータガバナンスのために設計された、強力で使いやすいオープンソースのマスターデータ管理（MDM）システムです。GoバックエンドとVue.jsフロントエンドで構築されており、包括的なデータ管理、ガバナンス、および統合機能を提供します。

**プロジェクトウェブサイト**: https://pieworks.github.io/piemdm/

## 🚀 機能

- データ管理と統合
- マスターデータモデリング
- データガバナンス
- システム統合
- アクセス制御
- ワークフロー管理

## 📋 必要条件

- Go 1.24.12以上
- Node.js 20以上
- MySQL 8.0以上
- Redis 6以上

## 🚀 クイックスタート

### 1. リポジトリのクローン

```bash
git clone https://github.com/pieworks/piemdm.git
cd piemdm
```

### 2. バックエンドのセットアップ

```bash
cd backend
# 依存関係のインストール
go mod tidy

# 設定ファイルのコピー
cp config/local.yml.example config/local.yml
# config/local.ymlを編集してデータベースなどを設定

# データベースマイグレーションの実行
go run cmd/migration/main.go

# バックエンドサービスの起動
go run cmd/server/main.go
```

### 3. フロントエンドのセットアップ

```bash
cd frontend
# 依存関係のインストール
npm install
# または
pnpm install

# 開発サーバーの起動
npm run dev
# または
pnpm dev
```

### 4. アプリケーションへのアクセス

- フロントエンド: http://localhost:8081
- バックエンドAPI: http://localhost:8787
- APIドキュメント: http://localhost:8787/swagger/index.html

## 🐳 Dockerデプロイ

### Docker Composeの使用

```bash
# 全サービスの起動
docker-compose -f deploy/docker-compose.yml up -d

# ログの表示
docker-compose -f deploy/docker-compose.yml logs -f

# サービスの停止
docker-compose -f deploy/docker-compose.yml down
```

### Dockerの手動ビルド

```bash
# バックエンドイメージのビルド
cd backend
docker build -t piemdm-api:latest -f scripts/build/Dockerfile .

# フロントエンドイメージのビルド
cd frontend
docker build -t piemdm-web:latest .
```

## 🔧 設定

### 環境変数

ルートディレクトリに `.env` ファイルを作成します：

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

### データベース設定

```sql
CREATE DATABASE piemdm CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 📖 APIドキュメント

APIドキュメントはSwaggerを使用して自動生成され、以下からアクセスできます：

- 開発環境: http://localhost:8787/swagger/index.html
- 本番環境: https://your-domain.com/swagger/index.html

## 🧪 テスト

### バックエンドテスト

```bash
cd backend
# 全テストの実行
make test

# カバレッジ付きテストの実行
go test -cover ./...

# モックの生成
make mock
```

### フロントエンドテスト

```bash
cd frontend
# ユニットテストの実行
pnpm test

# E2Eテストの実行
npm run test:e2e
```

## 🚀 デプロイ

### 本番ビルド

```bash
# バックエンドのビルド
cd backend
make build

# フロントエンドのビルド
cd frontend
pnpm build
```

### 環境固有の設定

- 開発環境: `config/local.yml`
- 本番環境: `config/prod.yml`

## 🤝 貢献

貢献を歓迎します！詳細は [貢献ガイド](CONTRIBUTING.md) をご覧ください。

### 開発ワークフロー

1. リポジトリをフォーク
2. 機能ブランチを作成
3. 変更をコミット
4. テストを追加
5. プルリクエストを送信

### コーディング規約

- Goのベストプラクティスと規約に従う
- フロントエンドコードにはESLintとPrettierを使用
- 包括的なテストを作成
- ドキュメントを更新

## 📄 ライセンス

このプロジェクトはMITライセンスの下でライセンスされています - 詳細は [LICENSE](LICENSE) ファイルをご覧ください。

## 📞 サポート

- 📧 メール: [jasen215@gmail.com]
- 🐛 Issues: [GitHub Issues](https://github.com/pieworks/piemdm/issues)
- 💬 ディスカッション: [GitHub Discussions](https://github.com/pieworks/piemdm/discussions)

## 🙏 謝辞

このプロジェクトを可能にしてくれたすべての貢献者とオープンソースコミュニティに感謝します。

---

**このプロジェクトが役立つと思ったら、Star ⭐ をお願いします！**
