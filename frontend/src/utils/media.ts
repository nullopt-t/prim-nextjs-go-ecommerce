/**
 * Normalizes media/image URLs so that they work consistently across devices:
 * - Replaces hardcoded localhost / 127.0.0.1 MinIO URLs with the client host or NEXT_PUBLIC_API_URL host.
 * - Resolves relative catalog paths (e.g., "products/xyz.jpg" or "/products/xyz.jpg").
 */
export function normalizeMediaUrl(url?: string | null): string {
  if (!url) return "/placeholder-product.png";

  let result = url.trim();

  // If client-side and URL points to MinIO port :9000,
  // align hostname to current window.location.hostname so images load whether accessed via
  // localhost, 127.0.0.1, or local network IP (192.168.x.x)
  if (typeof window !== "undefined" && window.location.hostname) {
    result = result.replace(/https?:\/\/[^/:]+:9000/g, `${window.location.protocol}//${window.location.hostname}:9000`);
  } else if (process.env.NEXT_PUBLIC_API_URL) {
    try {
      const parsedApi = new URL(process.env.NEXT_PUBLIC_API_URL);
      if (parsedApi.hostname) {
        result = result.replace(/https?:\/\/[^/:]+:9000/g, `${parsedApi.protocol}//${parsedApi.hostname}:9000`);
      }
    } catch {
      // Ignore URL parse failure
    }
  }

  // If path is a root-relative storage bucket path (e.g. /product-thumbnails/... or /catalog/...)
  if (result.startsWith("/")) {
    if (result.startsWith("/placeholder") || result.startsWith("/icons/") || result.startsWith("/images/")) {
      return result;
    }
    const host =
      (typeof window !== "undefined" && window.location.hostname) ||
      (process.env.NEXT_PUBLIC_API_URL ? new URL(process.env.NEXT_PUBLIC_API_URL).hostname : "localhost");
    const protocol = (typeof window !== "undefined" && window.location.protocol) || "http:";
    return `${protocol}//${host}:9000${result}`;
  }

  // If path is relative without leading slash (e.g. "catalog/xyz.jpg" or "xyz.jpg")
  if (!result.startsWith("http://") && !result.startsWith("https://")) {
    const host =
      (typeof window !== "undefined" && window.location.hostname) ||
      (process.env.NEXT_PUBLIC_API_URL ? new URL(process.env.NEXT_PUBLIC_API_URL).hostname : "localhost");
    const protocol = (typeof window !== "undefined" && window.location.protocol) || "http:";
    result = `${protocol}//${host}:9000/catalog/${result.replace(/^\/?catalog\//, "")}`;
  }

  return result;
}
