"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { Link } from "@/shared/i18n/navigation";
import { userService } from "@/modules/account/api/userService";
import { LazyImage } from "@/shared/ui/lazyImage";
import { normalizeMediaUrl } from "@/shared/lib/media";
import { useTranslations, useLocale } from "next-intl";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { Check, Clock, Truck, Package, ExternalLink, XCircle, Download } from "lucide-react";
import InvoicePrintView from "./InvoicePrintView";
import { isSafeUrl } from "@/shared/lib/safeUrl";

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
  status: "pending" | "processing" | "shipped" | "delivered" | "cancelled" | "refunded";
  totalAmount: number;
  discountAmount: number;
  shippingAmount?: number;
  taxAmount?: number;
  codFee?: number;
  couponCode?: string | null;
  couponId?: string | null;
  currency: string;
  carrier?: string | null;
  trackingNumber?: string | null;
  trackingUrl?: string | null;
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
  sellerName?: string;
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
  refunded: {
    className: "text-zinc-600 bg-zinc-100 dark:bg-zinc-800/50 dark:text-zinc-300 border border-zinc-300/50",
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
  const [cancelling, setCancelling] = useState(false);

  const handleCancelOrder = async () => {
    if (!order || !id) return;
    setCancelling(true);
    try {
      await userService.cancelOrder(String(id));
      setOrder({ ...order, status: "cancelled" });
      notify.success(t("cancelSuccess"));
    } catch (err: unknown) {
      notify.error(apiErrorMessage(err, t("cancelFailed")));
    } finally {
      setCancelling(false);
    }
  };

  useEffect(() => {
    let cancelled = false;

    async function load() {
      if (!id) return;
      setLoading(true);
      setError(null);
      try {
        const res = await userService.getOrderById(String(id));
        if (cancelled) return;
        const data = (res?.data ?? res) as Order;
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
  const shortId = order.id?.slice(0, 8).toUpperCase() ?? "";
  const placedDateRaw = new Date(order.createdAt);
  const placedDate = Number.isNaN(placedDateRaw.getTime())
    ? "—"
    : placedDateRaw.toLocaleDateString(locale, {
        year: "numeric",
        month: "long",
        day: "numeric",
      });
  const itemsSubtotalCents = items.reduce(
    (acc, it) => acc + (it.priceAtPurchase || 0) * (it.quantity || 1),
    0
  );
  const rawSubtotal = itemsSubtotalCents > 0
    ? (itemsSubtotalCents / 100).toFixed(2)
    : (((order.totalAmount ?? 0) + (order.discountAmount ?? 0) - (order.shippingAmount ?? 0) - (order.taxAmount ?? 0) - (order.codFee ?? 0)) / 100).toFixed(2);
  const discountVal = (order.discountAmount ?? 0) / 100;
  const formattedDiscount = discountVal > 0 ? discountVal.toFixed(2) : null;
  const shippingVal = (order.shippingAmount ?? 0) / 100;
  const formattedShipping = shippingVal > 0 ? shippingVal.toFixed(2) : null;
  const taxVal = (order.taxAmount ?? 0) / 100;
  const formattedTax = taxVal > 0 ? taxVal.toFixed(2) : null;
  const codVal = (order.codFee ?? 0) / 100;
  const formattedCod = codVal > 0 ? codVal.toFixed(2) : null;
  const totalPaid = ((order.totalAmount ?? 0) / 100).toFixed(2);
  const formattedAddress = formatAddress(order.shippingAddress);

  const statusStepMap: Record<Order["status"], number> = {
    pending: 0,
    processing: 1,
    shipped: 2,
    delivered: 3,
    cancelled: -1,
    refunded: -1,
  };
  const currentStep = statusStepMap[order.status] ?? 0;

  const steps = [
    { key: "placed", label: t("timeline.placed"), icon: Clock },
    { key: "processing", label: t("timeline.processing"), icon: Package },
    { key: "shipped", label: t("timeline.shipped"), icon: Truck },
    { key: "delivered", label: t("timeline.delivered"), icon: Check },
  ];

  return (
    <>
    <div className="w-full flex flex-col gap-6 pb-12 no-print">
      {/* ── Top Bar ── */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/70 pb-5">
        <div>
          <Link
            href="/user/orders"
            className="text-xs text-muted-foreground hover:text-accent-brand transition-colors mb-2 inline-flex items-center gap-1 group"
          >
            <span className="rtl:rotate-180 group-hover:-translate-x-0.5 rtl:group-hover:translate-x-0.5 transition-transform" aria-hidden>←</span>
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
          {order.status === "pending" && (
            <button
              type="button"
              onClick={handleCancelOrder}
              disabled={cancelling}
              className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-red-500/10 text-red-600 dark:text-red-400 hover:bg-red-500/20 border border-red-500/20 transition-colors disabled:opacity-50"
            >
              {cancelling ? t("cancelling") : t("cancelOrder")}
            </button>
          )}
          <button
            type="button"
            onClick={() => window.print()}
            className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-xs font-semibold border border-border bg-background text-foreground hover:bg-secondary transition-colors shadow-xs"
          >
            <Download className="size-3.5" aria-hidden="true" />
            {t("downloadInvoice")}
          </button>
          <Link
            href="/products"
            className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-accent-brand text-white hover:opacity-90 transition-opacity shadow-xs"
          >
            {t("buyAgain")}
          </Link>
        </div>
      </div>

      {/* ── Status Progress Stepper & Tracking ── */}
      {order.status === "cancelled" ? (
        <div className="flex items-center gap-3 p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-600 dark:text-red-400">
          <XCircle className="size-5 shrink-0" />
          <div className="text-xs">
            <span className="font-bold">{t("timeline.cancelled")}</span>
          </div>
        </div>
      ) : (
        <div className="p-5 rounded-2xl bg-card border border-border/70 shadow-2xs">
          <div className="relative flex items-center justify-between">
            <div className="absolute top-1/2 -translate-y-1/2 left-6 right-6 h-1 bg-secondary rounded-full -z-0" />
            <div
              className="absolute top-1/2 -translate-y-1/2 left-6 h-1 bg-accent-brand rounded-full transition-all duration-500 -z-0"
              style={{ width: `${Math.min(100, Math.max(0, (currentStep / (steps.length - 1)) * 100))}%` }}
            />
            {steps.map((st, idx) => {
              const Icon = st.icon;
              const isDone = idx < currentStep;
              const isCurrent = idx === currentStep;
              return (
                <div key={st.key} className="flex flex-col items-center gap-2 z-10">
                  <div
                    className={`size-9 rounded-full flex items-center justify-center transition-colors text-xs font-bold ${
                      isDone
                        ? "bg-accent-brand text-white shadow-xs"
                        : isCurrent
                        ? "bg-accent-brand text-white ring-4 ring-accent-brand/20 shadow-xs"
                        : "bg-secondary text-muted-foreground border border-border"
                    }`}
                  >
                    {isDone ? <Check className="size-4" /> : <Icon className="size-4" />}
                  </div>
                  <span
                    className={`text-[11px] sm:text-xs font-medium text-center ${
                      isDone || isCurrent ? "text-foreground font-semibold" : "text-muted-foreground"
                    }`}
                  >
                    {st.label}
                  </span>
                </div>
              );
            })}
          </div>

          {(order.carrier || order.trackingNumber || order.trackingUrl) && (
            <div className="mt-5 pt-4 border-t border-border/70 flex flex-wrap items-center justify-between gap-3 text-xs bg-secondary/30 -mx-5 -mb-5 p-4 rounded-b-2xl">
              <div className="flex flex-wrap items-center gap-3">
                <div className="flex items-center gap-1.5 text-muted-foreground">
                  <Truck className="size-4 text-accent-brand" />
                  <span className="font-semibold text-foreground">{t("tracking")}:</span>
                </div>
                {order.carrier && (
                  <span className="px-2 py-0.5 rounded-md bg-secondary text-foreground font-medium text-[11px] border border-border">
                    {order.carrier}
                  </span>
                )}
                {order.trackingNumber && (
                  <span className="font-mono text-muted-foreground text-[11px]">
                    {order.trackingNumber}
                  </span>
                )}
              </div>
              {isSafeUrl(order.trackingUrl) && (
                <a
                  href={order.trackingUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg bg-accent-brand text-white font-medium hover:opacity-90 transition-opacity text-xs"
                >
                  <span>{t("trackShipment")}</span>
                  <ExternalLink className="size-3.5" />
                </a>
              )}
            </div>
          )}
        </div>
      )}

      {/* ── Two-Column Details Layout ── */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
        {/* Left Column (8 cols): Line Items List */}
        <div className="lg:col-span-8 flex flex-col gap-3.5">
          <h2 className="text-sm font-bold text-foreground">
            {t("itemsTitle", { count: items.length })}
          </h2>

          <div className="flex flex-col gap-2.5">
            {items.map((item) => {
              const snap = parseSnapshot(item.productSnapshot);
              const title = snap.title || snap.name || "Purchased Product";
              const imgSrc = resolveItemImage(snap);
              const itemPrice = `${(item.priceAtPurchase / 100).toFixed(2)} ${(order.currency ?? "").toUpperCase()}`;
              const productTarget = snap.slug || snap.productId;
              const productHref = productTarget ? `/products/${productTarget}` : null;

              const imageBlock = (
                <div className="size-11 rounded-lg bg-background border border-border/60 overflow-hidden shrink-0 flex items-center justify-center p-1 shadow-2xs hover:scale-105 transition-transform duration-200">
                  {imgSrc ? (
                    <LazyImage
                      src={imgSrc}
                      alt={title}
                      containerClassName="w-full h-full"
                      imageClassName="w-full h-full object-contain"
                    />
                  ) : (
                    <span className="text-[10px] text-muted-foreground font-medium">
                      {t("item")}
                    </span>
                  )}
                </div>
              );

              return (
                <div
                  key={item.id}
                  className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3 sm:p-3.5 rounded-xl bg-secondary/30 border border-border/50 hover:bg-secondary/40 transition-colors shadow-2xs"
                >
                  <div className="flex items-center gap-3 min-w-0 flex-1">
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
                          className="text-xs sm:text-sm font-semibold text-foreground hover:text-accent-brand transition-colors line-clamp-1 block focus:outline-none"
                        >
                          {title}
                        </Link>
                      ) : (
                        <p className="text-xs sm:text-sm font-semibold text-foreground line-clamp-1">{title}</p>
                      )}
                      <div className="flex items-center gap-2 text-[11px] text-muted-foreground mt-0.5">
                        <span>{t("quantity", { count: item.quantity })}</span>
                        <span>•</span>
                        <span className="font-bold text-foreground text-xs">{itemPrice}</span>
                        {snap.sellerName && (
                          <>
                            <span>•</span>
                            <span className="font-medium text-accent-brand">
                              {t("soldBy", { seller: snap.sellerName })}
                            </span>
                          </>
                        )}
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 shrink-0 self-end sm:self-center">
                    <Link
                      href="/products"
                      className="px-2.5 py-1 rounded-md text-xs font-medium bg-background hover:bg-secondary text-foreground border border-border/70 transition-colors shadow-2xs"
                    >
                      {t("buyAgain")}
                    </Link>
                    <Link
                      href="/user/reviews"
                      className="px-2.5 py-1 rounded-md text-xs font-medium bg-background hover:bg-secondary text-foreground border border-border/70 transition-colors shadow-2xs"
                    >
                      {t("writeReview")}
                    </Link>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Right Column (4 cols): Delivery Destination & Order Summary */}
        <div className="lg:col-span-4 flex flex-col gap-3.5">
          {/* Order Summary */}
          <div className="bg-secondary/30 border border-border/50 rounded-xl p-4 shadow-2xs">
            <h3 className="text-xs font-semibold text-foreground uppercase tracking-wider mb-3">
              {t("paymentSummary")}
            </h3>
            <div className="space-y-2 text-xs">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">{t("subtotal")}</span>
                <span className="font-medium text-foreground">{rawSubtotal} {(order.currency ?? "").toUpperCase()}</span>
              </div>
              {formattedDiscount && (
                <div className="flex items-center justify-between text-emerald-600 dark:text-emerald-400">
                  <div className="flex items-center gap-1.5">
                    <span>{t("discount")}</span>
                    {order.couponCode && (
                      <span className="px-1.5 py-0.5 rounded text-[10px] font-mono font-semibold bg-emerald-500/10 border border-emerald-500/20">
                        {order.couponCode}
                      </span>
                    )}
                  </div>
                  <span className="font-medium font-mono">
                    -{formattedDiscount} {(order.currency ?? "").toUpperCase()}
                  </span>
                </div>
              )}
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">{t("shipping")}</span>
                {formattedShipping ? (
                  <span className="font-medium text-foreground">{formattedShipping} {(order.currency ?? "").toUpperCase()}</span>
                ) : (
                  <span className="text-emerald-600 dark:text-emerald-400 font-medium">{t("freeShipping")}</span>
                )}
              </div>
              {formattedTax && (
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">{t("invoice.tax")}</span>
                  <span className="font-medium text-foreground">{formattedTax} {(order.currency ?? "").toUpperCase()}</span>
                </div>
              )}
              {formattedCod && (
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">{t("invoice.codFee")}</span>
                  <span className="font-medium text-foreground">{formattedCod} {(order.currency ?? "").toUpperCase()}</span>
                </div>
              )}
              <div className="border-t border-border/50 pt-2.5 mt-2.5 flex items-center justify-between font-bold text-xs sm:text-sm">
                <span className="text-foreground">{t("total")}</span>
                <span className="text-foreground font-mono">{totalPaid} {(order.currency ?? "").toUpperCase()}</span>
              </div>
            </div>
          </div>

          {/* Delivery Details */}
          <div className="bg-secondary/30 border border-border/50 rounded-xl p-4 shadow-2xs">
            <h3 className="text-xs font-semibold text-foreground uppercase tracking-wider mb-2.5">
              {t("deliveryAddress")}
            </h3>
            <div className="space-y-1 text-xs">
              {formattedAddress ? (
                <p className="text-foreground/90 font-medium leading-relaxed">{formattedAddress}</p>
              ) : (
                <p className="text-muted-foreground">—</p>
              )}
            </div>

            <div className="border-t border-border/50 mt-3 pt-2.5">
              <h4 className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider mb-0.5">
                {t("customer")}
              </h4>
              <p className="text-xs font-mono text-foreground truncate">{order.customerEmail}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
    <InvoicePrintView order={order} />
    </>
  );
}
