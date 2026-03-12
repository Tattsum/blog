import Link from "next/link";
import { postClient } from "@/lib/api";
import type { Post } from "@/gen/blog/v1/post_pb";
import { Header } from "@/components/Header";

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
    <div className="container">
      <Header />

      {error && (
        <p style={{ color: "var(--error)", marginBottom: 16 }}>{error}</p>
      )}

      {!error && posts.length === 0 && (
        <p style={{ color: "var(--muted)" }}>まだ記事がありません。</p>
      )}

      {!error && posts.length > 0 && (
        <section>
          <ul className="article-list">
            {posts.map((post) => (
              <li key={post.id}>
                {post.thumbnailUrl && (
                  <Link
                    href={`/posts/${encodeURIComponent(post.slug)}`}
                    style={{ display: "block", marginBottom: 8 }}
                  >
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img
                      src={post.thumbnailUrl}
                      alt=""
                      width={640}
                      height={360}
                      style={{ width: "100%", maxWidth: 640, height: "auto", borderRadius: 4, objectFit: "cover" }}
                      loading="lazy"
                      decoding="async"
                    />
                  </Link>
                )}
                <Link
                  href={`/posts/${encodeURIComponent(post.slug)}`}
                  className="title"
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
                      color: "var(--muted)",
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
                      color: "var(--muted-foreground)",
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
            <p style={{ marginTop: 16, fontSize: "0.875rem", color: "var(--muted)" }}>
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
