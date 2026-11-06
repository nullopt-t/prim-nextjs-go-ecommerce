"use client";

import { useCallback, useEffect, useState } from "react";
import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { adminService } from "@/modules/admin/api/adminService";
import type { AdminSeller } from "@/modules/admin/types";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { Link } from "@/shared/i18n/navigation";
import { useLocale, useTranslations } from "next-intl";

const STATUS_LABEL_KEYS: Record<string, string> = {
  pending: "status.pending",
  active: "status.active",
  suspended: "status.suspended",
  rejected: "status.rejected",
};

export function AdminSellers() {
  const t = useTranslations("admin");
  const locale = useLocale();
  const statusLabel = (status: string) => {
    const key = STATUS_LABEL_KEYS[status] ?? status;
    return t(key);
  };
  const [sellers, setSellers] = useState<AdminSeller[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [statusFilter, setStatusFilter] = useState("all");

  const loadSellers = useCallback(async () => {
    setLoading(true);
    setLoadError(false);
    try {
      const res = await adminService.getSellers({
        status: statusFilter === "all" ? undefined : statusFilter,
        pageSize: 100,
      });
      setSellers(res.data || []);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [statusFilter]);

  useEffect(() => {
    void loadSellers();
  }, [loadSellers]);

  const handleTransition = async (seller: AdminSeller, to: string) => {
    try {
      const updated = await adminService.updateSellerStatus(seller.id, to);
      notify.success(t("sellers.toast.statusUpdated"));
      if (updated.data) {
        setSellers((prev) => prev.map((v) => (v.id === seller.id ? updated.data as AdminSeller : v)));
      }
      await loadSellers();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("sellers.toast.statusFailed")));
    }
  };

  if (loadError) {
    return (
      <ErrorState
        title={t("sellers.loadTitle")}
        description={t("sellers.loadDescription")}
        onRetry={() => void loadSellers()}
      />
    );
  }

  return (
    <div className="space-y-5">
      <PageHeader
        title={t("sellers.title")}
        actions={
          <button
            onClick={() => void loadSellers()}
            disabled={loading}
            className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-secondary disabled:opacity-50"
          >
            {t("actions.refresh")}
          </button>
        }
      />

      <div className="flex items-center justify-between gap-3">
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
          aria-label={t("status.all")}
        >
          <option value="all">{t("status.all")}</option>
          <option value="pending">{t("status.pending")}</option>
          <option value="active">{t("status.active")}</option>
          <option value="suspended">{t("status.suspended")}</option>
          <option value="rejected">{t("status.rejected")}</option>
        </select>
        <span className="text-xs text-muted-foreground">
          {t("sellers.count", { count: sellers.length })}
        </span>
      </div>

      <div className="overflow-hidden rounded-lg border border-border bg-background">
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <th className="py-2.5 ps-4 text-start font-medium">{t("sellers.colStore")}</th>
                <th className="py-2.5 text-start font-medium">{t("sellers.colStatus")}</th>
                <th className="py-2.5 text-start font-medium">{t("sellers.colCommission")}</th>
                <th className="py-2.5 text-start font-medium">{t("sellers.colRegistered")}</th>
                <th className="py-2.5 pe-4 text-end font-medium">{t("sellers.colAction")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {sellers.map((seller) => (
                <tr key={seller.id} className="transition-colors hover:bg-secondary/30">
                  <td className="py-3 ps-4">
                    <Link
                      href={`/admin/sellers/${seller.id}`}
                      className="font-medium text-foreground hover:underline"
                    >
                      {seller.storeName}
                    </Link>
                    <div className="font-mono text-xs text-muted-foreground">{seller.slug}</div>
                  </td>
                  <td className="py-3">
                    <StatusPill tone={statusTone(seller.status)}>{statusLabel(seller.status)}</StatusPill>
                  </td>
                  <td className="py-3 font-medium text-foreground tabular-nums">{(seller.commissionRateBps / 100).toFixed(2)}%</td>
                  <td className="py-3 text-xs text-muted-foreground">{new Date(seller.createdAt).toLocaleDateString(locale)}</td>
                  <td className="py-3 pe-4 text-end">
                    <div className="inline-flex items-center gap-1">
                      <Link
                        href={`/admin/sellers/${seller.id}`}
                        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                      >
                        {t("sellers.actions.view")}
                      </Link>
                      {seller.status === "pending" ? (
                        <>
                          <button
                            onClick={() => void handleTransition(seller, "active")}
                            className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                          >
                            {t("sellers.actions.approve")}
                          </button>
                          <button
                            onClick={() => void handleTransition(seller, "rejected")}
                            className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                          >
                            {t("sellers.actions.reject")}
                          </button>
                        </>
                      ) : null}
                      {seller.status === "active" ? (
                        <button
                          onClick={() => void handleTransition(seller, "suspended")}
                          className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                        >
                          {t("sellers.actions.suspend")}
                        </button>
                      ) : null}
                      {seller.status === "suspended" ? (
                        <button
                          onClick={() => void handleTransition(seller, "active")}
                          className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                        >
                          {t("sellers.actions.activate")}
                        </button>
                      ) : null}
                    </div>
                  </td>
                </tr>
              ))}
              {sellers.length === 0 ? (
                <tr>
                  <td colSpan={5} className="py-10 text-center text-sm text-muted-foreground">
                    {t("sellers.empty")}
                  </td>
                </tr>
              ) : null}
            </tbody>
          </ResponsiveTable>
        </div>
      </div>
    </div>
  );
}
