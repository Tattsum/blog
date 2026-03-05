import Link from "next/link";
import { AdminProvider } from "./AdminProvider";
import { AdminGate } from "./AdminGate";

export const metadata = {
  title: "管理 | ブログ",
  description: "ブログ管理画面",
};

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <AdminProvider>
      <div style={{ maxWidth: 900, margin: "0 auto", padding: 24 }}>
        <header style={{ marginBottom: 24, borderBottom: "1px solid #eee", paddingBottom: 16 }}>
          <h1 style={{ fontSize: "1.25rem", fontWeight: 700 }}>
            <Link href="/admin" style={{ color: "inherit", textDecoration: "none" }}>
              管理
            </Link>
          </h1>
          <AdminGate
            fallback={null}
            showNavWhenReady
          >
            <nav style={{ marginTop: 8, display: "flex", gap: 16 }}>
              <Link href="/admin/posts" style={{ color: "#666", textDecoration: "underline" }}>
                {"記事一覧"}
              </Link>
              <Link href="/admin/posts/new" style={{ color: "#666", textDecoration: "underline" }}>
                {"新規作成"}
              </Link>
              <Link href="/" style={{ color: "#666", textDecoration: "underline" }}>
                {"サイトへ"}
              </Link>
            </nav>
          </AdminGate>
        </header>
        <AdminGate>
          {children}
        </AdminGate>
      </div>
    </AdminProvider>
  );
}
