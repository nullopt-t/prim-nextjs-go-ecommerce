"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { Link } from "@/i18n/navigation";
import { userService } from "@/services/user";
import { LazyImage } from "@/components/ui/lazyImage";
import { normalizeMediaUrl } from "@/utils/media";
import { useTranslations, useLocale } from "next-intl";

interface OrderItem {
  id: string;
  variantId: string;
  quantity: number;
  priceAtPurchase: number;
  productSnapshot: string;
}

interface Address {
  street?: string;
  city?: string;
  state?: string;
  postalCode?: string;
  country?: string;
}

interface Order {
  id: string;
  customerEmail: string;
  shippingAddress?: Address;
  billingAddress?: Address;
  status: "pending" | "processing" | "shipped" | "delivered" | "cancelled";
  totalAmount: number;
  discountAmount: number;
  currency: string;
  items: OrderItem[];
  createdAt: string;
  updatedAt: string;
}

interface ParsedSnapshot {
  title?: string;
  name?: string;
  sku?: string;
  productId?: string;
  slug?: string;
  img?: string;
  image?: string;
  thumbnail?: string;
  thumbnailUrl?: string;
  price?: number;
  attributes?: Record<string, any>;
}

function resolveItemImage(snap: ParsedSnapshot): string {
  const raw = snap.img || snap.image || snap.thumbnail || snap.thumbnailUrl;
  return normalizeMediaUrl(raw);
}

function parseSnapshot(snapshot?: string): ParsedSnapshot {
  if (!snapshot) return {};
  try {
    return JSON.parse(snapshot) ?? {};
  } catch {
    return {};
  }
}

function formatAddress(addr?: Address): string {
  if (!addr) return "";
  const parts = [addr.street, addr.city, addr.state, addr.country].filter(Boolean);
  return parts.join(", ");
}

const STATUS_BADGE: Record<
  Order["status"],
  { className: string }
> = {
  pending: {
    className: "text-amber-600 bg-amber-50 dark:bg-amber-950/30 dark:text-amber-400 border border-amber-200/50",
  },
  processing: {
    className: "text-blue-600 bg-blue-50 dark:bg-blue-950/30 dark:text-blue-400 border border-blue-200/50",
  },
  shipped: {
    className: "text-violet-600 bg-violet-50 dark:bg-violet-950/30 dark:text-violet-400 border border-violet-200/50",
  },
  delivered: {
    className: "text-emerald-600 bg-emerald-50 dark:bg-emerald-950/30 dark:text-emerald-400 border border-emerald-200/50",
  },
  cancelled: {
    className: "text-red-600 bg-red-50 dark:bg-red-950/30 dark:text-red-400 border border-red-200/50",
  },
};

export default function OrderDetailsContent() {
  const params = useParams();
  const id = Array.isArray(params?.id) ? params.id[0] : params?.id;
  const t = useTranslations("orderDetails");
  const tOrders = useTranslations("orders");
  const locale = useLocale();

  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      if (!id) return;
      setLoading(true);
      setError(null);
      try {
        const res = await userService.getOrderById(String(id));
        if (cancelled) return;
        const data = res?.data ?? res;
        setOrder(data);
      } catch (err: unknown) {
        if (!cancelled) setError(err instanceof Error ? err.message : t("error"));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, [id, t]);

  if (loading) {
    return (
      <div className="w-full flex flex-col gap-6 py-6 animate-pulse">
        <div className="h-6 w-32 bg-secondary rounded" />
        <div className="h-10 w-64 bg-secondary rounded" />
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          <div className="lg:col-span-8 h-96 bg-secondary/50 rounded-xl" />
          <div className="lg:col-span-4 h-96 bg-secondary/50 rounded-xl" />
        </div>
      </div>
    );
  }

  if (error || !order) {
    return (
      <div className="w-full bg-card border border-border rounded-xl p-12 text-center flex flex-col items-center justify-center gap-3">
        <h2 className="text-lg font-bold text-foreground">{t("notFound")}</h2>
        <p className="text-xs text-muted-foreground max-w-sm">{error || t("notFoundDesc")}</p>
        <Link
          href="/user/orders"
          className="mt-3 px-4 py-2 text-xs font-semibold rounded-lg bg-accent-brand text-white hover:opacity-90 transition-opacity"
        >
          {t("backToOrders")}
        </Link>
      </div>
    );
  }

  const items = order.items ?? [];
  const statusBadge = STATUS_BADGE[order.status] ?? STATUS_BADGE.pending;
  const shortId = order.id.slice(0, 8).toUpperCase();
  const placedDate = new Date(order.createdAt).toLocaleDateString(locale, {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
  const subtotal = (order.totalAmount / 100).toFixed(2);
  const formattedAddress = formatAddress(order.shippingAddress);

  return (
    <div className="w-full flex flex-col gap-6 pb-12">
      {/* ── Top Bar ── */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/70 pb-5">
        <div>
          <Link
            href="/user/orders"
            className="text-xs text-muted-foreground hover:text-accent-brand transition-colors mb-2 inline-flex items-center gap-1 group"
          >
            <span className="rtl:rotate-180 group-hover:-translate-x-0.5 transition-transform">←</span>
            <span>{t("backToOrders")}</span>
          </Link>
          <div className="flex flex-wrap items-center gap-3 mt-1">
            <h1 className="text-2xl font-bold text-foreground tracking-tight">
              {t("orderTitle", { id: shortId })}
            </h1>
            <span className={`px-2.5 py-0.5 rounded-full text-xs font-semibold ${statusBadge.className}`}>
              {tOrders(`status.${order.status}`)}
            </span>
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            {t("placedOn", { date: placedDate })}
          </p>
        </div>

        <div className="flex items-center gap-2.5 shrink-0">
          <Link
            href="/products"
            className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-accent-brand text-white hover:opacity-90 transition-opacity shadow-xs"
          >
            {t("buyAgain")}
          </Link>
        </div>
      </div>

      {/* ── Two-Column Details Layout ── */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
        {/* Left Column (8 cols): Line Items List */}
        <div className="lg:col-span-8 flex flex-col gap-6">
          <div className="bg-card border border-border rounded-xl p-5 shadow-2xs">
            <h2 className="text-sm font-bold text-foreground mb-4">
              {t("itemsTitle", { count: items.length })}
            </h2>

            <div className="divide-y divide-border/60">
              {items.map((item) => {
                const snap = parseSnapshot(item.productSnapshot);
                const title = snap.title || snap.name || "Purchased Product";
                const imgSrc = resolveItemImage(snap);
                const itemPrice = `${(item.priceAtPurchase / 100).toFixed(2)} ${order.currency.toUpperCase()}`;
                const productTarget = snap.slug || snap.productId;
                const productHref = productTarget ? `/products/${productTarget}` : null;

                const imageBlock = (
                  <div className="size-16 rounded-lg border border-border/80 bg-secondary/30 overflow-hidden shrink-0 aspect-square flex items-center justify-center transition-transform hover:scale-105">
                    {imgSrc ? (
                      <LazyImage
                        src={imgSrc}
                        alt={title}
                        containerClassName="w-full h-full"
                        imageClassName="w-full h-full object-cover object-center"
                      />
                    ) : (
                      <span className="text-[10px] text-muted-foreground">Product</span>
                    )}
                  </div>
                );

                return (
                  <div
                    key={item.id}
                    className="py-4 first:pt-0 last:pb-0 flex flex-col sm:flex-row sm:items-center justify-between gap-4"
                  >
                    <div className="flex items-center gap-3.5 min-w-0 flex-1">
                      {productHref ? (
                        <Link href={productHref} className="shrink-0 focus:outline-none">
                          {imageBlock}
                        </Link>
                      ) : (
                        imageBlock
                      )}

                      <div className="min-w-0 flex-1">
                        {productHref ? (
                          <Link
                            href={productHref}
                            className="text-sm font-semibold text-foreground hover:text-accent-brand transition-colors line-clamp-1 block focus:outline-none"
                          >
                            {title}
                          </Link>
                        ) : (
                          <p className="text-sm font-semibold text-foreground line-clamp-1">{title}</p>
                        )}
                        <div className="flex items-center gap-3 text-xs text-muted-foreground mt-0.5">
                          <span>{t("quantity", { count: item.quantity })}</span>
                          <span>•</span>
                          <span className="font-semibold text-foreground">{itemPrice}</span>
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-2 shrink-0 self-end sm:self-center">
                      <Link
                        href="/products"
                        className="px-3 py-1.5 rounded-lg text-xs font-medium bg-secondary hover:bg-secondary/80 text-foreground border border-border transition-colors"
                      >
                        {t("buyAgain")}
                      </Link>
                      <Link
                        href="/user/reviews"
                        className="px-3 py-1.5 rounded-lg text-xs font-medium bg-secondary hover:bg-secondary/80 text-foreground border border-border transition-colors"
                      >
                        {t("writeReview")}
                      </Link>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Right Column (4 cols): Delivery Destination & Order Summary */}
        <div className="lg:col-span-4 flex flex-col gap-4">
          {/* Order Summary */}
          <div className="bg-card border border-border rounded-xl p-4.5 shadow-2xs">
            <h3 className="text-xs font-semibold text-foreground uppercase tracking-wider mb-3">
              {t("paymentSummary")}
            </h3>
            <div className="space-y-2 text-xs">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">{t("subtotal")}</span>
                <span className="font-medium text-foreground">{subtotal} {order.currency.toUpperCase()}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">{t("shipping")}</span>
                <span className="text-emerald-600 dark:text-emerald-400 font-medium">{t("freeShipping")}</span>
              </div>
              <div className="border-t border-border pt-2 mt-2 flex items-center justify-between font-bold text-sm">
                <span className="text-foreground">{t("total")}</span>
                <span className="text-foreground">{subtotal} {order.currency.toUpperCase()}</span>
              </div>
            </div>
          </div>

          {/* Delivery Details */}
          <div className="bg-card border border-border rounded-xl p-4.5 shadow-2xs">
            <h3 className="text-xs font-semibold text-foreground uppercase tracking-wider mb-3">
              {t("deliveryAddress")}
            </h3>
            <div className="space-y-1 text-xs">
              {formattedAddress ? (
                <p className="text-foreground leading-relaxed">{formattedAddress}</p>
              ) : (
                <p className="text-muted-foreground">—</p>
              )}
            </div>

            <div className="border-t border-border mt-3 pt-3">
              <h4 className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider mb-1">
                {t("customer")}
              </h4>
              <p className="text-xs font-mono text-foreground truncate">{order.customerEmail}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
