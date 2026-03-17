"use client";

import { createClient } from "@connectrpc/connect";
import type { Interceptor } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { PostService } from "@/gen/blog/v1/post_pb";
import { TagService } from "@/gen/blog/v1/tag_pb";
import { AIService } from "@/gen/blog/v1/ai_pb";
import { AuthService } from "@/gen/blog/v1/auth_pb";
import { edgeSafeFetch } from "@/lib/edge-safe-fetch";

const AI_PROVIDER_STORAGE_KEY = "blog-ai-provider";

const baseUrl =
  typeof window !== "undefined"
    ? (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080")
    : process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const baseTransport = () =>
  createConnectTransport({ baseUrl, fetch: edgeSafeFetch });

function getStoredAiProvider(): string {
  if (typeof window === "undefined") return "";
  return window.sessionStorage.getItem(AI_PROVIDER_STORAGE_KEY) ?? "";
}

/** ログイン用。認証ヘッダーなしで AuthService.Login を呼ぶ */
export function getLoginClient() {
  return createClient(AuthService, baseTransport());
}

/** Bearer トークンで管理者 API を呼ぶトランスポート */
function createBearerTransport(sessionToken: string) {
  const interceptor: Interceptor = (next) => async (req) => {
    req.header.set("Authorization", `Bearer ${sessionToken}`);
    const provider = getStoredAiProvider();
    if (provider) req.header.set("X-AI-Provider", provider);
    return await next(req);
  };
  return createConnectTransport({
    baseUrl,
    interceptors: [interceptor],
    fetch: edgeSafeFetch,
  });
}

/** X-Admin-Key で管理者 API を呼ぶトランスポート（従来方式・併用可能） */
function createAdminKeyTransport(adminKey: string) {
  const interceptor: Interceptor = (next) => async (req) => {
    req.header.set("X-Admin-Key", adminKey);
    const provider = getStoredAiProvider();
    if (provider) req.header.set("X-AI-Provider", provider);
    return await next(req);
  };
  return createConnectTransport({
    baseUrl,
    interceptors: [interceptor],
    fetch: edgeSafeFetch,
  });
}

/** セッショントークンで Post / Tag / AI / Auth クライアントを生成 */
export function createAdminClientsWithSession(sessionToken: string) {
  const transport = createBearerTransport(sessionToken);
  return {
    authClient: createClient(AuthService, transport),
    postClient: createClient(PostService, transport),
    tagClient: createClient(TagService, transport),
    aiClient: createClient(AIService, transport),
  };
}

/** X-Admin-Key で Post / Tag / AI クライアントを生成（従来方式） */
export function createAdminClients(adminKey: string) {
  const transport = createAdminKeyTransport(adminKey);
  return {
    postClient: createClient(PostService, transport),
    tagClient: createClient(TagService, transport),
    aiClient: createClient(AIService, transport),
  };
}

/** アップロード API のベース URL（管理画面の fetch 用） */
export function getUploadBaseUrl(): string {
  return baseUrl;
}

/** メディアアップロード。認証は adminKey または sessionToken のいずれかを渡す。 */
export async function uploadMedia(
  file: File,
  auth: { adminKey?: string; sessionToken?: string }
): Promise<{ url: string }> {
  const headers: Record<string, string> = {};
  if (auth.adminKey) {
    headers["X-Admin-Key"] = auth.adminKey;
  } else if (auth.sessionToken) {
    headers["Authorization"] = `Bearer ${auth.sessionToken}`;
  } else {
    throw new Error("認証が必要です（管理者キーまたはログイン）");
  }
  const form = new FormData();
  form.append("file", file);
  const res = await edgeSafeFetch(`${baseUrl}/upload`, {
    method: "POST",
    headers,
    body: form,
    credentials: "include",
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    const msg = (body as { error?: string }).error ?? res.statusText;
    throw new Error(msg);
  }
  const data = (await res.json()) as { url?: string };
  if (!data?.url) throw new Error("アップロード応答に URL が含まれていません");
  return { url: data.url };
}
