"use client";

import { ShieldCheck, CreditCard, Banknote, Truck, Clock } from "lucide-react";
import { useTranslations } from "next-intl";
import { usePublicSettings } from "@/modules/settings/model/publicSettingsContext";
import { formatMoney } from "@/shared/lib/money";

export function TrustBar() {
	const t = useTranslations("common.trustBar");
	const { settings } = usePublicSettings();

	const threshold = settings?.freeShippingThreshold ?? 0;
	const deliveryDays = settings?.estimatedDeliveryDays?.trim();

	return (
		<div className="mb-8 flex flex-wrap items-center gap-3 rounded-xl border border-border/70 bg-muted/30 px-4 py-3 text-xs text-muted-foreground">
			<span className="inline-flex items-center gap-1.5 font-semibold text-foreground">
				<ShieldCheck className="size-4" aria-hidden="true" />
				{t("secureCheckout")}
			</span>
			{settings?.allowCards && (
				<span className="inline-flex items-center gap-1.5">
					<CreditCard className="size-4" aria-hidden="true" />
					{t("cards")}
				</span>
			)}
			{settings?.allowCod && (
				<span className="inline-flex items-center gap-1.5">
					<Banknote className="size-4" aria-hidden="true" />
					{t("cod")}
				</span>
			)}
			{threshold > 0 && (
				<span className="inline-flex items-center gap-1.5">
					<Truck className="size-4" aria-hidden="true" />
					{t("freeDelivery", { amount: formatMoney(threshold, settings?.currency?.trim() ?? "") })}
				</span>
			)}
			{deliveryDays && (
				<span className="inline-flex items-center gap-1.5">
					<Clock className="size-4" aria-hidden="true" />
					{t("deliveryDays", { days: deliveryDays })}
				</span>
			)}
		</div>
	);
}
