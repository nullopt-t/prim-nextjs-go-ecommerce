"use client";

import { createContext, useContext, useState, useEffect, useCallback } from "react";
import { authService } from "@/modules/auth/api/authService";
import { mergeGuestHistory } from "@/modules/history/lib/mergeGuestHistory";

export interface User {
  id: string;
  email: string;
  name: string;
  role?: string;
  status?: string;
  phone?: string;
  avatar?: string;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  loading: boolean;
  startChallenge: (payload: { email: string }) => Promise<{ success: boolean; error?: string }>;
  resendChallenge: (payload: { email: string }) => Promise<{ success: boolean; error?: string }>;
  verifyChallenge: (payload: { email?: string; code: string; mergeCart?: boolean }) => Promise<{ success: boolean; error?: string }>;
  logout: () => Promise<{ success: boolean; error?: string }>;
  refreshUser: () => Promise<User | null>;
}

const AuthContext = createContext<AuthContextType | null>(null);

function formatUserData(rawData: any): User | null {
  const d = rawData?.data || rawData;
  if (!d || !d.id) return null;

  const email = d.identifier || d.email || "";
  const name = d.fullName || d.name || (email ? email.split("@")[0] : "User");

  return {
    id: String(d.id),
    email,
    name,
    role: d.role,
    status: d.status,
    phone: d.phone,
    avatar: d.avatar,
  };
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchUser = useCallback(async () => {
    setLoading(true);
    try {
      const data = await authService.getMe();
      const parsed = formatUserData(data);
      setUser(parsed);
      if (parsed) {
        // Reconcile guest views as soon as the server session is known, including
        // users who were already signed in on this device.
        await mergeGuestHistory();
      }
      return parsed;
    } catch {
      setUser(null);
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchUser();
  }, [fetchUser]);

  const startChallenge = async (payload: { email: string }) => {
    try {
      await authService.startChallenge(payload);
      return { success: true };
    } catch (err: any) {
      return { success: false, error: err.message };
    }
  };

  const resendChallenge = async (payload: { email: string }) => {
    try {
      await authService.resendChallenge(payload);
      return { success: true };
    } catch (err: any) {
      return { success: false, error: err.message };
    }
  };

  const verifyChallenge = async (payload: { email?: string; code: string; mergeCart?: boolean }) => {
    try {
      await authService.verifyChallenge({
        email: payload.email,
        code: payload.code,
        ...(payload.mergeCart !== undefined ? { merge_cart: payload.mergeCart } : {}),
      });
      // The backend returns 204 No Content on verify, so the session must be
      // re-read before success can be claimed. If the session cookie was not
      // accepted (e.g. cross-site cookie rejection), /me fails and returning
      // success would send the user to a guarded page that bounces them back
      // to /auth, which looks like a frozen page.
      const loaded = await fetchUser();
      if (!loaded) {
        return { success: false, error: "Session could not be established; please try again." };
      }
      return { success: true };
    } catch (err: any) {
      return { success: false, error: err.message };
    }
  };

  const logout = async () => {
    try {
      await authService.logout();
      setUser(null);
      return { success: true };
    } catch (err: any) {
      setUser(null);
      return { success: false, error: err.message };
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        loading,
        startChallenge,
        resendChallenge,
        verifyChallenge,
        logout,
        refreshUser: fetchUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuthContext() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuthContext must be used within an AuthProvider");
  }
  return context;
}
