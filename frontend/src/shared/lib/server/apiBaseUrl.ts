/**
 * Resolve the API base URL for server-side calls.
 *
 * Next.js inlines every `process.env.NEXT_PUBLIC_*` reference at build time,
 * including the references in server bundles. In a container deployment that
 * bakes the browser-facing host (for example the host LAN IP) into the server
 * bundle, where it may be unreachable, so SSR fetches hang until the TCP
 * connect times out.
 *
 * `API_URL` is the optional server-only override (the Compose network address,
 * e.g. `http://api:8081`). `NEXT_PUBLIC_API_URL` remains the single public
 * browser-facing value and is the fallback when no override is set.
 *
 * Reading through variable keys keeps both lookups dynamic, so the server
 * bundle reads the runtime environment instead of the build-time value. The
 * browser bundle still gets the build-time `NEXT_PUBLIC_API_URL` through the
 * direct references in `src/shared/api/client.ts`.
 */
const SERVER_API_URL_KEY = "API_URL";
const PUBLIC_API_URL_KEY = "NEXT_PUBLIC_API_URL";

export function getServerApiBaseUrl(): string | null {
	const value =
		process.env[SERVER_API_URL_KEY] || process.env[PUBLIC_API_URL_KEY];
	if (!value) return null;
	return value.replace(/\/$/, "");
}
