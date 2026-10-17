"use client";

import { toast } from "sonner";

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
 * Handles:
 * - `Error` thrown by `@/api/client` (`.message` holds the server message)
 * - service result objects that return `{ error }` instead of throwing
 * - plain strings
 * - axios-style `{ response: { data: { message | error } } }`
 */
export function apiErrorMessage(err: unknown, fallback: string): string {
  if (typeof err === "string") {
    return err.trim() || fallback;
  }

  if (err && typeof err === "object") {
    const e = err as Record<string, any>;
    const response = e.response?.data;
    const candidates = [e.error, e.message, response?.message, response?.error];

    for (const candidate of candidates) {
      if (typeof candidate === "string" && candidate.trim()) {
        return candidate;
      }
    }
  }

  return fallback;
}
