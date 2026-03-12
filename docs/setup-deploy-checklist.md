# セットアップ・デプロイ やること一覧（詳細手順）

**対象**: 個人ブログを GCP Cloud Run（API）と Cloudflare Pages（フロント）で本番運用するまでの作業一覧。  
**記載基準**: 2026年3月時点の公式ドキュメントおよび推奨されるやり方に基づく。

関連: [ADR-002（ホスティング方針）](adr/002-hosting-cloudflare-and-cloudrun.md)。

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

### 6.1.1 commit / push 前に手動で GCP（Cloud Run）を更新して確認する

CI を使わず、ローカルでイメージをビルド → Artifact Registry に push → Cloud Run にデプロイ → 動作確認する手順です。

**前提**: `gcloud` でログイン済み（`gcloud auth login`）、デフォルトプロジェクトまたは `GCP_PROJECT_ID` が本番用であること。

1. **環境変数を設定**（リポジトリルートで）:

   ```bash
   export GCP_PROJECT_ID=kano-blog-prod
   export REGION=asia-northeast1
   export CONNECTION_NAME=$GCP_PROJECT_ID:$REGION:blog-mysql
   ```

2. **イメージをビルド**（Cloud Run 用・linux/amd64）:

   ```bash
   make docker-build-api
   ```

   - イメージは `asia-northeast1-docker.pkg.dev/kano-blog-prod/blog-repo/blog-api:latest` にタグ付けされる。

3. **Artifact Registry に push**:

   ```bash
   make docker-push-api
   ```

   - 初回は `gcloud auth configure-docker` が走る。権限エラーなら `gcloud auth login` と Artifact Registry の権限を確認。

4. **Cloud Run にデプロイ**（既存サービス `blog-backend` を更新）:

   ```bash
   gcloud run deploy blog-backend \
     --project=$GCP_PROJECT_ID \
     --region=$REGION \
     --image=$REGION-docker.pkg.dev/$GCP_PROJECT_ID/blog-repo/blog-api:latest \
     --platform=managed \
     --allow-unauthenticated \
     --port=8080 \
     --add-cloudsql-instances=$CONNECTION_NAME \
     --set-secrets=DATABASE_DSN=DATABASE_DSN:latest,ADMIN_API_KEY=ADMIN_API_KEY:latest \
     --set-env-vars=GOOGLE_CLOUD_PROJECT=$GCP_PROJECT_ID,GOOGLE_CLOUD_LOCATION=$REGION
   ```

   - デプロイ完了時に **Service URL** が表示される。

5. **動作確認**（regional URL で `/health` を確認）:

   ```bash
   # プロジェクト番号は describe で確認するか、表示された Service URL から分かる
   curl -sS "https://blog-backend-1098008862560.asia-northeast1.run.app/health"
   # => ok が返れば成功
   ```

   - URL が不明な場合:  
     `gcloud run services describe blog-backend --project=$GCP_PROJECT_ID --region=$REGION --format='value(status.url)'`  
     で取得したドメインの **regional 形式**（`https://blog-backend-PROJECT_NUMBER.asia-northeast1.run.app`）で `/health` を叩く。

ここまで問題なければ、続けて commit / push してよい。

**Terraform state を既存の `blog-backend` に合わせる場合**（state にまだ `blog-api` が残っているとき）:  
既存の `blog-backend` を state に取り込む import 手順は [terraform-import-blog-backend.md](terraform-import-blog-backend.md) を参照。本番では実施済みのため、別環境や state 再構築時のみ参照する。

### 6.1.2 イメージのローカルビルド・実行で確認する（push 前の検証）

デプロイ前に、同じ Docker イメージをローカルで動かして `/health` 等が返ることを確認する手順です。

1. **イメージをビルド**（リポジトリルートで）:

   ```bash
   make docker-build-api-local
   ```

   - タグは `blog-api:local`。`--platform linux/amd64` でビルドする（Cloud Run と同じ）。

2. **コンテナを起動**（別ターミナルでも可）:

   ```bash
   make docker-run-api-local
   ```

   - ポート `8080` で待ち受け。`DATABASE_DSN` 未設定のため RPC ハンドラは登録されず、`/health` と `/healthz` のみ有効。

3. **動作確認**（起動したターミナルとは別のターミナルで）:

   ```bash
   curl -sS http://127.0.0.1:8080/health
   # => ok
   curl -sS -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8080/
   # => 404（アプリの 404。コンテナには届いている）
   ```

   - 停止はコンテナ起動中のターミナルで `Ctrl+C`。

**DB 付きでローカル実行する場合**（Cloud SQL Proxy 等で DB を用意しているとき）:

   ```bash
   docker run --rm -p 8080:8080 \
     -e DATABASE_DSN="mysql://user:pass@tcp(host.docker.internal:3306)/blog?parseTime=true" \
     -e ADMIN_API_KEY="dummy-for-local" \
     blog-api:local
   ```

### 6.2 詳細手順

1. **ローカルでイメージをビルド・push する場合**
   - **Artifact Registry のイメージ名**（例: `blog-api:latest`）と **Cloud Run サービス名**（本番推奨: `blog-backend`）は別。以下の `blog-api` はイメージタグ。
   - リポジトリのルートで実行（コンテキストはルート、Dockerfile は `backend/Dockerfile`）:

   ```bash
   export GCP_PROJECT_ID=your-project-id
   export REGION=asia-northeast1
   docker build -t ${REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/blog-repo/blog-api:latest -f backend/Dockerfile .
   gcloud auth configure-docker ${REGION}-docker.pkg.dev --quiet
   docker push ${REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/blog-repo/blog-api:latest
   ```

2. **Cloud Run にデプロイ（本番推奨: サービス名 `blog-backend`）**
   - **PORT は env に渡さない**（予約名。Cloud Run が自動設定する）。
   - **Cloud SQL の Unix ソケット接続**（推奨）＋ Secret ＋ Vertex 用 env の例:

   ```bash
   export GCP_PROJECT_ID=your-project-id
   export REGION=asia-northeast1
   export CONNECTION_NAME=$GCP_PROJECT_ID:$REGION:blog-mysql

   gcloud run deploy blog-backend \
     --project=$GCP_PROJECT_ID \
     --region=$REGION \
     --image=$REGION-docker.pkg.dev/$GCP_PROJECT_ID/blog-repo/blog-api:latest \
     --platform=managed \
     --allow-unauthenticated \
     --port=8080 \
     --add-cloudsql-instances=$CONNECTION_NAME \
     --set-secrets=DATABASE_DSN=DATABASE_DSN:latest,ADMIN_API_KEY=ADMIN_API_KEY:latest \
     --set-env-vars=GOOGLE_CLOUD_PROJECT=$GCP_PROJECT_ID,GOOGLE_CLOUD_LOCATION=$REGION
   ```

   - デプロイ完了時に **Service URL** が表示される（例: `https://blog-backend-PROJECT_NUMBER.asia-northeast1.run.app`）。
   - `DATABASE_DSN` の Secret 値は、ホストを Unix ソケット用（`/cloudsql/CONNECTION_NAME` 等）にした DSN にしておく。

3. **サービス URL の確認とヘルスチェック**
   - URL 取得:  
     `gcloud run services describe blog-backend --project=$GCP_PROJECT_ID --region=$REGION --format='value(status.url)'`
   - **生存確認**: 本番では **`/health`** を使用する（推奨）。**`/healthz` は Cloud Run で使えない**。Google Cloud の [Known issues - Reserved URL paths](https://cloud.google.com/run/docs/known-issues#reserved_url_paths) によれば、**末尾が `z` で終わるパスは予約**されており使用不可。衝突を避けるため **末尾が `z` のパスは避ける**ことが推奨されている。そのためアプリでは `/health` を用意し、ヘルスチェック・動作確認は `/health` を使う。
     - `curl -sS "https://blog-backend-PROJECT_NUMBER.asia-northeast1.run.app/health"` → **`ok`** が返れば API は生きている。
     - ルート **`/`** で `404 page not found`（text/plain）とアプリのヘッダーが返る場合も、コンテナには届いている。
   - **URL が 2 本ある場合**: サービスによっては **regional**（`....asia-northeast1.run.app`）と **短い方**（`...-an.a.run.app`）の両方が付く。**`/health` や `/` で Google の 404 HTML になるときは、もう一方の URL を試す**。フロントの **`NEXT_PUBLIC_API_URL` は実際に応答が返る方**に合わせる。
   - 取得した URL（末尾スラッシュなし）を **Cloudflare / `wrangler.jsonc` / `frontend/.env.production`** の `NEXT_PUBLIC_API_URL` に設定する。

### 6.3 管理ユーザー（seed）の作成

管理画面（`/admin`）でメール・パスワードでログインするための**初回管理者ユーザー**を 1 件登録する手順です。`backend/cmd/seed` が `users` テーブルに bcrypt ハッシュで INSERT します。

**前提**: Cloud SQL のマイグレーションが完了しており、**migrate ユーザーに `blog`  DB への権限が付与済み**（[8.3 の「migrate に権限付与」](#83-デプロイ時のマイグレーションci-で実行) を参照）であること。

1. **Cloud SQL Auth Proxy を起動する**（別ターミナルで起動したままにする）

   ```bash
   export GCP_PROJECT_ID=your-project-id
   export REGION=asia-northeast1
   export CONNECTION_NAME=$GCP_PROJECT_ID:$REGION:blog-mysql
   # プロキシをインストールしていない場合: https://cloud.google.com/sql/docs/mysql/connect-auth-proxy#install
   cloud-sql-proxy --port 3306 $CONNECTION_NAME
   ```

   - 接続名は `terraform output cloud_sql_connection_name` でも確認可能。

2. **環境変数を設定し、seed を実行する**（リポジトリルートで）

   - **migrate のパスワード**: Terraform の `terraform.tfvars` の **`db_root_password`** と同じ値（Terraform の `google_sql_user.migrate` で同じパスワードを設定している）。
   - **DSN 形式**: go-sql-driver/mysql は **`mysql://` スキーム非対応**のため、**`migrate:パスワード@tcp(127.0.0.1:3306)/blog?parseTime=true`** のように先頭に `mysql://` を付けないこと。付けるとユーザー名が `mysql` と解釈され `Access denied for user 'mysql'@'...'` になる。
   - **パスワードに `+` が含まれる場合**: **シェルからシングルクォートで渡すときは `+` をそのまま**でよい（`export DATABASE_DSN='migrate:パスワード@tcp(...)'`）。`%2B` にするとドライバがデコードせずそのまま送り、認証に失敗することがある。コード内で DSN を組み立てる場合は percent エンコード（`+` → `%2B` 等）を行う。

   ```bash
   export DATABASE_DSN='migrate:ここにdb_root_passwordを入れる@tcp(127.0.0.1:3306)/blog?parseTime=true'
   export SEED_ADMIN_EMAIL='admin@example.com'
   export SEED_ADMIN_PASSWORD='8文字以上のパスワード'
   # 任意: 表示名
   export SEED_ADMIN_DISPLAY_NAME='管理者'

   go run ./backend/cmd/seed
   ```

   - **SEED_ADMIN_PASSWORD** は 8 文字以上必須。既に同じメールのユーザーがいる場合は「user already exists」と表示され、INSERT はスキップされる。
   - 成功時は `created admin user: admin@example.com (id=...)` と表示される。

3. **管理画面でログイン**
   - フロント（本番またはローカル）の `/admin` で、上記のメールとパスワードでログインできる。

### 6.4 Vertex AI（Gemini）で要約・下書き支援を有効にする（任意）

管理画面の「本文から要約を生成」「下書き支援」は、**Cloud Run に `GOOGLE_CLOUD_PROJECT` が入っていれば** Vertex 上の Gemini を使う。未設定のままならローカル要約／プレースホルダにフォールバックする。

- **Terraform 利用時**: `terraform/vertex_ai.tf` で実行 SA に `roles/aiplatform.user` を付与済み。`cloudrun.tf` で `GOOGLE_CLOUD_PROJECT` と `GOOGLE_CLOUD_LOCATION`（リージョン）を env に渡している。`terraform apply` 後に API を再デプロイすれば有効。
- **手動デプロイ時**: 次を追加する。
  - `--set-env-vars=GOOGLE_CLOUD_PROJECT=$GCP_PROJECT_ID,GOOGLE_CLOUD_LOCATION=$REGION`（既存の `--set-env-vars` と結合する場合はカンマ区切りで併記）
  - 実行 SA に `roles/aiplatform.user` を付与:

    ```bash
    gcloud projects add-iam-policy-binding $GCP_PROJECT_ID \
      --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
      --role="roles/aiplatform.user"
    ```

- **モデル変更**: Cloud Run の環境変数 `VERTEX_GEMINI_MODEL`（例: `gemini-2.0-flash-001`）。リージョンによって利用可能モデルが異なる場合がある。

### 6.5 Vertex AI 上の Claude を使う（任意）

Gemini の代わりに **Partner モデル（Claude）** を使う場合:

- Cloud Run の環境変数 **`AI_PROVIDER=vertex-claude`**（または `claude`）を設定する。同一の `GOOGLE_CLOUD_PROJECT` / `GOOGLE_CLOUD_LOCATION` と `roles/aiplatform.user` を利用。
- 任意で **`VERTEX_CLAUDE_MODEL`**（例: SDK 定数に合わせ `claude-sonnet-4-5-20250929`）。未設定時はコード側デフォルトを使用。リージョンによって Model Garden で利用可否が異なる。
- 実装は `anthropic-sdk-go` の Vertex オプション（ADC）。本番では Claude が有効なリージョンを選ぶこと。

### 6.6 メディアアップロードを GCS または R2 で永続化する（推奨）

管理画面からアップロードしたサムネイル・本文画像の URL が **`https://<Cloud Run URL>/uploads/xxx`** の場合、**Cloud Run のコンテナはエフェメラル**なため、再起動・再デプロイでファイルが消え、画像が 404 になります。本番では **GCS または R2 に保存**してください。

#### GCS を選ぶ場合

- **GCS バケットを作成**（例: `blog-media`）:

  ```bash
  gcloud storage buckets create gs://blog-media --location=asia-northeast1
  ```

- **バケットを公開読取にする**（記事・サムネイルを読者が参照するため）:
  - [Making data public](https://cloud.google.com/storage/docs/access-control/making-data-public) に従い、Uniform bucket-level access で `allUsers` に `roles/storage.objectViewer` を付与するか、オブジェクトごとに ACL を設定する。
  - コンソール: バケット → 権限 → プリンシパルに `allUsers`、ロールに「ストレージ オブジェクトの閲覧者」を追加。
- **Cloud Run の実行サービスアカウントに書込権限を付与**:

  ```bash
  gcloud projects add-iam-policy-binding $GCP_PROJECT_ID \
    --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
    --role="roles/storage.objectCreator"
  ```

  - バケット単位で付与する場合は、バケットの IAM で上記 SA に「ストレージ オブジェクト作成者」を付与。
- **Cloud Run に環境変数を追加**:
  - `MEDIA_STORAGE=gcs`
  - `GCS_MEDIA_BUCKET=blog-media`（作成したバケット名）
  - **Terraform で Cloud Run を管理している場合**: [terraform/README.md](../terraform/README.md) を参照し、`terraform.tfvars` に `media_storage = "gcs"` と `gcs_media_bucket = "blog-media"` を設定して `terraform apply` する。
- 再デプロイ後、管理画面からアップロードすると URL は `https://storage.googleapis.com/blog-media/xxx` となり永続します。

#### R2 を選ぶ場合

- **Cloudflare ダッシュボード**で [R2](https://dash.cloudflare.com/?to=/:account/r2) を開き、**Create bucket** でバケットを作成（例: `blog-media`）。
- **R2 API トークン**を発行: **R2** → **Manage R2 API Tokens** → **Create API token**。権限は「Object Read & Write」、バケットは作成したバケットに限定可能。発行後に **Access Key ID** と **Secret Access Key** を控える。
- **パブリックアクセス**のいずれかを設定:
  - **r2.dev サブドメイン**: バケット設定で「Public development URL」を有効にすると Cloudflare が URL（例: `https://pub-xxxx.r2.dev`）を発行。開発・小規模向け。
  - **カスタムドメイン**: 同一 Cloudflare アカウント内のゾーンをバケットに接続（例: `https://media.example.com`）。本番推奨。
- **Cloud Run に環境変数を追加**:
  - `MEDIA_STORAGE=r2`
  - `R2_ACCOUNT_ID`（Cloudflare アカウント ID。ダッシュボードの URL や Overview で確認）
  - `R2_ACCESS_KEY_ID`（上記 API トークンの Access Key ID）
  - `R2_SECRET_ACCESS_KEY`（上記 API トークンの Secret Access Key）
  - `R2_BUCKET`（バケット名、例: `blog-media`）
  - `R2_PUBLIC_BASE_URL`（公開 URL のベース。r2.dev の場合は「Public development URL」の値、カスタムドメインの場合はその URL。末尾スラッシュなし）
  - **Terraform で Cloud Run を管理している場合**: `terraform.tfvars` に `media_storage = "r2"` と R2 用変数（`r2_account_id`, `r2_access_key_id`, `r2_secret_access_key`, `r2_bucket`, `r2_public_base_url`）を設定して `terraform apply` する。例は [terraform/terraform.tfvars.example](../terraform/terraform.tfvars.example) を参照。
- 再デプロイ後、管理画面からアップロードすると、設定した公開 URL ベース + オブジェクトキーで永続します。

#### R2 カスタムドメインで Error 1014（CNAME Cross-User Banned）が出る場合

画像 URL（例: `https://asset.tattsum.com/xxx.png`）にアクセスすると 403 と Error 1014 が返る場合、[Cloudflare 公式](https://developers.cloudflare.com/support/troubleshooting/http-status-codes/cloudflare-1xxx-errors/error-1014/)では次の原因が挙げられています。

1. **ゾーン hold（Zone Hold）**: カスタムドメインのゾーン（例: `tattsum.com`）に [Zone Hold](https://developers.cloudflare.com/fundamentals/account/account-security/zone-holds/) が有効だと、R2 用サブドメインが有効化されず 1014 になる。**対処**: Cloudflare ダッシュボードで該当ゾーン → セキュリティ／アカウント設定から Zone Hold を解除する（Enterprise プランで利用している場合など）。
2. **ゾーンが banned**: フィッシング報告や未払いなどでゾーンが制限されている。**対処**: [Abuse 報告](https://developers.cloudflare.com/fundamentals/reference/report-abuse/complaint-types/)の有無を確認し、必要なら審査依頼。未払いインボイスがあれば [支払い](https://developers.cloudflare.com/billing/pay-invoices-overdue-balances/)を完了する。
3. **別アカウントのゾーン**: CNAME 先が別 Cloudflare アカウントのゾーンの場合、デフォルトでは禁止。**対処**: R2 バケットとカスタムドメインのゾーンが**同一アカウント**であることを確認。別アカウントで運用している場合は [Cloudflare for SaaS](https://developers.cloudflare.com/cloudflare-for-platforms/cloudflare-for-saas/) の利用を検討。

同一アカウントで Zone Hold も問題ない場合でも 1014 が出る場合は、Cloudflare サポートまたはアカウント担当に問い合わせることを推奨します。暫定対応として、`r2_public_base_url` を **r2.dev の Public development URL**（例: `https://pub-xxxx.r2.dev`）に変更すると画像は表示されます（レート制限あり・開発用途向け）。

既に壊れた画像は、該当記事を編集してサムネイルを再アップロードするか、サムネイル欄を空にして保存してください。詳細は [post-thumbnail-and-media.md](post-thumbnail-and-media.md) の「5.5 本番環境（Cloud Run 等）での注意」を参照。

---

## 7. Cloudflare Pages の設定（手動）

Cloudflare Pages は Terraform での管理が難しいため、ここではダッシュボードでの手動設定手順を詳細に記載します。

### 7.1 やること

- [ ] Cloudflare ダッシュボードで Pages プロジェクトを作成し、Git リポジトリと接続する
- [ ] モノレポ用にルートディレクトリ・ビルドコマンド・ビルド出力ディレクトリを設定する
- [ ] **`NEXT_PUBLIC_API_URL`** は **リポジトリ**（`frontend/.env.production` と `frontend/wrangler.jsonc`）で管理する。Cloudflare ダッシュボードの「変数とシークレット」には **設定しない**（リポジトリを正とし、デプロイのたびに変更不要にする）。
- [ ] （任意）カスタムドメインを設定する

### 7.2 前提（事前に用意するもの）

- Cloudflare アカウント（[dash.cloudflare.com](https://dash.cloudflare.com/) でサインアップ可）
- GitHub に本リポジトリ（または fork）が push 済みであること
- **6. Cloud Run** まで完了し、Cloud Run のサービス URL を控えていること（例: `https://blog-backend-PROJECT_NUMBER.asia-northeast1.run.app`）。この URL を後で環境変数に設定します。

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
- **本番用**: `env.production.vars.NEXT_PUBLIC_API_URL` に Cloud Run の **regional URL**（例: `https://blog-backend-PROJECT_NUMBER.asia-northeast1.run.app`）を記載。あわせて Next のビルド時に読まれる **`frontend/.env.production`** にも同じ URL を記載する。**`terraform output cloud_run_url` は短い URL を返すため本番では使わず、`/health` が `ok` になる regional URL を直接記載する。**
- **Cloudflare ダッシュボードの「変数とシークレット」には `NEXT_PUBLIC_API_URL` を設定しない**。設定するとダッシュボードの値が優先され、リポジトリを更新しても反映されない。URL 変更時は `.env.production` と `wrangler.jsonc` の両方を更新し、push して再デプロイする。
- **ドメイン・ルート**: `workers_dev: true` で `blog.<アカウント>.workers.dev`（例: `blog.kurohari35.workers.dev`）を有効化。`routes` にカスタムドメイン（例: `tattsum.com`）を `custom_domain: true` で記載すると、デプロイ時に Cloudflare が DNS と証明書を整備する。プレビュー URL（`*-blog.xxx.workers.dev`）は Git 連携時のブランチごとのプレビュー用で、ダッシュボード側の挙動に依存する。
- ローカルで `npm run preview` や `npm run deploy` を実行するときは、このファイルが参照されます（`npx wrangler deploy` 単体では `.open-next` が無いため失敗します）。

#### Cloudflare ダッシュボードの「変数とシークレット」について

**`NEXT_PUBLIC_API_URL` はダッシュボードに設定しない。** リポジトリの `frontend/.env.production` と `frontend/wrangler.jsonc` の `env.production.vars` を正とし、ビルド・デプロイ時にここが使われる。ダッシュボードに同じ変数があるとそちらが優先され、リポジトリを更新しても反映されず、デプロイのたびに手動で変えなければならなくなる。

- 既にダッシュボードに `NEXT_PUBLIC_API_URL` が登録されている場合は **削除**する（Settings → Environment variables / 変数とシークレット → 該当変数を Remove）。
- URL を変更するときは、`frontend/.env.production` と `frontend/wrangler.jsonc` の両方を更新して push し、再デプロイすればよい。

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

### 7.8 手順 6: ドメインとルート（リポジトリ管理）

本プロジェクトでは **`frontend/wrangler.jsonc`** でドメイン・ルートを管理しています。

- **workers.dev**: `workers_dev: true` により本番は `blog.<アカウント>.workers.dev`（例: `blog.kurohari35.workers.dev`）で公開される。
- **カスタムドメイン**: `routes` に `pattern` と `custom_domain: true` を指定（例: `tattsum.com`）。ドメインのゾーンが Cloudflare にあり、`npm run deploy` でデプロイするとルートと証明書が設定される。
- ドメインを追加・変更する場合は `wrangler.jsonc` の `routes` を編集し、push して再デプロイする。ダッシュボードで手動追加する必要はない。

### 7.9 トラブルシューティング

| 現象 | 確認・対処 |
| --- | --- |
| ビルドが「ルートで実行されている」ようなエラー | Root directory が `frontend` になっているか確認。Set root directory にチェックが入っているか確認。 |
| `npm run build` で Next が見つからない | ルートが `frontend` か確認。`frontend/package.json` に `next` が入っているか確認。 |
| `.open-next/worker.js` was not found | **Deploy command** を **`npm run deploy`** に変更する。`npx wrangler deploy` だけでは OpenNext のビルドが走らず `.open-next` が作られない。 |
| Cloud Run の URL で `curl /healthz` が Google の 404 HTML | **`/healthz` は Cloud Run の予約パス**（[Known issues - Reserved URL paths](https://cloud.google.com/run/docs/known-issues#reserved_url_paths): 末尾 `z` のパスは使用不可）。**`/health`** を使う（アプリで用意済み）。本番は **`blog-backend`** ・**regional URL** を推奨。 |
| 本番で API に接続できない | リポジトリの **`frontend/.env.production`** と **`frontend/wrangler.jsonc`** の `NEXT_PUBLIC_API_URL` が **regional URL**（例: `https://blog-backend-1098008862560.asia-northeast1.run.app`）になっているか確認。末尾スラッシュなし。**Cloudflare ダッシュボードに同名の変数がある場合は削除**（リポジトリを正とする）。変更後は push して再デプロイ。 |
| Node のバージョン不一致 | `frontend/.nvmrc` に `20` または `22` を入れ、再デプロイ。 |
| 403 / CORS エラー | Cloud Run 側で CORS が許可されているか、および `NEXT_PUBLIC_API_URL` が正しいか確認。 |
| **POST /upload が 404** | 環境変数（`MEDIA_STORAGE=r2` 等）が入っていても、**動いているコンテナイメージが古い**と `/upload` は登録されない。**GitHub Actions の「Deploy API (Cloud Run)」ワークフローを手動実行**（Actions タブ → Deploy API → Run workflow）して最新コードでイメージをビルド・デプロイする。ローカルで `make docker-api` して push する場合は、Artifact Registry への `artifactregistry.repositories.uploadArtifacts` 権限が必要。 |
| **R2 カスタムドメインで画像が Error 1014（CNAME Cross-User Banned）** | 上記「6.6 R2 を選ぶ場合」直下の **R2 カスタムドメインで Error 1014** の項を参照。Zone Hold の解除、ゾーン制限（フィッシング報告・未払い）の解消、同一アカウント確認。暫定は `r2_public_base_url` を r2.dev の URL に変更。 |

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

**CI の Cloud Run デプロイの役割**: `gcloud run deploy` では **イメージ・リージョン・allow-unauthenticated のみ**を指定します。Secret（DATABASE_DSN, ADMIN_API_KEY）・Cloud SQL 接続・環境変数（GOOGLE_CLOUD_PROJECT 等）は **Terraform（`terraform/cloudrun.tf`）が唯一のソース・オブ・トゥルース**のため、CI では上書きしません。初回または設定変更時は `terraform apply` で Cloud Run を更新し、以降の push では CI がイメージだけ差し替えます。

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

5. **Cloud Run サービス名を CI と Terraform で揃える**
   - **デプロイ先サービス名**: `deploy-api.yml` は **`gcloud run deploy ${SERVICE_NAME}`** とし、**`SERVICE_NAME`** は GitHub Secret **`CLOUD_RUN_SERVICE_NAME`**（未設定時は **`blog-backend`**）。
   - **`terraform.tfvars`** の **`cloud_run_service_name`** も同じ値（例: `blog-backend`）にする。Artifact Registry のイメージ名（例: `blog-api`）とは別。
   - **CI の deploy はイメージのみ更新**（Secret・Cloud SQL・env は Terraform 管理のため `gcloud run deploy` では指定しない）。
   - **Terraform state がまだ `blog-api` のまま**の場合: `terraform state list | grep cloud_run` で確認。サービスを手動で `blog-backend` に作り直したあと Terraform を合わせるなら、`terraform import` で既存 `blog-backend` を取り込むか、state から古いリソースを外してから `apply` する（運用に合わせて [handover.md](handover.md) を参照）。

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

- [ ] Cloud Run のサービス URL に **`/health`** を GET して 200 と `ok` が返ることを確認する（`/healthz` は Cloud Run で 404 になる場合があるため `/health` を推奨）
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
| 9 | 動作確認（/health、フロント、管理画面） | 本文 9 |

---

## 11. 参考リンク（2026年3月時点）

- [Cloud Run ドキュメント](https://cloud.google.com/run/docs)
- [Cloud Run - Known issues（Reserved URL paths: 末尾 `z` のパスは使用不可）](https://cloud.google.com/run/docs/known-issues#reserved_url_paths)
- [Cloud Run - Configure container health checks](https://cloud.google.com/run/docs/configuring/healthchecks)
- [Connecting to Cloud SQL from Cloud Run](https://cloud.google.com/sql/docs/mysql/connect-instance-cloud-run)
- [Workload Identity Federation とデプロイパイプライン](https://cloud.google.com/iam/docs/workload-identity-federation-with-deployment-pipelines)
- [Configuring OpenID Connect in GCP (GitHub Docs)](https://docs.github.com/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-google-cloud-platform)
- [Cloudflare Pages - Build configuration](https://developers.cloudflare.com/pages/configuration/build-configuration)
- [Cloudflare Pages - Monorepos](https://developers.cloudflare.com/pages/configuration/monorepos)
- [Deploy Next.js on Cloudflare Pages](https://developers.cloudflare.com/pages/framework-guides/nextjs/deploy-a-nextjs-site)
