"use client";

import { useState } from "react";
import { getLoginClient } from "@/lib/admin-api";
import { useAdmin } from "./AdminProvider";

type Props = {
  children: React.ReactNode;
  fallback?: React.ReactNode;
  showNavWhenReady?: boolean;
};

export function AdminGate({ children, fallback, showNavWhenReady }: Props) {
  const admin = useAdmin();

  if (admin === null) {
    return fallback ?? null;
  }

  if (!admin.isReady) {
    if (showNavWhenReady) return null;
    return (
      fallback ?? (
        <AdminLoginForm
          onLogin={admin.setSessionToken}
          onUseApiKey={admin.setAdminKey}
        />
      )
    );
  }

  return <>{children}</>;
}

function AdminLoginForm({
  onLogin,
  onUseApiKey,
}: {
  onLogin: (sessionToken: string) => void;
  onUseApiKey: (key: string) => void;
}) {
  const [mode, setMode] = useState<"login" | "key">("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [key, setKey] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  if (mode === "key") {
    return (
      <div style={{ marginTop: 24 }}>
        <h2 style={{ fontSize: "1.125rem", fontWeight: 600, marginBottom: 12 }}>
          管理者キーを入力
        </h2>
        <p style={{ fontSize: "0.9375rem", color: "#666", marginBottom: 16 }}>
          記事の作成・編集・削除には管理者 API キー（X-Admin-Key）が必要です。
        </p>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            setError("");
            if (!key.trim()) {
              setError("キーを入力してください。");
              return;
            }
            onUseApiKey(key.trim());
          }}
          style={{
            display: "flex",
            flexDirection: "column",
            gap: 12,
            maxWidth: 400,
          }}
        >
          <input
            type="password"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            placeholder="ADMIN_API_KEY"
            autoComplete="off"
            style={{
              padding: "8px 12px",
              border: "1px solid #ccc",
              borderRadius: 4,
              fontSize: "1rem",
            }}
          />
          <button
            type="submit"
            style={{
              padding: "8px 16px",
              border: "1px solid #333",
              borderRadius: 4,
              background: "#333",
              color: "#fff",
              cursor: "pointer",
            }}
          >
            送信
          </button>
          <button
            type="button"
            onClick={() => setMode("login")}
            style={{
              padding: "8px 16px",
              border: "1px solid #ccc",
              borderRadius: 4,
              background: "#fff",
              cursor: "pointer",
              fontSize: "0.875rem",
            }}
          >
            メールでログインに切り替え
          </button>
          {error && (
            <p style={{ color: "#c00", fontSize: "0.875rem" }}>{error}</p>
          )}
        </form>
      </div>
    );
  }

  return (
    <div style={{ marginTop: 24 }}>
      <h2 style={{ fontSize: "1.125rem", fontWeight: 600, marginBottom: 12 }}>
        管理者ログイン
      </h2>
      <p style={{ fontSize: "0.9375rem", color: "#666", marginBottom: 16 }}>
        users テーブルに登録したメールとパスワードでログインしてください。
      </p>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          setError("");
          if (!email.trim() || !password) {
            setError("メールとパスワードを入力してください。");
            return;
          }
          setLoading(true);
          try {
            const client = getLoginClient();
            const res = await client.login({
              email: email.trim(),
              password,
            });
            const token = res.sessionToken;
            if (token) {
              onLogin(token);
            } else {
              setError("ログインに失敗しました。");
            }
          } catch (err) {
            const msg =
              err instanceof Error ? err.message : "ログインに失敗しました。";
            setError(msg);
          } finally {
            setLoading(false);
          }
        }}
        style={{
          display: "flex",
          flexDirection: "column",
          gap: 12,
          maxWidth: 400,
        }}
      >
        <input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="メールアドレス"
          autoComplete="email"
          style={{
            padding: "8px 12px",
            border: "1px solid #ccc",
            borderRadius: 4,
            fontSize: "1rem",
          }}
        />
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="パスワード"
          autoComplete="current-password"
          style={{
            padding: "8px 12px",
            border: "1px solid #ccc",
            borderRadius: 4,
            fontSize: "1rem",
          }}
        />
        <button
          type="submit"
          disabled={loading}
          style={{
            padding: "8px 16px",
            border: "1px solid #333",
            borderRadius: 4,
            background: "#333",
            color: "#fff",
            cursor: loading ? "wait" : "pointer",
          }}
        >
          {loading ? "ログイン中…" : "ログイン"}
        </button>
        <button
          type="button"
          onClick={() => setMode("key")}
          style={{
            padding: "8px 16px",
            border: "1px solid #ccc",
            borderRadius: 4,
            background: "#fff",
            cursor: "pointer",
            fontSize: "0.875rem",
          }}
        >
          API キーで入る
        </button>
        {error && (
          <p style={{ color: "#c00", fontSize: "0.875rem" }}>{error}</p>
        )}
      </form>
    </div>
  );
}
