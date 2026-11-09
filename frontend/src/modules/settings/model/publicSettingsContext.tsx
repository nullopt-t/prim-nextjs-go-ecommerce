"use client";

import { createContext, useCallback, useContext, useEffect, useState } from "react";
import {
	settingsService,
	type PublicSettings,
	type BackendCurrency,
} from "@/modules/settings/api/settingsService";
import { PUBLIC_SETTINGS_UPDATED_EVENT } from "@/shared/lib/settingsEvents";
import {
	getStoredCurrency,
	rememberCurrencyChoice,
	setBackendCurrencies,
} from "@/shared/lib/currency";

interface PublicSettingsContextValue {
	/** The storefront settings, or null while loading / on failure. */
	settings: PublicSettings | null;
	currencies: BackendCurrency[];
	loading: boolean;
	/** Visitor-selected display currency; falls back to the settings currency. */
	displayCurrency: string;
	setDisplayCurrency: (code: string) => void;
}

const PublicSettingsContext = createContext<PublicSettingsContextValue>({
	settings: null,
	currencies: [],
	loading: true,
	displayCurrency: "",
	setDisplayCurrency: () => {},
});

export function PublicSettingsProvider({
	children,
	initialSettings = null,
	initialDisplayCurrency,
}: {
	children: React.ReactNode;
	initialSettings?: PublicSettings | null;
	initialDisplayCurrency?: string;
}) {
	const [settings, setSettings] = useState<PublicSettings | null>(initialSettings);
	const [currencies, setCurrencies] = useState<BackendCurrency[]>([]);
	const [loading, setLoading] = useState(initialSettings === null);
	// SSR renders the store's base currency or visitor's cookie currency;
	// any localStorage choice syncs on mount without flash.
	const [displayCurrency, setDisplayCurrencyState] = useState<string>(
		initialDisplayCurrency || initialSettings?.currency?.trim() || "",
	);

	const loadSettings = useCallback(async () => {
		try {
			const [settingsRes, currenciesRes] = await Promise.allSettled([
				settingsService.getPublicSettings(),
				settingsService.getCurrencies(),
			]);
			if (settingsRes.status === "fulfilled" && settingsRes.value?.data) {
				setSettings(settingsRes.value.data);
			}
			if (currenciesRes.status === "fulfilled" && currenciesRes.value?.data) {
				setCurrencies(currenciesRes.value.data);
				setBackendCurrencies(currenciesRes.value.data);
			}
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
			value={{ settings, currencies, loading, displayCurrency, setDisplayCurrency }}
		>
			{children}
		</PublicSettingsContext.Provider>
	);
}

export function usePublicSettings() {
	return useContext(PublicSettingsContext);
}

/** Supported currencies and live exchange rates from the backend. */
export function useCurrencies(): BackendCurrency[] {
	const { currencies } = usePublicSettings();
	return currencies;
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
