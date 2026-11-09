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
	stock: number;
}

const CARD_CLASS =
	"rounded-xl border border-border/70 bg-card p-4 sm:p-5 hover:border-border hover:bg-muted/30 active:scale-[0.99] transition-all cursor-pointer select-none";

export function AdminProducts() {
	const navigate = useRouter();
	const t = useTranslations("admin");
	const [statsRows, setStatsRows] = useState<StatRow[]>([]);
	const [loading, setLoading] = useState(true);
	const [exporting, setExporting] = useState(false);
	const [loadError, setLoadError] = useState(false);

	const loadStats = useCallback(async () => {
		try {
			setLoading(true);
			setLoadError(false);
			const res = await adminService.getProducts({ pageSize: 100 });
			setStatsRows(
				(res?.data ?? []).map((p) => ({
					status: (p.status || "draft").toLowerCase(),
					stock: typeof p.stockQuantity === "number" ? p.stockQuantity : p.stockCount ?? p.stock ?? 0,
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
			navigate.push("/admin/products/search");
		};
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [navigate]);

	const stats = useMemo(() => ({
		all: statsRows.length,
		published: statsRows.filter((r) => r.status === "published").length,
		draft: statsRows.filter((r) => r.status === "draft").length,
		outOfStock: statsRows.filter((r) => r.stock <= 0).length,
	}), [statsRows]);

	const kpiCards = [
		{ key: "all", label: t("products.cardAll"), count: (stats as Record<string, number>).all, href: "/admin/products/search" },
		{ key: "published", label: t("products.cardPublished"), count: (stats as Record<string, number>).published, href: "/admin/products/search?status=published" },
		{ key: "draft", label: t("products.cardDraft"), count: (stats as Record<string, number>).draft, href: "/admin/products/search?status=draft" },
		{ key: "out", label: t("products.cardOutOfStock"), count: (stats as Record<string, number>).outOfStock, href: "/admin/products/search?status=outofstock" },
	];

	const actionCards = [
		{ id: "search", title: t("products.actionSearchTitle"), description: t("products.actionSearchDesc"), href: "/admin/products/search" },
		{ id: "add", title: t("products.actionAddTitle"), description: t("products.actionAddDesc"), href: "/admin/products/new" },
		{ id: "drafts", title: t("products.actionDraftsTitle"), description: t("products.actionDraftsDesc"), href: "/admin/products/drafts" },
		{ id: "categories", title: t("products.actionCategoriesTitle"), description: t("products.actionCategoriesDesc"), href: "/admin/categories" },
		{ id: "inventory", title: t("products.actionInventoryTitle"), description: t("products.actionInventoryDesc"), href: "/admin/inventory" },
	];

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

	if (loadError) {
		return (
			<div className="space-y-6">
				<PageHeader title={t("products.title")} description={t("products.subtitle")} />
				<ErrorState onRetry={loadStats} />
			</div>
		);
	}

	return (
		<div className="space-y-8">
			<PageHeader title={t("products.title")} description={t("products.subtitle")} />

			<section className="space-y-3">
				<h2 className="text-sm font-semibold text-foreground">{t("products.overviewTitle")}</h2>
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
				<h2 className="text-sm font-semibold text-foreground">{t("products.managementTitle")}</h2>
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
				<h2 className="text-sm font-semibold text-foreground">{t("products.dataTitle")}</h2>
				<button
					type="button"
					onClick={handleExport}
					disabled={exporting}
					className={`block w-full text-start disabled:opacity-50 ${CARD_CLASS}`}
				>
					<div className="text-sm font-semibold text-foreground">{t("reports.export")}</div>
					<p className="mt-1 text-xs leading-relaxed text-muted-foreground">
						{exporting ? t("common.loading") : t("orders.actionExportDesc")}
					</p>
				</button>
			</section>
		</div>
	);
}
