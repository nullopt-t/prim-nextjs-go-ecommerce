"use client";

import { useCallback, useEffect, useState } from "react";
import { ArrowLeft } from "lucide-react";
import { Link } from "@/shared/i18n/navigation";
import { ErrorState } from "@/shared/ui/errorState";
import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
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
  const [reasonTarget, setReasonTarget] = useState<string | null>(null);
  const [reason, setReason] = useState("");

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

  const handleTransition = async (to: string, reviewReason = "") => {
    if (!vendor) return;
    if ((to === "rejected" || to === "suspended") && !reviewReason.trim()) {
      notify.error(t("vendors.reason.required"));
      return;
    }
    try {
      const updated = await adminService.updateVendorStatus(vendor.id, to, reviewReason.trim());
      notify.success(t("vendors.toast.statusUpdated"));
      if (updated.data) setVendor(updated.data);
      setReasonTarget(null);
      setReason("");
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
          className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
        >
          {t("vendors.actions.approve")}
        </button>
        <button
          onClick={() => {
            setReasonTarget("rejected");
            setReason("");
          }}
          className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
        >
          {t("vendors.actions.reject")}
        </button>
      </>
    ) : vendor.status === "active" ? (
      <button
        onClick={() => {
          setReasonTarget("suspended");
          setReason("");
        }}
        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
      >
        {t("vendors.actions.suspend")}
      </button>
    ) : (
      <button
        onClick={() => void handleTransition("active")}
        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
      >
        {t("vendors.actions.activate")}
      </button>
    );

  const payoutAction = (payout: AdminPayout): string | null => {
    if (payout.status === "requested") return "approved";
    if (payout.status === "approved") return "processing";
    if (payout.status === "processing") return "paid";
    return null;
  };

  return (
    <div className="space-y-5">
      <div>
        <Link
          href="/admin/vendors"
          className="mb-2 inline-flex items-center gap-1.5 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-3.5 rtl:rotate-180" />
          {t("vendors.back")}
        </Link>
        <PageHeader
          title={vendor.storeName}
          actions={<StatusPill tone={statusTone(vendor.status)}>{statusLabel(vendor.status)}</StatusPill>}
        />
        <p className="mt-1 font-mono text-xs text-muted-foreground">{vendor.slug}</p>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <section className="space-y-4 rounded-lg border border-border bg-background p-4">
          <h2 className="text-sm font-semibold text-foreground">{t("vendors.commission")}</h2>
          <dl className="flex items-baseline justify-between gap-3 text-sm">
            <dt className="text-muted-foreground">{t("vendors.commission")}</dt>
            <dd className="font-medium text-foreground tabular-nums">{(vendor.commissionRateBps / 100).toFixed(2)}%</dd>
          </dl>
          <div className="space-y-2">
            <label className="text-sm font-medium text-foreground" htmlFor="vendor-commission">
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
              className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
            />
            <div className="text-xs text-muted-foreground">{t("vendors.commissionHint")}</div>
            <button
              onClick={() => void handleSaveCommission()}
              className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
            >
              {t("actions.save")}
            </button>
          </div>
        </section>

        <section className="space-y-2 rounded-lg border border-border bg-background p-4">
          <h2 className="text-sm font-semibold text-foreground">{t("vendors.subtitle")}</h2>
          <p className="text-sm text-muted-foreground">{vendor.description || "—"}</p>
          <dl className="flex items-baseline justify-between gap-3 text-sm">
            <dt className="text-muted-foreground">{t("vendors.registered")}</dt>
            <dd className="text-xs text-foreground">{new Date(vendor.createdAt).toLocaleDateString(locale)}</dd>
          </dl>
        </section>
      </div>

      <section className="space-y-4 rounded-lg border border-border bg-background p-4">
        <h2 className="text-sm font-semibold text-foreground">{t("vendors.statusTitle")}</h2>
        <div className="flex flex-wrap items-center gap-1">{statusAction}</div>

        {vendor.statusReason ? (
          <div className="rounded-md border border-border bg-muted/30 p-3">
            <div className="text-xs font-medium text-foreground">
              {t("vendors.reason.current")}
            </div>
            <p className="mt-1 text-sm text-muted-foreground">{vendor.statusReason}</p>
            {vendor.reviewedAt ? (
              <p className="mt-1 text-[11px] text-muted-foreground">
                {t("vendors.reason.reviewedAt", {
                  date: new Date(vendor.reviewedAt).toLocaleDateString(locale),
                })}
              </p>
            ) : null}
          </div>
        ) : null}

        {reasonTarget ? (
          <div className="space-y-2 rounded-md border border-border p-3">
            <label
              className="text-xs font-medium text-foreground"
              htmlFor="vendor-status-reason"
            >
              {t("vendors.reason.label", {
                action: t(
                  reasonTarget === "rejected"
                    ? "vendors.actions.reject"
                    : "vendors.actions.suspend",
                ),
              })}
            </label>
            <textarea
              id="vendor-status-reason"
              rows={3}
              maxLength={1000}
              value={reason}
              onChange={(event) => setReason(event.target.value)}
              placeholder={t("vendors.reason.placeholder")}
              className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
            />
            <div className="flex items-center gap-2">
              <button
                onClick={() => void handleTransition(reasonTarget, reason)}
                disabled={!reason.trim()}
                className="rounded-md bg-destructive px-3 py-1.5 text-xs font-medium text-white disabled:opacity-50"
              >
                {t("vendors.reason.confirm", {
                  action: t(
                    reasonTarget === "rejected"
                      ? "vendors.actions.reject"
                      : "vendors.actions.suspend",
                  ),
                })}
              </button>
              <button
                onClick={() => {
                  setReasonTarget(null);
                  setReason("");
                }}
                className="rounded-md border border-border px-3 py-1.5 text-xs font-medium hover:bg-secondary"
              >
                {t("vendors.reason.cancel")}
              </button>
            </div>
          </div>
        ) : null}
      </section>

      <section className="overflow-hidden rounded-lg border border-border bg-background">
        <div className="border-b border-border p-4">
          <h2 className="text-sm font-semibold text-foreground">{t("vendors.payoutsTitle")}</h2>
          <p className="mt-0.5 text-xs text-muted-foreground">{t("vendors.payoutsSubtitle")}</p>
        </div>
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <th className="py-2.5 ps-4 text-start font-medium">{t("vendors.payouts.colId")}</th>
                <th className="py-2.5 text-start font-medium">{t("vendors.payouts.colAmount")}</th>
                <th className="py-2.5 text-start font-medium">{t("vendors.payouts.colStatus")}</th>
                <th className="py-2.5 pe-4 text-end font-medium">{t("vendors.payouts.colAction")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {payouts.map((payout) => {
                const next = payoutAction(payout);
                return (
                  <tr key={payout.id} className="transition-colors hover:bg-secondary/30">
                    <td className="py-3 ps-4 font-mono text-xs text-muted-foreground">{payout.id}</td>
                    <td className="py-3 font-medium text-foreground tabular-nums">
                      {payout.currency} {((payout.amount ?? 0) / 100).toFixed(2)}
                    </td>
                    <td className="py-3">
                      <StatusPill tone={statusTone(payout.status)}>
                        {t(`vendors.payouts.status.${payout.status}`, { defaultValue: payout.status })}
                      </StatusPill>
                    </td>
                    <td className="py-3 pe-4 text-end">
                      {next ? (
                        <button
                          onClick={() => void handlePayoutTransition(payout, next)}
                          className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                        >
                          {t(`vendors.payouts.actions.${next}`)}
                        </button>
                      ) : (
                        <span className="text-xs text-muted-foreground">—</span>
                      )}
                    </td>
                  </tr>
                );
              })}
              {payouts.length === 0 && (
                <tr>
                  <td colSpan={4} className="py-10 text-center text-sm text-muted-foreground">
                    {t("vendors.payouts.empty")}
                  </td>
                </tr>
              )}
            </tbody>
          </ResponsiveTable>
        </div>
      </section>
    </div>
  );
}
