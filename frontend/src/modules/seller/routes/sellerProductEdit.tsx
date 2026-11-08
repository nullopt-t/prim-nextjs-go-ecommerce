"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { Link, useRouter } from "@/shared/i18n/navigation";
import { useTranslations } from "next-intl";
import { sellerService } from "@/modules/seller/api/sellerService";
import type { SellerProduct, SellerVariant } from "@/modules/seller/types";
import { parseMajorPrice } from "./wizard/wizardTypes";
import { PageHeader } from "@/shared/ui/pageHeader";
import { ErrorState } from "@/shared/ui/errorState";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { apiErrorMessage, notify } from "@/shared/lib/notify";

/**
 * Dedicated edit surface for an existing product. Products are immutable
 * after creation except for the (default variant's) price and lifecycle
 * status, so the wizard stays creation-only and this page owns price-only
 * edits. Everything else is display-only; deletion stays on the list.
 */
export function SellerProductEdit() {
	const t = useTranslations("seller");
	const router = useRouter();
	const params = useParams();
	const productId = typeof params?.id === "string" ? params.id : undefined;

	const [product, setProduct] = useState<SellerProduct | null>(null);
	const [variant, setVariant] = useState<SellerVariant | null>(null);
	const [newPrice, setNewPrice] = useState("");
	const [priceError, setPriceError] = useState<string | null>(null);
	const [loading, setLoading] = useState(true);
	const [saving, setSaving] = useState(false);
	const [loadError, setLoadError] = useState<Error | null>(null);

	useEffect(() => {
		let cancelled = false;
		const load = async () => {
			if (!productId) {
				setLoadError(new Error(t("products.loadTitle")));
				setLoading(false);
				return;
			}
			// No seller get-by-id endpoint; the list API is the summary source.
			let found: SellerProduct | undefined;
			for (let page = 1; page <= 10 && !found; page += 1) {
				const response = await sellerService.getProducts({ page, pageSize: 100 });
				found = response.data.find((item) => item.id === productId);
				if (response.data.length < 100) break;
			}
			if (!found) throw new Error(t("wizard.loadFailed"));
			const variantsResponse = await sellerService.getVariants(found.id, { pageSize: 100 });
			const variants = variantsResponse.data ?? [];
			if (cancelled) return;
			setProduct(found);
			setVariant(variants.find((item) => item.isDefault) ?? variants[0] ?? null);
		};
		load()
			.catch((error) => {
				if (!cancelled) setLoadError(error instanceof Error ? error : new Error(String(error)));
			})
			.finally(() => {
				if (!cancelled) setLoading(false);
			});
		return () => {
			cancelled = true;
		};
	}, [productId, t]);

	if (loading) {
		return (
			<div className="flex min-h-[50vh] items-center justify-center">
				<div className="size-6 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
			</div>
		);
	}

	if (loadError || !product) {
		return (
			<div className="space-y-5">
				<PageHeader title={t("products.editTitle")} />
				<ErrorState
					title={t("products.loadTitle")}
					description={loadError?.message || t("wizard.loadFailed")}
					onRetry={() => router.refresh()}
				/>
			</div>
		);
	}

	const currency = product.currency || variant?.currency || "EGP";
	const currentPriceMinor = variant?.price;
	const currentPriceDisplay =
		currentPriceMinor !== undefined && currentPriceMinor !== null
			? `${(currentPriceMinor / 100).toFixed(2)} ${currency}`
			: "—";
	const isPublished = product.status === "published";

	const handleSavePrice = async () => {
		if (!variant) return;
		const minor = parseMajorPrice(newPrice);
		if (minor === undefined) {
			setPriceError(t("products.validation.priceInvalid"));
			return;
		}
		setPriceError(null);
		setSaving(true);
		try {
			// Only the price field is sent; every other product field stays nil
			// (unchanged) server-side.
			await sellerService.updateVariant(variant.id, { price: minor });
			notify.success(t("toasts.productUpdated"));
			router.replace("/seller/products");
		} catch (error) {
			notify.error(apiErrorMessage(error, t("products.saveFailed")));
		} finally {
			setSaving(false);
		}
	};

	const handleLifecycle = async () => {
		setSaving(true);
		try {
			if (isPublished) {
				await sellerService.archiveProduct(product.id);
				notify.success(t("toasts.productArchived"));
			} else {
				await sellerService.publishProduct(product.id);
				notify.success(t("toasts.productPublished"));
			}
			router.replace("/seller/products");
		} catch (error) {
			notify.error(apiErrorMessage(error, t("products.saveFailed")));
		} finally {
			setSaving(false);
		}
	};

	return (
		<div className="space-y-6">
			<PageHeader title={t("products.editTitle")} />
			<Link href="/seller/products" className="text-xs font-medium text-muted-foreground hover:text-foreground">
				← {t("products.backToProducts")}
			</Link>

			<section className="space-y-4 rounded-xl border border-border bg-background p-5">
				<div className="flex flex-wrap items-center justify-between gap-3">
					<div className="min-w-0">
						<h2 className="truncate text-sm font-semibold text-foreground">{product.title}</h2>
						<p className="font-mono text-xs text-muted-foreground">{product.sku || product.slug}</p>
					</div>
					<div className="flex items-center gap-2">
						<StatusPill tone={statusTone(product.status)}>{product.status}</StatusPill>
						<button
							type="button"
							onClick={handleLifecycle}
							disabled={saving}
							className="rounded-md border border-border px-3 py-1.5 text-xs font-medium text-foreground hover:bg-secondary disabled:opacity-50"
						>
							{isPublished ? t("actions.archive") : t("actions.publish")}
						</button>
					</div>
				</div>

				{variant ? (
					<div className="space-y-3 border-t border-border pt-4">
						<p className="text-xs text-muted-foreground">
							{t("products.editPrice.defaultVariantNote", { variant: variant.title || variant.id })}
						</p>
						<div className="grid gap-3 sm:grid-cols-2">
							<div>
								<label className="block text-xs font-medium text-muted-foreground">
									{t("products.editPrice.current")}
								</label>
								<p className="mt-1 text-lg font-semibold tabular-nums text-foreground">{currentPriceDisplay}</p>
							</div>
							<div>
								<label htmlFor="edit-price" className="block text-xs font-medium text-muted-foreground">
									{t("products.editPrice.newPrice")}
								</label>
								<input
									id="edit-price"
									type="text"
									inputMode="decimal"
									value={newPrice}
									onChange={(event) => {
										setNewPrice(event.target.value);
										setPriceError(null);
									}}
									placeholder={(currentPriceMinor !== undefined && currentPriceMinor !== null ? (currentPriceMinor / 100).toFixed(2) : "0.00")}
									className="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-accent-brand/30"
									aria-invalid={priceError !== null}
								/>
								{priceError ? (
									<p className="mt-1 text-xs text-destructive" role="alert">
										{priceError}
									</p>
								) : null}
							</div>
						</div>
					</div>
				) : (
					<p className="border-t border-border pt-4 text-xs text-muted-foreground">
						{t("wizard.loadFailed")}
					</p>
				)}

				<div className="flex justify-end gap-2 border-t border-border pt-4">
					<Link
						href="/seller/products"
						className="rounded-md border border-border px-4 py-2 text-sm font-medium text-foreground hover:bg-secondary"
					>
						{t("actions.cancel")}
					</Link>
					<button
						type="button"
						onClick={handleSavePrice}
						disabled={saving || !variant}
						className="rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:opacity-90 disabled:opacity-50"
					>
						{t("actions.saveChanges")}
					</button>
				</div>
			</section>
		</div>
	);
}
