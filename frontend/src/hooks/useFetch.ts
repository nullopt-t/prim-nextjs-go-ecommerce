"use client";

import { useState, useEffect } from "react";

export function useFetch<T = any>(
  fetcher: (endpoint: string) => Promise<T>,
  endpoint: string
) {
  const [data, setData] = useState<T | undefined>(undefined);
  const [errorMsg, setErrorMsg] = useState<unknown>(undefined);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let isCancelled = false;

    const execute = async () => {
      setLoading(true);
      try {
        const result = await fetcher(endpoint);
        if (!isCancelled) {
          setData(result);
        }
      } catch (error) {
        if (!isCancelled) {
          setErrorMsg(error);
        }
      } finally {
        if (!isCancelled) {
          setLoading(false);
        }
      }
    };

    execute();

    return () => {
      isCancelled = true;
    };
  }, [fetcher, endpoint]);

  const refetch = async () => {
    setLoading(true);
    try {
      const result = await fetcher(endpoint);
      setData(result);
      return result;
    } catch (error) {
      setErrorMsg(error);
      throw error;
    } finally {
      setLoading(false);
    }
  };

  return { data, loading, errorMsg, refetch };
}

export default useFetch;
