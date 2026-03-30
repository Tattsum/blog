"use client";

import { useEffect, useMemo, useState } from "react";

type Props = {
  title: string;
  path: string;
};

function IconShare({ size = 18 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <path
        d="M12 3v12"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
      <path
        d="M7 8l5-5 5 5"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M5 21h14a2 2 0 0 0 2-2v-6"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  );
}

function IconCopy({ size = 18 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <path
        d="M9 9h10v10H9V9z"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinejoin="round"
      />
      <path
        d="M5 15H4a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1h10a1 1 0 0 1 1 1v1"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  );
}

function IconX({ size = 18 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <path
        d="M7 7l10 10"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
      <path
        d="M17 7L7 17"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  );
}

function IconBluesky({ size = 18 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <path
        d="M12 2l4 20H8L12 2z"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinejoin="round"
      />
      <path
        d="M12 7.5c2.5 0 4.5 1.2 4.5 3.3 0 2-2 3.2-4.5 3.2s-4.5-1.2-4.5-3.2C7.5 8.7 9.5 7.5 12 7.5z"
        stroke="currentColor"
        strokeWidth="1.7"
      />
    </svg>
  );
}

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

  const canShare = useMemo(() => {
    if (!mounted) return false;
    if (typeof navigator === "undefined") return false;
    if (!navigator.share) return false;
    if (!url) return false;
    if (typeof navigator.canShare !== "function") return true;
    try {
      return navigator.canShare({ url });
    } catch {
      // navigator.canShare の引数仕様が環境差でズレるケースに備える
      return true;
    }
  }, [mounted, url]);

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
      // Clipboard API の失敗時は、最低限のフォールバックとして prompt を出す
      try {
        window.prompt("このURLをコピーしてください", url);
      } catch {
        // ignore (prompt 自体も失敗しうる環境)
      }
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
      {canShare && (
        <button
          type="button"
          className="share-btn share-btn-primary"
          onClick={handleShare}
          aria-label="この記事を共有"
          disabled={!url}
        >
          <IconShare />
          <span className="share-btn-text">共有</span>
        </button>
      )}
      <button
        type="button"
        className="share-btn"
        onClick={handleCopy}
        aria-label="この記事のリンクをコピー"
        disabled={!url}
      >
        <IconCopy />
        {copied ? "コピーしました" : "リンクをコピー"}
      </button>
      {url && (
        <>
          <a
            className="share-btn"
            href={xIntentUrl}
            target="_blank"
            rel="noopener noreferrer"
            aria-label="X で共有"
          >
            <IconX />
            <span className="share-btn-text">X</span>
          </a>
          <a
            className="share-btn"
            href={blueskyIntentUrl}
            target="_blank"
            rel="noopener noreferrer"
            aria-label="Bluesky で共有"
          >
            <IconBluesky />
            <span className="share-btn-text">Bluesky</span>
          </a>
        </>
      )}
      {copyError && <span className="share-error">{copyError}</span>}
      {copied && !copyError && (
        <span className="share-toast" role="status" aria-live="polite">
          コピーしました
        </span>
      )}
    </div>
  );
}

