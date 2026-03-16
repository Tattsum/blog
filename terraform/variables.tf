variable "project_id" {
  type        = string
  description = "GCP プロジェクト ID"
}

variable "region" {
  type        = string
  default     = "asia-northeast1"
  description = "リージョン（Cloud SQL・Cloud Run などで使用）"
}

# --- Cloud SQL ---
variable "db_tier" {
  type        = string
  default     = "db-f1-micro"
  description = "Cloud SQL のマシンタイプ（開発: db-f1-micro, 本番: db-custom-* など）"
}

variable "db_name" {
  type        = string
  default     = "blog"
  description = "作成するデータベース名"
}

variable "db_root_password" {
  type        = string
  sensitive   = true
  description = "Cloud SQL root ユーザのパスワード（DATABASE_DSN に使用）。未設定時は random_password を使用"
  default     = null
}

# --- Secret Manager ---
variable "admin_api_key" {
  type        = string
  sensitive   = true
  description = "管理 API 用キー（ADMIN_API_KEY シークレットの値）"
}

# --- Cloud Run ---
variable "cloud_run_image" {
  type        = string
  description = "Cloud Run にデプロイするコンテナイメージ（例: asia-northeast1-docker.pkg.dev/PROJECT_ID/blog-repo/blog-api:latest）"
}

variable "cloud_run_service_name" {
  type        = string
  default     = "blog-backend"
  description = "Cloud Run サービス名（本番は blog-backend を推奨。既存 blog-api から移行する場合は tfvars で明示）"
}

variable "cors_allowed_origins" {
  type        = string
  default     = "https://tattsum.com,http://localhost:3000"
  description = "CORS で許可するオリジン（カンマ区切り。フロントのオリジン）"
}

# --- メディアストレージ（アップロード先）---
# 未設定の場合はバックエンドのデフォルト（ローカル＝Cloud Run 上では非永続）。本番では gcs または r2 を推奨。
variable "media_storage" {
  type        = string
  default     = ""
  description = "メディアストレージ種別: gcs / r2 / 空（ローカル）。gcs の場合は gcs_media_bucket を、r2 の場合は r2_* 変数を設定する"
}

variable "gcs_media_bucket" {
  type        = string
  default     = null
  description = "GCS メディア用バケット名（media_storage=gcs のとき必須）。バケットは手動または別リソースで作成すること"
}

variable "gcs_public_base_url" {
  type        = string
  default     = null
  description = "GCS メディアの公開 URL ベース（省略時は storage.googleapis.com/bucket 形式）。Load Balancer + カスタムドメイン（例: https://asset.example.com）で配信するときに指定。末尾スラッシュなし"
}

# R2 用（media_storage=r2 のとき必須）。r2_secret_access_key は機密のため tfvars を .gitignore に含めること
variable "r2_account_id" {
  type        = string
  default     = null
  description = "Cloudflare R2: アカウント ID（media_storage=r2 のとき必須）"
}

variable "r2_access_key_id" {
  type        = string
  default     = null
  sensitive   = true
  description = "Cloudflare R2: API トークン Access Key ID（media_storage=r2 のとき必須）"
}

variable "r2_secret_access_key" {
  type        = string
  default     = null
  sensitive   = true
  description = "Cloudflare R2: API トークン Secret Access Key（media_storage=r2 のとき必須）"
}

variable "r2_bucket" {
  type        = string
  default     = null
  description = "Cloudflare R2: バケット名（media_storage=r2 のとき必須）"
}

variable "r2_public_base_url" {
  type        = string
  default     = null
  description = "Cloudflare R2: 公開 URL ベース（例: https://pub-xxxx.r2.dev）。末尾スラッシュなし（media_storage=r2 のとき必須）"
}
