import { useTranslations } from "next-intl";
import { BadgePercent } from "lucide-react";
import { formatMoney } from "@/shared/lib/money";
import { resolveTier, type TierPricing } from "../../utils/tierPricing";
import { ProductSection } from "./ProductSection";

interface TierPricingTableProps {
	tiers: TierPricing[];
	baseUnitPrice: number;
	quantity: number;
	currency: string;
}

/**
 * Volume-discount reference table, rendered on the PDP when tier data exists
 * on the product/variant. The row matching the selected quantity is
 * highlighted so shoppers see their active saving at a glance.
 */
export function TierPricingTable({ tiers, baseUnitPrice, quantity, currency }: TierPricingTableProps) {
	const t = useTranslations("products");
	if (!tiers || tiers.length === 0) return null;

	const { activeTier } = resolveTier(baseUnitPrice, tiers, quantity);

	return (
		<ProductSection
			title={t("tierPricing.title")}
			icon={<BadgePercent className="size-4 text-accent-brand" aria-hidden="true" />}
		>
			<div className="overflow-x-auto rounded-xl border border-border/70">
				<table className="w-full text-start text-xs sm:text-sm">
					<thead>
						<tr className="bg-secondary/60 text-muted-foreground">
							<th className="px-3.5 py-2.5 text-start font-semibold">{t("tierPricing.quantity")}</th>
							<th className="px-3.5 py-2.5 text-start font-semibold">{t("tierPricing.unitPrice")}</th>
							<th className="px-3.5 py-2.5 text-start font-semibold">{t("tierPricing.savings")}</th>
						</tr>
					</thead>
					<tbody className="divide-y divide-border/70">
						{tiers.map((tier) => {
							const { unitPrice, savingsPerUnit } = resolveTier(baseUnitPrice, [tier], tier.minQuantity);
							const isActive = activeTier?.minQuantity === tier.minQuantity;
							return (
								<tr
									key={tier.minQuantity}
									className={
										isActive
											? "bg-accent-brand/10 font-semibold text-foreground"
											: "text-muted-foreground"
									}
								>
									<td className="px-3.5 py-2.5">{t("tierPricing.minQuantity", { count: tier.minQuantity })}</td>
									<td className="px-3.5 py-2.5">{formatMoney(unitPrice, currency)}</td>
									<td className="px-3.5 py-2.5">
										{tier.discountPercent !== undefined
											? t("tierPricing.savingsPercent", { percent: tier.discountPercent })
											: t("tierPricing.savingsAmount", { amount: formatMoney(savingsPerUnit, currency) })}
									</td>
								</tr>
							);
						})}
					</tbody>
				</table>
			</div>
		</ProductSection>
	);
}
