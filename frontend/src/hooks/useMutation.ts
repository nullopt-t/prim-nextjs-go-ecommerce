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
      setErrorMsg(error?.message || "An error occurred");
      return { success: false, error };
    } finally {
      setLoading(false);
    }
  };

  return { mutate, loading, errorMsg };
}
