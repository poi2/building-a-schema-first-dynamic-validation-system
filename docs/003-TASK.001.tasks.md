# Milestone 1: 基盤構築 (Infrastructure, Proto & Shared Types

## Task 1-1: モノレポ基盤と共通環境の構築

* **Background**: 複数言語・複数サービスを円滑に管理するため、Go Workspaces と Node.js Workspaces を導入します。また、ポート衝突を防ぐ動的マッピングと、`CELO_` プレフィックスを用いた環境変数規約を確立します。
* **Acceptance Criteria**:
* ルートに `go.work` が存在し、`./pkg/gen` および各サービスを認識していること。
* `docker compose up -d` で Postgres が起動し、`init.sh` により `irs_db`, `be_db` が作成されること。
* 各サービスが `CELO_DB_URL` などの統一的な環境変数で設定可能であること。


* **批判的視点への対策**: 接続文字列の形式を全サービスで統一（`postgres://user:pass@host:port/db`）し、環境変数の不一致による起動失敗を防止。

## Task 1-2: Proto定義と「型専用共有モジュール」の確立

* **Background**: 生成されたコードの配置場所がバラバラだと Import Cycle や補完エラーの原因になります。プロジェクトルートの `pkg/gen` を「型専用の Go モジュール」として定義し、全サービスがここを参照する構造を作ります。
* **Acceptance Criteria**:
* `buf generate` を実行すると、`pkg/gen/go/` および `pkg/gen/ts/` にコードが出力されること。
* `pkg/gen/go` に独自の `go.mod` があり、`be` や `isr` サービスから `import "github.com/user/celo/pkg/gen/go/..."` として参照できること。
* `buf lint` が CI 上でパスすること。


* **批判的視点への対策**: 生成コードを各サービス内に閉じ込めず、共有パッケージ (`pkg/gen`) とすることで、IDE の定義ジャンプや補完を確実に機能させる。

---

## Milestone 1 時点の想定ディレクトリ構造

```bash
celo/
├── go.work              # Go Workspaces (isr, be, pkg/gen を管理)
├── package.json         # Node Workspaces (bff, fe, e2e を管理)
├── buf.gen.yaml         # pkg/gen/ への出力を定義
├── pkg/
│   └── gen/             # 【重要】生成コード専用の共有モジュール
│       ├── go/
│       │   └── go.mod   # pkg/gen/go として独立
│       └── ts/
├── services/
│   ├── isr/
│   │   └── go.mod       # pkg/gen/go を参照
│   └── be/
│       └── go.mod       # pkg/gen/go を参照
├── init-db/
│   └── init.sh          # DB分離ロジック
└── docker-compose.yml   # 動的ポート割り当て

```

---

## Milestone 2: ISR (レジストリ) & Schema Push

### Task: ISR サービスの基本実装

* **Background**: スキーマバイナリを SemVer 管理・配信するハブを作成する。
* **Acceptance Criteria**:
* `UploadSchema` でバイナリ保存、`GetLatestPatch` で最新版を返却できること。



### Task: Local Upload スクリプトの実装

* **Background**: 開発環境から ISR へ UUID v7 を伴う最新スキーマを送り込む。
* **Acceptance Criteria**:
* スクリプト一発で `.proto` ビルドから ISR 登録までが完了すること。



---

## Milestone 3: Backend (BE) 実装 & ホットリロード

### Task: BE サービスの基盤と User API

* **Background**: ユーザー情報の保存と取得を実装し、IDに UUID v7 を採用する。
* **Acceptance Criteria**:
* ユーザーの新規作成と一覧取得ができること。



### Task: 動的スキーマ同期 (Hot Reload) 実装

* **Background**: 1分周期のポーリングと、バリデーターの Atomic Swap を実装する。
* **Acceptance Criteria**:
* アプリを止めずに、ログ上でスキーマバージョンの更新が確認できること。



### Task: Post API と Context Enrichment 実装

* **Background**: `_plan` 注入と `protovalidate` による検証を実装。
* **Acceptance Criteria**:
* **Testcontainers-go** による統合テストがパスすること。
* バリデーションエラーに `Design Doc 5` 仕様の `message_id` 等が含まれること。



---

## Milestone 4: Frontend & BFF & E2E

### Task: BFF (Node.js) の実装

* **Background**: フロントエンド向けのプロキシと、スキーマ配信エンドポイントを提供。
* **Acceptance Criteria**:
* レスポンスヘッダーに `X-Schema-Version` を付与し、APIプロキシが機能すること。



### Task: Frontend (React) と動的バリデーション UI

* **Background**: 即時フィードバックとバックグラウンドでのスキーマ更新を実装。
* **Acceptance Criteria**:
* スキーマ更新後、リロードなしで入力エラーの閾値が変化すること。



### Task: シナリオテスト (Playwright) の導入

* **Background**: YAML シナリオに基づく全系の自動検証を Testcontainers 上で行う。
* **Acceptance Criteria**:
* 全シナリオ（正常・異常・偽装リクエスト）が CI 上で完走すること。
