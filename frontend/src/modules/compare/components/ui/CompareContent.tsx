"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { Scale } from "lucide-react";
import { Link } from "@/shared/i18n/navigation";
import { useCompare } from "@/modules/compare/model/useCompare";
import { catalogService } from "@/modules/catalog/api/catalogService";
import { normalizeProduct } from "@/modules/catalog/model/normalizeProduct";
import { getProductDisplayName } from "@/modules/catalog/utils/productHelpers";
import { usePriceFormatter } from "@/modules/settings/model/usePriceFormatter";
import { Stars } from "@/shared/ui/stars";
import { CatalogImage } from "@/shared/ui/catalogImage";
import type { ProductDetailsEntity } from "@/modules/catalog/types";

interface LoadedProduct {
	slug: string;
	product: ProductDetailsEntity | null;
}

function resolvePrice(product: ProductDetailsEntity): number {
	const variant = product.variants?.find((v) => v.isDefault) ?? product.variants?.[0];
	const price = variant?.extractedPrice ?? variant?.price ?? product.price;
	const num = Number(price);
	return Number.isFinite(num) ? num : 0;
}

function resolveAttributes(product: ProductDetailsEntity): Record<string, string> {
	const variant = product.variants?.find((v) => v.isDefault) ?? product.variants?.[0];
	const raw = (variant?.attributes ?? (product as unknown as { attributes?: Record<string, unknown> }).attributes) || {};
	const out: Record<string, string> = {};
	for (const [key, value] of Object.entries(raw)) {
		if (value === undefined || value === null || value === "") continue;
		const str = typeof value === "object" ? JSON.stringify(value) : String(value);
		if (str.trim() !== "" && str !== "null" && str !== "undefined") {
			out[key] = str;
		}
	}
	return out;
}

export default function CompareContent() {
	const t = useTranslations("compare");
	const tProducts = useTranslations("products");
	const locale = useLocale();
	const formatPrice = usePriceFormatter();
	const { compareSlugs, removeFromCompare, clearCompare } = useCompare();
	const [rows, setRows] = useState<LoadedProduct[]>([]);
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const loadProducts = useCallback(async (slugs: string[]) => {
		if (!slugs.length) {
			setRows([]);
			setLoading(false);
			setError(null);
			return;
		}
		setLoading(true);
		setError(null);
		try {
			const results = await Promise.all(
				slugs.map(async (slug) => {
					try {
						const res = await catalogService.getProductBySlug(slug);
						const raw = (res as { data?: unknown })?.data ?? (res as unknown);
						const product = normalizeProduct(
							raw as Parameters<typeof normalizeProduct>[0],
						) as unknown as ProductDetailsEntity | null;
						return { slug, product };
					} catch {
						return { slug, product: null };
					}
				}),
			);
			setRows(results);
		} catch {
			setError(t("loadError"));
		} finally {
			setLoading(false);
		}
	}, [t]);

	useEffect(() => {
		loadProducts(compareSlugs);
	}, [compareSlugs, loadProducts]);

	const products = useMemo(
		() =>
			rows
				.filter((row): row is { slug: string; product: ProductDetailsEntity } => Boolean(row.product)),
		[rows],
	);

	const attributeKeys = useMemo(() => {
		const keys = new Set<string>();
		for (const { product } of products) {
			for (const key of Object.keys(resolveAttributes(product))) {
				keys.add(key);
			}
		}
		return Array.from(keys).sort();
	}, [products]);

	const missing = rows.length - products.length;

	return (
		<div className="mx-auto flex w-full max-w-7xl flex-col gap-6 pb-12">
			<div className="flex flex-col gap-4 border-b border-border/70 pb-5 sm:flex-row sm:items-end sm:justify-between">
				<div>
					<h1 className="text-2xl font-bold tracking-tight text-foreground">{t("title")}</h1>
					<p className="mt-1 text-sm text-muted-foreground">{t("subtitle")}</p>
				</div>
				<div className="flex items-center gap-2.5">
					<Link
						href="/products"
						className="rounded-lg border border-border bg-secondary px-3.5 py-1.5 text-xs font-medium text-foreground shadow-2xs transition-colors hover:bg-secondary/80"
					>
						{t("continueShopping")}
					</Link>
					{products.length > 0 && (
						<button
							type="button"
							onClick={clearCompare}
							className="rounded-lg border border-destructive/20 px-3.5 py-1.5 text-xs font-medium text-destructive shadow-2xs transition-colors hover:bg-destructive/10"
						>
							{t("clearAll")}
						</button>
					)}
				</div>
			</div>

			{loading && products.length === 0 ? (
				<div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4 md:gap-4">
					{Array.from({ length: 4 }).map((_, i) => (
						<div
							key={i}
							className="flex animate-pulse flex-col gap-3 rounded-2xl border border-border bg-card p-4"
						>
							<div className="aspect-square rounded-xl bg-secondary/50" />
							<div className="h-4 w-3/4 rounded bg-secondary/60" />
							<div className="h-4 w-1/3 rounded bg-secondary/40" />
						</div>
					))}
				</div>
			) : products.length === 0 ? (
				<div className="flex w-full flex-col items-center justify-center gap-3 rounded-2xl border border-border bg-card p-10 text-center shadow-2xs sm:p-16">
					<div className="mb-1 flex size-16 items-center justify-center rounded-full border border-border/80 bg-secondary/60 text-muted-foreground">
						<Scale className="size-6 opacity-70" aria-hidden="true" />
					</div>
					<h2 className="text-lg font-bold text-foreground">{t("empty.title")}</h2>
					<p className="max-w-md text-sm leading-relaxed text-muted-foreground">
						{error ?? t("empty.subtitle")}
					</p>
					<Link
						href="/products"
						className="mt-4 rounded-lg bg-accent-brand px-4 py-2 text-xs font-semibold text-white shadow-xs transition-opacity hover:opacity-90"
					>
						{t("empty.browse")}
					</Link>
				</div>
			) : (
				<div className="overflow-x-auto rounded-2xl border border-border bg-card shadow-2xs">
					<table className="w-full min-w-[640px] border-collapse text-sm">
						<thead>
							<tr className="border-b border-border/70">
								<th className="sticky start-0 z-20 min-w-36 bg-card p-4 text-start align-bottom text-xs font-semibold uppercase tracking-wider text-muted-foreground">
									{t("table.spec")}
								</th>
								{products.map(({ slug, product }) => (
									<th key={slug} className="min-w-44 p-4 align-top">
										<div className="flex flex-col items-center gap-2 text-center">
											<Link
												href={`/products/${product.slug || slug}`}
												className="block w-full"
											>
												<div className="relative mx-auto aspect-square w-full max-w-36 overflow-hidden rounded-xl bg-secondary/40">
													<CatalogImage
														src={product.img || "/placeholder-product.png"}
														alt={getProductDisplayName(product, locale)}
														fill
														className="object-cover"
													/>
												</div>
											</Link>
											<Link
												href={`/products/${product.slug || slug}`}
												className="line-clamp-2 text-sm font-semibold text-foreground hover:text-accent-brand"
											>
												{getProductDisplayName(product, locale)}
											</Link>
											<button
												type="button"
												onClick={() => removeFromCompare(slug)}
												className="text-xs font-medium text-destructive hover:underline"
											>
												{t("remove")}
											</button>
										</div>
									</th>
								))}
							</tr>
						</thead>
						<tbody className="divide-y divide-border/70">
							<tr>
								<td className="sticky start-0 z-10 bg-card p-4 font-medium text-muted-foreground">{t("table.price")}</td>
								{products.map(({ slug, product }) => (
									<td key={slug} className="p-4 text-center font-bold text-foreground">
										{formatPrice(resolvePrice(product))}
									</td>
								))}
							</tr>
							<tr>
								<td className="sticky start-0 z-10 bg-card p-4 font-medium text-muted-foreground">{t("table.brand")}</td>
								{products.map(({ slug, product }) => (
									<td key={slug} className="p-4 text-center text-foreground">
										{typeof product.brand === "string"
											? product.brand || "—"
											: (product.brand?.name ?? "—")}
									</td>
								))}
							</tr>
							<tr>
								<td className="sticky start-0 z-10 bg-card p-4 font-medium text-muted-foreground">{t("table.category")}</td>
								{products.map(({ slug, product }) => (
									<td key={slug} className="p-4 text-center capitalize text-foreground">
										{product.category?.replace("-", " ") || "—"}
									</td>
								))}
							</tr>
							<tr>
								<td className="sticky start-0 z-10 bg-card p-4 font-medium text-muted-foreground">{t("table.availability")}</td>
								{products.map(({ slug, product }) => (
									<td
										key={slug}
										className={`p-4 text-center font-medium ${
											product.inStock ? "text-emerald-700 dark:text-emerald-400" : "text-destructive"
										}`}
									>
										{product.inStock ? tProducts("actions.inStock", { count: product.stockCount ?? 0 }) : tProducts("actions.outOfStock")}
									</td>
								))}
							</tr>
							<tr>
								<td className="sticky start-0 z-10 bg-card p-4 font-medium text-muted-foreground">{t("table.rating")}</td>
								{products.map(({ slug, product }) => (
									<td key={slug} className="p-4">
										<div className="flex items-center justify-center gap-1.5">
											<Stars starsNum={Number(product.stars) || 0} />
											<span className="text-xs text-muted-foreground">({product.reviews ?? 0})</span>
										</div>
									</td>
								))}
							</tr>
							{attributeKeys.map((key) => (
								<tr key={key}>
									<td className="sticky start-0 z-10 bg-card p-4 font-medium capitalize text-muted-foreground">
										{key.replace(/[_-]+/g, " ")}
									</td>
									{products.map(({ slug, product }) => (
										<td key={slug} className="p-4 text-center text-foreground">
											{resolveAttributes(product)[key] ?? "—"}
										</td>
									))}
								</tr>
							))}
						</tbody>
					</table>
				</div>
			)}

			{missing > 0 && !loading && (
				<p className="text-sm text-muted-foreground">{t("missingProducts", { count: missing })}</p>
			)}
			{products.length === 1 && !loading && (
				<p className="text-sm text-muted-foreground">{t("addMore")}</p>
			)}
		</div>
	);
}
