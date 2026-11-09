"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { Tag } from "lucide-react";
import ProductShelf from "./productShelf";
import type { ProductDetailsEntity } from "@/modules/catalog/types";

function isOnSale(p: ProductDetailsEntity): boolean {
	const pct = Number(String(p.discountPercentage ?? "").replace("%", ""));
	if (pct > 0) return true;
	const price = p.extractedPrice ?? p.price;
	const old =
		p.extractedOriginalPrice ??
		p.oldPrice ??
		(p.originalPrice ? parseFloat(String(p.originalPrice).replace(/[^0-9.-]+/g, "")) : undefined);
	return typeof old === "number" && typeof price === "number" && price > 0 && old > price;
}

export default function OnSaleShelf({ products = [] }: { products?: ProductDetailsEntity[] }) {
	const t = useTranslations("home.onSale");

	const deals = useMemo(() => products.filter(isOnSale), [products]);

	if (deals.length === 0) return null;

	return (
		<ProductShelf
			title={t("title")}
			subtitle={t("subtitle")}
			badge={
				<span className="inline-flex shrink-0 items-center gap-1 rounded-full bg-red-600 px-2.5 py-1 text-[11px] font-bold uppercase tracking-wide text-white">
					<Tag className="size-3" />
					{t("badge")}
				</span>
			}
			link="/products"
			products={deals}
		/>
	);
}
