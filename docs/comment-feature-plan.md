# コメント機能 設計・実装プラン

公開記事に対する読者コメント機能を、既存の Clean Architecture（domain → application/repository IF → infrastructure → interface/rpc）に乗せて追加するための設計メモ。実装は本ドキュメントを前提に進める。

- **対象リポジトリ**: `blog`
- **前提ドキュメント**: [architecture.md](architecture.md) / [api-specification.md](api-specification.md) / [implementation-plan.md](implementation-plan.md)（フェーズ 6 で「コメント」は計画のみ）

---

## 1. 要件の整理

| 項目 | 内容 |
| --- | --- |
| **投稿対象** | 公開済み（`PUBLISHED`）記事に対してのみ匿名コメント可。 |
| **モデレーション** | 投稿直後は `PENDING`（非公開）。管理者が承認して `APPROVED` で初めて公開表示する。 |
| **入力項目** | 名前（必須）・本文（必須）・メールアドレス（任意・非表示で保存）。 |
| **表示** | 記事ページ下部に承認済みコメントを表示。本文はプレーンテキストとして描画する。 |
| **管理** | 管理画面で `PENDING` 一覧を確認し、承認・却下・削除を行う。 |

---

## 2. セキュリティ方針

| 観点 | 対策 |
| --- | --- |
| **XSS** | コメント本文はプレーンテキスト固定。記事本文の `MarkdownBody` と異なり `dangerouslySetInnerHTML` を一切通さず、React の自動エスケープで描画する。 |
| **スパム（bot）** | ハニーポット欄（`website`）を用意し、値が入っていたら成功を装って静かに破棄する。 |
| **スパム（量）** | IP 単位のインメモリ・トークンバケットでレート制限（例: 1 IP あたり 1 分 5 件・1 時間 20 件）。超過時は `ResourceExhausted`。 |
| **プライバシー** | `author_email` は保存するが公開レスポンスでは必ず空にし、管理用レスポンスでのみ実値を返す。 |
| **入力検証** | 名前 1〜80 字、本文 1〜2000 字、制御文字除去。メールは任意（入力時のみ形式・255 字を検証）。対象記事が存在し公開済みであることを確認する。 |

---

## 3. ドメイン層 `backend/internal/domain/comment/comment.go`

```go
package comment

type Status int32

const (
    StatusUnspecified Status = 0
    StatusPending     Status = 1 // 投稿直後（未公開）
    StatusApproved    Status = 2 // 承認済み（公開表示）
    StatusRejected    Status = 3 // 却下（スパム含む）
)

type Comment struct {
    ID          string
    PostID      string
    AuthorName  string
    AuthorEmail string // 空文字 = 未入力。公開レスポンスには含めない
    Body        string // プレーンテキスト（Markdown/HTML 不可）
    Status      Status
    CreatedAt   time.Time
}

func (c *Comment) IsVisible() bool { return c.Status == StatusApproved }
func (c *Comment) Approve()        { c.Status = StatusApproved }
func (c *Comment) Reject()         { c.Status = StatusRejected }
```

---

## 4. リポジトリ IF（`backend/internal/domain/repository/repository.go` に追記）

```go
type ListCommentsFilter struct {
    PostID   string
    Status   comment.Status // 0 = 全件（管理用）, それ以外で絞り込み
    Page     int32
    PageSize int32
}

type CommentRepository interface {
    Create(ctx context.Context, c *comment.Comment) error
    GetByID(ctx context.Context, id string) (*comment.Comment, error)
    List(ctx context.Context, filter ListCommentsFilter) ([]*comment.Comment, int64, error)
    Update(ctx context.Context, c *comment.Comment) error // 状態遷移の保存
    Delete(ctx context.Context, id string) error
}
```

命名は既存規約に準拠（複数返却は `List`、存在前提取得は `GetByID`、`Exists` は足さない）。

---

## 5. DB マイグレーション `backend/db/migrations/000003_create_comments`

```sql
-- up
CREATE TABLE comments (
    id VARCHAR(36) PRIMARY KEY,
    post_id VARCHAR(36) NOT NULL,
    author_name VARCHAR(80) NOT NULL,
    author_email VARCHAR(255) NULL,
    body VARCHAR(2000) NOT NULL,
    status TINYINT NOT NULL DEFAULT 1 COMMENT '0=unspecified, 1=pending, 2=approved, 3=rejected',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    INDEX idx_comments_post_status (post_id, status),
    CONSTRAINT fk_comments_post FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

```sql
-- down
DROP TABLE IF EXISTS comments;
```

- 複合 INDEX `(post_id, status)`: 公開表示クエリ `WHERE post_id=? AND status=2` をカバー（カーディナリティの高い `post_id` を先頭）。
- `DATETIME(6)` / `utf8mb4_unicode_ci` / PK は `VARCHAR(36)` と、既存テーブルの規約に統一。
- `author_email` は外部入力のため可変長 `VARCHAR(255)`。未入力と区別するため `NULL` 許容。
- **FK + ON DELETE CASCADE**: グローバル規約では FK 非推奨だが、本リポジトリは既存 `post_tags` が FK CASCADE を採用済みのため、リポジトリ内一貫性を優先して既存に揃える。記事削除時はコメントも DB 側で自動削除される。

---

## 6. proto `proto/blog/v1/comment.proto`

```proto
message Comment {
  string id = 1;
  string post_id = 2;
  string author_name = 3;
  string author_email = 4; // 公開レスポンスでは空。管理用レスポンスのみ実値
  string body = 5;
  string status = 6; // "PENDING" / "APPROVED" / "REJECTED"（既存 ListPosts と同じ文字列ステータス流儀）
  string created_at = 7;
}

service CommentService {
  rpc ListComments(ListCommentsRequest) returns (ListCommentsResponse) {}                                  // 公開: 承認済みのみ
  rpc CreateComment(CreateCommentRequest) returns (CreateCommentResponse) {}                                // 公開: PENDING で作成
  rpc ListCommentsForModeration(ListCommentsForModerationRequest) returns (ListCommentsResponse) {}         // 管理: status 絞り込み
  rpc ModerateComment(ModerateCommentRequest) returns (ModerateCommentResponse) {}                          // 管理: 承認/却下
  rpc DeleteComment(DeleteCommentRequest) returns (DeleteCommentResponse) {}                                // 管理: 削除
}
```

- `CreateCommentRequest`: `post_id` / `author_name` / `body` / `optional string author_email` / `string website`（ハニーポット）。
- 生成は `make proto`（Go + `frontend/src/gen`）。

---

## 7. interface/rpc `backend/internal/interface/rpc/comment_server.go`

- `ListComments`（公開）: `filter.Status = StatusApproved` 固定。認証不要。converter で `author_email` を空にする。
- `CreateComment`（公開・認証なし）:
  - ハニーポット `website` が非空 → 成功レスポンスを返しつつ破棄。
  - レート制限（IP 単位のインメモリ・トークンバケット、`rpc` 層に新設）。
  - `validateCommentFields`（`validation.go` に追加）で入力検証。対象記事を `GetByID` で取得し公開済みか確認。
  - `Status = PENDING` 固定で作成。
- `ListCommentsForModeration` / `ModerateComment` / `DeleteComment`: 先頭で `requireAdminOrSession(...)`。`ModerateComment` は `Approve()/Reject()` 後に `Update`。
- repo エラーは既存 `MapHandlerError` で正規化。

---

## 8. 配線 `backend/cmd/server/main.go`

```go
commentRepo := mysql.NewCommentRepository(db)
commentPath, commentHandler := blogv1connect.NewCommentServiceHandler(
    rpc.NewCommentServer(commentRepo, postRepo, adminKey, sessionStore, rateLimiter),
)
mux.Handle(commentPath, commentHandler)
```

---

## 9. フロントエンド

- **記事ページ** `frontend/src/app/posts/[slug]/page.tsx` 下部:
  - `CommentList`（Server Component）: `commentClient.listComments({ postId })` で承認済みを表示。本文は `{c.body}` を JSX 直挿し（自動エスケープ）。
  - `CommentForm`（Client Component）: 名前・本文・任意メール・不可視ハニーポット欄。送信後「承認待ちです」を表示。
- **管理画面** `frontend/src/app/admin/comments/page.tsx`: `PENDING` 一覧 → 承認 / 却下 / 削除（`admin-api.ts` 経由）。`AdminNav.tsx` に導線追加。

---

## 10. テスト

- **domain**: `Approve()` / `Reject()` / `IsVisible()` のテーブル駆動テスト。
- **rpc**: `CreateComment` の「未公開記事へのコメント拒否」「ハニーポット破棄」「レート制限超過で `ResourceExhausted`」「`PENDING` で作成」、`ListComments` が `APPROVED` のみ返すこと、モデレーション系の認証必須。
- **mysql**: `comment_repository` の Create / List（フィルタ）/ Update を実 DB 統合テスト（既存 repo テスト方針に準拠、テストデータはゼロ値を避け具体値を使う）。
- 期待値はリテラル直書き・具体値で assert する。

---

## 11. 影響範囲まとめ

| 種別 | ファイル |
| --- | --- |
| **新規** | `domain/comment/comment.go`, `infrastructure/mysql/comment_repository.go`, `interface/rpc/comment_server.go`, `db/migrations/000003_create_comments.{up,down}.sql`, `proto/blog/v1/comment.proto`, フロント `admin/comments/page.tsx`・コメント表示/投稿コンポーネント |
| **追記** | `domain/repository/repository.go`, `interface/rpc/validation.go`, `interface/rpc/converter.go`, `cmd/server/main.go`, `admin/AdminNav.tsx`, `posts/[slug]/page.tsx`, `lib/admin-api.ts`, 各テスト |
