"use client";

import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback, useRef } from "react";
import { useSearchParams } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { adminService, type AdminPayment } from "@/modules/admin/api/adminService";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";
import { Link } from "@/shared/i18n/navigation";

function toISODate(d: Date): string {
	return d.toISOString().slice(0, 10);
}
function applyPreset(preset: string): { from: string; to: string } {
	const now = new Date();
	const today = toISODate(now);
	switch (preset) {
		case "today":
			return { from: today, to: today };
		case "7d": {
			const d = new Date(now);
			d.setDate(d.getDate() - 6);
			return { from: toISODate(d), to: today };
		}
		case "30d": {
			const d = new Date(now);
			d.setDate(d.getDate() - 29);
			return { from: toISODate(d), to: today };
		}
		default:
			return { from: "", to: "" };
	}
}

const KNOWN_STATUSES = ["all", "captured", "pending", "refunded", "failed", "authorized"];

export function AdminPaymentsSearch() {
	const searchParams = useSearchParams();
	const initialStatus = (() => {
		const s = (searchParams.get("status") || "all").toLowerCase();
		return KNOWN_STATUSES.includes(s) ? s : "all";
	})();
	const t = useTranslations("admin");
	const locale = useLocale();
	const { user } = useAuthContext();
	const isSuperadmin = user?.role === "superadmin";

	const [payments, setPayments] = useState<AdminPayment[]>([]);
	const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);

	const [searchQuery, setSearchQuery] = useState("");
	const [debouncedSearch, setDebouncedSearch] = useState("");
	const [statusFilter, setStatusFilter] = useState(initialStatus);
	const [dateFrom, setDateFrom] = useState("");
	const [dateTo, setDateTo] = useState("");
	const [datePreset, setDatePreset] = useState("all");

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

	const loadPayments = useCallback(async () => {
		try {
			setLoading(true);
			setLoadError(false);
			const res = await adminService.getPayments({
				cursor: cursor || undefined,
				pageSize: 50,
				search: debouncedSearch || undefined,
				status: statusFilter !== "all" ? statusFilter : undefined,
				dateFrom: dateFrom || undefined,
				dateTo: dateTo || undefined,
			});
			setPayments(res.data ?? []);
			if (res.meta) setMeta(res.meta as CursorMeta);
		} catch {
			setLoadError(true);
		} finally {
			setLoading(false);
		}
	}, [cursor, debouncedSearch, statusFilter, dateFrom, dateTo]);

	useEffect(() => {
		setStatusFilter(initialStatus);
		reset();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [initialStatus]);

	useEffect(() => {
		reset();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [debouncedSearch, statusFilter, dateFrom, dateTo]);

	useEffect(() => {
		loadPayments();
	}, [loadPayments]);

	const handleIssueRefund = async (pay: AdminPayment) => {
		if (confirm(t("payments.confirmRefund", { id: pay.id }))) {
			try {
				await adminService.refundPayment(pay.id);
				notify.success(t("payments.toast.refundIssued", { customer: pay.customerEmail || "—" }));
				loadPayments();
			} catch (err) {
				notify.error(apiErrorMessage(err, t("payments.toast.refundFailed")));
			}
		}
	};

	const handleMarkPaid = async (pay: AdminPayment) => {
		if (confirm(t("payments.confirmMarkPaid", { id: pay.id }))) {
			try {
				await adminService.updatePaymentStatus(pay.id, "captured");
				notify.success(t("payments.toast.markedPaid", { customer: pay.customerEmail || "—" }));
				loadPayments();
			} catch (err) {
				notify.error(apiErrorMessage(err, t("payments.toast.statusFailed")));
			}
		}
	};

	const pickPreset = (preset: string) => {
		setDatePreset(preset);
		const range = applyPreset(preset);
		setDateFrom(range.from);
		setDateTo(range.to);
	};

	const statusOptions = [
		{ value: "all", label: t("status.all") },
		{ value: "captured", label: t("status.paid") },
		{ value: "pending", label: t("status.pending") },
		{ value: "refunded", label: t("status.refunded") },
		{ value: "failed", label: t("status.failed") },
	];
	const statusLabel = statusOptions.find((o) => o.value === statusFilter)?.label ?? statusFilter;
	const presets = [
		{ id: "all", label: t("orders.dateAllTime") },
		{ id: "today", label: t("orders.dateToday") },
		{ id: "7d", label: t("orders.dateLast7") },
		{ id: "30d", label: t("orders.dateLast30") },
	];

	return (
		<div className="space-y-6">
			<div className="flex flex-col gap-3">
				<Link href="/admin/payments" className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors">
					{t("payments.backToPaymentsOverview")}
				</Link>
				<PageHeader title={t("payments.searchTitle")} description={t("payments.searchSubtitle")} />
			</div>

			<div className="rounded-xl border border-border/70 bg-card p-4 sm:p-6 space-y-4">
				<div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
					<div className="space-y-1.5">
						<label htmlFor="paymentsSearchQuery" className="text-xs font-medium text-muted-foreground">
							{t("payments.searchLabel")}
						</label>
						<input
							id="paymentsSearchQuery"
							type="text"
							value={searchQuery}
							onChange={(e) => setSearchQuery(e.target.value)}
							onKeyDown={handleSearchKey}
							placeholder={t("payments.searchPlaceholderPayments")}
							className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground focus:outline-hidden focus:border-border hover:bg-muted/30 transition-colors"
						/>
					</div>

					<div className="space-y-1.5 relative" ref={statusRef}>
						<span id="paymentsSearchStatusLabel" className="text-xs font-medium text-muted-foreground">{t("payments.colStatus")}</span>
						<button
							type="button"
							onClick={() => setStatusOpen((v) => !v)}
							aria-expanded={statusOpen}
							aria-haspopup="listbox"
							aria-labelledby="paymentsSearchStatusLabel"
							className="w-full h-11 px-3 rounded-xl border border-border bg-background text-sm text-foreground text-start hover:bg-muted/30 transition-colors cursor-pointer"
						>
							{statusLabel}
						</button>
						{statusOpen && (
							<div role="listbox" aria-label={t("payments.colStatus")} className="absolute z-50 mt-2 w-full rounded-xl border border-border/70 bg-card py-1.5 shadow-lg overflow-hidden">
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

				<div className="flex flex-wrap items-center gap-2 pt-2 border-t border-border/50">
					{presets.map((p) => (
						<button
							key={p.id}
							type="button"
							onClick={() => pickPreset(p.id)}
							aria-pressed={datePreset === p.id}
							className={`h-9 px-3 rounded-full border text-xs font-medium transition-colors cursor-pointer ${datePreset === p.id ? "border-border bg-foreground text-background" : "border-border/70 text-muted-foreground hover:text-foreground hover:bg-muted/30"}`}
						>
							{p.label}
						</button>
					))}
				</div>
			</div>

			{loadError ? (
				<ErrorState onRetry={loadPayments} />
			) : (
				<div className="overflow-hidden rounded-xl border border-border/70 bg-card">
					<div className="overflow-x-auto">
						<ResponsiveTable className="w-full text-start text-sm">
							<thead className="text-xs text-muted-foreground bg-muted/30">
								<tr className="border-b border-border/70">
									<th className="py-3 ps-4 text-start font-medium">{t("payments.colTxn")}</th>
									<th className="py-3 text-start font-medium">{t("payments.colCustomer")}</th>
									<th className="py-3 text-start font-medium">{t("payments.colGross")}</th>
									<th className="py-3 text-start font-medium">{t("payments.colStatus")}</th>
									<th className="py-3 pe-4 text-end font-medium">{t("payments.colActions")}</th>
								</tr>
							</thead>
							<tbody className="divide-y divide-border/50">
								{loading ? (
									<tr>
										<td colSpan={5} className="py-12 text-center text-sm text-muted-foreground">{t("common.loading")}</td>
									</tr>
								) : payments.length === 0 ? (
									<tr>
										<td colSpan={5} className="py-12 text-center text-sm text-muted-foreground">{t("payments.empty")}</td>
									</tr>
								) : (
									payments.map((pay) => {
										const statusNorm = pay.status.toLowerCase();
										const amount = pay.amount != null ? pay.amount : 0;
										return (
											<tr key={pay.id} className="hover:bg-muted/40 transition-colors">
												<td className="py-3.5 ps-4">
													<div className="font-mono text-xs text-muted-foreground">{pay.id}</div>
													<div className="font-mono text-xs text-muted-foreground">{pay.orderId}</div>
												</td>
												<td className="py-3.5">
													<div className="font-medium text-foreground">{pay.customerEmail ? pay.customerEmail.split("@")[0] : "—"}</div>
													<div className="text-xs text-muted-foreground">{pay.customerEmail || "—"}</div>
												</td>
												<td className="py-3.5 font-medium tabular-nums">
													{amount.toLocaleString(locale, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}{pay.currency ? ` ${pay.currency}` : ""}
												</td>
												<td className="py-3.5">
													<StatusPill tone={statusTone(pay.status)}>{t(`status.${statusNorm}`) || pay.status}</StatusPill>
												</td>
												<td className="py-3.5 pe-4 text-end">
													<div className="inline-flex items-center justify-end gap-2">
														{isSuperadmin && statusNorm === "pending" && (
															<button type="button" onClick={() => handleMarkPaid(pay)} className="text-xs font-semibold hover:underline cursor-pointer">{t("payments.markPaid")}</button>
														)}
														{isSuperadmin && (statusNorm === "captured" || statusNorm === "succeeded" || statusNorm === "paid") && (
															<button type="button" onClick={() => handleIssueRefund(pay)} className="text-xs font-semibold text-destructive hover:underline cursor-pointer">{t("payments.issueRefund")}</button>
														)}
														{!isSuperadmin && pay.orderId && (
															<Link href={`/admin/orders/${pay.orderId}`} className="text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors">{t("payments.viewOrder")}</Link>
														)}
													</div>
												</td>
											</tr>
										);
									})
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
						shown={payments.length}
					/>
				</div>
			)}
		</div>
	);
}
