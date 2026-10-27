/**
 * Resolve the API base URL for server-side calls.
 *
 * Next.js inlines every `process.env.NEXT_PUBLIC_*` reference at build time,
 * including the references in server bundles. In a container deployment that
 * bakes the browser-facing host (for example the host LAN IP) into the server
 * bundle, where it may be unreachable, so SSR fetches hang until the TCP
 * connect times out.
 *
 * Reading the value through a variable key keeps the lookup dynamic, so the
 * server bundle reads the runtime environment instead of the build-time value.
 * The browser bundle still gets the build-time value through the direct
 * references in `src/shared/api/client.ts`.
 */
const API_BASE_URL_KEY = "NEXT_PUBLIC_API_URL";

export function getServerApiBaseUrl(): string | null {
	const value = process.env[API_BASE_URL_KEY];
	if (!value) return null;
	return value.replace(/\/$/, "");
}
