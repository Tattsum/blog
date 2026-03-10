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
