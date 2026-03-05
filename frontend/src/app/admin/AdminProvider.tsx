"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { createAdminClients } from "@/lib/admin-api";

const STORAGE_KEY = "blog-admin-key";

type AdminContextValue = {
  adminKey: string;
  setAdminKey: (key: string) => void;
  clearAdminKey: () => void;
  isReady: boolean;
  postClient: ReturnType<typeof createAdminClients>["postClient"] | null;
  tagClient: ReturnType<typeof createAdminClients>["tagClient"] | null;
  aiClient: ReturnType<typeof createAdminClients>["aiClient"] | null;
};

const AdminContext = createContext<AdminContextValue | null>(null);

function getStoredKey(): string {
  if (typeof window === "undefined") return "";
  return window.sessionStorage.getItem(STORAGE_KEY) ?? "";
}

export function AdminProvider({ children }: { children: React.ReactNode }) {
  const [adminKey, setAdminKeyState] = useState("");

  useEffect(() => {
    setAdminKeyState(getStoredKey());
  }, []);

  const setAdminKey = useCallback((key: string) => {
    setAdminKeyState(key);
    if (typeof window !== "undefined") {
      if (key) window.sessionStorage.setItem(STORAGE_KEY, key);
      else window.sessionStorage.removeItem(STORAGE_KEY);
    }
  }, []);

  const clearAdminKey = useCallback(() => setAdminKey(""), [setAdminKey]);

  const clients = useMemo(
    () => (adminKey ? createAdminClients(adminKey) : null),
    [adminKey]
  );

  const value = useMemo<AdminContextValue>(() => ({
    adminKey,
    setAdminKey,
    clearAdminKey,
    isReady: !!clients,
    postClient: clients?.postClient ?? null,
    tagClient: clients?.tagClient ?? null,
    aiClient: clients?.aiClient ?? null,
  }), [adminKey, setAdminKey, clearAdminKey, clients]);

  return (
    <AdminContext.Provider value={value}>
      {children}
    </AdminContext.Provider>
  );
}

export function useAdmin() {
  const ctx = useContext(AdminContext);
  return ctx;
}
