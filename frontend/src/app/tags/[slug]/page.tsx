import Link from "next/link";
import { notFound } from "next/navigation";
import { Header } from "@/components/Header";
import { postClient, tagClient } from "@/lib/api";

type Props = { params: Promise<{ slug: string }> };

export async function generateMetadata({ params }: Props) {
  const { slug } = await params;
  return { title: `タグ: ${decodeURIComponent(slug)} | ブログ` };
}

export default async function TagPostsPage({ params }: Props) {
  const { slug } = await params;
  const slugDecoded = decodeURIComponent(slug);
  const tagRes = await tagClient.listTags({ page: 1, pageSize: 500 });
  const tag = tagRes.tags?.find((t) => t.slug === slugDecoded);
  if (!tag) {
    notFound();
  }

  const listRes = await postClient.listPosts({
    status: "published",
    page: 1,
    pageSize: 20,
    tagId: tag.id,
  });

  return (
    <main className="container">
      <Header />
      <h1 style={{ fontSize: "1.5rem", fontWeight: 700, marginBottom: 8 }}>
        タグ: {tag.name}
      </h1>
      <p style={{ fontSize: "0.875rem", color: "var(--muted)", marginBottom: 24 }}>
        <Link href="/tags" style={{ color: "var(--muted)" }}>
          ← タグ一覧
        </Link>
      </p>
      {listRes.posts.length === 0 ? (
        <p style={{ color: "var(--muted)" }}>このタグの記事はまだありません。</p>
      ) : (
        <ul className="article-list">
          {listRes.posts.map((post) => (
            <li key={post.id}>
              {post.thumbnailUrl && (
                <Link
                  href={`/posts/${post.slug}`}
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
              <Link href={`/posts/${post.slug}`} className="title">
                {post.title}
              </Link>
              {post.summary && (
                <p style={{ marginTop: 4, fontSize: "0.875rem", color: "var(--muted)" }}>
                  {post.summary}
                </p>
              )}
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}
