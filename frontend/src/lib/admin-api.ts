"use client";

import { createClient } from "@connectrpc/connect";
import type { Interceptor } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { PostService } from "@/gen/blog/v1/post_pb";
import { TagService } from "@/gen/blog/v1/tag_pb";
import { AIService } from "@/gen/blog/v1/ai_pb";

const baseUrl =
  typeof window !== "undefined"
    ? (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080")
    : process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

function createAdminTransport(adminKey: string) {
  const authInterceptor: Interceptor = (next) => async (req) => {
    req.header.set("X-Admin-Key", adminKey);
    return await next(req);
  };

  return createConnectTransport({
    baseUrl,
    interceptors: [authInterceptor],
  });
}

export function createAdminClients(adminKey: string) {
  const transport = createAdminTransport(adminKey);
  return {
    postClient: createClient(PostService, transport),
    tagClient: createClient(TagService, transport),
    aiClient: createClient(AIService, transport),
  };
}
