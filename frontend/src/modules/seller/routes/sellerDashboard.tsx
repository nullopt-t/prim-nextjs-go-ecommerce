"use client";

import { useEffect, useState } from "react";
import { Link } from "@/shared/i18n/navigation";
import { sellerService } from "@/modules/seller/api/sellerService";
import type { SellerDashboard as SellerDashboardData } from "@/modules/seller/types";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";
import { useSellerContext } from "./sellerContext";

export function formatMinor(amount: number, currency = "EGP") {
  return `${currency} ${(amount / 100).toFixed(2)}`;
}

const EMPTY_DASHBOARD: SellerDashboardData = {
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

export function SellerDashboard() {
  const t = useTranslations("seller");
  const { seller } = useSellerContext();
  const [dashboard, setDashboard] = useState<SellerDashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<Error | null>(null);

  const loadDashboard = async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const response = await sellerService.getDashboard();
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
  const commissionPercent = seller ? seller.commissionRateBps / 100 : 0;

  const nextSteps = (() => {
    const items: { label: string; detail: string; href: string; done: boolean }[] = [
      {
        label: t("dashboard.nextSteps.addProduct"),
        detail: t("dashboard.nextSteps.addProductDetail"),
        href: "/seller/products/new",
        done: stats.productCount > 0,
      },
      {
        label: t("dashboard.nextSteps.finishDrafts"),
        detail: t("dashboard.nextSteps.finishDraftsDetail", { count: stats.draftProductCount }),
        href: "/seller/products?status=draft",
        done: stats.draftProductCount === 0,
      },
      {
        label: t("dashboard.nextSteps.restock"),
        detail: t("dashboard.nextSteps.restockDetail", { count: stats.lowStockVariantCount }),
        href: "/seller/inventory",
        done: stats.lowStockVariantCount === 0,
      },
      {
        label: t("dashboard.nextSteps.requestPayout"),
        detail: t("dashboard.nextSteps.requestPayoutDetail", { amount: formatMinor(stats.availableBalance) }),
        href: "/seller/payouts",
        done: stats.availableBalance <= 0,
      },
    ];
    return items;
  })();
  const openSteps = nextSteps.filter((item) => !item.done);

  const metrics = [
    {
      label: t("dashboard.grossSales"),
      value: formatMinor(stats.grossRevenue),
      detail: t("dashboard.acrossCatalog"),
      href: "/seller/orders" as const,
    },
    {
      label: t("dashboard.netEarnings"),
      value: formatMinor(stats.lifetimeEarnings),
      detail: t("dashboard.commissionNote", { rate: commissionPercent.toFixed(2) }),
      href: "/seller/payouts" as const,
    },
    {
      label: t("dashboard.availableBalance"),
      value: formatMinor(stats.availableBalance),
      detail: t("dashboard.outstanding", { amount: formatMinor(stats.outstandingPayouts) }),
      href: "/seller/payouts" as const,
    },
    {
      label: t("dashboard.catalogHealth"),
      value: String(stats.productCount),
      detail: t("dashboard.publishedCount", { count: stats.publishedProductCount }),
      href: "/seller/products" as const,
    },
  ];

  return (
    <div className="space-y-5">
      <PageHeader
        title={t("dashboard.title")}
        description={t("dashboard.subtitle")}
        actions={
          <>
            <Link
              href="/seller/products/new"
              className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
            >
              {t("actions.addNewProduct")}
            </Link>
            <Link
              href="/seller/products"
              className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-secondary"
            >
              {t("actions.manageProducts")}
            </Link>
          </>
        }
      />

      <section className="rounded-xl border border-border bg-background p-4">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold text-foreground">{t("dashboard.nextStepsTitle")}</h2>
          <span className="text-xs text-muted-foreground">
            {t("dashboard.nextStepsProgress", { done: nextSteps.length - openSteps.length, total: nextSteps.length })}
          </span>
        </div>
        <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-secondary" role="progressbar" aria-valuenow={nextSteps.length - openSteps.length} aria-valuemax={nextSteps.length}>
          <div
            className="h-full bg-primary transition-all"
            style={{ width: `${nextSteps.length ? Math.round(((nextSteps.length - openSteps.length) / nextSteps.length) * 100) : 0}%` }}
          />
        </div>
        {openSteps.length === 0 ? (
          <p className="mt-3 text-sm text-muted-foreground">{t("dashboard.nextStepsDone")}</p>
        ) : (
          <ul className="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-2">
            {openSteps.map((item) => (
              <li key={item.label}>
                <Link
                  href={item.href}
                  className="block rounded-lg border border-border px-3 py-2.5 text-sm transition-colors hover:bg-secondary/50"
                >
                  <span className="font-medium text-foreground">{item.label}</span>
                  <span className="block text-xs text-muted-foreground">{item.detail}</span>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </section>

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
              href="/seller/inventory"
              className="block rounded-md border border-amber-600/20 bg-amber-600/10 px-3 py-2 text-sm text-amber-700 hover:underline dark:text-amber-400"
            >
              {t("dashboard.attentionLowStock", { count: stats.lowStockVariantCount })}
            </Link>
          ) : null}
          {stats.draftProductCount > 0 ? (
            <Link
              href="/seller/products"
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
              href="/seller/payouts"
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
              href="/seller/inventory"
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
