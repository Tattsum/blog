"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAdmin } from "./AdminProvider";

export function AdminNav() {
  const admin = useAdmin();
  const router = useRouter();

  if (!admin?.isReady) return null;

  const handleLogout = async () => {
    if (admin.authClient) {
      try {
        await admin.authClient.logout({});
      } catch {
        // 無視
      }
    }
    admin.clearSession();
    admin.clearAdminKey();
    router.refresh();
  };

  return (
    <nav style={{ marginTop: 8, display: "flex", gap: 16, alignItems: "center" }}>
      <Link
        href="/admin/posts"
        style={{ color: "#666", textDecoration: "underline" }}
      >
        {"記事一覧"}
      </Link>
      <Link
        href="/admin/posts/new"
        style={{ color: "#666", textDecoration: "underline" }}
      >
        {"新規作成"}
      </Link>
      <Link
        href="/"
        style={{ color: "#666", textDecoration: "underline" }}
      >
        {"サイトへ"}
      </Link>
      <button
        type="button"
        onClick={handleLogout}
        style={{
          marginLeft: 8,
          padding: "4px 12px",
          fontSize: "0.875rem",
          border: "1px solid #999",
          borderRadius: 4,
          background: "#fff",
          cursor: "pointer",
          color: "#666",
        }}
      >
        ログアウト
      </button>
    </nav>
  );
}
