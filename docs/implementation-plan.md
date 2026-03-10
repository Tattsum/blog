# 個人ブログ実装プラン

このドキュメントは、既存のアーキテクチャ設計・API 仕様・ADR を前提に、実際のアプリケーション実装を段階的に進めるための実行プランをまとめたものです。

- **対象リポジトリ**: `blog`
- **前提ドキュメント**:
  - アーキテクチャ: `docs/architecture.md`
  - API 仕様: `docs/api-specification.md`
  - ADR: `docs/adr/001-connect-rpc.md`, `docs/adr/002-hosting-cloudflare-and-cloudrun.md`
- **設計方針**: Go のドメイン駆動設計（DDD）を意識し、可能な範囲でテスト駆動開発（TDD）を行う。セキュリティとパフォーマンスを常に考慮する。

---

## 進捗（現状）

| フェーズ | 状態 | 備考 |
| --- | --- | --- |
| フェーズ 0 | 完了 | Go/buf/Next.js/CI 整備済み。Renovate・golangci-lint v2 も導入済み。 |
| フェーズ 1 | 完了 | ドメイン・リポジトリIF・MySQL・seed コマンドまで実装済み。 |
| フェーズ 2 | 完了 | Post/Tag/Auth/AI 各サービス実装済み。Bearer セッションと X-Admin-Key 併存。残りはテスト・エラー共通化（任意）。 |
| フェーズ 3 | 完了 | 閲覧系・管理画面・タグ別一覧・AI 連携 UI まで実装済み。残りは Vertex AI 連携・SSG/ISR 最適化。 |
| フェーズ 4 | 一部 | 認証・バリデーションは実施済み。セッション固定/CSRF・N+1/キャッシュ調整は未実施。 |
| フェーズ 5 | 一部 | CI/CD・deploy-api.yml・setup-deploy-checklist.md まで完了。ログ・監視・アラートは未実施。 |
| フェーズ 6 | 未着手 | コメント・RSS・マルチテナント・監査ログは計画のみ。 |

---

## フェーズ 0: 基盤・ツールチェーン整備

- **Go / モジュール構成**
  - `backend/` に Go モジュールを作成（`go.mod`、`cmd/server/main.go`、`internal/`）。
  - DDD を意識して `internal/domain/`, `internal/application/`, `internal/infrastructure/`, `internal/interface/`（handler 層）などのレイヤを切る。
- **buf / コード生成**
  - ルートに `buf.yaml`, `buf.gen.yaml` を追加。
  - 既存の `proto/blog/v1/*.proto` から Go / Connect 用コードを `backend/gen/` に生成。
  - `make` もしくは `npm scripts` などで `buf lint`, `buf generate` をコマンド化。
- **Node / Next.js プロジェクト雛形**
  - `frontend/` に Next.js 16（App Router 前提）のプロジェクトを作成。
  - TypeScript / ESLint / Prettier / markdownlint 等を設定。
  - `@connectrpc/connect` / `@connectrpc/connect-web` などを追加。
- **CI の最初の土台（実施済み）**
  - `.github/workflows/ci.yml`: push/PR で `npm run lint:md`, `buf lint`, `npm run generate:proto`, `go test ./...`, `golangci-lint run ./...`, フロントエンド `npm run build` を実行。

---

## フェーズ 1: ドメインモデリングとリポジトリ層

- **ドメインモデル定義（Go）**
  - `backend/internal/domain/post`, `domain/tag`, `domain/user` を定義済み。Post/Tag/User エンティティ、Slug/Email 値オブジェクト、Post の公開判定（IsPublished / Publish / Unpublish）を実装。
- **リポジトリインターフェース**
  - `backend/internal/domain/repository` に `PostRepository`, `TagRepository`, `UserRepository`, `Clock` を定義済み。
- **MySQL 実装**
  - `backend/internal/infrastructure/mysql` に各リポジトリ実装を追加済み。`backend/db/migrations` に golang-migrate 用の初期スキーマ（users, tags, posts, post_tags）を配置。
  - 初回管理者ユーザ登録用に `backend/cmd/seed` を用意。`SEED_ADMIN_EMAIL` / `SEED_ADMIN_PASSWORD` / `DATABASE_DSN` を指定して実行すると、該当メールが未登録の場合に bcrypt ハッシュで 1 件 INSERT する。
- **テスト**
  - ドメイン層（post）のユニットテストを追加済み。リポジトリ層の統合テストは未実装（任意）。
- **フェーズ 1 完了条件**: ドメイン・リポジトリIF・MySQL 実装が存在し、`go test ./backend/...` および `golangci-lint run ./...` が通ること。

---

## フェーズ 2: バックエンド API 実装（connect-go）

- **サーバブートストラップ（実施済み）**
  - `cmd/server/main.go`: `DATABASE_DSN` で MySQL 接続、`ADMIN_API_KEY` を PostServer/TagServer に渡して登録。`/healthz` とセキュリティヘッダは常時有効。
  - `backend/internal/interface/rpc`: domain→proto 変換（converter.go）、管理者キー認証（auth.go、X-Admin-Key ヘッダ）、Slugify、PostServer・TagServer の全 RPC 実装。
- **サービスごとの実装状況**
  - PostService: ListPosts / GetPost（公開のみ／draft 一覧は要認証）、CreatePost / UpdatePost / DeletePost / SearchPosts / PublishPost 実装済み（作成・更新・削除・公開は X-Admin-Key 必須）。
  - TagService: ListTags / CreateTag / DeleteTag 実装済み（Create/Delete は X-Admin-Key 必須）。
  - AuthService: 実装済み（Login / Logout / GetMe、メモリセッション・bcrypt パスワード検証）。AIService: 実装済み（ダミー）。
- **今後の作業**
  - 管理画面を AuthService（ログイン・セッション）ベースに切り替え（バックエンドは実装済み。フロントは X-Admin-Key と併存可）。
  - **エラーハンドリングの共通化（実施）**: `rpc.MapHandlerError` で repo 由来の err を `CodeInternal`＋汎用メッセージに正規化。AuthServer（Login/GetMe）、PostServer、TagServer の repo 呼び出し失敗時はすべて MapHandlerError 経由に統一済み。
  - サービス層のユニットテスト・ハンドラの E2E テスト。

---

## フェーズ 3: フロントエンド実装（Next.js）

- **基盤（実施済み）**
  - Next.js 16（App Router）を `frontend/` に作成。`buf generate` で TypeScript を `frontend/src/gen` に出力し、`@connectrpc/connect-web` で `postClient` / `tagClient` を初期化。`NEXT_PUBLIC_API_URL` で API ベース URL を指定。
- **閲覧系機能（実施済み）**
  - トップページ（`/`）: `ListPosts` で公開記事一覧を表示。
  - 記事詳細（`/posts/[slug]`）: `GetPost` で slug 指定し、`react-markdown` で本文をレンダリング。
  - タグ一覧（`/tags`）: `ListTags` でタグ一覧表示。タグ別記事一覧（`/tags/[slug]`）: `ListTags` で slug 一致のタグを取得し、`ListPosts(tag_id)` で該当記事を表示。
  - 検索（`/search`）: `SearchPosts` で全文検索。共通ヘッダーに検索リンク・検索フォームを配置。
- **管理画面（実施済み・AuthService ログイン or X-Admin-Key）**
  - `/admin`: 管理者ログイン（メール・パスワードで AuthService.Login）または API キー入力。セッショントークンは sessionStorage に保存し、Bearer で Post/Tag/AI API を呼び出し。バックエンドは X-Admin-Key または有効な Bearer セッションのいずれかで管理者を許可。
  - `/admin/posts`: 記事一覧（下書き/公開フィルタ）。
  - `/admin/posts/new`: 新規記事作成（CreatePost）。
  - `/admin/posts/[id]/edit`: 記事編集、公開/下書き、削除。ログアウトボタンあり（AuthService.Logout + セッション削除）。
  - 共通ヘッダーに「管理」リンクを追加。
- **AI 連携 UI（実施済み・ダミー実装）**
  - バックエンド: `AIService`（Summarize / DraftSupport）を connect-go で実装。現時点では Vertex AI ではなくローカルロジックで要約・提案本文を生成。
  - 管理画面: `/admin/posts/[id]/edit` に「本文から要約を生成」（Summarize）ボタンと、「下書き支援」入力＋提案本文プレビュー（DraftSupport）を追加。提案本文はワンクリックで本文に反映可能。
- **今後の作業**
  - AI 実装の Vertex AI 連携への切り替え（GCP 認証・コスト・レイテンシを考慮）。
- **UX / パフォーマンス**
  - SSG/ISR 可能なページは極力静的生成し、SEO とパフォーマンスを最適化。
  - `use cache` など Next.js 16 のキャッシュ機能を適用（ADR で決定する場合は別途記録）。

---

## フェーズ 4: セキュリティ・パフォーマンスの強化

- **セキュリティ**
  - 管理者認証フローの見直し（セッション固定攻撃対策、CSRF 対策、パスワードハッシュなど）。
  - **Google アカウントでの管理者ログイン（OAuth 2.0 / OIDC）**: メール・パスワードに加え、Google で「このブログの管理者」としてログインできるようにする。実装時は Google OAuth クライアントの作成、バックエンドでの ID トークン検証、既存セッション機構との統合を検討。
  - 入力値のバリデーションとサニタイズ（タイトル、slug、Markdown 等）。
    - 記事 API ではタイトル長・slug 形式・summary/本文長・tag_ids 数/長さのバリデーションを実装済み。
  - ロール・権限チェック（今は単一管理者想定だが、将来の拡張を見据えてインターフェースを設計）。
- **パフォーマンス**
  - DB クエリの N+1 解消、必要なインデックスの追加。
  - Cloudflare 側のキャッシュ戦略（HTML/静的アセット・API レスポンス）を調整。
  - Cloud Run のコールドスタート影響をモニタし、必要に応じて最小インスタンス数・メモリ/CPU を調整。

---

## フェーズ 5: 運用・監視・CI/CD

- **ログ・メトリクス**
  - 構造化ログ（JSON）でリクエスト ID・ユーザー ID・重要イベント（ログイン・記事公開など）を記録。
  - Cloud Logging / Cloud Monitoring を用いてエラーレート・レイテンシ・リソース利用状況を可視化。
- **アラート**
  - 5xx レート、レスポンス時間、DB 接続エラーなどに対するアラートを設定。
- **CI/CD**
  - GitHub Actions で以下を自動化:
    - PR / push: `ci.yml` で Markdown/Proto lint、コード生成、Go テスト、golangci-lint、フロントビルド。
    - デプロイ: `deploy-api.yml` で `main` への push（backend 関連）または手動実行時に Cloud Run へ API をビルド・デプロイ。Cloudflare Pages は Git 連携で `main` に push すると自動ビルドされる想定。
  - インフラ設定手順は [docs/setup-deploy-checklist.md](setup-deploy-checklist.md) に記載。

---

## フェーズ 6: 機能拡張ロードマップ（例）

- **Google アカウントでの管理者ログイン**
  - 管理画面で「Google でログイン」を選べるようにする。OAuth 2.0 / OpenID Connect で Google の ID トークンを取得し、バックエンドで検証したうえで既存のセッション発行フローと統合。許可する Google アカウント（メール）を設定で絞る運用を想定。
- **コメント・いいね機能**
  - 新しい proto / サービス（CommentService, LikeService）を追加し、スパム対策・モデレーションも検討。
- **RSS / OGP**
  - `/rss.xml` の生成、記事ごとの OGP 画像自動生成（AI ベースのサムネイルなど）。
- **マルチテナント対応**
  - Blog エンティティ導入、ドメイン/サブドメインによるテナント切り替え。
- **監査ログ**
  - 管理者操作（ログイン、記事 CRUD、公開/非公開変更）を専用テーブルに保存し、検索 UI を提供。

---

## 実装順序の推奨

1. **フェーズ 0**: 基盤を固め、buf/codegen・Next.js・Go モジュールを最低限動かす。
2. **フェーズ 1〜2**: バックエンドのドメイン・API を優先実装し、最低限の API（記事一覧・詳細）を完成させる。
3. **フェーズ 3**: フロントから記事閲覧フローを通し、「読む」体験を先に完成させる。
4. **フェーズ 3（後半）〜4**: 管理画面・AI 連携・セキュリティ/パフォーマンス強化を進める。
5. **フェーズ 5 以降**: 運用・監視・CI/CD を整えつつ、コメントや RSS などの拡張を追加していく。

各フェーズで、主要な決定やトレードオフがあれば `docs/adr/ADR-00x-*.md` として ADR を追加し、将来の自分が経緯を追えるようにする。
