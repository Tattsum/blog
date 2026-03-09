import Link from "next/link";

export function Header() {
  return (
    <header style={{ marginBottom: 32 }}>
      <h1 style={{ fontSize: "1.5rem", fontWeight: 700 }}>
        <Link href="/" style={{ color: "var(--foreground)", textDecoration: "none" }}>
          ブログ
        </Link>
      </h1>
      <nav style={{ marginTop: 8, display: "flex", alignItems: "center", gap: 16, flexWrap: "wrap" }}>
        <Link href="/tags" style={{ color: "var(--muted)" }}>
          タグ一覧
        </Link>
        <Link href="/search" style={{ color: "var(--muted)" }}>
          検索
        </Link>
        <Link href="/admin" style={{ color: "var(--muted)" }}>
          管理
        </Link>
        <form action="/search" method="get" style={{ display: "inline-flex", gap: 8 }}>
          <input
            type="search"
            name="q"
            placeholder="記事を検索"
            aria-label="検索キーワード"
            style={{
              padding: "6px 10px",
              border: "1px solid var(--border)",
              borderRadius: 4,
              fontSize: "0.9375rem",
              minWidth: 180,
              background: "var(--card-bg)",
              color: "var(--foreground)",
            }}
          />
          <button
            type="submit"
            style={{
              padding: "6px 12px",
              border: "1px solid var(--foreground)",
              borderRadius: 4,
              background: "var(--foreground)",
              color: "var(--background)",
              fontSize: "0.9375rem",
              cursor: "pointer",
            }}
          >
            検索
          </button>
        </form>
      </nav>
    </header>
  );
}
