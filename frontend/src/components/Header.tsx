import Link from "next/link";

export function Header() {
  return (
    <header className="site-header">
      <h1 className="site-title">
        <Link href="/">ブログ</Link>
      </h1>
      <nav className="site-nav" aria-label="メインナビゲーション">
        <div className="nav-links">
          <Link href="/tags">タグ一覧</Link>
          <Link href="/search">検索</Link>
          <Link href="/admin">管理</Link>
        </div>
        <form action="/search" method="get" className="search-form" role="search">
          <input
            type="search"
            name="q"
            placeholder="記事を検索"
            aria-label="検索キーワード"
          />
          <button type="submit">検索</button>
        </form>
      </nav>
    </header>
  );
}
