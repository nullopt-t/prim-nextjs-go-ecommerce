"use client";

import { useEffect, useState } from "react";
import RecentlyCard from "@/modules/home/components/ui/recentlyCard";
import { useCatalogContext } from "@/modules/catalog/model/catalogContext";
import { getRecentlyViewedSlugs } from "@/shared/lib/recentlyViewed";
import { useLocale } from "next-intl";

export default function RecentlyGrid() {
	const { products } = useCatalogContext();
	const locale = useLocale();
	const [slugs, setSlugs] = useState<string[]>([]);

	useEffect(() => {
		setSlugs(getRecentlyViewedSlugs());
	}, []);

	const items = slugs
		.map((slug) => products.find((p: any) => p.slug === slug || p.id === slug))
		.filter(Boolean)
		.slice(0, 6);

	if (items.length === 0) {
		return null;
	}

	return (
		<div className="flex gap-4 sm:gap-6 mt-4 overflow-x-auto pb-2 scrollbar-none">
			{items.map((item: any) => {
				const itemTitle =
					locale === "ar"
						? item.product?.ar || item.title || item.product?.en
						: item.product?.en || item.title || item.product?.ar;

				return (
					<div key={item.id || item.slug} className="shrink-0">
						<RecentlyCard title={itemTitle} img={item.img} slug={item.slug || item.id} />
					</div>
				);
			})}
		</div>
	);
}
