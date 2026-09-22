"use client";

import { useTranslations } from "next-intl";
import { Title } from "@/components/ui/title";

export default function BeliefsDescription() {
  const t = useTranslations("about");

  return (
    <div>
      <Title title={t("belief.title")} />
      <p className="text-txt-sm md:text-txt-md lg:text-txt-lg text-muted-foreground leading-relaxed">
        {t("belief.description")}
      </p>
    </div>
  );
}
