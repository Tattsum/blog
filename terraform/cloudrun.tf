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

      # PORT は Cloud Run が自動設定するため指定しない
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

      # Vertex AI（Gemini）: AIService が genai SDK で呼び出すときに使用。
      # roles/aiplatform.user は terraform/vertex_ai.tf で付与。
      env {
        name  = "GOOGLE_CLOUD_PROJECT"
        value = var.project_id
      }
      env {
        name  = "GOOGLE_CLOUD_LOCATION"
        value = var.region
      }

      # CORS: ブラウザから別オリジン（tattsum.com 等）で API を呼ぶために必要
      env {
        name  = "CORS_ALLOWED_ORIGINS"
        value = var.cors_allowed_origins
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

# 未認証でアクセス可能にする（公開 API のため）
resource "google_cloud_run_v2_service_iam_member" "public" {
  project  = google_cloud_run_v2_service.blog_api.project
  location = google_cloud_run_v2_service.blog_api.location
  name     = google_cloud_run_v2_service.blog_api.name

  role   = "roles/run.invoker"
  member = "allUsers"
}
