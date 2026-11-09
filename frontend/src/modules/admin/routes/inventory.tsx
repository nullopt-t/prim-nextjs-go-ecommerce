"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter, Link } from "@/shared/i18n/navigation";
import { useTranslations } from "next-intl";
import { adminService } from "@/modules/admin/api/adminService";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";

interface StatRow {
	onHand: number;
	available: number;
}

const CARD_CLASS =
	"rounded-xl border border-border/70 bg-card p-4 sm:p-5 hover:border-border hover:bg-muted/30 active:scale-[0.99] transition-all cursor-pointer select-none";

export function AdminInventory() {
	const navigate = useRouter();
	const t = useTranslations("admin");
	const [statsRows, setStatsRows] = useState<StatRow[]>([]);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);

	const loadStats = useCallback(async () => {
		try {
			setLoading(true);
			setLoadError(false);
			const res = await adminService.getInventory({ pageSize: 100 });
			const items = (res?.data ?? []) as unknown as {
				onHandQuantity?: number;
				availableQuantity?: number;
			}[];
			setStatsRows(
				items.map((it) => ({
					onHand: it.onHandQuantity ?? 0,
					available: it.availableQuantity ?? 0,
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
			navigate.push("/admin/inventory/search");
		};
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [navigate]);

	const stats = useMemo(() => ({
		total: statsRows.length,
		inStock: statsRows.filter((r) => r.onHand > 10).length,
		lowStock: statsRows.filter((r) => r.onHand > 0 && r.onHand <= 10).length,
		outOfStock: statsRows.filter((r) => r.onHand <= 0).length,
	}), [statsRows]);

	const kpiCards = [
		{ key: "in", label: t("inventory.cardInStock"), count: stats.inStock, href: "/admin/inventory/search?stock=in" },
		{ key: "low", label: t("inventory.cardLowStock"), count: stats.lowStock, href: "/admin/inventory/search?stock=low" },
		{ key: "out", label: t("inventory.cardOutOfStock"), count: stats.outOfStock, href: "/admin/inventory/search?stock=out" },
		{ key: "total", label: t("inventory.catalogItems", { count: stats.total }), count: stats.total, href: "/admin/inventory/search" },
	];

	const actionCards = [
		{ id: "search", title: t("inventory.actionSearchTitle"), description: t("inventory.actionSearchDesc"), href: "/admin/inventory/search" },
		{ id: "adjust", title: t("inventory.actionAdjustTitle"), description: t("inventory.actionAdjustDesc"), href: "/admin/inventory/search" },
		{ id: "audit", title: t("inventory.actionAuditTitle"), description: t("inventory.actionAuditDesc"), href: "/admin/inventory/search" },
		{ id: "export", title: t("inventory.actionExportTitle"), description: t("inventory.actionExportDesc"), href: "/admin/reports" },
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
				<PageHeader title={t("inventory.title")} description={t("inventory.subtitle")} />
				<ErrorState onRetry={loadStats} />
			</div>
		);
	}

	return (
		<div className="space-y-8">
			<PageHeader title={t("inventory.title")} description={t("inventory.subtitle")} />

			<section className="space-y-3">
				<h2 className="text-sm font-semibold text-foreground">{t("inventory.overviewTitle")}</h2>
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
				<h2 className="text-sm font-semibold text-foreground">{t("inventory.operationsTitle")}</h2>
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
				<h2 className="text-sm font-semibold text-foreground">{t("inventory.dataTitle")}</h2>
				<button type="button" onClick={handleExport} className={`block w-full text-start ${CARD_CLASS}`}>
					<div className="text-sm font-semibold text-foreground">{t("inventory.actionExportTitle")}</div>
					<p className="mt-1 text-xs leading-relaxed text-muted-foreground">{t("inventory.actionExportDesc")}</p>
				</button>
			</section>
		</div>
	);
}
