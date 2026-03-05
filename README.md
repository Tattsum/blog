# blog

個人ブログシステムのモノレポ。管理者が記事を執筆・公開し、読者が一覧・検索・閲覧できる。Markdown と Vertex AI (Gemini) による要約・下書き支援を提供する。

---

## 1. プロジェクト概要

- **リポジトリ名**: blog
- **対象**: モノレポ（フロントエンド + バックエンド + Proto 定義）
- **概要**: 記事の CRUD・タグ管理・全文検索・管理者認証・AI 要約・下書き支援を備えた個人ブログ。フロントは Next.js、API は connect-go、インフラは GCP Cloud Run / Cloudflare Pages / Cloud SQL (MySQL) / Vertex AI を想定している。

---

## 2. アーキテクチャ（簡潔に）

- **フロント**: Next.js (TypeScript) → Cloudflare Pages で配信
- **API**: Go + connect-go → Cloud Run で稼働
- **通信**: Connect RPC（HTTP/JSON 互換）。proto から型安全なクライアント・サーバーを生成
- **永続化**: Cloud SQL (MySQL)
- **AI**: Vertex AI (Gemini) で要約・下書き支援（バックエンド経由のみ）

詳細は [docs/architecture.md](docs/architecture.md)、API 仕様は [docs/api-specification.md](docs/api-specification.md)、実装フェーズは [docs/implementation-plan.md](docs/implementation-plan.md) を参照。

---

## 3. 前提条件

いずれも **2026年3月時点の最新版** を想定している。

| ツール | 想定バージョン | 用途 |
| --- | --- | --- |
| Go | 1.26+ | バックエンド・コード生成 |
| Node.js | 24+ (LTS) | フロントエンド・ビルド |
| pnpm | 10+ | フロントエンドのパッケージ管理 |
| Docker | 29+ | ローカル DB・Cloud Run ビルド |
| buf | 1.66+ | Proto の lint・コード生成 |
| golangci-lint | 最新 | Go の lint（`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`） |
| MySQL | 8.4 (LTS) | ローカル開発用（Docker 可） |

---

## 4. ローカル環境のセットアップ手順

### 4.1 リポジトリのクローン

```bash
git clone https://github.com/Tattsum/blog.git
cd blog
```

### 4.2 依存ツールのインストール

```bash
# Go プラグイン（proto から Go / Connect 生成）
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest

# buf（推奨）
go install github.com/bufbuild/buf/cmd/buf@latest

# Go の lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 4.3 Proto のコード生成

```bash
# リポジトリルートで（Go: gen/、TypeScript: frontend/src/gen）
npm run generate:proto
```

または `PATH="$(pwd)/node_modules/.bin:$PATH" buf generate`。`buf.gen.yaml` で Go と protoc-gen-es（TS）の出力先を指定している。

### 4.4 バックエンドの起動

```bash
cd backend
go mod download
go run ./cmd/server
```

- ローカルでは環境変数で DB 接続先・Vertex AI 等を指定する（セクション 5 参照）。
- MySQL は Docker で立てる場合の例:

```bash
docker run -d --name blog-mysql -e MYSQL_ROOT_PASSWORD=local -e MYSQL_DATABASE=blog -p 3306:3306 mysql:8.4
```

- 初回は DB マイグレーションを実行する（[golang-migrate](https://github.com/golang-migrate/migrate) をインストール後）:

```bash
migrate -path backend/db/migrations -database "mysql://root:local@tcp(localhost:3306)/blog" up
```

### 4.5 フロントエンドの起動

```bash
cd frontend
npm install
npm run dev
```

- ブラウザで `http://localhost:3000` を開く。
- API のベース URL は環境変数 `NEXT_PUBLIC_API_URL` で指定（例: `http://localhost:8080`）。

---

## 5. 環境変数一覧

| 変数名 | 説明 | デフォルト | 必須 |
| --- | --- | --- | --- |
| **バックエンド** | | | |
| `DATABASE_DSN` | MySQL 接続文字列 | — | 必須 |
| `ADMIN_API_KEY` | 管理者 API キー（X-Admin-Key ヘッダで照合、記事・タグの作成・更新・削除・公開に必要） | — | 管理操作時 |
| `VERTEX_AI_PROJECT` | GCP プロジェクト ID（Vertex AI） | — | AI 利用時 |
| `VERTEX_AI_LOCATION` | Vertex AI リージョン | `us-central1` | 任意 |
| `SESSION_SECRET` | セッション署名用シークレット | — | 管理者認証時 |
| `PORT` | サーバー待ち受けポート | `8080` | 任意 |
| **フロントエンド** | | | |
| `NEXT_PUBLIC_API_URL` | Connect API のベース URL | `http://localhost:8080` | 必須 |

本番では `DATABASE_DSN` や `SESSION_SECRET` 等は GCP Secret Manager 等で注入する想定。

---

## 6. 開発時のよく使うコマンド

```bash
# Markdown の lint（ルートで実行）
npm run lint:md

# Proto の lint とコード生成
npm run lint:proto
buf generate

# Go の lint（要: golangci-lint インストール）
npm run lint:go

# バックエンドのテスト
cd backend && go test ./...

# フロントエンドの開発サーバー
cd frontend && npm run dev

# フロントエンドのビルド（本番用）
cd frontend && npm run build

# バックエンドのビルド（Docker 用）
docker build -t blog-api -f backend/Dockerfile .
```

---

## 7. デプロイ手順

### 7.1 バックエンド（Cloud Run）

- リポジトリルートまたは `backend/` に `Dockerfile` を置き、Cloud Build または GitHub Actions でビルドする想定。

```bash
# 例: Cloud Build でイメージビルド・Cloud Run へデプロイ
gcloud builds submit --tag gcr.io/PROJECT_ID/blog-api
gcloud run deploy blog-api --image gcr.io/PROJECT_ID/blog-api --platform managed --region REGION
```

- Cloud SQL への接続は VPC コネクタ＋プライベート IP、または Cloud SQL Auth Proxy を利用。
- 環境変数・シークレットは Cloud Run の「変数とシークレット」または Secret Manager 連携で設定。

### 7.2 フロントエンド（Cloudflare Pages）

- Next.js のビルド成果物を Cloudflare Pages にデプロイする。

```bash
# 例: Wrangler または Cloudflare ダッシュボードから
cd frontend
npm run build
# 出力ディレクトリ（例: .next または out）を Cloudflare Pages にアップロード
```

- 本番の API ベース URL を `NEXT_PUBLIC_API_URL` に設定し、ビルド時に埋め込む。

---

## 8. ディレクトリ構成

```text
blog/
├── .vscode/           # エディタ設定・推奨拡張
├── docs/              # 設計・API 仕様・ADR・実装プラン
│   ├── adr/
│   ├── architecture.md
│   ├── api-specification.md
│   └── implementation-plan.md
├── proto/             # Protocol Buffers 定義
│   └── blog/v1/
│       ├── post.proto
│       ├── tag.proto
│       ├── auth.proto
│       └── ai.proto
├── backend/           # Go API（connect-go）
│   ├── cmd/server/
│   ├── db/migrations/ # DB マイグレーション（golang-migrate）
│   └── internal/
│       ├── domain/    # ドメイン層（post, tag, user, repository IF）
│       ├── infrastructure/mysql/  # リポジトリ実装
│       └── interface/rpc/  # Connect RPC ハンドラ（PostService, TagService）
├── frontend/          # Next.js アプリ（App Router）
│   ├── src/
│   │   ├── app/       # ルート: /, /posts/[slug], /tags, /tags/[slug], /search
│   │   ├── gen/       # proto から生成した TypeScript（buf generate）
│   │   └── lib/       # API クライアント（Connect-Web）
│   └── package.json
├── .golangci.yaml      # Go の lint 設定（golangci-lint）
├── .markdownlint.json  # markdownlint ルール
├── .markdownlint-cli2.jsonc # markdownlint-cli2 の glob 設定
├── buf.gen.yaml        # buf コード生成設定
├── buf.yaml           # buf 設定
├── go.mod             # Go モジュール（monorepo 全体）
├── package.json       # ルート（Markdown lint 等）
└── README.md
```

`backend/` はドメイン層・API ハンドラまで実装済み。`frontend/` は閲覧系（トップ・記事詳細・タグ一覧）を実装済み。

---

## 9. コントリビューションガイド（個人開発・将来用）

- **ブランチ**: 機能は `feature/xxx`、修正は `fix/xxx` を推奨。`main` は常にデプロイ可能な状態を保つ。
- **コミット**: 意図が分かるメッセージにする（例: `feat(post): add ListPosts RPC`）。
- **Proto 変更**: `buf lint` と `buf generate` を実行し、生成コードをコミットするかどうかはポリシーに従う。
- **ドキュメント**: 設計変更時は [docs/architecture.md](docs/architecture.md)、API 変更時は [docs/api-specification.md](docs/api-specification.md) を更新する。
- **セキュリティ**: 認証情報・シークレットはコミットしない。ローカル用は `.env.example` のみを置き、`.env` は .gitignore する。

---

## 注意

- コマンドはすべてコードブロックで記載している。
- 初見でも迷わないよう、手順はステップ単位で分けてある。実際のディレクトリやファイルが未作成の場合は、上記を参考に作成すること。
