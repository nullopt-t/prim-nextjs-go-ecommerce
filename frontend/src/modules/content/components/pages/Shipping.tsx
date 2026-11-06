"use client";

import { useTranslations } from "next-intl";
import { usePublicSettings } from "@/modules/settings/model/publicSettingsContext";
import { formatMoney } from "@/shared/lib/money";

export function Shipping() {
  const t = useTranslations("shop");
  const { settings } = usePublicSettings();
  const currency = settings?.currency?.trim() ?? "";
  const days = settings?.estimatedDeliveryDays?.trim() ?? "—";
  const rate = settings ? formatMoney(settings.flatRateShipping, currency) : "—";
  const threshold = settings ? formatMoney(settings.freeShippingThreshold, currency) : "—";
  return (
  <div className="w-full flex flex-col gap-6 py-4 sm:py-6">
    <h1 className="text-3xl md:text-4xl font-bold text-foreground">
      {t("pages.shipping.title")}
    </h1>
    <div className="text-muted-foreground flex flex-col gap-4 text-base leading-relaxed">
      <p>{t("pages.shipping.intro")}</p>
      <h3 className="text-xl font-semibold text-foreground mt-4">{t("pages.shipping.standardTitle")}</h3>
      <p>{t("pages.shipping.standardDescription", { days, rate, threshold })}</p>
      <h3 className="text-xl font-semibold text-foreground mt-4">{t("pages.shipping.expressTitle")}</h3>
      <p>{t("pages.shipping.expressDescription")}</p>
    </div>
  </div>
  );
}
