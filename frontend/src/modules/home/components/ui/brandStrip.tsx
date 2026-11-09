"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/shared/i18n/navigation";
import type { ProductDetailsEntity } from "@/modules/catalog/types";

export default function BrandStrip({ products = [] }: { products?: ProductDetailsEntity[] }) {
	const t = useTranslations("home.brands");

	const brands = useMemo(() => {
		const counts = new Map<string, number>();
		for (const p of products) {
			const name = typeof p.brand === "string" ? p.brand : p.brand?.name;
			if (!name) continue;
			counts.set(name, (counts.get(name) ?? 0) + 1);
		}
		return [...counts.entries()].map(([name, count]) => ({ name, count })).slice(0, 8);
	}, [products]);

	if (brands.length === 0) return null;

	return (
		<section className="space-y-3">
			<div className="text-start">
				<h2 className="text-base sm:text-lg md:text-xl font-bold text-foreground">{t("title")}</h2>
				<p className="text-[11px] sm:text-xs text-muted-foreground mt-0.5">{t("subtitle")}</p>
			</div>
			<div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-3 sm:gap-4">
				{brands.map((brand) => (
					<Link
						key={brand.name}
						href={`/products?brand=${encodeURIComponent(brand.name)}`}
						className="flex flex-col items-center gap-2 rounded-2xl border border-border/80 bg-card p-4 text-center shadow-2xs hover:border-accent-brand/50 transition-colors"
					>
						<span
							aria-hidden
							className="inline-flex size-11 items-center justify-center rounded-full bg-accent-brand/10 text-lg font-bold text-accent-brand"
						>
							{brand.name.charAt(0)}
						</span>
						<span dir="auto" className="text-sm font-semibold text-foreground line-clamp-1">
							{brand.name}
						</span>
						<span className="text-[11px] text-muted-foreground tabular-nums">
							{t("productCount", { count: brand.count })}
						</span>
					</Link>
				))}
			</div>
		</section>
	);
}
