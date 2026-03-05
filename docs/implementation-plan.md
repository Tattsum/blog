# 個人ブログ実装プラン

このドキュメントは、既存のアーキテクチャ設計・API 仕様・ADR を前提に、実際のアプリケーション実装を段階的に進めるための実行プランをまとめたものです。

- **対象リポジトリ**: `blog`
- **前提ドキュメント**:
  - アーキテクチャ: `docs/architecture.md`
  - API 仕様: `docs/api-specification.md`
  - ADR: `docs/adr/001-connect-rpc.md`, `docs/adr/002-hosting-cloudflare-and-cloudrun.md`
- **設計方針**: Go のドメイン駆動設計（DDD）を意識し、可能な範囲でテスト駆動開発（TDD）を行う。セキュリティとパフォーマンスを常に考慮する。

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
- **CI の最初の土台（任意）**
  - GitHub Actions で `markdownlint`, `buf lint`, `go test` を最低限回すワークフローを追加（詳細は後続フェーズで拡張）。

---

## フェーズ 1: ドメインモデリングとリポジトリ層

- **ドメインモデル定義（Go）**
  - `backend/internal/domain/post`, `domain/tag`, `domain/user` を定義済み。Post/Tag/User エンティティ、Slug/Email 値オブジェクト、Post の公開判定（IsPublished / Publish / Unpublish）を実装。
- **リポジトリインターフェース**
  - `backend/internal/domain/repository` に `PostRepository`, `TagRepository`, `UserRepository`, `Clock` を定義済み。
- **MySQL 実装**
  - `backend/internal/infrastructure/mysql` に各リポジトリ実装を追加済み。`backend/db/migrations` に golang-migrate 用の初期スキーマ（users, tags, posts, post_tags）を配置。
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
  - AuthService / AIService: 未実装。
- **今後の作業**
  - セッション／ログインベースの認証（AuthService）への移行、エラーハンドリングの共通化、入力バリデーション強化。
  - サービス層のユニットテスト・ハンドラの E2E テスト。

---

## フェーズ 3: フロントエンド実装（Next.js）

- **基盤（実施済み）**
  - Next.js 16（App Router）を `frontend/` に作成。`buf generate` で TypeScript を `frontend/src/gen` に出力し、`@connectrpc/connect-web` で `postClient` / `tagClient` を初期化。`NEXT_PUBLIC_API_URL` で API ベース URL を指定。
- **閲覧系機能（実施済み）**
  - トップページ（`/`）: `ListPosts` で公開記事一覧を表示。
  - 記事詳細（`/posts/[slug]`）: `GetPost` で slug 指定し、`react-markdown` で本文をレンダリング。
  - タグ一覧（`/tags`）: `ListTags` でタグ一覧表示。タグ別記事一覧（`/tags/[slug]`）はプレースホルダのみ（将来拡張）。
  - 検索（`/search`）: `SearchPosts` で全文検索。共通ヘッダーに検索リンク・検索フォームを配置。
- **今後の作業**
  - 管理画面（ログイン・記事 CRUD・公開操作）、AI 連携 UI。
- **管理画面**
  - ログインフォーム（AuthService.Login）とセッション管理（Cookie or localStorage + HTTP-only Cookie）。
  - 記事一覧（下書き/公開のフィルタ）、作成・編集画面（Markdown エディタ）。
  - 公開/非公開切り替え（PublishPost）、削除（DeletePost）。
- **AI 連携 UI**
  - 要約生成ボタン（Summarize）: 現在の本文から要約を生成し、summary 欄に反映。
  - 下書き支援（DraftSupport）: プロンプトと本文を送信し、提案本文を差分表示して選択可能に。
- **UX / パフォーマンス**
  - SSG/ISR 可能なページは極力静的生成し、SEO とパフォーマンスを最適化。
  - `use cache` など Next.js 16 のキャッシュ機能を適用（ADR で決定する場合は別途記録）。

---

## フェーズ 4: セキュリティ・パフォーマンスの強化

- **セキュリティ**
  - 管理者認証フローの見直し（セッション固定攻撃対策、CSRF 対策、パスワードハッシュなど）。
  - 入力値のバリデーションとサニタイズ（タイトル、slug、Markdown 等）。
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
    - PR 時: `npm run lint:md`, `buf lint`, `go test ./...`、（将来）`frontend` テスト。
    - main マージ時: Cloud Run と Cloudflare Pages へのデプロイ（手動承認ステップを含めてもよい）。

---

## フェーズ 6: 機能拡張ロードマップ（例）

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
