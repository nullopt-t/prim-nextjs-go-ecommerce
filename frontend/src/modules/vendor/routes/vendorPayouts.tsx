"use client";

import { useEffect, useState } from "react";
import { X } from "lucide-react";
import { vendorService } from "@/modules/vendor/api/vendorService";
import type { VendorDashboard, VendorPayout } from "@/modules/vendor/types";
import { formatMinor } from "./vendorDashboard";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { formatDateTime } from "@/shared/lib/datetime";
import { useTranslations, useLocale } from "next-intl";
import { useVendorContext } from "./vendorContext";

export function VendorPayouts() {
	const t = useTranslations("vendor");
	const locale = useLocale();
	const { vendor } = useVendorContext();

	const [dashboard, setDashboard] = useState<VendorDashboard | null>(null);
	const [payouts, setPayouts] = useState<VendorPayout[]>([]);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState<Error | null>(null);
	const [isModalOpen, setIsModalOpen] = useState(false);
	const [payoutAmount, setPayoutAmount] = useState("");
	const [submitting, setSubmitting] = useState(false);

	const loadPayouts = async () => {
		setLoading(true);
		setLoadError(null);
		try {
			const [dashboardResponse, payoutResponse] = await Promise.all([
				vendorService.getDashboard(),
				vendorService.getPayouts(),
			]);
			setDashboard(dashboardResponse.data);
			setPayouts(payoutResponse.data);
			setPayoutAmount((dashboardResponse.data.availableBalance / 100).toFixed(2));
		} catch (error) {
			setLoadError(error instanceof Error ? error : new Error(String(error)));
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		void loadPayouts();
	}, []);

	const statusLabel = (status: string) => {
		const map: Record<string, string> = {
			requested: "status.requested",
			approved: "status.approved",
			processing: "status.processing",
			paid: "status.paid",
			rejected: "status.rejected",
		};
		return map[status] ? t(map[status]) : status;
	};

	const handleRequestPayout = async (event: React.FormEvent) => {
		event.preventDefault();
		const amountMinor = Math.round(Number(payoutAmount) * 100);
		if (!Number.isFinite(amountMinor) || amountMinor <= 0 || (dashboard && amountMinor > dashboard.availableBalance)) {
			notify.error(t("toasts.invalidPayoutAmount"));
			return;
		}

		setSubmitting(true);
		try {
			await vendorService.requestPayout({ amount: amountMinor });
			notify.success(t("toasts.payoutSubmitted", { amount: (amountMinor / 100).toFixed(2) }));
			setIsModalOpen(false);
			await loadPayouts();
		} catch (error) {
			notify.error(apiErrorMessage(error, t("payouts.requestFailed")));
		} finally {
			setSubmitting(false);
		}
	};

	if (loading) {
		return (
			<div className="space-y-5">
				<PageHeader title={t("payouts.title")} />
				<div className="flex min-h-[40vh] items-center justify-center">
					<div className="size-6 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
				</div>
			</div>
		);
	}

	if (loadError || !dashboard) {
		return (
			<div className="space-y-5">
				<PageHeader title={t("payouts.title")} />
				<ErrorState
					title={t("payouts.loadTitle")}
					description={loadError?.message || t("payouts.loadDescription")}
					onRetry={() => void loadPayouts()}
				/>
			</div>
		);
	}

	const metrics = [
		{
			label: t("payouts.availableBalance"),
			value: formatMinor(dashboard.availableBalance),
			detail: t("payouts.availableBalanceNote"),
		},
		{
			label: t("payouts.pendingClearance"),
			value: formatMinor(dashboard.pendingEarnings),
			detail: t("payouts.pendingClearanceNote"),
		},
		{
			label: t("payouts.outstandingClaims"),
			value: formatMinor(dashboard.outstandingPayouts),
			detail: t("payouts.outstandingClaimsNote"),
		},
	];

	return (
		<div className="space-y-5">
			<PageHeader
				title={t("payouts.title")}
				actions={
					<button
						onClick={() => setIsModalOpen(true)}
						className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
					>
						{t("actions.requestPayout")}
					</button>
				}
			/>

			<div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
				{metrics.map((metric) => (
					<div
						key={metric.label}
						className="rounded-lg border border-border bg-background p-4"
					>
						<div className="text-xs font-medium text-muted-foreground">
							{metric.label}
						</div>
						<div className="mt-2 text-2xl font-semibold tracking-tight text-foreground tabular-nums">
							{metric.value}
						</div>
						<div className="mt-1 text-xs text-muted-foreground">{metric.detail}</div>
					</div>
				))}
			</div>

			<section className="rounded-lg border border-border bg-background">
				<div className="flex flex-col gap-2 border-b border-border p-4 sm:flex-row sm:items-center sm:justify-between">
					<div>
						<div className="text-sm font-medium text-foreground">
							{vendor?.storeName ?? t("payouts.storeProfile")}
						</div>
						<div className="text-xs text-muted-foreground">{vendor?.slug ?? "—"}</div>
					</div>
					<div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
						<span>
							{t("payouts.commissionFee")}{" "}
							<strong className="font-medium text-foreground">
								{vendor ? (vendor.commissionRateBps / 100).toFixed(2) : "0.00"}%
							</strong>
						</span>
						<span>
							{t("payouts.lifetimeEarnings")}{" "}
							<strong className="font-medium text-foreground">
								{formatMinor(dashboard.lifetimeEarnings)}
							</strong>
						</span>
						<StatusPill tone={statusTone(vendor?.status)}>
							{statusLabel(vendor?.status ?? "")}
						</StatusPill>
					</div>
				</div>

				<div className="border-b border-border px-4 py-3">
					<h2 className="text-sm font-semibold text-foreground">
						{t("payouts.historyTitle")}
					</h2>
				</div>

				{payouts.length === 0 ? (
					<p className="p-10 text-center text-sm text-muted-foreground">
						{t("payouts.empty")}
					</p>
				) : (
					<div className="overflow-x-auto">
						<ResponsiveTable className="w-full text-start text-sm md:min-w-max">
							<thead className="text-xs text-muted-foreground">
								<tr className="border-b border-border">
									<th className="py-2.5 ps-4 text-start font-medium">
										{t("payouts.table.payoutId")}
									</th>
									<th className="py-2.5 text-start font-medium">
										{t("payouts.table.amount")}
									</th>
									<th className="py-2.5 text-start font-medium">
										{t("payouts.table.requestedDate")}
									</th>
									<th className="py-2.5 text-start font-medium">
										{t("payouts.table.balanceSnapshot")}
									</th>
									<th className="py-2.5 text-start font-medium">
										{t("payouts.table.reference")}
									</th>
									<th className="py-2.5 pe-4 text-end font-medium">
										{t("payouts.table.status")}
									</th>
								</tr>
							</thead>
							<tbody className="divide-y divide-border">
								{payouts.map((payout) => (
									<tr key={payout.id} className="transition-colors hover:bg-secondary/30">
										<td className="max-w-36 truncate py-3 ps-4 font-mono text-xs text-foreground">
											{payout.id}
										</td>
										<td className="py-3 font-medium text-foreground tabular-nums">
											{formatMinor(payout.amount, payout.currency)}
										</td>
										<td className="py-3 text-xs text-muted-foreground">
											{formatDateTime(payout.createdAt, locale)}
										</td>
										<td className="py-3 text-muted-foreground tabular-nums">
											{formatMinor(payout.snapshotAvailableBalance, payout.currency)}
										</td>
										<td className="py-3 font-mono text-xs text-muted-foreground">
											{payout.referenceNumber ?? "—"}
										</td>
										<td className="py-3 pe-4 text-end">
											<StatusPill tone={statusTone(payout.status)}>
												{statusLabel(payout.status)}
											</StatusPill>
										</td>
									</tr>
								))}
							</tbody>
						</ResponsiveTable>
					</div>
				)}
			</section>

			{isModalOpen ? (
				<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
					<form
						onSubmit={handleRequestPayout}
						className="w-full max-w-md space-y-4 rounded-lg border border-border bg-background p-5 shadow-xl"
					>
						<div className="flex items-center justify-between border-b border-border pb-3">
							<h3 className="text-sm font-semibold text-foreground">
								{t("payouts.modalTitle")}
							</h3>
							<button
								type="button"
								onClick={() => setIsModalOpen(false)}
								className="text-muted-foreground hover:text-foreground"
								aria-label={t("actions.cancel")}
							>
								<X className="size-4" />
							</button>
						</div>
						<div className="space-y-1">
							<label className="text-sm font-medium text-foreground">
								{t("payouts.withdrawalAmount")}
							</label>
							<input
								type="number"
								step="0.01"
								min="0.01"
								max={dashboard.availableBalance / 100}
								required
								value={payoutAmount}
								onChange={(event) => setPayoutAmount(event.target.value)}
								className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-medium text-foreground"
							/>
							<p className="text-xs text-muted-foreground">
								{t("payouts.maximumWithdrawable")}{" "}
								<strong className="font-medium text-foreground">
									{formatMinor(dashboard.availableBalance)}
								</strong>
							</p>
						</div>
						<p className="rounded-md border border-border bg-secondary/30 p-3 text-xs text-muted-foreground">
							{t("payouts.requestNote")}
						</p>
						<div className="flex justify-end gap-2 border-t border-border pt-3">
							<button
								type="button"
								onClick={() => setIsModalOpen(false)}
								className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground hover:bg-secondary"
							>
								{t("actions.cancel")}
							</button>
							<button
								type="submit"
								disabled={submitting}
								className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground disabled:opacity-60"
							>
								{t("actions.confirmRequest")}
							</button>
						</div>
					</form>
				</div>
			) : null}
		</div>
	);
}
