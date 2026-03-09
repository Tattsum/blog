# Secret Manager: シークレットリソースと初回バージョン
# 値は Terraform の variable から注入（tfvars は .gitignore 推奨）

resource "google_secret_manager_secret" "database_dsn" {
  secret_id = "DATABASE_DSN"
  project   = var.project_id

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "database_dsn" {
  secret      = google_secret_manager_secret.database_dsn.id
  secret_data = local.database_dsn
}

resource "google_secret_manager_secret" "admin_api_key" {
  secret_id = "ADMIN_API_KEY"
  project   = var.project_id

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "admin_api_key" {
  secret      = google_secret_manager_secret.admin_api_key.id
  secret_data = var.admin_api_key
}

# Cloud Run の実行用サービスアカウント（デフォルトの Compute SA）に Secret 参照権限を付与
resource "google_secret_manager_secret_iam_member" "database_dsn" {
  secret_id = google_secret_manager_secret.database_dsn.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}

resource "google_secret_manager_secret_iam_member" "admin_api_key" {
  secret_id = google_secret_manager_secret.admin_api_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}
