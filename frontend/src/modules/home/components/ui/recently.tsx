"use client";

import { useCallback, useEffect, useState } from "react";
import RecentlyGrid from "@/modules/home/components/ui/recentlyGrid";
import { useTranslations, useLocale } from "next-intl";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { useCatalogContext } from "@/modules/catalog/model/catalogContext";
import { historyService } from "@/modules/history/api/historyService";
import type { ViewedProduct } from "@/modules/history/types";
import {
	clearRecentlyViewed,
	getRecentlyViewedSlugs,
	removeRecentlyViewed,
	RECENTLY_VIEWED_UPDATED_EVENT,
} from "@/shared/lib/recentlyViewed";
import { notify } from "@/shared/lib/notify";

const RECENT_LIMIT = 6;

export default function Recently() {
	const t = useTranslations("home");
	const locale = useLocale();
	const { user, loading: authLoading } = useAuthContext();
	const { products } = useCatalogContext();
	const [items, setItems] = useState<ViewedProduct[]>([]);

	const load = useCallback(async () => {
		if (authLoading) return;

		if (user) {
			try {
				const response = await historyService.list(RECENT_LIMIT);
				setItems(response.data ?? []);
				return;
			} catch {
				// Fall through to the guest cache if the account history is unavailable.
			}
		}

		const localItems = getRecentlyViewedSlugs()
			.map((slug): ViewedProduct | null => {
				const product = products.find(
					(candidate) => candidate.slug === slug || candidate.id === slug
				);
				if (!product) return null;

				const localized =
					typeof product.product === "object" ? product.product : undefined;
				const title =
					locale === "ar"
						? localized?.ar || product.title || localized?.en
						: localized?.en || product.title || localized?.ar;

				return {
					productId: String(product.id),
					title: title || product.name || product.sku || "",
					slug: product.slug || String(product.id),
					basePrice: product.extractedPrice ?? product.price ?? 0,
					currency: product.currency ?? "",
					thumbnailUrl: product.img || product.images?.[0],
				};
			})
			.filter((item): item is ViewedProduct => item !== null)
			.slice(0, RECENT_LIMIT);

		setItems(localItems);
	}, [authLoading, locale, products, user]);

	useEffect(() => {
		void load();
	}, [load]);

	useEffect(() => {
		const handleUpdate = () => {
			void load();
		};
		window.addEventListener(RECENTLY_VIEWED_UPDATED_EVENT, handleUpdate);
		return () => {
			window.removeEventListener(RECENTLY_VIEWED_UPDATED_EVENT, handleUpdate);
		};
	}, [load]);

	const handleRemove = useCallback(
		async (item: ViewedProduct) => {
			setItems((prev) => prev.filter((i) => i.productId !== item.productId));
			try {
				if (user) {
					await historyService.remove(item.productId);
				} else {
					removeRecentlyViewed(item.slug);
				}
			} catch {
				// Best-effort: optimistic removal stands.
			}
			notify.info(t("recentlyViewed.removed", { name: item.title }));
		},
		[t, user],
	);

	const handleClear = useCallback(async () => {
		setItems([]);
		try {
			if (user) {
				await historyService.clear();
			} else {
				clearRecentlyViewed();
			}
		} catch {
			// Best-effort: optimistic clear stands.
		}
		notify.success(t("recentlyViewed.cleared"));
	}, [t, user]);

	if (items.length === 0) {
		return null;
	}

	return (
		<div className="w-full px-3 sm:px-6 lg:px-8 mb-5 sm:mb-6">
			<div className="bg-card/40 border border-border/80 rounded-2xl p-3.5 sm:p-5 shadow-xs">
				<div className="flex items-center justify-between pb-2.5 border-b border-border/50">
					<p className="text-foreground font-bold text-sm sm:text-base">
						{t("recentlyViewed.title")}
					</p>
					<button
						type="button"
						onClick={() => void handleClear()}
						className="cursor-pointer text-xs font-medium text-accent-brand hover:underline"
					>
						{t("recentlyViewed.clear")}
					</button>
				</div>
				<RecentlyGrid items={items} onRemove={handleRemove} />
			</div>
		</div>
	);
}
