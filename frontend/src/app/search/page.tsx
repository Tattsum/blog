import Link from "next/link";
import { postClient } from "@/lib/api";
import type { Post } from "@/gen/blog/v1/post_pb";
import { Header } from "@/components/Header";

export const metadata = {
  title: "検索 | ブログ",
  description: "記事の全文検索",
};

type Props = { searchParams: Promise<{ q?: string }> };

export default async function SearchPage({ searchParams }: Props) {
  const params = await searchParams;
  const query = params.q?.trim() ?? "";
  let posts: Post[] = [];
  let totalCount = 0;
  let error: string | null = null;

  if (query.length > 0) {
    try {
      const res = await postClient.searchPosts({
        query,
        page: 1,
        pageSize: 20,
      });
      posts = res.posts ?? [];
      totalCount = res.totalCount ?? 0;
    } catch (e) {
      error =
        e instanceof Error ? e.message : "検索に失敗しました。";
    }
  }

  return (
    <div className="container">
      <Header />

      <h2 style={{ fontSize: "1.25rem", fontWeight: 600, marginBottom: 16 }}>
        記事を検索
      </h2>

      <form action="/search" method="get" style={{ marginBottom: 24 }}>
        <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
          <input
            type="search"
            name="q"
            defaultValue={query}
            placeholder="キーワードを入力"
            aria-label="検索キーワード"
            style={{
              padding: "8px 12px",
              border: "1px solid var(--border)",
              borderRadius: 4,
              fontSize: "1rem",
              minWidth: 200,
              flex: 1,
              background: "var(--card-bg)",
              color: "var(--foreground)",
            }}
          />
          <button
            type="submit"
            style={{
              padding: "8px 16px",
              border: "1px solid var(--foreground)",
              borderRadius: 4,
              background: "var(--foreground)",
              color: "var(--background)",
              fontSize: "1rem",
              cursor: "pointer",
            }}
          >
            検索
          </button>
        </div>
      </form>

      {query.length === 0 && (
        <p style={{ color: "var(--muted)" }}>
          キーワードを入力して検索してください。タイトル・本文・要約から検索します。
        </p>
      )}

      {query.length > 0 && error && (
        <p style={{ color: "var(--error)", marginBottom: 16 }}>{error}</p>
      )}

      {query.length > 0 && !error && posts.length === 0 && (
        <p style={{ color: "var(--muted)" }}>「{query}」に一致する記事はありませんでした。</p>
      )}

      {query.length > 0 && !error && posts.length > 0 && (
        <section>
          <p style={{ marginBottom: 16, fontSize: "0.9375rem", color: "var(--muted)" }}>
            「{query}」で {totalCount} 件見つかりました。
          </p>
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
                <Link href={`/posts/${encodeURIComponent(post.slug)}`} className="title">
                  {post.title}
                </Link>
                {post.publishedAt && (
                  <time
                    dateTime={post.publishedAt}
                    style={{ display: "block", marginTop: 4, fontSize: "0.875rem", color: "var(--muted)" }}
                  >
                    {formatDate(post.publishedAt)}
                  </time>
                )}
                {post.summary && (
                  <p style={{ marginTop: 8, fontSize: "0.9375rem", color: "var(--muted-foreground)", lineHeight: 1.6 }}>
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

      <p style={{ marginTop: 24 }}>
        <Link href="/" style={{ color: "var(--muted)" }}>
          ← トップに戻る
        </Link>
      </p>
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
