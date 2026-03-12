# 記事のサムネイル・本文内メディア（画像・動画）設計

記事にサムネイル画像を付け、本文 Markdown で画像・動画を配置できるようにするための設計メモ。実装は本ドキュメントを前提に進める。

---

## 1. 要件の整理

| 項目 | 内容 |
| --- | --- |
| **サムネイル画像** | 記事ごとに 1 枚のサムネイルを設定し、一覧・詳細・OGP 等で利用する。 |
| **本文の画像・動画** | 本文は Markdown のまま。画像は Markdown 記法、動画は埋め込み or リンクで配置できるようにする。 |
| **アップローダー** | 管理画面上から画像・動画をアップロードし、取得した URL をサムネイル欄や本文に利用できるようにする。管理画面の操作だけで完結させる。 |

---

## 2. サムネイル画像

### 2.1 方針

- **保存形式**: 記事ごとに **URL を 1 件だけ保持**する方式は変わらない。**アップローダー**により、管理画面から画像をアップロードして取得した URL をそのままサムネイルに設定できるようにする。
- **画像のホスティング**: アップロード先は **Cloudflare R2** を推奨（[4. メディアのホスティング](#4-メディアのホスティング) 参照）。GCS 等も選択可能。アップロード API とストレージの設計は [5. アップローダー](#5-アップローダー) に記載する。

### 2.2 データモデル・API

- **DB**: `posts` に `thumbnail_url VARCHAR(1024) NULL` を追加（migration で対応）。実務的には 1024 で開始し、必要なら 2048 に拡張する。
- **ドメイン**: `Post` に `ThumbnailURL string` を追加。
- **Proto**: 次のフィールドを追加する。
  - `Post`: `string thumbnail_url = 11;`（`published_at = 10` の次）
  - `CreatePostRequest`: `string thumbnail_url = 6;`（`tag_ids = 5` の次）
  - `UpdatePostRequest`: `optional string thumbnail_url = 7;`（`tag_ids = 6` の次）
- **API**: Post を返すすべての RPC（GetPost / ListPosts / SearchPosts / CreatePost / UpdatePost / PublishPost）で `thumbnail_url` を扱う。未設定時は空文字 or 未設定のまま。空文字で「未設定」を表すか、`optional string` で「未設定」と「意図的に空にした」を区別するかは、実装時にフロント・バックエンドで解釈を揃えること。

### 2.3 フロント

- **管理画面**: 記事の新規作成・編集フォームに「サムネイル URL」入力欄を 1 つ追加。任意入力。あわせて **アップローダー**を用意し、ファイル選択→アップロード→取得した URL をサムネイル欄にセット（または手動で URL を貼り付け）できるようにする。
- **一覧・詳細**: `thumbnail_url` が設定されていれば表示（一覧ではカード上部、詳細ではタイトル上 or 直下など）。未設定なら従来どおりテキストのみ。
- **OGP**: 将来的に OGP を生成する場合、`thumbnail_url` を `og:image` に利用する。

### 2.4 セキュリティ・バリデーション

- URL の形式チェック（`http` / `https` のみ許可、スキーム・長さ制限）はバックエンドで行うとよい。詳細は実装時に決める。

---

## 3. 本文の画像・動画（Markdown）

### 3.1 画像

- **現状**: 本文は `body_markdown` で保存し、フロントで `ReactMarkdown` によりレンダリングしている。  
  Markdown の画像記法 `![alt](url)` は **すでに利用可能**。URL は外部ホスト（管理者がアップロードした先）を指定する。
- **アップローダー連携**: 管理画面の本文編集時に、画像をアップロードして取得した URL を Markdown に `![alt](url)` 形式で挿入できるようにする（アップローダーは [5. アップローダー](#5-アップローダー) で定義）。必要に応じて `react-markdown` の `components` で `<img>` に `loading="lazy"` や `decoding="async"` を付与する。

### 3.2 動画

- **方式 1（推奨・まずここ）**: 本文中に **動画の URL をリンク**として記述する（例: `[〇〇の動画](https://...)`）。クリックで遷移。管理画面から動画をアップロードして取得した URL を挿入できるようにする（アップローダーで対応）。
- **方式 2（埋め込み）**: 埋め込み表示したい場合は、Markdown の拡張 or 専用記法を決め、フロントで **URL を iframe に変換**する。
  - 例: `https://www.youtube.com/embed/VIDEO_ID` を検出して `<iframe>` を挿入するカスタムコンポーネントを `react-markdown` に渡す。
  - 対象: YouTube / Vimeo 等。どのドメインを許可するかはホワイトリストで管理する。

### 3.3 実装の優先度

1. **Phase 1**: サムネイル URL の追加（DB・API・管理画面・一覧・詳細）。本文の画像は既存の Markdown のまま利用。サムネイル・本文 URL は手動入力。
2. **Phase 2**: アップローダー（API + ストレージ R2/GCS + 管理画面のアップロード UI）。サムネイル・本文画像・動画を管理画面からアップロードして URL を取得・挿入できるようにする。
3. **Phase 3**: 動画埋め込み（任意）。`react-markdown` のカスタムレンダラで、特定 URL パターンを iframe に変換する。

---

## 4. メディアのホスティング

### 4.1 画質・オリジナルについて

本システムは画像・動画の **URL を保持するのみ** で、ファイルの取得・リサイズ・再エンコードは行わない。したがって、画素数や品質を落とす処理はブログ側にはない。デジタルカメラ（例: Canon EOS Kiss X10）のオリジナル解像度・ファイルサイズのまま利用可能である。実際の配信品質は、メディアを置くホスティング先の設定に依存する。

### 4.2 推奨: Cloudflare R2

- **推奨する構成**: サムネイル・本文画像・動画のファイルは **Cloudflare R2** にホスティングすることを推奨する。
- **理由**:
  - フロントエンドが Cloudflare のため、読者への配信経路が同一でキャッシュ・レイテンシの面で有利。
  - R2 は **エグレス料金が無料** のため、高解像度写真や 1080p 動画など大容量メディアの配信コストを抑えやすい。
  - バックエンドは Google Cloud でも、将来的なアップロード API では S3 互換 API で R2 にアップロード可能。
- 管理者は **アップローダー**により管理画面から R2 にアップロードし、取得した URL をサムネイル欄や本文 Markdown に利用する（Phase 2 で実装）。手動で URL を入力することも可能。

### 4.3 代替: Google Cloud Storage（GCS）

- すべてのインフラを GCP に寄せたい場合は **GCS** を選択してもよい。
- **注意**: ストレージからインターネットへのエグレスは課金対象のため、閲覧量・ファイルサイズが増えると転送コストが増えやすい。必要に応じて **Cloud CDN** を GCS の前に配置し、転送量・コストを抑えることを検討する。

---

## 5. アップローダー

管理画面上から画像・動画をアップロードし、取得した URL をサムネイル欄や本文に挿入できるようにする。管理画面の操作だけで完結させる。

### 5.1 目的

- 管理者が外部ツールや別タブでストレージにアップロードする手間をなくす。
- 記事編集フロー内で「ファイル選択 → アップロード → URL をサムネイル or 本文に反映」まで一貫して行えるようにする。

### 5.2 バックエンド

- **アップロード API**: 管理者認証（X-Admin-Key または Bearer セッション）必須。Multipart でファイルを受信するか、Presigned URL（S3/R2 互換）を発行してフロントから直接ストレージにアップロードする方式のいずれかを採用する。詳細は実装時に決める。
- **ストレージ**: [4. メディアのホスティング](#4-メディアのホスティング) に従い、**Cloudflare R2** を推奨。**R2 実装済み**（`MEDIA_STORAGE=r2` 時）。環境変数は [5.5 本番環境（Cloud Run 等）での注意](#55-本番環境cloud-run-等での注意) を参照。GCS を選ぶ場合は Go の GCS クライアントでアップロードする（`MEDIA_STORAGE=gcs`）。
- **レスポンス**: アップロード成功後、**公開 URL**（サムネイル・本文でそのまま参照できる URL）を返す。R2 の場合はパブリックアクセス用バケットまたはカスタムドメインの URL を返す。

### 5.3 フロント（管理画面）

- **サムネイル**: 記事編集フォームに「ファイルを選択してアップロード」ボタンまたはドラッグ＆ドロップを用意。アップロード完了後に返却された URL をサムネイル URL 欄にセットする。従来どおり URL を手動で入力することも可能。
- **本文**: 本文 Markdown 編集時に、画像・動画をアップロードして返却 URL を取得し、`![alt](url)` または `[表示テキスト](url)` の形式でカーソル位置に挿入する UI を用意する。
- いずれもアップロード中はローディング表示、失敗時はエラーメッセージを表示する。

### 5.4 セキュリティ・制約

- **認証**: アップロード API は管理者のみ利用可能とする（既存の X-Admin-Key または Bearer セッションで保護）。
- **ファイル種別**: 画像（JPEG / PNG / GIF / WebP 等）・動画（MP4 / WebM 等）に限定する。MIME タイプと拡張子の両方を検証する。
- **ファイルサイズ**: 上限を設ける（例: 画像 10MB、動画 100MB）。詳細は実装時に決める。
- **ストレージの公開範囲**: アップロードしたオブジェクトは読者が参照するため、バケットの読み取りは公開（または署名付き URL で配信する方式）とする。書き込みはバックエンドのみに限定する。

### 5.5 本番環境（Cloud Run 等）での注意

- **ローカルストレージ（`UPLOAD_DIR`）はコンテナ内の一時ディスクに保存される**ため、Cloud Run のインスタンス再起動・スケールダウン・再デプロイで **ファイルが消えます**。その結果、`https://<backend>/uploads/xxx.png` のような URL は 404 になり、画像が壊れて表示されます。
- **本番では GCS または R2 を使うこと**:
  - **GCS**: 環境変数 `MEDIA_STORAGE=gcs` と `GCS_MEDIA_BUCKET=<バケット名>` を設定。URL は `https://storage.googleapis.com/<bucket>/<key>` となり永続します。GCS バケットの公開読取設定と、Cloud Run のサービスアカウントにストレージ書込権限を付与する必要があります（[setup-deploy-checklist.md](setup-deploy-checklist.md) 等を参照）。
  - **R2**: 環境変数 `MEDIA_STORAGE=r2` と `R2_ACCOUNT_ID`・`R2_ACCESS_KEY_ID`・`R2_SECRET_ACCESS_KEY`・`R2_BUCKET`・`R2_PUBLIC_BASE_URL` を設定。`R2_PUBLIC_BASE_URL` は r2.dev の Public development URL（例: `https://pub-xxxx.r2.dev`）またはカスタムドメイン（例: `https://media.example.com`）を指定。手順は [setup-deploy-checklist.md](setup-deploy-checklist.md) の「6.6 メディアアップロードを GCS または R2 で永続化する」を参照。
- 既に壊れた画像（DB に保存された `thumbnail_url` がバックエンドの `/uploads/` を指している場合）は、管理画面で該当記事を編集し、サムネイルを再アップロードして GCS/R2 の URL に差し替えるか、サムネイル欄を空にして保存してください。

---

## 6. 関連ドキュメント

- [API 仕様](api-specification.md) — Post のフィールド追加時に更新する。
- [フロントエンドデザイン方針](frontend-design.md) — 一覧・詳細のレイアウトや `.post-body` のスタイル。
- [実装プラン](implementation-plan.md) — 本機能をフェーズとして追記する場合の参照。実装時には実装プランに本機能のフェーズ（サムネイル URL・アップローダー・動画埋め込み）を追記する。
- [メディアファインダー](media-finder.md) — アップロード済み画像の一覧・選択・再利用（将来拡張）の設計。

---

## 7. 実装状況

- **Phase 1（完了）**: サムネイル URL の追加
  - DB: `posts.thumbnail_url VARCHAR(1024) NULL`（migration `000002_add_thumbnail_url_to_posts`）
  - Proto: `Post.thumbnail_url = 11`, `CreatePostRequest.thumbnail_url = 6`, `UpdatePostRequest.optional thumbnail_url = 7`
  - バックエンド: ドメイン・converter・validation・repository・PostServer で Create/Update/Get/List/Search 対応。URL 検証（http/https・最大 1024 文字）
  - フロント: 管理画面の新規・編集にサムネイル URL 入力欄。一覧・詳細・検索・タグ別一覧で `thumbnail_url` があれば表示。管理一覧で 48x48 サムネイル表示
- **Phase 2（完了）**: アップローダー
  - バックエンド: `POST /upload`（multipart `file`）。管理者認証（X-Admin-Key または Bearer セッション）必須。`MediaStorage` 抽象化、ローカル（`UPLOAD_DIR`・任意で `BASE_URL`）と GCS（`MEDIA_STORAGE=gcs`・`GCS_MEDIA_BUCKET`）実装。MIME・拡張子・サイズ検証（画像 10MB・動画 100MB）。成功時 `{"url": "..."}` を返す。ローカル時は `/uploads/` を静的配信
  - フロント: 管理画面の記事新規・編集で「ファイルを選択してアップロード」（サムネイル用）と「画像・動画をアップロードして挿入」（本文用。`![画像](url)` をカーソル位置に挿入）。`uploadMedia()`（admin-api.ts）で X-Admin-Key または Bearer を付与して送信
  - 詳細: [api-specification.md](api-specification.md) の「メディアアップロード（POST /upload）」
- **Phase 3（完了）**: 動画埋め込み
  - 記事詳細（`/posts/[slug]`）で本文 Markdown を `MarkdownBody` コンポーネントでレンダリング。リンクの `href` が埋め込み許可 URL のとき iframe で表示。
  - 許可 URL（ホワイトリスト）: `https://www.youtube.com/embed/*`、`https://youtube.com/embed/*`、`https://www.youtube.com/watch?v=*`（embed に変換）、`https://player.vimeo.com/video/*`。それ以外のリンクは従来どおり `<a>` で新しいタブ表示。
  - 実装: `frontend/src/lib/embed-url.ts`（URL 判定）、`frontend/src/components/MarkdownBody.tsx`（react-markdown の `a` / `img` カスタムコンポーネント）。画像は `loading="lazy"` と `decoding="async"` を付与。
- 今後の拡張として、アップロード済み画像の一覧・選択・再利用（メディアファインダー）を [media-finder.md](media-finder.md) に記載する。

---

## 8. 変更履歴

- 2026-03-12: 初版（サムネイル URL・本文画像は既存 Markdown、動画はリンク or 埋め込み方針を記載）。同日レビュー反映（Proto フィールド番号明記・未設定時扱い・DB 長の注記）。
- 2026-03-12: メディアのホスティングを追記（画質・オリジナル保持の説明、Cloudflare R2 推奨、GCS を代替として記載）。
- 2026-03-12: アップローダーを要件に追加。管理画面から画像・動画をアップロードして操作を完結させる方針と、Phase 2 でアップローダー実装、セクション 5 で詳細を定義。
- 2026-03-12: 全体レビュー反映（認証表記を Bearer セッションに統一、API に SearchPosts / PublishPost を明記、実装プランとの相互参照を追加）。
- 2026-03-12: 実装状況を追記（Phase 1・Phase 2 完了内容と Phase 3 未着手）。
- 2026-03-12: Phase 3 完了（動画埋め込み）。MarkdownBody で YouTube / Vimeo の許可 URL を iframe 表示、実装状況を更新。
- 2026-03-12: 5.5 追加。本番（Cloud Run）でローカルストレージが非永続であることと、GCS 利用・壊れた画像の対処を記載。
- 2026-03-12: R2 実装に合わせて 5.2・5.5 を更新（本番では GCS または R2、R2 用環境変数）。6 に media-finder.md への参照、7 にメディアファインダー拡張の記載を追加。
