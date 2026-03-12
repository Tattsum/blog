"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAdmin } from "../../AdminProvider";
import { AdminGate } from "../../AdminGate";

export default function NewPostPage() {
  return (
    <AdminGate>
      <NewPostForm />
    </AdminGate>
  );
}

function NewPostForm() {
  const admin = useAdmin();
  const router = useRouter();
  const [title, setTitle] = useState("");
  const [slug, setSlug] = useState("");
  const [bodyMarkdown, setBodyMarkdown] = useState("");
  const [summary, setSummary] = useState("");
  const [thumbnailUrl, setThumbnailUrl] = useState("");
  const [tagIds, setTagIds] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const client = admin?.postClient;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!client) return;
    setError("");
    if (!title.trim()) {
      setError("タイトルを入力してください。");
      return;
    }
    setSubmitting(true);
    try {
      const res = await client.createPost({
        title: title.trim(),
        slug: slug.trim() || undefined,
        bodyMarkdown,
        summary,
        thumbnailUrl: thumbnailUrl.trim() || undefined,
        tagIds: tagIds.trim() ? tagIds.split(",").map((s) => s.trim()).filter(Boolean) : [],
      });
      const id = res.post?.id;
      if (id) router.push(`/admin/posts/${id}/edit`);
      else router.push("/admin/posts");
    } catch (e) {
      setError(e instanceof Error ? e.message : "作成に失敗しました");
    } finally {
      setSubmitting(false);
    }
  }

  if (!admin?.isReady) return null;

  return (
    <>
      <h2 style={{ fontSize: "1.25rem", fontWeight: 600, marginBottom: 16 }}>
        新規記事
      </h2>
      <form onSubmit={handleSubmit} style={{ display: "flex", flexDirection: "column", gap: 16, maxWidth: 600 }}>
        <div>
          <label style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>タイトル *</label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
            style={{ width: "100%", padding: "8px 12px", border: "1px solid #ccc", borderRadius: 4 }}
          />
        </div>
        <div>
          <label style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>スラグ（省略時はタイトルから自動）</label>
          <input
            type="text"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            placeholder="my-post"
            style={{ width: "100%", padding: "8px 12px", border: "1px solid #ccc", borderRadius: 4 }}
          />
        </div>
        <div>
          <label style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>本文（Markdown）</label>
          <textarea
            value={bodyMarkdown}
            onChange={(e) => setBodyMarkdown(e.target.value)}
            rows={12}
            style={{ width: "100%", padding: "8px 12px", border: "1px solid #ccc", borderRadius: 4, fontFamily: "monospace" }}
          />
        </div>
        <div>
          <label style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>要約</label>
          <textarea
            value={summary}
            onChange={(e) => setSummary(e.target.value)}
            rows={3}
            style={{ width: "100%", padding: "8px 12px", border: "1px solid #ccc", borderRadius: 4 }}
          />
        </div>
        <div>
          <label style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>サムネイル URL（任意）</label>
          <input
            type="url"
            value={thumbnailUrl}
            onChange={(e) => setThumbnailUrl(e.target.value)}
            placeholder="https://..."
            style={{ width: "100%", padding: "8px 12px", border: "1px solid #ccc", borderRadius: 4 }}
          />
        </div>
        <div>
          <label style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>タグ ID（カンマ区切り）</label>
          <input
            type="text"
            value={tagIds}
            onChange={(e) => setTagIds(e.target.value)}
            placeholder="tag-id-1, tag-id-2"
            style={{ width: "100%", padding: "8px 12px", border: "1px solid #ccc", borderRadius: 4 }}
          />
        </div>
        {error && <p style={{ color: "#c00" }}>{error}</p>}
        <div style={{ display: "flex", gap: 8 }}>
          <button
            type="submit"
            disabled={submitting}
            style={{
              padding: "8px 16px",
              border: "1px solid #333",
              borderRadius: 4,
              background: "#333",
              color: "#fff",
              cursor: submitting ? "not-allowed" : "pointer",
            }}
          >
            {submitting ? "作成中…" : "作成"}
          </button>
          <Link href="/admin/posts" style={{ padding: "8px 16px", color: "#666" }}>
            キャンセル
          </Link>
        </div>
      </form>
    </>
  );
}
