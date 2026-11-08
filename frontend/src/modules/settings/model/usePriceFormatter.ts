"use client";

import { useCallback } from "react";
import { useCurrency, useDisplayCurrency } from "@/modules/settings/model/publicSettingsContext";
import { convertCurrency } from "@/shared/lib/currency";
import { formatMoney } from "@/shared/lib/money";

/**
 * Formats a base-currency amount in the visitor's selected display currency.
 * The charged price is unaffected (server-side); this is display-only.
 */
export function usePriceFormatter() {
	const baseCurrency = useCurrency();
	const displayCurrency = useDisplayCurrency();
	return useCallback(
		(amount: number | string | undefined | null, from?: string): string => {
			const numeric = typeof amount === "string" ? Number(amount) : amount;
			const converted = convertCurrency(
				typeof numeric === "number" && Number.isFinite(numeric) ? numeric : 0,
				(from ?? baseCurrency) || null,
				displayCurrency || null,
			);
			return formatMoney(converted, displayCurrency);
		},
		[baseCurrency, displayCurrency],
	);
}
