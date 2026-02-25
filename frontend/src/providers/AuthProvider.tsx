import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { client, setAuthToken, setUnauthorizedHandler } from "../api/client";

const STORAGE_KEY = "bravo_token";
const STORAGE_USER = "bravo_user";

interface AuthContextType {
  token: string | null;
  userEmail: string | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, country: string, role: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(() => {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) setAuthToken(saved);
    return saved;
  });
  const [userEmail, setUserEmail] = useState<string | null>(
    () => localStorage.getItem(STORAGE_USER)
  );

  const logout = () => {
    setToken(null);
    setUserEmail(null);
    localStorage.removeItem(STORAGE_KEY);
    localStorage.removeItem(STORAGE_USER);
    setAuthToken(null);
  };

  useEffect(() => {
    setUnauthorizedHandler(logout);
    return () => setUnauthorizedHandler(null);
  });

  const login = async (email: string, password: string) => {
    const res = await client.post("/auth/login", { email, password });
    const jwt = res.data?.token;
    if (jwt) {
      setToken(jwt);
      setUserEmail(email);
      localStorage.setItem(STORAGE_KEY, jwt);
      localStorage.setItem(STORAGE_USER, email);
      setAuthToken(jwt);
    }
  };

  const register = async (email: string, password: string, country: string, role: string) => {
    const res = await client.post("/auth/register", { email, password, country, role });
    const jwt = res.data?.token;
    if (jwt) {
      setToken(jwt);
      setUserEmail(email);
      localStorage.setItem(STORAGE_KEY, jwt);
      localStorage.setItem(STORAGE_USER, email);
      setAuthToken(jwt);
    }
  };

  const value = useMemo(
    () => ({ token, userEmail, isAuthenticated: Boolean(token), login, register, logout }),
    [token, userEmail]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
