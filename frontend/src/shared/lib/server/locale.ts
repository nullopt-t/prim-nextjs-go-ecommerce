/**
 * Server-side locale resolution.
 *
 * Precedence (documented in AGENTS.md):
 * 1. Explicit visitor choice stored in the short `locale` cookie wins.
 * 2. Otherwise the store's configured `defaultLocale` applies.
 *
 * These helpers are pure so the precedence rules can be unit-tested without
 * mocking Next.js headers/cookies/redirect.
 */
export type SupportedLocale = "en" | "ar";

export const SUPPORTED_LOCALES: SupportedLocale[] = ["en", "ar"];

/**
 * Resolves the effective locale following the documented precedence. Returns
 * `undefined` when neither the explicit choice nor the store default is a
 * supported locale (no redirect should happen in that case).
 */
export function resolveTargetLocale(
	explicitLocale: string | undefined,
	storeDefault: string | undefined,
): SupportedLocale | undefined {
	// An explicit choice only counts when it is a supported locale; otherwise
	// fall back to the store default (matching the previous `chosenLocale ??
	// storeDefault` behavior).
	const explicit =
		explicitLocale === "en" || explicitLocale === "ar"
			? explicitLocale
			: undefined;
	const resolved = explicit ?? storeDefault;
	return resolved === "en" || resolved === "ar" ? resolved : undefined;
}

/**
 * Determines whether the layout should redirect to add a locale prefix.
 * Redirects only when the asked path has no locale prefix and the resolved
 * target locale differs from the current locale.
 */
export function shouldAddLocalePrefix(
	locale: SupportedLocale,
	askedPath: string,
	targetLocale: SupportedLocale | undefined,
): boolean {
	if (!targetLocale) return false;
	const hasLocalePrefix = /^\/(en|ar)(?=\/|$)/.test(askedPath);
	return !hasLocalePrefix && targetLocale !== locale;
}
