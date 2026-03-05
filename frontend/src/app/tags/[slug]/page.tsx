import Link from "next/link";
import { tagClient } from "@/lib/api";

type Props = { params: Promise<{ slug: string }> };

export async function generateMetadata({ params }: Props) {
  const { slug } = await params;
  return {
    title: `タグ: ${decodeURIComponent(slug)} | ブログ`,
  };
}

export default async function TagSlugPage({ params }: Props) {
  const { slug } = await params;
  const slugDecoded = decodeURIComponent(slug);

  let tagName = slugDecoded;
  try {
    const res = await tagClient.listTags({ page: 1, pageSize: 100 });
    const found = res.tags?.find((t) => t.slug === slugDecoded);
    if (found) tagName = found.name;
  } catch {
    // 一覧取得に失敗してもタグ名は slug のまま表示
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

      <h2 style={{ fontSize: "1.25rem", fontWeight: 600, marginBottom: 16 }}>
        タグ: {tagName}
      </h2>
      <p style={{ color: "#666", marginBottom: 24 }}>
        タグ別の記事一覧は今後対応予定です。
      </p>

      <p>
        <Link href="/tags" style={{ color: "#666", textDecoration: "underline" }}>
          ← タグ一覧に戻る
        </Link>
      </p>
    </div>
  );
}
