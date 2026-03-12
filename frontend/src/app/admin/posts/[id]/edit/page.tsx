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
      const msg = err instanceof Error ? err.message : "アップロードに失敗しました";
      setUploadThumbError(
        msg === "unauthorized"
          ? "認証エラーです。一度ログアウトして再ログインするか、管理者キーで入れ直してください。"
          : msg
      );
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
      const msg = err instanceof Error ? err.message : "画像のアップロードに失敗しました";
      setError(
        msg === "unauthorized"
          ? "認証エラーです。一度ログアウトして再ログインするか、管理者キーで入れ直してください。"
          : msg
      );
    }
  }

  if (!admin?.isReady) return null;
  if (loading) return <p className="admin-muted">読み込み中…</p>;
  if (error && !post) return <p className="admin-error">{error}</p>;
  if (!post) return <p className="admin-muted">記事が見つかりません。</p>;

  const isPublished = post.status === Post_Status.PUBLISHED;

  return (
    <>
      <h2 className="admin-page-title">編集: {post.title || "(無題)"}</h2>
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
          <label>スラグ</label>
          <input
            type="text"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
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
              id="body-image-upload-edit"
              onChange={handleBodyImageUpload}
            />
            <label htmlFor="body-image-upload-edit" className="admin-label-btn">
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
          <button
            type="button"
            onClick={handleSummarize}
            disabled={submitting || aiBusy}
            className="admin-btn-secondary"
            style={{ marginTop: 8 }}
          >
            {aiBusy ? "要約生成中…" : "本文から要約を生成"}
          </button>
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
            className="admin-input"
          />
        </div>
        <div className="admin-form-group">
          <label>下書き支援（AI）</label>
          <input
            type="text"
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="例: 結論を強くして"
            className="admin-input"
            style={{ marginBottom: 8 }}
          />
          <button
            type="button"
            onClick={handleDraftSupport}
            disabled={submitting || aiBusy}
            className="admin-btn-secondary"
            style={{ marginBottom: 8 }}
          >
            {aiBusy ? "提案取得中…" : "提案本文を取得"}
          </button>
          {suggestedBody && (
            <>
              <textarea
                value={suggestedBody}
                readOnly
                rows={6}
                className="admin-textarea mono"
                style={{ marginBottom: 8 }}
              />
              <button type="button" onClick={applySuggestedBody} className="admin-btn">
                提案を本文に反映
              </button>
            </>
          )}
        </div>
        {error && <p className="admin-error">{error}</p>}
        <div className="admin-form-actions">
          <button type="submit" disabled={submitting} className="admin-btn">
            {submitting ? "保存中…" : "保存"}
          </button>
          <button
            type="button"
            onClick={() => handlePublish(getUnpublishFlagForPublishButton(isPublished))}
            disabled={submitting}
            className="admin-btn admin-btn-secondary"
          >
            {isPublished ? "下書きに戻す" : "公開する"}
          </button>
          <button
            type="button"
            onClick={handleDelete}
            disabled={submitting}
            className="admin-btn admin-btn-danger"
          >
            削除
          </button>
          {isPublished && (
            <Link href={`/posts/${encodeURIComponent(post.slug)}`} className="admin-btn admin-btn-secondary">
              表示
            </Link>
          )}
          <Link href="/admin/posts" className="admin-btn admin-btn-secondary">
            一覧へ
          </Link>
        </div>
      </form>
    </>
  );
}
