provider "google" {
  project = var.project_id
  region  = var.region
}

data "google_project" "project" {
  project_id = var.project_id
}

resource "random_password" "db_root" {
  count   = var.db_root_password == null ? 1 : 0
  length  = 24
  special = true
}

locals {
  db_root_password = var.db_root_password != null ? var.db_root_password : random_password.db_root[0].result
  connection_name  = "${var.project_id}:${var.region}:${google_sql_database_instance.main.name}"
  database_dsn = "root:${local.db_root_password}@unix(/cloudsql/${local.connection_name})/${var.db_name}?parseTime=true"
  use_r2 = var.media_storage == "r2" && var.r2_account_id != null && var.r2_access_key_id != null && var.r2_secret_access_key != null && var.r2_bucket != null && var.r2_public_base_url != null
}
