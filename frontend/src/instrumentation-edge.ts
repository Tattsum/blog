/**
 * Edge (Cloudflare Workers) では fetch の redirect に "error" が指定されると
 * "Invalid redirect value, must be one of 'follow' or 'manual'" となるため、
 * "error" を "follow" に置き換えた fetch で上書きする。
 */
const originalFetch = globalThis.fetch;
globalThis.fetch = function (
  input: RequestInfo | URL,
  init?: RequestInit
): Promise<Response> {
  const opts: RequestInit = init ? { ...init } : {};
  if (opts.redirect === "error") {
    opts.redirect = "follow";
  }
  return originalFetch(input, opts);
};

export {};
