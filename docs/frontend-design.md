# フロントエンドデザイン方針

フロントの見た目とテーマの指針をまとめたドキュメントです。

---

## 1. デザインの方向性（Zenn 風）

- **参考**: [Zenn](https://zenn.dev) のような、読みやすさを優先したブログ・技術記事向けの見た目を目指す。
- **方針**:
  - コンテンツの可読性を最優先にする。
  - 余白とタイポグラフィ（行間・フォントサイズ）を整え、長文でも負担が少ないようにする。
  - 装飾は控えめにし、記事一覧・記事詳細・タグ一覧などで一貫したトーンを保つ。
- **実装**: `frontend/src/app/globals.css` で CSS 変数（`--background`, `--foreground`, `--muted`, `--border`, `--link`, `--code-bg` など）を定義し、各ページ・コンポーネントでこれを参照する。

---

## 2. ダークモード対応

- **方式**: `prefers-color-scheme: dark` に従うメディアクエリで、ライト／ダークのテーマを切り替える。
- **実装**:
  - `:root` にライト用の変数、`@media (prefers-color-scheme: dark)` 内の `:root` にダーク用の変数を定義。
  - 色はハードコードせず、`var(--foreground)` や `var(--muted)` などの変数で指定する。
- **今後の拡張**: ユーザーが手動でライト／ダークを切り替えたい場合は、`class` や `data-theme` を `<html>` に付与し、そのクラスに応じて変数を上書きする方式を検討する。

---

## 3. レイアウト・コンポーネント

- **コンテナ**: 本文幅は最大 720px 程度に制限し、中央寄せ。`.container` クラスで統一。
- **記事一覧**: 一覧項目は `.article-list` で区切り線と余白を統一。タイトルリンクは `.title` でホバー時にアクセント色にする。
- **コードブロック**: `.post-body pre` / `.post-body code` でモノスペースと背景色（`--code-bg`）を適用し、ダークモードでも読みやすくする。

---

## 4. Edge 環境での fetch redirect エラー対応

Cloudflare Workers（Edge）では、`fetch` の `redirect` オプションに `"error"` を指定すると **「Invalid redirect value, must be one of 'follow' or 'manual'」** が発生する。

- **対応（二重対策）**:
  1. **Connect トランスポート**: `frontend/src/lib/edge-safe-fetch.ts` で `redirect: "error"` を `"follow"` に正規化する `edgeSafeFetch` を定義し、`api.ts` と `admin-api.ts` の `createConnectTransport` に `fetch: edgeSafeFetch` を渡す。API 呼び出しはこちらで確実にカバーされる。
  2. **グローバル fetch**: `frontend/src/instrumentation.ts` で Edge ランタイム時のみ `instrumentation-edge.ts` を読み込み、グローバルな `fetch` をラップする。Next や他ライブラリが `redirect: "error"` を使う場合の保険。
- **参照**: [Cloudflare Workers Request](https://developers.cloudflare.com/workers/runtime-apis/request/)、[OpenNext troubleshooting](https://opennext.js.org/cloudflare/troubleshooting)。

---

## 5. 関連ファイル

| ファイル | 役割 |
| --- | --- |
| `frontend/src/app/globals.css` | テーマ変数・コンテナ・記事一覧・post-body のスタイル |
| `frontend/src/app/layout.tsx` | フォント変数・`suppressHydrationWarning`（テーマのちらつき軽減） |
| `frontend/src/instrumentation.ts` | Edge 時のみ fetch の redirect を正規化するモジュールを読み込み |
| `frontend/src/instrumentation-edge.ts` | fetch の `redirect: "error"` を `"follow"` に置き換えるラッパー（グローバル） |
| `frontend/src/lib/edge-safe-fetch.ts` | Connect 用に `redirect` を正規化する fetch（トランスポートに渡す） |

以上を踏まえ、新規ページやコンポーネントを追加する際は、色指定に CSS 変数を使い、必要に応じて `.container` や `.article-list` を利用すること。
