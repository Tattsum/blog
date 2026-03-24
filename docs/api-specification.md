# 個人ブログシステム API 仕様書

本ドキュメントは connect-go の HTTP/JSON 互換モードを前提とし、Protocol Buffers 定義に基づく API の仕様を記述する。

---

## Proto 定義

リポジトリの `proto/blog/v1/` に以下のファイルがある。

- `post.proto` — 記事 (PostService)
- `tag.proto` — タグ (TagService)
- `auth.proto` — 認証 (AuthService)
- `ai.proto` — AI 要約・下書き支援・校正 (AIService)

### post.proto（抜粋・要約）

```proto
syntax = "proto3";
package blog.v1;

message Post {
  string id = 1;
  string title = 2;
  string slug = 3;
  string body_markdown = 4;
  string summary = 5;
  repeated string tag_ids = 6;
  enum Status { STATUS_UNSPECIFIED = 0; DRAFT = 1; PUBLISHED = 2; }
  Status status = 7;
  string created_at = 8;
  string updated_at = 9;
  string published_at = 10;
  string thumbnail_url = 11; // サムネイル画像 URL（任意）
}

service PostService {
  rpc ListPosts(ListPostsRequest) returns (ListPostsResponse) {}
  rpc GetPost(GetPostRequest) returns (GetPostResponse) {}
  rpc CreatePost(CreatePostRequest) returns (CreatePostResponse) {}
  rpc UpdatePost(UpdatePostRequest) returns (UpdatePostResponse) {}
  rpc DeletePost(DeletePostRequest) returns (DeletePostResponse) {}
  rpc SearchPosts(SearchPostsRequest) returns (SearchPostsResponse) {}
  rpc PublishPost(PublishPostRequest) returns (PublishPostResponse) {}
}
```

（リクエスト/レスポンスメッセージ定義は [proto/blog/v1/post.proto](proto/blog/v1/post.proto) を参照。）

---

## 各 RPC メソッド

### PostService

#### ListPosts

- **概要**: 記事一覧をページングで取得する。未認証の場合は `status=published` のみ。管理者は `draft` や未指定（全件）を指定可能。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- | --- |
| リクエスト | page | int32 | 任意 | 1 始まりのページ番号。省略時は 1 |
| リクエスト | page_size | int32 | 任意 | 1 ページあたり件数。省略時 20、最大 100 |
| リクエスト | status | string | 任意 | `"published"` / `"draft"` / `""`（管理者のみ全件） |
| レスポンス | posts | Post[] | 必須 | 記事の配列 |
| レスポンス | total_count | int32 | 必須 | 条件に一致する総件数 |

#### ListPosts — エラーコード

| Code | 条件 |
| --- | --- |
| CodeInvalidArgument | page &lt; 1 または page_size が 0 / 100 超 |
| CodePermissionDenied | 管理者以外が status に draft または空を指定 |

#### ListPosts — リクエスト例（JSON）

```json
{
  "page": 1,
  "pageSize": 20,
  "status": "published"
}
```

#### ListPosts — レスポンス例（JSON）

```json
{
  "posts": [
    {
      "id": "post-01",
      "title": "サンプル記事",
      "slug": "sample-post",
      "bodyMarkdown": "# 本文",
      "summary": "要約文",
      "tagIds": ["tag-1"],
      "status": "PUBLISHED",
      "createdAt": "2025-03-01T00:00:00Z",
      "updatedAt": "2025-03-01T00:00:00Z",
      "publishedAt": "2025-03-01T00:00:00Z"
    }
  ],
  "totalCount": 1
}
```

---

#### GetPost

- **概要**: 記事 ID または slug で 1 件取得。未認証の場合は公開記事のみ。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- | --- |
| リクエスト | id | string | 必須 | 記事 ID または slug |
| レスポンス | post | Post | 必須 | 記事 |

#### GetPost — エラーコード

| Code | 条件 |
| --- | --- |
| CodeInvalidArgument | id が空 |
| CodeNotFound | 該当記事がない、または下書きで権限なし |

#### GetPost — リクエスト例（JSON）

```json
{
  "id": "sample-post"
}
```

#### GetPost — レスポンス例（JSON）

```json
{
  "post": {
    "id": "post-01",
    "title": "サンプル記事",
    "slug": "sample-post",
    "bodyMarkdown": "# 本文",
    "summary": "要約",
    "tagIds": ["tag-1"],
    "status": "PUBLISHED",
    "createdAt": "2025-03-01T00:00:00Z",
    "updatedAt": "2025-03-01T00:00:00Z",
    "publishedAt": "2025-03-01T00:00:00Z"
  }
}
```

---

#### CreatePost

- **概要**: 記事を新規作成。管理者認証必須。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- | --- |
| リクエスト | title | string | 必須 | タイトル |
| リクエスト | slug | string | 任意 | URL 用スラグ。省略時は title から生成。日本語（漢字/ひらがな/カタカナ）が含まれる場合は、選択中の AI プロバイダで英語スラグ化を試み、失敗または利用不可時は unidecode ベースの Slugify にフォールバックして正規化 |
| リクエスト | body_markdown | string | 必須 | 本文（Markdown） |
| リクエスト | summary | string | 任意 | 要約 |
| リクエスト | tag_ids | string[] | 任意 | タグ ID の配列 |
| リクエスト | thumbnail_url | string | 任意 | サムネイル画像の URL（http/https、最大 1024 文字） |
| レスポンス | post | Post | 必須 | 作成された記事（DRAFT） |

#### CreatePost — エラーコード

#### CreatePost — リクエスト例（JSON）

---

#### UpdatePost

- **概要**: 既存記事を更新。管理者認証必須。指定したフィールドのみ更新。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- | --- |
| リクエスト | id | string | 必須 | 記事 ID |
| リクエスト | title | string | 任意 | タイトル |
| リクエスト | slug | string | 任意 | スラグ |
| リクエスト | body_markdown | string | 任意 | 本文 |
| リクエスト | summary | string | 任意 | 要約 |
| リクエスト | tag_ids | string[] | 任意 | タグ ID 配列 |
| リクエスト | thumbnail_url | string | 任意 | サムネイル画像の URL（http/https、最大 1024 文字） |
| レスポンス | post | Post | 必須 | 更新後の記事 |

#### UpdatePost — エラーコード

---

#### DeletePost

- **概要**: 記事を削除。管理者認証必須。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- |
| リクエスト | id | string | 必須 | 記事 ID |
| レスポンス | （なし） | — | — | 空レスポンス |

#### DeletePost — エラーコード

- **概要**: 全文検索で記事一覧を取得。未認証の場合は公開記事のみ対象。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- | --- |
| リクエスト | query | string | 必須 | 検索クエリ |
| リクエスト | page | int32 | 任意 | ページ番号（1 始まり） |
| リクエスト | page_size | int32 | 任意 | 1 ページあたり件数 |
| レスポンス | posts | Post[] | 必須 | 記事配列 |
| レスポンス | total_count | int32 | 必須 | 総件数 |

#### SearchPosts — エラーコード

| Code | 条件 |
| --- | --- |
| CodeInvalidArgument | query が空、または page / page_size が不正 |

#### SearchPosts — リクエスト例（JSON）

```json
{
  "query": "Next.js",
  "page": 1,
  "pageSize": 20
}
```

---

#### PublishPost

- **概要**: 記事を公開または下書きに戻す。管理者認証必須。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- | --- |
| リクエスト | id | string | 必須 | 記事 ID |
| リクエスト | unpublish | bool | 任意 | true で下書きに戻す。省略時は公開 |
| レスポンス | post | Post | 必須 | 更新後の記事 |

#### PublishPost — エラーコード

| Code | 条件 |
| --- | --- |
| CodeUnauthenticated | 未認証 |
| CodeNotFound | 記事が存在しない |

---

### メディアアップロード（POST /upload）

- **概要**: 管理者認証必須の multipart アップロード。画像・動画をストレージに保存し、公開 URL を JSON で返す。サムネイルや本文 Markdown 用の URL 取得に利用する。
- **認証**: `X-Admin-Key` ヘッダー、または `Authorization: Bearer <session_token>` のいずれか。
- **リクエスト**: `Content-Type: multipart/form-data`。フィールド名 `file` で 1 ファイルを送信。
- **許可形式**: 画像（JPEG / PNG / GIF / WebP）、動画（MP4 / WebM）。画像は最大 10MB、動画は最大 100MB。
- **レスポンス**: 成功時 `200`、JSON `{"url": "https://..."}`。失敗時は `4xx` / `5xx` で JSON `{"error": "メッセージ"}`。

---

### TagService

#### ListTags

- **概要**: タグ一覧をページングで取得。認証不要。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- | --- |
| リクエスト | page | int32 | 任意 | ページ番号 |
| リクエスト | page_size | int32 | 任意 | 件数 |
| レスポンス | tags | Tag[] | 必須 | タグ配列 |
| レスポンス | total_count | int32 | 必須 | 総件数 |

**Tag フィールド**: id (string), name (string), slug (string), created_at (string, RFC3339)

#### ListTags — エラーコード

CodeInvalidArgument（page / page_size が不正）

---

#### CreateTag

- **概要**: タグを新規作成。管理者認証必須。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- | --- |
| リクエスト | name | string | 必須 | 表示名 |
| リクエスト | slug | string | 任意 | スラグ。省略時は name から生成（unidecode ベースの Slugify で正規化） |
| レスポンス | tag | Tag | 必須 | 作成されたタグ |

#### CreateTag — エラーコード

CodeUnauthenticated, CodeInvalidArgument, CodeAlreadyExists（同一 slug）

---

#### DeleteTag

- **概要**: タグを削除。管理者認証必須。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- |
| リクエスト | id | string | 必須 | タグ ID |
| レスポンス | （なし） | — | — | 空レスポンス |

#### DeleteTag — エラーコード

CodeUnauthenticated, CodeNotFound

---

### AuthService

- **概要**: 管理者としてログインし、セッショントークンを取得する。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- | --- |
| リクエスト | email | string | 必須 | メールアドレス |
| リクエスト | password | string | 必須 | パスワード |
| レスポンス | session_token | string | 必須 | セッション識別子 |
| レスポンス | expires_at | string | 必須 | 有効期限（RFC3339） |

#### Login — エラーコード

| Code | 条件 |
| --- | --- |
| CodeInvalidArgument | email / password が空 |
| CodeUnauthenticated | 認証失敗（メールまたはパスワード誤り） |

#### Login — リクエスト例（JSON）

```json
{
  "email": "admin@example.com",
  "password": "********"
}
```

#### Login — レスポンス例（JSON）

```json
{
  "sessionToken": "sess_xxxx",
  "expiresAt": "2025-03-02T00:00:00Z"
}
```

---

#### Logout

- **概要**: 現在のセッションを無効化。認証必須。

**リクエスト / レスポンス**: ともに空メッセージ。

#### Logout — エラーコード: CodeUnauthenticated（未認証）

---

#### GetMe

- **概要**: 認証中の管理者情報を取得。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- | --- |
| レスポンス | id | string | 必須 | ユーザー ID |
| レスポンス | email | string | 必須 | メールアドレス |
| レスポンス | display_name | string | 任意 | 表示名 |

#### GetMe — エラーコード: CodeUnauthenticated

---

### AIService

#### Summarize

- **概要**: 指定テキストの要約を Vertex AI (Gemini) で生成。管理者認証必須。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- | --- |
| リクエスト | text | string | 必須 | 要約対象の本文 |
| リクエスト | max_sentences | int32 | 任意 | 要約文の最大文数。省略時 3 |
| レスポンス | summary | string | 必須 | 要約文 |

#### Summarize — エラーコード

| Code | 条件 |
| --- | --- |
| CodeUnauthenticated | 未認証 |
| CodeInvalidArgument | text が空、または max_sentences が不正 |
| CodeResourceExhausted | Vertex AI のレート制限超過 |
| CodeUnavailable | Vertex AI 一時不可 |

#### Summarize — リクエスト例（JSON）

```json
{
  "text": "長い本文テキスト...",
  "maxSentences": 3
}
```

---

#### DraftSupport

- **概要**: 現在の本文とユーザーの指示に基づき、AI が提案本文を生成。管理者認証必須。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- | --- |
| リクエスト | prompt | string | 必須 | ユーザーの指示（例: 「結論を強くして」） |
| リクエスト | current_body | string | 必須 | 現在の本文 |
| レスポンス | suggested_body | string | 必須 | AI が提案する本文 |

#### DraftSupport — エラーコード: CodeUnauthenticated, CodeInvalidArgument, CodeResourceExhausted, CodeUnavailable

#### DraftSupport — リクエスト例（JSON）

```json
{
  "prompt": "結論を強くして",
  "currentBody": "# 記事\n本文..."
}
```

---

#### Proofread

- **概要**: 指定テキストの誤字脱字・表記などの指摘を AI が返す。管理者認証必須。Vertex AI 等が未設定の場合は利用不可。

| 種別 | フィールド | 型 | 必須/任意 | 説明 |
| --- | --- | --- | --- | --- |
| リクエスト | text | string | 必須 | 校正対象テキスト（空不可。長さ上限はサーバーで Unicode ルーン数により約 10 万まで。絵文字や日本語も 1 ルーンとして数える） |
| レスポンス | report | string | 必須 | 指摘内容のレポート（プレーンテキスト） |

#### Proofread — エラーコード

| Code | 条件 |
| --- | --- |
| CodeUnauthenticated / CodePermissionDenied | 未認証または管理者キー・セッション不正 |
| CodeInvalidArgument | text が空、または長すぎる（上限超過時はメッセージに最大ルーン数が含まれる） |
| CodeFailedPrecondition | AI（Vertex 等）未設定、または指定した AI プロバイダが利用不可 |
| CodeResourceExhausted | Vertex AI のレート制限超過 |
| CodeUnavailable | Vertex AI 一時不可 |

#### Proofread — リクエスト例（JSON）

```json
{
  "text": "# タイトル\n\n本文..."
}
```

---

## 共通事項

### 認証方式

- **読者向け API**（ListPosts, GetPost, SearchPosts, ListTags）: 認証不要。公開記事・タグのみ取得可能。
- **管理者向け API**（記事の Create/Update/Delete/Publish、タグの Create/Delete、AuthService, AIService）: セッション認証を前提とする。
  - Login で取得した `session_token` を Cookie または `Authorization` ヘッダー（例: `Bearer <token>`）で送信する。
  - 実装では OIDC 連携またはサーバー側セッション（HTTP-only Cookie）のいずれかを想定。

### エラーハンドリングの共通ルール

- connect-go の [Connect Error Codes](https://connectrpc.com/docs/go/errors/) に準拠する。
- HTTP/JSON モードでは、エラーは HTTP ステータスと JSON ボディで返す（Connect のエラー形式）。
- クライアントは `code`（文字列）と `message` を解釈し、CodeNotFound / CodeInvalidArgument / CodeUnauthenticated 等で分岐する。
- サーバーは 4xx/5xx 時に一貫したエラー形式を返し、本番では内部詳細を露出しない。

### 注意事項・制約

- **HTTP/JSON 互換**: Connect の HTTP/JSON モードを前提とする。フィールド名は JSON では camelCase（例: `tagIds`, `totalCount`）となる。
- **ページング**: `page` は 1 始まり。`page_size` の最大値は 100。指定しない場合はサービス側のデフォルト（例: 20）を用いる。
- **日時**: すべて RFC3339 文字列（例: `2025-03-01T00:00:00Z`）。
- **AI 利用**: Summarize / DraftSupport / Proofread は Vertex AI 等を利用するため、レート制限・入力長制限・利用料に注意する。入力テキストはサニタイズし、出力は必要に応じて検証すること。
- **slug**: 記事・タグの slug は URL パスに用いるため、重複不可・形式制約（英数字とハイフン等）をサーバーで検証すること。
