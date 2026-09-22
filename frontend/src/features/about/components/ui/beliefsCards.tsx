"use client";

import { useTranslations } from "next-intl";
import { LucideIcon } from "lucide-react";

export interface BeliefItem {
  id: string;
  title: string;
  subTitle: string;
  icon: LucideIcon;
}

export default function BeliefsCards({ belief }: { belief: BeliefItem }) {
  const t = useTranslations("about");
  const Icon = belief.icon;

  return (
    <div className="p-4 md:p-6 bg-card rounded-lg border border-border">
      <Icon className="mb-3 text-accent-brand size-6" />
      <p className="mb-2 text-card-foreground font-medium text-txt-sm md:text-txt-md lg:text-txt-lg">
        {t(belief.title)}
      </p>
      <p className="text-muted-foreground text-xs md:text-sm leading-relaxed">
        {t(belief.subTitle)}
      </p>
    </div>
  );
}
