provider "google" {
  project = var.project_id
  region  = var.region
}

data "google_project" "project" {
  project_id = var.project_id
}

# Cloud SQL root パスワード: 変数が未設定ならランダム生成（state に平文が残るため本番では var を推奨）
resource "random_password" "db_root" {
  count   = var.db_root_password == null ? 1 : 0
  length  = 24
  special = true
}

locals {
  db_root_password = var.db_root_password != null ? var.db_root_password : random_password.db_root[0].result
  connection_name  = "${var.project_id}:${var.region}:${google_sql_database_instance.main.name}"
  # go-sql-driver/mysql の Unix ソケット接続形式
  database_dsn = "root:${local.db_root_password}@unix(/cloudsql/${local.connection_name})/${var.db_name}?parseTime=true"
}
