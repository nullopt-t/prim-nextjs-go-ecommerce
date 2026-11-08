/**
 * True only for absolute http(s) URLs. Order/shipment payloads carry
 * server-stored tracking URLs; rendering anything else as an <a href> would
 * allow javascript: or data: payloads to execute on click.
 */
export function isSafeUrl(url?: string | null): url is string {
	return !!url && /^https?:\/\//i.test(url.trim());
}
