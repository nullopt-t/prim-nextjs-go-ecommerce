"use client";

import { useEffect, useState } from "react";
import { Loader2, Plus, Trash2 } from "lucide-react";
import { sellerService } from "@/modules/seller/api/sellerService";
import type { SellerBundle, SellerProduct, SellerVariant } from "@/modules/seller/types";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/form/select";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";

interface BundleItemDraft {
	productId: string;
	variantId: string;
	quantity: string;
}

export function SellerBundles() {
	const t = useTranslations("seller");
	const [bundles, setBundles] = useState<SellerBundle[]>([]);
	const [products, setProducts] = useState<SellerProduct[]>([]);
	const [variantsByProduct, setVariantsByProduct] = useState<Record<string, SellerVariant[]>>({});
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState<Error | null>(null);

	const [title, setTitle] = useState("");
	const [price, setPrice] = useState("");
	const [currency, setCurrency] = useState("EGP");
	const [thumbnailUrl, setThumbnailUrl] = useState("");
	const [uploading, setUploading] = useState(false);
	const [items, setItems] = useState<BundleItemDraft[]>([{ productId: "", variantId: "", quantity: "1" }]);
	const [creating, setCreating] = useState(false);
	const [editingId, setEditingId] = useState<string | null>(null);

	const load = async () => {
		setLoading(true);
		setLoadError(null);
		try {
			const [bundleResponse, productResponse] = await Promise.all([
				sellerService.getBundles(),
				sellerService.getProducts({ page: 1, pageSize: 100 }),
			]);
			setBundles(bundleResponse.data ?? []);
			setProducts(productResponse.data ?? []);
		} catch (error) {
			setLoadError(error instanceof Error ? error : new Error(String(error)));
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		void load();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, []);

	const loadVariants = async (productId: string) => {
		if (!productId || variantsByProduct[productId]) return;
		try {
			const response = await sellerService.getVariants(productId, { pageSize: 100 });
			setVariantsByProduct((current) => ({ ...current, [productId]: response.data ?? [] }));
		} catch {
			// Variant list is best-effort; the product itself still bundles.
		}
	};

	const uploadThumbnail = async (file: File | undefined) => {
		if (!file) return;
		setUploading(true);
		try {
			const uploaded = await sellerService.uploadMedia(file);
			setThumbnailUrl(uploaded.data.url);
		} catch (error) {
			notify.error(apiErrorMessage(error, t("products.saveFailed")));
		} finally {
			setUploading(false);
		}
	};

	const startEdit = (bundle: SellerBundle) => {
		setEditingId(bundle.id);
		setTitle(bundle.productTitle);
		setPrice(String(bundle.bundlePrice / 100));
		setCurrency(bundle.currency);
		setThumbnailUrl("");
		setItems(
			bundle.items.map((item) => ({
				productId: item.productId,
				variantId: item.variantId ?? "",
				quantity: String(item.quantity),
			})),
		);
		for (const item of bundle.items) void loadVariants(item.productId);
	};

	const cancelEdit = () => {
		setEditingId(null);
		setTitle("");
		setPrice("");
		setThumbnailUrl("");
		setItems([{ productId: "", variantId: "", quantity: "1" }]);
	};

	const createBundle = async () => {
		const priceMinor = Math.round(Number(price) * 100);
		if (!title.trim() || !Number.isFinite(priceMinor) || priceMinor <= 0) return;
		const payloadItems = items
			.filter((item) => item.productId && Number(item.quantity) > 0)
			.map((item) => ({
				productId: item.productId,
				variantId: item.variantId || undefined,
				quantity: Number(item.quantity),
			}));
		if (payloadItems.length === 0) return;
		setCreating(true);
		try {
			if (editingId) {
				await sellerService.updateBundle(editingId, {
					title: title.trim(),
					bundlePrice: priceMinor,
					currency: currency.trim().toUpperCase() || "EGP",
					items: payloadItems,
				});
				notify.success(t("bundles.updated"));
			} else {
				await sellerService.createBundle({
					title: title.trim(),
					bundlePrice: priceMinor,
					currency: currency.trim().toUpperCase() || "EGP",
					thumbnailUrl: thumbnailUrl || undefined,
					items: payloadItems,
				});
				notify.success(t("bundles.created"));
			}
			cancelEdit();
			await load();
		} catch (error) {
			notify.error(apiErrorMessage(error, t("products.saveFailed")));
		} finally {
			setCreating(false);
		}
	};

	const publishBundle = async (bundle: SellerBundle) => {
		try {
			await sellerService.publishProduct(bundle.productId);
			notify.success(t("bundles.published"));
			await load();
		} catch (error) {
			notify.error(apiErrorMessage(error, t("products.saveFailed")));
		}
	};

	const removeBundle = async (bundle: SellerBundle) => {
		try {
			await sellerService.deleteBundle(bundle.id);
			notify.success(t("bundles.deleted"));
			await load();
		} catch (error) {
			notify.error(apiErrorMessage(error, t("products.saveFailed")));
		}
	};

	if (loading) {
		return (
			<div className="flex min-h-[50vh] items-center justify-center">
				<div className="size-6 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
			</div>
		);
	}

	if (loadError) {
		return (
			<div className="space-y-5">
				<PageHeader title={t("bundles.title")} description={t("bundles.subtitle")} />
				<ErrorState title={t("bundles.title")} description={loadError.message || t("bundles.loadFailed")} onRetry={() => void load()} />
			</div>
		);
	}

	return (
		<div className="space-y-6">
			<PageHeader title={t("bundles.title")} description={t("bundles.subtitle")} />

			<section className="space-y-4 rounded-lg border border-border bg-background p-5">
				<h2 className="text-base font-semibold text-foreground">{t("bundles.createTitle")}</h2>
				<div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
					<div className="space-y-1">
						<label className="text-xs font-medium text-foreground">{t("bundles.bundleTitle")}</label>
						<input
							value={title}
							onChange={(e) => setTitle(e.target.value)}
							className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
						/>
					</div>
					<div className="space-y-1">
						<label className="text-xs font-medium text-foreground">{t("bundles.bundlePrice")}</label>
						<input
							type="number"
							min="0.01"
							step="0.01"
							value={price}
							onChange={(e) => setPrice(e.target.value)}
							className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
						/>
					</div>
					<div className="space-y-1">
						<label className="text-xs font-medium text-foreground">{t("bundles.currency")}</label>
						<input
							value={currency}
							onChange={(e) => setCurrency(e.target.value.toUpperCase())}
							className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
						/>
					</div>
					<div className="space-y-1">
						<label className="text-xs font-medium text-foreground">{t("bundles.thumbnail")}</label>
						<div className="flex items-center gap-2">
							{thumbnailUrl ? (
								<img src={thumbnailUrl} alt="" className="size-9 rounded-md border border-border object-cover" />
							) : (
								<div className="size-9 rounded-md border border-dashed border-border" />
							)}
							<label className="inline-flex cursor-pointer items-center rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground hover:bg-secondary">
								{uploading && <Loader2 className="size-3.5 animate-spin" />}
								{t("wizard.variants.chooseThumbnail")}
								<input
									type="file"
									accept="image/*"
									disabled={uploading}
									className="sr-only"
									onChange={(e) => {
										void uploadThumbnail(e.target.files?.[0]);
										e.target.value = "";
									}}
								/>
							</label>
						</div>
					</div>
				</div>

				<div className="space-y-2">
					<div className="flex items-center justify-between">
						<p className="text-xs font-semibold text-foreground">{t("bundles.itemsTitle")}</p>
						<button
							type="button"
							onClick={() => setItems((current) => [...current, { productId: "", variantId: "", quantity: "1" }])}
							className="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground hover:bg-secondary"
						>
							<Plus className="size-3.5" />
							{t("bundles.addItem")}
						</button>
					</div>
					{items.map((item, index) => {
						const variants = variantsByProduct[item.productId] ?? [];
						return (
							<div key={index} className="grid grid-cols-1 items-end gap-2 rounded-lg border border-border p-3 sm:grid-cols-[1fr_1fr_24_32px]">
								<div className="space-y-1">
									<label className="text-xs text-muted-foreground">{t("bundles.product")}</label>
									<Select
										value={item.productId}
										onValueChange={(value) => {
											void loadVariants(value);
											setItems((current) =>
												current.map((entry, i) => (i === index ? { ...entry, productId: value, variantId: "" } : entry)),
											);
										}}
									>
										<SelectTrigger>
											<SelectValue placeholder={t("bundles.product")} />
										</SelectTrigger>
										<SelectContent>
											{products.map((product) => (
												<SelectItem key={product.id} value={product.id}>{product.title}</SelectItem>
											))}
										</SelectContent>
									</Select>
								</div>
								<div className="space-y-1">
									<label className="text-xs text-muted-foreground">{t("bundles.variant")}</label>
									<Select
										value={item.variantId}
										onValueChange={(value) =>
											setItems((current) =>
												current.map((entry, i) => (i === index ? { ...entry, variantId: value } : entry)),
											)
										}
									>
										<SelectTrigger>
											<SelectValue placeholder={t("bundles.wholeProduct")} />
										</SelectTrigger>
										<SelectContent>
											{variants.map((variant) => (
												<SelectItem key={variant.id} value={variant.id}>{variant.title}</SelectItem>
											))}
										</SelectContent>
									</Select>
								</div>
								<div className="space-y-1">
									<label className="text-xs text-muted-foreground">{t("bundles.quantity")}</label>
									<input
										type="number"
										min="1"
										value={item.quantity}
										onChange={(e) =>
											setItems((current) =>
												current.map((entry, i) => (i === index ? { ...entry, quantity: e.target.value } : entry)),
											)
										}
										className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
									/>
								</div>
								<button
									type="button"
									disabled={items.length <= 1}
									onClick={() => setItems((current) => current.filter((_, i) => i !== index))}
									className="rounded-md border border-border p-2 text-muted-foreground hover:bg-destructive/10 hover:text-destructive disabled:opacity-50"
									aria-label={t("bundles.deleted")}
								>
									<Trash2 className="size-4" />
								</button>
							</div>
						);
					})}
				</div>

				<div className="flex flex-wrap gap-2">
					<button
						type="button"
						disabled={creating || !title.trim() || !price || items.some((item) => !item.productId)}
						onClick={() => void createBundle()}
						className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground disabled:opacity-50"
					>
						{editingId ? t("bundles.save") : t("bundles.create")}
					</button>
					{editingId && (
						<button
							type="button"
							onClick={cancelEdit}
							className="rounded-md border border-border px-4 py-2 text-sm font-medium text-foreground hover:bg-secondary"
						>
							{t("actions.cancel")}
						</button>
					)}
				</div>
			</section>

			{bundles.length === 0 ? (
				<p className="text-xs text-muted-foreground">{t("bundles.empty")}</p>
			) : (
				<ul className="divide-y divide-border rounded-lg border border-border bg-background">
					{bundles.map((bundle) => (
						<li key={bundle.id} className="flex flex-wrap items-center gap-3 px-4 py-3">
							<div className="min-w-48 flex-1">
								<p className="text-sm font-semibold text-foreground">{bundle.productTitle}</p>
								<p className="text-xs text-muted-foreground">
									{bundle.items.map((item) => `${item.quantity}× ${item.title || item.productId}`).join(" · ")}
								</p>
							</div>
							<span className="text-sm font-semibold tabular-nums text-foreground">
								{bundle.currency} {(bundle.bundlePrice / 100).toFixed(2)}
							</span>
							<StatusPill tone={statusTone(bundle.status)}>{bundle.status}</StatusPill>
							{bundle.status !== "published" && (
								<button
									type="button"
									onClick={() => void publishBundle(bundle)}
									className="rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground hover:bg-secondary"
								>
									{t("bundles.publish")}
								</button>
							)}
							<button
								type="button"
								onClick={() => startEdit(bundle)}
								className="rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground hover:bg-secondary"
							>
								{t("bundles.edit")}
							</button>
							<button
								type="button"
								onClick={() => void removeBundle(bundle)}
								className="rounded-md border border-border p-1.5 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
								aria-label={t("bundles.deleted")}
							>
								<Trash2 className="size-4" />
							</button>
						</li>
					))}
				</ul>
			)}
		</div>
	);
}
