"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAdmin } from "./AdminProvider";

const AI_PROVIDER_STORAGE_KEY = "blog-ai-provider";

export function AdminNav() {
  const admin = useAdmin();
  const router = useRouter();

  if (!admin?.isReady) return null;

  const currentProvider =
    typeof window !== "undefined"
      ? window.sessionStorage.getItem(AI_PROVIDER_STORAGE_KEY) ?? ""
      : "";

  const handleLogout = async () => {
    if (admin.authClient) {
      try {
        await admin.authClient.logout({});
      } catch {
      }
    }
    admin.clearSession();
    admin.clearAdminKey();
    router.refresh();
  };

  const handleProviderChange = (v: string) => {
    if (typeof window === "undefined") return;
    if (v) window.sessionStorage.setItem(AI_PROVIDER_STORAGE_KEY, v);
    else window.sessionStorage.removeItem(AI_PROVIDER_STORAGE_KEY);
    router.refresh();
  };

  return (
    <nav className="admin-nav" aria-label="管理メニュー">
      <Link href="/admin/posts">記事一覧</Link>
      <Link href="/admin/posts/new">新規作成</Link>
      <Link href="/">サイトへ</Link>
      <label className="admin-inline-field">
        <span className="admin-inline-label">AI</span>
        <select
          className="admin-select"
          value={currentProvider}
          onChange={(e) => handleProviderChange(e.target.value)}
        >
          <option value="">Gemini（デフォルト）</option>
          <option value="claude">Claude</option>
        </select>
      </label>
      <button type="button" onClick={handleLogout}>
        ログアウト
      </button>
    </nav>
  );
}
