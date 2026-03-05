import Link from "next/link";
import { postClient } from "@/lib/api";
import type { Post } from "@/gen/blog/v1/post_pb";

export const metadata = {
  title: "ブログ",
  description: "個人ブログのトップページ",
};

export default async function HomePage() {
  let posts: Post[] = [];
  let totalCount = 0;
  let error: string | null = null;

  try {
    const res = await postClient.listPosts({
      page: 1,
      pageSize: 20,
    });
    posts = res.posts ?? [];
    totalCount = res.totalCount ?? 0;
  } catch (e) {
    error =
      e instanceof Error ? e.message : "記事一覧の取得に失敗しました。";
  }

  return (
    <div style={{ maxWidth: 720, margin: "0 auto", padding: 24 }}>
      <header style={{ marginBottom: 32 }}>
        <h1 style={{ fontSize: "1.5rem", fontWeight: 700 }}>
          <Link href="/" style={{ color: "inherit", textDecoration: "none" }}>
            ブログ
          </Link>
        </h1>
        <nav style={{ marginTop: 8 }}>
          <Link
            href="/tags"
            style={{ marginRight: 16, color: "#666", textDecoration: "underline" }}
          >
            {"タグ一覧"}
          </Link>
        </nav>
      </header>

      {error && (
        <p style={{ color: "#c00", marginBottom: 16 }}>{error}</p>
      )}

      {!error && posts.length === 0 && (
        <p style={{ color: "#666" }}>まだ記事がありません。</p>
      )}

      {!error && posts.length > 0 && (
        <section>
          <ul style={{ listStyle: "none", padding: 0, margin: 0 }}>
            {posts.map((post) => (
              <li
                key={post.id}
                style={{
                  marginBottom: 24,
                  paddingBottom: 24,
                  borderBottom: "1px solid #eee",
                }}
              >
                <Link
                  href={`/posts/${encodeURIComponent(post.slug)}`}
                  style={{
                    fontSize: "1.125rem",
                    fontWeight: 600,
                    color: "inherit",
                    textDecoration: "none",
                  }}
                >
                  {post.title}
                </Link>
                {post.publishedAt && (
                  <time
                    dateTime={post.publishedAt}
                    style={{
                      display: "block",
                      marginTop: 4,
                      fontSize: "0.875rem",
                      color: "#666",
                    }}
                  >
                    {formatDate(post.publishedAt)}
                  </time>
                )}
                {post.summary && (
                  <p
                    style={{
                      marginTop: 8,
                      fontSize: "0.9375rem",
                      color: "#444",
                      lineHeight: 1.6,
                    }}
                  >
                    {post.summary}
                  </p>
                )}
              </li>
            ))}
          </ul>
          {totalCount > 20 && (
            <p style={{ marginTop: 16, fontSize: "0.875rem", color: "#666" }}>
              全 {totalCount} 件（1 ページ目を表示中）
            </p>
          )}
        </section>
      )}
    </div>
  );
}

function formatDate(rfc3339: string): string {
  try {
    const d = new Date(rfc3339);
    return d.toLocaleDateString("ja-JP", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  } catch {
    return rfc3339;
  }
}
