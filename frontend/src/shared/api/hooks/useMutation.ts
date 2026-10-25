"use client";

import { useState } from "react";

export function useMutation<T = any, A extends any[] = any[]>(
  mutationFn: (...args: A) => Promise<T>
) {
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const mutate = async (...args: A): Promise<{ success: boolean; data?: T; error?: any }> => {
    setLoading(true);
    setErrorMsg(null);
    try {
      const result = await mutationFn(...args);
      return { success: true, data: result };
    } catch (error: any) {
      const msg = error?.message || (typeof error === "string" ? error : "An error occurred");
      setErrorMsg(msg);
      return { success: false, error: msg };
    } finally {
      setLoading(false);
    }
  };

  return { mutate, loading, errorMsg };
}
