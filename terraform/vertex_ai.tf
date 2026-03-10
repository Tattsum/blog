# Cloud Run のデフォルト実行 SA（Compute デフォルト）に Vertex AI 利用権限を付与する。
# AIService が Gemini（Vertex）を呼ぶときに必須。未付与だと 403 Permission denied になる。
resource "google_project_iam_member" "blog_api_vertex_ai_user" {
  project = var.project_id
  role    = "roles/aiplatform.user"
  member  = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}
