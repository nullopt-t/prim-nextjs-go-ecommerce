"use client";

import { CustomButton } from "@/components/ui/button";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";

const currentDate = new Date().getFullYear();

export default function Hero() {
  const t = useTranslations("home");

  return (
    <div className="p-8 md:p-14 lg:p-16 rounded-2xl bg-footer text-white border border-border/20 shadow-md">
      <p className="text-accent-brand mb-4 font-semibold text-sm tracking-wider uppercase">
        {t("hero.badge", { year: currentDate })}
      </p>
      <h1 className="mb-4 text-white font-black text-3xl sm:text-4xl md:text-5xl lg:text-6xl tracking-tight max-w-2xl">
        <span>{t("hero.title")} </span>
        <span className="text-accent-brand">{t("hero.delivered")}</span>
      </h1>
      <p className="text-neutral-400 mb-8 max-w-xl text-sm md:text-base leading-relaxed">
        {t("hero.description")}
      </p>
      <div className="flex flex-wrap gap-4">
        <Link href="/products" className="inline-block">
          <div className="px-6 py-3 bg-accent-brand hover:opacity-90 text-white rounded-xl font-bold text-sm transition-opacity shadow-sm">
            {t("hero.shopNow")}
          </div>
        </Link>
        <Link href="/products" className="inline-block">
          <div className="px-6 py-3 bg-white/10 text-white hover:bg-white/20 rounded-xl font-bold text-sm transition-colors border border-white/15">
            {t("hero.viewDeals")}
          </div>
        </Link>
      </div>
    </div>
  );
}
