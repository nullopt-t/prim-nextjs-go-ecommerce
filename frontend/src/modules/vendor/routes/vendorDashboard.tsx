"use client";

import { useEffect, useState } from "react";
import { Link } from "@/shared/i18n/navigation";
import { vendorService } from "@/modules/vendor/api/vendorService";
import type { VendorDashboard as VendorDashboardData } from "@/modules/vendor/types";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
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
        <div className="size-6 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
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

  const metrics = [
    {
      label: t("dashboard.grossSales"),
      value: formatMinor(stats.grossRevenue),
      detail: t("dashboard.acrossCatalog"),
      href: "/vendor/orders" as const,
    },
    {
      label: t("dashboard.netEarnings"),
      value: formatMinor(stats.lifetimeEarnings),
      detail: t("dashboard.commissionNote", { rate: commissionPercent.toFixed(2) }),
      href: "/vendor/payouts" as const,
    },
    {
      label: t("dashboard.availableBalance"),
      value: formatMinor(stats.availableBalance),
      detail: t("dashboard.outstanding", { amount: formatMinor(stats.outstandingPayouts) }),
      href: "/vendor/payouts" as const,
    },
    {
      label: t("dashboard.catalogHealth"),
      value: String(stats.productCount),
      detail: t("dashboard.publishedCount", { count: stats.publishedProductCount }),
      href: "/vendor/products" as const,
    },
  ];

  return (
    <div className="space-y-5">
      <PageHeader
        title={t("dashboard.title")}
        actions={
          <Link
            href="/vendor/products"
            className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-secondary"
          >
            {t("actions.manageProducts")}
          </Link>
        }
      />

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {metrics.map((metric) => (
          <Link
            key={metric.label}
            href={metric.href}
            className="rounded-lg border border-border bg-background p-4 transition-colors hover:bg-secondary/40"
          >
            <div className="text-xs font-medium text-muted-foreground">
              {metric.label}
            </div>
            <div className="mt-2 text-2xl font-semibold tracking-tight text-foreground tabular-nums">
              {metric.value}
            </div>
            <div className="mt-1 text-xs text-muted-foreground">{metric.detail}</div>
          </Link>
        ))}
      </div>

      {stats.lowStockVariantCount > 0 || stats.draftProductCount > 0 ? (
        <div className="space-y-2">
          {stats.lowStockVariantCount > 0 ? (
            <Link
              href="/vendor/inventory"
              className="block rounded-md border border-amber-600/20 bg-amber-600/10 px-3 py-2 text-sm text-amber-700 hover:underline dark:text-amber-400"
            >
              {t("dashboard.attentionLowStock", { count: stats.lowStockVariantCount })}
            </Link>
          ) : null}
          {stats.draftProductCount > 0 ? (
            <Link
              href="/vendor/products"
              className="block rounded-md border border-border bg-background px-3 py-2 text-sm text-muted-foreground hover:text-foreground"
            >
              {t("dashboard.attentionDrafts", { count: stats.draftProductCount })}
            </Link>
          ) : null}
        </div>
      ) : null}

      <div className="grid grid-cols-1 gap-5 lg:grid-cols-2">
        <section className="rounded-lg border border-border bg-background">
          <div className="flex items-center justify-between border-b border-border p-4">
            <h2 className="text-sm font-semibold text-foreground">
              {t("dashboard.moneyPosition")}
            </h2>
            <Link
              href="/vendor/payouts"
              className="text-xs font-medium text-muted-foreground hover:text-foreground"
            >
              {t("actions.viewPayouts")}
            </Link>
          </div>
          <dl className="divide-y divide-border text-sm">
            <div className="flex items-center justify-between px-4 py-3">
              <dt className="text-muted-foreground">{t("dashboard.pendingEarnings")}</dt>
              <dd className="font-medium text-foreground tabular-nums">
                {formatMinor(stats.pendingEarnings)}
              </dd>
            </div>
            <div className="flex items-center justify-between px-4 py-3">
              <dt className="text-muted-foreground">{t("dashboard.outstandingPayouts")}</dt>
              <dd className="font-medium text-foreground tabular-nums">
                {formatMinor(stats.outstandingPayouts)}
              </dd>
            </div>
            <div className="flex items-center justify-between px-4 py-3">
              <dt className="text-muted-foreground">{t("dashboard.orderCount")}</dt>
              <dd className="font-medium text-foreground tabular-nums">{stats.orderCount}</dd>
            </div>
          </dl>
        </section>

        <section className="rounded-lg border border-border bg-background">
          <div className="flex items-center justify-between border-b border-border p-4">
            <h2 className="text-sm font-semibold text-foreground">
              {t("dashboard.catalogHealthTitle")}
            </h2>
            <Link
              href="/vendor/inventory"
              className="text-xs font-medium text-muted-foreground hover:text-foreground"
            >
              {t("actions.viewInventory")}
            </Link>
          </div>
          <dl className="divide-y divide-border text-sm">
            <div className="flex items-center justify-between px-4 py-3">
              <dt className="text-muted-foreground">{t("dashboard.published")}</dt>
              <dd className="font-medium text-foreground tabular-nums">
                {stats.publishedProductCount}
              </dd>
            </div>
            <div className="flex items-center justify-between px-4 py-3">
              <dt className="text-muted-foreground">{t("dashboard.drafts")}</dt>
              <dd className="font-medium text-foreground tabular-nums">
                {stats.draftProductCount}
              </dd>
            </div>
            <div className="flex items-center justify-between px-4 py-3">
              <dt className="text-muted-foreground">{t("dashboard.lowStockVariants")}</dt>
              <dd className="font-medium text-foreground tabular-nums">
                {stats.lowStockVariantCount}
              </dd>
            </div>
          </dl>
        </section>
      </div>
    </div>
  );
}
