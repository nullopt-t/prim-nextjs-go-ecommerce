"use client";

import { useCallback, useEffect, useRef, useState } from "react";

/**
 * Small data-fetching hook for client islands.
 *
 * The fetcher is kept in a ref so callers can pass an inline function without
 * retriggering the effect on every render; only `endpoint` changes refetch
 * automatically. Use `refetch` for manual refreshes.
 */
export function useFetch<T = unknown>(
	fetcher: (endpoint: string) => Promise<T>,
	endpoint: string
) {
	const [data, setData] = useState<T | undefined>(undefined);
	const [errorMsg, setErrorMsg] = useState<string | undefined>(undefined);
	const [loading, setLoading] = useState(true);

	const fetcherRef = useRef(fetcher);

	useEffect(() => {
		fetcherRef.current = fetcher;
	}, [fetcher]);

	const execute = useCallback(async () => {
		setLoading(true);
		setErrorMsg(undefined);
		try {
			const result = await fetcherRef.current(endpoint);
			setData(result);
			return result;
		} catch (error) {
			const message =
				error instanceof Error ? error.message : "An error occurred";
			setErrorMsg(message);
			throw error;
		} finally {
			setLoading(false);
		}
	}, [endpoint]);

	useEffect(() => {
		let cancelled = false;

		const run = async () => {
			setLoading(true);
			setErrorMsg(undefined);
			try {
				const result = await fetcherRef.current(endpoint);
				if (!cancelled) setData(result);
			} catch (error) {
				if (!cancelled) {
					setErrorMsg(
						error instanceof Error ? error.message : "An error occurred"
					);
				}
			} finally {
				if (!cancelled) setLoading(false);
			}
		};

		run();

		return () => {
			cancelled = true;
		};
	}, [endpoint]);

	return { data, loading, errorMsg, refetch: execute };
}

export default useFetch;
