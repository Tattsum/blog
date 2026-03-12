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
      <div className="admin-container">
        <header className="admin-header">
          <h1>
            <Link href="/admin">管理</Link>
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
