import Link from "next/link";
import { AdminGate } from "./AdminGate";

export default function AdminPage() {
  return (
    <AdminGate>
      <p className="admin-muted" style={{ marginBottom: 16 }}>
        記事の管理は「記事一覧」から行えます。
      </p>
      <Link href="/admin/posts" className="admin-card-link">
        記事一覧へ →
      </Link>
    </AdminGate>
  );
}
