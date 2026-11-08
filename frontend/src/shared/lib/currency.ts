/**
 * Multi-currency display support.
 *
 * Prices in the payload are in the store's base currency (public settings
 * `currency`). Display conversion uses static market estimates — good enough
 * to compare storefront prices, never for charging (the charged price is
 * always computed server-side). Refresh `UNITS_PER_EGP` whenever a live FX
 * source lands; nothing else needs to change.
 */

export const SUPPORTED_CURRENCIES = ["EGP", "USD", "EUR", "SAR", "AED"] as const;

export type SupportedCurrency = (typeof SUPPORTED_CURRENCIES)[number];

/** EGP value of one unit of each currency (display estimates). */
const UNITS_PER_EGP: Record<SupportedCurrency, number> = {
	EGP: 1,
	USD: 48.5,
	EUR: 52.0,
	SAR: 12.9,
	AED: 13.2,
};

export function isSupportedCurrency(code?: string | null): code is SupportedCurrency {
	if (!code) return false;
	return (SUPPORTED_CURRENCIES as readonly string[]).includes(code.trim().toUpperCase());
}

/** Normalizes a supported code; returns null for unsupported/unknown input. */
export function normalizeCurrency(code?: string | null): SupportedCurrency | null {
	if (!code) return null;
	const upper = code.trim().toUpperCase();
	return isSupportedCurrency(upper) ? upper : null;
}

/**
 * Converts an amount between supported currencies via EGP. Unknown
 * currencies pass through unchanged (never guess a rate).
 */
export function convertCurrency(amount: number, from: string | null | undefined, to: string | null | undefined): number {
	const fromCode = normalizeCurrency(from);
	const toCode = normalizeCurrency(to);
	if (!Number.isFinite(amount) || !fromCode || !toCode || fromCode === toCode) return amount;
	const amountInEgp = amount * UNITS_PER_EGP[fromCode];
	return amountInEgp / UNITS_PER_EGP[toCode];
}

/** Cookie + localStorage persistence for the visitor's explicit choice. */
export const CURRENCY_COOKIE_NAME = "prim_currency";
const CURRENCY_COOKIE_MAX_AGE = 60 * 60 * 24 * 365;
const CURRENCY_STORAGE_KEY = "prim:currency";

export function rememberCurrencyChoice(code: string) {
	const normalized = normalizeCurrency(code);
	if (!normalized || typeof window === "undefined") return;
	const secure = window.location.protocol === "https:" ? "; Secure" : "";
	document.cookie = `${CURRENCY_COOKIE_NAME}=${normalized}; Path=/; Max-Age=${CURRENCY_COOKIE_MAX_AGE}; SameSite=Lax${secure}`;
	try {
		window.localStorage.setItem(CURRENCY_STORAGE_KEY, normalized);
	} catch {
		// Storage can be unavailable in private mode; the cookie still persists.
	}
}

export function getStoredCurrency(): SupportedCurrency | null {
	if (typeof window === "undefined") return null;
	try {
		const match = document.cookie.match(new RegExp(`(?:^|; )${CURRENCY_COOKIE_NAME}=([^;]*)`));
		const fromCookie = normalizeCurrency(match?.[1] ? decodeURIComponent(match[1]) : null);
		if (fromCookie) return fromCookie;
	} catch {
		// Fall through to localStorage.
	}
	try {
		return normalizeCurrency(window.localStorage.getItem(CURRENCY_STORAGE_KEY));
	} catch {
		return null;
	}
}
