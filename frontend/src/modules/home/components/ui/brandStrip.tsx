"use client";

import { useMemo, useState } from "react";
import { useTranslations, useLocale } from "next-intl";
import { Link } from "@/shared/i18n/navigation";
import type { Category, ProductDetailsEntity } from "@/modules/catalog/types";
import { useCatalogContext } from "@/modules/catalog/model/catalogContext";
import { normalizeMediaUrl } from "@/shared/lib/media";
import { localizedName } from "@/shared/lib/localizedName";

type BrandCard = {
	name: string;
	logo?: string;
	count: number;
	categoryIds: string[];
};

function BrandLogo({ name, logo }: { name: string; logo?: string }) {
	const [failed, setFailed] = useState(false);

	if (!logo || failed) {
		return (
			<span className="flex h-20 w-full items-center justify-center rounded-2xl border border-border/40 bg-accent-brand/10">
				<span
					aria-hidden
					className="inline-flex size-12 items-center justify-center rounded-full bg-accent-brand/15 text-xl font-black text-accent-brand"
				>
					{name.charAt(0).toUpperCase()}
				</span>
			</span>
		);
	}

	return (
		<span className="flex h-20 w-full items-center justify-center overflow-hidden rounded-2xl border border-border/40 bg-white p-3.5 shadow-2xs transition-transform duration-200 group-hover:scale-105 dark:bg-white/95">
			{/* Brand logos come from arbitrary storage hosts, so next/image is not usable here. */}
			<img
				src={normalizeMediaUrl(logo)}
				alt=""
				loading="lazy"
				onError={() => setFailed(true)}
				className="h-12 w-auto max-w-[130px] object-contain"
			/>
		</span>
	);
}

export default function BrandStrip({
	products = [],
	categories = [],
}: {
	products?: ProductDetailsEntity[];
	categories?: Category[];
}) {
	const t = useTranslations("home.brands");
	const locale = useLocale();
	const { brands: catalogBrands } = useCatalogContext();

	// Category → brands present in the product feed. Mirrors the id/name
	// matching used by the category showcase cards.
	const categoryGroups = useMemo(() => {
		if (categories.length === 0 || products.length === 0) return [];

		return categories
			.map((category) => {
				const label = localizedName(category.name, locale);
				const enName = localizedName(category.name, "en", "");
				const brandNames = new Set<string>();

				for (const product of products) {
					const matchesCategory =
						product.categoryId === category.id ||
						(typeof product.category === "string" &&
							(product.category.toLowerCase() === label.toLowerCase() ||
								product.category.toLowerCase() === enName.toLowerCase()));
					if (!matchesCategory) continue;
					const brandName =
						typeof product.brand === "string" ? product.brand : product.brand?.name;
					if (brandName) brandNames.add(brandName);
				}

				return { id: category.id, label, brandNames };
			})
			.filter((group) => group.brandNames.size > 0);
	}, [categories, products, locale]);

	const brandCards = useMemo<BrandCard[]>(() => {
		const counts = new Map<string, number>();
		for (const product of products) {
			const name = typeof product.brand === "string" ? product.brand : product.brand?.name;
			if (name) counts.set(name, (counts.get(name) ?? 0) + 1);
		}

		// The catalog brand list carries logos and full product counts; brands
		// seen only in the feed are merged in so none disappear from the strip.
		const byName = new Map<string, { name: string; logo?: string; count?: number }>();
		for (const brand of catalogBrands) {
			byName.set(brand.name, {
				name: brand.name,
				logo: brand.logoUrl || brand.logo,
				count: brand.productCount,
			});
		}
		for (const [name, count] of counts) {
			const existing = byName.get(name);
			if (!existing) byName.set(name, { name, count });
			else if (existing.count === undefined) existing.count = count;
		}

		return [...byName.values()]
			.map((brand) => ({
				name: brand.name,
				logo: brand.logo,
				count: brand.count ?? counts.get(brand.name) ?? 0,
				categoryIds: categoryGroups
					.filter((group) => group.brandNames.has(brand.name))
					.map((group) => group.id),
			}))
			.sort((a, b) => b.count - a.count);
	}, [products, catalogBrands, categoryGroups]);

	const [activeCategoryId, setActiveCategoryId] = useState<string | null>(null);
	// Reset to "All" implicitly if the selected category disappears after refresh.
	const activeCategory = categoryGroups.some((group) => group.id === activeCategoryId)
		? activeCategoryId
		: null;
	const visibleBrands = (activeCategory
		? brandCards.filter((brand) => brand.categoryIds.includes(activeCategory))
		: brandCards
	).slice(0, 12);

	if (brandCards.length === 0) return null;

	const pillClass = (active: boolean) =>
		`shrink-0 cursor-pointer rounded-full border px-3.5 py-1.5 text-xs font-semibold transition-colors ${
			active
				? "border-accent-brand bg-accent-brand text-white"
				: "border-border/80 bg-card text-muted-foreground hover:border-accent-brand/50 hover:text-foreground"
		}`;

	return (
		<section className="space-y-3">
			<div className="flex items-start justify-between gap-3">
				<div className="text-start">
					<h2 className="text-base sm:text-lg md:text-xl font-bold text-foreground">{t("title")}</h2>
					<p className="text-[11px] sm:text-xs text-muted-foreground mt-0.5">{t("subtitle")}</p>
				</div>
				<Link
					href="/brands"
					className="shrink-0 pt-0.5 text-xs font-semibold text-accent-brand transition-colors hover:text-accent-brand/80 hover:underline sm:text-sm"
				>
					{t("viewAll")}
				</Link>
			</div>

			{categoryGroups.length > 0 && (
				<div className="scrollbar-none flex gap-2 overflow-x-auto pb-1">
					<button
						type="button"
						onClick={() => setActiveCategoryId(null)}
						aria-pressed={activeCategory === null}
						className={pillClass(activeCategory === null)}
					>
						{t("all")}
					</button>
					{categoryGroups.map((group) => (
						<button
							key={group.id}
							type="button"
							onClick={() => setActiveCategoryId(group.id)}
							aria-pressed={activeCategory === group.id}
							className={pillClass(activeCategory === group.id)}
						>
							{group.label}
						</button>
					))}
				</div>
			)}

			<div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4 sm:gap-5">
				{visibleBrands.map((brand) => (
					<Link
						key={brand.name}
						href={`/products?brand=${encodeURIComponent(brand.name)}`}
						className="group flex flex-col items-center justify-between rounded-2xl border border-border/80 bg-card p-4 sm:p-5 text-center shadow-xs transition-all duration-200 hover:-translate-y-1 hover:border-accent-brand/60 hover:shadow-md"
					>
						<BrandLogo name={brand.name} logo={brand.logo} />
						<span dir="auto" className="mt-3 text-sm sm:text-base font-bold text-foreground line-clamp-1 group-hover:text-accent-brand transition-colors">
							{brand.name}
						</span>
					</Link>
				))}
			</div>
		</section>
	);
}
