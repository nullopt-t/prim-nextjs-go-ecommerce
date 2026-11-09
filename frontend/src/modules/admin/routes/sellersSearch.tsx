"use client";

import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback, useRef } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { adminService, type AdminSeller } from "@/modules/admin/api/adminService";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";
import { Link } from "@/shared/i18n/navigation";

const STATUS_LABEL_KEYS: Record<string, string> = {
	pending: "status.pending",
	active: "status.active",
	suspended: "status.suspended",
	rejected: "status.rejected",
};
const KNOWN_STATUSES = ["all", "pending", "active", "suspended", "rejected"];

export function AdminSellersSearch() {
	const searchParams = useSearchParams();
	const initialStatus = (() => {
		const s = (searchParams.get("status") || "all").toLowerCase();
		return KNOWN_STATUSES.includes(s) ? s : "all";
	})();
	const t = useTranslations("admin");
	const [sellers, setSellers] = useState<AdminSeller[]>([]);
	const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);

	const [searchQuery, setSearchQuery] = useState("");
	const [debouncedSearch, setDebouncedSearch] = useState("");
	const [statusFilter, setStatusFilter] = useState(initialStatus);

	const [statusOpen, setStatusOpen] = useState(false);
	const statusRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const id = setTimeout(() => setDebouncedSearch(searchQuery.trim()), 300);
		return () => clearTimeout(id);
	}, [searchQuery]);

	const handleSearchKey = (e: React.KeyboardEvent) => {
		if (e.key === "Enter") setDebouncedSearch(searchQuery.trim());
	};

	useEffect(() => {
		const onDown = (e: MouseEvent) => {
			if (!statusRef.current?.contains(e.target as Node)) setStatusOpen(false);
		};
		const onKey = (e: KeyboardEvent) => {
			if (e.key === "Escape") setStatusOpen(false);
		};
		document.addEventListener("mousedown", onDown);
		document.addEventListener("keydown", onKey);
		return () => {
			document.removeEventListener("mousedown", onDown);
			document.removeEventListener("keydown", onKey);
		};
	}, []);

	const { cursor, hasPrev, next, prev, reset, page } = useCursorPagination();

	const loadSellers = useCallback(async () => {
		try {
			setLoading(true);
			setLoadError(false);
			const res = await adminService.getSellers({
				cursor: cursor || undefined,
				pageSize: 50,
				search: debouncedSearch || undefined,
				status: statusFilter !== "all" ? statusFilter : undefined,
			});
			setSellers(res.data ?? []);
			if (res.meta) setMeta(res.meta);
		} catch {
			setLoadError(true);
		} finally {
			setLoading(false);
		}
	}, [cursor, debouncedSearch, statusFilter]);

	useEffect(() => {
		setStatusFilter(initialStatus);
		reset();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [initialStatus]);

	useEffect(() => {
		reset();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [debouncedSearch, statusFilter]);

	useEffect(() => {
		loadSellers();
	}, [loadSellers]);

	const handleTransition = async (seller: AdminSeller, to: string) => {
		try {
			const updated = await adminService.updateSellerStatus(seller.id, to);
			notify.success(t("sellers.toast.statusUpdated"));
			if (updated.data) {
				setSellers((prev) => prev.map((v) => (v.id === seller.id ? (updated.data as AdminSeller) : v)));
			}
			await loadSellers();
		} catch (err) {
			notify.error(apiErrorMessage(err, t("sellers.toast.statusFailed")));
		}
	};

	const statusLabel = (status: string) => t(STATUS_LABEL_KEYS[status] ?? status);

	const statusOptions = [
		{ value: "all", label: t("status.all") },
		{ value: "pending", label: t("status.pending") },
		{ value: "active", label: t("status.active") },
		{ value: "suspended", label: t("status.suspended") },
		{ value: "rejected", label: t("status.rejected") },
	];
	const statusLabelText = statusOptions.find((o) => o.value === statusFilter)?.label ?? statusFilter;

	return (
		<div className="space-y-6">
			<div className="flex flex-col gap-3">
				<Link href="/admin/sellers" className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors">
					{t("sellers.backToSellersOverview")}
				</Link>
				<PageHeader title={t("sellers.searchTitle")} description={t("sellers.searchSubtitle")} />
			</div>

			<div className="rounded-xl border border-border/70 bg-card p-4 sm:p-6 space-y-4">
				<div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
					<div className="space-y-1.5">
						<label htmlFor="sellersSearchQuery" className="text-xs font-medium text-muted-foreground">
							{t("sellers.searchLabel")}
						</label>
						<input
							id="sellersSearchQuery"
							type="text"
							value={searchQuery}
							onChange={(e) => setSearchQuery(e.target.value)}
							onKeyDown={handleSearchKey}
							placeholder={t("sellers.searchPlaceholderSellers")}
							className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground focus:outline-hidden focus:border-border hover:bg-muted/30 transition-colors"
						/>
					</div>

					<div className="space-y-1.5 relative" ref={statusRef}>
						<span id="sellersSearchStatusLabel" className="text-xs font-medium text-muted-foreground">{t("sellers.colStatus")}</span>
						<button
							type="button"
							onClick={() => setStatusOpen((v) => !v)}
							aria-expanded={statusOpen}
							aria-haspopup="listbox"
							aria-labelledby="sellersSearchStatusLabel"
							className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground text-start hover:bg-muted/30 transition-colors cursor-pointer"
						>
							{statusLabelText}
						</button>
						{statusOpen && (
							<div role="listbox" aria-label={t("sellers.colStatus")} className="absolute z-50 mt-2 w-full rounded-xl border border-border/70 bg-card py-1.5 shadow-lg overflow-hidden">
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
				<ErrorState onRetry={loadSellers} />
			) : (
				<div className="overflow-hidden rounded-xl border border-border/70 bg-card">
					<div className="overflow-x-auto">
						<ResponsiveTable className="w-full text-start text-sm">
							<thead className="text-xs text-muted-foreground bg-muted/30">
								<tr className="border-b border-border/70">
									<th className="py-3 ps-4 text-start font-medium">{t("sellers.colStore")}</th>
									<th className="py-3 text-start font-medium">{t("sellers.store")}</th>
									<th className="py-3 text-start font-medium">{t("sellers.colCommission")}</th>
									<th className="py-3 text-start font-medium">{t("sellers.colStatus")}</th>
									<th className="py-3 pe-4 text-end font-medium">{t("sellers.colAction")}</th>
								</tr>
							</thead>
							<tbody className="divide-y divide-border/50">
								{loading ? (
									<tr>
										<td colSpan={5} className="py-12 text-center text-sm text-muted-foreground">{t("common.loading")}</td>
									</tr>
								) : sellers.length === 0 ? (
									<tr>
										<td colSpan={5} className="py-12 text-center text-sm text-muted-foreground">{t("sellers.empty")}</td>
									</tr>
								) : (
									sellers.map((seller) => (
										<tr key={seller.id} className="hover:bg-muted/40 transition-colors">
											<td className="py-3.5 ps-4">
												<div className="flex items-center gap-3">
													{seller.logoUrl ? (
														<img src={seller.logoUrl} alt="" className="size-10 shrink-0 rounded-md border border-border object-cover" />
													) : (
														<div className="flex size-10 shrink-0 items-center justify-center rounded-md border border-border bg-secondary text-xs font-semibold text-foreground">
															{seller.storeName.slice(0, 2).toUpperCase()}
														</div>
													)}
													<div className="min-w-0">
														<div className="truncate font-medium text-foreground">{seller.storeName}</div>
														<div className="truncate font-mono text-xs text-muted-foreground">{seller.slug}</div>
													</div>
												</div>
											</td>
											<td className="py-3.5">
												<div className="text-sm text-foreground">{seller.storeName}</div>
												<div className="truncate font-mono text-xs text-muted-foreground">{seller.slug}</div>
											</td>
											<td className="py-3.5 font-medium tabular-nums">{(seller.commissionRateBps / 100).toFixed(2)}%</td>
											<td className="py-3.5">
												<StatusPill tone={statusTone(seller.status)}>{statusLabel(seller.status)}</StatusPill>
											</td>
											<td className="py-3.5 pe-4 text-end">
												<div className="inline-flex items-center justify-end gap-2">
													<Link href={`/admin/sellers/${seller.id}`} className="text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors">
														{t("sellers.manageAction")}
													</Link>
													{seller.status === "pending" && (
														<>
															<button type="button" onClick={() => void handleTransition(seller, "active")} className="text-xs font-semibold hover:underline cursor-pointer">{t("sellers.actions.approve")}</button>
															<button type="button" onClick={() => void handleTransition(seller, "rejected")} className="text-xs font-semibold text-destructive hover:underline cursor-pointer">{t("sellers.actions.reject")}</button>
														</>
													)}
													{seller.status === "active" && (
														<button type="button" onClick={() => void handleTransition(seller, "suspended")} className="text-xs font-semibold hover:underline cursor-pointer">{t("sellers.actions.suspend")}</button>
													)}
													{seller.status === "suspended" && (
														<button type="button" onClick={() => void handleTransition(seller, "active")} className="text-xs font-semibold hover:underline cursor-pointer">{t("sellers.actions.activate")}</button>
													)}
												</div>
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
						shown={sellers.length}
					/>
				</div>
			)}
		</div>
	);
}
