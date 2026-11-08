"use client";

import {
	createContext,
	useCallback,
	useContext,
	useEffect,
	useState,
} from "react";
import {
	COMPARE_UPDATED_EVENT,
	addCompareSlug,
	clearCompareSlugs,
	getCompareSlugs,
	isCompareFull,
	removeCompareSlugStored,
} from "@/modules/compare/lib/compareStore";
import { MAX_COMPARE_ITEMS } from "@/modules/compare/types";

interface CompareContextType {
	compareSlugs: string[];
	compareCount: number;
	isCompared: (slug?: string | null) => boolean;
	canAdd: boolean;
	addToCompare: (slug?: string | null) => boolean;
	removeFromCompare: (slug?: string | null) => void;
	toggleCompare: (slug?: string | null) => boolean;
	clearCompare: () => void;
}

const CompareContext = createContext<CompareContextType | null>(null);

export function CompareProvider({ children }: { children: React.ReactNode }) {
	const [compareSlugs, setCompareSlugs] = useState<string[]>([]);

	const reload = useCallback(() => setCompareSlugs(getCompareSlugs()), []);

	useEffect(() => {
		reload();
		const onChange = () => reload();
		window.addEventListener(COMPARE_UPDATED_EVENT, onChange);
		return () => window.removeEventListener(COMPARE_UPDATED_EVENT, onChange);
	}, [reload]);

	const isCompared = useCallback(
		(slug?: string | null) => Boolean(slug && compareSlugs.includes(slug)),
		[compareSlugs],
	);

	const addToCompare = useCallback(
		(slug?: string | null): boolean => {
			if (!slug || compareSlugs.includes(slug)) return isCompared(slug);
			if (isCompareFull(compareSlugs)) return false;
			const next = addCompareSlug(slug);
			setCompareSlugs(next);
			return true;
		},
		[compareSlugs, isCompared],
	);

	const removeFromCompare = useCallback((slug?: string | null) => {
		setCompareSlugs(removeCompareSlugStored(slug));
	}, []);

	const toggleCompare = useCallback(
		(slug?: string | null): boolean => {
			if (!slug) return false;
			if (isCompared(slug)) {
				removeFromCompare(slug);
				return true;
			}
			return addToCompare(slug);
		},
		[isCompared, addToCompare, removeFromCompare],
	);

	const clearCompare = useCallback(() => {
		clearCompareSlugs();
		setCompareSlugs([]);
	}, []);

	return (
		<CompareContext.Provider
			value={{
				compareSlugs,
				compareCount: compareSlugs.length,
				isCompared,
				canAdd: !isCompareFull(compareSlugs),
				addToCompare,
				removeFromCompare,
				toggleCompare,
				clearCompare,
			}}
		>
			{children}
		</CompareContext.Provider>
	);
}

export function useCompare() {
	const context = useContext(CompareContext);
	if (!context) {
		throw new Error("useCompare must be used within a CompareProvider");
	}
	return context;
}

export { MAX_COMPARE_ITEMS };
