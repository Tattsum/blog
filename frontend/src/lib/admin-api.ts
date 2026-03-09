"use client";

import { createClient } from "@connectrpc/connect";
import type { Interceptor } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { PostService } from "@/gen/blog/v1/post_pb";
import { TagService } from "@/gen/blog/v1/tag_pb";
import { AIService } from "@/gen/blog/v1/ai_pb";
import { AuthService } from "@/gen/blog/v1/auth_pb";
import { edgeSafeFetch } from "@/lib/edge-safe-fetch";

const baseUrl =
  typeof window !== "undefined"
    ? (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080")
    : process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const baseTransport = () =>
  createConnectTransport({ baseUrl, fetch: edgeSafeFetch });

/** ログイン用。認証ヘッダーなしで AuthService.Login を呼ぶ */
export function getLoginClient() {
  return createClient(AuthService, baseTransport());
}

/** Bearer トークンで管理者 API を呼ぶトランスポート */
function createBearerTransport(sessionToken: string) {
  const interceptor: Interceptor = (next) => async (req) => {
    req.header.set("Authorization", `Bearer ${sessionToken}`);
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
