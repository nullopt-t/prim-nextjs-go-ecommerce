"use client";

import { CustomButton } from "@/components/ui/button";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";

const currentDate = new Date().getFullYear();

export default function Hero() {
  const t = useTranslations("home");

  return (
    <div className="p-8 md:p-14 lg:p-16 rounded-xl bg-footer text-white">
      <p className="text-accent-brand mb-4 font-semibold text-sm">
        {t("hero.badge", { year: currentDate })}
      </p>
      <h1 className="mb-4 text-white font-black text-title-sm md:text-title-md lg:text-title-lg max-w-2xl">
        <span>{t("hero.title")} </span>
        <span className="text-accent-brand">{t("hero.delivered")}</span>
      </h1>
      <p className="text-muted-foreground mb-8 max-w-xl text-txt-sm md:text-txt-md">
        {t("hero.description")}
      </p>
      <div className="flex flex-wrap gap-4">
        <Link href="/products" className="w-36 h-12 inline-block">
          <div className="w-full h-full bg-accent-brand hover:opacity-90 text-white rounded-md">
            <CustomButton text={t("hero.shopNow")} />
          </div>
        </Link>
        <Link href="/products" className="w-36 h-12 inline-block">
          <div className="w-full h-full bg-secondary text-secondary-foreground hover:opacity-90 rounded-md">
            <CustomButton text={t("hero.viewDeals")} />
          </div>
        </Link>
      </div>
    </div>
  );
}
