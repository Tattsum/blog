# 不要なものの削除・整理（作業メモ）

**実施済み（2026年3月）** … GCP の不要サービス削除とドキュメント整理を完了。本ファイルは記録として残す。

---

## 1. GCP 上で削除してよいリソース（手動） — 実施済み

以下を削除済み。本番で使用する Cloud Run サービスは **`blog-backend`** のみ。

- **blog-api** … 削除済み（Terraform state からも外れていた旧サービス）
- **blog-api-hello** … 削除済み（動作確認用に作成した hello イメージ）
- **blog-backend-test** … 削除済み（トラブルシュート用に作成したテストサービス）

今後、同様のテスト用サービスを作成した場合は `gcloud run services list` で確認し、不要なら `gcloud run services delete <名前> --project=... --region=...` で削除する。

---

## 2. リポジトリ内のドキュメント整理 — 実施済み

- **Terraform import 手順**: [terraform-import-blog-backend.md](terraform-import-blog-backend.md) に切り出し、[setup-deploy-checklist.md](setup-deploy-checklist.md) の 6.1.1 ではリンクのみに変更済み。
- **削除していないもの**: AGENTS.md / agent.md / .cursorrules、terraform.tfvars.example / .envrc.example、backend の `*_test.go`、skills/ はそのまま利用。
