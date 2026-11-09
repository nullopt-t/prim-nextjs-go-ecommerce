"use client";

import { useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { Search, Package } from "lucide-react";
import { Link } from "@/shared/i18n/navigation";
import { orderService } from "@/modules/orders/api/orderService";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { formatMoney } from "@/shared/lib/money";
import type { Order } from "@/modules/orders/types";
import { isSafeUrl } from "@/shared/lib/safeUrl";
import { OrderTimelineHistory } from "./OrderTimelineHistory";

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function parseSnapshotTitle(snapshot?: string): string {
  if (!snapshot) return "";
  try {
    const parsed = JSON.parse(snapshot) as { title?: string; name?: string };
    return parsed.title || parsed.name || "";
  } catch {
    return "";
  }
}

/**
 * Guest order lookup: email + order number against the public
 * `GET /api/v1/orders/lookup` contract (no session). Renders the same
 * Order shape used by the authenticated account views.
 */
export default function OrderLookupContent() {
  const t = useTranslations("orderLookup");
  const tOrders = useTranslations("orders");
  const tDetails = useTranslations("orderDetails");
  const locale = useLocale();

  const [email, setEmail] = useState("");
  const [orderNumber, setOrderNumber] = useState("");
  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(false);
  const [fieldError, setFieldError] = useState<{ email?: string; orderNumber?: string }>({});
  const [lookupError, setLookupError] = useState<string | null>(null);
  const [searched, setSearched] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const errors: typeof fieldError = {};
    if (!EMAIL_RE.test(email.trim())) errors.email = t("emailInvalid");
    if (!UUID_RE.test(orderNumber.trim())) errors.orderNumber = t("orderNumberInvalid");
    setFieldError(errors);
    if (Object.keys(errors).length > 0) return;

    setLoading(true);
    setSearched(true);
    setLookupError(null);
    setOrder(null);
    try {
      const res = await orderService.lookupOrder(email.trim().toLowerCase(), orderNumber.trim());
      setOrder((res?.data ?? null) as Order | null);
      if (!res?.data) setLookupError(t("notFound"));
    } catch (err) {
      setLookupError(apiErrorMessage(err, t("lookupFailed")));
      setOrder(null);
    } finally {
      setLoading(false);
    }
  };

  function placedDateLabel(orderCreatedAt: string): string {
    const raw = new Date(orderCreatedAt);
    return Number.isNaN(raw.getTime())
      ? "—"
      : raw.toLocaleDateString(locale, { year: "numeric", month: "long", day: "numeric" });
  }
  const statusLabel = (status: Order["status"]) => tOrders(`status.${status}`);

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 py-4 sm:py-6">
      <header>
        <h1 className="text-3xl font-bold text-foreground md:text-4xl">{t("title")}</h1>
        <p className="mt-2 text-base text-muted-foreground">{t("subtitle")}</p>
      </header>

      <form
        onSubmit={handleSubmit}
        className="flex flex-col gap-4 rounded-2xl border border-border bg-card p-5 shadow-sm sm:p-6"
      >
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div className="flex flex-col gap-1.5">
            <label htmlFor="lookup-email" className="text-sm font-semibold text-foreground">
              {t("emailLabel")}
            </label>
            <input
              id="lookup-email"
              type="email"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="customer@example.com"
              className="rounded-lg border border-border bg-background px-3.5 py-2.5 text-sm text-foreground placeholder:text-muted-foreground focus:border-accent-brand focus:outline-none"
            />
            {fieldError.email && <p className="text-xs text-destructive">{fieldError.email}</p>}
          </div>
          <div className="flex flex-col gap-1.5">
            <label htmlFor="lookup-order-number" className="text-sm font-semibold text-foreground">
              {t("orderNumberLabel")}
            </label>
            <input
              id="lookup-order-number"
              type="text"
              autoComplete="off"
              value={orderNumber}
              onChange={(e) => setOrderNumber(e.target.value)}
              placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
              dir="ltr"
              className="rounded-lg border border-border bg-background px-3.5 py-2.5 text-sm text-foreground placeholder:text-muted-foreground focus:border-accent-brand focus:outline-none"
            />
            {fieldError.orderNumber && <p className="text-xs text-destructive">{fieldError.orderNumber}</p>}
          </div>
        </div>
        <button
          type="submit"
          disabled={loading}
          className="inline-flex items-center justify-center gap-2 rounded-lg bg-accent-brand px-4 py-3 text-sm font-semibold text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
        >
          <Search className="size-4" aria-hidden="true" />
          {loading ? t("looking") : t("lookup")}
        </button>
      </form>

      {lookupError && !loading && (
        <div role="alert" className="flex items-start gap-2 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          <span className="mt-0.5 size-4 shrink-0 rounded-full border border-destructive/40" aria-hidden="true" />
          <div>
            <p className="font-semibold">{lookupError}</p>
            {!order && <p className="mt-1 text-xs opacity-80">{t("notFoundDesc")}</p>}
          </div>
        </div>
      )}

      {searched && !loading && !order && (
        <div className="flex flex-col items-center gap-3 rounded-2xl border border-border bg-card p-10 text-center shadow-2xs">
          <Package className="size-10 opacity-50 text-muted-foreground" aria-hidden="true" />
          <p className="text-sm font-semibold text-foreground">{t("notFound")}</p>
          <p className="max-w-sm text-sm text-muted-foreground">{t("notFoundDesc")}</p>
        </div>
      )}

      {order && (
        <section className="flex flex-col gap-4 rounded-2xl border border-border bg-card p-5 shadow-sm sm:p-6">
          <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border/70 pb-4">
            <h2 className="text-lg font-bold text-foreground">{tDetails("orderTitle", { id: order.id })}</h2>
            <span className="rounded-full bg-secondary px-3 py-1 text-xs font-semibold text-foreground">
              {statusLabel(order.status)}
            </span>
          </div>

          <div className="grid grid-cols-1 gap-3 text-sm sm:grid-cols-3">
            <div>
              <p className="text-muted-foreground">{tDetails("placedOn", { date: placedDateLabel(order.createdAt) })}</p>
            </div>
            <div>
              <p className="text-muted-foreground">{tOrders("itemCount", { count: order.items?.length ?? 0 })}</p>
            </div>
            <div className="sm:text-end">
              <p className="font-bold text-foreground">
                {tDetails("total")}: {formatMoney((order.totalAmount ?? 0) / 100, order.currency)}
              </p>
            </div>
          </div>

          {(order.trackingNumber || order.trackingUrl) && (
            <div className="flex flex-col gap-1 rounded-xl bg-secondary/50 p-4 text-sm">
              <p className="text-muted-foreground">{tDetails("trackingNumber")}: <span className="font-semibold text-foreground">{order.trackingNumber ?? "—"}</span></p>
              {order.carrier && <p className="text-muted-foreground">{tDetails("carrier")}: <span className="font-semibold text-foreground">{order.carrier}</span></p>}
              {isSafeUrl(order.trackingUrl) && (
                <a href={order.trackingUrl} target="_blank" rel="noreferrer" className="font-medium text-accent-brand hover:underline">
                  {tDetails("trackShipment")}
                </a>
              )}
            </div>
          )}

          <OrderTimelineHistory history={order.statusHistory} createdAt={order.createdAt} />

          <div className="divide-y divide-border/70">
            {(order.items ?? []).map((item) => {
              const snapshotTitle = parseSnapshotTitle(item.productSnapshot);
              const displayName = item.title || snapshotTitle || t("deletedProduct", { defaultMessage: "Item" }) + ` (#${(item.variantId ?? "").slice(0, 8)})`;
              return (
                <div key={item.id} className="flex items-center justify-between gap-4 py-3 text-sm">
                  <div className="min-w-0">
                    <p className="truncate font-semibold text-foreground">{displayName}</p>
                    <p className="text-xs text-muted-foreground">{tDetails("quantity", { count: item.quantity })}</p>
                  </div>
                  <p className="shrink-0 font-semibold text-foreground">{formatMoney(((item.priceAtPurchase || 0) * (item.quantity || 1)) / 100, order.currency)}</p>
                </div>
              );
            })}
          </div>

          <Link href="/products" className="text-sm font-medium text-accent-brand hover:underline">
            {t("continueShopping")}
          </Link>
        </section>
      )}
    </div>
  );
}
