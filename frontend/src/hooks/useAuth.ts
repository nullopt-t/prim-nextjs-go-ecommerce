"use client";

import useFetch from "./useFetch";
import { authService } from "@/services/auth";
import { useAuthContext } from "@/context/AuthContext";
import { useState } from "react";

export function useMe() {
  const { user, loading, isAuthenticated, refreshUser } = useAuthContext();
  return { user, loading, isAuthenticated, refreshUser };
}

export function useStartChallenge() {
  const { startChallenge } = useAuthContext();
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const mutate = async (payload: any) => {
    setLoading(true);
    setErrorMsg(null);
    const result = await startChallenge(payload);
    if (!result.success) setErrorMsg(result.error || "Failed to start challenge");
    setLoading(false);
    return result;
  };

  return { startChallenge: mutate, loading, errorMsg };
}

export function useResendChallenge() {
  const { resendChallenge } = useAuthContext();
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const mutate = async (payload: any) => {
    setLoading(true);
    setErrorMsg(null);
    const result = await resendChallenge(payload);
    if (!result.success) setErrorMsg(result.error || "Failed to resend challenge");
    setLoading(false);
    return result;
  };

  return { resendChallenge: mutate, loading, errorMsg };
}

export function useVerifyChallenge() {
  const { verifyChallenge } = useAuthContext();
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const mutate = async (payload: any) => {
    setLoading(true);
    setErrorMsg(null);
    const result = await verifyChallenge(payload);
    if (!result.success) setErrorMsg(result.error || "Failed to verify challenge");
    setLoading(false);
    return result;
  };

  return { verifyChallenge: mutate, loading, errorMsg };
}

export function useLogout() {
  const { logout } = useAuthContext();
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const mutate = async (_sessionId?: string) => {
    setLoading(true);
    setErrorMsg(null);
    const result = await logout();
    if (!result.success) setErrorMsg(result.error || "Failed to logout");
    setLoading(false);
    return result;
  };

  return { logout: mutate, loading, errorMsg };
}

export function useSessions() {
  const { data, loading, errorMsg, refetch } = useFetch(() => authService.getSessions(), "");
  return { sessions: data, loading, errorMsg, refreshSessions: refetch };
}
