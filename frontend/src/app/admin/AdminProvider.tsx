"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  createAdminClients,
  createAdminClientsWithSession,
} from "@/lib/admin-api";

const SESSION_STORAGE_KEY = "blog-session-token";
const ADMIN_KEY_STORAGE_KEY = "blog-admin-key";

type SessionClients = ReturnType<typeof createAdminClientsWithSession>;
type KeyClients = ReturnType<typeof createAdminClients>;

type AdminContextValue = {
  sessionToken: string;
  setSessionToken: (token: string) => void;
  clearSession: () => void;
  adminKey: string;
  setAdminKey: (key: string) => void;
  clearAdminKey: () => void;
  isReady: boolean;
  authClient: SessionClients["authClient"] | null;
  postClient: SessionClients["postClient"] | KeyClients["postClient"] | null;
  tagClient: SessionClients["tagClient"] | KeyClients["tagClient"] | null;
  aiClient: SessionClients["aiClient"] | KeyClients["aiClient"] | null;
};

const AdminContext = createContext<AdminContextValue | null>(null);

function getStoredSessionToken(): string {
  if (typeof window === "undefined") return "";
  return window.sessionStorage.getItem(SESSION_STORAGE_KEY) ?? "";
}

function getStoredAdminKey(): string {
  if (typeof window === "undefined") return "";
  return window.sessionStorage.getItem(ADMIN_KEY_STORAGE_KEY) ?? "";
}

export function AdminProvider({ children }: { children: React.ReactNode }) {
  const [sessionToken, setSessionTokenState] = useState("");
  const [adminKey, setAdminKeyState] = useState("");

  useEffect(() => {
    setSessionTokenState(getStoredSessionToken());
    setAdminKeyState(getStoredAdminKey());
  }, []);

  const setSessionToken = useCallback((token: string) => {
    setSessionTokenState(token);
    if (typeof window !== "undefined") {
      if (token) window.sessionStorage.setItem(SESSION_STORAGE_KEY, token);
      else window.sessionStorage.removeItem(SESSION_STORAGE_KEY);
    }
  }, []);

  const clearSession = useCallback(() => setSessionToken(""), [setSessionToken]);

  const setAdminKey = useCallback((key: string) => {
    setAdminKeyState(key);
    if (typeof window !== "undefined") {
      if (key) window.sessionStorage.setItem(ADMIN_KEY_STORAGE_KEY, key);
      else window.sessionStorage.removeItem(ADMIN_KEY_STORAGE_KEY);
    }
  }, []);

  const clearAdminKey = useCallback(() => setAdminKey(""), [setAdminKey]);

  const sessionClients = useMemo(
    () => (sessionToken ? createAdminClientsWithSession(sessionToken) : null),
    [sessionToken]
  );
  const keyClients = useMemo(
    () => (adminKey ? createAdminClients(adminKey) : null),
    [adminKey]
  );
  const clients = sessionClients ?? keyClients;

  const value = useMemo<AdminContextValue>(
    () => ({
      sessionToken,
      setSessionToken,
      clearSession,
      adminKey,
      setAdminKey,
      clearAdminKey,
      isReady: !!clients,
      authClient: sessionClients?.authClient ?? null,
      postClient: clients?.postClient ?? null,
      tagClient: clients?.tagClient ?? null,
      aiClient: clients?.aiClient ?? null,
    }),
    [
      sessionToken,
      setSessionToken,
      clearSession,
      adminKey,
      setAdminKey,
      clearAdminKey,
      sessionClients,
      clients,
    ]
  );

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
