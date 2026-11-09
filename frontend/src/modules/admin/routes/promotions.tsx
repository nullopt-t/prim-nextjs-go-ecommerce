"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter, Link } from "@/shared/i18n/navigation";
import { useTranslations } from "next-intl";
import { adminService } from "@/modules/admin/api/adminService";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";

interface StatRow {
	status: string;
	usageCount: number;
	expiresAt: string;
}

const CARD_CLASS =
	"rounded-xl border border-border/70 bg-card p-4 sm:p-5 hover:border-border hover:bg-muted/30 active:scale-[0.99] transition-all cursor-pointer select-none";

export function AdminPromotions() {
	const navigate = useRouter();
	const t = useTranslations("admin");
	const [statsRows, setStatsRows] = useState<StatRow[]>([]);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);

	const loadStats = useCallback(async () => {
		try {
			setLoading(true);
			setLoadError(false);
			const res = await adminService.getPromotions({ pageSize: 100 });
			setStatsRows(
				(res?.data ?? []).map((p) => ({
					status: (p.status ?? "").toLowerCase(),
					usageCount: p.usageCount ?? 0,
					expiresAt: p.expiresAt ?? "",
				}))
			);
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
			navigate.push("/admin/promotions/search");
		};
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [navigate]);

	const stats = useMemo(() => {
		const now = new Date();
		return {
			active: statsRows.filter((r) => r.status === "active").length,
			expired: statsRows.filter((r) => r.expiresAt && new Date(r.expiresAt) < now).length,
			usage: statsRows.reduce((acc, r) => acc + (r.usageCount || 0), 0),
		};
	}, [statsRows]);

	const kpiCards = [
		{ key: "active", label: t("promotions.cardActive"), count: stats.active, href: "/admin/promotions/search?status=active" },
		{ key: "expired", label: t("promotions.cardExpired"), count: stats.expired, href: "/admin/promotions/search?status=expired" },
		{ key: "usage", label: t("promotions.cardUsage"), count: stats.usage, href: "/admin/promotions/search" },
		{ key: "all", label: t("promotions.colCode"), count: statsRows.length, href: "/admin/promotions/search" },
	];

	const actionCards = [
		{ id: "search", title: t("promotions.actionSearchTitle"), description: t("promotions.actionSearchDesc"), href: "/admin/promotions/search" },
		{ id: "create", title: t("promotions.actionCreateTitle"), description: t("promotions.actionCreateDesc"), href: "/admin/promotions/search" },
		{ id: "expired", title: t("promotions.actionExpiredTitle"), description: t("promotions.actionExpiredDesc"), href: "/admin/promotions/search?status=expired" },
		{ id: "rules", title: t("promotions.actionRulesTitle"), description: t("promotions.actionRulesDesc"), href: "/admin/reports" },
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
				<PageHeader title={t("promotions.title")} description={t("promotions.subtitle")} />
				<ErrorState onRetry={loadStats} />
			</div>
		);
	}

	return (
		<div className="space-y-8">
			<PageHeader title={t("promotions.title")} description={t("promotions.subtitle")} />

			<section className="space-y-3">
				<h2 className="text-sm font-semibold text-foreground">{t("promotions.overviewTitle")}</h2>
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
				<h2 className="text-sm font-semibold text-foreground">{t("promotions.operationsTitle")}</h2>
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
				<h2 className="text-sm font-semibold text-foreground">{t("promotions.dataTitle")}</h2>
				<button type="button" onClick={handleExport} className={`block w-full text-start ${CARD_CLASS}`}>
					<div className="text-sm font-semibold text-foreground">{t("reports.export")}</div>
					<p className="mt-1 text-xs leading-relaxed text-muted-foreground">{t("orders.actionExportDesc")}</p>
				</button>
			</section>
		</div>
	);
}
