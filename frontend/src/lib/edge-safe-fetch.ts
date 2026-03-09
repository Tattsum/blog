/**
 * Edge (Cloudflare Workers) では fetch の redirect に "error" が指定されると
 * "Invalid redirect value, must be one of 'follow' or 'manual'" となる。
 * Connect や Next が内部で redirect: "error" を使う場合があるため、
 * 常に redirect を "follow" に正規化した fetch を渡す。
 */
export function edgeSafeFetch(
  input: RequestInfo | URL,
  init?: RequestInit
): Promise<Response> {
  const opts: RequestInit = { ...init };
  if (opts.redirect === "error") {
    opts.redirect = "follow";
  }
  return fetch(input, opts);
}
