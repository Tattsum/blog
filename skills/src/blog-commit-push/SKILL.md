---
name: blog-commit-push
description: blog モノレポで変更を git commit しリモートへ push する。タスク完了後やユーザーが「commit」「push」「コミットして」と依頼したときに使う。コミット前に blog-lint-and-test に従い lint・テストを通す。
---

# blog — Commit と Push

変更を **コミット** して **リモートへ push** する手順。コミット・push の依頼や、タスク完了後に使う。

## いつ使うか

- ユーザーが「commit して」「push して」「コミット・push して」と依頼したとき
- タスク完了後、変更をリポジトリに反映したいとき
- コードやドキュメントを変更し、lint・テストを通したあと

## 前提

コミット前に [blog-lint-and-test](../blog-lint-and-test/SKILL.md) に従い、`make lint` と `make test`（フロント変更時は `npm run build`）を成功させること。

## 手順（リポジトリルート）

### 1. 変更確認

```bash
git status
```

コミット対象のファイルを確認する。

### 2. ステージング

```bash
git add <ファイルまたはディレクトリ>
```

例: 特定ファイルだけ `git add path/to/file.go`、今回の変更すべて `git add -A` または `git add .`

### 3. コミット

```bash
git commit -m "種類: 簡潔な説明

- 箇条書きで変更内容を書いてもよい（任意）"
```

**メッセージの慣例**:
- 先頭に種類: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:` など
- 1 行目は 50 字程度で要約。詳細は 2 行目以降に。

### 4. プッシュ

```bash
git push
```

リモートが `origin` でブランチが `main` なら `git push origin main` または `git push` でよい。

## 権限

- `git commit`: リポジトリの `git_write` 権限が必要な環境では有効化する。
- `git push`: ネットワーク権限が必要。リモート認証（SSH やトークン）はユーザー環境に依存する。

## 参照

- コミット前の必須確認: [blog-lint-and-test](../blog-lint-and-test/SKILL.md)、[AGENTS.md](../../AGENTS.md)
