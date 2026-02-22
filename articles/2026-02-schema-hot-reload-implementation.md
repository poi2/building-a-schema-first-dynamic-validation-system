# Protocol Buffers バリデーションの Hot Reload を試みて失敗した話 - protovalidate の制約と学び

> **Status**: Draft
> **Target**: Zenn / 会社ブログ
> **Tags**: #protobuf #go #schemaregistry #protovalidate #失敗から学ぶ

## TL;DR

- Protocol Buffers のバリデーションスキーマを実行時更新（Hot Reload）しようとした
- ISR（Internal Schema Registry）からのポーリング、`atomic.Value`でのスレッドセーフな差し替えを実装
- **結果**: スキーマ取得とホットスワップは成功したが、バリデーションルールは更新されなかった
- **原因**: protovalidate は静的メッセージの埋め込みディスクリプタを使用し、動的ディスクリプタを無視する
- **学び**: ライブラリの内部動作を理解せずに設計すると、実装後に根本的な制約に気づく

## 背景・モチベーション

### やりたかったこと

マイクロサービスで Protocol Buffers を使った開発で、以下を実現したいと考えました：

1. **バリデーションルールの無停止更新**
   - 例: ユーザープランごとの投稿文字数制限（Free: 100文字 → 50文字）
   - アプリを再起動せずにルール変更を反映

2. **複数サービス間でのスキーマ同期**
   - FE、BE で同じバリデーションルールを共有
   - バージョンの不整合を防ぐ

3. **段階的なロールアウト**
   - A/B テストやカナリアリリースで異なるバリデーションルールを適用

### なぜ失敗が起きたか

結論から言うと、**事前調査が不足していました**。protovalidate のドキュメントを読んだだけで、実際のソースコードを確認せずに「`WithMessageDescriptors()` で動的ディスクリプタを渡せば動くだろう」と楽観的に考えていました。

## 実装したアーキテクチャ

### 全体構成

```
┌─────────┐     ┌─────────┐     ┌─────────┐
│   FE    │────▶│   BE    │────▶│   DB    │
└─────────┘     └────┬────┘     └─────────┘
                     │
                     │ ポーリング (1分間隔)
                     │ GetLatestPatch(major, minor)
                     ▼
                ┌─────────┐
                │   ISR   │ Internal Schema Registry
                │(Postgres)│ スキーマをバージョン管理
                └─────────┘
```

### コンポーネント設計

実装したコンポーネント：

1. **ISR (Internal Schema Registry)**
   - スキーマバイナリ（FileDescriptorSet）をバージョン付きで保存
   - SemVer（Major.Minor.Patch）で管理
   - `GetLatestPatch(major, minor)` API でパッチバージョンを取得

2. **SchemaManager** (`services/be/internal/schemamanager/`)
   - ISR クライアント管理
   - 1分間隔でのポーリング
   - バージョン比較とホットスワップトリガー

3. **SchemaAwareValidator** (`services/be/internal/validator/`)
   - `atomic.Value` でバリデーターを保持
   - スレッドセーフな読み取りと更新
   - バージョン情報の管理

## 実装詳細

### 1. スレッドセーフなバリデーター管理

```go
type validatorWithVersion struct {
    validator protovalidate.Validator
    version   string
}

type SchemaAwareValidator struct {
    v atomic.Value // *validatorWithVersion
}

func (s *SchemaAwareValidator) Validate(msg proto.Message, options ...protovalidate.ValidationOption) error {
    vwv := s.v.Load().(*validatorWithVersion)
    return vwv.validator.Validate(msg, options...)
}

func (s *SchemaAwareValidator) UpdateSchema(descriptorBytes []byte, version string) error {
    // 1. FileDescriptorSet をアンマーシャル
    fds := &descriptorpb.FileDescriptorSet{}
    if err := proto.Unmarshal(descriptorBytes, fds); err != nil {
        return err
    }

    // 2. Files registry を作成
    files, err := protodesc.NewFiles(fds)
    if err != nil {
        return err
    }

    // 3. Extension registry を作成（buf.validate 拡張のため）
    extensionRegistry := &protoregistry.Types{}
    files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
        extensions := fd.Extensions()
        for i := 0; i < extensions.Len(); i++ {
            ext := extensions.Get(i)
            extensionRegistry.RegisterExtension(dynamicpb.NewExtensionType(ext))
        }
        return true
    })

    // 4. メッセージディスクリプタを収集
    var descriptors []protoreflect.MessageDescriptor
    files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
        messages := fd.Messages()
        for i := 0; i < messages.Len(); i++ {
            descriptors = append(descriptors, messages.Get(i))
        }
        return true
    })

    // 5. 新しいバリデーターを作成
    validator, err := protovalidate.New(
        protovalidate.WithMessageDescriptors(descriptors...),
        protovalidate.WithExtensionTypeResolver(extensionRegistry),
    )
    if err != nil {
        return err
    }

    // 6. atomic.Value で差し替え
    s.v.Store(&validatorWithVersion{
        validator: validator,
        version:   version,
    })

    return nil
}
```

### 2. ISR ポーリングとホットスワップ

```go
func (m *SchemaManager) pollLoop(ctx context.Context) {
    ticker := time.NewTicker(m.config.PollingInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if err := m.checkAndUpdateSchema(ctx); err != nil {
                log.Printf("Schema polling error (will retry in %s): %v",
                    m.config.PollingInterval, err)
            }
        case <-m.stopCh:
            return
        case <-ctx.Done():
            return
        }
    }
}

func (m *SchemaManager) checkAndUpdateSchema(ctx context.Context) error {
    currentVersion := m.validator.GetCurrentVersion()

    req := connect.NewRequest(&isrv1.GetLatestPatchRequest{
        Major: m.config.Major,
        Minor: m.config.Minor,
    })

    resp, err := m.client.GetLatestPatch(ctx, req)
    if err != nil {
        return err
    }

    latestVersion := resp.Msg.Metadata.Version

    if currentVersion == latestVersion {
        return nil  // 更新不要
    }

    if err := m.validator.UpdateSchema(resp.Msg.SchemaBinary, latestVersion); err != nil {
        return err
    }

    log.Printf("Hot-swapped validator: %s -> %s", currentVersion, latestVersion)
    return nil
}
```

## 問題の発見 - 動いているように見えたが...

### 初期検証: 成功したと思った

```bash
# 1. v1.0.5 をアップロード
./scripts/upload-schema.sh 1.0.5

# 2. BE ログを確認
# ✅ "Schema initialized: target=1.0, loaded version=1.0.5"

# 3. v1.0.6 をアップロード（min_len: 1 → 5 に変更）
./scripts/upload-schema.sh 1.0.6

# 4. 1分後のログ
# ✅ "Hot-swapped validator: 1.0.5 -> 1.0.6"
```

この時点では「完璧だ！」と思いました。ログも正常、エラーもなし。

### 異変に気づく

バリデーションルールが本当に更新されたか確認するため、境界値テストを実施：

```bash
# min_len: 5 に変更したので、3文字の "Bob" は拒否されるはず
curl -X POST http://localhost:50052/user.v1.UserService/CreateUser \
  -H "Content-Type: application/json" \
  -d '{"name": "Bob", "email": "bob@example.com", "plan": "USER_PLAN_FREE"}'
```

**結果**: ✅ 成功（ユーザーが作成された）

**期待**: ❌ バリデーションエラー（name は最低5文字必要）

## 調査: なぜバリデーションルールが更新されないのか

### 仮説1: スキーマがISRに正しく保存されていない？

```bash
# ISRから直接スキーマを取得して確認
protoc --decode=google.protobuf.FileDescriptorSet \
  google/protobuf/descriptor.proto < /tmp/isr-schema.bin \
  | grep -A 5 "CreateUserRequest"

# 結果: min_len: 5 が確認できた → ISR は正常
```

### 仮説2: UpdateSchema() が失敗している？

デバッグログを追加：

```go
func (s *SchemaAwareValidator) UpdateSchema(...) error {
    // ...
    log.Printf("[DEBUG] Found %d message descriptors", len(descriptors))
    for i, desc := range descriptors {
        log.Printf("[DEBUG] Descriptor %d: %s", i, desc.FullName())
    }
    // ...
}
```

**結果**: 75個のディスクリプタが正常に読み込まれていた → UpdateSchema も正常

### 仮説3: Extension が登録されていない？

buf.validate のバリデーションルールは Protocol Buffer の extension として定義されています。拡張が登録されているか確認：

```go
extensionRegistry := &protoregistry.Types{}
extensionCount := 0
files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
    extensions := fd.Extensions()
    for i := 0; i < extensions.Len(); i++ {
        ext := extensions.Get(i)
        if err = extensionRegistry.RegisterExtension(...); err == nil {
            extensionCount++
        }
    }
    return true
})
log.Printf("[DEBUG] Total extensions registered: %d", extensionCount)
```

**結果**: 4個の拡張（buf.validate.message、buf.validate.field など）が正常に登録されていた → 拡張登録も正常

### 真犯人: protovalidate のソースコード

すべての可能性を消去法で潰した後、protovalidate のソースコードを直接確認しました：

```go
// github.com/bufbuild/protovalidate-go/validator.go
func (v *validator) Validate(
    msg proto.Message,
    options ...ValidationOption,
) error {
    // ...
    refl := msg.ProtoReflect()
    eval := v.builder.Load(refl.Descriptor())  // ← ここ！
    // ...
}
```

**衝撃の事実**: `msg.ProtoReflect().Descriptor()` を使用している！

つまり、`WithMessageDescriptors()` で渡した動的ディスクリプタは**使用されず**、メッセージ自身が持つ**静的ディスクリプタ**が使用されます。

## 根本原因: 静的メッセージの制約

### Protocol Buffers の2つのメッセージ形態

1. **静的メッセージ** (`*userv1.CreateUserRequest`)
   - `buf generate` で生成された Go 構造体
   - コンパイル時に `.pb.go` ファイルに埋め込まれたディスクリプタを使用
   - `ProtoReflect().Descriptor()` は常にコンパイル時のディスクリプタを返す

2. **動的メッセージ** (`*dynamicpb.Message`)
   - 実行時に `protoreflect.MessageDescriptor` から生成
   - ディスクリプタを外部から注入可能

### なぜ動かないか

```
┌──────────────────────────────────────┐
│ BE Service (Go Binary)               │
│                                      │
│  ┌────────────────────────────────┐ │
│  │ userv1.CreateUserRequest       │ │
│  │   ↓ ProtoReflect()             │ │
│  │ 埋め込まれたディスクリプタ      │ │  ← コンパイル時に固定
│  │ (min_len: 1)                   │ │
│  └────────────────────────────────┘ │
│                                      │
│  ┌────────────────────────────────┐ │
│  │ SchemaAwareValidator           │ │
│  │   動的ディスクリプタ            │ │  ← 実行時に更新可能
│  │   (min_len: 5)                 │ │
│  └────────────────────────────────┘ │
│                    ↑                 │
│                    使われない！       │
└──────────────────────────────────────┘
```

protovalidate は `msg.ProtoReflect().Descriptor()` を呼び出すため、常に**埋め込まれたディスクリプタ**（min_len: 1）が使用されます。

### 検証: リビルドすると動く

```bash
# 1. proto ファイルを編集（min_len: 1 → 5）
vim proto/user/v1/user.proto

# 2. コード生成
buf generate

# 3. BE サービスをリビルド・再起動
docker-compose build be
docker-compose up -d be

# 4. 同じリクエストを送信
curl -X POST http://localhost:50052/user.v1.UserService/CreateUser \
  -H "Content-Type: application/json" \
  -d '{"name": "Bob", "email": "bob@example.com", "plan": "USER_PLAN_FREE"}'
```

**結果**: ❌ バリデーションエラー（正しく拒否された）

これで確信しました。**Hot Reload は不可能**です。

## 代替アプローチ

### 1. 動的メッセージへの全面移行（理論上可能）

静的メッセージではなく、`dynamicpb.Message` を使用すれば Hot Reload は可能です。

**メリット**:
- 実行時にディスクリプタを差し替え可能
- 真の Hot Reload が実現できる

**デメリット**:
- 型安全性の喪失（コンパイル時のチェックがない）
- Connect RPC ハンドラーを含む全面的な書き換えが必要
- パフォーマンスオーバーヘッド
- 開発体験の悪化（IDE の補完が効かない）

### 2. カスタムバリデーター実装

protovalidate を使わず、動的ディスクリプタから直接バリデーションルールを読み取る独自実装。

**メリット**:
- 完全なコントロール
- Hot Reload 可能

**デメリット**:
- CEL エンジンの統合が必要
- protovalidate の豊富な機能を再実装
- メンテナンスコストが高い

### 3. 現実的なデプロイ戦略（推奨）

Hot Reload を諦め、標準的なデプロイ戦略で対応：

**Blue-Green Deployment**:
- 新スキーマでビルドした新バージョンを別環境に展開
- トラフィックを切り替え
- ダウンタイムほぼゼロ

**Rolling Update**:
- Kubernetes で順次ポッドを更新
- 段階的なロールアウト
- 問題発生時は即座にロールバック

**Canary Release**:
- 一部のトラフィックだけ新バージョンに流す
- リスクを最小化

## 学んだこと

### 1. 事前のソースコード調査の重要性

ドキュメントだけでなく、**実際のソースコードを読む**べきでした。

`WithMessageDescriptors()` のドキュメントには「warming up the Validator」と書いてあり、「動的ディスクリプタを使う」とは書いていませんでした。

### 2. プロトタイプでの早期検証

ISR やスキーママネージャーを実装する前に、**最小限のプロトタイプで検証**すべきでした：

```go
// 最小限の検証コード
func main() {
    // 動的ディスクリプタでバリデーターを作成
    validator, _ := protovalidate.New(
        protovalidate.WithMessageDescriptors(dynamicDescriptor),
    )

    // 静的メッセージで検証
    req := &userv1.CreateUserRequest{Name: "Bob"}
    err := validator.Validate(req)

    // これだけで「動かない」ことが分かる
}
```

### 3. 失敗も価値がある

実装したコンポーネント（ISR、SchemaManager、SchemaAwareValidator）は、将来的に以下で活用できます：

- スキーマバージョン管理の基盤
- マイグレーション戦略の検証
- 動的メッセージへの移行時の資産

### 4. 技術選定の判断基準

**ライブラリの内部動作を理解してから設計する**

今回は以下を怠りました：
- protovalidate が静的ディスクリプタに依存することの確認
- Protocol Buffers のリフレクション API の理解
- 動的メッセージと静的メッセージの違いの認識

## まとめ

Protocol Buffers のバリデーション Hot Reload を実装しようとして失敗しましたが、得られた知見は大きかったです：

- ✅ protovalidate の動作仕様を深く理解できた
- ✅ Protocol Buffers のリフレクション API を学んだ
- ✅ スキーマレジストリの実装経験を得た
- ✅ Go の並行処理（atomic.Value）の実践
- ❌ Hot Reload は現実的でないことを確認

**結論**: 静的メッセージを使う限り、バリデーションルールの Hot Reload は不可能。スキーマ更新には Blue-Green デプロイメントなど、実績ある手法を使うべき。

## 参考資料

- [protovalidate-go のソースコード](https://github.com/bufbuild/protovalidate-go/blob/main/validator.go)
- [Protocol Buffers Reflection API](https://pkg.go.dev/google.golang.org/protobuf/reflect/protoreflect)
- [プロジェクト実装詳細](../docs/001-DD.003.validation-strategy.md)

## 続編予告

次回は「失敗から学んで、実際に動くものを作る」編として、以下を検討中：

1. 動的メッセージでの Hot Reload 実装
2. カスタム CEL バリデーターの実装
3. Blue-Green デプロイメントでのスキーマ更新戦略

失敗は成功のもと。この経験が誰かの役に立てば幸いです。
