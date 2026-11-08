"use client";

import { createContext, useCallback, useContext, useEffect, useState } from "react";
import { settingsService, type PublicSettings } from "@/modules/settings/api/settingsService";
import { PUBLIC_SETTINGS_UPDATED_EVENT } from "@/shared/lib/settingsEvents";
import {
	getStoredCurrency,
	rememberCurrencyChoice,
} from "@/shared/lib/currency";

interface PublicSettingsContextValue {
	/** The storefront settings, or null while loading / on failure. */
	settings: PublicSettings | null;
	loading: boolean;
	/** Visitor-selected display currency; falls back to the settings currency. */
	displayCurrency: string;
	setDisplayCurrency: (code: string) => void;
}

const PublicSettingsContext = createContext<PublicSettingsContextValue>({
	settings: null,
	loading: true,
	displayCurrency: "",
	setDisplayCurrency: () => {},
});

export function PublicSettingsProvider({
	children,
	initialSettings = null,
}: {
	children: React.ReactNode;
	initialSettings?: PublicSettings | null;
}) {
	const [settings, setSettings] = useState<PublicSettings | null>(initialSettings);
	const [loading, setLoading] = useState(initialSettings === null);
	// SSR renders the store's base currency; the stored choice applies on mount
	// so hydration never flashes between the two.
	const [displayCurrency, setDisplayCurrencyState] = useState<string>(
		initialSettings?.currency?.trim() ?? "",
	);

	const loadSettings = useCallback(async () => {
		try {
			const res = await settingsService.getPublicSettings();
			setSettings(res?.data ?? null);
		} catch {
			// Storefront must keep working if settings cannot be loaded; the UI
			// simply falls back to its built-in defaults.
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		if (initialSettings === null) loadSettings();
	}, [initialSettings, loadSettings]);

	useEffect(() => {
		const stored = getStoredCurrency();
		if (stored) setDisplayCurrencyState(stored);
	}, []);

	const setDisplayCurrency = useCallback((code: string) => {
		const normalized = code.trim().toUpperCase();
		if (!normalized) return;
		setDisplayCurrencyState(normalized);
		rememberCurrencyChoice(normalized);
	}, []);

	// Keep storefront data fresh after an admin save and when returning to the tab.
	useEffect(() => {
		const refresh = () => {
			if (document.visibilityState === "visible") loadSettings();
		};
		const onSettingsUpdated = () => loadSettings();

		window.addEventListener("focus", refresh);
		document.addEventListener("visibilitychange", refresh);
		window.addEventListener(PUBLIC_SETTINGS_UPDATED_EVENT, onSettingsUpdated);
		return () => {
			window.removeEventListener("focus", refresh);
			document.removeEventListener("visibilitychange", refresh);
			window.removeEventListener(PUBLIC_SETTINGS_UPDATED_EVENT, onSettingsUpdated);
		};
	}, [loadSettings]);

	return (
		<PublicSettingsContext.Provider
			value={{ settings, loading, displayCurrency, setDisplayCurrency }}
		>
			{children}
		</PublicSettingsContext.Provider>
	);
}

export function usePublicSettings() {
	return useContext(PublicSettingsContext);
}

/** Display currency from public settings; "" while settings are unavailable. */
export function useCurrency(): string {
	const { settings } = usePublicSettings();
	return settings?.currency?.trim() ?? "";
}

/** Visitor-selected display currency (falls back to the settings currency). */
export function useDisplayCurrency(): string {
	const { displayCurrency, settings } = usePublicSettings();
	return displayCurrency || settings?.currency?.trim() || "";
}
