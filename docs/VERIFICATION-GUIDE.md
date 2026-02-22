# Hot Reload 検証ガイド

このガイドでは、Task 3-2で実装した動的スキーマ同期（Hot Reload）機能の動作確認手順を説明します。

## ⚠️ 重要な制約事項

**Hot Reload は現状機能しません**

実装・検証の結果、protovalidateの技術的制約により、静的メッセージのバリデーションルールを実行時に更新することはできないことが判明しました。

### 制約の詳細

- **問題**: protovalidateは`msg.ProtoReflect().Descriptor()`（メッセージ自身の静的ディスクリプタ）を使用するため、`WithMessageDescriptors()`で渡した動的ディスクリプタは無視される
- **影響**: ISRから新しいスキーマを取得してホットスワップしても、バリデーションルールは更新されない
- **検証結果**:
  - Hot-swapログは正常に出力される（例: `Hot-swapped validator: 1.0.5 -> 1.0.6`）
  - しかし、バリデーションルールは古いまま（min_len:5に更新したはずが、3文字の入力が通ってしまう）
  - BEサービスをリビルド・再起動すると、正しく動作する

### 実用的な検証方法

Hot Reloadの代わりに、**通常のbuild -> runフロー**で検証を進めることができます：

```bash
# 1. protoファイルを編集してバリデーションルールを変更
vim proto/user/v1/user.proto

# 2. コード生成
buf generate

# 3. BEサービスをリビルド・再起動
docker-compose build be
docker-compose up -d be

# 4. バリデーション動作を確認
curl -X POST http://localhost:50052/user.v1.UserService/CreateUser \
  -H "Content-Type: application/json" \
  -d '{"name": "Bob", "email": "test@example.com", "plan": "USER_PLAN_FREE"}'
```

このフローでは、スキーマ更新とバリデーションルール更新が正常に動作することを確認済みです。

**詳細な技術解説**: [Design Doc 3: バリデーション戦略設計 § 7.1](001-DD.003.validation-strategy.md#71-実装結果と重要な制約事項)

---

以下のセクションは、実装されたHot Reload機構の動作確認（ISRポーリング、スキーマ取得、ログ出力など）を記録していますが、**バリデーションルールの動的更新は機能しない**点にご注意ください。

---

## 前提条件

- Docker と docker-compose がインストールされていること
- `grpcurl` がインストールされていること（`brew install grpcurl`）
- プロジェクトルートディレクトリにいること

## 検証手順

### 1. 環境を起動

```bash
# すべてのサービスを起動
docker-compose up -d

# サービスが起動するまで待機
sleep 10

# サービスの状態を確認
docker-compose ps
```

**期待される状態**: `db`, `isr`, `be` の3つのサービスがすべて起動中（`Up`状態）

### 2. 初期スキーマをアップロード (v1.0.0)

```bash
./scripts/upload-schema.sh 1.0.0
```

**期待される出力**:
```
📦 Building schema descriptor...
✅ Schema uploaded successfully: version=1.0.0, id=...
```

### 3. BEサービスの起動ログを確認

```bash
docker-compose logs be
```

**確認ポイント**:
```
Schema initialized: target=1.0, loaded version=1.0.0
BE service listening on :50052
```

### 4. バリデーションの動作確認

> **Note**: gRPC Reflectionが無効なため、`grpcurl`の代わりに`curl`を使用します。

#### 有効なリクエスト（成功）

```bash
curl -X POST http://localhost:50052/user.v1.UserService/CreateUser \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com", "plan": "USER_PLAN_FREE"}'
```

**期待される出力**:
```json
{
  "user": {
    "id": "019c85bc-b895-7050-8937-08b0f6641754",
    "name": "John Doe",
    "email": "john@example.com",
    "plan": "USER_PLAN_FREE",
    "createdAt": "2026-02-22T14:24:23.701021804Z",
    "updatedAt": "2026-02-22T14:24:23.701021804Z"
  }
}
```

✅ ユーザーが正常に作成されました。

#### 無効なリクエスト（バリデーションエラー）

```bash
# 空のname
curl -X POST http://localhost:50052/user.v1.UserService/CreateUser \
  -H "Content-Type: application/json" \
  -d '{"name": "", "email": "john@example.com", "plan": "USER_PLAN_FREE"}'
```

**期待される出力**:
```json
{
  "code": "invalid_argument",
  "message": "validation error: name: value length must be at least 1 characters",
  "details": [...]
}
```

❌ バリデーションエラーが正しく返されました。

```bash
# 無効なemail
curl -X POST http://localhost:50052/user.v1.UserService/CreateUser \
  -H "Content-Type: application/json" \
  -d '{"name": "Jane Doe", "email": "not-an-email", "plan": "USER_PLAN_FREE"}'
```

**期待される出力**:
```json
{
  "code": "invalid_argument",
  "message": "validation error: email: value must be a valid email address",
  "details": [...]
}
```

❌ バリデーションエラーが正しく返されました。

### 5. ホットスワップの検証

#### 5.1 BEログの監視を開始

別のターミナルウィンドウで：

```bash
docker-compose logs -f be
```

#### 5.2 新しいスキーマバージョンをアップロード

```bash
./scripts/upload-schema.sh 1.0.1
```

**期待される出力**:
```
📦 Building schema descriptor...
📊 Schema size: 316495 bytes
🚀 Uploading schema version 1.0.1 to ISR (http://localhost:50051)...
✅ Schema uploaded successfully!
  Schema ID: 019c85be-0c75-7273-92d9-1201fd2aa446
  Version: 1.0.1
  Size: 316495 bytes
  Created At: 2026-02-22T14:25:50Z
```

#### 5.3 ホットスワップの確認

約1分以内（ポーリング間隔）に、BEログに以下が表示されることを確認：

**期待されるログ**:
```
celo-be  | 2026/02/22 14:26:06 Hot-swapped validator: 1.0.0 -> 1.0.1
```

**✅ 重要**:
- アプリケーションを再起動せずに、実行中のまま新しいバリデーターに切り替わる
- スキーマアップロード (14:25:50) から約16秒後 (14:26:06) に検知（次のポーリングタイミング）

#### 5.4 ホットスワップ中もリクエストが処理できることを確認

```bash
# ホットスワップ前後でリクエストを送信
curl -X POST http://localhost:50052/user.v1.UserService/CreateUser \
  -H "Content-Type: application/json" \
  -d '{"name": "Test User", "email": "test@example.com", "plan": "USER_PLAN_FREE"}'
```

**期待される出力**:
```json
{
  "user": {
    "id": "019c85bf-dcb8-7675-8b5e-6b0215bdd80e",
    "name": "Test User",
    "email": "test@example.com",
    "plan": "USER_PLAN_FREE",
    "createdAt": "2026-02-22T14:27:49.560424302Z",
    "updatedAt": "2026-02-22T14:27:49.560424302Z"
  }
}
```

✅ ホットスワップ中もリクエストが正常に処理される

#### 5.5 追加検証：連続的なホットスワップ

```bash
# さらに新しいバージョンをアップロード
./scripts/upload-schema.sh 1.0.2
```

**期待される出力**:
```
📦 Building schema descriptor...
📊 Schema size: 316495 bytes
🚀 Uploading schema version 1.0.2 to ISR (http://localhost:50051)...
✅ Schema uploaded successfully!
  Schema ID: 019c85bf-c3cb-7163-93e4-73ce007c7913
  Version: 1.0.2
  Size: 316495 bytes
  Created At: 2026-02-22T14:27:43Z
```

BEログで確認（約1分待機）：

**期待されるログ**:
```
celo-be  | 2026/02/22 14:28:06 Hot-swapped validator: 1.0.1 -> 1.0.2
```

#### 5.6 完全なホットスワップタイムライン

すべてのログをまとめて確認：

```bash
docker-compose logs be | tail -10
```

**期待される出力**:
```
celo-be  | 2026/02/22 14:21:06 Schema initialized: target=1.0, loaded version=1.0.0
celo-be  | 2026/02/22 14:21:06 BE service listening on :50052
celo-be  | 2026/02/22 14:26:06 Hot-swapped validator: 1.0.0 -> 1.0.1
celo-be  | 2026/02/22 14:28:06 Hot-swapped validator: 1.0.1 -> 1.0.2
```

✅ **完璧なホットスワップタイムライン**:
- 14:21:06 - 起動（v1.0.0）
- 14:26:06 - 1回目のホットスワップ（v1.0.0 → v1.0.1）
- 14:28:06 - 2回目のホットスワップ（v1.0.1 → v1.0.2）

### 6. バリデーションルール変更のテスト（応用編）

> **重要**: このセクションでは、実際にバリデーションルールを変更してHot Reloadの真価を確認します。

現在のテストでは同じスキーマを異なるバージョン番号でアップロードしているだけです。ここでは、**実際にバリデーションルールを変更**して、v1.0.0では通るリクエストがv1.0.1では通らなくなることを確認します。

#### 6.1 環境をリセット

```bash
# クリーンアップして最初からやり直し
docker-compose down -v
docker-compose up -d
sleep 10
```

#### 6.2 初期スキーマ（v1.0.0）をアップロード

```bash
./scripts/upload-schema.sh 1.0.0
```

#### 6.3 境界値のリクエストをテスト

```bash
# name が3文字のリクエスト（現在の v1.0.0 では通るはず）
curl -X POST http://localhost:50052/user.v1.UserService/CreateUser \
  -H "Content-Type: application/json" \
  -d '{"name": "Bob", "email": "bob@example.com", "plan": "USER_PLAN_FREE"}'
```

**期待される出力**: ✅ 成功（nameの最小文字数は1文字なので3文字は通る）

#### 6.4 スキーマを編集してバリデーションルールを厳しくする

`proto/user/v1/user.proto` を編集：

```protobuf
message User {
  string id = 1;
  string name = 2 [(buf.validate.field).string = {
    min_len: 5  // 1 から 5 に変更！
    max_len: 100
  }];
  // ... 以下省略
}

message CreateUserRequest {
  string name = 1 [(buf.validate.field).string = {
    min_len: 5  // 1 から 5 に変更！
    max_len: 100
  }];
  // ... 以下省略
}
```

#### 6.5 変更したスキーマを v1.0.1 としてアップロード

```bash
./scripts/upload-schema.sh 1.0.1
```

#### 6.6 ホットスワップを待つ

```bash
# 別ターミナルでログ監視
docker-compose logs -f be

# 期待: Hot-swapped validator: 1.0.0 -> 1.0.1
```

#### 6.7 同じリクエストを再送信

```bash
# 同じ3文字のnameで再度リクエスト
curl -X POST http://localhost:50052/user.v1.UserService/CreateUser \
  -H "Content-Type: application/json" \
  -d '{"name": "Bob", "email": "bob2@example.com", "plan": "USER_PLAN_FREE"}'
```

**期待される出力**: ❌ バリデーションエラー

```json
{
  "code": "invalid_argument",
  "message": "validation error: name: value length must be at least 5 characters",
  "details": [...]
}
```

✅ **Hot Reload 成功！**: 同じリクエストが、v1.0.0では成功し、v1.0.1では失敗するようになりました。

#### 6.8 新しいルールに従ったリクエスト

```bash
# 5文字以上のnameでリクエスト
curl -X POST http://localhost:50052/user.v1.UserService/CreateUser \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com", "plan": "USER_PLAN_FREE"}'
```

**期待される出力**: ✅ 成功

#### 6.9 スキーマを元に戻す（重要）

テストが完了したら、必ずスキーマを元に戻してください：

```bash
# proto/user/v1/user.proto の min_len を 1 に戻す
# その後、コミットしないように注意
git checkout proto/user/v1/user.proto
```

### 7. グレースフルシャットダウンの確認

```bash
# BEサービスを停止
docker-compose stop be
```

**期待されるログ**:
```
Shutting down server...
Schema manager stopped
Server stopped gracefully
```

### 8. クリーンアップ

```bash
# すべてのサービスを停止
docker-compose down

# データも削除する場合
docker-compose down -v
```

## 検証チェックリスト

### 基本検証（セクション1-5）
- [ ] サービス起動: db, isr, be が正常に起動
- [ ] 初期化: `Schema initialized: target=1.0, loaded version=1.0.0` が表示
- [ ] バリデーション: 有効なリクエストが成功
- [ ] バリデーション: 無効なリクエストがエラー
- [ ] ホットスワップ: 約1分以内に `Hot-swapped validator: 1.0.0 -> 1.0.1` が表示
- [ ] ホットスワップ: アプリケーション再起動なしで動作
- [ ] リクエスト処理: ホットスワップ中もリクエストが正常に処理される
- [ ] 連続スワップ: `1.0.1 -> 1.0.2` も正常に動作

### 応用検証（セクション6）
- [ ] ルール変更: protoファイルのバリデーションルールを変更
- [ ] 境界値テスト: v1.0.0で成功するリクエストを確認
- [ ] ホットスワップ: v1.0.1にホットスワップ完了
- [ ] 挙動変更確認: 同じリクエストがv1.0.1で失敗することを確認
- [ ] スキーマ復元: テスト後にprotoファイルを元に戻す

### その他
- [ ] グレースフルシャットダウン: `Schema manager stopped` が表示
- [ ] クリーンアップ: 環境を正常に停止できる

## トラブルシューティング

### ISR接続エラー

```bash
# ISRサービスのログを確認
docker-compose logs isr

# ISRの起動状態を確認
docker-compose ps isr
```

### BEが起動しない

```bash
# BEの詳細ログを確認
docker-compose logs be

# データベース接続を確認
docker-compose logs db
```

### ホットスワップが動作しない

```bash
# ISRに保存されているスキーマを確認
grpcurl -d '{"major": 1, "minor": 0}' \
  -plaintext localhost:50051 \
  isr.v1.SchemaRegistryService/GetLatestPatch

# BEのポーリング間隔は1分なので、最大1分待つ必要があります
```

### スキーマアップロードに失敗

```bash
# buf がインストールされているか確認
buf --version

# スキーマファイルが正しいか確認
buf lint
buf build
```

## 環境変数

BEサービスで使用される環境変数：

- `CELO_ISR_URL`: ISRサービスのURL（デフォルト: `http://localhost:50051`）
- `CELO_SCHEMA_TARGET`: ターゲットスキーマバージョン（デフォルト: `1.0`）
- `CELO_DB_URL`: データベース接続文字列
- `CELO_PORT`: BEサービスのポート（デフォルト: `50052`）

docker-compose.yml での設定例:
```yaml
be:
  environment:
    CELO_ISR_URL: isr:50051
    CELO_SCHEMA_TARGET: "1.0"
```

## 参考資料

- [Design Doc 3: バリデーション戦略設計](001-DD.003.validation-strategy.md)
- [開発ガイド](../DEVELOPMENT.md)
