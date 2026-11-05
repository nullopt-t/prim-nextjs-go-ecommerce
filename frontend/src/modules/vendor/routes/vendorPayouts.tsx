"use client";

import { useEffect, useState } from "react";
import {
  DollarSign,
  Landmark,
  CheckCircle2,
  Clock,
  X,
  Loader2,
} from "lucide-react";
import { vendorService } from "@/modules/vendor/api/vendorService";
import type { VendorDashboard, VendorPayout } from "@/modules/vendor/types";
import { formatMinor } from "./vendorDashboard";
import { ErrorState } from "@/shared/ui/errorState";
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
      <div className="flex min-h-[50vh] items-center justify-center">
        <Loader2 className="size-8 animate-spin text-accent-brand" />
      </div>
    );
  }

  if (loadError || !dashboard) {
    return (
      <ErrorState
        title={t("payouts.loadTitle")}
        description={loadError?.message || t("payouts.loadDescription")}
        onRetry={() => void loadPayouts()}
      />
    );
  }

  return (
    <div className="space-y-6 w-full">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("payouts.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t("payouts.subtitle")}</p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-2 px-5 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0"
        >
          <DollarSign className="size-4" />
          <span>{t("actions.requestPayout")}</span>
        </button>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-2">
          <div className="flex items-center justify-between text-muted-foreground text-xs font-semibold">
            <span>{t("payouts.availableBalance")}</span>
            <CheckCircle2 className="size-4 text-emerald-500" />
          </div>
          <div className="text-3xl font-bold text-foreground">{formatMinor(dashboard.availableBalance)}</div>
          <p className="text-[11px] text-muted-foreground">{t("payouts.availableBalanceNote")}</p>
        </div>

        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-2">
          <div className="flex items-center justify-between text-muted-foreground text-xs font-semibold">
            <span>{t("payouts.pendingClearance")}</span>
            <Clock className="size-4 text-amber-500" />
          </div>
          <div className="text-3xl font-bold text-foreground">{formatMinor(dashboard.pendingEarnings)}</div>
          <p className="text-[11px] text-muted-foreground">{t("payouts.pendingClearanceNote")}</p>
        </div>

        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-2">
          <div className="flex items-center justify-between text-muted-foreground text-xs font-semibold">
            <span>{t("payouts.outstandingClaims")}</span>
            <Landmark className="size-4 text-accent-brand" />
          </div>
          <div className="text-3xl font-bold text-foreground">{formatMinor(dashboard.outstandingPayouts)}</div>
          <p className="text-[11px] text-muted-foreground">{t("payouts.outstandingClaimsNote")}</p>
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-3">
        <div className="flex items-center justify-between border-b border-border pb-3">
          <div>
            <div className="font-bold text-sm text-foreground">{vendor?.storeName ?? t("payouts.storeProfile")}</div>
            <div className="text-xs text-muted-foreground">{vendor?.slug ?? "—"}</div>
          </div>
          <span className="px-3 py-1 bg-secondary text-foreground border border-border text-xs font-bold rounded-full">
            {statusLabel(vendor?.status ?? "")}
          </span>
        </div>
        <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground">
          <span>
            {t("payouts.commissionFee")}{" "}
            <strong>{vendor ? (vendor.commissionRateBps / 100).toFixed(2) : "0.00"}%</strong>
          </span>
          <span>
            {t("payouts.lifetimeEarnings")}{" "}
            <strong>{formatMinor(dashboard.lifetimeEarnings)}</strong>
          </span>
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs space-y-4">
        <div className="p-6 border-b border-border">
          <h3 className="font-bold text-base text-foreground">{t("payouts.historyTitle")}</h3>
          <p className="text-xs text-muted-foreground mt-0.5">{t("payouts.historySubtitle")}</p>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full min-w-max text-left text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 pl-6">{t("payouts.table.payoutId")}</th>
                <th className="py-3.5">{t("payouts.table.amount")}</th>
                <th className="py-3.5">{t("payouts.table.requestedDate")}</th>
                <th className="py-3.5">{t("payouts.table.balanceSnapshot")}</th>
                <th className="py-3.5">{t("payouts.table.reference")}</th>
                <th className="py-3.5 pr-6 text-right">{t("payouts.table.status")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {payouts.map((payout) => (
                <tr key={payout.id} className="hover:bg-secondary/20 transition-colors">
                  <td className="py-3.5 pl-6 font-mono font-bold text-foreground">{payout.id}</td>
                  <td className="py-3.5 font-bold text-foreground text-sm">
                    {formatMinor(payout.amount, payout.currency)}
                  </td>
                  <td className="py-3.5 text-muted-foreground">{formatDateTime(payout.createdAt, locale)}</td>
                  <td className="py-3.5 text-muted-foreground">
                    {formatMinor(payout.snapshotAvailableBalance, payout.currency)}
                  </td>
                  <td className="py-3.5 font-mono text-muted-foreground">{payout.referenceNumber ?? "—"}</td>
                  <td className="py-3.5 pr-6 text-right">
                    <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-secondary text-foreground border border-border">
                      {statusLabel(payout.status)}
                    </span>
                  </td>
                </tr>
              ))}
              {payouts.length === 0 && (
                <tr>
                  <td colSpan={6} className="py-10 text-center text-muted-foreground">
                    {t("payouts.empty")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <form onSubmit={handleRequestPayout} className="bg-background border border-border rounded-2xl max-w-md w-full p-6 shadow-xl space-y-4">
            <div className="flex justify-between items-center border-b border-border pb-3">
              <h3 className="font-bold text-base text-foreground">{t("payouts.modalTitle")}</h3>
              <button type="button" onClick={() => setIsModalOpen(false)} className="text-muted-foreground">
                <X className="size-5" />
              </button>
            </div>
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">{t("payouts.withdrawalAmount")}</label>
              <input
                type="number"
                step="0.01"
                min="0.01"
                max={dashboard.availableBalance / 100}
                required
                value={payoutAmount}
                onChange={(event) => setPayoutAmount(event.target.value)}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-base font-bold text-foreground"
              />
              <div className="text-[11px] text-muted-foreground mt-1">
                {t("payouts.maximumWithdrawable")} <strong>{formatMinor(dashboard.availableBalance)}</strong>
              </div>
            </div>
            <div className="p-3 bg-secondary/30 rounded-xl border border-border text-xs text-muted-foreground">
              {t("payouts.requestNote")}
            </div>
            <div className="flex justify-end gap-3 pt-3 border-t border-border">
              <button type="button" onClick={() => setIsModalOpen(false)} className="px-4 py-2 text-xs border border-border rounded-xl">
                {t("actions.cancel")}
              </button>
              <button type="submit" disabled={submitting} className="inline-flex items-center gap-2 px-4 py-2 text-xs bg-accent-brand text-white font-semibold rounded-xl disabled:opacity-60">
                {submitting && <Loader2 className="size-4 animate-spin" />}
                {t("actions.confirmRequest")}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}
