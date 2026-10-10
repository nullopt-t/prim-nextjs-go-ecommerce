/**
 * Normalizes media/image URLs so that they work consistently across devices:
 * - Replaces hardcoded localhost / 127.0.0.1 MinIO URLs with the client host or NEXT_PUBLIC_API_URL host.
 * - Resolves relative catalog paths (e.g., "products/xyz.jpg" or "/products/xyz.jpg").
 */
export function normalizeMediaUrl(url?: string | null): string {
  if (!url) return "/placeholder-product.png";

  let result = url.trim();

  // If client-side and URL points to localhost:9000 / 127.0.0.1:9000,
  // rewrite hostname to match window.location.hostname so phones on the local network can load it.
  if (typeof window !== "undefined" && window.location.hostname && window.location.hostname !== "localhost") {
    result = result.replace(/http:\/\/(localhost|127\.0\.0\.1):9000/g, `http://${window.location.hostname}:9000`);
  } else if (process.env.NEXT_PUBLIC_API_URL) {
    try {
      const parsedApi = new URL(process.env.NEXT_PUBLIC_API_URL);
      if (parsedApi.hostname && parsedApi.hostname !== "localhost") {
        result = result.replace(/http:\/\/(localhost|127\.0\.0\.1):9000/g, `http://${parsedApi.hostname}:9000`);
      }
    } catch {
      // Ignore URL parse failure
    }
  }

  // If path is relative and not beginning with /
  if (!result.startsWith("http://") && !result.startsWith("https://") && !result.startsWith("/")) {
    const host =
      (typeof window !== "undefined" && window.location.hostname) ||
      (process.env.NEXT_PUBLIC_API_URL ? new URL(process.env.NEXT_PUBLIC_API_URL).hostname : "localhost");
    result = `http://${host}:9000/catalog/${result.replace(/^\/?catalog\//, "")}`;
  }

  return result;
}
