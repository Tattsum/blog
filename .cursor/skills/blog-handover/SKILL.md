---
name: blog-handover
description: blog モノレポの続きの作業・引き継ぎ。インフラ状態、Cloudflare、Cloud Run、次のタスクを把握するときに docs/handover.md と implementation-plan を読む。
---

# blog — 引き継ぎと続きの作業

新規エージェントや担当が「いまの状態と次にやること」を把握するための skill。

## 必ず読む

1. **`docs/handover.md`** … GCP / Cloudflare の状態、済んだ手動項目、つまずきポイント
2. **`docs/implementation-plan.md`** … フェーズ進捗、AI・Vertex・セキュリティの残り

## よく触るパス

- **デプロイ手順** … `docs/setup-deploy-checklist.md`
- **Edge fetch redirect** … `docs/frontend-design.md`
- **AI 複数モデル方針** … `docs/ai-model-providers.md`
- **API 入口** … `backend/cmd/server/main.go`
- **Makefile** … `make docker-api` / `make lint` / `make test`

## Cloud Run / Vertex

- Terraform で `GOOGLE_CLOUD_PROJECT` / `LOCATION` が注入され、`roles/aiplatform.user` が付いていれば AIService が Gemini を利用
- 未設定時はローカル要約にフォールバック

## 作業完了前

`blog-lint-and-test` skill に従い **lint / test を通してから** 報告すること。
