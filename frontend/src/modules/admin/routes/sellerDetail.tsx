"use client";

import { useCallback, useEffect, useState } from "react";
import { ArrowLeft, FileText } from "lucide-react";
import { Link } from "@/shared/i18n/navigation";
import { ErrorState } from "@/shared/ui/errorState";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/form/dialog";
import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { adminService } from "@/modules/admin/api/adminService";
import type { AdminPayout, AdminSeller, SellerDocument } from "@/modules/admin/types";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { normalizeMediaUrl } from "@/shared/lib/media";
import { useLocale, useTranslations } from "next-intl";

const STATUS_LABEL_KEYS: Record<string, string> = {
  pending: "status.pending",
  active: "status.active",
  suspended: "status.suspended",
  rejected: "status.rejected",
};

export function AdminSellerDetail({ sellerId }: { sellerId: string }) {
  const t = useTranslations("admin");
  const locale = useLocale();
  const statusLabel = (status: string) => {
    const key = STATUS_LABEL_KEYS[status] ?? status;
    return t(key);
  };
  const [seller, setSeller] = useState<AdminSeller | null>(null);
  const [payouts, setPayouts] = useState<AdminPayout[]>([]);
  const [documents, setDocuments] = useState<SellerDocument[]>([]);
  const [reviewTarget, setReviewTarget] = useState<{
    doc: SellerDocument;
    status: "approved" | "rejected";
  } | null>(null);
  const [reviewReason, setReviewReason] = useState("");
  const [reviewing, setReviewing] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [commissionInput, setCommissionInput] = useState("");
  const [reasonTarget, setReasonTarget] = useState<string | null>(null);
  const [reason, setReason] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setLoadError(false);
    try {
      const [sellerRes, payoutsRes, documentsRes] = await Promise.all([
        adminService.getSellerById(sellerId),
        adminService.getSellerPayouts(),
        adminService.getSellerDocuments(sellerId).catch(() => ({ data: [] as SellerDocument[] })),
      ]);
      const nextSeller = sellerRes.data;
      setSeller(nextSeller);
      setPayouts(payoutsRes.data || []);
      setDocuments(Array.isArray(documentsRes?.data) ? documentsRes.data : []);
      setCommissionInput(((nextSeller?.commissionRateBps ?? 0) / 100).toFixed(2));
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [sellerId]);

  useEffect(() => {
    void load();
  }, [load]);

  const handleTransition = async (to: string, reviewReason = "") => {
    if (!seller) return;
    if ((to === "rejected" || to === "suspended") && !reviewReason.trim()) {
      notify.error(t("sellers.reason.required"));
      return;
    }
    try {
      const updated = await adminService.updateSellerStatus(seller.id, to, reviewReason.trim());
      notify.success(t("sellers.toast.statusUpdated"));
      if (updated.data) setSeller(updated.data);
      setReasonTarget(null);
      setReason("");
      await load();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("sellers.toast.statusFailed")));
    }
  };

  const handleSaveCommission = async () => {
    if (!seller) return;
    const rateBps = Math.round(Number(commissionInput) * 100);
    if (!Number.isFinite(rateBps) || rateBps < 0 || rateBps > 10000) {
      notify.error(t("sellers.commissionInvalid"));
      return;
    }
    try {
      const updated = await adminService.updateSellerCommission(seller.id, rateBps);
      notify.success(t("sellers.toast.statusUpdated"));
      if (updated.data) setSeller(updated.data);
    } catch (err) {
      notify.error(apiErrorMessage(err, t("sellers.toast.statusFailed")));
    }
  };

  const handlePayoutTransition = async (payout: AdminPayout, status: string) => {
    try {
      await adminService.transitionSellerPayout(payout.id, status);
      notify.success(t("sellers.toast.statusUpdated"));
      const payoutsRes = await adminService.getSellerPayouts();
      setPayouts(payoutsRes.data || []);
    } catch (err) {
      notify.error(apiErrorMessage(err, t("sellers.toast.statusFailed")));
    }
  };

  const handleReviewDocument = async () => {
    if (!reviewTarget) return;
    if (reviewTarget.status === "rejected" && !reviewReason.trim()) {
      notify.error(t("sellers.kyc.reasonRequired"));
      return;
    }
    setReviewing(true);
    try {
      await adminService.reviewSellerDocument(
        sellerId,
        reviewTarget.doc.id,
        reviewTarget.status,
        reviewReason.trim(),
      );
      notify.success(t("sellers.kyc.reviewed"));
      setReviewTarget(null);
      setReviewReason("");
      await load();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("sellers.kyc.reviewFailed")));
    } finally {
      setReviewing(false);
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
        title={t("sellers.loadTitle")}
        description={t("sellers.loadDescription")}
        onRetry={() => void load()}
      />
    );
  }

  if (!seller) {
    return null;
  }

  const statusAction =
    seller.status === "pending" ? (
      <>
        <button
          onClick={() => void handleTransition("active")}
          className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
        >
          {t("sellers.actions.approve")}
        </button>
        <button
          onClick={() => {
            setReasonTarget("rejected");
            setReason("");
          }}
          className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
        >
          {t("sellers.actions.reject")}
        </button>
      </>
    ) : seller.status === "active" ? (
      <button
        onClick={() => {
          setReasonTarget("suspended");
          setReason("");
        }}
        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
      >
        {t("sellers.actions.suspend")}
      </button>
    ) : (
      <button
        onClick={() => void handleTransition("active")}
        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
      >
        {t("sellers.actions.activate")}
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
          href="/admin/sellers"
          className="mb-2 inline-flex items-center gap-1.5 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-3.5 rtl:rotate-180" />
          {t("sellers.back")}
        </Link>
        <PageHeader
          title={seller.storeName}
          actions={<StatusPill tone={statusTone(seller.status)}>{statusLabel(seller.status)}</StatusPill>}
        />
        <p className="mt-1 font-mono text-xs text-muted-foreground">{seller.slug}</p>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <section className="space-y-4 rounded-lg border border-border bg-background p-4">
          <h2 className="text-sm font-semibold text-foreground">{t("sellers.commission")}</h2>
          <dl className="flex items-baseline justify-between gap-3 text-sm">
            <dt className="text-muted-foreground">{t("sellers.commission")}</dt>
            <dd className="font-medium text-foreground tabular-nums">{(seller.commissionRateBps / 100).toFixed(2)}%</dd>
          </dl>
          <div className="space-y-2">
            <label className="text-sm font-medium text-foreground" htmlFor="seller-commission">
              {t("sellers.commission")}
            </label>
            <input
              id="seller-commission"
              type="number"
              step="0.01"
              min="0"
              max="100"
              value={commissionInput}
              onChange={(event) => setCommissionInput(event.target.value)}
              className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
            />
            <div className="text-xs text-muted-foreground">{t("sellers.commissionHint")}</div>
            <button
              onClick={() => void handleSaveCommission()}
              className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
            >
              {t("actions.save")}
            </button>
          </div>
        </section>

        <section className="space-y-2 rounded-lg border border-border bg-background p-4">
          <h2 className="text-sm font-semibold text-foreground">{t("sellers.subtitle")}</h2>
          <p className="text-sm text-muted-foreground">{seller.description || "—"}</p>
          <dl className="space-y-2 text-sm">
            <div className="flex items-baseline justify-between gap-3">
              <dt className="text-muted-foreground">{t("sellers.kyc.legalName")}</dt>
              <dd className="text-foreground">{seller.legalName || "—"}</dd>
            </div>
            <div className="flex items-baseline justify-between gap-3">
              <dt className="text-muted-foreground">{t("sellers.kyc.taxNumber")}</dt>
              <dd className="font-mono text-xs text-foreground" dir="ltr">{seller.taxNumber || "—"}</dd>
            </div>
            <div className="flex items-baseline justify-between gap-3">
              <dt className="text-muted-foreground">{t("sellers.kyc.commercialRegister")}</dt>
              <dd className="font-mono text-xs text-foreground" dir="ltr">{seller.commercialRegister || "—"}</dd>
            </div>
            <div className="flex items-baseline justify-between gap-3">
              <dt className="text-muted-foreground">{t("sellers.registered")}</dt>
              <dd className="text-xs text-foreground">{new Date(seller.createdAt).toLocaleDateString(locale)}</dd>
            </div>
          </dl>
        </section>
      </div>

      <section className="space-y-4 rounded-lg border border-border bg-background p-4">
        <h2 className="text-sm font-semibold text-foreground">{t("sellers.statusTitle")}</h2>
        <div className="flex flex-wrap items-center gap-1">{statusAction}</div>

        {seller.statusReason ? (
          <div className="rounded-md border border-border bg-muted/30 p-3">
            <div className="text-xs font-medium text-foreground">
              {t("sellers.reason.current")}
            </div>
            <p className="mt-1 text-sm text-muted-foreground">{seller.statusReason}</p>
            {seller.reviewedAt ? (
              <p className="mt-1 text-[11px] text-muted-foreground">
                {t("sellers.reason.reviewedAt", {
                  date: new Date(seller.reviewedAt).toLocaleDateString(locale),
                })}
              </p>
            ) : null}
          </div>
        ) : null}

        {reasonTarget ? (
          <div className="space-y-2 rounded-md border border-border p-3">
            <label
              className="text-xs font-medium text-foreground"
              htmlFor="seller-status-reason"
            >
              {t("sellers.reason.label", {
                action: t(
                  reasonTarget === "rejected"
                    ? "sellers.actions.reject"
                    : "sellers.actions.suspend",
                ),
              })}
            </label>
            <textarea
              id="seller-status-reason"
              rows={3}
              maxLength={1000}
              value={reason}
              onChange={(event) => setReason(event.target.value)}
              placeholder={t("sellers.reason.placeholder")}
              className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
            />
            <div className="flex items-center gap-2">
              <button
                onClick={() => void handleTransition(reasonTarget, reason)}
                disabled={!reason.trim()}
                className="rounded-md bg-destructive px-3 py-1.5 text-xs font-medium text-white disabled:opacity-50"
              >
                {t("sellers.reason.confirm", {
                  action: t(
                    reasonTarget === "rejected"
                      ? "sellers.actions.reject"
                      : "sellers.actions.suspend",
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
                {t("sellers.reason.cancel")}
              </button>
            </div>
          </div>
        ) : null}
      </section>

      <section className="overflow-hidden rounded-lg border border-border bg-background">
        <div className="border-b border-border p-4">
          <h2 className="text-sm font-semibold text-foreground">{t("sellers.payoutsTitle")}</h2>
          <p className="mt-0.5 text-xs text-muted-foreground">{t("sellers.payoutsSubtitle")}</p>
        </div>
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <th className="py-2.5 ps-4 text-start font-medium">{t("sellers.payouts.colId")}</th>
                <th className="py-2.5 text-start font-medium">{t("sellers.payouts.colAmount")}</th>
                <th className="py-2.5 text-start font-medium">{t("sellers.payouts.colStatus")}</th>
                <th className="py-2.5 pe-4 text-end font-medium">{t("sellers.payouts.colAction")}</th>
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
                        {t(`sellers.payouts.status.${payout.status}`, { defaultValue: payout.status })}
                      </StatusPill>
                    </td>
                    <td className="py-3 pe-4 text-end">
                      {next ? (
                        <button
                          onClick={() => void handlePayoutTransition(payout, next)}
                          className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                        >
                          {t(`sellers.payouts.actions.${next}`)}
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
                    {t("sellers.payouts.empty")}
                  </td>
                </tr>
              )}
            </tbody>
          </ResponsiveTable>
        </div>
      </section>

      <section className="overflow-hidden rounded-lg border border-border bg-background">
        <div className="border-b border-border p-4">
          <h2 className="text-sm font-semibold text-foreground">{t("sellers.kyc.title")}</h2>
          <p className="mt-0.5 text-xs text-muted-foreground">{t("sellers.kyc.subtitle")}</p>
        </div>
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <th className="py-2.5 ps-4 text-start font-medium">{t("sellers.kyc.colType")}</th>
                <th className="py-2.5 text-start font-medium">{t("sellers.kyc.colFile")}</th>
                <th className="py-2.5 text-start font-medium">{t("sellers.kyc.colStatus")}</th>
                <th className="py-2.5 pe-4 text-end font-medium">{t("sellers.kyc.colAction")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {documents.map((doc) => (
                <tr key={doc.id} className="transition-colors hover:bg-secondary/30">
                  <td className="py-3 ps-4 text-foreground">
                    {t(`sellers.kyc.types.${doc.documentType}`, { defaultValue: doc.documentType })}
                  </td>
                  <td className="py-3">
                    {doc.documentUrl ? (
                      <a
                        href={normalizeMediaUrl(doc.documentUrl)}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex items-center gap-1.5 text-xs text-accent-brand hover:underline"
                      >
                        <FileText className="size-3.5" />
                        <span className="max-w-48 truncate">{doc.fileName}</span>
                      </a>
                    ) : (
                      <span className="text-xs text-muted-foreground">{doc.fileName}</span>
                    )}
                    {doc.status === "rejected" && doc.rejectionReason ? (
                      <p className="mt-1 text-xs text-destructive">{doc.rejectionReason}</p>
                    ) : null}
                  </td>
                  <td className="py-3">
                    <StatusPill tone={statusTone(doc.status)}>
                      {t(`sellers.kyc.status.${doc.status}`, { defaultValue: doc.status })}
                    </StatusPill>
                  </td>
                  <td className="py-3 pe-4 text-end">
                    <div className="flex justify-end gap-1">
                      <button
                        onClick={() => {
                          setReviewTarget({ doc, status: "approved" });
                          setReviewReason("");
                        }}
                        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                      >
                        {t("sellers.actions.approve")}
                      </button>
                      <button
                        onClick={() => {
                          setReviewTarget({ doc, status: "rejected" });
                          setReviewReason("");
                        }}
                        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                      >
                        {t("sellers.actions.reject")}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
              {documents.length === 0 && (
                <tr>
                  <td colSpan={4} className="py-10 text-center text-sm text-muted-foreground">
                    {t("sellers.kyc.empty")}
                  </td>
                </tr>
              )}
            </tbody>
          </ResponsiveTable>
        </div>
      </section>

      <Dialog
        open={reviewTarget !== null}
        onOpenChange={(open) => {
          if (!open) {
            setReviewTarget(null);
            setReviewReason("");
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {t(
                reviewTarget?.status === "rejected"
                  ? "sellers.kyc.rejectTitle"
                  : "sellers.kyc.approveTitle",
              )}
            </DialogTitle>
            <DialogDescription>{t("sellers.kyc.reviewDescription")}</DialogDescription>
          </DialogHeader>
          {reviewTarget?.status === "rejected" ? (
            <div className="space-y-2">
              <label htmlFor="kyc-review-reason" className="text-sm font-medium text-foreground">
                {t("sellers.kyc.reasonLabel")}
              </label>
              <textarea
                id="kyc-review-reason"
                rows={3}
                maxLength={1000}
                value={reviewReason}
                onChange={(event) => setReviewReason(event.target.value)}
                placeholder={t("sellers.reason.placeholder")}
                className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
              />
            </div>
          ) : null}
          <DialogFooter>
            <button
              onClick={() => {
                setReviewTarget(null);
                setReviewReason("");
              }}
              className="rounded-md border border-border px-3 py-2 text-sm font-medium hover:bg-secondary"
            >
              {t("actions.cancel")}
            </button>
            <button
              onClick={() => void handleReviewDocument()}
              disabled={
                reviewing || (reviewTarget?.status === "rejected" && !reviewReason.trim())
              }
              className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-50"
            >
              {t(
                reviewTarget?.status === "rejected"
                  ? "sellers.actions.reject"
                  : "sellers.actions.approve",
              )}
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
