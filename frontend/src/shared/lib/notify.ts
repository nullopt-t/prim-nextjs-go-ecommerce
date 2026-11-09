"use client";

import { toast } from "sonner";
import { getErrorMessageByCode } from "@/shared/i18n/errors";

/**
 * Centralized toast helpers.
 *
 * Every user-facing notification in the app should go through `notify` so that
 * durations, dismissal and future behaviour (dedupe ids, analytics, …) stay
 * consistent across features. Prefer `apiErrorMessage` to turn anything the API
 * layer throws or returns into a displayable string.
 */
const DURATIONS = {
  success: 3000,
  error: 5000,
  info: 4000,
} as const;

type ToastOptions = Parameters<typeof toast.success>[1];

export const notify = {
  success: (message: string, options?: ToastOptions) =>
    toast.success(message, { duration: DURATIONS.success, ...options }),
  error: (message: string, options?: ToastOptions) =>
    toast.error(message, { duration: DURATIONS.error, ...options }),
  info: (message: string, options?: ToastOptions) =>
    toast.info(message, { duration: DURATIONS.info, ...options }),
};

/**
 * Normalizes the various error shapes produced across the codebase into a
 * single human-readable message, falling back to `fallback` when nothing
 * useful is present.
 *
 * Looks up error translations by canonical error code first (e.g., INSUFFICIENT_INVENTORY,
 * INVALID_ORDER_STATUS, COUPON_NOT_FOUND, UNAUTHORIZED, INVALID_CREDENTIALS)
 * before falling back to raw backend strings or fallback message.
 *
 * Handles:
 * - `Error` thrown by `@/shared/api/client` (`.code` and `.message`)
 * - service result objects that return `{ code, error }` instead of throwing
 * - plain strings (either error codes or messages)
 * - axios-style `{ response: { data: { code, message, error } } }`
 */
export function apiErrorMessage(err: unknown, fallback: string, locale?: string): string {
  if (typeof err === "string") {
    const trimmed = err.trim();
    const translated = getErrorMessageByCode(trimmed, locale);
    if (translated) return translated;
    return trimmed || fallback;
  }

  if (err && typeof err === "object") {
    const e = err as Record<string, unknown>;
    const response = (e.response as { data?: { code?: string; message?: string; error?: string } } | undefined)?.data;
    const dataObj = e.data as { code?: string; message?: string; error?: string } | undefined;

    // 1. Try resolving translated message via canonical error code
    const codeCandidates = [
      e.code,
      response?.code,
      dataObj?.code,
      e.errorCode,
      e.error_code,
    ];

    for (const code of codeCandidates) {
      if (typeof code === "string" && code.trim()) {
        const translated = getErrorMessageByCode(code, locale);
        if (translated) {
          return translated;
        }
      }
    }

    // 2. Check if error string itself is an error code
    const rawStringCandidates = [
      e.error as string | undefined,
      e.message as string | undefined,
      response?.message,
      response?.error,
      dataObj?.message,
      dataObj?.error,
    ];

    for (const candidate of rawStringCandidates) {
      if (typeof candidate === "string" && candidate.trim()) {
        const translated = getErrorMessageByCode(candidate.trim(), locale);
        if (translated) {
          return translated;
        }
      }
    }

    // 3. Fall back to raw backend message
    for (const candidate of rawStringCandidates) {
      if (typeof candidate === "string" && candidate.trim()) {
        return candidate.trim();
      }
    }
  }

  return fallback;
}
