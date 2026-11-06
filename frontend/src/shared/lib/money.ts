/**
 * Money formatting for the storefront.
 *
 * Prices are rendered as "{currency} {amount}" using the currency from public
 * settings (or the cart/product payload). When no currency is known the amount
 * is rendered bare instead of guessing a symbol.
 */
export function formatMoney(
	value: number | undefined | null,
	currency?: string | null,
): string {
	if (value === undefined || value === null || Number.isNaN(value)) return "—";
	const amount = value.toFixed(2);
	const code = currency?.trim() ?? "";
	return code ? `${code} ${amount}` : amount;
}

/** Trimmed currency code from settings/payload, or "" when unknown. */
export function resolveCurrency(currency?: string | null): string {
	return currency?.trim() ?? "";
}
