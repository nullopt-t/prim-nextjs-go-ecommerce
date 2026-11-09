"use client";

import { getErrorMessageByCode } from "@/shared/i18n/errors";

export interface ToastOptions {
  duration?: number;
  id?: string | number;
  description?: React.ReactNode;
  action?: {
    label: string;
    onClick?: () => void;
  };
  [key: string]: unknown;
}

/**
 * Centralized notification and error dispatching.
 *
 * Requirements:
 * 1. No toasts: Floating toasts are completely removed from the site.
 * 2. Success actions are trusted and silent: No success toasts or banners.
 * 3. Errors are embedded in the page: Errors dispatch to in-page alert banners
 *    and local component error states, ensuring actions never fail silently.
 */
export const notify = {
  success: (_message: string, _options?: ToastOptions) => {
    // Succession is silent across the entire site. No toasts or popups.
  },
  info: (_message: string, _options?: ToastOptions) => {
    // Info notifications are silent.
  },
  error: (message: string, _options?: ToastOptions) => {
    if (typeof window !== "undefined" && message) {
      window.dispatchEvent(
        new CustomEvent("prim:embedded-error", { detail: { message } })
      );
    }
  },
  clearError: () => {
    if (typeof window !== "undefined") {
      window.dispatchEvent(
        new CustomEvent("prim:embedded-error", { detail: { message: null } })
      );
    }
  },
};

/**
 * Normalizes the various error shapes produced across the codebase into a
 * single human-readable message, falling back to `fallback` when nothing
 * useful is present.
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
