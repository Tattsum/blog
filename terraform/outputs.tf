output "cloud_run_url" {
  value       = google_cloud_run_v2_service.blog_api.uri
  description = "Cloud Run サービスの URL（NEXT_PUBLIC_API_URL に設定）"
}

output "cloud_sql_connection_name" {
  value       = local.connection_name
  description = "Cloud SQL 接続名（Cloud SQL Auth Proxy や gcloud で使用）"
}

output "db_root_password" {
  value       = var.db_root_password != null ? null : random_password.db_root[0].result
  sensitive   = true
  description = "Terraform で生成した root パスワード（var.db_root_password 未設定時のみ）。マイグレーション実行時に使用"
}
