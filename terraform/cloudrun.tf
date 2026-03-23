resource "google_cloud_run_v2_service" "blog_api" {
  name     = var.cloud_run_service_name
  project  = var.project_id
  location = var.region

  template {
    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }

    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [local.connection_name]
      }
    }

    containers {
      image = var.cloud_run_image

      volume_mounts {
        name       = "cloudsql"
        mount_path = "/cloudsql"
      }

      env {
        name = "DATABASE_DSN"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.database_dsn.id
            version = "latest"
          }
        }
      }

      env {
        name = "ADMIN_API_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.admin_api_key.id
            version = "latest"
          }
        }
      }

      env {
        name  = "GOOGLE_CLOUD_PROJECT"
        value = var.project_id
      }
      env {
        name  = "GOOGLE_CLOUD_LOCATION"
        value = "us-central1"
      }
      env {
        name  = "AI_PROVIDER"
        value = var.ai_provider
      }
      env {
        name  = "VERTEX_GEMINI_MODEL"
        value = var.vertex_gemini_model
      }
      dynamic "env" {
        for_each = var.vertex_claude_model != "" ? [1] : []
        content {
          name  = "VERTEX_CLAUDE_MODEL"
          value = var.vertex_claude_model
        }
      }

      env {
        name  = "CORS_ALLOWED_ORIGINS"
        value = var.cors_allowed_origins
      }

      dynamic "env" {
        for_each = var.media_storage == "gcs" && var.gcs_media_bucket != null && var.gcs_media_bucket != "" ? [1] : []
        content {
          name  = "MEDIA_STORAGE"
          value = "gcs"
        }
      }
      dynamic "env" {
        for_each = var.media_storage == "gcs" && var.gcs_media_bucket != null && var.gcs_media_bucket != "" ? [1] : []
        content {
          name  = "GCS_MEDIA_BUCKET"
          value = var.gcs_media_bucket
        }
      }
      dynamic "env" {
        for_each = var.media_storage == "gcs" && var.gcs_media_bucket != null && var.gcs_media_bucket != "" && var.gcs_public_base_url != null && var.gcs_public_base_url != "" ? [1] : []
        content {
          name  = "GCS_PUBLIC_BASE_URL"
          value = var.gcs_public_base_url
        }
      }

      dynamic "env" {
        for_each = local.use_r2 ? [1] : []
        content {
          name  = "MEDIA_STORAGE"
          value = "r2"
        }
      }
      dynamic "env" {
        for_each = local.use_r2 ? [1] : []
        content {
          name  = "R2_ACCOUNT_ID"
          value = var.r2_account_id
        }
      }
      dynamic "env" {
        for_each = local.use_r2 ? [1] : []
        content {
          name  = "R2_ACCESS_KEY_ID"
          value = var.r2_access_key_id
        }
      }
      dynamic "env" {
        for_each = local.use_r2 ? [1] : []
        content {
          name  = "R2_SECRET_ACCESS_KEY"
          value = var.r2_secret_access_key
        }
      }
      dynamic "env" {
        for_each = local.use_r2 ? [1] : []
        content {
          name  = "R2_BUCKET"
          value = var.r2_bucket
        }
      }
      dynamic "env" {
        for_each = local.use_r2 ? [1] : []
        content {
          name  = "R2_PUBLIC_BASE_URL"
          value = var.r2_public_base_url
        }
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }
}

resource "google_cloud_run_v2_service_iam_member" "public" {
  project  = google_cloud_run_v2_service.blog_api.project
  location = google_cloud_run_v2_service.blog_api.location
  name     = google_cloud_run_v2_service.blog_api.name

  role   = "roles/run.invoker"
  member = "allUsers"
}
