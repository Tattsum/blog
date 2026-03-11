# 引き継ぎ・続きの作業用ドキュメント

新しいエージェントや担当者が、このリポジトリの「いまの状態」と「次にやること」を把握し、スムーズに作業を続けられるようにするためのドキュメントです。

**更新目安**: 大きなマイルストーンやインフラ変更のたびに更新する。

---

## 1. プロジェクト概要

- **リポジトリ**: blog（モノレポ）
- **構成**: フロント（Next.js / `frontend/`）、API（Go + connect-go / `backend/`）、Proto 定義（`proto/`）、インフラ（Terraform / `terraform/`）
- **本番想定**: API → GCP Cloud Run、フロント → Cloudflare Pages、DB → Cloud SQL (MySQL)
- **関連ドキュメント**:
  - [実装プラン・進捗](implementation-plan.md)
  - [セットアップ・デプロイ手順（詳細）](setup-deploy-checklist.md)
  - [フロントエンドデザイン方針（Zenn 風・ダークモード）](frontend-design.md)
  - [AI モデル・プロバイダ拡張（Claude / DeepSeek / OpenAI 等・設計メモ）](ai-model-providers.md)

---

## 2. 直近の作業内容（ここまでの流れ）

- **Terraform で GCP を管理**: Cloud SQL (MySQL)、Secret Manager（DATABASE_DSN, ADMIN_API_KEY）、Cloud Run を `terraform/` で定義。Artifact Registry と WIF は手順書どおり手動で済ませ済み。
- **terraform apply の実施**: Cloud SQL インスタンス `blog-mysql` と DB `blog`、Secret Manager、**Cloud Run サービス**まで作成完了。Cloud Run は一度 PORT/amd64 で失敗したが、起動順序修正と Dockerfile の proto 生成追加で解消済み。
- **Cloud Run まわりの修正**:
  - **サービス名**: 本番 API は **`blog-backend`** にデプロイする（`blog-api` 名で 404 になるケースを避けるため手順書・CI・Terraform 既定を揃え済み）。
  - **PORT**: Cloud Run v2 では `PORT` は **env に渡さない**（予約）。コンテナは **8080 で待ち受け**、`gcloud run deploy` には **`--port=8080`** を付ける。`terraform/cloudrun.tf` から `PORT` の env は削除済み。
  - **起動順序**: `backend/cmd/server/main.go` で DB 接続より先に `ListenAndServe` を開始するよう変更（Cloud Run の起動タイムアウト対策）。
  - **イメージアーキテクチャ**: Cloud Run は **linux/amd64** のみ。`make docker-api` が `--platform linux/amd64` でビルドするよう記載済み。
- **マイグレーション・CI**:
  - Terraform で **migrate** ユーザー（host `%`）を追加。root は Proxy 経由で Access denied になるため、CI では `migrate` を使用。
  - GitHub Secrets は **`MIGRATION_PASSWORD`**（推奨・tfvars の `db_root_password` をそのまま。ワークフロー側で URL エンコード）または `MIGRATION_DSN`。初回のみ DB に `GRANT ... ON blog.* TO 'migrate'@'%'` を実行する必要あり（[setup-deploy-checklist.md 8.3](docs/setup-deploy-checklist.md#83-デプロイ時のマイグレーションci-で実行)）。
  - マイグレーション `000001`: `summary` を MySQL 8 対応のため `TEXT` → `VARCHAR(2000)` に変更済み。dirty 状態の解消は `DROP TABLE schema_migrations` 等でリセットしてから再実行。
- **Docker ビルド**: `gen/` は .gitignore のため CI に無い。`backend/Dockerfile` で buf + protoc-gen-go/connect-go により **ビルド時に proto から Go コードを生成**。`buf.gen.go.yaml`（Go 専用）をリポジトリに追加済み。
- **Edge fetch redirect エラー（Cloudflare Workers）**:
  - Workers では `fetch(..., { redirect: "error" })` が使えず Invalid redirect となる。対策として **グローバル fetch のラップ**（`instrumentation-edge.ts`）に加え、**Connect のトランスポートに `edgeSafeFetch` を渡す**（`frontend/src/lib/edge-safe-fetch.ts`）二重対策を入れ済み。詳細は [frontend-design.md](frontend-design.md) 参照。
- **フロント（Cloudflare Pages）**:
  - `.gitignore` を `gen/` → `/gen/` に変更し、**frontend/src/gen/**（proto から生成した TypeScript）をリポジトリにコミット済み。Cloudflare の clone に含まれるためビルドが通る。
  - Cloudflare のビルド設定: ルート `frontend`。**デプロイコマンドは必ず `npm run deploy`**（OpenNext の build で `.open-next` を生成してから deploy）。`npx wrangler deploy` のみだと `.open-next/worker.js was not found` で失敗する。
  - **`frontend/package.json` の `name` を `blog` に変更済み**。OpenNext が `WORKER_SELF_REFERENCE` のサービス名に package name を使うため、`frontend` のままだと「Worker 'frontend' が存在しない」でデプロイ失敗する問題を解消。
- **Makefile**: ルートに `Makefile` を追加。`make docker-api`、`make terraform-*`、`make migrate-up`、`make proto` / `make lint` / `make test` など。デフォルトの `GCP_PROJECT_ID` は `kano-blog-prod`、`REGION` は `asia-northeast1`。

---

## 3. 現在のインフラ・アプリ状態

| 項目 | 状態 | 備考 |
| --- | --- | --- |
| GCP プロジェクト | 利用中 | `kano-blog-prod` |
| Workload Identity / デプロイ用 SA | 済 | `blog-deploy`、Artifact Registry・Cloud Run デプロイ用 |
| Artifact Registry | 済 | `blog-repo`（asia-northeast1） |
| Cloud SQL (MySQL) | Terraform で作成済み | インスタンス `blog-mysql`、DB `blog`。migrate ユーザー（host %）も Terraform で作成。 |
| Secret Manager | Terraform で作成済み | `DATABASE_DSN`、`ADMIN_API_KEY` |
| Cloud Run サービス | **作成済み** | サービス名 **`blog-backend`**（`gcloud run deploy blog-backend ...`）。デプロイ後の **Service URL** は regional 形式（例: `https://blog-backend-1098008862560.asia-northeast1.run.app`）。Terraform では `cloud_run_service_name` 既定を `blog-backend` に合わせ済み。URL は `terraform output cloud_run_url` または `gcloud run services describe blog-backend --format='value(status.url)'` で取得。 |
| デプロイ用 GitHub Actions | 設定済み | `deploy-api.yml`（MIGRATION_PASSWORD → migrate → ビルド・push → **`gcloud run deploy`** でイメージのみ更新）。Cloud Run の Secret・Cloud SQL・env は Terraform（`cloudrun.tf`）が管理するため CI では指定しない。サービス名は Secret **`CLOUD_RUN_SERVICE_NAME`**（未設定時 **`blog-backend`**）。`terraform.tfvars` の `cloud_run_service_name` と一致させる。 |
| Cloudflare Pages / Workers | **設定・動作確認済み** | ルート `frontend`。Deploy command は `npm run deploy`。**`NEXT_PUBLIC_API_URL`** はリポジトリの **`frontend/.env.production`** と **`frontend/wrangler.jsonc` の `env.production.vars`** で管理（本番は regional URL）。Cloudflare ダッシュボードの「変数とシークレット」には **設定しない**（リポジトリを正とする）。生存確認は **`/health`**（Cloud Run では `/healthz` は予約のため使用不可）。 |

---

## 4. 設定の所在（機密は書かず「どこに何があるか」のみ）

- **Terraform 変数**: `terraform/terraform.tfvars`（`.gitignore` 済み）。`project_id`、`db_root_password`、`admin_api_key`、`cloud_run_image` など。サンプルは `terraform/terraform.tfvars.example`。
- **ローカル環境変数**: `.envrc`（`.gitignore` 済み）。サンプルは `.envrc.example`。GCP_PROJECT_ID、REGION、GITHUB_ORG など。
- **GitHub Actions**:
  - **Secrets**: `GCP_PROJECT_ID`、`GCP_SA_KEY` が必須。マイグレーション用に `MIGRATION_PASSWORD`（推奨）または `MIGRATION_DSN`。任意で `GCP_REGION`、`IMAGE_NAME`。
  - **Variables**: ワークフローは現状 `secrets.*` を参照しているため、上記は Secrets に登録する必要あり。
- **GCP IAM**: デプロイ用 SA `blog-deploy` に **Cloud SQL Client**（`roles/cloudsql.client`）を付与すること（CI マイグレーション用）。手順は [setup-deploy-checklist.md 8.3](setup-deploy-checklist.md#83-デプロイ時のマイグレーションci-で実行)。

---

## 5. 次にやること（続きの作業チェックリスト）

**次のエージェント向け**: Cloudflare 本番まわり（デプロイ・環境変数・動作確認）は対応済み。残りは任意項目と実装プラン上のコードタスクが中心です。

### 済（手動対応完了）

- ~~Cloudflare Pages のデプロイ確認~~ … Deploy command を `npm run deploy` にし、デプロイ成功まで確認済み。
- ~~Cloudflare の環境変数~~ … **`NEXT_PUBLIC_API_URL`** は **リポジトリ**（`.env.production` と `wrangler.jsonc`）で管理。ダッシュボードには設定しない（デプロイのたびに変えずに済む）。
- ~~動作確認~~ … **`/health`** で 200、フロントで API 動作、管理画面ログイン／API キーまで確認済み。

### 残り

- **（任意）管理ユーザーの seed**  
  - ローカルで Cloud SQL Auth Proxy を起動したうえで、`DATABASE_DSN='mysql://migrate:パスワード@tcp(127.0.0.1:3306)/blog?parseTime=true'` と `SEED_ADMIN_EMAIL` / `SEED_ADMIN_PASSWORD` を設定し、`go run ./backend/cmd/seed` を実行。管理画面用の初回ユーザーを 1 件登録できる。

**参考（すでに実施済み）**: Cloud Run 用イメージ push、Terraform apply（blog-backend を import して state 統一）、migrate 権限付与、MIGRATION_PASSWORD 設定、CI マイグレーション成功、Docker ビルド時の proto 生成、frontend/src/gen のコミット、package.json name の修正、**Cloudflare デプロイ・NEXT_PUBLIC_API_URL（リポジトリ管理）・本番動作確認（2026-03 頃）**。

---

## 6. 主要ファイル・ディレクトリ

| パス | 役割 |
| --- | --- |
| `Makefile` | ビルド・push・terraform・migrate・lint などのエントリポイント。Cloud Run 用は `make docker-api`。 |
| `terraform/*.tf` | GCP リソース定義（Cloud SQL, Secret Manager, Cloud Run）。`terraform.tfvars` は gitignore。 |
| `terraform/.terraform-version` | tfenv 用（1.6.6）。 |
| `.github/workflows/deploy-api.yml` | デプロイ用ワークフロー。マイグレーション → ビルド・push → Cloud Run デプロイ。 |
| `.github/workflows/ci.yml` | PR/push 時の lint・test・ビルド。 |
| `backend/cmd/server/main.go` | API エントリポイント。`PORT` は env 未設定時 8080。 |
| `backend/db/migrations/` | golang-migrate 用の SQL。 |
| `docs/setup-deploy-checklist.md` | セットアップ・デプロイの詳細手順（GCP・Cloudflare・Secrets・マイグレーション CI 含む）。 |
| `docs/implementation-plan.md` | 実装フェーズと進捗。 |

---

## 7. 注意事項（つまずきやすい点）

- **Cloud Run のイメージ**: 必ず **linux/amd64** でビルドする。ローカルが arm64 の場合は `make docker-api`（内部で `--platform linux/amd64` を付与）を使う。
- **Cloud Run の PORT**: Terraform で `PORT` を env に指定してはいけない。Cloud Run が自動設定する。
- **Terraform apply の順序**: イメージを先に push してから `terraform apply` する。未 push や arm64 イメージのまま apply するとエラーになる。
- **GitHub Secrets**: マイグレーション用は **`MIGRATION_PASSWORD`**（推奨・`db_root_password` をそのまま、URL エンコード不要）または **`MIGRATION_DSN`** を Secrets に登録。初回のみ [setup-deploy-checklist.md 8.3](docs/setup-deploy-checklist.md#83-デプロイ時のマイグレーションci-で実行) の「migrate に権限付与」を実行すること。
- **Cloudflare Pages / OpenNext**: `frontend/package.json` の **`name` は `blog` にすること**。`frontend` のままだと wrangler が生成する `WORKER_SELF_REFERENCE` が存在しない Worker 名を参照し、デプロイが失敗する。
- **Cloudflare の `NEXT_PUBLIC_API_URL`**: **ダッシュボードの「変数とシークレット」には設定しない**。リポジトリの `frontend/.env.production` と `frontend/wrangler.jsonc` の `env.production.vars` を正とし、URL 変更時はここだけ更新して push すればよい。
- **記事の公開範囲**: 記事はログイン不要で URL を知っていれば閲覧可能（公開記事のみ）。投稿・編集・削除は管理者のみ（計画書に Google ログインの拡張を追記済み）。

---

## 8. コマンド早見（よく使うもの）

```bash
# イメージをビルドして push（apply の前に実行）
make docker-api

# Terraform
make terraform-plan
make terraform-apply

# 本番 DB マイグレーション（Cloud SQL Proxy 起動後、DATABASE_DSN を export してから）
export DATABASE_DSN='mysql://root:パスワード@tcp(127.0.0.1:3306)/blog?parseTime=true'
make migrate-up

# コード生成・lint・テスト
make proto
make lint
make test
```

以上を把握しておけば、続きの作業（seed やフェーズ 5 のログ／監視など）をスムーズに進められます。

---

## 9. レビュー・指摘メモ

- 直近の PR やコードレビューで指摘された内容があれば、ここに要約を追記すること。
- Cloudflare のデプロイ・`NEXT_PUBLIC_API_URL`（リポジトリの .env.production / wrangler.jsonc で管理・ダッシュボードには設定しない）・本番動作確認は対応済み。次は **（任意）seed** や **実装プランのフェーズ 5（ログ・監視）** などを優先してよい。
