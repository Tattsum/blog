export function getEmbedUrl(href: string): string | null {
  if (!href || typeof href !== "string") return null;
  const s = href.trim();
  if (!s.startsWith("https://")) return null;

  try {
    const url = new URL(s);

    if (
      (url.hostname === "www.youtube.com" || url.hostname === "youtube.com") &&
      url.pathname.startsWith("/embed/")
    ) {
      const id = url.pathname.slice(7).split("/")[0];
      if (id) return `https://www.youtube.com/embed/${id}`;
    }
    if (
      (url.hostname === "www.youtube.com" || url.hostname === "youtube.com") &&
      url.pathname === "/watch" &&
      url.searchParams.has("v")
    ) {
      const id = url.searchParams.get("v");
      if (id) return `https://www.youtube.com/embed/${id}`;
    }
    if (url.hostname === "player.vimeo.com" && url.pathname.startsWith("/video/")) {
      const id = url.pathname.slice(7).split("/")[0];
      if (id) return `https://player.vimeo.com/video/${id}`;
    }
  } catch {
    return null;
  }
  return null;
}
