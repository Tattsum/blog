# blog

個人ブログのモノレポ。Next.js（フロント）と Go + connect-go（API）、MySQL。管理画面から記事の執筆・公開、要約・下書き支援（Vertex AI / 未設定時はローカルフォールバック）。

## ドキュメント

| 内容 | 場所 |
| --- | --- |
| アーキテクチャ・ADR・API 仕様・実装プラン | [docs/](docs/) |
| セットアップ・デプロイ（GCP / Cloudflare） | [docs/setup-deploy-checklist.md](docs/setup-deploy-checklist.md) |
| 引き継ぎ・インフラの現状 | [docs/handover.md](docs/handover.md) |
| エージェント向け（lint / test 必須） | [AGENTS.md](AGENTS.md) |

## 早わかりコマンド

リポジトリルートで:

```bash
make proto          # buf 生成（Go + frontend/src/gen）
make lint && make test
```

バックエンド起動・DB・シードの手順は **docs/setup-deploy-checklist.md** または **docs/handover.md** を参照。

## ディレクトリ

| パス | 説明 |
| --- | --- |
| `backend/` | Go API — 詳細は [backend/internal/README.md](backend/internal/README.md) |
| `frontend/` | Next.js — [frontend/README.md](frontend/README.md) |
| `terraform/` | GCP — [terraform/README.md](terraform/README.md) |
| `proto/` | Connect 定義 |
| `skills/` | Cursor 用 Skill ソース — [skills/README.md](skills/README.md) |

CI は `.github/workflows/ci.yml`。Docker ビルドはルートで `make docker-api`（Makefile 参照）。
