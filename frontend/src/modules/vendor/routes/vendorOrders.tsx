"use client";

import { CircleAlert } from "lucide-react";
import { useTranslations } from "next-intl";
import { useVendorContext } from "./vendorContext";

export function VendorOrders() {
  const t = useTranslations("vendor");
  const { vendor } = useVendorContext();

  return (
    <div className="space-y-6 w-full">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("orders.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t("orders.subtitle", { store: vendor?.storeName || t("orders.storeFallback") })}
          </p>
        </div>
      </div>

      <div className="rounded-2xl border border-amber-500/30 bg-amber-500/10 p-6 text-center shadow-xs">
        <CircleAlert className="mx-auto mb-3 size-7 text-amber-600 dark:text-amber-400" />
        <h2 className="font-bold text-foreground">{t("orders.unavailableTitle")}</h2>
        <p className="mx-auto mt-2 max-w-xl text-sm leading-relaxed text-muted-foreground">
          {t("orders.unavailableDescription")}
        </p>
        <p className="mx-auto mt-3 max-w-xl rounded-xl border border-border bg-background px-4 py-3 text-xs leading-relaxed text-muted-foreground">
          {t("orders.noFulfillmentChanges")}
        </p>
      </div>
    </div>
  );
}
