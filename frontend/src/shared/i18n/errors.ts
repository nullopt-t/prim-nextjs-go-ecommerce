import enMessages from "../../../messages/en.json";
import arMessages from "../../../messages/ar.json";

export type CanonicalErrorCode =
  | "INSUFFICIENT_INVENTORY"
  | "INVALID_ORDER_STATUS"
  | "COUPON_NOT_FOUND"
  | "UNAUTHORIZED"
  | "INVALID_CREDENTIALS";

export const ERROR_MESSAGES: Record<"en" | "ar", Record<string, string>> = {
  en: enMessages.errors,
  ar: arMessages.errors,
};

/**
 * Resolves active browser locale from path or DOM attributes.
 * Defaults to "en" when running on server or without explicit locale.
 */
export function getActiveLocale(): "en" | "ar" {
  if (typeof window !== "undefined") {
    const pathname = window.location.pathname;
    if (pathname.startsWith("/ar") || pathname === "/ar") {
      return "ar";
    }
    if (
      document.documentElement.lang === "ar" ||
      document.documentElement.dir === "rtl"
    ) {
      return "ar";
    }
  }
  return "en";
}

/**
 * Looks up translated error message by canonical backend error code.
 * Returns translated message in target locale, falls back to EN, or undefined if not found.
 */
export function getErrorMessageByCode(
  code: string | undefined | null,
  locale?: string
): string | undefined {
  if (!code || typeof code !== "string") return undefined;
  const normalized = code.trim().toUpperCase();
  const currentLocale = (locale || getActiveLocale()) === "ar" ? "ar" : "en";
  return ERROR_MESSAGES[currentLocale]?.[normalized] || ERROR_MESSAGES.en?.[normalized];
}
