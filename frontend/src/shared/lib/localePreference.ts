/** Cookie read by the next-intl middleware to resolve the locale server-side. */
export const LOCALE_COOKIE_NAME = "NEXT_LOCALE";
/** Set only when the visitor explicitly switches language (not by middleware). */
export const EXPLICIT_LOCALE_COOKIE_NAME = "locale";
const LOCALE_COOKIE_MAX_AGE = 60 * 60 * 24 * 365;

/** Persists a language explicitly chosen by the visitor. */
export function rememberLocaleChoice(locale: string) {
	if (typeof window === "undefined") return;
	if (locale !== "ar" && locale !== "en") return;

	const secure = window.location.protocol === "https:" ? "; Secure" : "";
	document.cookie = `${LOCALE_COOKIE_NAME}=${locale}; Path=/; Max-Age=${LOCALE_COOKIE_MAX_AGE}; SameSite=Lax${secure}`;
	document.cookie = `${EXPLICIT_LOCALE_COOKIE_NAME}=${locale}; Path=/; Max-Age=${LOCALE_COOKIE_MAX_AGE}; SameSite=Lax${secure}`;
}
