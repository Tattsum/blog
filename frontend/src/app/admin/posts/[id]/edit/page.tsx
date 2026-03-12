"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter, useParams } from "next/navigation";
import Link from "next/link";
import { useAdmin } from "../../../AdminProvider";
import { AdminGate } from "../../../AdminGate";
import { getUnpublishFlagForPublishButton } from "./publish-utils";
import { uploadMedia } from "@/lib/admin-api";
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
  const [thumbnailUrl, setThumbnailUrl] = useState("");
  const [tagIds, setTagIds] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [aiBusy, setAiBusy] = useState(false);
  const [prompt, setPrompt] = useState("");
  const [suggestedBody, setSuggestedBody] = useState("");
  const [uploadingThumb, setUploadingThumb] = useState(false);
  const [uploadThumbError, setUploadThumbError] = useState("");
  const thumbInputRef = useRef<HTMLInputElement>(null);
  const bodyInputRef = useRef<HTMLTextAreaElement>(null);

  const client = admin?.postClient;
  const aiClient = admin?.aiClient;

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
          setThumbnailUrl(p.thumbnailUrl ?? "");
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
        thumbnailUrl: thumbnailUrl.trim() || undefined,
        tagIds: tagIds.trim() ? tagIds.split(",").map((s) => s.trim()).filter(Boolean) : [],
      });
      setPost((prev) => (prev ? { ...prev, title: title.trim(), slug: slug.trim(), bodyMarkdown, summary, thumbnailUrl: thumbnailUrl.trim() } : null));
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

  async function handleSummarize() {
    if (!aiClient) return;
    if (!bodyMarkdown.trim()) {
      setError("本文が空のため要約できません。");
      return;
    }
    setError("");
    setAiBusy(true);
    try {
      const res = await aiClient.summarize({
        text: bodyMarkdown,
        maxSentences: 3,
      });
      setSummary(res.summary ?? "");
    } catch (e) {
      setError(e instanceof Error ? e.message : "要約の生成に失敗しました");
    } finally {
      setAiBusy(false);
    }
  }

  async function handleDraftSupport() {
    if (!aiClient) return;
    if (!bodyMarkdown.trim()) {
      setError("本文が空のため下書き支援を実行できません。");
      return;
    }
    setError("");
    setAiBusy(true);
    try {
      const res = await aiClient.draftSupport({
        prompt,
        currentBody: bodyMarkdown,
      });
      setSuggestedBody(res.suggestedBody ?? "");
    } catch (e) {
      setError(e instanceof Error ? e.message : "下書き支援の取得に失敗しました");
    } finally {
      setAiBusy(false);
    }
  }

  function applySuggestedBody() {
    if (!suggestedBody) return;
    setBodyMarkdown(suggestedBody);
  }

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
              id="body-image-upload-edit"
              onChange={handleBodyImageUpload}
            />
            <label
              htmlFor="body-image-upload-edit"
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
          <button
            type="button"
            onClick={handleSummarize}
            disabled={submitting || aiBusy}
            style={{
              marginTop: 8,
              padding: "6px 12px",
              border: "1px solid #333",
              borderRadius: 4,
              background: "transparent",
              cursor: submitting || aiBusy ? "not-allowed" : "pointer",
              fontSize: "0.875rem",
            }}
          >
            {aiBusy ? "要約生成中…" : "本文から要約を生成"}
          </button>
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
            style={{ width: "100%", padding: "8px 12px", border: "1px solid #ccc", borderRadius: 4 }}
          />
        </div>
        <div>
          <label style={{ display: "block", marginBottom: 4, fontWeight: 500 }}>下書き支援（AI）</label>
          <input
            type="text"
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="例: 結論を強くして"
            style={{ width: "100%", padding: "8px 12px", border: "1px solid #ccc", borderRadius: 4, marginBottom: 8 }}
          />
          <button
            type="button"
            onClick={handleDraftSupport}
            disabled={submitting || aiBusy}
            style={{
              padding: "6px 12px",
              border: "1px solid #333",
              borderRadius: 4,
              background: "transparent",
              cursor: submitting || aiBusy ? "not-allowed" : "pointer",
              fontSize: "0.875rem",
              marginBottom: 8,
            }}
          >
            {aiBusy ? "提案取得中…" : "提案本文を取得"}
          </button>
          {suggestedBody && (
            <>
              <textarea
                value={suggestedBody}
                readOnly
                rows={6}
                style={{
                  width: "100%",
                  padding: "8px 12px",
                  border: "1px solid #ccc",
                  borderRadius: 4,
                  fontFamily: "monospace",
                  marginBottom: 8,
                }}
              />
              <button
                type="button"
                onClick={applySuggestedBody}
                style={{
                  padding: "6px 12px",
                  border: "1px solid #333",
                  borderRadius: 4,
                  background: "#333",
                  color: "#fff",
                  fontSize: "0.875rem",
                  cursor: "pointer",
                }}
              >
                提案を本文に反映
              </button>
            </>
          )}
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
            onClick={() => handlePublish(getUnpublishFlagForPublishButton(isPublished))}
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
