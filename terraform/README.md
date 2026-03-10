# Terraform: ブログ GCP インフラ（Cloud SQL / Secret Manager / Cloud Run）

[セットアップ・デプロイ やること一覧](../docs/setup-deploy-checklist.md) の **4. Cloud SQL**・**5. Secret Manager**・**6. Cloud Run** を Terraform で管理するための設定です。

## 前提

- **1〜3**（GCP プロジェクト・API 有効化・WIF・Artifact Registry）は手順書どおり手動で完了していること
- 初回 `apply` の前に、**コンテナイメージを Artifact Registry に push 済み**であること

## 使い方

### 1. 変数ファイルの準備

```bash
cp terraform.tfvars.example terraform.tfvars
# 編集: project_id, admin_api_key, cloud_run_image を必須で設定
# 任意: db_root_password（未設定ならランダム生成。本番では変数で渡すことを推奨）
```

### 2. 初回: イメージの push

**重要**: Cloud Run は **linux/amd64** のイメージのみ対応しています。M1/M2 Mac などでは `--platform linux/amd64` を付けてビルドしてください。

リポジトリルートで（Makefile 使用例）:

```bash
export GCP_PROJECT_ID=your-project-id
export REGION=asia-northeast1
make docker-api
```

または手動で:

```bash
export GCP_PROJECT_ID=your-project-id
export REGION=asia-northeast1
docker build --platform linux/amd64 -t ${REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/blog-repo/blog-api:latest -f backend/Dockerfile .
gcloud auth configure-docker ${REGION}-docker.pkg.dev --quiet
docker push ${REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/blog-repo/blog-api:latest
```

### 3. Terraform の実行

```bash
cd terraform
terraform init
terraform plan   # 変更内容を確認
terraform apply  # 実行（Cloud SQL 作成は数分かかります）
```

### 4. マイグレーション

Cloud SQL 作成後、ルートパスワードは次のいずれかで確認します。

- `terraform.tfvars` で `db_root_password` を設定した場合: その値
- 未設定でランダム生成した場合: `terraform output -raw db_root_password`

[Cloud SQL Auth Proxy](https://cloud.google.com/sql/docs/mysql/connect-auth-proxy) で接続し、リポジトリルートで:

```bash
migrate -path backend/db/migrations \
  -database "mysql://root:PASSWORD@tcp(localhost:3306)/blog?parseTime=true" \
  up
```

### 5. 出力の利用

- **Cloud Run URL**: `terraform output cloud_run_url` は **短い URL**（`...-an.a.run.app`）を返す。本番の `NEXT_PUBLIC_API_URL` には **regional URL**（`https://SERVICE-PROJECT_NUMBER.REGION.run.app`）を **`frontend/.env.production`** と **`frontend/wrangler.jsonc`** に直接記載する（`/health` が通る方）。Cloudflare ダッシュボードの変数には設定しない。
- **接続名**: `terraform output cloud_sql_connection_name` → 手動接続時などに参照

## リソース一覧

| リソース | 説明 |
| --- | --- |
| Cloud SQL (MySQL 8.0) | インスタンス `blog-mysql` + データベース `blog` |
| Secret Manager | `DATABASE_DSN`（Unix ソケット用 DSN）、`ADMIN_API_KEY` |
| Cloud Run v2 | サービス名は `cloud_run_service_name`（既定 `blog-backend`）。シークレット・Cloud SQL 接続付き |

## 注意

- `terraform.tfvars` に機密を書く場合は、`.gitignore` に追加してください。
- 本番では `db_root_password` を変数で渡し、state に平文が残らないようにすることを推奨します。
- Cloud Run のイメージを更新する場合は、CI で push したうえで `terraform apply` で `cloud_run_image` を新しいタグに変更するか、CI から `gcloud run deploy` で上書きする運用でも構いません。
