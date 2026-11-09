"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter, Link } from "@/shared/i18n/navigation";
import { useTranslations } from "next-intl";
import { adminService, type AdminUser } from "@/modules/admin/api/adminService";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";

const CARD_CLASS =
	"rounded-xl border border-border/70 bg-card p-4 sm:p-5 hover:border-border hover:bg-muted/30 active:scale-[0.99] transition-all cursor-pointer select-none";

export function AdminCustomers() {
	const navigate = useRouter();
	const t = useTranslations("admin");
	const [statsUsers, setStatsUsers] = useState<AdminUser[]>([]);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);

	const loadStats = useCallback(async () => {
		try {
			setLoading(true);
			setLoadError(false);
			const res = await adminService.getUsers({ pageSize: 100 });
			setStatsUsers(res.data ?? []);
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
			navigate.push("/admin/customers/search");
		};
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [navigate]);

	const stats = useMemo(() => ({
		total: statsUsers.length,
		active: statsUsers.filter((u) => u.status === "active").length,
		suspended: statsUsers.filter((u) => u.status === "suspended").length,
		admins: statsUsers.filter((u) => u.role === "admin" || u.role === "superadmin").length,
	}), [statsUsers]);

	const kpiCards = [
		{ key: "total", label: t("customers.cardTotal"), count: stats.total, href: "/admin/customers/search" },
		{ key: "active", label: t("customers.cardActive"), count: stats.active, href: "/admin/customers/search?status=active" },
		{ key: "suspended", label: t("customers.cardSuspended"), count: stats.suspended, href: "/admin/customers/search?status=suspended" },
		{ key: "admins", label: t("customers.cardAdmins"), count: stats.admins, href: "/admin/customers/search?role=admin" },
	];

	const actionCards = [
		{ id: "search", title: t("customers.actionSearchTitle"), description: t("customers.actionSearchDesc"), href: "/admin/customers/search" },
		{ id: "customers", title: t("customers.actionCustomersTitle"), description: t("customers.actionCustomersDesc"), href: "/admin/customers/search?role=customer" },
		{ id: "sellers", title: t("customers.actionSellersTitle"), description: t("customers.actionSellersDesc"), href: "/admin/customers/search?role=seller" },
		{ id: "admins", title: t("customers.actionAdminsTitle"), description: t("customers.actionAdminsDesc"), href: "/admin/customers/search?role=admin" },
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
				<PageHeader title={t("customers.title")} description={t("customers.subtitle")} />
				<ErrorState onRetry={loadStats} />
			</div>
		);
	}

	return (
		<div className="space-y-8">
			<PageHeader title={t("customers.title")} description={t("customers.subtitle")} />

			<section className="space-y-3">
				<h2 className="text-sm font-semibold text-foreground">{t("customers.overviewTitle")}</h2>
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
				<h2 className="text-sm font-semibold text-foreground">{t("customers.managementTitle")}</h2>
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
				<h2 className="text-sm font-semibold text-foreground">{t("customers.dataTitle")}</h2>
				<button type="button" onClick={handleExport} className={`block w-full text-start ${CARD_CLASS}`}>
					<div className="text-sm font-semibold text-foreground">{t("reports.export")}</div>
					<p className="mt-1 text-xs leading-relaxed text-muted-foreground">{t("orders.actionExportDesc")}</p>
				</button>
			</section>
		</div>
	);
}
