"use client";

import { useEffect, useState } from "react";
import { Search } from "lucide-react";
import { vendorService } from "@/modules/vendor/api/vendorService";
import type {
	StockReason,
	VendorProduct,
	VendorVariant,
} from "@/modules/vendor/types";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";

interface InventoryRow {
	product: VendorProduct;
	variant: VendorVariant;
}

export function VendorInventory() {
	const t = useTranslations("vendor");
	const [rows, setRows] = useState<InventoryRow[]>([]);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState<Error | null>(null);
	const [search, setSearch] = useState("");
	const [filterLowStock, setFilterLowStock] = useState(false);
	const [draftAdjustments, setDraftAdjustments] = useState<Record<string, string>>({});
	const [reasonByVariant, setReasonByVariant] = useState<Record<string, StockReason>>({});
	const [savingVariantId, setSavingVariantId] = useState<string | null>(null);

	const loadInventory = async () => {
		setLoading(true);
		setLoadError(null);
		try {
			const productResponse = await vendorService.getProducts({ pageSize: 100 });
			const productVariantLists = await Promise.all(
				productResponse.data.map(async (product) => ({
					product,
					variants: (await vendorService.getVariants(product.id, { pageSize: 100 })).data,
				})),
			);

			setRows(
				productVariantLists.flatMap(({ product, variants }) =>
					variants.map((variant) => ({ product, variant })),
				),
			);
		} catch (error) {
			setLoadError(error instanceof Error ? error : new Error(String(error)));
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		void loadInventory();
	}, []);

	const adjustStock = async (variant: VendorVariant, delta: number) => {
		if (delta === 0) return;
		setSavingVariantId(variant.id);
		try {
			await vendorService.adjustStock(variant.id, {
				quantity: delta,
				reason: reasonByVariant[variant.id] ?? "restock",
			});
			notify.success(t("toasts.inventoryUpdated"));
			await loadInventory();
		} catch (error) {
			notify.error(apiErrorMessage(error, t("variants.stockFailed")));
		} finally {
			setSavingVariantId(null);
		}
	};

	const handleDraftSubmit = (variant: VendorVariant) => {
		const parsed = Number.parseInt(draftAdjustments[variant.id] ?? "", 10);
		if (!Number.isFinite(parsed) || parsed === 0) {
			notify.error(t("variants.invalidStockAmount"));
			return;
		}
		void adjustStock(variant, parsed);
	};

	const filtered = rows.filter(({ product, variant }) => {
		const matchesSearch =
			product.title.toLowerCase().includes(search.toLowerCase()) ||
			variant.title.toLowerCase().includes(search.toLowerCase()) ||
			variant.sku.toLowerCase().includes(search.toLowerCase());
		const matchesLowStock = filterLowStock ? product.stock <= 5 : true;
		return matchesSearch && matchesLowStock;
	});

	const lowStockCount = rows.filter(({ product }) => product.stock <= 5).length;

	return (
		<div className="space-y-5">
			<PageHeader title={t("inventory.title")} />

			{lowStockCount > 0 ? (
				<p className="rounded-md border border-amber-600/20 bg-amber-600/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-400">
					{t("inventory.lowStockWarning", { count: lowStockCount })}
				</p>
			) : null}

			{loading ? (
				<div className="flex min-h-[40vh] items-center justify-center">
					<div className="size-6 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
				</div>
			) : loadError ? (
				<ErrorState
					title={t("inventory.loadTitle")}
					description={loadError.message || t("inventory.loadDescription")}
					onRetry={() => void loadInventory()}
				/>
			) : (
				<>
					<div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
						<div className="relative w-full sm:max-w-xs">
							<Search className="absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
							<input
								type="text"
								placeholder={t("inventory.searchPlaceholder")}
								value={search}
								onChange={(event) => setSearch(event.target.value)}
								className="w-full rounded-md border border-border bg-background py-2 pe-3 ps-9 text-sm text-foreground placeholder:text-muted-foreground focus:outline-hidden focus:ring-2 focus:ring-ring/30"
							/>
						</div>
						<button
							onClick={() => setFilterLowStock(!filterLowStock)}
							className={`rounded-md border px-3 py-2 text-sm font-medium transition-colors ${
								filterLowStock
									? "border-amber-600/30 bg-amber-600/10 text-amber-700 dark:text-amber-400"
									: "border-border bg-background text-foreground hover:bg-secondary"
							}`}
						>
							{filterLowStock ? t("actions.showAllItems") : t("actions.filterLowStock")}
						</button>
					</div>

					<section className="overflow-hidden rounded-lg border border-border bg-background">
						<p className="border-b border-border px-4 py-2.5 text-xs text-muted-foreground">
							{t("inventory.variantLevelNote")}
						</p>
						<div className="overflow-x-auto">
							<ResponsiveTable className="w-full text-start text-sm md:min-w-max">
								<thead className="text-xs text-muted-foreground">
									<tr className="border-b border-border">
										<th className="py-2.5 ps-4 text-start font-medium">
											{t("inventory.table.productVariant")}
										</th>
										<th className="py-2.5 text-start font-medium">
											{t("inventory.table.sku")}
										</th>
										<th className="py-2.5 text-start font-medium">
											{t("inventory.table.skuStock")}
										</th>
										<th className="py-2.5 text-start font-medium">
											{t("inventory.table.stockReason")}
										</th>
										<th className="py-2.5 text-start font-medium">
											{t("inventory.table.status")}
										</th>
										<th className="py-2.5 pe-4 text-end font-medium">
											{t("inventory.table.adjustment")}
										</th>
									</tr>
								</thead>
								<tbody className="divide-y divide-border">
									{filtered.map(({ product, variant }) => (
										<tr key={variant.id} className="transition-colors hover:bg-secondary/30">
											<td className="py-3 ps-4">
												<div className="font-medium text-foreground">{product.title}</div>
												<div className="text-xs text-muted-foreground">{variant.title}</div>
											</td>
											<td className="py-3 font-mono text-xs text-muted-foreground">
												{variant.sku}
											</td>
											<td className="py-3 font-medium text-foreground tabular-nums">
												{product.stock}
											</td>
											<td className="py-3">
												<select
													value={reasonByVariant[variant.id] ?? "restock"}
													onChange={(event) =>
														setReasonByVariant((current) => ({
															...current,
															[variant.id]: event.target.value as StockReason,
														}))
													}
													className="rounded-md border border-border bg-background px-2 py-1.5 text-xs text-foreground"
													aria-label={t("inventory.table.stockReason")}
												>
													<option value="restock">{t("inventory.reasonRestock")}</option>
													<option value="return">{t("inventory.reasonReturn")}</option>
													<option value="adjustment">{t("inventory.reasonAdjustment")}</option>
												</select>
											</td>
											<td className="py-3">
												<StatusPill
													tone={
														product.stock === 0
															? "danger"
															: product.stock <= 5
																? "warning"
																: statusTone("in stock")
													}
												>
													{product.stock === 0
														? t("status.outOfStock")
														: product.stock <= 5
															? t("status.lowStock")
															: t("status.inStock")}
												</StatusPill>
											</td>
											<td className="py-3 pe-4 text-end">
												<div className="inline-flex items-center gap-1.5">
													<button
														onClick={() => void adjustStock(variant, -1)}
														disabled={savingVariantId === variant.id}
														className="rounded-md border border-border bg-background px-2 py-1 text-sm text-foreground transition-colors hover:bg-secondary disabled:opacity-50"
														title={t("inventory.decrease")}
													>
														−
													</button>
													<input
														type="number"
														value={draftAdjustments[variant.id] ?? ""}
														onChange={(event) =>
															setDraftAdjustments((current) => ({
																...current,
																[variant.id]: event.target.value,
															}))
														}
														placeholder={t("inventory.deltaPlaceholder")}
														className="w-16 rounded-md border border-border bg-background px-2 py-1 text-center text-sm text-foreground"
														aria-label={t("inventory.deltaPlaceholder")}
													/>
													<button
														onClick={() => handleDraftSubmit(variant)}
														disabled={savingVariantId === variant.id}
														className="rounded-md bg-primary px-2 py-1 text-xs font-medium text-primary-foreground disabled:opacity-50"
													>
														{t("inventory.applyDelta")}
													</button>
													<button
														onClick={() => void adjustStock(variant, 1)}
														disabled={savingVariantId === variant.id}
														className="rounded-md border border-border bg-background px-2 py-1 text-sm text-foreground transition-colors hover:bg-secondary disabled:opacity-50"
														title={t("inventory.increase")}
													>
														+
													</button>
												</div>
											</td>
										</tr>
									))}
									{filtered.length === 0 ? (
										<tr>
											<td colSpan={6} className="py-10 text-center text-sm text-muted-foreground">
												{t("inventory.empty")}
											</td>
										</tr>
									) : null}
								</tbody>
							</ResponsiveTable>
						</div>
					</section>
				</>
			)}
		</div>
	);
}
