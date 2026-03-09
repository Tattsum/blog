import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { PostService } from "@/gen/blog/v1/post_pb";
import { TagService } from "@/gen/blog/v1/tag_pb";
import { edgeSafeFetch } from "@/lib/edge-safe-fetch";

const baseUrl =
  typeof window !== "undefined"
    ? (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080")
    : process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const transport = createConnectTransport({
  baseUrl,
  fetch: edgeSafeFetch,
});

export const postClient = createClient(PostService, transport);
export const tagClient = createClient(TagService, transport);
