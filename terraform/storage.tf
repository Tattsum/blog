resource "google_storage_bucket_iam_member" "media_writer" {
  count  = var.media_storage == "gcs" && var.gcs_media_bucket != null && var.gcs_media_bucket != "" ? 1 : 0
  bucket = var.gcs_media_bucket
  role   = "roles/storage.objectCreator"
  member = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}
