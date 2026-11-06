"use client";

import { useCallback, useEffect, useState } from "react";
import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { adminService } from "@/modules/admin/api/adminService";
import type { AdminVendor } from "@/modules/admin/types";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { Link } from "@/shared/i18n/navigation";
import { useLocale, useTranslations } from "next-intl";

const STATUS_LABEL_KEYS: Record<string, string> = {
  pending: "status.pending",
  active: "status.active",
  suspended: "status.suspended",
  rejected: "status.rejected",
};

export function AdminVendors() {
  const t = useTranslations("admin");
  const locale = useLocale();
  const statusLabel = (status: string) => {
    const key = STATUS_LABEL_KEYS[status] ?? status;
    return t(key);
  };
  const [vendors, setVendors] = useState<AdminVendor[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [statusFilter, setStatusFilter] = useState("all");

  const loadVendors = useCallback(async () => {
    setLoading(true);
    setLoadError(false);
    try {
      const res = await adminService.getVendors({
        status: statusFilter === "all" ? undefined : statusFilter,
        pageSize: 100,
      });
      setVendors(res.data || []);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [statusFilter]);

  useEffect(() => {
    void loadVendors();
  }, [loadVendors]);

  const handleTransition = async (vendor: AdminVendor, to: string) => {
    try {
      const updated = await adminService.updateVendorStatus(vendor.id, to);
      notify.success(t("vendors.toast.statusUpdated"));
      if (updated.data) {
        setVendors((prev) => prev.map((v) => (v.id === vendor.id ? updated.data as AdminVendor : v)));
      }
      await loadVendors();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("vendors.toast.statusFailed")));
    }
  };

  if (loadError) {
    return (
      <ErrorState
        title={t("vendors.loadTitle")}
        description={t("vendors.loadDescription")}
        onRetry={() => void loadVendors()}
      />
    );
  }

  return (
    <div className="space-y-5">
      <PageHeader
        title={t("vendors.title")}
        actions={
          <button
            onClick={() => void loadVendors()}
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
          {t("vendors.count", { count: vendors.length })}
        </span>
      </div>

      <div className="overflow-hidden rounded-lg border border-border bg-background">
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <th className="py-2.5 ps-4 text-start font-medium">{t("vendors.colStore")}</th>
                <th className="py-2.5 text-start font-medium">{t("vendors.colStatus")}</th>
                <th className="py-2.5 text-start font-medium">{t("vendors.colCommission")}</th>
                <th className="py-2.5 text-start font-medium">{t("vendors.colRegistered")}</th>
                <th className="py-2.5 pe-4 text-end font-medium">{t("vendors.colAction")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {vendors.map((vendor) => (
                <tr key={vendor.id} className="transition-colors hover:bg-secondary/30">
                  <td className="py-3 ps-4">
                    <Link
                      href={`/admin/vendors/${vendor.id}`}
                      className="font-medium text-foreground hover:underline"
                    >
                      {vendor.storeName}
                    </Link>
                    <div className="font-mono text-xs text-muted-foreground">{vendor.slug}</div>
                  </td>
                  <td className="py-3">
                    <StatusPill tone={statusTone(vendor.status)}>{statusLabel(vendor.status)}</StatusPill>
                  </td>
                  <td className="py-3 font-medium text-foreground tabular-nums">{(vendor.commissionRateBps / 100).toFixed(2)}%</td>
                  <td className="py-3 text-xs text-muted-foreground">{new Date(vendor.createdAt).toLocaleDateString(locale)}</td>
                  <td className="py-3 pe-4 text-end">
                    <div className="inline-flex items-center gap-1">
                      <Link
                        href={`/admin/vendors/${vendor.id}`}
                        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                      >
                        {t("vendors.actions.view")}
                      </Link>
                      {vendor.status === "pending" ? (
                        <>
                          <button
                            onClick={() => void handleTransition(vendor, "active")}
                            className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                          >
                            {t("vendors.actions.approve")}
                          </button>
                          <button
                            onClick={() => void handleTransition(vendor, "rejected")}
                            className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                          >
                            {t("vendors.actions.reject")}
                          </button>
                        </>
                      ) : null}
                      {vendor.status === "active" ? (
                        <button
                          onClick={() => void handleTransition(vendor, "suspended")}
                          className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                        >
                          {t("vendors.actions.suspend")}
                        </button>
                      ) : null}
                      {vendor.status === "suspended" ? (
                        <button
                          onClick={() => void handleTransition(vendor, "active")}
                          className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                        >
                          {t("vendors.actions.activate")}
                        </button>
                      ) : null}
                    </div>
                  </td>
                </tr>
              ))}
              {vendors.length === 0 ? (
                <tr>
                  <td colSpan={5} className="py-10 text-center text-sm text-muted-foreground">
                    {t("vendors.empty")}
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
