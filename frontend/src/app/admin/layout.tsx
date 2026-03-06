import Link from "next/link";
import { AdminProvider } from "./AdminProvider";
import { AdminGate } from "./AdminGate";
import { AdminNav } from "./AdminNav";

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
          <AdminGate fallback={null} showNavWhenReady>
            <AdminNav />
          </AdminGate>
        </header>
        <AdminGate>{children}</AdminGate>
      </div>
    </AdminProvider>
  );
}
