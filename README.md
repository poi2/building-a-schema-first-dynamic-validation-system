# Building a Schema-First Dynamic Validation System

A proof-of-concept for schema-first validation using ConnectRPC and protovalidate, demonstrating **protovalidate-based validation** and **Context Enrichment** across FE and BE.

## Overview

This PoC focuses on validating the core concepts of schema-driven validation:

* `.proto` files serve as the **single source of truth** for validation rules
* **Context Enrichment**: Business logic (user plan-based restrictions) using永続層の値
* **protovalidate**: Declarative validation rules using CEL expressions
* **FE/BE validation**: Both layers use protovalidate for consistent validation

### What this PoC demonstrates

✅ **protovalidate** による宣言的バリデーション
✅ **Context Enrichment** (YAML から user.plan を取得して動的バリデーション)
✅ FE/BE 両方での protovalidate 動作確認
✅ curl/grpcurl による API 検証

### What is out of scope

❌ Hot Reload (技術的制約により実現不可 - [詳細](docs/001-DD.003.validation-strategy.md#71-実装結果と重要な制約事項))
❌ 動的スキーマ配信・同期機構
❌ FE-BE 間の実際の通信 (各層で独立してバリデーション検証)
❌ Server Streaming
❌ E2Eテスト (Playwright)

## Architecture (Simplified)

* **Backend**: Go, `connect-go`, `protovalidate-go`, YAML 永続化
* **Frontend**: TypeScript, `protovalidate-ts`, YAML 永続化（疑似的な実装）
* **ISR**: Go service (参考実装として維持、BE との連携なし)
* **Persistence**: YAML ファイル (PostgreSQL は使用しない)
* **Services**: 個別起動 (docker-compose は使用しない)

## Getting Started

### Prerequisites

* Node.js 20+ (FE 開発用)
* Go 1.24+ (BE 開発用)
* [Buf CLI](https://docs.buf.build/installation) (proto コード生成用)

**注**: docker-compose は不要です（サービス間の依存関係がないため）

### Setup

**注**: 現状の BE 実装は PostgreSQL ベースのため、YAML 永続化への移行（Task 3-3）完了後に以下の手順が有効になります。現時点では `docker-compose up -d` または `CELO_DB_URL` 環境変数の設定が必要です。

```bash
# 1. Generate code from proto files
make proto-generate

# 2. Start BE service (YAML 永続化移行後)
cd services/be
go run main.go
# -> http://localhost:50052 で起動

# 3. (Optional) Start FE service (実装後)
# 別ターミナルで実行:
cd services/fe
npm run dev
# -> http://localhost:3000 で起動

# 4. (Optional) Start ISR service (参考)
# 別ターミナルで実行:
cd services/isr
go run main.go
# -> http://localhost:50051 で起動
```

### YAML Persistence

データは YAML ファイルに保存されます:

```bash
services/be/data/
├── user.yaml    # ユーザー情報
└── post.yaml    # 投稿情報

services/fe/data/  # (実装後)
├── user.yaml
└── post.yaml
```

### API Testing

#### curl (HTTP/JSON)

```bash
# Create User
curl -X POST http://localhost:50052/user.v1.UserService/CreateUser \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "plan": "USER_PLAN_PRO"
  }'

# List Users
curl -X POST http://localhost:50052/user.v1.UserService/ListUsers \
  -H "Content-Type: application/json" \
  -d '{"page": 1, "pageSize": 10}'
```

#### grpcurl (gRPC)

```bash
# Create User
grpcurl -plaintext \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "plan": "USER_PLAN_PRO"
  }' \
  localhost:50052 \
  user.v1.UserService/CreateUser

# List Users
grpcurl -plaintext \
  -d '{"page": 1, "pageSize": 10}' \
  localhost:50052 \
  user.v1.UserService/ListUsers
```

## Documentation

### For Developers

* **[Development Guide](DEVELOPMENT.md)** - Setup, Git hooks, coding conventions

### Design Documentation

* [Requirements](docs/000.requirement.md) - PoC の目的とスコープ（更新済み）
* [High-Level Design](docs/001-DD.001.high-level-design.md) - 簡略化されたアーキテクチャ（更新済み）
* [Validation Strategy](docs/001-DD.003.validation-strategy.md) - Context Enrichment パターン
* [Data Model (YAML)](docs/001-DD.004.data-model-and-api-interface.md) - YAML フォーマットと API（更新済み）
* [Tasks](docs/003-TASK.001.tasks.md) - タスク一覧（更新済み）

## Key Features

### 1. Schema-First Validation

`.proto` ファイルに定義された validation ルールが FE/BE で自動的に適用されます:

```protobuf
message CreateUserRequest {
  string name = 1 [(buf.validate.field).string = {
    min_len: 1,
    max_len: 100
  }];
  string email = 2 [(buf.validate.field).string.email = true];
  common.v1.UserPlan plan = 3;
}
```

### 2. Context Enrichment

BE では YAML から user.plan を取得し、plan に応じた動的バリデーションを実現:

* **free**: message は 100文字以内
* **pro**: message は 200文字以内
* **enterprise**: message は 300文字以内

```go
// user.yaml から User を取得
user, err := h.userRepo.GetByID(ctx, req.Msg.UserId)

// user.plan に基づいて message の長さを検証
maxLen := getMaxLengthForPlan(user.Plan)
if len(req.Msg.Message) > maxLen {
    return connect.NewError(connect.CodeInvalidArgument, ...)
}
```

### 3. Dual-Layer Validation

* **FE**: protovalidate-ts で即時バリデーション（ユーザー体験向上）
* **BE**: protovalidate-go で最終バリデーション（セキュリティ保証）

## Development

### Code Generation

```bash
# Generate code from proto files
make proto-generate

# Lint proto files
make proto-lint
```

### Testing

```bash
# Run all tests
make test

# Format code
make fmt

# Lint code
make lint

# Run all CI checks
make ci
```

### Linting

```bash
# Lint markdown files
npm run lint:md

# Auto-fix markdown issues
npm run lint:md:fix
```

### Project Structure

```text
building-a-schema-first-dynamic-validation-system/
├── go.work              # Go Workspaces configuration
├── package.json         # Node.js Workspaces configuration
├── buf.yaml             # Buf configuration
├── buf.gen.yaml         # Code generation settings
├── Makefile             # Common development tasks
├── proto/               # Proto definitions (single source of truth)
│   ├── common/v1/
│   ├── user/v1/
│   ├── post/v1/
│   └── isr/v1/
├── pkg/
│   └── gen/             # Generated code (shared module)
│       ├── go/
│       └── ts/
├── services/
│   ├── be/              # Backend service
│   │   ├── main.go
│   │   ├── internal/
│   │   └── data/        # YAML files
│   ├── fe/              # Frontend (実装中)
│   │   └── data/        # YAML files
│   └── isr/             # ISR (参考実装)
└── docs/                # Design documentation
```

**注**: `docker-compose.yml` と `docker/init-db/` は今後削除予定です（PostgreSQL 不要のため）

## PoC Verification Scenarios

### Scenario 1: Basic Validation

1. User を作成
2. protovalidate が name, email をバリデーション
3. user.yaml に保存

### Scenario 2: Context Enrichment

1. free ユーザーで 100文字以内の Post を作成 → 成功
2. free ユーザーで 101文字の Post を作成 → エラー
3. pro ユーザーで 200文字以内の Post を作成 → 成功

### Scenario 3: Security (plan 偽装の防御)

1. リクエストで plan を偽装
2. BE が user.yaml から正しい plan を取得
3. 正しい plan でバリデーション → 偽装を検知してエラー

## License

This is a proof-of-concept project for educational purposes.
