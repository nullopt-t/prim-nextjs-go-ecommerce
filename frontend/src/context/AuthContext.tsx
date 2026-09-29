"use client";

import { createContext, useContext, useState, useEffect, useCallback } from "react";
import { authService } from "@/services/auth";

export interface User {
  id: string | number;
  email: string;
  name: string;
  phone?: string;
  avatar?: string;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  loading: boolean;
  startChallenge: (payload: { email: string }) => Promise<{ success: boolean; error?: string }>;
  resendChallenge: (payload: { email: string }) => Promise<{ success: boolean; error?: string }>;
  verifyChallenge: (payload: { email?: string; code: string }) => Promise<{ success: boolean; error?: string }>;
  logout: () => Promise<{ success: boolean; error?: string }>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchUser = useCallback(async () => {
    setLoading(true);
    try {
      const data = await authService.getMe();
      setUser(data);
    } catch {
      setUser(null);
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

  const verifyChallenge = async (payload: { email?: string; code: string }) => {
    try {
      const res: any = await authService.verifyChallenge(payload);
      if (res?.user) {
        setUser(res.user);
      } else {
        await fetchUser();
      }
      return { success: true };
    } catch (err: any) {
      return { success: false, error: err.message };
    }
  };

  const logout = async () => {
    try {
      setUser(null);
      return { success: true };
    } catch (err: any) {
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
