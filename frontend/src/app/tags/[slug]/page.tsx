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
    <main style={{ maxWidth: 720, margin: "0 auto", padding: 24 }}>
      <Header />
      <h1 style={{ fontSize: "1.5rem", fontWeight: 700, marginBottom: 8 }}>
        タグ: {tag.name}
      </h1>
      <p style={{ fontSize: "0.875rem", color: "#666", marginBottom: 24 }}>
        <Link href="/tags" style={{ color: "#666" }}>
          {"← タグ一覧"}
        </Link>
      </p>
      {listRes.posts.length === 0 ? (
        <p style={{ color: "#666" }}>このタグの記事はまだありません。</p>
      ) : (
        <ul style={{ listStyle: "none", padding: 0 }}>
          {listRes.posts.map((post) => (
            <li
              key={post.id}
              style={{
                borderBottom: "1px solid #eee",
                padding: "12px 0",
              }}
            >
              <Link
                href={`/posts/${post.slug}`}
                style={{ fontWeight: 600, textDecoration: "none", color: "#333" }}
              >
                {post.title}
              </Link>
              {post.summary && (
                <p style={{ marginTop: 4, fontSize: "0.875rem", color: "#666" }}>
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
