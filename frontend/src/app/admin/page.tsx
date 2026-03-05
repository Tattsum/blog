import Link from "next/link";
import { AdminGate } from "./AdminGate";

export default function AdminPage() {
  return (
    <AdminGate>
      <p style={{ color: "#666", marginBottom: 16 }}>
        記事の管理は「記事一覧」から行えます。
      </p>
      <Link href="/admin/posts" style={{ color: "#333", textDecoration: "underline" }}>
        {"記事一覧へ →"}
      </Link>
    </AdminGate>
  );
}
