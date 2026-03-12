# GCS メディアバケット: Cloud Run のデフォルト SA にオブジェクト作成権限を付与
# バケットは手動または別手順で作成し、公開読取設定も別途行う（setup-deploy-checklist 参照）
resource "google_storage_bucket_iam_member" "media_writer" {
  count  = var.media_storage == "gcs" && var.gcs_media_bucket != null && var.gcs_media_bucket != "" ? 1 : 0
  bucket = var.gcs_media_bucket
  role   = "roles/storage.objectCreator"
  member = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}
