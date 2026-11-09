"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter, Link } from "@/shared/i18n/navigation";
import { useTranslations } from "next-intl";
import { adminService, type AdminSeller } from "@/modules/admin/api/adminService";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";

const CARD_CLASS =
	"rounded-xl border border-border/70 bg-card p-4 sm:p-5 hover:border-border hover:bg-muted/30 active:scale-[0.99] transition-all cursor-pointer select-none";

export function AdminSellers() {
	const navigate = useRouter();
	const t = useTranslations("admin");
	const [statsRows, setStatsRows] = useState<AdminSeller[]>([]);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);

	const loadStats = useCallback(async () => {
		try {
			setLoading(true);
			setLoadError(false);
			const res = await adminService.getSellers({ pageSize: 100 });
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
			navigate.push("/admin/sellers/search");
		};
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [navigate]);

	const stats = useMemo(() => ({
		all: statsRows.length,
		active: statsRows.filter((s) => s.status === "active").length,
		pending: statsRows.filter((s) => s.status === "pending").length,
		suspended: statsRows.filter((s) => s.status === "suspended").length,
	}), [statsRows]);

	const kpiCards = [
		{ key: "all", label: t("sellers.cardAll"), count: stats.all, href: "/admin/sellers/search" },
		{ key: "active", label: t("sellers.cardActive"), count: stats.active, href: "/admin/sellers/search?status=active" },
		{ key: "pending", label: t("sellers.cardPending"), count: stats.pending, href: "/admin/sellers/search?status=pending" },
		{ key: "suspended", label: t("sellers.cardSuspended"), count: stats.suspended, href: "/admin/sellers/search?status=suspended" },
	];

	const actionCards = [
		{ id: "search", title: t("sellers.actionSearchTitle"), description: t("sellers.actionSearchDesc"), href: "/admin/sellers/search" },
		{ id: "pending", title: t("sellers.actionPendingTitle"), description: t("sellers.actionPendingDesc"), href: "/admin/sellers/search?status=pending" },
		{ id: "active", title: t("sellers.actionActiveTitle"), description: t("sellers.actionActiveDesc"), href: "/admin/sellers/search?status=active" },
		{ id: "payouts", title: t("sellers.actionPayoutsTitle"), description: t("sellers.actionPayoutsDesc"), href: "/admin/reports" },
	];

	const handleExport = async () => {
		try {
			await adminService.exportReport("orders", {});
			notify.success(t("reports.toast.exported"));
		} catch (err) {
			notify.error(apiErrorMessage(err, t("reports.toast.exportFailed")));
		}
	};

	if (loadError) {
		return (
			<div className="space-y-6">
				<PageHeader title={t("sellers.title")} description={t("sellers.subtitle")} />
				<ErrorState onRetry={loadStats} />
			</div>
		);
	}

	return (
		<div className="space-y-8">
			<PageHeader title={t("sellers.title")} description={t("sellers.subtitle")} />

			<section className="space-y-3">
				<h2 className="text-sm font-semibold text-foreground">{t("sellers.overviewTitle")}</h2>
				<div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
					{kpiCards.map((card) => (
						<Link key={card.key} href={card.href} className={`block text-start ${CARD_CLASS}`}>
							<div className="text-xs font-medium text-muted-foreground">{card.label}</div>
							<div className="mt-2 text-2xl font-semibold text-foreground tracking-tight tabular-nums leading-none">
								{loading ? "…" : card.count}
							</div>
						</Link>
					))}
				</div>
			</section>

			<section className="space-y-3">
				<h2 className="text-sm font-semibold text-foreground">{t("sellers.operationsTitle")}</h2>
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
				<h2 className="text-sm font-semibold text-foreground">{t("sellers.dataTitle")}</h2>
				<button type="button" onClick={handleExport} className={`block w-full text-start ${CARD_CLASS}`}>
					<div className="text-sm font-semibold text-foreground">{t("reports.export")}</div>
					<p className="mt-1 text-xs leading-relaxed text-muted-foreground">{t("orders.actionExportDesc")}</p>
				</button>
			</section>
		</div>
	);
}
