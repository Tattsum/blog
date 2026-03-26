"use client";

import { useEffect, useMemo, useState } from "react";

type Props = {
  title: string;
  path: string;
};

export function ShareBar({ title, path }: Props) {
  const [mounted, setMounted] = useState(false);
  const [copied, setCopied] = useState(false);
  const [copyError, setCopyError] = useState("");

  useEffect(() => setMounted(true), []);
  useEffect(() => {
    if (!copied) return;
    const t = window.setTimeout(() => setCopied(false), 1500);
    return () => window.clearTimeout(t);
  }, [copied]);
  useEffect(() => {
    if (!copyError) return;
    const t = window.setTimeout(() => setCopyError(""), 3000);
    return () => window.clearTimeout(t);
  }, [copyError]);

  const url = useMemo(() => {
    if (!mounted) return "";
    const origin = window.location.origin;
    return new URL(path, origin).toString();
  }, [mounted, path]);

  const shareText = useMemo(() => {
    const t = title.trim();
    return t ? t : "ブログ記事";
  }, [title]);

  const canShare =
    mounted &&
    typeof navigator !== "undefined" &&
    !!navigator.share &&
    (typeof navigator.canShare !== "function" || navigator.canShare({ url }));

  async function handleShare() {
    if (!url) return;
    if (!navigator.share) return;
    try {
      await navigator.share({ title: shareText, text: shareText, url });
    } catch {
      // user cancelled or unsupported
    }
  }

  async function handleCopy() {
    if (!url) return;
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(url);
      } else {
        // Clipboard API が使えない環境（権限・HTTP・ブラウザ制限）向けフォールバック
        window.prompt("このURLをコピーしてください", url);
      }
      setCopyError("");
      setCopied(true);
    } catch {
      setCopyError("コピーに失敗しました。URLを長押ししてコピーしてください。");
    }
  }

  const xIntentUrl = useMemo(() => {
    if (!url) return "";
    const q = new URL("https://twitter.com/intent/tweet");
    q.searchParams.set("text", shareText);
    q.searchParams.set("url", url);
    return q.toString();
  }, [shareText, url]);

  const blueskyIntentUrl = useMemo(() => {
    if (!url) return "";
    const q = new URL("https://bsky.app/intent/compose");
    q.searchParams.set("text", `${shareText}\n${url}`);
    return q.toString();
  }, [shareText, url]);

  if (!mounted) return null;

  return (
    <div className="share-bar" role="group" aria-label="SNS で共有">
      <span className="share-label">共有</span>
      {canShare && (
        <button type="button" className="share-btn" onClick={handleShare}>
          共有…
        </button>
      )}
      <button type="button" className="share-btn" onClick={handleCopy}>
        {copied ? "コピーしました" : "リンクをコピー"}
      </button>
      {url && (
        <>
          <a
            className="share-btn"
            href={xIntentUrl}
            target="_blank"
            rel="noopener noreferrer"
          >
            X
          </a>
          <a
            className="share-btn"
            href={blueskyIntentUrl}
            target="_blank"
            rel="noopener noreferrer"
          >
            Bluesky
          </a>
        </>
      )}
      {copyError && <span className="share-error">{copyError}</span>}
    </div>
  );
}

