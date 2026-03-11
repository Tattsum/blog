# Lint・テスト拡充の検討（2026年3月）

「公開する」ボタンで `unpublish` の真偽が逆に渡っていた不具合をきっかけに、同様のミスを防ぐための Lint 強化とテスト拡充を検討する。

---

## 1. 現状

| 対象 | Lint | テスト | CI |
| --- | --- | --- | --- |
| ルート | markdownlint, buf, golangci-lint | - | ✅ |
| Backend (Go) | golangci-lint | ドメイン・AI・RPC の一部 | ✅ go test, golangci-lint |
| Frontend (Next.js) | ESLint (eslint-config-next) ※CI 未実行 | **Vitest 1本**（公開ボタン） | ビルド + **frontend test** |

- フロントの **ESLint が CI で回っていない**ため、`npm run lint` の失敗が PR で検知されない。
- フロントの **単体テストが存在しない**ため、ボタンと API の対応（例: 公開 → `unpublish: false`）をテストで担保できていない。

---

## 2. 実施・検討項目

### 2.1 実施済み

- **CI で frontend test を実行**  
  `ci.yml` に `cd frontend && npm run test` を追加し、PR/push で Vitest が通ることを必須にする。
- **CI で frontend lint**  
  ESLint 10 と eslint-config-next / eslint-plugin-react の互換性で現状 `npm run lint` が失敗するため、CI には **test のみ**追加。lint は互換性解消後に CI に追加する。
- **Frontend 単体テストの導入**  
  Vitest を導入し、管理画面の「公開する」で `publishPost({ id, unpublish: false })` となるよう `getUnpublishFlagForPublishButton(isPublished)` をテストで担保する（1本目）。
- **本ドキュメント**  
  今後のルール追加・テスト拡充のたたき台とする。

### 2.2 今後の検討

- **ESLint ルールの強化**  
  TypeScript の厳格化（`strict` や `no-floating-promises` など）  
  boolean の取り違えを減らすための命名・コメント規約（Lint で直接「逆だ」と検知するのは難しいため、テストと組み合わせる）
- **Frontend ESLint を CI で実行**  
  ESLint 10 と eslint-plugin-react の互換性を解消し、`cd frontend && npm run lint` を CI に追加する。
- **Backend: PublishPost のテスト**  
  - `post_server_test.go` などで、`unpublish: true` で下書きに戻る／`unpublish: false` で公開になることを RPC レベルでテストする。
- **Frontend: テストの拡充**  
  - ログイン画面、記事一覧のフィルタ、保存／削除ボタンなど、クリックと API 呼び出しの対応をテストで増やす。
- **E2E テスト（任意）**  
  - Playwright 等で「ログイン → 編集 → 公開する → 一覧で公開になる」を流す。

---

## 3. 運用方針

- **PR 時**: ルートの `make lint` と `make test` に加え、**frontend の `npm run test`**（現状 CI では test のみ。lint は互換性解消後に CI 追加予定）を通してからマージする。
- **新機能・修正時**: 「クリックと API の対応」が自明でない箇所は、可能な範囲で単体テストを 1 本ずつ追加する。

---

## 4. 参考

- 不具合の修正: 「公開する」で `handlePublish(!isPublished)` → `handlePublish(isPublished)`（[edit/page.tsx](../../frontend/src/app/admin/posts/[id]/edit/page.tsx)）
- 実装プラン: [implementation-plan.md](implementation-plan.md) フェーズ 2「テスト・エラー共通化（任意）」
