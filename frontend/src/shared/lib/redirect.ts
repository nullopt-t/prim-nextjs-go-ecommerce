/**
 * Returns a safe internal path for post-auth navigation. Protocol-relative
 * paths, absolute URLs, backslashes, and control characters are rejected.
 */
export function safeRedirectPath(raw: string | null | undefined): string {
  if (!raw) return "/";

  const value = raw;
  if (!value.startsWith("/")) return "/";
  if (value.startsWith("//") || value.startsWith("/\\")) return "/";
  if (value.includes("\\") || value.includes(":")) return "/";
  // Reject any control characters or whitespace that could alter parsing.
  if (/[\u0000-\u001f\u007f\s]/.test(value)) return "/";

  return value;
}
