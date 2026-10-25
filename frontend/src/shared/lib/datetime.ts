/** Coerces a value to a valid Date, or returns null for missing/invalid input. */
function toDate(value?: string | number | Date | null): Date | null {
  if (value === null || value === undefined || value === "") return null;
  const d = value instanceof Date ? value : new Date(value);
  return Number.isNaN(d.getTime()) ? null : d;
}

/**
 * Formats a timestamp as a localized date and time, e.g. "Oct 9, 2026, 3:45 PM".
 * Use for anything backed by a timestamp column (*_at); returns "" for
 * missing/invalid values so callers can fall back to a placeholder.
 */
export function formatDateTime(
  value?: string | number | Date | null,
  locale?: string,
): string {
  const d = toDate(value);
  if (!d) return "";
  return d.toLocaleString(locale, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}
