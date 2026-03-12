"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useAdmin } from "../AdminProvider";
import type { Post } from "@/gen/blog/v1/post_pb";
import { Post_Status } from "@/gen/blog/v1/post_pb";
import { AdminGate } from "../AdminGate";

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
      .catch((e) => setError(e instanceof Error ? e.message : "一覧の取得に失敗しました"))
      .finally(() => setLoading(false));
  }, [client, statusFilter]);

  if (!admin?.isReady) return null;

  return (
    <>
      <h2 style={{ fontSize: "1.25rem", fontWeight: 600, marginBottom: 16 }}>
        記事一覧
      </h2>
      <div style={{ marginBottom: 16, display: "flex", gap: 8, alignItems: "center" }}>
        <span style={{ fontSize: "0.9375rem" }}>表示:</span>
        <button
          type="button"
          onClick={() => setStatusFilter("published")}
          style={{
            padding: "6px 12px",
            border: "1px solid #ccc",
            borderRadius: 4,
            background: statusFilter === "published" ? "#333" : "#fff",
            color: statusFilter === "published" ? "#fff" : "#333",
            cursor: "pointer",
          }}
        >
          公開
        </button>
        <button
          type="button"
          onClick={() => setStatusFilter("draft")}
          style={{
            padding: "6px 12px",
            border: "1px solid #ccc",
            borderRadius: 4,
            background: statusFilter === "draft" ? "#333" : "#fff",
            color: statusFilter === "draft" ? "#fff" : "#333",
            cursor: "pointer",
          }}
        >
          下書き
        </button>
      </div>
      {error && <p style={{ color: "#c00", marginBottom: 16 }}>{error}</p>}
      {loading && <p style={{ color: "#666" }}>読み込み中…</p>}
      {!loading && !error && posts.length === 0 && (
        <p style={{ color: "#666" }}>記事がありません。</p>
      )}
      {!loading && posts.length > 0 && (
        <ul style={{ listStyle: "none", padding: 0, margin: 0 }}>
          {posts.map((post) => (
            <li
              key={post.id}
              style={{
                marginBottom: 12,
                padding: 12,
                border: "1px solid #eee",
                borderRadius: 4,
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                flexWrap: "wrap",
                gap: 8,
              }}
            >
              <div>
                {post.thumbnailUrl && (
                  <Link
                    href={`/admin/posts/${post.id}/edit`}
                    style={{ display: "inline-block", marginRight: 12, verticalAlign: "middle" }}
                  >
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img
                      src={post.thumbnailUrl}
                      alt=""
                      width={48}
                      height={48}
                      style={{ width: 48, height: 48, objectFit: "cover", borderRadius: 4 }}
                    />
                  </Link>
                )}
                <span style={{ verticalAlign: "middle" }}>
                  <Link
                  href={`/admin/posts/${post.id}/edit`}
                  style={{ fontWeight: 600, color: "inherit", textDecoration: "none" }}
                >
                  {post.title || "(無題)"}
                </Link>
                <span
                  style={{
                    marginLeft: 8,
                    fontSize: "0.75rem",
                    color: post.status === Post_Status.PUBLISHED ? "#0a0" : "#666",
                  }}
                >
                  {post.status === Post_Status.PUBLISHED ? "公開" : "下書き"}
                </span>
              </span>
              </div>
              <div style={{ display: "flex", gap: 8 }}>
                {post.status === Post_Status.PUBLISHED ? (
                  <Link
                    href={`/posts/${encodeURIComponent(post.slug)}`}
                    style={{ fontSize: "0.875rem", color: "#666" }}
                  >
                    表示
                  </Link>
                ) : null}
                <Link
                  href={`/admin/posts/${post.id}/edit`}
                  style={{ fontSize: "0.875rem", color: "#666" }}
                >
                  編集
                </Link>
              </div>
            </li>
          ))}
        </ul>
      )}
      {!loading && totalCount > 0 && (
        <p style={{ marginTop: 16, fontSize: "0.875rem", color: "#666" }}>
          全 {totalCount} 件
        </p>
      )}
    </>
  );
}
