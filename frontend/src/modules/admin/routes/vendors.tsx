"use client";

import { useCallback, useEffect, useState } from "react";
import { Ban, CheckCheck, CheckCircle2, Eye, RefreshCw, XCircle } from "lucide-react";
import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { ErrorState } from "@/shared/ui/errorState";
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
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("vendors.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t("vendors.subtitle")}</p>
        </div>
        <button
          onClick={() => void loadVendors()}
          className="inline-flex items-center gap-2 px-4 py-2 border border-border rounded-xl text-xs font-semibold text-foreground hover:bg-secondary/40 transition-colors"
        >
          <RefreshCw className={`size-3.5 ${loading ? "animate-spin" : ""}`} />
          <span>{t("actions.refresh")}</span>
        </button>
      </div>

      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <div className="p-6 border-b border-border flex flex-col md:flex-row gap-3 md:items-center justify-between">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
            aria-label={t("status.all")}
          >
            <option value="all">{t("status.all")}</option>
            <option value="pending">{t("status.pending")}</option>
            <option value="active">{t("status.active")}</option>
            <option value="suspended">{t("status.suspended")}</option>
            <option value="rejected">{t("status.rejected")}</option>
          </select>
        </div>

        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full sm:min-w-max text-start text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 ps-4">{t("vendors.colStore")}</th>
                <th className="py-3.5">{t("vendors.colStatus")}</th>
                <th className="py-3.5">{t("vendors.colCommission")}</th>
                <th className="py-3.5">{t("vendors.colRegistered")}</th>
                <th className="py-3.5 pe-4 text-end">{t("vendors.colAction")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {vendors.map((vendor) => (
                <tr key={vendor.id} className="hover:bg-secondary/20 transition-colors">
                  <td className="py-3.5 ps-4">
                    <Link
                      href={`/admin/vendors/${vendor.id}`}
                      className="font-bold text-foreground hover:text-accent-brand transition-colors"
                    >
                      {vendor.storeName}
                    </Link>
                    <div className="text-[10px] text-muted-foreground font-mono">{vendor.slug}</div>
                  </td>
                  <td className="py-3.5">
                    <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-secondary text-foreground border border-border">
                      {statusLabel(vendor.status)}
                    </span>
                  </td>
                  <td className="py-3.5 font-bold text-foreground">{(vendor.commissionRateBps / 100).toFixed(2)}%</td>
                  <td className="py-3.5 text-muted-foreground">{new Date(vendor.createdAt).toLocaleDateString(locale)}</td>
                  <td className="py-3.5 pe-4 text-end">
                    <div className="inline-flex items-center gap-2">
                      <Link
                        href={`/admin/vendors/${vendor.id}`}
                        className="inline-flex items-center gap-1.5 px-3 py-1.5 border border-border rounded-xl text-[11px] font-semibold text-foreground hover:bg-secondary/40 transition-colors"
                      >
                        <Eye className="size-3.5" />
                        <span>{t("vendors.actions.view")}</span>
                      </Link>
                      {vendor.status === "pending" && (
                        <>
                          <button
                            onClick={() => void handleTransition(vendor, "active")}
                            className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20 rounded-xl text-[11px] font-semibold hover:bg-emerald-500/20 transition-colors"
                          >
                            <CheckCircle2 className="size-3.5" />
                            <span>{t("vendors.actions.approve")}</span>
                          </button>
                          <button
                            onClick={() => void handleTransition(vendor, "rejected")}
                            className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-red-500/10 text-red-600 dark:text-red-400 border border-red-500/20 rounded-xl text-[11px] font-semibold hover:bg-red-500/20 transition-colors"
                          >
                            <XCircle className="size-3.5" />
                            <span>{t("vendors.actions.reject")}</span>
                          </button>
                        </>
                      )}
                      {vendor.status === "active" && (
                        <button
                          onClick={() => void handleTransition(vendor, "suspended")}
                          className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20 rounded-xl text-[11px] font-semibold hover:bg-amber-500/20 transition-colors"
                        >
                          <Ban className="size-3.5" />
                          <span>{t("vendors.actions.suspend")}</span>
                        </button>
                      )}
                      {vendor.status === "suspended" && (
                        <button
                          onClick={() => void handleTransition(vendor, "active")}
                          className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20 rounded-xl text-[11px] font-semibold hover:bg-emerald-500/20 transition-colors"
                        >
                          <CheckCheck className="size-3.5" />
                          <span>{t("vendors.actions.activate")}</span>
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
              {vendors.length === 0 && (
                <tr>
                  <td colSpan={5} className="py-10 text-center text-muted-foreground">
                    {t("vendors.empty")}
                  </td>
                </tr>
              )}
            </tbody>
          </ResponsiveTable>
        </div>
      </div>
    </div>
  );
}
