---
name: blog-lint-and-test
description: blog モノレポでコード変更後、必ず lint・テスト・フロントビルドを通す。CI と同水準で完了報告する前に実行する。make lint / make test に失敗したら修正してから終了しない。
---

# blog — Lint とテストを必ず通す

このリポジトリでは **コミット・タスク完了の前に** lint と test を実行すること。失敗のまま「完了」としない。

## いつ使うか

- backend / frontend / docs / proto を変更したあと
- PR や push の前
- エージェントが「対応しました」と言う直前

## 手順（リポジトリルート）

### 1. Lint

```bash
make lint
```

内訳: `npm run lint:md` → `buf lint` → `npm run lint:go`（golangci-lint）。

### 2. テスト

```bash
make test
```

### 3. フロントビルド

`frontend/` を変えた場合、または全体確認として:

```bash
cd frontend && npm run build
```

## 一括（CI に近い）

```bash
npm run lint:md && npm run lint:proto && npm run generate:proto && \
  go test ./... -count=1 && golangci-lint run ./... && \
  (cd frontend && npm run build)
```

## 参照

- 詳細・トラブル: リポジトリルートの `AGENTS.md`
- CI 定義: `.github/workflows/ci.yml`
