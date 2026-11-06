"use client";

import { useEffect, useState } from "react";
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

export function VendorOrders() {
	const t = useTranslations("vendor");
	const locale = useLocale();

	const [orders, setOrders] = useState<VendorOrder[] | null>(null);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState<Error | null>(null);

	const loadOrders = async () => {
		setLoading(true);
		setLoadError(null);
		try {
			const response = await vendorService.getOrders({ pageSize: 100 });
			setOrders(response.data);
		} catch (error) {
			setLoadError(error instanceof Error ? error : new Error(String(error)));
			setOrders(null);
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		void loadOrders();
	}, []);

	const statusLabel = (status: string) => {
		const key = ORDER_STATUS_KEYS[status];
		return key ? t(key) : status;
	};

	if (loading) {
		return (
			<div className="space-y-5">
				<PageHeader title={t("orders.title")} />
				<div className="flex min-h-[40vh] items-center justify-center">
					<div className="size-6 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
				</div>
			</div>
		);
	}

	if (loadError || !orders) {
		return (
			<div className="space-y-5">
				<PageHeader title={t("orders.title")} />
				<ErrorState
					title={t("orders.loadTitle")}
					description={loadError?.message || t("orders.loadDescription")}
					onRetry={() => void loadOrders()}
				/>
			</div>
		);
	}

	return (
		<div className="space-y-5">
			<PageHeader title={t("orders.title")} />

			<section className="overflow-hidden rounded-lg border border-border bg-background">
				<div className="flex flex-col gap-1 border-b border-border p-4 sm:flex-row sm:items-center sm:justify-between">
					<h2 className="text-sm font-semibold text-foreground">
						{t("orders.historyTitle")}
					</h2>
					<p className="text-xs text-muted-foreground">
						{t("orders.noFulfillmentChanges")}
					</p>
				</div>

				{orders.length === 0 ? (
					<p className="p-10 text-center text-sm text-muted-foreground">
						{t("orders.empty")}
					</p>
				) : (
					<div className="overflow-x-auto">
						<ResponsiveTable className="w-full text-start text-sm md:min-w-max">
							<thead className="text-xs text-muted-foreground">
								<tr className="border-b border-border">
									<th className="py-2.5 ps-4 text-start font-medium">
										{t("orders.table.orderRef")}
									</th>
									<th className="py-2.5 text-start font-medium">
										{t("orders.table.date")}
									</th>
									<th className="py-2.5 text-start font-medium">
										{t("orders.table.fulfillmentStatus")}
									</th>
									<th className="py-2.5 text-start font-medium">
										{t("orders.table.item")}
									</th>
									<th className="py-2.5 text-start font-medium">
										{t("orders.table.grossRevenue")}
									</th>
									<th className="py-2.5 pe-4 text-end font-medium">
										{t("orders.table.netEarnings")}
									</th>
								</tr>
							</thead>
							<tbody className="divide-y divide-border">
								{orders.map((order) => (
									<tr key={order.id} className="transition-colors hover:bg-secondary/30">
										<td className="max-w-36 truncate py-3 ps-4 font-mono text-xs text-foreground">
											{order.id}
										</td>
										<td className="py-3 text-xs text-muted-foreground">
											{formatDateTime(order.createdAt, locale)}
										</td>
										<td className="py-3">
											<StatusPill tone={statusTone(order.status)}>
												{statusLabel(order.status)}
											</StatusPill>
										</td>
										<td className="py-3 text-muted-foreground">
											<div className="flex flex-col gap-0.5">
												{order.items.map((item) => (
													<span key={item.id} className="text-xs">
														{item.title || item.variantId} × {item.quantity}
													</span>
												))}
											</div>
										</td>
										<td className="py-3 font-medium text-foreground tabular-nums">
											{formatMinor(order.subtotal, order.currency)}
										</td>
										<td className="py-3 pe-4 text-end font-medium text-foreground tabular-nums">
											{formatMinor(order.earnings, order.currency)}
										</td>
									</tr>
								))}
							</tbody>
						</ResponsiveTable>
					</div>
				)}
			</section>
		</div>
	);
}
