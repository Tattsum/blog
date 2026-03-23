# エージェント向けガイド（blog モノレポ）

このリポジトリでコードを変更したエージェント・担当者は、**作業完了前に必ず lint / テストを通す**こと。CI（`.github/workflows/ci.yml`）と同じ水準で確認する。

---

## 一次ソースの調査（対応前の必須）

**2026年現在の一次ソースを十分に調査してから対応すること。**

- 技術仕様・エラー・ベストプラクティスは、**公式ドキュメント・公式リポジトリ・公式 API リファレンス**を優先して確認する。
- 検索結果やサードパーティの記事だけに頼らず、可能な限り **公式（Google Cloud、Cloudflare、Go、Next.js 等）の一次ソース**に当たる。
- エラー文や仕様が不明な場合は、`site:cloud.google.com` や `site:developer.mozilla.org` 等で公式ドメインに絞って検索し、該当する公式ページの記述を根拠に判断する。
- 試行錯誤や推測で対応しない。一次ソースに基づいた設計・修正を行う。

---

## 必須確認（タスク完了前に実行）

以下を**すべて成功**させてからコミット・報告すること。

### 1. Lint（CI と同等）

リポジトリルートで:

```bash
npm run lint:md
npm run lint:proto
npm run generate:proto
golangci-lint run ./...
```

または Makefile 一括（`lint:go` は `golangci-lint run ./...`）:

```bash
make lint
```

### 2. テスト

```bash
make test
# または
go test ./... -count=1
```

### 3. フロントビルド（CI で実行されているため、フロント変更時は必須）

```bash
cd frontend && npm ci && npm run build
```

フロントを触らない場合でも、念のため `frontend` で `npm run build` が通るか確認すると安全。

---

## 一発で CI に近づけるコマンド例

```bash
cd /path/to/blog
npm run lint:md && npm run lint:proto && npm run generate:proto && \
  go test ./... -count=1 && golangci-lint run ./... && \
  (cd frontend && npm run build)
```

`make lint` は `npm run lint:go` を含むが、CI では `generate:proto` の後に `go test` が走る。**proto を変えた場合は `make proto` または `npm run generate:proto` の後にテストすること。**

---

## よくある失敗

- **`markdownlint` が docs 内でエラー** … 見出しは `###` 等を使う。番号付きリストは MD029 に注意（連番を崩すなら箇条書きにする）。
- **`buf lint`** … `proto/` 修正後は `buf lint` で確認。
- **`golangci-lint`** … `backend/` 変更後は必ず実行。
- **フロントビルド失敗** … `frontend` で `npm ci` 済みか、`gen/` の TypeScript がコミットされているか確認。

---

## 関連ドキュメント

- [docs/handover.md](docs/handover.md) … 引き継ぎ・インフラ状態
- [docs/implementation-plan.md](docs/implementation-plan.md) … 実装フェーズ
- [Makefile](Makefile) … `make lint` / `make test`

---

## セキュリティ・パフォーマンス（常に考慮）

- **セキュリティ**: 認証・認可の抜けがないか、入力の検証・サニタイズ、機密情報の露出防止、依存関係の脆弱性確認を行う。外部入力を信頼しない。
- **パフォーマンス**: N+1 の回避、キャッシュ・ページングの検討、レスポンスサイズとクエリコストの意識。変更後は遅延やリソース使用の影響を確認してから完了とする。
- 詳細は [.cursorrules](.cursorrules) および [.cursor/rules/security-and-performance.mdc](.cursor/rules/security-and-performance.mdc) を参照。

---

## ルールの要約

- **対応前に 2026 年現在の一次ソース（公式ドキュメント・API・リポジトリ）を十分に調査し、その根拠に基づいて対応する。試行錯誤で対応しない。**
- **セキュリティとパフォーマンスを常に考慮する。** 認証・入力検証・情報漏れ防止・N+1 回避・レスポンス負荷を確認してから完了とする。
- **lint / test /（フロント変更時は build）を通さずに「完了」としない**
- **Go のテスト規約**: `*_test.go` は原則として外部パッケージである `package <pkg>_test` を使用する（必要ならテスト用の薄いラッパーを追加する）。
- 失敗した場合はログを確認し、修正してから再度実行する
