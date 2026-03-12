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
    <nav className="admin-nav" aria-label="管理メニュー">
      <Link href="/admin/posts">記事一覧</Link>
      <Link href="/admin/posts/new">新規作成</Link>
      <Link href="/">サイトへ</Link>
      <button type="button" onClick={handleLogout}>
        ログアウト
      </button>
    </nav>
  );
}
