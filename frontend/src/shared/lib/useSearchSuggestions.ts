"use client";

import { useEffect, useState } from "react";
import { catalogService } from "@/modules/catalog/api/catalogService";
import type { ProductSuggestion } from "@/modules/catalog/types";

interface SuggestionsState {
	suggestions: ProductSuggestion[];
	loading: boolean;
	error: Error | null;
}

const EMPTY: SuggestionsState = { suggestions: [], loading: false, error: null };
const cache = new Map<string, ProductSuggestion[]>();

/** Debounced product suggestions for a search query of at least two characters. */
export function useSearchSuggestions(query: string): SuggestionsState {
	const term = query.trim();
	const [state, setState] = useState<SuggestionsState & { term: string }>({ ...EMPTY, term: "" });

	useEffect(() => {
		if (term.length < 2 || cache.has(term)) return;

		let cancelled = false;
		const timer = setTimeout(() => {
			catalogService
				.getSuggestions(term)
				.then((res) => {
					if (cancelled) return;
					const list = Array.isArray(res?.data) ? res.data : [];
					cache.set(term, list);
					setState({ term, suggestions: list, loading: false, error: null });
				})
				.catch((error: unknown) => {
					if (cancelled) return;
					setState({
						term,
						suggestions: [],
						loading: false,
						error: error instanceof Error ? error : new Error(String(error)),
					});
				});
		}, 180);

		return () => {
			cancelled = true;
			clearTimeout(timer);
		};
	}, [term]);

	if (term.length < 2) return EMPTY;
	const cached = cache.get(term);
	if (cached) return { suggestions: cached, loading: false, error: null };
	if (state.term !== term) return { suggestions: [], loading: true, error: null };
	return { suggestions: state.suggestions, loading: state.loading, error: state.error };
}
