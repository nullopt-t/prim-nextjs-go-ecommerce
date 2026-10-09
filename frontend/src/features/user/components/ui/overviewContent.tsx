"use client";

import { useEffect, useState } from "react";
import { useAuthContext } from "@/context/AuthContext";
import { useWishlist } from "@/hooks/useWishlist";
import { Link } from "@/i18n/navigation";
import { userService } from "@/services/user";
import { useTranslations, useLocale } from "next-intl";

// ── Types ──────────────────────────────────────────────────────────────────────
interface OrderItem {
  id: string;
  variantId: string;
  quantity: number;
  priceAtPurchase: number;
  productSnapshot: string;
}

interface Order {
  id: string;
  customerEmail: string;
  status: "pending" | "processing" | "shipped" | "delivered" | "cancelled";
  totalAmount: number;
  discountAmount: number;
  currency: string;
  items: OrderItem[];
  createdAt: string;
  updatedAt: string;
}

// ── Helpers ────────────────────────────────────────────────────────────────────
function getGreeting() {
  const hour = new Date().getHours();
  if (hour < 12) return "Good morning";
  if (hour < 17) return "Good afternoon";
  return "Good evening";
}

function parseSnapshot(snapshot?: string): { title?: string; name?: string; sku?: string } {
  if (!snapshot) return {};
  try {
    return JSON.parse(snapshot) ?? {};
  } catch {
    return {};
  }
}

const STATUS_BADGE: Record<
  Order["status"],
  { label: string; className: string }
> = {
  pending: {
    label: "Pending",
    className: "text-amber-600 bg-amber-50 dark:bg-amber-950/30 dark:text-amber-400 border border-amber-200/50",
  },
  processing: {
    label: "Processing",
    className: "text-blue-600 bg-blue-50 dark:bg-blue-950/30 dark:text-blue-400 border border-blue-200/50",
  },
  shipped: {
    label: "Shipped",
    className: "text-violet-600 bg-violet-50 dark:bg-violet-950/30 dark:text-violet-400 border border-violet-200/50",
  },
  delivered: {
    label: "Delivered",
    className: "text-emerald-600 bg-emerald-50 dark:bg-emerald-950/30 dark:text-emerald-400 border border-emerald-200/50",
  },
  cancelled: {
    label: "Cancelled",
    className: "text-red-600 bg-red-50 dark:bg-red-950/30 dark:text-red-400 border border-red-200/50",
  },
};



const tbHeaders = ["Order #", "Placed On", "Items", "Total", "Status", "Actions"];

function TableSkeleton() {
  return (
    <div className="bg-card border border-border rounded-xl overflow-hidden shadow-xs">
      <div className="animate-pulse bg-secondary/50 h-10 border-b border-border" />
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="animate-pulse bg-secondary/40 rounded-lg h-14 mx-4 my-2.5" />
      ))}
    </div>
  );
}

// ── Component ──────────────────────────────────────────────────────────────────
export default function OverviewContent() {
  const { user } = useAuthContext();
  const { wishlistCount } = useWishlist();
  const t = useTranslations("overview");
  const locale = useLocale();
  const [orders, setOrders] = useState<Order[]>([]);
  const [reviewCount, setReviewCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const [ordersRes, reviewsRes] = await Promise.all([
          userService.getOrders(),
          userService.getMyReviews(),
        ]);

        if (cancelled) return;

        const rawOrders: Order[] = Array.isArray(ordersRes?.data)
          ? ordersRes.data
          : Array.isArray(ordersRes)
          ? ordersRes
          : [];

        const rawReviews: unknown[] = Array.isArray(reviewsRes?.data)
          ? reviewsRes.data
          : Array.isArray(reviewsRes)
          ? reviewsRes
          : [];

        setOrders(rawOrders);
        setReviewCount(rawReviews.length);
      } catch (err: unknown) {
        if (!cancelled) setError(err instanceof Error ? err.message : "Failed to load data.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, []);

  // Sorting & highlights
  const sortedOrders = [...orders].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );
  const latestOrder = sortedOrders[0] ?? null;
  const recentOrders = sortedOrders.slice(0, 5);

  const deliveredCount = orders.filter((o) => o.status === "delivered").length;
  const inTransitCount = orders.filter(
    (o) => o.status === "processing" || o.status === "shipped" || o.status === "pending"
  ).length;


  const localizedTbHeaders = [
    t("recentOrders.columns.order"),
    t("recentOrders.columns.placedOn"),
    t("recentOrders.columns.items"),
    t("recentOrders.columns.total"),
    t("recentOrders.columns.status"),
    t("recentOrders.columns.actions"),
  ];

  return (
    <div className="w-full flex flex-col gap-6 pb-12">
      {/* ── Top Header Strip ── */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-border/70 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-foreground tracking-tight">
            {t("heading")}
          </h1>
        </div>

        <div className="flex items-center gap-2.5 shrink-0">
          <Link
            href="/products"
            className="px-3.5 py-1.5 rounded-lg text-xs font-medium bg-secondary hover:bg-secondary/80 text-foreground border border-border transition-colors"
          >
            {t("browseCatalog")}
          </Link>
          <Link
            href="/"
            className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-accent-brand text-white hover:opacity-90 transition-opacity"
          >
            {t("storefront")}
          </Link>
        </div>
      </div>

      {/* ── High-Density Metrics Ribbon ── */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3.5">
        <div className="bg-card border border-border rounded-xl p-4 shadow-2xs flex flex-col justify-between">
          <span className="text-xs font-medium text-muted-foreground">{t("metrics.totalOrders")}</span>
          <div className="flex items-baseline justify-between mt-2">
            <span className="text-2xl font-bold tracking-tight text-foreground">{orders.length}</span>
            <span className="text-[11px] text-muted-foreground">{t("metrics.lifetime")}</span>
          </div>
        </div>

        <div className="bg-card border border-border rounded-xl p-4 shadow-2xs flex flex-col justify-between">
          <span className="text-xs font-medium text-muted-foreground">{t("metrics.inProgress")}</span>
          <div className="flex items-baseline justify-between mt-2">
            <span className="text-2xl font-bold tracking-tight text-foreground">{inTransitCount}</span>
            <span className="text-[11px] text-blue-600 dark:text-blue-400 font-medium">{t("metrics.activeTracking")}</span>
          </div>
        </div>

        <div className="bg-card border border-border rounded-xl p-4 shadow-2xs flex flex-col justify-between">
          <span className="text-xs font-medium text-muted-foreground">{t("metrics.delivered")}</span>
          <div className="flex items-baseline justify-between mt-2">
            <span className="text-2xl font-bold tracking-tight text-foreground">{deliveredCount}</span>
            <span className="text-[11px] text-emerald-600 dark:text-emerald-400 font-medium">{t("metrics.completed")}</span>
          </div>
        </div>

        <div className="bg-card border border-border rounded-xl p-4 shadow-2xs flex flex-col justify-between">
          <span className="text-xs font-medium text-muted-foreground">{t("metrics.savedWishlist")}</span>
          <div className="flex items-baseline justify-between mt-2">
            <span className="text-2xl font-bold tracking-tight text-foreground">{wishlistCount}</span>
            <Link href="/user/wishlist" className="text-[11px] text-accent-brand hover:underline font-medium">
              {t("metrics.viewItems")}
            </Link>
          </div>
        </div>
      </div>

      {/* ── Two-Column Widescreen Split ── */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
        {/* Left Column (8 cols): Active Shipment Spotlight & Recent Orders Table */}
        <div className="lg:col-span-8 flex flex-col gap-6">
          {/* Active / Latest Order Spotlight Card */}
          {latestOrder && (
            <div className="bg-card border border-border rounded-xl p-5 shadow-2xs">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3.5 border-b border-border">
                <div className="flex items-center gap-2.5">
                  <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                    {t("shipment.recentShipment")}
                  </span>
                  <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_BADGE[latestOrder.status]?.className}`}>
                    {STATUS_BADGE[latestOrder.status]?.label ?? latestOrder.status}
                  </span>
                </div>
                <div className="text-xs text-muted-foreground">
                  {t("shipment.order")} <span className="font-mono font-semibold text-foreground">#{latestOrder.id.slice(0, 8).toUpperCase()}</span> • {new Date(latestOrder.createdAt).toLocaleDateString(locale)}
                </div>
              </div>

              <div className="pt-3.5 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div className="min-w-0 flex-1">
                  <p className="text-xs text-muted-foreground mb-1">{t("shipment.purchasedItems")}</p>
                  <p className="text-sm font-medium text-foreground truncate">
                    {(latestOrder.items ?? []).map((item) => {
                      const snap = parseSnapshot(item.productSnapshot);
                      return snap.title || snap.name || "Item";
                    }).slice(0, 3).join(", ")}
                    {(latestOrder.items?.length ?? 0) > 3 && ` ${t("shipment.moreItems", { count: (latestOrder.items?.length ?? 0) - 3 })}`}
                  </p>
                </div>

                <div className="flex items-center gap-2 shrink-0">
                  <Link
                    href="/products"
                    className="px-3 py-1.5 rounded-lg text-xs font-medium bg-secondary hover:bg-secondary/80 text-foreground border border-border transition-colors"
                  >
                    {t("shipment.buyAgain")}
                  </Link>
                  <Link
                    href={`/user/orders/${latestOrder.id}`}
                    className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-accent-brand text-white hover:opacity-90 transition-opacity"
                  >
                    {t("shipment.viewOrder")}
                  </Link>
                </div>
              </div>
            </div>
          )}

          {/* Recent Orders Table */}
          <div className="bg-card border border-border rounded-xl overflow-hidden shadow-2xs">
            <div className="p-4 border-b border-border flex items-center justify-between">
              <div>
                <h2 className="text-sm font-bold text-foreground tracking-tight">{t("recentOrders.title")}</h2>
                <p className="text-xs text-muted-foreground">{t("recentOrders.subtitle")}</p>
              </div>
              <Link
                href="/user/orders"
                className="text-xs font-medium text-accent-brand hover:underline"
              >
                {t("recentOrders.viewAll", { count: orders.length })}
              </Link>
            </div>

            {loading ? (
              <TableSkeleton />
            ) : error ? (
              <div className="p-4 border-destructive/20 bg-destructive/10 text-destructive text-sm">
                {error}
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left rtl:text-right">
                  <thead className="bg-secondary/40 border-b border-border">
                    <tr>
                      {localizedTbHeaders.map((th, idx) => (
                        <th
                          key={idx}
                          className="px-4 py-2.5 text-[11px] font-semibold text-muted-foreground uppercase tracking-wider"
                        >
                          {th}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    {recentOrders.length === 0 ? (
                      <tr>
                        <td colSpan={6} className="px-4 py-8 text-center text-xs text-muted-foreground">
                          {t("recentOrders.empty")}
                        </td>
                      </tr>
                    ) : (
                      recentOrders.map((order) => {
                        const badge = STATUS_BADGE[order.status] ?? STATUS_BADGE.pending;
                        const dateStr = new Date(order.createdAt).toLocaleDateString(locale, {
                          month: "short",
                          day: "numeric",
                          year: "numeric",
                        });
                        const itemsCount = order.items?.length ?? 0;
                        const itemsLabel = itemsCount === 1 
                          ? t("recentOrders.itemCount", { count: itemsCount })
                          : t("recentOrders.itemCountPlural", { count: itemsCount });

                        return (
                          <tr
                            key={order.id}
                            className="text-xs hover:bg-secondary/30 transition-colors"
                          >
                            <td className="px-4 py-3 text-foreground font-semibold font-mono">
                              #{order.id.slice(0, 8).toUpperCase()}
                            </td>
                            <td className="px-4 py-3 text-muted-foreground">
                              {dateStr}
                            </td>
                            <td className="px-4 py-3 text-foreground">
                              {itemsLabel}
                            </td>
                            <td className="px-4 py-3 text-foreground font-semibold">
                              {(order.totalAmount / 100).toFixed(2)} {order.currency.toUpperCase()}
                            </td>
                            <td className="px-4 py-3">
                              <span
                                className={`px-2 py-0.5 rounded-full text-[11px] font-medium ${badge.className}`}
                              >
                                {badge.label}
                              </span>
                            </td>
                            <td className="px-4 py-3">
                              <Link
                                href={`/user/orders/${order.id}`}
                                className="text-accent-brand hover:underline font-medium"
                              >
                                {t("recentOrders.details")}
                              </Link>
                            </td>
                          </tr>
                        );
                      })
                    )}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>

        {/* Right Column (4 cols): Quick Direct Access, Security & Details Panel */}
        <div className="lg:col-span-4 flex flex-col gap-4">
          {/* Account Profile Summary Card */}
          <div className="bg-card border border-border rounded-xl p-4.5 shadow-2xs">
            <h3 className="text-xs font-semibold text-foreground uppercase tracking-wider mb-3">
              {t("accountOverview.title")}
            </h3>
            <div className="space-y-2.5 text-xs">
              <div className="flex items-center justify-between pb-2 border-b border-border/60">
                <span className="text-muted-foreground">{t("accountOverview.name")}</span>
                <span className="font-medium text-foreground">{user?.name || t("defaultUser")}</span>
              </div>
              <div className="flex items-center justify-between pb-2 border-b border-border/60">
                <span className="text-muted-foreground">{t("accountOverview.email")}</span>
                <span className="font-mono text-foreground truncate max-w-[180px]">{user?.email || "—"}</span>
              </div>
              <div className="flex items-center justify-between pb-2 border-b border-border/60">
                <span className="text-muted-foreground">{t("accountOverview.security")}</span>
                <span className="text-emerald-600 dark:text-emerald-400 font-medium">{t("accountOverview.protected")}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">{t("accountOverview.reviews")}</span>
                <span className="font-medium text-foreground">{t("accountOverview.reviewsPublished", { count: reviewCount })}</span>
              </div>
            </div>

            <div className="mt-4 pt-3 border-t border-border flex items-center justify-between">
              <Link
                href="/user/settings"
                className="text-xs text-accent-brand hover:underline font-medium"
              >
                {t("accountOverview.editSettings")}
              </Link>
            </div>
          </div>

          {/* Quick Shortcuts Side Panel */}
          <div className="bg-card border border-border rounded-xl p-4.5 shadow-2xs">
            <h3 className="text-xs font-semibold text-foreground uppercase tracking-wider mb-3">
              {t("shortcuts.title")}
            </h3>
            <div className="flex flex-col divide-y divide-border/60 text-xs">
              <Link
                href="/user/orders"
                className="py-2.5 flex items-center justify-between text-foreground hover:text-accent-brand transition-colors font-medium group"
              >
                <span>{t("shortcuts.trackPackages")}</span>
                <span className="text-muted-foreground group-hover:text-accent-brand transition-transform rtl:rotate-180">→</span>
              </Link>
              <Link
                href="/user/address"
                className="py-2.5 flex items-center justify-between text-foreground hover:text-accent-brand transition-colors font-medium group"
              >
                <span>{t("shortcuts.manageAddresses")}</span>
                <span className="text-muted-foreground group-hover:text-accent-brand transition-transform rtl:rotate-180">→</span>
              </Link>
              <Link
                href="/user/wishlist"
                className="py-2.5 flex items-center justify-between text-foreground hover:text-accent-brand transition-colors font-medium group"
              >
                <span>{t("shortcuts.viewWishlist", { count: wishlistCount })}</span>
                <span className="text-muted-foreground group-hover:text-accent-brand transition-transform rtl:rotate-180">→</span>
              </Link>
              <Link
                href="/user/payment"
                className="py-2.5 flex items-center justify-between text-foreground hover:text-accent-brand transition-colors font-medium group"
              >
                <span>{t("shortcuts.billingMethods")}</span>
                <span className="text-muted-foreground group-hover:text-accent-brand transition-transform rtl:rotate-180">→</span>
              </Link>
              <Link
                href="/user/reviews"
                className="py-2.5 flex items-center justify-between text-foreground hover:text-accent-brand transition-colors font-medium group"
              >
                <span>{t("shortcuts.myReviews")}</span>
                <span className="text-muted-foreground group-hover:text-accent-brand transition-transform rtl:rotate-180">→</span>
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
