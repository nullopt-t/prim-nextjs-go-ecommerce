"use client";

import { useEffect, useState } from "react";
import { ArrowLeft } from "lucide-react";
import { useParams } from "next/navigation";
import { Link } from "@/shared/i18n/navigation";
import { vendorService } from "@/modules/vendor/api/vendorService";
import type { VendorOrder } from "@/modules/vendor/types";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { formatDateTime } from "@/shared/lib/datetime";
import { formatMinor } from "./vendorDashboard";
import { useTranslations, useLocale } from "next-intl";

const ORDER_STATUS_KEYS: Record<string, string> = {
	pending: "status.pending",
	paid: "status.paid",
	processing: "status.processing",
	shipped: "status.shipped",
	delivered: "status.delivered",
	cancelled: "status.cancelled",
	refunded: "status.refunded",
};

export function VendorOrderDetail() {
	const t = useTranslations("vendor");
	const locale = useLocale();
	const params = useParams();
	const orderId = String(params?.id ?? "");

	const [order, setOrder] = useState<VendorOrder | null>(null);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState<Error | null>(null);

	const loadOrder = async () => {
		if (!orderId) return;
		setLoading(true);
		setLoadError(null);
		try {
			const response = await vendorService.getOrder(orderId);
			setOrder(response.data);
		} catch (error) {
			setLoadError(error instanceof Error ? error : new Error(String(error)));
			setOrder(null);
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		void loadOrder();
	}, [orderId]);

	if (loading) {
		return (
			<div className="space-y-5">
				<PageHeader title={t("orders.detail.title")} />
				<div className="flex min-h-[40vh] items-center justify-center">
					<div className="size-6 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
				</div>
			</div>
		);
	}

	if (loadError || !order) {
		return (
			<div className="space-y-5">
				<PageHeader title={t("orders.detail.title")} />
				<ErrorState
					title={t("orders.detail.loadTitle")}
					description={loadError?.message || t("orders.detail.loadDescription")}
					onRetry={() => void loadOrder()}
				/>
			</div>
		);
	}

	const statusLabel = ORDER_STATUS_KEYS[order.status] ? t(ORDER_STATUS_KEYS[order.status]) : order.status;
	const commissionTotal = order.items.reduce((sum, item) => sum + item.commissionAmount, 0);
	const itemCount = order.items.reduce((sum, item) => sum + item.quantity, 0);

	return (
		<div className="space-y-5">
			<PageHeader
				title={t("orders.detail.title")}
				description={t("orders.detail.subtitle", { id: order.id })}
				actions={
					<Link
						href="/vendor/orders"
						className="inline-flex items-center gap-1.5 rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground hover:bg-secondary"
					>
						<ArrowLeft className="size-4 rtl:rotate-180" />
						{t("orders.detail.back")}
					</Link>
				}
			/>

			<div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
				<div className="rounded-lg border border-border bg-background p-4">
					<div className="text-xs font-medium text-muted-foreground">{t("orders.detail.status")}</div>
					<div className="mt-2">
						<StatusPill tone={statusTone(order.status)}>{statusLabel}</StatusPill>
					</div>
					<div className="mt-2 text-xs text-muted-foreground">{formatDateTime(order.createdAt, locale)}</div>
				</div>
				<div className="rounded-lg border border-border bg-background p-4">
					<div className="text-xs font-medium text-muted-foreground">{t("orders.detail.gross")}</div>
					<div className="mt-2 text-2xl font-semibold tabular-nums text-foreground">{formatMinor(order.subtotal, order.currency)}</div>
				</div>
				<div className="rounded-lg border border-border bg-background p-4">
					<div className="text-xs font-medium text-muted-foreground">{t("orders.detail.earnings")}</div>
					<div className="mt-2 text-2xl font-semibold tabular-nums text-foreground">{formatMinor(order.earnings, order.currency)}</div>
				</div>
				<div className="rounded-lg border border-border bg-background p-4">
					<div className="text-xs font-medium text-muted-foreground">{t("orders.detail.commission")}</div>
					<div className="mt-2 text-2xl font-semibold tabular-nums text-foreground">{formatMinor(commissionTotal, order.currency)}</div>
					<div className="mt-1 text-xs text-muted-foreground">{t("orders.detail.itemCount", { count: itemCount })}</div>
				</div>
			</div>

			<section className="overflow-hidden rounded-lg border border-border bg-background">
				<div className="border-b border-border p-4">
					<h2 className="text-sm font-semibold text-foreground">{t("orders.detail.linesTitle")}</h2>
					<p className="mt-1 text-xs text-muted-foreground">{t("orders.detail.linesNote")}</p>
				</div>
				<div className="overflow-x-auto">
					<ResponsiveTable className="w-full text-start text-sm md:min-w-max">
						<thead className="text-xs text-muted-foreground">
							<tr className="border-b border-border">
								<th className="py-2.5 ps-4 text-start font-medium">{t("orders.detail.table.item")}</th>
								<th className="py-2.5 text-start font-medium">{t("orders.detail.table.qty")}</th>
								<th className="py-2.5 text-start font-medium">{t("orders.detail.table.unitPrice")}</th>
								<th className="py-2.5 text-start font-medium">{t("orders.detail.table.lineTotal")}</th>
								<th className="py-2.5 text-start font-medium">{t("orders.detail.table.commission")}</th>
								<th className="py-2.5 pe-4 text-end font-medium">{t("orders.detail.table.earning")}</th>
							</tr>
						</thead>
						<tbody className="divide-y divide-border">
							{order.items.map((item) => (
								<tr key={item.id} className="transition-colors hover:bg-secondary/30">
									<td className="py-3 ps-4">
										<div className="font-medium text-foreground">{item.title || item.variantId}</div>
										<div className="font-mono text-xs text-muted-foreground">{item.variantId}</div>
									</td>
									<td className="py-3 tabular-nums text-foreground">{item.quantity}</td>
									<td className="py-3 tabular-nums text-muted-foreground">{formatMinor(item.priceAtPurchase, order.currency)}</td>
									<td className="py-3 font-medium tabular-nums text-foreground">{formatMinor(item.lineTotal, order.currency)}</td>
									<td className="py-3 tabular-nums text-muted-foreground">
										{formatMinor(item.commissionAmount, order.currency)}
										<span className="ms-1 text-xs">({(item.commissionRateBps / 100).toFixed(2)}%)</span>
									</td>
									<td className="py-3 pe-4 text-end font-medium tabular-nums text-foreground">{formatMinor(item.vendorEarning, order.currency)}</td>
								</tr>
							))}
						</tbody>
					</ResponsiveTable>
				</div>
			</section>
		</div>
	);
}
