"use client";

import { useTranslations } from "next-intl";

export default function AboutHero() {
  const t = useTranslations("about");

  return (
    <div className="bg-footer rounded-xl p-8 md:p-14 flex flex-col items-center justify-center text-center">
      <p className="uppercase text-accent-brand text-xs md:text-sm font-semibold mb-3 tracking-widest">
        {t("hero.badge")}
      </p>
      <h1 className="text-title-sm md:text-title-md lg:text-title-lg text-white font-black mb-3 max-w-xl">
        {t("hero.title")}
      </h1>
      <p className="text-muted-foreground text-txt-sm md:text-txt-md max-w-lg">
        {t("hero.description")}
      </p>
    </div>
  );
}
