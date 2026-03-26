import Link from "next/link";
import { notFound } from "next/navigation";
import { postClient } from "@/lib/api";
import { Header } from "@/components/Header";
import { MarkdownBody } from "@/components/MarkdownBody";
import { ShareBar } from "@/components/ShareBar";

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
    <div className="container article-page">
      <Header />

      <article>
        {thumbnailUrl && (
          <div className="article-hero">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={thumbnailUrl}
              alt=""
              width={800}
              height={450}
              decoding="async"
            />
          </div>
        )}
        <h1 className="article-title">{title}</h1>
        {publishedAt && (
          <time dateTime={publishedAt} className="article-meta">
            {formatDate(publishedAt)}
          </time>
        )}
        <ShareBar title={title} path={`/posts/${encodeURIComponent(slug)}`} />
        <div className="post-body">
          <MarkdownBody>{bodyMarkdown}</MarkdownBody>
        </div>
      </article>

      <Link href="/" className="back-link">
        ← 一覧に戻る
      </Link>
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
