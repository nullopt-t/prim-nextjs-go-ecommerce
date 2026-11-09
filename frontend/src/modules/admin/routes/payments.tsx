"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter, Link } from "@/shared/i18n/navigation";
import { useTranslations } from "next-intl";
import { adminService, type AdminPayment } from "@/modules/admin/api/adminService";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";

const CARD_CLASS =
	"rounded-xl border border-border/70 bg-card p-4 sm:p-5 hover:border-border hover:bg-muted/30 active:scale-[0.99] transition-all cursor-pointer select-none";

export function AdminPayments() {
	const navigate = useRouter();
	const t = useTranslations("admin");
	const [statsRows, setStatsRows] = useState<AdminPayment[]>([]);
	const [loading, setLoading] = useState(true);
	const [exporting, setExporting] = useState(false);
	const [loadError, setLoadError] = useState(false);

	const loadStats = useCallback(async () => {
		try {
			setLoading(true);
			setLoadError(false);
			const res = await adminService.getPayments({ pageSize: 100 });
			setStatsRows(res.data ?? []);
		} catch {
			setLoadError(true);
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		loadStats();
	}, [loadStats]);

	useEffect(() => {
		const onKeyDown = (e: KeyboardEvent) => {
			const target = e.target as HTMLElement | null;
			if (
				e.key !== "/" ||
				e.metaKey || e.ctrlKey || e.altKey ||
				target?.closest("input, textarea, select, [contenteditable]")
			) {
				return;
			}
			e.preventDefault();
			navigate.push("/admin/payments/search");
		};
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [navigate]);

	const stats = useMemo(() => {
		const totalVolumeCents = statsRows.reduce((acc, p) => acc + (p.amount || 0), 0);
		return {
			all: statsRows.length,
			succeeded: statsRows.filter((p) => p.status === "captured" || p.status.toLowerCase() === "succeeded" || p.status.toLowerCase() === "paid").length,
			pending: statsRows.filter((p) => p.status.toLowerCase() === "pending" || p.status.toLowerCase() === "authorized").length,
			refunded: statsRows.filter((p) => p.status.toLowerCase() === "refunded").length,
			grossVolume: `EGP ${(totalVolumeCents / 100).toFixed(2)}`,
		};
	}, [statsRows]);

	const kpiCards = [
		{ key: "all", label: t("payments.cardAll"), count: stats.all, href: "/admin/payments/search" },
		{ key: "succeeded", label: t("payments.cardSucceeded"), count: stats.succeeded, href: "/admin/payments/search?status=captured" },
		{ key: "pending", label: t("payments.cardPending"), count: stats.pending, href: "/admin/payments/search?status=pending" },
		{ key: "refunded", label: t("payments.cardRefunded"), count: stats.refunded, href: "/admin/payments/search?status=refunded" },
	];

	const actionCards = [
		{ id: "search", title: t("payments.actionSearchTitle"), description: t("payments.actionSearchDesc"), href: "/admin/payments/search" },
		{ id: "pending", title: t("payments.actionPendingTitle"), description: t("payments.actionPendingDesc"), href: "/admin/payments/search?status=pending" },
		{ id: "refunds", title: t("payments.actionRefundsTitle"), description: t("payments.actionRefundsDesc"), href: "/admin/payments/search?status=refunded" },
		{ id: "export", title: t("payments.actionExportTitle"), description: t("payments.actionExportDesc"), href: "/admin/reports" },
	];

	const handleExport = async () => {
		try {
			setExporting(true);
			await adminService.exportReport("payments", {});
			notify.success(t("reports.toast.exported"));
		} catch (err) {
			notify.error(apiErrorMessage(err, t("reports.toast.exportFailed")));
		} finally {
			setExporting(false);
		}
	};

	if (loadError) {
		return (
			<div className="space-y-6">
				<PageHeader title={t("payments.title")} description={t("payments.subtitle")} />
				<ErrorState onRetry={loadStats} />
			</div>
		);
	}

	return (
		<div className="space-y-8">
			<PageHeader title={t("payments.title")} description={t("payments.subtitle")} />

			<section className="space-y-3">
				<h2 className="text-sm font-semibold text-foreground">{t("payments.overviewTitle")}</h2>
				<div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
					{kpiCards.map((card) => (
						<Link key={card.key} href={card.href} className={`block text-start ${CARD_CLASS}`}>
							<div className="text-xs font-medium text-muted-foreground">{card.label}</div>
							<div className="mt-2 text-2xl font-semibold text-foreground tracking-tight tabular-nums leading-none">
								{loading ? "…" : card.count}
							</div>
							{card.key === "all" && (
								<div className="mt-1.5 text-xs text-muted-foreground tabular-nums">
									{t("payments.grossVolume")}: {stats.grossVolume}
								</div>
							)}
						</Link>
					))}
				</div>
			</section>

			<section className="space-y-3">
				<h2 className="text-sm font-semibold text-foreground">{t("payments.operationsTitle")}</h2>
				<div className="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
					{actionCards.map((card) => (
						<Link key={card.id} href={card.href} className={`block ${CARD_CLASS}`}>
							<div className="text-sm font-semibold text-foreground">{card.title}</div>
							<p className="mt-1 text-xs leading-relaxed text-muted-foreground">{card.description}</p>
						</Link>
					))}
				</div>
			</section>

			<section className="space-y-3">
				<h2 className="text-sm font-semibold text-foreground">{t("payments.dataTitle")}</h2>
				<button type="button" onClick={handleExport} disabled={exporting} className={`block w-full text-start disabled:opacity-50 ${CARD_CLASS}`}>
					<div className="text-sm font-semibold text-foreground">{t("reports.export")}</div>
					<p className="mt-1 text-xs leading-relaxed text-muted-foreground">
						{exporting ? t("common.loading") : t("payments.actionExportDesc")}
					</p>
				</button>
			</section>
		</div>
	);
}
