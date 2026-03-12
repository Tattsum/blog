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
      <h2 className="admin-page-title">新規記事</h2>
      <form onSubmit={handleSubmit} className="admin-form">
        <div className="admin-form-group">
          <label>タイトル *</label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
            className="admin-input"
          />
        </div>
        <div className="admin-form-group">
          <label>スラグ（省略時はタイトルから自動）</label>
          <input
            type="text"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            placeholder="my-post"
            className="admin-input"
          />
        </div>
        <div className="admin-form-group">
          <label>本文（Markdown）</label>
          <textarea
            ref={bodyInputRef}
            value={bodyMarkdown}
            onChange={(e) => setBodyMarkdown(e.target.value)}
            rows={12}
            className="admin-textarea mono"
          />
          <div style={{ marginTop: 8 }}>
            <input
              type="file"
              accept="image/jpeg,image/png,image/gif,image/webp,video/mp4,video/webm"
              style={{ display: "none" }}
              id="body-image-upload-new"
              onChange={handleBodyImageUpload}
            />
            <label htmlFor="body-image-upload-new" className="admin-label-btn">
              画像・動画をアップロードして挿入
            </label>
          </div>
        </div>
        <div className="admin-form-group">
          <label>要約</label>
          <textarea
            value={summary}
            onChange={(e) => setSummary(e.target.value)}
            rows={3}
            className="admin-textarea"
          />
        </div>
        <div className="admin-form-group">
          <label>サムネイル URL（任意）</label>
          <div style={{ display: "flex", gap: 8, alignItems: "center", flexWrap: "wrap" }}>
            <input
              type="url"
              value={thumbnailUrl}
              onChange={(e) => setThumbnailUrl(e.target.value)}
              placeholder="https://..."
              className="admin-input"
              style={{ flex: 1, minWidth: 200 }}
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
              className="admin-btn-secondary"
            >
              {uploadingThumb ? "アップロード中…" : "ファイルを選択してアップロード"}
            </button>
          </div>
          {uploadThumbError && <p className="admin-error" style={{ fontSize: "0.875rem", marginTop: 4 }}>{uploadThumbError}</p>}
        </div>
        <div className="admin-form-group">
          <label>タグ ID（カンマ区切り）</label>
          <input
            type="text"
            value={tagIds}
            onChange={(e) => setTagIds(e.target.value)}
            placeholder="tag-id-1, tag-id-2"
            className="admin-input"
          />
        </div>
        {error && <p className="admin-error">{error}</p>}
        <div className="admin-form-actions">
          <button type="submit" disabled={submitting} className="admin-btn">
            {submitting ? "作成中…" : "作成"}
          </button>
          <Link href="/admin/posts" className="admin-btn admin-btn-secondary">
            キャンセル
          </Link>
        </div>
      </form>
    </>
  );
}
