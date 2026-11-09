"use client";

import { useEffect, useState } from "react";
import { Trash2 } from "lucide-react";
import { sellerService } from "@/modules/seller/api/sellerService";
import type { SellerPromotion, SellerVariant } from "@/modules/seller/types";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/form/select";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";

interface SaleRow {
	variant: SellerVariant;
	productTitle: string;
	compareAt: string;
	saving: boolean;
}

export function SellerPromotions() {
	const t = useTranslations("seller");
	const [promotions, setPromotions] = useState<SellerPromotion[]>([]);
	const [sales, setSales] = useState<SaleRow[]>([]);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState<Error | null>(null);

	const [code, setCode] = useState("");
	const [discountType, setDiscountType] = useState<"percentage" | "fixed_amount">("percentage");
	const [discountValue, setDiscountValue] = useState("");
	const [minCartAmount, setMinCartAmount] = useState("");
	const [usageLimit, setUsageLimit] = useState("");
	const [creating, setCreating] = useState(false);

	const load = async () => {
		setLoading(true);
		setLoadError(null);
		try {
			const promoResponse = await sellerService.getPromotions();
			setPromotions(promoResponse.data ?? []);

			const productResponse = await sellerService.getProducts({ page: 1, pageSize: 100 });
			const rows: SaleRow[] = [];
			for (const product of productResponse.data) {
				const variants = await sellerService.getVariants(product.id, { pageSize: 100 });
				for (const variant of variants.data) {
					rows.push({
						variant,
						productTitle: product.title,
						compareAt: variant.crossedOutPrice ? String(variant.crossedOutPrice / 100) : "",
						saving: false,
					});
				}
			}
			setSales(rows);
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

	const createCoupon = async () => {
		const value = Number(discountValue);
		if (!code.trim() || !Number.isFinite(value) || value <= 0) return;
		setCreating(true);
		try {
			await sellerService.createPromotion({
				code: code.trim().toUpperCase(),
				discountType,
				discountValue: discountType === "percentage" ? value : Math.round(value * 100),
				minCartAmount: minCartAmount ? Math.round(Number(minCartAmount) * 100) : undefined,
				usageLimit: usageLimit ? Number.parseInt(usageLimit, 10) : undefined,
			});
			notify.success(t("promotions.created"));
			setCode("");
			setDiscountValue("");
			setMinCartAmount("");
			setUsageLimit("");
			await load();
		} catch (error) {
			notify.error(apiErrorMessage(error, t("products.saveFailed")));
		} finally {
			setCreating(false);
		}
	};

	const toggleStatus = async (promotion: SellerPromotion) => {
		const nextStatus = promotion.status === "active" ? "disabled" : "active";
		try {
			await sellerService.setPromotionStatus(promotion.id, nextStatus);
			notify.success(t("promotions.statusUpdated"));
			await load();
		} catch (error) {
			notify.error(apiErrorMessage(error, t("products.saveFailed")));
		}
	};

	const removeCoupon = async (promotion: SellerPromotion) => {
		try {
			await sellerService.deletePromotion(promotion.id);
			notify.success(t("promotions.deleted"));
			await load();
		} catch (error) {
			notify.error(apiErrorMessage(error, t("products.saveFailed")));
		}
	};

	const saveSalePrice = async (row: SaleRow) => {
		const compareAt = row.compareAt.trim() ? Math.round(Number(row.compareAt) * 100) : undefined;
		setSales((current) =>
			current.map((item) => (item.variant.id === row.variant.id ? { ...item, saving: true } : item)),
		);
		try {
			await sellerService.updateVariant(row.variant.id, { crossedOutPrice: compareAt });
			notify.success(t("promotions.saved"));
			await load();
		} catch (error) {
			notify.error(apiErrorMessage(error, t("products.saveFailed")));
		} finally {
			setSales((current) =>
				current.map((item) => (item.variant.id === row.variant.id ? { ...item, saving: false } : item)),
			);
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
				<PageHeader title={t("promotions.title")} description={t("promotions.subtitle")} />
				<ErrorState title={t("promotions.title")} description={loadError.message || t("promotions.loadFailed")} onRetry={() => void load()} />
			</div>
		);
	}

	return (
		<div className="space-y-6">
			<PageHeader title={t("promotions.title")} description={t("promotions.subtitle")} />

			<section className="space-y-4 rounded-lg border border-border bg-background p-5">
				<div>
					<h2 className="text-base font-semibold text-foreground">{t("promotions.couponsTitle")}</h2>
					<p className="mt-1 text-xs text-muted-foreground">{t("promotions.couponsSubtitle")}</p>
				</div>
				<div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-5">
					<div className="space-y-1">
						<label className="text-xs font-medium text-foreground">{t("promotions.code")}</label>
						<input
							value={code}
							onChange={(e) => setCode(e.target.value.toUpperCase())}
							dir="ltr"
							placeholder="SUMMER20"
							className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
						/>
					</div>
					<div className="space-y-1">
						<label className="text-xs font-medium text-foreground">{t("promotions.discountType")}</label>
						<Select value={discountType} onValueChange={(value) => setDiscountType(value as "percentage" | "fixed_amount")}>
							<SelectTrigger>
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="percentage">{t("promotions.percentage")}</SelectItem>
								<SelectItem value="fixed_amount">{t("promotions.fixedAmount")}</SelectItem>
							</SelectContent>
						</Select>
					</div>
					<div className="space-y-1">
						<label className="text-xs font-medium text-foreground">{t("promotions.discountValue")}</label>
						<input
							type="number"
							min="0"
							step={discountType === "percentage" ? "1" : "0.01"}
							value={discountValue}
							onChange={(e) => setDiscountValue(e.target.value)}
							className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
						/>
					</div>
					<div className="space-y-1">
						<label className="text-xs font-medium text-foreground">{t("promotions.minCartAmount")}</label>
						<input
							type="number"
							min="0"
							step="0.01"
							value={minCartAmount}
							onChange={(e) => setMinCartAmount(e.target.value)}
							className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
						/>
					</div>
					<div className="space-y-1">
						<label className="text-xs font-medium text-foreground">{t("promotions.usageLimit")}</label>
						<input
							type="number"
							min="0"
							value={usageLimit}
							onChange={(e) => setUsageLimit(e.target.value)}
							className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
						/>
					</div>
				</div>
				<button
					type="button"
					disabled={creating || !code.trim() || !discountValue}
					onClick={() => void createCoupon()}
					className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground disabled:opacity-50"
				>
					{t("promotions.create")}
				</button>

				{promotions.length === 0 ? (
					<p className="text-xs text-muted-foreground">{t("promotions.emptyCoupons")}</p>
				) : (
					<ul className="divide-y divide-border rounded-md border border-border">
						{promotions.map((promotion) => (
							<li key={promotion.id} className="flex flex-wrap items-center gap-3 px-4 py-3">
								<div className="min-w-40 flex-1">
									<p className="text-sm font-semibold text-foreground" dir="ltr">{promotion.code}</p>
									<p className="text-xs text-muted-foreground">
										{promotion.discountType === "percentage"
											? `${promotion.discountValue}%`
											: `${(promotion.discountValue / 100).toFixed(2)} ${promotion.currency}`}
										{" · "}
										{t("promotions.usageCount", { count: promotion.usageCount })}
									</p>
								</div>
								<StatusPill tone={statusTone(promotion.status)}>
									{promotion.status === "active" ? t("promotions.active") : t("promotions.disabled")}
								</StatusPill>
								<button
									type="button"
									onClick={() => void toggleStatus(promotion)}
									className="rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground hover:bg-secondary"
								>
									{promotion.status === "active" ? t("promotions.disable") : t("promotions.enable")}
								</button>
								<button
									type="button"
									onClick={() => void removeCoupon(promotion)}
									className="rounded-md border border-border p-1.5 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
									aria-label={t("promotions.delete")}
								>
									<Trash2 className="size-4" />
								</button>
							</li>
						))}
					</ul>
				)}
			</section>

			<section className="space-y-4 rounded-lg border border-border bg-background p-5">
				<div>
					<h2 className="text-base font-semibold text-foreground">{t("promotions.salesTitle")}</h2>
					<p className="mt-1 text-xs text-muted-foreground">{t("promotions.salesSubtitle")}</p>
				</div>
				{sales.length === 0 ? (
					<p className="text-xs text-muted-foreground">{t("promotions.emptySales")}</p>
				) : (
					<ul className="divide-y divide-border rounded-md border border-border">
						{sales.map((row) => (
							<li key={row.variant.id} className="flex flex-wrap items-center gap-3 px-4 py-3">
								<div className="min-w-40 flex-1">
									<p className="text-sm font-medium text-foreground">{row.productTitle}</p>
									<p className="text-xs text-muted-foreground">
										{row.variant.title}
										{row.variant.price ? ` · ${(row.variant.price / 100).toFixed(2)}` : ""}
									</p>
								</div>
								<div className="flex flex-wrap items-center gap-2">
									<label className="text-xs text-muted-foreground">{t("promotions.compareAtPrice")}</label>
									<input
										type="number"
										min="0"
										step="0.01"
										value={row.compareAt}
										onChange={(e) =>
											setSales((current) =>
												current.map((item) =>
													item.variant.id === row.variant.id ? { ...item, compareAt: e.target.value } : item,
												),
											)
										}
										className="w-28 rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
									/>
									<button
										type="button"
										disabled={row.saving}
										onClick={() => void saveSalePrice(row)}
										className="rounded-md bg-primary px-3 py-2 text-xs font-medium text-primary-foreground disabled:opacity-50"
									>
										{t("promotions.save")}
									</button>
								</div>
							</li>
						))}
					</ul>
				)}
			</section>
		</div>
	);
}
