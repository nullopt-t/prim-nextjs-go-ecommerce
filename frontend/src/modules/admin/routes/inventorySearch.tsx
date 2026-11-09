"use client";

import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback, useRef } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { adminService } from "@/modules/admin/api/adminService";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill } from "@/shared/ui/statusPill";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";
import { Link } from "@/shared/i18n/navigation";

interface InventoryProduct {
	id: string;
	variantId: string;
	productId: string;
	name: string;
	sku: string;
	stock: number;
	available: number;
	image: string;
	lowStockThreshold: number;
}

const KNOWN_STOCK_FILTERS = ["all", "in", "low", "out"];

function stockBucket(stock: number, low: number): string {
	if (stock <= 0) return "out";
	if (stock <= low) return "low";
	return "in";
}

export function AdminInventorySearch() {
	const searchParams = useSearchParams();
	const initialStock = (() => {
		const s = (searchParams.get("stock") || "all").toLowerCase();
		return KNOWN_STOCK_FILTERS.includes(s) ? s : "all";
	})();
	const t = useTranslations("admin");

	const [products, setProducts] = useState<InventoryProduct[]>([]);
	const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);

	const [searchQuery, setSearchQuery] = useState("");
	const [debouncedSearch, setDebouncedSearch] = useState("");
	const [stockFilter, setStockFilter] = useState(initialStock);

	const [stockOpen, setStockOpen] = useState(false);
	const stockRef = useRef<HTMLDivElement>(null);

	const [selectedProduct, setSelectedProduct] = useState<InventoryProduct | null>(null);
	const [adjustAmount, setAdjustAmount] = useState(0);
	const [adjustReason, setAdjustReason] = useState("restock");
	const [reasonOpen, setReasonOpen] = useState(false);
	const reasonRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const id = setTimeout(() => setDebouncedSearch(searchQuery.trim()), 300);
		return () => clearTimeout(id);
	}, [searchQuery]);

	const handleSearchKey = (e: React.KeyboardEvent) => {
		if (e.key === "Enter") setDebouncedSearch(searchQuery.trim());
	};

	useEffect(() => {
		const onDown = (e: MouseEvent) => {
			if (!stockRef.current?.contains(e.target as Node)) setStockOpen(false);
			if (!reasonRef.current?.contains(e.target as Node)) setReasonOpen(false);
		};
		const onKey = (e: KeyboardEvent) => {
			if (e.key === "Escape") {
				setStockOpen(false);
				setReasonOpen(false);
			}
		};
		document.addEventListener("mousedown", onDown);
		document.addEventListener("keydown", onKey);
		return () => {
			document.removeEventListener("mousedown", onDown);
			document.removeEventListener("keydown", onKey);
		};
	}, []);

	const { cursor, hasPrev, next, prev, reset, page } = useCursorPagination();

	const loadInventoryData = useCallback(async () => {
		try {
			setLoading(true);
			setLoadError(false);
			const res = await adminService.getInventory({
				cursor: cursor || undefined,
				pageSize: 50,
				search: debouncedSearch || undefined,
			});
			const items = (res?.data ?? []) as unknown as {
				variantId: string;
				productId: string;
				productName: string;
				variantTitle?: string;
				sku: string;
				thumbnail?: string;
				onHandQuantity: number;
				availableQuantity: number;
			}[];
			setProducts(
				items.map((it) => ({
					id: it.variantId || it.productId,
					variantId: it.variantId,
					productId: it.productId,
					name: it.variantTitle ? `${it.productName} (${it.variantTitle})` : it.productName,
					sku: it.sku || "—",
					stock: it.onHandQuantity ?? 0,
					available: it.availableQuantity ?? 0,
					image: it.thumbnail || "",
					lowStockThreshold: 10,
				}))
			);
			if (res?.meta) setMeta(res.meta as CursorMeta);
		} catch {
			setLoadError(true);
		} finally {
			setLoading(false);
		}
	}, [cursor, debouncedSearch]);

	useEffect(() => {
		setStockFilter(initialStock);
		reset();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [initialStock]);

	useEffect(() => {
		reset();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [debouncedSearch]);

	useEffect(() => {
		loadInventoryData();
	}, [loadInventoryData]);

	const handleAdjustStock = (prod: InventoryProduct) => {
		setSelectedProduct(prod);
		setAdjustAmount(0);
		setAdjustReason("restock");
	};

	const handleApplyAdjustment = async () => {
		if (!selectedProduct || adjustAmount === 0) return;
		try {
			if (selectedProduct.variantId) {
				await adminService.adjustInventory(selectedProduct.variantId, adjustAmount, adjustReason);
			} else if (selectedProduct.productId) {
				await adminService.adjustProductInventory(selectedProduct.productId, adjustAmount, adjustReason);
			} else {
				notify.error(t("inventory.toast.noId"));
				return;
			}
			notify.success(t("inventory.toast.adjusted", { name: selectedProduct.name }));
			setSelectedProduct(null);
			loadInventoryData();
		} catch (err) {
			notify.error(apiErrorMessage(err, t("inventory.toast.adjustFailed")));
		}
	};

	const filtered =
		stockFilter === "all"
			? products
			: products.filter((p) => stockBucket(p.stock, p.lowStockThreshold) === stockFilter);

	const stockOptions = [
		{ value: "all", label: t("inventory.stockAll") },
		{ value: "in", label: t("inventory.stockInStock") },
		{ value: "low", label: t("inventory.stockLow") },
		{ value: "out", label: t("inventory.stockOut") },
	];
	const stockLabel = stockOptions.find((o) => o.value === stockFilter)?.label ?? stockFilter;

	const reasonOptions = [
		{ value: "restock", label: t("inventory.reasonRestock") },
		{ value: "adjustment", label: t("inventory.reasonAdjustment") },
		{ value: "return", label: t("inventory.reasonReturn") },
		{ value: "sale", label: t("inventory.reasonSale") },
	];
	const reasonLabel = reasonOptions.find((o) => o.value === adjustReason)?.label ?? adjustReason;

	return (
		<div className="space-y-6">
			<div className="flex flex-col gap-3">
				<Link href="/admin/inventory" className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors">
					{t("inventory.backToInventoryOverview")}
				</Link>
				<PageHeader title={t("inventory.searchTitle")} description={t("inventory.searchSubtitle")} />
			</div>

			<div className="rounded-xl border border-border/70 bg-card p-4 sm:p-6 space-y-4">
				<div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
					<div className="space-y-1.5">
						<label htmlFor="inventorySearchQuery" className="text-xs font-medium text-muted-foreground">
							{t("inventory.searchLabel")}
						</label>
						<input
							id="inventorySearchQuery"
							type="text"
							value={searchQuery}
							onChange={(e) => setSearchQuery(e.target.value)}
							onKeyDown={handleSearchKey}
							placeholder={t("inventory.searchPlaceholderInventory")}
							className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground focus:outline-hidden focus:border-border hover:bg-muted/30 transition-colors"
						/>
					</div>

					<div className="space-y-1.5 relative" ref={stockRef}>
						<span id="inventorySearchStockLabel" className="text-xs font-medium text-muted-foreground">{t("inventory.stockFilter")}</span>
						<button
							type="button"
							onClick={() => setStockOpen((v) => !v)}
							aria-expanded={stockOpen}
							aria-haspopup="listbox"
							aria-labelledby="inventorySearchStockLabel"
							className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground text-start hover:bg-muted/30 transition-colors cursor-pointer"
						>
							{stockLabel}
						</button>
						{stockOpen && (
							<div role="listbox" aria-label={t("inventory.stockFilter")} className="absolute z-50 mt-2 w-full rounded-xl border border-border/70 bg-card py-1.5 shadow-lg overflow-hidden">
								{stockOptions.map((o) => (
									<button
										key={o.value}
										type="button"
										role="option"
										aria-selected={o.value === stockFilter}
										onClick={() => {
											setStockFilter(o.value);
											setStockOpen(false);
										}}
										className={`w-full px-3 py-2 text-start text-sm transition-colors cursor-pointer ${o.value === stockFilter ? "font-semibold text-foreground bg-muted/40" : "text-muted-foreground hover:text-foreground hover:bg-muted/30"}`}
									>
										{o.label}
									</button>
								))}
							</div>
						)}
					</div>
				</div>
			</div>

			{loadError ? (
				<ErrorState onRetry={loadInventoryData} />
			) : (
				<div className="overflow-hidden rounded-xl border border-border/70 bg-card">
					<div className="overflow-x-auto">
						<ResponsiveTable className="w-full text-start text-sm">
							<thead className="text-xs text-muted-foreground bg-muted/30">
								<tr className="border-b border-border/70">
									<th className="py-3 ps-4 text-start font-medium">{t("inventory.colProduct")}</th>
									<th className="py-3 text-start font-medium">{t("inventory.colStockOnHand")}</th>
									<th className="py-3 text-start font-medium">{t("inventory.colSku")}</th>
									<th className="py-3 text-start font-medium">{t("inventory.colStatus")}</th>
									<th className="py-3 pe-4 text-end font-medium">{t("inventory.colActions")}</th>
								</tr>
							</thead>
							<tbody className="divide-y divide-border/50">
								{loading ? (
									<tr>
										<td colSpan={5} className="py-12 text-center text-sm text-muted-foreground">{t("common.loading")}</td>
									</tr>
								) : filtered.length === 0 ? (
									<tr>
										<td colSpan={5} className="py-12 text-center text-sm text-muted-foreground">{t("inventory.catalogItems", { count: 0 })}</td>
									</tr>
								) : (
									filtered.map((prod) => (
										<tr key={prod.id} className="hover:bg-muted/40 transition-colors">
											<td className="py-3.5 ps-4">
												<div className="flex items-center gap-3">
													{prod.image ? (
														<img src={prod.image} alt="" loading="lazy" className="size-10 shrink-0 rounded-md border border-border object-cover" />
													) : (
														<div className="size-10 shrink-0 rounded-md border border-border bg-secondary/40" />
													)}
													<div className="min-w-0">
														<div className="truncate font-medium text-foreground">{prod.name}</div>
														<div className="truncate font-mono text-xs text-muted-foreground">{prod.sku}</div>
													</div>
												</div>
											</td>
											<td className="py-3.5 font-medium tabular-nums">{t("inventory.units", { count: prod.stock })}</td>
											<td className="py-3.5 font-medium tabular-nums">{t("inventory.units", { count: prod.available })}</td>
											<td className="py-3.5">
												<StatusPill tone={prod.stock <= prod.lowStockThreshold ? "warning" : "success"}>
													{prod.stock <= prod.lowStockThreshold ? t("status.lowStock") : t("status.inStock")}
												</StatusPill>
											</td>
											<td className="py-3.5 pe-4 text-end">
												<button type="button" onClick={() => handleAdjustStock(prod)} className="text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors cursor-pointer">
													{t("inventory.adjust")}
												</button>
											</td>
										</tr>
									))
								)}
							</tbody>
						</ResponsiveTable>
					</div>
					<CursorPager
						hasPrev={hasPrev}
						hasMore={meta.hasMore}
						page={page}
						onPrev={prev}
						onNext={() => next(meta.nextCursor)}
						loading={loading}
						shown={filtered.length}
					/>
				</div>
			)}

			{selectedProduct && (
				<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
					<div className="w-full max-w-sm space-y-4 rounded-xl border border-border/70 bg-card p-5">
						<h3 className="text-sm font-semibold text-foreground">{t("inventory.adjustTitle", { name: selectedProduct.name })}</h3>
						<div className="text-sm text-muted-foreground">
							{t("inventory.currentStockLabel")} <strong className="font-medium text-foreground">{t("inventory.units", { count: selectedProduct.stock })}</strong>
						</div>
						<div className="space-y-3">
							<div>
								<span className="text-sm font-medium text-foreground">{t("inventory.stockChange")}</span>
								<div className="mt-1 flex items-center gap-2">
									<button type="button" onClick={() => setAdjustAmount((prev) => prev - 1)} className="rounded-xl border border-border/70 px-3 py-2 hover:bg-muted/30 transition-colors cursor-pointer">−</button>
									<input
										type="number"
										value={adjustAmount}
										onChange={(e) => setAdjustAmount(parseInt(e.target.value, 10) || 0)}
										className="w-full rounded-xl border border-border bg-background py-2 text-center text-sm font-medium tabular-nums text-foreground"
									/>
									<button type="button" onClick={() => setAdjustAmount((prev) => prev + 1)} className="rounded-xl border border-border/70 px-3 py-2 hover:bg-muted/30 transition-colors cursor-pointer">+</button>
								</div>
							</div>
							<div className="relative" ref={reasonRef}>
								<span className="text-sm font-medium text-foreground">{t("inventory.reasonLabel")}</span>
								<button
									type="button"
									onClick={() => setReasonOpen((v) => !v)}
									aria-expanded={reasonOpen}
									aria-haspopup="listbox"
									className="mt-1 w-full rounded-xl border border-border bg-background px-3 py-2 text-sm text-foreground text-start hover:bg-muted/30 transition-colors cursor-pointer"
								>
									{reasonLabel}
								</button>
								{reasonOpen && (
									<div role="listbox" aria-label={t("inventory.reasonLabel")} className="absolute z-50 mt-2 w-full rounded-xl border border-border/70 bg-card py-1.5 shadow-lg overflow-hidden">
										{reasonOptions.map((o) => (
											<button
												key={o.value}
												type="button"
												role="option"
												aria-selected={o.value === adjustReason}
												onClick={() => {
													setAdjustReason(o.value);
													setReasonOpen(false);
												}}
												className={`w-full px-3 py-2 text-start text-sm transition-colors cursor-pointer ${o.value === adjustReason ? "font-semibold text-foreground bg-muted/40" : "text-muted-foreground hover:text-foreground hover:bg-muted/30"}`}
											>
												{o.label}
											</button>
										))}
									</div>
								)}
							</div>
						</div>
						<div className="flex justify-end gap-2 border-t border-border/50 pt-3">
							<button type="button" onClick={() => setSelectedProduct(null)} className="rounded-xl border border-border/70 px-3 py-2 text-sm font-medium text-foreground hover:bg-muted/30 transition-colors cursor-pointer">{t("actions.cancel")}</button>
							<button type="button" onClick={handleApplyAdjustment} className="rounded-xl bg-foreground px-3 py-2 text-sm font-medium text-background transition-colors cursor-pointer">{t("inventory.applyAdjustment")}</button>
						</div>
					</div>
				</div>
			)}
		</div>
	);
}
