"use client";

import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback, useRef } from "react";
import { useSearchParams } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { adminService } from "@/modules/admin/api/adminService";
import { catalogService } from "@/modules/catalog/api/catalogService";
import type { AdminCategory } from "@/modules/catalog/types";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";
import { normalizeMediaUrl } from "@/shared/lib/media";
import { Link } from "@/shared/i18n/navigation";

interface ProductView {
	id: string;
	name: string;
	sku: string;
	brand: string;
	category: string;
	price: string;
	status: string;
	image: string;
}

const KNOWN_STATUSES = ["all", "published", "draft", "archived", "outofstock"];

function categoryName(category: unknown): string {
	if (typeof category === "string") return category;
	if (category && typeof category === "object" && "name" in category) return String((category as { name?: string }).name || "—");
	return "—";
}
function brandName(brand: unknown): string {
	if (typeof brand === "string") return brand;
	if (brand && typeof brand === "object" && "name" in brand) return String((brand as { name?: string }).name || "—");
	return "—";
}

export function AdminProductsSearch() {
	const searchParams = useSearchParams();
	const initialStatus = (() => {
		const s = (searchParams.get("status") || "all").toLowerCase();
		return KNOWN_STATUSES.includes(s) ? s : "all";
	})();
	const t = useTranslations("admin");
	const locale = useLocale();
	const [rows, setRows] = useState<ProductView[]>([]);
	const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
	const [categories, setCategories] = useState<AdminCategory[]>([]);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);

	const [searchQuery, setSearchQuery] = useState("");
	const [debouncedSearch, setDebouncedSearch] = useState("");
	const [categoryFilter, setCategoryFilter] = useState("all");
	const [statusFilter, setStatusFilter] = useState(initialStatus);

	const [categoryOpen, setCategoryOpen] = useState(false);
	const [statusOpen, setStatusOpen] = useState(false);
	const categoryRef = useRef<HTMLDivElement>(null);
	const statusRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const id = setTimeout(() => setDebouncedSearch(searchQuery.trim()), 300);
		return () => clearTimeout(id);
	}, [searchQuery]);

	const handleSearchKey = (e: React.KeyboardEvent) => {
		if (e.key === "Enter") {
			setDebouncedSearch(searchQuery.trim());
		}
	};

	useEffect(() => {
		let cancelled = false;
		catalogService
			.getAdminCategories(100)
			.then((res) => {
				if (cancelled) return;
				const list = Array.isArray(res) ? res : (res.data ?? []);
				setCategories(
					(list as (AdminCategory & { deletedAt?: string })[]).filter((c) => !c.deletedAt)
				);
			})
			.catch(() => undefined);
		return () => {
			cancelled = true;
		};
	}, []);

	useEffect(() => {
		const onDown = (e: MouseEvent) => {
			if (!categoryRef.current?.contains(e.target as Node)) setCategoryOpen(false);
			if (!statusRef.current?.contains(e.target as Node)) setStatusOpen(false);
		};
		const onKey = (e: KeyboardEvent) => {
			if (e.key === "Escape") {
				setCategoryOpen(false);
				setStatusOpen(false);
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

	const loadProducts = useCallback(async () => {
		try {
			setLoading(true);
			setLoadError(false);
			const res = await adminService.getProducts({
				cursor: cursor || undefined,
				pageSize: 50,
				search: debouncedSearch || undefined,
				category: categoryFilter !== "all" ? categoryFilter : undefined,
				status: statusFilter !== "all" && statusFilter !== "outofstock" ? statusFilter : undefined,
			});
			const list = res?.data ?? [];
			const formatted = list.map((p) => ({
				id: p.id || p.slug || "—",
				name: p.title || (p.product && (locale === "ar" ? p.product.ar || p.product.en : p.product.en || p.product.ar)) || p.name || "—",
				sku: p.sku || "—",
				brand: brandName(p.brand),
				category: categoryName(p.category),
				price: `EGP ${(((p.extractedPrice ?? (typeof p.price === "number" ? p.price : 0)) || 0)).toFixed(2)}`,
				status: p.status || "draft",
				image: normalizeMediaUrl(p.thumbnail || p.img || p.image),
			}));
			setRows(formatted);
			if (res?.meta && "nextCursor" in res.meta && "hasMore" in res.meta) setMeta(res.meta as CursorMeta);
		} catch {
			setLoadError(true);
		} finally {
			setLoading(false);
		}
	}, [cursor, debouncedSearch, categoryFilter, statusFilter, locale]);

	useEffect(() => {
		setStatusFilter(initialStatus);
		reset();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [initialStatus]);

	useEffect(() => {
		reset();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [debouncedSearch, categoryFilter, statusFilter]);

	useEffect(() => {
		loadProducts();
	}, [loadProducts]);

	const categoryOptions = [
		{ value: "all", label: t("products.allCategories") },
		...categories.map((c) => ({ value: c.name, label: c.name })),
	];
	const statusOptions = [
		{ value: "all", label: t("status.all") },
		{ value: "published", label: t("status.published") },
		{ value: "draft", label: t("status.draft") },
		{ value: "archived", label: t("status.archived") },
		{ value: "outofstock", label: t("products.cardOutOfStock") },
	];
	const categoryLabel = categoryOptions.find((o) => o.value === categoryFilter)?.label ?? categoryFilter;
	const statusLabel = statusOptions.find((o) => o.value === statusFilter)?.label ?? statusFilter;

	return (
		<div className="space-y-6">
			<div className="flex flex-col gap-3">
				<Link
					href="/admin/products"
					className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
				>
					{t("products.backToProductsOverview")}
				</Link>
				<PageHeader title={t("products.searchTitle")} description={t("products.searchSubtitle")} />
			</div>

			<div className="rounded-xl border border-border/70 bg-card p-4 sm:p-6 space-y-4">
				<div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
					<div className="sm:col-span-1 space-y-1.5">
						<label htmlFor="productsSearchQuery" className="text-xs font-medium text-muted-foreground">
							{t("products.searchLabel")}
						</label>
						<input
							id="productsSearchQuery"
							type="text"
							value={searchQuery}
							onChange={(e) => setSearchQuery(e.target.value)}
							onKeyDown={handleSearchKey}
							placeholder={t("products.searchPlaceholderProducts")}
							className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground focus:outline-hidden focus:border-border hover:bg-muted/30 transition-colors"
						/>
					</div>

					<div className="space-y-1.5 relative" ref={categoryRef}>
						<span id="productsSearchCategoryLabel" className="text-xs font-medium text-muted-foreground">
							{t("products.colCategory")}
						</span>
						<button
							type="button"
							onClick={() => setCategoryOpen((v) => !v)}
							aria-expanded={categoryOpen}
							aria-haspopup="listbox"
							aria-labelledby="productsSearchCategoryLabel"
							className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground text-start hover:bg-muted/30 transition-colors cursor-pointer"
						>
							{categoryLabel}
						</button>
						{categoryOpen && (
							<div role="listbox" aria-label={t("products.colCategory")} className="absolute z-50 mt-2 w-full rounded-xl border border-border/70 bg-card py-1.5 shadow-lg overflow-hidden">
								{categoryOptions.map((o) => (
									<button
										key={o.value}
										type="button"
										role="option"
										aria-selected={o.value === categoryFilter}
										onClick={() => {
											setCategoryFilter(o.value);
											setCategoryOpen(false);
										}}
										className={`w-full px-3 py-2 text-start text-sm transition-colors cursor-pointer ${o.value === categoryFilter ? "font-semibold text-foreground bg-muted/40" : "text-muted-foreground hover:text-foreground hover:bg-muted/30"}`}
									>
										{o.label}
									</button>
								))}
							</div>
						)}
					</div>

					<div className="space-y-1.5 relative" ref={statusRef}>
						<span id="productsSearchStatusLabel" className="text-xs font-medium text-muted-foreground">
							{t("products.colStatus")}
						</span>
						<button
							type="button"
							onClick={() => setStatusOpen((v) => !v)}
							aria-expanded={statusOpen}
							aria-haspopup="listbox"
							aria-labelledby="productsSearchStatusLabel"
							className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground text-start hover:bg-muted/30 transition-colors cursor-pointer"
						>
							{statusLabel}
						</button>
						{statusOpen && (
							<div role="listbox" aria-label={t("products.colStatus")} className="absolute z-50 mt-2 w-full rounded-xl border border-border/70 bg-card py-1.5 shadow-lg overflow-hidden">
								{statusOptions.map((o) => (
									<button
										key={o.value}
										type="button"
										role="option"
										aria-selected={o.value === statusFilter}
										onClick={() => {
											setStatusFilter(o.value);
											setStatusOpen(false);
										}}
										className={`w-full px-3 py-2 text-start text-sm transition-colors cursor-pointer ${o.value === statusFilter ? "font-semibold text-foreground bg-muted/40" : "text-muted-foreground hover:text-foreground hover:bg-muted/30"}`}
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
				<ErrorState onRetry={loadProducts} />
			) : (
				<div className="overflow-hidden rounded-xl border border-border/70 bg-card">
					<div className="overflow-x-auto">
						<ResponsiveTable className="w-full text-start text-sm">
							<thead className="text-xs text-muted-foreground bg-muted/30">
								<tr className="border-b border-border/70">
									<th className="py-3 ps-4 text-start font-medium">{t("products.colProduct")}</th>
									<th className="py-3 text-start font-medium">{t("products.colCategory")}</th>
									<th className="py-3 text-start font-medium">{t("products.colPrice")}</th>
									<th className="py-3 text-start font-medium">{t("products.colStatus")}</th>
									<th className="py-3 pe-4 text-end font-medium">{t("products.colActions")}</th>
								</tr>
							</thead>
							<tbody className="divide-y divide-border/50">
								{loading ? (
									<tr>
										<td colSpan={5} className="py-12 text-center text-sm text-muted-foreground">{t("common.loading")}</td>
									</tr>
								) : rows.length === 0 ? (
									<tr>
										<td colSpan={5} className="py-12 text-center text-sm text-muted-foreground">{t("products.empty")}</td>
									</tr>
								) : (
									rows.map((row) => (
										<tr key={row.id} className="hover:bg-muted/40 transition-colors">
											<td className="py-3.5 ps-4">
												<div className="flex items-center gap-3">
													<img src={row.image} alt="" className="size-10 shrink-0 rounded-md border border-border object-cover" />
													<div className="min-w-0">
														<div className="truncate text-sm font-medium text-foreground">{row.name}</div>
														<div className="truncate text-xs text-muted-foreground">{row.sku} · {row.brand}</div>
													</div>
												</div>
											</td>
											<td className="py-3.5">{row.category}</td>
											<td className="py-3.5 font-semibold tabular-nums">{row.price}</td>
											<td className="py-3.5">
												<StatusPill tone={statusTone(row.status)}>{t(`status.${String(row.status).toLowerCase()}`)}</StatusPill>
											</td>
											<td className="py-3.5 pe-4 text-end">
												<Link href={`/admin/products/${row.id}`} className="text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors">
													{t("products.editAction")}
												</Link>
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
						shown={rows.length}
					/>
				</div>
			)}
		</div>
	);
}
