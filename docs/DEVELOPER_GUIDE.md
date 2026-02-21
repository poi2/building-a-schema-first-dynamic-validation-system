# Developer Guide

このドキュメントは、プロジェクトの開発に参加する開発者向けのガイドです。

## 目次

- [開発環境のセットアップ](#開発環境のセットアップ)
- [Git Hooks](#git-hooks)
- [コミット規約](#コミット規約)
- [開発ワークフロー](#開発ワークフロー)
- [コーディング規約](#コーディング規約)
- [テスト](#テスト)
- [PR作成とレビュー](#pr作成とレビュー)

## 開発環境のセットアップ

### 必須ツール

以下のツールをインストールしてください：

- **Docker & Docker Compose**: コンテナ環境
- **Node.js 20+**: BFF/FE開発用
- **Go 1.23+**: Backend/ISR開発用
- **Buf CLI**: Proto code generation
  - インストール: <https://docs.buf.build/installation>

### 初回セットアップ

```bash
# 1. リポジトリをクローン
git clone https://github.com/poi2/building-a-schema-first-dynamic-validation-system.git
cd building-a-schema-first-dynamic-validation-system

# 2. Git hooksをセットアップ（重要！）
bash .github/git-hooks/setup-hooks.sh

# 3. 依存関係のインストール
npm install

# 4. Protoコード生成
make proto-generate

# 5. Dockerサービスの起動
docker compose up -d

# 6. サービスの動作確認
docker compose ps
```

### ディレクトリ構成

```text
.
├── .github/
│   ├── git-hooks/           # Git hooks (commit-msg, setup script)
│   └── workflows/           # GitHub Actions CI/CD
├── docs/                    # 設計ドキュメント
├── proto/                   # Proto定義（Single Source of Truth）
├── pkg/gen/                 # 生成コード（共有モジュール）
│   ├── go/                  # Go generated code
│   └── ts/                  # TypeScript generated code
├── services/
│   ├── isr/                 # Internal Schema Registry (Go)
│   ├── be/                  # Backend service (Go)
│   ├── bff/                 # Backend for Frontend (Node.js)
│   └── fe/                  # Frontend (React)
└── tests/e2e/               # End-to-end tests
```

## Git Hooks

### セットアップ

**最初に必ず実行してください：**

```bash
bash .github/git-hooks/setup-hooks.sh
```

このスクリプトは `.github/git-hooks/commit-msg` を `.git/hooks/` にコピーして実行権限を付与します。

### commit-msg フック

コミット時に以下をチェックします：

1. **1行ルール**: コミットメッセージは1行で記述
2. **Conventional Commits**: 規定のフォーマットに従う

違反した場合、コミットは拒否されます。

## コミット規約

### Conventional Commits

このプロジェクトでは [Conventional Commits](https://www.conventionalcommits.org/) を採用しています。

#### フォーマット

```text
<type>(<scope>): <description>
```

- **type**: 必須
- **scope**: オプション
- **description**: 必須（短く簡潔に）

#### Type一覧

| Type | 説明 | 例 |
|------|------|-----|
| `feat` | 新機能追加 | `feat: Add protovalidate interceptor` |
| `fix` | バグ修正 | `fix: Update Go version to 1.23` |
| `docs` | ドキュメント変更 | `docs: Add developer guide` |
| `style` | コードスタイル修正（動作変更なし） | `style: Format code with gofmt` |
| `refactor` | リファクタリング | `refactor: Extract test helper function` |
| `perf` | パフォーマンス改善 | `perf: Optimize schema validation` |
| `test` | テスト追加/修正 | `test: Add validation integration tests` |
| `build` | ビルドシステム変更 | `build: Update Dockerfile` |
| `ci` | CI設定変更 | `ci: Add linter workflow` |
| `chore` | その他の変更 | `chore: Update dependencies` |
| `revert` | revert | `revert: Revert "feat: Add feature"` |

#### 例

**✅ Good**

```bash
git commit -m "feat: Add protovalidate interceptor to ISR service"
git commit -m "fix: Update Go version to 1.23 for Docker compatibility"
git commit -m "test: Add validation tests for 10MB size limit"
git commit -m "docs: Update setup instructions in README"
```

**❌ Bad**

```bash
# 複数行はNG
git commit -m "feat: Add feature

This is a detailed description"

# type がない
git commit -m "Added new feature"

# 形式が間違っている
git commit -m "feat Add feature"  # コロンがない
git commit -m "feature: Add feature"  # type が間違い
```

### Co-Authored-By

Claude Code と共同で作業した場合、以下をコミットメッセージに追加できます（オプション）：

```text
Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

ただし、git hookは1行ルールを強制するため、GitHub UI で後から追加することを推奨します。

## 開発ワークフロー

### ブランチ戦略

1. **main**: 本番環境に対応する安定ブランチ
2. **feature/***: 機能追加用ブランチ
3. **fix/***: バグ修正用ブランチ
4. **docs/***: ドキュメント更新用ブランチ

### 作業の流れ

```bash
# 1. mainブランチを最新化
git checkout main
git pull

# 2. 作業ブランチを作成
git checkout -b feat/add-new-feature

# 3. コード変更とテスト
# ... 開発作業 ...
make test

# 4. コミット（git hookが自動で検証）
git add .
git commit -m "feat: Add new feature"

# 5. プッシュ
git push -u origin feat/add-new-feature

# 6. PRを作成
gh pr create --title "feat: Add new feature" --body "..."
```

### CI/CDチェック

PRを作成すると、以下のチェックが自動実行されます：

- **Proto Lint**: `buf lint`
- **Go Test**: `go test -v -race -coverprofile=coverage.out ./...`
- **Go Lint**: `go vet` + `staticcheck`
- **Build Docker Image**: `docker compose build`

すべてのチェックがパスするまでマージできません。

## コーディング規約

### Go

- **フォーマット**: `gofmt` を使用
- **Linter**: `go vet` + `staticcheck`
- **命名規則**: Go標準に従う
  - Exported: PascalCase
  - Unexported: camelCase
- **エラーハンドリング**: エラーは適切にラップして返す

```go
// Good
if err != nil {
    return fmt.Errorf("failed to upload schema: %w", err)
}

// Bad
if err != nil {
    return err  // コンテキストがない
}
```

### TypeScript

- **フォーマット**: Prettier
- **Linter**: ESLint
- **命名規則**:
  - 変数/関数: camelCase
  - 型/インターフェース: PascalCase
  - 定数: UPPER_SNAKE_CASE

### Proto

- **Style Guide**: [Buf Style Guide](https://buf.build/docs/best-practices/style-guide)
- **Lint**: `buf lint` を実行
- **命名規則**:
  - Message: PascalCase
  - Field: snake_case
  - Service: PascalCase
  - RPC: PascalCase

## テスト

### Go

#### ユニットテスト

```go
func TestSchemaHandler_UploadSchema_Success(t *testing.T) {
    // Arrange
    mockRepo := &mockSchemaRepository{...}
    handler := NewSchemaHandler(mockRepo)

    // Act
    resp, err := handler.UploadSchema(ctx, req)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Msg.Version != "1.0.0" {
        t.Errorf("got %v, want 1.0.0", resp.Msg.Version)
    }
}
```

#### 統合テスト

```go
func TestUploadSchema_ValidationError(t *testing.T) {
    // httptest.NewServer でサーバーを起動
    client, cleanup := newTestClient(t, handler)
    defer cleanup()

    // テスト実行
    _, err := client.UploadSchema(ctx, req)

    // エラーコード検証
    var connectErr *connect.Error
    if !errors.As(err, &connectErr) {
        t.Fatalf("expected connect.Error")
    }
}
```

#### テスト実行

```bash
# 全テスト実行
go test ./...

# カバレッジ付き
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 特定パッケージのみ
go test ./internal/handler -v
```

### TypeScript

```bash
# ユニットテスト
npm test

# カバレッジ付き
npm run test:coverage
```

### E2Eテスト

```bash
# Playwrightでシナリオテスト実行（将来実装予定）
npm run test:e2e
```

## PR作成とレビュー

### PR作成前のチェックリスト

- [ ] すべてのテストが通る (`make test`)
- [ ] コードがフォーマットされている (`make fmt`)
- [ ] Lintエラーがない (`make lint`)
- [ ] CIが全てパスしている
- [ ] コミットメッセージがConventional Commitsに従っている

### PRテンプレート

```markdown
## Summary
この変更の概要を記述

## Changes
- 変更内容1
- 変更内容2

## Test Plan
- [ ] テスト項目1
- [ ] テスト項目2

## Related Issues
- Closes #123
```

### レビュープロセス

1. **自己レビュー**: PR作成後、自分でコードを再確認
2. **CIチェック**: すべてのCIが緑になるまで修正
3. **Copilot Review**: 自動レビューコメントに対応
4. **マージ**: Approve後、mainにマージ

### PRサイズの目安

- **Small**: ~100行（理想）
- **Medium**: 100-300行
- **Large**: 300-500行（避けるべき）
- **Too Large**: 500行以上（分割を検討）

大きな機能は複数のPRに分割することを推奨します。

## トラブルシューティング

### Git hookが動作しない

```bash
# フックを再セットアップ
bash .github/git-hooks/setup-hooks.sh

# 権限を確認
ls -la .git/hooks/commit-msg

# 期待される出力: -rwxr-xr-x
```

### Protoコード生成エラー

```bash
# Bufのバージョン確認
buf --version

# キャッシュクリア後に再生成
rm -rf pkg/gen/
make proto-generate
```

### Dockerサービスが起動しない

```bash
# ログ確認
docker compose logs

# クリーンアップして再起動
make docker-clean
make docker-up
```

## 参考リンク

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Buf Documentation](https://buf.build/docs)
- [Connect Documentation](https://connectrpc.com/docs/introduction)
- [protovalidate](https://github.com/bufbuild/protovalidate)
- [GitHub Flow](https://docs.github.com/en/get-started/quickstart/github-flow)
