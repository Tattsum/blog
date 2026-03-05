"use client";

import { useState, useEffect } from "react";
import { useRouter, useParams } from "next/navigation";
import Link from "next/link";
import { useAdmin } from "../../../AdminProvider";
import { AdminGate } from "../../../AdminGate";
import type { Post } from "@/gen/blog/v1/post_pb";
import { Post_Status } from "@/gen/blog/v1/post_pb";

export default function EditPostPage() {
  return (
    <AdminGate>
      <EditPostForm />
    </AdminGate>
  );
}

function EditPostForm() {
  const admin = useAdmin();
  const router = useRouter();
  const params = useParams();
  const id = typeof params.id === "string" ? params.id : "";
  const [post, setPost] = useState<Post | null>(null);
  const [title, setTitle] = useState("");
  const [slug, setSlug] = useState("");
  const [bodyMarkdown, setBodyMarkdown] = useState("");
  const [summary, setSummary] = useState("");
  const [tagIds, setTagIds] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const client = admin?.postClient;

  useEffect(() => {
    if (!client || !id) {
      setLoading(false);
      return;
    }
    client
      .getPost({ id })
      .then((res) => {
        const p = res.post;
        if (p) {
          setPost(p);
          setTitle(p.title ?? "");
          setSlug(p.slug ?? "");
          setBodyMarkdown(p.bodyMarkdown ?? "");
          setSummary(p.summary ?? "");
          setTagIds((p.tagIds ?? []).join(", "));
        }
      })
      .catch(() => setError("記事の取得に失敗しました"))
      .finally(() => setLoading(false));
  }, [client, id]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!client || !id) return;
    setError("");
    if (!title.trim()) {
      setError("タイトルを入力してください。");
      return;
    }
    setSubmitting(true);
    try {
      await client.updatePost({
        id,
        title: title.trim(),
        slug: slug.trim() || undefined,
        bodyMarkdown,
        summary,
        tagIds: tagIds.trim() ? tagIds.split(",").map((s) => s.trim()).filter(Boolean) : [],
      });
      setPost((prev) => (prev ? { ...prev, title: title.trim(), slug: slug.trim(), bodyMarkdown, summary } : null));
    } catch (e) {
      setError(e instanceof Error ? e.message : "更新に失敗しました");
    } finally {
      setSubmitting(false);
    }
  }

  async function handlePublish(unpublish: boolean) {
    if (!client || !id) return;
    setError("");
    setSubmitting(true);
    try {
      await client.publishPost({ id, unpublish });
      router.refresh();
      const res = await client.getPost({ id });
      if (res.post) setPost(res.post);
    } catch (e) {
      setError(e instanceof Error ? e.message : "公開状態の変更に失敗しました");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete() {
    if (!client || !id) return;
    if (!confirm("この記事を削除してもよろしいですか？")) return;
    setError("");
    setSubmitting(true);
    try {
      await client.deletePost({ id });
      router.push("/admin/posts");
    } catch (e) {
      setError(e instanceof Error ? e.message : "削除に失敗しました");
      setSubmitting(false);
    }
  }

  if (!admin?.isReady) return null;
  if (loading) return <p style={{ color: "#666" }}>読み込み中…</p>;
  if (error && !post) return <p style={{ color: "#c00" }}>{error}</p>;
  if (!post) return <p style={{ color: "#666" }}>記事が見つかりません。</p>;

  const isPublished = post.status === Post_Status.PUBLISHED;

  return (
    <>
      <h2 style={{ fontSize: "1.25rem", fontWeight: 600, marginBottom: 16 }}>
        編集: {post.title || "(無題)"}
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
          <label style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>スラグ</label>
          <input
            type="text"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
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
          <label style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>タグ ID（カンマ区切り）</label>
          <input
            type="text"
            value={tagIds}
            onChange={(e) => setTagIds(e.target.value)}
            style={{ width: "100%", padding: "8px 12px", border: "1px solid #ccc", borderRadius: 4 }}
          />
        </div>
        {error && <p style={{ color: "#c00" }}>{error}</p>}
        <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
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
            {submitting ? "保存中…" : "保存"}
          </button>
          <button
            type="button"
            onClick={() => handlePublish(!isPublished)}
            disabled={submitting}
            style={{
              padding: "8px 16px",
              border: "1px solid #333",
              borderRadius: 4,
              background: "transparent",
              cursor: submitting ? "not-allowed" : "pointer",
            }}
          >
            {isPublished ? "下書きに戻す" : "公開する"}
          </button>
          <button
            type="button"
            onClick={handleDelete}
            disabled={submitting}
            style={{
              padding: "8px 16px",
              border: "1px solid #c00",
              borderRadius: 4,
              background: "transparent",
              color: "#c00",
              cursor: submitting ? "not-allowed" : "pointer",
            }}
          >
            削除
          </button>
          {isPublished && (
            <Link href={`/posts/${encodeURIComponent(post.slug)}`} style={{ padding: "8px 16px", color: "#666" }}>
              表示
            </Link>
          )}
          <Link href="/admin/posts" style={{ padding: "8px 16px", color: "#666" }}>
            一覧へ
          </Link>
        </div>
      </form>
    </>
  );
}
