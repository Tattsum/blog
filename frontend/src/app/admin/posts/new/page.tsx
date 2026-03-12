"use client";

import { useState, useRef } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAdmin } from "../../AdminProvider";
import { AdminGate } from "../../AdminGate";
import { uploadMedia } from "@/lib/admin-api";

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
  const [uploadingThumb, setUploadingThumb] = useState(false);
  const [uploadThumbError, setUploadThumbError] = useState("");
  const thumbInputRef = useRef<HTMLInputElement>(null);
  const bodyInputRef = useRef<HTMLTextAreaElement>(null);

  const client = admin?.postClient;

  async function handleThumbUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file || !admin) return;
    e.target.value = "";
    setUploadThumbError("");
    setUploadingThumb(true);
    try {
      const { url } = await uploadMedia(file, {
        adminKey: admin.adminKey || undefined,
        sessionToken: admin.sessionToken || undefined,
      });
      setThumbnailUrl(url);
    } catch (err) {
      setUploadThumbError(err instanceof Error ? err.message : "アップロードに失敗しました");
    } finally {
      setUploadingThumb(false);
    }
  }

  async function handleBodyImageUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file || !admin) return;
    e.target.value = "";
    setError("");
    try {
      const { url } = await uploadMedia(file, {
        adminKey: admin.adminKey || undefined,
        sessionToken: admin.sessionToken || undefined,
      });
      const insert = `![画像](${url})`;
      const ta = bodyInputRef.current;
      if (ta) {
        const start = ta.selectionStart;
        const end = ta.selectionEnd;
        const before = bodyMarkdown.slice(0, start);
        const after = bodyMarkdown.slice(end);
        setBodyMarkdown(before + insert + after);
      } else {
        setBodyMarkdown((prev) => prev + insert);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "画像のアップロードに失敗しました");
    }
  }

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
            ref={bodyInputRef}
            value={bodyMarkdown}
            onChange={(e) => setBodyMarkdown(e.target.value)}
            rows={12}
            style={{ width: "100%", padding: "8px 12px", border: "1px solid #ccc", borderRadius: 4, fontFamily: "monospace" }}
          />
          <div style={{ marginTop: 8 }}>
            <input
              type="file"
              accept="image/jpeg,image/png,image/gif,image/webp,video/mp4,video/webm"
              style={{ display: "none" }}
              id="body-image-upload-new"
              onChange={handleBodyImageUpload}
            />
            <label
              htmlFor="body-image-upload-new"
              style={{
                display: "inline-block",
                padding: "6px 12px",
                border: "1px solid #333",
                borderRadius: 4,
                background: "transparent",
                cursor: "pointer",
                fontSize: "0.875rem",
              }}
            >
              画像・動画をアップロードして挿入
            </label>
          </div>
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
          <div style={{ display: "flex", gap: 8, alignItems: "center", flexWrap: "wrap" }}>
            <input
              type="url"
              value={thumbnailUrl}
              onChange={(e) => setThumbnailUrl(e.target.value)}
              placeholder="https://..."
              style={{ flex: 1, minWidth: 200, padding: "8px 12px", border: "1px solid #ccc", borderRadius: 4 }}
            />
            <input
              type="file"
              accept="image/jpeg,image/png,image/gif,image/webp"
              ref={thumbInputRef}
              style={{ display: "none" }}
              onChange={handleThumbUpload}
            />
            <button
              type="button"
              disabled={uploadingThumb}
              onClick={() => thumbInputRef.current?.click()}
              style={{
                padding: "8px 12px",
                border: "1px solid #333",
                borderRadius: 4,
                background: "transparent",
                cursor: uploadingThumb ? "not-allowed" : "pointer",
                fontSize: "0.875rem",
              }}
            >
              {uploadingThumb ? "アップロード中…" : "ファイルを選択してアップロード"}
            </button>
          </div>
          {uploadThumbError && <p style={{ color: "#c00", fontSize: "0.875rem", marginTop: 4 }}>{uploadThumbError}</p>}
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
