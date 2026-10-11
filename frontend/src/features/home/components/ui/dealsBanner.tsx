"use client";

import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";
import {
  Tag,
  Truck,
  ShieldCheck,
  Zap,
  ArrowRight,
  Sparkles,
} from "lucide-react";

export default function DealsBanner() {
  const t = useTranslations("home.dealsBanner");

  return (
    <section className="w-full">
      <div className="relative overflow-hidden rounded-2xl bg-linear-to-r from-secondary/70 via-secondary/40 to-secondary/70 border border-border/70 p-4 sm:p-5 md:p-6 shadow-2xs">
        {/* Subtle decorative color accents */}
        <div className="pointer-events-none absolute -top-12 -left-12 size-40 rounded-full bg-accent-brand/15 blur-2xl" />
        <div className="pointer-events-none absolute -bottom-12 -right-12 size-40 rounded-full bg-sky-500/15 blur-2xl" />
        <div className="pointer-events-none absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-3/4 h-20 bg-amber-500/10 blur-3xl" />

        <div className="relative z-10 flex flex-col lg:flex-row items-center justify-between gap-5 md:gap-6">
          {/* Left: Lead Offer & Callout */}
          <div className="flex flex-col sm:flex-row items-center sm:items-start text-center sm:text-left rtl:sm:text-right gap-3.5 sm:gap-4 max-w-xl">
            <div className="size-12 sm:size-14 rounded-2xl bg-accent-brand/10 text-accent-brand border border-accent-brand/20 flex items-center justify-center shrink-0 shadow-2xs">
              <Zap className="size-6 sm:size-7 fill-accent-brand/20" />
            </div>

            <div className="flex flex-col">
              <div className="inline-flex items-center justify-center sm:justify-start gap-1.5 mb-1">
                <span className="px-2 py-0.5 rounded-full bg-red-500/10 text-red-600 dark:text-red-400 border border-red-500/20 text-[10px] font-bold uppercase tracking-wider">
                  {t("badge")}
                </span>
                <span className="text-xs font-semibold text-muted-foreground flex items-center gap-1">
                  <Sparkles className="size-3 text-amber-500" />
                  {t("limitedTime")}
                </span>
              </div>

              <h2 className="text-lg sm:text-xl md:text-2xl font-black text-foreground tracking-tight leading-snug">
                {t("headline")}
              </h2>

              <p className="text-xs sm:text-sm text-muted-foreground mt-1 max-w-md">
                {t("subtext")}
              </p>
            </div>
          </div>

          {/* Middle: Key Trust Guarantees (Fast Delivery, Quality Guarantee, Money Back) */}
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-2.5 sm:gap-3 w-full lg:w-auto">
            {/* Coupon tile */}
            <div className="flex items-center gap-2.5 p-2.5 rounded-xl bg-background/80 border border-border/60 shadow-2xs backdrop-blur-xs">
              <div className="size-8 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center shrink-0">
                <Tag className="size-4" />
              </div>
              <div className="flex flex-col min-w-0">
                <span className="text-xs font-bold text-foreground leading-tight truncate">
                  {t("couponTitle")}
                </span>
                <span className="text-[10px] text-muted-foreground truncate">
                  {t("couponDesc")}
                </span>
              </div>
            </div>

            {/* Free Shipping tile */}
            <div className="flex items-center gap-2.5 p-2.5 rounded-xl bg-background/80 border border-border/60 shadow-2xs backdrop-blur-xs">
              <div className="size-8 rounded-lg bg-sky-500/10 text-sky-600 dark:text-sky-400 flex items-center justify-center shrink-0">
                <Truck className="size-4" />
              </div>
              <div className="flex flex-col min-w-0">
                <span className="text-xs font-bold text-foreground leading-tight truncate">
                  {t("shippingTitle")}
                </span>
                <span className="text-[10px] text-muted-foreground truncate">
                  {t("shippingDesc")}
                </span>
              </div>
            </div>

            {/* Buyer Guarantee tile */}
            <div className="col-span-2 sm:col-span-1 flex items-center gap-2.5 p-2.5 rounded-xl bg-background/80 border border-border/60 shadow-2xs backdrop-blur-xs">
              <div className="size-8 rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400 flex items-center justify-center shrink-0">
                <ShieldCheck className="size-4" />
              </div>
              <div className="flex flex-col min-w-0">
                <span className="text-xs font-bold text-foreground leading-tight truncate">
                  {t("guaranteeTitle")}
                </span>
                <span className="text-[10px] text-muted-foreground truncate">
                  {t("guaranteeDesc")}
                </span>
              </div>
            </div>
          </div>

          {/* Right: Action CTA */}
          <div className="shrink-0 w-full sm:w-auto flex justify-center">
            <Link
              href="/products"
              className="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl bg-accent-brand text-white font-bold text-xs sm:text-sm hover:bg-accent-brand/90 transition-all shadow-xs hover:shadow-sm cursor-pointer group"
            >
              <span>{t("cta")}</span>
              <ArrowRight className="size-4 rtl:rotate-180 group-hover:translate-x-0.5 rtl:group-hover:-translate-x-0.5 transition-transform" />
            </Link>
          </div>
        </div>
      </div>
    </section>
  );
}
