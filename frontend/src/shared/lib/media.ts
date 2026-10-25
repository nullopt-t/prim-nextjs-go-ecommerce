/**
 * Normalizes media/image URLs so they are identical on the server and in the
 * browser. Using only `NEXT_PUBLIC_*` env values here matters: deriving the
 * host from `window.location` during hydration produces different markup on
 * the server and client.
 *
 * Set `NEXT_PUBLIC_MEDIA_BASE_URL` to the public object-storage origin (for
 * example `https://cdn.example.com`). When it is not set, the API hostname with
 * the default MinIO port is used.
 */
export const PLACEHOLDER_PRODUCT_IMAGE = "/placeholder-product.png";

function getMediaOrigin(): string {
	const explicit = process.env.NEXT_PUBLIC_MEDIA_BASE_URL;
	if (explicit) return explicit.replace(/\/$/, "");

	const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8081";
	try {
		const parsed = new URL(apiUrl);
		if (parsed.hostname) {
			return `${parsed.protocol}//${parsed.hostname}:9000`;
		}
	} catch {
		// Fall through to the local default.
	}
	return "http://localhost:9000";
}

export function normalizeMediaUrl(url?: string | null): string {
	if (!url) return PLACEHOLDER_PRODUCT_IMAGE;

	const result = url.trim();
	if (!result) return PLACEHOLDER_PRODUCT_IMAGE;

	// Absolute URLs are passed through untouched.
	if (/^https?:\/\//i.test(result) || result.startsWith("data:") || result.startsWith("blob:")) {
		return result;
	}

	// App-served assets (placeholders, icons, bundled images) stay same-origin.
	if (
		result.startsWith("/placeholder") ||
		result.startsWith("/icons/") ||
		result.startsWith("/images/") ||
		result.startsWith("/favicon")
	) {
		return result;
	}

	// Root-relative storage paths and bare relative paths both live in the
	// `catalog` bucket.
	const path = result.startsWith("/")
		? result
		: `/catalog/${result.replace(/^\/?catalog\//, "")}`;

	return `${getMediaOrigin()}${path}`;
}
