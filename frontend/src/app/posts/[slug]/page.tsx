import Link from "next/link";
import { notFound } from "next/navigation";
import { postClient } from "@/lib/api";
import { Header } from "@/components/Header";
import { MarkdownBody } from "@/components/MarkdownBody";

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
  let thumbnailUrl = "";
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
      thumbnailUrl = post.thumbnailUrl ?? "";
      publishedAt = post.publishedAt ?? "";
    }
  } catch {
    notFoundErr = true;
  }

  if (notFoundErr) notFound();

  return (
    <div className="container">
      <Header />

      <article>
        {thumbnailUrl && (
          <div style={{ marginBottom: 24 }}>
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={thumbnailUrl}
              alt=""
              width={800}
              height={450}
              style={{ width: "100%", maxWidth: 800, height: "auto", borderRadius: 4, objectFit: "cover" }}
              decoding="async"
            />
          </div>
        )}
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
              color: "var(--muted)",
            }}
          >
            {formatDate(publishedAt)}
          </time>
        )}
        <div className="post-body" style={{ lineHeight: 1.8, fontSize: "1rem" }}>
          <MarkdownBody>{bodyMarkdown}</MarkdownBody>
        </div>
      </article>

      <p style={{ marginTop: 32 }}>
        <Link href="/" style={{ color: "var(--muted)" }}>
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
