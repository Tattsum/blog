"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useAdmin } from "../AdminProvider";
import type { Post } from "@/gen/blog/v1/post_pb";
import { Post_Status } from "@/gen/blog/v1/post_pb";
import { AdminGate } from "../AdminGate";
import { toAdminErrorMessage } from "@/lib/admin-error";

export default function AdminPostsPage() {
  return (
    <AdminGate>
      <AdminPostsList />
    </AdminGate>
  );
}

function AdminPostsList() {
  const admin = useAdmin();
  const [posts, setPosts] = useState<Post[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [statusFilter, setStatusFilter] = useState<"draft" | "published">("published");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const client = admin?.postClient ?? null;

  useEffect(() => {
    if (!client) return;
    setLoading(true);
    setError("");
    client
      .listPosts({
        page: 1,
        pageSize: 50,
        status: statusFilter,
      })
      .then((res) => {
        setPosts(res.posts ?? []);
        setTotalCount(res.totalCount ?? 0);
      })
      .catch((e) => setError(toAdminErrorMessage(e, "一覧の取得に失敗しました")))
      .finally(() => setLoading(false));
  }, [client, statusFilter]);

  if (!admin?.isReady) return null;

  return (
    <>
      <h2 className="admin-page-title">記事一覧</h2>
      <div className="admin-filter-buttons">
        <span className="admin-muted">表示:</span>
        <button
          type="button"
          onClick={() => setStatusFilter("published")}
          className={statusFilter === "published" ? "active" : undefined}
        >
          公開
        </button>
        <button
          type="button"
          onClick={() => setStatusFilter("draft")}
          className={statusFilter === "draft" ? "active" : undefined}
        >
          下書き
        </button>
      </div>
      {error && <p className="admin-error">{error}</p>}
      {loading && <p className="admin-muted">読み込み中…</p>}
      {!loading && !error && posts.length === 0 && (
        <p className="admin-muted">記事がありません。</p>
      )}
      {!loading && posts.length > 0 && (
        <ul style={{ listStyle: "none", padding: 0, margin: 0 }}>
          {posts.map((post) => (
            <li key={post.id} className="admin-card">
              <div style={{ display: "flex", alignItems: "center", gap: 12, flexWrap: "wrap" }}>
                {post.thumbnailUrl && (
                  <Link href={`/admin/posts/${post.id}/edit`}>
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img
                      src={post.thumbnailUrl}
                      alt=""
                      width={48}
                      height={48}
                      className="admin-thumbnail"
                    />
                  </Link>
                )}
                <span>
                  <Link href={`/admin/posts/${post.id}/edit`} className="admin-card-link">
                    {post.title || "(無題)"}
                  </Link>
                  <span
                    className="admin-muted"
                    style={{
                      marginLeft: 8,
                      fontSize: "0.75rem",
                      color: post.status === Post_Status.PUBLISHED ? "var(--link)" : undefined,
                    }}
                  >
                    {post.status === Post_Status.PUBLISHED ? "公開" : "下書き"}
                  </span>
                </span>
              </div>
              <div className="admin-card-actions">
                {post.status === Post_Status.PUBLISHED ? (
                  <Link href={`/posts/${encodeURIComponent(post.slug)}`}>表示</Link>
                ) : null}
                <Link href={`/admin/posts/${post.id}/edit`}>編集</Link>
              </div>
            </li>
          ))}
        </ul>
      )}
      {!loading && totalCount > 0 && (
        <p className="admin-muted" style={{ marginTop: 16 }}>
          全 {totalCount} 件
        </p>
      )}
    </>
  );
}
