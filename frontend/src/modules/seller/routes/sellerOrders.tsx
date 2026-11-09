"use client";

import { useEffect, useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Link } from "@/shared/i18n/navigation";
import { sellerService } from "@/modules/seller/api/sellerService";
import type { SellerOrder } from "@/modules/seller/types";
import type { OffsetMeta } from "@/shared/api/types";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { formatDateTime } from "@/shared/lib/datetime";
import { formatMinor } from "./sellerDashboard";
import { useTranslations, useLocale } from "next-intl";
import { CustomSelect } from "@/shared/ui/customSelect";

const ORDER_STATUS_KEYS: Record<string, string> = {
	pending: "status.pending",
	paid: "status.paid",
	processing: "status.processing",
	shipped: "status.shipped",
	delivered: "status.delivered",
	cancelled: "status.cancelled",
	refunded: "status.refunded",
};

export function SellerOrders() {
	const t = useTranslations("seller");
	const locale = useLocale();

	const [orders, setOrders] = useState<SellerOrder[] | null>(null);
	const [meta, setMeta] = useState<OffsetMeta | null>(null);
	const [page, setPage] = useState(1);
	const [statusFilter, setStatusFilter] = useState("all");
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState<Error | null>(null);

	const loadOrders = async (nextPage = page, nextStatus = statusFilter) => {
		setLoading(true);
		setLoadError(null);
		try {
			const response = await sellerService.getOrders({
				page: nextPage,
				pageSize: 20,
				status: nextStatus === "all" ? undefined : nextStatus,
			});
			setOrders(response.data);
			const nextMeta = response.meta as OffsetMeta | undefined;
			if (nextMeta && typeof nextMeta.page === "number") setMeta(nextMeta);
			else setMeta({ page: nextPage, pageSize: 20, totalItems: response.data.length, totalPages: 1, hasPrevious: nextPage > 1, hasNext: false });
		} catch (error) {
			setLoadError(error instanceof Error ? error : new Error(String(error)));
			setOrders(null);
			setMeta(null);
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		void loadOrders(1, "all");
		// eslint-disable-next-line react-hooks/exhaustive-deps
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
				<div className="flex flex-col gap-3 border-b border-border p-4 sm:flex-row sm:items-center sm:justify-between">
					<div>
						<h2 className="text-sm font-semibold text-foreground">
							{t("orders.historyTitle")}
						</h2>
						<p className="mt-1 text-xs text-muted-foreground">
							{t("orders.noFulfillmentChanges")}
						</p>
					</div>
					<CustomSelect
						value={statusFilter}
						onChange={(value) => {
							setStatusFilter(value);
							setPage(1);
							void loadOrders(1, value);
						}}
						options={[
							{ value: "all", label: t("status.all") },
							...Object.keys(ORDER_STATUS_KEYS).map((status) => ({
							value: status,
							label: t(ORDER_STATUS_KEYS[status]),
							})),
						]}
						ariaLabel={t("orders.table.fulfillmentStatus")}
/>
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
									<th className="py-2.5 text-start font-medium">
										{t("orders.table.netEarnings")}
									</th>
									<th className="py-2.5 pe-4 text-end font-medium">
										{t("orders.table.actions")}
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
											{order.couponCode && (
												<span className="mt-0.5 block rounded-full bg-emerald-600/10 px-2 py-0.5 text-[11px] font-medium text-emerald-700">
													{t("orders.couponApplied", {
														code: order.couponCode,
														amount: ((order.discountAmount ?? 0) / 100).toFixed(2),
													})}
												</span>
											)}
										</td>
										<td className="py-3 text-end font-medium text-foreground tabular-nums">
											{formatMinor(order.earnings, order.currency)}
										</td>
										<td className="py-3 pe-4 text-end">
											<Link
												href={`/seller/orders/${order.id}`}
												className="text-xs font-medium text-foreground underline-offset-4 hover:underline"
											>
												{t("orders.detail.view")}
											</Link>
										</td>
									</tr>
								))}
							</tbody>
						</ResponsiveTable>
					</div>
				)}
				{meta && meta.totalPages > 1 ? (
					<div className="flex flex-wrap items-center justify-between gap-3 border-t border-border px-4 py-3 text-sm">
						<span className="text-xs text-muted-foreground">
							{t("orders.pagination", { page: meta.page, total: meta.totalPages, count: meta.totalItems })}
						</span>
						<div className="flex items-center gap-2">
							<button
								onClick={() => { const next = Math.max(1, page - 1); setPage(next); void loadOrders(next); }}
								disabled={page <= 1 || loading}
								className="inline-flex items-center gap-1 rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground disabled:opacity-50"
							>
								<ChevronLeft className="size-4 rtl:rotate-180" />
								{t("actions.previous")}
							</button>
							<button
								onClick={() => { const next = page + 1; setPage(next); void loadOrders(next); }}
								disabled={page >= meta.totalPages || loading}
								className="inline-flex items-center gap-1 rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground disabled:opacity-50"
							>
								{t("actions.next")}
								<ChevronRight className="size-4 rtl:rotate-180" />
							</button>
						</div>
					</div>
				) : null}
			</section>
		</div>
	);
}
