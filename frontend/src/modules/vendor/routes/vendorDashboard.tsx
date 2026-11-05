"use client";

import { useEffect, useState } from "react";
import { Link } from "@/shared/i18n/navigation";
import {
  DollarSign,
  ShoppingBag,
  TrendingUp,
  Package,
  ArrowUpRight,
  Clock,
  BadgeDollarSign,
} from "lucide-react";
import { vendorService } from "@/modules/vendor/api/vendorService";
import type { VendorDashboard as VendorDashboardData } from "@/modules/vendor/types";
import { ErrorState } from "@/shared/ui/errorState";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";
import { useVendorContext } from "./vendorContext";

export function formatMinor(amount: number, currency = "EGP") {
  return `${currency} ${(amount / 100).toFixed(2)}`;
}

const EMPTY_DASHBOARD: VendorDashboardData = {
  availableBalance: 0,
  pendingEarnings: 0,
  lifetimeEarnings: 0,
  outstandingPayouts: 0,
  grossRevenue: 0,
  orderCount: 0,
  productCount: 0,
  publishedProductCount: 0,
  draftProductCount: 0,
  lowStockVariantCount: 0,
};

export function VendorDashboard() {
  const t = useTranslations("vendor");
  const { vendor } = useVendorContext();
  const [dashboard, setDashboard] = useState<VendorDashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<Error | null>(null);

  const loadDashboard = async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const response = await vendorService.getDashboard();
      setDashboard(response.data);
    } catch (error) {
      setLoadError(error instanceof Error ? error : new Error(String(error)));
      notify.error(apiErrorMessage(error, t("loadFailed")));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadDashboard();
  }, []);

  if (loading) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-accent-brand border-t-transparent" />
      </div>
    );
  }

  if (loadError || !dashboard) {
    return (
      <ErrorState
        title={t("dashboardLoadTitle")}
        description={loadError?.message || t("dashboardLoadDescription")}
        onRetry={() => void loadDashboard()}
      />
    );
  }

  const stats = dashboard ?? EMPTY_DASHBOARD;
  const commissionPercent = vendor ? vendor.commissionRateBps / 100 : 0;

  return (
    <div className="space-y-6 w-full">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold text-foreground">{t("dashboard.title")}</h1>
            <span className="px-2.5 py-0.5 rounded-full text-xs font-bold bg-accent-brand/10 text-accent-brand border border-accent-brand/20">
              {vendor?.storeName ?? "—"}
            </span>
          </div>
          <p className="text-sm text-muted-foreground mt-1">{t("dashboard.subtitle")}</p>
        </div>
        <Link
          href="/vendor/products"
          className="inline-flex items-center gap-1.5 px-4 py-2 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all"
        >
          <Package className="size-4" />
          <span>{t("actions.manageProducts")}</span>
        </Link>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex items-center justify-between">
          <div>
            <div className="text-xs font-semibold text-muted-foreground">{t("dashboard.grossSales")}</div>
            <div className="text-2xl font-bold text-foreground mt-1">{formatMinor(stats.grossRevenue)}</div>
          </div>
          <div className="size-11 rounded-2xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center">
            <DollarSign className="size-5" />
          </div>
        </div>

        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex items-center justify-between">
          <div>
            <div className="text-xs font-semibold text-muted-foreground">{t("dashboard.netEarnings")}</div>
            <div className="text-2xl font-bold text-foreground mt-1">{formatMinor(stats.lifetimeEarnings)}</div>
            <div className="text-[11px] text-muted-foreground mt-1">
              {t("dashboard.commissionNote", { rate: commissionPercent.toFixed(2) })}
            </div>
          </div>
          <div className="size-11 rounded-2xl bg-blue-500/10 text-blue-600 dark:text-blue-400 flex items-center justify-center">
            <TrendingUp className="size-5" />
          </div>
        </div>

        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex items-center justify-between">
          <div>
            <div className="text-xs font-semibold text-muted-foreground">{t("dashboard.availableBalance")}</div>
            <div className="text-2xl font-bold text-foreground mt-1">{formatMinor(stats.availableBalance)}</div>
            <div className="text-[11px] text-muted-foreground mt-1">
              {t("dashboard.outstanding", { amount: formatMinor(stats.outstandingPayouts) })}
            </div>
          </div>
          <div className="size-11 rounded-2xl bg-purple-500/10 text-purple-600 dark:text-purple-400 flex items-center justify-center">
            <BadgeDollarSign className="size-5" />
          </div>
        </div>

        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex items-center justify-between">
          <div>
            <div className="text-xs font-semibold text-muted-foreground">{t("dashboard.catalogHealth")}</div>
            <div className="text-2xl font-bold text-foreground mt-1">
              {stats.productCount}
            </div>
            <div className="text-[11px] text-muted-foreground mt-1">
              {t("dashboard.publishedCount", { count: stats.publishedProductCount })}
            </div>
          </div>
          <div className="size-11 rounded-2xl bg-amber-500/10 text-amber-600 dark:text-amber-400 flex items-center justify-center">
            <Package className="size-5" />
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
          <div className="flex justify-between items-center border-b border-border pb-3">
            <h3 className="font-bold text-base text-foreground">{t("dashboard.moneyPosition")}</h3>
            <Link href="/vendor/payouts" className="text-xs font-semibold text-accent-brand hover:underline flex items-center gap-1">
              <span>{t("actions.viewPayouts")}</span>
              <ArrowUpRight className="size-3.5" />
            </Link>
          </div>
          <div className="space-y-3 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">{t("dashboard.pendingEarnings")}</span>
              <span className="font-bold">{formatMinor(stats.pendingEarnings)}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">{t("dashboard.outstandingPayouts")}</span>
              <span className="font-bold">{formatMinor(stats.outstandingPayouts)}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">{t("dashboard.orderCount")}</span>
              <span className="font-bold">{stats.orderCount}</span>
            </div>
          </div>
        </div>

        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
          <div className="flex justify-between items-center border-b border-border pb-3">
            <h3 className="font-bold text-base text-foreground">{t("dashboard.catalogHealthTitle")}</h3>
            <Link href="/vendor/inventory" className="text-xs font-semibold text-accent-brand hover:underline flex items-center gap-1">
              <span>{t("actions.viewInventory")}</span>
              <ArrowUpRight className="size-3.5" />
            </Link>
          </div>
          <div className="space-y-3 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">{t("dashboard.published")}</span>
              <span className="font-bold">{stats.publishedProductCount}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">{t("dashboard.drafts")}</span>
              <span className="font-bold">{stats.draftProductCount}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">{t("dashboard.lowStockVariants")}</span>
              <span className="font-bold">{stats.lowStockVariantCount}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">{t("dashboard.orders")}</span>
              <span className="inline-flex items-center gap-1.5 font-bold">
                <ShoppingBag className="size-4 text-muted-foreground" />
                {stats.orderCount}
              </span>
            </div>
          </div>
          <div className="flex items-center gap-1 text-[11px] text-muted-foreground">
            <Clock className="size-3" />
            <span>{t("dashboard.dashboardNote")}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
