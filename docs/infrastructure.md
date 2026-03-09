# インフラ設定ガイド

本ドキュメントは、個人ブログを **GCP Cloud Run**（API）と **Cloudflare Pages**（フロント）で運用するためのインフラ設定手順をまとめたものです。ADR-002 に基づき、API は GCP、フロントは Cloudflare に配置します。

**詳細な手順（やること一覧・2026年3月時点のやり方）** は [セットアップ・デプロイ やること一覧（詳細手順）](setup-deploy-checklist.md) を参照してください。

---

## 1. 前提条件

| 項目 | 内容 |
| --- | --- |
| GCP | アカウント、課金の有効なプロジェクト |
| Cloudflare | アカウント（Pages 利用可能） |
| ツール | `gcloud` CLI、Docker（ローカルビルド時）、Node.js（フロントビルド） |
| リポジトリ | 本番デプロイ用の GitHub リポジトリ（main ブランチ） |

---

## 2. GCP 側の設定

### 2.1 プロジェクトと API の有効化

```bash
# プロジェクト ID を設定
export GCP_PROJECT_ID=your-project-id
gcloud config set project $GCP_PROJECT_ID

# 必要な API を有効化
gcloud services enable run.googleapis.com
gcloud services enable sqladmin.googleapis.com
gcloud services enable secretmanager.googleapis.com
gcloud services enable artifactregistry.googleapis.com
# Cloud Build でビルドする場合
gcloud services enable cloudbuild.googleapis.com
```

### 2.2 Artifact Registry リポジトリ（コンテナイメージ用）

```bash
export REGION=asia-northeast1
gcloud artifacts repositories create blog-repo \
  --repository-format=docker \
  --location=$REGION \
  --description="Blog API container images"
```

### 2.3 Cloud SQL (MySQL) インスタンス

- **コンソール**: [Cloud SQL](https://console.cloud.google.com/sql) で「インスタンスを作成」→ MySQL 8.4 を選択。
- **推奨**: 開発時は「開発」プリセット、本番は「本番」でパブリック IP またはプライベート IP を選択。
- **ネットワーク**: Cloud Run から接続する場合は VPC ピアリングまたはプライベート IP + VPC コネクタを検討。
- **データベース**: インスタンス作成後に `blog` データベースを作成し、[golang-migrate](https://github.com/golang-migrate/migrate) で `backend/db/migrations` を適用。

```bash
# 例: 接続名でマイグレーション（Cloud SQL Auth Proxy 使用時）
migrate -path backend/db/migrations -database "mysql://USER:PASS@tcp(localhost:3306)/blog" up
```

### 2.4 Secret Manager（シークレット）

Cloud Run に渡す機密情報は Secret Manager に格納し、Cloud Run の「シークレット」で参照します。

```bash
# 例: データベース DSN
echo -n "mysql://user:password@/blog?parseTime=true" | \
  gcloud secrets create DATABASE_DSN --data-file=- --project=$GCP_PROJECT_ID

# 管理者 API キー（X-Admin-Key 用）
echo -n "your-admin-api-key" | \
  gcloud secrets create ADMIN_API_KEY --data-file=- --project=$GCP_PROJECT_ID
```

- 本番では `DATABASE_DSN`、`ADMIN_API_KEY` を必ず設定。Cloud SQL の接続文字列はインスタンスの接続名に合わせて記載。

### 2.5 Cloud Run へのデプロイ

- **ビルド**: リポジトリルートで `backend/Dockerfile` をビルドするため、コンテキストはルートにすること。

```bash
# リポジトリルートで実行
docker build -t $REGION-docker.pkg.dev/$GCP_PROJECT_ID/blog-repo/blog-api:latest -f backend/Dockerfile .
docker push $REGION-docker.pkg.dev/$GCP_PROJECT_ID/blog-repo/blog-api:latest
```

- **デプロイ**:

```bash
gcloud run deploy blog-api \
  --image=$REGION-docker.pkg.dev/$GCP_PROJECT_ID/blog-repo/blog-api:latest \
  --platform=managed \
  --region=$REGION \
  --allow-unauthenticated \
  --set-secrets=DATABASE_DSN=DATABASE_DSN:latest,ADMIN_API_KEY=ADMIN_API_KEY:latest \
  --set-env-vars=PORT=8080
```

- Cloud SQL にプライベート接続する場合は `--vpc-connector` と `--add-cloudsql-instances` のいずれか（または VPC コネクタ＋プライベート IP）を設定。
- デプロイ後に表示される **サービス URL**（例: `https://blog-api-xxxxx-an.a.run.app`）を控え、フロントの `NEXT_PUBLIC_API_URL` に設定する。

---

## 3. Cloudflare 側の設定

### 3.1 Cloudflare Pages プロジェクト

- [Cloudflare Dashboard](https://dash.cloudflare.com/) → **Workers & Pages** → **Create** → **Pages** → **Connect to Git**。
- 本リポジトリを選択し、ブランチは `main`。
- **Build configuration**:
  - Framework preset: **Next.js**（または None で手動設定）。
  - Build command: `cd frontend && npm ci && npm run build`
  - Build output directory: `frontend/.next`（Next.js の場合は通常、Cloudflare が自動検出）。
  - Root directory: リポジトリルートのまま（モノレポのため `frontend` をサブディレクトリに指定する場合あり。Cloudflare Pages の「Root directory」を `frontend` にすると、Build command は `npm run build` でよい）。

### 3.2 環境変数（Cloudflare Pages）

- **Settings** → **Environment variables** で本番用に追加:
  - `NEXT_PUBLIC_API_URL`: 上記 Cloud Run のサービス URL（例: `https://blog-api-xxxxx-an.a.run.app`）。
- 変更後は再ビルドが必要。

### 3.3 カスタムドメイン（任意）

- **Custom domains** でドメインを追加し、DNS の CNAME を案内に従って設定。

---

## 4. デプロイの流れ

### 4.1 手動デプロイ

1. **API**: 上記のとおり `docker build` → `docker push` → `gcloud run deploy`。
2. **フロント**: Cloudflare の「Create deployment」で main をビルド・デプロイするか、ローカルで `cd frontend && npm run build` のうえ、Wrangler などでアップロード。

### 4.2 CI/CD（GitHub Actions）

- **デプロイ用ワークフロー**: `.github/workflows/deploy-api.yml` が `main` への push（backend 関連の変更）または手動実行で、Cloud Run へ API をビルド・デプロイする。
- **必要な GitHub Secrets**:
  - `GCP_PROJECT_ID`: GCP プロジェクト ID
  - `GCP_SA_KEY`: サービスアカウントの JSON キー（Artifact Registry への push と Cloud Run のデプロイ権限を持つこと）
  - 任意: `GCP_REGION`（未設定時は `asia-northeast1`）、`IMAGE_NAME`（未設定時は `blog-api`）
- **事前準備**: GCP 側で Artifact Registry リポジトリ `blog-repo` の作成、Cloud Run サービス・シークレットの初回設定が必要。詳細は本ドキュメントの「2. GCP 側の設定」を参照。
- **フロント**: Cloudflare Pages は Git 連携で `main` への push 時に自動ビルドされる想定。環境変数 `NEXT_PUBLIC_API_URL` を Cloudflare のダッシュボードで設定する。

---

## 5. チェックリスト

- [ ] GCP プロジェクトで課金・API 有効化
- [ ] Artifact Registry リポジトリ作成
- [ ] Cloud SQL インスタンス作成・DB 作成・マイグレーション適用
- [ ] Secret Manager に `DATABASE_DSN`・`ADMIN_API_KEY` 作成
- [ ] Cloud Run に API をデプロイし、サービス URL を確認
- [ ] Cloudflare Pages でリポジトリ連携・ビルド設定・`NEXT_PUBLIC_API_URL` 設定
- [ ] フロントから API へリクエストが通ることを確認（CORS は Connect のデフォルトで許可されている想定。必要なら Cloud Run でレスポンスヘッダを調整）

---

## 6. 参考

- [Cloud Run ドキュメント](https://cloud.google.com/run/docs)
- [Cloudflare Pages - Documentation](https://developers.cloudflare.com/pages/)
- [ADR-002: フロントを Cloudflare Pages、API を GCP Cloud Run でホスティングする](adr/002-hosting-cloudflare-and-cloudrun.md)
