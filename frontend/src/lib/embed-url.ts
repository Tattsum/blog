/**
 * 動画埋め込み用 URL のホワイトリスト。
 * 許可したパターンのリンクは記事本文で iframe として表示する。
 */

/** 埋め込み可能な URL かどうか、および変換後の embed URL を返す。該当しなければ null。 */
export function getEmbedUrl(href: string): string | null {
  if (!href || typeof href !== "string") return null;
  const s = href.trim();
  if (!s.startsWith("https://")) return null;

  try {
    const url = new URL(s);

    // https://www.youtube.com/embed/VIDEO_ID または https://youtube.com/embed/VIDEO_ID
    if (
      (url.hostname === "www.youtube.com" || url.hostname === "youtube.com") &&
      url.pathname.startsWith("/embed/")
    ) {
      const id = url.pathname.slice(7).split("/")[0];
      if (id) return `https://www.youtube.com/embed/${id}`;
    }
    // https://www.youtube.com/watch?v=VIDEO_ID
    if (
      (url.hostname === "www.youtube.com" || url.hostname === "youtube.com") &&
      url.pathname === "/watch" &&
      url.searchParams.has("v")
    ) {
      const id = url.searchParams.get("v");
      if (id) return `https://www.youtube.com/embed/${id}`;
    }

    // https://player.vimeo.com/video/VIDEO_ID
    if (url.hostname === "player.vimeo.com" && url.pathname.startsWith("/video/")) {
      const id = url.pathname.slice(7).split("/")[0];
      if (id) return `https://player.vimeo.com/video/${id}`;
    }
  } catch {
    return null;
  }
  return null;
}
