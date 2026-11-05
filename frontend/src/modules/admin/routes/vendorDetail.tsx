"use client";

import { useCallback, useEffect, useState } from "react";
import { ArrowLeft, Ban, CheckCheck, CheckCircle2, XCircle } from "lucide-react";
import { Link } from "@/shared/i18n/navigation";
import { ErrorState } from "@/shared/ui/errorState";
import { adminService } from "@/modules/admin/api/adminService";
import type { AdminPayout, AdminVendor } from "@/modules/admin/types";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { useLocale, useTranslations } from "next-intl";

const STATUS_LABEL_KEYS: Record<string, string> = {
  pending: "status.pending",
  active: "status.active",
  suspended: "status.suspended",
  rejected: "status.rejected",
};

export function AdminVendorDetail({ vendorId }: { vendorId: string }) {
  const t = useTranslations("admin");
  const locale = useLocale();
  const statusLabel = (status: string) => {
    const key = STATUS_LABEL_KEYS[status] ?? status;
    return t(key);
  };
  const [vendor, setVendor] = useState<AdminVendor | null>(null);
  const [payouts, setPayouts] = useState<AdminPayout[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [commissionInput, setCommissionInput] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setLoadError(false);
    try {
      const [vendorRes, payoutsRes] = await Promise.all([
        adminService.getVendorById(vendorId),
        adminService.getVendorPayouts(),
      ]);
      const nextVendor = vendorRes.data;
      setVendor(nextVendor);
      setPayouts(payoutsRes.data || []);
      setCommissionInput(((nextVendor?.commissionRateBps ?? 0) / 100).toFixed(2));
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [vendorId]);

  useEffect(() => {
    void load();
  }, [load]);

  const handleTransition = async (to: string) => {
    if (!vendor) return;
    try {
      const updated = await adminService.updateVendorStatus(vendor.id, to);
      notify.success(t("vendors.toast.statusUpdated"));
      if (updated.data) setVendor(updated.data);
      await load();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("vendors.toast.statusFailed")));
    }
  };

  const handleSaveCommission = async () => {
    if (!vendor) return;
    const rateBps = Math.round(Number(commissionInput) * 100);
    if (!Number.isFinite(rateBps) || rateBps < 0 || rateBps > 10000) {
      notify.error(t("vendors.commissionInvalid"));
      return;
    }
    try {
      const updated = await adminService.updateVendorCommission(vendor.id, rateBps);
      notify.success(t("vendors.toast.statusUpdated"));
      if (updated.data) setVendor(updated.data);
    } catch (err) {
      notify.error(apiErrorMessage(err, t("vendors.toast.statusFailed")));
    }
  };

  const handlePayoutTransition = async (payout: AdminPayout, status: string) => {
    try {
      await adminService.transitionVendorPayout(payout.id, status);
      notify.success(t("vendors.toast.statusUpdated"));
      const payoutsRes = await adminService.getVendorPayouts();
      setPayouts(payoutsRes.data || []);
    } catch (err) {
      notify.error(apiErrorMessage(err, t("vendors.toast.statusFailed")));
    }
  };

  if (loading) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-accent-brand border-t-transparent" />
      </div>
    );
  }

  if (loadError) {
    return (
      <ErrorState
        title={t("vendors.loadTitle")}
        description={t("vendors.loadDescription")}
        onRetry={() => void load()}
      />
    );
  }

  if (!vendor) {
    return null;
  }

  const statusAction =
    vendor.status === "pending" ? (
      <>
        <button
          onClick={() => void handleTransition("active")}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20 rounded-xl text-[11px] font-semibold hover:bg-emerald-500/20 transition-colors"
        >
          <CheckCircle2 className="size-3.5" />
          <span>{t("vendors.actions.approve")}</span>
        </button>
        <button
          onClick={() => void handleTransition("rejected")}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-red-500/10 text-red-600 dark:text-red-400 border border-red-500/20 rounded-xl text-[11px] font-semibold hover:bg-red-500/20 transition-colors"
        >
          <XCircle className="size-3.5" />
          <span>{t("vendors.actions.reject")}</span>
        </button>
      </>
    ) : vendor.status === "active" ? (
      <button
        onClick={() => void handleTransition("suspended")}
        className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20 rounded-xl text-[11px] font-semibold hover:bg-amber-500/20 transition-colors"
      >
        <Ban className="size-3.5" />
        <span>{t("vendors.actions.suspend")}</span>
      </button>
    ) : (
      <button
        onClick={() => void handleTransition("active")}
        className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20 rounded-xl text-[11px] font-semibold hover:bg-emerald-500/20 transition-colors"
      >
        <CheckCheck className="size-3.5" />
        <span>{t("vendors.actions.activate")}</span>
      </button>
    );

  const payoutAction = (payout: AdminPayout): string | null => {
    if (payout.status === "requested") return "approved";
    if (payout.status === "approved") return "processing";
    if (payout.status === "processing") return "paid";
    return null;
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <Link
            href="/admin/vendors"
            className="inline-flex items-center gap-1.5 text-xs font-semibold text-accent-brand hover:underline"
          >
            <ArrowLeft className="size-3.5" />
            {t("vendors.back")}
          </Link>
          <h1 className="text-2xl font-bold text-foreground mt-3">{vendor.storeName}</h1>
          <p className="text-sm text-muted-foreground mt-1 font-mono">{vendor.slug}</p>
        </div>
        <span className="px-3 py-1 bg-secondary text-foreground border border-border text-xs font-bold rounded-full">
          {statusLabel(vendor.status)}
        </span>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
          <h3 className="font-bold text-base text-foreground">{t("vendors.commission")}</h3>
          <div className="text-sm text-muted-foreground">
            {t("vendors.commission")}:{" "}
            <strong className="text-foreground">{(vendor.commissionRateBps / 100).toFixed(2)}%</strong>
          </div>
          <div className="space-y-2">
            <label className="text-xs font-semibold text-foreground" htmlFor="vendor-commission">
              {t("vendors.commission")}
            </label>
            <input
              id="vendor-commission"
              type="number"
              step="0.01"
              min="0"
              max="100"
              value={commissionInput}
              onChange={(event) => setCommissionInput(event.target.value)}
              className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-base font-bold text-foreground"
            />
            <div className="text-[11px] text-muted-foreground">{t("vendors.commissionHint")}</div>
            <button
              onClick={() => void handleSaveCommission()}
              className="inline-flex items-center gap-2 px-4 py-2 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all"
            >
              <span>{t("actions.save")}</span>
            </button>
          </div>
        </div>

        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
          <h3 className="font-bold text-base text-foreground">{t("vendors.subtitle")}</h3>
          <p className="text-xs text-muted-foreground">{vendor.description || "—"}</p>
          <div className="text-xs text-muted-foreground">{t("vendors.registered")}: {new Date(vendor.createdAt).toLocaleDateString(locale)}</div>
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
        <h3 className="font-bold text-base text-foreground">{t("vendors.statusTitle")}</h3>
        <div className="flex flex-wrap items-center gap-2">{statusAction}</div>
      </div>

      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <div className="p-6 border-b border-border">
          <h3 className="font-bold text-base text-foreground">{t("vendors.payoutsTitle")}</h3>
          <p className="text-xs text-muted-foreground mt-0.5">{t("vendors.payoutsSubtitle")}</p>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full min-w-max text-left text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 ps-4">{t("vendors.payouts.colId")}</th>
                <th className="py-3.5">{t("vendors.payouts.colAmount")}</th>
                <th className="py-3.5">{t("vendors.payouts.colStatus")}</th>
                <th className="py-3.5 pe-4 text-end">{t("vendors.payouts.colAction")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {payouts.map((payout) => {
                const next = payoutAction(payout);
                return (
                  <tr key={payout.id} className="hover:bg-secondary/20 transition-colors">
                    <td className="py-3.5 ps-4 font-mono font-bold text-foreground">{payout.id}</td>
                    <td className="py-3.5 font-bold text-foreground">
                      {payout.currency} {((payout.amount ?? 0) / 100).toFixed(2)}
                    </td>
                    <td className="py-3.5">
                      <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-secondary text-foreground border border-border">
                        {t(`vendors.payouts.status.${payout.status}`, { defaultValue: payout.status })}
                      </span>
                    </td>
                    <td className="py-3.5 pe-4 text-end">
                      {next ? (
                        <button
                          onClick={() => void handlePayoutTransition(payout, next)}
                          className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-accent-brand/10 text-accent-brand border border-accent-brand/20 rounded-xl text-[11px] font-semibold hover:bg-accent-brand/20 transition-colors"
                        >
                          <span>{t(`vendors.payouts.actions.${next}`)}</span>
                        </button>
                      ) : (
                        <span className="text-[11px] text-muted-foreground">—</span>
                      )}
                    </td>
                  </tr>
                );
              })}
              {payouts.length === 0 && (
                <tr>
                  <td colSpan={4} className="py-10 text-center text-muted-foreground">
                    {t("vendors.payouts.empty")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
