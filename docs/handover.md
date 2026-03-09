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
  - [インフラ概要](infrastructure.md)

---

## 2. 直近の作業内容（ここまでの流れ）

- **Terraform で GCP を管理**: Cloud SQL (MySQL)、Secret Manager（DATABASE_DSN, ADMIN_API_KEY）、Cloud Run を `terraform/` で定義。Artifact Registry と WIF は手順書どおり手動で済ませ済み。
- **terraform apply の実施**: Cloud SQL インスタンス `blog-mysql` と DB `blog`、Secret Manager のシークレット・バージョンは作成完了。**Cloud Run サービスは一度失敗**（後述の理由で修正済み）。
- **Cloud Run まわりの修正**:
  - **PORT**: Cloud Run v2 では `PORT` は予約のため指定不可。`terraform/cloudrun.tf` から `PORT` の env を削除済み。
  - **イメージアーキテクチャ**: Cloud Run は **linux/amd64** のみ対応。M1/M2 等でビルドすると arm64 になりエラーになるため、**必ず `--platform linux/amd64` でビルド**する必要あり。ルートの `Makefile` で `make docker-api` がそのようにビルドするよう記載済み。
- **CI でマイグレーション実行**: `.github/workflows/deploy-api.yml` に「Run migrations」ステップを追加。Cloud SQL Auth Proxy + golang-migrate で本番 DB に `migrate up` を実行。GitHub Secrets の `MIGRATION_DSN` と、デプロイ用 SA への **Cloud SQL Client** ロールが必要。
- **Makefile の追加**: ルートに `Makefile` を追加。`make docker-api`（ビルド＋push）、`make terraform-*`、`make migrate-up`、`make proto` / `make lint` / `make test` などを用意。デフォルトの `GCP_PROJECT_ID` は `kano-blog-prod`、`REGION` は `asia-northeast1`。
- **計画書の更新**: Google アカウントでの管理者ログインをフェーズ 4・6 に追記済み。

---

## 3. 現在のインフラ・アプリ状態

| 項目 | 状態 | 備考 |
| --- | --- | --- |
| GCP プロジェクト | 利用中 | `kano-blog-prod` |
| Workload Identity / デプロイ用 SA | 済 | `blog-deploy`、Artifact Registry・Cloud Run デプロイ用 |
| Artifact Registry | 済 | `blog-repo`（asia-northeast1） |
| Cloud SQL (MySQL) | Terraform で作成済み | インスタンス `blog-mysql`、DB `blog` |
| Secret Manager | Terraform で作成済み | `DATABASE_DSN`、`ADMIN_API_KEY` |
| Cloud Run サービス | **未作成** | 上記の PORT/amd64 対応後に apply で作成予定 |
| デプロイ用 GitHub Actions | 設定済み | `deploy-api.yml`（マイグレーション＋ビルド＋デプロイ） |
| Cloudflare Pages | 手動想定 | 手順は [setup-deploy-checklist.md セクション 7](setup-deploy-checklist.md#7-cloudflare-pages-の設定手動) |

---

## 4. 設定の所在（機密は書かず「どこに何があるか」のみ）

- **Terraform 変数**: `terraform/terraform.tfvars`（`.gitignore` 済み）。`project_id`、`db_root_password`、`admin_api_key`、`cloud_run_image` など。サンプルは `terraform/terraform.tfvars.example`。
- **ローカル環境変数**: `.envrc`（`.gitignore` 済み）。サンプルは `.envrc.example`。GCP_PROJECT_ID、REGION、GITHUB_ORG など。
- **GitHub Actions**:
  - **Secrets**: `GCP_PROJECT_ID`、`GCP_SA_KEY`、`MIGRATION_DSN` が必須。任意で `GCP_REGION`、`IMAGE_NAME`。
  - **Variables**: ワークフローは現状 `secrets.*` を参照しているため、上記は Secrets に登録する必要あり。
- **GCP IAM**: デプロイ用 SA `blog-deploy` に **Cloud SQL Client**（`roles/cloudsql.client`）を付与すること（CI マイグレーション用）。手順は [setup-deploy-checklist.md 8.3](setup-deploy-checklist.md#83-デプロイ時のマイグレーションci-で実行)。

---

## 5. 次にやること（続きの作業チェックリスト）

1. **Cloud Run 用イメージを linux/amd64 で push する**
   - リポジトリルートで `make docker-api`（要 `GCP_PROJECT_ID` / `REGION` の環境変数または Makefile デフォルト）。
   - これを行わないと Terraform の Cloud Run 作成が「イメージが amd64 でない」と失敗する。

2. **Terraform で Cloud Run を作成する**
   - `cd terraform && terraform apply`
   - 作成されるのは `google_cloud_run_v2_service.blog_api` と `google_cloud_run_v2_service_iam_member.public`。

3. **（初回のみ）migrate ユーザーに DB 権限を付与**
   - [setup-deploy-checklist.md 8.3](docs/setup-deploy-checklist.md#83-デプロイ時のマイグレーションci-で実行) の「マイグレーション用ユーザーに権限を付与」を実行。その後、GitHub Secrets の `MIGRATION_DSN` を `mysql://migrate:パスワード@tcp(127.0.0.1:3306)/blog?parseTime=true`（パスワードは URL エンコード）に設定。

4. **本番 DB にマイグレーションをかける（CI または手動）**
   - CI では `MIGRATION_DSN` 設定で自動実行。手動の場合は Cloud SQL Auth Proxy で `localhost:3306` に接続した状態で、`DATABASE_DSN`（または `MIGRATION_DSN` 相当）を設定して `make migrate-up`。パスワードは `terraform.tfvars` の `db_root_password` と同じ（migrate ユーザーも同じパスワード）。

5. **Cloud Run の URL を控える**
   - `terraform output cloud_run_url` を実行し、フロントの本番環境変数 `NEXT_PUBLIC_API_URL` に設定する（Cloudflare Pages の環境変数など）。

6. **Cloudflare Pages の設定（未実施なら）**
   - [setup-deploy-checklist.md セクション 7](setup-deploy-checklist.md#7-cloudflare-pages-の設定手動) の手順で、リポジトリ連携・ルート `frontend`・ビルドコマンド・`NEXT_PUBLIC_API_URL` を設定。

7. **（任意）管理ユーザーの seed**
   - ローカルで `go run ./backend/cmd/seed` を実行し、管理画面用のメール/パスワードを 1 件登録。

8. **動作確認**
   - Cloud Run の `/healthz`、フロントの表示、管理画面ログイン（メール/パスワードまたは API キー）を確認。

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
- **GitHub Secrets**: `MIGRATION_DSN` は **Secrets** に登録する（Variables だとワークフローから参照されない＋機密のため）。値は **`migrate` ユーザー**で `mysql://migrate:パスワード@tcp(127.0.0.1:3306)/blog?parseTime=true` 形式。パスワードは `db_root_password` と同じ。特殊文字は URL エンコード（例: `+` → `%2B`）。初回のみ [setup-deploy-checklist.md 8.3](docs/setup-deploy-checklist.md#83-デプロイ時のマイグレーションci-で実行) の「migrate に権限付与」を実行すること。
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

以上を把握しておけば、続きの作業（apply → マイグレーション → Cloudflare → 動作確認）をスムーズに進められます。
