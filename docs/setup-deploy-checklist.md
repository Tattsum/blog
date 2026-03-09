# セットアップ・デプロイ やること一覧（詳細手順）

**対象**: 個人ブログを GCP Cloud Run（API）と Cloudflare Pages（フロント）で本番運用するまでの作業一覧。  
**記載基準**: 2026年3月時点の公式ドキュメントおよび推奨されるやり方に基づく。

関連: [インフラ設定ガイド](infrastructure.md)、[ADR-002（ホスティング方針）](adr/002-hosting-cloudflare-and-cloudrun.md)。

---

## 全体の流れ

1. **GCP 準備** … プロジェクト・課金・API 有効化
2. **GCP 認証（CI 用）** … Workload Identity Federation（鍵レス）またはサービスアカウントキー
3. **Artifact Registry** … コンテナイメージ保存用リポジトリ
4. **Cloud SQL** … MySQL インスタンス・DB・マイグレーション（**Terraform 可**: 4〜6 を [terraform/README.md](../terraform/README.md) で一括実施可能）
5. **Secret Manager** … DSN や API キーなどのシークレット
6. **Cloud Run** … API の初回デプロイとサービス URL の確認
7. **Cloudflare Pages** … リポジトリ連携・ビルド設定・環境変数（手動）
8. **動作確認** … フロントから API への通信・管理画面ログイン

---

## 1. GCP プロジェクトの準備

### 1.1 やること

- [ ] GCP プロジェクトを作成する（または既存を利用する）
- [ ] 課金を有効にする
- [ ] 必要な API を有効にする

### 1.2 詳細手順（2026年3月時点）

1. **プロジェクト作成**
   - [Google Cloud Console](https://console.cloud.google.com/) にログイン
   - 画面上部のプロジェクト選択 → 「新しいプロジェクト」→ プロジェクト名・ID を入力（例: `myblog-prod`）
   - 作成後、プロジェクト ID を控える（以降 `GCP_PROJECT_ID` として参照）

2. **課金の有効化**
   - 「お支払い」→ 課金アカウントをリンク（未作成の場合は作成）
   - プロジェクトに課金アカウントが紐づいていることを確認

3. **必要な API の一括有効化**

   ```bash
   export GCP_PROJECT_ID=your-project-id   # 実際の ID に置き換え
   gcloud config set project $GCP_PROJECT_ID

   gcloud services enable \
     run.googleapis.com \
     sqladmin.googleapis.com \
     secretmanager.googleapis.com \
     artifactregistry.googleapis.com \
     iamcredentials.googleapis.com \
     cloudbuild.googleapis.com \
     servicenetworking.googleapis.com \
     compute.googleapis.com
   ```

   - `iamcredentials.googleapis.com`: Workload Identity Federation（OIDC）で GitHub Actions から鍵なし認証する場合に必要
   - `servicenetworking.googleapis.com` / `compute.googleapis.com`: Cloud SQL をプライベート IP で使う場合に必要（後述）

---

## 2. CI/CD 用の GCP 認証（GitHub Actions）

GitHub Actions から Cloud Run へデプロイするには、**Workload Identity Federation（OIDC）** を使う方法（鍵なし・推奨）と、**サービスアカウントキー** を GitHub Secrets に登録する方法があります。2026年時点では鍵レス（OIDC）が推奨です。

**機密情報について**: プロジェクト ID や GitHub の org/repo などは `.envrc` に設定し、リポジトリにはコミットしないことを推奨します。`.envrc` は `.gitignore` に含まれており、サンプルは `.envrc.example` を参照してください。

### 2.1 やること

- [ ] **方法 A（推奨）**: Workload Identity Federation を設定し、GitHub の OIDC で GCP に認証する
- [ ] **方法 B**: サービスアカウントキーを発行し、GitHub Secrets に登録する
- [ ] デプロイ用サービスアカウントに必要な IAM ロールを付与する

### 2.2 方法 A: Workload Identity Federation（OIDC）の詳細手順

1. **Workload Identity プールの作成**

   ```bash
   # 未設定なら .envrc.example を .envrc にコピーして値を設定し、source .envrc または direnv allow
   export GCP_PROJECT_NUMBER=$(gcloud projects describe $GCP_PROJECT_ID --format='value(projectNumber)')

   gcloud iam workload-identity-pools create $POOL_NAME \
     --project=$GCP_PROJECT_ID \
     --location=global \
     --display-name="GitHub Actions"
   ```

2. **OIDC プロバイダの追加**

   **重要**: GCP では attribute condition が必須です。条件で参照する claim は attribute-mapping に含める必要があります。

   ```bash
   # 自分の GitHub の org / リポジトリ名に合わせて変更（.envrc で export している場合は不要）
   # GITHUB_REPO はリポジトリ名のみ（例: blog）。principalSet の path が attribute.repository/ORG/REPO になる
   export GITHUB_ORG=YOUR_ORG_OR_USERNAME
   export GITHUB_REPO=blog

   gcloud iam workload-identity-pools providers create-oidc $PROVIDER_NAME \
     --project=$GCP_PROJECT_ID \
     --location=global \
     --workload-identity-pool=$POOL_NAME \
     --display-name="GitHub OIDC" \
     --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository,attribute.repository_owner=assertion.repository_owner" \
     --attribute-condition="assertion.repository_owner=='$GITHUB_ORG'" \
     --issuer-uri="https://token.actions.githubusercontent.com"
   ```

3. **デプロイ用サービスアカウントの作成とロール付与**

   必ず先にサービスアカウントを作成してから、ロールを付与してください。

   ```bash
   export SA_NAME=blog-deploy
   export SA_EMAIL=${SA_NAME}@${GCP_PROJECT_ID}.iam.gserviceaccount.com

   # 3.1 サービスアカウントを作成（未作成の場合のみ）
   gcloud iam service-accounts create $SA_NAME \
     --project=$GCP_PROJECT_ID \
     --display-name="Blog API Deploy"

   # 3.2 ロールを付与（上記の作成後に実行）
   gcloud projects add-iam-policy-binding $GCP_PROJECT_ID \
     --member="serviceAccount:${SA_EMAIL}" \
     --role="roles/artifactregistry.writer"
   gcloud projects add-iam-policy-binding $GCP_PROJECT_ID \
     --member="serviceAccount:${SA_EMAIL}" \
     --role="roles/run.admin"
   gcloud projects add-iam-policy-binding $GCP_PROJECT_ID \
     --member="serviceAccount:${SA_EMAIL}" \
     --role="roles/iam.serviceAccountUser"
   ```

4. **GitHub リポジトリにだけ権限を渡す（プールの IAM バインド）**

   ```bash
   gcloud iam service-accounts add-iam-policy-binding $SA_EMAIL \
     --project=$GCP_PROJECT_ID \
     --role="roles/iam.workloadIdentityUser" \
     --member="principalSet://iam.googleapis.com/projects/${GCP_PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_NAME}/attribute.repository/${GITHUB_ORG}/${GITHUB_REPO}"
   ```

5. **GitHub Actions ワークフローでの利用**

   - リポジトリの **Settings → Secrets and variables → Actions** で **Variables** を追加:
     - `GCP_PROJECT_ID`: 上記のプロジェクト ID
     - （任意）`GCP_REGION`: 例 `asia-northeast1`
   - **Secrets には GCP の鍵を登録しない**（OIDC で認証するため）
   - ワークフロー側では `google-github-actions/auth` で `workload_identity_provider` と `service_account` を指定する（後述「ワークフロー例」参照）

### 2.3 方法 B: サービスアカウントキーを使う場合

1. 上記と同様にデプロイ用サービスアカウント `blog-deploy` を作成し、`roles/artifactregistry.writer` / `roles/run.admin` / `roles/iam.serviceAccountUser` を付与する。
2. キーを発行:

   ```bash
   gcloud iam service-accounts keys create key.json \
     --project=$GCP_PROJECT_ID \
     --iam-account=blog-deploy@${GCP_PROJECT_ID}.iam.gserviceaccount.com
   ```

3. `key.json` の内容を **GitHub → Settings → Secrets and variables → Actions** で新規 Secret を作成し、名前を `GCP_SA_KEY`、値に JSON 全体を貼り付けて保存。
4. ローカルの `key.json` は削除し、リポジトリにコミットしない。

---

## 3. Artifact Registry の作成

### 3.1 やること

- [ ] Docker 形式の Artifact Registry リポジトリを 1 つ作成する

### 3.2 詳細手順

```bash
export GCP_PROJECT_ID=your-project-id
export REGION=asia-northeast1

gcloud artifacts repositories create blog-repo \
  --project=$GCP_PROJECT_ID \
  --repository-format=docker \
  --location=$REGION \
  --description="Blog API container images"
```

- リポジトリ名は `blog-repo` のままにすると、既存の `.github/workflows/deploy-api.yml` のイメージパスと一致します。
- 別名にする場合は、ワークフロー内の `blog-repo` をその名前に合わせて変更してください。

---

## 4. Cloud SQL (MySQL) の作成とマイグレーション

**Terraform で実施する場合**: 4・5・6 を一括で管理する場合は [terraform/README.md](../terraform/README.md) を参照し、`terraform/` で Cloud SQL・Secret Manager・Cloud Run を作成できます。以下は手動での手順です。

### 4.1 やること

- [ ] Cloud SQL for MySQL インスタンスを作成する
- [ ] インスタンスに `blog` データベースを作成する
- [ ] マイグレーションを実行してスキーマを適用する
- [ ] （任意）本番ではプライベート IP と VPC コネクタを検討する

### 4.2 詳細手順（2026年3月時点）

1. **インスタンス作成（コンソール）**
   - [Cloud SQL](https://console.cloud.google.com/sql) → 「インスタンスを作成」→ **MySQL** を選択
   - **MySQL のバージョン**: 8.4（LTS 推奨）
   - **マシンタイプ**: 開発なら「共有コア」、本番なら「専用コア」を選択
   - **ストレージ**: 20 GB 以上（必要に応じて自動増量を有効化）
   - **接続**: 開発では「パブリック IP」で十分。本番でプライベート IP を使う場合は「プライベート IP」を有効にし、VPC と割り当てられた IP 範囲を設定
   - ルートパスワードを設定し、接続名（`PROJECT:REGION:INSTANCE`）を控える

2. **データベース `blog` の作成**
   - 作成したインスタンスを開く → 「データベース」→「データベースを作成」→ 名前: `blog`

3. **接続情報の確認**
   - 「接続」タブで「このインスタンスへの接続」の接続名（例: `myblog-prod:asia-northeast1:blog-mysql`）を確認
   - パブリック IP を使う場合: `HOST` は「このインスタンスの IP アドレス」、ポートは `3306`
   - 接続文字列の例（Go 用）:
     - パブリック IP: `USER:PASSWORD@tcp(IP:3306)/blog?parseTime=true`
     - Cloud Run から接続する場合（後述）: Unix ソケット `PROJECT:REGION:INSTANCE` を使う

4. **マイグレーションの実行**
   - ローカルから接続する場合（Cloud SQL Auth Proxy 推奨）:
     - [Cloud SQL Auth Proxy](https://cloud.google.com/sql/docs/mysql/connect-auth-proxy) をインストールし、プロキシ経由で `localhost:3306` に接続
     - [golang-migrate](https://github.com/golang-migrate/migrate) をインストール後:

     ```bash
     migrate -path backend/db/migrations \
       -database "mysql://USER:PASSWORD@tcp(localhost:3306)/blog?parseTime=true" \
       up
     ```

   - 接続ユーザは Cloud SQL の「ユーザー」で作成したもの（ルートでも可。本番では専用ユーザを推奨）

5. **（本番）Cloud Run から Cloud SQL へ接続する場合**
   - Cloud Run には **Cloud SQL の接続** を追加する方法が公式で推奨されています（自動で Unix ソケットがマウントされ、IAM で認証も可能）
   - デプロイ時に `--add-cloudsql-instances=PROJECT:REGION:INSTANCE` を指定する（後述「6. Cloud Run へのデプロイ」）
   - プライベート IP のみのインスタンスの場合は、**Serverless VPC Access コネクタ** を作成し、Cloud Run に `--vpc-connector` でそのコネクタを指定する必要があります（[Connecting to Private Cloud SQL from Cloud Run](https://cloud.google.com/sql/docs/mysql/connect-run) を参照）

---

## 5. Secret Manager の作成

**Terraform で実施する場合**: [terraform/README.md](../terraform/README.md) を参照（4 とあわせて実行）。

### 5.1 やること

- [ ] Cloud Run に渡すシークレットを Secret Manager に登録する
- [ ] 少なくとも `DATABASE_DSN` と `ADMIN_API_KEY` を作成する

### 5.2 詳細手順

1. **DATABASE_DSN**
   - Cloud SQL の接続文字列をそのまま格納します。
   - **パブリック IP の場合**（ローカルや Cloud Run から TCP で接続）:

     ```bash
     echo -n 'mysql://USER:PASSWORD@tcp(INSTANCE_IP:3306)/blog?parseTime=true' | \
       gcloud secrets create DATABASE_DSN --data-file=- --project=$GCP_PROJECT_ID
     ```

   - **Cloud Run で Unix ソケットを使う場合**（推奨）:
     - 接続名を `MYPROJECT:REGION:INSTANCE` とすると、アプリ内では `USER:PASSWORD@unix(/cloudsql/MYPROJECT:REGION:INSTANCE)/blog?parseTime=true` のような形式になります（使用する Go ドライバの仕様に合わせてください。`go-sql-driver/mysql` では `/cloudsql/CONNECTION_NAME` をホストに指定）

2. **ADMIN_API_KEY**

   ```bash
   echo -n 'your-secure-admin-api-key' | \
     gcloud secrets create ADMIN_API_KEY --data-file=- --project=$GCP_PROJECT_ID
   ```

3. **Cloud Run のサービスアカウントに Secret アクセス権を付与**
   - Cloud Run はデフォルトで「Compute Engine のデフォルトサービスアカウント」などで動くため、その SA に `roles/secretmanager.secretAccessor` を付与するか、カスタム SA を使いその SA に付与します。

   ```bash
   export PROJECT_NUMBER=$(gcloud projects describe $GCP_PROJECT_ID --format='value(projectNumber)')
   gcloud secrets add-iam-policy-binding DATABASE_DSN \
     --project=$GCP_PROJECT_ID \
     --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
     --role="roles/secretmanager.secretAccessor"
   gcloud secrets add-iam-policy-binding ADMIN_API_KEY \
     --project=$GCP_PROJECT_ID \
     --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
     --role="roles/secretmanager.secretAccessor"
   ```

---

## 6. Cloud Run へのデプロイ

**Terraform で実施する場合**: [terraform/README.md](../terraform/README.md) を参照（4・5 とあわせて実行）。イメージを push したうえで `terraform apply` でデプロイ可能です。

### 6.1 やること

- [ ] リポジトリルートから `backend/Dockerfile` でイメージをビルドし、Artifact Registry に push する
- [ ] Cloud Run にサービスをデプロイし、シークレットと（必要なら）Cloud SQL 接続を設定する
- [ ] 発行されたサービス URL を控える

### 6.2 詳細手順

1. **ローカルでイメージをビルド・push する場合**
   - リポジトリのルートで実行（コンテキストはルート、Dockerfile は `backend/Dockerfile`）:

   ```bash
   export GCP_PROJECT_ID=your-project-id
   export REGION=asia-northeast1
   docker build -t ${REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/blog-repo/blog-api:latest -f backend/Dockerfile .
   gcloud auth configure-docker ${REGION}-docker.pkg.dev --quiet
   docker push ${REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/blog-repo/blog-api:latest
   ```

2. **Cloud Run にデプロイ**
   - 最小限の例（パブリック IP の Cloud SQL を DSN で指定している場合）:

   ```bash
   gcloud run deploy blog-api \
     --project=$GCP_PROJECT_ID \
     --image=$REGION-docker.pkg.dev/$GCP_PROJECT_ID/blog-repo/blog-api:latest \
     --platform=managed \
     --region=$REGION \
     --allow-unauthenticated \
     --set-secrets=DATABASE_DSN=DATABASE_DSN:latest,ADMIN_API_KEY=ADMIN_API_KEY:latest \
     --set-env-vars=PORT=8080
   ```

   - **Cloud SQL の Unix ソケット接続を使う場合**（推奨）:
     - 接続名を `CONNECTION_NAME`（例: `myblog-prod:asia-northeast1:blog-mysql`）とすると:

     ```bash
     gcloud run deploy blog-api \
       --project=$GCP_PROJECT_ID \
       --image=$REGION-docker.pkg.dev/$GCP_PROJECT_ID/blog-repo/blog-api:latest \
       --platform=managed \
       --region=$REGION \
       --allow-unauthenticated \
       --add-cloudsql-instances=CONNECTION_NAME \
       --set-secrets=DATABASE_DSN=DATABASE_DSN:latest,ADMIN_API_KEY=ADMIN_API_KEY:latest \
       --set-env-vars=PORT=8080
     ```

     - このとき `DATABASE_DSN` の値は、ホスト部分を Unix ソケット用にしたもの（ドライバに合わせて `/cloudsql/CONNECTION_NAME` など）にします。

3. **サービス URL の確認**
   - デプロイ後、コンソールまたは `gcloud run services describe blog-api --region=$REGION --format='value(status.url)'` で URL を取得し、**フロントの本番環境変数 `NEXT_PUBLIC_API_URL` に設定**します（例: `https://blog-api-xxxxx-an.a.run.app`）。

---

## 7. Cloudflare Pages の設定（手動）

Cloudflare Pages は Terraform での管理が難しいため、ここではダッシュボードでの手動設定手順を詳細に記載します。

### 7.1 やること

- [ ] Cloudflare ダッシュボードで Pages プロジェクトを作成し、Git リポジトリと接続する
- [ ] モノレポ用にルートディレクトリ・ビルドコマンド・ビルド出力ディレクトリを設定する
- [ ] 環境変数 `NEXT_PUBLIC_API_URL` を設定する
- [ ] （任意）カスタムドメインを設定する

### 7.2 前提（事前に用意するもの）

- Cloudflare アカウント（[dash.cloudflare.com](https://dash.cloudflare.com/) でサインアップ可）
- GitHub に本リポジトリ（または fork）が push 済みであること
- **6. Cloud Run** まで完了し、Cloud Run のサービス URL を控えていること（例: `https://blog-api-xxxxx-an.a.run.app`）。この URL を後で環境変数に設定します。

### 7.3 手順 1: Pages プロジェクトの作成と Git 接続

1. [Cloudflare Dashboard](https://dash.cloudflare.com/) にログインする。
2. 左サイドバーで **Workers & Pages** をクリックする。
3. **Create** ボタンをクリックし、**Pages** を選択する。
4. **Connect to Git** を選択する。
5. **GitHub** が表示されたら **Connect GitHub** をクリックし、表示される手順に従って GitHub と連携する（初回のみ。Organization の場合は「All repositories」または対象リポジトリへのアクセスを許可する）。
6. リポジトリ一覧から **Tattsum/blog**（または自分の fork の名前）を選択する。
7. **Begin setup** をクリックする。
8. 次の「7.4 ビルド設定」の画面に進む。

### 7.4 手順 2: ビルド設定（モノレポ・Next.js）

このリポジトリはモノレポのため、**ルートディレクトリを `frontend` に変更**する必要があります。

1. **Project name**  
   - 任意（例: `blog` や `myblog`）。サブドメインは `プロジェクト名.pages.dev` になる。

2. **Production branch**  
   - `main` のままにする（本番デプロイ対象のブランチ）。

3. **Build configuration** で **Framework preset**  
   - **Next.js** を選択する。  
   - 一覧にない場合は **None** を選び、以降の項目を手動で入力する。

4. **Root directory（重要）**  
   - **Set root directory** にチェックを入れる。  
   - 値に **`frontend`** と入力する。  
   - これにより、ビルドはリポジトリルートではなく `frontend` ディレクトリ直下で実行される。

5. **Build command と Deploy command（OpenNext / Cloudflare Workers の場合・重要）**  
   - 本プロジェクトは **OpenNext**（`@opennextjs/cloudflare`）で Workers にデプロイするため、**`.open-next/worker.js` はデプロイ時に自動作成されません**。  
   - **Build command**: 依存インストールのみでよい場合（Cloudflare が先に `npm clean-install` するなら）**空**で可。手動なら **`npm ci`**。  
   - **Deploy command**: 必ず **`npm run deploy`** を指定する。  
     - `npm run deploy` は `opennextjs-cloudflare build`（Next.js ビルド＋`.open-next` 生成）の後に `opennextjs-cloudflare deploy` を実行する。  
     - **`npx wrangler deploy` だけにしていると** `.open-next/worker.js` が存在せず **「The entry-point file at ".open-next/worker.js" was not found」** で失敗する。  
   - 従来の **Next.js（Pages 静的＋Functions）** のみ使う場合は **Build command**: `npm run build`、Deploy は Cloudflare のデフォルトのままでよい。

6. **Build output directory**  
   - **Next.js** プリセットの場合は、Cloudflare の Next.js 統合により自動設定されることが多い。表示されている場合はそのまま。  
   - 手動の場合は、Next.js 13+ App Router では **`.next`** が使われる。静的エクスポート（`output: 'export'`）にしている場合は **`out`** を指定する。  
   - 本プロジェクトはデフォルトの Next.js 設定のため、自動設定または **`.next`** を採用する。

7. **Environment variables**  
   - この段階では追加しなくてよい。**Save** 後、**7.6 手順 4** で環境変数を追加する。

8. **Save and Deploy** をクリックする。  
   - 初回ビルドが開始される。ルートを `frontend` にしているため、`frontend` 内の `package.json` と `next.config.ts` が使われる。

### 7.5 手順 3: Node バージョンの指定（推奨）

Cloudflare Pages のビルドで使う Node バージョンを固定すると、ビルドが安定しやすいです。

1. リポジトリの **`frontend`** ディレクトリ直下に **`.nvmrc`** を作成する（未作成の場合）。  
   - 中身は 1 行で **`20`** または **`22`**（例: `20`）。  
   - 多くの Cloudflare ビルド環境では `.nvmrc` が読まれ、そのバージョンが使われる。

2. **`frontend/package.json`** に **`engines`** を追加する（任意だが推奨）:

   ```json
   "engines": {
     "node": ">=20.0.0"
   }
   ```

3. 上記をコミット・push すると、次回のデプロイからその Node バージョンが使われる。

### 7.6 手順 4: 環境変数と wrangler 設定

#### wrangler 設定ファイルと本番 URL のリポジトリ管理

リポジトリの `frontend/wrangler.jsonc` に設定を置き、**本番用 API URL もリポジトリで管理**しています。

- **ローカル用**: `vars.NEXT_PUBLIC_API_URL` に `http://localhost:8080` を記載。
- **本番用**: `env.production.vars.NEXT_PUBLIC_API_URL` に Cloud Run の URL を記載。あわせて Next のビルド時に読まれる **`frontend/.env.production`** にも同じ URL を記載済み（`terraform output cloud_run_url` と一致させてください）。Cloudflare のダッシュボードで変数を設定しなくてもデプロイで本番 URL が使われます。
- Cloud Run の URL を変更した場合は、`wrangler.jsonc` の `env.production.vars` と `.env.production` の両方を更新し、push して再デプロイしてください。
- ローカルで `npm run preview` や `npm run deploy` を実行するときは、このファイルが参照されます（`npx wrangler deploy` 単体では `.open-next` が無いため失敗します）。

#### ダッシュボードで環境変数を設定する場合

Cloud Run の API をフロントから呼ぶために、本番用の環境変数を設定します。

1. Cloudflare Dashboard の **Workers & Pages** から、作成した **Pages プロジェクト**（例: blog）をクリックする。
2. 上部タブで **Settings** をクリックする。
3. 左メニューで **Environment variables** をクリックする。
4. **Add variable** または **Add** をクリックする。
5. 以下を入力する:
   - **Variable name**: `NEXT_PUBLIC_API_URL`
   - **Value**: Cloud Run のサービス URL（例: `https://blog-api-xxxxx-an.a.run.app`）。末尾のスラッシュは付けない。
   - **Environment**: **Production** にチェック。Preview でも同じ API を使う場合は **Preview** にもチェックする。
6. **Save** する。
7. 環境変数を追加・変更した場合は、**Deployments** タブで最新のデプロイの **⋯** → **Retry deployment** を実行するか、main に空コミットを push して再ビルドする。

### 7.7 手順 5: 初回デプロイとビルド確認

1. **Deployments** タブで、初回または再デプロイの **Status** が **Success** になるまで待つ。
2. 失敗した場合は **View build logs** を開き、以下を確認する:
   - **Root directory**: `frontend` になっているか。  
   - **Deploy command**: OpenNext 利用時は **`npm run deploy`** になっているか。`npx wrangler deploy` のみだと `.open-next/worker.js was not found` になる。  
   - **Build command**: 依存インストールが行われているか（Cloudflare が事前に `npm clean-install` する場合は Deploy で `npm run deploy` がビルドも実行する）。  
   - **Node のバージョン**: `.nvmrc` がある場合はそのバージョンになっているか。  
   - エラーメッセージ: `MODULE_NOT_FOUND` の場合は依存不足、`Cannot find module 'next'` の場合はルートディレクトリ誤りや `npm install` 不足の可能性がある。
3. 成功したら、**View site** または **Open production URL** で `https://プロジェクト名.pages.dev` を開き、トップページや記事一覧が表示されることを確認する。
4. ブラウザの開発者ツールのネットワークタブで、API リクエストが `NEXT_PUBLIC_API_URL` で設定した Cloud Run の URL に向かっていることを確認する。

### 7.8 手順 6: カスタムドメイン（任意）

1. 同じ Pages プロジェクトの **Custom domains** を開く。
2. **Set up a custom domain** をクリックし、取得済みのドメイン（例: `blog.example.com`）を入力する。
3. 案内に従って DNS で **CNAME** を設定する（Cloudflare にネームサーバーを移管している場合はプロキシ設定も可能）。  
   - ターゲットは通常 `プロジェクト名.pages.dev` または案内に表示される値。
4. 検証が完了すると、そのドメインで Pages のサイトにアクセスできるようになる。

### 7.9 トラブルシューティング

| 現象 | 確認・対処 |
| --- | --- |
| ビルドが「ルートで実行されている」ようなエラー | Root directory が `frontend` になっているか確認。Set root directory にチェックが入っているか確認。 |
| `npm run build` で Next が見つからない | ルートが `frontend` か確認。`frontend/package.json` に `next` が入っているか確認。 |
| `.open-next/worker.js` was not found | **Deploy command** を **`npm run deploy`** に変更する。`npx wrangler deploy` だけでは OpenNext のビルドが走らず `.open-next` が作られない。 |
| 本番で API に接続できない | **Settings → Environment variables** で `NEXT_PUBLIC_API_URL` が Production に設定されているか確認。値の末尾にスラッシュがないか、Cloud Run の URL が正しいか確認。変更後は再デプロイが必要。 |
| Node のバージョン不一致 | `frontend/.nvmrc` に `20` または `22` を入れ、再デプロイ。 |
| 403 / CORS エラー | Cloud Run 側で CORS が許可されているか、および `NEXT_PUBLIC_API_URL` が正しいか確認。 |

---

## 8. GitHub Actions ワークフロー（OIDC 利用時）

Workload Identity Federation を使う場合、`.github/workflows/deploy-api.yml` では **Secrets に `GCP_SA_KEY` を置かず**、代わりに **Variables** で `GCP_PROJECT_ID` を渡し、`google-github-actions/auth` で OIDC を指定します。

### 8.1 やること

- [ ] 上記「2.2 方法 A」で WIF とサービスアカウントを用意する
- [ ] GitHub の Variables に `GCP_PROJECT_ID`（と任意で `GCP_REGION`）を設定する
- [ ] ワークフローで `workload_identity_provider` と `service_account` を指定する

### 8.2 ワークフロー例（OIDC）

既存の `deploy-api.yml` の「Authenticate to Google Cloud」を次のように差し替えます（プール名・プロバイダ名・サービスアカウントは「2.2」で設定した値に合わせてください）:

```yaml
- name: Authenticate to Google Cloud (OIDC)
  uses: google-github-actions/auth@v2
  with:
    workload_identity_provider: projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/github-actions-pool/providers/github-oidc
    service_account: blog-deploy@GCP_PROJECT_ID.iam.gserviceaccount.com
```

- `PROJECT_NUMBER` は `gcloud projects describe $GCP_PROJECT_ID --format='value(projectNumber)'` で取得
- 先頭の `permissions:` に `id-token: write` が必要です（OIDC トークン発行のため）:

```yaml
jobs:
  deploy-api:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write
```

### 8.3 デプロイ時のマイグレーション（CI で実行）

`deploy-api.yml` では、**デプロイ前に** 本番 Cloud SQL に対して `backend/db/migrations` のマイグレーションを CI 上で実行します（`paths` に `backend/**` が含まれるため、migrations の変更でもデプロイが走り、その際に migrate が実行されます）。

#### 必要な設定

1. **（初回のみ）マイグレーション用ユーザー `migrate` に権限を付与**
   - Terraform で `migrate`@'%' ユーザーが作成される（`terraform/cloudsql.tf`）。Proxy 経由（`cloudsqlproxy~IP`）からの接続は、host が `%` のユーザーでないと「Access denied」になるため、root ではなくこのユーザーを CI で使う。
   - **初回のみ**、Cloud SQL に接続して以下を実行する（GCP Console の「Cloud Shell で接続」や、ローカルで Cloud SQL Auth Proxy を起動したうえで `mysql` クライアントから root で接続して実行）。

     ```sql
     GRANT ALL PRIVILEGES ON `blog`.* TO 'migrate'@'%';
     FLUSH PRIVILEGES;
     ```

2. **GitHub Secrets にマイグレーション用の Secret を追加（いずれか一方）**
   - **推奨: `MIGRATION_PASSWORD`** — 値は Terraform の `db_root_password` を**そのまま**（URL エンコード不要）。ワークフロー側で DSN を組み立てるため、`+` などの特殊文字でもそのままでよい。
   - **代替: `MIGRATION_DSN`** — 全文を指定する場合。形式: `mysql://migrate:パスワード@tcp(127.0.0.1:3306)/blog?parseTime=true`。パスワードに `+` などが含まれる場合は URL エンコード（例: `+` → `%2B`）が必要。

3. **デプロイ用サービスアカウントに Cloud SQL Client ロールを付与**
   - マイグレーション実行時に Cloud SQL Auth Proxy がインスタンスに接続するため、デプロイ用 SA（例: `blog-deploy@PROJECT_ID.iam.gserviceaccount.com`）に **Cloud SQL Client**（`roles/cloudsql.client`）を付与する。
   - 例（Terraform で作成した `blog-mysql` の場合）:

     ```bash
     gcloud projects add-iam-policy-binding $GCP_PROJECT_ID \
       --member="serviceAccount:blog-deploy@${GCP_PROJECT_ID}.iam.gserviceaccount.com" \
       --role="roles/cloudsql.client"
     ```

4. **（上記 1 の GRANT が未実施の場合）** 先に「1. マイグレーション用ユーザーに権限を付与」を実行してから、GitHub Secrets の `MIGRATION_PASSWORD`（推奨）または `MIGRATION_DSN` を設定し、デプロイを実行する。

#### 8.3.1 初回セットアップ詳細（Step 2・Step 3）

以下は「migrate に権限付与」と「MIGRATION_DSN 設定」を、初回だけ確実に行うための詳細手順です。

---

##### Step 2: 初回のみ — migrate ユーザーに DB 権限を付与

Terraform で作成した `migrate` ユーザーは、作成直後はどのデータベースにもアクセスできません。root で Cloud SQL に接続し、`blog` データベースに対する権限を 1 回だけ付与する必要があります。

###### 2-1. Cloud SQL に接続する（いずれか一方でよい）

- **方法 A: GCP Console の Cloud Shell から接続（推奨）**
  1. [Google Cloud Console](https://console.cloud.google.com/) を開き、プロジェクト（例: `kano-blog-prod`）を選択する。
  2. 左メニューから **SQL**（または「データベース」→「SQL」）を開く。
  3. インスタンス **blog-mysql** をクリックする。
  4. 画面上部の **「Cloud Shell で接続」** をクリックする（または、あらかじめ Cloud Shell を開いておき、次のコマンドを実行する）。
  5. 接続方法で **「gcloud sql connect を使用」** を選び、表示されるコマンドを実行する（MySQL では `--database` は使えないため付けない）。例:

     ```bash
     gcloud sql connect blog-mysql --user=root --project=kano-blog-prod
     ```

  6. パスワードを聞かれたら、**Terraform の `db_root_password`**（`terraform/terraform.tfvars` に記載の値）を入力する。
  7. 接続できたら、MySQL のプロンプト（`mysql>`）が表示される。

- **方法 B: ローカル PC から Cloud SQL Auth Proxy + mysql クライアントで接続**
  1. [Cloud SQL Auth Proxy](https://cloud.google.com/sql/docs/mysql/connect-auth-proxy) をダウンロードし、PATH の通った場所に置く（または `curl` で取得）。
  2. ターミナルでプロキシを起動する（接続名は Terraform の `output cloud_sql_connection_name` で確認可能）:

     ```bash
     cloud-sql-proxy --port 3306 kano-blog-prod:asia-northeast1:blog-mysql
     ```

  3. **別のターミナル**で、mysql クライアントをインストール済みであれば:

     ```bash
     mysql -h 127.0.0.1 -P 3306 -u root -p blog
     ```

  4. パスワードを聞かれたら、**Terraform の `db_root_password`** を入力する。
  5. `mysql>` プロンプトが表示されれば接続成功。

###### 2-2. root で接続できない場合（Access denied for 'root'@'cloudsqlproxy~...'）

**原因**: Cloud SQL Proxy 経由だと接続元が `cloudsqlproxy~<IP>` と見え、root がその host を許可していないと拒否されます。パスワードが正しくても「Access denied」になります。

**対処（公開 IP で直接接続する）**: Proxy を使わず、インスタンスの**公開 IP** に mysql クライアントで接続すると、接続元が単なる IP になり、`root@'%'` で通ることがあります。

1. **認証済みネットワークに Cloud Shell の IP を追加**
   - GCP Console → **SQL** → **blog-mysql** → **接続**（Connections）タブ。
   - **認証済みネットワーク** → **ネットワークを追加**。
   - 名前: 例 `cloudshell`。ネットワーク: **34.81.32.136/32**（Cloud Shell の外向き IP。別の環境ならそのマシンの IP を指定）。**保存**。

2. **インスタンスの公開 IP を確認**
   - 同じ **接続** タブの「このインスタンスの接続名」付近に、**公開 IP アドレス**が表示されます。例: `34.84.xxx.xxx`。メモする。

3. **Cloud Shell で mysql クライアントを入れて接続**
   - Cloud Shell で次を実行（`<PUBLIC_IP>` は上記の公開 IP に置き換え）:

     ```bash
     sudo apt-get update && sudo apt-get install -y default-mysql-client
     mysql -h <PUBLIC_IP> -u root -p
     ```

   - パスワードを聞かれたら、**Terraform の `db_root_password`**（`terraform.tfvars` の値）を入力する。

4. **接続できたら 2-3 の SQL を実行**
   - `mysql>` プロンプトで、前述の `GRANT` と `FLUSH PRIVILEGES` を実行する。

5. **（任意）セキュリティのため認証済みネットワークを削除**
   - 接続・GRANT が終わったら、**接続** タブで追加したネットワークを削除してよい。

**それでも接続できない場合**: root のパスワードが Terraform の `db_root_password` と一致しているか確認する。GCP Console の **SQL → blog-mysql → ユーザー** で root のパスワードを**リセット**し、`terraform.tfvars` の `db_root_password` と同一にしてから再度試す。

###### 2-3. SQL を実行する

MySQL のプロンプト（`mysql>`）が表示されている状態で、以下を 1 行ずつ実行します。

```sql
GRANT ALL PRIVILEGES ON `blog`.* TO 'migrate'@'%';
FLUSH PRIVILEGES;
```

- `Query OK` や `Rows affected: 0` などが出れば成功です。
- 終了する場合は `exit` または `\q` で MySQL を抜け、Cloud Shell の場合はそのまま、ローカルの場合はプロキシを Ctrl+C で止めてください。

これで、CI から `migrate` ユーザーで接続したときに `blog` データベースに対してマイグレーション（CREATE TABLE 等）を実行できるようになります。

---

##### Step 3: GitHub Secrets にマイグレーション用 Secret を設定する

CI（deploy-api ワークフロー）が本番 DB に接続するために、GitHub のリポジトリに Secret を登録します。**MIGRATION_PASSWORD を推奨**（パスワードをそのまま入れればよく、URL エンコード不要）。

###### 3-1. 推奨: MIGRATION_PASSWORD を登録

1. GitHub でリポジトリ（例: `Tattsum/blog`）を開く。
2. **Settings** → **Secrets and variables** → **Actions** を開く。
3. **New repository secret** をクリックする。
4. **Name** に **`MIGRATION_PASSWORD`** と入力する。
5. **Secret** に、**Terraform の `db_root_password` の値をそのまま**貼り付ける（`+` などの特殊文字もそのままでよい。ワークフロー側で URL エンコードする）。
6. **Add secret** で保存する。

###### 3-2. 代替: MIGRATION_DSN を登録

- **Name**: `MIGRATION_DSN`
- **Secret**: `mysql://migrate:パスワード@tcp(127.0.0.1:3306)/blog?parseTime=true`
- パスワードに `+`, `=`, `/`, `?`, `#`, `@` が含まれる場合は URL エンコード（例: `+` → `%2B`）が必要。

###### 3-3. 確認

- `MIGRATION_PASSWORD` と `MIGRATION_DSN` の両方がある場合は、**MIGRATION_PASSWORD** が使われます。
- 設定後、`main` に push するか、Actions から「Deploy API (Cloud Run)」ワークフローを手動実行すると、「Run migrations」ステップでマイグレーションが実行されます。

---

##### このあと（Step 4 以降）

- デプロイ用サービスアカウントに **Cloud SQL Client** ロールが付与されていること（本セクション「3. デプロイ用サービスアカウントに Cloud SQL Client ロールを付与」）を確認する。
- 以上で、CI 上のマイグレーションとデプロイが通る状態になります。

#### 挙動

- `MIGRATION_PASSWORD` も `MIGRATION_DSN` も未設定の場合は「Run migrations」ステップはスキップされ、ビルド・デプロイのみ実行される。
- いずれか設定済みの場合: Auth → Cloud SQL Proxy 起動 → `migrate up` → プロキシ終了 → ビルド・Push → Cloud Run デプロイ、の順で実行される。

---

## 9. 動作確認

### 9.1 やること

- [ ] Cloud Run のサービス URL に `/healthz` を GET して 200 が返ることを確認する
- [ ] Cloudflare Pages の URL を開き、トップ・記事一覧が表示されることを確認する
- [ ] フロントから API（記事一覧・検索など）が呼ばれていることを確認する（ブラウザの開発者ツールのネットワークタブなど）
- [ ] 管理画面（`/admin`）でログインまたは API キー入力ができ、記事一覧・編集ができることを確認する

### 9.2 CORS

- Connect および本プロジェクトの API は、ブラウザからのクロスオリジンリクエストを許可する設定になっている想定です。別ドメインで 403 などが出る場合は、Cloud Run のレスポンスヘッダや Connect の CORS 設定を確認してください。

---

## 10. チェックリスト（一覧）

| # | 項目 | 参照 |
| --- | --- | --- |
| 1 | GCP プロジェクト作成・課金・API 有効化 | 本文 1 |
| 2 | CI 用認証（WIF 推奨 or SA キー） | 本文 2 |
| 3 | Artifact Registry リポジトリ `blog-repo` 作成 | 本文 3 |
| 4 | Cloud SQL インスタンス・DB・マイグレーション | 本文 4 |
| 5 | Secret Manager に `DATABASE_DSN`・`ADMIN_API_KEY` | 本文 5 |
| 6 | Cloud Run へ API デプロイ・サービス URL 確認 | 本文 6 |
| 7 | Cloudflare Pages リポジトリ連携・ビルド設定・環境変数 | 本文 7 |
| 8 | （OIDC 時）ワークフローで WIF 利用に変更 | 本文 8 |
| 9 | 動作確認（/healthz、フロント、管理画面） | 本文 9 |

---

## 11. 参考リンク（2026年3月時点）

- [Cloud Run ドキュメント](https://cloud.google.com/run/docs)
- [Connecting to Cloud SQL from Cloud Run](https://cloud.google.com/sql/docs/mysql/connect-instance-cloud-run)
- [Workload Identity Federation とデプロイパイプライン](https://cloud.google.com/iam/docs/workload-identity-federation-with-deployment-pipelines)
- [Configuring OpenID Connect in GCP (GitHub Docs)](https://docs.github.com/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-google-cloud-platform)
- [Cloudflare Pages - Build configuration](https://developers.cloudflare.com/pages/configuration/build-configuration)
- [Cloudflare Pages - Monorepos](https://developers.cloudflare.com/pages/configuration/monorepos)
- [Deploy Next.js on Cloudflare Pages](https://developers.cloudflare.com/pages/framework-guides/nextjs/deploy-a-nextjs-site)
