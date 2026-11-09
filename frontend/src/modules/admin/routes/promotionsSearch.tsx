"use client";

import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback, useRef } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { adminService } from "@/modules/admin/api/adminService";
import type { Promotion } from "@/modules/promotions/types";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";
import { Link } from "@/shared/i18n/navigation";
import { ConfirmDialog } from "@/shared/ui/confirmDialog";
import { buildPromotionPayload } from "@/modules/admin/lib/promotionPayload";

interface PromotionView {
	id: string;
	code: string;
	discountType: string;
	value: string;
	usageLimit: number;
	currentUsage: number;
	status: string;
	startsAt: string;
	expiresAt: string;
}

const KNOWN_STATUSES = ["all", "active", "disabled", "expired"];

export function AdminPromotionsSearch() {
	const searchParams = useSearchParams();
	const initialStatus = (() => {
		const s = (searchParams.get("status") || "all").toLowerCase();
		return KNOWN_STATUSES.includes(s) ? s : "all";
	})();
	const t = useTranslations("admin");

	const [promotions, setPromotions] = useState<PromotionView[]>([]);
	const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);

	const [searchQuery, setSearchQuery] = useState("");
	const [debouncedSearch, setDebouncedSearch] = useState("");
	const [statusFilter, setStatusFilter] = useState(initialStatus);

	const [statusOpen, setStatusOpen] = useState(false);
	const statusRef = useRef<HTMLDivElement>(null);

	const [isModalOpen, setIsModalOpen] = useState(false);
	const [formData, setFormData] = useState({
		code: "",
		type: "Percentage",
		value: "15",
		usageLimit: 200,
		startDate: "2026-08-01",
		endDate: "2026-12-31",
	});
	const [pendingDelete, setPendingDelete] = useState<string | null>(null);
	const [deleting, setDeleting] = useState(false);

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

	const loadPromos = useCallback(async () => {
		try {
			setLoading(true);
			setLoadError(false);
			const res = await adminService.getPromotions({
				cursor: cursor || undefined,
				pageSize: 50,
				search: debouncedSearch || undefined,
				status: statusFilter !== "all" ? statusFilter : undefined,
			});
			const list = res?.data ?? [];
			setPromotions(
				list.map((p: Promotion) => ({
					id: p.id,
					code: p.code,
					discountType: p.discountType,
					value: p.discountType === "percentage" ? `${p.discountValue}%` : `EGP ${(p.discountValue / 100).toFixed(2)}`,
					usageLimit: p.usageLimit ?? 0,
					currentUsage: p.usageCount ?? 0,
					status: p.status ?? "",
					startsAt: p.startsAt ?? "",
					expiresAt: p.expiresAt ?? "",
				}))
			);
			if (res?.meta) setMeta(res.meta);
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
		loadPromos();
	}, [loadPromos]);

	const handleToggleStatus = async (id: string, currentStatus: string) => {
		const isActive = currentStatus.toLowerCase() === "active";
		const nextStatus = isActive ? "disabled" : "active";
		try {
			await adminService.updatePromotionStatus(id, nextStatus);
			notify.success(t("promotions.toast.statusUpdated"));
			loadPromos();
		} catch (err) {
			notify.error(apiErrorMessage(err, t("promotions.toast.statusFailed")));
		}
	};

	const confirmDelete = async () => {
		if (!pendingDelete) return;
		setDeleting(true);
		try {
			await adminService.deletePromotion(pendingDelete);
			notify.success(t("promotions.toast.deleted"));
			setPendingDelete(null);
			loadPromos();
		} catch (err) {
			notify.error(apiErrorMessage(err, t("promotions.toast.deleteFailed")));
		} finally {
			setDeleting(false);
		}
	};

	const handleCreatePromo = async (e: React.FormEvent) => {
		e.preventDefault();
		const payload = buildPromotionPayload(formData);
		if (!payload) {
			notify.error(t("promotions.toast.createFailed"));
			return;
		}
		try {
			await adminService.createPromotion(payload);
			notify.success(t("promotions.toast.created", { code: payload.code }));
			setIsModalOpen(false);
			loadPromos();
		} catch (err) {
			notify.error(apiErrorMessage(err, t("promotions.toast.createFailed")));
		}
	};

	const statusOptions = [
		{ value: "all", label: t("status.all") },
		{ value: "active", label: t("status.active") },
		{ value: "disabled", label: t("status.disabled") },
		{ value: "expired", label: t("promotions.cardExpired") },
	];
	const statusLabel = statusOptions.find((o) => o.value === statusFilter)?.label ?? statusFilter;

	return (
		<div className="space-y-6">
			<div className="flex flex-col gap-3">
				<Link href="/admin/promotions" className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors">
					{t("promotions.backToPromotionsOverview")}
				</Link>
				<PageHeader title={t("promotions.searchTitle")} description={t("promotions.searchSubtitle")} />
			</div>

			<div className="flex justify-end">
				<button type="button" onClick={() => setIsModalOpen(true)} className="inline-flex items-center rounded-xl border border-border/70 bg-card px-4 py-2.5 text-sm font-semibold text-foreground hover:bg-muted/30 transition-colors cursor-pointer active:scale-[0.99] select-none">
					{t("promotions.create")}
				</button>
			</div>

			<div className="rounded-xl border border-border/70 bg-card p-4 sm:p-6 space-y-4">
				<div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
					<div className="space-y-1.5">
						<label htmlFor="promotionsSearchQuery" className="text-xs font-medium text-muted-foreground">
							{t("promotions.searchLabel")}
						</label>
						<input
							id="promotionsSearchQuery"
							type="text"
							value={searchQuery}
							onChange={(e) => setSearchQuery(e.target.value)}
							onKeyDown={handleSearchKey}
							placeholder={t("promotions.searchPlaceholderPromotions")}
							className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground focus:outline-hidden focus:border-border hover:bg-muted/30 transition-colors"
						/>
					</div>

					<div className="space-y-1.5 relative" ref={statusRef}>
						<span id="promotionsSearchStatusLabel" className="text-xs font-medium text-muted-foreground">{t("promotions.colStatus")}</span>
						<button
							type="button"
							onClick={() => setStatusOpen((v) => !v)}
							aria-expanded={statusOpen}
							aria-haspopup="listbox"
							aria-labelledby="promotionsSearchStatusLabel"
							className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground text-start hover:bg-muted/30 transition-colors cursor-pointer"
						>
							{statusLabel}
						</button>
						{statusOpen && (
							<div role="listbox" aria-label={t("promotions.colStatus")} className="absolute z-50 mt-2 w-full rounded-xl border border-border/70 bg-card py-1.5 shadow-lg overflow-hidden">
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
				<ErrorState onRetry={loadPromos} />
			) : (
				<div className="overflow-hidden rounded-xl border border-border/70 bg-card">
					<div className="overflow-x-auto">
						<ResponsiveTable className="w-full text-start text-sm">
							<thead className="text-xs text-muted-foreground bg-muted/30">
								<tr className="border-b border-border/70">
									<th className="py-3 ps-4 text-start font-medium">{t("promotions.colCode")}</th>
									<th className="py-3 text-start font-medium">{t("promotions.colDiscount")}</th>
									<th className="py-3 text-start font-medium">{t("promotions.colUsage")}</th>
									<th className="py-3 text-start font-medium">{t("promotions.colStatus")}</th>
									<th className="py-3 pe-4 text-end font-medium">{t("promotions.colActions")}</th>
								</tr>
							</thead>
							<tbody className="divide-y divide-border/50">
								{loading ? (
									<tr>
										<td colSpan={5} className="py-12 text-center text-sm text-muted-foreground">{t("common.loading")}</td>
									</tr>
								) : promotions.length === 0 ? (
									<tr>
										<td colSpan={5} className="py-12 text-center text-sm text-muted-foreground">{t("promotions.empty")}</td>
									</tr>
								) : (
									promotions.map((promo) => (
										<tr key={promo.id} className="hover:bg-muted/40 transition-colors">
											<td className="py-3.5 ps-4">
												<span className="inline-flex rounded-md border border-border/70 bg-secondary/40 px-2 py-0.5 font-mono text-xs font-semibold tracking-wider text-foreground">
													{promo.code}
												</span>
											</td>
											<td className="py-3.5">
												<div className="font-medium text-foreground">{promo.value}</div>
												<div className="text-xs text-muted-foreground">
													{t("promotions.typeLabel", { type: promo.discountType === "percentage" ? t("promotions.type.percentage") : t("promotions.type.fixedAmount") })}
												</div>
											</td>
											<td className="py-3.5 text-xs text-muted-foreground">{t("promotions.redeemed", { used: promo.currentUsage, limit: promo.usageLimit })}</td>
											<td className="py-3.5">
												<StatusPill tone={statusTone(promo.status)}>{t(`status.${String(promo.status).toLowerCase()}`) || promo.status}</StatusPill>
											</td>
											<td className="py-3.5 pe-4 text-end">
												<div className="inline-flex items-center justify-end gap-2">
													<button type="button" onClick={() => handleToggleStatus(promo.id, promo.status)} className="text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors cursor-pointer">
														{promo.status.toLowerCase() === "active" ? t("promotions.actions.deactivate") : t("promotions.actions.activate")}
													</button>
													<button type="button" onClick={() => setPendingDelete(promo.id)} className="text-xs font-semibold text-destructive hover:underline cursor-pointer">
														{t("actions.delete")}
													</button>
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
						shown={promotions.length}
					/>
				</div>
			)}

			{isModalOpen && (
				<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
					<div className="w-full max-w-md space-y-4 rounded-xl border border-border/70 bg-card p-5 shadow-xl">
						<div className="flex items-center justify-between border-b border-border/50 pb-3">
							<h3 className="text-sm font-semibold text-foreground">{t("promotions.create")}</h3>
							<button type="button" onClick={() => setIsModalOpen(false)} className="text-muted-foreground hover:text-foreground" aria-label={t("actions.cancel")}>✕</button>
						</div>
						<form onSubmit={handleCreatePromo} className="space-y-4">
							<div className="space-y-1">
								<label className="text-sm font-medium text-foreground">{t("promotions.form.code")}</label>
								<input type="text" required placeholder={t("promotions.form.codePlaceholder")} value={formData.code} onChange={(e) => setFormData({ ...formData, code: e.target.value })} className="w-full rounded-xl border border-border bg-background px-3 py-2 font-mono text-sm uppercase text-foreground" />
							</div>
							<div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
								<div className="space-y-1">
									<label className="text-sm font-medium text-foreground">{t("promotions.form.discountType")}</label>
									<div className="flex gap-2">
										{["Percentage", "Fixed Amount"].map((v) => (
											<button key={v} type="button" onClick={() => setFormData({ ...formData, type: v })} aria-pressed={formData.type === v} className={`flex-1 rounded-xl border px-3 py-2 text-sm transition-colors cursor-pointer ${formData.type === v ? "border-border bg-foreground text-background" : "border-border/70"}`}>
												{v === "Percentage" ? t("promotions.form.typePercentage") : t("promotions.form.typeFixed")}
											</button>
										))}
									</div>
								</div>
								<div className="space-y-1">
									<label className="text-sm font-medium text-foreground">{t("promotions.form.valueLabel")}</label>
									<input type="text" required placeholder={t("promotions.form.valuePlaceholder")} value={formData.value} onChange={(e) => setFormData({ ...formData, value: e.target.value })} className="w-full rounded-xl border border-border bg-background px-3 py-2 text-sm text-foreground" />
								</div>
							</div>
							<div className="space-y-1">
								<label className="text-sm font-medium text-foreground">{t("promotions.form.usageLimit")}</label>
								<input type="number" value={formData.usageLimit} onChange={(e) => setFormData({ ...formData, usageLimit: Number(e.target.value) || 0 })} className="w-full rounded-xl border border-border bg-background px-3 py-2 text-sm text-foreground" />
							</div>
							<div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
								<div className="space-y-1">
									<label className="text-sm font-medium text-foreground">{t("promotions.form.startDate")}</label>
									<input type="date" value={formData.startDate} onChange={(e) => setFormData({ ...formData, startDate: e.target.value })} className="w-full rounded-xl border border-border bg-background px-3 py-2 text-sm text-foreground" />
								</div>
								<div className="space-y-1">
									<label className="text-sm font-medium text-foreground">{t("promotions.form.endDate")}</label>
									<input type="date" value={formData.endDate} min={formData.startDate || undefined} onChange={(e) => setFormData({ ...formData, endDate: e.target.value })} className="w-full rounded-xl border border-border bg-background px-3 py-2 text-sm text-foreground" />
								</div>
							</div>
							<div className="flex justify-end gap-2 border-t border-border/50 pt-3">
								<button type="button" onClick={() => setIsModalOpen(false)} className="rounded-xl border border-border/70 px-3 py-2 text-sm font-medium text-foreground hover:bg-muted/30 transition-colors cursor-pointer">{t("actions.cancel")}</button>
								<button type="submit" className="rounded-xl bg-foreground px-3 py-2 text-sm font-medium text-background transition-colors cursor-pointer">{t("promotions.form.submit")}</button>
							</div>
						</form>
					</div>
				</div>
			)}

			<ConfirmDialog
				isOpen={pendingDelete !== null}
				title={t("promotions.confirmDelete")}
				confirmLabel={t("actions.delete")}
				cancelLabel={t("actions.cancel")}
				loading={deleting}
				onConfirm={confirmDelete}
				onClose={() => setPendingDelete(null)}
			/>
		</div>
	);
}
