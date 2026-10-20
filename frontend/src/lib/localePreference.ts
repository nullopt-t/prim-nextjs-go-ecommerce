export const LOCALE_CHOICE_KEY = "prim_locale_choice";

/** Persists a language explicitly chosen by the visitor. */
export function rememberLocaleChoice(locale: string) {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(LOCALE_CHOICE_KEY, locale);
}

/** Returns the visitor's explicit language choice, if any. */
export function getRememberedLocaleChoice(): string | null {
  if (typeof window === "undefined") return null;
  const value = window.localStorage.getItem(LOCALE_CHOICE_KEY);
  return value === "ar" || value === "en" ? value : null;
}
