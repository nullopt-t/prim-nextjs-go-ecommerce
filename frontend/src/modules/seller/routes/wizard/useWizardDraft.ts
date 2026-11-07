"use client";

import { useEffect } from "react";
import type { ProductForm, VariantOverride } from "./wizardTypes";

interface DraftState {
	formData: ProductForm;
	variantOverrides: Record<string, VariantOverride>;
	savedCombinations: Record<string, string>[];
	swatchImages: Record<string, string>;
}

// Persists the staged draft so back/forward navigation and refresh keep
// state (form, per-variant overrides, saved combinations, swatch images).
export function usePersistDraft(
	draftKey: string,
	state: DraftState,
	hydrated: boolean,
	setLastSavedAt: (value: number) => void,
) {
	useEffect(() => {
		if (!hydrated) return;
		sessionStorage.setItem(draftKey, JSON.stringify(state));
		setLastSavedAt(Date.now());
	}, [draftKey, state.formData, state.variantOverrides, state.savedCombinations, state.swatchImages, hydrated, setLastSavedAt]);
}
