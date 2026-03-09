terraform {
  required_version = ">= 1.5.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }

  # 本番では GCS バックエンドを推奨（state の共有・ロック）
  # backend "gcs" {
  #   bucket = "your-tfstate-bucket"
  #   prefix = "blog"
  # }
}
