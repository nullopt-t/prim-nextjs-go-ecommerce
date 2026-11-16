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
			return `${parsed.protocol}//${parsed.hostname}:4566`;
		}
	} catch {
		// Fall through to the local default.
	}
	return "http://localhost:4566";
}

export function normalizeMediaUrl(url?: string | null): string {
	if (!url) return PLACEHOLDER_PRODUCT_IMAGE;

	let result = url.trim();
	if (!result) return PLACEHOLDER_PRODUCT_IMAGE;

	const mediaOrigin = getMediaOrigin();

	// Rewrite legacy MinIO port 9000 or container-internal hostnames (minio, floci) to the public media origin.
	result = result.replace(/^https?:\/\/(localhost|127\.0\.0\.1|minio):9000/, mediaOrigin);
	result = result.replace(/^https?:\/\/(localhost|127\.0\.0\.1):4566/, mediaOrigin);
	result = result.replace(/^https?:\/\/floci:4566/, mediaOrigin);

	// In the browser on a mobile or remote LAN device, replace localhost with current hostname
	if (
		typeof window !== "undefined" &&
		window.location.hostname &&
		window.location.hostname !== "localhost" &&
		window.location.hostname !== "127.0.0.1"
	) {
		result = result.replace(
			/^http:\/\/(localhost|127\.0\.0\.1):4566/,
			`${window.location.protocol}//${window.location.hostname}:4566`
		);
	}

	// Absolute URLs are passed through untouched.
	// Recognize S3, Cloudflare R2 and Floci/LocalStack (port 4566) storage URLs.
	if (
		/^https?:\/\//i.test(result) &&
		/(amazonaws\.com|r2\.cloudflarestorage\.com|:(4566|9000)\b)/.test(result)
	) {
		return result;
	}

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

export function getStoragePublicUrl(key: string): string {
	const trimmed = (key || "").trim();
	if (!trimmed) return PLACEHOLDER_PRODUCT_IMAGE;
	const path = trimmed.startsWith("/") ? trimmed : `/${trimmed}`;
	return `${getMediaOrigin()}${path}`;
}
