"use client";

import { HiMiniArrowLongRight } from "react-icons/hi2";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";

export default function SectionTitle({ title, link = "/products" }: { title: string; link?: string }) {
  const t = useTranslations("home");

  return (
    <div className="flex justify-between items-center">
      <p className="font-medium text-title-sm md:text-title-md text-foreground">
        {title.includes(".") ? t(title) : title}
      </p>
      <Link
        href={link}
        className="flex gap-2 items-center text-txt-sm md:text-txt-md lg:text-txt-lg cursor-pointer text-accent-brand hover:underline hover:underline-offset-4"
      >
        <span>{t("categories.seeAll")}</span>
        <HiMiniArrowLongRight className="size-5" />
      </Link>
    </div>
  );
}
