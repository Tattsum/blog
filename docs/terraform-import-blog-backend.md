# Terraform: 既存の blog-backend を state に取り込む（import）

手動で `gcloud run deploy blog-backend` した既存サービスを、Terraform の state に取り込む手順です。**既に本番で import 済みの場合は不要**。別環境や state を失った場合の参考用。

## 前提

- `terraform.tfvars` の `cloud_run_service_name = "blog-backend"` になっている
- Terraform state にはまだ古い `blog-api` が残っている（`terraform plan` で「blog-api を destroy → blog-backend を create」と出る）

## 手順

```bash
cd terraform
terraform state rm google_cloud_run_v2_service.blog_api
terraform state rm google_cloud_run_v2_service_iam_member.public
terraform import google_cloud_run_v2_service.blog_api <PROJECT_ID>/<REGION>/blog-backend
terraform import google_cloud_run_v2_service_iam_member.public "projects/<PROJECT_ID>/locations/<REGION>/services/blog-backend roles/run.invoker allUsers"
terraform plan
```

**例（kano-blog-prod / asia-northeast1）**:

```bash
terraform import google_cloud_run_v2_service.blog_api kano-blog-prod/asia-northeast1/blog-backend
terraform import google_cloud_run_v2_service_iam_member.public "projects/kano-blog-prod/locations/asia-northeast1/services/blog-backend roles/run.invoker allUsers"
```

import 後は `terraform plan` が No changes または軽微な drift のみになる想定。drift があれば `terraform apply` で Terraform の定義に合わせられる。
