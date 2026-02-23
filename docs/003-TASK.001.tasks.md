# Milestone 1: 基盤構築 (Infrastructure, Proto & Shared Types)

## Task 1-1: モノレポ基盤と共通環境の構築

* **Background**: 複数言語・複数サービスを円滑に管理するため、Go Workspaces と Node.js Workspaces を導入します。また、`CELO_` プレフィックスを用いた環境変数規約を確立します。
* **Acceptance Criteria**:
  * ルートに `go.work` が存在し、各サービスを認識していること。
  * 各サービスが `CELO_DB_URL` などの統一的な環境変数で設定可能であること（PostgreSQL は使用しないが、ISR は既存実装を維持）。

* **批判的視点への対策**: 環境変数の形式を全サービスで統一し、不一致による起動失敗を防止。

## Task 1-2: Proto定義と「型専用共有モジュール」の確立

* **Background**: 生成されたコードの配置場所がバラバラだと Import Cycle や補完エラーの原因になります。プロジェクトルートの `pkg/gen` を「型専用の共有ディレクトリ」として定義し、全サービスがここを参照する構造を作ります。
* **Acceptance Criteria**:
  * `buf generate` を実行すると、`pkg/gen/go/` および `pkg/gen/ts/` にコードが出力されること。
  * `pkg/gen/go` には `go.mod` を置かず、Go Workspaces でルートおよび各サービスから参照できること。
  * `buf lint` が CI 上でパスすること。

* **批判的視点への対策**: 生成コードを各サービス内に閉じ込めず、共有ディレクトリ (`pkg/gen`) とすることで、IDE の定義ジャンプや補完を確実に機能させる。

---

## Milestone 1 時点の想定ディレクトリ構造

```bash
building-a-schema-first-dynamic-validation-system/
├── go.work              # Go Workspaces (isr, be, pkg/gen を管理)
├── package.json         # Node Workspaces (fe, e2e を管理)
├── buf.gen.yaml         # pkg/gen/ への出力を定義
├── pkg/
│   └── gen/             # 【重要】生成コード専用の共有モジュール
│       ├── go/
│       │   └── go.mod   # pkg/gen/go として独立
│       └── ts/
├── services/
│   ├── isr/
│   │   └── go.mod       # pkg/gen/go を参照
│   ├── be/
│   │   └── go.mod       # pkg/gen/go を参照
│   └── fe/
└── README.md
```

**注**: docker-compose.yml は今後削除予定です（サービス間依存がないため）

---

## Milestone 2: ISR (レジストリ) & Schema Push

### Task 2-1: ISR サービスの基本実装

* **Background**: スキーマバイナリを SemVer 管理・配信するハブを作成する（参考実装として維持）。
* **Acceptance Criteria**:
  * `UploadSchema` でバイナリ保存、`GetLatestPatch` で最新版を返却できること。
  * PostgreSQL を使用した実装（BE との連携は行わない）。

### Task 2-2: Local Upload スクリプトの実装

* **Background**: 開発環境から ISR へ UUID v7 を伴う最新スキーマを送り込む。
* **Acceptance Criteria**:
  * スクリプト一発で `.proto` ビルドから ISR 登録までが完了すること。

**注**: ISR は参考実装として残しますが、BE との連携は行いません。

---

## Milestone 3: Backend (BE) 実装

### Task 3-1: BE サービスの基盤と User API ✅

* **Background**: ユーザー情報の保存と取得を実装し、IDに UUID v7 を採用する。
* **Acceptance Criteria**:
  * ユーザーの新規作成と一覧取得ができること。

**Status**: 完了（PostgreSQL版として実装済み。Task 3-3 で YAML に移行予定）

### Task 3-2: 動的スキーマ同期 (Hot Reload) 実装 ❌ 制約により不可

* **Background**: 1分周期のポーリングと、バリデーターの Atomic Swap を実装する。
* **実装結果**: ISRポーリング機構とスキーマ取得は実装完了。ログ出力も正常に動作。しかし、protovalidateが静的メッセージの`msg.ProtoReflect().Descriptor()`を参照するため、**バリデーションルールの動的更新は不可能**と判明。詳細は[Design Doc 3 § 7.1](001-DD.003.validation-strategy.md#71-実装結果と重要な制約事項)参照。

**Status**: 技術的制約により中止

### Task 3-3: BE の YAML 永続化実装

* **Background**: PostgreSQL を YAML ファイルに置き換え、シンプルな永続化を実現する。
* **Acceptance Criteria**:
  * User API (CreateUser, ListUsers) が YAML ファイルで動作すること。
  * `services/be/data/user.yaml` にデータが保存されること。
  * 既存の Repository インターフェースを維持すること。

**実装内容**:

* `YAMLUserRepository` の実装
* user.yaml のフォーマット定義
* 起動時の YAML ロード処理

### Task 3-4: Post API と Context Enrichment 実装

* **Background**: Post API を実装し、Context Enrichment（user.plan を使った動的バリデーション）を実現する。
* **Acceptance Criteria**:
  * CreatePost, ListPosts が実装されていること。
  * CreatePost 時に user.yaml から user.plan を取得し、plan に応じた message 長制限を適用すること。
  * free: 100文字、pro: 200文字、enterprise: 300文字
  * `services/be/data/post.yaml` にデータが保存されること。
  * curl/grpcurl でテストできること。

**実装内容**:

* `PostHandler` の実装
* `YAMLPostRepository` の実装
* Context Enrichment ロジック
* post.yaml のフォーマット定義

---

## Milestone 4: Frontend 実装

### Task 4-1: FE の簡易実装

* **Background**: protovalidate-ts を使ったクライアント側バリデーションを実装する。実際の BE 通信は行わず、YAML 永続化による疑似的な実装を行う。
* **Acceptance Criteria**:
  * ユーザー作成画面が実装されていること。
  * 投稿作成・一覧画面が実装されていること。
  * protovalidate-ts によるバリデーションが動作すること。
  * `services/fe/data/user.yaml`, `services/fe/data/post.yaml` にデータが保存されること。
  * BE への実際の通信は行わないこと。

**実装内容**:

* React/Vue/Svelte のいずれかで実装
* protovalidate-ts の組み込み
* YAML 読み書き処理
* 基本的な UI コンポーネント

**スコープ外**:

* ~~BE との実際の通信~~
* ~~スキーマの定期取得~~
* ~~Server Streaming~~
* ~~E2Eテスト~~

---

## Milestone 5: ドキュメント整備

### Task 5-1: 検証手順書の作成

* **Background**: curl/grpcurl を使った検証手順を整備する。
* **Acceptance Criteria**:
  * User API の検証手順が記載されていること。
  * Post API の検証手順が記載されていること（Context Enrichment の検証を含む）。
  * セキュリティ検証（plan 偽装）の手順が記載されていること。

### Task 5-2: README の更新

* **Background**: PoC のスコープ変更を反映し、起動方法を明記する。
* **Acceptance Criteria**:
  * docker-compose.yml が不要になったことが記載されていること。
  * 各サービスの個別起動方法が記載されていること。
  * YAML 永続化の仕組みが説明されていること。
  * PoC の目的（protovalidate + Context Enrichment の検証）が明記されていること。

---

## タスクの優先順位

| Priority | Task | Milestone |
|----------|------|-----------|
| 高 | Task 3-3: BE の YAML 永続化実装 | M3 |
| 高 | Task 3-4: Post API と Context Enrichment 実装 | M3 |
| 中 | Task 4-1: FE の簡易実装 | M4 |
| 中 | Task 5-1: 検証手順書の作成 | M5 |
| 中 | Task 5-2: README の更新 | M5 |
| 低 | Task 2-1, 2-2: ISR 関連 | M2 |

**注**: Milestone 1 は完了済み、Milestone 2 の ISR は参考実装として残すが優先度は低い。
