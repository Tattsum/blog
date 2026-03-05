"use client";

import { useState } from "react";
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
    return fallback ?? <AdminKeyForm onSetKey={admin.setAdminKey} />;
  }

  return <>{children}</>;
}

function AdminKeyForm({ onSetKey }: { onSetKey: (key: string) => void }) {
  const [key, setKey] = useState("");
  const [error, setError] = useState("");

  return (
    <div style={{ marginTop: 24 }}>
      <h2 style={{ fontSize: "1.125rem", fontWeight: 600, marginBottom: 12 }}>
        管理者キーを入力
      </h2>
      <p style={{ fontSize: "0.9375rem", color: "#666", marginBottom: 16 }}>
        記事の作成・編集・削除には管理者 API キー（X-Admin-Key）が必要です。バックエンドの ADMIN_API_KEY と一致するキーを入力してください。
      </p>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          setError("");
          if (!key.trim()) {
            setError("キーを入力してください。");
            return;
          }
          onSetKey(key.trim());
        }}
        style={{ display: "flex", flexDirection: "column", gap: 12, maxWidth: 400 }}
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
        {error && <p style={{ color: "#c00", fontSize: "0.875rem" }}>{error}</p>}
      </form>
    </div>
  );
}
