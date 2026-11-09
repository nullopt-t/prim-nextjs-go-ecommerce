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
	rawTotal: number;
}

const CARD_CLASS =
	"rounded-xl border border-border/70 bg-card p-4 sm:p-5 hover:border-border hover:bg-muted/30 active:scale-[0.99] transition-all cursor-pointer select-none";

export function AdminOrders() {
	const navigate = useRouter();
	const t = useTranslations("admin");
	const [statsOrders, setStatsOrders] = useState<StatRow[]>([]);
	const [loading, setLoading] = useState(true);
	const [exporting, setExporting] = useState(false);
	const [loadError, setLoadError] = useState(false);

	// KPI counts come from a single unfiltered snapshot (main page shows no list).
	const loadStats = useCallback(async () => {
		try {
			setLoading(true);
			setLoadError(false);
			const res = await adminService.getOrders({ pageSize: 100 });
			setStatsOrders(
				(res?.data ?? []).map((so) => ({
					status: so.status ? so.status.toLowerCase() : "pending",
					rawTotal: so.totalAmount || 0,
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

	// On-spot refresh when a new order event is received in the admin console
	useEffect(() => {
		const handleNewOrder = () => loadStats();
		window.addEventListener("admin-order-placed", handleNewOrder);
		return () => window.removeEventListener("admin-order-placed", handleNewOrder);
	}, [loadStats]);

	// "/" shortcut opens the dedicated search page (ignored while typing).
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
			navigate.push("/admin/orders/search");
		};
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [navigate]);

	const handleExport = async () => {
		try {
			setExporting(true);
			await adminService.exportReport("orders", {});
			notify.success(t("reports.toast.exported"));
		} catch (err) {
			notify.error(apiErrorMessage(err, t("reports.toast.exportFailed")));
		} finally {
			setExporting(false);
		}
	};

	const stats = useMemo(() => {
		const totalVolumeCents = statsOrders.reduce((acc, o) => acc + o.rawTotal, 0);
		return {
			all: statsOrders.length,
			processing: statsOrders.filter((o) => o.status === "processing").length,
			pending: statsOrders.filter((o) => o.status === "pending").length,
			delivered: statsOrders.filter((o) => o.status === "delivered").length,
			grossVolume: `EGP ${(totalVolumeCents / 100).toFixed(2)}`,
		};
	}, [statsOrders]);

	const kpiCards = [
		{ status: "all", label: t("orders.cardAll"), count: stats.all },
		{ status: "processing", label: t("orders.cardProcessing"), count: stats.processing },
		{ status: "pending", label: t("orders.cardPending"), count: stats.pending },
		{ status: "delivered", label: t("orders.cardDelivered"), count: stats.delivered },
	];

	const actionCards = [
		{
			id: "search",
			title: t("orders.actionSearchTitle"),
			description: t("orders.actionSearchDesc"),
			href: "/admin/orders/search",
		},
		{
			id: "fulfill",
			title: t("orders.actionFulfillTitle"),
			description: t("orders.actionFulfillDesc"),
			href: "/admin/orders/search?status=processing",
		},
		{
			id: "review",
			title: t("orders.actionPendingTitle"),
			description: t("orders.actionPendingDesc"),
			href: "/admin/orders/search?status=pending",
		},
		{
			id: "returns",
			title: t("orders.actionReturnsTitle"),
			description: t("orders.actionReturnsDesc"),
			href: "/admin/orders/search?status=cancelled",
		},
	];

	if (loadError) {
		return (
			<div className="space-y-6">
				<PageHeader title={t("orders.title")} description={t("orders.subtitle")} />
				<ErrorState
					title={t("orders.errorTitle")}
					description={t("orders.errorMessage")}
					onRetry={loadStats}
				/>
			</div>
		);
	}

	return (
		<div className="space-y-8">
			<PageHeader title={t("orders.title")} description={t("orders.subtitle")} />

			{/* Section 1: Overview & Metrics */}
			<section className="space-y-3">
				<h2 className="text-sm font-semibold text-foreground">{t("orders.overviewTitle")}</h2>
				<div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
					{kpiCards.map((card) => (
						<Link
							key={card.status}
							href={`/admin/orders/search?status=${card.status}`}
							className={`block text-start ${CARD_CLASS}`}
						>
							<div className="text-xs font-medium text-muted-foreground">{card.label}</div>
							<div className="mt-2 text-2xl font-semibold text-foreground tracking-tight tabular-nums leading-none">
								{loading ? "…" : card.count}
							</div>
							{card.status === "all" && (
								<div className="mt-1.5 text-xs text-muted-foreground tabular-nums">
									{stats.grossVolume}
								</div>
							)}
						</Link>
					))}
				</div>
			</section>

			{/* Section 2: Order Management & Search */}
			<section className="space-y-3">
				<h2 className="text-sm font-semibold text-foreground">{t("orders.managementTitle")}</h2>
				<div className="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
					{actionCards.map((card) => (
						<Link key={card.id} href={card.href} className={`block ${CARD_CLASS}`}>
							<div className="text-sm font-semibold text-foreground">{card.title}</div>
							<p className="mt-1 text-xs leading-relaxed text-muted-foreground">{card.description}</p>
						</Link>
					))}
				</div>
			</section>

			{/* Section 3: Data & Export */}
			<section className="space-y-3">
				<h2 className="text-sm font-semibold text-foreground">{t("orders.dataTitle")}</h2>
				<button
					type="button"
					onClick={handleExport}
					disabled={exporting}
					className={`block w-full text-start disabled:opacity-50 ${CARD_CLASS}`}
				>
					<div className="text-sm font-semibold text-foreground">{t("orders.actionExportTitle")}</div>
					<p className="mt-1 text-xs leading-relaxed text-muted-foreground">
						{exporting ? t("common.loading") : t("orders.actionExportDesc")}
					</p>
				</button>
			</section>
		</div>
	);
}
