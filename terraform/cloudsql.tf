resource "google_sql_database_instance" "main" {
  project          = var.project_id
  name             = "blog-mysql"
  database_version = "MYSQL_8_0"
  region           = var.region

  deletion_protection = false

  settings {
    tier = var.db_tier

    ip_configuration {
      ipv4_enabled = true
    }

    backup_configuration {
      enabled            = true
      start_time         = "03:00"
      point_in_time_recovery_enabled = false
    }

    disk_size = 20
    disk_type = "PD_SSD"
  }
}

resource "google_sql_database" "blog" {
  name     = var.db_name
  instance = google_sql_database_instance.main.name
  project  = var.project_id
}

resource "google_sql_user" "migrate" {
  name     = "migrate"
  instance = google_sql_database_instance.main.name
  project  = var.project_id
  host     = "%"
  password = local.db_root_password
}
