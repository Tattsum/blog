import Link from "next/link";
import { tagClient } from "@/lib/api";
import type { Tag } from "@/gen/blog/v1/tag_pb";
import { Header } from "@/components/Header";

export const metadata = {
  title: "タグ一覧 | ブログ",
  description: "ブログのタグ一覧",
};

export default async function TagsPage() {
  let tags: Tag[] = [];
  let error: string | null = null;

  try {
    const res = await tagClient.listTags({
      page: 1,
      pageSize: 100,
    });
    tags = res.tags ?? [];
  } catch (e) {
    error =
      e instanceof Error ? e.message : "タグ一覧の取得に失敗しました。";
  }

  return (
    <div style={{ maxWidth: 720, margin: "0 auto", padding: 24 }}>
      <Header />

      <h2 style={{ fontSize: "1.25rem", fontWeight: 600, marginBottom: 16 }}>
        タグ一覧
      </h2>

      {error && (
        <p style={{ color: "#c00", marginBottom: 16 }}>{error}</p>
      )}

      {!error && tags.length === 0 && (
        <p style={{ color: "#666" }}>まだタグがありません。</p>
      )}

      {!error && tags.length > 0 && (
        <ul
          style={{
            listStyle: "none",
            padding: 0,
            margin: 0,
            display: "flex",
            flexWrap: "wrap",
            gap: 12,
          }}
        >
          {tags.map((tag) => (
            <li key={tag.id}>
              <Link
                href={`/tags/${encodeURIComponent(tag.slug)}`}
                style={{
                  display: "inline-block",
                  padding: "6px 12px",
                  backgroundColor: "#f0f0f0",
                  borderRadius: 4,
                  color: "inherit",
                  textDecoration: "none",
                  fontSize: "0.9375rem",
                }}
              >
                {tag.name}
              </Link>
            </li>
          ))}
        </ul>
      )}

      <p style={{ marginTop: 24 }}>
        <Link href="/" style={{ color: "#666", textDecoration: "underline" }}>
          ← トップに戻る
        </Link>
      </p>
    </div>
  );
}
