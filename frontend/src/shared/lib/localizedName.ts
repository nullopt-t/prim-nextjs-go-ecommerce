/**
 * Resolve a localized name (backend `{ en, ar }` object or plain string) for
 * the active locale. Falls back to the other locale, then to `fallback`.
 */
export function localizedName(
	name: string | { en?: string; ar?: string } | undefined | null,
	locale: string,
	fallback = "—",
): string {
	if (typeof name === "string") return name.trim() ? name : fallback;
	if (!name) return fallback;
	const value = locale === "ar" ? name.ar || name.en : name.en || name.ar;
	return value?.trim() ? value : fallback;
}
