"use client";

import ReactMarkdown from "react-markdown";
import { getEmbedUrl } from "@/lib/embed-url";

type MarkdownBodyProps = {
  children: string;
};

/**
 * 記事本文用の Markdown レンダラ。
 * - リンクの href が埋め込み許可 URL（YouTube / Vimeo）の場合は iframe で表示する。
 */
export function MarkdownBody({ children }: MarkdownBodyProps) {
  return (
    <ReactMarkdown
      components={{
        a: ({ href, children: linkChildren, ...props }) => {
          const embedSrc = href ? getEmbedUrl(href) : null;
          if (embedSrc) {
            return (
              <span className="embed-video-wrapper" style={{ display: "block", marginBlock: "1rem" }}>
                <iframe
                  src={embedSrc}
                  title="動画埋め込み"
                  allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
                  allowFullScreen
                  style={{ width: "100%", maxWidth: 560, aspectRatio: "16/9", border: 0, borderRadius: 4 }}
                />
              </span>
            );
          }
          return (
            <a href={href} {...props} target="_blank" rel="noopener noreferrer">
              {linkChildren}
            </a>
          );
        },
        img: ({ src, alt, ...props }) => (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={src}
            alt={alt ?? ""}
            loading="lazy"
            decoding="async"
            style={{ maxWidth: "100%", height: "auto" }}
            {...props}
          />
        ),
      }}
    >
      {children}
    </ReactMarkdown>
  );
}
