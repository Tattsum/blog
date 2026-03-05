import Link from "next/link";
import { notFound } from "next/navigation";
import ReactMarkdown from "react-markdown";
import { postClient } from "@/lib/api";

type Props = { params: Promise<{ slug: string }> };

export async function generateMetadata({ params }: Props) {
  const { slug } = await params;
  try {
    const res = await postClient.getPost({ id: slug });
    const post = res.post;
    if (!post) return { title: "記事が見つかりません" };
    return {
      title: `${post.title} | ブログ`,
      description: post.summary || undefined,
    };
  } catch {
    return { title: "記事が見つかりません" };
  }
}

export default async function PostPage({ params }: Props) {
  const { slug } = await params;
  let title = "";
  let bodyMarkdown = "";
  let publishedAt = "";
  let notFoundErr = false;

  try {
    const res = await postClient.getPost({ id: slug });
    const post = res.post;
    if (!post) {
      notFoundErr = true;
    } else {
      title = post.title;
      bodyMarkdown = post.bodyMarkdown ?? "";
      publishedAt = post.publishedAt ?? "";
    }
  } catch {
    notFoundErr = true;
  }

  if (notFoundErr) notFound();

  return (
    <div style={{ maxWidth: 720, margin: "0 auto", padding: 24 }}>
      <header style={{ marginBottom: 32 }}>
        <h1 style={{ fontSize: "1.5rem", fontWeight: 700 }}>
          <Link href="/" style={{ color: "inherit", textDecoration: "none" }}>
            ブログ
          </Link>
        </h1>
        <nav style={{ marginTop: 8 }}>
          <Link
            href="/tags"
            style={{ marginRight: 16, color: "#666", textDecoration: "underline" }}
          >
            {"タグ一覧"}
          </Link>
        </nav>
      </header>

      <article>
        <h1 style={{ fontSize: "1.75rem", fontWeight: 700, marginBottom: 8 }}>
          {title}
        </h1>
        {publishedAt && (
          <time
            dateTime={publishedAt}
            style={{
              display: "block",
              marginBottom: 24,
              fontSize: "0.875rem",
              color: "#666",
            }}
          >
            {formatDate(publishedAt)}
          </time>
        )}
        <div
          className="post-body"
          style={{
            lineHeight: 1.8,
            fontSize: "1rem",
          }}
        >
          <ReactMarkdown>{bodyMarkdown}</ReactMarkdown>
        </div>
      </article>

      <p style={{ marginTop: 32 }}>
        <Link href="/" style={{ color: "#666", textDecoration: "underline" }}>
          ← 一覧に戻る
        </Link>
      </p>
    </div>
  );
}

function formatDate(rfc3339: string): string {
  try {
    const d = new Date(rfc3339);
    return d.toLocaleDateString("ja-JP", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  } catch {
    return rfc3339;
  }
}
