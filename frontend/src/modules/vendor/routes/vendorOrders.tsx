"use client";

import { useEffect, useState } from "react";
import { Loader2, ShoppingBag } from "lucide-react";
import { vendorService } from "@/modules/vendor/api/vendorService";
import type { VendorOrder } from "@/modules/vendor/types";
import { ErrorState } from "@/shared/ui/errorState";
import { formatDateTime } from "@/shared/lib/datetime";
import { formatMinor } from "./vendorDashboard";
import { useTranslations, useLocale } from "next-intl";
import { useVendorContext } from "./vendorContext";

const ORDER_STATUS_KEYS: Record<string, string> = {
  pending: "status.pending",
  paid: "status.paid",
  processing: "status.processing",
  shipped: "status.shipped",
  delivered: "status.delivered",
  cancelled: "status.cancelled",
  refunded: "status.refunded",
};

export function VendorOrders() {
  const t = useTranslations("vendor");
  const locale = useLocale();
  const { vendor } = useVendorContext();

  const [orders, setOrders] = useState<VendorOrder[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<Error | null>(null);

  const loadOrders = async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const response = await vendorService.getOrders({ pageSize: 100 });
      setOrders(response.data);
    } catch (error) {
      setLoadError(error instanceof Error ? error : new Error(String(error)));
      setOrders(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadOrders();
  }, []);

  const statusLabel = (status: string) => {
    const key = ORDER_STATUS_KEYS[status];
    return key ? t(key) : status;
  };

  const header = (
    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
      <div>
        <h1 className="text-2xl font-bold text-foreground">{t("orders.title")}</h1>
        <p className="text-sm text-muted-foreground mt-1">
          {t("orders.subtitle", { store: vendor?.storeName || t("orders.storeFallback") })}
        </p>
      </div>
    </div>
  );

  if (loading) {
    return (
      <div className="space-y-6 w-full">
        {header}
        <div className="flex min-h-[50vh] items-center justify-center">
          <Loader2 className="size-8 animate-spin text-accent-brand" />
        </div>
      </div>
    );
  }

  if (loadError || !orders) {
    return (
      <div className="space-y-6 w-full">
        {header}
        <ErrorState
          title={t("orders.loadTitle")}
          description={loadError?.message || t("orders.loadDescription")}
          onRetry={() => void loadOrders()}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6 w-full">
      {header}

      <p className="text-xs text-muted-foreground">{t("orders.noFulfillmentChanges")}</p>

      {orders.length === 0 ? (
        <div className="bg-background border border-border rounded-2xl p-10 text-center shadow-xs">
          <ShoppingBag className="mx-auto mb-3 size-7 text-muted-foreground" />
          <h2 className="font-bold text-foreground">{t("orders.empty")}</h2>
        </div>
      ) : (
        <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
          <div className="p-6 border-b border-border">
            <h3 className="font-bold text-base text-foreground">{t("orders.historyTitle")}</h3>
            <p className="text-xs text-muted-foreground mt-0.5">{t("orders.historySubtitle")}</p>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full min-w-max text-left text-xs">
              <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
                <tr>
                  <th className="py-3.5 pl-6">{t("orders.table.orderRef")}</th>
                  <th className="py-3.5">{t("orders.table.date")}</th>
                  <th className="py-3.5">{t("orders.table.fulfillmentStatus")}</th>
                  <th className="py-3.5">{t("orders.table.item")}</th>
                  <th className="py-3.5 text-right">{t("orders.table.grossRevenue")}</th>
                  <th className="py-3.5 pr-6 text-right">{t("orders.table.netEarnings")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {orders.map((order) => (
                  <tr key={order.id} className="hover:bg-secondary/20 transition-colors">
                    <td className="py-3.5 pl-6 font-mono font-bold text-foreground">{order.id}</td>
                    <td className="py-3.5 text-muted-foreground">{formatDateTime(order.createdAt, locale)}</td>
                    <td className="py-3.5">
                      <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-secondary text-foreground border border-border">
                        {statusLabel(order.status)}
                      </span>
                    </td>
                    <td className="py-3.5 text-muted-foreground">
                      <div className="flex flex-col gap-1">
                        {order.items.map((item) => (
                          <span key={item.id}>
                            {item.title || item.variantId} × {item.quantity}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="py-3.5 text-right font-bold text-foreground">
                      {formatMinor(order.subtotal, order.currency)}
                    </td>
                    <td className="py-3.5 pr-6 text-right font-bold text-foreground">
                      {formatMinor(order.earnings, order.currency)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
